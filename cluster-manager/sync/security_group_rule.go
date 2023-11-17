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

func compareSecurityGroupRule(dbSecurityGroupRule *model.ClusterSecurityGroupRule, result *client.GetSecurityGroupRuleResult) bool {
	if dbSecurityGroupRule.UUID != result.SecurityGroupRule.UUID {
		return false
	}

	if dbSecurityGroupRule.Direction != result.SecurityGroupRule.Direction {
		return false
	}

	if !reflect.DeepEqual(dbSecurityGroupRule.Description, result.SecurityGroupRule.Description) {
		return false
	}

	if !reflect.DeepEqual(dbSecurityGroupRule.NetworkCidr, result.SecurityGroupRule.NetworkCidr) {
		return false
	}

	if !reflect.DeepEqual(dbSecurityGroupRule.PortRangeMin, result.SecurityGroupRule.PortRangeMin) {
		return false
	}

	if !reflect.DeepEqual(dbSecurityGroupRule.PortRangeMax, result.SecurityGroupRule.PortRangeMax) {
		return false
	}

	if !reflect.DeepEqual(dbSecurityGroupRule.Protocol, result.SecurityGroupRule.Protocol) {
		return false
	}

	if dbSecurityGroupRule.EtherType != result.SecurityGroupRule.EtherType {
		return false
	}

	var securityGroup model.ClusterSecurityGroup
	var rSecurityGroup model.ClusterSecurityGroup

	if result.RemoteSecurityGroup == nil {
		if dbSecurityGroupRule.RemoteSecurityGroupID == nil {
			return true
		}

		return false
	}

	if dbSecurityGroupRule.RemoteSecurityGroupID == nil {
		return false
	}

	if err := database.Execute(func(db *gorm.DB) error {
		if err := db.Find(&securityGroup, &model.ClusterSecurityGroup{ID: dbSecurityGroupRule.SecurityGroupID}).Error; err != nil {
			return err
		}

		if err := db.Find(&rSecurityGroup, &model.ClusterSecurityGroup{ID: *dbSecurityGroupRule.RemoteSecurityGroupID}).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return false
	}

	if securityGroup.UUID != result.SecurityGroup.UUID {
		return false
	}

	if rSecurityGroup.UUID != result.RemoteSecurityGroup.UUID {
		return false
	}

	return true
}

func (s *Synchronizer) deleteSecurityGroupRule(rule *model.ClusterSecurityGroupRule) (err error) {
	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(rule).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteSecurityGroupRule] Could not delete cluster(%d) security group rule(%s). Cause: %+v", s.clusterID, rule.UUID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterSecurityGroupRuleDeleted, &queue.DeleteClusterSecurityGroupRule{
		Cluster:               &model.Cluster{ID: s.clusterID},
		OrigSecurityGroupRule: rule,
	}); err != nil {
		logger.Warnf("[Sync-deleteSecurityGroupRule] Could not publish cluster(%d) security group rule(%s) deleted message. Cause: %+v", s.clusterID, rule.UUID, err)
	}

	// 해당 security group rule 을 쓰고 있는 shared task 삭제
	migrator.ClearSharedTask(s.clusterID, rule.ID, rule.UUID, drConstant.MigrationTaskTypeCreateSecurityGroupRule)

	return nil
}

func (s *Synchronizer) syncDeletedSecurityGroupRuleList(securityGroup *model.ClusterSecurityGroup, securityGroupRules []model.ClusterSecurityGroupRule) (err error) {
	var ruleList []*model.ClusterSecurityGroupRule

	// 동기화되어있는 보안 그룹 규칙 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterSecurityGroupRule{SecurityGroupID: securityGroup.ID}).Find(&ruleList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 보안 그룹 규칙 제거
	for _, rule := range ruleList {
		exists := false
		for _, clusterSecurityGroupRule := range securityGroupRules {
			exists = exists || (clusterSecurityGroupRule.UUID == rule.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteSecurityGroupRule(rule); err != nil {
			return err
		}
	}

	return nil
}

// syncClusterSecurityGroupRuleList 보안 그룹 규칙 목록 동기화
func (s *Synchronizer) syncClusterSecurityGroupRuleList(securityGroup *model.ClusterSecurityGroup, rules []model.ClusterSecurityGroupRule, opts ...Option) (err error) {
	// 클러스터에서 삭제된 보안 그룹 규칙 제거
	if err = s.syncDeletedSecurityGroupRuleList(securityGroup, rules); err != nil {
		return err
	}

	// 보안 그룹 규칙 상세 동기화
	for _, rule := range rules {
		if e := s.SyncClusterSecurityGroupRule(&model.ClusterSecurityGroupRule{UUID: rule.UUID}, opts...); e != nil {
			logger.Warnf("[syncClusterSecurityGroupRuleList] Failed to sync cluster(%d) security group rule(%s). Cause: %+v", s.clusterID, rule.UUID, e)
			err = e
		}
	}

	return err
}

// SyncClusterSecurityGroupRule 보안 그룹 규칙 상세 동기화
func (s *Synchronizer) SyncClusterSecurityGroupRule(rule *model.ClusterSecurityGroupRule, opts ...Option) error {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.TenantUUID != "" {
		if ok, err := s.isDeletedTenant(options.TenantUUID); ok {
			// 삭제된 tenant
			logger.Infof("[SyncClusterSecurityGroupRule] Already sync and deleted tenant: cluster(%d) securityGroupRule(%s) tenant(%s)",
				s.clusterID, rule.UUID, options.TenantUUID)
			return nil
		} else if err != nil {
			logger.Errorf("[SyncClusterSecurityGroupRule] Could not get tenant from db: cluster(%d) securityGroupRule(%s) tenant(%s). Cause: %+v",
				s.clusterID, rule.UUID, options.TenantUUID, err)
			return err
		}
		// 삭제안된 tenant 라면 동기화 계속 진행
	}

	rsp, err := s.Cli.GetSecurityGroupRule(client.GetSecurityGroupRuleRequest{SecurityGroupRule: model.ClusterSecurityGroupRule{UUID: rule.UUID}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	var sg *model.ClusterSecurityGroup
	if rsp != nil {
		sg, err = s.getSecurityGroup(&model.ClusterSecurityGroup{UUID: rsp.Result.SecurityGroup.UUID}, opts...)
		if err != nil {
			return err
		}

		rsp.Result.SecurityGroupRule.SecurityGroupID = sg.ID
	}

	err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterSecurityGroupRule{}.TableName()).
			Joins("join cdm_cluster_security_group on cdm_cluster_security_group_rule.security_group_id = cdm_cluster_security_group.id").
			Joins("join cdm_cluster_tenant on cdm_cluster_security_group.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterSecurityGroupRule{UUID: rule.UUID}).First(rule).Error
	})
	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		return nil

	case rsp == nil: // 보안 그룹 규칙 삭제
		if err = s.deleteSecurityGroupRule(rule); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 보안 그룹 규칙 추가
		if rsp.Result.RemoteSecurityGroup != nil {
			r, err := s.getSecurityGroup(&model.ClusterSecurityGroup{UUID: rsp.Result.RemoteSecurityGroup.UUID}, opts...)
			if err != nil {
				return err
			}
			rsp.Result.SecurityGroupRule.RemoteSecurityGroupID = &r.ID
		}

		*rule = rsp.Result.SecurityGroupRule

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(rule).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterSecurityGroupRule] Done - added: cluster(%d) security group rule(%s)", s.clusterID, rule.UUID)

		if err = publishMessage(constant.QueueNoticeClusterSecurityGroupRuleCreated, &queue.CreateClusterSecurityGroupRule{
			Cluster:           &model.Cluster{ID: s.clusterID},
			SecurityGroupRule: rule,
		}); err != nil {
			logger.Warnf("[SyncClusterSecurityGroupRule] Could not publish cluster(%d) security group rule(%s) created message. Cause: %+v", s.clusterID, rule.UUID, err)
		}

	case options.Force || !compareSecurityGroupRule(rule, &rsp.Result): //보안 그룹 규칙 정보 수정
		if rsp.Result.RemoteSecurityGroup != nil {
			r, err := s.getSecurityGroup(&model.ClusterSecurityGroup{UUID: rsp.Result.RemoteSecurityGroup.UUID}, opts...)
			if err != nil {
				return err
			}
			rsp.Result.SecurityGroupRule.RemoteSecurityGroupID = &r.ID
		}

		rsp.Result.SecurityGroupRule.ID = rule.ID
		*rule = rsp.Result.SecurityGroupRule

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(rule).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterSecurityGroupRule] Done - updated: cluster(%d) security group rule(%s)", s.clusterID, rule.UUID)

		if err = publishMessage(constant.QueueNoticeClusterSecurityGroupRuleUpdated, &queue.UpdateClusterSecurityGroupRule{
			Cluster:           &model.Cluster{ID: s.clusterID},
			SecurityGroupRule: rule,
		}); err != nil {
			logger.Warnf("[SyncClusterSecurityGroupRule] Could not publish cluster(%d) security group rule(%s) updated message. Cause: %+v", s.clusterID, rule.UUID, err)
		}
	}

	return nil
}
