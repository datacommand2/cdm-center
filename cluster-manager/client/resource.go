package client

import (
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
)

const (
	// QuotaComputeInstances 인스턴스 수 제한
	QuotaComputeInstances = "instances"
	// QuotaComputeKeyPairs keypair 수 제한
	QuotaComputeKeyPairs = "key_pairs"
	// QuotaComputeMetadataItems 메타데이터 수 제한
	QuotaComputeMetadataItems = "metadata_items"
	// QuotaComputeRAMSize 램 크기 제한
	QuotaComputeRAMSize = "ram"
	// QuotaComputeServerGroups 서버 그룹 수 제한
	QuotaComputeServerGroups = "server_groups"
	// QuotaComputeServerGroupMembers 서버 그룹의 맴버 수 제한
	QuotaComputeServerGroupMembers = "server_group_members"
	// QuotaComputeVCPUs Virtual CPU 수 제한
	QuotaComputeVCPUs = "vcpus"
	// QuotaComputeInjectedFileContentBytes 삽입 파일 크기 제한
	QuotaComputeInjectedFileContentBytes = "injected_file_content_bytes"
	// QuotaComputeInjectedFilePathBytes 삽입 파일 이름 크기 제한
	QuotaComputeInjectedFilePathBytes = "injected_file_path_bytes"
	// QuotaComputeInjectedFiles 삽입 파일 수 제한
	QuotaComputeInjectedFiles = "injected_files"

	// QuotaNetworkFloatingIPs 부동 IP 수 제한
	QuotaNetworkFloatingIPs = "floating_ip"
	// QuotaNetworkNetworks 네트워크 수 제한
	QuotaNetworkNetworks = "networks"
	// QuotaNetworkSecurityGroupRules 보안 그룹의 규칙 수 제한
	QuotaNetworkSecurityGroupRules = "security_group_rules"
	// QuotaNetworkSecurityGroups 보안 그룹 수 제한
	QuotaNetworkSecurityGroups = "security_groups"
	// QuotaNetworkPorts 포트 수 제한
	QuotaNetworkPorts = "ports"
	// QuotaNetworkRouters 라우터 수 제한
	QuotaNetworkRouters = "routers"
	// QuotaNetworkSubnets 서브넷 수 제한
	QuotaNetworkSubnets = "subnets"
	// QuotaNetworkSubnetPools 서브넷 풀 수 제한
	QuotaNetworkSubnetPools = "subnet_pools"
	// QuotaNetworkRBACPolicies RBAC 규칙 수 제한
	QuotaNetworkRBACPolicies = "rbac_policies"
	// QuotaNetworkTrunks 네트워크 트렁크 수 제한
	//QuotaNetworkTrunks = "trunk"

	// QuotaStorageBackupGigabytes 백업 크기 제한
	QuotaStorageBackupGigabytes = "backup_gigabytes"
	// QuotaStorageBackups 백업 수 제한
	QuotaStorageBackups = "backups"
	// QuotaStorageGigabytes 스토리지 크기 제한
	QuotaStorageGigabytes = "gigabytes"
	// QuotaStorageGroups 스토리지 그룹 수 제한
	QuotaStorageGroups = "groups"
	// QuotaStoragePerVolumeGigabytes 볼륨 당 크기 제한
	QuotaStoragePerVolumeGigabytes = "per_volume_gigabytes"
	// QuotaStorageSnapshots 스냅샷 수 제한
	QuotaStorageSnapshots = "snapshots"
	// QuotaStorageVolumes 볼륨 수 제한
	QuotaStorageVolumes = "volumes"
)

// QuotaKeys 정의된 Quota 키 명칭
var QuotaKeys = []interface{}{
	QuotaComputeInstances,
	QuotaComputeKeyPairs,
	QuotaComputeMetadataItems,
	QuotaComputeRAMSize,
	QuotaComputeServerGroups,
	QuotaComputeServerGroupMembers,
	QuotaComputeVCPUs,
	QuotaComputeInjectedFileContentBytes,
	QuotaComputeInjectedFilePathBytes,
	QuotaComputeInjectedFiles,

	QuotaNetworkFloatingIPs,
	QuotaNetworkNetworks,
	QuotaNetworkSecurityGroupRules,
	QuotaNetworkSecurityGroups,
	QuotaNetworkPorts,
	QuotaNetworkRouters,
	QuotaNetworkSubnets,
	QuotaNetworkSubnetPools,
	QuotaNetworkRBACPolicies,
	//QuotaNetworkTrunks,

	QuotaStorageBackupGigabytes,
	QuotaStorageBackups,
	QuotaStorageGigabytes,
	QuotaStorageGroups,
	QuotaStoragePerVolumeGigabytes,
	QuotaStorageSnapshots,
	QuotaStorageVolumes,
}

// IsQuotaKey 유효한 quota key 판단
func IsQuotaKey(key string) bool {
	for _, quotaKey := range QuotaKeys {
		if quotaKey == key {
			return true
		}
	}

	return false
}

// CreateTenantRequest 테넌트 생성 요청
type CreateTenantRequest struct {
	Tenant    model.ClusterTenant
	QuotaList []model.ClusterQuota
}

// CreateTenantResponse 테넌트 생성 응답
type CreateTenantResponse struct {
	Tenant    model.ClusterTenant
	QuotaList []model.ClusterQuota
}

// GetTenantResult 테넌트 조회 Result 구조체
type GetTenantResult struct {
	Tenant    model.ClusterTenant
	QuotaList []model.ClusterQuota
}

// GetTenantRequest 테넌트 조회 Request 구조체
type GetTenantRequest struct {
	Tenant model.ClusterTenant
}

// GetTenantListResponse 테넌트 목록 조회 Response 구조체
type GetTenantListResponse struct {
	ResultList []GetTenantResult
}

// DeleteTenantRequest 테넌트 삭제 Request 구조체
type DeleteTenantRequest struct {
	Tenant model.ClusterTenant
}

// GetTenantResponse 테넌트 조회 Response 구조체
type GetTenantResponse struct {
	Result GetTenantResult
}

// GetHypervisorResult 하이퍼바이저 조회 Result 구조체
type GetHypervisorResult struct {
	Hypervisor       model.ClusterHypervisor
	AvailabilityZone model.ClusterAvailabilityZone
}

// GetHypervisorRequest 하이퍼바이저 조회 Request 구조체
type GetHypervisorRequest struct {
	Hypervisor model.ClusterHypervisor
}

// GetHypervisorListResponse 하이퍼바이저 목록 조회 Response 구조체
type GetHypervisorListResponse struct {
	ResultList []GetHypervisorResult
}

// GetHypervisorResponse 하이퍼바이저 조회 Response 구조체
type GetHypervisorResponse struct {
	Result GetHypervisorResult
}

// GetAvailabilityZoneResult 가용구역 조회 Result 구조체
type GetAvailabilityZoneResult struct {
	AvailabilityZone model.ClusterAvailabilityZone
}

// GetAvailabilityZoneListResponse 가용구역 목록 조회 Response 구조체
type GetAvailabilityZoneListResponse struct {
	ResultList []GetAvailabilityZoneResult
}

// GetInstanceVolumeResult 인스턴스 볼륨 result 구조체
type GetInstanceVolumeResult struct {
	InstanceVolume model.ClusterInstanceVolume
	Volume         model.ClusterVolume
}

// GetInstanceNetworkResult 인스턴스 네트워크 result 구조체
type GetInstanceNetworkResult struct {
	InstanceNetwork model.ClusterInstanceNetwork
	Network         model.ClusterNetwork
	Subnet          model.ClusterSubnet
	FloatingIP      *model.ClusterFloatingIP
}

// GetInstanceResult instance 조회 Result 구조체
type GetInstanceResult struct {
	Instance            model.ClusterInstance
	Tenant              model.ClusterTenant
	InstanceSpec        model.ClusterInstanceSpec
	KeyPair             *model.ClusterKeypair
	Hypervisor          model.ClusterHypervisor
	AvailabilityZone    model.ClusterAvailabilityZone
	SecurityGroupList   []model.ClusterSecurityGroup
	InstanceVolumeList  []GetInstanceVolumeResult
	InstanceNetworkList []GetInstanceNetworkResult
}

// GetInstanceRequest instance 조회 Request 구조체
type GetInstanceRequest struct {
	Instance model.ClusterInstance
}

// GetInstanceListResponse instance 목록 조회 Response 구조체
type GetInstanceListResponse struct {
	ResultList []GetInstanceResult
}

// GetInstanceResponse instance 조회 Response 구조체
type GetInstanceResponse struct {
	Result GetInstanceResult
}

// DeleteInstanceRequest instance 삭제 Request 구조체
type DeleteInstanceRequest struct {
	Tenant   model.ClusterTenant
	Instance model.ClusterInstance
}

// GetInstanceSpecResult instance spec 조회 Result 구조체
type GetInstanceSpecResult struct {
	InstanceSpec  model.ClusterInstanceSpec
	ExtraSpecList []model.ClusterInstanceExtraSpec
}

// GetInstanceSpecRequest instance spec 조회 Request 구조체
type GetInstanceSpecRequest struct {
	InstanceSpec model.ClusterInstanceSpec
}

// GetInstanceSpecResponse instance spec 조회 Response 구조체
type GetInstanceSpecResponse struct {
	Result GetInstanceSpecResult
}

// GetInstanceExtraSpecResult instance extra spec 조회 Result 구조체
type GetInstanceExtraSpecResult struct {
	InstanceExtraSpec model.ClusterInstanceExtraSpec
}

// GetInstanceExtraSpecRequest instance extra spec 목록 조회 Request 구조체
type GetInstanceExtraSpecRequest struct {
	InstanceSpec      model.ClusterInstanceSpec
	InstanceExtraSpec model.ClusterInstanceExtraSpec
}

// GetInstanceExtraSpecResponse instance extra spec 조회 Response 구조체
type GetInstanceExtraSpecResponse struct {
	Result GetInstanceExtraSpecResult
}

// DeleteInstanceSpecRequest instance 삭제 Request 구조체
type DeleteInstanceSpecRequest struct {
	InstanceSpec model.ClusterInstanceSpec
}

// StartInstanceRequest instance 기동 Request 구조체
type StartInstanceRequest struct {
	Tenant   model.ClusterTenant
	Instance model.ClusterInstance
}

// StopInstanceRequest instance 중지 Request 구조체
type StopInstanceRequest struct {
	Tenant   model.ClusterTenant
	Instance model.ClusterInstance
}

// GetKeyPairResult 키페어 조회 Result 구조체
type GetKeyPairResult struct {
	KeyPair model.ClusterKeypair
}

// GetKeyPairRequest 키페어 조회 Request 구조체
type GetKeyPairRequest struct {
	KeyPair model.ClusterKeypair
}

// GetKeyPairListResponse 키페어 목록 조회 Response 구조체
type GetKeyPairListResponse struct {
	ResultList []GetKeyPairResult
}

// GetKeyPairResponse 키페어 조회 Response 구조체
type GetKeyPairResponse struct {
	Result GetKeyPairResult
}

// CreateSecurityGroupRequest security group 생성 요청
type CreateSecurityGroupRequest struct {
	Tenant        model.ClusterTenant
	SecurityGroup model.ClusterSecurityGroup
}

// CreateSecurityGroupResponse security group 생성 응답
type CreateSecurityGroupResponse struct {
	SecurityGroup model.ClusterSecurityGroup
}

// CreateSecurityGroupRuleRequest security group rule 생성 요청
type CreateSecurityGroupRuleRequest struct {
	Tenant              model.ClusterTenant
	SecurityGroup       model.ClusterSecurityGroup
	Rule                model.ClusterSecurityGroupRule
	RemoteSecurityGroup *model.ClusterSecurityGroup
}

// CreateSecurityGroupRuleResponse security group rule 생성 응답
type CreateSecurityGroupRuleResponse struct {
	Rule                model.ClusterSecurityGroupRule
	RemoteSecurityGroup *model.ClusterSecurityGroup
}

// DeleteSecurityGroupRequest security group 삭제 요청
type DeleteSecurityGroupRequest struct {
	Tenant        model.ClusterTenant
	SecurityGroup model.ClusterSecurityGroup
}

// DeleteSecurityGroupRuleRequest security group rule 삭제 요청
type DeleteSecurityGroupRuleRequest struct {
	Tenant        model.ClusterTenant
	SecurityGroup model.ClusterSecurityGroup
	Rule          model.ClusterSecurityGroupRule
}

// GetSecurityGroupResult security group 조회 Result 구조체
type GetSecurityGroupResult struct {
	SecurityGroup model.ClusterSecurityGroup
	Tenant        model.ClusterTenant
	RuleList      []model.ClusterSecurityGroupRule
}

// GetSecurityGroupRequest security group 조회 Request 구조체
type GetSecurityGroupRequest struct {
	SecurityGroup model.ClusterSecurityGroup
}

// GetSecurityGroupListResponse security group 목록 조회 Response 구조체
type GetSecurityGroupListResponse struct {
	ResultList []GetSecurityGroupResult
}

// GetSecurityGroupResponse security group 조회 Response 구조체
type GetSecurityGroupResponse struct {
	Result GetSecurityGroupResult
}

// GetSecurityGroupRuleResult security group rule 조회 Result 구조체
type GetSecurityGroupRuleResult struct {
	SecurityGroupRule   model.ClusterSecurityGroupRule
	SecurityGroup       model.ClusterSecurityGroup
	RemoteSecurityGroup *model.ClusterSecurityGroup
}

// GetSecurityGroupRuleRequest security group rule 조회 Request 구조체
type GetSecurityGroupRuleRequest struct {
	SecurityGroupRule model.ClusterSecurityGroupRule
}

// GetSecurityGroupRuleResponse security group rule 조회 Response 구조체
type GetSecurityGroupRuleResponse struct {
	Result GetSecurityGroupRuleResult
}

// CreateNetworkRequest 네트워크 생성 요청
type CreateNetworkRequest struct {
	Tenant  model.ClusterTenant
	Network model.ClusterNetwork
}

// CreateNetworkResponse 네트워크 생성 응답
type CreateNetworkResponse struct {
	Network model.ClusterNetwork
}

// DeleteNetworkRequest 네트워크 삭제 요청
type DeleteNetworkRequest struct {
	Tenant  model.ClusterTenant
	Network model.ClusterNetwork
}

// CreateSubnetRequest 서브넷 생성 요청
type CreateSubnetRequest struct {
	Tenant         model.ClusterTenant
	Network        model.ClusterNetwork
	Subnet         model.ClusterSubnet
	NameserverList []model.ClusterSubnetNameserver
	PoolsList      []model.ClusterSubnetDHCPPool
}

// CreateSubnetResponse 서브넷 생성 응답
type CreateSubnetResponse struct {
	Subnet         model.ClusterSubnet
	NameserverList []model.ClusterSubnetNameserver
	PoolsList      []model.ClusterSubnetDHCPPool
}

// DeleteSubnetRequest 서브넷 삭제 요청
type DeleteSubnetRequest struct {
	Tenant  model.ClusterTenant
	Network model.ClusterNetwork
	Subnet  model.ClusterSubnet
}

// GetNetworkResult 네트워크 조회 Result 구조체
type GetNetworkResult struct {
	Network        model.ClusterNetwork
	Tenant         model.ClusterTenant
	SubnetList     []model.ClusterSubnet
	FloatingIPList []model.ClusterFloatingIP
}

// GetNetworkRequest 네트워크 조회 Request 구조체
type GetNetworkRequest struct {
	Network model.ClusterNetwork
}

// GetNetworkListResponse 네트워크 목록 조회 Response 구조체
type GetNetworkListResponse struct {
	ResultList []GetNetworkResult
}

// GetNetworkResponse 네트워크 조회 Response 구조체
type GetNetworkResponse struct {
	Result GetNetworkResult
}

// GetSubnetResult 서브넷 조회 Result 구조체
type GetSubnetResult struct {
	Subnet         model.ClusterSubnet // GatewayEnabled 는 GatewayIPAddress 값이 null 일때 true
	Network        model.ClusterNetwork
	NameserverList []model.ClusterSubnetNameserver
	PoolList       []model.ClusterSubnetDHCPPool
}

// GetSubnetRequest 서브넷 조회 Request 구조체
type GetSubnetRequest struct {
	Subnet model.ClusterSubnet
}

// GetSubnetResponse 서브넷 조회 Response 구조체
type GetSubnetResponse struct {
	Result GetSubnetResult
}

// RoutingInterfaceResult 라우팅 인터페이스 Result 구조체
type RoutingInterfaceResult struct {
	Subnet           model.ClusterSubnet
	RoutingInterface model.ClusterNetworkRoutingInterface
}

// GetRouterResult 라우터 조회 Result 구조체
type GetRouterResult struct {
	Router                       model.ClusterRouter
	Tenant                       model.ClusterTenant
	ExtraRoute                   []model.ClusterRouterExtraRoute
	ExternalRoutingInterfaceList []RoutingInterfaceResult
	InternalRoutingInterfaceList []RoutingInterfaceResult
}

// GetRouterRequest 라우터 조회 Request 구조체
type GetRouterRequest struct {
	Router model.ClusterRouter
}

// CreateRoutingInterface 라우터 인터페이스
type CreateRoutingInterface struct {
	Network          model.ClusterNetwork
	Subnet           model.ClusterSubnet
	RoutingInterface model.ClusterNetworkRoutingInterface
}

// CreateRouterRequest 라우터 생성 요청
type CreateRouterRequest struct {
	Tenant                       model.ClusterTenant
	Router                       model.ClusterRouter
	ExtraRouteList               []model.ClusterRouterExtraRoute
	ExternalRoutingInterfaceList []CreateRoutingInterface
	InternalRoutingInterfaceList []CreateRoutingInterface
}

// CreateRouterResponse 라우터 생성 응답
type CreateRouterResponse struct {
	Router                       model.ClusterRouter
	ExtraRouteList               []model.ClusterRouterExtraRoute
	ExternalRoutingInterfaceList []CreateRoutingInterface
	InternalRoutingInterfaceList []CreateRoutingInterface
}

// DeleteRouterRequest 라우터 삭제 요청
type DeleteRouterRequest struct {
	Tenant model.ClusterTenant
	Router model.ClusterRouter
}

// GetRouterListResponse 라우터 목록 조회 Response 구조체
type GetRouterListResponse struct {
	ResultList []GetRouterResult
}

// GetRouterResponse 라우터 조회 Response 구조체
type GetRouterResponse struct {
	Result GetRouterResult
}

// DeleteFloatingIPRequest 부동 IP 삭제 요청
type DeleteFloatingIPRequest struct {
	Tenant     model.ClusterTenant
	FloatingIP model.ClusterFloatingIP
}

// CreateFloatingIPRequest 부동 IP 생성 요청
type CreateFloatingIPRequest struct {
	Tenant     model.ClusterTenant
	Network    model.ClusterNetwork
	FloatingIP model.ClusterFloatingIP
}

// CreateFloatingIPResponse 부동 IP 생성 응답
type CreateFloatingIPResponse struct {
	FloatingIP model.ClusterFloatingIP
}

// GetFloatingIPResult floating ip 조회 Result 구조체
type GetFloatingIPResult struct {
	Tenant     model.ClusterTenant
	FloatingIP model.ClusterFloatingIP
	Network    model.ClusterNetwork
}

// GetFloatingIPRequest floating ip 조회 Request 구조체
type GetFloatingIPRequest struct {
	FloatingIP model.ClusterFloatingIP
}

// GetFloatingIPResponse floating ip 조회 Response 구조체
type GetFloatingIPResponse struct {
	Result GetFloatingIPResult
}

// GetStorageResult 스토리지 조회 Result 구조체
type GetStorageResult struct {
	Storage           model.ClusterStorage
	VolumeBackendName string
	Metadata          map[string]interface{}
}

// GetStorageRequest 스토리지 조회 Request 구조체
// TODO:
type GetStorageRequest struct {
	Storage model.ClusterStorage
	Agent   Agent
}

// GetStorageListResponse 스토리지 목록 조회 Response 구조체
type GetStorageListResponse struct {
	ResultList []GetStorageResult
}

// GetStorageResponse 스토리지 조회 Response 구조체
type GetStorageResponse struct {
	Result GetStorageResult
}

// GetVolumeSnapshotResult 볼륨 스냅샷 조회 Result 구조체
type GetVolumeSnapshotResult struct {
	VolumeSnapshot model.ClusterVolumeSnapshot
	Volume         model.ClusterVolume
}

// GetVolumeSnapshotRequest 볼륨 스냅샷 조회 Request 구조체
type GetVolumeSnapshotRequest struct {
	VolumeSnapshot model.ClusterVolumeSnapshot
}

// GetVolumeSnapshotListResponse 볼륨 스냅샷 목록 조회 Response 구조체
type GetVolumeSnapshotListResponse struct {
	ResultList []GetVolumeSnapshotResult
}

// GetVolumeSnapshotResponse 볼륨 스냅샷 조회 Response 구조체
type GetVolumeSnapshotResponse struct {
	Result GetVolumeSnapshotResult
}

// GetVolumeResult 볼륨 조회 Result 구조체
type GetVolumeResult struct {
	Volume   model.ClusterVolume
	Tenant   model.ClusterTenant
	Storage  model.ClusterStorage
	Metadata map[string]interface{}
}

// GetVolumeRequest 볼륨 조회 Request 구조체
type GetVolumeRequest struct {
	Volume model.ClusterVolume
}

// GetVolumeListResponse 볼륨 목록 조회 Response 구조체
type GetVolumeListResponse struct {
	ResultList []GetVolumeResult
}

// GetVolumeResponse 볼륨 조회 Response 구조체
type GetVolumeResponse struct {
	Result GetVolumeResult
}

// VolumeGroup 볼륨 그룹
type VolumeGroup struct {
	UUID        string
	Name        string
	Description string
	StorageList []model.ClusterStorage
	VolumeList  []model.ClusterVolume
}

// GetVolumeGroupRequest 볼륨 그룹 조회 요청
type GetVolumeGroupRequest struct {
	Tenant      model.ClusterTenant
	VolumeGroup VolumeGroup
}

// GetVolumeGroupResponse 볼륨 그룹 조회 응답
type GetVolumeGroupResponse struct {
	VolumeGroup VolumeGroup
}

// GetVolumeGroupListResponse 볼륨 목록 조회 Response 구조체
type GetVolumeGroupListResponse struct {
	VolumeGroupList []VolumeGroup
}

// CreateVolumeGroupRequest 볼륨 그룹 생성 요청
type CreateVolumeGroupRequest struct {
	Tenant      model.ClusterTenant
	VolumeGroup VolumeGroup
}

// CreateVolumeGroupResponse 볼륨 그룹 생성 응답
type CreateVolumeGroupResponse struct {
	VolumeGroup VolumeGroup
}

// DeleteVolumeGroupRequest 볼륨 그룹 삭제 요청
type DeleteVolumeGroupRequest struct {
	Tenant      model.ClusterTenant
	VolumeGroup VolumeGroup
}

// UpdateVolumeGroupRequest 볼륨 그룹 수정 요청
type UpdateVolumeGroupRequest struct {
	Tenant      model.ClusterTenant
	VolumeGroup VolumeGroup
	AddVolumes  []model.ClusterVolume
	DelVolumes  []model.ClusterVolume
}

// VolumeGroupSnapshot 볼륨 그룹 스냅샷
type VolumeGroupSnapshot struct {
	UUID        string
	Name        string
	Description string
}

// CreateVolumeGroupSnapshotRequest 볼륨 그룹 스냅샷 생성 요청
type CreateVolumeGroupSnapshotRequest struct {
	Tenant              model.ClusterTenant
	VolumeGroup         VolumeGroup
	VolumeGroupSnapshot VolumeGroupSnapshot
}

// CreateVolumeGroupSnapshotResponse 볼륨 그룹 스냅샷 생성 응답
type CreateVolumeGroupSnapshotResponse struct {
	VolumeGroupSnapshot VolumeGroupSnapshot
}

// DeleteVolumeGroupSnapshotRequest 볼륨 그룹 스냅샷 삭제 요청
type DeleteVolumeGroupSnapshotRequest struct {
	Tenant              model.ClusterTenant
	VolumeGroup         VolumeGroup
	VolumeGroupSnapshot VolumeGroupSnapshot
}

// GetVolumeGroupSnapshotListResult 볼륨 그룹 스냅샷 목록 조회 결과
type GetVolumeGroupSnapshotListResult struct {
	VolumeGroup         VolumeGroup
	VolumeGroupSnapshot VolumeGroupSnapshot
}

// GetVolumeGroupSnapshotListResponse 볼륨 그룹 스냅샷 목록 조회 응답
type GetVolumeGroupSnapshotListResponse struct {
	VolumeGroupSnapshots []GetVolumeGroupSnapshotListResult
}

// CreateVolumeRequest 볼륨 생성 Request 구조체
type CreateVolumeRequest struct {
	Tenant  model.ClusterTenant
	Storage model.ClusterStorage
	Volume  model.ClusterVolume
}

// CreateVolumeResponse 볼륨 생성 Response 구조체
type CreateVolumeResponse struct {
	Volume model.ClusterVolume
}

// ImportVolumeRequest 볼륨 import Request 구조체
type ImportVolumeRequest struct {
	Tenant          model.ClusterTenant
	SourceStorage   model.ClusterStorage
	TargetStorage   model.ClusterStorage
	TargetMetadata  map[string]string
	Volume          model.ClusterVolume
	VolumeSnapshots []model.ClusterVolumeSnapshot
}

// ImportVolumeResponse 볼륨 import Response 구조체
type ImportVolumeResponse struct {
	VolumePair    VolumePair
	SnapshotPairs []SnapshotPair
}

// CopyVolumeRequest 볼륨 copy Request 구조체
type CopyVolumeRequest struct {
	Tenant         model.ClusterTenant
	SourceStorage  model.ClusterStorage
	TargetStorage  model.ClusterStorage
	TargetMetadata map[string]string
	Volume         model.ClusterVolume
	Snapshots      []model.ClusterVolumeSnapshot
}

// CopyVolumeResponse 볼륨 copy Response 구조체
type CopyVolumeResponse struct {
	Volume    model.ClusterVolume
	Snapshots []model.ClusterVolumeSnapshot
}

// DeleteVolumeCopyRequest 볼륨 copy 삭제 Request 구조체
type DeleteVolumeCopyRequest struct {
	Tenant         model.ClusterTenant
	SourceStorage  model.ClusterStorage
	TargetStorage  model.ClusterStorage
	TargetMetadata map[string]string
	Volume         model.ClusterVolume
	VolumeID       uint64
	Snapshots      []model.ClusterVolumeSnapshot
}

// DeleteVolumeRequest 볼륨 삭제 Request 구조체
type DeleteVolumeRequest struct {
	Tenant model.ClusterTenant
	Volume model.ClusterVolume
}

// VolumePair 볼륨 source, target 구조체
type VolumePair struct {
	Source     model.ClusterVolume
	Target     model.ClusterVolume
	SourceFile string
	TargetFile string
}

// SnapshotPair 스냅샷 source, target 구조체
type SnapshotPair struct {
	Source               model.ClusterVolumeSnapshot
	Target               model.ClusterVolumeSnapshot
	SourceFile           string
	TargetFile           string
	FailModifyHeaderFlag bool
}

// UnmanageVolumeRequest 볼륨 unmanage 요청
type UnmanageVolumeRequest struct {
	Tenant         model.ClusterTenant
	SourceStorage  model.ClusterStorage
	TargetStorage  model.ClusterStorage
	TargetMetadata map[string]string
	VolumePair     VolumePair
	VolumeID       uint64
	SnapshotPairs  []SnapshotPair
}

// CreateVolumeSnapshotRequest 볼륨 스냅샷 생성 Request 구조체
type CreateVolumeSnapshotRequest struct {
	Tenant         model.ClusterTenant
	Volume         model.ClusterVolume
	VolumeSnapshot model.ClusterVolumeSnapshot
}

// CreateVolumeSnapshotResponse 볼륨 스냅샷 생성 Response 구조체
type CreateVolumeSnapshotResponse struct {
	VolumeSnapshot model.ClusterVolumeSnapshot
}

// DeleteVolumeSnapshotRequest 볼륨 스냅샷 삭제 Request 구조체
type DeleteVolumeSnapshotRequest struct {
	Tenant         model.ClusterTenant
	VolumeSnapshot model.ClusterVolumeSnapshot
}

// CreateInstanceSpecRequest 인스턴스 스펙 생성 Request 구조체
type CreateInstanceSpecRequest struct {
	InstanceSpec model.ClusterInstanceSpec
	ExtraSpec    []model.ClusterInstanceExtraSpec
}

// CreateInstanceSpecResponse 인스턴스 스펙 생성 Response 구조체
type CreateInstanceSpecResponse struct {
	InstanceSpec model.ClusterInstanceSpec
	ExtraSpec    []model.ClusterInstanceExtraSpec
}

// CreateKeypairRequest keypair 생성 Request 구조체
type CreateKeypairRequest struct {
	Keypair model.ClusterKeypair
}

// CreateKeypairResponse keypair 생성 Response 구조체
type CreateKeypairResponse struct {
	Keypair model.ClusterKeypair
}

// DeleteKeypairRequest keypair 삭제 Request 구조체
type DeleteKeypairRequest struct {
	Keypair model.ClusterKeypair
}

// AssignedNetwork 인스턴스에 할당된 네트워크 구조체
type AssignedNetwork struct {
	Network    model.ClusterNetwork
	Subnet     model.ClusterSubnet
	FloatingIP *model.ClusterFloatingIP
	DhcpFlag   bool
	IPAddress  string
}

// AssignedSecurityGroup 인스턴스에 할당된 보안 그룹 구조체
type AssignedSecurityGroup struct {
	SecurityGroup model.ClusterSecurityGroup
}

// AttachedVolume 인스턴스에 attach 된 볼륨 구조체
type AttachedVolume struct {
	Volume     model.ClusterVolume
	DevicePath string
	BootIndex  int64
}

// CreateInstanceRequest 인스턴스 생성 Request 구조체
type CreateInstanceRequest struct {
	Tenant                 model.ClusterTenant
	Hypervisor             model.ClusterHypervisor
	AvailabilityZone       model.ClusterAvailabilityZone
	InstanceSpec           model.ClusterInstanceSpec
	Keypair                model.ClusterKeypair
	Instance               model.ClusterInstance
	InstanceUserScript     model.ClusterInstanceUserScript
	AssignedNetworks       []AssignedNetwork
	AssignedSecurityGroups []AssignedSecurityGroup
	AttachedVolumes        []AttachedVolume
}

// CreateInstanceResponse 인스턴스 생성 Response 구조체
type CreateInstanceResponse struct {
	Instance model.ClusterInstance
}

// PreprocessCreatingInstanceRequest 인스턴스 생성 전처리 Request 구조체
type PreprocessCreatingInstanceRequest struct {
	Storages   []*model.ClusterStorage
	Hypervisor *model.ClusterHypervisor
}
