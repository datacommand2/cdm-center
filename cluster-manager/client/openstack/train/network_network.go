package train

import (
	"encoding/json"
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"net/http"
	"time"
)

// CreateNetworkResult 네트워크
type CreateNetworkResult struct {
	UUID string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`

	// true 만 해야됨, 왜냐면 instance 기준이니, network 가 disable 이면 안됨
	AdminStateUp bool   `json:"admin_state_up,omitempty"`
	ProjectUUID  string `json:"project_id,omitempty"`

	// False 만 해야됨, 외부 네트워크는 생성하지 않고 매핑만 함
	RouterExternal bool    `json:"router:external,omitempty"`
	Description    *string `json:"description,omitempty"`

	// FIXME: recovery openstack cluster 에 설정된 networking backend 를 몰라서 입력값이 유효한지 알 수 없음.
	//  recovery openstack cluster 의 default 를 사용
	Type   string `json:"provider:network_type,omitempty"`
	Status string `json:"status,omitempty"`
}

// CreateNetworkRequest 네트워크 생성 요청
type CreateNetworkRequest struct {
	Network CreateNetworkResult `json:"network"`
}

// CreateNetworkResponse 네트워크 생성 응답
type CreateNetworkResponse struct {
	Network CreateNetworkResult `json:"network"`
	Raw     string
}

// DeleteNetworkRequest 네트워크 삭제 요청
type DeleteNetworkRequest struct {
	NetworkUUID string
}

// NetworkResult 네트워크 Result
type NetworkResult struct {
	UUID         string   `json:"id,omitempty"`
	Name         string   `json:"name,omitempty"`
	Description  *string  `json:"description,omitempty"`
	ProjectID    string   `json:"project_id,omitempty"`
	TypeCode     string   `json:"provider:network_type,omitempty"`
	ExternalFlag bool     `json:"router:external,omitempty"`
	State        bool     `json:"admin_state_up,omitempty"`
	Status       string   `json:"status,omitempty"`
	Subnets      []string `json:"subnets,omitempty"`
}

// ShowNetworkDetailsRequest 네트워크 조회 request
type ShowNetworkDetailsRequest struct {
	NetworkID string
}

// ListNetworksResponse 네트워크 목록 조회 response
type ListNetworksResponse struct {
	Networks []NetworkResult `json:"networks,omitempty"`
}

// ShowNetworkDetailsResponse 네트워크 조회 response
type ShowNetworkDetailsResponse struct {
	Network NetworkResult `json:"network,omitempty"`
	Raw     string
}

// CreateNetwork 네트워크 생성
func CreateNetwork(c client.Client, in CreateNetworkRequest) (*CreateNetworkResponse, error) {
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
		URL:    endpoint + "/v2.0/networks",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateNetwork", t, c, request, response, err)
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
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	// convert
	var ret CreateNetworkResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual network deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}
	ret.Raw = response.Body

	return &ret, nil
}

// DeleteNetwork 네트워크 삭제
func DeleteNetwork(c client.Client, in DeleteNetworkRequest) error {
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
		URL:    fmt.Sprintf("%s/v2.0/networks/%s", endpoint, in.NetworkUUID),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteNetwork", t, c, request, response, err)
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
			logger.Warnf("Could not delete network. Cause: not found network")
		case response.Code == 409:
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

// ListNetworks 네트워크 목록 조회
func ListNetworks(c client.Client) (*ListNetworksResponse, error) {
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
		URL:    endpoint + "/v2.0/networks",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListNetworks", t, c, request, response, err)
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

	var rsp ListNetworksResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ShowNetworkDetails 네트워크 상세 조회
func ShowNetworkDetails(c client.Client, req ShowNetworkDetailsRequest) (*ShowNetworkDetailsResponse, error) {
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
		URL:    endpoint + "/v2.0/networks/" + req.NetworkID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowNetworkDetails", t, c, request, response, err)
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

	var rsp ShowNetworkDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}
