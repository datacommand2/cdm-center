package spec

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

func validateInstanceSpecRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterInstanceSpecRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterInstanceSpecId() == 0 {
		return errors.RequiredParameter("cluster_instance_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateInstanceSpecByUUIDRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterInstanceSpecByUUIDRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetUuid() == "" {
		return errors.RequiredParameter("uuid")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateInstanceSpecListRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterInstanceSpecListRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if utf8.RuneCountInString(req.Name) > 255 {
		return errors.LengthOverflowParameterValue("name", req.Name, 255)
	}

	if len(req.Uuid) > 255 {
		return errors.LengthOverflowParameterValue("uuid", req.Uuid, 255)
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateCreateClusterInstanceSpecRequest(req *cms.CreateClusterInstanceSpecRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	// spec
	if req.GetSpec() == nil {
		return errors.RequiredParameter("spec")
	}

	specName := req.GetSpec().GetName()
	if specName == "" {
		return errors.RequiredParameter("spec.name")
	}

	if req.GetSpec().GetMemTotalBytes() == 0 {
		return errors.RequiredParameter("spec.mem_total_bytes")
	}

	// TODO: 콘트라베이스에서 TotalBytes 가 0으로 설정되어 있어 임시로 1 입력
	//if req.GetSpec().GetDiskTotalBytes() == 0 {
	//	return errors.RequiredParameter("spec.disk_total_bytes")
	//}

	if req.GetSpec().GetVcpuTotalCnt() == 0 {
		return errors.RequiredParameter("spec.vcpu_total_cnt")
	}

	if utf8.RuneCountInString(specName) > 255 {
		return errors.LengthOverflowParameterValue("spec.name", specName, 255)
	}

	specDescription := req.GetSpec().GetDescription()
	if utf8.RuneCountInString(specDescription) > 255 {
		return errors.LengthOverflowParameterValue("spec.description", specDescription, 255)
	}

	// extra specs
	for _, extra := range req.GetExtraSpecs() {
		if extra.Key == "" {
			return errors.RequiredParameter("extra_specs.key")
		}
		if utf8.RuneCountInString(extra.Key) > 255 {
			return errors.LengthOverflowParameterValue("extra_specs.key", extra.Key, 255)
		}
		if extra.Value == "" {
			return errors.RequiredParameter("extra_specs.value")
		}
		if utf8.RuneCountInString(extra.Value) > 255 {
			return errors.LengthOverflowParameterValue("extra_specs.value", extra.Value, 255)
		}
	}

	return nil
}

func validateDeleteClusterInstanceSpecRequest(req *cms.DeleteClusterInstanceSpecRequest) error {
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	// spec
	if req.GetSpec() == nil {
		return errors.RequiredParameter("spec")
	}

	if req.Spec.GetCluster() == nil {
		return errors.RequiredParameter("cluster")
	}

	if req.Spec.Cluster.GetId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	instanceSpecUUID := req.GetSpec().GetUuid()
	if instanceSpecUUID == "" {
		return errors.RequiredParameter("spec.uuid")
	}
	if len(instanceSpecUUID) > 255 {
		return errors.LengthOverflowParameterValue("spec.uuid", instanceSpecUUID, 255)
	}

	return nil
}

func validateCheckIsExistInstanceSpec(ctx context.Context, db *gorm.DB, req *cms.CheckIsExistByNameRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetName() == "" {
		return errors.RequiredParameter("name")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}
