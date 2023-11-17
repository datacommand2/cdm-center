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

// VolumeGroupSnapshot 볼륨 그룹 스냅샷
type VolumeGroupSnapshot struct {
	// 생성 요청
	GroupID     string `json:"group_id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`

	ID          string `json:"id,omitempty"`
	Status      string `json:"status,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	GroupTypeID string `json:"group_type_id,omitempty"`
	ProjectUUID string `json:"project_id,omitempty"`
}

// ListVolumeGroupSnapshotsWithDetailsResponse 볼륨 그룹 스냅샷 목록 상세 조회 응답
type ListVolumeGroupSnapshotsWithDetailsResponse struct {
	VolumeGroupSnapshots []VolumeGroupSnapshot `json:"group_snapshots,omitempty"`
}

// ListVolumeGroupSnapshotsWithDetails 볼륨 그룹 스냅샷 목록 상세 조회
func ListVolumeGroupSnapshotsWithDetails(c client.Client) (*ListVolumeGroupSnapshotsWithDetailsResponse, error) {
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
		URL:    endpoint + "/group_snapshots/detail?all_tenants=true",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListVolumeGroupSnapshotsWithDetails", t, c, request, response, err)
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
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp ListVolumeGroupSnapshotsWithDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// DeleteVolumeGroupSnapshotRequest 볼륨 그룹 스냅샷 제거 요청
type DeleteVolumeGroupSnapshotRequest struct {
	VolumeGroupSnapshotID string
}

// DeleteVolumeGroupSnapshot 볼륨 그룹 스냅샷 제거
func DeleteVolumeGroupSnapshot(c client.Client, in DeleteVolumeGroupSnapshotRequest) error {
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
		URL:    endpoint + "/group_snapshots/" + in.VolumeGroupSnapshotID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteVolumeGroupSnapshot", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return client.Unknown(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			logger.Warnf("Could not delete volume group snapshot. Cause: not found volume group snapshot")
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkVolumeGroupSnapshotDeletion(c, in.VolumeGroupSnapshotID)
}

func checkVolumeGroupSnapshotDeletion(c client.Client, groupSnapshotUUID string) error {
	timeout := time.After(time.Second * constant.VolumeGroupSnapshotDeletionCheckTimeOut)

	for {
		res, err := ShowVolumeGroupSnapshotDetails(c, ShowVolumeGroupSnapshotDetailsRequest{VolumeGroupSnapshotUUID: groupSnapshotUUID})
		if errors.Equal(err, client.ErrNotFound) {
			return nil
		} else if err != nil {
			return err
		}

		if res.VolumeGroupSnapshot.Status == "error_deleting" {
			return client.RemoteServerError(errors.New("error occurred during deletion of volume group snapshot"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.VolumeGroupSnapshotDeletionCheckTimeInterval):
			continue
		}
	}
}

// ShowVolumeGroupSnapshotDetailsRequest 볼륨 그룹 스냅샷 상세 조회 요청
type ShowVolumeGroupSnapshotDetailsRequest struct {
	VolumeGroupSnapshotUUID string
}

// VolumeGroupSnapshotResponse 볼륨 그룹 스냅샷 생성 및 조회 응답
type VolumeGroupSnapshotResponse struct {
	VolumeGroupSnapshot VolumeGroupSnapshot `json:"group_snapshot,omitempty"`
}

// ShowVolumeGroupSnapshotDetails 볼륨 그룹 스냅샷 상세 조회
func ShowVolumeGroupSnapshotDetails(c client.Client, in ShowVolumeGroupSnapshotDetailsRequest) (*VolumeGroupSnapshotResponse, error) {
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
		URL:    endpoint + "/group_snapshots/" + in.VolumeGroupSnapshotUUID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowVolumeGroupSnapshotDetails", t, c, request, response, err)
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
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	var rsp VolumeGroupSnapshotResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// CreateVolumeGroupSnapshotRequest 볼륨 그룹 스냅샷 생성 요청
type CreateVolumeGroupSnapshotRequest struct {
	VolumeGroupSnapshot VolumeGroupSnapshot `json:"group_snapshot"`
}

// CreateVolumeGroupSnapshot 볼륨 그룹 스냅샷 생성
func CreateVolumeGroupSnapshot(c client.Client, in CreateVolumeGroupSnapshotRequest) (*VolumeGroupSnapshotResponse, error) {
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
		URL:    endpoint + "/group_snapshots",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateVolumeGroupSnapshot", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			// group_uuid 가 잘못된 경우
			// group 에 볼륨 멤버가 없는 경우
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

	var ret VolumeGroupSnapshotResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual volume group snapshot deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	if err := checkVolumeGroupSnapshotCreation(c, ret.VolumeGroupSnapshot.ID); err != nil {
		return &ret, err
	}

	return &ret, nil
}

func checkVolumeGroupSnapshotCreation(c client.Client, groupSnapshotUUID string) error {
	timeout := time.After(time.Second * constant.VolumeGroupSnapshotCreationCheckTimeOut)
	now := time.Now()

	for {
		res, err := ShowVolumeGroupSnapshotDetails(c, ShowVolumeGroupSnapshotDetailsRequest{VolumeGroupSnapshotUUID: groupSnapshotUUID})
		if err != nil {
			return err
		}

		switch res.VolumeGroupSnapshot.Status {
		case "available":
			logger.Infof("[checkVolumeGroupSnapshotCreation] volume group snapshot(%s) creation elapsed time: %v", groupSnapshotUUID, time.Since(now))
			return nil
		case "error":
			return client.RemoteServerError(errors.New("error occurred during creation of volume group snapshot"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.VolumeGroupSnapshotCreationCheckTimeInterval):
			continue
		}
	}
}
