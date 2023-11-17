package keypair

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/metadata"
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

func getKeypair(db *gorm.DB, cid, kid uint64) (*model.ClusterKeypair, error) {
	var keypair model.ClusterKeypair
	err := db.Table(model.ClusterKeypair{}.TableName()).
		Where(&model.ClusterKeypair{ID: kid, ClusterID: cid}).
		First(&keypair).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterKeyPair(kid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &keypair, nil
}

// Get 클러스터 KeyPair 조회
func Get(ctx context.Context, req *cms.ClusterKeyPairRequest) (*cms.ClusterKeypair, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateGetRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[KeyPair-Get] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	var c *model.Cluster
	var k *model.ClusterKeypair

	if err = database.Execute(func(db *gorm.DB) error {
		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[KeyPair-Get] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		if k, err = getKeypair(db, req.ClusterId, req.ClusterKeypairId); err != nil {
			logger.Errorf("[KeyPair-Get] Could not get the cluster(%d) keypair. Cause: %+v", req.ClusterId, req.ClusterKeypairId, err)
			return err
		}

		return nil
	}); err != nil {

	}

	var rsp = cms.ClusterKeypair{
		Cluster: new(cms.Cluster),
	}

	if err := rsp.SetFromModel(k); err != nil {
		return nil, err
	}

	if err := rsp.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}

	return &rsp, nil
}

// CreateKeypair 클러스터 Keypair 생성
func CreateKeypair(req *cms.CreateClusterKeypairRequest) (*cms.ClusterKeypair, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateKeypair] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err := validateCreateKeypairRequest(req); err != nil {
		logger.Errorf("[CreateKeypair] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateKeypair] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[CreateKeypair] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateKeypair] Could not close client. Cause: %+v", err)
		}
	}()

	mKeypair, err := req.GetKeypair().Model()
	if err != nil {
		return nil, err
	}

	res, err := cli.CreateKeypair(client.CreateKeypairRequest{
		Keypair: *mKeypair,
	})
	if err != nil {
		logger.Errorf("[CreateKeypair] Could not create the cluster keypair. Cause: %+v", err)
		return nil, err
	}

	var keypair cms.ClusterKeypair
	if err = keypair.SetFromModel(&res.Keypair); err != nil {
		return nil, err
	}

	return &keypair, nil
}

// DeleteKeypair 클러스터 Keypair 삭제
func DeleteKeypair(req *cms.DeleteClusterKeypairRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteKeypair] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteKeypairRequest(req); err != nil {
		logger.Errorf("[DeleteKeypair] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteKeypair] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[DeleteKeypair] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteKeypair] Could not close client. Cause: %+v", err)
		}
	}()

	if err = cli.DeleteKeypair(client.DeleteKeypairRequest{
		Keypair: model.ClusterKeypair{
			Name: req.GetKeypair().GetName(),
		},
	}); err != nil {
		logger.Errorf("[DeleteKeypair] Could not delete the cluster keypair. Cause: %+v", err)
		return err
	}

	return nil
}

// CheckIsExistKeypair Keypair 이름을 통한 클러스터 Keypair 존재유무 확인
func CheckIsExistKeypair(ctx context.Context, req *cms.CheckIsExistByNameRequest) (bool, error) {
	var (
		err     error
		existed bool
	)

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateCheckIsExistKeypair(ctx, db, req); err != nil {
			return err
		}

		if existed, err = checkIsExistKeypairByName(db, req.ClusterId, req.GetName()); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return false, err
	}

	return existed, nil
}

func checkIsExistKeypairByName(db *gorm.DB, cid uint64, name string) (bool, error) {
	var cnt int
	err := db.Table(model.ClusterKeypair{}.TableName()).Where(&model.ClusterKeypair{ClusterID: cid, Name: name}).
		Count(&cnt).Error

	if err != nil {
		return false, errors.UnusableDatabase(err)
	}

	return cnt > 0, nil
}
