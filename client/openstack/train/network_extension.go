package train

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"net/http"
	"time"
)

// Extension 확장판
type Extension struct {
	Name        string   `json:"name,omitempty"`
	Alias       string   `json:"alias,omitempty"`
	Description string   `json:"description,omitempty"`
	Updated     string   `json:"updated,omitempty"`
	Links       []string `json:"links,omitempty"`
}

// ListExtensionsResponse 네트워크 확장판 목록 조회 결과
type ListExtensionsResponse struct {
	Extensions []Extension `json:"extensions,omitempty"`
}

// ListExtensions 네트워크 확장판 목록 조회
func ListExtensions(c client.Client) (*ListExtensionsResponse, error) {
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
		URL:    endpoint + "/v2.0/extensions",
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("ListExtensions", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 401:
			return nil, client.Unknown(response)
		case response.Code == 403:
			return nil, client.UnAuthorized(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}
	// convert
	var ret ListExtensionsResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}
