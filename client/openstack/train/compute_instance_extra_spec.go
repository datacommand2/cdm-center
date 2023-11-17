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

// ShowFlavorExtraSpecRequest flavor extra spec 상세 조회 request
type ShowFlavorExtraSpecRequest struct {
	FlavorID     string
	ExtraSpecKey string
}

// ShowFlavorExtraSpecResponse flavor extra spec 상세 조회 response
type ShowFlavorExtraSpecResponse struct {
	Key   string
	Value string
}

// CreateExtraSpecsRequest flavor Extra spec 생성 request
type CreateExtraSpecsRequest struct {
	FlavorUUID string            `json:"-"`
	ExtraSpecs map[string]string `json:"extra_specs"`
}

// CreateExtraSpecsResponse flavor Extra spec 생성 response
type CreateExtraSpecsResponse struct {
	ExtraSpecs map[string]string `json:"extra_specs"`
}

// ShowFlavorExtraSpec flavor extra spec 상세 조회
func ShowFlavorExtraSpec(c client.Client, req ShowFlavorExtraSpecRequest) (*ShowFlavorExtraSpecResponse, error) {
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
		URL:    endpoint + "/flavors/" + req.FlavorID + "/os-extra_specs/" + req.ExtraSpecKey,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowFlavorExtraSpec", t, c, request, response, err)
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

	var dat map[string]string
	if err := json.Unmarshal([]byte(response.Body), &dat); err != nil {
		return nil, errors.Unknown(err)
	}

	var rsp ShowFlavorExtraSpecResponse

	for k, v := range dat {
		rsp.Key = k
		rsp.Value = v
	}

	return &rsp, nil
}

// CreateExtraSpecsForFlavor 인스턴스에 extra 스팩을 추가한다
func CreateExtraSpecsForFlavor(c client.Client, req CreateExtraSpecsRequest) (*CreateExtraSpecsResponse, error) {
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
		URL:    endpoint + "/flavors/" + req.FlavorUUID + "/os-extra_specs",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateExtraSpecsForFlavor", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
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

	var ret CreateExtraSpecsResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual flavor extra spec deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}
