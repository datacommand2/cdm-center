package train

import (
	"encoding/json"
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack/train/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"net/http"
	"strconv"
	"time"
)

// VolumeAttachment 볼륨 attachment 정보 구조체
type VolumeAttachment struct {
	UUID string `json:"attachment_id"`
}

// VolumeResult 볼륨 Result 구조체
type VolumeResult struct {
	ProjectID      string                 `json:"os-vol-tenant-attr:tenant_id,omitempty"`
	VolumeTypeName string                 `json:"volume_type,omitempty"`
	UUID           string                 `json:"id,omitempty"`
	Name           string                 `json:"name,omitempty"`
	Description    *string                `json:"description,omitempty"`
	SizeGibiBytes  uint64                 `json:"size,omitempty"`
	Multiattach    bool                   `json:"multiattach,omitempty"`
	Bootable       string                 `json:"bootable,omitempty"`
	Status         string                 `json:"status,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Attachments    []VolumeAttachment     `json:"attachments,omitempty"`
}

// ShowVolumeDetailsRequest 볼륨 상세 조회 request
type ShowVolumeDetailsRequest struct {
	VolumeID string
}

// ListAccessibleVolumesResponse 볼륨 목록 조회 response
type ListAccessibleVolumesResponse struct {
	Volumes []VolumeResult `json:"volumes,omitempty"`
}

// ShowVolumeDetailsResponse 볼륨 상세 조회 response
type ShowVolumeDetailsResponse struct {
	Volume VolumeResult `json:"volume,omitempty"`
	Raw    string
}

// CreateVolumeRequest 볼륨의 CreateVolumeRequest 구조체
type CreateVolumeRequest struct {
	Volume VolumeResult `json:"volume"`
}

// CreateVolumeResponse 볼륨의 CreateVolumeResponse 구조체
type CreateVolumeResponse struct {
	Volume VolumeResult `json:"volume"`
	Raw    *string
}

// UpdatesVolumeReadOnlyAccessModeFlagResult 볼륨의 UpdatesVolumeReadOnlyAccessModeFlagResult 구조체
type UpdatesVolumeReadOnlyAccessModeFlagResult struct {
	ReadOnly bool `json:"readonly,omitempty"`
}

// UpdatesVolumeReadOnlyAccessModeFlagRequest 볼륨의 UpdatesVolumeReadOnlyAccessModeFlagRequest 구조체
type UpdatesVolumeReadOnlyAccessModeFlagRequest struct {
	VolumeID       string                                    `json:"-"`
	VolumeReadOnly UpdatesVolumeReadOnlyAccessModeFlagResult `json:"os-update_readonly_flag"`
}

// UpdateVolumeBootableStatusResult 볼륨의 UpdateVolumeBootableStatusResult 구조체
type UpdateVolumeBootableStatusResult struct {
	Bootable bool `json:"bootable,omitempty"`
}

// UpdateVolumeBootableStatusRequest 볼륨의 UpdateVolumeBootableStatusRequest 구조체
type UpdateVolumeBootableStatusRequest struct {
	VolumeID       string                           `json:"-"`
	VolumeBootable UpdateVolumeBootableStatusResult `json:"os-set_bootable"`
}

// DeleteVolumeRequest 볼륨의 DeleteVolumeRequest 구조체
type DeleteVolumeRequest struct {
	VolumeID string
	Cascade  bool
}

// VolumeReference 이미 존재하는 볼륨에 대한 VolumeReference 구조체
type VolumeReference struct {
	SourceName string `json:"source-name"`
}

// ManageVolumeRequest 볼륨의 ManageVolumeRequest 구조체
type ManageVolumeRequest struct {
	Name           string          `json:"name,omitempty"`
	Description    *string         `json:"description,omitempty"`
	VolumeTypeUUID string          `json:"volume_type,omitempty"`
	Bootable       bool            `json:"bootable,omitempty"`
	Host           string          `json:"host"`
	Ref            VolumeReference `json:"ref"`
}

// ManageExistingVolumeRequest 볼륨 import Request 구조체
type ManageExistingVolumeRequest struct {
	Volume ManageVolumeRequest `json:"volume"`
}

// ManageExistingVolumeResponse 볼륨 import Response 구조체
type ManageExistingVolumeResponse struct {
	Volume VolumeResult `json:"volume"`
	Raw    *string
}

// ServicesResult 서비스의 ServicesResult 구조체
type ServicesResult struct {
	Binary    string `json:"binary,omitempty"`
	Host      string `json:"host,omitempty"`
	State     string `json:"state,omitempty"`
	Status    string `json:"status,omitempty"`
	Zone      string `json:"zone,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ListAllCinderServicesResponse 서비스 조회 Response 구조체
type ListAllCinderServicesResponse struct {
	Services []ServicesResult `json:"services"`
}

// ListAccessibleVolumes 볼륨 목록 조회
func ListAccessibleVolumes(c client.Client) (*ListAccessibleVolumesResponse, error) {
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
		URL:    endpoint + "/volumes?all_tenants=true", // admin only
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListAccessibleVolumes", t, c, request, response, err)
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

	var rsp ListAccessibleVolumesResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ShowVolumeDetails 볼륨 상세 조회
func ShowVolumeDetails(c client.Client, req ShowVolumeDetailsRequest) (*ShowVolumeDetailsResponse, error) {
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
		URL:    endpoint + "/volumes/" + req.VolumeID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowVolumeDetails", t, c, request, response, err)
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

	var rsp ShowVolumeDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}

// 볼륨 readonly flag 설정 완료 여부 확인
func checkVolumeReadOnlyFlag(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.VolumeReadOnlyCheckTimeOut)
	for {
		res, err := ShowVolumeDetails(c, ShowVolumeDetailsRequest{
			VolumeID: uuid,
		})
		if err != nil {
			return err
		}
		if v, ok := res.Volume.Metadata["readonly"]; ok {
			b, err := strconv.ParseBool(v.(string))
			if err != nil {
				return err
			} else if !b {
				return client.RemoteServerError(errors.New("failed to set volume to read only"))
			}
			return nil
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.VolumeReadOnlyCheckTimeInterval):
			continue
		}
	}
}

// 볼륨 생성 완료 여부 확인
func checkVolumeCreation(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.VolumeCreationCheckTimeOut)
	now := time.Now()

	for {
		res, err := ShowVolumeDetails(c, ShowVolumeDetailsRequest{
			VolumeID: uuid,
		})
		if err != nil {
			return err
		}

		if res != nil {
			switch res.Volume.Status {
			case "available":
				logger.Infof("[checkVolumeCreation] volume(%s) creation elapsed time: %v", uuid, time.Since(now))
				return nil

			case "error":
				return client.RemoteServerError(errors.New("error occurred during creation of volume"))

			default:
				logger.Infof("[checkVolumeCreation] ShowVolumeDetails res.Volume.Status: %s", res.Volume.Status)
			}
		} else {
			logger.Warn("[checkVolumeCreation] ShowVolumeDetails res.Volume.Status is nil")
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.VolumeCreationCheckTimeInterval):
			continue
		}
	}
}

// CreateVolume 볼륨을 생성한다
func CreateVolume(c client.Client, req CreateVolumeRequest) (*CreateVolumeResponse, error) {
	body, err := json.Marshal(&req)
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
		URL:    endpoint + "/volumes",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateVolume", t, c, request, response, err)
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

	var ret CreateVolumeResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual volume deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	if err := checkVolumeCreation(c, ret.Volume.UUID); err != nil {
		return &ret, err
	}

	ret.Raw = &response.Body

	return &ret, nil
}

// UpdatesVolumeReadOnlyAccessModeFlag 볼륨에 read-only flag 를 설정한다
func UpdatesVolumeReadOnlyAccessModeFlag(c client.Client, req UpdatesVolumeReadOnlyAccessModeFlagRequest) error {
	body, err := json.Marshal(&req)
	if err != nil {
		return errors.Unknown(err)
	}

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
		Method: http.MethodPost,
		URL:    endpoint + "/volumes/" + req.VolumeID + "/action",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("UpdatesVolumeReadOnlyAccessModeFlag", t, c, request, response, err)
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
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkVolumeReadOnlyFlag(c, req.VolumeID)
}

// UpdateVolumeBootableStatus 볼륨의 bootable 을 변경한다
func UpdateVolumeBootableStatus(c client.Client, req UpdateVolumeBootableStatusRequest) error {
	body, err := json.Marshal(&req)
	if err != nil {
		return errors.Unknown(err)
	}

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
		Method: http.MethodPost,
		URL:    endpoint + "/volumes/" + req.VolumeID + "/action",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("UpdateVolumeBootableStatus", t, c, request, response, err)
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
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return nil
}

// 볼륨 삭제 완료 여부 확인
func checkVolumeDeletion(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.VolumeDeletionCheckTimeOut)

	for {
		res, err := ShowVolumeDetails(c, ShowVolumeDetailsRequest{
			VolumeID: uuid,
		})
		switch {
		case errors.Equal(err, client.ErrNotFound):
			return nil
		case err != nil:
			return err
		}

		if res.Volume.Status == "error_deleting" {
			return client.RemoteServerError(errors.New("error occurred during deletion of volume"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.VolumeDeletionCheckTimeInterval):
			continue
		}
	}
}

// DeleteVolume 볼륨을 삭제한다
func DeleteVolume(c client.Client, req DeleteVolumeRequest) error {
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
		URL:    fmt.Sprintf("%s/volumes/%s?cascade=%v", endpoint, req.VolumeID, req.Cascade),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteVolume", t, c, request, response, err)
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
			logger.Warnf("Could not delete volume. Cause: not found volume")
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkVolumeDeletion(c, req.VolumeID)
}

// UnmanageVolumeRequest 볼륨 unmanage 요청
type UnmanageVolumeRequest struct {
	VolumeUUID string
}

// UnmanageVolume 볼륨 unmanage
func UnmanageVolume(c client.Client, req UnmanageVolumeRequest) error {
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
		Method: http.MethodPost,
		URL:    fmt.Sprintf("%s/volumes/%s/action", endpoint, req.VolumeUUID),
		Body:   []byte("{\"os-unmanage\":{}}"),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("UnmanageVolume", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			logger.Warnf("Could not unmanage volume. Cause: not found volume")
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkVolumeUnmanaging(c, req.VolumeUUID)
}

func checkVolumeUnmanaging(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.VolumeUnmanagingCheckTimeOut)

	for {
		res, err := ShowVolumeDetails(c, ShowVolumeDetailsRequest{
			VolumeID: uuid,
		})
		switch {
		case errors.Equal(err, client.ErrNotFound):
			return nil
		case err != nil:
			return err
		}

		if res.Volume.Status != "unmanaging" {
			return client.RemoteServerError(errors.New("error occurred during unmanaging of volume"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.VolumeUnmanagingCheckTimeInterval):
			continue
		}
	}
}

// ManageExistingVolume 기존 볼륨 파일을 이용하여 볼륨을 가져온다
func ManageExistingVolume(c client.Client, req ManageExistingVolumeRequest) (*ManageExistingVolumeResponse, error) {
	body, err := json.Marshal(&req)
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
		URL:    endpoint + "/manageable_volumes",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ManageExistingVolume", t, c, request, response, err)
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

	var ret ManageExistingVolumeResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual volume deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	if err := checkVolumeCreation(c, ret.Volume.UUID); err != nil {
		return &ret, err
	}

	ret.Raw = &response.Body

	return &ret, nil
}

// ListAllCinderServices host 조회
func ListAllCinderServices(c client.Client) (*ListAllCinderServicesResponse, error) {
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
		URL:    endpoint + "/os-services",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListAllCinderServices", t, c, request, response, err)
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

	var rsp ListAllCinderServicesResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ManageableVolume volume manage 정보
type ManageableVolume struct {
	SafeToManage bool            `json:"safe_to_manage"`
	Reference    VolumeReference `json:"reference"`
	Size         uint64          `json:"size"`
}

// ListSummaryOfVolumesAvailableToManageResponse volume manage 목록 조회 응답
type ListSummaryOfVolumesAvailableToManageResponse struct {
	ManageableVolumes []ManageableVolume `json:"manageable-volumes"`
}

// ListSummaryOfVolumesAvailableToManageRequest volume manage 목록 조회 요청
type ListSummaryOfVolumesAvailableToManageRequest struct {
	Host string
}

// ListSummaryOfVolumesAvailableToManage volume manage 목록 조회
func ListSummaryOfVolumesAvailableToManage(c client.Client, req ListSummaryOfVolumesAvailableToManageRequest) (*ListSummaryOfVolumesAvailableToManageResponse, error) {
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
		URL:    endpoint + "/manageable_volumes?host=" + req.Host,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListSummaryOfVolumesAvailableToManage", t, c, request, response, err)
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

	var rsp ListSummaryOfVolumesAvailableToManageResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}
