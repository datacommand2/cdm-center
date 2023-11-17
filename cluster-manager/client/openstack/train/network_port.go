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

// FixedIP 포트 고정 IP
type FixedIP struct {
	SubnetUUID string `json:"subnet_id,omitempty"`
	IPAddress  string `json:"ip_address,omitempty"`
}

// PortResult 포트
type PortResult struct {
	UUID        string    `json:"id,omitempty"`
	NetworkUUID string    `json:"network_id,omitempty"`
	FixedIPs    []FixedIP `json:"fixed_ips,omitempty"`
	DeviceID    string    `json:"device_id,omitempty"`

	// network:router_interface
	DeviceOwner string `json:"device_owner,omitempty"`
	//TODO: ProjectID 추가로 내부 외부 구분 가능 검토
	ProjectID      string   `json:"project_id,omitempty"`
	SecurityGroups []string `json:"security_groups,omitempty"`
}

// CreatePortRequest 포트 생성 요청
type CreatePortRequest struct {
	Port PortResult `json:"port"`
}

// CreatePortResponse 포트 생성 응답
type CreatePortResponse struct {
	Port PortResult `json:"port"`
}

// ShowPortDetailsRequest 네트워크 포트 조회 Request 구조체
type ShowPortDetailsRequest struct {
	PortID string
}

// ShowPortDetailsResponse 네트워크 포트 조회 Response 구조체
type ShowPortDetailsResponse struct {
	Port PortResult `json:"port"`
	Raw  string
}

// CreatePort 네트워크 포트를 생성한다
func CreatePort(c client.Client, req CreatePortRequest) (*CreatePortResponse, error) {
	body, err := json.Marshal(&req)
	if err != nil {
		return nil, errors.Unknown(err)
	}

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
		URL:    endpoint + "/v2.0/ports",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreatePort", t, c, request, response, err)
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
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var ret CreatePortResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual port deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}

// DeletePort 네트워크 포트를 삭제한다
func DeletePort(c client.Client, req DeletePortRequest) error {
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
		URL:    endpoint + "/v2.0/ports/" + req.PortID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeletePort", t, c, request, response, err)
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
			logger.Warnf("Could not delete port. Cause: not found port")
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

// ShowPortDetails 네트워크 포트 상세 조회
func ShowPortDetails(c client.Client, req ShowPortDetailsRequest) (*ShowPortDetailsResponse, error) {
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
		URL:    endpoint + "/v2.0/ports/" + req.PortID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowPortDetails", t, c, request, response, err)
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

	var rsp ShowPortDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}
