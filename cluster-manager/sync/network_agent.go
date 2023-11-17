package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack/train"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/logger"
)

// GetNetworkAgentServices Get Network Agent list
func GetNetworkAgentServices(c client.Client) ([]queue.Agent, error) {
	// get network agents info
	result, err := train.ListAllNetworkAgents(c)
	if err != nil {
		logger.Errorf("[GetNetworkAgentServices] Could not get all network agent list. Cause: %+v", err)
		return nil, err
	}

	var agents []queue.Agent
	if result.Agents != nil {
		for _, agent := range result.Agents {
			status := "unavailable"
			if agent.AdminStateUp && agent.Alive {
				status = "available"
			}

			agents = append(agents, queue.Agent{
				ID:                 agent.UUID,
				Type:               agent.AgentType,
				Binary:             agent.Binary,
				Host:               agent.Host,
				Status:             status,
				HeartbeatTimestamp: agent.HeartbeatTimestamp,
			})
		}
	}

	return agents, err
}
