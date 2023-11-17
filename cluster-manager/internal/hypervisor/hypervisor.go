package hypervisor

import (
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/metadata"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"

	"context"
)

type hypervisorRecord struct {
	Hypervisor       model.ClusterHypervisor       `gorm:"embedded"`
	AvailabilityZone model.ClusterAvailabilityZone `gorm:"embedded"`
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

func getHypervisorRecordList(db *gorm.DB, clusterID uint64, filters ...hypervisorFilter) ([]hypervisorRecord, error) {
	var err error
	var cond = db

	for _, f := range filters {
		cond, err = f.Apply(cond)
		if err != nil {
			return nil, err
		}
	}

	var records []hypervisorRecord

	err = cond.Table(model.ClusterHypervisor{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_availability_zone on cdm_cluster_availability_zone.id = cdm_cluster_hypervisor.cluster_availability_zone_id").
		Where(&model.ClusterHypervisor{ClusterID: clusterID}).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}

func getHypervisorsPagination(db *gorm.DB, clusterID uint64, filters ...hypervisorFilter) (*cms.Pagination, error) {
	var err error
	var offset, limit uint64
	var cond = db.Where(&model.ClusterHypervisor{ClusterID: clusterID})

	for _, f := range filters {
		if _, ok := f.(*paginationFilter); ok {
			offset = f.(*paginationFilter).Offset.GetValue()
			limit = f.(*paginationFilter).Limit.GetValue()
			continue
		}

		cond, err = f.Apply(cond)
		if err != nil {
			return nil, errors.UnusableDatabase(err)
		}
	}

	var total uint64
	if err = cond.Model(&model.ClusterHypervisor{}).Count(&total).Error; err != nil {
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

func getHypervisorRecord(db *gorm.DB, cid, nid uint64) (*hypervisorRecord, error) {
	var record hypervisorRecord
	err := db.Table(model.ClusterHypervisor{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_availability_zone on cdm_cluster_availability_zone.id = cdm_cluster_hypervisor.cluster_availability_zone_id").
		Where(&model.ClusterHypervisor{ID: nid, ClusterID: cid}).
		First(&record).Error

	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterHypervisor(nid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &record, nil
}

// GetList 클러스터 하이퍼바이저 목록 조회
func GetList(ctx context.Context, req *cms.ClusterHypervisorListRequest) ([]*cms.ClusterHypervisor, *cms.Pagination, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateHypervisorListRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[Hypervisor-GetList] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	var c *model.Cluster
	var hypervisors []hypervisorRecord

	var filters []hypervisorFilter
	if err = database.Execute(func(db *gorm.DB) error {
		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Hypervisor-GetList] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		filters = makeClusterHypervisorListFilters(req)

		if hypervisors, err = getHypervisorRecordList(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[Hypervisor-GetList] Could not get the cluster(%d) hypervisor record list. Cause: %+v", req.ClusterId, err)
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	var rsp []*cms.ClusterHypervisor
	for _, h := range hypervisors {
		var hypervisor = cms.ClusterHypervisor{
			Cluster:          new(cms.Cluster),
			AvailabilityZone: new(cms.ClusterAvailabilityZone),
		}

		if err = hypervisor.SetFromModel(&h.Hypervisor); err != nil {
			return nil, nil, err
		}
		if err = hypervisor.Cluster.SetFromModel(c); err != nil {
			return nil, nil, err
		}
		if err = hypervisor.AvailabilityZone.SetFromModel(&h.AvailabilityZone); err != nil {
			return nil, nil, err
		}

		rsp = append(rsp, &hypervisor)
	}

	var pagination *cms.Pagination
	if err = database.Execute(func(db *gorm.DB) error {
		if pagination, err = getHypervisorsPagination(db, c.ID, filters...); err != nil {
			logger.Errorf("[Hypervisor-GetList] Could not get the cluster(%d) hypervisors pagination. Cause: %+v", req.ClusterId, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}

	return rsp, pagination, nil
}

// Get 클러스터 하이퍼바이저 조회
func Get(ctx context.Context, req *cms.ClusterHypervisorRequest) (*cms.ClusterHypervisor, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateHypervisorRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[Hypervisor-Get] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	var c *model.Cluster
	var h *hypervisorRecord

	if err = database.Execute(func(db *gorm.DB) error {
		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Hypervisor-Get] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		if h, err = getHypervisorRecord(db, req.ClusterId, req.ClusterHypervisorId); err != nil {
			logger.Errorf("[Hypervisor-Get] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var rsp = cms.ClusterHypervisor{
		Cluster:          new(cms.Cluster),
		AvailabilityZone: new(cms.ClusterAvailabilityZone),
	}

	if err = rsp.SetFromModel(&h.Hypervisor); err != nil {
		return nil, err
	}

	if err = rsp.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}

	if err = rsp.AvailabilityZone.SetFromModel(&h.AvailabilityZone); err != nil {
		return nil, err
	}

	return &rsp, nil
}

func getHypervisorExtraInfo(hypervisor *cms.ClusterHypervisor) *model.ClusterHypervisor {
	var sshAccount, sshPassword, agentVersion *string
	var agentInstalledAt, agentLastUpgradedAt *int64

	sshPort := hypervisor.GetSshPort()
	agentPort := hypervisor.GetAgentPort()

	if hypervisor.GetSshAccount() != "" {
		sshAccount = &hypervisor.SshAccount
	}

	if hypervisor.GetSshPassword() != "" {
		sshPassword = &hypervisor.SshPassword
	}

	if hypervisor.GetAgentVersion() != "" {
		agentVersion = &hypervisor.AgentVersion
	}

	if hypervisor.GetAgentInstalledAt() != 0 {
		agentInstalledAt = &hypervisor.AgentInstalledAt
	}

	if hypervisor.GetAgentLastUpgradedAt() != 0 {
		agentLastUpgradedAt = &hypervisor.AgentLastUpgradedAt
	}

	return &model.ClusterHypervisor{
		SSHPort:             &sshPort,
		SSHAccount:          sshAccount,
		SSHPassword:         sshPassword,
		AgentPort:           &agentPort,
		AgentVersion:        agentVersion,
		AgentInstalledAt:    agentInstalledAt,
		AgentLastUpgradedAt: agentLastUpgradedAt,
	}
}

// Update 클러스터 하이퍼바이저 추가 정보 수정
func Update(ctx context.Context, req *cms.UpdateClusterHypervisorRequest) (*cms.ClusterHypervisor, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateUpdateHypervisorRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[Hypervisor-Update] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	if err = checkValidHypervisor(req.GetHypervisor()); err != nil {
		logger.Errorf("[Hypervisor-Update] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	h := getHypervisorExtraInfo(req.GetHypervisor())
	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Model(&model.ClusterHypervisor{}).Where(&model.ClusterHypervisor{ClusterID: req.GetClusterId(), ID: req.GetClusterHypervisorId()}).Update(h).Error
	}); err != nil {
		logger.Errorf("[Hypervisor-Update] Could not update the cluster(%d) hypervisor(%s). Cause: %+v", req.ClusterId, req.Hypervisor.GetHostname(), err)
		return nil, errors.UnusableDatabase(err)
	}

	return Get(ctx, &cms.ClusterHypervisorRequest{ClusterId: req.GetClusterId(), ClusterHypervisorId: req.GetClusterHypervisorId()})
}
