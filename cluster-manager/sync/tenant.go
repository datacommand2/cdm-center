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

func compareTenant(dbTenant, clusterTenant model.ClusterTenant) bool {
	if dbTenant.UUID != clusterTenant.UUID {
		return false
	}

	if dbTenant.Name != clusterTenant.Name {
		return false
	}

	if dbTenant.Enabled != clusterTenant.Enabled {
		return false
	}

	if !reflect.DeepEqual(dbTenant.Description, clusterTenant.Description) {
		return false
	}

	return true
}

// TODO: 이 기능을 사용하는 모든 부분들 refactoring 필요: tenant 조회를 중복으로 하고있음.
// isDeletedTenant sync 중 삭제된 tenant 인지 확인하는 함수
// true: 삭제됨. false: 삭제안됨(존재함).
func (s *Synchronizer) isDeletedTenant(uuid string) (bool, error) {
	var t model.ClusterTenant
	err := database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterTenant{ClusterID: s.clusterID, UUID: uuid}).First(&t).Error
	})
	switch {
	case err == gorm.ErrRecordNotFound:
		return true, nil

	case err != nil:
		return false, errors.UnusableDatabase(err)
	}

	return false, nil
}

func (s *Synchronizer) getTenant(tenant *model.ClusterTenant, opts ...Option) (*model.ClusterTenant, error) {
	var err error
	var t model.ClusterTenant

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterTenant{ClusterID: s.clusterID, UUID: tenant.UUID}).First(&t).Error
	}); err == nil {
		return &t, nil
	}

	if err = s.SyncClusterTenant(tenant, opts...); err != nil {
		return nil, err
	}

	err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterTenant{ClusterID: s.clusterID, UUID: tenant.UUID}).First(&t).Error
	})
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, nil

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &t, nil
}

func (s *Synchronizer) deleteTenant(tenant *model.ClusterTenant) (err error) {
	logger.Infof("[Sync-deleteTenant] Start: cluster(%d) tenant(%s)", s.clusterID, tenant.UUID)

	var (
		instanceList      []model.ClusterInstance
		securityGroupList []model.ClusterSecurityGroup
		volumeList        []model.ClusterVolume
		floatingIPList    []model.ClusterFloatingIP
		networkList       []model.ClusterNetwork
		routerList        []model.ClusterRouter
	)

	if err = database.Execute(func(db *gorm.DB) error {
		if err = db.Where(&model.ClusterInstance{ClusterTenantID: tenant.ID}).Find(&instanceList).Error; err != nil {
			return err
		}

		if err = db.Where(&model.ClusterSecurityGroup{ClusterTenantID: tenant.ID}).Find(&securityGroupList).Error; err != nil {
			return err
		}

		if err = db.Where(&model.ClusterVolume{ClusterTenantID: tenant.ID}).Find(&volumeList).Error; err != nil {
			return err
		}

		if err = db.Where(&model.ClusterFloatingIP{ClusterTenantID: tenant.ID}).Find(&floatingIPList).Error; err != nil {
			return err
		}

		if err = db.Where(&model.ClusterNetwork{ClusterTenantID: tenant.ID}).Find(&networkList).Error; err != nil {
			return err
		}

		if err = db.Where(&model.ClusterRouter{ClusterTenantID: tenant.ID}).Find(&routerList).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	for _, instance := range instanceList {
		if err = s.deleteInstance(&instance); err != nil {
			return err
		}
	}

	for _, sg := range securityGroupList {
		if err = s.deleteSecurityGroup(&sg); err != nil {
			return err
		}
	}

	for _, volume := range volumeList {
		if err = s.deleteVolume(&volume); err != nil {
			return err
		}
	}

	for _, floatingIP := range floatingIPList {
		if err = s.deleteFloatingIP(&floatingIP); err != nil {
			return err
		}
	}

	for _, network := range networkList {
		if err = s.deleteNetwork(&network); err != nil {
			return err
		}
	}

	for _, router := range routerList {
		if err = s.deleteRouter(&router); err != nil {
			return err
		}
	}

	if err = database.GormTransaction(func(db *gorm.DB) error {
		if err = db.Where(&model.ClusterQuota{ClusterTenantID: tenant.ID}).Delete(&model.ClusterQuota{}).Error; err != nil {
			logger.Errorf("[Sync-deleteTenant] Could not delete cluster(%d) tenant(%s)'s quotas. Cause: %+v",
				s.clusterID, tenant.UUID, err)
			return errors.UnusableDatabase(err)
		}

		if err = db.Delete(tenant).Error; err != nil {
			logger.Errorf("[Sync-deleteTenant] Could not delete cluster(%d) tenant(%s). Cause: %+v", s.clusterID, tenant.UUID, err)
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return err
	}

	if err = publishMessage(constant.QueueNoticeClusterTenantDeleted, &queue.DeleteClusterTenant{
		Cluster:    &model.Cluster{ID: s.clusterID},
		OrigTenant: tenant,
	}); err != nil {
		logger.Warnf("[Sync-deleteTenant] Could not publish cluster(%d) tenant(%s) deleted message. Cause: %+v",
			s.clusterID, tenant.UUID, err)
	}

	// 해당 tenant 를 쓰고 있는 shared task 삭제
	migrator.ClearSharedTask(s.clusterID, tenant.ID, tenant.UUID, drConstant.MigrationTaskTypeCreateTenant)

	logger.Infof("[Sync-deleteTenant] Success: cluster(%d) tenant(%s)", s.clusterID, tenant.UUID)
	return nil
}

func (s *Synchronizer) syncDeletedTenantList(rsp *client.GetTenantListResponse) error {
	var err error
	var tenantList []*model.ClusterTenant

	if err = database.Execute(func(db *gorm.DB) error {
		// 동기화되어있는 테넌트 목록 조회
		return db.Where(&model.ClusterTenant{ClusterID: s.clusterID}).Find(&tenantList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 테넌트 제거
	for _, tenant := range tenantList {
		exists := false
		for _, clusterTenant := range rsp.ResultList {
			exists = exists || (clusterTenant.Tenant.UUID == tenant.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteTenant(tenant); err != nil {
			return err
		}
	}

	return nil
}

// SyncClusterTenantList 테넌트 목록 동기화
func (s *Synchronizer) SyncClusterTenantList(opts ...Option) error {
	logger.Infof("[SyncClusterTenantList] Start: cluster(%d)", s.clusterID)

	// 클러스터의 테넌트 목록 조회
	rsp, err := s.Cli.GetTenantList()
	if err != nil {
		return err
	}

	// 클러스터에서 삭제된 테넌트 제거
	if err = s.syncDeletedTenantList(rsp); err != nil {
		return err
	}

	// 테넌트 상세 동기화
	for _, clusterTenant := range rsp.ResultList {
		if e := s.SyncClusterTenant(&model.ClusterTenant{UUID: clusterTenant.Tenant.UUID}, opts...); e != nil {
			logger.Warnf("[SyncClusterTenantList] Failed to sync cluster(%d) tenant(%s). Cause: %+v",
				s.clusterID, clusterTenant.Tenant.UUID, e)
			err = e
		}
	}

	logger.Infof("[SyncClusterTenantList] Completed: cluster(%d)", s.clusterID)
	return err
}

// syncClusterTenantQuota 쿼타 상세 동기화
func (s *Synchronizer) syncClusterTenantQuota(db *gorm.DB, tenant *model.ClusterTenant, quota *model.ClusterQuota) (*model.ClusterQuota, error) {
	var err error
	var m model.ClusterQuota
	err = db.Where(&model.ClusterQuota{ClusterTenantID: tenant.ID, Key: quota.Key}).Find(&m).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.UnusableDatabase(err)
	}

	if m.Value == quota.Value {
		return &m, nil
	}

	quota.ClusterTenantID = tenant.ID
	if err = db.Save(quota).Error; err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return quota, nil
}

// SyncClusterTenant 테넌트 상세 동기화
func (s *Synchronizer) SyncClusterTenant(tenant *model.ClusterTenant, opts ...Option) error {
	logger.Infof("[SyncClusterTenant] Start: cluster(%d) tenant(%s)", s.clusterID, tenant.UUID)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	rsp, err := s.Cli.GetTenant(client.GetTenantRequest{Tenant: model.ClusterTenant{UUID: tenant.UUID}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterTenant{ClusterID: s.clusterID, UUID: tenant.UUID}).First(tenant).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		logger.Infof("[SyncClusterTenant] Success - nothing was updated: cluster(%d) tenant(%s)", s.clusterID, tenant.UUID)
		return nil

	case rsp == nil: // 테넌트 삭제
		if err = s.deleteTenant(tenant); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 테넌트 추가
		rsp.Result.Tenant.ClusterID = s.clusterID
		*tenant = rsp.Result.Tenant

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(tenant).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		if err = publishMessage(constant.QueueNoticeClusterTenantCreated, &queue.CreateClusterTenant{
			Cluster: &model.Cluster{ID: s.clusterID},
			Tenant:  tenant,
		}); err != nil {
			logger.Warnf("[SyncClusterTenant] Could not publish cluster(%d) tenant(%s) created message. Cause: %+v",
				s.clusterID, tenant.UUID, err)
		}

	case options.Force || !compareTenant(*tenant, rsp.Result.Tenant): // 테넌트 정보 수정
		origTenant := *tenant
		rsp.Result.Tenant.ID = tenant.ID
		rsp.Result.Tenant.ClusterID = s.clusterID
		*tenant = rsp.Result.Tenant

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(tenant).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		if err = publishMessage(constant.QueueNoticeClusterTenantUpdated, &queue.UpdateClusterTenant{
			Cluster:    &model.Cluster{ID: s.clusterID},
			OrigTenant: &origTenant,
			Tenant:     tenant,
		}); err != nil {
			logger.Warnf("[SyncClusterTenant] Could not publish cluster(%d) tenant(%s) updated message. Cause: %+v",
				s.clusterID, tenant.UUID, err)
		}
	}

	// 신규, 수정인 경우 Quota 동기화
	if rsp != nil {
		if err = database.GormTransaction(func(db *gorm.DB) error {
			for _, quota := range rsp.Result.QuotaList {
				if _, err = s.syncClusterTenantQuota(db, tenant, &quota); err != nil {
					logger.Warnf("[SyncClusterTenant] Failed to sync cluster(%d) tenant(%s) quota. Cause: %+v",
						s.clusterID, tenant.UUID, err)
				}
			}
			return nil
		}); err != nil {
			logger.Infof("[SyncClusterTenant] Cluster(%d) tenant(%s) synchronization completed.", s.clusterID, tenant.UUID)
			return err
		}
	}

	logger.Infof("[SyncClusterTenant] Success: cluster(%d) tenant(%s)", s.clusterID, tenant.UUID)
	return nil
}
