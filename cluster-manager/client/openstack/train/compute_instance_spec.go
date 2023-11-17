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

// FlavorResult flavor result 구조체
type FlavorResult struct {
	UUID                    string                 `json:"id,omitempty"`
	Name                    string                 `json:"name,omitempty"`
	Description             *string                `json:"description,omitempty"`
	VcpuTotalCnt            uint64                 `json:"vcpus,omitempty"`
	MemTotalMibiBytes       uint64                 `json:"ram,omitempty"`
	DiskTotalGibiBytes      uint64                 `json:"disk"`
	SwapTotalMibiBytes      uint64                 `json:"swap,omitempty"`
	EphemeralTotalGibiBytes uint64                 `json:"OS-FLV-EXT-DATA:ephemeral,omitempty"`
	ExtraSpec               map[string]interface{} `json:"extra_specs,omitempty"`
}

// ShowFlavorDetailsRequest flavor 상세 조회 request
type ShowFlavorDetailsRequest struct {
	FlavorID string
}

// ShowFlavorDetailsResponse flavor 상세 조회 response
type ShowFlavorDetailsResponse struct {
	Flavor FlavorResult `json:"flavor,omitempty"`
	Raw    string
}

// CreateFlavorRequest flavor 생성 request 구조체
type CreateFlavorRequest struct {
	Flavor FlavorResult `json:"flavor"`
}

// CreateFlavorResponse flavor 생성 response 구조체
type CreateFlavorResponse struct {
	Flavor FlavorResult `json:"flavor"`
}

// DeleteFlavorRequest flavor 삭제 request 구조체
type DeleteFlavorRequest struct {
	FlavorID string
}

// ShowFlavorDetails flavor 상세 조회
func ShowFlavorDetails(c client.Client, req ShowFlavorDetailsRequest) (*ShowFlavorDetailsResponse, error) {
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
		URL:    endpoint + "/flavors/" + req.FlavorID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowFlavorDetails", t, c, request, response, err)
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

	var rsp ShowFlavorDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}

// CreateFlavor Flavor 를 생성한다
func CreateFlavor(c client.Client, req CreateFlavorRequest) (*CreateFlavorResponse, error) {
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
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodPost,
		URL:    endpoint + "/flavors",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateFlavor", t, c, request, response, err)
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
			return nil, client.Conflict(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var ret CreateFlavorResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual flavor deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}

// Flavor 삭제 완료 여부 확인
func checkFlavorDeletion(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.FlavorDeletionCheckTimeOut)

	for {
		_, err := ShowFlavorDetails(c, ShowFlavorDetailsRequest{
			FlavorID: uuid,
		})
		switch {
		case errors.Equal(err, client.ErrNotFound):
			return nil
		case err != nil:
			return err
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.FlavorDeletionCheckTimeInterval):
			continue
		}
	}
}

// DeleteFlavor Flavor 를 삭제한다
func DeleteFlavor(c client.Client, req DeleteFlavorRequest) error {
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
		URL:    endpoint + "/flavors/" + req.FlavorID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteFlavor", t, c, request, response, err)
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
			logger.Warnf("spec not found")
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	if err = checkFlavorDeletion(c, req.FlavorID); err != nil {
		return err
	}

	return nil
}
