package tenant

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	"github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

func validateGetClusterTenantRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterTenantRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterTenantId() == 0 {
		return errors.RequiredParameter("cluster_tenant_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateGetClusterTenantByUUIDRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterTenantByUUIDRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetUuid() == "" {
		return errors.RequiredParameter("uuid")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateClusterTenantListRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterTenantListRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateCreateClusterTenantRequest(req *cms.CreateClusterTenantRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	if req.GetTenant() == nil {
		return errors.RequiredParameter("tenant")
	}

	tenantName := req.GetTenant().GetName()
	if tenantName == "" {
		return errors.RequiredParameter("tenant.name")
	}
	if utf8.RuneCountInString(tenantName) > 64 {
		return errors.LengthOverflowParameterValue("tenant.name", tenantName, 64)
	}

	// FIXME: 사용할 수 있는 quota key 를 모두 알지 못해 validation 항목에서 제외함
	//for i, qts := range req.GetTenant().GetQuotas() {
	//	key := qts.GetKey()
	//	if !client.IsQuotaKey(key) {
	//		return errors.UnavailableParameterValue(
	//			fmt.Sprintf("tenant.quota[%v].key", i), key,
	//			client.QuotaKeys)
	//	}
	//}
	return nil
}

func validateDeleteClusterTenantRequest(req *cms.DeleteClusterTenantRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	if req.GetTenant() == nil {
		return errors.RequiredParameter("tenant")
	}

	tenantUUID := req.GetTenant().GetUuid()
	if tenantUUID == "" {
		return errors.RequiredParameter("tenant.uuid")
	}
	if len(tenantUUID) > 36 {
		return errors.LengthOverflowParameterValue("tenant.uuid", tenantUUID, 36)
	}

	return nil
}

func validateCheckIsExistTenant(ctx context.Context, db *gorm.DB, req *cms.CheckIsExistByNameRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetName() == "" {
		return errors.RequiredParameter("name")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}
