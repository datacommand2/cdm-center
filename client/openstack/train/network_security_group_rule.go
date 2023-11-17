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

// CreateSecurityGroupRuleResult 보안 그룹 규칙 생성
type CreateSecurityGroupRuleResult struct {
	UUID              string  `json:"id,omitempty"`
	Direction         string  `json:"direction,omitempty"`
	SecurityGroupUUID string  `json:"security_group_id,omitempty"`
	RemoteGroupUUID   string  `json:"remote_group_id,omitempty"`
	Protocol          *string `json:"protocol,omitempty"`
	EtherType         *string `json:"ethertype,omitempty"`
	PortRangeMax      *uint64 `json:"port_range_max,omitempty"`
	PortRangeMin      *uint64 `json:"port_range_min,omitempty"`
	RemoteIPPrefix    *string `json:"remote_ip_prefix,omitempty"`
	Description       *string `json:"description,omitempty"`
}

// CreateSecurityGroupRuleRequest 보안 그룹 규칙 생성 요청
type CreateSecurityGroupRuleRequest struct {
	Rule CreateSecurityGroupRuleResult `json:"security_group_rule"`
}

// CreateSecurityGroupRuleResponse 보안 그룹 규칙 생성 응답
type CreateSecurityGroupRuleResponse struct {
	Rule CreateSecurityGroupRuleResult `json:"security_group_rule"`
	Raw  string
}

// SecurityGroupRuleResult 보안 그룹 규칙 result
type SecurityGroupRuleResult struct {
	SecurityGroupID string  `json:"security_group_id,omitempty"`
	RemoteGroupID   *string `json:"remote_group_id,omitempty"`
	UUID            string  `json:"id,omitempty"`
	Description     *string `json:"description,omitempty"`
	NetworkCidr     *string `json:"remote_ip_prefix,omitempty"`
	Direction       string  `json:"direction,omitempty"`
	PortRangeMin    *uint64 `json:"port_range_min,omitempty"`
	PortRangeMax    *uint64 `json:"port_range_max,omitempty"`
	Protocol        *string `json:"protocol,omitempty"`
	EtherType       string  `json:"ethertype,omitempty"`
}

// ShowSecurityGroupRuleRequest 보안 그룹 규칙 조회 request
type ShowSecurityGroupRuleRequest struct {
	SecurityGroupRuleID string
}

// ShowSecurityGroupRuleResponse 보안 그룹 규칙 조회 response
type ShowSecurityGroupRuleResponse struct {
	SecurityGroupRule SecurityGroupRuleResult `json:"security_group_rule,omitempty"`
	Raw               string
}

// DeleteSecurityGroupRuleRequest 보안 그룹 규칙 삭제 요청
type DeleteSecurityGroupRuleRequest struct {
	RuleUUID string
}

// Uint32EtherType 숫자를 IPv4 혹은 IPv6 로 변경
func Uint32EtherType(t uint32) *string {
	var ipType string
	if t == 4 {
		ipType = "IPv4"
	} else {
		ipType = "IPv6"
	}
	return &ipType
}

// StringEtherType IPv4 혹은 IPv6를 숫자로 변경
func StringEtherType(t string) uint32 {
	if t == "IPv6" {
		return 6
	}
	return 4
}

// CreateSecurityGroupRule 보안 그룹 규칙 생성
func CreateSecurityGroupRule(c client.Client, in CreateSecurityGroupRuleRequest) (*CreateSecurityGroupRuleResponse, error) {
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
		URL:    endpoint + "/v2.0/security-group-rules",
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("CreateSecurityGroupRule", t, c, request, response, err)
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

	// convert
	var ret CreateSecurityGroupRuleResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		// 실패 시, security group 이 삭제되므로 에러 처리 X
		return nil, errors.Unknown(err)
	}
	ret.Raw = response.Body

	return &ret, nil
}

// DeleteSecurityGroupRule 보안 그룹 규칙 삭제
func DeleteSecurityGroupRule(c client.Client, in DeleteSecurityGroupRuleRequest) error {
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
		URL:    fmt.Sprintf("%s/v2.0/security-group-rules/%s", endpoint, in.RuleUUID),
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("DeleteSecurityGroupRule", t, c, request, response, err)
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
			logger.Warnf("Could not delete security group rule. Cause: not found security group rule")
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

// ShowSecurityGroupRule 보안 그룹 규칙 상세 조회
func ShowSecurityGroupRule(c client.Client, req ShowSecurityGroupRuleRequest) (*ShowSecurityGroupRuleResponse, error) {
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
		URL:    endpoint + "/v2.0/security-group-rules/" + req.SecurityGroupRuleID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowSecurityGroupRule", t, c, request, response, err)
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

	var rsp ShowSecurityGroupRuleResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}
	rsp.Raw = response.Body

	return &rsp, nil
}
