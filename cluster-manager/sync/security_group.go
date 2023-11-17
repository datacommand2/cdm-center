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

func compareSecurityGroup(dbSecurityGroup *model.ClusterSecurityGroup, result *client.GetSecurityGroupResult) bool {
	if dbSecurityGroup.UUID != result.SecurityGroup.UUID {
		return false
	}

	if dbSecurityGroup.Name != result.SecurityGroup.Name {
		return false
	}

	if !reflect.DeepEqual(dbSecurityGroup.Description, result.SecurityGroup.Description) {
		return false
	}

	var t model.ClusterTenant
	if err := database.Execute(func(db *gorm.DB) error {
		return db.Find(&t, &model.ClusterTenant{ID: dbSecurityGroup.ClusterTenantID}).Error
	}); err != nil {
		return false
	}

	if t.UUID != result.Tenant.UUID {
		return false
	}

	return true
}

func (s *Synchronizer) getSecurityGroup(securityGroup *model.ClusterSecurityGroup, opts ...Option) (*model.ClusterSecurityGroup, error) {
	var err error
	var sg model.ClusterSecurityGroup

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterSecurityGroup{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_security_group.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterSecurityGroup{UUID: securityGroup.UUID}).First(&sg).Error
	}); err == nil {
		return &sg, nil
	}

	if err = s.SyncClusterSecurityGroup(&model.ClusterSecurityGroup{UUID: securityGroup.UUID}, opts...); err != nil {
		return nil, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterSecurityGroup{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_security_group.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterSecurityGroup{UUID: securityGroup.UUID}).First(&sg).Error
	}); err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return &sg, nil
}

func (s *Synchronizer) deleteSecurityGroup(sg *model.ClusterSecurityGroup) (err error) {
	logger.Infof("[Sync-deleteSecurityGroup] Start: cluster(%d) security group(%s)", s.clusterID, sg.UUID)

	var instanceGroupList []model.ClusterInstanceSecurityGroup
	var ruleList []model.ClusterSecurityGroupRule

	if err = database.Execute(func(db *gorm.DB) error {
		if err = db.Where(&model.ClusterInstanceSecurityGroup{ClusterSecurityGroupID: sg.ID}).Find(&instanceGroupList).Error; err != nil {
			return err
		}

		if err = db.Where(&model.ClusterSecurityGroupRule{SecurityGroupID: sg.ID}).Find(&ruleList).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// delete instance security group
	for _, group := range instanceGroupList {
		if err = database.GormTransaction(func(db *gorm.DB) error {
			return s.deleteInstanceSecurityGroup(db, &group)
		}); err != nil {
			return err
		}
	}

	// delete security group rule
	for _, rule := range ruleList {
		if err = s.deleteSecurityGroupRule(&rule); err != nil {
			return err
		}
	}

	// delete security group
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Delete(sg).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteSecurityGroup] Could not delete cluster(%d) security group(%s). Cause: %+v",
			s.clusterID, sg.UUID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterSecurityGroupDeleted, &queue.DeleteClusterSecurityGroup{
		Cluster:           &model.Cluster{ID: s.clusterID},
		OrigSecurityGroup: sg,
	}); err != nil {
		logger.Warnf("[Sync-deleteSecurityGroup] Could not publish cluster(%d) security group(%s) deleted message. Cause: %+v",
			s.clusterID, sg.UUID, err)
	}

	// 해당 security group 을 쓰고 있는 shared task 삭제
	migrator.ClearSharedTask(s.clusterID, sg.ID, sg.UUID, drConstant.MigrationTaskTypeCreateSecurityGroup)

	logger.Infof("[Sync-deleteSecurityGroup] Success: cluster(%d) security group(%s)", s.clusterID, sg.UUID)
	return nil
}

func (s *Synchronizer) syncDeletedSecurityGroupList(rsp *client.GetSecurityGroupListResponse) (err error) {
	var securityGroupList []*model.ClusterSecurityGroup

	// 동기화되어있는 보안 그룹 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterSecurityGroup{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_security_group.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Find(&securityGroupList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 보안 그룹 제거
	for _, securityGroup := range securityGroupList {
		exists := false
		for _, clusterSecurityGroup := range rsp.ResultList {
			exists = exists || (clusterSecurityGroup.SecurityGroup.UUID == securityGroup.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteSecurityGroup(securityGroup); err != nil {
			return err
		}
	}

	return nil
}

// SyncClusterSecurityGroupList 보안 그룹 목록 동기화
func (s *Synchronizer) SyncClusterSecurityGroupList(opts ...Option) error {
	logger.Infof("[SyncClusterSecurityGroupList] Start: cluster(%d)", s.clusterID)

	// 클러스터의 보안 그룹 목록 조회
	rsp, err := s.Cli.GetSecurityGroupList()
	if err != nil {
		return err
	}

	// 클러스터에서 삭제된 보안 그룹 제거
	if err = s.syncDeletedSecurityGroupList(rsp); err != nil {
		return err
	}

	// 보안 그룹 상세 동기화
	for _, securityGroup := range rsp.ResultList {
		if err = s.SyncClusterSecurityGroup(&model.ClusterSecurityGroup{UUID: securityGroup.SecurityGroup.UUID}, opts...); err != nil {
			logger.Warnf("[SyncClusterSecurityGroupList] Failed to sync cluster(%d) security group(%s). Cause: %+v", s.clusterID, securityGroup.SecurityGroup.UUID, err)
		}
	}

	logger.Infof("[SyncClusterSecurityGroupList] Success: cluster(%d)", s.clusterID)
	return err
}

// SyncClusterSecurityGroup 보안 그룹 상세 동기화
func (s *Synchronizer) SyncClusterSecurityGroup(securityGroup *model.ClusterSecurityGroup, opts ...Option) error {
	logger.Infof("[SyncClusterSecurityGroup] Start: cluster(%d) security group(%s)", s.clusterID, securityGroup.UUID)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.TenantUUID != "" {
		if ok, err := s.isDeletedTenant(options.TenantUUID); ok {
			// 삭제된 tenant
			logger.Infof("[SyncClusterSecurityGroup] Already sync and deleted tenant: cluster(%d) securityGroup(%s) tenant(%s)",
				s.clusterID, securityGroup.UUID, options.TenantUUID)
			return nil
		} else if err != nil {
			logger.Errorf("[SyncClusterSecurityGroup] Could not get tenant from db: cluster(%d) securityGroup(%s) tenant(%s). Cause: %+v",
				s.clusterID, securityGroup.UUID, options.TenantUUID, err)
			return err
		}
		// 삭제안된 tenant 라면 동기화 계속 진행
	}

	rsp, err := s.Cli.GetSecurityGroup(client.GetSecurityGroupRequest{SecurityGroup: model.ClusterSecurityGroup{UUID: securityGroup.UUID}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterSecurityGroup{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_security_group.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterSecurityGroup{UUID: securityGroup.UUID}).
			First(securityGroup).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		logger.Infof("[SyncClusterSecurityGroup] Success - nothing was updated: cluster(%d) securityGroup(%s)", s.clusterID, securityGroup.UUID)
		return nil

	case rsp == nil: // 보안 그룹 삭제
		if err = s.deleteSecurityGroup(securityGroup); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 보안 그룹 추가
		var tenant *model.ClusterTenant
		if tenant, err = s.getTenant(&model.ClusterTenant{UUID: rsp.Result.Tenant.UUID}, opts...); err != nil {
			return err
		} else if tenant == nil {
			logger.Warnf("[SyncClusterSecurityGroup] Could not sync cluster(%d) security group(%s). Cause: not found tenant(%s)",
				s.clusterID, securityGroup.UUID, rsp.Result.Tenant.UUID)
			return nil
		}

		rsp.Result.SecurityGroup.ClusterTenantID = tenant.ID
		*securityGroup = rsp.Result.SecurityGroup

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(securityGroup).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterSecurityGroup] Done - added: cluster(%d) security group(%s)", s.clusterID, securityGroup.UUID)

		if err = publishMessage(constant.QueueNoticeClusterSecurityGroupCreated, &queue.CreateClusterSecurityGroup{
			Cluster:       &model.Cluster{ID: s.clusterID},
			SecurityGroup: securityGroup,
		}); err != nil {
			logger.Warnf("[SyncClusterSecurityGroup] Could not publish cluster(%d) security group(%s) created message. Cause: %+v",
				s.clusterID, rsp.Result.SecurityGroup.UUID, rsp.Result.Tenant.UUID)
		}

	case options.Force || !compareSecurityGroup(securityGroup, &rsp.Result): //보안 그룹 정보 수정
		var tenant *model.ClusterTenant
		if tenant, err = s.getTenant(&model.ClusterTenant{UUID: rsp.Result.Tenant.UUID}, opts...); err != nil {
			return err
		} else if tenant == nil {
			logger.Warnf("[SyncClusterSecurityGroup] Could not sync cluster(%d) security group(%s). Cause: not found tenant(%s)",
				s.clusterID, securityGroup.UUID, rsp.Result.Tenant.UUID)
			return nil
		}

		rsp.Result.SecurityGroup.ID = securityGroup.ID
		rsp.Result.SecurityGroup.ClusterTenantID = tenant.ID
		*securityGroup = rsp.Result.SecurityGroup

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(securityGroup).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterSecurityGroup] Done - updated: cluster(%d) security group(%s)", s.clusterID, securityGroup.UUID)

		if err = publishMessage(constant.QueueNoticeClusterSecurityGroupUpdated, &queue.UpdateClusterSecurityGroup{
			Cluster:       &model.Cluster{ID: s.clusterID},
			SecurityGroup: securityGroup,
		}); err != nil {
			logger.Warnf("[SyncClusterSecurityGroup] Could not publish cluster(%d) security group(%s) updated message. Cause: %+v",
				s.clusterID, securityGroup.UUID, err)
		}
	}

	// 추가, 수정일 경우 보안 그룹 규칙 동기화
	if rsp != nil {
		if err = s.syncClusterSecurityGroupRuleList(securityGroup, rsp.Result.RuleList, opts...); err != nil {
			return err
		}
	}

	logger.Infof("[SyncClusterSecurityGroup] Success: cluster(%d) security group(%s)", s.clusterID, securityGroup.UUID)
	return nil
}
