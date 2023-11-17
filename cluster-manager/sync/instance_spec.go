package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	drConstant "github.com/datacommand2/cdm-disaster-recovery/common/constant"
	"github.com/datacommand2/cdm-disaster-recovery/common/migrator"
	"github.com/jinzhu/gorm"
	"reflect"
)

func (s *Synchronizer) deleteInstanceExtraSpec(extraSpec *model.ClusterInstanceExtraSpec) (err error) {
	logger.Infof("[Sync-deleteInstanceExtraSpec] Start: cluster(%d) instance extra spec(%d)", s.clusterID, extraSpec.ClusterInstanceSpecID)

	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(extraSpec).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteInstanceExtraSpec] Could not delete cluster(%d) instance extra spec(%d). Cause: %+v", s.clusterID, extraSpec.ClusterInstanceSpecID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterInstanceExtraSpecDeleted, queue.DeleteClusterInstanceExtraSpec{
		Cluster:       &model.Cluster{ID: s.clusterID},
		OrigExtraSpec: extraSpec,
	}); err != nil {
		logger.Warnf("[Sync-deleteInstanceExtraSpec] Could not publish cluster(%d) instance extra spec(%d) deleted message. Cause: %+v",
			s.clusterID, extraSpec.ClusterInstanceSpecID, err)
	}

	logger.Infof("[Sync-deleteInstanceExtraSpec] Success: cluster(%d) instance extra spec(%d)", s.clusterID, extraSpec.ClusterInstanceSpecID)
	return nil
}

// SyncClusterInstanceExtraSpec 인스턴스 엑스트라 스팩 상세 동기화
func (s *Synchronizer) SyncClusterInstanceExtraSpec(spec *model.ClusterInstanceSpec, extraSpec *model.ClusterInstanceExtraSpec, opts ...Option) error {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	rsp, err := s.Cli.GetInstanceExtraSpec(client.GetInstanceExtraSpecRequest{InstanceSpec: model.ClusterInstanceSpec{UUID: spec.UUID}, InstanceExtraSpec: model.ClusterInstanceExtraSpec{Key: extraSpec.Key}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterInstanceExtraSpec{ClusterInstanceSpecID: spec.ID, Key: extraSpec.Key}).First(extraSpec).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		return nil

	case rsp == nil: // 인스턴스 엑스트라 스팩 삭제
		if err = s.deleteInstanceExtraSpec(extraSpec); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 인스턴스 엑스트라 스팩 추가
		rsp.Result.InstanceExtraSpec.ClusterInstanceSpecID = spec.ID
		*extraSpec = rsp.Result.InstanceExtraSpec

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(extraSpec).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterInstanceExtraSpec] Done - added: cluster(%d) instance extra spec(%s)", s.clusterID, spec.UUID)

		if err = publishMessage(constant.QueueNoticeClusterInstanceExtraSpecCreated, queue.CreateClusterInstanceExtraSpec{
			Cluster:   &model.Cluster{ID: s.clusterID},
			ExtraSpec: extraSpec,
		}); err != nil {
			logger.Warnf("[SyncClusterInstanceExtraSpec] Could not publish cluster instance extra spec created message. Cause: %+v", err)
		}

	case options.Force || !(extraSpec.Value == rsp.Result.InstanceExtraSpec.Value): // 인스턴스 엑스트라 스팩 정보 수정
		rsp.Result.InstanceExtraSpec.ID = extraSpec.ID
		rsp.Result.InstanceExtraSpec.ClusterInstanceSpecID = spec.ID
		*extraSpec = rsp.Result.InstanceExtraSpec

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(extraSpec).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterInstanceExtraSpec] Done - updated: cluster(%d) instance extra spec(%s)", s.clusterID, spec.UUID)

		if err = publishMessage(constant.QueueNoticeClusterInstanceExtraSpecUpdated, queue.UpdateClusterInstanceExtraSpec{
			Cluster:   &model.Cluster{ID: s.clusterID},
			ExtraSpec: extraSpec,
		}); err != nil {
			logger.Warnf("[SyncClusterInstanceExtraSpec] Could not publish cluster instance extra spec updated message. Cause: %+v", err)
		}
	}

	return nil
}

func compareInstanceSpec(dbSpec, result *model.ClusterInstanceSpec) bool {
	if dbSpec.UUID != result.UUID {
		return false
	}

	if dbSpec.Name != result.Name {
		return false
	}

	if dbSpec.VcpuTotalCnt != result.VcpuTotalCnt {
		return false
	}

	if dbSpec.MemTotalBytes != result.MemTotalBytes {
		return false
	}

	if dbSpec.DiskTotalBytes != result.DiskTotalBytes {
		return false
	}

	if dbSpec.SwapTotalBytes != result.SwapTotalBytes {
		return false
	}

	if dbSpec.EphemeralTotalBytes != result.EphemeralTotalBytes {
		return false
	}

	if !reflect.DeepEqual(dbSpec.Description, result.Description) {
		return false
	}

	return true
}

func (s *Synchronizer) getInstanceSpecList() ([]*model.ClusterInstanceSpec, error) {
	var err error
	var result []*model.ClusterInstanceSpec

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterInstanceSpec{ClusterID: s.clusterID}).Find(&result).Error
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Synchronizer) getInstanceSpec(spec *model.ClusterInstanceSpec, opts ...Option) (*model.ClusterInstanceSpec, error) {
	var err error
	var result model.ClusterInstanceSpec

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterInstanceSpec{ClusterID: s.clusterID, UUID: spec.UUID}).First(&result).Error
	}); err == nil {
		return &result, nil
	}

	if err = s.SyncClusterInstanceSpec(spec, opts...); err != nil {
		return nil, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterInstanceSpec{ClusterID: s.clusterID, UUID: spec.UUID}).First(&result).Error
	}); err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return &result, nil
}

func (s *Synchronizer) deleteInstanceSpec(spec *model.ClusterInstanceSpec) (err error) {
	logger.Infof("[Sync-deleteInstanceSpec] Start: cluster(%d) instance spec(%s)", s.clusterID, spec.UUID)

	var (
		count         uint64
		extraSpecList []model.ClusterInstanceExtraSpec
	)

	if err = database.Execute(func(db *gorm.DB) error {
		if err = db.Table(model.ClusterInstance{}.TableName()).
			Where(&model.ClusterInstance{ClusterInstanceSpecID: spec.ID}).
			Count(&count).Error; err != nil {
			return err
		}

		if count > 0 {
			return nil
		}

		if err = db.Where(&model.ClusterInstanceExtraSpec{ClusterInstanceSpecID: spec.ID}).Find(&extraSpecList).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	if count > 0 {
		return nil
	}

	// 해당 instance spec 을 쓰고 있는 shared task 삭제
	migrator.ClearSharedTask(s.clusterID, spec.ID, spec.UUID, drConstant.MigrationTaskTypeCreateSpec)

	// delete instance extra spec
	errFlag := false
	// delete instance extra spec
	for _, extraSpec := range extraSpecList {
		if err = s.deleteInstanceExtraSpec(&extraSpec); err != nil {
			errFlag = true
			continue
		}
	}

	if errFlag {
		logger.Errorf("[Sync-deleteInstanceSpec] Could not complete delete instance spec: instance spec(%s:%s)", spec.Name, spec.UUID)
		return errors.UnusableDatabase(err)
	}

	// delete instance spec
	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(spec).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteInstanceSpec] Could not delete cluster(%d) instance spec(%s). Cause: %+v", s.clusterID, spec.UUID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterInstanceSpecDeleted, queue.DeleteClusterInstanceSpec{
		Cluster:  &model.Cluster{ID: s.clusterID},
		OrigSpec: spec,
	}); err != nil {
		logger.Warnf("[Sync-deleteInstanceSpec] Could not publish cluster(%d) instance spec(%s) deleted message. Cause: %+v", s.clusterID, spec.UUID, err)
	}

	logger.Infof("[Sync-deleteInstanceSpec] Success: cluster(%d) instance spec(%s)", s.clusterID, spec.UUID)
	return nil
}

// SyncClusterInstanceSpecs 인스턴스 스팩들 동기화
func (s *Synchronizer) SyncClusterInstanceSpecs(specList []*model.ClusterInstanceSpec) error {
	logger.Infof("[SyncClusterInstanceSpecs] Start: cluster(%d)", s.clusterID)

	for _, spec := range specList {
		_ = s.SyncClusterInstanceSpec(spec)
	}

	logger.Infof("[SyncClusterInstanceSpecs] Done: cluster(%d)", s.clusterID)
	return nil
}

// SyncClusterInstanceSpecList 인스턴스 스팩 목록 동기화
func (s *Synchronizer) SyncClusterInstanceSpecList() error {
	logger.Infof("[SyncClusterInstanceSpecList] Start: cluster(%d)", s.clusterID)

	specList, err := s.getInstanceSpecList()
	if err != nil {
		logger.Errorf("[SyncClusterInstanceSpecList] Could not get cluster(%d) instance spec list. Cause: %+v", s.clusterID, err)
		return err
	}

	for _, spec := range specList {
		if e := s.SyncClusterInstanceSpec(spec); e != nil {
			logger.Warnf("[SyncClusterInstanceSpecList] Failed to sync cluster(%d) instance spec(%s). Cause: %+v", s.clusterID, spec.Name, e)
			err = e
		}
	}

	logger.Infof("[SyncClusterInstanceSpecList] Completed: cluster(%d)", s.clusterID)
	return nil
}

// SyncClusterInstanceSpec 인스턴스 스팩 상세 동기화
func (s *Synchronizer) SyncClusterInstanceSpec(spec *model.ClusterInstanceSpec, opts ...Option) error {
	logger.Infof("[SyncClusterInstanceSpec] Start: cluster(%d) instance spec(%s)", s.clusterID, spec.UUID)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	rsp, err := s.Cli.GetInstanceSpec(client.GetInstanceSpecRequest{InstanceSpec: model.ClusterInstanceSpec{UUID: spec.UUID}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterInstanceSpec{ClusterID: s.clusterID, UUID: spec.UUID}).First(spec).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		return nil

	case rsp == nil: // 인스턴스 스팩 삭제
		if err = s.deleteInstanceSpec(spec); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 인스턴스 스팩 추가
		rsp.Result.InstanceSpec.ClusterID = s.clusterID
		*spec = rsp.Result.InstanceSpec

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(spec).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterInstanceSpec] Done - added: cluster(%d) instance spec(%s)", s.clusterID, spec.UUID)

		if err = publishMessage(constant.QueueNoticeClusterInstanceSpecCreated, queue.CreateClusterInstanceSpec{
			Cluster: &model.Cluster{ID: s.clusterID},
			Spec:    spec,
		}); err != nil {
			logger.Warnf("[SyncClusterInstanceSpec] Could not publish cluster(%d) instance spec(%s) created message. Cause: %+v", s.clusterID, spec.UUID, err)
		}

	case options.Force || !compareInstanceSpec(spec, &rsp.Result.InstanceSpec): // 인스턴스 스팩 정보 수정
		rsp.Result.InstanceSpec.ID = spec.ID
		rsp.Result.InstanceSpec.ClusterID = s.clusterID
		*spec = rsp.Result.InstanceSpec

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(spec).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterInstanceSpec] Done - updated: cluster(%d) instance spec(%s)", s.clusterID, spec.UUID)

		if err = publishMessage(constant.QueueNoticeClusterInstanceSpecUpdated, queue.UpdateClusterInstanceSpec{
			Cluster: &model.Cluster{ID: s.clusterID},
			Spec:    spec,
		}); err != nil {
			logger.Warnf("[SyncClusterInstanceSpec] Could not publish cluster(%d) instance spec(%s) updated message. Cause: %+v", s.clusterID, spec.UUID, err)
		}
	}

	// 신규, 수정인 경우 익스트라 스팩 동기화
	if rsp != nil {
		for _, extraSpec := range rsp.Result.ExtraSpecList {
			if err = s.SyncClusterInstanceExtraSpec(spec, &extraSpec, opts...); err != nil {
				logger.Warnf("[SyncClusterInstanceSpec] Failed to sync cluster(%d) instance extra spec(%d). Cause: %+v", s.clusterID, extraSpec.ClusterInstanceSpecID, err)
			}
		}
	}

	logger.Infof("[SyncClusterInstanceSpec] Success: cluster(%d) instance spec(%s)", s.clusterID, spec.UUID)
	return err
}
