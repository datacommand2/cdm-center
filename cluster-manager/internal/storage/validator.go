package storage

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-center/cluster-manager/storage"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

func validateStorageRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterStorageRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterStorageId() == 0 {
		return errors.RequiredParameter("cluster_storage_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateStorageListRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterStorageListRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.TypeCode != "" && !storage.IsClusterStorageTypeCode(req.TypeCode) {
		return errors.UnavailableParameterValue("type_code", req.TypeCode, storage.ClusterStorageTypeCodes)
	}

	if utf8.RuneCountInString(req.Name) > 255 {
		return errors.LengthOverflowParameterValue("name", req.Name, 255)
	}
	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}
