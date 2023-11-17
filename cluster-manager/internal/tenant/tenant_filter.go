package tenant

import (
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
)

type tenantFilter interface {
	Apply(*gorm.DB) (*gorm.DB, error)
}

type uuidFilter struct {
	UUID string
}

func (f *uuidFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("uuid = ?", f.UUID), nil
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

func makeClusterTenantListFilter(req *cms.ClusterTenantListRequest) []tenantFilter {
	var filters []tenantFilter

	if len(req.Uuid) != 0 {
		filters = append(filters, &uuidFilter{UUID: req.Uuid})
	}

	if req.GetOffset() != nil || req.GetLimit() != nil {
		filters = append(filters, &paginationFilter{Offset: req.GetOffset(), Limit: req.GetLimit()})
	}

	return filters
}
