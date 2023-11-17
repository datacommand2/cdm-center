package spec

import (
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

type instanceSpecFilter interface {
	Apply(*gorm.DB) (*gorm.DB, error)
}

type uuidFilter struct {
	UUID string
}

func (f *uuidFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_instance_spec.uuid = ?", f.UUID), nil
}

type nameFilter struct {
	Name string
}

func (f *nameFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_instance_spec.name LIKE ?", "%"+f.Name+"%"), nil
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

func makeClusterInstanceSpecListFilters(req *cms.ClusterInstanceSpecListRequest) []instanceSpecFilter {
	var filters []instanceSpecFilter

	if len(req.Uuid) != 0 {
		filters = append(filters, &uuidFilter{UUID: req.Uuid})
	}

	if utf8.RuneCountInString(req.Name) != 0 {
		filters = append(filters, &nameFilter{Name: req.Name})
	}

	if req.GetOffset() != nil || req.GetLimit() != nil {
		filters = append(filters, &paginationFilter{Offset: req.GetOffset(), Limit: req.GetLimit()})
	}

	return filters
}
