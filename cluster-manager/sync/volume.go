package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	clusterVolume "github.com/datacommand2/cdm-center/cluster-manager/volume"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
	"reflect"
)

func compareVolume(dbVolume *model.ClusterVolume, result *client.GetVolumeResult) bool {
	if dbVolume.UUID != result.Volume.UUID {
		return false
	}

	if dbVolume.Name != result.Volume.Name {
		return false
	}

	if dbVolume.SizeBytes != result.Volume.SizeBytes {
		return false
	}

	if dbVolume.Multiattach != result.Volume.Multiattach {
		return false
	}

	if dbVolume.Bootable != result.Volume.Bootable {
		return false
	}

	if dbVolume.Readonly != result.Volume.Readonly {
		return false
	}

	if dbVolume.Status != result.Volume.Status {
		return false
	}

	if !reflect.DeepEqual(dbVolume.Description, result.Volume.Description) {
		return false
	}

	var t model.ClusterTenant
	var s model.ClusterStorage
	if err := database.Execute(func(db *gorm.DB) error {
		if err := db.Find(&t, &model.ClusterTenant{ID: dbVolume.ClusterTenantID}).Error; err != nil {
			return err
		}

		if err := db.Find(&s, &model.ClusterStorage{ID: dbVolume.ClusterStorageID}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return false
	}

	if t.UUID != result.Tenant.UUID {
		return false
	}

	if s.UUID != result.Storage.UUID {
		return false
	}

	return true
}

// getVolume 볼륨 정보를 반환한다.
func (s *Synchronizer) getVolume(volume *model.ClusterVolume, opts ...Option) (*model.ClusterVolume, error) {
	var err error
	var v model.ClusterVolume

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterVolume{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_volume.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterVolume{UUID: volume.UUID}).
			First(&v).Error
	}); err == nil {
		return &v, nil
	}

	if err = s.SyncClusterVolume(volume, opts...); err != nil {
		return nil, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterVolume{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_volume.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterVolume{UUID: volume.UUID}).
			First(&v).Error
	}); err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return &v, nil
}

func (s *Synchronizer) deleteVolume(volume *model.ClusterVolume) (err error) {
	logger.Infof("[Sync-deleteVolume] Start: cluster(%d) volume(%s)", s.clusterID, volume.UUID)

	var instanceVolumeList []model.ClusterInstanceVolume
	var snapshotList []model.ClusterVolumeSnapshot

	if err = database.Execute(func(db *gorm.DB) error {
		if err = db.Where(&model.ClusterInstanceVolume{ClusterVolumeID: volume.ID}).Find(&instanceVolumeList).Error; err != nil {
			return err
		}

		if err = db.Where(&model.ClusterVolumeSnapshot{ClusterVolumeID: volume.ID}).Find(&snapshotList).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// delete instance volume
	for _, instanceVolume := range instanceVolumeList {
		if err = s.deleteInstanceVolume(volume, &instanceVolume); err != nil {
			return err
		}
	}

	// delete volume snapshot
	for _, snapshot := range snapshotList {
		if err = s.deleteVolumeSnapshot(volume, &snapshot); err != nil {
			return err
		}
	}

	// delete volume
	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(volume).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteVolume] Could not delete cluster(%d) volume(%s). Cause: %+v", s.clusterID, volume.UUID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterVolumeDeleted, &queue.DeleteClusterVolume{
		Cluster:    &model.Cluster{ID: s.clusterID},
		OrigVolume: volume,
	}); err != nil {
		logger.Warnf("[Sync-deleteVolume] Could not publish cluster(%d) volume(%s) deleted message. Cause: %+v",
			s.clusterID, volume.UUID, err)
	}

	v, err := clusterVolume.GetVolume(s.clusterID, volume.ID)
	if err != nil {
		if !errors.Equal(err, clusterVolume.ErrNotFoundVolume) {
			logger.Warnf("[Sync-deleteVolume] Could not delete cluster(%d) volume(%s) metadata. Cause: %+v",
				s.clusterID, volume.UUID, err)
		}
	} else {
		if err = v.Delete(); err != nil {
			logger.Errorf("[Sync-deleteVolume] Could not delete cluster(%d) volume(%s) kv. Cause: %+v", s.clusterID, volume.UUID, err)
			return err
		}
	}

	logger.Infof("[Sync-deleteVolume] Success: cluster(%d) volume(%s)", s.clusterID, volume.UUID)
	return nil
}

func (s *Synchronizer) syncDeletedVolumeList(rsp *client.GetVolumeListResponse) error {
	var err error
	var volumeList []*model.ClusterVolume

	// 동기화되어있는 볼륨 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterVolume{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_volume.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Find(&volumeList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 볼륨 제거
	for _, volume := range volumeList {
		exists := false
		for _, result := range rsp.ResultList {
			exists = exists || (result.Volume.UUID == volume.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteVolume(volume); err != nil {
			return err
		}
	}

	return nil
}

// SyncClusterVolumeList 볼륨 목록 동기화
func (s *Synchronizer) SyncClusterVolumeList(opts ...Option) error {
	logger.Infof("[SyncClusterVolumeList] Start: cluster(%d)", s.clusterID)

	// 클러스터의 볼륨 목록 조회
	rsp, err := s.Cli.GetVolumeList()
	if err != nil {
		return err
	}

	// 클러스터에서 삭제된 볼륨 제거
	if err = s.syncDeletedVolumeList(rsp); err != nil {
		return err
	}

	// 볼륨 상세 동기화
	for _, volume := range rsp.ResultList {
		if e := s.SyncClusterVolume(&model.ClusterVolume{UUID: volume.Volume.UUID}, opts...); e != nil {
			logger.Warnf("[SyncClusterVolumeList] Failed to sync cluster(%d) volume(%s). Cause: %+v", s.clusterID, volume.Volume.UUID, e)
			err = e
		}
	}

	logger.Infof("[SyncClusterVolumeList] Completed: cluster(%d)", s.clusterID)
	return err
}

// SyncClusterVolume 볼륨 상세 동기화
func (s *Synchronizer) SyncClusterVolume(volume *model.ClusterVolume, opts ...Option) error {
	logger.Infof("[SyncClusterVolume] Start: cluster(%d) volume(%s)", s.clusterID, volume.UUID)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.TenantUUID != "" {
		if ok, err := s.isDeletedTenant(options.TenantUUID); ok {
			// 삭제된 tenant
			logger.Infof("[SyncClusterVolume] Already sync and deleted tenant: cluster(%d) volume(%s) tenant(%s)",
				s.clusterID, volume.UUID, options.TenantUUID)
			return nil
		} else if err != nil {
			logger.Errorf("[SyncClusterVolume] Could not get tenant from db: cluster(%d) volume(%s) tenant(%s). Cause: %+v",
				s.clusterID, volume.UUID, options.TenantUUID, err)
			return err
		}
		// 삭제안된 tenant 라면 동기화 계속 진행
	}

	rsp, err := s.Cli.GetVolume(client.GetVolumeRequest{Volume: model.ClusterVolume{UUID: volume.UUID}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterVolume{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_volume.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterVolume{UUID: volume.UUID}).
			First(volume).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		logger.Infof("[SyncClusterVolume] Success - nothing was updated: cluster(%d) volume(%s)", s.clusterID, volume.UUID)
		return nil

	case rsp == nil: // 볼륨 삭제
		if err = s.deleteVolume(volume); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 볼륨 추가
		var tenant *model.ClusterTenant
		if tenant, err = s.getTenant(&model.ClusterTenant{UUID: rsp.Result.Tenant.UUID}, opts...); err != nil {
			return err
		} else if tenant == nil {
			logger.Warnf("[SyncClusterVolume] Could not sync cluster(%d) volume(%s). Cause: not found tenant(%s)",
				s.clusterID, volume.UUID, rsp.Result.Tenant.UUID)
			return nil
		}

		var storage *model.ClusterStorage
		if storage, err = s.getStorage(&model.ClusterStorage{UUID: rsp.Result.Storage.UUID}, opts...); err != nil {
			return err
		}

		rsp.Result.Volume.ClusterTenantID = tenant.ID
		rsp.Result.Volume.ClusterStorageID = storage.ID
		*volume = rsp.Result.Volume

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(volume).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterVolume] Done - added: cluster(%d) volume(%s)", s.clusterID, volume.UUID)

		// volume attachment 정보 동기화
		if err = syncClusterVolumeMetadata(&clusterVolume.ClusterVolume{
			ClusterID:  s.clusterID,
			VolumeID:   volume.ID,
			VolumeUUID: volume.UUID,
		}, rsp.Result.Metadata); err != nil {
			return err
		}

		if err = publishMessage(constant.QueueNoticeClusterVolumeCreated, &queue.CreateClusterVolume{
			Cluster: &model.Cluster{ID: s.clusterID},
			Volume:  volume,
		}); err != nil {
			logger.Warnf("[SyncClusterVolume] Could not publish cluster(%d) volume(%s) created message. Cause: %+v",
				s.clusterID, volume.UUID, err)
		}

	case options.Force || !compareVolume(volume, &rsp.Result): // 볼륨 정보 변경
		origVolume := *volume
		var tenant *model.ClusterTenant
		if tenant, err = s.getTenant(&model.ClusterTenant{UUID: rsp.Result.Tenant.UUID}, opts...); err != nil {
			return err
		} else if tenant == nil {
			logger.Warnf("[SyncClusterVolume] Could not sync cluster(%d) volume(%s). Cause: not found tenant(%s)",
				s.clusterID, volume.UUID, rsp.Result.Tenant.UUID)
			return nil
		}

		var storage *model.ClusterStorage
		if storage, err = s.getStorage(&model.ClusterStorage{UUID: rsp.Result.Storage.UUID}, opts...); err != nil {
			return err
		}

		rsp.Result.Volume.ID = volume.ID
		rsp.Result.Volume.ClusterTenantID = tenant.ID
		rsp.Result.Volume.ClusterStorageID = storage.ID
		*volume = rsp.Result.Volume

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(volume).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterVolume] Done - updated: cluster(%d) volume(%s)", s.clusterID, volume.UUID)

		// volume attachment 정보 동기화
		if err = syncClusterVolumeMetadata(&clusterVolume.ClusterVolume{
			ClusterID:  s.clusterID,
			VolumeID:   volume.ID,
			VolumeUUID: volume.UUID,
		}, rsp.Result.Metadata); err != nil {
			return err
		}

		if err = publishMessage(constant.QueueNoticeClusterVolumeUpdated, &queue.UpdateClusterVolume{
			Cluster:    &model.Cluster{ID: s.clusterID},
			OrigVolume: &origVolume,
			Volume:     volume,
		}); err != nil {
			logger.Warnf("[SyncClusterVolume] Could not publish cluster(%d) volume(%s) updated message. Cause: %+v",
				s.clusterID, volume.UUID, err)
		}
	}

	logger.Infof("[SyncClusterVolume] Success: cluster(%d) volume(%s)", s.clusterID, volume.UUID)
	return nil
}

func syncClusterVolumeMetadata(volume *clusterVolume.ClusterVolume, metadata map[string]interface{}) error {
	err := clusterVolume.PutVolume(volume)
	if err != nil {
		return err
	}

	v, err := clusterVolume.GetVolume(volume.ClusterID, volume.VolumeID)
	if err != nil {
		return err
	}

	return v.MergeMetadata(metadata)
}
