package train

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"net/http"
	"time"
)

// ExternalGatewayInfo 라우터 외부 게이트웨이 정보
type ExternalGatewayInfo struct {
	NetworkUUID      string    `json:"network_id,omitempty"`
	ExternalFixedIPs []FixedIP `json:"external_fixed_ips,omitempty"`
}

// RouterResult 라우터 result
type RouterResult struct {
	UUID                string              `json:"id,omitempty"`
	Name                string              `json:"name,omitempty"`
	Description         *string             `json:"description,omitempty"`
	ProjectID           string              `json:"project_id,omitempty"`
	State               bool                `json:"admin_state_up,omitempty"`
	Status              string              `json:"status,omitempty"`
	Routes              []ExtraRoute        `json:"routes,omitempty"`
	ExternalGatewayInfo ExternalGatewayInfo `json:"external_gateway_info,omitempty"`
}

// ShowRouterDetailsRequest 라우터 조회 request
type ShowRouterDetailsRequest struct {
	RouterID string
}

// ListPortsRequest 포트 목록 조회 request
type ListPortsRequest struct {
	DeviceOwner string
	DeviceID    string
}

// ListRoutersResponse 라우터 목록 조회 response
type ListRoutersResponse struct {
	Routers []RouterResult `json:"routers,omitempty"`
}

// ShowRouterDetailsResponse 라우터 조회 response
type ShowRouterDetailsResponse struct {
	Router RouterResult `json:"router,omitempty"`
	Raw    string
}

// ListPortsResponse 포트 목록 조회 response
type ListPortsResponse struct {
	Ports []PortResult `json:"ports,omitempty"`
}

// CreateRouterRequest 라우터 생성 요청
type CreateRouterRequest struct {
	Router RouterResult `json:"router"`
}

// CreateRouterResponse 라우터 생성 요청
type CreateRouterResponse struct {
	Router RouterResult `json:"router"`
	Raw    string       `json:"-"`
}

// CreateRouter 라우터 생성
func CreateRouter(c client.Client, in CreateRouterRequest) (*CreateRouterResponse, error) {
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
		URL:    endpoint + "/v2.0/routers",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateRouter", t, c, request, response, err)
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
		case response.Code == 409:
			// Quota exceeded for resources
			return nil, client.Conflict(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	// convert
	var ret CreateRouterResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual router deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}
	ret.Raw = response.Body

	return &ret, nil
}

// DeleteRouterRequest 라우터 삭제 요청
type DeleteRouterRequest struct {
	RouterUUID string
}

// DeleteRouter 라우터 삭제
func DeleteRouter(c client.Client, in DeleteRouterRequest) error {
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
		URL:    endpoint + "/v2.0/routers/" + in.RouterUUID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteRouter", t, c, request, response, err)
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
			logger.Warnf("Could not delete router. Cause: not found router")
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

// ListRouters 라우터 목록 조회
func ListRouters(c client.Client) (*ListRoutersResponse, error) {
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
		URL:    endpoint + "/v2.0/routers",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListRouters", t, c, request, response, err)
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

	var rsp ListRoutersResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ShowRouterDetails 라우터 상세 조회
func ShowRouterDetails(c client.Client, req ShowRouterDetailsRequest) (*ShowRouterDetailsResponse, error) {
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
		URL:    endpoint + "/v2.0/routers/" + req.RouterID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowRouterDetails", t, c, request, response, err)
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

	var rsp ShowRouterDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}

// ListPorts 포트 목록 조회
func ListPorts(c client.Client, req ListPortsRequest) (*ListPortsResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeNetwork)
	if err != nil {
		return nil, err
	}
	var reqURL string
	if req.DeviceOwner == "" {
		reqURL = endpoint + "/v2.0/ports?device_id=" + req.DeviceID
	} else {
		reqURL = endpoint + "/v2.0/ports?device_id=" + req.DeviceID + "&device_owner=" + req.DeviceOwner
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodGet,
		//URL:    endpoint + "/v2.0/ports?device_id=" + req.DeviceID + "&device_owner=" + req.DeviceOwner,
		//TODO: QA 요청으로 DeviceOwner 임시 삭제
		//URL: endpoint + "/v2.0/ports?device_id=" + req.DeviceID,
		URL: reqURL,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListPorts", t, c, request, response, err)
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

	var rsp ListPortsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}
