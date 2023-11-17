package availabilityzone

import (
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
)

type availabilityZoneFilter interface {
	Apply(*gorm.DB) (*gorm.DB, error)
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

func makeAvailabilityZoneListFilter(req *cms.ClusterAvailabilityZoneListRequest) []availabilityZoneFilter {
	var filters []availabilityZoneFilter

	if req.GetOffset() != nil || req.GetLimit() != nil {
		filters = append(filters, &paginationFilter{Offset: req.GetOffset(), Limit: req.GetLimit()})
	}

	return filters
}
