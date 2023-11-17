package train

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"net/http"
	"time"
)

// ShowNetworkQuotaRequest Network quota-set request
type ShowNetworkQuotaRequest struct {
	ProjectID string
}

// ShowNetworkQuotaResponse Network quota-set response
type ShowNetworkQuotaResponse struct {
	QuotaSets map[string]int64
}

// UpdateNetworkQuotaResult network 쿼타
type UpdateNetworkQuotaResult struct {
	FloatingIP        int64 `json:"floatingip,omitempty"`
	Network           int64 `json:"network,omitempty"`
	Port              int64 `json:"port,omitempty"`
	RbacPolicy        int64 `json:"rbac_policy,omitempty"`
	Router            int64 `json:"router,omitempty"`
	SecurityGroup     int64 `json:"security_group,omitempty"`
	SecurityGroupRule int64 `json:"security_group_rule,omitempty"`
	Subnet            int64 `json:"subnet,omitempty"`
	SubnetPool        int64 `json:"subnetpool,omitempty"`
	//Trunk             int64 `json:"trunk,omitempty"`
}

// UpdateNetworkQuotaRequest network 쿼타 갱신 요청
type UpdateNetworkQuotaRequest struct {
	TenantUUID string                   `json:"-"`
	QuotaSet   UpdateNetworkQuotaResult `json:"quota,omitempty"`
}

// UpdateNetworkQuotaResponse network 쿼타 갱신 응답
type UpdateNetworkQuotaResponse struct {
	QuotaSet UpdateNetworkQuotaResult `json:"quota,omitempty"`
}

// Set 네트워크 쿼타 설정
func (u *UpdateNetworkQuotaRequest) Set(key string, value int64) {
	switch key {
	case client.QuotaNetworkFloatingIPs:
		u.QuotaSet.FloatingIP = value
	case client.QuotaNetworkNetworks:
		u.QuotaSet.Network = value
	case client.QuotaNetworkSecurityGroupRules:
		u.QuotaSet.SecurityGroupRule = value
	case client.QuotaNetworkSecurityGroups:
		u.QuotaSet.SecurityGroup = value
	case client.QuotaNetworkPorts:
		u.QuotaSet.Port = value
	case client.QuotaNetworkRouters:
		u.QuotaSet.Router = value
	case client.QuotaNetworkSubnets:
		u.QuotaSet.Subnet = value
	case client.QuotaNetworkSubnetPools:
		u.QuotaSet.SubnetPool = value
	case client.QuotaNetworkRBACPolicies:
		u.QuotaSet.RbacPolicy = value
		//case client.QuotaNetworkTrunks:
		//	u.QuotaSet.Trunk = value
	}
}

// ToMap UpdateNetworkQuotaResponse -> map
func (u *UpdateNetworkQuotaResponse) ToMap() map[string]int64 {
	if u == nil {
		return nil
	}
	return map[string]int64{
		client.QuotaNetworkFloatingIPs:        u.QuotaSet.FloatingIP,
		client.QuotaNetworkNetworks:           u.QuotaSet.Network,
		client.QuotaNetworkSecurityGroupRules: u.QuotaSet.SecurityGroupRule,
		client.QuotaNetworkSecurityGroups:     u.QuotaSet.SecurityGroup,
		client.QuotaNetworkPorts:              u.QuotaSet.Port,
		client.QuotaNetworkRouters:            u.QuotaSet.Router,
		client.QuotaNetworkSubnets:            u.QuotaSet.Subnet,
		client.QuotaNetworkSubnetPools:        u.QuotaSet.SubnetPool,
		client.QuotaNetworkRBACPolicies:       u.QuotaSet.RbacPolicy,
		//client.QuotaNetworkTrunks:             u.QuotaSet.Trunk,
	}
}

// UpdateNetworkQuotaForAProject 네트워크 쿼타 생성
func UpdateNetworkQuotaForAProject(c client.Client, in UpdateNetworkQuotaRequest) (*UpdateNetworkQuotaResponse, error) {
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
		Method: http.MethodPut,
		URL:    endpoint + "/v2.0/quotas/" + in.TenantUUID,
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("UpdateNetworkQuotaForAProject", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 401:
			return nil, client.Unknown(response)
		case response.Code == 403:
			return nil, client.Unknown(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	// convert
	var ret UpdateNetworkQuotaResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}

// ShowNetworkQuota Network quota-set 조회
func ShowNetworkQuota(c client.Client, req ShowNetworkQuotaRequest) (*ShowNetworkQuotaResponse, error) {
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
		URL:    endpoint + "/v2.0/quotas/" + req.ProjectID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowNetworkQuota", t, c, request, response, err)
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

	var dat map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &dat); err != nil {
		return nil, errors.Unknown(err)
	}

	quotaSets := make(map[string]int64)
	for k, v := range dat["quota"].(map[string]interface{}) {
		quotaSets[k] = int64(v.(float64))
	}

	return &ShowNetworkQuotaResponse{QuotaSets: quotaSets}, nil
}
