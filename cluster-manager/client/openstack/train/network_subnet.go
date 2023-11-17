package train

import (
	"encoding/json"
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"net"
	"net/http"
	"time"
)

// SubnetAllocationPool 서브넷 풀
type SubnetAllocationPool struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

// CreateSubnetResult 서브넷
type CreateSubnetResult struct {
	NetworkUUID string `json:"network_id,omitempty"`
	IPVersion   uint8  `json:"ip_version,omitempty"`
	CIDR        string `json:"cidr,omitempty"`
	UUID        string `json:"id,omitempty"`

	ProjectID       string                 `json:"project_id,omitempty"`
	Name            string                 `json:"name,omitempty"`
	EnableDHCP      bool                   `json:"enable_dhcp,omitempty"`
	DNSNameservers  []string               `json:"dns_nameservers,omitempty"`
	AllocationPools []SubnetAllocationPool `json:"allocation_pools,omitempty"`
	GatewayIP       *string                `json:"gateway_ip,omitempty"`
	Description     *string                `json:"description,omitempty"`
	IPv6AddressMode *string                `json:"ipv6_address_mode,omitempty"`
	IPv6RAMode      *string                `json:"ipv6_ra_mode,omitempty"`
}

// CreateSubnetRequest 서브넷 생성 요청
type CreateSubnetRequest struct {
	Subnet CreateSubnetResult `json:"subnet"`
}

// CreateSubnetResponse 서브넷 생성 응답
type CreateSubnetResponse struct {
	Subnet CreateSubnetResult `json:"subnet"`
	Raw    string
}

// DeleteSubnetRequest 서브넷 삭제 요청
type DeleteSubnetRequest struct {
	SubnetUUID string
}

// AllocationPool 네트워크 서브넷 allocation pool 구조체
type AllocationPool struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

// SubnetResult 네트워크 서브넷 Result
type SubnetResult struct {
	NetworkID           string           `json:"network_id,omitempty"`
	UUID                string           `json:"id,omitempty"`
	Name                string           `json:"name,omitempty"`
	Description         *string          `json:"description,omitempty"`
	NetworkCidr         string           `json:"cidr,omitempty"`
	DHCPEnabled         bool             `json:"enable_dhcp,omitempty"`
	GatewayIPAddress    *string          `json:"gateway_ip,omitempty"`
	Ipv6AddressModeCode *string          `json:"ipv6_address_mode,omitempty"`
	Ipv6RaModeCode      *string          `json:"ipv6_ra_mode,omitempty"`
	Nameservers         []string         `json:"dns_nameservers,omitempty"`
	AllocationPools     []AllocationPool `json:"allocation_pools,omitempty"`
}

// ShowSubnetDetailsRequest 네트워크 서브넷 상세 조회 request
type ShowSubnetDetailsRequest struct {
	SubnetID string
}

// ShowSubnetDetailsResponse 네트워크 서브넷 상세 조회 response
type ShowSubnetDetailsResponse struct {
	Subnet SubnetResult `json:"subnet,omitempty"`
	Raw    string
}

// IPVersion CIDR 로 IP version 반환 함수
func IPVersion(CIDR string) (uint8, error) {
	ip, _, err := net.ParseCIDR(CIDR)
	if err != nil {
		return 0, errors.Unknown(err)
	}

	if ip.To4() != nil {
		return 4, nil
	}

	return 6, nil
}

// CreateSubnet 서브넷 생성
func CreateSubnet(c client.Client, in CreateSubnetRequest) (*CreateSubnetResponse, error) {
	// convert
	body, err := json.Marshal(&in)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	// request
	endpoint, err := c.GetEndpoint(ServiceTypeNetwork)
	if err != nil {
		return nil, err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodPost,
		URL:    endpoint + "/v2.0/subnets",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateSubnet", t, c, request, response, err)
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

	// convert
	var ret CreateSubnetResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		// 로그 처리 안함
		// 에러 발생시, rollback 을 통해 network 삭제시, 같이 삭제됨
		return nil, errors.Unknown(err)
	}
	ret.Raw = response.Body

	return &ret, nil
}

// DeleteSubnet 서브넷 삭제
func DeleteSubnet(c client.Client, in DeleteSubnetRequest) error {
	// request
	endpoint, err := c.GetEndpoint(ServiceTypeNetwork)
	if err != nil {
		return err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("%s/v2.0/subnets/%s", endpoint, in.SubnetUUID),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteSubnet", t, c, request, response, err)
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
			logger.Warnf("Could not delete subnet. Cause: not found subnet")
		case response.Code == 409:
			// 문서에는 없으나, 서브넷에 할당된 포트가 사용 중일 때 발생
			return client.Conflict(response)
		case response.Code == 412:
			return client.Unknown(response)
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return nil
}

// ShowSubnetDetails 네트워크 서브넷 상세 조회
func ShowSubnetDetails(c client.Client, req ShowSubnetDetailsRequest) (*ShowSubnetDetailsResponse, error) {
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
		URL:    endpoint + "/v2.0/subnets/" + req.SubnetID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowSubnetDetails", t, c, request, response, err)
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

	var rsp ShowSubnetDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}
