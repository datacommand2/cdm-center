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
	"time"
)

// VolumeSnapshotResult 볼륨 스냅샷 Result 구조체
type VolumeSnapshotResult struct {
	UUID            string  `json:"id,omitempty"`
	Name            string  `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	SizeGibiBytes   uint64  `json:"size,omitempty"`
	Status          string  `json:"status,omitempty"`
	CreatedAt       string  `json:"created_at,omitempty"`
	VolumeID        string  `json:"volume_id,omitempty"`
	GroupSnapshotID *string `json:"group_snapshot_id,omitempty"`
}

// CreateVolumeSnapshotRequest 볼륨의 CreateVolumeSnapshotRequest 구조체
type CreateVolumeSnapshotRequest struct {
	VolumeSnapshot VolumeSnapshotResult `json:"snapshot"`
}

// CreateVolumeSnapshotResponse 볼륨의 CreateVolumeSnapshotResponse 구조체
type CreateVolumeSnapshotResponse struct {
	VolumeSnapshot VolumeSnapshotResult `json:"snapshot"`
	Raw            *string
}

// ShowVolumeSnapshotDetailsRequest 볼륨 스냅샷 조회 request 구조체
type ShowVolumeSnapshotDetailsRequest struct {
	SnapshotID string
}

// ListAccessibleSnapshotsResponse 볼륨 스냅샷 목록 조회 response 구조체
type ListAccessibleSnapshotsResponse struct {
	VolumeSnapshots []VolumeSnapshotResult `json:"snapshots,omitempty"`
}

// ShowVolumeSnapshotDetailsResponse 볼륨 스냅샷 조회 response 구조체
type ShowVolumeSnapshotDetailsResponse struct {
	VolumeSnapshot VolumeSnapshotResult `json:"snapshot,omitempty"`
	Raw            string
}

// DeleteVolumeSnapshotRequest 볼륨 스냅샷 DeleteVolumeSnapshotRequest 구조체
type DeleteVolumeSnapshotRequest struct {
	SnapshotID string
}

// VolumeSnapshotReference 이미 존재하는 볼륨 스냅샷에 대한 VolumeSnapshotReference 구조체
type VolumeSnapshotReference struct {
	SourceName string `json:"source-name"`
}

// ManageVolumeSnapshotRequest 볼륨의 ManageVolumeSnapshotRequest 구조체
type ManageVolumeSnapshotRequest struct {
	Name        string                  `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	VolumeID    string                  `json:"volume_id,omitempty"`
	Ref         VolumeSnapshotReference `json:"ref"`
}

// ManageExistingVolumeSnapshotRequest 볼륨 스냅샷 import Request 구조체
type ManageExistingVolumeSnapshotRequest struct {
	Snapshot ManageVolumeSnapshotRequest `json:"snapshot"`
}

// ManageExistingVolumeSnapshotResponse 볼륨 스냅샷 import Response 구조체
type ManageExistingVolumeSnapshotResponse struct {
	Snapshot VolumeSnapshotResult `json:"snapshot"`
	Raw      *string
}

// ListAccessibleSnapshots 볼륨 스냅샷 목록 조회
func ListAccessibleSnapshots(c client.Client) (*ListAccessibleSnapshotsResponse, error) {
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
		URL:    endpoint + "/snapshots?all_tenants=true",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListAccessibleSnapshots", t, c, request, response, err)
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

	var rsp ListAccessibleSnapshotsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ShowVolumeSnapshotDetails 볼륨 스냅샷 상세 조회
func ShowVolumeSnapshotDetails(c client.Client, req ShowVolumeSnapshotDetailsRequest) (*ShowVolumeSnapshotDetailsResponse, error) {
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
		URL:    endpoint + "/snapshots/" + req.SnapshotID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowVolumeSnapshotDetails", t, c, request, response, err)
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

	var rsp ShowVolumeSnapshotDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}

// async 동작 수행 완료 확인을 위해 볼륨 스냅샷의 Status 확인
func checkVolumeSnapshotCreation(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.VolumeSnapshotCreationCheckTimeOut)

	for {
		res, err := ShowVolumeSnapshotDetails(c, ShowVolumeSnapshotDetailsRequest{
			SnapshotID: uuid,
		})
		if err != nil {
			return err
		}

		switch res.VolumeSnapshot.Status {
		case "available":
			return nil

		case "error":
			return client.RemoteServerError(errors.New("error occurred during creation of volume snapshot"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.VolumeSnapshotCreationCheckTimeInterval):
			continue
		}
	}
}

// CreateVolumeSnapshot 볼륨 스냅샷을 생성한다
func CreateVolumeSnapshot(c client.Client, req CreateVolumeSnapshotRequest) (*CreateVolumeSnapshotResponse, error) {
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
		URL:    endpoint + "/snapshots",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateVolumeSnapshot", t, c, request, response, err)
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
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var ret CreateVolumeSnapshotResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual volume snapshot deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	if response.Code == 202 {
		if err := checkVolumeSnapshotCreation(c, ret.VolumeSnapshot.UUID); err != nil {
			return &ret, err
		}
	}

	ret.Raw = &response.Body

	return &ret, nil
}

// ManageExistingVolumeSnapshot 기존 볼륨 스냅샷 파일을 이용하여 볼륨 스냅샷을 가져온다
func ManageExistingVolumeSnapshot(c client.Client, req ManageExistingVolumeSnapshotRequest) (*ManageExistingVolumeSnapshotResponse, error) {
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
		URL:    endpoint + "/manageable_snapshots",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ManageExistingVolumeSnapshot", t, c, request, response, err)
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

	var ret ManageExistingVolumeSnapshotResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual volume snapshot deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	if err := checkVolumeSnapshotCreation(c, ret.Snapshot.UUID); err != nil {
		return &ret, err
	}

	ret.Raw = &response.Body

	return &ret, nil
}

// 볼륨 스냅샷 삭제 완료 여부 확인
func checkVolumeSnapshotDeletion(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.VolumeSnapshotDeletionCheckTimeOut)

	for {
		res, err := ShowVolumeSnapshotDetails(c, ShowVolumeSnapshotDetailsRequest{
			SnapshotID: uuid,
		})
		switch {
		case errors.Equal(err, client.ErrNotFound):
			return nil
		case err != nil:
			return err
		}

		switch res.VolumeSnapshot.Status {
		case "deleted":
			return nil

		case "error_deleting":
			return client.RemoteServerError(errors.New("error occurred during deletion of volume snapshot"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.VolumeSnapshotDeletionCheckTimeInterval):
			continue
		}
	}
}

// DeleteVolumeSnapshot 볼륨 스냅샷을 삭제한다
func DeleteVolumeSnapshot(c client.Client, req DeleteVolumeSnapshotRequest) error {
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
		URL:    endpoint + "/snapshots/" + req.SnapshotID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteVolumeSnapshot", t, c, request, response, err)
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
			logger.Warnf("Could not delete volume snapshot. Cause: not found volume snapshot")
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkVolumeSnapshotDeletion(c, req.SnapshotID)
}

// UnmanageVolumeSnapshotRequest 볼륨 스냅샷 unmanage 요청
type UnmanageVolumeSnapshotRequest struct {
	SnapshotUUID string
}

// UnmanageVolumeSnapshot 볼륨 스냅샷 unmanage
func UnmanageVolumeSnapshot(c client.Client, req UnmanageVolumeSnapshotRequest) error {
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
		URL:    fmt.Sprintf("%s/snapshots/%s/action", endpoint, req.SnapshotUUID),
		Body:   []byte("{\"os-unmanage\":{}}"),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("UnmanageVolumeSnapshot", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			logger.Warnf("Could not unmanage volume snapshot. Cause: not found volume snapshot")
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkVolumeSnapshotUnmanaging(c, req.SnapshotUUID)
}

// 볼륨 스냅샷 Unmanage 완료 여부 확인
func checkVolumeSnapshotUnmanaging(c client.Client, uuid string) error {
	timeout := time.After(time.Second * constant.VolumeSnapshotUnmanagingCheckTimeOut)

	for {
		res, err := ShowVolumeSnapshotDetails(c, ShowVolumeSnapshotDetailsRequest{
			SnapshotID: uuid,
		})
		switch {
		case errors.Equal(err, client.ErrNotFound):
			return nil
		case err != nil:
			return err
		}

		if res.VolumeSnapshot.Status != "unmanaging" {
			return client.RemoteServerError(errors.New("error occurred during unmanaging of volume snapshot"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.VolumeSnapshotUnmanagingCheckTimeInterval):
			continue
		}
	}
}

// SnapshotReference snapshot reference 정보
type SnapshotReference struct {
	SourceName string `json:"source-name"`
}

// ManageableSnapshot snapshot manage 정보
type ManageableSnapshot struct {
	VolumeReference   VolumeReference   `json:"source_reference"`
	SafeToManage      bool              `json:"safe_to_manage"`
	SnapshotReference SnapshotReference `json:"reference"`
	Size              uint64            `json:"size"`
}

// ListSummaryOfSnapshotsAvailableToManageResponse snapshot manage 목록 조회 요청
type ListSummaryOfSnapshotsAvailableToManageResponse struct {
	ManageableSnapshots []ManageableSnapshot `json:"manageable-snapshots"`
}

// ListSummaryOfSnapshotsAvailableToManageRequest snapshot manage 목록 조회 응답
type ListSummaryOfSnapshotsAvailableToManageRequest struct {
	Host string
}

// ListSummaryOfSnapshotsAvailableToManage snapshot manage 목록 조회
func ListSummaryOfSnapshotsAvailableToManage(c client.Client, req ListSummaryOfSnapshotsAvailableToManageRequest) (*ListSummaryOfSnapshotsAvailableToManageResponse, error) {
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
		URL:    endpoint + "/manageable_snapshots?host=" + req.Host,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListSummaryOfSnapshotsAvailableToManage", t, c, request, response, err)
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

	var rsp ListSummaryOfSnapshotsAvailableToManageResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}
