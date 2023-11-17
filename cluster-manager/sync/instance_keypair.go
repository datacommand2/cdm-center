package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	drConstant "github.com/datacommand2/cdm-disaster-recovery/common/constant"
	"github.com/datacommand2/cdm-disaster-recovery/common/migrator"
	"github.com/jinzhu/gorm"
)

func (s *Synchronizer) getKeypair(kp *model.ClusterKeypair) (*model.ClusterKeypair, error) {
	var err error
	var keypair model.ClusterKeypair

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterKeypair{ClusterID: s.clusterID, Name: kp.Name}).First(&keypair).Error
	}); err == nil {
		return &keypair, nil
	}

	if err = s.SyncClusterKeypair(kp); err != nil {
		return nil, err
	}

	err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterKeypair{ClusterID: s.clusterID, Name: kp.Name}).First(&keypair).Error
	})
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, nil

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &keypair, nil
}

func (s *Synchronizer) deleteKeypair(keypair *model.ClusterKeypair) (err error) {
	logger.Infof("[Sync-deleteKeypair] Start: cluster(%d) keypair(%s)", s.clusterID, keypair.Name)

	// update instance keypair
	var instanceList []model.ClusterInstance
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterInstance{ClusterKeypairID: &keypair.ID}).Find(&instanceList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	for _, instance := range instanceList {
		if err = s.deleteInstanceKeypair(&instance); err != nil {
			return err
		}
	}

	// delete keypair
	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(keypair).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteKeypair] Could not delete cluster(%d) keypair(%s). Cause: %+v", s.clusterID, keypair.Name, err)
		return errors.UnusableDatabase(err)
	}

	// 해당 keypair 를 쓰고 있는 shared task 삭제
	migrator.ClearSharedTask(s.clusterID, keypair.ID, keypair.Name, drConstant.MigrationTaskTypeCreateKeypair)

	logger.Infof("[Sync-deleteKeypair] Success: cluster(%d) keypair(%s)", s.clusterID, keypair.Name)
	return nil
}

func (s *Synchronizer) syncDeletedKeypairList(rsp *client.GetKeyPairListResponse) error {
	var err error
	var keypairList []*model.ClusterKeypair

	// 동기화되어있는 keypair 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterKeypair{ClusterID: s.clusterID}).Find(&keypairList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 keypair 제거
	for _, keypair := range keypairList {
		exists := false
		for _, clusterKeypair := range rsp.ResultList {
			exists = exists || (clusterKeypair.KeyPair.Name == keypair.Name)
		}

		if exists {
			continue
		}

		if err = s.deleteKeypair(keypair); err != nil {
			return err
		}
	}

	return nil
}

// SyncClusterKeypairList keypair 목록 동기화
func (s *Synchronizer) SyncClusterKeypairList() error {
	logger.Infof("[SyncClusterKeypairList] Start: cluster(%d)", s.clusterID)

	// 클러스터의 keypair 목록 조회
	rsp, err := s.Cli.GetKeyPairList()
	if err != nil {
		return err
	}

	// 클러스터에서 삭제된 keypair 제거
	if err = s.syncDeletedKeypairList(rsp); err != nil {
		return err
	}

	// keypair 상세 동기화
	for _, clusterKeypair := range rsp.ResultList {
		if e := s.SyncClusterKeypair(&model.ClusterKeypair{Name: clusterKeypair.KeyPair.Name}); e != nil {
			logger.Warnf("[SyncClusterKeypairList] Failed to sync cluster(%d) keypair(%s). Cause: %+v", s.clusterID, clusterKeypair.KeyPair.Name, e)
			err = e
		}
	}

	logger.Infof("[SyncClusterKeypairList] Completed: cluster(%d)", s.clusterID)
	return err
}

// SyncClusterKeypair keypair 상세 동기화
func (s *Synchronizer) SyncClusterKeypair(keypair *model.ClusterKeypair) error {
	logger.Infof("[SyncClusterKeypair] Start: cluster(%d) keypair(%s)", s.clusterID, keypair.Name)

	rsp, err := s.Cli.GetKeyPair(client.GetKeyPairRequest{KeyPair: model.ClusterKeypair{Name: keypair.Name}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterKeypair{ClusterID: s.clusterID, Name: keypair.Name}).First(keypair).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		logger.Infof("[SyncClusterKeypair] Success - nothing was updated: cluster(%d) keypair(%s)", s.clusterID, keypair.Name)
		return nil

	case rsp == nil: // keypair 삭제
		if err = s.deleteKeypair(keypair); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // keypair 추가
		rsp.Result.KeyPair.ClusterID = s.clusterID
		*keypair = rsp.Result.KeyPair

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(keypair).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterKeypair] Done - added: cluster(%d) keypair(%s)", s.clusterID, keypair.Name)
	}

	logger.Infof("[SyncClusterKeypair] Success: cluster(%d) keypair(%s)", s.clusterID, keypair.Name)
	return nil
}
