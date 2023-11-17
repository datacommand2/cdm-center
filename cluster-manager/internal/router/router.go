package router

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
)

type routerRecord struct {
	Router model.ClusterRouter `gorm:"embedded"`
	Tenant model.ClusterTenant `gorm:"embedded"`
}

type routerInterfaceRecord struct {
	NetworkRoutingInterface model.ClusterNetworkRoutingInterface `gorm:"embedded"`
	Subnet                  model.ClusterSubnet                  `gorm:"embedded"`
	Network                 model.ClusterNetwork                 `gorm:"embedded"`
}

func checkIsExistRoutingInterfaceByIPAddress(cid uint64, ipAddress string) (bool, error) {
	var cnt int
	if err := database.Execute(func(db *gorm.DB) error {
		return db.Table(model.ClusterNetworkRoutingInterface{}.TableName()).
			Joins("join cdm_cluster_router on cdm_cluster_router.id = cdm_cluster_routing_interface.cluster_router_id").
			Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_router.cluster_tenant_id").
			Where(&model.ClusterTenant{ClusterID: cid}).
			Where(&model.ClusterNetworkRoutingInterface{IPAddress: ipAddress}).
			Count(&cnt).Error
	}); err != nil {
		return false, errors.UnusableDatabase(err)
	}

	return cnt > 0, nil
}

// CheckIsExistRoutingInterface ip address 를 통한 클러스터 router 존재유무 확인
func CheckIsExistRoutingInterface(ctx context.Context, req *cms.CheckIsExistClusterRoutingInterfaceRequest) (bool, error) {
	var err error
	if err = validateCheckIsExistRoutingInterface(ctx, req); err != nil {
		logger.Errorf("[CheckIsExistRoutingInterface] Errors occurred during validating the request. Cause: %+v", err)
		return false, err
	}

	return checkIsExistRoutingInterfaceByIPAddress(req.ClusterId, req.GetClusterRoutingInterfaceIpAddress())
}

// CreateRouter 라우터 생성
func CreateRouter(req *cms.CreateClusterRouterRequest) (*cms.ClusterRouter, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateRouter] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err = validateCreateClusterRouterRequest(req); err != nil {
		logger.Errorf("[CreateRouter] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateRouter] Could not create synchronizer client. Cause: %+v", err)
		return nil, err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[CreateRouter] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateRouter] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return nil, err
	}
	mRouter, err := req.GetRouter().Model()
	if err != nil {
		return nil, err
	}
	var mExtraRoutes []model.ClusterRouterExtraRoute
	for _, extra := range req.GetRouter().GetExtraRoutes() {
		extraRoute, err := extra.Model()
		if err != nil {
			return nil, err
		}

		mExtraRoutes = append(mExtraRoutes, *extraRoute)
	}

	externals, err := convertRoutingInterfacesToClient(req.GetRouter().GetExternalRoutingInterfaces(), true)
	if err != nil {
		return nil, err
	}

	internals, err := convertRoutingInterfacesToClient(req.GetRouter().GetInternalRoutingInterfaces(), false)
	if err != nil {
		return nil, err
	}

	routerRsp, err := cli.CreateRouter(client.CreateRouterRequest{
		Tenant:                       *mTenant,
		Router:                       *mRouter,
		ExtraRouteList:               mExtraRoutes,
		ExternalRoutingInterfaceList: externals,
		InternalRoutingInterfaceList: internals,
	})
	if err != nil {
		logger.Errorf("[CreateRouter] Could not create the cluster router. Cause: %+v", err)
		return nil, err
	}

	var rollback = true
	defer func() {
		if rollback {
			if err := cli.DeleteRouter(client.DeleteRouterRequest{
				Tenant: *mTenant,
				Router: routerRsp.Router,
			}); err != nil {
				logger.Warnf("[CreateRouter] Could not delete cluster router. Cause: %+v", err)
			}
		}
	}()

	var ret cms.ClusterRouter
	if err := ret.SetFromModel(&routerRsp.Router); err != nil {
		return nil, err
	}

	var cExtraRoutes []*cms.ClusterRouterExtraRoute
	for _, extra := range routerRsp.ExtraRouteList {
		var cExtra cms.ClusterRouterExtraRoute
		if err := cExtra.SetFromModel(&extra); err != nil {
			return nil, err
		}

		cExtraRoutes = append(cExtraRoutes, &cExtra)
	}
	ret.ExtraRoutes = cExtraRoutes

	ret.InternalRoutingInterfaces, err = convertRoutingInterfacesToCMS(routerRsp.InternalRoutingInterfaceList)
	if err != nil {
		return nil, err
	}

	ret.ExternalRoutingInterfaces, err = convertRoutingInterfacesToCMS(routerRsp.ExternalRoutingInterfaceList)
	if err != nil {
		return nil, err
	}

	rollback = false
	return &ret, nil
}

// DeleteRouter 라우터 삭제
func DeleteRouter(req *cms.DeleteClusterRouterRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteRouter] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteClusterRouterRequest(req); err != nil {
		logger.Errorf("[DeleteRouter] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteRouter] Could not create client. Cause: %+v", err)
		return err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[DeleteRouter] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteRouter] Could not close client. Cause: %+v", err)
		}
	}()

	tenant, err := req.GetTenant().Model()
	if err != nil {
		return err
	}

	router, err := req.GetRouter().Model()
	if err != nil {
		return err
	}

	return cli.DeleteRouter(client.DeleteRouterRequest{
		Tenant: *tenant,
		Router: *router,
	})
}

func convertRoutingInterfacesToCMS(in []client.CreateRoutingInterface) ([]*cms.ClusterNetworkRoutingInterface, error) {
	var ret []*cms.ClusterNetworkRoutingInterface
	for _, item := range in {
		var network cms.ClusterNetwork
		var subnet cms.ClusterSubnet
		if err := network.SetFromModel(&item.Network); err != nil {
			return nil, err
		}
		if err := subnet.SetFromModel(&item.Subnet); err != nil {
			return nil, err
		}

		ret = append(ret, &cms.ClusterNetworkRoutingInterface{
			Network:   &network,
			Subnet:    &subnet,
			IpAddress: item.RoutingInterface.IPAddress,
		})
	}
	return ret, nil
}

func convertRoutingInterfacesToClient(in []*cms.ClusterNetworkRoutingInterface, isExternal bool) ([]client.CreateRoutingInterface, error) {
	var routingInterfaces []client.CreateRoutingInterface
	for _, item := range in {
		mNetwork, err := item.GetNetwork().Model()
		if err != nil {
			return nil, err
		}

		mSubnet, err := item.GetSubnet().Model()
		if err != nil {
			return nil, err
		}

		routingInterfaces = append(routingInterfaces, client.CreateRoutingInterface{
			Network: *mNetwork,
			Subnet:  *mSubnet,
			RoutingInterface: model.ClusterNetworkRoutingInterface{
				IPAddress:    item.GetIpAddress(),
				ExternalFlag: isExternal,
			},
		})
	}
	return routingInterfaces, nil
}

func getRouter(db *gorm.DB, cid, rid uint64) (*routerRecord, error) {
	var m routerRecord
	err := db.Table(model.ClusterRouter{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_router.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: cid}).
		Where(&model.ClusterRouter{ID: rid}).
		Find(&m).Error

	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterRouter(rid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &m, nil
}

func getRouterByUUID(db *gorm.DB, cid uint64, uuid string) (*routerRecord, error) {
	var m routerRecord
	err := db.Table(model.ClusterRouter{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_router.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: cid}).
		Where(&model.ClusterRouter{UUID: uuid}).
		First(&m).Error

	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterRouterByUUID(uuid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &m, nil
}

func getRouters(db *gorm.DB, cid uint64, filters ...routerFilter) ([]*routerRecord, error) {
	var err error
	cond := db
	for _, f := range filters {
		if cond, err = f.Apply(cond); err != nil {
			return nil, err
		}
	}

	var list []*routerRecord
	err = cond.Table(model.ClusterRouter{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_router.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: cid}).
		Find(&list).Error

	if err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return list, err
}

func getRoutingInterfaceRecords(db *gorm.DB, rid uint64) ([]routerInterfaceRecord, error) {
	var records []routerInterfaceRecord
	err := db.Table(model.ClusterNetworkRoutingInterface{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_subnet on cdm_cluster_subnet.id = cdm_cluster_routing_interface.cluster_subnet_id").
		Joins("join cdm_cluster_network on cdm_cluster_network.id = cdm_cluster_subnet.cluster_network_id").
		Where(&model.ClusterNetworkRoutingInterface{ClusterRouterID: rid}).
		Find(&records).Error
	if err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return records, err
}

// GetList 클러스터 라우터 목록을 조회하는 함수
func GetList(ctx context.Context, req *cms.ClusterRouterListRequest) ([]*cms.ClusterRouter, *cms.Pagination, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateGetClusterRouterListRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[Router-GetList] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	var rsp []*cms.ClusterRouter
	var pagination *cms.Pagination
	var filters []routerFilter

	if err = database.Execute(func(db *gorm.DB) error {
		filters = makeClusterRouterListFilters(db, req)

		var routers []*routerRecord
		if routers, err = getRouters(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[Router-GetList] Could not get the cluster(%d) routers. Cause: %+v", req.ClusterId, err)
			return err
		}

		for _, r := range routers {
			var router = cms.ClusterRouter{
				Tenant: new(cms.ClusterTenant),
			}

			if err = router.SetFromModel(&r.Router); err != nil {
				return err
			}
			if err = router.Tenant.SetFromModel(&r.Tenant); err != nil {
				return err
			}

			var records []routerInterfaceRecord
			if records, err = getRoutingInterfaceRecords(db, r.Router.ID); err != nil {
				return err
			}

			for _, rec := range records {
				var ri cms.ClusterNetworkRoutingInterface
				if err = ri.SetFromModel(&rec.NetworkRoutingInterface); err != nil {
					return err
				}

				ri.Network = new(cms.ClusterNetwork)
				if err = ri.Network.SetFromModel(&rec.Network); err != nil {
					return err
				}

				ri.Subnet = new(cms.ClusterSubnet)
				if err = ri.Subnet.SetFromModel(&rec.Subnet); err != nil {
					return err
				}

				var ns []model.ClusterSubnetNameserver
				if ns, err = rec.Subnet.GetSubnetNameservers(db); err != nil {
					return errors.UnusableDatabase(err)
				}

				for _, n := range ns {
					var cns cms.ClusterSubnetNameserver
					if err = cns.SetFromModel(&n); err != nil {
						return err
					}
					ri.Subnet.Nameservers = append(ri.Subnet.Nameservers, &cns)
				}

				if rec.NetworkRoutingInterface.ExternalFlag == true {
					router.ExternalRoutingInterfaces = append(router.ExternalRoutingInterfaces, &ri)
				} else {
					router.InternalRoutingInterfaces = append(router.InternalRoutingInterfaces, &ri)
				}
			}

			extraRoutes, err := r.Router.GetExtraRoutes(db)
			if err != nil {
				return errors.UnusableDatabase(err)
			}

			for _, route := range extraRoutes {
				var extraRoute cms.ClusterRouterExtraRoute
				if err = extraRoute.SetFromModel(&route); err != nil {
					return err
				}

				router.ExtraRoutes = append(router.ExtraRoutes, &extraRoute)
			}

			rsp = append(rsp, &router)
		}

		if pagination, err = getRoutersPagination(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[Router-GetList] Could not get the cluster routers pagination. Cause: %+v", err)
			return err
		}

		return nil
	}); err != nil {

	}

	return rsp, pagination, nil
}

func getRoutersPagination(db *gorm.DB, clusterID uint64, filters ...routerFilter) (*cms.Pagination, error) {
	var err error
	var offset, limit uint64

	var conditions = db

	for _, f := range filters {
		if _, ok := f.(*paginationFilter); ok {
			offset = f.(*paginationFilter).Offset.GetValue()
			limit = f.(*paginationFilter).Limit.GetValue()
			continue
		}

		conditions, err = f.Apply(conditions)
		if err != nil {
			return nil, errors.UnusableDatabase(err)
		}
	}

	var total uint64
	if err = conditions.Table(model.ClusterRouter{}.TableName()).
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_router.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: clusterID}).
		Count(&total).Error; err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	if limit == 0 {
		return &cms.Pagination{
			Page:       &wrappers.UInt64Value{Value: 1},
			TotalPage:  &wrappers.UInt64Value{Value: 1},
			TotalItems: &wrappers.UInt64Value{Value: total},
		}, nil
	}

	return &cms.Pagination{
		Page:       &wrappers.UInt64Value{Value: offset/limit + 1},
		TotalPage:  &wrappers.UInt64Value{Value: (total + limit - 1) / limit},
		TotalItems: &wrappers.UInt64Value{Value: total},
	}, nil
}

// Get 클러스터 라우터를 조회하는 함수
func Get(ctx context.Context, req *cms.ClusterRouterRequest) (*cms.ClusterRouter, error) {
	var err error
	var r *routerRecord

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateGetClusterRouterRequest(ctx, db, req); err != nil {
			logger.Errorf("[Router-Get] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if r, err = getRouter(db, req.ClusterId, req.ClusterRouterId); err != nil {
			logger.Errorf("[Router-Get] Could not get the cluster(%d) router(%d). Cause: %+v", req.ClusterId, req.ClusterRouterId, err)
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var router = cms.ClusterRouter{
		Tenant: new(cms.ClusterTenant),
	}

	if err = router.SetFromModel(&r.Router); err != nil {
		return nil, err
	}
	if err = router.Tenant.SetFromModel(&r.Tenant); err != nil {
		return nil, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		var records []routerInterfaceRecord
		if records, err = getRoutingInterfaceRecords(db, r.Router.ID); err != nil {
			logger.Errorf("[Router-Get] Could not get the cluster(%d) router(%d) records. Cause: %+v", req.ClusterId, r.Router.ID, err)
			return err
		}

		for _, rec := range records {
			var ri cms.ClusterNetworkRoutingInterface
			if err = ri.SetFromModel(&rec.NetworkRoutingInterface); err != nil {
				return err
			}

			ri.Network = new(cms.ClusterNetwork)
			if err = ri.Network.SetFromModel(&rec.Network); err != nil {
				return err
			}

			ri.Subnet = new(cms.ClusterSubnet)
			if err = ri.Subnet.SetFromModel(&rec.Subnet); err != nil {
				return err
			}

			var ns []model.ClusterSubnetNameserver
			if ns, err = rec.Subnet.GetSubnetNameservers(db); err != nil {
				return errors.UnusableDatabase(err)
			}

			for _, n := range ns {
				var cns cms.ClusterSubnetNameserver
				if err = cns.SetFromModel(&n); err != nil {
					return err
				}
				ri.Subnet.Nameservers = append(ri.Subnet.Nameservers, &cns)
			}

			if rec.NetworkRoutingInterface.ExternalFlag == true {
				router.ExternalRoutingInterfaces = append(router.ExternalRoutingInterfaces, &ri)
			} else {
				router.InternalRoutingInterfaces = append(router.InternalRoutingInterfaces, &ri)
			}
		}

		extraRoutes, err := r.Router.GetExtraRoutes(db)
		if err != nil {
			return errors.UnusableDatabase(err)
		}

		for _, route := range extraRoutes {
			var extraRoute cms.ClusterRouterExtraRoute
			if err = extraRoute.SetFromModel(&route); err != nil {
				return err
			}

			router.ExtraRoutes = append(router.ExtraRoutes, &extraRoute)
		}

		return nil
	}); err != nil {

	}

	return &router, nil
}

// GetByUUID uuid 를 통해 클러스터 라우터를 조회하는 함수
func GetByUUID(ctx context.Context, req *cms.ClusterRouterByUUIDRequest) (*cms.ClusterRouter, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateGetClusterRouterByUUIDRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[Router-GetByUUID] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	var router = cms.ClusterRouter{
		Tenant: new(cms.ClusterTenant),
	}

	if err = database.Execute(func(db *gorm.DB) error {
		var r *routerRecord
		if r, err = getRouterByUUID(db, req.ClusterId, req.Uuid); err != nil {
			logger.Errorf("[Router-Get] Could not get the cluster(%d) router by uuid(%s). Cause: %+v", req.ClusterId, req.Uuid, err)
			return err
		}

		if err = router.SetFromModel(&r.Router); err != nil {
			return err
		}
		if err = router.Tenant.SetFromModel(&r.Tenant); err != nil {
			return err
		}

		var records []routerInterfaceRecord
		if records, err = getRoutingInterfaceRecords(db, r.Router.ID); err != nil {
			logger.Errorf("[Router-Get] Could not get the cluster(%d) routing(%d) interface records. Cause: %+v", req.ClusterId, r.Router.ID, err)
			return err
		}

		for _, rec := range records {
			var ri cms.ClusterNetworkRoutingInterface
			if err = ri.SetFromModel(&rec.NetworkRoutingInterface); err != nil {
				return err
			}

			ri.Network = new(cms.ClusterNetwork)
			if err = ri.Network.SetFromModel(&rec.Network); err != nil {
				return err
			}

			ri.Subnet = new(cms.ClusterSubnet)
			if err = ri.Subnet.SetFromModel(&rec.Subnet); err != nil {
				return err
			}

			var ns []model.ClusterSubnetNameserver
			if ns, err = rec.Subnet.GetSubnetNameservers(db); err != nil {
				return errors.UnusableDatabase(err)
			}

			for _, n := range ns {
				var cns cms.ClusterSubnetNameserver
				if err = cns.SetFromModel(&n); err != nil {
					return err
				}
				ri.Subnet.Nameservers = append(ri.Subnet.Nameservers, &cns)
			}

			if rec.NetworkRoutingInterface.ExternalFlag == true {
				router.ExternalRoutingInterfaces = append(router.ExternalRoutingInterfaces, &ri)
			} else {
				router.InternalRoutingInterfaces = append(router.InternalRoutingInterfaces, &ri)
			}
		}

		extraRoutes, err := r.Router.GetExtraRoutes(db)
		if err != nil {
			return errors.UnusableDatabase(err)
		}

		for _, route := range extraRoutes {
			var extraRoute cms.ClusterRouterExtraRoute
			if err = extraRoute.SetFromModel(&route); err != nil {
				return err
			}

			router.ExtraRoutes = append(router.ExtraRoutes, &extraRoute)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &router, nil
}

func validateCheckIsExistRoutingInterface(ctx context.Context, req *cms.CheckIsExistClusterRoutingInterfaceRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetClusterRoutingInterfaceIpAddress() == "" {
		return errors.RequiredParameter("cluster_routing_interface_address")
	}

	return database.Execute(func(db *gorm.DB) error {
		return cluster.IsClusterOwner(ctx, db, req.ClusterId)
	})
}
