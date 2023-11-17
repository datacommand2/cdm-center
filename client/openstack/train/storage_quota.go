package train

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"net/http"
	"time"
)

// ShowStorageQuotaRequest Storage quota-set request
type ShowStorageQuotaRequest struct {
	ProjectID string
}

// ShowStorageQuotaResponse Storage quota-set response
type ShowStorageQuotaResponse struct {
	QuotaSets map[string]int64
}

// UpdateStorageQuotaResult storage 쿼타
type UpdateStorageQuotaResult struct {
	Volumes            int64 `json:"volumes,omitempty"`
	Snapshots          int64 `json:"snapshots,omitempty"`
	Groups             int64 `json:"groups,omitempty"`
	PerVolumeGigabytes int64 `json:"per_volume_gigabytes,omitempty"`
	Gigabytes          int64 `json:"gigabytes,omitempty"`
	Backups            int64 `json:"backups,omitempty"`
	BackupGigabytes    int64 `json:"backup_gigabytes,omitempty"`
}

// UpdateStorageQuotaRequest storage 쿼타 갱신 요청
type UpdateStorageQuotaRequest struct {
	TenantUUID string                   `json:"-"`
	QuotaSet   UpdateStorageQuotaResult `json:"quota_set,omitempty"`
}

// UpdateStorageQuotaResponse storage 쿼타 갱신 응답
type UpdateStorageQuotaResponse struct {
	QuotaSet UpdateStorageQuotaResult `json:"quota_set,omitempty"`
}

// Set 스토리지 쿼타 설정
func (u *UpdateStorageQuotaRequest) Set(key string, value int64) {
	switch key {
	case client.QuotaStorageBackupGigabytes:
		u.QuotaSet.BackupGigabytes = value
	case client.QuotaStorageBackups:
		u.QuotaSet.Backups = value
	case client.QuotaStorageGigabytes:
		u.QuotaSet.Gigabytes = value
	case client.QuotaStorageGroups:
		u.QuotaSet.Groups = value
	case client.QuotaStoragePerVolumeGigabytes:
		u.QuotaSet.PerVolumeGigabytes = value
	case client.QuotaStorageSnapshots:
		u.QuotaSet.Snapshots = value
	case client.QuotaStorageVolumes:
		u.QuotaSet.Volumes = value
	}
}

// ToMap UpdateStorageQuotaResponse -> map
func (u *UpdateStorageQuotaResponse) ToMap() map[string]int64 {
	if u == nil {
		return nil
	}
	return map[string]int64{
		client.QuotaStorageBackupGigabytes:    u.QuotaSet.BackupGigabytes,
		client.QuotaStorageBackups:            u.QuotaSet.Backups,
		client.QuotaStorageGigabytes:          u.QuotaSet.Gigabytes,
		client.QuotaStorageGroups:             u.QuotaSet.Groups,
		client.QuotaStoragePerVolumeGigabytes: u.QuotaSet.PerVolumeGigabytes,
		client.QuotaStorageSnapshots:          u.QuotaSet.Snapshots,
		client.QuotaStorageVolumes:            u.QuotaSet.Volumes,
	}
}

// UpdateStorageQuotasForAProject 스토리지 쿼타 생성
func UpdateStorageQuotasForAProject(c client.Client, in UpdateStorageQuotaRequest) (*UpdateStorageQuotaResponse, error) {
	// convert
	body, err := json.Marshal(&in)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	// request
	endpoint, err := c.GetEndpoint(ServiceTypeVolumeV3)
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
	requestLog("UpdateStorageQuotasForAProject", t, c, request, response, err)
	if err != nil {
		return nil, err
	}
	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}

	// convert
	var ret UpdateStorageQuotaResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}

// ShowStorageQuota Storage quota-set 조회
func ShowStorageQuota(c client.Client, req ShowStorageQuotaRequest) (*ShowStorageQuotaResponse, error) {
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
		URL:    endpoint + "/os-quota-sets/" + req.ProjectID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowStorageQuota", t, c, request, response, err)
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

	return &ShowStorageQuotaResponse{QuotaSets: quotaSets}, nil
}
