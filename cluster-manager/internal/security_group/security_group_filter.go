package securitygroup

import (
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
)

type securityGroupFilter interface {
	Apply(*gorm.DB) (*gorm.DB, error)
}

type clusterTenantIDFilter struct {
	ID uint64
}

func (f *clusterTenantIDFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cluster_tenant_id = ?", f.ID), nil
}

type clusterInstanceIDFilter struct {
	ID uint64
	DB *gorm.DB
}

func (f *clusterInstanceIDFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	var relations []model.ClusterInstanceSecurityGroup
	if err := f.DB.Where("cluster_instance_id = ?", f.ID).Find(&relations).Error; err != nil {
		return nil, err
	}

	var id []uint64
	for _, rel := range relations {
		id = append(id, rel.ClusterSecurityGroupID)
	}

	return db.Where("cdm_cluster_security_group.id IN (?)", id), nil
}

type uuidFilter struct {
	UUID string
}

func (f *uuidFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_security_group.uuid = ?", f.UUID), nil
}

type paginationFilter struct {
	Offset *wrappers.UInt64Value
	Limit  *wrappers.UInt64Value
}

func (f *paginationFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	cond := db
	if f.Offset != nil {
		cond = cond.Offset(f.Offset.GetValue())
	}

	if f.Limit != nil {
		cond = cond.Limit(f.Limit.GetValue())
	}

	return cond, nil
}

func makeSecurityGroupListFilter(db *gorm.DB, req *cms.ClusterSecurityGroupListRequest) []securityGroupFilter {
	var filters []securityGroupFilter

	if req.ClusterTenantId != 0 {
		filters = append(filters, &clusterTenantIDFilter{ID: req.ClusterTenantId})
	}

	if req.ClusterInstanceId != 0 {
		filters = append(filters, &clusterInstanceIDFilter{DB: db, ID: req.ClusterInstanceId})
	}

	if len(req.Uuid) != 0 {
		filters = append(filters, &uuidFilter{UUID: req.Uuid})
	}

	if req.GetOffset() != nil || req.GetLimit() != nil {
		filters = append(filters, &paginationFilter{Offset: req.GetOffset(), Limit: req.GetLimit()})
	}

	return filters
}
