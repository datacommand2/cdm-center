package network

import (
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
)

type networkFilter interface {
	Apply(*gorm.DB) (*gorm.DB, error)
}

type clusterNetworkExternalOnlyFilter struct {
}

func (f *clusterNetworkExternalOnlyFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("external_flag = true"), nil
}

type clusterNetworkTenantFilter struct {
	id uint64
}

func (f *clusterNetworkTenantFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cluster_tenant_id = ?", f.id), nil
}

type uuidFilter struct {
	UUID string
}

func (f *uuidFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_network.uuid = ?", f.UUID), nil
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

func makeClusterNetworkFilters(req *cms.ClusterNetworkListRequest) []networkFilter {
	var filters []networkFilter
	if req.ClusterTenantId != 0 {
		filters = append(filters, &clusterNetworkTenantFilter{req.ClusterTenantId})
	}

	if len(req.Uuid) != 0 {
		filters = append(filters, &uuidFilter{UUID: req.Uuid})
	}

	if req.ExternalOnly {
		filters = append(filters, &clusterNetworkExternalOnlyFilter{})
	}

	if req.GetOffset() != nil || req.GetLimit() != nil {
		filters = append(filters, &paginationFilter{Offset: req.GetOffset(), Limit: req.GetLimit()})
	}

	return filters
}
