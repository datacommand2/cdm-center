package train

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"net/http"
	"time"
)

// VolumeTypeResult 볼륨 타입 Result 구조체
type VolumeTypeResult struct {
	UUID        string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	ExtraSpec   map[string]interface{} `json:"extra_specs,omitempty"`
}

// ShowVolumeTypeDetailRequest 볼륨 타입 조회 request 구조체
type ShowVolumeTypeDetailRequest struct {
	VolumeTypeID string
}

// ListAllVolumeTypesResponse 볼륨 타입 목록 조회 response 구조체
type ListAllVolumeTypesResponse struct {
	VolumeTypes []VolumeTypeResult `json:"volume_types,omitempty"`
}

// ShowVolumeTypeDetailResponse 볼륨 타입 조회 response 구조체
type ShowVolumeTypeDetailResponse struct {
	VolumeType VolumeTypeResult `json:"volume_type,omitempty"`
	Raw        string
}

// Capabilities 볼륨 스토리지 풀 capabilities 구조체
type Capabilities struct {
	VolumeBackendName      string      `json:"volume_backend_name,omitempty"`
	TotalCapacityGigaBytes interface{} `json:"total_capacity_gb,omitempty"`
	FreeCapacityGigaBytes  interface{} `json:"free_capacity_gb,omitempty"`
}

// BackendStoragePoolResult 볼륨 스토리지 풀 result
type BackendStoragePoolResult struct {
	Capabilities Capabilities `json:"capabilities,omitempty"`
}

// ListAllBackendStoragePoolsResponse 볼륨 스토리지 풀 목록 조회 response 구조체
type ListAllBackendStoragePoolsResponse struct {
	StoragePools []BackendStoragePoolResult `json:"pools,omitempty"`
}

// ListAllVolumeTypes 볼륨 타입 목록 조회
func ListAllVolumeTypes(c client.Client) (*ListAllVolumeTypesResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeVolumeV3)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/types",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListAllVolumeTypes", t, c, request, response, err)
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

	var rsp ListAllVolumeTypesResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ShowVolumeTypeDetail 볼륨 타입 상세 조회
func ShowVolumeTypeDetail(c client.Client, req ShowVolumeTypeDetailRequest) (*ShowVolumeTypeDetailResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeVolumeV3)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/types/" + req.VolumeTypeID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowVolumeTypeDetail", t, c, request, response, err)
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

	var rsp ShowVolumeTypeDetailResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}

// ShowDefaultVolumeType 기본 volume type 조회
func ShowDefaultVolumeType(c client.Client) (*ShowVolumeTypeDetailResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeVolumeV3)
	if err != nil {
		return nil, err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/types/default",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowDefaultVolumeType", t, c, request, response, err)
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

	var rsp ShowVolumeTypeDetailResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}

// ListAllBackendStoragePools 볼륨 스토리지 풀 목록 조회
func ListAllBackendStoragePools(c client.Client) (*ListAllBackendStoragePoolsResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeVolumeV3)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/scheduler-stats/get_pools?detail=true",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListAllBackendStoragePools", t, c, request, response, err)
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

	var rsp ListAllBackendStoragePoolsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}
