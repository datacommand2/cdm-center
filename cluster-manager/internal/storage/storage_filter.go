package storage

import (
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
)

type storageFilter interface {
	Apply(*gorm.DB) (*gorm.DB, error)
}

type storageTypeCodeFilter struct {
	StorageType string
}

func (f *storageTypeCodeFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("type_code = ?", f.StorageType), nil
}

type storageNameFilter struct {
	Name string
}

func (f *storageNameFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("name LIKE ?", "%"+f.Name+"%"), nil
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

func makeClusterStorageListFilter(req *cms.ClusterStorageListRequest) []storageFilter {
	var filters []storageFilter

	if req.TypeCode != "" {
		filters = append(filters, &storageTypeCodeFilter{StorageType: req.TypeCode})
	}

	if req.Name != "" {
		filters = append(filters, &storageNameFilter{Name: req.Name})
	}

	if req.GetOffset() != nil || req.GetLimit() != nil {
		filters = append(filters, &paginationFilter{Offset: req.GetOffset(), Limit: req.GetLimit()})
	}

	return filters
}
