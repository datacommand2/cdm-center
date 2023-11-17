package train

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"net/http"
	"time"
)

// UpdateComputeQuotaResult compute 쿼타
type UpdateComputeQuotaResult struct {
	Instances                int64 `json:"instances,omitempty"`
	KeyPairs                 int64 `json:"key_pairs,omitempty"`
	MetadataItems            int64 `json:"metadata_items,omitempty"`
	RAM                      int64 `json:"ram,omitempty"`
	ServerGroups             int64 `json:"server_groups,omitempty"`
	ServerGroupMembers       int64 `json:"server_group_members,omitempty"`
	VCPUs                    int64 `json:"cores,omitempty"`
	InjectedFileContentBytes int64 `json:"injected_file_content_bytes,omitempty"`
	InjectedFilePathBytes    int64 `json:"injected_file_path_bytes,omitempty"`
	InjectedFiles            int64 `json:"injected_files,omitempty"`
}

// UpdateComputeQuotaRequest compute 쿼타 갱신 요청
type UpdateComputeQuotaRequest struct {
	TenantUUID string                   `json:"-"`
	QuotaSet   UpdateComputeQuotaResult `json:"quota_set,omitempty"`
}

// UpdateComputeQuotaResponse compute 쿼타 갱신 응답
type UpdateComputeQuotaResponse struct {
	QuotaSet UpdateComputeQuotaResult `json:"quota_set,omitempty"`
}

// ShowComputeQuotaRequest Compute quota-set request
type ShowComputeQuotaRequest struct {
	ProjectID string
}

// ShowComputeQuotaResponse Compute quota-set response
type ShowComputeQuotaResponse struct {
	QuotaSets map[string]int64
}

// Set 컴퓨트 쿼타 설정
func (u *UpdateComputeQuotaRequest) Set(key string, value int64) {
	switch key {
	case client.QuotaComputeInstances:
		u.QuotaSet.Instances = value
	case client.QuotaComputeKeyPairs:
		u.QuotaSet.KeyPairs = value
	case client.QuotaComputeMetadataItems:
		u.QuotaSet.MetadataItems = value
	case client.QuotaComputeRAMSize:
		u.QuotaSet.RAM = value
	case client.QuotaComputeServerGroups:
		u.QuotaSet.ServerGroups = value
	case client.QuotaComputeServerGroupMembers:
		u.QuotaSet.ServerGroupMembers = value
	case client.QuotaComputeVCPUs:
		u.QuotaSet.VCPUs = value
	case client.QuotaComputeInjectedFileContentBytes:
		u.QuotaSet.InjectedFileContentBytes = value
	case client.QuotaComputeInjectedFilePathBytes:
		u.QuotaSet.InjectedFilePathBytes = value
	case client.QuotaComputeInjectedFiles:
		u.QuotaSet.InjectedFiles = value
	}
}

// ToMap UpdateComputeQuotaResponse -> map
func (u *UpdateComputeQuotaResponse) ToMap() map[string]int64 {
	if u == nil {
		return nil
	}
	return map[string]int64{
		client.QuotaComputeInstances:                u.QuotaSet.Instances,
		client.QuotaComputeKeyPairs:                 u.QuotaSet.KeyPairs,
		client.QuotaComputeMetadataItems:            u.QuotaSet.MetadataItems,
		client.QuotaComputeRAMSize:                  u.QuotaSet.RAM,
		client.QuotaComputeServerGroups:             u.QuotaSet.ServerGroups,
		client.QuotaComputeServerGroupMembers:       u.QuotaSet.ServerGroupMembers,
		client.QuotaComputeVCPUs:                    u.QuotaSet.VCPUs,
		client.QuotaComputeInjectedFileContentBytes: u.QuotaSet.InjectedFileContentBytes,
		client.QuotaComputeInjectedFilePathBytes:    u.QuotaSet.InjectedFilePathBytes,
		client.QuotaComputeInjectedFiles:            u.QuotaSet.InjectedFiles,
	}
}

// ShowComputeQuota Compute quota-set 조회
func ShowComputeQuota(c client.Client, req ShowComputeQuotaRequest) (*ShowComputeQuotaResponse, error) {
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}

	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":          []string{"application/json"},
			HeaderXAuthToken:        []string{c.GetTokenKey()},
			"OpenStack-API-Version": []string{"compute 2.79"},
		},
		Method: http.MethodGet,
		URL:    endpoint + "/os-quota-sets/" + req.ProjectID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowComputeQuota", t, c, request, response, err)
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
	for k, v := range dat["quota_set"].(map[string]interface{}) {
		if k == "id" {
			continue
		}
		quotaSets[k] = int64(v.(float64))
	}

	return &ShowComputeQuotaResponse{QuotaSets: quotaSets}, nil
}

// UpdateComputeQuotas 컴퓨트 쿼타 생성
func UpdateComputeQuotas(c client.Client, in UpdateComputeQuotaRequest) (*UpdateComputeQuotaResponse, error) {
	// convert
	body, err := json.Marshal(&in)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	// request
	endpoint, err := c.GetEndpoint(ServiceTypeCompute)
	if err != nil {
		return nil, err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodPut,
		URL:    endpoint + "/os-quota-sets/" + in.TenantUUID,
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("UpdateComputeQuotas", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			// 400: 테넌트 ID가 잘못된 경우 in doc
			return nil, client.Unknown(response)
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
	var ret UpdateComputeQuotaResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}
