package handler

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/config"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	availabilityZone "github.com/datacommand2/cdm-center/cluster-manager/internal/availability_zone"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	floatingip "github.com/datacommand2/cdm-center/cluster-manager/internal/floating_ip"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/hypervisor"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/instance"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/keypair"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/network"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/router"
	securityGroup "github.com/datacommand2/cdm-center/cluster-manager/internal/security_group"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/spec"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/storage"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/tenant"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/volume"
	volumegroup "github.com/datacommand2/cdm-center/cluster-manager/internal/volume_group"
	"github.com/datacommand2/cdm-center/cluster-manager/leader"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	clusterStorage "github.com/datacommand2/cdm-center/cluster-manager/storage"
	"github.com/datacommand2/cdm-center/cluster-manager/sync"
	"github.com/datacommand2/cdm-cloud/common/broker"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"

	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/jinzhu/gorm"
	"io/ioutil"
	"net"
	"net/url"
	"path"
	"strconv"
	"time"
)

const (
	defaultPrivateKeyPath = "/etc/ssl/cdm-center"
	defaultPrivateKeyFile = "cluster-manager.pem"
)

// ClusterManagerHandler Cluster Manager 서비스 RPC 핸들러
type ClusterManagerHandler struct {
	privateKey  *rsa.PrivateKey
	syncStatus  map[uint64]queue.SyncClusterStatus
	checkStatus map[uint64]queue.ClusterServiceStatus
	updatedAt   map[uint64]time.Time
}

// GetPublicKey 공개키 조회
func (h *ClusterManagerHandler) GetPublicKey(ctx context.Context, _ *cms.Empty, rsp *cms.PublicKey) error {
	logger.Debug("Received ClusterManager.GetPublicKey request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetPublicKey] Could not get public key. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_public_key.failure-validate_request", err)
	}

	pub, err := x509.MarshalPKIXPublicKey(h.privateKey.Public())
	if err != nil {
		err = errors.Unknown(err)
		return createError(ctx, "cdm-center.manager.get_public_key.failure-marshal", err)
	}

	p := pem.Block{Type: "RSA PUBLIC KEY", Bytes: pub}
	b := pem.EncodeToMemory(&p)
	rsp.Key = string(b)

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_public_key.success"}
	logger.Debug("Respond ClusterManager.GetPublicKey request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_public_key.success", nil)
}

// CheckConnection 클러스터 연결 확인
func (h *ClusterManagerHandler) CheckConnection(ctx context.Context, req *cms.ClusterConnectionInfo, rsp *cms.ClusterConnectionResponse) error {
	logger.Debug("Received ClusterManager.CheckConnection request")
	var err error

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-CheckConnection] Could not check connection. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.check_connection.failure-validate_request", err)
	}

	req.Credential, err = client.DecryptCredentialPassword(req.GetCredential())
	if err != nil {
		return createError(ctx, "cdm-center.manager.check_connection.failure-credential_password", err)
	}

	if err = cluster.CheckConnection(req); err != nil {
		return createError(ctx, "cdm-center.manager.check_connection.failure-check", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.check_connection.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.check_connection.success", nil)
}

// GetCenterConfig center 의 config 조회
func (h *ClusterManagerHandler) GetCenterConfig(ctx context.Context, req *cms.GetCenterConfigRequest, rsp *cms.GetCenterConfigResponse) error {
	logger.Debug("Received ClusterManager.GetCenterConfig request")
	var err error

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetCenterConfig] Could not get config. Cause: %+v", err)
		}
	}()

	cfg, _ := config.GetConfig(req.ClusterId)
	if cfg != nil {
		rsp.Config = &cms.Config{
			TimestampInterval:    cfg.TimestampInterval,
			ReservedSyncInterval: cfg.ReservedSyncInterval,
		}

	} else {
		// Etcd 에 저장된 config 값이 없어서 기본값으로 출력하고 저장.
		rsp.Config = &cms.Config{
			TimestampInterval:    5,
			ReservedSyncInterval: 1,
		}

		cfg = &config.Config{
			ClusterID:            req.ClusterId,
			ReservedSyncInterval: 1,
			TimestampInterval:    5,
		}
		if err = cfg.Put(); err != nil {
			return createError(ctx, "cdm-center.manager.get_center_config.failure-put", err)
		}
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_center_config.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.get_center_config.success", nil)
}

// UpdateCenterConfig center 의 config 업데이트
func (h *ClusterManagerHandler) UpdateCenterConfig(ctx context.Context, req *cms.UpdateCenterConfigRequest, rsp *cms.UpdateCenterConfigResponse) error {
	logger.Debug("Received ClusterManager.UpdateCenterConfig request")
	var err error

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetCenterConfig] Could not update config. Cause: %+v", err)
		}
	}()

	cfg, _ := config.GetConfig(req.ClusterId)
	// config 업데이트 값 중 TimestampInterval 이 있으면 1 부터 10까지의 값만 수정할 수 있다.
	if req.Config.TimestampInterval != 0 {
		if req.Config.TimestampInterval >= 1 || req.Config.TimestampInterval <= 10 {
			cfg.TimestampInterval = req.Config.TimestampInterval
		} else {
			logger.Warnf("[Handler-GetCenterConfig] Could not set config of value. Cause: %+v",
				errors.OutOfRangeParameterValue("", req.Config.TimestampInterval, 1, 10))
		}
	}

	// config 업데이트 값 중 ReservedSyncInterval 이 있으면 1 부터 3까지의 값만 수정할 수 있다.
	if req.Config.ReservedSyncInterval != 0 {
		if req.Config.ReservedSyncInterval >= 1 || req.Config.ReservedSyncInterval <= 3 {
			cfg.ReservedSyncInterval = req.Config.ReservedSyncInterval
		} else {
			logger.Warnf("[Handler-GetCenterConfig] Could not set config of value. Cause: %+v",
				errors.OutOfRangeParameterValue("", req.Config.ReservedSyncInterval, 1, 3))
		}
	}

	if err = cfg.Put(); err != nil {
		return createError(ctx, "cdm-center.manager.update_center_config.failure-put", err)
	}

	b, _ := json.Marshal(queue.ClusterConfig{
		ClusterID: req.ClusterId,
		Config:    cfg,
	})

	// broker 로 메시지를 보내 메모리 데이터에 저장하기 위함
	if err = broker.Publish(constant.TopicNoticeCenterConfigUpdated, &broker.Message{Body: b}); err != nil {
		logger.Warnf("[Handler-UpdateCenterConfig] Unable to publish updated Config. Cause: %+v", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.update_center_config.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.update_center_config.success", nil)
}

// CheckCluster Cluster 의 nova, cinder, neutron 상태조회
func (h *ClusterManagerHandler) CheckCluster(ctx context.Context, req *cms.CheckClusterRequest, rsp *cms.CheckClusterResponse) error {
	logger.Debug("Received ClusterManager.CheckCluster request")
	var err error

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-CheckCluster] Could not check cluster. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.check_cluster.failure-validate_request", err)
	}

	var c *model.Cluster
	if err = database.Transaction(func(db *gorm.DB) error {
		c, err = cluster.ValidateSyncCluster(ctx, db, req.ClusterId)
		return err
	}); err != nil {
		return createError(ctx, "cdm-center.manager.check_cluster.failure-validate_sync_cluster", err)
	}

	logger.Debugf("[Handler-CheckCluster] Gets the latest service status for the cluster(%d:%s).", req.ClusterId, c.Name)
	// 메모리 데이터이기 때문에 함수가 시작하는 시점의 데이터를 저장하기 위해 변수에 할당
	data, ok := h.checkStatus[req.ClusterId]
	if !ok {
		logger.Warnf("[Handler-CheckCluster] Cluster(%d) status data does not exist in memory.", req.ClusterId)
	}

	updatedAt, ok := h.updatedAt[req.ClusterId]
	if !ok {
		logger.Warnf("[Handler-CheckCluster] Cluster(%d) UpdatedAt data does not exist in memory.", req.ClusterId)
	}

	var status string
	rsp.Computes, rsp.Storages, rsp.Networks, status, err = cluster.CheckCluster(c, data)
	if err != nil {
		return createError(ctx, "cdm-center.manager.check_cluster.failure-check_cluster", err)
	}

	if data.Network == nil && data.Compute == nil && data.Storage == nil && status != constant.ClusterStateInactive {
		rsp.Status = constant.ClusterStateLoading
	} else {
		rsp.Status = status
	}

	if data.ComputeError != "" {
		rsp.ComputeError = data.ComputeError
	}
	if data.StorageError != "" {
		rsp.StorageError = data.StorageError
	}
	if data.NetworkError != "" {
		rsp.NetworkError = data.NetworkError
	}
	if updatedAt.String() != "" {
		rsp.UpdatedAt = cluster.DurationToString(time.Now().Sub(updatedAt))
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.check_cluster.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.check_cluster.success", nil)
}

// UpdateCredential 클러스터 인증 정보 수정
func (h *ClusterManagerHandler) UpdateCredential(ctx context.Context, req *cms.UpdateCredentialRequest, rsp *cms.UpdateCredentialResponse) error {
	logger.Debug("Received ClusterManager.UpdateCredential request")
	var err error

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-UpdateCredential] Cloud not update credential. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.update_credential.failure-validate_request", err)
	}

	if err = database.Transaction(func(db *gorm.DB) error {
		err = cluster.UpdateCredential(ctx, db, req)
		return err
	}); err != nil {
		return createError(ctx, "cdm-center.manager.update_credential.failure-update_credential", err)
	}

	return errors.StatusOK(ctx, "cdm-center.manager.update_credential.success", nil)
}

// SyncCluster 클러스터 동기화
func (h *ClusterManagerHandler) SyncCluster(ctx context.Context, req *cms.SyncClusterRequest, rsp *cms.SyncClusterResponse) (err error) {
	logger.Debug("Received ClusterManager.SyncCluster request")

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-SyncCluster] Could not sync cluster. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.sync_cluster.failure-validate_request", err)
	}

	logger.Infof("[Handler-SyncCluster] Start to cluster(%d) manual sync.", req.GetClusterId())
	var c *model.Cluster
	if err = database.Transaction(func(db *gorm.DB) error {
		c, err = cluster.ValidateSyncCluster(ctx, db, req.ClusterId)
		return err
	}); err != nil {
		return createError(ctx, "cdm-center.manager.sync_cluster.failure-sync", err)
	}

	ok, status := cluster.IsAbleToSync(c.ID)
	if !ok {
		logger.Warnf("[Handler-SyncCluster] Cannot synchronize because it is a synchronizing cluster(%d:%s).", c.ID, c.Name)
		return createError(ctx, "cdm-center.manager.sync_cluster.failure-sync", internal.UnableSynchronizeStatus(c.ID, status.StatusCode))
	}

	go func() {
		unlock, err := internal.ClusterDBLock(c.ID)
		if err != nil {
			logger.Warnf("[Handler-SyncCluster] Could not cluster(%d) lock. Cause: %+v", c.ID, err)
			return
		}

		defer unlock()

		if err = sync.Cluster(&model.Cluster{ID: c.ID}); err != nil {
			logger.Errorf("[Handler-SyncCluster] Some errors occurred during synchronizing the cluster(%d:%s).", c.ID, c.Name)
		}
	}()

	return errors.StatusOK(ctx, "cdm-center.manager.sync_cluster.success", nil)
}

// GetClusterSyncStatus 클러스터 동기화 상태 조회
func (h *ClusterManagerHandler) GetClusterSyncStatus(ctx context.Context, req *cms.SyncClusterRequest, rsp *cms.GetClusterSyncStatusResponse) error {
	logger.Debug("Received ClusterManager.GetClusterSyncStatus request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterSyncStatus] Could not get cluster sync status. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_sync_status.failure-validate_request", err)
	}

	var c *model.Cluster
	if err = database.Transaction(func(db *gorm.DB) error {
		c, err = cluster.ValidateSyncCluster(ctx, db, req.ClusterId)
		return err
	}); err != nil {
		return createError(ctx, "cdm-center.manager.get_sync_cluster_status.failure-validate_sync_cluster", err)
	}

	rsp.Id = c.ID
	status, err := cluster.GetStatus(c.ID)
	if err != nil {
		logger.Warnf("[Handler-GetClusterSyncStatus] Could not get status of cluster(%d:%s). Cause: %+v", c.ID, c.Name, err)
		status.StatusCode = constant.ClusterSyncStateCodeUnknown
	}

	if status.StatusCode == constant.ClusterSyncStateCodeInit || status.StatusCode == constant.ClusterSyncStateCodeUnknown {
		// 동기화를 새롭게 시작할때 진행상태 데이터를 초기화하여 UI 에서 처음에 100으로 들어가는 일이 없게 map 데이터 초기화
		h.syncStatus[c.ID] = queue.SyncClusterStatus{
			ClusterID:  c.ID,
			Completion: nil,
			Progress:   0,
		}
	}

	// 메모리 데이터에 저장되어 있는 클러스터 동기화 상태값 들고오기
	syncStatus := h.syncStatus[c.ID]

	rsp.Progress = syncStatus.Progress
	rsp.Status = status.StatusCode

	for k, v := range syncStatus.Completion {
		rsp.Completion = append(rsp.Completion, &cms.Completion{
			Resource:       k,
			ProgressStatus: v,
		})
	}

	for _, r := range status.Reasons {
		rsp.Reasons = append(rsp.Reasons, &cms.Message{
			Code:     r.Code,
			Contents: r.Contents,
		})
	}

	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_sync_status.success", nil)
}

// SyncException 클러스터 상태 제외 대상 추가, 삭제
func (h *ClusterManagerHandler) SyncException(ctx context.Context, req *cms.SyncExceptionRequest, rsp *cms.SyncExceptionResponse) error {
	logger.Debug("Received ClusterManager.SyncException request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-SyncException] Could not sync exception. Cause: %+v", err)
		}
	}()

	var c *model.Cluster
	if err = database.Transaction(func(db *gorm.DB) error {
		c, err = cluster.ValidateSyncCluster(ctx, db, req.ClusterId)
		return err
	}); err != nil {
		return createError(ctx, "cdm-center.manager.sync_exception.failure-sync", err)
	}

	networks, computes, storages, err := cluster.SyncException(req)
	if err != nil {
		return createError(ctx, "cdm-center.manager.sync_exception.failure-set_sync_exception", err)
	}

	ex := &queue.ExcludedCluster{
		ClusterID: req.ClusterId,
		Storage:   storages,
		Compute:   computes,
		Network:   networks,
	}

	b, err := json.Marshal(ex)
	if err != nil {
		return createError(ctx, "cdm-center.manager.sync_exception.failure-marshal", err)
	}

	if err = broker.Publish(constant.QueueClusterServiceException, &broker.Message{Body: b}); err != nil {
		return createError(ctx, "cdm-center.manager.sync_exception.failure-publish", err)
	}

	// 제외된 서비스를 바로 적용
	l := leader.NewLeader(
		leader.LoadDataFromOpenStack(true),
		leader.PutCheckCluster(h.checkStatus[req.ClusterId]),
		leader.PutExcluded(&queue.ExcludedCluster{Storage: storages, Compute: computes, Network: networks}),
	)

	c.StateCode = l.CheckCluster(req.ClusterId)

	if err = database.Transaction(func(db *gorm.DB) error {
		return db.Save(c).Error
	}); err != nil {
		return createError(ctx, "cdm-center.manager.sync_exception.failure-update", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.sync_exception.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.sync_exception.success", nil)
}

// GetClusterList 클러스터 목록 조회
func (h *ClusterManagerHandler) GetClusterList(ctx context.Context, req *cms.ClusterListRequest, rsp *cms.ClusterListResponse) error {
	logger.Debug("Received ClusterManager.GetClusterList request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterList] Could not get cluster list. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_list.failure-validate_request", err)
	}

	if err = database.Transaction(func(db *gorm.DB) error {
		rsp.Clusters, rsp.Pagination, err = cluster.GetList(ctx, db, req)
		return err
	}); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_list.failure-get", err)
	}

	if len(rsp.Clusters) == 0 {
		return createError(ctx, "cdm-center.manager.get_cluster_list.success-get", errors.ErrNoContent)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_list.success"}
	logger.Debug("Respond ClusterManager.GetClusterList request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_list.success", nil)
}

// AddCluster 클러스터 등록
func (h *ClusterManagerHandler) AddCluster(ctx context.Context, req *cms.AddClusterRequest, rsp *cms.ClusterResponse) error {
	logger.Debug("Received ClusterManager.AddCluster request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-AddCluster] Could not add cluster. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.add_cluster.failure-validate_request", err)
	}

	if err = database.Transaction(func(db *gorm.DB) error {
		rsp.Cluster, err = cluster.Add(ctx, db, req)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return createError(ctx, "cdm-center.manager.add_cluster.failure-add", err)
	}

	u, _ := url.Parse(rsp.Cluster.ApiServerUrl)
	ip, _, _ := net.SplitHostPort(u.Host)

	// TODO:
	var agent = Agent{
		IP:             ip,
		Port:           61001,
		Version:        "1.0.0",
		InstalledAt:    0,
		LastUpgradedAt: 0,
	}
	if err = cluster.SetAgent(rsp.Cluster.Id, agent); err != nil {
		logger.Warnf("[Handler-AddCluster] Could not set connection info of agent on cinder")
	}

	b, _ := json.Marshal(rsp.Cluster.Id)
	if err = broker.Publish(constant.TopicNoticeClusterCreated, &broker.Message{Body: b}); err != nil {
		logger.Warnf("[Handler-AddCluster] Could not publish cluster created. Cause: %+v", err)
	}

	go func() {
		unlock, err := internal.ClusterDBLock(rsp.Cluster.Id)
		if err != nil {
			logger.Warnf("[Handler-AddCluster] Could not cluster(%d) lock. Cause: %+v", rsp.Cluster.Id, err)
			return
		}

		defer unlock()

		if err = sync.Cluster(&model.Cluster{ID: rsp.Cluster.Id}); err != nil {
			logger.Warnf("[Handler-AddCluster] Some errors occurred during synchronizing the cluster(%d:%s).", rsp.Cluster.Id, rsp.Cluster.Name)
		}
	}()

	rsp.Message = &cms.Message{Code: "cdm-center.manager.add_cluster.success"}
	logger.Debug("Respond ClusterManager.AddCluster request")
	return errors.StatusOK(ctx, "cdm-center.manager.add_cluster.success", nil)
}

// GetCluster 클러스터 조회
func (h *ClusterManagerHandler) GetCluster(ctx context.Context, req *cms.ClusterRequest, rsp *cms.ClusterResponse) error {
	logger.Debug("Received ClusterManager.GetCluster request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetCluster] Could not get cluster. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster.failure-validate_request", err)
	}

	if err = database.Transaction(func(db *gorm.DB) error {
		rsp.Cluster, err = cluster.Get(ctx, db, req)
		return err
	}); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster.success"}
	logger.Debug("Respond ClusterManager.GetCluster request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster.success", nil)
}

// UpdateCluster 클러스터 수정
func (h *ClusterManagerHandler) UpdateCluster(ctx context.Context, req *cms.UpdateClusterRequest, rsp *cms.ClusterResponse) error {
	logger.Debug("Received ClusterManager.UpdateCluster request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-UpdateCluster] Could not update cluster. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.update_cluster.failure-validate_request", err)
	}

	unlock, err := internal.ClusterDBLock(req.ClusterId)
	if err != nil {
		return createError(ctx, "cdm-center.manager.update_cluster.failure-cluster_db_lock", err)
	}

	defer unlock()

	if err = database.Transaction(func(db *gorm.DB) error {
		rsp.Cluster, err = cluster.Update(ctx, db, req)
		return err
	}); err != nil {
		return createError(ctx, "cdm-center.manager.update_cluster.failure-update", err)
	}

	logger.Debug("Publishing cluster updated notice to broker")
	msg := broker.Message{Body: []byte(strconv.FormatUint(req.ClusterId, 10))}
	if err := broker.Publish(constant.TopicNoticeClusterUpdated, &msg); err != nil {
		logger.Warnf("[Handler-UpdateCluster] Could not publish cluster updated notice. Cause: %+v", errors.UnusableBroker(err))
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.update_cluster.success"}
	logger.Debug("Respond ClusterManager.UpdateCluster request")
	return errors.StatusOK(ctx, "cdm-center.manager.update_cluster.success", nil)
}

// DeleteCluster 클러스터 제거
func (h *ClusterManagerHandler) DeleteCluster(ctx context.Context, req *cms.DeleteClusterRequest, rsp *cms.DeleteClusterResponse) (err error) {
	logger.Debug("Received ClusterManager.DeleteCluster request")

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-DeleteCluster] Could not delete cluster. Cause: %+v", err)
		}
	}()

	unlock, err := internal.ClusterDBLock(req.ClusterId)
	if err != nil {
		return createError(ctx, "cdm-center.manager.delete_cluster.failure-cluster_db_lock", err)
	}

	defer unlock()

	if err = cluster.Delete(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.delete_cluster.failure-delete", err)
	}

	if err = cluster.DeleteException(req.ClusterId); err != nil {
		logger.Errorf("[Cluster-Delete] Could not delete the cluster_exception. Cause: %+v", err)
	}

	if err = config.DeleteConfig(req.ClusterId); err != nil {
		logger.Errorf("[Cluster-Delete] Could not delete the cluster_config. Cause: %+v", err)
	}

	b, _ := json.Marshal(req.ClusterId)

	if err = broker.Publish(constant.QueueClusterServiceDelete, &broker.Message{Body: b}); err != nil {
		logger.Warnf("[Handler-DeleteCluster] Could not publish cluster deleted. Cause: %+v", err)
	}

	if err = broker.Publish(constant.TopicNoticeClusterDeleted, &broker.Message{Body: b}); err != nil {
		logger.Warnf("[Handler-DeleteCluster] Could not publish cluster deleted. Cause: %+v", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster.success"}
	logger.Debug("Respond ClusterManager.DeleteCluster request")
	return errors.StatusOK(ctx, "cdm-center.manager.delete_cluster.success", nil)
}

// GetClusterHypervisorList 클러스터 하이퍼바이저 목록 조회
func (h *ClusterManagerHandler) GetClusterHypervisorList(ctx context.Context, req *cms.ClusterHypervisorListRequest, rsp *cms.ClusterHypervisorListResponse) error {
	logger.Debug("Received ClusterManager.GetClusterHypervisorList request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterHypervisorList] Could not get cluster hypervisor list. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_hypervisor_list.failure-validate_request", err)
	}

	if rsp.Hypervisors, rsp.Pagination, err = hypervisor.GetList(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_hypervisor_list.failure-get", err)
	}

	if len(rsp.Hypervisors) == 0 {
		return createError(ctx, "cdm-center.manager.get_cluster_hypervisor_list.success-get", errors.ErrNoContent)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_hypervisor_list.success"}
	logger.Debug("Respond ClusterManager.GetClusterHypervisorList request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_hypervisor_list.success", nil)
}

// GetClusterHypervisor 클러스터 하이퍼바이저 조회
func (h *ClusterManagerHandler) GetClusterHypervisor(ctx context.Context, req *cms.ClusterHypervisorRequest, rsp *cms.ClusterHypervisorResponse) error {
	logger.Debug("Received ClusterManager.GetClusterHypervisor request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterHypervisor] Could not get cluster hypervisor. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_hypervisor.failure-validate_request", err)
	}

	if rsp.Hypervisor, err = hypervisor.Get(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_hypervisor.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_hypervisor.success"}
	logger.Debug("Respond ClusterManager.GetClusterHypervisor request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_hypervisor.success", nil)
}

// UpdateClusterHypervisor 클러스터 하이퍼바이저 추가정보 수정
func (h *ClusterManagerHandler) UpdateClusterHypervisor(ctx context.Context, req *cms.UpdateClusterHypervisorRequest, rsp *cms.ClusterHypervisorResponse) error {
	logger.Debug("Received ClusterManager.UpdateClusterHypervisor request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-UpdateClusterHypervisor] Could not update cluster hypervisor. Cause: %+v", err)
		}
	}()

	unlock, err := internal.ClusterDBLock(req.ClusterHypervisorId)
	if err != nil {
		return createError(ctx, "cdm-center.manager.update_cluster_hypervisor.failure-cluster_db_lock", err)
	}

	defer unlock()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.update_cluster_hypervisor.failure-validate_request", err)
	}

	if rsp.Hypervisor, err = hypervisor.Update(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.update_cluster_hypervisor.failure-update", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.update_cluster_hypervisor.success"}
	logger.Debug("Respond ClusterManager.UpdateClusterHypervisor request")
	return errors.StatusOK(ctx, "cdm-center.manager.update_cluster_hypervisor.success", nil)
}

// GetClusterAvailabilityZoneList 클러스터 가용구역 목록 조회
func (h *ClusterManagerHandler) GetClusterAvailabilityZoneList(ctx context.Context, req *cms.ClusterAvailabilityZoneListRequest, rsp *cms.ClusterAvailabilityZoneListResponse) error {
	logger.Debug("Received ClusterManager.GetClusterAvailabilityZoneList request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterAvailabilityZoneList] Could not get availability zone list. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_availability_zone_list.failure-validate_request", err)
	}

	if rsp.AvailabilityZones, rsp.Pagination, err = availabilityZone.GetList(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_availability_zone_list.failure-get", err)
	}

	if len(rsp.AvailabilityZones) == 0 {
		return createError(ctx, "cdm-center.manager.get_cluster_availability_zone_list.success-get", errors.ErrNoContent)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_availability_zone_list.success"}
	logger.Debug("Respond ClusterManager.GetClusterAvailabilityZoneList request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_availability_zone_list.success", nil)
}

// GetClusterAvailabilityZone 클러스터 가용구역 조회
func (h *ClusterManagerHandler) GetClusterAvailabilityZone(ctx context.Context, req *cms.ClusterAvailabilityZoneRequest, rsp *cms.ClusterAvailabilityZoneResponse) error {
	logger.Debug("Received ClusterManager.ClusterAvailabilityZone request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterAvailabilityZone] Could not get cluster availability zone. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_availability_zone.failure-validate_request", err)
	}

	if rsp.AvailabilityZone, err = availabilityZone.Get(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_availability_zone.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_availability_zone.success"}
	logger.Debug("Respond ClusterManager.GetClusterAvailabilityZone request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_availability_zone.success", nil)
}

// GetClusterTenantList 클러스터 테넌트 목록 조회
func (h *ClusterManagerHandler) GetClusterTenantList(ctx context.Context, req *cms.ClusterTenantListRequest, rsp *cms.ClusterTenantListResponse) error {
	logger.Debug("Received ClusterManager.GetClusterTenantList request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterTenantList] Could not get cluster tenant list. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_tenant_list.failure-validate_request", err)
	}

	if rsp.Tenants, rsp.Pagination, err = tenant.GetList(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_tenant_list.failure-get", err)
	}

	if len(rsp.Tenants) == 0 {
		return createError(ctx, "cdm-center.manager.get_cluster_tenant_list.success-get", errors.ErrNoContent)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_tenant_list.success"}
	logger.Debug("Respond ClusterManager.GetClusterTenantList request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_tenant_list.success", nil)
}

// CreateClusterTenant 클러스터 테넌트 생성
func (h *ClusterManagerHandler) CreateClusterTenant(ctx context.Context, req *cms.CreateClusterTenantRequest, rsp *cms.ClusterTenantResponse) error {
	logger.Debug("Received ClusterManager.CreateClusterTenant request")
	defer logger.Debug("Respond ClusterManager.CreateClusterTenant request")

	var err error
	rsp.Tenant, err = tenant.CreateTenant(req)
	if err != nil {
		logger.Errorf("[Handler-CreateClusterTenant] Could not create cluster tenant. Cause: %+v", err)
		return createError(ctx, "cdm-center.manager.create_cluster_tenant.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_tenant.success"}
	return nil
}

// GetClusterTenant 클러스터 테넌트 조회
func (h *ClusterManagerHandler) GetClusterTenant(ctx context.Context, req *cms.ClusterTenantRequest, rsp *cms.ClusterTenantResponse) error {
	logger.Debug("Received ClusterManager.GetClusterTenant request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterTenant] Could not get cluster tenant. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_tenant.failure-validate_request", err)
	}

	if rsp.Tenant, err = tenant.Get(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_tenant.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_tenant.success"}
	logger.Debug("Respond ClusterManager.GetClusterTenant request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_tenant.success", nil)
}

/* 미사용으로 주석처리
// GetClusterTenantByUUID uuid 를 통해 클러스터 테넌트 조회
func (h *ClusterManagerHandler) GetClusterTenantByUUID(ctx context.Context, req *cms.ClusterTenantByUUIDRequest, rsp *cms.ClusterTenantResponse) error {
	logger.Debug("Received ClusterManager.GetClusterTenantByUUID request")
	defer logger.Debug("Respond ClusterManager.GetClusterTenantByUUID request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterTenantByUUID] Could not get cluster tenant. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_tenant_by_uuid.failure-validate_request", err)
	}

	if rsp.Tenant, err = tenant.GetByUUID(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_tenant_by_uuid.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_tenant_by_uuid.success"}
	return nil
}
*/

// DeleteClusterTenant 클러스터 테넌트 삭제
func (h *ClusterManagerHandler) DeleteClusterTenant(ctx context.Context, req *cms.DeleteClusterTenantRequest, rsp *cms.DeleteClusterTenantResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterTenant request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterTenant request")

	if err := tenant.DeleteTenant(req); err != nil {
		logger.Errorf("[Handler-DeleteClusterTenant] Could not delete cluster tenant. Cause: %+v", err)
		return createError(ctx, "cdm-center.manager.delete_cluster_tenant.failure-delete", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_tenant.success"}
	return nil
}

// CheckIsExistClusterTenant 테넌트 이름을 통한 클러스터 테넌트 존재유무 확인
func (h *ClusterManagerHandler) CheckIsExistClusterTenant(ctx context.Context, req *cms.CheckIsExistByNameRequest, rsp *cms.CheckIsExistResponse) error {
	logger.Debug("Received ClusterManager.CheckIsExistClusterTenant request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-CheckIsExistClusterTenant] Could not check cluster tenant exist. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_tenant.failure-validate_request", err)
	}

	if rsp.IsExist, err = tenant.CheckIsExistTenant(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_tenant.failure-get", err)
	}

	logger.Debug("Respond ClusterManager.CheckIsExistClusterTenant request")
	rsp.Message = &cms.Message{Code: "cdm-center.manager.check_is_exist_cluster_tenant.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.check_is_exist_cluster_tenant.success", nil)
}

// GetClusterNetworkList 클러스터 네트워크 목록 조회
func (h *ClusterManagerHandler) GetClusterNetworkList(ctx context.Context, req *cms.ClusterNetworkListRequest, rsp *cms.ClusterNetworkListResponse) error {
	logger.Debug("Received ClusterManager.GetClusterNetworkList request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterNetworkList] Could not get cluster network list. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_network_list.failure-validate_request", err)
	}

	if rsp.Networks, rsp.Pagination, err = network.GetList(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_network_list.failure-get", err)
	}

	if len(rsp.Networks) == 0 {
		return createError(ctx, "cdm-center.manager.get_cluster_network_list.success-get", errors.ErrNoContent)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_network_list.success"}
	logger.Debug("Respond ClusterManager.GetClusterNetworkList request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_network_list.success", nil)
}

// CreateClusterNetwork 클러스터 네트워크 생성
func (h *ClusterManagerHandler) CreateClusterNetwork(ctx context.Context, req *cms.CreateClusterNetworkRequest, rsp *cms.ClusterNetworkResponse) error {
	logger.Debug("Received ClusterManager.CreateClusterNetwork request")
	defer logger.Debug("Respond ClusterManager.CreateClusterNetwork request")

	var err error
	rsp.Network, err = network.CreateNetwork(req)
	if err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-CreateClusterNetwork] Could not create cluster network. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-CreateClusterNetwork] Could not create cluster network. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.create_cluster_network.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_network.success"}
	return nil
}

// DeleteClusterNetwork 클러스터 네트워크 삭제
func (h *ClusterManagerHandler) DeleteClusterNetwork(ctx context.Context, req *cms.DeleteClusterNetworkRequest, rsp *cms.DeleteClusterNetworkResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterNetwork request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterNetwork request")

	if err := network.DeleteNetwork(req); err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-DeleteClusterNetwork] Could not delete cluster network. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-DeleteClusterNetwork] Could not delete cluster network. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.delete_cluster_network.failure-delete", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_network.success"}
	return nil
}

// GetClusterNetwork 클러스터 네트워크 조회
func (h *ClusterManagerHandler) GetClusterNetwork(ctx context.Context, req *cms.ClusterNetworkRequest, rsp *cms.ClusterNetworkResponse) error {
	logger.Debug("Received ClusterManager.GetClusterNetwork request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterNetwork] Could not get cluster network. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_network.failure-validate_request", err)
	}

	if rsp.Network, err = network.Get(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_network.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_network.success"}
	logger.Debug("Respond ClusterManager.GetClusterNetwork request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_network.success", nil)
}

// GetClusterNetworkByUUID uuid 를 통해 클러스터 네트워크 조회
func (h *ClusterManagerHandler) GetClusterNetworkByUUID(ctx context.Context, req *cms.ClusterNetworkByUUIDRequest, rsp *cms.ClusterNetworkResponse) error {
	logger.Debug("Received ClusterManager.GetClusterNetworkByUUID request")
	defer logger.Debug("Respond ClusterManager.GetClusterNetworkByUUID request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterNetworkByUUID] Could not get cluster network. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_network_by_uuid.failure-validate_request", err)
	}

	if rsp.Network, err = network.GetByUUID(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_network_by_uuid.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_network_by_uuid.success"}
	return nil
}

// CreateClusterSubnet 클러스터 서브넷 생성
func (h *ClusterManagerHandler) CreateClusterSubnet(ctx context.Context, req *cms.CreateClusterSubnetRequest, rsp *cms.ClusterSubnetResponse) error {
	logger.Debug("Received ClusterManager.CreateClusterSubnet request")
	defer logger.Debug("Respond ClusterManager.CreateClusterSubnet request")

	var err error
	rsp.Subnet, err = network.CreateSubnet(req)
	if err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-CreateClusterSubnet] Could not create cluster subnet. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-CreateClusterSubnet] Could not create cluster subnet. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.create_cluster_subnet.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_subnet.success"}
	return nil
}

/* 미사용으로 주석처리
// DeleteClusterSubnet 클러스터 서브넷 삭제
func (h *ClusterManagerHandler) DeleteClusterSubnet(ctx context.Context, req *cms.DeleteClusterSubnetRequest, rsp *cms.DeleteClusterSubnetResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterSubnet request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterSubnet request")

	if err := network.DeleteSubnet(req); err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-DeleteClusterSubnet] Could not delete cluster subnet. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-DeleteClusterSubnet] Could not delete cluster subnet. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.delete_cluster_subnet.failure-delete", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_subnet.success"}
	return nil
}
*/

// GetClusterSubnet 클러스터 서브넷 조회
func (h *ClusterManagerHandler) GetClusterSubnet(ctx context.Context, req *cms.ClusterSubnetRequest, rsp *cms.ClusterSubnetResponse) error {
	logger.Debug("Received ClusterManager.GetClusterSubnet request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterSubnet] Could not get cluster subnet. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_subnet.failure-validate_request", err)
	}

	if rsp.Subnet, err = network.GetSubnet(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_subnet.failure-get", err)
	}

	logger.Debug("Respond ClusterManager.GetClusterSubnet request")
	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_subnet.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_subnet.success", nil)
}

// CreateClusterFloatingIP 클러스터 FloatingIP 생성
func (h *ClusterManagerHandler) CreateClusterFloatingIP(ctx context.Context, req *cms.CreateClusterFloatingIPRequest, rsp *cms.ClusterFloatingIPResponse) error {
	logger.Debug("Received ClusterManager.CreateClusterFloatingIP request")
	defer logger.Debug("Respond ClusterManager.CreateClusterFloatingIP request")

	var err error
	rsp.FloatingIp, err = floatingip.CreateFloatingIP(req)
	if err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-CreateClusterFloatingIP] Could not create cluster floating ip. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-CreateClusterFloatingIP] Could not create cluster floating ip. Cause: %+v", err)
		}
		return createError(ctx, "cdm-center.manager.create_cluster_floating_ip.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_floating_ip.success"}
	return nil
}

// DeleteClusterFloatingIP 클러스터 FloatingIP 삭제
func (h *ClusterManagerHandler) DeleteClusterFloatingIP(ctx context.Context, req *cms.DeleteClusterFloatingIPRequest, rsp *cms.DeleteClusterFloatingIPResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterFloatingIP request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterFloatingIP request")

	if err := floatingip.DeleteFloatingIP(req); err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-DeleteClusterFloatingIP] Could not delete cluster floating ip. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-DeleteClusterFloatingIP] Could not delete cluster floating ip. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.delete_cluster_floating_ip.failure-delete", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_floating_ip.success"}
	return nil
}

// GetClusterFloatingIP 클러스터 FloatingIP 조회
func (h *ClusterManagerHandler) GetClusterFloatingIP(ctx context.Context, req *cms.ClusterFloatingIPRequest, rsp *cms.ClusterFloatingIPResponse) error {
	logger.Debug("Received ClusterManager.GetClusterFloatingIP request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterFloatingIP] Could not get cluster floating ip. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_floating_ip.failure-validate_request", err)
	}

	if rsp.FloatingIp, err = floatingip.GetFloatingIP(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_floating_ip.failure-get", err)
	}

	logger.Debug("Respond ClusterManager.GetClusterFloatingIP request")
	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_floating_ip.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_floating_ip.success", nil)
}

// CheckIsExistClusterFloatingIP ip address 를 통한 클러스터 FloatingIP 존재유무 확인
func (h *ClusterManagerHandler) CheckIsExistClusterFloatingIP(ctx context.Context, req *cms.CheckIsExistClusterFloatingIPRequest, rsp *cms.CheckIsExistResponse) error {
	logger.Debug("Received ClusterManager.CheckIsExistClusterFloatingIP request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-CheckIsExistClusterFloatingIP] Could not check cluster floating ip exist. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_floating_ip.failure-validate_request", err)
	}

	if rsp.IsExist, err = floatingip.CheckIsExistFloatingIP(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_floating_ip.failure-get", err)
	}

	logger.Debug("Respond ClusterManager.CheckIsExistClusterFloatingIP request")
	rsp.Message = &cms.Message{Code: "cdm-center.manager.check_is_exist_cluster_floating_ip.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.check_is_exist_cluster_floating_ip.success", nil)
}

// CreateClusterRouter 클러스터 라우터 생성
func (h *ClusterManagerHandler) CreateClusterRouter(ctx context.Context, req *cms.CreateClusterRouterRequest, rsp *cms.ClusterRouterResponse) error {
	logger.Debug("Received ClusterManager.CreateClusterRouter request")
	defer logger.Debug("Respond ClusterManager.CreateClusterRouter request")

	var err error
	rsp.Router, err = router.CreateRouter(req)
	if err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-CreateClusterRouter] Could not create cluster router. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-CreateClusterRouter] Could not create cluster router. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.create_cluster_router.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_router.success"}
	return nil
}

// DeleteClusterRouter 클러스터 라우터 삭제
func (h *ClusterManagerHandler) DeleteClusterRouter(ctx context.Context, req *cms.DeleteClusterRouterRequest, rsp *cms.DeleteClusterRouterResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterRouter request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterRouter request")

	if err := router.DeleteRouter(req); err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-DeleteClusterRouter] Could not delete cluster router. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-DeleteClusterRouter] Could not delete cluster router. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.delete_cluster_router.failure-delete", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_router.success"}
	return nil
}

// GetClusterRouterList 클러스터 네트워크 라우터 목록 조회
func (h *ClusterManagerHandler) GetClusterRouterList(ctx context.Context, req *cms.ClusterRouterListRequest, rsp *cms.ClusterRouterListResponse) error {
	logger.Debug("Respond ClusterManager.GetClusterRouterList request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterRouterList] Could not get cluster network router list. Cause: %+v", err)
		}
	}()
	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_router_list.failure-validate_request", err)
	}

	if rsp.Routers, rsp.Pagination, err = router.GetList(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_router_list.failure-get", err)
	}

	if len(rsp.Routers) == 0 {
		return createError(ctx, "cdm-center.manager.get_cluster_router_list.success-get", errors.ErrNoContent)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_router_list.success"}
	logger.Debug("Respond ClusterManager.GetClusterRouterList request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_router_list.success", nil)
}

// GetClusterRouter 클러스터 네트워크 라우터 조회
func (h *ClusterManagerHandler) GetClusterRouter(ctx context.Context, req *cms.ClusterRouterRequest, rsp *cms.ClusterRouterResponse) error {
	logger.Debug("Respond ClusterManager.GetClusterRouter request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterRouter] Could not get cluster router. Cause: %+v", err)
		}
	}()
	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_router.failure-validate_request", err)
	}

	if rsp.Router, err = router.Get(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_router.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_router.success"}
	logger.Debug("Respond ClusterManager.GetClusterRouter request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_router.success", nil)
}

/* 미사용으로 주석처리
// GetClusterRouterByUUID uuid 를 통해 클러스터 네트워크 라우터 조회
func (h *ClusterManagerHandler) GetClusterRouterByUUID(ctx context.Context, req *cms.ClusterRouterByUUIDRequest, rsp *cms.ClusterRouterResponse) error {
	logger.Debug("Received ClusterManager.GetClusterRouterByUUID request")
	defer logger.Debug("Respond ClusterManager.GetClusterRouterByUUID request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterRouterByUUID] Could not get cluster router. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_router_by_uuid.failure-validate_request", err)
	}

	if rsp.Router, err = router.GetByUUID(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_router_by_uuid.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_router_by_uuid.success"}
	return nil
}
*/

// CheckIsExistClusterRoutingInterface ip address 를 통한 클러스터 Routing Interface 존재유무 확인
func (h *ClusterManagerHandler) CheckIsExistClusterRoutingInterface(ctx context.Context, req *cms.CheckIsExistClusterRoutingInterfaceRequest, rsp *cms.CheckIsExistResponse) error {
	logger.Debug("Received ClusterManager.CheckIsExistClusterRoutingInterface request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-CheckIsExistClusterRoutingInterface] Could not check cluster routing interface exist. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_routing_interface.failure-validate_request", err)
	}

	if rsp.IsExist, err = router.CheckIsExistRoutingInterface(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_routing_interface.failure-get", err)
	}

	logger.Debug("Respond ClusterManager.CheckIsExistClusterRoutingInterface request")
	rsp.Message = &cms.Message{Code: "cdm-center.manager.check_is_exist_cluster_routing_interface.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.check_is_exist_cluster_routing_interface.success", nil)
}

// GetClusterStorageList 클러스터 볼륨타입 목록 조회
func (h *ClusterManagerHandler) GetClusterStorageList(ctx context.Context, req *cms.ClusterStorageListRequest, rsp *cms.ClusterStorageListResponse) error {
	logger.Debug("Received ClusterManager.GetClusterStorageList request")
	var err error

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterStorageList] Could not get cluster storage list. Cause: %+v", err)
		}
	}()

	if rsp.Storages, rsp.Pagination, err = storage.GetList(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_storage_list.failure-get", err)
	}

	if len(rsp.Storages) == 0 {
		return createError(ctx, "cdm-center.manager.get_cluster_storage_list.success-get", errors.ErrNoContent)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_storage_list.success"}
	logger.Debug("Respond ClusterManager.GetClusterStorageList request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_storage_list.success", nil)
}

// GetClusterStorage 클러스터 볼륨타입 조회
func (h *ClusterManagerHandler) GetClusterStorage(ctx context.Context, req *cms.ClusterStorageRequest, rsp *cms.ClusterStorageResponse) error {
	var err error
	logger.Debug("Received ClusterManager.GetClusterStorage request")

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterStorage] Could not get cluster storage. Cause: %+v", err)
		}
	}()

	if rsp.Storage, err = storage.Get(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_storage.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_storage.success"}
	logger.Debug("Respond ClusterManager.GetClusterStorage request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_storage.success", nil)
}

// UpdateClusterStorageMetadata 클러스터 볼륨타입 메타데이터 수정
func (h *ClusterManagerHandler) UpdateClusterStorageMetadata(ctx context.Context, req *cms.UpdateClusterStorageMetadataRequest, rsp *cms.ClusterStorageMetadataResponse) error {
	var err error
	logger.Debug("Received ClusterManager.UpdateClusterStorageMetadata request")

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-UpdateClusterStorageMetadata] Could not update cluster storage metadata. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.update_cluster_storage_metadata.failure-validate_request", err)
	}

	if req.GetClusterId() == 0 {
		err = errors.RequiredParameter("cluster_id")
		return createError(ctx, "cdm-center.manager.update_cluster_storage_metadata.failure-get_storage", err)
	}

	if req.GetClusterStorageId() == 0 {
		err = errors.RequiredParameter("cluster_storage_id")
		return createError(ctx, "cdm-center.manager.update_cluster_storage_metadata.failure-get_storage", err)
	}

	s, err := clusterStorage.GetStorage(req.GetClusterId(), req.GetClusterStorageId())
	if err != nil {
		return createError(ctx, "cdm-center.manager.update_cluster_storage_metadata.failure-get_storage", err)
	}

	md := make(map[string]interface{})
	for k, v := range req.GetMetadata() {
		md[k] = v
	}

	if err = s.MergeMetadata(md); err != nil {
		return createError(ctx, "cdm-center.manager.update_cluster_storage_metadata.failure-merge", err)
	}

	result, err := s.GetMetadata()
	if err != nil {
		return createError(ctx, "cdm-center.manager.update_cluster_storage_metadata.failure-get", err)
	}

	rsp.Metadata = make(map[string]string)
	for k, v := range result {
		rsp.Metadata[k] = fmt.Sprint(v)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.update_cluster_storage_metadata.success"}
	logger.Debug("Respond ClusterManager.UpdateClusterStorageMetadata request")
	return errors.StatusOK(ctx, "cdm-center.manager.update_cluster_storage_metadata.success", nil)
}

// GetClusterVolumeGroupList 클러스터 볼륨 그룹 목록을 조회한다.
func (h *ClusterManagerHandler) GetClusterVolumeGroupList(ctx context.Context, req *cms.GetClusterVolumeGroupListRequest, rsp *cms.GetClusterVolumeGroupListResponse) error {
	logger.Debug("Received ClusterManager.GetClusterVolumeGroupList request")
	defer logger.Debug("Respond ClusterManager.GetClusterVolumeGroupList request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterVolumeGroupList] Could not get cluster volume group list. Cause: %+v", err)
		}
	}()

	if err = database.Transaction(func(db *gorm.DB) error {
		rsp.VolumeGroups, err = volumegroup.GetVolumeGroupList(db, req)
		return err
	}); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_volume_group_list.failure-get", err)
	}

	if len(rsp.VolumeGroups) == 0 {
		return createError(ctx, "cdm-center.manager.get_cluster_volume_group_list.success-get", errors.ErrNoContent)
	}

	logger.Debug("Respond ClusterManager.GetClusterVolumeGroupList request")
	return nil
}

// GetClusterVolumeGroupByUUID 클러스터 볼륨 그룹을 조회한다.
func (h *ClusterManagerHandler) GetClusterVolumeGroupByUUID(ctx context.Context, req *cms.GetClusterVolumeGroupByUUIDRequest, rsp *cms.GetClusterVolumeGroupResponse) error {
	logger.Debug("Received ClusterManager.GetClusterVolumeGroupByUUID request")
	defer logger.Debug("Respond ClusterManager.GetClusterVolumeGroupByUUID request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterVolumeGroupByUUID] Could not get cluster volume group. Cause: %+v", err)
		}
	}()

	if err = database.Transaction(func(db *gorm.DB) error {
		rsp.VolumeGroup, err = volumegroup.GetVolumeGroupByUUID(ctx, db, req)
		return err
	}); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_volume_group.failure-get", err)
	}

	logger.Debug("Respond ClusterManager.GetClusterVolumeGroupByUUID request")
	return nil
}

// CreateClusterVolumeGroup 클러스터 볼륨 그룹 생성
func (h *ClusterManagerHandler) CreateClusterVolumeGroup(ctx context.Context, req *cms.CreateClusterVolumeGroupRequest, rsp *cms.CreateClusterVolumeGroupResponse) error {
	logger.Debug("Received ClusterManager.CreateClusterVolumeGroup request")
	defer logger.Debug("Respond ClusterManager.CreateClusterVolumeGroup request")

	var err error
	rsp.VolumeGroup, err = volumegroup.CreateVolumeGroup(req)
	if err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-CreateClusterVolumeGroup] Could not create cluster volume group. tenant: %s, Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-CreateClusterVolumeGroup] Could not create cluster volume group. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.create_cluster_volume_group.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_volume_group.success"}
	return nil
}

// DeleteClusterVolumeGroup 클러스터 볼륨 그룹 삭제
func (h *ClusterManagerHandler) DeleteClusterVolumeGroup(ctx context.Context, req *cms.DeleteClusterVolumeGroupRequest, rsp *cms.DeleteClusterVolumeGroupResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterVolumeGroup request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterVolumeGroup request")

	if err := volumegroup.DeleteVolumeGroup(req); err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-DeleteClusterVolumeGroup] Could not delete cluster volume group. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-DeleteClusterVolumeGroup] Could not delete cluster volume group. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.delete_cluster_volume_group.failure-delete", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_volume_group.success"}
	return nil
}

// UpdateClusterVolumeGroup 클러스터 볼륨 그룹 수정
func (h *ClusterManagerHandler) UpdateClusterVolumeGroup(ctx context.Context, req *cms.UpdateClusterVolumeGroupRequest, rsp *cms.UpdateClusterVolumeGroupResponse) error {
	logger.Debug("Received ClusterManager.UpdateClusterVolumeGroup request")
	defer logger.Debug("Respond ClusterManager.UpdateClusterVolumeGroup request")

	if err := volumegroup.UpdateVolumeGroup(req); err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-UpdateClusterVolumeGroup] Could not update cluster volume group. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-UpdateClusterVolumeGroup] Could not update cluster volume group. Cause: %+v", err)
		}
		return createError(ctx, "cdm-center.manager.update_cluster_volume_group.failure-update", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.update_cluster_volume_group.success"}
	return nil
}

// CreateClusterVolumeGroupSnapshot 클러스터 볼륨 그룹 스냅샷 생성
func (h *ClusterManagerHandler) CreateClusterVolumeGroupSnapshot(ctx context.Context, req *cms.CreateClusterVolumeGroupSnapshotRequest, rsp *cms.CreateClusterVolumeGroupSnapshotResponse) error {
	logger.Debug("Received ClusterManager.CreateClusterVolumeGroupSnapshot request")
	defer logger.Debug("Respond ClusterManager.CreateClusterVolumeGroupSnapshot request")

	return createError(ctx, "cdm-center.manager.create_cluster_volume_group_snapshot.failure-create", clusterStorage.ErrUnsupportedStorageType)
}

// GetClusterVolumeGroupSnapshotList 클러스터 볼륨 그룹 목록 조회
func (h *ClusterManagerHandler) GetClusterVolumeGroupSnapshotList(ctx context.Context, req *cms.GetClusterVolumeGroupSnapshotListRequest, rsp *cms.GetClusterVolumeGroupSnapshotListResponse) error {
	logger.Debug("Received ClusterManager.GetClusterVolumeGroupSnapshotList request")
	defer logger.Debug("Respond ClusterManager.GetClusterVolumeGroupSnapshotList request")

	var err error
	if rsp.VolumeGroupSnapshots, err = volumegroup.GetClusterVolumeGroupSnapshotList(req); err != nil {
		logger.Errorf("[Handler-GetClusterVolumeGroupSnapshotList] Could not get cluster volume group snapshot list. Cause: %+v", err)
		return createError(ctx, "cdm-center.manager.get_cluster_volume_group_snapshot_list.failure", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_volume_group_snapshot_list.success"}
	return nil
}

// DeleteClusterVolumeGroupSnapshot 클러스터 볼륨 그룹 스냅샷 삭제
func (h *ClusterManagerHandler) DeleteClusterVolumeGroupSnapshot(ctx context.Context, req *cms.DeleteClusterVolumeGroupSnapshotRequest, rsp *cms.DeleteClusterVolumeGroupSnapshotResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterVolumeGroupSnapshot request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterVolumeGroupSnapshot request")

	unlock, err := internal.ClusterVolumeGroupSnapshotLock(req.VolumeGroup.Uuid)
	if err != nil {
		if req.GetTenant() != nil {
			logger.Warnf("[Handler-DeleteClusterVolumeGroupSnapshot] Could not lock the cluster volume group snapshot(%s). tenant: %s Cause: %+v", req.VolumeGroup.Uuid, req.GetTenant().GetUuid(), err)
		} else {
			logger.Warnf("[Handler-DeleteClusterVolumeGroupSnapshot] Could not lock the cluster volume group snapshot(%s). Cause: %+v", req.VolumeGroup.Uuid, err)
		}

		return createError(ctx, "cdm-center.manager.delete_cluster_volume_group_snapshot.failure-volume_group_snapshot_db_lock", err)
	}

	defer func() {
		unlock()
	}()

	if err = volumegroup.DeleteVolumeGroupSnapshot(req); err != nil {
		logger.Errorf("[Handler-DeleteClusterVolumeGroupSnapshot] Could not delete cluster volume group snapshot. Cause: %+v", err)
		return createError(ctx, "cdm-center.manager.delete_cluster_volume_group_snapshot.failure-delete", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_volume_group_snapshot.success"}
	return nil
}

// GetClusterVolumeList 클러스터 볼륨 목록 조회
func (h *ClusterManagerHandler) GetClusterVolumeList(ctx context.Context, req *cms.ClusterVolumeListRequest, rsp *cms.ClusterVolumeListResponse) error {
	var err error
	logger.Debug("Received ClusterManager.GetClusterVolumeList request")

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterVolumeList] Could not get cluster volume list. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_volume_list.failure-validate_request", err)
	}

	if rsp.Volumes, rsp.Pagination, err = volume.GetList(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_volume_list.failure-get", err)
	}

	if len(rsp.Volumes) == 0 {
		return createError(ctx, "cdm-center.manager.get_cluster_volume_list.success-get", errors.ErrNoContent)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_volume_list.success"}
	logger.Debug("Respond ClusterManager.GetClusterVolumeList request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_volume_list.success", nil)
}

// GetClusterVolume 클러스터 볼륨 조회
func (h *ClusterManagerHandler) GetClusterVolume(ctx context.Context, req *cms.ClusterVolumeRequest, rsp *cms.ClusterVolumeResponse) error {
	var err error
	logger.Debug("Received ClusterManager.GetClusterVolume request")

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterVolume] Could not get cluster volume. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_volume.failure-validate_request", err)
	}

	rsp.Volume, err = volume.Get(ctx, req)
	if err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_volume.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_volume.success"}
	logger.Debug("Respond ClusterManager.GetClusterVolume request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_volume.success", nil)
}

/* 미사용으로 주석처리
// GetClusterVolumeByUUID uuid 를 통해 클러스터 볼륨 조회
func (h *ClusterManagerHandler) GetClusterVolumeByUUID(ctx context.Context, req *cms.ClusterVolumeByUUIDRequest, rsp *cms.ClusterVolumeResponse) error {
	logger.Debug("Received ClusterManager.GetClusterVolumeByUUID request")
	defer logger.Debug("Respond ClusterManager.GetClusterVolumeByUUID request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterVolumeByUUID] Could not get cluster volume. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_volume_by_uuid.failure-validate_request", err)
	}

	if rsp.Volume, err = volume.GetByUUID(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_volume_by_uuid.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_volume_by_uuid.success"}
	return nil
}
*/

/* 미사용으로 주석처리
// CreateClusterVolume 클러스터 볼륨 생성
func (h *ClusterManagerHandler) CreateClusterVolume(ctx context.Context, req *cms.CreateClusterVolumeRequest, rsp *cms.CreateClusterVolumeResponse) error {
	logger.Debug("Received ClusterManager.CreateClusterVolume request")
	defer logger.Debug("Respond ClusterManager.CreateClusterVolume request")

	var err error
	rsp.Volume, err = volume.CreateVolume(req)
	if err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-CreateClusterVolume] Could not create cluster volume. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-CreateClusterVolume] Could not create cluster volume. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.create_cluster_volume.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_volume.success"}
	return nil
}
*/

// ImportClusterVolume 클러스터 볼륨 import
func (h *ClusterManagerHandler) ImportClusterVolume(ctx context.Context, req *cms.ImportClusterVolumeRequest, rsp *cms.ImportClusterVolumeResponse) error {
	logger.Debug("Received ClusterManager.ImportClusterVolume request")
	defer logger.Debug("Respond ClusterManager.ImportClusterVolume request")

	var err error
	rsp.VolumePair, rsp.SnapshotPairs, err = volume.ImportVolume(req)
	if err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-ImportClusterVolume] Could not import cluster volume. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-ImportClusterVolume] Could not import cluster volume. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.import_cluster_volume.failure-import", err)
	}

	rsp.SourceStorage = req.SourceStorage
	rsp.TargetStorage = req.TargetStorage
	rsp.VolumePair.Target.Storage = req.TargetStorage
	rsp.Message = &cms.Message{Code: "cdm-center.manager.import_cluster_volume.success"}
	return nil
}

// CopyClusterVolume 클러스터 볼륨 copy
// 성능상에 문제로 실제 copy 는 하지 않는다.
// misnomer: PrepareMigration 정도가 적절해보임
func (h *ClusterManagerHandler) CopyClusterVolume(ctx context.Context, req *cms.CopyClusterVolumeRequest, rsp *cms.CopyClusterVolumeResponse) error {
	logger.Debug("Received ClusterManager.CopyClusterVolume request")
	defer logger.Debug("Respond ClusterManager.CopyClusterVolume request")

	//var err error
	//rsp.Volume, rsp.Snapshots, err = volume.CopyVolume(req)
	//if err != nil {
	//	if req.GetTenant() != nil {
	//		logger.Errorf("[Handler-CopyClusterVolume] Could not copy cluster volume. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
	//	} else {
	//		logger.Errorf("[Handler-CopyClusterVolume] Could not copy cluster volume. Cause: %+v", err)
	//	}
	//	return createError(ctx, "cdm-center.manager.copy_cluster_volume.failure-copy", err)
	//}

	rsp.Volume = req.Volume
	rsp.SourceStorage = req.SourceStorage
	rsp.TargetStorage = req.TargetStorage
	rsp.Message = &cms.Message{Code: "cdm-center.manager.copy_cluster_volume.success"}
	return nil
}

// DeleteClusterVolumeCopy 클러스터 볼륨 copy 삭제
func (h *ClusterManagerHandler) DeleteClusterVolumeCopy(ctx context.Context, req *cms.DeleteClusterVolumeCopyRequest, rsp *cms.DeleteClusterVolumeCopyResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterVolumeCopy request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterVolumeCopy request")

	/*if err := volume.DeleteVolumeCopy(req); err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-DeleteClusterVolumeCopy] Could not delete cluster volume copy. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-DeleteClusterVolumeCopy] Could not delete cluster volume copy. Cause: %+v", err)
		}
		return createError(ctx, "cdm-center.manager.delete_cluster_volume_copy.failure-delete", err)
	}*/

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_volume_copy.success"}
	return nil
}

// DeleteClusterVolume 클러스터 볼륨 삭제
func (h *ClusterManagerHandler) DeleteClusterVolume(ctx context.Context, req *cms.DeleteClusterVolumeRequest, rsp *cms.DeleteClusterVolumeResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterVolume request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterVolume request")

	//if err := volume.DeleteVolume(req); err != nil {
	//	if req.GetTenant() != nil {
	//		logger.Errorf("[Handler-DeleteClusterVolume] Could not delete cluster volume. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
	//	} else {
	//		logger.Errorf("[Handler-DeleteClusterVolume] Could not delete cluster volume. Cause: %+v", err)
	//	}
	//	return createError(ctx, "cdm-center.manager.delete_cluster_volume.failure-delete", err)
	//}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_volume.success"}
	return nil
}

// UnmanageClusterVolume 클러스터 볼륨 unmanage
func (h *ClusterManagerHandler) UnmanageClusterVolume(ctx context.Context, req *cms.UnmanageClusterVolumeRequest, rsp *cms.UnmanageClusterVolumeResponse) error {
	logger.Debug("Received ClusterManager.UnmanageClusterVolume request")
	defer logger.Debug("Respond ClusterManager.UnmanageClusterVolume request")

	if err := volume.UnmanageVolume(req); err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-UnmanageClusterVolume] Could not unmanage cluster volume. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-UnmanageClusterVolume] Could not unmanage cluster volume. Cause: %+v", err)
		}
		return createError(ctx, "cdm-center.manager.unmanage_cluster_volume.failure-unmanage", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.unmanage_cluster_volume.success"}
	return nil
}

// SyncClusterVolumeSnapshotList 클러스터 볼륨 스냅샷 목록 동기화
func (h *ClusterManagerHandler) SyncClusterVolumeSnapshotList(ctx context.Context, req *cms.SyncClusterVolumeSnapshotListRequest, rsp *cms.SyncClusterVolumeSnapshotListResponse) error {
	logger.Debug("Received ClusterManager.SyncClusterVolumeSnapshotList request")
	defer logger.Debug("Respond ClusterManager.SyncClusterVolumeSnapshotList request")

	unlock, err := internal.ClusterDBLock(req.Cluster.Id)
	if err != nil {
		logger.Warnf("[Handler-SyncClusterVolumeSnapshotList] Could not cluster(%d) lock. Cause: %+v", req.Cluster.Id, err)
		return createError(ctx, "cdm-center.manager.sync_cluster_volume_snapshot_list.failure-cluster_db_lock", err)
	}

	defer unlock()

	if err = volume.SyncVolumeSnapshotList(req); err != nil {
		logger.Errorf("[Handler-SyncClusterVolumeSnapshotList] Could not sync cluster volume snapshot. Cause: %+v", err)
		return createError(ctx, "cdm-center.manager.sync_cluster_volume_snapshot_list.failure-sync", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.sync_cluster_volume_snapshot_list.success"}
	return nil
}

/* 미사용으로 주석처리
// CreateClusterVolumeSnapshot 클러스터 볼륨 스냅샷 생성
func (h *ClusterManagerHandler) CreateClusterVolumeSnapshot(ctx context.Context, req *cms.CreateClusterVolumeSnapshotRequest, rsp *cms.CreateClusterVolumeSnapshotResponse) error {
	logger.Debug("Received ClusterManager.CreateClusterVolumeSnapshot request")
	defer logger.Debug("Respond ClusterManager.CreateClusterVolumeSnapshot request")

	var err error
	rsp.Snapshot, err = volume.CreateVolumeSnapshot(req)
	if err != nil {
		logger.Errorf("[Handler-CreateClusterVolumeSnapshot] Could not create cluster volume snapshot. Cause: %+v", err)
		return createError(ctx, "cdm-center.manager.create_cluster_volume_snapshot.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_volume_snapshot.success"}
	return nil
}
*/

/* 미사용으로 주석처리
// DeleteClusterVolumeSnapshot 클러스터 볼륨 스냅샷 삭제
func (h *ClusterManagerHandler) DeleteClusterVolumeSnapshot(ctx context.Context, req *cms.DeleteClusterVolumeSnapshotRequest, rsp *cms.DeleteClusterVolumeSnapshotResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterVolumeSnapshot request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterVolumeSnapshot request")

	if err := volume.DeleteVolumeSnapshot(req); err != nil {
		logger.Errorf("[Handler-DeleteClusterVolumeSnapshot] Could not delete cluster volume snapshot. Cause: %+v", err)
		return createError(ctx, "cdm-center.manager.delete_cluster_volume_snapshot.failure-delete", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_volume_snapshot.success"}
	return nil
}
*/

// GetClusterInstanceList 클러스터 인스턴스 목록 조회
func (h *ClusterManagerHandler) GetClusterInstanceList(ctx context.Context, req *cms.ClusterInstanceListRequest, rsp *cms.ClusterInstanceListResponse) error {
	logger.Debug("Received ClusterManager.GetClusterInstanceList request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterInstanceList] Could not get cluster instance list. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_list.failure-validate_request", err)
	}

	if rsp.Instances, rsp.Pagination, err = instance.GetList(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_list.failure-get", err)
	}

	if len(rsp.Instances) == 0 {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_list.success-get", errors.ErrNoContent)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_instance_list.success"}
	logger.Debug("Respond ClusterManager.GetClusterInstanceList request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_instance_list.success", nil)
}

// GetClusterInstance 클러스터 인스턴스 조회
func (h *ClusterManagerHandler) GetClusterInstance(ctx context.Context, req *cms.ClusterInstanceRequest, rsp *cms.ClusterInstanceResponse) error {
	logger.Debug("Received ClusterManager.GetClusterInstance request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterInstance] Could not get cluster instance. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance.failure-validate_request", err)
	}

	if rsp.Instance, err = instance.Get(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_instance.success"}
	logger.Debug("Respond ClusterManager.GetClusterInstance request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_instance.success", nil)
}

// GetClusterInstanceByUUID uuid 를 통해 클러스터 인스턴스 조회
func (h *ClusterManagerHandler) GetClusterInstanceByUUID(ctx context.Context, req *cms.ClusterInstanceByUUIDRequest, rsp *cms.ClusterInstanceResponse) error {
	logger.Debug("Received ClusterManager.GetClusterInstanceByUUID request")
	defer logger.Debug("Respond ClusterManager.GetClusterInstanceByUUID request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterInstanceByUUID] Could not get cluster instance. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_by_uuid.failure-validate_request", err)
	}

	if rsp.Instance, err = instance.GetByUUID(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_by_uuid.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_instance_by_uuid.success"}
	return nil
}

// GetClusterInstanceNumber 클러스터 인스턴스 수 조회
// req.ClusterId 가 지정되지 않은 경우, 권한을 가지는 모든 클러스터의 인스턴스 수를 조회한다.
func (h *ClusterManagerHandler) GetClusterInstanceNumber(ctx context.Context, req *cms.ClusterInstanceNumberRequest, rsp *cms.ClusterInstanceNumberResponse) error {
	logger.Debug("Received ClusterManager.GetClusterInstanceNumber request")
	defer logger.Debug("Respond ClusterManager.GetClusterInstanceNumber request")

	var err error
	if err = database.Transaction(func(db *gorm.DB) error {
		rsp.InstanceNumber, err = instance.GetInstanceNumber(ctx, db, req)
		return err
	}); err != nil {
		logger.Errorf("[Handler-GetClusterInstanceNumber] Could not get cluster instance number. Cause: %+v", err)
		return createError(ctx, "cdm-center.manager.get_cluster_instance_number.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_instance_number.success"}
	return nil
}

// CheckIsExistClusterInstance instance uuid 를 통한 클러스터 Instance 존재유무 확인
func (h *ClusterManagerHandler) CheckIsExistClusterInstance(ctx context.Context, req *cms.CheckIsExistClusterInstanceRequest, rsp *cms.CheckIsExistResponse) error {
	logger.Debug("Received ClusterManager.CheckIsExistClusterInstance request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-CheckIsExistClusterInstance] Could not check cluster instance exist. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_instance.failure-validate_request", err)
	}

	if rsp.IsExist, err = instance.CheckIsExistInstance(req); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_instance.failure-get", err)
	}

	logger.Debug("Respond ClusterManager.CheckIsExistClusterInstance request")
	rsp.Message = &cms.Message{Code: "cdm-center.manager.check_is_exist_cluster_instance.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.check_is_exist_cluster_instance.success", nil)
}

// CreateClusterInstance 클러스터 인스턴스 생성
func (h *ClusterManagerHandler) CreateClusterInstance(ctx context.Context, req *cms.CreateClusterInstanceRequest, rsp *cms.CreateClusterInstanceResponse) error {
	logger.Debug("Received ClusterManager.CreateClusterInstance request")
	defer logger.Debug("Respond ClusterManager.CreateClusterInstance request")

	var err error
	rsp.Instance, err = instance.CreateInstance(req)
	if err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-CreateClusterInstance] Could not create cluster instance. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-CreateClusterInstance] Could not create cluster instance. Cause: %+v", err)
		}
		return createError(ctx, "cdm-center.manager.create_cluster_instance.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_instance.success"}
	return nil
}

// DeleteClusterInstance 클러스터 인스턴스 삭제
func (h *ClusterManagerHandler) DeleteClusterInstance(ctx context.Context, req *cms.DeleteClusterInstanceRequest, rsp *cms.DeleteClusterInstanceResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterInstance request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterInstance request")

	if err := instance.DeleteInstance(req); err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-DeleteClusterInstance] Could not delete cluster instance. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-DeleteClusterInstance] Could not delete cluster instance. Cause: %+v", err)
		}
		return createError(ctx, "cdm-center.manager.delete_cluster_instance.failure-delete", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_instance.success"}
	return nil
}

/* 미사용으로 주석처리
// StartClusterInstance 클러스터 인스턴스 기동
func (h *ClusterManagerHandler) StartClusterInstance(ctx context.Context, req *cms.StartClusterInstanceRequest, rsp *cms.StartClusterInstanceResponse) error {
	logger.Debug("Received ClusterManager.StartClusterInstance request")
	defer logger.Debug("Respond ClusterManager.StartClusterInstance request")

	if err := instance.StartInstance(req); err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-StartClusterInstance] Could not start cluster instance. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-StartClusterInstance] Could not start cluster instance. Cause: %+v", err)
		}
		return createError(ctx, "cdm-center.manager.start_cluster_instance.failure-start", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.start_cluster_instance.success"}
	return nil
}
*/

// StopClusterInstance 클러스터 인스턴스 중지
func (h *ClusterManagerHandler) StopClusterInstance(ctx context.Context, req *cms.StopClusterInstanceRequest, rsp *cms.StopClusterInstanceResponse) error {
	logger.Debug("Received ClusterManager.StopClusterInstance request")
	defer logger.Debug("Respond ClusterManager.StopClusterInstance request")

	if err := instance.StopInstance(req); err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-StopClusterInstance] Could not stop cluster instance. tenant : %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-StopClusterInstance] Could not stop cluster instance. Cause: %+v", err)
		}
		return createError(ctx, "cdm-center.manager.stop_cluster_instance.failure-stop", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.stop_cluster_instance.success"}
	return nil
}

/* 미사용으로 주석처리
// GetClusterInstanceSpecList 클러스터 인스턴스 spec 목록 조회
func (h *ClusterManagerHandler) GetClusterInstanceSpecList(ctx context.Context, req *cms.ClusterInstanceSpecListRequest, rsp *cms.ClusterInstanceSpecListResponse) error {
	logger.Debug("Received ClusterManager.GetClusterInstanceSpecList request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterInstanceSpecList] Could not get cluster instance spec list. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_spec_list.failure-validate_request", err)
	}

	if err = database.Transaction(func(db *gorm.DB) error {
		rsp.Specs, rsp.Pagination, err = spec.GetList(ctx, db, req)
		return err
	}); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_spec_list.failure-get", err)
	}

	if len(rsp.Specs) == 0 {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_spec_list.success-get", errors.ErrNoContent)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_instance_spec_list.success"}
	logger.Debug("Respond ClusterManager.GetClusterInstanceSpecList request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_instance_spec_list.success", nil)
}
*/

/* 미사용으로 주석처리
// GetClusterInstanceSpec 클러스터 인스턴스 Spec 조회
func (h *ClusterManagerHandler) GetClusterInstanceSpec(ctx context.Context, req *cms.ClusterInstanceSpecRequest, rsp *cms.ClusterInstanceSpecResponse) error {
	logger.Debug("Received ClusterManager.GetClusterInstanceSpec request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterInstanceSpec] Could not get cluster instance spec. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_spec.failure-validate_request", err)
	}

	if rsp.Spec, err = spec.Get(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_spec.failure-get", err)
	}

	logger.Debug("Respond ClusterManager.GetClusterInstanceSpec request")
	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_instance_spec.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_instance_spec.success", nil)
}
*/

// GetClusterInstanceSpecByUUID uuid 를 통해 클러스터 인스턴스 Spec 조회
func (h *ClusterManagerHandler) GetClusterInstanceSpecByUUID(ctx context.Context, req *cms.ClusterInstanceSpecByUUIDRequest, rsp *cms.ClusterInstanceSpecResponse) error {
	logger.Debug("Received ClusterManager.GetClusterInstanceSpecByUUID request")
	defer logger.Debug("Respond ClusterManager.GetClusterInstanceSpecByUUID request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterInstanceSpecByUUID] Could not get cluster instance spec. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_spec_by_uuid.failure-validate_request", err)
	}

	if rsp.Spec, err = spec.GetByUUID(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_spec_by_uuid.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_instance_spec_by_uuid.success"}
	return nil
}

// CheckIsExistClusterInstanceSpec 인스턴스 스팩 이름을 통한 클러스터 인스턴스 스팩 존재유무 확인
func (h *ClusterManagerHandler) CheckIsExistClusterInstanceSpec(ctx context.Context, req *cms.CheckIsExistByNameRequest, rsp *cms.CheckIsExistResponse) error {
	logger.Debug("Received ClusterManager.CheckIsExistClusterInstanceSpec request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-CheckIsExistClusterInstanceSpec] Could not check cluster instance spec exist. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_instance_spec.failure-validate_request", err)
	}

	if rsp.IsExist, err = spec.CheckIsExistInstanceSpec(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_instance_spec.failure-get", err)
	}

	logger.Debug("Respond ClusterManager.CheckIsExistClusterInstanceSpec request")
	rsp.Message = &cms.Message{Code: "cdm-center.manager.check_is_exist_cluster_instance_spec.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.check_is_exist_cluster_instance_spec.success", nil)
}

// CreateClusterInstanceSpec 클러스터 인스턴스 Spec 생성
func (h *ClusterManagerHandler) CreateClusterInstanceSpec(ctx context.Context, req *cms.CreateClusterInstanceSpecRequest, rsp *cms.CreateClusterInstanceSpecResponse) (err error) {
	logger.Debug("Received ClusterManager.CreateClusterInstanceSpec request")
	defer logger.Debug("Respond ClusterManager.CreateClusterInstanceSpec request")

	rsp.Spec, err = spec.CreateInstanceSpec(req)
	if err != nil {
		logger.Errorf("[Handler-CreateClusterInstanceSpec] Could not create cluster instance spec. Cause: %+v", err)
		return createError(ctx, "cdm-center.manager.create_cluster_instance_spec.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_instance_spec.success"}
	return nil
}

// DeleteClusterInstanceSpec 클러스터 인스턴스 Spec 삭제
func (h *ClusterManagerHandler) DeleteClusterInstanceSpec(ctx context.Context, req *cms.DeleteClusterInstanceSpecRequest, rsp *cms.DeleteClusterInstanceSpecResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterInstanceSpec request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterInstanceSpec request")

	if err := spec.DeleteInstanceSpec(req); err != nil {
		logger.Errorf("[Handler-DeleteClusterInstanceSpec] Could not delete cluster instance spec. Cause: %+v", err)
		return createError(ctx, "cdm-center.manager.delete_cluster_instance_spec.failure-delete", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_instance_spec.success"}
	return nil
}

// GetClusterKeyPair 클러스터 KeyPair 조회
func (h *ClusterManagerHandler) GetClusterKeyPair(ctx context.Context, req *cms.ClusterKeyPairRequest, rsp *cms.ClusterKeyPairResponse) error {
	logger.Debug("Received ClusterManager.GetClusterKeyPair request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterKeyPair] Could not get cluster keypair. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_keypair.failure-validate_request", err)
	}

	if rsp.Keypair, err = keypair.Get(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_keypair.failure-get", err)
	}

	logger.Debug("Respond ClusterManager.GetClusterKeyPair request")
	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_keypair.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_keypair.success", nil)
}

// CheckIsExistClusterKeypair Keypair 이름을 통한 클러스터 Keypair 존재유무 확인
func (h *ClusterManagerHandler) CheckIsExistClusterKeypair(ctx context.Context, req *cms.CheckIsExistByNameRequest, rsp *cms.CheckIsExistResponse) error {
	logger.Debug("Received ClusterManager.CheckIsExistClusterKeypair request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-CheckIsExistClusterKeypair] Could not check cluster keypair exist. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_keypair.failure-validate_request", err)
	}

	if rsp.IsExist, err = keypair.CheckIsExistKeypair(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_keypair.failure-get", err)
	}

	logger.Debug("Respond ClusterManager.CheckIsExistClusterKeypair request")
	rsp.Message = &cms.Message{Code: "cdm-center.manager.check_is_exist_cluster_keypair.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.check_is_exist_cluster_keypair.success", nil)
}

// CreateClusterKeypair 클러스터 Keypair 생성
func (h *ClusterManagerHandler) CreateClusterKeypair(ctx context.Context, req *cms.CreateClusterKeypairRequest, rsp *cms.CreateClusterKeypairResponse) (err error) {
	logger.Debug("Received ClusterManager.CreateClusterKeypair request")
	defer logger.Debug("Respond ClusterManager.CreateClusterKeypair request")

	rsp.Keypair, err = keypair.CreateKeypair(req)
	if err != nil {
		logger.Errorf("[Handler-CreateClusterKeypair] Could not create cluster keypair. Cause: %+v", err)
		return createError(ctx, "cdm-center.manager.create_cluster_keypair.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_keypair.success"}
	return nil
}

// DeleteClusterKeypair 클러스터 Keypair 삭제
func (h *ClusterManagerHandler) DeleteClusterKeypair(ctx context.Context, req *cms.DeleteClusterKeypairRequest, rsp *cms.DeleteClusterKeypairResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterKeypair request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterKeypair request")

	if err := keypair.DeleteKeypair(req); err != nil {
		logger.Errorf("[Handler-DeleteClusterKeypair] Could not delete cluster keypair. Cause: %+v", err)
		return createError(ctx, "cdm-center.manager.delete_cluster_keypair.failure-delete", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_keypair.success"}
	return nil
}

// GetClusterInstanceUserScript 클러스터 Instance User Script 조회
func (h *ClusterManagerHandler) GetClusterInstanceUserScript(ctx context.Context, req *cms.ClusterInstanceUserScriptRequest, rsp *cms.ClusterInstanceUserScriptResponse) (err error) {
	logger.Debug("Received ClusterManager.GetClusterInstanceUserScript request")
	defer logger.Debug("Respond ClusterManager.GetClusterInstanceUserScript request")

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterInstanceUserScript] Could not get cluster instance user script. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_user_script.failure-validate_request", err)
	}

	if rsp.UserData, err = instance.GetInstanceUserScript(req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_instance_user_script.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_instance_user_script.success"}
	return nil
}

// UpdateClusterInstanceUserScript 클러스터 Instance User Script 수정
func (h *ClusterManagerHandler) UpdateClusterInstanceUserScript(ctx context.Context, req *cms.UpdateClusterInstanceUserScriptRequest, rsp *cms.UpdateClusterInstanceUserScriptResponse) (err error) {
	logger.Debug("Received ClusterManager.UpdateClusterInstanceUserScript request")
	defer logger.Debug("Respond ClusterManager.UpdateClusterInstanceUserScript request")

	defer func() {
		if err != nil {
			logger.Errorf("[Handler-UpdateClusterInstanceUserScript] Could not update cluster instance user script. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.update_cluster_instance_user_script.failure-validate_request", err)
	}

	if err = instance.UpdateInstanceUserScript(req); err != nil {
		return createError(ctx, "cdm-center.manager.update_cluster_instance_user_script.failure-update", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.update_cluster_instance_user_script.success"}
	return nil
}

// CreateClusterSecurityGroup 클러스터 보안 그룹 생성
func (h *ClusterManagerHandler) CreateClusterSecurityGroup(ctx context.Context, req *cms.CreateClusterSecurityGroupRequest, rsp *cms.ClusterSecurityGroupResponse) error {
	logger.Debug("Received ClusterManager.CreateClusterSecurityGroup request")
	defer logger.Debug("Respond ClusterManager.CreateClusterSecurityGroup request")

	var err error
	rsp.SecurityGroup, err = securityGroup.CreateSecurityGroup(req)
	if err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-CreateClusterSecurityGroup] Could not create cluster security group. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-CreateClusterSecurityGroup] Could not create cluster security group. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.create_cluster_security_group.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_security_group.success"}
	return nil
}

// CreateClusterSecurityGroupRule 클러스터 보안 그룹 규칙 생성
func (h *ClusterManagerHandler) CreateClusterSecurityGroupRule(ctx context.Context, req *cms.CreateClusterSecurityGroupRuleRequest, rsp *cms.ClusterSecurityGroupRuleResponse) error {
	logger.Debug("Received ClusterManager.CreateClusterSecurityGroupRule request")
	defer logger.Debug("Respond ClusterManager.CreateClusterSecurityGroupRule request")

	var err error
	rsp.SecurityGroupRule, err = securityGroup.CreateSecurityGroupRule(req)
	if err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-CreateClusterSecurityGroupRule] Could not create cluster security group rule. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-CreateClusterSecurityGroupRule] Could not create cluster security group rule. Cause: %+v", err)
		}

		return createError(ctx, "cdm-center.manager.create_cluster_security_group_rule.failure-create", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.create_cluster_security_group_rule.success"}
	return nil
}

// DeleteClusterSecurityGroup 클러스터 보안 그룹 삭제
func (h *ClusterManagerHandler) DeleteClusterSecurityGroup(ctx context.Context, req *cms.DeleteClusterSecurityGroupRequest, rsp *cms.DeleteClusterSecurityGroupResponse) error {
	logger.Debug("Received ClusterManager.DeleteClusterSecurityGroup request")
	defer logger.Debug("Respond ClusterManager.DeleteClusterSecurityGroup request")

	if err := securityGroup.DeleteSecurityGroup(req); err != nil {
		if req.GetTenant() != nil {
			logger.Errorf("[Handler-DeleteClusterSecurityGroup] Could not delete cluster security group. tenant: %s Cause: %+v", req.GetTenant().GetUuid(), err)
		} else {
			logger.Errorf("[Handler-DeleteClusterSecurityGroup] Could not delete cluster security group. Cause: %+v", err)
		}
		return createError(ctx, "cdm-center.manager.delete_cluster_security_group.failure-delete", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.delete_cluster_security_group.success"}
	return nil
}

// GetClusterSecurityGroupList 클러스터 보안그룹 목록 조회
func (h *ClusterManagerHandler) GetClusterSecurityGroupList(ctx context.Context, req *cms.ClusterSecurityGroupListRequest, rsp *cms.ClusterSecurityGroupListResponse) error {
	logger.Debug("Received ClusterManager.GetClusterSecurityGroupList request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterSecurityGroupList] Could not get cluster security group list. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_security_group_list.failure-validate_request", err)
	}

	if rsp.SecurityGroups, rsp.Pagination, err = securityGroup.GetList(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_security_group_list.failure-get", err)
	}

	if len(rsp.SecurityGroups) == 0 {
		return createError(ctx, "cdm-center.manager.get_cluster_security_group_list.success-get", errors.ErrNoContent)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_security_group_list.success"}
	logger.Debug("Respond ClusterManager.GetClusterSecurityGroupList request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_security_group_list.success", nil)
}

// GetClusterSecurityGroup 클러스터 보안그룹 조회
func (h *ClusterManagerHandler) GetClusterSecurityGroup(ctx context.Context, req *cms.ClusterSecurityGroupRequest, rsp *cms.ClusterSecurityGroupResponse) error {
	logger.Debug("Received ClusterManager.GetClusterSecurityGroup request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterSecurityGroup] Could not get cluster security group. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_security_group.failure-validate_request", err)
	}

	if rsp.SecurityGroup, err = securityGroup.Get(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_security_group.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_security_group.success"}
	logger.Debug("Respond ClusterManager.GetClusterSecurityGroup request")
	return errors.StatusOK(ctx, "cdm-center.manager.get_cluster_security_group.success", nil)
}

/* 미사용으로 주석처리
// GetClusterSecurityGroupByUUID uuid 를 통해 클러스터 보안 그룹 조회
func (h *ClusterManagerHandler) GetClusterSecurityGroupByUUID(ctx context.Context, req *cms.ClusterSecurityGroupByUUIDRequest, rsp *cms.ClusterSecurityGroupResponse) error {
	logger.Debug("Received ClusterManager.GetClusterSecurityGroupByUUID request")
	defer logger.Debug("Respond ClusterManager.GetClusterSecurityGroupByUUID request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-GetClusterSecurityGroupByUUID] Could not get cluster security group. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_security_group_by_uuid.failure-validate_request", err)
	}

	if rsp.SecurityGroup, err = securityGroup.GetByUUID(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.get_cluster_security_group_by_uuid.failure-get", err)
	}

	rsp.Message = &cms.Message{Code: "cdm-center.manager.get_cluster_security_group_by_uuid.success"}
	return nil
}
*/

// CheckIsExistClusterSecurityGroup 보안그룹 이름을 통한 클러스터 보안그룹 존재유무 확인
func (h *ClusterManagerHandler) CheckIsExistClusterSecurityGroup(ctx context.Context, req *cms.CheckIsExistByNameRequest, rsp *cms.CheckIsExistResponse) error {
	logger.Debug("Received ClusterManager.CheckIsExistClusterSecurityGroup request")

	var err error
	defer func() {
		if err != nil {
			logger.Errorf("[Handler-CheckIsExistClusterSecurityGroup] Could not check cluster security group exist. Cause: %+v", err)
		}
	}()

	if err = validateRequest(ctx); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_security_group.failure-validate_request", err)
	}

	if rsp.IsExist, err = securityGroup.CheckIsExistSecurityGroup(ctx, req); err != nil {
		return createError(ctx, "cdm-center.manager.check_is_exist_cluster_security_group.failure-get", err)
	}

	logger.Debug("Respond ClusterManager.CheckIsExistClusterSecurityGroup request")
	rsp.Message = &cms.Message{Code: "cdm-center.manager.check_is_exist_cluster_security_group.success"}
	return errors.StatusOK(ctx, "cdm-center.manager.check_is_exist_cluster_security_group.success", nil)
}

// NewClusterManagerHandler Cluster Manager 서비스 RPC 핸들러 생성
func NewClusterManagerHandler() (*ClusterManagerHandler, error) {
	f := path.Join(defaultPrivateKeyPath, defaultPrivateKeyFile)
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New(fmt.Sprintf("invalid pem formatted file. file: %s", f))
	}

	pkey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	client.Key = "cdm"

	return &ClusterManagerHandler{
		privateKey:  pkey,
		updatedAt:   make(map[uint64]time.Time),
		syncStatus:  make(map[uint64]queue.SyncClusterStatus),
		checkStatus: make(map[uint64]queue.ClusterServiceStatus),
	}, nil
}
