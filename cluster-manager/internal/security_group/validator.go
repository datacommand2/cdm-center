package securitygroup

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

func validateGetSecurityGroup(ctx context.Context, db *gorm.DB, req *cms.ClusterSecurityGroupRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterSecurityGroupId() == 0 {
		return errors.RequiredParameter("cluster_security_group_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateGetSecurityGroupByUUID(ctx context.Context, db *gorm.DB, req *cms.ClusterSecurityGroupByUUIDRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetUuid() == "" {
		return errors.RequiredParameter("uuid")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateGetSecurityGroupList(ctx context.Context, db *gorm.DB, req *cms.ClusterSecurityGroupListRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateCreateSecurityGroup(req *cms.CreateClusterSecurityGroupRequest) error {
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

	// security group
	if req.GetSecurityGroup() == nil {
		return errors.RequiredParameter("security_group")
	}

	securityGroupName := req.GetSecurityGroup().GetName()
	if securityGroupName == "" {
		return errors.RequiredParameter("security_group.name")
	}

	if utf8.RuneCountInString(securityGroupName) > 255 {
		return errors.LengthOverflowParameterValue("security_group.name", securityGroupName, 255)
	}

	securityGroupDescription := req.GetSecurityGroup().GetDescription()
	if utf8.RuneCountInString(securityGroupDescription) > 255 {
		return errors.LengthOverflowParameterValue("security_group.description", securityGroupDescription, 255)
	}

	return nil
}

func validateCreateSecurityGroupRule(req *cms.CreateClusterSecurityGroupRuleRequest) error {
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

	// security group
	if req.GetSecurityGroup() == nil {
		return errors.RequiredParameter("security_group")
	}

	securityGroupUUID := req.GetSecurityGroup().GetUuid()
	if securityGroupUUID == "" {
		return errors.RequiredParameter("security_group.uuid")
	}
	if _, err := uuid.Parse(securityGroupUUID); err != nil {
		return errors.FormatMismatchParameterValue("security_group.uuid", securityGroupUUID, "UUID")
	}

	// security group rule
	if req.GetSecurityGroupRule() == nil {
		return errors.RequiredParameter("security_group_rule")
	}

	ruleDescription := req.GetSecurityGroupRule().GetDescription()
	if utf8.RuneCountInString(ruleDescription) > 255 {
		return errors.LengthOverflowParameterValue("security_group_rule.description", ruleDescription, 255)
	}

	ruleRemoteSecurityGroupUUID := req.GetSecurityGroupRule().GetRemoteSecurityGroup().GetUuid()
	ruleNetworkCIDR := req.GetSecurityGroupRule().GetNetworkCidr()
	if ruleRemoteSecurityGroupUUID != "" && ruleNetworkCIDR != "" {
		return errors.InvalidParameterValue("security_group_rule.network_cidr", ruleNetworkCIDR, "either the remote_security_group.uuid or network_cidr")
	}

	if ruleRemoteSecurityGroupUUID != "" {
		if _, err := uuid.Parse(ruleRemoteSecurityGroupUUID); err != nil {
			return errors.FormatMismatchParameterValue("security_group_rule.remote_security_group.uuid", ruleRemoteSecurityGroupUUID, "UUID")
		}
	}

	ruleEtherType := req.GetSecurityGroupRule().GetEtherType()
	if ruleEtherType != 4 && ruleEtherType != 6 {
		return errors.UnavailableParameterValue("security_group_rule.ether_type", ruleEtherType, []interface{}{4, 6})
	}

	if ruleNetworkCIDR != "" && !govalidator.IsCIDR(ruleNetworkCIDR) {
		return errors.FormatMismatchParameterValue("security_group_rule.network_cidr", ruleNetworkCIDR, "CIDR")
	}

	ruleDirection := req.GetSecurityGroupRule().GetDirection()
	if ruleDirection != "ingress" && ruleDirection != "egress" {
		return errors.UnavailableParameterValue("security_group_rule.direction", ruleDirection, []interface{}{"ingress", "egress"})
	}

	rulePortMax := req.GetSecurityGroupRule().GetPortRangeMax()
	rulePortMin := req.GetSecurityGroupRule().GetPortRangeMin()
	if rulePortMax != 0 || rulePortMin != 0 {
		if !(0 < rulePortMax && rulePortMax < 65536) {
			return errors.OutOfRangeParameterValue("security_group_rule.port_range_max", rulePortMax, 1, 65535)
		}
		if !(0 < rulePortMin && rulePortMin < 65536) {
			return errors.OutOfRangeParameterValue("security_group_rule.port_range_min", rulePortMin, 1, 65535)
		}
		if rulePortMax < rulePortMin {
			return errors.InvalidParameterValue("security_group_rule.port_range_min", rulePortMax, "port_range_max value is must equal larger than port_range_min value")
		}
	}

	protocols := []interface{}{
		"any", "ah", "dccp", "egp", "esp", "gre", "icmp", "icmpv6", "igmp", "ipip",
		"ipv6-encap", "ipv6-frag", "ipv6-icmp", "ipv6-nonxt", "ipv6-opts", "ipv6-route",
		"ospf", "pgm", "rsvp", "sctp", "tcp", "udp", "udplite", "vrrp",
	}

	ruleProtocol := req.GetSecurityGroupRule().GetProtocol()
	matched := false
	for _, p := range protocols {
		matched = matched || (ruleProtocol == p.(string))
	}
	if !matched && ruleProtocol != "" {
		return errors.UnavailableParameterValue("security_group_rule.protocol", ruleProtocol, protocols)
	}

	return nil
}

func validateDeleteSecurityGroup(req *cms.DeleteClusterSecurityGroupRequest) error {
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

	if req.GetSecurityGroup() == nil {
		return errors.RequiredParameter("security_group")
	}
	groupUUID := req.GetSecurityGroup().GetUuid()
	if groupUUID == "" {
		return errors.RequiredParameter("security_group.uuid")
	}
	if _, err := uuid.Parse(groupUUID); err != nil {
		return errors.FormatMismatchParameterValue("security_group.uuid", groupUUID, "UUID")
	}

	for i, rule := range req.GetSecurityGroup().GetRules() {
		ruleUUID := rule.GetUuid()
		if ruleUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("security_group.rules[%d].uuid", i))
		}
		if _, err := uuid.Parse(ruleUUID); err != nil {
			return errors.FormatMismatchParameterValue(
				fmt.Sprintf("security_group.rules[%d].uuid", i),
				ruleUUID, "UUID",
			)
		}
	}

	return nil
}

func validateCheckIsExistSecurityGroup(ctx context.Context, db *gorm.DB, req *cms.CheckIsExistByNameRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("tenant_id")
	}

	if req.GetName() == "" {
		return errors.RequiredParameter("name")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}
