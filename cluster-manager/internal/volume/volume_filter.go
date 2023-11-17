package volume

import (
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

type volumeFilter interface {
	Apply(*gorm.DB) (*gorm.DB, error)
}

type volumeTenantIDFilter struct {
	ID uint64
}

func (f *volumeTenantIDFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cluster_tenant_id = ?", f.ID), nil
}

type volumeStorageIDFilter struct {
	ID uint64
}

func (f *volumeStorageIDFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cluster_storage_id = ?", f.ID), nil
}

type volumeInstanceIDFilter struct {
	DB *gorm.DB
	ID uint64
}

func (f *volumeInstanceIDFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	ci := model.ClusterInstance{ID: f.ID}
	volumes, err := ci.GetVolumes(f.DB)
	if err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	var id []uint64
	for _, v := range volumes {
		id = append(id, v.ID)
	}

	return db.Where("cdm_cluster_volume.id in (?)", id), nil
}

type uuidFilter struct {
	UUID string
}

func (f *uuidFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_volume.uuid = ?", f.UUID), nil
}

type nameFilter struct {
	Name string
}

func (f *nameFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_volume.name LIKE ?", "%"+f.Name+"%"), nil
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

func makeVolumeFilter(db *gorm.DB, req *cms.ClusterVolumeListRequest) []volumeFilter {
	var filters []volumeFilter

	if req.ClusterTenantId != 0 {
		filters = append(filters, &volumeTenantIDFilter{ID: req.ClusterTenantId})
	}

	if req.ClusterStorageId != 0 {
		filters = append(filters, &volumeStorageIDFilter{ID: req.ClusterStorageId})
	}

	if req.ClusterInstanceId != 0 {
		filters = append(filters, &volumeInstanceIDFilter{DB: db, ID: req.ClusterInstanceId})
	}

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
