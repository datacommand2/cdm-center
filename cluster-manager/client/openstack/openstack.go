package openstack

import (
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack/train"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack/train/constant"
	cmsConstant "github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/storage"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	client.RegisterClusterClientCreationFunc(cmsConstant.ClusterTypeOpenstack, New)
}

// Client 는 Openstack Client 구조체
// TODO: 버전 별 connection 을 다르게 둬야 함. 하지만 지금은 train 만 고려해서 구현 진행 예정.
type Client struct {
	url        string
	credential string
	tenantID   string
	conn       *train.Connection
}

func (c *Client) getAdminRoleTenantID() (string, error) {

	auth := fmt.Sprintf("{ \"auth\": { \"identity\": %s}}", c.credential)
	if err := c.conn.Authentication(train.AuthenticationRequest{URL: c.url, Auth: auth}); err != nil {
		return "", err
	}

	defer func() {
		c.conn.Mutex.Lock()
		c.conn.Token = nil
		c.conn.Mutex.Unlock()
	}()

	projectScopesRsp, err := train.GetAvailableProjectScopes(c, c.url)
	if err != nil {
		return "", err
	}

	for _, project := range projectScopesRsp.Projects {
		auth = fmt.Sprintf("{ \"auth\": { \"identity\": %s, \"scope\": { \"project\": { \"id\": \"%s\"}}}}", c.credential, project.UUID)

		if err := c.conn.Authentication(train.AuthenticationRequest{URL: c.url, Auth: auth}); err != nil {
			return "", err
		}
		if err = c.conn.CheckAuthority(); err == nil {
			return project.UUID, nil
		}
	}

	return "", err
}

func (c *Client) getTenantID() (string, error) {
	if c.tenantID != "" {
		return c.tenantID, nil
	}

	p, err := c.getAdminRoleTenantID()
	if err != nil {
		return "", err
	}

	return p, nil
}

// Connect Openstack 에 연결한다.
func (c *Client) Connect() error {
	if c.conn.IsConnected() {
		logger.Debug("Already connected to Openstack server.")
		return nil
	}

	tenantID, err := c.getTenantID()
	if err != nil {
		return err
	}

	auth := fmt.Sprintf("{ \"auth\": { \"identity\": %s, \"scope\": { \"project\": { \"id\": \"%s\"}}}}", c.credential, tenantID)
	if err := c.conn.Authentication(train.AuthenticationRequest{URL: c.url, Auth: auth}); err != nil {
		logger.Errorf("[Connect] Could not connect to Openstack. Cause: %+v", err)
		return err
	}

	//c.conn.Close = make(chan interface{})
	//go c.conn.Reconnect(c.url, auth)

	return nil
}

// Close Openstack 연결을 종료한다.
func (c *Client) Close() error {
	if !c.conn.IsConnected() {
		return nil
	}
	//close(c.conn.Close)

	c.conn.Mutex.Lock()
	defer c.conn.Mutex.Unlock()

	c.conn.Token = nil

	return nil
}

// CheckAuthority 고객사에서 제공한 credential 내 사용자에게 api 요청 권한이 있는지 확인한다.
func (c *Client) CheckAuthority() error {
	if !c.conn.IsConnected() {
		return client.NotConnected(cmsConstant.ClusterTypeOpenstack)
	}

	return c.conn.CheckAuthority()
}

// GetTokenKey 토큰 키 조회
func (c *Client) GetTokenKey() string {
	return c.conn.GetTokenKey()
}

// GetEndpoint 엔드포인트 조회
func (c *Client) GetEndpoint(serviceType string) (string, error) {
	return c.conn.GetEndpoint(serviceType)
}

// New 는 Openstack Client 구조체를 생성하는 함수이다.
func New(url, credential, tenantID string) client.Client {
	return &Client{
		url:        url,
		credential: credential,
		conn:       &train.Connection{Mutex: sync.Mutex{}},
		tenantID:   tenantID,
	}
}

func (c *Client) updateQuotas(uuid string, quotaList []model.ClusterQuota) ([]model.ClusterQuota, error) {
	var computeReq = train.UpdateComputeQuotaRequest{TenantUUID: uuid}
	var networkReq = train.UpdateNetworkQuotaRequest{TenantUUID: uuid}
	var storageReq = train.UpdateStorageQuotaRequest{TenantUUID: uuid}

	var computeRsp *train.UpdateComputeQuotaResponse
	var networkRsp *train.UpdateNetworkQuotaResponse
	var storageRsp *train.UpdateStorageQuotaResponse

	var err error
	var result []model.ClusterQuota

	for _, quota := range quotaList {
		computeReq.Set(quota.Key, quota.Value)
		networkReq.Set(quota.Key, quota.Value)
		storageReq.Set(quota.Key, quota.Value)
	}

	if computeReq.QuotaSet != (train.UpdateComputeQuotaResult{}) {
		if computeRsp, err = train.UpdateComputeQuotas(c, computeReq); err != nil {
			return nil, err
		}
	}
	if networkReq.QuotaSet != (train.UpdateNetworkQuotaResult{}) {
		if networkRsp, err = train.UpdateNetworkQuotaForAProject(c, networkReq); err != nil {
			return nil, err
		}
	}
	if storageReq.QuotaSet != (train.UpdateStorageQuotaResult{}) {
		if storageRsp, err = train.UpdateStorageQuotasForAProject(c, storageReq); err != nil {
			return nil, err
		}
	}

	for k, v := range computeRsp.ToMap() {
		result = append(result, model.ClusterQuota{Key: k, Value: v})
	}
	for k, v := range networkRsp.ToMap() {
		result = append(result, model.ClusterQuota{Key: k, Value: v})
	}
	for k, v := range storageRsp.ToMap() {
		result = append(result, model.ClusterQuota{Key: k, Value: v})
	}
	return result, nil
}

// CreateTenant 테넌트 생성
func (c *Client) CreateTenant(req client.CreateTenantRequest) (*client.CreateTenantResponse, error) {
	var description string
	if req.Tenant.Description != nil {
		description = *req.Tenant.Description
	}
	projectResponse, err := train.CreateProject(c, train.CreateProjectRequest{
		Tenant: train.CreateProjectResult{
			Name:        req.Tenant.Name,
			Description: description,
			Enabled:     req.Tenant.Enabled,
		},
	})
	if err != nil {
		return nil, err
	}

	var rollback = true
	defer func() {
		if rollback {
			if err := train.DeleteProject(c, train.DeleteProjectRequest{
				ProjectUUID: projectResponse.Tenant.UUID,
			}); err != nil {
				logger.Errorf("Could not delete tenant. Cause: %+v", err)
			}
		}
	}()

	// get admin role uuid, 참조: CheckAuthority
	var roleUUID string
	for _, r := range c.conn.Token.Payload.Roles {
		if r.Name == "admin" {
			roleUUID = r.ID
			break
		}
	}
	if roleUUID == "" {
		return nil, errors.Unknown(errors.New("not found admin role"))
	}

	err = train.AssignRoleToUserOnProject(c, train.AssignRoleToGroupOnProjectRequest{
		ProjectUUID: projectResponse.Tenant.UUID,
		UserUUID:    c.conn.Token.Payload.User.ID,
		RoleUUID:    roleUUID,
	})
	if err != nil {
		return nil, err
	}

	modelQuotaList, err := c.updateQuotas(projectResponse.Tenant.UUID, req.QuotaList)
	if err != nil {
		return nil, err
	}

	rollback = false
	return &client.CreateTenantResponse{
		Tenant: model.ClusterTenant{
			UUID:        projectResponse.Tenant.UUID,
			Name:        projectResponse.Tenant.Name,
			Description: &projectResponse.Tenant.Description,
			Enabled:     projectResponse.Tenant.Enabled,
			Raw:         &projectResponse.Raw,
		},
		QuotaList: modelQuotaList,
	}, nil
}

// GetAvailabilityZoneList 가용 구역 목록 조회
func (c *Client) GetAvailabilityZoneList() (*client.GetAvailabilityZoneListResponse, error) {
	rsp, err := train.GetAvailabilityZoneInformation(c)
	if err != nil {
		return nil, err
	}

	var azList []client.GetAvailabilityZoneResult
	for _, az := range rsp.AvailabilityZones {
		azList = append(azList, client.GetAvailabilityZoneResult{
			AvailabilityZone: model.ClusterAvailabilityZone{
				Name:      az.ZoneName,
				Available: az.ZoneState.Available,
				Raw:       &az.Raw,
			},
		})
	}

	return &client.GetAvailabilityZoneListResponse{ResultList: azList}, nil
}

// GetHypervisorList 하이퍼바이저 목록 조회
func (c *Client) GetHypervisorList() (*client.GetHypervisorListResponse, error) {
	rsp, err := train.ListHypervisors(c)
	if err != nil {
		return nil, err
	}

	var hypervisorList []client.GetHypervisorResult
	for _, Hypervisor := range rsp.Hypervisors {
		hypervisorList = append(hypervisorList, client.GetHypervisorResult{
			Hypervisor: model.ClusterHypervisor{
				UUID: Hypervisor.UUID,
			},
		})
	}

	return &client.GetHypervisorListResponse{ResultList: hypervisorList}, nil
}

// GetHypervisor 하이퍼바이저 상세 조회
func (c *Client) GetHypervisor(req client.GetHypervisorRequest) (*client.GetHypervisorResponse, error) {
	hypervisorRsp, err := train.ShowHypervisorDetails(c, train.ShowHypervisorDetailsRequest{HypervisorID: req.Hypervisor.UUID})
	if err != nil {
		return nil, err
	}

	csRsp, err := train.ListComputeServices(c, train.ListComputeServicesRequest{Binary: "nova-compute", HostName: hypervisorRsp.Hypervisor.Service.Host})
	if err != nil {
		return nil, err
	}

	return &client.GetHypervisorResponse{Result: client.GetHypervisorResult{
		Hypervisor: model.ClusterHypervisor{
			UUID:           hypervisorRsp.Hypervisor.UUID,
			TypeCode:       hypervisorRsp.Hypervisor.TypeCode,
			Hostname:       hypervisorRsp.Hypervisor.Hostname,
			IPAddress:      hypervisorRsp.Hypervisor.IPAddress,
			VcpuTotalCnt:   hypervisorRsp.Hypervisor.VcpuTotalCnt,
			VcpuUsedCnt:    hypervisorRsp.Hypervisor.VcpuUsedCnt,
			MemTotalBytes:  hypervisorRsp.Hypervisor.MemTotalMibiBytes * 1024 * 1024,
			MemUsedBytes:   hypervisorRsp.Hypervisor.MemUsedMibiBytes * 1024 * 1024,
			DiskTotalBytes: hypervisorRsp.Hypervisor.DiskTotalGibiBytes * 1024 * 1024 * 1024,
			DiskUsedBytes:  hypervisorRsp.Hypervisor.DiskUsedGibiBytes * 1024 * 1024 * 1024,
			State:          hypervisorRsp.Hypervisor.State,
			Status:         hypervisorRsp.Hypervisor.Status,
			Raw:            &hypervisorRsp.Raw,
		},
		AvailabilityZone: model.ClusterAvailabilityZone{
			Name: csRsp.ComputeServices[0].ZoneName,
		},
	}}, nil
}

// GetInstanceList 인스턴스 목록 조회
func (c *Client) GetInstanceList() (*client.GetInstanceListResponse, error) {
	result, err := train.ListServers(c)
	if err != nil {
		return nil, err
	}

	var rsp client.GetInstanceListResponse
	for _, server := range result.Servers {
		rsp.ResultList = append(rsp.ResultList, client.GetInstanceResult{
			Instance: model.ClusterInstance{
				UUID: server.UUID,
			},
		})
	}

	return &rsp, nil
}

// GetInstance 인스턴스 상세 조회
func (c *Client) GetInstance(req client.GetInstanceRequest) (*client.GetInstanceResponse, error) {
	result, err := train.ShowServerDetails(c, train.ShowServerDetailsRequest{ServerID: req.Instance.UUID})
	if err != nil {
		if errors.Equal(err, client.ErrUnAuthenticated) {
			logger.Errorf("[Openstack-ShowServerDetails] tenantID: %d. err: %+v", c.tenantID, err)
		}
		return nil, err
	}

	// get keypair
	var keypair *model.ClusterKeypair
	if result.Server.KeyName != nil {
		keypair = &model.ClusterKeypair{
			Name: *result.Server.KeyName,
		}
	}

	// get instance state
	state, ok := InstanceStateMap[result.Server.State]
	if !ok {
		return nil, UnknownInstanceState(result.Server.Name, result.Server.State)
	}

	// get hypervisor
	hypervisorListRsp, err := train.ListHypervisors(c)
	if err != nil {
		return nil, err
	}

	var instanceHypervisor model.ClusterHypervisor
	for _, hypervisor := range hypervisorListRsp.Hypervisors {
		for _, server := range hypervisor.Servers {
			if server.UUID == result.Server.UUID {
				instanceHypervisor = model.ClusterHypervisor{
					UUID: hypervisor.UUID,
				}
				break
			}
		}
		if instanceHypervisor.UUID != "" {
			break
		}
	}
	if instanceHypervisor.UUID == "" {
		return nil, NotFoundInstanceHypervisor(result.Server.Name)
	}

	// get securityGroup
	var instanceSecurityGroups []model.ClusterSecurityGroup
	sgListRsp, err := train.ListSecurityGroupsByServer(c, train.ListSecurityGroupsByServerRequest{ServerID: result.Server.UUID})
	if err != nil {
		return nil, err
	}

	for _, sg := range sgListRsp.SecurityGroups {
		instanceSecurityGroups = append(instanceSecurityGroups, model.ClusterSecurityGroup{
			UUID: sg.UUID,
		})
	}

	// get instance spec
	flavorsRsp, err := train.ListFlavors(c)
	if err != nil {
		return nil, err
	}

	var instanceSpec model.ClusterInstanceSpec
	for _, flavor := range flavorsRsp.Flavors {
		if flavor.Name == result.Server.Flavor.Name {
			instanceSpec = model.ClusterInstanceSpec{
				UUID: flavor.UUID,
			}
			break
		}
	}
	if instanceSpec.UUID == "" {
		return nil, NotFoundInstanceSpec(result.Server.Name, result.Server.Flavor.Name)
	}

	var instanceVolumes []client.GetInstanceVolumeResult
	for _, attachedVolume := range result.Server.AttachedVolumes {
		attachVolumeRsp, err := train.ShowDetailVolumeAttachment(c, train.ShowDetailVolumeAttachmentRequest{ServerID: result.Server.UUID, VolumeID: attachedVolume.VolumeID})
		if err != nil {
			return nil, err
		}
		instanceVolumes = append(instanceVolumes, client.GetInstanceVolumeResult{
			InstanceVolume: model.ClusterInstanceVolume{
				DevicePath: attachVolumeRsp.VolumeAttachment.DevicePath,
			},
			Volume: model.ClusterVolume{
				UUID: attachVolumeRsp.VolumeAttachment.VolumeID,
			},
		})
	}

	var instanceNetworkFloatingIP *model.ClusterFloatingIP
	var instanceNetworks []client.GetInstanceNetworkResult
	instancePortInterfacesRsp, err := train.ListPortInterfaces(c, train.ListPortInterfacesRequest{ServerID: result.Server.UUID})
	if err != nil {
		return nil, err
	}

	for _, portInterface := range instancePortInterfacesRsp.PortInterfaces {
		for _, fixedIP := range portInterface.FixedIPs {
			floatingIPsRsp, err := train.ListFloatingIPsByPortInterface(c, train.ListFloatingIPsByPortInterfaceRequest{PortID: portInterface.PortID, FixedIP: fixedIP.IPAddress})
			if err != nil {
				return nil, err
			}

			for _, floatingIP := range floatingIPsRsp.FloatingIPs {
				instanceNetworkFloatingIP = &model.ClusterFloatingIP{UUID: floatingIP.UUID}
			}

			subnetRsp, err := train.ShowSubnetDetails(c, train.ShowSubnetDetailsRequest{SubnetID: fixedIP.SubnetUUID})
			if err != nil {
				return nil, err
			}

			instanceNetworks = append(instanceNetworks, client.GetInstanceNetworkResult{
				InstanceNetwork: model.ClusterInstanceNetwork{
					DhcpFlag:  subnetRsp.Subnet.DHCPEnabled,
					IPAddress: fixedIP.IPAddress,
				},
				Subnet: model.ClusterSubnet{
					UUID: fixedIP.SubnetUUID,
				},
				Network: model.ClusterNetwork{
					UUID: portInterface.NetworkID,
				},
				FloatingIP: instanceNetworkFloatingIP,
			})
		}
	}

	return &client.GetInstanceResponse{
		Result: client.GetInstanceResult{
			Instance: model.ClusterInstance{
				UUID:        result.Server.UUID,
				Name:        result.Server.Name,
				Description: result.Server.Description,
				State:       state,
				Status:      result.Server.Status,
				Raw:         &result.Raw,
			},
			Tenant: model.ClusterTenant{
				UUID: result.Server.TenantID,
			},
			InstanceSpec: instanceSpec,
			KeyPair:      keypair,
			Hypervisor:   instanceHypervisor,
			AvailabilityZone: model.ClusterAvailabilityZone{
				Name: result.Server.AzName,
			},
			SecurityGroupList:   instanceSecurityGroups,
			InstanceVolumeList:  instanceVolumes,
			InstanceNetworkList: instanceNetworks,
		},
	}, nil
}

// GetInstanceSpec 인스턴스 spec 상세 조회
func (c *Client) GetInstanceSpec(req client.GetInstanceSpecRequest) (*client.GetInstanceSpecResponse, error) {
	result, err := train.ShowFlavorDetails(c, train.ShowFlavorDetailsRequest{FlavorID: req.InstanceSpec.UUID})
	if err != nil {
		return nil, err
	}

	var extraSpecs []model.ClusterInstanceExtraSpec
	for k := range result.Flavor.ExtraSpec {
		extraSpecs = append(extraSpecs, model.ClusterInstanceExtraSpec{
			Key: k,
		})
	}

	return &client.GetInstanceSpecResponse{
		Result: client.GetInstanceSpecResult{
			InstanceSpec: model.ClusterInstanceSpec{
				UUID:                result.Flavor.UUID,
				Name:                result.Flavor.Name,
				Description:         result.Flavor.Description,
				VcpuTotalCnt:        result.Flavor.VcpuTotalCnt,
				MemTotalBytes:       result.Flavor.MemTotalMibiBytes * 1024 * 1024,
				DiskTotalBytes:      result.Flavor.DiskTotalGibiBytes * 1024 * 1024 * 1024,
				SwapTotalBytes:      result.Flavor.SwapTotalMibiBytes * 1024 * 1024,
				EphemeralTotalBytes: result.Flavor.EphemeralTotalGibiBytes * 1024 * 1024 * 1024,
			},
			ExtraSpecList: extraSpecs,
		},
	}, nil
}

// GetInstanceExtraSpec 인스턴스 extra spec 상세 조회
func (c *Client) GetInstanceExtraSpec(req client.GetInstanceExtraSpecRequest) (*client.GetInstanceExtraSpecResponse, error) {
	result, err := train.ShowFlavorExtraSpec(c, train.ShowFlavorExtraSpecRequest{FlavorID: req.InstanceSpec.UUID, ExtraSpecKey: req.InstanceExtraSpec.Key})
	if err != nil {
		return nil, err
	}

	return &client.GetInstanceExtraSpecResponse{
		Result: client.GetInstanceExtraSpecResult{
			InstanceExtraSpec: model.ClusterInstanceExtraSpec{
				Key:   result.Key,
				Value: result.Value,
			},
		},
	}, nil
}

// DeleteInstance 인스턴스 삭제
func (c *Client) DeleteInstance(req client.DeleteInstanceRequest) error {
	var err error

	// 포트 목록 조회
	rsp, err := train.ListPortInterfaces(c, train.ListPortInterfacesRequest{ServerID: req.Instance.UUID})
	if err != nil {
		return err
	}

	defer func() {
		// 포트 삭제
		for _, p := range rsp.PortInterfaces {
			if err = train.DeletePort(c, train.DeletePortRequest{PortID: p.PortID}); err != nil {
				logger.Warnf("Could not delete created port. uuid: %s, Cause: %+v", p.PortID, err)
			}
		}
	}()

	// 인스턴스 삭제
	if err = train.DeleteServer(c, train.DeleteServerRequest{ServerID: req.Instance.UUID}); err != nil {
		// 인스턴스 강제 삭제
		if err = train.ForceDeleteServer(c, train.DeleteServerRequest{ServerID: req.Instance.UUID}); err != nil {
			return err
		}
	}

	return nil
}

// DeleteInstanceSpec 인스턴스 스팩 삭제
func (c *Client) DeleteInstanceSpec(req client.DeleteInstanceSpecRequest) error {
	return train.DeleteFlavor(c, train.DeleteFlavorRequest{
		FlavorID: req.InstanceSpec.UUID,
	})
}

// StartInstance 인스턴스 기동
func (c *Client) StartInstance(req client.StartInstanceRequest) error {
	return train.StartServer(c, train.StartServerRequest{
		ServerID: req.Instance.UUID,
	})
}

// StopInstance 인스턴스 중지
func (c *Client) StopInstance(req client.StopInstanceRequest) error {
	return train.StopServer(c, train.StopServerRequest{
		ServerID: req.Instance.UUID,
	})
}

// GetKeyPairList Keypair 목록 조회
func (c *Client) GetKeyPairList() (*client.GetKeyPairListResponse, error) {
	result, err := train.ListKeypairs(c)
	if err != nil {
		return nil, err
	}

	var rsp client.GetKeyPairListResponse
	for _, l := range result.Keypairs {
		rsp.ResultList = append(rsp.ResultList, client.GetKeyPairResult{
			KeyPair: model.ClusterKeypair{
				Name: l.Keypair.Name,
			},
		})
	}

	return &rsp, nil
}

// GetKeyPair Keypair 상세 조회
func (c *Client) GetKeyPair(req client.GetKeyPairRequest) (*client.GetKeyPairResponse, error) {
	result, err := train.ShowKeypairDetails(c, train.ShowKeypairDetailsRequest{KeypairName: req.KeyPair.Name})
	if err != nil {
		return nil, err
	}

	typeCode := cmsConstant.OpenstackKeypairTypeSSH
	if result.Keypair.TypeCode == "" {
		// openstack version 2.2 이하에서는 type 정보가 따로 들어오지 않고
		// public_key 가 ssh-rsa 로 시작하는지에 따라 구분할 수 있음
		if strings.Index(result.Keypair.PublicKey, "ssh-rsa") != 0 {
			typeCode = cmsConstant.OpenstackKeypairTypeX509
		}
	} else {
		if result.Keypair.TypeCode == "x509" {
			typeCode = cmsConstant.OpenstackKeypairTypeX509
		}
	}

	return &client.GetKeyPairResponse{
		Result: client.GetKeyPairResult{
			KeyPair: model.ClusterKeypair{
				Name:        result.Keypair.Name,
				Fingerprint: result.Keypair.Fingerprint,
				PublicKey:   result.Keypair.PublicKey,
				TypeCode:    typeCode,
			},
		},
	}, nil
}

// DeleteTenant 테넌트 삭제
func (c *Client) DeleteTenant(req client.DeleteTenantRequest) error {
	return train.DeleteProject(c, train.DeleteProjectRequest{
		ProjectUUID: req.Tenant.UUID,
	})
}

// GetTenantList 테넌트 목록 조회
func (c *Client) GetTenantList() (*client.GetTenantListResponse, error) {
	result, err := train.ListProjects(c)
	if err != nil {
		return nil, err
	}

	var rsp client.GetTenantListResponse
	for _, project := range result.Projects {
		rsp.ResultList = append(rsp.ResultList, client.GetTenantResult{
			Tenant: model.ClusterTenant{
				UUID: project.UUID,
			},
		})
	}

	return &rsp, nil
}

// GetTenant 테넌트 상세 조회
func (c *Client) GetTenant(req client.GetTenantRequest) (*client.GetTenantResponse, error) {
	result, err := train.ShowProjectDetails(c, train.ShowProjectDetailsRequest{ProjectID: req.Tenant.UUID})
	if err != nil {
		return nil, err
	}

	var quotaList []model.ClusterQuota
	// get compute quota
	computeQuota, err := train.ShowComputeQuota(c, train.ShowComputeQuotaRequest{ProjectID: req.Tenant.UUID})
	if err != nil {
		return nil, err
	}
	for k, v := range computeQuota.QuotaSets {
		quotaList = append(quotaList, model.ClusterQuota{Key: k, Value: v})
	}

	// get network quota
	networkQuota, err := train.ShowNetworkQuota(c, train.ShowNetworkQuotaRequest{ProjectID: req.Tenant.UUID})
	if err != nil {
		return nil, err
	}
	for k, v := range networkQuota.QuotaSets {
		quotaList = append(quotaList, model.ClusterQuota{Key: k, Value: v})
	}

	// get storage quota
	storageQuota, err := train.ShowStorageQuota(c, train.ShowStorageQuotaRequest{ProjectID: req.Tenant.UUID})
	if err != nil {
		return nil, err
	}
	for k, v := range storageQuota.QuotaSets {
		quotaList = append(quotaList, model.ClusterQuota{Key: k, Value: v})
	}

	return &client.GetTenantResponse{
		Result: client.GetTenantResult{
			Tenant: model.ClusterTenant{
				UUID:        result.Project.UUID,
				Name:        result.Project.Name,
				Description: result.Project.Description,
				Enabled:     result.Project.Enabled,
				Raw:         &result.Raw,
			},
			QuotaList: quotaList,
		},
	}, nil
}

// CreateFloatingIP 부동 IP 생성
func (c *Client) CreateFloatingIP(req client.CreateFloatingIPRequest) (*client.CreateFloatingIPResponse, error) {
	ip, err := train.CreateFloatingIP(c, train.CreateFloatingIPRequest{
		FloatingIP: train.CreateFloatingIPResult{
			ProjectUUID:       req.Tenant.UUID,
			FloatingNetworkID: req.Network.UUID,
			FloatingIPAddress: req.FloatingIP.IPAddress,
			Description:       req.FloatingIP.Description,
		},
	})
	if err != nil {
		return nil, err
	}

	if ip.FloatingIP.Status == "ERROR" {
		if err := train.DeleteFloatingIP(c, train.DeleteFloatingIPRequest{
			FloatingUUID: ip.FloatingIP.UUID,
		}); err != nil {
			logger.Warnf("Could not delete created floating ip. uuid: %s, Cause: %+v", ip.FloatingIP.UUID, err)
		}
		return nil, client.RemoteServerError(errors.New("error occurred during creation of floating ip"))
	}

	return &client.CreateFloatingIPResponse{
		FloatingIP: model.ClusterFloatingIP{
			UUID:        ip.FloatingIP.UUID,
			Description: ip.FloatingIP.Description,
			IPAddress:   ip.FloatingIP.FloatingIPAddress,
			Status:      ip.FloatingIP.Status,
			Raw:         &ip.Raw,
		},
	}, nil
}

// DeleteFloatingIP 부동 IP 삭제
func (c *Client) DeleteFloatingIP(req client.DeleteFloatingIPRequest) error {
	return train.DeleteFloatingIP(c, train.DeleteFloatingIPRequest{
		FloatingUUID: req.FloatingIP.UUID,
	})
}

// GetFloatingIP floating ip 상세 조회
func (c *Client) GetFloatingIP(req client.GetFloatingIPRequest) (*client.GetFloatingIPResponse, error) {
	result, err := train.ShowFloatingIPDetails(c, train.ShowFloatingIPDetailsRequest{FloatingIPID: req.FloatingIP.UUID})
	if err != nil {
		return nil, err
	}

	return &client.GetFloatingIPResponse{
		Result: client.GetFloatingIPResult{
			FloatingIP: model.ClusterFloatingIP{
				UUID:        result.FloatingIP.UUID,
				Description: result.FloatingIP.Description,
				IPAddress:   result.FloatingIP.IPAddress,
				Status:      result.FloatingIP.Status,
				Raw:         &result.Raw,
			},
			Network: model.ClusterNetwork{
				UUID: result.FloatingIP.NetworkID,
			},
			Tenant: model.ClusterTenant{
				UUID: result.FloatingIP.ProjectID,
			},
		},
	}, nil
}

// CreateNetwork 네트워크 생성
func (c *Client) CreateNetwork(req client.CreateNetworkRequest) (*client.CreateNetworkResponse, error) {
	var stateBool bool
	switch req.Network.State {
	case "up":
		stateBool = true
	case "down":
		stateBool = false
	}

	rsp, err := train.CreateNetwork(c, train.CreateNetworkRequest{
		Network: train.CreateNetworkResult{
			ProjectUUID:    req.Tenant.UUID,
			Name:           req.Network.Name,
			AdminStateUp:   stateBool,
			RouterExternal: req.Network.ExternalFlag,
			Description:    req.Network.Description,
		},
	})
	if err != nil {
		return nil, err
	}

	var stateString string
	if rsp.Network.AdminStateUp {
		stateString = "up"
	} else {
		stateString = "down"
	}
	return &client.CreateNetworkResponse{
		Network: model.ClusterNetwork{
			UUID:         rsp.Network.UUID,
			Name:         rsp.Network.Name,
			Description:  rsp.Network.Description,
			TypeCode:     rsp.Network.Type,
			ExternalFlag: rsp.Network.RouterExternal,
			State:        stateString,
			Status:       rsp.Network.Status,
			Raw:          &rsp.Raw,
		},
	}, nil
}

// CreateSubnet 서브넷 생성
func (c *Client) CreateSubnet(req client.CreateSubnetRequest) (*client.CreateSubnetResponse, error) {
	var nameservers []string
	for _, nameserver := range req.NameserverList {
		nameservers = append(nameservers, nameserver.Nameserver)
	}

	var pools []train.SubnetAllocationPool
	for _, pool := range req.PoolsList {
		pools = append(pools, train.SubnetAllocationPool{
			Start: pool.StartIPAddress,
			End:   pool.EndIPAddress,
		})
	}

	ipVersion, err := train.IPVersion(req.Subnet.NetworkCidr)
	if err != nil {
		return nil, err
	}
	rsp, err := train.CreateSubnet(c, train.CreateSubnetRequest{
		Subnet: train.CreateSubnetResult{
			NetworkUUID:     req.Network.UUID,
			IPVersion:       ipVersion,
			CIDR:            req.Subnet.NetworkCidr,
			ProjectID:       req.Tenant.UUID,
			Name:            req.Subnet.Name,
			EnableDHCP:      req.Subnet.DHCPEnabled,
			DNSNameservers:  nameservers,
			AllocationPools: pools,
			GatewayIP:       req.Subnet.GatewayIPAddress,
			Description:     req.Subnet.Description,
			IPv6AddressMode: req.Subnet.Ipv6AddressModeCode,
			IPv6RAMode:      req.Subnet.Ipv6RaModeCode,
		},
	})
	if err != nil {
		return nil, err
	}

	var modelNameServers []model.ClusterSubnetNameserver
	for _, nameserver := range rsp.Subnet.DNSNameservers {
		modelNameServers = append(modelNameServers, model.ClusterSubnetNameserver{
			Nameserver: nameserver,
		})
	}
	var modelPools []model.ClusterSubnetDHCPPool
	for _, pool := range rsp.Subnet.AllocationPools {
		modelPools = append(modelPools, model.ClusterSubnetDHCPPool{
			StartIPAddress: pool.Start,
			EndIPAddress:   pool.End,
		})
	}
	return &client.CreateSubnetResponse{
		Subnet: model.ClusterSubnet{
			UUID:                rsp.Subnet.UUID,
			Name:                rsp.Subnet.Name,
			Description:         rsp.Subnet.Description,
			NetworkCidr:         rsp.Subnet.CIDR,
			DHCPEnabled:         rsp.Subnet.EnableDHCP,
			GatewayEnabled:      rsp.Subnet.GatewayIP != nil,
			GatewayIPAddress:    rsp.Subnet.GatewayIP,
			Ipv6AddressModeCode: rsp.Subnet.IPv6AddressMode,
			Ipv6RaModeCode:      rsp.Subnet.IPv6RAMode,
			Raw:                 &rsp.Raw,
		},
		NameserverList: modelNameServers,
		PoolsList:      modelPools,
	}, nil
}

// DeleteNetwork 네트워크 삭제
func (c *Client) DeleteNetwork(req client.DeleteNetworkRequest) error {
	return train.DeleteNetwork(c, train.DeleteNetworkRequest{
		NetworkUUID: req.Network.UUID,
	})
}

// DeleteSubnet 서브넷 삭제
func (c *Client) DeleteSubnet(req client.DeleteSubnetRequest) error {
	return train.DeleteSubnet(c, train.DeleteSubnetRequest{
		SubnetUUID: req.Subnet.UUID,
	})
}

// GetNetworkList 네트워크 목록 조회
func (c *Client) GetNetworkList() (*client.GetNetworkListResponse, error) {
	result, err := train.ListNetworks(c)
	if err != nil {
		return nil, err
	}

	var rsp client.GetNetworkListResponse
	for _, network := range result.Networks {
		rsp.ResultList = append(rsp.ResultList, client.GetNetworkResult{
			Network: model.ClusterNetwork{
				UUID: network.UUID,
			},
		})
	}

	return &rsp, nil
}

// GetNetwork 네트워크 상세 조회
func (c *Client) GetNetwork(req client.GetNetworkRequest) (*client.GetNetworkResponse, error) {

	//get network
	result, err := train.ShowNetworkDetails(c, train.ShowNetworkDetailsRequest{NetworkID: req.Network.UUID})
	if err != nil {
		return nil, err
	}

	var state string
	if result.Network.State {
		state = "up"
	} else {
		state = "down"
	}

	//get subnets
	var subnetList []model.ClusterSubnet
	for _, subnetID := range result.Network.Subnets {
		subnetList = append(subnetList, model.ClusterSubnet{UUID: subnetID})
	}

	//get floating ips
	var floatingIPList []model.ClusterFloatingIP
	floatingIPs, err := train.ListFloatingIPs(c, train.ListFloatingIPsRequest{NetworkID: req.Network.UUID})
	if err != nil {
		return nil, err
	}
	for _, floatingIP := range floatingIPs.FloatingIPs {
		floatingIPList = append(floatingIPList, model.ClusterFloatingIP{UUID: floatingIP.UUID})
	}

	return &client.GetNetworkResponse{
		Result: client.GetNetworkResult{
			Network: model.ClusterNetwork{
				UUID:         result.Network.UUID,
				Name:         result.Network.Name,
				Description:  result.Network.Description,
				TypeCode:     result.Network.TypeCode,
				ExternalFlag: result.Network.ExternalFlag,
				State:        state,
				Status:       result.Network.Status,
				Raw:          &result.Raw,
			},
			Tenant: model.ClusterTenant{
				UUID: result.Network.ProjectID,
			},
			SubnetList:     subnetList,
			FloatingIPList: floatingIPList,
		},
	}, nil
}

// CreateRouter 라우터 생성
func (c *Client) CreateRouter(req client.CreateRouterRequest) (*client.CreateRouterResponse, error) {
	var boolState bool
	switch req.Router.State {
	case "up", "":
		boolState = true
	case "down":
		boolState = false
	}

	var networkUUID string
	if len(req.ExternalRoutingInterfaceList) > 0 {
		networkUUID = req.ExternalRoutingInterfaceList[0].Network.UUID
	}
	var fixedIPs []train.FixedIP
	for _, routingInterface := range req.ExternalRoutingInterfaceList {
		// openstack 외부 router 인터페이스의 network 는 같아야 한다.
		if networkUUID != routingInterface.Network.UUID {
			return nil, errors.Unknown(errors.New("error occurred during creation of router"))
		}

		fixedIPs = append(fixedIPs, train.FixedIP{
			IPAddress:  routingInterface.RoutingInterface.IPAddress,
			SubnetUUID: routingInterface.Subnet.UUID,
		})
	}

	routerRsp, err := train.CreateRouter(c, train.CreateRouterRequest{
		Router: train.RouterResult{
			ProjectID:   req.Tenant.UUID,
			Name:        req.Router.Name,
			Description: req.Router.Description,
			State:       boolState,
			ExternalGatewayInfo: train.ExternalGatewayInfo{
				NetworkUUID:      networkUUID,
				ExternalFixedIPs: fixedIPs,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	var rollback = true
	var deleteInterfacePorts []string

	defer func() {
		if rollback {
			if err := train.RemoveExtraRoutesFromRouter(c, train.RemoveExtraRoutesFromRouterRequest{
				RouterUUID: routerRsp.Router.UUID,
				Router: train.RouterResult{
					Routes: routerRsp.Router.Routes,
				},
			}); err != nil {
				logger.Warnf("Could not delete created router. uuid: %s, Cause: %+v", routerRsp.Router.UUID, err)
				return
			}

			for _, portUUID := range deleteInterfacePorts {
				if err := train.RemoveInterfaceFromRouter(c, train.RemoveInterfaceFromRouterRequest{
					RouterUUID: routerRsp.Router.UUID,
					PortUUID:   portUUID,
				}); err != nil {
					logger.Warnf("Could not delete created router. uuid: %s, Cause: %+v", routerRsp.Router.UUID, err)
					return
				}
			}

			if err := train.DeleteRouter(c, train.DeleteRouterRequest{
				RouterUUID: routerRsp.Router.UUID,
			}); err != nil {
				logger.Warnf("Could not delete created router. uuid: %s, Cause: %+v", routerRsp.Router.UUID, err)
				return
			}
		}
	}()

	for _, routingInterface := range req.InternalRoutingInterfaceList {
		port, err := train.CreatePort(c, train.CreatePortRequest{
			Port: train.PortResult{
				NetworkUUID: routingInterface.Network.UUID,
				FixedIPs: []train.FixedIP{
					{
						SubnetUUID: routingInterface.Subnet.UUID,
						IPAddress:  routingInterface.RoutingInterface.IPAddress,
					},
				},
			},
		})
		if err != nil {
			return nil, err
		}

		_, err = train.AddInterfaceToRouter(c, train.AddInterfaceToRouterRequest{
			RouterUUID: routerRsp.Router.UUID,
			PortUUID:   &port.Port.UUID,
		})
		if err != nil {
			return nil, err
		}
		deleteInterfacePorts = append(deleteInterfacePorts, port.Port.UUID)
	}

	var externals []client.CreateRoutingInterface
	var internals []client.CreateRoutingInterface

	//portListRsp, err := train.ListPorts(c, train.ListPortsRequest{DeviceOwner: "network:router_interface", DeviceID: routerRsp.Router.UUID})
	//TODO: QA 요청으로 DeviceOwner 임시 삭제
	portListRsp, err := train.ListPorts(c, train.ListPortsRequest{DeviceID: routerRsp.Router.UUID})
	if err != nil {
		return nil, err
	}

	for _, port := range portListRsp.Ports {
		//TODO: ProjectID 추가로 내부 외부 구분 가능 검토
		if port.ProjectID != "" {
			internals = append(internals, client.CreateRoutingInterface{
				Network: model.ClusterNetwork{UUID: port.NetworkUUID},
				Subnet:  model.ClusterSubnet{UUID: port.FixedIPs[0].SubnetUUID},
				RoutingInterface: model.ClusterNetworkRoutingInterface{
					IPAddress:    port.FixedIPs[0].IPAddress,
					ExternalFlag: false,
				},
			})
		}

	}

	for _, p := range routerRsp.Router.ExternalGatewayInfo.ExternalFixedIPs {
		externals = append(externals, client.CreateRoutingInterface{
			Network: model.ClusterNetwork{
				UUID: routerRsp.Router.ExternalGatewayInfo.NetworkUUID,
			},
			Subnet: model.ClusterSubnet{
				UUID: p.SubnetUUID,
			},
			RoutingInterface: model.ClusterNetworkRoutingInterface{
				IPAddress:    p.IPAddress,
				ExternalFlag: true,
			},
		})
	}

	// extraroute 라는 extension 이 필요
	// extraroute 활성화 여부 판단 방법 못찾음
	var routes []train.ExtraRoute
	for _, route := range req.ExtraRouteList {
		routes = append(routes, train.ExtraRoute{
			DestinationCIDR: route.Destination,
			NextHopIP:       route.Nexthop,
		})
	}

	var mExtraRoutes []model.ClusterRouterExtraRoute
	if len(routes) > 0 {
		extraRoutes, err := train.AddExtraRoutesToRouter(c, train.AddExtraRoutesToRouterRequest{
			RouterUUID: routerRsp.Router.UUID,
			Router: train.RouterResult{
				Routes: routes,
			},
		})
		if errors.Equal(err, client.ErrRemoteServerError) {
			logger.Warnf("Could not create extra route. uuid: %s, Cause: %+v", routerRsp.Router.UUID, err)
		} else if err != nil {
			return nil, err
		} else {
			for _, route := range extraRoutes.Router.Routes {
				mExtraRoutes = append(mExtraRoutes, model.ClusterRouterExtraRoute{
					Destination: route.DestinationCIDR,
					Nexthop:     route.NextHopIP,
				})
			}
		}
	}

	var stringState string
	if routerRsp.Router.State {
		stringState = "up"
	} else {
		stringState = "down"
	}

	rollback = false
	return &client.CreateRouterResponse{
		Router: model.ClusterRouter{
			UUID:        routerRsp.Router.UUID,
			Name:        routerRsp.Router.Name,
			Description: routerRsp.Router.Description,
			State:       stringState,
			Status:      routerRsp.Router.Status,
			Raw:         &routerRsp.Raw,
		},
		ExtraRouteList:               mExtraRoutes,
		ExternalRoutingInterfaceList: externals,
		InternalRoutingInterfaceList: internals,
	}, nil
}

// DeleteRouter 라우터 삭제
func (c *Client) DeleteRouter(req client.DeleteRouterRequest) error {
	// port 및 router 조회
	//portListRsp, err := train.ListPorts(c, train.ListPortsRequest{DeviceOwner: "network:router_interface", DeviceID: req.Router.UUID})
	//TODO: QA 요청으로 DeviceOwner 임시 삭제
	portListRsp, err := train.ListPorts(c, train.ListPortsRequest{DeviceID: req.Router.UUID})
	if err != nil {
		return err
	}

	details, err := train.ShowRouterDetails(c, train.ShowRouterDetailsRequest{
		RouterID: req.Router.UUID,
	})
	if err != nil {
		return err
	}

	if len(details.Router.Routes) > 0 {
		if err := train.RemoveExtraRoutesFromRouter(c, train.RemoveExtraRoutesFromRouterRequest{
			RouterUUID: details.Router.UUID,
			Router: train.RouterResult{
				Routes: details.Router.Routes,
			},
		}); err != nil {
			return err
		}
	}

	for _, port := range portListRsp.Ports {
		if err := train.RemoveInterfaceFromRouter(c, train.RemoveInterfaceFromRouterRequest{
			RouterUUID: req.Router.UUID,
			PortUUID:   port.UUID,
		}); err != nil {
			return err
		}
	}

	return train.DeleteRouter(c, train.DeleteRouterRequest{
		RouterUUID: req.Router.UUID,
	})
}

// GetRouterList 라우터 목록 조회
func (c *Client) GetRouterList() (*client.GetRouterListResponse, error) {
	result, err := train.ListRouters(c)
	if err != nil {
		return nil, err
	}

	var rsp client.GetRouterListResponse
	for _, router := range result.Routers {
		rsp.ResultList = append(rsp.ResultList, client.GetRouterResult{
			Router: model.ClusterRouter{
				UUID: router.UUID,
			},
		})
	}

	return &rsp, nil
}

// GetRouter 라우터 상세 조회
func (c *Client) GetRouter(req client.GetRouterRequest) (*client.GetRouterResponse, error) {
	result, err := train.ShowRouterDetails(c, train.ShowRouterDetailsRequest{RouterID: req.Router.UUID})
	if err != nil {
		return nil, err
	}

	var state string
	if result.Router.State {
		state = "up"
	} else {
		state = "down"
	}

	//get extra route
	var extraRouteList []model.ClusterRouterExtraRoute
	for _, route := range result.Router.Routes {
		extraRouteList = append(extraRouteList, model.ClusterRouterExtraRoute{
			Destination: route.DestinationCIDR,
			Nexthop:     route.NextHopIP,
		})
	}

	// get external routing interface
	var exRouteInterfaceList []client.RoutingInterfaceResult
	for _, fixedIP := range result.Router.ExternalGatewayInfo.ExternalFixedIPs {
		exRouteInterfaceList = append(exRouteInterfaceList, client.RoutingInterfaceResult{
			Subnet: model.ClusterSubnet{UUID: fixedIP.SubnetUUID},
			RoutingInterface: model.ClusterNetworkRoutingInterface{
				IPAddress:    fixedIP.IPAddress,
				ExternalFlag: true,
			},
		})
	}

	// get internal routing interface
	//portListRsp, err := train.ListPorts(c, train.ListPortsRequest{DeviceOwner: "network:router_interface", DeviceID: req.Router.UUID})
	//TODO: QA 요청으로 DeviceOwner 임시 삭제
	portListRsp, err := train.ListPorts(c, train.ListPortsRequest{DeviceID: req.Router.UUID})
	if err != nil {
		return nil, err
	}

	var inRouteInterfaceList []client.RoutingInterfaceResult
	for _, port := range portListRsp.Ports {
		//TODO: ProjectID 추가로 내부 외부 구분 가능 검토
		if port.ProjectID != "" {
			for _, fixedIP := range port.FixedIPs {
				inRouteInterfaceList = append(inRouteInterfaceList, client.RoutingInterfaceResult{
					Subnet: model.ClusterSubnet{UUID: fixedIP.SubnetUUID},
					RoutingInterface: model.ClusterNetworkRoutingInterface{
						IPAddress:    fixedIP.IPAddress,
						ExternalFlag: false,
					},
				})
			}
		}
	}

	return &client.GetRouterResponse{
		Result: client.GetRouterResult{
			Router: model.ClusterRouter{
				UUID:        result.Router.UUID,
				Name:        result.Router.Name,
				Description: result.Router.Description,
				State:       state,
				Status:      result.Router.Status,
				Raw:         &result.Raw,
			},
			Tenant:                       model.ClusterTenant{UUID: result.Router.ProjectID},
			ExtraRoute:                   extraRouteList,
			ExternalRoutingInterfaceList: exRouteInterfaceList,
			InternalRoutingInterfaceList: inRouteInterfaceList,
		},
	}, nil
}

// CreateSecurityGroup 보안 그룹 생성
func (c *Client) CreateSecurityGroup(req client.CreateSecurityGroupRequest) (*client.CreateSecurityGroupResponse, error) {
	group, err := train.CreateSecurityGroup(c, train.CreateSecurityGroupRequest{
		SecurityGroup: train.CreateSecurityGroupResult{
			ProjectUUID: req.Tenant.UUID,
			Name:        req.SecurityGroup.Name,
			Description: req.SecurityGroup.Description,
		},
	})
	if err != nil {
		return nil, err
	}

	var rollback = true
	defer func() {
		if rollback {
			if err := train.DeleteSecurityGroup(c, train.DeleteSecurityGroupRequest{
				SecurityGroupUUID: group.SecurityGroup.UUID,
			}); err != nil {
				logger.Warnf("Could not delete created security group. uuid: %s, name: %s, Cause: %+v", group.SecurityGroup.UUID, group.SecurityGroup.Name, err)
			}
		}
	}()

	for _, rule := range group.SecurityGroup.Rules {
		if err := train.DeleteSecurityGroupRule(c, train.DeleteSecurityGroupRuleRequest{
			RuleUUID: rule.UUID,
		}); err != nil {
			return nil, err
		}
	}

	rollback = false
	return &client.CreateSecurityGroupResponse{
		SecurityGroup: model.ClusterSecurityGroup{
			UUID:        group.SecurityGroup.UUID,
			Name:        group.SecurityGroup.Name,
			Description: group.SecurityGroup.Description,
			Raw:         &group.Raw,
		},
	}, nil
}

// CreateSecurityGroupRule 보안 그룹 규칙 생성
func (c *Client) CreateSecurityGroupRule(req client.CreateSecurityGroupRuleRequest) (*client.CreateSecurityGroupRuleResponse, error) {
	var remoteGroupUUID string
	if req.RemoteSecurityGroup != nil {
		remoteGroupUUID = req.RemoteSecurityGroup.UUID
	}
	rule, err := train.CreateSecurityGroupRule(c, train.CreateSecurityGroupRuleRequest{
		Rule: train.CreateSecurityGroupRuleResult{
			SecurityGroupUUID: req.SecurityGroup.UUID,
			RemoteGroupUUID:   remoteGroupUUID,
			Direction:         req.Rule.Direction,
			Protocol:          req.Rule.Protocol,
			RemoteIPPrefix:    req.Rule.NetworkCidr,
			EtherType:         train.Uint32EtherType(req.Rule.EtherType),
			PortRangeMax:      req.Rule.PortRangeMax,
			PortRangeMin:      req.Rule.PortRangeMin,
			Description:       req.Rule.Description,
		},
	})
	if err != nil {
		return nil, err
	}

	var mRemoteGroup *model.ClusterSecurityGroup
	if rule.Rule.RemoteGroupUUID != "" {
		mRemoteGroup = &model.ClusterSecurityGroup{UUID: rule.Rule.RemoteGroupUUID}
	}

	return &client.CreateSecurityGroupRuleResponse{
		Rule: model.ClusterSecurityGroupRule{
			UUID:         rule.Rule.UUID,
			Description:  rule.Rule.Description,
			EtherType:    train.StringEtherType(*rule.Rule.EtherType),
			NetworkCidr:  rule.Rule.RemoteIPPrefix,
			Direction:    rule.Rule.Direction,
			PortRangeMin: rule.Rule.PortRangeMin,
			PortRangeMax: rule.Rule.PortRangeMax,
			Protocol:     rule.Rule.Protocol,
			Raw:          &rule.Raw,
		},
		RemoteSecurityGroup: mRemoteGroup,
	}, nil
}

// DeleteSecurityGroup 보안 그룹 삭제
func (c *Client) DeleteSecurityGroup(req client.DeleteSecurityGroupRequest) error {
	return train.DeleteSecurityGroup(c, train.DeleteSecurityGroupRequest{
		SecurityGroupUUID: req.SecurityGroup.UUID,
	})
}

// DeleteSecurityGroupRule 보안 그룹 규칙 삭제
func (c *Client) DeleteSecurityGroupRule(req client.DeleteSecurityGroupRuleRequest) error {
	return train.DeleteSecurityGroupRule(c, train.DeleteSecurityGroupRuleRequest{
		RuleUUID: req.Rule.UUID,
	})
}

// GetSecurityGroupList security group 목록 조회
func (c *Client) GetSecurityGroupList() (*client.GetSecurityGroupListResponse, error) {
	result, err := train.ListSecurityGroups(c)
	if err != nil {
		return nil, err
	}

	var rsp client.GetSecurityGroupListResponse
	for _, securityGroup := range result.SecurityGroups {
		rsp.ResultList = append(rsp.ResultList, client.GetSecurityGroupResult{SecurityGroup: model.ClusterSecurityGroup{UUID: securityGroup.UUID}})
	}

	return &rsp, nil
}

// GetSecurityGroup security group 상세 조회
func (c *Client) GetSecurityGroup(req client.GetSecurityGroupRequest) (*client.GetSecurityGroupResponse, error) {
	result, err := train.ShowSecurityGroup(c, train.ShowSecurityGroupRequest{SecurityGroupID: req.SecurityGroup.UUID})
	if err != nil {
		return nil, err
	}

	var ruleList []model.ClusterSecurityGroupRule
	for _, rule := range result.SecurityGroup.Rules {
		ruleList = append(ruleList, model.ClusterSecurityGroupRule{UUID: rule.UUID})
	}

	return &client.GetSecurityGroupResponse{
		Result: client.GetSecurityGroupResult{
			SecurityGroup: model.ClusterSecurityGroup{
				UUID:        result.SecurityGroup.UUID,
				Name:        result.SecurityGroup.Name,
				Description: result.SecurityGroup.Description,
				Raw:         &result.Raw,
			},
			Tenant: model.ClusterTenant{
				UUID: result.SecurityGroup.ProjectID,
			},
			RuleList: ruleList,
		},
	}, nil
}

// GetSecurityGroupRule security group rule 상세 조회
func (c *Client) GetSecurityGroupRule(req client.GetSecurityGroupRuleRequest) (*client.GetSecurityGroupRuleResponse, error) {
	rsp, err := train.ShowSecurityGroupRule(c, train.ShowSecurityGroupRuleRequest{SecurityGroupRuleID: req.SecurityGroupRule.UUID})
	if err != nil {
		return nil, err
	}

	var remoteSecurityGroup *model.ClusterSecurityGroup
	if rsp.SecurityGroupRule.RemoteGroupID != nil {
		remoteSecurityGroup = &model.ClusterSecurityGroup{
			UUID: *rsp.SecurityGroupRule.RemoteGroupID,
		}
	}

	return &client.GetSecurityGroupRuleResponse{
		Result: client.GetSecurityGroupRuleResult{
			SecurityGroupRule: model.ClusterSecurityGroupRule{
				UUID:         rsp.SecurityGroupRule.UUID,
				Description:  rsp.SecurityGroupRule.Description,
				NetworkCidr:  rsp.SecurityGroupRule.NetworkCidr,
				Direction:    rsp.SecurityGroupRule.Direction,
				PortRangeMin: rsp.SecurityGroupRule.PortRangeMin,
				PortRangeMax: rsp.SecurityGroupRule.PortRangeMax,
				Protocol:     rsp.SecurityGroupRule.Protocol,
				EtherType:    train.StringEtherType(rsp.SecurityGroupRule.EtherType),
				Raw:          &rsp.Raw,
			},
			SecurityGroup: model.ClusterSecurityGroup{
				UUID: rsp.SecurityGroupRule.SecurityGroupID,
			},
			RemoteSecurityGroup: remoteSecurityGroup,
		},
	}, nil
}

// GetSubnet 서브넷 상세 조회
func (c *Client) GetSubnet(req client.GetSubnetRequest) (*client.GetSubnetResponse, error) {
	result, err := train.ShowSubnetDetails(c, train.ShowSubnetDetailsRequest{SubnetID: req.Subnet.UUID})
	if err != nil {
		return nil, err
	}

	var nameServerList []model.ClusterSubnetNameserver
	for _, nameServer := range result.Subnet.Nameservers {
		nameServerList = append(nameServerList, model.ClusterSubnetNameserver{
			Nameserver: nameServer,
		})
	}

	var dhcpPoolList []model.ClusterSubnetDHCPPool
	for _, dhcpPool := range result.Subnet.AllocationPools {
		dhcpPoolList = append(dhcpPoolList, model.ClusterSubnetDHCPPool{
			StartIPAddress: dhcpPool.Start,
			EndIPAddress:   dhcpPool.End,
		})
	}

	var gatewayEnabled bool
	if result.Subnet.GatewayIPAddress != nil {
		gatewayEnabled = true
	}

	return &client.GetSubnetResponse{
		Result: client.GetSubnetResult{
			Subnet: model.ClusterSubnet{
				UUID:                result.Subnet.UUID,
				Name:                result.Subnet.Name,
				Description:         result.Subnet.Description,
				NetworkCidr:         result.Subnet.NetworkCidr,
				DHCPEnabled:         result.Subnet.DHCPEnabled,
				GatewayEnabled:      gatewayEnabled,
				GatewayIPAddress:    result.Subnet.GatewayIPAddress,
				Ipv6AddressModeCode: result.Subnet.Ipv6AddressModeCode,
				Ipv6RaModeCode:      result.Subnet.Ipv6RaModeCode,
				Raw:                 &result.Raw,
			},
			Network: model.ClusterNetwork{
				UUID: result.Subnet.NetworkID,
			},
			NameserverList: nameServerList,
			PoolList:       dhcpPoolList,
		},
	}, nil
}

// GetStorageList 스토리지 목록 조회
func (c *Client) GetStorageList() (*client.GetStorageListResponse, error) {
	result, err := train.ListAllVolumeTypes(c)
	if err != nil {
		return nil, err
	}

	var rsp client.GetStorageListResponse
	for _, volumeType := range result.VolumeTypes {
		rsp.ResultList = append(rsp.ResultList, client.GetStorageResult{
			Storage: model.ClusterStorage{UUID: volumeType.UUID},
		})
	}

	return &rsp, nil
}

// GetStorage 스토리지 상세 조회
func (c *Client) GetStorage(req client.GetStorageRequest) (*client.GetStorageResponse, error) {
	result, err := train.ShowVolumeTypeDetail(c, train.ShowVolumeTypeDetailRequest{VolumeTypeID: req.Storage.UUID})
	if err != nil {
		return nil, err
	}

	var volumeBackendName string
	v, ok := result.VolumeType.ExtraSpec["volume_backend_name"]
	if !ok {
		logger.Warnf("GetStorage result.Raw : %v", result.Raw)
		logger.Warnf("GetStorage result.ExtraSpec : %+v", result.VolumeType)
		if result.VolumeType.Name == "__DEFAULT__" {
			v = "ceph"
			logger.Warnf("GetStorage result.ExtraSpec is empty : changed spec (%s)", v.(string))
		} else {
			return nil, NotFoundVolumeBackendName(result.VolumeType.Name)
		}
	}
	volumeBackendName = v.(string)

	logger.Infof("[GetStorage] Start : ListAllBackendStoragePools [%s]", volumeBackendName)
	storagePoolsRsp, err := train.ListAllBackendStoragePools(c)
	if err != nil {
		//return nil, err
		logger.Warnf("[GetStorage] Failed to sync storage(%s). Cause: %+v", volumeBackendName, err)
	}

	var capacityBytes, freeBytes, usedBytes *uint64
	if err == nil {
		for _, pool := range storagePoolsRsp.StoragePools {
			if volumeBackendName == pool.Capabilities.VolumeBackendName {
				totalCapacityGigaBytes, ok := pool.Capabilities.TotalCapacityGigaBytes.(float64)
				if ok {
					totalCapacityBytes := uint64(totalCapacityGigaBytes * 1000 * 1000 * 1000)
					capacityBytes = &totalCapacityBytes
				}
				freeCapacityGigaBytes, ok := pool.Capabilities.FreeCapacityGigaBytes.(float64)
				if ok {
					freeCapacityBytes := uint64(freeCapacityGigaBytes * 1000 * 1000 * 1000)
					freeBytes = &freeCapacityBytes
				}
				if capacityBytes != nil && freeBytes != nil {
					usedBytes = new(uint64)
					*usedBytes = *capacityBytes - *freeBytes
				}
				break
			}
		}
	}
	logger.Info("[GetStorage] Done : ListAllBackendStoragePools")

	typeCode := storage.ClusterStorageTypeUnKnown

	md := map[string]interface{}{"backend": volumeBackendName}
	if strings.Contains(volumeBackendName, "ceph") == true {
		typeCode = storage.ClusterStorageTypeCeph
		md["driver"] = "cinder.volume.drivers.rbd.RBDDriver"
		md["client"] = "cinder"
		md["pool"] = "volumes"
	} else if strings.Contains(volumeBackendName, "nfs") == true {
		typeCode = storage.ClusterStorageTypeNFS
	} else if strings.Contains(volumeBackendName, "iscsi") == true {
		typeCode = storage.ClusterStorageTypeLvm
	} else if strings.Contains(volumeBackendName, "lvm") == true {
		typeCode = storage.ClusterStorageTypeLvm
	}

	logger.Info("[GetStorage] Done : getBackendMetadata")

	return &client.GetStorageResponse{
		Result: client.GetStorageResult{
			Storage: model.ClusterStorage{
				UUID:          result.VolumeType.UUID,
				Name:          result.VolumeType.Name,
				Description:   result.VolumeType.Description,
				TypeCode:      typeCode,
				CapacityBytes: capacityBytes,
				UsedBytes:     usedBytes,
				Credential:    nil,
				Raw:           &result.Raw,
			},
			VolumeBackendName: volumeBackendName,
			Metadata:          md,
		},
	}, nil
}

// GetVolumeList 볼륨 목록 조회
func (c *Client) GetVolumeList() (*client.GetVolumeListResponse, error) {
	result, err := train.ListAccessibleVolumes(c)
	if err != nil {
		return nil, err
	}

	var rsp client.GetVolumeListResponse
	for _, volume := range result.Volumes {
		rsp.ResultList = append(rsp.ResultList, client.GetVolumeResult{
			Volume: model.ClusterVolume{UUID: volume.UUID},
		})
	}

	return &rsp, nil
}

// GetVolume 볼륨 상세 조회
func (c *Client) GetVolume(req client.GetVolumeRequest) (*client.GetVolumeResponse, error) {
	result, err := train.ShowVolumeDetails(c, train.ShowVolumeDetailsRequest{VolumeID: req.Volume.UUID})
	if err != nil {
		return nil, err
	}

	bootable, err := strconv.ParseBool(result.Volume.Bootable)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var readOnly bool
	if v, ok := result.Volume.Metadata["readonly"]; ok {
		readOnly, err = strconv.ParseBool(v.(string))
		if err != nil {
			return nil, errors.Unknown(err)
		}
	}

	//get volume type
	volumeTypesRsp, err := train.ListAllVolumeTypes(c)
	if err != nil {
		return nil, err
	}

	var storageUUID string
	for _, volumeType := range volumeTypesRsp.VolumeTypes {
		if volumeType.Name == result.Volume.VolumeTypeName {
			storageUUID = volumeType.UUID
			break
		}
	}
	if storageUUID == "" {
		return nil, NotFoundVolumeType(result.Volume.Name, result.Volume.VolumeTypeName)
	}

	var md map[string]interface{}
	if len(result.Volume.Attachments) > 0 {
		attachRsp, err := train.ShowAttachmentDetails(c, train.ShowAttachmentDetailsRequest{AttachmentID: result.Volume.Attachments[0].UUID})
		if err != nil {
			return nil, err
		}
		md = attachRsp.VolumeAttachment.ConnectionInfo
	}

	return &client.GetVolumeResponse{
		Result: client.GetVolumeResult{
			Volume: model.ClusterVolume{
				UUID:        result.Volume.UUID,
				Name:        result.Volume.Name,
				Description: result.Volume.Description,
				SizeBytes:   result.Volume.SizeGibiBytes * 1024 * 1024 * 1024,
				Multiattach: result.Volume.Multiattach,
				Bootable:    bootable,
				Readonly:    readOnly,
				Status:      result.Volume.Status,
				Raw:         &result.Raw,
			},
			Storage: model.ClusterStorage{
				UUID: storageUUID,
			},
			Tenant: model.ClusterTenant{
				UUID: result.Volume.ProjectID,
			},
			Metadata: md,
		},
	}, nil
}

// GetVolumeGroupList 볼륨 그룹 목록 조회
func (c *Client) GetVolumeGroupList() (*client.GetVolumeGroupListResponse, error) {
	result, err := train.ListAllVolumeGroups(c)
	if err != nil {
		return nil, err
	}

	var (
		rsp          client.GetVolumeGroupListResponse
		volumeGroups []client.VolumeGroup
	)
	for _, group := range result.Groups {
		volumeGroup := client.VolumeGroup{UUID: group.UUID, Name: group.Name}
		volumeGroups = append(volumeGroups, volumeGroup)
	}

	rsp.VolumeGroupList = volumeGroups

	return &rsp, nil
}

// GetVolumeGroup 볼륨 그룹 조회
func (c *Client) GetVolumeGroup(req client.GetVolumeGroupRequest) (*client.GetVolumeGroupResponse, error) {
	showRsp, err := train.ShowVolumeGroupDetails(c, train.ShowVolumeGroupDetailsRequest{
		GroupUUID:   req.VolumeGroup.UUID,
		ShowVolumes: true,
	})
	if err != nil {
		return nil, err
	}

	rsp := client.GetVolumeGroupResponse{
		VolumeGroup: client.VolumeGroup{
			UUID:        showRsp.Group.ID,
			Name:        showRsp.Group.Name,
			Description: showRsp.Group.Description,
		},
	}

	for _, id := range showRsp.Group.VolumeTypes {
		rsp.VolumeGroup.StorageList = append(rsp.VolumeGroup.StorageList, model.ClusterStorage{UUID: id})
	}

	for _, id := range showRsp.Group.Volumes {
		rsp.VolumeGroup.VolumeList = append(rsp.VolumeGroup.VolumeList, model.ClusterVolume{UUID: id})
	}

	return &rsp, nil
}

// CreateVolumeGroup 볼륨 그룹 생성
func (c *Client) CreateVolumeGroup(req client.CreateVolumeGroupRequest) (*client.CreateVolumeGroupResponse, error) {
	groupTypeRsp, err := train.CreateVolumeGroupType(c, train.CreateVolumeGroupTypeRequest{
		GroupType: train.VolumeGroupType{
			Name:        req.VolumeGroup.Name,
			Description: req.VolumeGroup.Description,
			IsPublic:    true,
		},
	})
	if err != nil {
		return nil, err
	}

	var storagesIn []string
	for _, item := range req.VolumeGroup.StorageList {
		storagesIn = append(storagesIn, item.UUID)
	}

	var groupRsp *train.VolumeGroupResponse
	var rollback = true
	defer func() {
		if !rollback {
			return
		}

		if groupRsp != nil {
			if err := train.DeleteVolumeGroup(c, train.DeleteVolumeGroupRequest{
				GroupID: groupRsp.Group.ID,
				Delete:  train.DeleteWrapper{DeleteVolumes: false},
			}); err != nil {
				logger.Warnf("Could not delete created volume group. uuid: %s, Cause: %+v", groupRsp.Group.ID, err)
			}
		}

		if err := train.DeleteVolumeGroupType(c, train.DeleteVolumeGroupTypeRequest{VolumeGroupTypeID: groupTypeRsp.GroupType.ID}); err != nil {
			logger.Warnf("Could not delete created volume group type. uuid: %s, Cause: %+v", groupTypeRsp.GroupType.ID, err)
		}
	}()

	groupRsp, err = train.CreateVolumeGroup(c, train.CreateVolumeGroupRequest{
		Group: train.VolumeGroup{
			Name:        req.VolumeGroup.Name,
			Description: req.VolumeGroup.Description,
			GroupType:   groupTypeRsp.GroupType.ID,
			VolumeTypes: storagesIn,
		},
	})
	if err != nil {
		return nil, err
	}

	rollback = false
	return &client.CreateVolumeGroupResponse{
		VolumeGroup: client.VolumeGroup{
			UUID:        groupRsp.Group.ID,
			Name:        groupRsp.Group.Name,
			Description: groupRsp.Group.Description,
		},
	}, nil
}

// CreateVolumeGroupSnapshot 볼륨 그룹 스냅샷 생성
func (c *Client) CreateVolumeGroupSnapshot(req client.CreateVolumeGroupSnapshotRequest) (*client.CreateVolumeGroupSnapshotResponse, error) {
	var (
		rollback = true
		rsp      *train.VolumeGroupSnapshotResponse
		err      error
	)
	defer func() {
		if !rollback || rsp == nil {
			return
		}

		if err := train.DeleteVolumeGroupSnapshot(c, train.DeleteVolumeGroupSnapshotRequest{
			VolumeGroupSnapshotID: rsp.VolumeGroupSnapshot.ID,
		}); err != nil {
			logger.Warnf("Could not delete created volume group snapshot. uuid: %s, Cause: %+v", rsp.VolumeGroupSnapshot.ID, err)
		}
	}()

	rsp, err = train.CreateVolumeGroupSnapshot(c, train.CreateVolumeGroupSnapshotRequest{
		VolumeGroupSnapshot: train.VolumeGroupSnapshot{
			GroupID:     req.VolumeGroup.UUID,
			Name:        req.VolumeGroupSnapshot.Name,
			Description: req.VolumeGroupSnapshot.Description,
		},
	})
	if err != nil {
		return nil, err
	}

	rollback = false
	return &client.CreateVolumeGroupSnapshotResponse{
		VolumeGroupSnapshot: client.VolumeGroupSnapshot{
			UUID:        rsp.VolumeGroupSnapshot.ID,
			Name:        rsp.VolumeGroupSnapshot.Name,
			Description: rsp.VolumeGroupSnapshot.Description,
		},
	}, nil
}

// DeleteVolumeGroupSnapshot 볼륨 그룹 스냅샷 삭제
func (c *Client) DeleteVolumeGroupSnapshot(req client.DeleteVolumeGroupSnapshotRequest) error {
	return train.DeleteVolumeGroupSnapshot(c, train.DeleteVolumeGroupSnapshotRequest{
		VolumeGroupSnapshotID: req.VolumeGroupSnapshot.UUID,
	})
}

// GetVolumeGroupSnapshotList 볼륨 그룹 스냅샷 목록 조회
func (c *Client) GetVolumeGroupSnapshotList() (*client.GetVolumeGroupSnapshotListResponse, error) {
	rsp, err := train.ListVolumeGroupSnapshotsWithDetails(c)
	if err != nil {
		return nil, err
	}

	var snapshots []client.GetVolumeGroupSnapshotListResult
	for _, vgs := range rsp.VolumeGroupSnapshots {
		snapshots = append(snapshots, client.GetVolumeGroupSnapshotListResult{
			VolumeGroup: client.VolumeGroup{
				UUID: vgs.GroupID,
			},
			VolumeGroupSnapshot: client.VolumeGroupSnapshot{
				UUID:        vgs.ID,
				Name:        vgs.Name,
				Description: vgs.Description,
			},
		})
	}

	return &client.GetVolumeGroupSnapshotListResponse{
		VolumeGroupSnapshots: snapshots,
	}, nil
}

// CreateVolume 볼륨 생성
func (c *Client) CreateVolume(req client.CreateVolumeRequest) (*client.CreateVolumeResponse, error) {
	var rsp *train.CreateVolumeResponse
	var err error
	var rollback = true

	defer func() {
		if !rollback || rsp == nil {
			return
		}

		if err := train.DeleteVolume(c, train.DeleteVolumeRequest{
			VolumeID: rsp.Volume.UUID,
		}); err != nil {
			logger.Warnf("Could not delete created volume. uuid: %s, name: %s, Cause: %+v", rsp.Volume.UUID, rsp.Volume.Name, err)
		}
	}()

	rsp, err = train.CreateVolume(c, train.CreateVolumeRequest{
		Volume: train.VolumeResult{
			Description:    req.Volume.Description,
			VolumeTypeName: req.Storage.Name,
			Name:           req.Volume.Name,
			SizeGibiBytes:  req.Volume.SizeBytes / (1024 * 1024 * 1024),
			Multiattach:    req.Volume.Multiattach,
		},
	})
	if err != nil {
		return nil, err
	}

	if err := updateVolumeMode(c, updateVolumeModeRequest{
		VolumeUUID: rsp.Volume.UUID,
		Bootable:   req.Volume.Bootable,
		Readonly:   req.Volume.Readonly,
	}); err != nil {
		return nil, err
	}

	rollback = false
	return &client.CreateVolumeResponse{
		Volume: model.ClusterVolume{
			UUID:        rsp.Volume.UUID,
			Name:        rsp.Volume.Name,
			Description: rsp.Volume.Description,
			SizeBytes:   rsp.Volume.SizeGibiBytes * 1024 * 1024 * 1024,
			Multiattach: rsp.Volume.Multiattach,
			Readonly:    req.Volume.Readonly,
			Bootable:    req.Volume.Bootable,
			Raw:         rsp.Raw,
		},
	}, nil
}

type updateVolumeModeRequest struct {
	VolumeUUID string
	Bootable   bool
	Readonly   bool
}

func updateVolumeMode(c client.Client, req updateVolumeModeRequest) error {
	if req.Readonly == true {
		if err := train.UpdatesVolumeReadOnlyAccessModeFlag(c, train.UpdatesVolumeReadOnlyAccessModeFlagRequest{
			VolumeID: req.VolumeUUID,
			VolumeReadOnly: train.UpdatesVolumeReadOnlyAccessModeFlagResult{
				ReadOnly: req.Readonly,
			},
		}); err != nil {
			return err
		}
	}

	if req.Bootable == true {
		if err := train.UpdateVolumeBootableStatus(c, train.UpdateVolumeBootableStatusRequest{
			VolumeID: req.VolumeUUID,
			VolumeBootable: train.UpdateVolumeBootableStatusResult{
				Bootable: req.Bootable,
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

// UUIDPair ID 쌍
type UUIDPair struct {
	SourceUUID      string
	DestinationUUID string
}

// ImportVolume 볼륨 가져오기
func (c *Client) ImportVolume(req client.ImportVolumeRequest) (*client.ImportVolumeResponse, error) {
	switch req.TargetStorage.TypeCode {
	case storage.ClusterStorageTypeCeph:
		return importCEPH(c, &req)
	default:
		return nil, errors.UnavailableParameterValue("target_storage.type_code", req.TargetStorage.TypeCode, storage.ClusterStorageTypeCodes)
	}
}

// DeleteVolume 볼륨 삭제
func (c *Client) DeleteVolume(req client.DeleteVolumeRequest) error {
	return train.DeleteVolume(c, train.DeleteVolumeRequest{
		VolumeID: req.Volume.UUID,
		Cascade:  true,
	})
}

// UnmanageVolume 볼륨 unmanage
func (c *Client) UnmanageVolume(req client.UnmanageVolumeRequest) error {
	_, err := train.ShowVolumeDetails(c, train.ShowVolumeDetailsRequest{
		VolumeID: req.VolumePair.Target.UUID,
	})
	switch {
	case errors.Equal(err, client.ErrNotFound):
		logger.Warnf("Could not unmanage volume. Cause: %+v", err)
		return nil
	case err != nil:
		return err
	}

	switch req.TargetStorage.TypeCode {
	case storage.ClusterStorageTypeCeph:
		return unmanageCEPH(c, &req)
	default:
		return errors.UnavailableParameterValue("storage.type_code", req.TargetStorage.TypeCode, storage.ClusterStorageTypeCodes)
	}
}

// GetVolumeSnapshotList 볼륨 스냅샷 목록 조회
func (c *Client) GetVolumeSnapshotList() (*client.GetVolumeSnapshotListResponse, error) {
	result, err := train.ListAccessibleSnapshots(c)
	if err != nil {
		return nil, err
	}

	var rsp client.GetVolumeSnapshotListResponse
	for _, snapshot := range result.VolumeSnapshots {
		rsp.ResultList = append(rsp.ResultList, client.GetVolumeSnapshotResult{
			VolumeSnapshot: model.ClusterVolumeSnapshot{UUID: snapshot.UUID},
		})
	}

	return &rsp, nil
}

// GetVolumeSnapshot 볼륨 스냅샷 상세 조회
func (c *Client) GetVolumeSnapshot(req client.GetVolumeSnapshotRequest) (*client.GetVolumeSnapshotResponse, error) {
	result, err := train.ShowVolumeSnapshotDetails(c, train.ShowVolumeSnapshotDetailsRequest{SnapshotID: req.VolumeSnapshot.UUID})
	if err != nil {
		return nil, err
	}

	createdAt, err := time.Parse("2006-01-02T15:04:05.000000", result.VolumeSnapshot.CreatedAt)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	return &client.GetVolumeSnapshotResponse{
		Result: client.GetVolumeSnapshotResult{
			VolumeSnapshot: model.ClusterVolumeSnapshot{
				UUID:                           result.VolumeSnapshot.UUID,
				ClusterVolumeGroupSnapshotUUID: result.VolumeSnapshot.GroupSnapshotID,
				Name:                           result.VolumeSnapshot.Name,
				Description:                    result.VolumeSnapshot.Description,
				SizeBytes:                      result.VolumeSnapshot.SizeGibiBytes * 1024 * 1024 * 1024,
				Status:                         result.VolumeSnapshot.Status,
				CreatedAt:                      createdAt.Unix(),
				Raw:                            &result.Raw,
			},
			Volume: model.ClusterVolume{
				UUID: result.VolumeSnapshot.VolumeID,
			},
		},
	}, nil
}

// CreateVolumeSnapshot 볼륨 스냅샷 생성
func (c *Client) CreateVolumeSnapshot(req client.CreateVolumeSnapshotRequest) (*client.CreateVolumeSnapshotResponse, error) {
	var rsp *train.CreateVolumeSnapshotResponse
	var err error
	var rollback = true

	defer func() {
		if !rollback || rsp == nil {
			return
		}

		if err := train.DeleteVolumeSnapshot(c, train.DeleteVolumeSnapshotRequest{
			SnapshotID: rsp.VolumeSnapshot.UUID,
		}); err != nil {
			logger.Warnf("Could not delete created volume snapshot. uuid: %s, name: %s, Cause: %+v", rsp.VolumeSnapshot.UUID, rsp.VolumeSnapshot.Name, err)
		}
	}()

	rsp, err = train.CreateVolumeSnapshot(c, train.CreateVolumeSnapshotRequest{
		VolumeSnapshot: train.VolumeSnapshotResult{
			Description: req.VolumeSnapshot.Description,
			Name:        req.VolumeSnapshot.Name,
			VolumeID:    req.Volume.UUID,
		},
	})
	if err != nil {
		return nil, err
	}

	createdAt, err := time.Parse("2006-01-02T15:04:05.000000", rsp.VolumeSnapshot.CreatedAt)
	if err != nil {
		return nil, err
	}

	rollback = false
	return &client.CreateVolumeSnapshotResponse{
		VolumeSnapshot: model.ClusterVolumeSnapshot{
			UUID:        rsp.VolumeSnapshot.UUID,
			Name:        rsp.VolumeSnapshot.Name,
			Description: rsp.VolumeSnapshot.Description,
			SizeBytes:   rsp.VolumeSnapshot.SizeGibiBytes * 1024 * 1024 * 1024,
			CreatedAt:   createdAt.Unix(),
			Raw:         rsp.Raw,
		},
	}, nil
}

// DeleteVolumeGroup 볼륨 그룹 삭제
func (c *Client) DeleteVolumeGroup(req client.DeleteVolumeGroupRequest) error {
	groupRsp, err := train.ShowVolumeGroupDetails(c, train.ShowVolumeGroupDetailsRequest{
		GroupUUID:   req.VolumeGroup.UUID,
		ShowVolumes: true,
	})
	if errors.Equal(err, client.ErrNotFound) {
		logger.Warnf("Could not delete volume group(%s). Cause: not found volume group", req.VolumeGroup.UUID)
		return nil
	} else if err != nil {
		return err
	}

	vgSnapshotsRsp, err := train.ListVolumeGroupSnapshotsWithDetails(c)
	if err != nil {
		return err
	}

	for _, snapshot := range vgSnapshotsRsp.VolumeGroupSnapshots {
		if snapshot.GroupID != req.VolumeGroup.UUID {
			continue
		}

		if err := train.DeleteVolumeGroupSnapshot(c, train.DeleteVolumeGroupSnapshotRequest{
			VolumeGroupSnapshotID: snapshot.ID,
		}); err != nil {
			return err
		}
	}

	getPartVolumes := func(index int, volumes []string) []string {
		var limit = 6
		var retry = len(volumes) / limit

		switch {
		case index > retry:
			return nil
		case index == retry:
			if len(volumes)%limit == 0 {
				return nil
			}
			return volumes[index*limit:]
		}

		return volumes[index*limit : (index+1)*limit]
	}

	if len(groupRsp.Group.Volumes) > 0 {
		var index int
		for {
			partDeleteVolumes := getPartVolumes(index, groupRsp.Group.Volumes)

			if partDeleteVolumes == nil {
				break
			}

			if err := train.UpdateVolumeGroup(c, train.UpdateVolumeGroupRequest{
				VolumeGroupUUID: req.VolumeGroup.UUID,
				Group: train.UpdateVolumeGroupWrapper{
					RemoveVolumes: strings.Join(partDeleteVolumes, ","),
				},
			}); err != nil {
				return err
			}
			index++
		}
	}

	if err := train.DeleteVolumeGroup(c, train.DeleteVolumeGroupRequest{
		GroupID: req.VolumeGroup.UUID,
		Delete: train.DeleteWrapper{
			DeleteVolumes: false,
		},
	}); err != nil {
		return err
	}

	return train.DeleteVolumeGroupType(c, train.DeleteVolumeGroupTypeRequest{
		VolumeGroupTypeID: groupRsp.Group.GroupType,
	})
}

// CopyVolume 볼륨 copy
func (c *Client) CopyVolume(req client.CopyVolumeRequest) (*client.CopyVolumeResponse, error) {
	switch req.TargetStorage.TypeCode {
	case storage.ClusterStorageTypeCeph:
		return copyCEPH(c, &req)
	default:
		return nil, errors.UnavailableParameterValue("target_storage.type_code", req.TargetStorage.TypeCode, storage.ClusterStorageTypeCodes)
	}
}

// DeleteVolumeCopy 볼륨 copy 삭제
func (c *Client) DeleteVolumeCopy(req client.DeleteVolumeCopyRequest) error {
	switch req.TargetStorage.TypeCode {
	case storage.ClusterStorageTypeCeph:
		return deleteCopyCEPH(c, &req)
	default:
		return errors.UnavailableParameterValue("target_storage.type_code", req.TargetStorage.TypeCode, storage.ClusterStorageTypeCodes)
	}
}

func isContainUUID(uuid string, uuids []string) bool {
	for _, item := range uuids {
		if uuid == item {
			return true
		}
	}

	return false
}

// UpdateVolumeGroup 볼륨 그룹 수정
func (c *Client) UpdateVolumeGroup(req client.UpdateVolumeGroupRequest) error {
	groupRsp, err := train.ShowVolumeGroupDetails(c, train.ShowVolumeGroupDetailsRequest{
		GroupUUID:   req.VolumeGroup.UUID,
		ShowVolumes: true,
	})
	if errors.Equal(err, client.ErrNotFound) {
		logger.Warnf("Could not update volume group(%s). Cause: not found volume group", req.VolumeGroup.UUID)
		return nil
	} else if err != nil {
		return err
	}

	var (
		addVolumes []string
		delVolumes []string
	)

	for _, volume := range req.AddVolumes {
		if !isContainUUID(volume.UUID, groupRsp.Group.Volumes) {
			addVolumes = append(addVolumes, volume.UUID)
		} else {
			logger.Warnf("Could not add volume %s in volume group %s. Cause: already exists volume in volume group", volume.UUID, req.VolumeGroup.UUID)
		}
	}

	for _, volume := range req.DelVolumes {
		if isContainUUID(volume.UUID, groupRsp.Group.Volumes) {
			delVolumes = append(delVolumes, volume.UUID)
		} else {
			logger.Warnf("Could not delete volume %s in volume group %s. Cause: not exists volume in volume group", volume.UUID, req.VolumeGroup.UUID)
		}
	}

	getPartVolumes := func(index int, volumes []string) []string {
		var limit = 6
		var retry = len(volumes) / limit

		switch {
		case index > retry:
			return nil
		case index == retry:
			if len(volumes)%limit == 0 {
				return nil
			}
			return volumes[index*limit:]
		}

		return volumes[index*limit : (index+1)*limit]
	}

	var index int
	for {
		partAddVolumes := getPartVolumes(index, addVolumes)
		partDeleteVolumes := getPartVolumes(index, delVolumes)

		if partAddVolumes == nil && partDeleteVolumes == nil {
			return nil
		}

		if err := train.UpdateVolumeGroup(c, train.UpdateVolumeGroupRequest{
			VolumeGroupUUID: req.VolumeGroup.UUID,
			Group: train.UpdateVolumeGroupWrapper{
				Name:          req.VolumeGroup.Name,
				Description:   req.VolumeGroup.Description,
				AddVolumes:    strings.Join(partAddVolumes, ","),
				RemoveVolumes: strings.Join(partDeleteVolumes, ","),
			},
		}); err != nil {
			return err
		}

		index++
	}
}

// DeleteVolumeSnapshot 볼륨 스냅샷 삭제
func (c *Client) DeleteVolumeSnapshot(req client.DeleteVolumeSnapshotRequest) error {
	return train.DeleteVolumeSnapshot(c, train.DeleteVolumeSnapshotRequest{
		SnapshotID: req.VolumeSnapshot.UUID,
	})
}

// CreateInstanceSpec 인스턴스 스팩 생성
func (c *Client) CreateInstanceSpec(req client.CreateInstanceSpecRequest) (*client.CreateInstanceSpecResponse, error) {
	var rollback = true
	rspFlavor, err := train.CreateFlavor(c, train.CreateFlavorRequest{
		Flavor: train.FlavorResult{
			Name:               req.InstanceSpec.Name,
			MemTotalMibiBytes:  req.InstanceSpec.MemTotalBytes / (1024 * 1024),
			DiskTotalGibiBytes: req.InstanceSpec.DiskTotalBytes / (1024 * 1024 * 1024),
			//DiskTotalGibiBytes:      1024 * 1024 * 1024,
			VcpuTotalCnt:            req.InstanceSpec.VcpuTotalCnt,
			EphemeralTotalGibiBytes: req.InstanceSpec.EphemeralTotalBytes / (1024 * 1024 * 1024),
			SwapTotalMibiBytes:      req.InstanceSpec.SwapTotalBytes / (1024 * 1024),
		},
	})
	if err != nil {
		return nil, err
	}

	defer func() {
		if rollback {
			if err = train.DeleteFlavor(c, train.DeleteFlavorRequest{
				FlavorID: rspFlavor.Flavor.UUID,
			}); err != nil {
				logger.Warnf("Could not delete created Flavor. uuid: %s, Cause: %+v", rspFlavor.Flavor.UUID, err)
			}
		}
	}()

	m := make(map[string]string)
	for _, e := range req.ExtraSpec {
		m[e.Key] = e.Value
	}

	rspExtraSpec, err := train.CreateExtraSpecsForFlavor(c, train.CreateExtraSpecsRequest{
		FlavorUUID: rspFlavor.Flavor.UUID,
		ExtraSpecs: m,
	})
	if err != nil {
		return nil, err
	}

	var mExtraSpecs []model.ClusterInstanceExtraSpec
	for k, v := range rspExtraSpec.ExtraSpecs {
		mExtraSpecs = append(mExtraSpecs, model.ClusterInstanceExtraSpec{
			Key:   k,
			Value: v,
		})
	}

	rollback = false

	return &client.CreateInstanceSpecResponse{
		InstanceSpec: model.ClusterInstanceSpec{
			UUID:                rspFlavor.Flavor.UUID,
			Name:                rspFlavor.Flavor.Name,
			Description:         rspFlavor.Flavor.Description,
			VcpuTotalCnt:        rspFlavor.Flavor.VcpuTotalCnt,
			MemTotalBytes:       rspFlavor.Flavor.MemTotalMibiBytes * 1024 * 1024,
			DiskTotalBytes:      rspFlavor.Flavor.DiskTotalGibiBytes * 1024 * 1024 * 1024,
			SwapTotalBytes:      rspFlavor.Flavor.SwapTotalMibiBytes * 1024 * 1024,
			EphemeralTotalBytes: rspFlavor.Flavor.EphemeralTotalGibiBytes * 1024 * 1024 * 1024,
		},
		ExtraSpec: mExtraSpecs,
	}, nil
}

// 인스턴스에 할당할 네트워크 포트 생성
func createPorts(c client.Client, networks []client.AssignedNetwork, securityGroups []client.AssignedSecurityGroup) ([]string, error) {
	var portIDList []string
	var securityGroupList []string

	//포트 생성시 보안그룹 추가
	for _, sg := range securityGroups {
		securityGroupList = append(securityGroupList, sg.SecurityGroup.UUID)
	}

	for _, n := range networks {
		port := train.PortResult{
			NetworkUUID: n.Network.UUID,
		}
		if len(securityGroupList) > 0 {
			port.SecurityGroups = securityGroupList
		}

		port.FixedIPs = append(port.FixedIPs, train.FixedIP{
			IPAddress:  n.IPAddress,
			SubnetUUID: n.Subnet.UUID,
		})

		portRsp, err := train.CreatePort(c, train.CreatePortRequest{Port: port})
		if err != nil {
			return portIDList, err
		}

		portIDList = append(portIDList, portRsp.Port.UUID)

		if n.FloatingIP == nil {
			continue
		}

		// 생성된 포트에 floating IP 연결
		_, err = train.UpdateFloatingIP(c, train.UpdateFloatingIPRequest{
			FloatingIPID: n.FloatingIP.UUID,
			FloatingIP: train.UpdateFloatingIPResult{
				PortID:         portRsp.Port.UUID,
				FixedIPAddress: n.IPAddress,
				Description:    n.FloatingIP.Description,
			},
		})
		if err != nil {
			return portIDList, err
		}
	}

	return portIDList, nil
}

// 인스턴스에 보안 그룹 추가
func addSecurityGroup(c client.Client, uuid string, req []client.AssignedSecurityGroup) error {
	for _, sg := range req {
		if err := train.AddSecurityGroupToServer(c, train.AddSecurityGroupToServerRequest{
			ServerID: uuid,
			Action: train.AddSecurityGroup{
				Name: sg.SecurityGroup.Name,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

// 인스턴스에 보안 그룹 제거
func removeSecurityGroup(c client.Client, uuid string, name string) error {

	//기본 default 보안그룹 삭제
	if err := train.RemoveSecurityGroupToServer(c, train.RemoveSecurityGroupToServerRequest{
		ServerID: uuid,
		Action: train.RemoveSecurityGroup{
			Name: name,
		},
	}); err != nil {
		logger.Warnf("[addSecurityGroup - RemoveSecurityGroup] (%s) : default", uuid)
	}

	return nil
}

// PreprocessCreatingInstance 인스턴스 생성 전처리 처리
// NFS volume 의 경우 hypervisor 에 NFS export 마운트 처리 진행
func (c *Client) PreprocessCreatingInstance(req client.PreprocessCreatingInstanceRequest) error {
	var agentPort uint
	if req.Hypervisor.AgentPort == nil || *req.Hypervisor.AgentPort == 0 {
		agentPort = 61001
	} else {
		agentPort = uint(*req.Hypervisor.AgentPort)
	}

	// TODO:
	var agent = Agent{
		IP:   req.Hypervisor.IPAddress,
		Port: agentPort,
	}

	for _, s := range req.Storages {
		if s.TypeCode != storage.ClusterStorageTypeNFS {
			continue
		}

		targetStorage, err := storage.GetStorage(s.ClusterID, s.ID)
		if err != nil {
			return err
		}

		md, err := targetStorage.GetMetadata()
		if err != nil {
			return err
		}

		if len(md["exports"].([]interface{})) == 0 {
			return errors.Unknown(errors.New("not found export"))
		}

		var exports = make([]string, len(md["exports"].([]interface{})))
		for i, val := range md["exports"].([]interface{}) {
			exports[i] = val.(string)
		}
	}

	return nil
}

// CreateInstance 인스턴스 생성
func (c *Client) CreateInstance(req client.CreateInstanceRequest) (*client.CreateInstanceResponse, error) {
	var err error
	var rollback = true
	var portIDList []string
	var networks = make([]train.NetworksArray, 0)
	var volumes = make([]train.BlockDeviceMappingV2Array, 0)
	var createRsp *train.CreateServerResponse
	var detailRsp *train.ShowServerDetailsResponse

	var securityGroups = make([]train.SecurityGroupArray, 0)

	defer func() {
		if !rollback {
			return
		}

		if createRsp != nil {
			if err = train.DeleteServer(c, train.DeleteServerRequest{ServerID: createRsp.Server.UUID}); err != nil {
				logger.Warnf("Could not delete created instance. uuid: %s, Cause: %+v", createRsp.Server.UUID, err)
			}
		}

		for _, p := range portIDList {
			if err = train.DeletePort(c, train.DeletePortRequest{PortID: p}); err != nil {
				logger.Warnf("Could not delete created port. uuid: %s, Cause: %+v", p, err)
			}
		}
	}()

	//인스턴스 생성시 보안그룹 추가
	for _, sg := range req.AssignedSecurityGroups {
		securityGroups = append(securityGroups, train.SecurityGroupArray{
			Name: sg.SecurityGroup.Name,
		})
	}

	// 인스턴스에 할당할 네트워크 포트 생성
	if portIDList, err = createPorts(c, req.AssignedNetworks, req.AssignedSecurityGroups); err != nil {
		return nil, err
	}

	// 인스턴스 네트워크
	for _, pid := range portIDList {
		networks = append(networks, train.NetworksArray{PortID: pid})
	}

	// 인스턴스 볼륨
	for _, vol := range req.AttachedVolumes {
		volumes = append(volumes, train.BlockDeviceMappingV2Array{
			UUID:            vol.Volume.UUID,
			BootIndex:       vol.BootIndex,
			DeviceName:      vol.DevicePath,
			SourceType:      constant.SourceTypeVolume,
			DestinationType: constant.DestinationTypeVolume,
		})
	}

	// 인스턴스 생성
	if createRsp, err = train.CreateServer(c, train.CreateServerRequest{
		Server: train.ServerRequest{
			Name:                 req.Instance.Name,
			FlavorRef:            req.InstanceSpec.UUID,
			KeyName:              req.Keypair.Name,
			AvailabilityZoneName: req.AvailabilityZone.Name,
			HypervisorHostname:   req.Hypervisor.Hostname,
			UserData:             req.InstanceUserScript.UserData,
			Description:          req.Instance.Description,
			Networks:             networks,
			BlockDeviceMappingV2: volumes,
			SecurityGroups:       securityGroups,
		},
	}); err != nil {
		return nil, err
	}

	// 보안 그룹 추가
	//if err = addSecurityGroup(c, createRsp.Server.UUID, req.AssignedSecurityGroups); err != nil {
	//	return nil, err
	//}
	//
	//// default 보안 그룹 제거
	//if err = removeSecurityGroup(c, createRsp.Server.UUID, "default"); err != nil {
	//	return nil, err
	//}

	// 생성한 인스턴스 조회
	if detailRsp, err = train.ShowServerDetails(c, train.ShowServerDetailsRequest{ServerID: createRsp.Server.UUID}); err != nil {
		return nil, err
	}

	state, ok := InstanceStateMap[detailRsp.Server.State]
	if !ok {
		return nil, UnknownInstanceState(detailRsp.Server.Name, detailRsp.Server.State)
	}

	rollback = false

	return &client.CreateInstanceResponse{
		Instance: model.ClusterInstance{
			UUID:        detailRsp.Server.UUID,
			Name:        detailRsp.Server.Name,
			Description: detailRsp.Server.Description,
			State:       state,
			Status:      detailRsp.Server.Status,
			Raw:         &detailRsp.Raw,
		},
	}, nil
}

// CreateKeypair Keypair 생성
func (c *Client) CreateKeypair(req client.CreateKeypairRequest) (*client.CreateKeypairResponse, error) {
	rspKeypair, err := train.ImportOrCreateKeypair(c, train.CreateKeypairRequest{
		Keypair: train.KeypairResult{
			Name:      req.Keypair.Name,
			PublicKey: req.Keypair.PublicKey,
			TypeCode:  req.Keypair.TypeCode,
		},
	})
	if err != nil {
		return nil, err
	}

	return &client.CreateKeypairResponse{
		Keypair: model.ClusterKeypair{
			Name:        rspKeypair.Keypair.Name,
			Fingerprint: rspKeypair.Keypair.Fingerprint,
			PublicKey:   rspKeypair.Keypair.PublicKey,
			TypeCode:    rspKeypair.Keypair.TypeCode,
		},
	}, nil
}

// DeleteKeypair Keypair 삭제
func (c *Client) DeleteKeypair(req client.DeleteKeypairRequest) error {
	return train.DeleteKeypair(c, train.DeleteKeypairRequest{
		KeypairName: req.Keypair.Name,
	})
}
