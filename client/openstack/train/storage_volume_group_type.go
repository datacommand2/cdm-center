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

// VolumeGroupType 볼륨 그룹 타입
type VolumeGroupType struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	IsPublic    bool   `json:"is_public,omitempty"`
}

// ListVolumeGroupTypesResponse 볼륨 그룹 타입 조회 응답
type ListVolumeGroupTypesResponse struct {
	GroupTypes []VolumeGroupType `json:"group_types"`
}

// CreateVolumeGroupTypeRequest 볼륨 그룹 타입 생성 요청
type CreateVolumeGroupTypeRequest struct {
	GroupType VolumeGroupType `json:"group_type"`
}

// VolumeGroupTypeResponse 볼륨 그룹 타입 생성 응답
type VolumeGroupTypeResponse struct {
	GroupType VolumeGroupType `json:"group_type"`
}

// DeleteVolumeGroupTypeRequest 볼륨 그룹 삭제 요청
type DeleteVolumeGroupTypeRequest struct {
	VolumeGroupTypeID string
}

// ShowVolumeGroupTypeDetailsRequest 볼륨 그룹 타입 상세 조회 요청
type ShowVolumeGroupTypeDetailsRequest struct {
	VolumeGroupTypeID string
}

// CreateVolumeGroupType (create group type) 볼륨 그룹 타입 생성, 202로 떨어지나 의미 없음
func CreateVolumeGroupType(c client.Client, in CreateVolumeGroupTypeRequest) (*VolumeGroupTypeResponse, error) {
	body, err := json.Marshal(&in)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	endpoint, err := c.GetEndpoint(ServiceTypeVolumeV3)
	if err != nil {
		return nil, err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"volume 3.59"},
		},
		Method: http.MethodPost,
		URL:    endpoint + "/group_types",
		Body:   body,
	}

	logger.Infof("[CreateVolumeGroupType] method : %+v", request.Method)
	logger.Infof("[CreateVolumeGroupType] URL : %+v", request.URL)
	logger.Infof("[CreateVolumeGroupType] BODY : %+v", string(request.Body[:]))
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateVolumeGroupType", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
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

	var ret VolumeGroupTypeResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual volume group type deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}

// DeleteVolumeGroupType 볼륨 그룹 타입 삭제, 202로 떨어지나 의미 없음
func DeleteVolumeGroupType(c client.Client, in DeleteVolumeGroupTypeRequest) error {
	endpoint, err := c.GetEndpoint(ServiceTypeVolumeV3)
	if err != nil {
		return err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"volume 3.59"},
		},
		Method: http.MethodDelete,
		URL:    endpoint + "/group_types/" + in.VolumeGroupTypeID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteVolumeGroupType", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			// {"badRequest": {"message": "Target group type is still in use. Group Type ac6536a4-019e-49b2-abbf-ce82e22780df deletion is not allowed with groups present with the type.", "code": 400}}
			return client.Unknown(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			logger.Warnf("Could not delete volume group type. Cause: not found volume group type")
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return nil
}

// ListVolumeGroupTypes 볼륨 그룹 타입 조회
func ListVolumeGroupTypes(c client.Client) (*ListVolumeGroupTypesResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeVolumeV3)
	if err != nil {
		return nil, err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"volume 3.59"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/group_types",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListVolumeGroupTypes", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 403:
			return nil, client.UnAuthorized(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var ret ListVolumeGroupTypesResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}

// ShowVolumeGroupTypeDetails 볼륨 그룹 타입 상세 조회
func ShowVolumeGroupTypeDetails(c client.Client, in ShowVolumeGroupTypeDetailsRequest) (*VolumeGroupTypeResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeVolumeV3)
	if err != nil {
		return nil, err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"volume 3.59"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/group_types/" + in.VolumeGroupTypeID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowVolumeGroupTypeDetails", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
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

	var ret VolumeGroupTypeResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}
