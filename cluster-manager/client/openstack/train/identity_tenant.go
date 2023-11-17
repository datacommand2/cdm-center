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

// CreateProjectResult 프로젝트
type CreateProjectResult struct {
	UUID        string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
}

// CreateProjectRequest 프로젝트 생성 요청
type CreateProjectRequest struct {
	Tenant CreateProjectResult `json:"project,omitempty"`
}

// CreateProjectResponse 프로젝트 생성 응답
type CreateProjectResponse struct {
	Tenant CreateProjectResult `json:"project,omitempty"`
	Raw    string
}

// ProjectResult 프로젝트 Result
type ProjectResult struct {
	UUID        string  `json:"id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Enabled     bool    `json:"enabled,omitempty"`
}

// ShowProjectDetailsRequest 프로젝트 조회 request
type ShowProjectDetailsRequest struct {
	ProjectID string
}

// ListProjectsResponse 프로젝트 목록 조회 response
type ListProjectsResponse struct {
	Projects []ProjectResult `json:"projects,omitempty"`
}

// ShowProjectDetailsResponse 프로젝트 조회 response
type ShowProjectDetailsResponse struct {
	Project ProjectResult `json:"project,omitempty"`
	Raw     string
}

// DeleteProjectRequest 프로젝트 삭제 요청
type DeleteProjectRequest struct {
	ProjectUUID string
}

// CreateProject 테넌트 생성
func CreateProject(c client.Client, in CreateProjectRequest) (*CreateProjectResponse, error) {
	// convert
	body, err := json.Marshal(&in)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	// request
	endpoint, err := c.GetEndpoint(ServiceTypeIdentity)
	if err != nil {
		return nil, err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodPost,
		URL:    endpoint + "/projects",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateProject", t, c, request, response, err)
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
		case response.Code == 409: // 테넌트 이름 충돌
			return nil, client.Conflict(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	// convert
	var ret CreateProjectResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		return nil, errors.Unknown(err)
	}
	ret.Raw = response.Body

	logger.Infof("CreateProject : %+v", response.Body)
	return &ret, nil
}

// DeleteProject 프로젝트 삭제
// 프로젝트 삭제 시, 자동으로 생성된 보안 그룹도 같이 삭제한다.
// Web 대시보드에서는 보안 그룹도 같이 삭제되는 것처럼 보이나, openstack command 로 검색하면 해당 프로젝트의 보안 그룹이 그대로 남아있다.
func DeleteProject(c client.Client, in DeleteProjectRequest) error {
	// request
	endpoint, err := c.GetEndpoint(ServiceTypeIdentity)
	if err != nil {
		return err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodDelete,
		URL:    endpoint + "/projects/" + in.ProjectUUID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteProject", t, c, request, response, err)
	if err != nil {
		return err
	}

	logger.Infof("DeleteProject : %s", in.ProjectUUID)

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return client.Unknown(response)
		case response.Code == 401:
			return client.Unknown(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404: // 삭제하려는 테넌트 없음
			logger.Warnf("Could not delete tenant. Cause: not found tenant")
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	// 삭제된 테넌트로 로그인되어 있는 경우
	// 이후 명령이 동작하지 않으므로 reconnect 수행
	if err := c.Close(); err != nil {
		logger.Warnf("Could not close client. Cause: %+v", err)
	}

	if err := c.Connect(); err != nil {
		return err
	}

	// 남아있는 보안 그룹을 제거한다.
	groups, err := ListSecurityGroups(c)
	if err != nil {
		logger.Warnf("Security groups may remains for tenant. uuid: %s, Cause: %+v", in.ProjectUUID, err)
		return nil
	}

	for _, group := range groups.SecurityGroups {
		if group.ProjectID == in.ProjectUUID {
			err := DeleteSecurityGroup(c, DeleteSecurityGroupRequest{SecurityGroupUUID: group.UUID})
			if err != nil {
				logger.Warnf("Could not delete security group. uuid: %s, Cause: %+v", group.UUID, err)
			}
		}
	}

	return nil
}

// ListProjects 프로젝트 목록 조회
func ListProjects(c client.Client) (*ListProjectsResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeIdentity)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/projects",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListProjects", t, c, request, response, err)
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

	var rsp ListProjectsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ShowProjectDetails 프로젝트 상세 조회
func ShowProjectDetails(c client.Client, req ShowProjectDetailsRequest) (*ShowProjectDetailsResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeIdentity)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/projects/" + req.ProjectID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowProjectDetails", t, c, request, response, err)
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

	var rsp ShowProjectDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}
