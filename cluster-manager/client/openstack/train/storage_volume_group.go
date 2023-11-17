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

// VolumeGroup 볼륨 그룹
type VolumeGroup struct {
	Name             string   `json:"name,omitempty"`
	Description      string   `json:"description,omitempty"`
	AvailabilityZone string   `json:"availability_zone,omitempty"`
	GroupType        string   `json:"group_type,omitempty"`
	VolumeTypes      []string `json:"volume_types,omitempty"`
	Volumes          []string `json:"volumes,omitempty"`

	// 생성 요청시 X
	ID        string `json:"id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
	Status    string `json:"status,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// VolumeGroups 볼륨 그룹 목록 조회 구조체
type VolumeGroups struct {
	UUID string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// ListAllVolumeGroupsResponse 볼륨 그룹 목록 조회 응답
type ListAllVolumeGroupsResponse struct {
	Groups []VolumeGroups `json:"consistencygroups"`
}

// CreateVolumeGroupRequest 볼륨 그룹 생성 요청
type CreateVolumeGroupRequest struct {
	Group VolumeGroup `json:"group"`
}

// VolumeGroupResponse 볼륨 그룹 생성 및 조회 응답
type VolumeGroupResponse struct {
	Group VolumeGroup `json:"group"`
}

// ShowVolumeGroupDetailsRequest 볼륨 그룹 상세 조회 요청
type ShowVolumeGroupDetailsRequest struct {
	GroupUUID   string `json:"-"`
	ShowVolumes bool   `json:"-"`
}

// DeleteWrapper 볼륨 그룹의 볼륨 삭제 여부
type DeleteWrapper struct {
	DeleteVolumes bool `json:"delete-volumes"`
}

// DeleteVolumeGroupRequest 볼륨 그룹 삭제 요청
type DeleteVolumeGroupRequest struct {
	GroupID string        `json:"-"`
	Delete  DeleteWrapper `json:"delete"`
}

// ListAllVolumeGroups 볼륨 타입 목록 조회
func ListAllVolumeGroups(c client.Client) (*ListAllVolumeGroupsResponse, error) {
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
		URL:    endpoint + "/consistencygroups",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListAllVolumeGroups", t, c, request, response, err)
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

	var rsp ListAllVolumeGroupsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// CreateVolumeGroup (create group) 볼륨 그룹 생성
func CreateVolumeGroup(c client.Client, in CreateVolumeGroupRequest) (*VolumeGroupResponse, error) {
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
		URL:    endpoint + "/groups",
		Body:   body,
	}
	logger.Infof("[CreateVolumeGroup] method : %+v", request.Method)
	logger.Infof("[CreateVolumeGroup] URL : %+v", request.URL)
	logger.Infof("[CreateVolumeGroup] BODY : %+v", string(request.Body[:]))
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateVolumeGroup", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			// group.group_type 지정 안 된 경우
			// group.volume_types 지정 안 된 경우
			// 기본 볼륨 그룹 타입은 사용 불가
			//   - Group_type 22478632-040c-4513-8080-3ff3774650f1 is reserved for migrating CGs to groups. Migrated group can only be operated by CG APIs.
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

	var ret VolumeGroupResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual volume group deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	if err := checkVolumeGroupCreation(c, ret.Group.ID); err != nil {
		return &ret, err
	}

	return &ret, nil
}

func checkVolumeGroupCreation(c client.Client, volumeGroupUUID string) error {
	timeout := time.After(time.Second * constant.VolumeGroupCreationCheckTimeOut)

	for {
		res, err := ShowVolumeGroupDetails(c, ShowVolumeGroupDetailsRequest{GroupUUID: volumeGroupUUID})
		if err != nil {
			return err
		}

		switch res.Group.Status {
		case "available":
			return nil
		case "error":
			return client.RemoteServerError(errors.New("error occurred during creation of volume group"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.VolumeGroupCreationCheckTimeInterval):
			continue
		}
	}
}

// DeleteVolumeGroup (delete group) 볼륨 그룹 삭제
// 그룹에 속한 볼륨을 제거할 것인지를 묻는 옵션이 있음
func DeleteVolumeGroup(c client.Client, in DeleteVolumeGroupRequest) error {
	body, err := json.Marshal(&in)
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
		URL:    fmt.Sprintf("%s/groups/%s/action", endpoint, in.GroupID),
		Body:   body,
	}
	logger.Infof("[DeleteVolumeGroup] method : %+v", request.Method)
	logger.Infof("[DeleteVolumeGroup] URL : %+v", request.URL)
	logger.Infof("[DeleteVolumeGroup] BODY : %+v", string(request.Body[:]))
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteVolumeGroup", t, c, request, response, err)
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
			logger.Warnf("Could not delete volume group. Cause: not found volume group")
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkVolumeGroupDeletion(c, in.GroupID)
}

func checkVolumeGroupDeletion(c client.Client, groupUUID string) error {
	timeout := time.After(time.Second * constant.VolumeGroupDeletionCheckTimeOut)
	now := time.Now()

	for {
		res, err := ShowVolumeGroupDetails(c, ShowVolumeGroupDetailsRequest{GroupUUID: groupUUID})
		if errors.Equal(err, client.ErrNotFound) {
			logger.Infof("[checkVolumeGroupDeletion] volume group(%s) deletion elapsed time: %v", groupUUID, time.Since(now))
			return nil
		} else if err != nil {
			return err
		}

		if res.Group.Status == "error_deleting" {
			return client.RemoteServerError(errors.New("error occurred during deletion of volume group"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.VolumeGroupDeletionCheckTimeInterval):
			continue
		}
	}
}

// ShowVolumeGroupDetails (Show group details) 볼륨 그룹 상세 조회
func ShowVolumeGroupDetails(c client.Client, in ShowVolumeGroupDetailsRequest) (*VolumeGroupResponse, error) {
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
		URL:    fmt.Sprintf("%s/groups/%s?list_volume=%v", endpoint, in.GroupUUID, in.ShowVolumes),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowVolumeGroupDetails", t, c, request, response, err)
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

	var ret VolumeGroupResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}

// UpdateVolumeGroupWrapper 볼륨 그룹
type UpdateVolumeGroupWrapper struct {
	Name          string `json:"name,omitempty"`
	Description   string `json:"description,omitempty"`
	AddVolumes    string `json:"add_volumes,omitempty"`
	RemoveVolumes string `json:"remove_volumes,omitempty"`
}

// UpdateVolumeGroupRequest 볼륨 그룹 수정 요청
type UpdateVolumeGroupRequest struct {
	VolumeGroupUUID string                   `json:"-"`
	Group           UpdateVolumeGroupWrapper `json:"group,omitempty"`
}

func checkVolumeGroupUpdate(c client.Client, volumeGroupUUID string) error {
	timeout := time.After(time.Second * constant.VolumeGroupUpdateCheckTimeOut)

	for {
		res, err := ShowVolumeGroupDetails(c, ShowVolumeGroupDetailsRequest{GroupUUID: volumeGroupUUID})
		if err != nil {
			return err
		}

		switch res.Group.Status {
		case "available":
			return nil
		case "error":
			return client.RemoteServerError(errors.New("error occurred during update of volume group"))
		}

		select {
		case <-timeout:
			return errors.Unknown(errors.New("asynchronous processing is timeout"))

		case <-time.After(time.Second * constant.VolumeGroupUpdateCheckTimeInterval):
			continue
		}
	}
}

// UpdateVolumeGroup 볼륨 그룹 수정
func UpdateVolumeGroup(c client.Client, in UpdateVolumeGroupRequest) error {
	body, err := json.Marshal(&in)
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
		Method: http.MethodPut,
		URL:    endpoint + "/groups/" + in.VolumeGroupUUID,
		Body:   body,
	}
	logger.Infof("[UpdateVolumeGroup] method : %+v", request.Method)
	logger.Infof("[UpdateVolumeGroup] URL : %+v", request.URL)
	logger.Infof("[UpdateVolumeGroup] BODY : %+v", string(request.Body[:]))
	t := time.Now()
	response, err := request.Request()
	requestLog("UpdateVolumeGroup", t, c, request, response, err)
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
			return client.NotFound(response)
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return checkVolumeGroupUpdate(c, in.VolumeGroupUUID)
}
