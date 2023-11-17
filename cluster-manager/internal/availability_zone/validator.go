package availabilityzone

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/jinzhu/gorm"
)

func validateGetAvailabilityZoneRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterAvailabilityZoneRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterAvailabilityZoneId() == 0 {
		return errors.RequiredParameter("cluster_availability_zone_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateGetAvailabilityZoneListRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterAvailabilityZoneListRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}
