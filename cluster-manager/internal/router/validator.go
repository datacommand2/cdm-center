package router

import (
	"context"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

func validateCreateClusterRouterRequest(req *cms.CreateClusterRouterRequest) error {
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

	// router
	if req.GetRouter() == nil {
		return errors.RequiredParameter("router")
	}

	state := req.GetRouter().GetState()
	if state != "" && state != "up" && state != "down" {
		return errors.UnavailableParameterValue("router.state", state, []interface{}{"up", "down"})
	}

	routerName := req.GetRouter().GetName()
	if utf8.RuneCountInString(routerName) > 255 {
		return errors.LengthOverflowParameterValue("router.name", routerName, 255)
	}

	routerDesc := req.GetRouter().GetDescription()
	if utf8.RuneCountInString(routerDesc) > 255 {
		return errors.LengthOverflowParameterValue("router.description", routerDesc, 255)
	}

	for i, extra := range req.GetRouter().GetExtraRoutes() {
		destination := extra.GetDestination()
		nexthop := extra.GetNexthop()

		if destination == "" {
			return errors.RequiredParameter(fmt.Sprintf("router.extra_routes[%d].destination", i))
		}
		if !govalidator.IsCIDR(destination) {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("router.extra_routes[%d].destination", i), destination, "CIDR")
		}

		if nexthop == "" {
			return errors.RequiredParameter(fmt.Sprintf("router.extra_routes[%d].nexthop", i))
		}
		if !govalidator.IsIP(nexthop) {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("router.extra_routes[%d].nexthop", i), nexthop, "IP")
		}
	}

	for i, routingInterface := range req.GetRouter().GetExternalRoutingInterfaces() {
		networkUUID := routingInterface.GetNetwork().GetUuid()
		address := routingInterface.GetIpAddress()
		subnetUUID := routingInterface.GetSubnet().GetUuid()

		if networkUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("router.external_routing_interfaces[%d].network.uuid", i))
		}
		if _, err := uuid.Parse(networkUUID); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("router.external_routing_interfaces[%d].network.uuid", i), networkUUID, "UUID")
		}

		if address != "" && !govalidator.IsIP(address) {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("router.external_routing_interfaces[%d].ip_address", i), address, "IP")
		}

		if subnetUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("router.external_routing_interfaces[%d].subnet.uuid", i))
		}
		if _, err := uuid.Parse(subnetUUID); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("router.external_routing_interfaces[%d].subnet.uuid", i), subnetUUID, "UUID")
		}
	}

	for i, routingInterface := range req.GetRouter().GetInternalRoutingInterfaces() {
		networkUUID := routingInterface.GetNetwork().GetUuid()
		address := routingInterface.GetIpAddress()
		subnetUUID := routingInterface.GetSubnet().GetUuid()

		if networkUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("router.internal_routing_interfaces[%d].network.uuid", i))
		}
		if _, err := uuid.Parse(networkUUID); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("router.internal_routing_interfaces[%d].network.uuid", i), networkUUID, "UUID")
		}

		if address != "" && !govalidator.IsIPv4(address) && !govalidator.IsIPv6(address) {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("router.internal_routing_interfaces[%d].ip_address", i), address, "IP")
		}

		if subnetUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("router.internal_routing_interfaces[%d].subnet.uuid", i))
		}
		if _, err := uuid.Parse(subnetUUID); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("router.internal_routing_interfaces[%d].subnet.uuid", i), subnetUUID, "UUID")
		}
	}

	return nil
}

func validateDeleteClusterRouterRequest(req *cms.DeleteClusterRouterRequest) error {
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

	if req.GetRouter() == nil {
		return errors.RequiredParameter("router")
	}

	routerUUID := req.GetRouter().GetUuid()
	if routerUUID == "" {
		return errors.RequiredParameter("router.uuid")
	}
	if _, err := uuid.Parse(routerUUID); err != nil {
		return errors.FormatMismatchParameterValue("router.uuid", routerUUID, "UUID")
	}

	return nil
}

func validateGetClusterRouterRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterRouterRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterRouterId() == 0 {
		return errors.RequiredParameter("cluster_router_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateGetClusterRouterByUUIDRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterRouterByUUIDRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetUuid() == "" {
		return errors.RequiredParameter("uuid")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateGetClusterRouterListRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterRouterListRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}
