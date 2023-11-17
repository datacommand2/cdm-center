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

// ExtraRoute route
type ExtraRoute struct {
	DestinationCIDR string `json:"destination"`
	NextHopIP       string `json:"nexthop"`
}

// AddExtraRoutesToRouterRequest extra route 생성 요청
type AddExtraRoutesToRouterRequest struct {
	RouterUUID string       `json:"-"`
	Router     RouterResult `json:"router"`
}

// AddExtraRoutesToRouterResponse extra route 생성 응답
type AddExtraRoutesToRouterResponse struct {
	Router RouterResult `json:"router"`
}

// AddExtraRoutesToRouter 라우터에 extra route 추가
func AddExtraRoutesToRouter(c client.Client, in AddExtraRoutesToRouterRequest) (*AddExtraRoutesToRouterResponse, error) {
	// The extra routes configuration for L3 router.
	// A list of dictionaries with destination and nexthop parameters.
	// It is available when extraroute extension is enabled.
	body, err := json.Marshal(&in)
	if err != nil {
		return nil, errors.Unknown(err)
	}

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
		URL:    fmt.Sprintf("%s/v2.0/routers/%s/add_extraroutes", endpoint, in.RouterUUID),
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("AddExtraRoutesToRouter", t, c, request, response, err)
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
		case response.Code == 412:
			return nil, client.Unknown(response)
		case response.Code >= 500:
			return nil, client.RemoteServerError(response)
		case response.Code >= 400:
			return nil, client.Unknown(response)
		}
	}
	// convert
	var ret AddExtraRoutesToRouterResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual extra route deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}

// RemoveExtraRoutesFromRouterRequest extra route 제거 요청
type RemoveExtraRoutesFromRouterRequest struct {
	RouterUUID string       `json:"-"`
	Router     RouterResult `json:"router"`
}

// RemoveExtraRoutesFromRouter extra route 제거
func RemoveExtraRoutesFromRouter(c client.Client, in RemoveExtraRoutesFromRouterRequest) error {
	body, err := json.Marshal(&in)
	if err != nil {
		return errors.Unknown(err)
	}

	endpoint, err := c.GetEndpoint(ServiceTypeNetwork)
	if err != nil {
		return err
	}
	request := util.HTTPRequest{
		Header: http.Header{
			"Content-Type":   []string{"application/json"},
			HeaderXAuthToken: []string{c.GetTokenKey()},
		},
		Method: http.MethodPut,
		URL:    fmt.Sprintf("%s/v2.0/routers/%s/remove_extraroutes", endpoint, in.RouterUUID),
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("RemoveExtraRoutesFromRouter", t, c, request, response, err)
	if err != nil {
		return err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			return client.Unknown(response)
		case response.Code == 401:
			return client.Unknown(response)
		case response.Code == 403:
			return client.UnAuthorized(response)
		case response.Code == 404:
			logger.Warnf("Could not delete extra route. Cause: not found extra route")
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
