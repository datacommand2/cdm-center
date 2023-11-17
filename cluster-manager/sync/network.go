package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	drConstant "github.com/datacommand2/cdm-disaster-recovery/common/constant"
	"github.com/datacommand2/cdm-disaster-recovery/common/migrator"
	"github.com/jinzhu/gorm"
	"reflect"
)

func compareNetwork(dbNetwork *model.ClusterNetwork, result *client.GetNetworkResult) bool {
	if dbNetwork.UUID != result.Network.UUID {
		return false
	}

	if dbNetwork.Name != result.Network.Name {
		return false
	}

	if dbNetwork.TypeCode != result.Network.TypeCode {
		return false
	}

	if dbNetwork.ExternalFlag != result.Network.ExternalFlag {
		return false
	}

	if dbNetwork.State != result.Network.State {
		return false
	}

	if dbNetwork.Status != result.Network.Status {
		return false
	}

	if !reflect.DeepEqual(dbNetwork.Description, result.Network.Description) {
		return false
	}

	var t model.ClusterTenant
	if err := database.Execute(func(db *gorm.DB) error {
		return db.Find(&t, &model.ClusterTenant{ID: dbNetwork.ClusterTenantID}).Error
	}); err != nil {
		return false
	}

	if t.UUID != result.Tenant.UUID {
		return false
	}

	return true
}

func (s *Synchronizer) getNetwork(network *model.ClusterNetwork, opts ...Option) (*model.ClusterNetwork, error) {
	var err error
	var n model.ClusterNetwork

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterNetwork{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_network.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterNetwork{UUID: network.UUID}).
			First(&n).Error
	}); err == nil {
		return &n, nil
	}

	err = s.SyncClusterNetwork(&model.ClusterNetwork{UUID: network.UUID}, opts...)
	if err != nil {
		return nil, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterNetwork{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_network.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterNetwork{UUID: network.UUID}).
			First(&n).Error
	}); err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return &n, nil
}

func (s *Synchronizer) deleteNetwork(network *model.ClusterNetwork) (err error) {
	logger.Infof("[Sync-deleteNetwork] Start: cluster(%d) network(%s)", s.clusterID, network.UUID)

	var instanceNetworkList []model.ClusterInstanceNetwork
	var floatingIPList []model.ClusterFloatingIP
	var subnetList []model.ClusterSubnet

	if err = database.Execute(func(db *gorm.DB) error {
		if err = db.Where(&model.ClusterInstanceNetwork{ClusterNetworkID: network.ID}).Find(&instanceNetworkList).Error; err != nil {
			return err
		}

		if err = db.Where(&model.ClusterFloatingIP{ClusterNetworkID: network.ID}).Find(&floatingIPList).Error; err != nil {
			return err
		}

		if err = db.Where(&model.ClusterSubnet{ClusterNetworkID: network.ID}).Find(&subnetList).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// delete instance network
	for _, instanceNetwork := range instanceNetworkList {
		if err = s.deleteInstanceNetwork(&instanceNetwork); err != nil {
			return err
		}
	}

	// delete floating ip
	for _, floatingIP := range floatingIPList {
		if err = s.deleteFloatingIP(&floatingIP); err != nil {
			return err
		}
	}

	// delete subnet
	for _, subnet := range subnetList {
		if err = s.deleteSubnet(&subnet); err != nil {
			return err
		}
	}

	// delete network
	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(network).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteNetwork] Could not delete cluster(%d) network(%s). Cause: %+v", s.clusterID, network.UUID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterNetworkDeleted, &queue.DeleteClusterNetwork{
		Cluster:     &model.Cluster{ID: s.clusterID},
		OrigNetwork: network,
	}); err != nil {
		logger.Warnf("[Sync-deleteNetwork] Could not publish cluster(%d) network(%s) deleted message. Cause: %+v", s.clusterID, network.UUID, err)
	}

	// 해당 network uuid 를 쓰고 있는 shared task 삭제
	migrator.ClearSharedTask(s.clusterID, network.ID, network.UUID, drConstant.MigrationTaskTypeCreateNetwork)

	logger.Infof("[Sync-deleteNetwork] Success: cluster(%d) network(%s)", s.clusterID, network.UUID)
	return nil
}

func (s *Synchronizer) syncDeletedNetworkList(rsp *client.GetNetworkListResponse) error {
	var err error
	var networkList []*model.ClusterNetwork

	// 동기화되어있는 네트워크 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterNetwork{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_network.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Find(&networkList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 네트워크 제거
	for _, network := range networkList {
		exists := false
		for _, clusterNetwork := range rsp.ResultList {
			exists = exists || (clusterNetwork.Network.UUID == network.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteNetwork(network); err != nil {
			return err
		}
	}

	return nil
}

// SyncClusterNetworkList 네트워크 목록 동기화
func (s *Synchronizer) SyncClusterNetworkList(opts ...Option) error {
	logger.Infof("[SyncClusterNetworkList] Start: cluster(%d)", s.clusterID)

	// 클러스터의 네트워크 목록 조회
	rsp, err := s.Cli.GetNetworkList()
	if err != nil {
		return err
	}

	// 클러스터에서 삭제된 네트워크 제거
	if err = s.syncDeletedNetworkList(rsp); err != nil {
		return err
	}

	// 네트워크 상세 동기화
	for _, network := range rsp.ResultList {
		if e := s.SyncClusterNetwork(&model.ClusterNetwork{UUID: network.Network.UUID}, opts...); e != nil {
			logger.Warnf("[SyncClusterNetworkList] Failed to sync cluster(%d) network(%s). Cause: %+v", s.clusterID, network.Network.UUID, e)
			err = e
		}
	}

	logger.Infof("[SyncClusterNetworkList] Completed: cluster(%d)", s.clusterID)
	return err
}

// SyncClusterNetwork 네트워크 상세 동기화
func (s *Synchronizer) SyncClusterNetwork(network *model.ClusterNetwork, opts ...Option) error {
	logger.Infof("[SyncClusterNetwork] Start: cluster(%d) network(%s)", s.clusterID, network.UUID)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.TenantUUID != "" {
		if ok, err := s.isDeletedTenant(options.TenantUUID); ok {
			// 삭제된 tenant
			logger.Infof("[SyncClusterNetwork] Already sync and deleted tenant: cluster(%d) network(%s) tenant(%s)",
				s.clusterID, network.UUID, options.TenantUUID)
			return nil
		} else if err != nil {
			logger.Errorf("[SyncClusterNetwork] Could not get tenant from db: cluster(%d) network(%s) tenant(%s). Cause: %+v",
				s.clusterID, network.UUID, options.TenantUUID, err)
			return err
		}
		// 삭제안된 tenant 라면 동기화 계속 진행
	}

	rsp, err := s.Cli.GetNetwork(client.GetNetworkRequest{Network: model.ClusterNetwork{UUID: network.UUID}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterNetwork{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_network.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterNetwork{UUID: network.UUID}).
			First(network).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		logger.Infof("[SyncClusterNetwork] Success - nothing was updated: cluster(%d) network(%s)", s.clusterID, network.UUID)
		return nil

	case rsp == nil: // 네트워크 삭제
		// 네트워크 삭제 시 해당 네트워크의 floating ip와 subnet, subnetNameserver, subnetDHCPPool 지움
		if err = s.deleteNetwork(network); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 네트워크 추가
		var tenant *model.ClusterTenant
		if tenant, err = s.getTenant(&model.ClusterTenant{UUID: rsp.Result.Tenant.UUID}, opts...); err != nil {
			return err
		} else if tenant == nil {
			logger.Warnf("[SyncClusterNetwork] Could not sync cluster(%d) network(%s). Cause: not found tenant(%s)", s.clusterID, network.UUID, rsp.Result.Tenant.UUID)
			return nil
		}

		rsp.Result.Network.ClusterTenantID = tenant.ID
		*network = rsp.Result.Network

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(network).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterNetwork] Done - added: cluster(%d) network(%s)", s.clusterID, network.UUID)

		if err = publishMessage(constant.QueueNoticeClusterNetworkCreated, &queue.CreateClusterNetwork{
			Cluster: &model.Cluster{ID: s.clusterID},
			Network: network,
		}); err != nil {
			logger.Warnf("[SyncClusterNetwork] Could not publish cluster(%d) network(%s) created message. Cause: %+v", s.clusterID, network.UUID, err)
		}

	case options.Force || !compareNetwork(network, &rsp.Result): // 네트워크 정보 수정
		var tenant *model.ClusterTenant
		if tenant, err = s.getTenant(&model.ClusterTenant{UUID: rsp.Result.Tenant.UUID}, opts...); err != nil {
			return err
		} else if tenant == nil {
			logger.Warnf("[SyncClusterNetwork] Could not sync cluster(%d) network(%s). Cause: not found tenant(%s)", s.clusterID, network.UUID, rsp.Result.Tenant.UUID)
			return nil
		}

		origNetwork := *network
		rsp.Result.Network.ID = network.ID
		rsp.Result.Network.ClusterTenantID = tenant.ID
		*network = rsp.Result.Network

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(network).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterNetwork] Done - updated: cluster(%d) network(%s)", s.clusterID, network.UUID)

		if err = publishMessage(constant.QueueNoticeClusterNetworkUpdated, &queue.UpdateClusterNetwork{
			Cluster:     &model.Cluster{ID: s.clusterID},
			OrigNetwork: &origNetwork,
			Network:     network,
		}); err != nil {
			logger.Warnf("[SyncClusterNetwork] Could not publish cluster(%d) network(%s) updated message. Cause: %+v", s.clusterID, network.UUID, err)
		}
	}

	// 신규, 수정인 경우 subnet, floating ip 동기화
	if rsp != nil {
		if err = s.syncClusterSubnetList(network, rsp.Result.SubnetList, opts...); err != nil {
			logger.Errorf("[SyncClusterNetwork] Failed to sync cluster(%d) network(%s) subnet list. Cause: %+v", s.clusterID, network.UUID, err)
		}

		if err = s.syncClusterFloatingIPList(network, rsp.Result.FloatingIPList, opts...); err != nil {
			logger.Errorf("[SyncClusterNetwork] Failed to sync cluster(%d) network(%s) floating IP list. Cause: %+v", s.clusterID, network.UUID, err)
		}

		if err != nil {
			logger.Infof("[SyncClusterNetwork] Cluster(%d) network(%s) synchronization completed.", s.clusterID, network.UUID)
			return err
		}
	}

	logger.Infof("[SyncClusterNetwork] Success: cluster(%d) network(%s)", s.clusterID, network.UUID)
	return nil
}
