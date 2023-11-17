package tenant

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
)

// DeleteTenant 테넌트 삭제
func DeleteTenant(req *cms.DeleteClusterTenantRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteTenant] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteClusterTenantRequest(req); err != nil {
		logger.Errorf("[DeleteTenant] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteTenant] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[DeleteTenant] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteTenant] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return err
	}

	return cli.DeleteTenant(client.DeleteTenantRequest{
		Tenant: *mTenant,
	})
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

func getClusterTenant(db *gorm.DB, clusterID, tenantID uint64) (*model.ClusterTenant, error) {
	var tenant model.ClusterTenant
	err := db.First(&tenant, &model.ClusterTenant{ClusterID: clusterID, ID: tenantID}).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterTenant(tenantID)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &tenant, nil
}

func getClusterTenantByUUID(db *gorm.DB, clusterID uint64, uuid string) (*model.ClusterTenant, error) {
	var tenant model.ClusterTenant
	err := db.First(&tenant, &model.ClusterTenant{ClusterID: clusterID, UUID: uuid}).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterTenantByUUID(uuid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &tenant, nil
}

// Get 클러스터 테넌트 조회
func Get(ctx context.Context, req *cms.ClusterTenantRequest) (*cms.ClusterTenant, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateGetClusterTenantRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[Tenant-Get] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	var c *model.Cluster
	var rsp cms.ClusterTenant
	var tenant *model.ClusterTenant
	var quotas []model.ClusterQuota

	if err = database.Execute(func(db *gorm.DB) error {
		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Tenant-Get] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		if tenant, err = getClusterTenant(db, req.ClusterId, req.ClusterTenantId); err != nil {
			logger.Errorf("[Tenant-Get] Could not get the cluster(%d) tenant(%d). Cause: %+v", req.ClusterId, req.ClusterTenantId, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		if quotas, err = tenant.GetQuotas(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if err = rsp.SetFromModel(tenant); err != nil {
		return nil, err
	}

	rsp.Cluster = new(cms.Cluster)
	if err = rsp.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}

	for _, q := range quotas {
		quota := new(cms.ClusterQuota)
		if err = quota.SetFromModel(&q); err != nil {
			return nil, err
		}
		rsp.Quotas = append(rsp.Quotas, quota)
	}

	return &rsp, nil
}

// GetByUUID uuid 를 통해 클러스터 테넌트 조회(UUID)
func GetByUUID(ctx context.Context, req *cms.ClusterTenantByUUIDRequest) (*cms.ClusterTenant, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateGetClusterTenantByUUIDRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[Tenant-GetByUUID] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	var c *model.Cluster
	var rsp cms.ClusterTenant
	var tenant *model.ClusterTenant
	var quotas []model.ClusterQuota

	if err = database.Execute(func(db *gorm.DB) error {
		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Tenant-GetByUUID] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		if tenant, err = getClusterTenantByUUID(db, req.ClusterId, req.Uuid); err != nil {
			logger.Errorf("[Tenant-GetByUUID] Could not get the cluster(%d) tenant by uuid(%s). Cause: %+v", req.ClusterId, req.Uuid, err)
			return err
		}

		if quotas, err = tenant.GetQuotas(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if err = rsp.SetFromModel(tenant); err != nil {
		return nil, err
	}

	rsp.Cluster = new(cms.Cluster)
	if err = rsp.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}

	for _, q := range quotas {
		quota := new(cms.ClusterQuota)
		if err = quota.SetFromModel(&q); err != nil {
			return nil, err
		}
		rsp.Quotas = append(rsp.Quotas, quota)
	}

	return &rsp, nil
}

func getClusterTenantList(db *gorm.DB, clusterID uint64, filters ...tenantFilter) ([]*model.ClusterTenant, error) {
	var err error
	condition := db.Where(&model.ClusterTenant{ClusterID: clusterID})

	for _, f := range filters {
		if condition, err = f.Apply(condition); err != nil {
			return nil, err
		}
	}

	var tenants []*model.ClusterTenant
	if err = condition.Select("id, uuid, name, enabled").Find(&tenants).Error; err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return tenants, nil
}

// GetList 클러스터 테넌트 목록 조회
func GetList(ctx context.Context, req *cms.ClusterTenantListRequest) ([]*cms.ClusterTenant, *cms.Pagination, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateClusterTenantListRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[Tenant-GetList] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	var c *model.Cluster
	var rsp []*cms.ClusterTenant
	var tenants []*model.ClusterTenant

	if err = database.Execute(func(db *gorm.DB) error {
		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Tenant-GetList] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}

	filters := makeClusterTenantListFilter(req)

	var pagination *cms.Pagination
	if err = database.Execute(func(db *gorm.DB) error {
		if tenants, err = getClusterTenantList(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[Tenant-GetList] Could not get the cluster(%d) tenant list. Cause: %+v", req.ClusterId, err)
			return err
		}

		for _, t := range tenants {
			var tenant cms.ClusterTenant
			var quotas []model.ClusterQuota

			if err = tenant.SetFromModel(t); err != nil {
				return err
			}

			if quotas, err = t.GetQuotas(db); err != nil {
				return errors.UnusableDatabase(err)
			}

			tenant.Cluster = new(cms.Cluster)
			if err = tenant.Cluster.SetFromModel(c); err != nil {
				return err
			}

			for _, q := range quotas {
				quota := new(cms.ClusterQuota)
				if err = quota.SetFromModel(&q); err != nil {
					return err
				}

				tenant.Quotas = append(tenant.Quotas, quota)
			}

			rsp = append(rsp, &tenant)
		}

		if pagination, err = getTenantPagination(db, c.ID, filters...); err != nil {
			logger.Errorf("[Tenant-GetList] Could not get the cluster(%d) tenant list pagination. Cause: %+v", req.ClusterId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return rsp, pagination, nil
}

func getTenantPagination(db *gorm.DB, clusterID uint64, filters ...tenantFilter) (*cms.Pagination, error) {
	var err error
	var offset, limit uint64

	var conditions = db.Where(&model.ClusterTenant{ClusterID: clusterID})

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
	if err = conditions.Model(&model.ClusterTenant{}).Count(&total).Error; err != nil {
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

// CreateTenant 클러스터 테넌트 생성
func CreateTenant(req *cms.CreateClusterTenantRequest) (*cms.ClusterTenant, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateTenant] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err = validateCreateClusterTenantRequest(req); err != nil {
		logger.Errorf("[CreateTenant] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateTenant] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[CreateTenant] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateTenant] Could not close client. Cause: %+v", err)
		}
	}()

	var mQuotas []model.ClusterQuota
	for _, item := range req.GetTenant().GetQuotas() {
		mQuota, err := item.Model()
		if err != nil {
			return nil, err
		}
		mQuotas = append(mQuotas, *mQuota)
	}

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return nil, err
	}
	tenantResponse, err := cli.CreateTenant(client.CreateTenantRequest{
		Tenant:    *mTenant,
		QuotaList: mQuotas,
	})
	if err != nil {
		return nil, err
	}
	var rollback = true
	defer func() {
		if rollback {
			if err := cli.DeleteTenant(client.DeleteTenantRequest{
				Tenant: model.ClusterTenant{
					UUID: tenantResponse.Tenant.UUID,
				},
			}); err != nil {
				logger.Warnf("[CreateTenant] Could not delete cluster tenant. Cause: %+v", err)
			}
		}
	}()

	var cTenant cms.ClusterTenant
	if err := cTenant.SetFromModel(&tenantResponse.Tenant); err != nil {
		return nil, err
	}

	var cQuotas []*cms.ClusterQuota
	for _, item := range tenantResponse.QuotaList {
		var cQuota cms.ClusterQuota
		if err := cQuota.SetFromModel(&item); err != nil {
			return nil, err
		}
		cQuotas = append(cQuotas, &cQuota)
	}
	cTenant.Quotas = cQuotas

	rollback = false
	return &cTenant, nil
}

// CheckIsExistTenant 테넌트 이름을 통한 클러스터 테넌트 존재유무 확인
func CheckIsExistTenant(ctx context.Context, req *cms.CheckIsExistByNameRequest) (bool, error) {
	var (
		err     error
		existed bool
	)

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateCheckIsExistTenant(ctx, db, req); err != nil {
			return err
		}

		if existed, err = checkIsExistTenantByName(db, req.ClusterId, req.GetName()); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return false, err
	}

	return existed, nil
}

func checkIsExistTenantByName(db *gorm.DB, cid uint64, name string) (bool, error) {
	var cnt int
	err := db.Table(model.ClusterTenant{}.TableName()).Where(&model.ClusterTenant{ClusterID: cid, Name: name}).
		Count(&cnt).Error

	if err != nil {
		return false, errors.UnusableDatabase(err)
	}

	return cnt > 0, nil
}
