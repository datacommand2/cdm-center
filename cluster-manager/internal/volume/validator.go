package volume

import (
	"context"
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-center/cluster-manager/storage"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
	"unicode/utf8"
)

func validateVolumeRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterVolumeRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterVolumeId() == 0 {
		return errors.RequiredParameter("cluster_volume_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateVolumeByUUIDRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterVolumeByUUIDRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetUuid() == "" {
		return errors.RequiredParameter("uuid")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateVolumeListRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterVolumeListRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if utf8.RuneCountInString(req.Name) > 255 {
		return errors.LengthOverflowParameterValue("name", req.Name, 255)
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateCreateClusterVolumeRequest(req *cms.CreateClusterVolumeRequest) error {
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

	// storage
	if req.GetStorage() == nil {
		return errors.RequiredParameter("storage")
	}

	storageName := req.GetStorage().GetName()
	if storageName == "" {
		return errors.RequiredParameter("storage.name")
	}
	if utf8.RuneCountInString(storageName) > 255 {
		return errors.LengthOverflowParameterValue("storage.name", storageName, 255)
	}

	storageTypeCode := req.GetStorage().GetTypeCode()
	if !storage.IsClusterStorageTypeCode(storageTypeCode) {
		return errors.UnavailableParameterValue("storage.type_code", storageTypeCode, storage.ClusterStorageTypeCodes)
	}

	// volume
	if req.GetVolume() == nil {
		return errors.RequiredParameter("volume")
	}

	volumeName := req.GetVolume().GetName()
	if utf8.RuneCountInString(volumeName) > 255 {
		return errors.LengthOverflowParameterValue("volume.name", volumeName, 255)
	}

	volumeDescription := req.GetVolume().GetDescription()
	if utf8.RuneCountInString(volumeDescription) > 255 {
		return errors.LengthOverflowParameterValue("volume.description", volumeDescription, 255)
	}

	volumeSize := req.GetVolume().GetSizeBytes()
	if volumeSize < 1024*1024*1024 {
		return errors.OutOfRangeParameterValue("volume.size_bytes", volumeSize, 1073741824, nil)
	}

	return nil
}

func validateImportClusterVolumeRequest(req *cms.ImportClusterVolumeRequest) error {
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

	// source storage
	if req.GetSourceStorage() == nil {
		return errors.RequiredParameter("source_storage")
	}

	sourceStorageID := req.GetSourceStorage().GetId()
	if sourceStorageID == 0 {
		return errors.RequiredParameter("source_storage.id")
	}

	sourceStorageName := req.GetSourceStorage().GetName()
	if sourceStorageName == "" {
		return errors.RequiredParameter("source_storage.name")
	}
	if utf8.RuneCountInString(sourceStorageName) > 255 {
		return errors.LengthOverflowParameterValue("source_storage.name", sourceStorageName, 255)
	}

	sourceStorageUUID := req.GetSourceStorage().GetUuid()
	if sourceStorageUUID == "" {
		return errors.RequiredParameter("source_storage.uuid")
	}
	if _, err := uuid.Parse(sourceStorageUUID); err != nil {
		return errors.FormatMismatchParameterValue("source_storage.uuid", sourceStorageUUID, "UUID")
	}

	sourceStorageTypeCode := req.GetSourceStorage().GetTypeCode()
	if !storage.IsClusterStorageTypeCode(sourceStorageTypeCode) {
		return errors.UnavailableParameterValue("source_storage.type_code", sourceStorageTypeCode, storage.ClusterStorageTypeCodes)
	}

	// target storage
	if req.GetTargetStorage() == nil {
		return errors.RequiredParameter("target_storage")
	}

	targetStorageID := req.GetTargetStorage().GetId()
	if targetStorageID == 0 {
		return errors.RequiredParameter("target_storage.id")
	}

	targetStorageName := req.GetTargetStorage().GetName()
	if targetStorageName == "" {
		return errors.RequiredParameter("target_storage.name")
	}
	if utf8.RuneCountInString(targetStorageName) > 255 {
		return errors.LengthOverflowParameterValue("target_storage.name", targetStorageName, 255)
	}

	targetStorageUUID := req.GetTargetStorage().GetUuid()
	if targetStorageUUID == "" {
		return errors.RequiredParameter("target_storage.uuid")
	}
	if _, err := uuid.Parse(targetStorageUUID); err != nil {
		return errors.FormatMismatchParameterValue("target_storage.uuid", targetStorageUUID, "UUID")
	}

	targetStorageTypeCode := req.GetTargetStorage().GetTypeCode()
	if !storage.IsClusterStorageTypeCode(targetStorageTypeCode) {
		return errors.UnavailableParameterValue("target_storage.type_code", targetStorageTypeCode, storage.ClusterStorageTypeCodes)
	}

	// volume
	if req.GetVolume() == nil {
		return errors.RequiredParameter("volume")
	}

	volumeID := req.GetVolume().GetId()
	if volumeID == 0 {
		return errors.RequiredParameter("volume.id")
	}
	volumeUUID := req.GetVolume().GetUuid()
	if volumeUUID == "" {
		return errors.RequiredParameter("volume.uuid")
	}

	slice := strings.Split(volumeUUID, "@")
	if len(slice) > 2 {
		return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
	}

	if _, err := uuid.Parse(slice[0]); err != nil {
		return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
	}

	if len(slice) == 2 {
		_, err := time.Parse("snap-2006-01-02-15:04:05", slice[1])
		if err != nil {
			return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
		}
	}

	volumeName := req.GetVolume().GetName()
	if utf8.RuneCountInString(volumeName) > 255 {
		return errors.LengthOverflowParameterValue("volume.name", volumeName, 255)
	}

	volumeDescription := req.GetVolume().GetDescription()
	if utf8.RuneCountInString(volumeDescription) > 255 {
		return errors.LengthOverflowParameterValue("volume.description", volumeDescription, 255)
	}

	volumeSize := req.GetVolume().GetSizeBytes()
	if volumeSize < 1024*1024*1024 {
		return errors.OutOfRangeParameterValue("volume.size_bytes", volumeSize, 1073741824, nil)
	}

	for i, snapshot := range req.GetSnapshots() {
		snapshotUUID := snapshot.GetUuid()
		snapshotName := snapshot.GetName()
		snapshotDescription := snapshot.GetDescription()

		// snapshot
		if snapshotUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("snapshots[%d].uuid", i))
		}
		if _, err := uuid.Parse(snapshotUUID); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("snapshots[%d].uuid", i), snapshotUUID, "UUID")
		}

		if utf8.RuneCountInString(snapshotName) > 255 {
			return errors.LengthOverflowParameterValue(fmt.Sprintf("snapshots[%d].name", i), snapshotName, 255)
		}
		if utf8.RuneCountInString(snapshotDescription) > 255 {
			return errors.LengthOverflowParameterValue(fmt.Sprintf("snapshots[%d].description", i), snapshotDescription, 255)
		}
	}

	return nil
}

func validateDeleteClusterVolumeRequest(req *cms.DeleteClusterVolumeRequest) error {
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

	if req.GetVolume() == nil {
		return errors.RequiredParameter("volume")
	}
	volumeUUID := req.GetVolume().GetUuid()
	if volumeUUID == "" {
		return errors.RequiredParameter("volume.uuid")
	}

	slice := strings.Split(volumeUUID, "@")
	if len(slice) > 2 {
		return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
	}

	if _, err := uuid.Parse(slice[0]); err != nil {
		return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
	}

	if len(slice) == 2 {
		_, err := time.Parse("snap-2006-01-02-15:04:05", slice[1])
		if err != nil {
			return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
		}
	}

	return nil
}

func validateCreateClusterVolumeSnapshotRequest(req *cms.CreateClusterVolumeSnapshotRequest) error {
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

	//volume
	if req.GetVolume() == nil {
		return errors.RequiredParameter("volume")
	}

	volumeUUID := req.GetVolume().GetUuid()
	if volumeUUID == "" {
		return errors.RequiredParameter("volume.uuid")
	}

	slice := strings.Split(volumeUUID, "@")
	if len(slice) > 2 {
		return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
	}

	if _, err := uuid.Parse(slice[0]); err != nil {
		return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
	}

	if len(slice) == 2 {
		_, err := time.Parse("snap-2006-01-02-15:04:05", slice[1])
		if err != nil {
			return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
		}
	}

	// snapshot
	if req.GetSnapshot() == nil {
		return errors.RequiredParameter("snapshot")
	}

	snapshotName := req.GetSnapshot().GetName()
	snapshotDescription := req.GetSnapshot().GetDescription()
	if utf8.RuneCountInString(snapshotName) > 255 {
		return errors.LengthOverflowParameterValue("snapshot.name", snapshotName, 255)
	}
	if utf8.RuneCountInString(snapshotDescription) > 255 {
		return errors.LengthOverflowParameterValue("snapshot.description", snapshotDescription, 255)
	}

	return nil
}

func validateDeleteClusterVolumeSnapshotRequest(req *cms.DeleteClusterVolumeSnapshotRequest) error {
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

	if req.GetSnapshot() == nil {
		return errors.RequiredParameter("snapshot")
	}
	volumeSnapshotUUID := req.GetSnapshot().GetUuid()
	if volumeSnapshotUUID == "" {
		return errors.RequiredParameter("snapshot.uuid")
	}
	if _, err := uuid.Parse(volumeSnapshotUUID); err != nil {
		return errors.FormatMismatchParameterValue("snapshot.uuid", volumeSnapshotUUID, "UUID")
	}

	return nil
}

func validateUnmanageClusterVolumeRequest(req *cms.UnmanageClusterVolumeRequest) error {
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

	// source storage
	if req.GetSourceStorage() == nil {
		return errors.RequiredParameter("source_storage")
	}

	sourceStorageID := req.GetSourceStorage().GetId()
	if sourceStorageID == 0 {
		return errors.RequiredParameter("source_storage.id")
	}

	sourceStorageName := req.GetSourceStorage().GetName()
	if sourceStorageName == "" {
		return errors.RequiredParameter("source_storage.name")
	}
	if utf8.RuneCountInString(sourceStorageName) > 255 {
		return errors.LengthOverflowParameterValue("source_storage.name", sourceStorageName, 255)
	}

	sourceStorageUUID := req.GetSourceStorage().GetUuid()
	if sourceStorageUUID == "" {
		return errors.RequiredParameter("source_storage.uuid")
	}
	if _, err := uuid.Parse(sourceStorageUUID); err != nil {
		return errors.FormatMismatchParameterValue("source_storage.uuid", sourceStorageUUID, "UUID")
	}

	sourceStorageTypeCode := req.GetSourceStorage().GetTypeCode()
	if !storage.IsClusterStorageTypeCode(sourceStorageTypeCode) {
		return errors.UnavailableParameterValue("source_storage.type_code", sourceStorageTypeCode, storage.ClusterStorageTypeCodes)
	}

	// target storage
	if req.GetTargetStorage() == nil {
		return errors.RequiredParameter("target_storage")
	}

	targetStorageID := req.GetTargetStorage().GetId()
	if targetStorageID == 0 {
		return errors.RequiredParameter("target_storage.id")
	}

	targetStorageName := req.GetTargetStorage().GetName()
	if targetStorageName == "" {
		return errors.RequiredParameter("target_storage.name")
	}
	if utf8.RuneCountInString(targetStorageName) > 255 {
		return errors.LengthOverflowParameterValue("target_storage.name", targetStorageName, 255)
	}

	targetStorageUUID := req.GetTargetStorage().GetUuid()
	if targetStorageUUID == "" {
		return errors.RequiredParameter("target_storage.uuid")
	}
	if _, err := uuid.Parse(targetStorageUUID); err != nil {
		return errors.FormatMismatchParameterValue("target_storage.uuid", targetStorageUUID, "UUID")
	}

	targetStorageTypeCode := req.GetTargetStorage().GetTypeCode()
	if !storage.IsClusterStorageTypeCode(targetStorageTypeCode) {
		return errors.UnavailableParameterValue("target_storage.type_code", targetStorageTypeCode, storage.ClusterStorageTypeCodes)
	}

	// volume
	if req.GetVolumePair() == nil {
		return errors.RequiredParameter("volume_pair")
	}

	if req.GetVolumePair().GetSource() == nil {
		return errors.RequiredParameter("volume_pair.source")
	}

	volumePairSourceUUID := req.GetVolumePair().GetSource().GetUuid()
	if volumePairSourceUUID == "" {
		return errors.RequiredParameter("volume_pair.source.uuid")
	}

	slice := strings.Split(volumePairSourceUUID, "@")
	if len(slice) > 2 {
		return errors.FormatMismatchParameterValue("volume.uuid", volumePairSourceUUID, "UUID")
	}

	if _, err := uuid.Parse(slice[0]); err != nil {
		return errors.FormatMismatchParameterValue("volume.uuid", volumePairSourceUUID, "UUID")
	}

	if len(slice) == 2 {
		_, err := time.Parse("snap-2006-01-02-15:04:05", slice[1])
		if err != nil {
			return errors.FormatMismatchParameterValue("volume.uuid", volumePairSourceUUID, "UUID")
		}
	}

	volumeID := req.GetVolumeId()
	if volumeID == 0 {
		return errors.RequiredParameter("volume_id")
	}

	if req.GetVolumePair().GetTarget() == nil {
		return errors.RequiredParameter("volume_pair.target")
	}

	volumePairTargetUUID := req.GetVolumePair().GetTarget().GetUuid()
	if volumePairTargetUUID == "" {
		return errors.RequiredParameter("volume_pair.target.uuid")
	}

	slice = strings.Split(volumePairTargetUUID, "@")
	if len(slice) > 2 {
		return errors.FormatMismatchParameterValue("volume.uuid", volumePairTargetUUID, "UUID")
	}

	if _, err := uuid.Parse(slice[0]); err != nil {
		return errors.FormatMismatchParameterValue("volume.uuid", volumePairTargetUUID, "UUID")
	}

	if len(slice) == 2 {
		_, err := time.Parse("snap-2006-01-02-15:04:05", slice[1])
		if err != nil {
			return errors.FormatMismatchParameterValue("volume.uuid", volumePairTargetUUID, "UUID")
		}
	}

	// snapshots
	for i, snpPair := range req.GetSnapshotPairs() {
		if snpPair.GetSource() == nil {
			return errors.RequiredParameter(fmt.Sprintf("snapshot_pairs[%d].source", i))
		}

		snpSrcUUID := snpPair.GetSource().GetUuid()
		if snpSrcUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("snapshot_pairs[%d].source.uuid", i))
		}
		if _, err := uuid.Parse(snpSrcUUID); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("snapshot_pairs[%d].source.uuid", i), snpSrcUUID, "UUID")
		}

		if snpPair.GetTarget() == nil {
			return errors.RequiredParameter(fmt.Sprintf("snapshot_pairs[%d].target", i))
		}

		snpTgtUUID := snpPair.GetTarget().GetUuid()
		if snpTgtUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("snapshot_pairs[%d].target.uuid", i))
		}
		if _, err := uuid.Parse(snpTgtUUID); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("snapshot_pairs[%d].target.uuid", i), snpTgtUUID, "UUID")
		}
	}

	return nil
}

func validateCopyClusterVolumeRequest(req *cms.CopyClusterVolumeRequest) error {
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

	// source storage
	if req.GetSourceStorage() == nil {
		return errors.RequiredParameter("source_storage")
	}

	sourceStorageID := req.GetSourceStorage().GetId()
	if sourceStorageID == 0 {
		return errors.RequiredParameter("source_storage.id")
	}

	sourceStorageName := req.GetSourceStorage().GetName()
	if sourceStorageName == "" {
		return errors.RequiredParameter("source_storage.name")
	}
	if utf8.RuneCountInString(sourceStorageName) > 255 {
		return errors.LengthOverflowParameterValue("source_storage.name", sourceStorageName, 255)
	}

	sourceStorageUUID := req.GetSourceStorage().GetUuid()
	if sourceStorageUUID == "" {
		return errors.RequiredParameter("source_storage.uuid")
	}
	if _, err := uuid.Parse(sourceStorageUUID); err != nil {
		return errors.FormatMismatchParameterValue("source_storage.uuid", sourceStorageUUID, "UUID")
	}

	sourceStorageTypeCode := req.GetSourceStorage().GetTypeCode()
	if !storage.IsClusterStorageTypeCode(sourceStorageTypeCode) {
		return errors.UnavailableParameterValue("source_storage.type_code", sourceStorageTypeCode, storage.ClusterStorageTypeCodes)
	}

	// target storage
	if req.GetTargetStorage() == nil {
		return errors.RequiredParameter("target_storage")
	}

	targetStorageID := req.GetTargetStorage().GetId()
	if targetStorageID == 0 {
		return errors.RequiredParameter("target_storage.id")
	}

	targetStorageName := req.GetTargetStorage().GetName()
	if targetStorageName == "" {
		return errors.RequiredParameter("target_storage.name")
	}
	if utf8.RuneCountInString(targetStorageName) > 255 {
		return errors.LengthOverflowParameterValue("target_storage.name", targetStorageName, 255)
	}

	targetStorageUUID := req.GetTargetStorage().GetUuid()
	if targetStorageUUID == "" {
		return errors.RequiredParameter("target_storage.uuid")
	}
	if _, err := uuid.Parse(targetStorageUUID); err != nil {
		return errors.FormatMismatchParameterValue("target_storage.uuid", targetStorageUUID, "UUID")
	}

	targetStorageTypeCode := req.GetTargetStorage().GetTypeCode()
	if !storage.IsClusterStorageTypeCode(targetStorageTypeCode) {
		return errors.UnavailableParameterValue("target_storage.type_code", targetStorageTypeCode, storage.ClusterStorageTypeCodes)
	}

	// volume
	if req.GetVolume() == nil {
		return errors.RequiredParameter("volume")
	}

	volumeID := req.GetVolume().GetId()
	if volumeID == 0 {
		return errors.RequiredParameter("volume.id")
	}
	volumeUUID := req.GetVolume().GetUuid()
	if volumeUUID == "" {
		return errors.RequiredParameter("volume.uuid")
	}

	slice := strings.Split(volumeUUID, "@")
	if len(slice) > 2 {
		return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
	}

	if _, err := uuid.Parse(slice[0]); err != nil {
		return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
	}

	if len(slice) == 2 {
		_, err := time.Parse("snap-2006-01-02-15:04:05", slice[1])
		if err != nil {
			return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
		}
	}

	volumeName := req.GetVolume().GetName()
	if utf8.RuneCountInString(volumeName) > 255 {
		return errors.LengthOverflowParameterValue("volume.name", volumeName, 255)
	}

	volumeDescription := req.GetVolume().GetDescription()
	if utf8.RuneCountInString(volumeDescription) > 255 {
		return errors.LengthOverflowParameterValue("volume.description", volumeDescription, 255)
	}

	volumeSize := req.GetVolume().GetSizeBytes()
	if volumeSize < 1024*1024*1024 {
		return errors.OutOfRangeParameterValue("volume.size_bytes", volumeSize, 1073741824, nil)
	}

	for i, snapshot := range req.GetSnapshots() {
		snapshotUUID := snapshot.GetUuid()
		snapshotName := snapshot.GetName()
		snapshotDescription := snapshot.GetDescription()

		// snapshot
		if snapshotUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("snapshots[%d].uuid", i))
		}
		if _, err := uuid.Parse(snapshotUUID); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("snapshots[%d].uuid", i), snapshotUUID, "UUID")
		}

		if utf8.RuneCountInString(snapshotName) > 255 {
			return errors.LengthOverflowParameterValue(fmt.Sprintf("snapshots[%d].name", i), snapshotName, 255)
		}
		if utf8.RuneCountInString(snapshotDescription) > 255 {
			return errors.LengthOverflowParameterValue(fmt.Sprintf("snapshots[%d].description", i), snapshotDescription, 255)
		}
	}

	return nil
}

func validateDeleteClusterVolumeCopyRequest(req *cms.DeleteClusterVolumeCopyRequest) error {
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

	// source storage
	if req.GetSourceStorage() == nil {
		return errors.RequiredParameter("source_storage")
	}

	sourceStorageID := req.GetSourceStorage().GetId()
	if sourceStorageID == 0 {
		return errors.RequiredParameter("source_storage.id")
	}

	sourceStorageName := req.GetSourceStorage().GetName()
	if sourceStorageName == "" {
		return errors.RequiredParameter("source_storage.name")
	}
	if utf8.RuneCountInString(sourceStorageName) > 255 {
		return errors.LengthOverflowParameterValue("source_storage.name", sourceStorageName, 255)
	}

	sourceStorageUUID := req.GetSourceStorage().GetUuid()
	if sourceStorageUUID == "" {
		return errors.RequiredParameter("source_storage.uuid")
	}
	if _, err := uuid.Parse(sourceStorageUUID); err != nil {
		return errors.FormatMismatchParameterValue("source_storage.uuid", sourceStorageUUID, "UUID")
	}

	sourceStorageTypeCode := req.GetSourceStorage().GetTypeCode()
	if !storage.IsClusterStorageTypeCode(sourceStorageTypeCode) {
		return errors.UnavailableParameterValue("source_storage.type_code", sourceStorageTypeCode, storage.ClusterStorageTypeCodes)
	}

	// target storage
	if req.GetTargetStorage() == nil {
		return errors.RequiredParameter("target_storage")
	}

	targetStorageID := req.GetTargetStorage().GetId()
	if targetStorageID == 0 {
		return errors.RequiredParameter("target_storage.id")
	}
	targetStorageName := req.GetTargetStorage().GetName()
	if targetStorageName == "" {
		return errors.RequiredParameter("target_storage.name")
	}
	if utf8.RuneCountInString(targetStorageName) > 255 {
		return errors.LengthOverflowParameterValue("target_storage.name", targetStorageName, 255)
	}

	targetStorageUUID := req.GetTargetStorage().GetUuid()
	if targetStorageUUID == "" {
		return errors.RequiredParameter("target_storage.uuid")
	}
	if _, err := uuid.Parse(targetStorageUUID); err != nil {
		return errors.FormatMismatchParameterValue("target_storage.uuid", targetStorageUUID, "UUID")
	}

	targetStorageTypeCode := req.GetTargetStorage().GetTypeCode()
	if !storage.IsClusterStorageTypeCode(targetStorageTypeCode) {
		return errors.UnavailableParameterValue("target_storage.type_code", targetStorageTypeCode, storage.ClusterStorageTypeCodes)
	}

	// volume
	volumeID := req.GetVolumeId()
	if volumeID == 0 {
		return errors.RequiredParameter("volume_id")
	}

	if req.GetVolume() == nil {
		return errors.RequiredParameter("volume")
	}

	volumeUUID := req.GetVolume().GetUuid()
	if volumeUUID == "" {
		return errors.RequiredParameter("volume.uuid")
	}
	slice := strings.Split(volumeUUID, "@")
	if len(slice) > 2 {
		return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
	}

	if _, err := uuid.Parse(slice[0]); err != nil {
		return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
	}

	if len(slice) == 2 {
		_, err := time.Parse("snap-2006-01-02-15:04:05", slice[1])
		if err != nil {
			return errors.FormatMismatchParameterValue("volume.uuid", volumeUUID, "UUID")
		}
	}

	for i, snapshot := range req.GetSnapshots() {
		snapshotUUID := snapshot.GetUuid()
		// snapshot
		if snapshotUUID == "" {
			return errors.RequiredParameter(fmt.Sprintf("snapshots[%d].uuid", i))
		}
		if _, err := uuid.Parse(snapshotUUID); err != nil {
			return errors.FormatMismatchParameterValue(fmt.Sprintf("snapshots[%d].uuid", i), snapshotUUID, "UUID")
		}
	}

	return nil
}
