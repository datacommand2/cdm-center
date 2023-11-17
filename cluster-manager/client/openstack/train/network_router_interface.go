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

// AddInterfaceToRouterRequest 라우터 인퍼테이스 생성 요청
type AddInterfaceToRouterRequest struct {
	RouterUUID string  `json:"-"`
	SubnetUUID *string `json:"subnet_id,omitempty"`
	PortUUID   *string `json:"port_id,omitempty"` // IP address 를 지정해야 한다면, 포트생성해야 함
}

// AddInterfaceToRouterResponse 라우터 인터페이스 생성 응답
type AddInterfaceToRouterResponse struct {
	UUID        string `json:"id"`
	NetworkUUID string `json:"network_id"`
	PortUUID    string `json:"port_id"`
	SubnetUUID  string `json:"subnet_id"`
	ProjectUUID string `json:"project_id"`
}

// AddInterfaceToRouter 라우터 인터페이스 생성
func AddInterfaceToRouter(c client.Client, in AddInterfaceToRouterRequest) (*AddInterfaceToRouterResponse, error) {
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
		URL:    fmt.Sprintf("%s/v2.0/routers/%s/add_router_interface", endpoint, in.RouterUUID),
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("AddInterfaceToRouter", t, c, request, response, err)
	if err != nil {
		return nil, err
	}

	if response.Code >= 400 {
		DeleteToken(c.GetTokenKey())
		switch {
		case response.Code == 400:
			// 한개의 내부 인터페이스는 하나의 서브넷만을 지정할 수 있음
			// 동일한 서브넷에 여러개의 IP를 주는 것도 불가함
			// {"NeutronError": {"message": "Bad router request: Cannot have multiple IPv4 subnets on router port.", "type": "BadRequest", "detail": ""}}
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
	var ret AddInterfaceToRouterResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual router interface deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}

// RemoveInterfaceFromRouterRequest 라우터 인터페이스 삭제 요청
type RemoveInterfaceFromRouterRequest struct {
	RouterUUID string `json:"-"`

	// port 혹은 subnet 중 하나를 지정하거나, 두개 다 지정 가능
	PortUUID   string `json:"port_id,omitempty"`
	SubnetUUID string `json:"subnet_id,omitempty"`
}

// RemoveInterfaceFromRouter 라우터 인터페이스 삭제
func RemoveInterfaceFromRouter(c client.Client, in RemoveInterfaceFromRouterRequest) error {
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
		URL:    fmt.Sprintf("%s/v2.0/routers/%s/remove_router_interface", endpoint, in.RouterUUID),
		Body:   body,
	}
	t := time.Now()
	response, err := request.Request()
	requestLog("RemoveInterfaceFromRouter", t, c, request, response, err)
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
			logger.Warnf("Could not delete router interface. Cause: not found router interface")
		case response.Code == 409:
			// extra route 에서 사용하는 interface 를 지우려는 경우
			return client.Conflict(response)
		case response.Code >= 500:
			return client.RemoteServerError(response)
		case response.Code >= 400:
			return client.Unknown(response)
		}
	}

	return nil
}
