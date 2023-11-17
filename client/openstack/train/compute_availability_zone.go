package train

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"net/http"
	"time"
)

// ZoneState 가용구역 zoneState 구조체
type ZoneState struct {
	Available bool `json:"available,omitempty"`
}

// AvailabilityZoneResult 가용구역 Result
type AvailabilityZoneResult struct {
	ZoneName  string    `json:"zoneName,omitempty"`
	ZoneState ZoneState `json:"zoneState,omitempty"`
	Raw       string
}

// GetAvailabilityZoneInformationResponse 가용구역 목록 조회 response
type GetAvailabilityZoneInformationResponse struct {
	AvailabilityZones []AvailabilityZoneResult `json:"availabilityZoneInfo,omitempty"`
}

// GetAvailabilityZoneInformation 가용구역 목록 조회
func GetAvailabilityZoneInformation(c client.Client) (*GetAvailabilityZoneInformationResponse, error) {
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
		URL:    endpoint + "/os-availability-zone",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("GetAvailabilityZoneInformation", t, c, request, response, err)
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

	var rsp GetAvailabilityZoneInformationResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	for i, az := range rsp.AvailabilityZones {
		b, err := json.Marshal(az)
		if err != nil {
			return nil, errors.Unknown(err)
		}
		rsp.AvailabilityZones[i].Raw = string(b)
	}

	return &rsp, nil
}
