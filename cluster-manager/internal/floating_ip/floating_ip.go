package floatingip

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	"github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
)

type floatingIPRecord struct {
	FloatingIP model.ClusterFloatingIP `gorm:"embedded"`
	Network    model.ClusterNetwork    `gorm:"embedded"`
	Tenant     model.ClusterTenant     `gorm:"embedded"`
}

func checkIsExistFloatingIPByIPAddress(db *gorm.DB, cid uint64, ipAddress string) (bool, error) {
	var cnt int
	err := db.Table(model.ClusterFloatingIP{}.TableName()).
		Joins("join cdm_cluster_network on cdm_cluster_network.id = cdm_cluster_floating_ip.cluster_network_id").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_floating_ip.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: cid}).
		Where(&model.ClusterFloatingIP{IPAddress: ipAddress}).
		Count(&cnt).Error

	if err != nil {
		return false, errors.UnusableDatabase(err)
	}

	return cnt > 0, nil
}

func getClusterFloatingIP(db *gorm.DB, cid, fid uint64) (*floatingIPRecord, error) {
	var r floatingIPRecord
	err := db.Table(model.ClusterFloatingIP{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_network on cdm_cluster_network.id = cdm_cluster_floating_ip.cluster_network_id").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_floating_ip.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: cid}).
		Where(&model.ClusterFloatingIP{ID: fid}).First(&r).Error

	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterFloatingIP(fid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &r, nil
}

// GetFloatingIP 클러스터 FloatingIP 조회
func GetFloatingIP(ctx context.Context, req *cms.ClusterFloatingIPRequest) (*cms.ClusterFloatingIP, error) {
	var err error
	var f *floatingIPRecord

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateGetClusterFloatingIP(ctx, db, req); err != nil {
			logger.Errorf("[GetFloatingIP] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if f, err = getClusterFloatingIP(db, req.ClusterId, req.ClusterFloatingIpId); err != nil {
			logger.Errorf("[GetFloatingIP] Could not get the cluster(%d) floating IP(%d). Cause: %+v", req.ClusterId, req.ClusterFloatingIpId, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var floatingIP = cms.ClusterFloatingIP{
		Tenant:  new(cms.ClusterTenant),
		Network: new(cms.ClusterNetwork),
	}

	if err = floatingIP.SetFromModel(&f.FloatingIP); err != nil {
		return nil, err
	}
	if err = floatingIP.Tenant.SetFromModel(&f.Tenant); err != nil {
		return nil, err
	}
	if err = floatingIP.Network.SetFromModel(&f.Network); err != nil {
		return nil, err
	}

	return &floatingIP, nil
}

// CheckIsExistFloatingIP ip address 를 통한 클러스터 FloatingIP 존재유무 확인
func CheckIsExistFloatingIP(ctx context.Context, req *cms.CheckIsExistClusterFloatingIPRequest) (bool, error) {
	var (
		err     error
		existed bool
	)

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateCheckIsExistFloatingIP(ctx, db, req); err != nil {
			logger.Errorf("[CheckIsExistFloatingIP] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if existed, err = checkIsExistFloatingIPByIPAddress(db, req.ClusterId, req.GetClusterFloatingIpIpAddress()); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return false, err
	}

	return existed, nil
}

// CreateFloatingIP 부동 IP 생성
func CreateFloatingIP(req *cms.CreateClusterFloatingIPRequest) (*cms.ClusterFloatingIP, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateFloatingIP] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err = validateCreateClusterFloatingIPRequest(req); err != nil {
		logger.Errorf("[CreateFloatingIP] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateFloatingIP] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[CreateFloatingIP] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateFloatingIP] Could not close client. Cause: %+v", err)
		}
	}()

	mFloatingIP, err := req.GetFloatingIp().Model()
	if err != nil {
		return nil, err
	}
	tenant, err := req.GetTenant().Model()
	if err != nil {
		return nil, err
	}
	network, err := req.GetNetwork().Model()
	if err != nil {
		return nil, err
	}
	floatingIPResponse, err := cli.CreateFloatingIP(client.CreateFloatingIPRequest{
		Tenant:     *tenant,
		Network:    *network,
		FloatingIP: *mFloatingIP,
	})
	if err != nil {
		logger.Errorf("[CreateFloatingIP] Could not create the cluster floating IP. Cause: %+v", err)
		return nil, err
	}

	var rollback = true
	defer func() {
		if rollback {
			if err := cli.DeleteFloatingIP(client.DeleteFloatingIPRequest{
				Tenant:     *tenant,
				FloatingIP: floatingIPResponse.FloatingIP,
			}); err != nil {
				logger.Warnf("[CreateFloatingIP] Could not delete the cluster floating ip during rollback. Cause: %+v", err)
			}
		}
	}()

	var cFloatingIP cms.ClusterFloatingIP
	if err := cFloatingIP.SetFromModel(&floatingIPResponse.FloatingIP); err != nil {
		return nil, err
	}

	rollback = false
	return &cFloatingIP, nil
}

// DeleteFloatingIP 부동 IP 삭제
func DeleteFloatingIP(req *cms.DeleteClusterFloatingIPRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteFloatingIP] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteClusterFloatingIPRequest(req); err != nil {
		logger.Errorf("[DeleteFloatingIP] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteFloatingIP] Could not create client. Cause: %+v", err)
		return err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[DeleteFloatingIP] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteFloatingIP] Could not close client. Cause: %+v", err)
		}
	}()

	tenant, err := req.GetTenant().Model()
	if err != nil {
		return err
	}
	ip, err := req.GetFloatingIp().Model()
	if err != nil {
		return err
	}
	return cli.DeleteFloatingIP(client.DeleteFloatingIPRequest{
		Tenant:     *tenant,
		FloatingIP: *ip,
	})
}
