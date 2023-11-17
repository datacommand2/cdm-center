package network

import (
	"context"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

func validateNetworkRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterNetworkRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterNetworkId() == 0 {
		return errors.RequiredParameter("cluster_network_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateNetworkByUUIDRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterNetworkByUUIDRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetUuid() == "" {
		return errors.RequiredParameter("uuid")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateNetworkListRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterNetworkListRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if err := cluster.IsClusterOwner(ctx, db, req.ClusterId); err != nil {
		return err
	}

	if req.ClusterTenantId != 0 {
		err := db.First(&model.ClusterTenant{}, &model.ClusterTenant{ID: req.ClusterTenantId, ClusterID: req.ClusterId}).Error
		switch {
		case errors.Equal(err, gorm.ErrRecordNotFound):
			return errors.InvalidParameterValue("cluster_tenant_id", req.ClusterTenantId, "not found tenant")

		case err != nil:
			return errors.UnusableDatabase(err)
		}
	}

	return nil
}

func validateCreateClusterNetworkRequest(req *cms.CreateClusterNetworkRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	// tenant
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

	// network
	if req.GetNetwork() == nil {
		return errors.RequiredParameter("network")
	}

	networkName := req.GetNetwork().GetName()
	if utf8.RuneCountInString(networkName) > 255 {
		return errors.LengthOverflowParameterValue("network.name", networkName, 255)
	}

	networkDescription := req.GetNetwork().GetDescription()
	if utf8.RuneCountInString(networkDescription) > 255 {
		return errors.LengthOverflowParameterValue("network.description", networkDescription, 255)
	}

	return nil
}

func validateDeleteClusterNetworkRequest(req *cms.DeleteClusterNetworkRequest) error {
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

	return nil
}

func validateGetClusterSubnetRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterSubnetRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterSubnetId() == 0 {
		return errors.RequiredParameter("cluster_subnet_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateCreateClusterSubnetRequest(req *cms.CreateClusterSubnetRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	// tenant
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

	// network
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

	// subnet
	if req.GetSubnet() == nil {
		return errors.RequiredParameter("subnet")
	}

	subnetName := req.GetSubnet().GetName()
	if utf8.RuneCountInString(subnetName) > 255 {
		return errors.LengthOverflowParameterValue("subnet.name",
			subnetName, 255)
	}

	subnetDescription := req.GetSubnet().GetDescription()
	if utf8.RuneCountInString(subnetDescription) > 255 {
		return errors.LengthOverflowParameterValue("subnet.description",
			subnetDescription, 255)
	}

	subnetCIDR := req.GetSubnet().GetNetworkCidr()
	if subnetCIDR == "" {
		return errors.RequiredParameter("subnet.network_cidr")
	}
	if !govalidator.IsCIDR(subnetCIDR) {
		return errors.FormatMismatchParameterValue("subnet.network_cidr",
			subnetCIDR, "CIDR")
	}

	for i, pool := range req.GetSubnet().GetDhcpPools() {
		startAddress := pool.GetStartIpAddress()
		endAddress := pool.GetEndIpAddress()

		if startAddress == "" {
			return errors.RequiredParameter(fmt.Sprintf("subnet.dhcp_pools[%d].start_ip_address", i))
		}
		if !govalidator.IsIP(startAddress) {
			return errors.FormatMismatchParameterValue(
				fmt.Sprintf("subnet.dhcp_pools[%d].start_ip_address", i),
				startAddress, "IP")
		}

		if endAddress == "" {
			return errors.RequiredParameter(fmt.Sprintf("subnet.dhcp_pools[%d].end_ip_address", i))
		}
		if !govalidator.IsIP(endAddress) {
			return errors.FormatMismatchParameterValue(
				fmt.Sprintf("subnet.dhcp_pools[%d].end_ip_address", i),
				endAddress, "IP")
		}
	}

	for i, nameserver := range req.GetSubnet().GetNameservers() {
		server := nameserver.GetNameserver()
		if server == "" {
			return errors.RequiredParameter(fmt.Sprintf("subnet.nameservers[%d].nameserver", i))
		}
		if !govalidator.IsIP(server) {
			return errors.FormatMismatchParameterValue(
				fmt.Sprintf("subnet.nameservers[%d].nameserver", i),
				server, "IP")
		}
	}

	return nil
}

func validateDeleteClusterSubnetRequest(req *cms.DeleteClusterSubnetRequest) error {
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

	if req.GetSubnet() == nil {
		return errors.RequiredParameter("subnet")
	}

	subnetUUID := req.GetSubnet().GetUuid()
	if subnetUUID == "" {
		return errors.RequiredParameter("subnet.uuid")
	}
	if _, err := uuid.Parse(subnetUUID); err != nil {
		return errors.FormatMismatchParameterValue("subnet.uuid", subnetUUID, "UUID")
	}

	return nil
}
