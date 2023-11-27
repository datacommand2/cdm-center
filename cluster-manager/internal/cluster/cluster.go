package cluster

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/config"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/database"
	commonModel "github.com/datacommand2/cdm-cloud/common/database/model"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	drConstant "github.com/datacommand2/cdm-disaster-recovery/common/constant"
	"github.com/datacommand2/cdm-disaster-recovery/common/migrator"
	//drms "github.com/datacommand2/cdm-disaster-recovery/services/manager/proto"

	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
	"strings"
	"time"

	// for openstack client
	_ "github.com/datacommand2/cdm-center/cluster-manager/client/openstack"
)

const (
	updateLayout        = "2006-01-02T15:04:05.000000"
	heartbeatTimeLayout = "2006-01-02 15:04:05"
)

// DurationToString Duration 타입을 계산하여 string 으로 내보내는 함수
func DurationToString(d time.Duration) string {
	year := d / (365 * 24 * time.Hour)
	month := (d % (365 * 24 * time.Hour)) / (30 * 24 * time.Hour)
	day := (d % (30 * 24 * time.Hour)) / (24 * time.Hour)
	hour := (d % (24 * time.Hour)) / time.Hour
	minute := (d % time.Hour) / time.Minute

	if year > 0 {
		return fmt.Sprintf("%d년 %d개월 %d일 %d시간 %d분", year, month, day, hour, minute)
	} else if month > 0 {
		return fmt.Sprintf("%d개월 %d일 %d시간 %d분", month, day, hour, minute)
	} else if day > 0 {
		return fmt.Sprintf("%d일 %d시간 %d분", day, hour, minute)
	} else if hour > 0 {
		return fmt.Sprintf("%d시간 %d분", hour, minute)
	} else {
		return fmt.Sprintf("%d분", minute)
	}
}

/*
CheckCluster 클러스터 상태 조회
data 매개 변수는 메모리에 저장되어있는 데이터입니다.
LastUpdated 값은 DurationToString 함수를 거쳐 현재시간과 업데이트 된 시간을 빼고 그 시간에 맞게 규정화된 스트링값으로 반환합니다.
*/
func CheckCluster(c *model.Cluster, data queue.ClusterServiceStatus) ([]*cms.ComputeStatus, []*cms.StorageStatus, []*cms.NetworkStatus, string, error) {
	var err error
	var updatedAt time.Time
	var storages []*cms.StorageStatus
	var computes []*cms.ComputeStatus
	var networks []*cms.NetworkStatus

	exception, err := GetException(c.ID)
	if err != nil && !errors.Equal(err, internal.ErrNotFoundClusterException) {
		logger.Errorf("[Cluster-CheckCluster] Could not get exception: clusters/%d/check/exception. Cause: %+v", c.ID, err)
		return nil, nil, nil, "", err
	}

	// 스토리지 서비스 상태
	if data.Storage != nil {
		for _, v := range data.Storage {
			if updatedAt, err = time.Parse(updateLayout, v.UpdatedAt); err != nil {
				logger.Errorf("[Cluster-CheckCluster] Could not parse update_at of storage. Cause : ", err)
				return nil, nil, nil, "", errors.InvalidParameterValue("cluster.updated_at", v.UpdatedAt, err.Error())
			}

			except := false
			if exception != nil {
				for _, ex := range exception.Storage {
					if v.Binary == ex.Binary && v.Host == ex.Host && v.Zone == ex.Zone && ex.Exception {
						except = true
					}
				}
			}

			storages = append(storages, &cms.StorageStatus{
				Id:          v.ID,
				Binary:      v.Binary,
				BackendName: v.BackendName,
				Host:        v.Host,
				Zone:        v.Zone,
				Status:      v.Status,
				LastUpdated: DurationToString(time.Now().Sub(updatedAt)),
				Exception:   except,
				Deleted:     v.Deleted,
			})
		}
	}

	// 컴퓨트 서비스 상태
	if data.Compute != nil {
		for _, v := range data.Compute {
			if updatedAt, err = time.Parse(updateLayout, v.UpdatedAt); err != nil {
				logger.Errorf("[Cluster-CheckCluster] Could not parse update_at of compute. Cause : ", err)
				return nil, nil, nil, "", errors.InvalidParameterValue("cluster.updated_at", v.UpdatedAt, err.Error())
			}

			except := false
			if exception != nil {
				if ex, ok := exception.Compute[v.ID]; ok && ex.Exception {
					except = true
				}
			}

			computes = append(computes, &cms.ComputeStatus{
				Id:          v.ID,
				Binary:      v.Binary,
				Host:        v.Host,
				Zone:        v.Zone,
				Status:      v.Status,
				LastUpdated: DurationToString(time.Now().Sub(updatedAt)),
				Exception:   except,
				Deleted:     v.Deleted,
			})
		}
	}

	// 네트워크 에이전트 서비스 상태
	if data.Network != nil {
		for _, v := range data.Network {
			if updatedAt, err = time.Parse(heartbeatTimeLayout, v.HeartbeatTimestamp); err != nil {
				logger.Errorf("[Cluster-CheckCluster] Could not parse heartbeat_timestamp of network. Cause : ", err)
				return nil, nil, nil, "", errors.InvalidParameterValue("cluster.heartbeat_timestamp", v.HeartbeatTimestamp, err.Error())
			}

			except := false
			if exception != nil {
				if ex, ok := exception.Network[v.ID]; ok && ex.Exception {
					except = true
				}
			}

			networks = append(networks, &cms.NetworkStatus{
				Id:          v.ID,
				Type:        v.Type,
				Binary:      v.Binary,
				Host:        v.Host,
				Status:      v.Status,
				LastUpdated: DurationToString(time.Now().Sub(updatedAt)),
				Exception:   except,
				Deleted:     v.Deleted,
			})
		}
	}

	return computes, storages, networks, data.Status, nil
}

// SyncException 클러스터 nova, cinder, compute 상태를 조회할때 제외하고싶은 서비스를 적용
func SyncException(req *cms.SyncExceptionRequest) (map[string]*cms.NetworkStatus, map[string]*cms.ComputeStatus, map[string]*cms.StorageStatus, error) {
	networks := make(map[string]*cms.NetworkStatus)
	storages := make(map[string]*cms.StorageStatus)
	computes := make(map[string]*cms.ComputeStatus)

	if _, err := GetException(req.ClusterId); err == nil {
		// Storage 는 서비스가 한번 꺼진 후 켜졌다면 ID가 다르기 때문에 데이터가 쌓이는걸 방지하기 위해 삭제후 저장한다.
		if err = DeleteException(req.ClusterId); err != nil {
			return nil, nil, nil, err
		}
	} else if err != nil && !errors.Equal(err, internal.ErrNotFoundClusterException) {
		return nil, nil, nil, err
	}

	if req.Networks != nil {
		for _, r := range req.Networks {
			networks[r.Id] = r
			if r.Exception {
				logger.Infof("[Cluster-SyncException] Network(%s:%s) service exception processing is complete.", r.Type, r.Host)
			}
		}
	}

	if req.Storages != nil {
		for _, r := range req.Storages {
			storages[r.Id] = r
			if r.Exception {
				logger.Infof("[Cluster-SyncException] Storage(%s:%s) service exception processing is complete.", r.Binary, r.BackendName)
			}
		}
	}

	if req.Computes != nil {
		for _, r := range req.Computes {
			computes[r.Id] = r
			if r.Exception {
				logger.Infof("[Cluster-SyncException] Compute(%s:%s) service exception processing is complete.", r.Binary, r.Host)
			}
		}
	}

	if err := SetException(req.ClusterId, queue.ExcludedCluster{
		Network: networks,
		Compute: computes,
		Storage: storages,
	}); err != nil {
		logger.Errorf("[Cluster-SyncException] Could not set sync exception: clusters/%d/check/exception. Cause: %+v", req.ClusterId, err)
		return nil, nil, nil, err
	}
	logger.Infof("[Cluster-SyncException] Success - set sync exception: clusters/%d/check/exception.", req.ClusterId)

	return networks, computes, storages, nil
}

// CheckConnection 클러스터 접속 가능여부 확인
func CheckConnection(conn *cms.ClusterConnectionInfo) error {
	if err := CheckValidClusterConnectionInfo(conn); err != nil {
		logger.Errorf("[Cluster-CheckConnection] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(conn.GetTypeCode(), conn.GetApiServerUrl(), conn.GetCredential(), conn.GetTenantId())
	if err != nil {
		logger.Errorf("[Cluster-CheckConnection] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[Cluster-CheckConnection] Could not connect client. Cause: %+v", err)
		return err
	}

	if err = cli.Close(); err != nil {
		logger.Warnf("[Cluster-CheckConnection] Could not close client. Cause: %+v", err)
	}

	return nil
}

func validateUpdateCredential(req *cms.UpdateCredentialRequest, c *model.Cluster) error {
	cred, err := client.DecryptCredentialPassword(req.NewCredential)
	if err != nil {
		return err
	}

	if err := CheckConnection(&cms.ClusterConnectionInfo{
		TypeCode:     c.TypeCode,
		ApiServerUrl: c.APIServerURL,
		Credential:   cred,
	}); err != nil {
		return err
	}
	return nil
}

// UpdateCredential 클러스터 인증 정보 수정
func UpdateCredential(ctx context.Context, db *gorm.DB, req *cms.UpdateCredentialRequest) error {
	var c *model.Cluster
	var err error
	if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
		logger.Errorf("[Cluster-UpdateCredential] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
		return err
	}

	if err = validateUpdateCredential(req, c); err != nil {
		logger.Errorf("[Cluster-UpdateCredential] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	c.Credential = req.NewCredential

	if err = db.Save(&c).Error; err != nil {
		logger.Errorf("[Cluster-UpdateCredential] Could not update the cluster(%d) credential. Cause: %+v", req.ClusterId, err)
		return errors.UnusableDatabase(err)
	}

	return nil
}

func getClusters(db *gorm.DB, filters ...clusterFilter) ([]model.Cluster, error) {
	var cond = db
	var err error

	for _, f := range filters {
		cond, err = f.Apply(cond)
		if err != nil {
			return nil, err
		}
	}

	var clusters []model.Cluster
	items := []string{
		"id",
		"owner_group_id",
		"name",
		"type_code",
		"state_code",
		"remarks",
		"created_at",
		"updated_at",
		"synchronized_at",
	}
	if err = cond.Select(strings.Join(items, ",")).Find(&clusters).Error; err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return clusters, nil
}

// GetList 클러스터 목록 조회
func GetList(ctx context.Context, db *gorm.DB, req *cms.ClusterListRequest) ([]*cms.Cluster, *cms.Pagination, error) {
	var err error
	if err = validateGetClusters(ctx, req); err != nil {
		logger.Errorf("[Cluster-GetList] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	if req.Sync {
		// TODO: not implements
	}

	filters := makeClusterFilter(ctx, req)

	var cs []model.Cluster
	if cs, err = getClusters(db, filters...); err != nil {
		logger.Errorf("[Cluster-GetList] Could not get the clusters. Cause: %+v", err)
		return nil, nil, err
	}

	var rsp []*cms.Cluster
	for _, c := range cs {
		var tmp cms.Cluster

		if err = tmp.SetFromModel(&c); err != nil {
			return nil, nil, err
		}

		var owner *commonModel.Group
		if owner, err = getGroup(ctx, db, c.OwnerGroupID); err != nil {
			logger.Errorf("[Cluster-GetList] Could not get the group(%d). Cause: %+v", c.OwnerGroupID, err)
			return nil, nil, err
		}

		tmp.OwnerGroup = new(cms.Group)
		if err = tmp.OwnerGroup.SetFromModel(owner); err != nil {
			return nil, nil, err
		}

		rsp = append(rsp, &tmp)
	}

	var pagination *cms.Pagination
	if pagination, err = getClustersPagination(db, filters...); err != nil {
		logger.Errorf("[Cluster-GetList] Could not get the clusters pagination. Cause: %+v", err)
		return nil, nil, err
	}

	return rsp, pagination, nil
}

func getClustersPagination(db *gorm.DB, filters ...clusterFilter) (*cms.Pagination, error) {
	var err error
	var offset, limit uint64

	var cond = db

	for _, f := range filters {
		if _, ok := f.(*paginationFilter); ok {
			offset = f.(*paginationFilter).Offset.GetValue()
			limit = f.(*paginationFilter).Limit.GetValue()
			continue
		}

		cond, err = f.Apply(cond)
		if err != nil {
			return nil, errors.UnusableDatabase(err)
		}
	}

	var total uint64
	if err = cond.Model(&model.Cluster{}).Count(&total).Error; err != nil {
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

func getGroup(ctx context.Context, db *gorm.DB, id uint64) (*commonModel.Group, error) {
	tid, _ := metadata.GetTenantID(ctx)

	var m commonModel.Group

	items := []string{
		"id",
		"name",
		"remarks",
		"deleted_flag",
	}

	err := db.Select(strings.Join(items, ",")).First(&m, &commonModel.Group{ID: id, TenantID: tid}).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundGroup(id)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &m, nil
}

func getClusterPermissions(_ context.Context, db *gorm.DB, clusterID uint64) ([]model.ClusterPermission, error) {
	var m []model.ClusterPermission
	err := db.Find(&m, &model.ClusterPermission{ClusterID: clusterID}).Error
	switch {
	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return m, nil
}

func getCluster(ctx context.Context, db *gorm.DB, id uint64) (*model.Cluster, error) {
	tid, _ := metadata.GetTenantID(ctx)

	var m model.Cluster
	err := db.First(&m, &model.Cluster{ID: id, TenantID: tid}).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundCluster(id)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &m, nil
}

// Get 클러스터 조회
func Get(ctx context.Context, db *gorm.DB, req *cms.ClusterRequest) (*cms.Cluster, error) {
	var c *model.Cluster
	var owner *commonModel.Group
	var perms []model.ClusterPermission
	var err error

	if req.Sync {
		// TODO: not implements
	}

	if req.GetClusterId() == 0 {
		err = errors.RequiredParameter("cluster_id")
		logger.Errorf("[Cluster-Get] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
		logger.Errorf("[Cluster-Get] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
		return nil, err
	}
	if owner, err = getGroup(ctx, db, c.OwnerGroupID); err != nil {
		logger.Errorf("[Cluster-Get] Could not get the group(%d). Cause: %+v", c.OwnerGroupID, err)
		return nil, err
	}
	if perms, err = getClusterPermissions(ctx, db, c.ID); err != nil {
		logger.Errorf("[Cluster-Get] Could not get the cluster(%d) permissions. Cause: %+v", c.ID, err)
		return nil, err
	}

	user, _ := metadata.GetAuthenticatedUser(ctx)
	if !internal.IsAdminUser(user) && !internal.IsGroupUser(user, c.OwnerGroupID) {
		logger.Errorf("[Cluster-Get] Errors occurred during checking the authentication of the user.")
		return nil, errors.UnauthorizedRequest(ctx)
	}

	var rsp cms.Cluster
	if err := rsp.SetFromModel(c); err != nil {
		return nil, err
	}

	rsp.OwnerGroup = new(cms.Group)
	if err := rsp.OwnerGroup.SetFromModel(owner); err != nil {
		return nil, err
	}

	for _, p := range perms {
		var grp *commonModel.Group
		if grp, err = getGroup(ctx, db, p.GroupID); err != nil {
			logger.Errorf("[Cluster-Get] Could not get the group(%d). Cause: %+v", p.GroupID, err)
			return nil, err
		}

		var perm cms.ClusterPermission
		if err := perm.SetFromModel(&p); err != nil {
			return nil, err
		}

		perm.Group = new(cms.Group)
		if err := perm.Group.SetFromModel(grp); err != nil {
			return nil, err
		}

		rsp.Permissions = append(rsp.Permissions, &perm)
	}

	return &rsp, nil
}

// Add 클러스터 등록
func Add(ctx context.Context, db *gorm.DB, req *cms.AddClusterRequest) (*cms.Cluster, error) {
	var err error
	credential, err := client.DecryptCredentialPassword(req.Cluster.Credential)
	if err != nil {
		logger.Errorf("[Cluster-Add] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err = validateAddClusterRequest(ctx, db, req, credential); err != nil {
		logger.Errorf("[Cluster-Add] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	if err = CheckConnection(&cms.ClusterConnectionInfo{
		TypeCode:     req.Cluster.TypeCode,
		ApiServerUrl: req.Cluster.ApiServerUrl,
		Credential:   credential,
	}); err != nil {
		logger.Errorf("[Cluster-Add] Errors occurred during checking the connection. Cause: %+v", err)
		return nil, err
	}

	var c *model.Cluster
	if c, err = req.Cluster.Model(); err != nil {
		return nil, err
	}

	c.TenantID, _ = metadata.GetTenantID(ctx)
	c.SynchronizedAt = 0
	c.StateCode = constant.ClusterStateActive

	if err = db.Save(c).Error; err != nil {
		logger.Errorf("[Cluster-Add] Could not create the cluster. Cause: %+v", err)
		return nil, errors.UnusableDatabase(err)
	}

	// 처음 Cluster 를 생성할 때 etcd 에 config 값이 기본값으로 저장
	cfg := &config.Config{
		ClusterID:            c.ID,
		TimestampInterval:    5,
		ReservedSyncInterval: 1,
	}

	if err = cfg.Put(); err != nil {
		logger.Warnf("[Cluster-Add] Unable to put the config. Cause : ", err)
	}

	return Get(ctx, db, &cms.ClusterRequest{ClusterId: c.ID})
}

// Update 클러스터 수정
func Update(ctx context.Context, db *gorm.DB, req *cms.UpdateClusterRequest) (*cms.Cluster, error) {
	if err := validateUpdateClusterRequest(ctx, db, req); err != nil {
		logger.Errorf("[Cluster-Update] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	orig, err := getCluster(ctx, db, req.ClusterId)
	if err != nil {
		logger.Errorf("[Cluster-Update] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
		return nil, err
	}

	user, _ := metadata.GetAuthenticatedUser(ctx)
	if !internal.IsAdminUser(user) && !internal.IsGroupUser(user, orig.OwnerGroupID) {
		logger.Errorf("[Cluster-Update] Errors occurred during checking the authentication of the user.")
		return nil, errors.UnauthorizedRequest(ctx)
	}

	if err := checkUpdatableCluster(orig, req.Cluster); err != nil {
		logger.Errorf("[Cluster-Update] Errors occurred during checking updatable status of the cluster(%d). Cause:%+v", req.ClusterId, err)
		return nil, err
	}

	var c *model.Cluster
	if c, err = req.Cluster.Model(); err != nil {
		return nil, err
	}

	orig.Name = c.Name
	orig.Remarks = c.Remarks

	if err := db.Save(orig).Error; err != nil {
		logger.Errorf("[Cluster-Update] Could not update the cluster. Cause: %+v", err)
		return nil, errors.UnusableDatabase(err)
	}

	return Get(ctx, db, &cms.ClusterRequest{ClusterId: orig.ID})
}

func deleteSecurityGroup(db *gorm.DB, tid []uint64) error {
	var sgs []*model.ClusterSecurityGroup
	if err := db.Where("cluster_tenant_id in (?)", tid).Find(&sgs).Error; err != nil {
		return err
	}

	var sgid []uint64
	for _, sg := range sgs {
		sgid = append(sgid, sg.ID)

		var sgrs []*model.ClusterSecurityGroupRule
		if err := db.Where(&model.ClusterSecurityGroupRule{SecurityGroupID: sg.ID}).Find(&sgrs).Error; err != nil {
			return err
		}
		for _, sgr := range sgrs {
			// 해당 security group rule 을 쓰고 있는 shared task 삭제
			migrator.ClearSharedTaskById(sgr.ID, drConstant.MigrationTaskTypeCreateSecurityGroupRule)
		}

		// 해당 security group 을 쓰고 있는 shared task 삭제
		migrator.ClearSharedTaskById(sg.ID, drConstant.MigrationTaskTypeCreateSecurityGroup)
	}

	// Delete cluster instance security group
	if err := db.Where("cluster_security_group_id in (?)", sgid).Delete(&model.ClusterInstanceSecurityGroup{}).Error; err != nil {
		return err
	}

	// Delete cluster instance security group rule
	if err := db.Where("security_group_id in (?)", sgid).Delete(&model.ClusterSecurityGroupRule{}).Error; err != nil {
		return err
	}

	// Delete security groups
	if err := db.Where("cluster_tenant_id in (?)", tid).Delete(&model.ClusterSecurityGroup{}).Error; err != nil {
		return err
	}

	return nil
}

func deleteClusterNetwork(db *gorm.DB, tid []uint64) error {
	var ns []*model.ClusterNetwork
	if err := db.Where("cluster_tenant_id in (?)", tid).Find(&ns).Error; err != nil {
		return err
	}

	var nid []uint64
	for _, n := range ns {
		nid = append(nid, n.ID)

		// 해당 network 를 쓰고 있는 shared task 삭제
		migrator.ClearSharedTaskById(n.ID, drConstant.MigrationTaskTypeCreateNetwork)
	}

	var subs []*model.ClusterSubnet
	if err := db.Where("cluster_network_id in (?)", nid).Find(&subs).Error; err != nil {
		return err
	}

	var subid []uint64
	for _, sub := range subs {
		subid = append(subid, sub.ID)

		// 해당 subnet 을 쓰고 있는 shared task 삭제
		migrator.ClearSharedTaskById(sub.ID, drConstant.MigrationTaskTypeCreateSubnet)
	}

	// Delete cluster instance network
	if err := db.Where("cluster_network_id in (?)", nid).Delete(&model.ClusterInstanceNetwork{}).Error; err != nil {
		return err
	}

	// Delete cluster floatingIP
	if err := db.Where("cluster_network_id in (?)", nid).Delete(&model.ClusterFloatingIP{}).Error; err != nil {
		return err
	}

	// Delete cluster network routing interface
	if err := db.Where("cluster_subnet_id in (?)", subid).Delete(&model.ClusterNetworkRoutingInterface{}).Error; err != nil {
		return err
	}

	// Delete cluster subnet name server
	if err := db.Where("cluster_subnet_id in (?)", subid).Delete(&model.ClusterSubnetNameserver{}).Error; err != nil {
		return err
	}

	// Delete cluster subnet dhcp pools
	if err := db.Where("cluster_subnet_id in (?)", subid).Delete(&model.ClusterSubnetDHCPPool{}).Error; err != nil {
		return err
	}

	// Delete cluster subnet
	if err := db.Where("cluster_network_id in (?)", nid).Delete(&model.ClusterSubnet{}).Error; err != nil {
		return err
	}

	// Delete cluster network
	if err := db.Where("cluster_tenant_id in (?)", tid).Delete(&model.ClusterNetwork{}).Error; err != nil {
		return err
	}

	return nil
}

func deleteClusterRouter(db *gorm.DB, tid []uint64) error {
	var rs []*model.ClusterRouter
	if err := db.Where("cluster_tenant_id in (?)", tid).Find(&rs).Error; err != nil {
		return err
	}

	var rid []uint64
	for _, r := range rs {
		rid = append(rid, r.ID)

		// 해당 router 를 쓰고 있는 shared task 삭제
		migrator.ClearSharedTaskById(r.ID, drConstant.MigrationTaskTypeCreateRouter)
	}

	if err := db.Where("cluster_router_id in (?)", rid).Delete(&model.ClusterRouterExtraRoute{}).Error; err != nil {
		return err
	}

	if err := db.Where("id in (?)", rid).Delete(&model.ClusterRouter{}).Error; err != nil {
		return err
	}

	return nil
}

func deleteClusterStorage(db *gorm.DB, id uint64) error {
	var css []*model.ClusterStorage
	if err := db.Where("cluster_id = ?", id).Find(&css).Error; err != nil {
		return err
	}

	var sid []uint64
	for _, cs := range css {
		sid = append(sid, cs.ID)
	}

	// Delete related volume when cluster storages delete
	var vs []*model.ClusterVolume
	if err := db.Where("cluster_storage_id in (?)", sid).Find(&vs).Error; err != nil {
		return err
	}

	var vid []uint64
	for _, v := range vs {
		vid = append(vid, v.ID)
	}

	// Delete volume snapshot
	if err := db.Where("cluster_volume_id in (?)", vid).Delete(&model.ClusterVolumeSnapshot{}).Error; err != nil {
		return err
	}

	// Delete instance volume
	if err := db.Where("cluster_volume_id in (?)", vid).Delete(&model.ClusterInstanceVolume{}).Error; err != nil {
		return err
	}

	// Delete volume
	if err := db.Where("id in (?)", vid).Delete(&model.ClusterVolume{}).Error; err != nil {
		return err
	}

	// Delete storage
	if err := db.Where("id in (?)", sid).Delete(&model.ClusterStorage{}).Error; err != nil {
		return err
	}

	return nil
}

func deleteInstance(db *gorm.DB, tid []uint64, cid uint64) error {
	var specs []*model.ClusterInstanceSpec
	if err := db.Where("cluster_id = ?", cid).Find(&specs).Error; err != nil {
		return err
	}

	var specID []uint64
	for _, spec := range specs {
		specID = append(specID, spec.ID)

		// 해당 instance spec 를 쓰고 있는 shared task 삭제
		migrator.ClearSharedTaskById(spec.ID, drConstant.MigrationTaskTypeCreateSpec)
	}

	// delete instances
	if err := db.Where("cluster_tenant_id in (?)", tid).Delete(&model.ClusterInstance{}).Error; err != nil {
		return err
	}

	// delete instance extra specs
	if err := db.Where("cluster_instance_spec_id in (?)", specID).Delete(&model.ClusterInstanceExtraSpec{}).Error; err != nil {
		return err
	}

	// delete instance spec
	if err := db.Where("cluster_id = ?", cid).Delete(&model.ClusterInstanceSpec{}).Error; err != nil {
		return err
	}

	var keys []*model.ClusterKeypair
	if err := db.Where("cluster_id = ?", cid).Find(&keys).Error; err != nil {
		return err
	}
	for _, key := range keys {
		// 해당 keypair 를 쓰고 있는 shared task 삭제
		migrator.ClearSharedTaskById(key.ID, drConstant.MigrationTaskTypeCreateKeypair)
	}

	// delete instance keypair
	if err := db.Where("cluster_id = ?", cid).Delete(&model.ClusterKeypair{}).Error; err != nil {
		return err
	}

	return nil
}

func deleteRelation(db *gorm.DB, cid uint64) error {
	// Delete permissions
	if err := db.Where("cluster_id = ?", cid).Delete(&model.ClusterPermission{}).Error; err != nil {
		return err
	}

	if err := deleteClusterStorage(db, cid); err != nil {
		return err
	}

	var ts []*model.ClusterTenant
	if err := db.Where("cluster_id = ?", cid).Find(&ts).Error; err != nil {
		return err
	}

	var tid []uint64
	for _, t := range ts {
		tid = append(tid, t.ID)

		// 해당 tenant 를 쓰고 있는 shared task 삭제
		migrator.ClearSharedTaskById(t.ID, drConstant.MigrationTaskTypeCreateTenant)
	}

	if err := deleteSecurityGroup(db, tid); err != nil {
		return err
	}

	if err := deleteClusterNetwork(db, tid); err != nil {
		return err
	}

	if err := deleteInstance(db, tid, cid); err != nil {
		return err
	}

	if err := deleteClusterRouter(db, tid); err != nil {
		return err
	}

	// Delete quotas
	if err := db.Where("cluster_tenant_id in (?)", tid).Delete(&model.ClusterQuota{}).Error; err != nil {
		return err
	}

	// Delete tenants
	if err := db.Where("cluster_id = ?", cid).Delete(&model.ClusterTenant{}).Error; err != nil {
		return err
	}

	// Delete hypervisors
	if err := db.Where("cluster_id = ?", cid).Delete(&model.ClusterHypervisor{}).Error; err != nil {
		return err
	}

	// Delete availability zones
	if err := db.Where("cluster_id = ?", cid).Delete(&model.ClusterAvailabilityZone{}).Error; err != nil {
		return err
	}

	return nil
}

// Delete 클러스터 제거
func Delete(ctx context.Context, req *cms.DeleteClusterRequest) error {
	var c *model.Cluster
	var err error

	if req.GetClusterId() == 0 {
		err = errors.RequiredParameter("cluster_id")
		logger.Errorf("[Cluster-Delete] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		c, err = getCluster(ctx, db, req.ClusterId)
		return err
	}); err != nil {
		logger.Errorf("[Cluster-Delete] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
		return err
	}

	user, _ := metadata.GetAuthenticatedUser(ctx)
	if !internal.IsAdminUser(user) && !internal.IsGroupUser(user, c.OwnerGroupID) {
		logger.Errorf("[Cluster-Delete] Errors occurred during checking the authentication of the user.")
		return errors.UnauthorizedRequest(ctx)
	}

	//cli := drms.NewDisasterRecoveryManagerService(drConstant.ServiceManagerName, grpc.NewClient())
	//_, err = cli.CheckDeletableCluster(ctx, &drms.CheckDeletableClusterRequest{ClusterId: c.ID})
	//if err != nil {
	//	logger.Errorf("[Cluster-Delete] Errors occurred during checking deletable status of the cluster(%d). Cause:%+v", req.ClusterId, err)
	//	return errors.IPCFailed(err)
	//}

	if err = database.GormTransaction(func(db *gorm.DB) error {
		if err = deleteRelation(db, req.ClusterId); err != nil {
			return err
		}

		return db.Where(&model.Cluster{ID: req.ClusterId}).Delete(&model.Cluster{}).Error
	}); err != nil {
		logger.Errorf("[Cluster-Delete] Could not delete the relation. Cause: %+v", err)
		return errors.UnusableDatabase(err)
	}

	if err = DeleteClusters(req.ClusterId); err != nil {
		logger.Errorf("[Cluster-Delete] Could not delete the cluster in key-value store. Cause: %+v", err)
		return err
	}

	// 해당 cluster id 를 쓰고 있는 shared task 삭제
	migrator.ClearSharedTaskByClusterId(req.ClusterId)

	return nil
}

// ValidateSyncCluster 클러스터 동기화를 위한 유효성 확인
func ValidateSyncCluster(ctx context.Context, db *gorm.DB, cid uint64) (*model.Cluster, error) {
	var c *model.Cluster
	var err error

	if cid == 0 {
		err = errors.RequiredParameter("cluster_id")
		logger.Errorf("[Cluster-ValidateSyncCluster] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	if c, err = getCluster(ctx, db, cid); err != nil {
		logger.Errorf("[Cluster-ValidateSyncCluster] Could not get the cluster(%d). Cause: %+v", cid, err)
		return nil, err
	}

	user, _ := metadata.GetAuthenticatedUser(ctx)
	if !internal.IsAdminUser(user) && !internal.IsGroupUser(user, c.OwnerGroupID) {
		logger.Errorf("[Cluster-ValidateSyncCluster] Errors occurred during checking the authentication of the user.")
		return nil, errors.UnauthorizedRequest(ctx)
	}

	return c, nil
}
