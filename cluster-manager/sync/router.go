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

func compareRouter(dbRouter *model.ClusterRouter, result *client.GetRouterResult) bool {
	if dbRouter.UUID != result.Router.UUID {
		return false
	}

	if dbRouter.Name != result.Router.Name {
		return false
	}

	if dbRouter.State != result.Router.State {
		return false
	}

	if dbRouter.Status != result.Router.Status {
		return false
	}

	if !reflect.DeepEqual(dbRouter.Description, result.Router.Description) {
		return false
	}

	var t model.ClusterTenant
	if err := database.Execute(func(db *gorm.DB) error {
		return db.Find(&t, &model.ClusterTenant{ID: dbRouter.ClusterTenantID}).Error
	}); err != nil {
		return false
	}

	if t.UUID != result.Tenant.UUID {
		return false
	}

	return true
}

func (s *Synchronizer) deleteRouter(router *model.ClusterRouter) (err error) {
	logger.Infof("[Sync-deleteRouter] Start: cluster(%d) router(%s)", s.clusterID, router.UUID)

	var routingInterfaceList []model.ClusterNetworkRoutingInterface
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterNetworkRoutingInterface{ClusterRouterID: router.ID}).Find(&routingInterfaceList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// delete routing interface
	for _, routingInterface := range routingInterfaceList {
		if err = s.deleteRoutingInterface(&routingInterface); err != nil {
			return err
		}
	}

	if err = database.GormTransaction(func(db *gorm.DB) error {
		// delete router extra route
		err = db.Where(&model.ClusterRouterExtraRoute{ClusterRouterID: router.ID}).Delete(&model.ClusterRouterExtraRoute{}).Error
		if err != nil {
			logger.Errorf("[Sync-deleteRouter] Could not delete cluster(%d) router(%s)'s extraRoutes. Cause: %+v", s.clusterID, router.UUID, err)
			return err
		}

		// delete router
		if err = db.Delete(router).Error; err != nil {
			logger.Errorf("[Sync-deleteRouter] Could not delete cluster(%d) router(%s). Cause: %+v", s.clusterID, router.UUID, err)
			return err
		}

		return nil
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterRouterDeleted, &queue.DeleteClusterRouter{
		Cluster:    &model.Cluster{ID: s.clusterID},
		OrigRouter: router,
	}); err != nil {
		logger.Warnf("[Sync-deleteRouter] Could not publish cluster(%d) router(%s) deleted message. Cause: %+v",
			s.clusterID, router.UUID, err)
	}

	// 해당 router 를 쓰고 있는 shared task 삭제
	migrator.ClearSharedTask(s.clusterID, router.ID, router.UUID, drConstant.MigrationTaskTypeCreateRouter)

	logger.Infof("[Sync-deleteRouter] Success: cluster(%d) router(%s)", s.clusterID, router.UUID)
	return nil
}

func (s *Synchronizer) syncDeletedRouterList(rsp *client.GetRouterListResponse) (err error) {
	var routerList []*model.ClusterRouter

	// 동기화되어있는 라우터 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterRouter{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_router.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Find(&routerList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 라우터 제거
	for _, router := range routerList {
		exists := false
		for _, clusterRouter := range rsp.ResultList {
			exists = exists || (clusterRouter.Router.UUID == router.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteRouter(router); err != nil {
			return err
		}
	}

	return nil
}

// SyncClusterRouterList 라우터 목록 동기화
func (s *Synchronizer) SyncClusterRouterList(opts ...Option) error {
	logger.Infof("[SyncClusterRouterList] Start: cluster(%d)", s.clusterID)

	// 클러스터의 라우터 목록 조회
	rsp, err := s.Cli.GetRouterList()
	if err != nil {
		return err
	}

	// 클러스터에서 삭제된 라우터 제거
	if err = s.syncDeletedRouterList(rsp); err != nil {
		return err
	}

	// 라우터 상세 동기화
	for _, router := range rsp.ResultList {
		if e := s.SyncClusterRouter(&model.ClusterRouter{UUID: router.Router.UUID}, opts...); e != nil {
			logger.Warnf("[SyncClusterRouterList] Failed to sync cluster(%d) router(%s). Cause: %+v",
				s.clusterID, router.Router.UUID, e)
			err = e
		}
	}

	logger.Infof("[SyncClusterRouterList] Completed: cluster(%d)", s.clusterID)
	return err
}

// SyncClusterRouter 라우터 상세 동기화
func (s *Synchronizer) SyncClusterRouter(router *model.ClusterRouter, opts ...Option) error {
	logger.Infof("[SyncClusterRouter] Start: cluster(%d) router(%s)", s.clusterID, router.UUID)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	if options.TenantUUID != "" {
		if ok, err := s.isDeletedTenant(options.TenantUUID); ok {
			// 삭제된 tenant
			logger.Infof("[SyncClusterRouter] Already sync and deleted tenant: cluster(%d) router(%s) tenant(%s)",
				s.clusterID, router.UUID, options.TenantUUID)
			return nil
		} else if err != nil {
			logger.Errorf("[SyncClusterRouter] Could not get tenant from db: cluster(%d) router(%s) tenant(%s). Cause: %+v",
				s.clusterID, router.UUID, options.TenantUUID, err)
			return err
		}
		// 삭제안된 tenant 라면 동기화 계속 진행
	}

	rsp, err := s.Cli.GetRouter(client.GetRouterRequest{Router: model.ClusterRouter{UUID: router.UUID}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterRouter{}.TableName()).
			Joins("join cdm_cluster_tenant on cdm_cluster_router.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Where(&model.ClusterRouter{UUID: router.UUID}).First(router).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		logger.Infof("[SyncClusterRouter] Success - nothing was updated: cluster(%d) router(%s)", s.clusterID, router.UUID)
		return nil

	case rsp == nil: // 라우터 삭제
		// 라우터 삭제 시 해당 라우터의 networkRoutingInterface 와 extraRoute 지움
		if err = s.deleteRouter(router); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 라우터 추가
		var tenant *model.ClusterTenant
		if tenant, err = s.getTenant(&model.ClusterTenant{UUID: rsp.Result.Tenant.UUID}, opts...); err != nil {
			return err
		} else if tenant == nil {
			logger.Warnf("[SyncClusterRouter] Could not sync cluster(%d) router(%s). Cause: not found tenant(%s)",
				s.clusterID, rsp.Result.Router.UUID, rsp.Result.Tenant.UUID)
			return nil
		}

		rsp.Result.Router.ClusterTenantID = tenant.ID
		*router = rsp.Result.Router

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(router).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterRouter] Done - added: cluster(%d) router(%s)", s.clusterID, router.UUID)

		if err = publishMessage(constant.QueueNoticeClusterRouterCreated, &queue.CreateClusterRouter{
			Cluster: &model.Cluster{ID: s.clusterID},
			Router:  router,
		}); err != nil {
			logger.Warnf("[SyncClusterRouter] Could not publish cluster(%d) router(%s) created message. Cause: %+v",
				s.clusterID, rsp.Result.Router.UUID, err)
		}

	case options.Force || !compareRouter(router, &rsp.Result): // 라우터 정보 변경
		var tenant *model.ClusterTenant
		if tenant, err = s.getTenant(&model.ClusterTenant{UUID: rsp.Result.Tenant.UUID}, opts...); err != nil {
			return err
		} else if tenant == nil {
			logger.Warnf("[SyncClusterRouter] Could not sync cluster(%d) router(%s). Cause: not found tenant(%s)",
				s.clusterID, rsp.Result.Router.UUID, rsp.Result.Tenant.UUID)
			return nil
		}

		origRouter := *router
		rsp.Result.Router.ID = router.ID
		rsp.Result.Router.ClusterTenantID = tenant.ID
		*router = rsp.Result.Router

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(router).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterRouter] Done - updated: cluster(%d) router(%s)", s.clusterID, router.UUID)

		if err = publishMessage(constant.QueueNoticeClusterRouterUpdated, &queue.UpdateClusterRouter{
			Cluster:    &model.Cluster{ID: s.clusterID},
			OrigRouter: &origRouter,
			Router:     router,
		}); err != nil {
			logger.Warnf("[SyncClusterRouter] Could not publish cluster(%d) router(%s) updated message. Cause: %+v",
				s.clusterID, router.UUID, err)
		}
	}

	// 신규, 수정인 경우 extraRoute, routing interface 동기화
	if rsp != nil {
		if err = s.syncClusterRouterExtraRouteList(router, rsp.Result.ExtraRoute); err != nil {
			logger.Errorf("[SyncClusterRouter] Failed to sync cluster(%d) router(%s) extra route list. Cause: %+v",
				s.clusterID, router.UUID, err)
		}

		if err = s.syncClusterNetworkRoutingInterfaceList(router, &rsp.Result); err != nil {
			logger.Errorf("[SyncClusterRouter] Failed to sync cluster(%d) router(%s) routing interface list. Cause: %+v",
				s.clusterID, router.UUID, err)
		}

		if err != nil {
			logger.Infof("[SyncClusterRouter] Cluster(%d) router(%s) synchronization completed.", s.clusterID, router.UUID)
			return err
		}
	}

	logger.Infof("[SyncClusterRouter] Success: cluster(%d) router(%s)", s.clusterID, router.UUID)
	return nil
}

// syncClusterRouterExtraRouteList router extraRoute 목록 동기화
func (s *Synchronizer) syncClusterRouterExtraRouteList(router *model.ClusterRouter, routes []model.ClusterRouterExtraRoute) (err error) {
	if err = database.GormTransaction(func(db *gorm.DB) error {
		// router extraRoute 상세 동기화
		for _, r := range routes {
			r.ClusterRouterID = router.ID
			if err = db.Save(&r).Error; err != nil {
				logger.Errorf("[syncClusterRouterExtraRouteList] Could not sync cluster(%d) router(%s) extra route. Cause: %+v", s.clusterID, router.UUID, err)
				return errors.UnusableDatabase(err)
			}
		}

		// 라우터의 extraRoute 전체 삭제
		if err = db.Where(&model.ClusterRouterExtraRoute{ClusterRouterID: router.ID}).Delete(&model.ClusterRouterExtraRoute{}).Error; err != nil {
			logger.Errorf("[syncClusterRouterExtraRouteList] Could not sync cluster(%d) router(%s) extra route. Cause: %+v", s.clusterID, router.UUID, err)
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func compareRoutingInterface(dbRoutingInterface model.ClusterNetworkRoutingInterface, result client.RoutingInterfaceResult) bool {
	if dbRoutingInterface.IPAddress != result.RoutingInterface.IPAddress {
		return false
	}

	if dbRoutingInterface.ExternalFlag != result.RoutingInterface.ExternalFlag {
		return false
	}

	return true
}

func (s *Synchronizer) deleteRoutingInterface(routingInterface *model.ClusterNetworkRoutingInterface) (err error) {
	logger.Infof("[Sync-deleteRoutingInterface] Start: cluster(%d) router(%d) subnet(%d)",
		s.clusterID, routingInterface.ClusterRouterID, routingInterface.ClusterSubnetID)

	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(routingInterface).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteRoutingInterface] Could not delete routing interfaces: cluster(%d) router(%d) subnet(%d). Cause: %+v",
			s.clusterID, routingInterface.ClusterRouterID, routingInterface.ClusterSubnetID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterRoutingInterfaceDeleted, &queue.DeleteClusterRoutingInterface{
		Cluster:              &model.Cluster{ID: s.clusterID},
		OrigRoutingInterface: routingInterface,
	}); err != nil {
		logger.Warnf("[Sync-deleteRoutingInterface] Could not publish cluster routing interface deleted message: cluster(%d) router(%d) subnet(%d). Cause: %+v",
			s.clusterID, routingInterface.ClusterRouterID, routingInterface.ClusterSubnetID, err)
	}

	logger.Infof("[Sync-deleteRoutingInterface] Success: cluster(%d) router(%d) subnet(%d)",
		s.clusterID, routingInterface.ClusterRouterID, routingInterface.ClusterSubnetID)
	return nil
}

func (s *Synchronizer) syncDeletedRoutingInterfaceList(routerID uint64, resultList []client.RoutingInterfaceResult) error {
	var err error
	var records []struct {
		RoutingInterface *model.ClusterNetworkRoutingInterface `gorm:"embedded"`
		Subnet           *model.ClusterSubnet                  `gorm:"embedded"`
	}

	// 동기화되어있는 라우팅 인터페이스 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterNetworkRoutingInterface{}.TableName()).
			Select("*").
			Joins("join cdm_cluster_subnet on cdm_cluster_routing_interface.cluster_subnet_id = cdm_cluster_subnet.id").
			Joins("join cdm_cluster_router on cdm_cluster_routing_interface.cluster_router_id = cdm_cluster_router.id").
			Joins("join cdm_cluster_tenant on cdm_cluster_router.cluster_tenant_id = cdm_cluster_tenant.id").
			Where(&model.ClusterRouter{ID: routerID}).
			Where(&model.ClusterTenant{ClusterID: s.clusterID}).
			Find(&records).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	for _, record := range records {
		exists := false
		for _, r := range resultList {
			exists = exists || (r.Subnet.UUID == record.Subnet.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteRoutingInterface(record.RoutingInterface); err != nil {
			return err
		}
	}

	return nil
}

// syncClusterNetworkRoutingInterfaceList network routing interface 목록 동기화
func (s *Synchronizer) syncClusterNetworkRoutingInterfaceList(router *model.ClusterRouter, result *client.GetRouterResult) (err error) {
	resultList := append(result.InternalRoutingInterfaceList, result.ExternalRoutingInterfaceList...)

	// 클러스터에서 삭제된  network routing interface 제거
	if err = s.syncDeletedRoutingInterfaceList(router.ID, resultList); err != nil {
		return err
	}

	// routing interface 상세 동기화
	for _, r := range resultList {
		var subnet *model.ClusterSubnet
		subnet, err = s.getSubnet(&model.ClusterSubnet{UUID: r.Subnet.UUID})
		if err != nil {
			return err
		}

		var routingInterface model.ClusterNetworkRoutingInterface
		err = database.Execute(func(db *gorm.DB) error {
			return db.Where(&model.ClusterNetworkRoutingInterface{ClusterRouterID: router.ID, ClusterSubnetID: subnet.ID}).First(&routingInterface).Error
		})

		switch {
		case err == gorm.ErrRecordNotFound:
			// 새로운 routing interface 추가
			routingInterface.ClusterSubnetID = subnet.ID
			routingInterface.ClusterRouterID = router.ID
			routingInterface.ExternalFlag = r.RoutingInterface.ExternalFlag
			routingInterface.IPAddress = r.RoutingInterface.IPAddress

			if err = database.GormTransaction(func(db *gorm.DB) error {
				return db.Save(&routingInterface).Error
			}); err != nil {
				return errors.UnusableDatabase(err)
			}

			logger.Infof("[syncClusterNetworkRoutingInterfaceList] Done - added: cluster(%d) router(%s) subnet(%s)",
				s.clusterID, router.UUID, r.Subnet.UUID)

			if err = publishMessage(constant.QueueNoticeClusterRoutingInterfaceCreated, &queue.CreateClusterRoutingInterface{
				Cluster:          &model.Cluster{ID: s.clusterID},
				RoutingInterface: &routingInterface,
			}); err != nil {
				logger.Warnf("[syncClusterNetworkRoutingInterfaceList] Could not publish cluster routing interface created message: cluster(%d) router(%s) subnet(%s). Cause: %+v",
					s.clusterID, router.UUID, r.Subnet.UUID, err)
			}

		case err == nil && !compareRoutingInterface(routingInterface, r):
			origRoutingInterface := routingInterface

			routingInterface.ExternalFlag = r.RoutingInterface.ExternalFlag
			routingInterface.IPAddress = r.RoutingInterface.IPAddress

			if err = database.GormTransaction(func(db *gorm.DB) error {
				return db.Save(&routingInterface).Error
			}); err != nil {
				return errors.UnusableDatabase(err)
			}

			logger.Infof("[syncClusterNetworkRoutingInterfaceList] Done - updated: cluster(%d) router(%s) subnet(%s)",
				s.clusterID, router.UUID, r.Subnet.UUID)

			if err = publishMessage(constant.QueueNoticeClusterRoutingInterfaceUpdated, &queue.UpdateClusterRoutingInterface{
				Cluster:              &model.Cluster{ID: s.clusterID},
				OrigRoutingInterface: &origRoutingInterface,
				RoutingInterface:     &routingInterface,
			}); err != nil {
				logger.Warnf("[syncClusterNetworkRoutingInterfaceList] Could not publish cluster routing interface updated message: cluster(%d) router(%s) subnet(%s). Cause: %+v",
					s.clusterID, router.UUID, r.Subnet.UUID, err)
			}

		case err != nil:
			return errors.UnusableDatabase(err)
		}
	}

	return nil
}
