package volumegroup

import (
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

func validateGetClusterVolumeGroupRequest(db *gorm.DB, req *cms.GetClusterVolumeGroupByUUIDRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster.id")
	}

	if req.GetClusterTenantId() == 0 {
		return errors.RequiredParameter("tenant.id")
	}

	if req.GetClusterVolumeGroupUuid() == "" {
		return errors.RequiredParameter("volume_group.uuid")
	}

	if err := db.Where(&model.ClusterTenant{ClusterID: req.ClusterId, ID: req.ClusterTenantId}).First(&model.ClusterTenant{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			internal.NotFoundClusterTenantByUUID(req.ClusterVolumeGroupUuid)
		}
		return errors.UnusableDatabase(err)
	}

	return nil
}

func validateCreateClusterVolumeGroupRequest(req *cms.CreateClusterVolumeGroupRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	tenantUUID := req.GetTenant().GetUuid()
	volumeGroupName := req.GetVolumeGroup().GetName()
	volumeGroupDescription := req.GetVolumeGroup().GetDescription()

	if tenantUUID == "" {
		return errors.RequiredParameter("tenant.uuid")
	}
	if _, err := uuid.Parse(tenantUUID); err != nil {
		return errors.FormatMismatchParameterValue("tenant.uuid", tenantUUID, "UUID")
	}

	if volumeGroupName == "" {
		return errors.RequiredParameter("volume_group.name")
	}
	if utf8.RuneCountInString(volumeGroupName) > 255 {
		return errors.LengthOverflowParameterValue("volume_group.name", volumeGroupName, 255)
	}

	if utf8.RuneCountInString(volumeGroupDescription) > 255 {
		return errors.LengthOverflowParameterValue("volume_group.description", volumeGroupDescription, 255)
	}

	if req.GetVolumeGroup().GetStorages() == nil {
		return errors.RequiredParameter("volume_group.storages")
	}

	for i, storage := range req.GetVolumeGroup().GetStorages() {
		storageUUID := storage.GetUuid()
		if storageUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("volume_group.storages[%d].uuid", i))
		}
		if _, err := uuid.Parse(storageUUID); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("volume_group.storages[%d].uuid", i), storageUUID, "UUID")
		}
	}

	return nil
}

func validateDeleteClusterVolumeGroupRequest(req *cms.DeleteClusterVolumeGroupRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	tenantUUID := req.GetTenant().GetUuid()
	volumeGroupUUID := req.GetVolumeGroup().GetUuid()

	if tenantUUID == "" {
		return errors.RequiredParameter("tenant.uuid")
	}
	if _, err := uuid.Parse(tenantUUID); err != nil {
		return errors.FormatMismatchParameterValue("tenant.uuid", tenantUUID, "UUID")
	}

	if volumeGroupUUID == "" {
		return errors.RequiredParameter("volume_group.uuid")
	}
	if _, err := uuid.Parse(volumeGroupUUID); err != nil {
		return errors.FormatMismatchParameterValue("volume_group.uuid", volumeGroupUUID, "UUID")
	}

	return nil
}

func validateUpdateClusterVolumeGroupRequest(req *cms.UpdateClusterVolumeGroupRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	tenantUUID := req.GetTenant().GetUuid()
	volumeGroupUUID := req.GetVolumeGroup().GetUuid()

	if tenantUUID == "" {
		return errors.RequiredParameter("tenant.uuid")
	}
	if _, err := uuid.Parse(tenantUUID); err != nil {
		return errors.FormatMismatchParameterValue("tenant.uuid", tenantUUID, "UUID")
	}

	if volumeGroupUUID == "" {
		return errors.RequiredParameter("volume_group.uuid")
	}
	if _, err := uuid.Parse(volumeGroupUUID); err != nil {
		return errors.FormatMismatchParameterValue("volume_group.uuid", volumeGroupUUID, "UUID")
	}

	for i, volume := range req.GetAddVolumes() {
		addVolumeUUID := volume.GetUuid()
		if addVolumeUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("add_volumes[%d].uuid", i))
		}
		if _, err := uuid.Parse(addVolumeUUID); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("add_volumes[%d].uuid", i), addVolumeUUID, "UUID")
		}
	}

	for i, volume := range req.GetDeleteVolumes() {
		delVolumeUUID := volume.GetUuid()
		if delVolumeUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("delete_volumes[%d].uuid", i))
		}
		if _, err := uuid.Parse(delVolumeUUID); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("delete_volumes[%d].uuid", i), delVolumeUUID, "UUID")
		}
	}

	return nil
}

func validateCreateClusterVolumeGroupSnapshotRequest(req *cms.CreateClusterVolumeGroupSnapshotRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	tenantUUID := req.GetTenant().GetUuid()
	volumeGroupUUID := req.GetVolumeGroup().GetUuid()
	snapshotName := req.GetVolumeGroupSnapshot().GetName()
	snapshotDescription := req.GetVolumeGroupSnapshot().GetDescription()

	if tenantUUID == "" {
		return errors.RequiredParameter("tenant.uuid")
	}
	if _, err := uuid.Parse(tenantUUID); err != nil {
		return errors.FormatMismatchParameterValue("tenant.uuid", tenantUUID, "UUID")
	}

	if volumeGroupUUID == "" {
		return errors.RequiredParameter("volume_group.uuid")
	}
	if _, err := uuid.Parse(volumeGroupUUID); err != nil {
		return errors.FormatMismatchParameterValue("volume_group.uuid", volumeGroupUUID, "UUID")
	}

	if utf8.RuneCountInString(snapshotName) > 255 {
		return errors.LengthOverflowParameterValue("volume_group_snapshot.name", snapshotName, 255)
	}
	if utf8.RuneCountInString(snapshotDescription) > 255 {
		return errors.LengthOverflowParameterValue("volume_group_snapshot.description", snapshotDescription, 255)
	}

	return nil
}

func validateDeleteClusterVolumeGroupSnapshotRequest(req *cms.DeleteClusterVolumeGroupSnapshotRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	tenantUUID := req.GetTenant().GetUuid()
	volumeGroupUUID := req.GetVolumeGroup().GetUuid()
	snapshotUUID := req.GetVolumeGroupSnapshot().GetUuid()

	if tenantUUID == "" {
		return errors.RequiredParameter("tenant.uuid")
	}
	if _, err := uuid.Parse(tenantUUID); err != nil {
		return errors.FormatMismatchParameterValue("tenant.uuid", tenantUUID, "UUID")
	}

	if volumeGroupUUID == "" {
		return errors.RequiredParameter("volume_group.uuid")
	}
	if _, err := uuid.Parse(volumeGroupUUID); err != nil {
		return errors.FormatMismatchParameterValue("volume_group.uuid", volumeGroupUUID, "UUID")
	}

	if snapshotUUID == "" {
		return errors.RequiredParameter("volume_group_snapshot.uuid")
	}
	if _, err := uuid.Parse(snapshotUUID); err != nil {
		return errors.FormatMismatchParameterValue("volume_group_snapshot.uuid", snapshotUUID, "UUID")
	}

	return nil
}

func validateGetClusterVolumeGroupSnapshotListRequest(req *cms.GetClusterVolumeGroupSnapshotListRequest) error {
	return cluster.CheckValidClusterConnectionInfo(req.GetConn())
}
