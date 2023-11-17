package train

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/util"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"time"

	"encoding/json"
	"net/http"
)

// AgentResult 구조체
type AgentResult struct {
	UUID               string  `json:"id,omitempty"`
	AgentType          string  `json:"agent_type,omitempty"`
	Binary             string  `json:"binary,omitempty"`
	Host               string  `json:"host,omitempty"`
	AdminStateUp       bool    `json:"admin_state_up,omitempty"`
	Description        *string `json:"description,omitempty"`
	Alive              bool    `json:"alive,omitempty"`
	HeartbeatTimestamp string  `json:"heartbeat_timestamp,omitempty"`
}

// ListAllNetworkAgentsResponse 네트워크 에이전트 목록 조회 결과
type ListAllNetworkAgentsResponse struct {
	Agents []AgentResult `json:"agents,omitempty"`
}

// ListAllNetworkAgents 네트워크 에이전트 목록 조회
func ListAllNetworkAgents(c client.Client) (*ListAllNetworkAgentsResponse, error) {
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
		URL:    endpoint + "/v2.0/agents",
	}

	t := time.Now()
	response, err := request.Request()
	requestLog("ListAllNetworkAgents", t, c, request, response, err)
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
	var ret ListAllNetworkAgentsResponse
	if err := json.Unmarshal([]byte(response.Body), &ret); err != nil {
		logger.Warnf("Manual network agent deletion is required. Cause: %+v", err)
		logger.Warnf(response.Body)
		return nil, errors.Unknown(err)
	}

	return &ret, nil
}
