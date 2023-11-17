package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"reflect"
)

type volumeSnapshotRecord struct {
	Snapshot model.ClusterVolumeSnapshot `gorm:"embedded"`
	Volume   model.ClusterVolume         `gorm:"embedded"`
}

func compareVolumeSnapshot(dbVolumeSnapshot *model.ClusterVolumeSnapshot, result *client.GetVolumeSnapshotResult) bool {
	if dbVolumeSnapshot.UUID != result.VolumeSnapshot.UUID {
		return false
	}

	if !reflect.DeepEqual(dbVolumeSnapshot.ClusterVolumeGroupSnapshotUUID, result.VolumeSnapshot.ClusterVolumeGroupSnapshotUUID) {
		return false
	}

	if dbVolumeSnapshot.Name != result.VolumeSnapshot.Name {
		return false
	}

	if dbVolumeSnapshot.SizeBytes != result.VolumeSnapshot.SizeBytes {
		return false
	}

	if dbVolumeSnapshot.Status != result.VolumeSnapshot.Status {
		return false
	}

	if dbVolumeSnapshot.CreatedAt != result.VolumeSnapshot.CreatedAt {
		return false
	}

	if !reflect.DeepEqual(dbVolumeSnapshot.Description, result.VolumeSnapshot.Description) {
		return false
	}

	var v model.ClusterVolume
	if err := database.Execute(func(db *gorm.DB) error {
		return db.Find(&v, &model.ClusterVolume{ID: dbVolumeSnapshot.ClusterVolumeID}).Error
	}); err != nil {
		return false
	}

	if v.UUID != result.Volume.UUID {
		return false
	}

	return true
}

func (s *Synchronizer) deleteVolumeSnapshot(volume *model.ClusterVolume, snapshot *model.ClusterVolumeSnapshot) (err error) {
	logger.Infof("[Sync-deleteVolumeSnapshot] Start: cluster(%d) volumeSnapshot(%s)", s.clusterID, snapshot.UUID)

	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(&snapshot).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteVolumeSnapshot] Could not delete cluster(%d) volumeSnapshot(%s). Cause: %+v",
			s.clusterID, snapshot.UUID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterVolumeSnapshotDeleted, &queue.DeleteClusterVolumeSnapshot{
		Cluster:            &model.Cluster{ID: s.clusterID},
		Volume:             volume,
		OrigVolumeSnapshot: snapshot,
	}); err != nil {
		logger.Warnf("[Sync-deleteVolumeSnapshot] Could not publish cluster(%d) volumeSnapshot(%s) deleted message. Cause: %+v",
			s.clusterID, snapshot.UUID, err)
	}

	logger.Infof("[Sync-deleteVolumeSnapshot] Success: cluster(%d) volumeSnapshot(%s)", s.clusterID, snapshot.UUID)
	return nil
}

func (s *Synchronizer) syncDeletedVolumeSnapshotList(rsp *client.GetVolumeSnapshotListResponse) error {
	var err error
	var volumeSnapshotRecordList []*volumeSnapshotRecord

	// 동기화되어있는 볼륨 스냅샷 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterVolumeSnapshot{}.TableName()).
			Select("*").
			Joins("join cdm_cluster_volume on cdm_cluster_volume_snapshot.cluster_volume_id = cdm_cluster_volume.id").
			Joins("join cdm_cluster_tenant on cdm_cluster_volume.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Find(&volumeSnapshotRecordList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 볼륨 스냅샷 제거
	for _, record := range volumeSnapshotRecordList {
		exists := false
		for _, clusterVolumeSnapshot := range rsp.ResultList {
			exists = exists || (clusterVolumeSnapshot.VolumeSnapshot.UUID == record.Snapshot.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteVolumeSnapshot(&record.Volume, &record.Snapshot); err != nil {
			return err
		}
	}

	return nil
}

// SyncClusterVolumeSnapshotList 볼륨 스냅샷 목록 동기화
func (s *Synchronizer) SyncClusterVolumeSnapshotList(opts ...Option) error {
	logger.Infof("[SyncClusterVolumeSnapshotList] Start: cluster(%d)", s.clusterID)

	// 클러스터의 볼륨 스냅샷 목록 조회
	rsp, err := s.Cli.GetVolumeSnapshotList()
	if err != nil {
		return err
	}

	// 클러스터에서 삭제된 볼륨 스냅샷 제거
	if err = s.syncDeletedVolumeSnapshotList(rsp); err != nil {
		return err
	}

	// 볼륨 스냅샷 상세 동기화
	for _, volumeSnapshot := range rsp.ResultList {
		if e := s.SyncClusterVolumeSnapshot(&model.ClusterVolumeSnapshot{UUID: volumeSnapshot.VolumeSnapshot.UUID}, opts...); e != nil {
			logger.Warnf("[SyncClusterVolumeSnapshotList] Failed to sync cluster(%d) volumeSnapshot(%s). Cause: %+v", s.clusterID, volumeSnapshot.VolumeSnapshot.UUID, e)
			err = e
		}
	}

	logger.Infof("[SyncClusterVolumeSnapshotList] Success: cluster(%d)", s.clusterID)
	return err
}

// SyncClusterVolumeSnapshot 볼륨 스냅샷 상세 동기화
func (s *Synchronizer) SyncClusterVolumeSnapshot(volumeSnapshot *model.ClusterVolumeSnapshot, opts ...Option) error {
	logger.Infof("[SyncClusterVolumeSnapshot] Start: cluster(%d) volumeSnapshot(%s)", s.clusterID, volumeSnapshot.UUID)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.TenantUUID != "" {
		if ok, err := s.isDeletedTenant(options.TenantUUID); ok {
			// 삭제된 tenant
			logger.Infof("[SyncClusterVolumeSnapshot] Already sync and deleted tenant: cluster(%d) volumeSnapshot(%s) tenant(%s)",
				s.clusterID, volumeSnapshot.UUID, options.TenantUUID)
			return nil
		} else if err != nil {
			logger.Errorf("[SyncClusterVolumeSnapshot] Could not get tenant from db: cluster(%d) volumeSnapshot(%s) tenant(%s). Cause: %+v",
				s.clusterID, volumeSnapshot.UUID, options.TenantUUID, err)
			return err
		}
		// 삭제안된 tenant 라면 동기화 계속 진행
	}

	rsp, err := s.Cli.GetVolumeSnapshot(client.GetVolumeSnapshotRequest{VolumeSnapshot: model.ClusterVolumeSnapshot{UUID: volumeSnapshot.UUID}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	var record volumeSnapshotRecord
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterVolumeSnapshot{}.TableName()).
			Select("*").
			Joins("join cdm_cluster_volume on cdm_cluster_volume_snapshot.cluster_volume_id = cdm_cluster_volume.id").
			Joins("join cdm_cluster_tenant on cdm_cluster_volume.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterVolumeSnapshot{UUID: volumeSnapshot.UUID}).
			First(&record).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	if err = copier.Copy(volumeSnapshot, &record.Snapshot); err != nil {
		return errors.Unknown(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		logger.Infof("[SyncClusterVolumeSnapshot] Success - nothing was updated: cluster(%d) volumeSnapshot(%s)", s.clusterID, volumeSnapshot.UUID)
		return nil

	case rsp == nil: // 볼륨 스냅샷 삭제
		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Delete(volumeSnapshot).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		if err = publishMessage(constant.QueueNoticeClusterVolumeSnapshotDeleted, &queue.DeleteClusterVolumeSnapshot{
			Cluster:            &model.Cluster{ID: s.clusterID},
			Volume:             &record.Volume,
			OrigVolumeSnapshot: &record.Snapshot,
		}); err != nil {
			logger.Warnf("[SyncClusterVolumeSnapshot] Could not publish cluster(%d) volumeSnapshot(%s) deleted message. Cause: %+v",
				s.clusterID, volumeSnapshot.UUID, err)
		}

	case err == gorm.ErrRecordNotFound: // 신규 볼륨 스냅샷이 추가됨
		var v *model.ClusterVolume
		if v, err = s.getVolume(&model.ClusterVolume{UUID: rsp.Result.Volume.UUID}, opts...); err != nil {
			return err
		}

		rsp.Result.VolumeSnapshot.ClusterVolumeID = v.ID
		*volumeSnapshot = rsp.Result.VolumeSnapshot

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(volumeSnapshot).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterVolumeSnapshot] Done - added: cluster(%d) volumeSnapshot(%s)", s.clusterID, volumeSnapshot.UUID)

		if err = publishMessage(constant.QueueNoticeClusterVolumeSnapshotCreated, &queue.CreateClusterVolumeSnapshot{
			Cluster:        &model.Cluster{ID: s.clusterID},
			Volume:         v,
			VolumeSnapshot: volumeSnapshot,
		}); err != nil {
			logger.Warnf("[SyncClusterVolumeSnapshot] Could not publish cluster(%d) volumeSnapshot(%s) created message. Cause: %+v",
				s.clusterID, volumeSnapshot.UUID, err)
		}

	case options.Force || !compareVolumeSnapshot(volumeSnapshot, &rsp.Result): // 기존 동기화되어있는 볼륨 스냅샷 정보가 변경됨
		var v *model.ClusterVolume
		if v, err = s.getVolume(&model.ClusterVolume{UUID: rsp.Result.Volume.UUID}, opts...); err != nil {
			return err
		}

		rsp.Result.VolumeSnapshot.ID = volumeSnapshot.ID
		rsp.Result.VolumeSnapshot.ClusterVolumeID = v.ID
		*volumeSnapshot = rsp.Result.VolumeSnapshot

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(volumeSnapshot).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterVolumeSnapshot] Done - updated: cluster(%d) volumeSnapshot(%s)", s.clusterID, volumeSnapshot.UUID)

		if err = publishMessage(constant.QueueNoticeClusterVolumeSnapshotUpdated, &queue.UpdateClusterVolumeSnapshot{
			Cluster:        &model.Cluster{ID: s.clusterID},
			Volume:         v,
			VolumeSnapshot: volumeSnapshot,
		}); err != nil {
			logger.Warnf("[SyncClusterVolumeSnapshot] Could not publish cluster(%d) volumeSnapshot(%s) updated message. Cause: %+v",
				s.clusterID, volumeSnapshot.UUID, err)
		}
	}

	logger.Infof("[SyncClusterVolumeSnapshot] Success: cluster(%d) volumeSnapshot(%s)", s.clusterID, volumeSnapshot.UUID)

	return nil
}
