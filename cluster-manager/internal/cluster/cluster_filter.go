package cluster

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	identity "github.com/datacommand2/cdm-cloud/services/identity/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

type clusterFilter interface {
	Apply(*gorm.DB) (*gorm.DB, error)
}

type clusterOwnerGroupListFilter struct {
	OwnerGroups []*identity.Group
}

func (f *clusterOwnerGroupListFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	var idList []uint64
	for _, g := range f.OwnerGroups {
		idList = append(idList, g.Id)
	}

	return db.Where("owner_group_id in (?)", idList), nil
}

type clusterTypeCodeFilter struct {
	TypeCode string
}

func (f *clusterTypeCodeFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("type_code = ?", f.TypeCode), nil
}

type clusterNameFilter struct {
	Name string
}

func (f *clusterNameFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("name LIKE ?", "%"+f.Name+"%"), nil
}

type clusterOwnerGroupFilter struct {
	OwnerGroupID uint64
}

func (f *clusterOwnerGroupFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("owner_group_id = ?", f.OwnerGroupID), nil
}

type clusterTenantIDFilter struct {
	TenantID uint64
}

func (f *clusterTenantIDFilter) Apply(db *gorm.DB) (*gorm.DB, error) {
	return db.Where("tenant_id = ?", f.TenantID), nil
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

func makeClusterFilter(ctx context.Context, req *cms.ClusterListRequest) []clusterFilter {
	var filters []clusterFilter

	// FIXME admin 의 경우 tenant 가 지정 되면 안됨
	// tenant 정책이 정해 진 후 코드 변경이 필요
	tid, _ := metadata.GetTenantID(ctx)
	filters = append(filters, &clusterTenantIDFilter{TenantID: tid})

	u, _ := metadata.GetAuthenticatedUser(ctx)
	if req.OwnerGroupId != 0 {
		filters = append(filters, &clusterOwnerGroupFilter{OwnerGroupID: req.OwnerGroupId})
	} else {
		if !internal.IsAdminUser(u) {
			filters = append(filters, &clusterOwnerGroupListFilter{OwnerGroups: u.Groups})
		}
	}

	if len(req.TypeCode) != 0 {
		filters = append(filters, &clusterTypeCodeFilter{TypeCode: req.TypeCode})
	}

	if utf8.RuneCountInString(req.Name) != 0 {
		filters = append(filters, &clusterNameFilter{Name: req.Name})
	}

	if req.GetOffset() != nil || req.GetLimit() != nil {
		filters = append(filters, &paginationFilter{Offset: req.GetOffset(), Limit: req.GetLimit()})
	}

	return filters
}
