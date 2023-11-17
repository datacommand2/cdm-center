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

// CreateSecurityGroupResult 보안 그룹 생성
type CreateSecurityGroupResult struct {
	UUID        string                          `json:"id,omitempty"`
	Name        string                          `json:"name,omitempty"`
	ProjectUUID string                          `json:"project_id,omitempty"`
	Description *string                         `json:"description,omitempty"`
	Rules       []CreateSecurityGroupRuleResult `json:"security_group_rules,omitempty"`
}

// CreateSecurityGroupRequest 보안 그룹 생성 요청
type CreateSecurityGroupRequest struct {
	SecurityGroup CreateSecurityGroupResult `json:"security_group"`
}

// CreateSecurityGroupResponse 보안 그룹 생성 응답
type CreateSecurityGroupResponse struct {
	SecurityGroup CreateSecurityGroupResult `json:"security_group"`
	Raw           string                    `json:"-"`
}

// SecurityGroupResult 보안 그룹 Result
type SecurityGroupResult struct {
	UUID        string                    `json:"id,omitempty"`
	Name        string                    `json:"name,omitempty"`
	Description *string                   `json:"description,omitempty"`
	ProjectID   string                    `json:"project_id,omitempty"`
	Rules       []SecurityGroupRuleResult `json:"security_group_rules,omitempty"`
}

// ShowSecurityGroupRequest 보안 그룹 조회 request
type ShowSecurityGroupRequest struct {
	SecurityGroupID string
}

// ListSecurityGroupsResponse 보안 그룹 목록 조회 response
type ListSecurityGroupsResponse struct {
	SecurityGroups []SecurityGroupResult `json:"security_groups,omitempty"`
}

// ShowSecurityGroupResponse 보안 그룹 조회 response
type ShowSecurityGroupResponse struct {
	SecurityGroup SecurityGroupResult `json:"security_group,omitempty"`
	Raw           string
}

// CreateSecurityGroup 보안 그룹 생성
func CreateSecurityGroup(c client.Client, in CreateSecurityGroupRequest) (*CreateSecurityGroupResponse, error) {
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
		URL:    endpoint + "/v2.0/security-groups",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateSecurityGroup", t, c, request, response, err)
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
			// Quota exceeded for resources
			return nil, client.Conflict(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	// convert
	var ret CreateSecurityGroupResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual security group deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}
	ret.Raw = response.Body

	return &ret, nil
}

// DeleteSecurityGroupRequest 보안 그룹 삭제 요청
type DeleteSecurityGroupRequest struct {
	SecurityGroupUUID string
}

// DeleteSecurityGroup 보안 그룹 삭제
func DeleteSecurityGroup(c client.Client, in DeleteSecurityGroupRequest) error {
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
		URL:    fmt.Sprintf("%s/v2.0/security-groups/%s", endpoint, in.SecurityGroupUUID),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteSecurityGroup", t, c, request, response, err)
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
			logger.Warnf("Could not delete security group. Cause: not found security group")
		case response.Code == 409:
			return client.Conflict(response)
		case response.Code == 412:
			return client.Unknown(response)
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return nil
}

// ListSecurityGroups 보안 그룹 목록 조회
func ListSecurityGroups(c client.Client) (*ListSecurityGroupsResponse, error) {
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
		URL:    endpoint + "/v2.0/security-groups",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListSecurityGroups", t, c, request, response, err)
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

	var rsp ListSecurityGroupsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}

// ShowSecurityGroup 보안 그룹 상세 조회
func ShowSecurityGroup(c client.Client, req ShowSecurityGroupRequest) (*ShowSecurityGroupResponse, error) {
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
		URL:    endpoint + "/v2.0/security-groups/" + req.SecurityGroupID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowSecurityGroup", t, c, request, response, err)
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

	var rsp ShowSecurityGroupResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}
