package train

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack/train/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"net/http"
	"time"
)

// ServerFlavor server flavor 구조체
type ServerFlavor struct {
	Name string `json:"original_name,omitempty"`
}

// AttachedVolumes attached volumes 구조체
type AttachedVolumes struct {
	VolumeID string `json:"id,omitempty"`
}

// SecurityGroupArray SecurityGroupArray 구조체
type SecurityGroupArray struct {
	Name string `json:"name,omitempty"`
}

// ServerResult 서버 result 구조체
type ServerResult struct {
	UUID            string               `json:"id,omitempty"`
	Name            string               `json:"name,omitempty"`
	Description     *string              `json:"description,omitempty"`
	State           int64                `json:"OS-EXT-STS:power_state,omitempty"`
	Status          string               `json:"status,omitempty"`
	KeyName         *string              `json:"key_name,omitempty"`
	TenantID        string               `json:"tenant_id,omitempty"`
	Flavor          ServerFlavor         `json:"flavor,omitempty"`
	AzName          string               `json:"OS-EXT-AZ:availability_zone,omitempty"`
	AttachedVolumes []AttachedVolumes    `json:"os-extended-volumes:volumes_attached,omitempty"`
	SecurityGroups  []SecurityGroupArray `json:"security_groups,omitempty"`
}

// ListServersResponse 서버 목록 조회 response 구조체
type ListServersResponse struct {
	Servers []ServerResult `json:"servers,omitempty"`
}

// ShowServerDetailsRequest 서버 상세 조회 request 구조체
type ShowServerDetailsRequest struct {
	ServerID string
}

// ShowServerDetailsResponse 서버 상세 조회 response 구조체
type ShowServerDetailsResponse struct {
	Server ServerResult `json:"server,omitempty"`
	Raw    string
}

// ListSecurityGroupsByServerRequest 서버의 보안 그룹 목록 조회 request 구조체
type ListSecurityGroupsByServerRequest struct {
	ServerID string
}

// ListFlavorsResponse flavor 목록 조회 response 구조체
type ListFlavorsResponse struct {
	Flavors []FlavorResult `json:"flavors,omitempty"`
}

// ShowDetailVolumeAttachmentResult 서버의 볼륨 상세 조회 result 구조체
type ShowDetailVolumeAttachmentResult struct {
	DevicePath string `json:"device,omitempty"`
	VolumeID   string `json:"volumeId,omitempty"`
}

// ShowDetailVolumeAttachmentRequest 서버의 볼륨 상세 조회 request 구조체
type ShowDetailVolumeAttachmentRequest struct {
	ServerID string
	VolumeID string
}

// ShowDetailVolumeAttachmentResponse 서버의 볼륨 상세 조회 response 구조체
type ShowDetailVolumeAttachmentResponse struct {
	VolumeAttachment ShowDetailVolumeAttachmentResult `json:"volumeAttachment"`
}

// ListFloatingIPsByPortInterfaceRequest PortID를 통한 네트워크 floating ip 목록 조회 request 구조체
type ListFloatingIPsByPortInterfaceRequest struct {
	PortID  string
	FixedIP string
}

// PortInterfaceResult 서버 포트 인터페이스 result 구조체
type PortInterfaceResult struct {
	FixedIPs  []FixedIP `json:"fixed_ips,omitempty"`
	NetworkID string    `json:"net_id,omitempty"`
	PortID    string    `json:"port_id,omitempty"`
}

// ListPortInterfacesRequest 서버 포트 인터페이스 목록 조회 request 구조체
type ListPortInterfacesRequest struct {
	ServerID string
}

// ListPortInterfacesResponse 서버 포트 인터페이스 목록 조회 response 구조체
type ListPortInterfacesResponse struct {
	PortInterfaces []PortInterfaceResult `json:"interfaceAttachments,omitempty"`
}

// NetworksArray 인스턴스에 연결된 네트워크 구조체
type NetworksArray struct {
	PortID string `json:"port"`
}

// BlockDeviceMappingV2Array 인스턴스에 연결된 볼륨 구조체
type BlockDeviceMappingV2Array struct {
	UUID            string `json:"uuid"`
	BootIndex       int64  `json:"boot_index"`
	DeviceName      string `json:"device_name,omitempty"`
	SourceType      string `json:"source_type"`
	DestinationType string `json:"destination_type,omitempty"`
	VolumeType      string `json:"volume_type,omitempty"`
}

// ServerRequest 인스턴스 ServerRequest 구조체
type ServerRequest struct {
	Name                 string                      `json:"name,omitempty"`
	FlavorRef            string                      `json:"flavorRef"`
	KeyName              string                      `json:"key_name,omitempty"`
	AvailabilityZoneName string                      `json:"availability_zone,omitempty"`
	HypervisorHostname   string                      `json:"hypervisor_hostname,omitempty"`
	UserData             string                      `json:"user_data,omitempty"`
	Description          *string                     `json:"description,omitempty"`
	Networks             []NetworksArray             `json:"networks"`
	BlockDeviceMappingV2 []BlockDeviceMappingV2Array `json:"block_device_mapping_v2"`
	SecurityGroups       []SecurityGroupArray        `json:"security_groups,omitempty"`
}

// ServerResponse 인스턴스 ServerResponse 구조체
type ServerResponse struct {
	UUID string `json:"id,omitempty"`
}

// CreateServerRequest 인스턴스 생성 request 구조체
type CreateServerRequest struct {
	Server ServerRequest `json:"server"`
}

// CreateServerResponse 인스턴스 생성 response 구조체
type CreateServerResponse struct {
	Server ServerResponse `json:"server"`
	Raw    *string
}

// DeletePortRequest 네트워크 포트 삭제 Request 구조체
type DeletePortRequest struct {
	PortID string
}

// AddSecurityGroup 인스턴스에 보안 그룹 추가
type AddSecurityGroup struct {
	Name string `json:"name"`
}

// AddSecurityGroupToServerRequest 인스턴스에 보안 그룹 추가 Request 구조체
type AddSecurityGroupToServerRequest struct {
	ServerID string           `json:"-"`
	Action   AddSecurityGroup `json:"addSecurityGroup"`
}

// DeleteServerRequest 인스턴스 삭제 request 구조체
type DeleteServerRequest struct {
	ServerID string
}

// StartServerRequest 볼륨의 StartServerRequest 구조체
type StartServerRequest struct {
	ServerID string
}

// StopServerRequest 볼륨의 StopServerRequest 구조체
type StopServerRequest struct {
	ServerID string
}

// RemoveSecurityGroup 인스턴스에 보안 그룹 삭제
type RemoveSecurityGroup struct {
	Name string `json:"name"`
}

// RemoveSecurityGroupToServerRequest 인스턴스에 보안 그룹 삭제 Request 구조체
type RemoveSecurityGroupToServerRequest struct {
	ServerID string              `json:"-"`
	Action   RemoveSecurityGroup `json:"removeSecurityGroup"`
}

// ListServers 서버 목록 조회
func ListServers(c client.Client) (*ListServersResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/servers?all_tenants=true",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListServers", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return nil, client.Unknown(response)
		case response.Code == 401:
			return nil, client.UnAuthenticated(response)
		case response.Code == 403:
			return nil, client.UnAuthorized(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp ListServersResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ShowServerDetails 서버 상세 조회
func ShowServerDetails(c client.Client, req ShowServerDetailsRequest) (*ShowServerDetailsResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/servers/" + req.ServerID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowServerDetails", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return nil, client.Unknown(response)
		case response.Code == 401:
			return nil, client.UnAuthenticated(response)
		case response.Code == 403:
			return nil, client.UnAuthorized(response)
		case response.Code == 404:
			return nil, client.NotFound(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp ShowServerDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	rsp.Raw = response.Body

	return &rsp, nil
}

// ListSecurityGroupsByServer 서버의 보안 그룹 목록 조회
func ListSecurityGroupsByServer(c client.Client, req ListSecurityGroupsByServerRequest) (*ListSecurityGroupsResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/servers/" + req.ServerID + "/os-security-groups",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListSecurityGroupsByServer", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return nil, client.Unknown(response)
		case response.Code == 401:
			return nil, client.UnAuthenticated(response)
		case response.Code == 403:
			return nil, client.UnAuthorized(response)
		case response.Code == 404:
			return nil, client.NotFound(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp ListSecurityGroupsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ListFlavors flavor 목록 조회
func ListFlavors(c client.Client) (*ListFlavorsResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/flavors",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListFlavors", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return nil, client.Unknown(response)
		case response.Code == 401:
			return nil, client.UnAuthenticated(response)
		case response.Code == 403:
			return nil, client.UnAuthorized(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp ListFlavorsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ShowDetailVolumeAttachment 서버 볼륨 상세 조회
func ShowDetailVolumeAttachment(c client.Client, req ShowDetailVolumeAttachmentRequest) (*ShowDetailVolumeAttachmentResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/servers/" + req.ServerID + "/os-volume_attachments/" + req.VolumeID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowDetailVolumeAttachment", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return nil, client.Unknown(response)
		case response.Code == 401:
			return nil, client.UnAuthenticated(response)
		case response.Code == 403:
			return nil, client.UnAuthorized(response)
		case response.Code == 404:
			return nil, client.NotFound(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}

	}

	var rsp ShowDetailVolumeAttachmentResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ListPortInterfaces 서버 포트 인터페이스 목록 조회
func ListPortInterfaces(c client.Client, req ListPortInterfacesRequest) (*ListPortInterfacesResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/servers/" + req.ServerID + "/os-interface",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListPortInterfaces", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return nil, client.Unknown(response)
		case response.Code == 401:
			return nil, client.UnAuthenticated(response)
		case response.Code == 403:
			return nil, client.UnAuthorized(response)
		case response.Code == 404:
			return nil, client.NotFound(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp ListPortInterfacesResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ListFloatingIPsByPortInterface PortID를 통한 네트워크 floating ip 목록 조회
func ListFloatingIPsByPortInterface(c client.Client, req ListFloatingIPsByPortInterfaceRequest) (*ListFloatingIPsResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeNetwork)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/v2.0/floatingips?port_id=" + req.PortID + "&fixed_ip_address=" + req.FixedIP,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListFloatingIPsByPortInterface", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return nil, client.Unknown(response)
		case response.Code == 401:
			return nil, client.UnAuthenticated(response)
		case response.Code == 403:
			return nil, client.UnAuthorized(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}

	}

	var rsp ListFloatingIPsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// 인스턴스 생성 완료 여부 확인
func checkInstanceCreation(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.InstanceCreationCheckTimeOut)
	now := time.Now()

	for {
		res, err := ShowServerDetails(c, ShowServerDetailsRequest{
			ServerID: uuid,
		})
		if err != nil {
			return err
		}

		switch res.Server.Status {
		case "ACTIVE":
			logger.Infof("[checkInstanceCreation] instance(%s) creation elapsed time: %v", uuid, time.Since(now))
			return nil

		case "ERROR":
			logger.Errorf("[checkInstanceCreation] ShowServerDetails response body: %+v", res.Raw)
			return client.RemoteServerError(errors.New("error occurred during creation of instance"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.InstanceCreationCheckTimeInterval):
			continue
		}
	}
}

// CreateServer 인스턴스를 생성한다
func CreateServer(c client.Client, req CreateServerRequest) (*CreateServerResponse, error) {
	body, err := json.Marshal(&req)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"}, //하이퍼바이저 할당 2.74 이상
		},
		Method: http.MethodPost,
		URL:    endpoint + "/servers",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateServer", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return nil, client.Unknown(response)
		case response.Code == 401:
			return nil, client.Unknown(response)
		case response.Code == 403:
			return nil, client.UnAuthorized(response)
		case response.Code == 404:
			return nil, client.NotFound(response)
		case response.Code == 409:
			return nil, client.Conflict(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var ret CreateServerResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual instance deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	if err := checkInstanceCreation(c, ret.Server.UUID); err != nil {
		return &ret, err
	}

	ret.Raw = &response.Body

	return &ret, nil
}

// 보안 그룹 추가 완료 여부 확인
func checkAddSecurityGroupToServer(c client.Client, uuid, securityGroupName string) error {
	timeout := time.After(time.Second * constant.AddSecurityGroupToServerCheckTimeOut)

	for {
		res, err := ShowServerDetails(c, ShowServerDetailsRequest{
			ServerID: uuid,
		})
		if err != nil {
			return err
		}

		for _, sg := range res.Server.SecurityGroups {
			switch {
			case securityGroupName == sg.Name:
				return nil
			}
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.AddSecurityGroupToServerCheckTimeInterval):
			continue
		}
	}
}

// AddSecurityGroupToServer 인스턴스에 보안그룹을 추가한다
func AddSecurityGroupToServer(c client.Client, req AddSecurityGroupToServerRequest) error {
	body, err := json.Marshal(&req)
	if err != nil {
		return errors.Unknown(err)
	}

	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodPost,
		URL:    endpoint + "/servers/" + req.ServerID + "/action",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("AddSecurityGroupToServer", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return client.Unknown(response)
		case response.Code == 401:
			return client.Unknown(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			return client.NotFound(response)
		case response.Code == 409:
			return client.Conflict(response)
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkAddSecurityGroupToServer(c, req.ServerID, req.Action.Name)
}

// 인스턴스 삭제 완료 여부 확인
func checkInstanceDeletion(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.InstanceDeletionCheckTimeOut)

	for {
		res, err := ShowServerDetails(c, ShowServerDetailsRequest{
			ServerID: uuid,
		})
		switch {
		case errors.Equal(err, client.ErrNotFound):
			return nil
		case err != nil:
			return err
		}

		switch res.Server.Status {
		case "DELETED":
			return nil
		case "SOFT_DELETED":
			return client.RemoteServerError(errors.New("instance was not completely deleted"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.InstanceDeletionCheckTimeInterval):
			continue
		}
	}
}

// DeleteServer 인스턴스를 삭제한다
func DeleteServer(c client.Client, req DeleteServerRequest) error {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodDelete,
		URL:    endpoint + "/servers/" + req.ServerID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteServer", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 401:
			return client.Unknown(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			logger.Warnf("instance not found")
		case response.Code == 409:
			return client.Conflict(response)
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkInstanceDeletion(c, req.ServerID)
}

// 인스턴스 강제 삭제 완료 여부 확인
func checkInstanceForceDeletion(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.InstanceForceDeletionCheckTimeOut)

	for {
		res, err := ShowServerDetails(c, ShowServerDetailsRequest{
			ServerID: uuid,
		})
		switch {
		case errors.Equal(err, client.ErrNotFound):
			return nil
		case err != nil:
			return err
		}

		if res.Server.Status == "DELETED" {
			return nil
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.InstanceForceDeletionCheckTimeInterval):
			continue
		}
	}
}

// ForceDeleteServer 인스턴스를 강제로 삭제한다
func ForceDeleteServer(c client.Client, req DeleteServerRequest) error {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodPost,
		URL:    endpoint + "/servers/" + req.ServerID + "/action",
		Body:   []byte("{ \"forceDelete\": null }"),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ForceDeleteServer", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 401:
			return client.Unknown(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			logger.Warnf("instance not found")
		case response.Code == 409:
			return client.Conflict(response)
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkInstanceForceDeletion(c, req.ServerID)
}

// 인스턴스 기동 완료 여부 확인
func checkInstanceStarted(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.InstanceStartedCheckTimeOut)

	for {
		res, err := ShowServerDetails(c, ShowServerDetailsRequest{
			ServerID: uuid,
		})
		if err != nil {
			return err
		}

		switch res.Server.Status {
		case "ACTIVE":
			return nil

		case "ERROR":
			return client.RemoteServerError(errors.New("error occurred during start of instance"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.InstanceStartedCheckTimeInterval):
			continue
		}
	}
}

// StartServer 인스턴스를 기동한다
func StartServer(c client.Client, req StartServerRequest) error {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodPost,
		URL:    endpoint + "/servers/" + req.ServerID + "/action",
		Body:   []byte("{ \"os-start\": null }"),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("StartServer", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 401:
			return client.Unknown(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			return client.NotFound(response)
		case response.Code == 409:
			return client.Conflict(response)
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkInstanceStarted(c, req.ServerID)
}

// 인스턴스 중지 완료 여부 확인
func checkInstanceStopped(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.InstanceStoppedCheckTimeOut)

	for {
		res, err := ShowServerDetails(c, ShowServerDetailsRequest{
			ServerID: uuid,
		})
		if err != nil {
			return err
		}

		if res.Server.Status == "SHUTOFF" {
			return nil
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.InstanceStoppedCheckTimeInterval):
			continue
		}
	}
}

// StopServer 인스턴스를 중지한다
func StopServer(c client.Client, req StopServerRequest) error {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodPost,
		URL:    endpoint + "/servers/" + req.ServerID + "/action",
		Body:   []byte("{ \"os-stop\": null }"),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("StopServer", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 401:
			return client.Unknown(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			return client.NotFound(response)
		case response.Code == 409:
			return client.Conflict(response)
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkInstanceStopped(c, req.ServerID)
}

// RemoveSecurityGroupToServer 인스턴스에 보안그룹을 삭제한다
func RemoveSecurityGroupToServer(c client.Client, req RemoveSecurityGroupToServerRequest) error {
	body, err := json.Marshal(&req)
	if err != nil {
		return errors.Unknown(err)
	}

	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodPost,
		URL:    endpoint + "/servers/" + req.ServerID + "/action",
		Body:   body,
	}

	t := time.Now()
	response, err := request.Request()
	requestLog("RemoveSecurityGroupToServer", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return client.Unknown(response)
		case response.Code == 401:
			return client.Unknown(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			return client.NotFound(response)
		case response.Code == 409:
			return client.Conflict(response)
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkRemoveSecurityGroupToServer(c, req.ServerID, req.Action.Name)
}

// checkRemoveSecurityGroupToServer 보안 그룹 삭제 완료 여부 확인
func checkRemoveSecurityGroupToServer(c client.Client, uuid, securityGroupName string) error {
	timeout := time.After(time.Second * constant.AddSecurityGroupToServerCheckTimeOut)

	for {
		res, err := ShowServerDetails(c, ShowServerDetailsRequest{
			ServerID: uuid,
		})
		if err != nil {
			return err
		}

		var isFind bool = false
		for _, sg := range res.Server.SecurityGroups {
			if securityGroupName == sg.Name {
				isFind = true
				break
			}
		}

		if isFind == false {
			return nil
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.AddSecurityGroupToServerCheckTimeInterval):
			continue
		}
	}
}
