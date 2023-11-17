package floatingip

import (
	"context"
	"github.com/asaskevich/govalidator"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	"github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

func validateCreateClusterFloatingIPRequest(req *cms.CreateClusterFloatingIPRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	if req.GetTenant() == nil {
		return errors.RequiredParameter("tenant")
	}

	tenantUUID := req.GetTenant().GetUuid()
	if tenantUUID == "" {
		return errors.RequiredParameter("tenant.uuid")
	}
	if _, err := uuid.Parse(tenantUUID); err != nil {
		return errors.FormatMismatchParameterValue("tenant.uuid", tenantUUID, "UUID")
	}

	if req.GetNetwork() == nil {
		return errors.RequiredParameter("network")
	}

	networkUUID := req.GetNetwork().GetUuid()
	if networkUUID == "" {
		return errors.RequiredParameter("network.uuid")
	}
	if _, err := uuid.Parse(networkUUID); err != nil {
		return errors.FormatMismatchParameterValue("network.uuid", networkUUID, "UUID")
	}

	if req.GetFloatingIp() == nil {
		return errors.RequiredParameter("floating_ip")
	}

	description := req.GetFloatingIp().GetDescription()
	if utf8.RuneCountInString(description) > 255 {
		return errors.LengthOverflowParameterValue("floating_ip.description", description, 255)
	}

	address := req.GetFloatingIp().GetIpAddress()
	if address != "" && !govalidator.IsIPv4(address) && !govalidator.IsIPv6(address) {
		return errors.FormatMismatchParameterValue("floating_ip.ip_address", address, "IP")
	}

	return nil
}

func validateDeleteClusterFloatingIPRequest(req *cms.DeleteClusterFloatingIPRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	if req.GetTenant() == nil {
		return errors.RequiredParameter("tenant")
	}
	tenantUUID := req.GetTenant().GetUuid()
	if tenantUUID == "" {
		return errors.RequiredParameter("tenant.uuid")
	}
	if _, err := uuid.Parse(tenantUUID); err != nil {
		return errors.FormatMismatchParameterValue("tenant.uuid", tenantUUID, "UUID")
	}

	if req.GetFloatingIp() == nil {
		return errors.RequiredParameter("floating_ip")
	}

	floatingUUID := req.GetFloatingIp().GetUuid()
	if floatingUUID == "" {
		return errors.RequiredParameter("floating_ip.uuid")
	}
	if _, err := uuid.Parse(floatingUUID); err != nil {
		return errors.FormatMismatchParameterValue("floating_ip.uuid", floatingUUID, "UUID")
	}

	return nil
}

func validateGetClusterFloatingIP(ctx context.Context, db *gorm.DB, req *cms.ClusterFloatingIPRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterFloatingIpId() == 0 {
		return errors.RequiredParameter("cluster_floating_ip")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateCheckIsExistFloatingIP(ctx context.Context, db *gorm.DB, req *cms.CheckIsExistClusterFloatingIPRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterFloatingIpIpAddress() == "" {
		return errors.RequiredParameter("cluster_floating_ip_address")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}
