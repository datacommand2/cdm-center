package hypervisor

import (
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
)

type hypervisorFilter interface {
	Apply(*gorm.DB) (*gorm.DB, error)
}

type hypervisorHostnameFilter struct {
	Hostname string
}

func (f *hypervisorHostnameFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("hostname LIKE ?", "%"+f.Hostname+"%"), nil
}

type hypervisorIPAddressFilter struct {
	IPAddress string
}

func (f *hypervisorIPAddressFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("ip_address = ?", f.IPAddress), nil
}

type availabilityZoneFilter struct {
	ID uint64
}

func (f *availabilityZoneFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("cluster_availability_zone_id = ?", f.ID), nil
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

func makeClusterHypervisorListFilters(req *cms.ClusterHypervisorListRequest) []hypervisorFilter {
	var filters []hypervisorFilter

	if req.Hostname != "" {
		filters = append(filters, &hypervisorHostnameFilter{Hostname: req.Hostname})
	}

	if req.IpAddress != "" {
		filters = append(filters, &hypervisorIPAddressFilter{IPAddress: req.IpAddress})
	}

	if req.ClusterAvailabilityZoneId != 0 {
		filters = append(filters, &availabilityZoneFilter{ID: req.ClusterAvailabilityZoneId})
	}

	if req.GetOffset() != nil || req.GetLimit() != nil {
		filters = append(filters, &paginationFilter{Offset: req.GetOffset(), Limit: req.GetLimit()})
	}

	return filters
}
