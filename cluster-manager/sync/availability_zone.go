package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
)

func compareAvailabilityZone(dbAvailabilityZone, clusterAvailabilityZone *model.ClusterAvailabilityZone) bool {
	if dbAvailabilityZone.Name != clusterAvailabilityZone.Name {
		return false
	}

	if dbAvailabilityZone.Available != clusterAvailabilityZone.Available {
		return false
	}

	return true
}

func (s *Synchronizer) getAvailabilityZone(az *model.ClusterAvailabilityZone, opts ...Option) (*model.ClusterAvailabilityZone, error) {
	var err error
	var zone model.ClusterAvailabilityZone

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterAvailabilityZone{ClusterID: s.clusterID, Name: az.Name}).First(&zone).Error
	}); err == nil {
		return &zone, nil
	}

	// Availability Zone 이 없으면 다시 한번 Sync 를 해서 다시 조회한다.
	if err = s.SyncClusterAvailabilityZoneList(opts...); err != nil {
		return nil, err
	}

	err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterAvailabilityZone{ClusterID: s.clusterID, Name: az.Name}).First(&zone).Error
	})
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, nil

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &zone, nil
}

func (s *Synchronizer) deleteAvailabilityZone(zone *model.ClusterAvailabilityZone) (err error) {
	logger.Infof("[Sync-deleteAvailabilityZone] Start: cluster(%d) availability zone(%s)", s.clusterID, zone.Name)

	var instanceList []model.ClusterInstance
	var hypervisorList []model.ClusterHypervisor

	if err = database.Execute(func(db *gorm.DB) error {
		if err = db.Where(&model.ClusterInstance{ClusterAvailabilityZoneID: zone.ID}).Find(&instanceList).Error; err != nil {
			return errors.UnusableDatabase(err)
		}

		if err = db.Where(&model.ClusterHypervisor{ClusterAvailabilityZoneID: zone.ID}).Find(&hypervisorList).Error; err != nil {
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return err
	}

	// delete instance
	for _, instance := range instanceList {
		if err = s.deleteInstance(&instance); err != nil {
			return err
		}
	}

	// delete hypervisor
	for _, hypervisor := range hypervisorList {
		if err = s.deleteHypervisor(&hypervisor); err != nil {
			return err
		}
	}

	// delete availability zone
	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(zone).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteAvailabilityZone] Could not delete cluster(%d) availability zone(%s). Cause: %+v",
			s.clusterID, zone.Name, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterAvailabilityZoneDeleted, &queue.DeleteClusterAvailabilityZone{
		Cluster:              &model.Cluster{ID: s.clusterID},
		OrigAvailabilityZone: zone,
	}); err != nil {
		logger.Warnf("[Sync-deleteAvailabilityZone] Could not publish cluster availability zone deleted message: cluster(%d) availability zone(%s). Cause: %+v",
			s.clusterID, zone.Name, err)
	}

	logger.Infof("[Sync-deleteAvailabilityZone] Success: cluster(%d) availability zone(%s)", s.clusterID, zone.Name)
	return nil
}

func (s *Synchronizer) syncDeletedAvailabilityZoneList(rsp *client.GetAvailabilityZoneListResponse) error {
	var err error
	var zoneList []*model.ClusterAvailabilityZone

	// 동기화되어있는 가용구역 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterAvailabilityZone{ClusterID: s.clusterID}).Find(&zoneList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 가용구역 제거
	for _, zone := range zoneList {
		exists := false
		for _, result := range rsp.ResultList {
			exists = exists || (result.AvailabilityZone.Name == zone.Name)
		}
		if exists {
			continue
		}

		if err = s.deleteAvailabilityZone(zone); err != nil {
			return err
		}
	}

	return nil
}

// SyncClusterAvailabilityZoneList 가용 구역 목록 동기화
func (s *Synchronizer) SyncClusterAvailabilityZoneList(opts ...Option) error {
	logger.Infof("[SyncClusterAvailabilityZoneList] Start: cluster(%d)", s.clusterID)

	// 클러스터의 가용구역 목록 조회
	rsp, err := s.Cli.GetAvailabilityZoneList()
	if err != nil {
		return err
	}

	// 클러스터에서 삭제된 가용구역 제거
	if err = s.syncDeletedAvailabilityZoneList(rsp); err != nil {
		return err
	}

	// 가용구역 상세 동기화
	for _, result := range rsp.ResultList {
		if e := s.SyncClusterAvailabilityZone(&result.AvailabilityZone, opts...); e != nil {
			logger.Warnf("[SyncClusterAvailabilityZoneList] Failed to sync cluster(%d) availability zone(%s). Cause: %+v",
				s.clusterID, result.AvailabilityZone.Name, e)
			err = e
		}
	}

	logger.Infof("[SyncClusterAvailabilityZoneList] Completed: cluster(%d)", s.clusterID)
	return err
}

// SyncClusterAvailabilityZone 가용구역 상세 동기화
func (s *Synchronizer) SyncClusterAvailabilityZone(zone *model.ClusterAvailabilityZone, opts ...Option) (err error) {
	logger.Infof("[SyncClusterAvailabilityZone] Start: cluster(%d) availability zone(%s)", s.clusterID, zone.Name)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	var orig model.ClusterAvailabilityZone
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterAvailabilityZone{ClusterID: s.clusterID, Name: zone.Name}).First(&orig).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case err == gorm.ErrRecordNotFound: // 신규 가용 구역이 추가됨
		zone.ClusterID = s.clusterID

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(zone).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterAvailabilityZone] Done - added: cluster(%d) availability zone(%s)", s.clusterID, zone.Name)

		if err = publishMessage(constant.QueueNoticeClusterAvailabilityZoneCreated, &queue.CreateClusterAvailabilityZone{
			Cluster:          &model.Cluster{ID: s.clusterID},
			AvailabilityZone: zone,
		}); err != nil {
			logger.Warnf("[SyncClusterAvailabilityZone] Could not publish cluster(%d) availability zone(%s) created message. Cause: %+v",
				s.clusterID, zone.Name, err)
		}

	case options.Force || !compareAvailabilityZone(&orig, zone): // 기존 가용 구역이 변경됨
		zone.ID = orig.ID
		zone.ClusterID = s.clusterID

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(zone).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterAvailabilityZone] Done - updated: cluster(%d) availability zone(%s)", s.clusterID, zone.Name)

		if err = publishMessage(constant.QueueNoticeClusterAvailabilityZoneUpdated, &queue.UpdateClusterAvailabilityZone{
			Cluster:          &model.Cluster{ID: s.clusterID},
			AvailabilityZone: zone,
		}); err != nil {
			logger.Warnf("[SyncClusterAvailabilityZone] Could not publish cluster(%d) availability zone(%s) zone updated message. Cause: %+v",
				s.clusterID, zone.Name, err)
		}
	}

	logger.Infof("[SyncClusterAvailabilityZone] Success: cluster(%d) availability zone(%s)", s.clusterID, zone.Name)
	return nil
}
