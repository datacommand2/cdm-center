package network

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
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
)

type networkRecord struct {
	Network model.ClusterNetwork `gorm:"embedded"`
	Tenant  model.ClusterTenant  `gorm:"embedded"`
}

// CreateNetwork 네트워크 생성
func CreateNetwork(req *cms.CreateClusterNetworkRequest) (*cms.ClusterNetwork, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateNetwork] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err = validateCreateClusterNetworkRequest(req); err != nil {
		logger.Errorf("[CreateNetwork] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateNetwork] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[CreateNetwork] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateNetwork] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return nil, err
	}
	mNetwork, err := req.GetNetwork().Model()
	if err != nil {
		return nil, err
	}

	networkRsp, err := cli.CreateNetwork(client.CreateNetworkRequest{
		Tenant:  *mTenant,
		Network: *mNetwork,
	})
	if err != nil {
		logger.Errorf("[CreateNetwork] Could not create the cluster network. Cause: %+v", err)
		return nil, err
	}

	var rollback = true
	defer func() {
		if rollback {
			if err := cli.DeleteNetwork(client.DeleteNetworkRequest{
				Tenant:  *mTenant,
				Network: networkRsp.Network,
			}); err != nil {
				logger.Warnf("[CreateNetwork] Could not delete cluster network. Cause: %+v", err)
			}
		}
	}()

	var cNetwork cms.ClusterNetwork
	if err := cNetwork.SetFromModel(&networkRsp.Network); err != nil {
		return nil, err
	}

	rollback = false
	return &cNetwork, nil
}

// CreateSubnet 서브넷 생성
func CreateSubnet(req *cms.CreateClusterSubnetRequest) (*cms.ClusterSubnet, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateSubnet] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err = validateCreateClusterSubnetRequest(req); err != nil {
		logger.Errorf("[CreateSubnet] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateSubnet] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[CreateSubnet] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateSubnet] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return nil, err
	}

	mNetwork, err := req.GetNetwork().Model()
	if err != nil {
		return nil, err
	}

	mSubnet, err := req.GetSubnet().Model()
	if err != nil {
		return nil, err
	}

	var mPools []model.ClusterSubnetDHCPPool
	for _, pool := range req.GetSubnet().GetDhcpPools() {
		mPool, err := pool.Model()
		if err != nil {
			return nil, err
		}
		mPools = append(mPools, *mPool)
	}

	var mNameservers []model.ClusterSubnetNameserver
	for _, nameserver := range req.GetSubnet().GetNameservers() {
		mNameserver, err := nameserver.Model()
		if err != nil {
			return nil, err
		}
		mNameservers = append(mNameservers, *mNameserver)
	}

	subnetRsp, err := cli.CreateSubnet(client.CreateSubnetRequest{
		Tenant:         *mTenant,
		Network:        *mNetwork,
		Subnet:         *mSubnet,
		NameserverList: mNameservers,
		PoolsList:      mPools,
	})
	if err != nil {
		logger.Errorf("[CreateSubnet] Could not create the cluster subnet. Cause: %+v", err)
		return nil, err
	}

	var rollback = true
	defer func() {
		if rollback {
			if err := cli.DeleteSubnet(client.DeleteSubnetRequest{
				Tenant:  *mTenant,
				Network: *mNetwork,
				Subnet:  subnetRsp.Subnet,
			}); err != nil {
				logger.Warnf("[CreateSubnet] Could not delete the cluster subnet. Cause: %+v", err)
			}
		}
	}()

	var cSubnet cms.ClusterSubnet
	if err := cSubnet.SetFromModel(&subnetRsp.Subnet); err != nil {
		return nil, err
	}

	var cPools []*cms.ClusterSubnetDHCPPool
	for _, pool := range subnetRsp.PoolsList {
		var cPool cms.ClusterSubnetDHCPPool
		if err := cPool.SetFromModel(&pool); err != nil {
			return nil, err
		}

		cPools = append(cPools, &cPool)
	}
	cSubnet.DhcpPools = cPools

	var cNameservers []*cms.ClusterSubnetNameserver
	for _, nameserver := range subnetRsp.NameserverList {
		var cNameserver cms.ClusterSubnetNameserver
		if err := cNameserver.SetFromModel(&nameserver); err != nil {
			return nil, err
		}
		cNameservers = append(cNameservers, &cNameserver)
	}
	cSubnet.Nameservers = cNameservers

	rollback = false
	return &cSubnet, nil
}

// DeleteNetwork 네트워크 삭제
func DeleteNetwork(req *cms.DeleteClusterNetworkRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteNetwork] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err := validateDeleteClusterNetworkRequest(req); err != nil {
		logger.Errorf("[DeleteNetwork] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteNetwork] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[DeleteNetwork] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteNetwork] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return err
	}

	mNetwork, err := req.GetNetwork().Model()
	if err != nil {
		return err
	}

	return cli.DeleteNetwork(client.DeleteNetworkRequest{
		Tenant:  *mTenant,
		Network: *mNetwork,
	})
}

// DeleteSubnet 서브넷 삭제
func DeleteSubnet(req *cms.DeleteClusterSubnetRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteSubnet] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteClusterSubnetRequest(req); err != nil {
		logger.Errorf("[DeleteSubnet] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteSubnet] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[DeleteSubnet] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteSubnet] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return err
	}

	mNetwork, err := req.GetNetwork().Model()
	if err != nil {
		return err
	}

	mSubnet, err := req.GetSubnet().Model()
	if err != nil {
		return err
	}

	return cli.DeleteSubnet(client.DeleteSubnetRequest{
		Tenant:  *mTenant,
		Network: *mNetwork,
		Subnet:  *mSubnet,
	})
}

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

func getClusterNetwork(db *gorm.DB, cid, nid uint64) (*networkRecord, error) {
	var (
		err    error
		record networkRecord
	)
	err = db.Table(model.ClusterNetwork{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_network.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: cid}).
		Where(&model.ClusterNetwork{ID: nid}).First(&record).Error

	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterNetwork(nid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &record, nil
}

func getClusterNetworkByUUID(db *gorm.DB, cid uint64, uuid string) (*networkRecord, error) {
	var (
		err    error
		record networkRecord
	)
	err = db.Table(model.ClusterNetwork{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_network.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: cid}).
		Where(&model.ClusterNetwork{UUID: uuid}).First(&record).Error

	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterNetworkByUUID(uuid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &record, nil
}

func getClusterSubnet(db *gorm.DB, cid, sid uint64) (*model.ClusterSubnet, error) {
	var s model.ClusterSubnet
	err := db.Table(model.ClusterSubnet{}.TableName()).
		Joins("join cdm_cluster_network on cdm_cluster_network.id = cdm_cluster_subnet.cluster_network_id").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_network.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: cid}).
		Where(&model.ClusterSubnet{ID: sid}).First(&s).Error

	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterSubnet(sid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &s, nil
}

// Get 클러스터 네트워크 조회
func Get(ctx context.Context, req *cms.ClusterNetworkRequest) (*cms.ClusterNetwork, error) {
	var err error
	var c *model.Cluster
	var record *networkRecord

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateNetworkRequest(ctx, db, req); err != nil {
			logger.Errorf("[Network-Get] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Network-Get] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		if record, err = getClusterNetwork(db, req.ClusterId, req.ClusterNetworkId); err != nil {
			logger.Errorf("[Network-Get] Could not get the cluster(%d) network(%d). Cause: %+v", req.ClusterId, req.ClusterNetworkId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var rsp cms.ClusterNetwork
	if err = rsp.SetFromModel(&record.Network); err != nil {
		return nil, err
	}

	rsp.Cluster = new(cms.Cluster)
	if err = rsp.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}

	rsp.Tenant = new(cms.ClusterTenant)
	if err = rsp.Tenant.SetFromModel(&record.Tenant); err != nil {
		return nil, err
	}

	var subnets []model.ClusterSubnet
	if err = database.Execute(func(db *gorm.DB) error {
		if subnets, err = record.Network.GetSubnets(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		for _, s := range subnets {
			subnet := new(cms.ClusterSubnet)
			if err = subnet.SetFromModel(&s); err != nil {
				return err
			}
			rsp.Subnets = append(rsp.Subnets, subnet)

			var dhcpPool []model.ClusterSubnetDHCPPool
			if dhcpPool, err = s.GetDHCPPools(db); err != nil {
				return errors.UnusableDatabase(err)
			}

			for _, d := range dhcpPool {
				dhcp := new(cms.ClusterSubnetDHCPPool)
				if err = dhcp.SetFromModel(&d); err != nil {
					return err
				}
				subnet.DhcpPools = append(subnet.DhcpPools, dhcp)
			}

			var nss []model.ClusterSubnetNameserver
			if nss, err = s.GetSubnetNameservers(db); err != nil {
				return errors.UnusableDatabase(err)
			}

			for _, n := range nss {
				nameserver := new(cms.ClusterSubnetNameserver)
				if err = nameserver.SetFromModel(&n); err != nil {
					return err
				}
				subnet.Nameservers = append(subnet.Nameservers, nameserver)
			}
		}

		var floatingIPs []model.ClusterFloatingIP
		if floatingIPs, err = record.Network.GetFloatingIPs(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		for _, f := range floatingIPs {
			floatingIP := new(cms.ClusterFloatingIP)
			if err = floatingIP.SetFromModel(&f); err != nil {
				return err
			}
			rsp.FloatingIps = append(rsp.FloatingIps, floatingIP)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &rsp, nil
}

// GetByUUID uuid 를 통해 클러스터 네트워크 조회
func GetByUUID(ctx context.Context, req *cms.ClusterNetworkByUUIDRequest) (*cms.ClusterNetwork, error) {
	var err error
	var c *model.Cluster

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateNetworkByUUIDRequest(ctx, db, req); err != nil {
			logger.Errorf("[Network-GetByUUID] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Network-GetByUUID] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var record *networkRecord
	if err = database.Execute(func(db *gorm.DB) error {
		if record, err = getClusterNetworkByUUID(db, req.ClusterId, req.Uuid); err != nil {
			logger.Errorf("[Network-GetByUUID] Could not get the cluster(%d) network by uuid(%s). Cause: %+v", req.ClusterId, req.Uuid, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var rsp cms.ClusterNetwork
	if err = rsp.SetFromModel(&record.Network); err != nil {
		return nil, err
	}

	rsp.Cluster = new(cms.Cluster)
	if err = rsp.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}

	rsp.Tenant = new(cms.ClusterTenant)
	if err = rsp.Tenant.SetFromModel(&record.Tenant); err != nil {
		return nil, err
	}

	var subnets []model.ClusterSubnet
	if err = database.Execute(func(db *gorm.DB) error {
		if subnets, err = record.Network.GetSubnets(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		for _, s := range subnets {
			subnet := new(cms.ClusterSubnet)
			if err = subnet.SetFromModel(&s); err != nil {
				return err
			}
			rsp.Subnets = append(rsp.Subnets, subnet)

			var dhcpPool []model.ClusterSubnetDHCPPool
			if dhcpPool, err = s.GetDHCPPools(db); err != nil {
				return errors.UnusableDatabase(err)
			}

			for _, d := range dhcpPool {
				dhcp := new(cms.ClusterSubnetDHCPPool)
				if err = dhcp.SetFromModel(&d); err != nil {
					return err
				}
				subnet.DhcpPools = append(subnet.DhcpPools, dhcp)
			}

			var nss []model.ClusterSubnetNameserver
			if nss, err = s.GetSubnetNameservers(db); err != nil {
				return errors.UnusableDatabase(err)
			}

			for _, n := range nss {
				nameserver := new(cms.ClusterSubnetNameserver)
				if err = nameserver.SetFromModel(&n); err != nil {
					return err
				}
				subnet.Nameservers = append(subnet.Nameservers, nameserver)
			}
		}

		var floatingIPs []model.ClusterFloatingIP
		if floatingIPs, err = record.Network.GetFloatingIPs(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		for _, f := range floatingIPs {
			floatingIP := new(cms.ClusterFloatingIP)
			if err = floatingIP.SetFromModel(&f); err != nil {
				return err
			}
			rsp.FloatingIps = append(rsp.FloatingIps, floatingIP)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &rsp, nil
}

func getClusterNetworks(db *gorm.DB, cid uint64, filters ...networkFilter) ([]networkRecord, error) {
	var err error
	var condition = db
	for _, f := range filters {
		if condition, err = f.Apply(condition); err != nil {
			return nil, err
		}
	}

	var records []networkRecord
	err = condition.Table(model.ClusterNetwork{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_network.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: cid}).Find(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetList 클러스터 네트워크 목록 조회
func GetList(ctx context.Context, req *cms.ClusterNetworkListRequest) ([]*cms.ClusterNetwork, *cms.Pagination, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateNetworkListRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[Network-GetList] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	var c *model.Cluster
	if err = database.Execute(func(db *gorm.DB) error {
		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Network-GetList] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	var cmsCluster cms.Cluster
	if err = cmsCluster.SetFromModel(c); err != nil {
		return nil, nil, err
	}

	filters := makeClusterNetworkFilters(req)

	var records []networkRecord
	if err = database.Execute(func(db *gorm.DB) error {
		if records, err = getClusterNetworks(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[Network-GetList] Could not get the cluster(%d) networks. Cause: %+v", req.ClusterId, err)
			return errors.UnusableDatabase(err)
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}

	var rsp []*cms.ClusterNetwork
	for _, record := range records {
		var n cms.ClusterNetwork
		if err = n.SetFromModel(&record.Network); err != nil {
			return nil, nil, err
		}

		n.Cluster = &cmsCluster

		n.Tenant = new(cms.ClusterTenant)
		if err = n.Tenant.SetFromModel(&record.Tenant); err != nil {
			return nil, nil, err
		}

		rsp = append(rsp, &n)
	}

	var pagination *cms.Pagination
	if err = database.Execute(func(db *gorm.DB) error {
		if pagination, err = getNetworksPagination(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[Network-GetList] Could not get the cluster networks pagination. Cause: %+v", err)
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}

	return rsp, pagination, nil
}

func getNetworksPagination(db *gorm.DB, clusterID uint64, filters ...networkFilter) (*cms.Pagination, error) {
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
	if err = conditions.Table(model.ClusterNetwork{}.TableName()).
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_network.cluster_tenant_id").
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

// GetSubnet 클러스터 서브넷 조회
func GetSubnet(ctx context.Context, req *cms.ClusterSubnetRequest) (*cms.ClusterSubnet, error) {
	var err error
	var s *model.ClusterSubnet

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateGetClusterSubnetRequest(ctx, db, req); err != nil {
			logger.Errorf("[GetSubnet] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if s, err = getClusterSubnet(db, req.ClusterId, req.ClusterSubnetId); err != nil {
			logger.Errorf("[GetSubnet] Could not get the cluster(%d) subnet(%d). Cause: %+v", req.ClusterId, req.ClusterSubnetId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var subnet cms.ClusterSubnet
	if err = subnet.SetFromModel(s); err != nil {
		return nil, err
	}

	var pools []model.ClusterSubnetDHCPPool
	var nameservers []model.ClusterSubnetNameserver

	if err = database.Execute(func(db *gorm.DB) error {
		if pools, err = s.GetDHCPPools(db); err != nil {
			return errors.UnusableDatabase(err)
		}
		if nameservers, err = s.GetSubnetNameservers(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	for _, p := range pools {
		var pool cms.ClusterSubnetDHCPPool
		if err = pool.SetFromModel(&p); err != nil {
			return nil, err
		}
		subnet.DhcpPools = append(subnet.DhcpPools, &pool)
	}

	for _, ns := range nameservers {
		var nameserver cms.ClusterSubnetNameserver
		if err = nameserver.SetFromModel(&ns); err != nil {
			return nil, err
		}
		subnet.Nameservers = append(subnet.Nameservers, &nameserver)
	}

	return &subnet, nil
}
