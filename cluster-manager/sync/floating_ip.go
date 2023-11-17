package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
	"reflect"
)

func compareFloatingIP(dbFloatingIP *model.ClusterFloatingIP, result *client.GetFloatingIPResult) bool {
	if dbFloatingIP.UUID != result.FloatingIP.UUID {
		return false
	}

	if dbFloatingIP.IPAddress != result.FloatingIP.IPAddress {
		return false
	}

	if dbFloatingIP.Status != result.FloatingIP.Status {
		return false
	}

	if !reflect.DeepEqual(dbFloatingIP.Description, result.FloatingIP.Description) {
		return false
	}

	var network model.ClusterNetwork
	if err := database.Execute(func(db *gorm.DB) error {
		return db.Find(&network, &model.ClusterNetwork{ID: dbFloatingIP.ClusterNetworkID}).Error
	}); err != nil {
		return false
	}

	if network.UUID != result.Network.UUID {
		return false
	}

	return true
}

func (s *Synchronizer) getFloatingIP(floatingIP *model.ClusterFloatingIP, opts ...Option) (*model.ClusterFloatingIP, error) {
	var err error
	var f model.ClusterFloatingIP
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterFloatingIP{}.TableName()).
			Joins("join cdm_cluster_network on cdm_cluster_floating_ip.cluster_network_id = cdm_cluster_network.id").
			Joins("join cdm_cluster_tenant on cdm_cluster_floating_ip.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterFloatingIP{UUID: floatingIP.UUID}).
			First(&f).Error
	}); err == nil {
		return &f, nil
	}

	if err = s.SyncClusterFloatingIP(&model.ClusterFloatingIP{UUID: floatingIP.UUID}, opts...); err != nil {
		return nil, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterFloatingIP{}.TableName()).
			Joins("join cdm_cluster_network on cdm_cluster_floating_ip.cluster_network_id = cdm_cluster_network.id").
			Joins("join cdm_cluster_tenant on cdm_cluster_floating_ip.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterFloatingIP{UUID: floatingIP.UUID}).
			First(&f).Error
	}); err != nil {
		return nil, err
	}

	return &f, nil
}

func (s *Synchronizer) deleteFloatingIP(floatingIP *model.ClusterFloatingIP) (err error) {
	logger.Infof("[Sync-deleteFloatingIP] Start: cluster(%d) floatingIP(%s)", s.clusterID, floatingIP.UUID)

	// update instance network's floating ip
	var instanceNetworkList []model.ClusterInstanceNetwork
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterInstanceNetwork{ClusterFloatingIPID: &floatingIP.ID}).Find(&instanceNetworkList).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteFloatingIP] Could not find cluster(%d) floatingIP(%s) in instance network. Cause: %+v",
			s.clusterID, floatingIP.UUID, err)
		return errors.UnusableDatabase(err)
	}

	for _, instanceNetwork := range instanceNetworkList {
		if err = s.deleteInstanceNetworkFloatingIP(&instanceNetwork); err != nil {
			return err
		}
	}

	// delete floating ip
	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(floatingIP).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteFloatingIP] Could not delete cluster(%d) floatingIP(%s) in db. Cause: %+v",
			s.clusterID, floatingIP.UUID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterFloatingIPDeleted, &queue.DeleteClusterFloatingIP{
		Cluster:        &model.Cluster{ID: s.clusterID},
		OrigFloatingIP: floatingIP,
	}); err != nil {
		logger.Warnf("[Sync-deleteFloatingIP] Could not publish cluster(%d) floatingIP(%s) deleted message. Cause: %+v",
			s.clusterID, floatingIP.UUID, err)
	}

	logger.Infof("[Sync-deleteFloatingIP] Success: cluster(%d) floatingIP(%s)", s.clusterID, floatingIP.UUID)
	return nil
}

func (s *Synchronizer) syncDeletedFloatingIPList(network *model.ClusterNetwork, rsp []model.ClusterFloatingIP) (err error) {
	var floatingIPList []*model.ClusterFloatingIP

	// 동기화되어있는 floating ip 목록 조회
	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Where(&model.ClusterFloatingIP{ClusterNetworkID: network.ID}).Find(&floatingIPList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 floating ip 제거
	for _, floatingIP := range floatingIPList {
		exists := false
		for _, clusterFloatingIP := range rsp {
			exists = exists || (clusterFloatingIP.UUID == floatingIP.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteFloatingIP(floatingIP); err != nil {
			return err
		}

	}

	return nil
}

// syncClusterFloatingIPList floating ip 목록 동기화
func (s *Synchronizer) syncClusterFloatingIPList(network *model.ClusterNetwork, floatingIPList []model.ClusterFloatingIP, opts ...Option) (err error) {
	logger.Infof("[syncClusterFloatingIPList] Start: cluster(%d)", s.clusterID)

	// 클러스터에서 삭제된 floating ip 제거
	if err = s.syncDeletedFloatingIPList(network, floatingIPList); err != nil {
		return err
	}

	// floating ip 상세 동기화
	for _, floatingIP := range floatingIPList {
		if e := s.SyncClusterFloatingIP(&model.ClusterFloatingIP{UUID: floatingIP.UUID}, opts...); e != nil {
			logger.Warnf("[syncClusterFloatingIPList] Failed to sync cluster(%d) floatingIP(%s). Cause: %+v", s.clusterID, floatingIP.UUID, e)
			err = e
		}
	}

	logger.Infof("[syncClusterFloatingIPList] Completed: cluster(%d)", s.clusterID)
	return err
}

// SyncClusterFloatingIP floating ip 상세 동기화
func (s *Synchronizer) SyncClusterFloatingIP(floatingIP *model.ClusterFloatingIP, opts ...Option) error {
	logger.Infof("[SyncClusterFloatingIP] Start: cluster(%d) floatingIP(%s)", s.clusterID, floatingIP.UUID)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.TenantUUID != "" {
		if ok, err := s.isDeletedTenant(options.TenantUUID); ok {
			// 삭제된 tenant
			logger.Infof("[SyncClusterFloatingIP] Already sync and deleted tenant: cluster(%d) floatingIP(%s) tenant(%s)",
				s.clusterID, floatingIP.UUID, options.TenantUUID)
			return nil
		} else if err != nil {
			logger.Errorf("[SyncClusterFloatingIP] Could not get tenant from db: cluster(%d) floatingIP(%s) tenant(%s). Cause: %+v",
				s.clusterID, floatingIP.UUID, options.TenantUUID, err)
			return err
		}
		// 삭제안된 tenant 라면 동기화 계속 진행
	}

	rsp, err := s.Cli.GetFloatingIP(client.GetFloatingIPRequest{FloatingIP: model.ClusterFloatingIP{UUID: floatingIP.UUID}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	var t *model.ClusterTenant
	var n *model.ClusterNetwork
	if rsp != nil {
		t, err = s.getTenant(&model.ClusterTenant{UUID: rsp.Result.Tenant.UUID}, opts...)
		if err != nil {
			logger.Errorf("[SyncClusterFloatingIP] Could not get tenant from db: cluster(%d) floatingIP(%s) tenant(%s). Cause: %+v",
				s.clusterID, floatingIP.UUID, rsp.Result.Tenant.UUID, err)
			return err
		}

		n, err = s.getNetwork(&model.ClusterNetwork{UUID: rsp.Result.Network.UUID}, opts...)
		if err != nil {
			logger.Errorf("[SyncClusterFloatingIP] Could not get network from db: cluster(%d) floatingIP(%s) network(%s). Cause: %+v",
				s.clusterID, floatingIP.UUID, rsp.Result.Network.UUID, err)
			return err
		}

		if t != nil && n != nil {
			rsp.Result.FloatingIP.ClusterTenantID = t.ID
			rsp.Result.FloatingIP.ClusterNetworkID = n.ID
		} else {
			logger.Warnf("[SyncClusterFloatingIP] Could not publish cluster(%d) floatingIP(%s) created message. Cause: tenant(%s) or network(%s) is nil",
				s.clusterID, floatingIP.UUID, rsp.Result.Tenant.UUID, rsp.Result.Network.UUID)
			return errors.Unknown(errors.New("tenant or network is nil"))
		}

	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterFloatingIP{}.TableName()).
			Joins("join cdm_cluster_network on cdm_cluster_floating_ip.cluster_network_id = cdm_cluster_network.id").
			Joins("join cdm_cluster_tenant on cdm_cluster_floating_ip.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterFloatingIP{UUID: floatingIP.UUID}).
			First(floatingIP).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		logger.Errorf("[SyncClusterFloatingIP] Could not find floatingIP in db: cluster(%d) floatingIP(%s). Cause: %+v",
			s.clusterID, floatingIP.UUID, err)
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		logger.Infof("[SyncClusterFloatingIP] Success - nothing was updated: cluster(%d) floatingIP(%s)", s.clusterID, floatingIP.UUID)
		return nil

	case rsp == nil: // 삭제
		if err = s.deleteFloatingIP(floatingIP); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 신규 생성
		*floatingIP = rsp.Result.FloatingIP

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(floatingIP).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterFloatingIP] Done - added: cluster(%d) floatingIP(%s)", s.clusterID, floatingIP.UUID)

		if err = publishMessage(constant.QueueNoticeClusterFloatingIPCreated, &queue.CreateClusterFloatingIP{
			Cluster:    &model.Cluster{ID: s.clusterID},
			FloatingIP: floatingIP,
		}); err != nil {
			logger.Warnf("[SyncClusterFloatingIP] Could not publish cluster(%d) floatingIP(%s) created message. Cause: %+v", s.clusterID, floatingIP.UUID, err)
		}

	case options.Force || !compareFloatingIP(floatingIP, &rsp.Result): // 업데이트
		origFloatingIP := *floatingIP
		rsp.Result.FloatingIP.ID = floatingIP.ID
		*floatingIP = rsp.Result.FloatingIP

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(floatingIP).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterFloatingIP] Done - updated: cluster(%d) floatingIP(%s)", s.clusterID, floatingIP.UUID)

		if err = publishMessage(constant.QueueNoticeClusterFloatingIPUpdated, &queue.UpdateClusterFloatingIP{
			Cluster:        &model.Cluster{ID: s.clusterID},
			OrigFloatingIP: &origFloatingIP,
			FloatingIP:     floatingIP,
		}); err != nil {
			logger.Warnf("[SyncClusterFloatingIP] Could not publish cluster(%d) floatingIP(%s) updated message. Cause: %+v", s.clusterID, floatingIP.UUID, err)
		}
	}

	logger.Infof("[SyncClusterFloatingIP] Success: cluster(%d) floatingIP(%s)", s.clusterID, floatingIP.UUID)
	return nil
}
