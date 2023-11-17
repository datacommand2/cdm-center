package storage

import (
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/metadata"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"

	"context"
)

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

func getClusterStorage(db *gorm.DB, cid, sid uint64) (*model.ClusterStorage, error) {
	var m model.ClusterStorage
	err := db.First(&m, &model.ClusterStorage{ID: sid, ClusterID: cid}).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterStorage(sid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &m, nil
}

func getClusterStorageList(db *gorm.DB, clusterID uint64, filters ...storageFilter) ([]*model.ClusterStorage, error) {
	var err error
	var cond = db.Where(&model.ClusterStorage{ClusterID: clusterID})

	for _, f := range filters {
		cond, err = f.Apply(cond)
		if err != nil {
			return nil, err
		}
	}

	var list []*model.ClusterStorage
	if err = cond.Find(&list).Error; err != nil {
		return nil, err
	}

	return list, nil
}

// GetList 볼륨 타입 목록 조회
func GetList(ctx context.Context, req *cms.ClusterStorageListRequest) ([]*cms.ClusterStorage, *cms.Pagination, error) {
	var err error

	if err = database.Execute(func(db *gorm.DB) error {
		return validateStorageListRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[Storage-GetList] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	var rsp []*cms.ClusterStorage
	var pagination *cms.Pagination
	if err = database.Execute(func(db *gorm.DB) error {
		var c *model.Cluster
		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Storage-GetList] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		filters := makeClusterStorageListFilter(req)

		// TODO  Credential 과 Raw 는 필요할 것으로 예상되어 데이터 베이스에 추가하였으나 어떻게 사용할지 확실하지 않아 proto 구조체에서는 추가 안됨.
		var storages []*model.ClusterStorage
		if storages, err = getClusterStorageList(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[Storage-GetList] Could not get the cluster(%d) storage list. Cause: %+v", req.ClusterId, err)
			return errors.UnusableDatabase(err)
		}

		for _, tmp := range storages {
			var s cms.ClusterStorage
			if err = s.SetFromModel(tmp); err != nil {
				return err
			}

			s.Cluster = new(cms.Cluster)
			if err = s.Cluster.SetFromModel(c); err != nil {
				return err
			}

			rsp = append(rsp, &s)
		}

		if pagination, err = getStoragePagination(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[Storage-GetList] Could not get the cluster(%d) storage list pagination. Cause: %+v", req.ClusterId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return rsp, pagination, nil
}

func getStoragePagination(db *gorm.DB, clusterID uint64, filters ...storageFilter) (*cms.Pagination, error) {
	var err error
	var offset, limit uint64

	var conditions = db.Where(&model.ClusterStorage{ClusterID: clusterID})

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
	if err = conditions.Model(&model.ClusterStorage{}).Count(&total).Error; err != nil {
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

// Get 볼륨 타입 조회
func Get(ctx context.Context, req *cms.ClusterStorageRequest) (*cms.ClusterStorage, error) {
	var err error
	var c *model.Cluster
	var s *model.ClusterStorage

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateStorageRequest(ctx, db, req); err != nil {
			logger.Errorf("[Storage-Get] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Storage-Get] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		if s, err = getClusterStorage(db, req.ClusterId, req.ClusterStorageId); err != nil {
			logger.Errorf("[Storage-Get] Could not get the cluster(%d) storage(%d). Cause: %+v", req.ClusterId, req.ClusterStorageId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var rsp cms.ClusterStorage
	if err = rsp.SetFromModel(s); err != nil {
		return nil, err
	}

	rsp.Cluster = new(cms.Cluster)
	if err = rsp.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}

	return &rsp, nil
}
