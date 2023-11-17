package securitygroup

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
)

type securityGroupRecord struct {
	SecurityGroup model.ClusterSecurityGroup `gorm:"embedded"`
	Tenant        model.ClusterTenant        `gorm:"embedded"`
}

// CreateSecurityGroup 보안 그룹 생성
func CreateSecurityGroup(req *cms.CreateClusterSecurityGroupRequest) (*cms.ClusterSecurityGroup, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateSecurityGroup] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err = validateCreateSecurityGroup(req); err != nil {
		logger.Errorf("[CreateSecurityGroup] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateSecurityGroup] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[CreateSecurityGroup] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateSecurityGroup] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return nil, err
	}

	mSecurityGroup, err := req.GetSecurityGroup().Model()
	if err != nil {
		return nil, err
	}

	sgRsp, err := cli.CreateSecurityGroup(client.CreateSecurityGroupRequest{
		Tenant:        *mTenant,
		SecurityGroup: *mSecurityGroup,
	})
	if err != nil {
		logger.Errorf("[CreateSecurityGroup] Could not create the cluster security group. Cause: %+v", err)
		return nil, err
	}

	var rollback = true
	defer func() {
		if rollback {
			if err := cli.DeleteSecurityGroup(client.DeleteSecurityGroupRequest{
				Tenant:        *mTenant,
				SecurityGroup: sgRsp.SecurityGroup,
			}); err != nil {
				logger.Warnf("[CreateSecurityGroup] Could not delete cluster security group. Cause: %+v", err)
			}
		}
	}()

	var cSG cms.ClusterSecurityGroup
	if err := cSG.SetFromModel(&sgRsp.SecurityGroup); err != nil {
		return nil, err
	}

	logger.Infof("[CreateSecurityGroup] rsp: %+v", *sgRsp)

	rollback = false
	return &cSG, nil
}

// CreateSecurityGroupRule 보안 그룹 규칙 생성
func CreateSecurityGroupRule(req *cms.CreateClusterSecurityGroupRuleRequest) (*cms.ClusterSecurityGroupRule, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateSecurityGroupRule] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err = validateCreateSecurityGroupRule(req); err != nil {
		logger.Errorf("[CreateSecurityGroupRule] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateSecurityGroupRule] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[CreateSecurityGroupRule] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateSecurityGroupRule] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return nil, err
	}

	mSecurityGroup, err := req.GetSecurityGroup().Model()
	if err != nil {
		return nil, err
	}

	mRule, err := req.GetSecurityGroupRule().Model()
	if err != nil {
		return nil, err
	}

	var remote *model.ClusterSecurityGroup
	if req.GetSecurityGroupRule().GetRemoteSecurityGroup().GetUuid() != "" {
		remote = &model.ClusterSecurityGroup{
			UUID: req.GetSecurityGroupRule().GetRemoteSecurityGroup().GetUuid(),
		}
	}

	ruleRsp, err := cli.CreateSecurityGroupRule(client.CreateSecurityGroupRuleRequest{
		Tenant:              *mTenant,
		SecurityGroup:       *mSecurityGroup,
		Rule:                *mRule,
		RemoteSecurityGroup: remote,
	})
	if err != nil {
		logger.Errorf("[CreateSecurityGroup] Could not create the cluster security group rule. Cause: %+v", err)
		return nil, err
	}

	var rollback = true
	defer func() {
		if rollback {
			if err := cli.DeleteSecurityGroupRule(client.DeleteSecurityGroupRuleRequest{
				Tenant:        *mTenant,
				SecurityGroup: *mSecurityGroup,
				Rule:          ruleRsp.Rule,
			}); err != nil {
				logger.Warnf("[CreateSecurityGroupRule] Could not delete cluster security group rule. Cause: %+v", err)
			}
		}
	}()

	var cRule cms.ClusterSecurityGroupRule
	if err := cRule.SetFromModel(&ruleRsp.Rule); err != nil {
		return nil, err
	}

	if ruleRsp.RemoteSecurityGroup != nil {
		var rsGroup cms.ClusterSecurityGroup
		if err := rsGroup.SetFromModel(ruleRsp.RemoteSecurityGroup); err != nil {
			return nil, err
		}
		cRule.RemoteSecurityGroup = &rsGroup
	}

	rollback = false
	return &cRule, nil
}

// DeleteSecurityGroup 보안 그룹 삭제
func DeleteSecurityGroup(req *cms.DeleteClusterSecurityGroupRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteSecurityGroup] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteSecurityGroup(req); err != nil {
		logger.Errorf("[DeleteSecurityGroup] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteSecurityGroup] Could not create client. Cause: %+v", err)
		return err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[DeleteSecurityGroup] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteSecurityGroup] Could not close client. Cause: %+v", err)
		}
	}()

	tenant, err := req.GetTenant().Model()
	if err != nil {
		return err
	}
	group, err := req.GetSecurityGroup().Model()
	if err != nil {
		return err
	}

	for _, rule := range req.GetSecurityGroup().GetRules() {
		groupRule, err := rule.Model()
		if err != nil {
			return err
		}
		if err := cli.DeleteSecurityGroupRule(client.DeleteSecurityGroupRuleRequest{
			Tenant:        *tenant,
			SecurityGroup: *group,
			Rule:          *groupRule,
		}); err != nil {
			logger.Errorf("[DeleteSecurityGroup] Could not delete cluster security group rule. Cause: %+v", err)
			return err
		}
	}

	return cli.DeleteSecurityGroup(client.DeleteSecurityGroupRequest{
		Tenant:        *tenant,
		SecurityGroup: *group,
	})
}

func getCluster(ctx context.Context, db *gorm.DB, id uint64) (*model.Cluster, error) {
	tid, _ := metadata.GetTenantID(ctx)

	var m model.Cluster
	err := db.Model(&m).
		Select("id, name, type_code, remarks").
		Where(&model.Cluster{ID: id, TenantID: tid}).
		First(&m).Error

	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundCluster(id)
	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}
	return &m, nil
}

func getSecurityGroups(db *gorm.DB, cid uint64, filters ...securityGroupFilter) ([]*securityGroupRecord, error) {
	var err error
	cond := db
	for _, f := range filters {
		if cond, err = f.Apply(cond); err != nil {
			return nil, err
		}
	}

	var records []*securityGroupRecord
	if err = cond.Table(model.ClusterSecurityGroup{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_security_group.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: cid}).
		Find(&records).Error; err != nil {
		return nil, err
	}

	return records, nil
}

func getSecurityGroupPagination(db *gorm.DB, clusterID uint64, filters ...securityGroupFilter) (*cms.Pagination, error) {
	var err error
	var offset, limit uint64

	var conditions = db

	for _, f := range filters {
		if _, ok := f.(*paginationFilter); ok {
			offset = f.(*paginationFilter).Offset.GetValue()
			limit = f.(*paginationFilter).Limit.GetValue()
			continue
		}

		conditions, err = f.Apply(conditions)
		if err != nil {
			return nil, errors.UnusableDatabase(err)
		}
	}

	var total uint64
	if err = conditions.Table(model.ClusterSecurityGroup{}.TableName()).
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_security_group.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: clusterID}).
		Count(&total).Error; err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	if limit == 0 {
		return &cms.Pagination{
			Page:       &wrappers.UInt64Value{Value: 1},
			TotalPage:  &wrappers.UInt64Value{Value: 1},
			TotalItems: &wrappers.UInt64Value{Value: total},
		}, nil
	}

	return &cms.Pagination{
		Page:       &wrappers.UInt64Value{Value: offset/limit + 1},
		TotalPage:  &wrappers.UInt64Value{Value: (total + limit - 1) / limit},
		TotalItems: &wrappers.UInt64Value{Value: total},
	}, nil
}

// GetList 클러스터 보안그룹 목록을 조회한다.
func GetList(ctx context.Context, req *cms.ClusterSecurityGroupListRequest) ([]*cms.ClusterSecurityGroup, *cms.Pagination, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateGetSecurityGroupList(ctx, db, req)
	}); err != nil {
		logger.Errorf("[SecurityGroup-GetList] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	var c *model.Cluster
	var filters []securityGroupFilter
	var securityGroups []*securityGroupRecord

	if err = database.Execute(func(db *gorm.DB) error {
		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[SecurityGroup-GetList] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		filters = makeSecurityGroupListFilter(db, req)

		if securityGroups, err = getSecurityGroups(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[SecurityGroup-GetList] Could not get the cluster(%d) security groups. Cause: %+v", req.ClusterId, err)
			return errors.UnusableDatabase(err)
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}

	var rsp []*cms.ClusterSecurityGroup
	for _, sg := range securityGroups {
		var securityGroup = cms.ClusterSecurityGroup{
			Cluster: new(cms.Cluster),
			Tenant:  new(cms.ClusterTenant),
		}

		if err = securityGroup.SetFromModel(&sg.SecurityGroup); err != nil {
			return nil, nil, err
		}
		if err = securityGroup.Cluster.SetFromModel(c); err != nil {
			return nil, nil, err
		}
		if err = securityGroup.Tenant.SetFromModel(&sg.Tenant); err != nil {
			return nil, nil, err
		}

		rsp = append(rsp, &securityGroup)
	}

	var pagination *cms.Pagination
	if err = database.Execute(func(db *gorm.DB) error {
		if pagination, err = getSecurityGroupPagination(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[SecurityGroup-GetList] Could not get the cluster(%d) security group pagination. Cause: %+v", req.ClusterId, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}

	return rsp, pagination, nil
}

func getSecurityGroup(db *gorm.DB, cid, id uint64) (*securityGroupRecord, error) {
	var m securityGroupRecord

	err := db.Table(model.ClusterSecurityGroup{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_security_group.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: cid}).
		Where(&model.ClusterSecurityGroup{ID: id}).First(&m).Error

	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterSecurityGroup(id)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &m, nil
}

func getSecurityGroupByUUID(db *gorm.DB, cid uint64, uuid string) (*securityGroupRecord, error) {
	var m securityGroupRecord

	err := db.Table(model.ClusterSecurityGroup{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_security_group.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: cid}).
		Where(&model.ClusterSecurityGroup{UUID: uuid}).First(&m).Error

	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterSecurityGroupByUUID(uuid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &m, nil
}

// Get 클러스터 보안그룹을 조회한다.
func Get(ctx context.Context, req *cms.ClusterSecurityGroupRequest) (*cms.ClusterSecurityGroup, error) {
	var err error
	var c *model.Cluster
	var sg *securityGroupRecord

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateGetSecurityGroup(ctx, db, req); err != nil {
			logger.Errorf("[SecurityGroup-Get] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[SecurityGroup-Get] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		if sg, err = getSecurityGroup(db, req.ClusterId, req.ClusterSecurityGroupId); err != nil {
			logger.Errorf("[SecurityGroup-Get] Could not get the cluster(%d) security group(%d). Cause: %+v", req.ClusterId, req.ClusterSecurityGroupId, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	securityGroup := cms.ClusterSecurityGroup{
		Cluster: new(cms.Cluster),
		Tenant:  new(cms.ClusterTenant),
	}

	if err = securityGroup.SetFromModel(&sg.SecurityGroup); err != nil {
		return nil, err
	}
	if err = securityGroup.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}
	if err = securityGroup.Tenant.SetFromModel(&sg.Tenant); err != nil {
		return nil, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		var rules []model.ClusterSecurityGroupRule
		if rules, err = sg.SecurityGroup.GetSecurityGroupRules(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		for _, r := range rules {
			var rule cms.ClusterSecurityGroupRule
			if err = rule.SetFromModel(&r); err != nil {
				return err
			}

			if r.RemoteSecurityGroupID != nil {
				rsg, err := r.GetRemoteSecurityGroup(db)
				switch {
				case err == gorm.ErrRecordNotFound:
					return errors.Unknown(err)

				case err != nil:
					return errors.UnusableDatabase(err)
				}

				rule.RemoteSecurityGroup = new(cms.ClusterSecurityGroup)
				if err = rule.RemoteSecurityGroup.SetFromModel(rsg); err != nil {
					return err
				}
			}
			securityGroup.Rules = append(securityGroup.Rules, &rule)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &securityGroup, nil
}

// GetByUUID uuid 를 통해 클러스터 보안그룹을 조회한다.
func GetByUUID(ctx context.Context, req *cms.ClusterSecurityGroupByUUIDRequest) (*cms.ClusterSecurityGroup, error) {
	var err error
	var c *model.Cluster

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateGetSecurityGroupByUUID(ctx, db, req); err != nil {
			logger.Errorf("[SecurityGroup-GetByUUID] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[SecurityGroup-GetByUUID] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var sg *securityGroupRecord
	if err = database.Execute(func(db *gorm.DB) error {
		if sg, err = getSecurityGroupByUUID(db, req.ClusterId, req.Uuid); err != nil {
			logger.Errorf("[SecurityGroup-GetByUUID] Could not get the cluster(%d) security group by uuid(%s). Cause: %+v", req.ClusterId, req.Uuid, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	securityGroup := cms.ClusterSecurityGroup{
		Cluster: new(cms.Cluster),
		Tenant:  new(cms.ClusterTenant),
	}

	if err = securityGroup.SetFromModel(&sg.SecurityGroup); err != nil {
		return nil, err
	}
	if err = securityGroup.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}
	if err = securityGroup.Tenant.SetFromModel(&sg.Tenant); err != nil {
		return nil, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		var rules []model.ClusterSecurityGroupRule
		if rules, err = sg.SecurityGroup.GetSecurityGroupRules(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		for _, r := range rules {
			var rule cms.ClusterSecurityGroupRule
			if err = rule.SetFromModel(&r); err != nil {
				return err
			}

			if r.RemoteSecurityGroupID != nil {
				rsg, err := r.GetRemoteSecurityGroup(db)
				switch {
				case err == gorm.ErrRecordNotFound:
					return errors.Unknown(err)

				case err != nil:
					return errors.UnusableDatabase(err)
				}

				rule.RemoteSecurityGroup = new(cms.ClusterSecurityGroup)
				if err = rule.RemoteSecurityGroup.SetFromModel(rsg); err != nil {
					return err
				}
			}
			securityGroup.Rules = append(securityGroup.Rules, &rule)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &securityGroup, nil
}

// CheckIsExistSecurityGroup 보안그룹 이름을 통한 클러스터 보안그룹 존재유무 확인
func CheckIsExistSecurityGroup(ctx context.Context, req *cms.CheckIsExistByNameRequest) (bool, error) {
	var (
		err     error
		existed bool
	)

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateCheckIsExistSecurityGroup(ctx, db, req); err != nil {
			return err
		}

		if existed, err = checkIsExistSecurityGroupByName(db, req.ClusterId, req.GetName()); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return false, err
	}

	return existed, nil
}

func checkIsExistSecurityGroupByName(db *gorm.DB, tid uint64, name string) (bool, error) {
	var cnt int
	err := db.Table(model.ClusterSecurityGroup{}.TableName()).Where(&model.ClusterSecurityGroup{ClusterTenantID: tid, Name: name}).
		Count(&cnt).Error

	if err != nil {
		return false, errors.UnusableDatabase(err)
	}

	return cnt > 0, nil
}
