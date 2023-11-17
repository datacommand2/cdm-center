package availabilityzone

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
)

func getClusterHypervisors(db *gorm.DB, zoneID uint64) ([]model.ClusterHypervisor, error) {
	var nodes []model.ClusterHypervisor
	if err := db.
		Select("id, type_code, hostname, ip_address").
		Where("cluster_availability_zone_id = ?", zoneID).
		Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func getCluster(ctx context.Context, db *gorm.DB, id uint64) (*model.Cluster, error) {
	tid, _ := metadata.GetTenantID(ctx)

	var m model.Cluster
	err := db.Model(&m).
		Select("id, name, type_code, remarks").
		Where(&model.Cluster{ID: id, TenantID: tid}).
		First(&m).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundCluster(id)
	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}
	return &m, nil
}

func getAvailabilityZone(db *gorm.DB, cid, azid uint64) (*model.ClusterAvailabilityZone, error) {
	var m model.ClusterAvailabilityZone
	err := db.Model(&m).
		Where(&model.ClusterAvailabilityZone{ID: azid, ClusterID: cid}).
		First(&m).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterAvailabilityZone(azid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &m, nil
}

// Get 클러스터 가용 구역 조회
func Get(ctx context.Context, req *cms.ClusterAvailabilityZoneRequest) (*cms.ClusterAvailabilityZone, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateGetAvailabilityZoneRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[AvailabilityZone-Get] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	var c *model.Cluster
	var az *model.ClusterAvailabilityZone
	var hypervisors []model.ClusterHypervisor

	if err = database.Execute(func(db *gorm.DB) error {
		if az, err = getAvailabilityZone(db, req.ClusterId, req.ClusterAvailabilityZoneId); err != nil {
			logger.Errorf("[AvailabilityZone-Get] Could not get the availability zone(%d). Cause: %+v", req.ClusterAvailabilityZoneId, err)
			return err
		}

		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[AvailabilityZone-Get] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		// 하이퍼바이저의 일부 필드만 조회하므로 modeL_relation 의 함수를 호출하지 않고 별도 함수 호출
		if hypervisors, err = getClusterHypervisors(db, req.ClusterAvailabilityZoneId); err != nil {
			logger.Errorf("[AvailabilityZone-Get] Could not get the cluster hypervisors. Cause: %+v", err)
			return errors.UnusableDatabase(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var rsp cms.ClusterAvailabilityZone
	if err = rsp.SetFromModel(az); err != nil {
		return nil, err
	}

	rsp.Cluster = new(cms.Cluster)
	if err = rsp.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}

	for _, h := range hypervisors {
		var hypervisor cms.ClusterHypervisor
		if err = hypervisor.SetFromModel(&h); err != nil {
			return nil, err
		}
		rsp.Hypervisors = append(rsp.Hypervisors, &hypervisor)
	}

	return &rsp, nil
}

func getAvailabilityZoneList(db *gorm.DB, clusterID uint64, filters ...availabilityZoneFilter) ([]model.ClusterAvailabilityZone, error) {
	var err error
	condition := db.Where(&model.ClusterAvailabilityZone{ClusterID: clusterID})

	for _, f := range filters {
		if condition, err = f.Apply(condition); err != nil {
			return nil, err
		}
	}
	var availabilityZones []model.ClusterAvailabilityZone

	if err = condition.Find(&availabilityZones).Error; err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return availabilityZones, nil
}

// GetList 가용구역 목록을 조회한다.
func GetList(ctx context.Context, req *cms.ClusterAvailabilityZoneListRequest) ([]*cms.ClusterAvailabilityZone, *cms.Pagination, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateGetAvailabilityZoneListRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[AvailabilityZone-GetList] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	var c *model.Cluster
	if err = database.Execute(func(db *gorm.DB) error {
		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[AvailabilityZone-GetList] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}

	var rspCluster cms.Cluster
	if err = rspCluster.SetFromModel(c); err != nil {
		return nil, nil, err
	}

	filters := makeAvailabilityZoneListFilter(req)

	var availabilityZones []model.ClusterAvailabilityZone
	if err = database.Execute(func(db *gorm.DB) error {
		if availabilityZones, err = getAvailabilityZoneList(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[AvailabilityZone-GetList] Could not get the availability zone list. Cause: %+v", err)
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}

	var rsp []*cms.ClusterAvailabilityZone
	for _, az := range availabilityZones {
		var availabilityZone cms.ClusterAvailabilityZone
		if err = availabilityZone.SetFromModel(&az); err != nil {
			return nil, nil, err
		}
		availabilityZone.Cluster = &rspCluster
		rsp = append(rsp, &availabilityZone)
	}

	var pagination *cms.Pagination
	if err = database.Execute(func(db *gorm.DB) error {
		if pagination, err = getAvailabilityZonesPagination(db, c.ID, filters...); err != nil {
			logger.Errorf("[AvailabilityZone-GetList] Could not get the availability zones pagination. Cause: %+v", err)
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}

	return rsp, pagination, err
}

func getAvailabilityZonesPagination(db *gorm.DB, clusterID uint64, filters ...availabilityZoneFilter) (*cms.Pagination, error) {
	var err error
	var offset, limit uint64

	var conditions = db.Where(&model.ClusterAvailabilityZone{ClusterID: clusterID})

	for _, f := range filters {
		if _, ok := f.(*paginationFilter); ok {
			offset = f.(*paginationFilter).Offset.GetValue()
			limit = f.(*paginationFilter).Limit.GetValue()
			continue
		}

		conditions, err = f.Apply(conditions)
		if err != nil {
			return nil, errors.UnusableDatabase(err)
		}
	}

	var total uint64
	if err = conditions.Model(&model.ClusterAvailabilityZone{}).Count(&total).Error; err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	if limit == 0 {
		return &cms.Pagination{
			Page:       &wrappers.UInt64Value{Value: 1},
			TotalPage:  &wrappers.UInt64Value{Value: 1},
			TotalItems: &wrappers.UInt64Value{Value: total},
		}, nil
	}

	return &cms.Pagination{
		Page:       &wrappers.UInt64Value{Value: offset/limit + 1},
		TotalPage:  &wrappers.UInt64Value{Value: (total + limit - 1) / limit},
		TotalItems: &wrappers.UInt64Value{Value: total},
	}, nil
}
