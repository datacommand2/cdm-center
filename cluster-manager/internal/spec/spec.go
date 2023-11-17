package spec

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-center/cluster-manager/sync"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
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

func getInstanceSpec(db *gorm.DB, cid, sid uint64) (*model.ClusterInstanceSpec, error) {
	var spec model.ClusterInstanceSpec
	err := db.Where(&model.ClusterInstanceSpec{ClusterID: cid, ID: sid}).First(&spec).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterInstanceSpec(sid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &spec, nil
}

func getInstanceSpecByUUID(db *gorm.DB, cid uint64, uuid string) (*model.ClusterInstanceSpec, error) {
	var spec model.ClusterInstanceSpec
	err := db.Where(&model.ClusterInstanceSpec{ClusterID: cid, UUID: uuid}).First(&spec).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterInstanceSpecByUUID(uuid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &spec, nil
}

func getClusterInstanceSpecList(db *gorm.DB, cid uint64, filters ...instanceSpecFilter) ([]*model.ClusterInstanceSpec, error) {
	var err error

	for _, f := range filters {
		if db, err = f.Apply(db); err != nil {
			return nil, err
		}
	}

	var specs []*model.ClusterInstanceSpec
	err = db.Table(model.ClusterInstanceSpec{}.TableName()).
		Select("*").
		Where(&model.ClusterInstanceSpec{ClusterID: cid}).
		Find(&specs).Error
	if err != nil {
		return nil, err
	}

	return specs, nil
}

func getInstanceSpecsPagination(db *gorm.DB, clusterID uint64, filters ...instanceSpecFilter) (*cms.Pagination, error) {
	var err error
	var offset, limit uint64

	var conditions = db.Where(&model.ClusterInstanceSpec{ClusterID: clusterID})

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
	if err = conditions.Model(&model.ClusterInstanceSpec{}).Count(&total).Error; err != nil {
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

// GetList 클러스터 인스턴스 spec 목록 조회
func GetList(ctx context.Context, db *gorm.DB, req *cms.ClusterInstanceSpecListRequest) ([]*cms.ClusterInstanceSpec, *cms.Pagination, error) {
	if err := validateInstanceSpecListRequest(ctx, db, req); err != nil {
		logger.Errorf("[InstanceSpec-GetList] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	var err error
	var c *model.Cluster
	var rsp []*cms.ClusterInstanceSpec

	if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
		logger.Errorf("[InstanceSpec-GetList] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
		return nil, nil, err
	}

	var (
		filters = makeClusterInstanceSpecListFilters(req)
		specs   []*model.ClusterInstanceSpec
	)
	if specs, err = getClusterInstanceSpecList(db, req.ClusterId, filters...); err != nil {
		logger.Errorf("[InstanceSpec-GetList] Could not get the cluster(%d) instance spec list. Cause: %+v", req.ClusterId, err)
		return nil, nil, errors.UnusableDatabase(err)
	}

	var cmsCluster cms.Cluster
	if err = cmsCluster.SetFromModel(c); err != nil {
		return nil, nil, err
	}

	for _, s := range specs {
		var spec = cms.ClusterInstanceSpec{
			Cluster: &cmsCluster,
		}

		if err = spec.SetFromModel(s); err != nil {
			return nil, nil, err
		}

		var extraSpecs []model.ClusterInstanceExtraSpec
		if extraSpecs, err = s.GetExtraSpecs(db); err != nil {
			return nil, nil, errors.UnusableDatabase(err)
		}

		for _, e := range extraSpecs {
			extra := new(cms.ClusterInstanceExtraSpec)
			if err = extra.SetFromModel(&e); err != nil {
				return nil, nil, err
			}
			spec.ExtraSpecs = append(spec.ExtraSpecs, extra)
		}

		rsp = append(rsp, &spec)
	}

	var pagination *cms.Pagination
	if pagination, err = getInstanceSpecsPagination(db, req.ClusterId, filters...); err != nil {
		logger.Errorf("[InstanceSpec-GetList] Could not get the cluster(%d) instance specs pagination. Cause: %+v", req.ClusterId, err)
		return nil, nil, err
	}

	return rsp, pagination, nil
}

// Get 클러스터 인스턴스 Spec 조회
func Get(ctx context.Context, req *cms.ClusterInstanceSpecRequest) (*cms.ClusterInstanceSpec, error) {
	var err error
	var c *model.Cluster
	var s *model.ClusterInstanceSpec

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateInstanceSpecRequest(ctx, db, req); err != nil {
			logger.Errorf("[InstanceSpec-Get] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[InstanceSpec-Get] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		if s, err = getInstanceSpec(db, req.ClusterId, req.ClusterInstanceSpecId); err != nil {
			logger.Errorf("[InstanceSpec-Get] Could not get the cluster(%d) instance spec(%d). Cause: %+v", req.ClusterId, req.ClusterInstanceSpecId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var extraSpecs []model.ClusterInstanceExtraSpec
	if err = database.Execute(func(db *gorm.DB) error {
		if extraSpecs, err = s.GetExtraSpecs(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var spec = cms.ClusterInstanceSpec{
		Cluster: new(cms.Cluster),
	}

	if err = spec.SetFromModel(s); err != nil {
		return nil, err
	}

	if err = spec.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}

	for _, e := range extraSpecs {
		var extraSpec cms.ClusterInstanceExtraSpec
		if err = extraSpec.SetFromModel(&e); err != nil {
			return nil, err
		}

		spec.ExtraSpecs = append(spec.ExtraSpecs, &extraSpec)
	}

	return &spec, nil
}

// GetByUUID uuid 를 통해 클러스터 인스턴스 Spec 조회
func GetByUUID(ctx context.Context, req *cms.ClusterInstanceSpecByUUIDRequest) (*cms.ClusterInstanceSpec, error) {
	var (
		err error
		c   *model.Cluster
	)

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateInstanceSpecByUUIDRequest(ctx, db, req); err != nil {
			logger.Errorf("[InstanceSpec-GetByUUID] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[InstanceSpec-GetByUUID] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var (
		s          *model.ClusterInstanceSpec
		extraSpecs []model.ClusterInstanceExtraSpec
	)

	if err = database.Execute(func(db *gorm.DB) error {
		if s, err = getInstanceSpecByUUID(db, req.ClusterId, req.Uuid); err != nil {
			logger.Errorf("[InstanceSpec-GetByUUID] Could not get the cluster(%d) instance spec by uuid(%s). Cause: %+v", req.ClusterId, req.Uuid, err)
			return err
		}

		if extraSpecs, err = s.GetExtraSpecs(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var spec = cms.ClusterInstanceSpec{
		Cluster: new(cms.Cluster),
	}

	if err = spec.SetFromModel(s); err != nil {
		return nil, err
	}

	if err = spec.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}

	for _, e := range extraSpecs {
		var extraSpec cms.ClusterInstanceExtraSpec
		if err = extraSpec.SetFromModel(&e); err != nil {
			return nil, err
		}

		spec.ExtraSpecs = append(spec.ExtraSpecs, &extraSpec)
	}

	return &spec, nil
}

// CreateInstanceSpec 클러스터 인스턴스 Spec 생성
func CreateInstanceSpec(req *cms.CreateClusterInstanceSpecRequest) (*cms.ClusterInstanceSpec, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateInstanceSpec] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	//req.Spec.DiskTotalBytes = 1
	if err = validateCreateClusterInstanceSpecRequest(req); err != nil {
		logger.Errorf("[CreateInstanceSpec] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateInstanceSpec] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[CreateInstanceSpec] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateInstanceSpec] Could not close client. Cause: %+v", err)
		}
	}()

	mSpec, err := req.GetSpec().Model()
	if err != nil {
		return nil, err
	}

	var mExtraSpecs []model.ClusterInstanceExtraSpec
	for _, extra := range req.GetExtraSpecs() {
		mExtraSpecs = append(mExtraSpecs, model.ClusterInstanceExtraSpec{
			Key:   extra.Key,
			Value: extra.Value,
		})
	}

	res, err := cli.CreateInstanceSpec(client.CreateInstanceSpecRequest{
		InstanceSpec: *mSpec,
		ExtraSpec:    mExtraSpecs,
	})
	if err != nil {
		logger.Errorf("[CreateInstanceSpec] Could not create the cluster instance spec. Cause: %+v", err)
		return nil, err
	}

	var instanceSpec cms.ClusterInstanceSpec
	if err = instanceSpec.SetFromModel(&res.InstanceSpec); err != nil {
		return nil, err
	}

	return &instanceSpec, nil
}

// DeleteInstanceSpec 인스턴스 Spec 삭제
func DeleteInstanceSpec(req *cms.DeleteClusterInstanceSpecRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteInstanceSpec] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteClusterInstanceSpecRequest(req); err != nil {
		logger.Errorf("[DeleteInstanceSpec] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteInstanceSpec] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[DeleteInstanceSpec] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteInstanceSpec] Could not close client. Cause: %+v", err)
		}
	}()

	if err = cli.DeleteInstanceSpec(client.DeleteInstanceSpecRequest{
		InstanceSpec: model.ClusterInstanceSpec{
			UUID: req.GetSpec().GetUuid(),
		},
	}); err != nil {
		logger.Errorf("[DeleteInstanceSpec] Could not delete the cluster instance spec. Cause: %+v", err)
		return err
	}

	if err = func() error {
		synchronizer, err := sync.NewSynchronizer(req.Spec.Cluster.Id)
		if err != nil {
			return err
		}

		defer func() {
			if err = synchronizer.Close(); err != nil {
				logger.Warnf("[DeleteInstanceSpec] Could not close synchronizer client. Cause: %+v", err)
			}
		}()

		return synchronizer.SyncClusterInstanceSpec(&model.ClusterInstanceSpec{UUID: req.Spec.Uuid}, sync.Force(true))

	}(); err != nil {
		logger.Warnf("[DeleteInstanceSpec] Could not sync cluster(%d) instance spec(%s). Cause: %+v", req.Spec.Cluster.Id, req.Spec.Uuid, err)
	}

	return nil
}

// CheckIsExistInstanceSpec 인스턴스 Spec 이름을 통한 클러스터 인스턴스 Spec 존재유무 확인
func CheckIsExistInstanceSpec(ctx context.Context, req *cms.CheckIsExistByNameRequest) (bool, error) {
	var (
		err     error
		existed bool
	)

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateCheckIsExistInstanceSpec(ctx, db, req); err != nil {
			return err
		}

		if existed, err = checkIsExistInstanceSpecByName(db, req.ClusterId, req.GetName()); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return false, err
	}

	return existed, nil
}

func checkIsExistInstanceSpecByName(db *gorm.DB, cid uint64, name string) (bool, error) {
	var (
		err     error
		filters = makeClusterInstanceSpecListFilters(&cms.ClusterInstanceSpecListRequest{Name: name})
		specs   []*model.ClusterInstanceSpec
	)
	if specs, err = getClusterInstanceSpecList(db, cid, filters...); err != nil {
		logger.Errorf("[checkIsExistInstanceSpecByName] Could not get the cluster(%d) instance spec list. Cause: %+v", cid, err)
		return false, errors.UnusableDatabase(err)
	}

	synchronizer, err := sync.NewSynchronizer(cid)
	if err != nil {
		logger.Errorf("[checkIsExistInstanceSpecByName] Could not create synchronizer client. Cause: %+v", err)
		return false, err
	}

	defer func() {
		if err = synchronizer.Close(); err != nil {
			logger.Warnf("[checkIsExistInstanceSpecByName] Could not close synchronizer client. Cause: %+v", err)
		}
	}()

	_ = synchronizer.SyncClusterInstanceSpecs(specs)

	var cnt int
	err = db.Table(model.ClusterInstanceSpec{}.TableName()).Where(&model.ClusterInstanceSpec{ClusterID: cid, Name: name}).
		Count(&cnt).Error
	if err != nil {
		return false, errors.UnusableDatabase(err)
	}

	return cnt > 0, nil
}
