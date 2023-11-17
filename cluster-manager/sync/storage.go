package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack/train"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	clusterStorage "github.com/datacommand2/cdm-center/cluster-manager/storage"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"

	"github.com/jinzhu/gorm"

	"reflect"
	"strings"
)

func compareStorage(dbStorage *model.ClusterStorage, result *client.GetStorageResult) bool {
	if dbStorage.UUID != result.Storage.UUID {
		return false
	}

	if dbStorage.Name != result.Storage.Name {
		return false
	}

	if dbStorage.TypeCode != result.Storage.TypeCode {
		return false
	}

	if !reflect.DeepEqual(dbStorage.Description, result.Storage.Description) {
		return false
	}

	if !reflect.DeepEqual(dbStorage.CapacityBytes, result.Storage.CapacityBytes) {
		return false
	}

	if !reflect.DeepEqual(dbStorage.UsedBytes, result.Storage.UsedBytes) {
		return false
	}

	return true
}

func getStorageStatus(cli client.Client, typeCode string) string {
	_, backendMap, err := GetStorageServiceList(cli)
	if err != nil {
		logger.Warnf("[getStorageStatus] Error occurred during get storage backend map. Cause: %+v", err)
		return "unknown"
	}

	code := strings.Split(typeCode, "openstack.storage.type.")
	status := backendMap[code[1]]
	if status == "" {
		status = "unknown"
	}

	return status
}

// GetStorageServiceList 시스템 정보의 Storage 가져오기
func GetStorageServiceList(cli client.Client) ([]queue.StorageServices, map[string]string, error) {
	// get cinder service info
	rsp, err := train.ListAllCinderServices(cli)
	if err != nil {
		logger.Errorf("[GetCinderServices] Could not get all cinder service list. Cause: %+v", err)
		return nil, nil, err
	}

	var storages []queue.StorageServices
	backendMap := make(map[string]string)
	backendMap["unknown"] = "unknown"

	if rsp.Services != nil {
		for _, service := range rsp.Services {
			if service.Binary == "cinder-volume" {
				arr := strings.Split(service.Host, "@")

				v, ok := backendMap[arr[1]]
				if service.State == "up" && service.Status == "enabled" {
					if !ok || v == "unavailable" {
						backendMap[arr[1]] = "available"
					}
				} else {
					if !ok {
						backendMap[arr[1]] = "unavailable"
					}
				}

				storages = append(storages, queue.StorageServices{
					Binary:      service.Binary,
					BackendName: arr[1],
					Host:        service.Host,
					Zone:        service.Zone,
					Status:      backendMap[arr[1]],
					UpdatedAt:   service.UpdatedAt,
				})

			} else {
				v, ok := backendMap[service.Binary]
				if service.State == "up" && service.Status == "enabled" {
					if !ok || v == "unavailable" {
						backendMap[service.Binary] = "available"
					}
				} else {
					if !ok {
						backendMap[service.Binary] = "unavailable"
					}
				}

				storages = append(storages, queue.StorageServices{
					Binary:    service.Binary,
					Host:      service.Host,
					Zone:      service.Zone,
					Status:    backendMap[service.Binary],
					UpdatedAt: service.UpdatedAt,
				})
			}
		}
	}

	return storages, backendMap, nil
}

// SetUnknownClusterStoragesStatus 클러스터의 스토리지상태 모두 unknown 으로 설정
func SetUnknownClusterStoragesStatus(id uint64) (err error) {
	var storageList []*model.ClusterStorage
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterStorage{ClusterID: id}).Find(&storageList).Error
	}); err != nil {
		logger.Errorf("[SetUnknownClusterStoragesStatus] Unable to get storage list for cluster(%d). Cause: %+v", id, err)
		return errors.UnusableDatabase(err)
	}

	if len(storageList) == 0 {
		logger.Infof("[SetUnknownClusterStoragesStatus] storageList is empty: cluster(%d)", id)
		return nil
	}

	for _, s := range storageList {
		if s.Status == "unknown" {
			continue
		}

		s.Status = "unknown"

		if err = database.Transaction(func(db *gorm.DB) error {
			return db.Save(s).Error
		}); err != nil {
			logger.Warnf("[SetUnknownClusterStoragesStatus] The storage(%d) status of the cluster(%d) cannot be updated to 'Unknown'. Cause: %+v", s.ID, id, err)
			continue
		}

		logger.Infof("[SetUnknownClusterStoragesStatus] The storage(%d) status of the cluster(%d) has changed to 'Unknown'.", s.ID, id)
	}

	return nil
}

// UpdateClusterStoragesStatus 스토리지 status update
func (s *Synchronizer) UpdateClusterStoragesStatus(storages []queue.StorageServices) (err error) {
	// get storage list
	var storageList []*model.ClusterStorage
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterStorage{ClusterID: s.clusterID}).Find(&storageList).Error
	}); err != nil {
		logger.Errorf("[UpdateClusterStoragesStatus] Could not get storage list: cluster(%d). Cause: %+v", s.clusterID, err)
		return errors.UnusableDatabase(err)
	}

	if len(storageList) == 0 {
		logger.Infof("[UpdateClusterStoragesStatus] storageList is empty: cluster(%d)", s.clusterID)
		return nil
	}

	var cs *clusterStorage.ClusterStorage
	for _, dbStorage := range storageList {
		if cs, err = clusterStorage.GetStorage(s.clusterID, dbStorage.ID); err != nil {
			continue
		}

		if storages != nil {
			for _, value := range storages {
				// 시스템정보의 스토리지 서비스 상태중 cinder-volume 만 적용하고 같은 BackendName 만 적용.
				// 볼륨타입 으로 비교하면 같은 볼륨타입이 여러개일때 중복된다.
				if value.BackendName == cs.BackendName {
					// storage 의 상태가 이전 상태와 같을때 수정하지 않음
					if dbStorage.Status == value.Status {
						continue
					}

					if value.Status == "" {
						// storage 가 openstack.storage.type.unknown 이거나 조회되지 않은 storage 의 경우
						logger.Infof("[UpdateClusterStoragesStatus] %s storageMap[%s] is empty -> unknown: cluster(%d).", dbStorage.TypeCode, value.BackendName, s.clusterID)
						value.Status = "unknown"
					}

					dbStorage.Status = value.Status

					if err = database.Transaction(func(db *gorm.DB) error {
						return db.Save(dbStorage).Error
					}); err != nil {
						logger.Warnf("[UpdateClusterStoragesStatus] Could not update storage(%d) : cluster(%d). Cause: %+v", dbStorage.ID, s.clusterID, err)
						continue
					}

					logger.Infof("[UpdateClusterStoragesStatus] The storage(%s) in the cluster(%d) has changed status(%s).", dbStorage.TypeCode, s.clusterID, dbStorage.Status)
				}
			}
		}
	}

	return nil
}

func (s *Synchronizer) getStorage(storage *model.ClusterStorage, opts ...Option) (*model.ClusterStorage, error) {
	var err error
	var result model.ClusterStorage
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterStorage{ClusterID: s.clusterID, UUID: storage.UUID}).First(&result).Error
	}); err == nil {
		return &result, nil
	}

	if err = s.SyncClusterStorage(storage, opts...); err != nil {
		return nil, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterStorage{ClusterID: s.clusterID, UUID: storage.UUID}).First(&result).Error
	}); err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return &result, nil
}

func (s *Synchronizer) deleteStorage(storage *model.ClusterStorage) (err error) {
	logger.Infof("[Sync-deleteStorage] Start: cluster(%d) storage(%s)", s.clusterID, storage.UUID)

	// delete volume
	var volumeList []model.ClusterVolume
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterVolume{ClusterStorageID: storage.ID}).Find(&volumeList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	for _, volume := range volumeList {
		if err = s.deleteVolume(&volume); err != nil {
			return err
		}
	}

	// delete storage
	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(&storage).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteStorage] Could not delete cluster(%d) storage(%s). Cause: %+v", s.clusterID, storage.UUID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterStorageDeleted, &queue.DeleteClusterStorage{
		Cluster:     &model.Cluster{ID: s.clusterID},
		OrigStorage: storage,
	}); err != nil {
		logger.Warnf("[Sync-deleteStorage] Could not publish cluster(%d) storage(%s) deleted message. Cause: %+v", s.clusterID, storage.UUID, err)
	}

	ss, err := clusterStorage.GetStorage(s.clusterID, storage.ID)
	if err != nil {
		if !errors.Equal(err, clusterStorage.ErrNotFoundStorage) {
			logger.Warnf("[Sync-deleteStorage] Could not delete cluster(%d) storage(%s) metadata. Cause: %+v", s.clusterID, storage.UUID, err)
		}
	} else {
		if err = ss.Delete(); err != nil {
			logger.Errorf("[Sync-deleteStorage] Could not delete cluster(%d) storage(%s) kv. Cause: %+v", s.clusterID, storage.UUID, err)
			return err
		}
	}

	logger.Infof("[Sync-deleteStorage] Success: cluster(%d) storage(%s)", s.clusterID, storage.UUID)
	return nil
}

func (s *Synchronizer) syncDeletedStorageList(rsp *client.GetStorageListResponse) (err error) {
	var storageList []*model.ClusterStorage

	// 동기화되어있는 스토리지 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterStorage{ClusterID: s.clusterID}).Find(&storageList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 스토리지 제거
	for _, storage := range storageList {
		exists := false
		for _, cStorage := range rsp.ResultList {
			exists = exists || (cStorage.Storage.UUID == storage.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteStorage(storage); err != nil {
			return err
		}
	}

	return nil
}

// SyncClusterStorageList 스토리지 목록 동기화
func (s *Synchronizer) SyncClusterStorageList(opts ...Option) error {
	logger.Infof("[SyncClusterStorageList] Start: cluster(%d)", s.clusterID)

	// 클러스터의 스토리지 목록 조회
	rsp, err := s.Cli.GetStorageList()
	if err != nil {
		return err
	}

	// 클러스터에서 삭제된 스토리지 제거
	if err = s.syncDeletedStorageList(rsp); err != nil {
		return err
	}

	// 스토리지 상세 동기화
	for _, storage := range rsp.ResultList {
		if e := s.SyncClusterStorage(&model.ClusterStorage{UUID: storage.Storage.UUID}, opts...); e != nil {
			logger.Warnf("[SyncClusterStorageList] Failed to sync cluster(%d) storage(%s). Cause: %+v", s.clusterID, storage.Storage.UUID, e)
			err = e
		}
	}

	logger.Infof("[SyncClusterStorageList] Success: cluster(%d)", s.clusterID)
	return err
}

// SyncClusterStorage 스토리지 상세 동기화
func (s *Synchronizer) SyncClusterStorage(storage *model.ClusterStorage, opts ...Option) error {
	logger.Infof("[SyncClusterStorage] Start: cluster(%d) storage(%s)", s.clusterID, storage.UUID)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	// FIXME: 임시로 cinder 의 agent 접속정보를 kv store 에서 가져온다.
	agent, err := cluster.GetAgent(s.clusterID)
	if err != nil {
		return err
	}

	rsp, err := s.Cli.GetStorage(client.GetStorageRequest{Storage: model.ClusterStorage{UUID: storage.UUID}, Agent: *agent})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterStorage{ClusterID: s.clusterID, UUID: storage.UUID}).First(storage).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		logger.Infof("[SyncClusterStorage] Success - nothing was updated: cluster(%d) storage(%s)", s.clusterID, storage.UUID)
		return nil

	case rsp == nil: // 스토리지 삭제
		if err = s.deleteStorage(storage); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 스토리지 추가
		rsp.Result.Storage.ClusterID = s.clusterID
		*storage = rsp.Result.Storage

		storage.Status = getStorageStatus(s.Cli, rsp.Result.Storage.TypeCode)

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(storage).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterStorage] Done - added: cluster(%d) storage(%s)", s.clusterID, storage.UUID)

		if err = syncClusterStorageMetadata(&clusterStorage.ClusterStorage{
			ClusterID:   s.clusterID,
			StorageID:   storage.ID,
			Type:        storage.TypeCode,
			BackendName: rsp.Result.VolumeBackendName,
		}, rsp.Result.Metadata); err != nil {
			return err
		}

		if err = publishMessage(constant.QueueNoticeClusterStorageCreated, &queue.CreateClusterStorage{
			Cluster: &model.Cluster{ID: s.clusterID},
			Storage: storage,
		}); err != nil {
			logger.Warnf("[SyncClusterStorage] Could not publish cluster(%d) storage(%s) created message. Cause: %+v",
				s.clusterID, storage.UUID, err)
		}

	case options.Force || !compareStorage(storage, &rsp.Result): // 스토리지 정보 변경
		origStorage := *storage
		rsp.Result.Storage.ID = storage.ID
		rsp.Result.Storage.ClusterID = s.clusterID
		*storage = rsp.Result.Storage

		storage.Status = getStorageStatus(s.Cli, rsp.Result.Storage.TypeCode)

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(storage).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterStorage] Done - updated: cluster(%d) storage(%s)", s.clusterID, storage.UUID)

		origStorageMetadata, err := clusterStorage.GetStorage(origStorage.ClusterID, origStorage.ID)
		if err != nil {
			return err
		}

		if err = syncClusterStorageMetadata(&clusterStorage.ClusterStorage{
			ClusterID:   s.clusterID,
			StorageID:   storage.ID,
			Type:        storage.TypeCode,
			BackendName: rsp.Result.VolumeBackendName,
		}, rsp.Result.Metadata); err != nil {
			return err
		}

		if err = publishMessage(constant.QueueNoticeClusterStorageUpdated, &queue.UpdateClusterStorage{
			Cluster:               &model.Cluster{ID: s.clusterID},
			OrigStorage:           &origStorage,
			Storage:               storage,
			OrigVolumeBackendName: origStorageMetadata.BackendName,
			VolumeBackendName:     rsp.Result.VolumeBackendName,
		}); err != nil {
			logger.Warnf("[SyncClusterStorage] Could not publish cluster(%d) storage(%s) updated message. Cause: %+v",
				s.clusterID, storage.UUID, err)
		}
	}

	logger.Infof("[SyncClusterStorage] Success: cluster(%d) storage(%s)", s.clusterID, storage.UUID)
	return nil
}

func syncClusterStorageMetadata(storage *clusterStorage.ClusterStorage, metadata map[string]interface{}) error {
	err := clusterStorage.PutStorage(storage)
	if err != nil {
		return err
	}

	s, err := clusterStorage.GetStorage(storage.ClusterID, storage.StorageID)
	if err != nil {
		return err
	}

	return s.MergeMetadata(metadata)
}
