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

// CreateFloatingIPResult 부동 IP 생성
type CreateFloatingIPResult struct {
	UUID              string `json:"id,omitempty"`
	FloatingNetworkID string `json:"floating_network_id,omitempty"`
	ProjectUUID       string `json:"project_id,omitempty"`

	// FloatingIPAddress 지정: 고정 IP 할당, 미지정: dynamic IP 할당
	FloatingIPAddress string  `json:"floating_ip_address,omitempty"`
	Description       *string `json:"description,omitempty"`

	Status string `json:"status,omitempty"`
}

// UpdateFloatingIPResult 부동 IP 할당
type UpdateFloatingIPResult struct {
	PortID         string  `json:"port_id"`
	FixedIPAddress string  `json:"fixed_ip_address,omitempty"`
	Description    *string `json:"description,omitempty"`
}

// CreateFloatingIPRequest 부동 IP 생성 요청
type CreateFloatingIPRequest struct {
	FloatingIP CreateFloatingIPResult `json:"floatingip"`
}

// CreateFloatingIPResponse 부동 IP 생성 응답
type CreateFloatingIPResponse struct {
	FloatingIP CreateFloatingIPResult `json:"floatingip"`
	Raw        string                 `json:"-"`
}

// UpdateFloatingIPRequest 인스턴스에 부동 IP 할당 요청
type UpdateFloatingIPRequest struct {
	FloatingIPID string                 `json:"-"`
	FloatingIP   UpdateFloatingIPResult `json:"floatingip"`
}

// UpdateFloatingIPResponse 인스턴스에 부동 IP 할당 응답
type UpdateFloatingIPResponse struct {
	FloatingIP FloatingIPResult `json:"floatingip"`
}

// FloatingIPResult 네트워크 floating ip Result
type FloatingIPResult struct {
	ProjectID   string  `json:"project_id,omitempty"`
	NetworkID   string  `json:"floating_network_id,omitempty"`
	UUID        string  `json:"id,omitempty"`
	Description *string `json:"description,omitempty"`
	IPAddress   string  `json:"floating_ip_address,omitempty"`
	Status      string  `json:"status,omitempty"`
}

// ShowFloatingIPDetailsRequest 네트워크 floating ip 조회 request
type ShowFloatingIPDetailsRequest struct {
	FloatingIPID string
}

// ListFloatingIPsRequest 네트워크 floating ip 목록 조회 request
type ListFloatingIPsRequest struct {
	NetworkID string
}

// ListFloatingIPsResponse 네트워크 floating ip 목록 조회 response
type ListFloatingIPsResponse struct {
	FloatingIPs []FloatingIPResult `json:"floatingips,omitempty"`
}

// ShowFloatingIPDetailsResponse 네트워크 floating ip 조회 response
type ShowFloatingIPDetailsResponse struct {
	FloatingIP FloatingIPResult `json:"floatingip,omitempty"`
	Raw        string
}

// CreateFloatingIP 부동 IP 생성
func CreateFloatingIP(c client.Client, in CreateFloatingIPRequest) (*CreateFloatingIPResponse, error) {
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
		URL:    endpoint + "/v2.0/floatingips",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateFloatingIP", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			// floatingip.status 가 지정된 경우
			return nil, client.Unknown(response)
		case response.Code == 401:
			return nil, client.Unknown(response)
		case response.Code == 403:
			return nil, client.UnAuthorized(response)
		case response.Code == 404:
			return nil, client.NotFound(response)
		case response.Code == 409:
			// floatingip.floating_ip_address 가 floating_network_id 에서 이미 사용되는 경우
			return nil, client.Conflict(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	// convert
	var ret CreateFloatingIPResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual floating ip deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}
	ret.Raw = response.Body

	return &ret, nil
}

// ListFloatingIPs 네트워크 floating ip 목록 조회
func ListFloatingIPs(c client.Client, req ListFloatingIPsRequest) (*ListFloatingIPsResponse, error) {
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
		URL:    endpoint + "/v2.0/floatingips?floating_network_id=" + req.NetworkID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListFloatingIPs", t, c, request, response, err)
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

// ShowFloatingIPDetails floating ip 상세 조회
func ShowFloatingIPDetails(c client.Client, req ShowFloatingIPDetailsRequest) (*ShowFloatingIPDetailsResponse, error) {
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
		URL:    endpoint + "/v2.0/floatingips/" + req.FloatingIPID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowFloatingIPDetails", t, c, request, response, err)
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

	var rsp ShowFloatingIPDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}

// DeleteFloatingIPRequest 부동 IP 삭제 요청
type DeleteFloatingIPRequest struct {
	FloatingUUID string
}

// DeleteFloatingIP 부동 IP 삭제
func DeleteFloatingIP(c client.Client, in DeleteFloatingIPRequest) error {
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
		URL:    fmt.Sprintf("%s/v2.0/floatingips/%s", endpoint, in.FloatingUUID),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteFloatingIP", t, c, request, response, err)
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
			logger.Warnf("Could not delete floating ip. Cause: not found floating ip")
		case response.Code == 412:
			// instance 에 연결된 floating 을 제거하는 경우 여기로 안떨어짐
			// 문서상 Update/Delete the target resource has been denied due to the mismatch of revision number.
			return client.Unknown(response)
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return nil
}

// UpdateFloatingIP 인스턴스에 부동 IP 할당
func UpdateFloatingIP(c client.Client, req UpdateFloatingIPRequest) (*UpdateFloatingIPResponse, error) {
	// convert
	body, err := json.Marshal(&req)
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
		Method: http.MethodPut,
		URL:    endpoint + "/v2.0/floatingips/" + req.FloatingIPID,
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("UpdateFloatingIP", t, c, request, response, err)
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
		case response.Code == 412:
			return nil, client.Unknown(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	// convert
	var ret UpdateFloatingIPResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual floating ip deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}
