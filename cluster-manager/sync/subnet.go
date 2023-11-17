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

func compareSubnet(dbSubnet *model.ClusterSubnet, result *client.GetSubnetResult) bool {
	if dbSubnet.UUID != result.Subnet.UUID {
		return false
	}

	if dbSubnet.Name != result.Subnet.Name {
		return false
	}

	if dbSubnet.NetworkCidr != result.Subnet.NetworkCidr {
		return false
	}

	if dbSubnet.DHCPEnabled != result.Subnet.DHCPEnabled {
		return false
	}

	if dbSubnet.GatewayEnabled != result.Subnet.GatewayEnabled {
		return false
	}

	if !reflect.DeepEqual(dbSubnet.Description, result.Subnet.Description) {
		return false
	}

	if !reflect.DeepEqual(dbSubnet.GatewayIPAddress, result.Subnet.GatewayIPAddress) {
		return false
	}

	if !reflect.DeepEqual(dbSubnet.Ipv6AddressModeCode, result.Subnet.Ipv6AddressModeCode) {
		return false
	}

	if !reflect.DeepEqual(dbSubnet.Ipv6RaModeCode, result.Subnet.Ipv6RaModeCode) {
		return false
	}

	var network model.ClusterNetwork
	if err := database.Execute(func(db *gorm.DB) error {
		return db.Find(&network, &model.ClusterNetwork{ID: dbSubnet.ClusterNetworkID}).Error
	}); err != nil {
		return false
	}

	if network.UUID != result.Network.UUID {
		return false
	}

	return true
}

func (s *Synchronizer) getSubnet(subnet *model.ClusterSubnet, opts ...Option) (*model.ClusterSubnet, error) {
	var err error
	var result model.ClusterSubnet

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterSubnet{}.TableName()).
			Joins("join cdm_cluster_network on cdm_cluster_subnet.cluster_network_id = cdm_cluster_network.id").
			Joins("join cdm_cluster_tenant on cdm_cluster_network.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterSubnet{UUID: subnet.UUID}).First(&result).Error
	}); err == nil {
		return &result, nil
	}

	if err = s.SyncClusterSubnet(&model.ClusterSubnet{UUID: subnet.UUID}, opts...); err != nil {
		return nil, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterSubnet{}.TableName()).
			Joins("join cdm_cluster_network on cdm_cluster_subnet.cluster_network_id = cdm_cluster_network.id").
			Joins("join cdm_cluster_tenant on cdm_cluster_network.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterSubnet{UUID: subnet.UUID}).First(&result).Error
	}); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *Synchronizer) deleteSubnet(subnet *model.ClusterSubnet) (err error) {
	logger.Infof("[Sync-deleteSubnet] Start: cluster(%d) subnet(%s)", s.clusterID, subnet.UUID)

	var instanceNetworkList []model.ClusterInstanceNetwork
	var routingInterfaceList []model.ClusterNetworkRoutingInterface
	if err = database.Execute(func(db *gorm.DB) error {
		if err = db.Where(&model.ClusterInstanceNetwork{ClusterSubnetID: subnet.ID}).Find(&instanceNetworkList).Error; err != nil {
			return err
		}

		if err = db.Where(&model.ClusterNetworkRoutingInterface{ClusterSubnetID: subnet.ID}).Find(&routingInterfaceList).Error; err != nil {
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

	// delete routing interface
	for _, routingInterface := range routingInterfaceList {
		if err = s.deleteRoutingInterface(&routingInterface); err != nil {
			return err
		}
	}

	if err = database.GormTransaction(func(db *gorm.DB) error {
		// delete subnet DHCP pool
		if err = db.Where(&model.ClusterSubnetDHCPPool{ClusterSubnetID: subnet.ID}).Delete(&model.ClusterSubnetDHCPPool{}).Error; err != nil {
			logger.Errorf("[Sync-deleteSubnet] Could not delete cluster(%d) subnet(%s) dhcp pool list. Cause: %+v",
				s.clusterID, subnet.UUID, err)
			return err
		}

		// delete subnet name server
		if err = db.Where(&model.ClusterSubnetNameserver{ClusterSubnetID: subnet.ID}).Delete(&model.ClusterSubnetNameserver{}).Error; err != nil {
			logger.Errorf("[Sync-deleteSubnet] Could not delete cluster(%d) subnet(%s) nameserver list. Cause: %+v",
				s.clusterID, subnet.UUID, err)
			return err
		}

		// delete subnet
		if err = db.Delete(subnet).Error; err != nil {
			logger.Errorf("[Sync-deleteSubnet] Could not delete cluster(%d) subnet(%s). Cause: %+v", s.clusterID, subnet.UUID, err)
			return err
		}

		return nil
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterSubnetDeleted, &queue.DeleteClusterSubnet{
		Cluster:    &model.Cluster{ID: s.clusterID},
		OrigSubnet: subnet,
	}); err != nil {
		logger.Warnf("[Sync-deleteSubnet] Could not publish cluster(%d) subnet(%s) deleted message. Cause: %+v",
			s.clusterID, subnet.UUID, err)
	}

	// 해당 subnet 을 쓰고 있는 shared task 삭제
	migrator.ClearSharedTask(s.clusterID, subnet.ID, subnet.UUID, drConstant.MigrationTaskTypeCreateSubnet)

	logger.Infof("[Sync-deleteSubnet] Success: cluster(%d) subnet(%s)", s.clusterID, subnet.UUID)
	return nil
}

func (s *Synchronizer) syncDeletedSubnetList(network *model.ClusterNetwork, rsp []model.ClusterSubnet) error {
	var err error
	var subnetList []*model.ClusterSubnet

	// 동기화되어있는 서브넷 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterSubnet{ClusterNetworkID: network.ID}).Find(&subnetList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 서브넷 제거
	for _, subnet := range subnetList {
		exists := false
		for _, clusterSubnet := range rsp {
			exists = exists || (clusterSubnet.UUID == subnet.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteSubnet(subnet); err != nil {
			return err
		}
	}

	return nil
}

// syncClusterSubnetList subnet 목록 동기화
func (s *Synchronizer) syncClusterSubnetList(network *model.ClusterNetwork, subnetList []model.ClusterSubnet, opts ...Option) (err error) {
	logger.Infof("[syncClusterSubnetList] Start: cluster(%d)", s.clusterID)

	// 클러스터에서 삭제된 서브넷 제거
	if err = s.syncDeletedSubnetList(network, subnetList); err != nil {
		return err
	}

	// 서브넷 상세 동기화
	for _, subnet := range subnetList {
		if e := s.SyncClusterSubnet(&model.ClusterSubnet{UUID: subnet.UUID}, opts...); e != nil {
			logger.Warnf("[syncClusterSubnetList] Failed to sync cluster(%d) subnet(%s). Cause: %+v", s.clusterID, subnet.UUID, e)
			err = e
		}
	}

	logger.Infof("[syncClusterSubnetList] Completed: cluster(%d)", s.clusterID)
	return err
}

// syncClusterSubnetDHCPPoolList 서브넷 dhcp pool 목록 동기화
func (s *Synchronizer) syncClusterSubnetDHCPPoolList(subnet *model.ClusterSubnet, pools []model.ClusterSubnetDHCPPool) (err error) {
	// 특정 서브넷의 dhcp pool 전체 삭제
	if err = database.GormTransaction(func(db *gorm.DB) error {
		if err = db.Where(&model.ClusterSubnetDHCPPool{ClusterSubnetID: subnet.ID}).Delete(&model.ClusterSubnetDHCPPool{}).Error; err != nil {
			logger.Errorf("[syncClusterSubnetDHCPPoolList] Could not sync cluster(%d) subnet(%s) dhcp pool. Cause: %+v",
				s.clusterID, subnet.UUID, err)
			return errors.UnusableDatabase(err)
		}

		// subnet nameserver 상세 동기화
		for _, p := range pools {
			p.ClusterSubnetID = subnet.ID
			if err = db.Save(&p).Error; err != nil {
				logger.Errorf("[syncClusterSubnetDHCPPoolList] Could not sync cluster(%d) subnet(%s) dhcp pool. Cause: %+v",
					s.clusterID, subnet.UUID, err)
				return errors.UnusableDatabase(err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// syncClusterSubnetNameserverList 서브넷 nameserver 목록 동기화
func (s *Synchronizer) syncClusterSubnetNameserverList(subnet *model.ClusterSubnet, nameservers []model.ClusterSubnetNameserver) (err error) {
	// 특정 서브넷의 nameserver 전체 삭제
	if err = database.GormTransaction(func(db *gorm.DB) error {
		if err = db.Where(&model.ClusterSubnetNameserver{ClusterSubnetID: subnet.ID}).Delete(&model.ClusterSubnetNameserver{}).Error; err != nil {
			logger.Errorf("[syncClusterSubnetNameserverList] Could not sync cluster(%d) subnet(%s) nameserver. Cause: %+v",
				s.clusterID, subnet.UUID, err)
			return errors.UnusableDatabase(err)
		}

		// subnet nameserver 상세 동기화
		for _, n := range nameservers {
			n.ClusterSubnetID = subnet.ID
			if err = db.Save(&n).Error; err != nil {
				logger.Errorf("[syncClusterSubnetNameserverList] Could not sync cluster(%d) subnet(%s) name server. Cause: %+v",
					s.clusterID, subnet.UUID, err)
				return errors.UnusableDatabase(err)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// SyncClusterSubnet subnet 상세 동기화
func (s *Synchronizer) SyncClusterSubnet(subnet *model.ClusterSubnet, opts ...Option) error {
	logger.Infof("[SyncClusterSubnet] Start: cluster(%d) subnet(%s)", s.clusterID, subnet.UUID)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.TenantUUID != "" {
		if ok, err := s.isDeletedTenant(options.TenantUUID); ok {
			// 삭제된 tenant
			logger.Infof("[SyncClusterSubnet] Already sync and deleted tenant: cluster(%d) subnet(%s) tenant(%s)",
				s.clusterID, subnet.UUID, options.TenantUUID)
			return nil
		} else if err != nil {
			logger.Errorf("[SyncClusterSubnet] Could not get tenant from db: cluster(%d) subnet(%s) tenant(%s). Cause: %+v",
				s.clusterID, subnet.UUID, options.TenantUUID, err)
			return err
		}
		// 삭제안된 tenant 라면 동기화 계속 진행
	}

	rsp, err := s.Cli.GetSubnet(client.GetSubnetRequest{Subnet: model.ClusterSubnet{UUID: subnet.UUID}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	if rsp != nil {
		n, err := s.getNetwork(&model.ClusterNetwork{UUID: rsp.Result.Network.UUID}, opts...)
		if err != nil {
			return err
		}

		rsp.Result.Subnet.ClusterNetworkID = n.ID
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterSubnet{}.TableName()).
			Joins("join cdm_cluster_network on cdm_cluster_subnet.cluster_network_id = cdm_cluster_network.id").
			Joins("join cdm_cluster_tenant on cdm_cluster_network.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterSubnet{UUID: subnet.UUID}).
			First(subnet).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		logger.Infof("[SyncClusterSubnet] Success - nothing was updated: cluster(%d) subnet(%s)", s.clusterID, subnet.UUID)
		return nil

	case rsp == nil: // 삭제
		// 서브넷 삭제시 서브넷과 관련된 name server 와 pool list, router interface 도 삭제
		if err = s.deleteSubnet(subnet); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 신규 생성
		*subnet = rsp.Result.Subnet

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(subnet).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterSubnet] Done - added: cluster(%d) subnet(%s)", s.clusterID, subnet.UUID)

		if err = publishMessage(constant.QueueNoticeClusterSubnetCreated, &queue.CreateClusterSubnet{
			Cluster: &model.Cluster{ID: s.clusterID},
			Subnet:  subnet,
		}); err != nil {
			logger.Warnf("[SyncClusterSubnet] Could not publish cluster(%d) subnet(%s) created message. Cause: %+v",
				s.clusterID, subnet.UUID, err)
		}

	case options.Force || !compareSubnet(subnet, &rsp.Result): // 업데이트
		rsp.Result.Subnet.ID = subnet.ID
		*subnet = rsp.Result.Subnet

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(subnet).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterSubnet] Done - updated: cluster(%d) subnet(%s)", s.clusterID, subnet.UUID)

		if err = publishMessage(constant.QueueNoticeClusterSubnetUpdated, &queue.UpdateClusterSubnet{
			Cluster: &model.Cluster{ID: s.clusterID},
			Subnet:  subnet,
		}); err != nil {
			logger.Warnf("[SyncClusterSubnet] Could not publish cluster(%d) subnet(%s) updated message. Cause: %+v",
				s.clusterID, subnet.UUID, err)
		}
	}

	// 신규, 수정인 경우 nameServer, dhcp pool list 동기화
	if rsp != nil {
		if err = s.syncClusterSubnetNameserverList(subnet, rsp.Result.NameserverList); err != nil {
			logger.Errorf("[SyncClusterSubnet] Failed to sync cluster(%d) subnet(%s) name server list. Cause: %+v",
				s.clusterID, subnet.UUID, err)
		}

		if err = s.syncClusterSubnetDHCPPoolList(subnet, rsp.Result.PoolList); err != nil {
			logger.Errorf("[SyncClusterSubnet] Failed to sync cluster(%d) subnet(%s) DHCP pool list. Cause: %+v",
				s.clusterID, subnet.UUID, err)
		}

		if err != nil {
			logger.Infof("[SyncClusterSubnet] Cluster(%d) subnet(%s) synchronization completed.", s.clusterID, subnet.UUID)
			return err
		}
	}

	logger.Infof("[SyncClusterSubnet] Success: cluster(%d) subnet(%s)", s.clusterID, subnet.UUID)
	return nil
}
