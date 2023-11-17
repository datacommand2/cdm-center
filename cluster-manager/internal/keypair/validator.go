package keypair

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/jinzhu/gorm"
	"unicode/utf8"
)

func validateGetRequest(ctx context.Context, db *gorm.DB, req *cms.ClusterKeyPairRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterKeypairId() == 0 {
		return errors.RequiredParameter("cluster_keypair_id")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}

func validateCreateKeypairRequest(req *cms.CreateClusterKeypairRequest) error {
	// 클러스터 연결 상태 검사
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	// keypair 파라미터 검사
	if req.GetKeypair() == nil {
		return errors.RequiredParameter("keypair")
	}

	keypairName := req.GetKeypair().GetName()
	if keypairName == "" {
		return errors.RequiredParameter("keypair.name")
	}
	if utf8.RuneCountInString(keypairName) > 255 {
		return errors.LengthOverflowParameterValue("keypair.name", keypairName, 255)
	}

	keypairPublicKey := req.GetKeypair().GetPublicKey()
	if keypairPublicKey == "" {
		return errors.RequiredParameter("keypair.public_key")
	}
	if len(keypairPublicKey) > 2048 {
		return errors.LengthOverflowParameterValue("keypair.public_key", keypairPublicKey, 2048)
	}

	keypairTypeCode := req.GetKeypair().GetTypeCode()
	if keypairTypeCode == "" {
		return errors.RequiredParameter("keypair.type_code")
	}
	if !internal.IsOpenstackKeypairTypeCode(keypairTypeCode) {
		return errors.UnavailableParameterValue("keypair.type_code", keypairTypeCode, internal.OpenstackKeypairTypeCodes)
	}

	return nil
}

func validateDeleteKeypairRequest(req *cms.DeleteClusterKeypairRequest) error {
	// 클러스터 연결 상태 검사
	if err := cluster.CheckValidClusterConnectionInfo(req.GetConn()); err != nil {
		return err
	}

	// keypair 파라미터 검사
	if req.GetKeypair() == nil {
		return errors.RequiredParameter("keypair")
	}

	KeypairName := req.GetKeypair().GetName()
	if KeypairName == "" {
		return errors.RequiredParameter("keypair.name")
	}
	if utf8.RuneCountInString(KeypairName) > 255 {
		return errors.LengthOverflowParameterValue("keypair.name", KeypairName, 255)
	}

	return nil
}

func validateCheckIsExistKeypair(ctx context.Context, db *gorm.DB, req *cms.CheckIsExistByNameRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetName() == "" {
		return errors.RequiredParameter("name")
	}

	return cluster.IsClusterOwner(ctx, db, req.ClusterId)
}
