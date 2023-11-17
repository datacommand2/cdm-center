package router

import (
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
)

type routerFilter interface {
	Apply(*gorm.DB) (*gorm.DB, error)
}

type clusterTenantIDFilter struct {
	ID uint64
}

func (f *clusterTenantIDFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cluster_tenant_id = ?", f.ID), nil
}

type clusterNetworkFilter struct {
	DB *gorm.DB
	ID uint64
}

func (f *clusterNetworkFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	var routingInterface []model.ClusterNetworkRoutingInterface

	err := f.DB.Table(model.ClusterNetworkRoutingInterface{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_subnet on cdm_cluster_subnet.id = cdm_cluster_routing_interface.cluster_subnet_id").
		Joins("join cdm_cluster_network on cdm_cluster_network.id = cdm_cluster_subnet.cluster_network_id").
		Where("cdm_cluster_network.external_flag = false").
		Where(&model.ClusterNetwork{ID: f.ID}).
		Find(&routingInterface).Error

	if err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	var ids []uint64
	for _, ri := range routingInterface {
		ids = append(ids, ri.ClusterRouterID)
	}

	return db.Where("cdm_cluster_router.id in (?)", ids), nil
}

type uuidFilter struct {
	UUID string
}

func (f *uuidFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_router.uuid = ?", f.UUID), nil
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

func makeClusterRouterListFilters(db *gorm.DB, req *cms.ClusterRouterListRequest) []routerFilter {
	var filters []routerFilter

	if req.ClusterTenantId != 0 {
		filters = append(filters, &clusterTenantIDFilter{ID: req.ClusterTenantId})
	}

	if req.ClusterNetworkId != 0 {
		filters = append(filters, &clusterNetworkFilter{DB: db, ID: req.ClusterNetworkId})
	}

	if len(req.Uuid) != 0 {
		filters = append(filters, &uuidFilter{UUID: req.Uuid})
	}

	if req.GetOffset() != nil || req.GetLimit() != nil {
		filters = append(filters, &paginationFilter{Offset: req.GetOffset(), Limit: req.GetLimit()})
	}

	return filters
}
