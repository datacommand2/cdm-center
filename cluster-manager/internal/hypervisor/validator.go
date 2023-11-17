package hypervisor

import (
	"context"
	"github.com/asaskevich/govalidator"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

func validateHypervisorRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterHypervisorRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterHypervisorId() == 0 {
		return errors.RequiredParameter("cluster_hypervisor_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateHypervisorListRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterHypervisorListRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if utf8.RuneCountInString(req.Hostname) > 255 {
		return errors.LengthOverflowParameterValue("hostname", req.Hostname, 255)
	}

	if req.IpAddress != "" && !govalidator.IsIP(req.IpAddress) {
		return errors.FormatMismatchParameterValue("ip_address", req.IpAddress, "IP Address")
	}

	if err := cluster.IsClusterOwner(ctx, db, req.ClusterId); err != nil {
		return err
	}

	if req.ClusterAvailabilityZoneId != 0 {
		err := db.Where(&model.ClusterAvailabilityZone{
			ClusterID: req.ClusterId, ID: req.ClusterAvailabilityZoneId,
		}).First(&model.ClusterAvailabilityZone{}).Error

		switch {
		case errors.Equal(err, gorm.ErrRecordNotFound):
			return errors.InvalidParameterValue("cluster_availability_zone_id", req.ClusterAvailabilityZoneId, "not found availability zone")

		case err != nil:
			return errors.UnusableDatabase(err)
		}
	}

	return nil
}

func validateUpdateHypervisorRequest(ctx context.Context, db *gorm.DB, req *cms.UpdateClusterHypervisorRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterHypervisorId() == 0 {
		return errors.RequiredParameter("cluster_hypervisor_id")
	}

	if req.GetClusterHypervisorId() != req.GetHypervisor().GetId() {
		return errors.UnchangeableParameter("cluster.hypervisor.id")
	}

	return cluster.IsClusterOwner(ctx, db, req.GetClusterId())
}

func checkValidHypervisor(h *cms.ClusterHypervisor) error {
	SSHPort := h.GetSshPort()
	SSHAccount := h.GetSshAccount()
	SSHPassword := h.GetSshPassword()
	AgentPort := h.GetAgentPort()
	AgentVersion := h.GetAgentVersion()

	if SSHPort == 0 || SSHPort > 65535 {
		return errors.OutOfRangeParameterValue("ssh_port", SSHPort, 1, 65535)
	}

	if AgentPort == 0 || AgentPort > 65535 {
		return errors.OutOfRangeParameterValue("agent_port", AgentPort, 1, 65535)
	}

	if SSHAccount != "" && len(SSHAccount) > 255 {
		return errors.LengthOverflowParameterValue("ssh_account", SSHAccount, 255)
	}

	if SSHPassword != "" && len(SSHPassword) > 255 {
		return errors.LengthOverflowParameterValue("ssh_password", SSHPassword, 255)
	}

	if AgentVersion != "" && len(AgentVersion) > 10 {
		return errors.LengthOverflowParameterValue("agent_version", AgentVersion, 10)
	}

	return nil
}
