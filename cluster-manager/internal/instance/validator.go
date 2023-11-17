package instance

import (
	"context"
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

func validateInstanceRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterInstanceRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterInstanceId() == 0 {
		return errors.RequiredParameter("cluster_instance_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateInstanceByUUIDRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterInstanceByUUIDRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetUuid() == "" {
		return errors.RequiredParameter("uuid")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateInstanceListRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterInstanceListRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if utf8.RuneCountInString(req.Name) > 255 {
		return errors.LengthOverflowParameterValue("name", req.Name, 255)
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateGetKeyPairRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterKeyPairRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterKeypairId() == 0 {
		return errors.RequiredParameter("cluster_keypair_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateCreateClusterInstanceRequest(req *cms.CreateClusterInstanceRequest) error {
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

	// hypervisor
	if req.GetHypervisor() == nil {
		return errors.RequiredParameter("hypervisor")
	}

	hypervisorHostname := req.GetHypervisor().GetHostname()
	if utf8.RuneCountInString(hypervisorHostname) > 255 {
		return errors.LengthOverflowParameterValue("hypervisor.hostname", hypervisorHostname, 255)
	}

	// availabilityZone
	if req.GetAvailabilityZone() == nil {
		return errors.RequiredParameter("availability_zone")
	}

	availabilityZoneName := req.GetAvailabilityZone().GetName()
	if availabilityZoneName == "" {
		return errors.RequiredParameter("availability_zone.name")
	}
	if utf8.RuneCountInString(availabilityZoneName) > 255 {
		return errors.LengthOverflowParameterValue("availability_zone.name", availabilityZoneName, 255)
	}

	// Spec
	if req.GetSpec() == nil {
		return errors.RequiredParameter("spec")
	}

	specUUID := req.GetSpec().GetUuid()
	if specUUID == "" {
		return errors.RequiredParameter("spec.uuid")
	}
	if len(specUUID) > 255 {
		return errors.LengthOverflowParameterValue("spec.uuid", specUUID, 255)
	}

	// keypair
	keypairName := req.GetKeypair().GetName()
	if utf8.RuneCountInString(keypairName) > 255 {
		return errors.LengthOverflowParameterValue("keypair.name", keypairName, 255)
	}

	// Networks
	for i, n := range req.GetNetworks() {
		if n.GetNetwork().GetUuid() == "" {
			return errors.RequiredParameter(fmt.Sprintf("networks[%d].network.uuid", i))
		}

		if _, err := uuid.Parse(n.GetNetwork().GetUuid()); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("networks[%d].network.uuid", i), n.GetNetwork().GetUuid(), "UUID")
		}

		if n.GetSubnet().GetUuid() == "" {
			return errors.RequiredParameter(fmt.Sprintf("networks[%d].subnet.uuid", i))
		}

		if _, err := uuid.Parse(n.GetSubnet().GetUuid()); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("networks[%d].subnet.uuid", i), n.GetSubnet().GetUuid(), "UUID")
		}

		if n.GetFloatingIp().GetUuid() != "" {
			if _, err := uuid.Parse(n.GetFloatingIp().GetUuid()); err != nil {
				return errors.FormatMismatchParameterValue(fmt.Sprintf("networks[%d].floating_ip.uuid", i), n.GetFloatingIp().GetUuid(), "UUID")
			}
		}

		if n.IpAddress == "" {
			return errors.RequiredParameter(fmt.Sprintf("networks[%d].ip_address", i))
		}
	}

	// Volumes
	if len(req.Volumes) == 0 {
		return errors.RequiredParameter("volumes")
	}

	for i, v := range req.GetVolumes() {
		if v.GetVolume().GetStorage() == nil {
			return errors.RequiredParameter(fmt.Sprintf("volumes[%d].volume.storage", i))
		}

		if v.GetVolume().GetUuid() == "" {
			return errors.RequiredParameter(fmt.Sprintf("volumes[%d].volume.uuid", i))
		}

		if _, err := uuid.Parse(v.GetVolume().GetUuid()); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("volumes[%d].volume.uuid", i), v.GetVolume().GetUuid(), "UUID")
		}

		if v.DevicePath == "" {
			return errors.RequiredParameter(fmt.Sprintf("volumes[%d].device_path", i))
		}
	}

	// SecurityGroups
	for i, sg := range req.GetSecurityGroups() {
		if sg.Name == "" {
			return errors.RequiredParameter(fmt.Sprintf("security_groups[%d].name", i))
		}

		if utf8.RuneCountInString(sg.Name) > 255 {
			return errors.LengthOverflowParameterValue(fmt.Sprintf("security_groups[%d].name", i), sg.Name, 255)
		}
	}

	// Instance
	if req.GetInstance() == nil {
		return errors.RequiredParameter("instance")
	}

	instanceName := req.GetInstance().GetName()
	if instanceName == "" {
		return errors.RequiredParameter("instance.name")
	}

	if utf8.RuneCountInString(instanceName) > 255 {
		return errors.LengthOverflowParameterValue("instance.name", instanceName, 255)
	}

	instanceDescription := req.GetInstance().GetDescription()
	if utf8.RuneCountInString(instanceDescription) > 255 {
		return errors.LengthOverflowParameterValue("instance.description", instanceDescription, 255)
	}

	return nil
}

func validateDeleteClusterInstanceRequest(req *cms.DeleteClusterInstanceRequest) error {
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

	// instance
	if req.GetInstance() == nil {
		return errors.RequiredParameter("instance")
	}

	instanceUUID := req.GetInstance().GetUuid()
	if instanceUUID == "" {
		return errors.RequiredParameter("instance.uuid")
	}
	if _, err := uuid.Parse(instanceUUID); err != nil {
		return errors.FormatMismatchParameterValue("instance.uuid", instanceUUID, "UUID")
	}

	return nil
}

func validateStartClusterInstanceRequest(req *cms.StartClusterInstanceRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	tenantUUID := req.GetTenant().GetUuid()
	instanceUUID := req.GetInstance().GetUuid()

	// tenant
	if tenantUUID == "" {
		return errors.RequiredParameter("tenant.uuid")
	}
	if _, err := uuid.Parse(tenantUUID); err != nil {
		return errors.FormatMismatchParameterValue("tenant.uuid", tenantUUID, "UUID")
	}

	// instance
	if instanceUUID == "" {
		return errors.RequiredParameter("instance.uuid")
	}
	if _, err := uuid.Parse(instanceUUID); err != nil {
		return errors.FormatMismatchParameterValue("instance.uuid", instanceUUID, "UUID")
	}

	return nil
}

func validateStopClusterInstanceRequest(req *cms.StopClusterInstanceRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	tenantUUID := req.GetTenant().GetUuid()
	instanceUUID := req.GetInstance().GetUuid()

	// tenant
	if tenantUUID == "" {
		return errors.RequiredParameter("tenant.uuid")
	}
	if _, err := uuid.Parse(tenantUUID); err != nil {
		return errors.FormatMismatchParameterValue("tenant.uuid", tenantUUID, "UUID")
	}

	// instance
	if instanceUUID == "" {
		return errors.RequiredParameter("instance.uuid")
	}
	if _, err := uuid.Parse(instanceUUID); err != nil {
		return errors.FormatMismatchParameterValue("instance.uuid", instanceUUID, "UUID")
	}

	return nil
}
