package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack"
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

type instanceVolumeRecord struct {
	InstanceVolume model.ClusterInstanceVolume `gorm:"embedded"`
	Volume         model.ClusterVolume         `gorm:"embedded"`
}

type instanceNetworkRecord struct {
	InstanceNetwork model.ClusterInstanceNetwork `gorm:"embedded"`
	Network         model.ClusterNetwork         `gorm:"embedded"`
	FloatingIP      model.ClusterFloatingIP      `gorm:"embedded"`
}

func compareInstance(dbInstance *model.ClusterInstance, result *client.GetInstanceResult) bool {
	var err error

	if dbInstance.UUID != result.Instance.UUID {
		return false
	}

	if dbInstance.Name != result.Instance.Name {
		return false
	}

	if dbInstance.State != result.Instance.State {
		return false
	}

	if dbInstance.Status != result.Instance.Status {
		return false
	}

	if !reflect.DeepEqual(dbInstance.Description, result.Instance.Description) {
		return false
	}

	if result.KeyPair == nil {
		if dbInstance.ClusterKeypairID == nil {
			return true
		}

		return false
	}

	if dbInstance.ClusterKeypairID == nil {
		return false
	}

	var t model.ClusterTenant
	var az model.ClusterAvailabilityZone
	var h model.ClusterHypervisor
	var s model.ClusterInstanceSpec
	var k model.ClusterKeypair

	if err = database.Execute(func(db *gorm.DB) error {
		if err = db.Find(&t, &model.ClusterTenant{ID: dbInstance.ClusterTenantID}).Error; err != nil {
			return err
		}

		if err = db.Find(&az, &model.ClusterAvailabilityZone{ID: dbInstance.ClusterAvailabilityZoneID}).Error; err != nil {
			return err
		}

		if err = db.Find(&h, &model.ClusterHypervisor{ID: dbInstance.ClusterHypervisorID}).Error; err != nil {
			return err
		}

		if err = db.Find(&s, &model.ClusterInstanceSpec{ID: dbInstance.ClusterInstanceSpecID}).Error; err != nil {
			return err
		}

		if err = db.Find(&k, &model.ClusterKeypair{ID: *dbInstance.ClusterKeypairID}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return false
	}

	if t.UUID != result.Tenant.UUID {
		return false
	}

	if az.Name != result.AvailabilityZone.Name {
		return false
	}

	if h.UUID != result.Hypervisor.UUID {
		return false
	}

	if s.UUID != result.InstanceSpec.UUID {
		return false
	}

	if k.Name != result.KeyPair.Name {
		return false
	}

	return true
}

func (s *Synchronizer) deleteInstanceSecurityGroup(db *gorm.DB, instanceGroup *model.ClusterInstanceSecurityGroup) (err error) {
	logger.Infof("[Sync-deleteInstanceSecurityGroup] Start: cluster(%d) instance(%d) security group(%d)",
		s.clusterID, instanceGroup.ClusterInstanceID, instanceGroup.ClusterSecurityGroupID)

	if err = db.Delete(instanceGroup).Error; err != nil {
		logger.Errorf("[Sync-deleteInstanceSecurityGroup] Could not delete cluster(%d) instance(%d) security group(%d). Cause: %+v",
			s.clusterID, instanceGroup.ClusterInstanceID, instanceGroup.ClusterSecurityGroupID, err)
		return errors.UnusableDatabase(err)
	}

	logger.Infof("[Sync-deleteInstanceSecurityGroup] Success: cluster(%d) instance(%d) security group(%d)",
		s.clusterID, instanceGroup.ClusterInstanceID, instanceGroup.ClusterSecurityGroupID)
	return nil
}

func (s *Synchronizer) deleteInstanceKeypair(instance *model.ClusterInstance) (err error) {
	logger.Infof("[Sync-deleteInstanceKeypair] Start: cluster(%d) instance(%s)", s.clusterID, instance.UUID)

	origInstance := instance
	instance.ClusterKeypairID = nil

	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Save(&instance).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteInstanceKeypair] Could not update cluster(%d) instance(%s) by deleted keypair. Cause: %+v", s.clusterID, instance.UUID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterInstanceUpdated, &queue.UpdateClusterInstance{
		Cluster:      &model.Cluster{ID: s.clusterID},
		OrigInstance: origInstance,
		Instance:     instance,
	}); err != nil {
		logger.Warnf("[Sync-deleteInstanceKeypair] Could not publish cluster(%d) instance(%s) updated message. Cause: %+v", s.clusterID, instance.UUID, err)
	}

	logger.Infof("[Sync-deleteInstanceKeypair] Success: cluster(%d) instance(%s)", s.clusterID, instance.UUID)
	return nil
}

func (s *Synchronizer) deleteInstanceUserScript(db *gorm.DB, instanceUserScript *model.ClusterInstanceUserScript) (err error) {
	logger.Infof("[Sync-deleteInstanceUserScript] Start: cluster(%d) instance(%d) user script.",
		s.clusterID, instanceUserScript.ClusterInstanceID)

	if err = db.Delete(instanceUserScript).Error; err != nil {
		logger.Errorf("[Sync-deleteInstanceUserScript] Could not delete cluster(%d) instance(%d) user script. Cause: %+v",
			s.clusterID, instanceUserScript.ClusterInstanceID, err)
		return errors.UnusableDatabase(err)
	}

	logger.Infof("[Sync-deleteInstanceUserScript] Success: cluster(%d) instance(%d) user script",
		s.clusterID, instanceUserScript.ClusterInstanceID)
	return nil
}

func (s *Synchronizer) deleteInstance(instance *model.ClusterInstance) (err error) {
	logger.Infof("[Sync-deleteInstance] Start: cluster(%d) instance(%s)", s.clusterID, instance.UUID)

	var (
		instanceNetworkList []model.ClusterInstanceNetwork
		instanceGroupList   []model.ClusterInstanceSecurityGroup
		instanceUserScript  model.ClusterInstanceUserScript
		instanceSpec        model.ClusterInstanceSpec
	)
	isExist := true

	if err = database.Execute(func(db *gorm.DB) error {
		if err = db.Where(&model.ClusterInstanceNetwork{ClusterInstanceID: instance.ID}).Find(&instanceNetworkList).Error; err != nil {
			return err
		}

		if err = db.Where(&model.ClusterInstanceSecurityGroup{ClusterInstanceID: instance.ID}).Find(&instanceGroupList).Error; err != nil {
			return err
		}

		if err = db.Where(&model.ClusterInstanceUserScript{ClusterInstanceID: instance.ID}).First(&instanceUserScript).Error; err == gorm.ErrRecordNotFound {
			isExist = false
		} else if err != nil {
			return err
		}

		_ = db.Where(&model.ClusterInstanceSpec{ID: instance.ClusterInstanceSpecID}).First(&instanceSpec).Error
		return nil
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	instanceVolumeRecords, err := s.getInstanceVolumeRecords(instance.ID)
	if err != nil {
		return err
	}

	// delete instance network
	for _, instanceNetwork := range instanceNetworkList {
		if err = s.deleteInstanceNetwork(&instanceNetwork); err != nil {
			return err
		}
	}

	// delete instance volume
	for _, record := range instanceVolumeRecords {
		if err = s.deleteInstanceVolume(&record.Volume, &record.InstanceVolume); err != nil {
			return err
		}
	}

	// delete instance security group
	for _, group := range instanceGroupList {
		if err = database.GormTransaction(func(db *gorm.DB) error {
			return s.deleteInstanceSecurityGroup(db, &group)
		}); err != nil {
			return err
		}
	}

	// If there is an instance user script, delete it
	if isExist {
		if err = database.GormTransaction(func(db *gorm.DB) error {
			return s.deleteInstanceUserScript(db, &instanceUserScript)
		}); err != nil {
			return err
		}
	}

	// delete instance
	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(instance).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteInstance] Could not delete cluster(%d) instance(%s). Cause: %+v", s.clusterID, instance.UUID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterInstanceDeleted, &queue.DeleteClusterInstance{
		Cluster:      &model.Cluster{ID: s.clusterID},
		OrigInstance: instance,
	}); err != nil {
		logger.Warnf("[Sync-deleteInstance] Could not publish cluster(%d) instance(%s) deleted message. Cause: %+v", s.clusterID, instance.UUID, err)
	}

	if instanceSpec.ID != 0 {
		_ = s.SyncClusterInstanceSpec(&instanceSpec)
	}

	logger.Infof("[Sync-deleteInstance] Success: cluster(%d) instance(%s)", s.clusterID, instance.UUID)
	return nil
}

func (s *Synchronizer) syncDeletedInstanceList(rsp *client.GetInstanceListResponse) error {
	var err error
	var instanceList []*model.ClusterInstance

	// 동기화되어있는 인스턴스 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterInstance{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_instance.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Find(&instanceList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 인스턴스 제거
	for _, instance := range instanceList {
		exists := false
		for _, clusterInstance := range rsp.ResultList {
			exists = exists || (clusterInstance.Instance.UUID == instance.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteInstance(instance); err != nil {
			return err
		}
	}

	return nil
}

// SyncClusterInstanceList 인스턴스 목록 동기화
func (s *Synchronizer) SyncClusterInstanceList(opts ...Option) error {
	logger.Infof("[SyncClusterInstanceList] Start: cluster(%d)", s.clusterID)

	// 클러스터의 인스턴스 목록 조회
	rsp, err := s.Cli.GetInstanceList()
	if err != nil {
		return err
	}

	// 클러스터에서 삭제된 인스턴스 제거
	if err = s.syncDeletedInstanceList(rsp); err != nil {
		return err
	}

	// 인스턴스 상세 동기화
	for _, clusterInstance := range rsp.ResultList {
		if e := s.SyncClusterInstance(&model.ClusterInstance{UUID: clusterInstance.Instance.UUID}, opts...); e != nil {
			logger.Warnf("[SyncClusterInstanceList] Failed to sync cluster(%d) instance(%s). Cause: %+v", s.clusterID, clusterInstance.Instance.UUID, e)
			err = e
		}
	}

	logger.Infof("[SyncClusterInstanceList] Completed: cluster(%d)", s.clusterID)
	return err
}

// SyncClusterInstance 인스턴스 상세 동기화
func (s *Synchronizer) SyncClusterInstance(instance *model.ClusterInstance, opts ...Option) error {
	logger.Infof("[SyncClusterInstance] Start: cluster(%d) instance(%s)", s.clusterID, instance.UUID)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.TenantUUID != "" {
		if ok, err := s.isDeletedTenant(options.TenantUUID); ok {
			// 삭제된 tenant
			logger.Infof("[SyncClusterInstance] Already sync and deleted tenant: cluster(%d) instance(%s) tenant(%s)",
				s.clusterID, instance.UUID, options.TenantUUID)
			return nil
		} else if err != nil {
			logger.Errorf("[SyncClusterInstance] Could not get tenant from db: cluster(%d) instance(%s) tenant(%s). Cause: %+v",
				s.clusterID, instance.UUID, options.TenantUUID, err)
			return err
		}
		// 삭제안된 tenant 라면 동기화 계속 진행
	}

	rsp, err := s.Cli.GetInstance(client.GetInstanceRequest{Instance: model.ClusterInstance{UUID: instance.UUID}})
	if errors.Equal(err, openstack.ErrNotFoundInstanceHypervisor) || errors.Equal(err, openstack.ErrNotFoundInstanceSpec) {
		// openstack 에서 instance spec 이 없거나 hypervisor 정보가 없는 instance 는 db 에서 삭제
		logger.Warnf("[SyncClusterInstance] Could not sync cluster(%d) instance(%s). Cause: %+v", s.clusterID, instance.UUID, err)
	} else if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	if rsp != nil {
		if rsp.Result.AvailabilityZone.Name == "" {
			logger.Warnf("[SyncClusterInstance] Could not sync cluster(%d) instance(%s). Cause: empty availability zone", s.clusterID, rsp.Result.Instance.UUID)
			return nil
		}

		var t *model.ClusterTenant
		if t, err = s.getTenant(&model.ClusterTenant{UUID: rsp.Result.Tenant.UUID}); err != nil {
			return err
		} else if t == nil {
			logger.Warnf("[SyncClusterInstance] Could not sync cluster(%d) instance(%s). Cause: not found tenant(%s)", s.clusterID, rsp.Result.Instance.UUID, rsp.Result.Tenant.UUID)
			return nil
		}

		var az *model.ClusterAvailabilityZone
		if az, err = s.getAvailabilityZone(&model.ClusterAvailabilityZone{Name: rsp.Result.AvailabilityZone.Name}); err != nil {
			return err
		} else if az == nil {
			logger.Warnf("[SyncClusterInstance] Could not sync cluster(%d) instance(%s). Cause: not found availability zone(%s)",
				s.clusterID, rsp.Result.Instance.UUID, rsp.Result.AvailabilityZone.Name)
			return nil
		}

		var h *model.ClusterHypervisor
		if h, err = s.getHypervisor(&model.ClusterHypervisor{UUID: rsp.Result.Hypervisor.UUID}); err != nil {
			return err
		} else if h == nil {
			logger.Warnf("[SyncClusterInstance] Could not sync cluster(%d) instance(%s). Cause: not found hypervisor(%s)", s.clusterID, rsp.Result.Instance.UUID, rsp.Result.Hypervisor.UUID)
			return nil
		}

		var spec *model.ClusterInstanceSpec
		if spec, err = s.getInstanceSpec(&model.ClusterInstanceSpec{UUID: rsp.Result.InstanceSpec.UUID}); err != nil {
			return err
		} else if spec == nil {
			logger.Warnf("[SyncClusterInstance] Could not sync cluster(%d) instance(%s). Cause: not found instance spec(%s)", s.clusterID, rsp.Result.Instance.UUID, rsp.Result.InstanceSpec.UUID)
			return nil
		}

		if rsp.Result.KeyPair != nil {
			k, err := s.getKeypair(&model.ClusterKeypair{Name: rsp.Result.KeyPair.Name})
			if err != nil {
				logger.Warnf("[SyncClusterInstance] Errors occurred during getting cluster(%d) keypair(%s). Cause: %+v", s.clusterID, rsp.Result.KeyPair.Name, err)
			}
			if k == nil {
				logger.Warnf("[SyncClusterInstance] Could not get cluster(%d) keypair(%s). Cause: not found keypair", s.clusterID, rsp.Result.KeyPair.Name)
			} else {
				rsp.Result.Instance.ClusterKeypairID = &k.ID
			}
		}

		rsp.Result.Instance.ClusterTenantID = t.ID
		rsp.Result.Instance.ClusterAvailabilityZoneID = az.ID
		rsp.Result.Instance.ClusterHypervisorID = h.ID
		rsp.Result.Instance.ClusterInstanceSpecID = spec.ID
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterInstance{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_instance.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterInstance{UUID: instance.UUID}).
			First(instance).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		logger.Infof("[SyncClusterInstance] Success - nothing was updated: cluster(%d) instance(%s)", s.clusterID, instance.UUID)
		return nil

	case rsp == nil: // 인스턴스 삭제
		if err = s.deleteInstance(instance); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 인스턴스 추가
		*instance = rsp.Result.Instance

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(instance).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterInstance] Done - added: cluster(%d) instance(%s)", s.clusterID, instance.UUID)

		if err = publishMessage(constant.QueueNoticeClusterInstanceCreated, &queue.CreateClusterInstance{
			Cluster:  &model.Cluster{ID: s.clusterID},
			Instance: instance,
		}); err != nil {
			logger.Warnf("[SyncClusterInstance] Could not publish cluster(%d) instance(%s) created message. Cause: %+v", s.clusterID, instance.UUID, err)
		}

	case options.Force || !compareInstance(instance, &rsp.Result): // 인스턴스 정보 수정
		origInstance := *instance
		rsp.Result.Instance.ID = instance.ID
		*instance = rsp.Result.Instance

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(instance).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterInstance] Done - updated: cluster(%d) instance(%s)", s.clusterID, instance.UUID)

		if err = publishMessage(constant.QueueNoticeClusterInstanceUpdated, &queue.UpdateClusterInstance{
			Cluster:      &model.Cluster{ID: s.clusterID},
			OrigInstance: &origInstance,
			Instance:     instance,
		}); err != nil {
			logger.Warnf("[SyncClusterInstance] Could not publish cluster(%d) instance(%s) updated message. Cause: %+v", s.clusterID, instance.UUID, err)
		}
	}

	// 신규, 수정인 경우 동기화
	if rsp != nil {
		if err = s.syncClusterInstanceVolumeList(instance, rsp.Result.InstanceVolumeList); err != nil {
			logger.Errorf("[SyncClusterInstance] Failed to sync cluster(%d) instance(%s) volume list. Cause: %+v", s.clusterID, instance.UUID, err)
		}

		if err = s.syncClusterInstanceNetworkList(instance, rsp.Result.InstanceNetworkList); err != nil {
			logger.Errorf("[SyncClusterInstance] Failed to sync cluster(%d) instance(%s) network list. Cause: %+v", s.clusterID, instance.UUID, err)
		}

		if err = s.syncClusterInstanceSecurityGroupList(instance, rsp.Result.SecurityGroupList); err != nil {
			logger.Errorf("[SyncClusterInstance] Failed to sync cluster(%d) instance(%s) security group list. Cause: %+v", s.clusterID, instance.UUID, err)
		}

		if err != nil {
			logger.Infof("[SyncClusterInstance] Cluster(%d) instance(%s) synchronization completed.", s.clusterID, instance.UUID)
			return err
		}
	}

	logger.Infof("[SyncClusterInstance] Success: cluster(%d) instance(%s)", s.clusterID, instance.UUID)
	return nil
}

func (s *Synchronizer) getInstanceVolumeRecords(instanceID uint64) ([]instanceVolumeRecord, error) {
	var records []instanceVolumeRecord

	if err := database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterInstanceVolume{}.TableName()).
			Select("*").
			Joins("join cdm_cluster_volume on cdm_cluster_volume.id = cdm_cluster_instance_volume.cluster_volume_id").
			Where(&model.ClusterInstanceVolume{ClusterInstanceID: instanceID}).
			Find(&records).Error
	}); err != nil {
		logger.Errorf("[Sync-getInstanceVolumeRecords] Could not get instance volume records. Cause: %+v", err)
		return nil, errors.UnusableDatabase(err)
	}

	return records, nil
}

func (s *Synchronizer) deleteInstanceVolume(volume *model.ClusterVolume, instanceVolume *model.ClusterInstanceVolume) (err error) {
	logger.Infof("[Sync-deleteInstanceVolume] Start: cluster(%d) instance(%d) volume(%d)", s.clusterID, instanceVolume.ClusterInstanceID, instanceVolume.ClusterVolumeID)

	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(instanceVolume).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteInstanceVolume] Could not delete cluster(%d) instance(%d) volume(%d). Cause: %+v", s.clusterID, instanceVolume.ClusterInstanceID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterInstanceVolumeDetached, &queue.DetachClusterInstanceVolume{
		Cluster:            &model.Cluster{ID: s.clusterID},
		OrigInstanceVolume: instanceVolume,
		OrigVolume:         volume,
	}); err != nil {
		logger.Warnf("[Sync-deleteInstanceVolume] Could not publish cluster cluster(%d) instance(%d) volume(%d) deleted message. Cause: %+v",
			s.clusterID, instanceVolume.ClusterInstanceID, err)
	}

	logger.Infof("[Sync-deleteInstanceVolume] Success: cluster(%d) instance(%d) volume(%d)", s.clusterID, instanceVolume.ClusterInstanceID, instanceVolume.ClusterVolumeID)
	return nil
}

func (s *Synchronizer) syncDeletedInstanceVolumeList(instance *model.ClusterInstance, instanceVolumeResults []client.GetInstanceVolumeResult) error {
	var err error

	instanceVolumeRecords, err := s.getInstanceVolumeRecords(instance.ID)
	if err != nil {
		return err
	}

	// 인스턴스에서 detach 된 볼륨 제거
	for _, record := range instanceVolumeRecords {
		exists := false
		for _, instanceVolumeResult := range instanceVolumeResults {
			exists = exists || (instanceVolumeResult.Volume.UUID == record.Volume.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteInstanceVolume(&record.Volume, &record.InstanceVolume); err != nil {
			return err
		}
	}

	return nil
}

// syncClusterInstanceVolumeList 인스턴스 볼륨 목록 동기화
func (s *Synchronizer) syncClusterInstanceVolumeList(instance *model.ClusterInstance, results []client.GetInstanceVolumeResult) error {
	var err error

	// 인스턴스 에서 detach 된 볼륨 삭제
	if err = s.syncDeletedInstanceVolumeList(instance, results); err != nil {
		return err
	}

	// 인스턴스 볼륨 상세 동기화
	for _, r := range results {
		var v *model.ClusterVolume
		if v, err = s.getVolume(&model.ClusterVolume{UUID: r.Volume.UUID}); err != nil {
			return err
		}

		// db 내 인스턴스 볼륨 조회
		var instanceVolume model.ClusterInstanceVolume
		err = database.Execute(func(db *gorm.DB) error {
			return db.Where(&model.ClusterInstanceVolume{ClusterInstanceID: instance.ID, ClusterVolumeID: v.ID}).First(&instanceVolume).Error
		})
		switch {
		case err == gorm.ErrRecordNotFound:
			// 새로운 인스턴스 볼륨 추가
			instanceVolume.ClusterInstanceID = instance.ID
			instanceVolume.ClusterVolumeID = v.ID
			instanceVolume.DevicePath = r.InstanceVolume.DevicePath

			if err = database.GormTransaction(func(db *gorm.DB) error {
				return db.Save(&instanceVolume).Error
			}); err != nil {
				return errors.UnusableDatabase(err)
			}

			logger.Infof("[syncClusterInstanceVolumeList] Done - added: cluster(%d) instance(%s) volume(%s)", s.clusterID, instance.Name, r.Volume.UUID)

			if err = publishMessage(constant.QueueNoticeClusterInstanceVolumeAttached, &queue.AttachClusterInstanceVolume{
				Cluster:        &model.Cluster{ID: s.clusterID},
				InstanceVolume: &instanceVolume,
				Volume:         v,
			}); err != nil {
				logger.Warnf("[syncClusterInstanceVolumeList] Could not publish cluster(%d) instance(%s) volume(%s) created message. Cause: %+v",
					s.clusterID, instance.Name, r.Volume.UUID, err)
			}

		case err == nil && instanceVolume.DevicePath != r.InstanceVolume.DevicePath:
			// 인스턴스 볼륨 수정
			instanceVolume.DevicePath = r.InstanceVolume.DevicePath

			if err = database.GormTransaction(func(db *gorm.DB) error {
				return db.Save(&instanceVolume).Error
			}); err != nil {
				return errors.UnusableDatabase(err)
			}

			logger.Infof("[syncClusterInstanceVolumeList] Done - updated: cluster(%d) instance(%s) volume(%s)", s.clusterID, instance.Name, r.Volume.UUID)

			if err = publishMessage(constant.QueueNoticeClusterInstanceVolumeUpdated, &queue.UpdateClusterInstanceVolume{
				Cluster:        &model.Cluster{ID: s.clusterID},
				InstanceVolume: &instanceVolume,
				Volume:         v,
			}); err != nil {
				logger.Warnf("[syncClusterInstanceVolumeList] Could not publish cluster(%d) instance(%s) volume(%s) updated message. Cause: %+v",
					s.clusterID, instance.Name, r.Volume.UUID, err)
			}

		case err != nil:
			return errors.UnusableDatabase(err)
		}
	}

	return nil
}

func (s *Synchronizer) getInstanceNetworkRecords(instanceID uint64) ([]instanceNetworkRecord, error) {
	var records []instanceNetworkRecord

	if err := database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterInstanceNetwork{}.TableName()).
			Select("*").
			Joins("join cdm_cluster_network on cdm_cluster_network.id = cdm_cluster_instance_network.cluster_network_id").
			Joins("left join cdm_cluster_floating_ip on cdm_cluster_floating_ip.id = cdm_cluster_instance_network.cluster_floating_ip_id").
			Where(&model.ClusterInstanceNetwork{ClusterInstanceID: instanceID}).
			Find(&records).Error
	}); err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return records, nil
}

func (s *Synchronizer) deleteInstanceNetworkFloatingIP(instanceNetwork *model.ClusterInstanceNetwork) (err error) {
	logger.Infof("[Sync-deleteInstanceNetworkFloatingIP] Start: cluster(%d) instance(%d)", s.clusterID, instanceNetwork.ClusterInstanceID)

	origInstanceNetwork := new(model.ClusterInstanceNetwork)
	if err := copier.CopyWithOption(origInstanceNetwork, instanceNetwork, copier.Option{DeepCopy: true}); err != nil {
		return errors.Unknown(err)
	}
	instanceNetwork.ClusterFloatingIPID = nil

	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Save(&instanceNetwork).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteInstanceNetworkFloatingIP] Could not update cluster(%d) instance(%d) by deleted floating IP. Cause: %+v",
			s.clusterID, instanceNetwork.ClusterInstanceID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterInstanceNetworkUpdated, &queue.UpdateClusterInstanceNetwork{
		Cluster:             &model.Cluster{ID: s.clusterID},
		OrigInstanceNetwork: origInstanceNetwork,
		InstanceNetwork:     instanceNetwork,
	}); err != nil {
		logger.Warnf("[Sync-deleteInstanceNetworkFloatingIP] Could not publish cluster(%d) instance(%d) network updated message. Cause: %+v",
			s.clusterID, instanceNetwork.ClusterInstanceID, err)
	}

	logger.Infof("[Sync-deleteInstanceNetworkFloatingIP] Success: cluster(%d) instance(%d)", s.clusterID, instanceNetwork.ClusterInstanceID)
	return nil
}

func (s *Synchronizer) deleteInstanceNetwork(instanceNetwork *model.ClusterInstanceNetwork) (err error) {
	logger.Infof("[Sync-deleteInstanceNetwork] Start: cluster(%d) instance(%d) network(%d)", s.clusterID, instanceNetwork.ClusterInstanceID, instanceNetwork.ClusterNetworkID)

	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(instanceNetwork).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteInstanceNetwork] Could not delete cluster(%d) instance(%d) network(%d). Cause: %+v",
			s.clusterID, instanceNetwork.ClusterInstanceID, instanceNetwork.ClusterNetworkID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterInstanceNetworkDetached, &queue.DetachClusterInstanceNetwork{
		Cluster:             &model.Cluster{ID: s.clusterID},
		OrigInstanceNetwork: instanceNetwork,
	}); err != nil {
		logger.Warnf("[Sync-deleteInstanceNetwork] Could not publish cluster(%d) instance(%d) network(%d) deleted message. Cause: %+v",
			s.clusterID, instanceNetwork.ClusterInstanceID, instanceNetwork.ClusterNetworkID, err)
	}

	logger.Infof("[Sync-deleteInstanceNetwork] Success: cluster(%d) instance(%d) network(%d)", s.clusterID, instanceNetwork.ClusterInstanceID, instanceNetwork.ClusterNetworkID)
	return nil
}

func (s *Synchronizer) syncDeletedInstanceNetworkList(instance *model.ClusterInstance, instanceNetworkResults []client.GetInstanceNetworkResult) error {
	var err error

	instanceNetworkRecords, err := s.getInstanceNetworkRecords(instance.ID)
	if err != nil {
		return err
	}

	// 인스턴스에서 detach 된 네트워크 제거
	for _, record := range instanceNetworkRecords {
		exists := false
		for _, instanceNetworkResult := range instanceNetworkResults {
			exists = exists || (instanceNetworkResult.Network.UUID == record.Network.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteInstanceNetwork(&record.InstanceNetwork); err != nil {
			return err
		}
	}

	return nil
}

func compareInstanceNetwork(dbInstanceNetwork *model.ClusterInstanceNetwork, result *client.GetInstanceNetworkResult) bool {
	if dbInstanceNetwork.DhcpFlag != result.InstanceNetwork.DhcpFlag {
		return false
	}

	if dbInstanceNetwork.IPAddress != result.InstanceNetwork.IPAddress {
		return false
	}

	if (result.FloatingIP == nil && dbInstanceNetwork.ClusterFloatingIPID != nil) ||
		(result.FloatingIP != nil && dbInstanceNetwork.ClusterFloatingIPID == nil) {
		return false
	}

	if result.FloatingIP != nil && dbInstanceNetwork.ClusterFloatingIPID != nil {
		var floatingIP model.ClusterFloatingIP
		if err := database.Execute(func(db *gorm.DB) error {
			return db.First(&floatingIP, &model.ClusterFloatingIP{ID: *dbInstanceNetwork.ClusterFloatingIPID}).Error
		}); err != nil {
			return false
		}

		if floatingIP.UUID != result.FloatingIP.UUID {
			return false
		}
	}

	return true
}

// syncClusterInstanceNetworkList 인스턴스 네트워크 목록 동기화
func (s *Synchronizer) syncClusterInstanceNetworkList(instance *model.ClusterInstance, results []client.GetInstanceNetworkResult) error {
	var err error

	if err = s.syncDeletedInstanceNetworkList(instance, results); err != nil {
		return err
	}

	// 인스턴스 네트워크 상세 동기화
	for _, r := range results {
		var network *model.ClusterNetwork
		var subnet *model.ClusterSubnet

		network, err = s.getNetwork(&model.ClusterNetwork{UUID: r.Network.UUID})
		if err != nil {
			return err
		}

		subnet, err = s.getSubnet(&model.ClusterSubnet{UUID: r.Subnet.UUID})
		if err != nil {
			return err
		}

		var f *model.ClusterFloatingIP
		if r.FloatingIP != nil {
			f, err = s.getFloatingIP(&model.ClusterFloatingIP{UUID: r.FloatingIP.UUID})
			if err != nil {
				return err
			}
		}

		// db 내 인스턴스 네트워크 조회
		var instanceNetwork model.ClusterInstanceNetwork
		err = database.Execute(func(db *gorm.DB) error {
			return db.Where(&model.ClusterInstanceNetwork{ClusterInstanceID: instance.ID, ClusterNetworkID: network.ID}).First(&instanceNetwork).Error
		})

		switch {
		case err == gorm.ErrRecordNotFound:
			// 새로운 인스턴스 네트워크 추가
			instanceNetwork.ClusterInstanceID = instance.ID
			instanceNetwork.ClusterNetworkID = network.ID
			instanceNetwork.ClusterSubnetID = subnet.ID
			if r.FloatingIP != nil {
				instanceNetwork.ClusterFloatingIPID = &f.ID
			}

			instanceNetwork.DhcpFlag = r.InstanceNetwork.DhcpFlag
			instanceNetwork.IPAddress = r.InstanceNetwork.IPAddress

			if err = database.GormTransaction(func(db *gorm.DB) error {
				return db.Save(&instanceNetwork).Error
			}); err != nil {
				return errors.UnusableDatabase(err)
			}

			logger.Infof("[syncClusterInstanceNetworkList] Done - added: cluster(%d) instance(%s) network(%s)", s.clusterID, instance.Name, r.Network.UUID)

			if err = publishMessage(constant.QueueNoticeClusterInstanceNetworkAttached, &queue.AttachClusterInstanceNetwork{
				Cluster:         &model.Cluster{ID: s.clusterID},
				InstanceNetwork: &instanceNetwork,
			}); err != nil {
				logger.Warnf("[syncClusterInstanceNetworkList] Could not publish cluster(%d) instance(%s) network(%s) created message. Cause: %+v",
					s.clusterID, instance.Name, r.Network.UUID, err)
			}

		case err == nil && !compareInstanceNetwork(&instanceNetwork, &r):
			origInstanceNetwork := new(model.ClusterInstanceNetwork)
			if err := copier.CopyWithOption(origInstanceNetwork, instanceNetwork, copier.Option{DeepCopy: true}); err != nil {
				return errors.Unknown(err)
			}

			// 인스턴스 네트워크 수정
			instanceNetwork.DhcpFlag = r.InstanceNetwork.DhcpFlag
			instanceNetwork.IPAddress = r.InstanceNetwork.IPAddress
			if r.FloatingIP != nil {
				instanceNetwork.ClusterFloatingIPID = &f.ID
			} else {
				instanceNetwork.ClusterFloatingIPID = nil
			}

			if err = database.GormTransaction(func(db *gorm.DB) error {
				return db.Save(&instanceNetwork).Error
			}); err != nil {
				return errors.UnusableDatabase(err)
			}

			logger.Infof("[syncClusterInstanceNetworkList] Done - updated: cluster(%d) instance(%s) network(%s)", s.clusterID, instance.Name, r.Network.UUID)

			if err = publishMessage(constant.QueueNoticeClusterInstanceNetworkUpdated, &queue.UpdateClusterInstanceNetwork{
				Cluster:             &model.Cluster{ID: s.clusterID},
				OrigInstanceNetwork: origInstanceNetwork,
				InstanceNetwork:     &instanceNetwork,
			}); err != nil {
				logger.Warnf("[syncClusterInstanceNetworkList] Could not publish cluster(%d) instance(%s) network(%s) updated message. Cause: %+v",
					s.clusterID, instance.Name, r.Network.UUID, err)
			}

		case err != nil:
			return errors.UnusableDatabase(err)
		}

	}

	return nil
}

// syncClusterInstanceSecurityGroupList 인스턴스 보안 그룹 목록 동기화
func (s *Synchronizer) syncClusterInstanceSecurityGroupList(instance *model.ClusterInstance, results []model.ClusterSecurityGroup) (err error) {
	if err = database.GormTransaction(func(db *gorm.DB) error {
		// 인스턴스 보안 그룹 목록 전체 삭제
		return db.Where(&model.ClusterInstanceSecurityGroup{ClusterInstanceID: instance.ID}).Delete(&model.ClusterInstanceSecurityGroup{}).Error
	}); err != nil {
		logger.Errorf("[syncClusterInstanceSecurityGroupList] Could not sync cluster(%d) instance(%s) security groups. Cause: %+v", s.clusterID, instance.Name, err)
		return errors.UnusableDatabase(err)
	}

	// 인스턴스 보안 그룹 상세 동기화
	for _, result := range results {
		sg, err := s.getSecurityGroup(&model.ClusterSecurityGroup{UUID: result.UUID})
		if err != nil {
			return err
		}

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(&model.ClusterInstanceSecurityGroup{ClusterInstanceID: instance.ID, ClusterSecurityGroupID: sg.ID}).Error
		}); err != nil {
			logger.Errorf("[syncClusterInstanceSecurityGroupList] Could not sync cluster(%d) instance(%s) security group(%d). Cause: %+v", s.clusterID, instance.Name, sg.ID, err)
			return errors.UnusableDatabase(err)
		}
	}

	return nil
}
