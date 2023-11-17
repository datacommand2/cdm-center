package instance

import (
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

type instanceFilter interface {
	Apply(*gorm.DB) (*gorm.DB, error)
}

type clusterIDListFiler struct {
	IDList []uint64
}

func (c *clusterIDListFiler) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_tenant.cluster_id IN (?)", c.IDList), nil
}

type tenantFilter struct {
	ID uint64
}

func (f *tenantFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_instance.cluster_tenant_id = ?", f.ID), nil
}

type availabilityZoneFilter struct {
	ID uint64
}

func (f *availabilityZoneFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_instance.cluster_availability_zone_id = ?", f.ID), nil
}

type hypervisorFilter struct {
	ID uint64
}

func (f *hypervisorFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_instance.cluster_hypervisor_id = ?", f.ID), nil
}

type volumeFilter struct {
	DB *gorm.DB
	ID uint64
}

func (f *volumeFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	var relations []model.ClusterInstanceVolume
	if err := f.DB.Where("cluster_volume_id = ?", f.ID).Find(&relations).Error; err != nil {
		return nil, err
	}

	var id []uint64
	for _, rel := range relations {
		id = append(id, rel.ClusterInstanceID)
	}

	return db.Where("cdm_cluster_instance.id IN (?)", id), nil
}

type uuidFilter struct {
	UUID string
}

func (f *uuidFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_instance.uuid = ?", f.UUID), nil
}

type nameFilter struct {
	Name string
}

func (f *nameFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cdm_cluster_instance.name LIKE ?", "%"+f.Name+"%"), nil
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

func makeClusterInstanceListFilters(db *gorm.DB, req *cms.ClusterInstanceListRequest) []instanceFilter {
	var filters []instanceFilter

	if req.ClusterTenantId != 0 {
		filters = append(filters, &tenantFilter{ID: req.ClusterTenantId})
	}

	if req.ClusterAvailabilityZoneId != 0 {
		filters = append(filters, &availabilityZoneFilter{ID: req.ClusterAvailabilityZoneId})
	}

	if req.ClusterHypervisorId != 0 {
		filters = append(filters, &hypervisorFilter{ID: req.ClusterHypervisorId})
	}

	if req.ClusterVolumeId != 0 {
		filters = append(filters, &volumeFilter{DB: db, ID: req.ClusterVolumeId})
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

func makeClusterInstanceNumberFilters(clusterIDList []uint64) []instanceFilter {
	return []instanceFilter{
		&clusterIDListFiler{IDList: clusterIDList},
	}
}
