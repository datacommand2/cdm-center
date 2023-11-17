package train

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"net/http"
	"time"
)

// VolumeAttachmentResult 볼륨 attachment Result 구조체
type VolumeAttachmentResult struct {
	ConnectionInfo map[string]interface{} `json:"connection_info,omitempty"`
}

// ShowAttachmentDetailsRequest 볼륨 attachment 상세 조회 request
type ShowAttachmentDetailsRequest struct {
	AttachmentID string
}

// ShowAttachmentDetailsResponse 볼륨 attachment 상세 조회 response
type ShowAttachmentDetailsResponse struct {
	VolumeAttachment VolumeAttachmentResult `json:"attachment,omitempty"`
}

// ShowAttachmentDetails 볼륨 attachment 상세 조회
func ShowAttachmentDetails(c client.Client, req ShowAttachmentDetailsRequest) (*ShowAttachmentDetailsResponse, error) {
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
		URL:    endpoint + "/attachments/" + req.AttachmentID,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ShowAttachmentDetails", t, c, request, response, err)
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

	var rsp ShowAttachmentDetailsResponse
	if err := json.Unmarshal([]byte(response.Body), &rsp); err != nil {
		return nil, errors.Unknown(err)
	}

	return &rsp, nil
}
