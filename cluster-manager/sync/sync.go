package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	c "github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
	"sync"
	"time"
)

const (
	// ResourceTenantList 테넌트 목록 동기화
	ResourceTenantList = "cluster.tenant.list.sync"
	// ResourceTenant 테넌트 동기화
	ResourceTenant = "cluster.tenant.sync"

	// ResourceAvailabilityZoneList 가용구역 목록 동기화
	ResourceAvailabilityZoneList = "cluster.availability_zone.list.sync"

	// ResourceHypervisorList 하이퍼바이저 목록 동기화
	ResourceHypervisorList = "cluster.hypervisor.list.sync"

	// ResourceSecurityGroupList 보안 그룹 목록 동기화
	ResourceSecurityGroupList = "cluster.security_group.list.sync"

	// ResourceSecurityGroup 보안 그룹 동기화
	ResourceSecurityGroup = "cluster.security_group.sync"
	// ResourceSecurityGroupRule 보안 그룹 규칙 동기화
	ResourceSecurityGroupRule = "cluster.security_group_rule.sync"

	// ResourceNetworkList 네트워크 목록 동기화
	ResourceNetworkList = "cluster.network.list.sync"
	// ResourceNetwork 네트워크 동기화
	ResourceNetwork = "cluster.network.sync"

	// ResourceSubnet 서브넷 동기화
	ResourceSubnet = "cluster.subnet.sync"

	// ResourceRouterList 라우터 목록 동기화
	ResourceRouterList = "cluster.router.list.sync"
	// ResourceRouter 라우터 동기화
	ResourceRouter = "cluster.router.sync"

	// ResourceFloatingIP floating ip 동기화
	ResourceFloatingIP = "cluster.floating-ip.sync"

	// ResourceStorageList 스토리지 목록 동기화
	ResourceStorageList = "cluster.storage.list.sync"
	// ResourceStorage 볼륨 타입 동기화
	ResourceStorage = "cluster.storage.sync"

	// ResourceVolumeList 볼륨 목록 동기화
	ResourceVolumeList = "cluster.volume.list.sync"
	// ResourceVolume 볼륨 동기화
	ResourceVolume = "cluster.volume.sync"

	// ResourceVolumeSnapshotList 볼륨 스냅샷 목록 동기화
	ResourceVolumeSnapshotList = "cluster.volume-snapshot.list.sync"
	// ResourceVolumeSnapshot 볼륨 스냅샷 동기화
	ResourceVolumeSnapshot = "cluster.volume-snapshot.sync"

	// ResourceKeypairList keypair 목록 동기화
	ResourceKeypairList = "cluster.keypair.list.sync"
	// ResourceKeypair keypair 동기화
	ResourceKeypair = "cluster.keypair.sync"

	// ResourceInstanceList 인스턴스 목록 동기화
	ResourceInstanceList = "cluster.instance.list.sync"
	// ResourceInstance 인스턴스 동기화
	ResourceInstance = "cluster.instance.sync"

	// ResourceInstanceSpecList 인스턴스 스팩 목록 동기화
	ResourceInstanceSpecList = "cluster.instance.spec.list.sync"

	// Completed 동기화 완료 상태
	Completed = "cluster.sync.completed"
	// Failed 동기화 실패 상태
	Failed = "cluster.sync.failed"
	// Running 동기화 진행중인 상태
	Running = "cluster.sync.running"
)

var reservedSynchronizerMapLock = sync.Mutex{}
var reservedSynchronizerMap map[uint64]*ReservedSynchronizer

// ReservedSynchronizer 예약 동기화 구조체
type ReservedSynchronizer struct {
	availabilityZoneList bool
	hypervisorList       bool
	projects             map[string]bool
	instances            map[string]string
	keypairs             map[string]bool
	volumes              map[string]string
	volumeSnapshots      map[string]string
	volumeTypes          map[string]bool
	networks             map[string]string
	subnets              map[string]string
	securityGroups       map[string]string
	securityGroupRules   map[string]string
	routers              map[string]string
	floatingIPs          map[string]string
}

// Synchronizer sync 구조체
type Synchronizer struct {
	Cli client.Client

	clusterID   uint64
	clusterName string
}

// ReservedSync 는 예약된 클러스터의 동기화를 하는 함수이다.
func ReservedSync() {
	m := make(map[uint64]*ReservedSynchronizer)

	reservedSynchronizerMapLock.Lock()

	for k, v := range reservedSynchronizerMap {
		m[k] = v
	}

	reservedSynchronizerMap = nil

	reservedSynchronizerMapLock.Unlock()

	for clusterID, rs := range m {
		if err := rs.reservedSync(clusterID); err != nil {
			logger.Warnf("[ReservedSync] Could not sync cluster(%d) reserved for synchronization. Cause: %+v", clusterID, err)
		}
	}
}

func (rs *ReservedSynchronizer) reservedSync(clusterID uint64) error {
	logger.Debugf("starting reserved synchronization of cluster(%d).", clusterID)

	unlock, err := internal.ClusterDBLock(clusterID)
	if err != nil {
		logger.Errorf("[reservedSync] Could not cluster lock (%d). Cause: %+v", clusterID, err)
		return err
	}

	defer unlock()

	now := time.Now()
	if err = c.SetStatus(clusterID, c.SyncStatus{StatusCode: constant.ClusterSyncStateCodeInit, StartedAt: now}); err != nil {
		logger.Warnf("[reservedSync] Could not set sync status(init): cluster(%d). Cause: %+v", clusterID, err)
	}

	var s *Synchronizer
	if s, err = NewSynchronizer(clusterID); err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if err = c.SetStatus(clusterID, c.SyncStatus{StatusCode: constant.ClusterSyncStateCodeFailed}); err != nil {
				logger.Warnf("[reservedSync] Could not set sync status(failed): cluster(%d). Cause: %+v", clusterID, err)
			}

			if err = s.Close(); err != nil {
				logger.Warnf("[reservedSync] Could not close synchronizer client. Cause: %+v", err)
			}
		}
	}()

	if err = c.SetStatus(clusterID, c.SyncStatus{StatusCode: constant.ClusterSyncStateCodeRunning}); err != nil {
		logger.Warnf("[reservedSync] Could not set sync status(running): cluster(%d). Cause: %+v", clusterID, err)
	}

	var failedReason []*c.Message

	for uuid := range rs.projects {
		if err := s.SyncClusterTenant(&model.ClusterTenant{UUID: uuid}); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceTenant))
			Reserve(clusterID, ResourceTenant, uuid)
			logger.Warnf("[reservedSync] Could not sync cluster(%d) tenant(%s). Cause: %+v", clusterID, uuid, err)
		}
	}

	if rs.availabilityZoneList {
		if err := s.SyncClusterAvailabilityZoneList(); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceAvailabilityZoneList))
			Reserve(clusterID, ResourceAvailabilityZoneList, "")
			logger.Warnf("[reservedSync] Could not sync cluster(%d) availability zone list. Cause: %+v", clusterID, err)
		}
		return nil
	}

	if rs.hypervisorList {
		if err := s.SyncClusterHypervisorList(); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceHypervisorList))
			Reserve(clusterID, ResourceHypervisorList, "")
			logger.Warnf("[reservedSync] Could not sync cluster(%d) hypervisor list. Cause: %+v", clusterID, err)
		}
		return nil
	}

	for uuid, tenantUUID := range rs.securityGroups {
		if err := s.SyncClusterSecurityGroup(&model.ClusterSecurityGroup{UUID: uuid}, TenantUUID(tenantUUID)); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceSecurityGroup))
			Reserve(clusterID, ResourceSecurityGroup, uuid, TenantUUID(tenantUUID))
			logger.Warnf("[reservedSync] Could not sync cluster(%d) security group(%s). Cause: %+v", clusterID, uuid, err)
		}
	}

	for uuid, tenantUUID := range rs.securityGroupRules {
		if err := s.SyncClusterSecurityGroupRule(&model.ClusterSecurityGroupRule{UUID: uuid}, TenantUUID(tenantUUID)); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceSecurityGroupRule))
			Reserve(clusterID, ResourceSecurityGroupRule, uuid, TenantUUID(tenantUUID))
			logger.Warnf("[reservedSync] Could not sync cluster(%d) security group rule(%s). Cause: %+v", clusterID, uuid, err)
		}
	}

	for uuid, tenantUUID := range rs.networks {
		if err := s.SyncClusterNetwork(&model.ClusterNetwork{UUID: uuid}, TenantUUID(tenantUUID)); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceNetwork))
			Reserve(clusterID, ResourceNetwork, uuid, TenantUUID(tenantUUID))
			logger.Warnf("[reservedSync] Could not sync cluster(%d) network(%s). Cause: %+v", clusterID, uuid, err)
		}
	}

	for uuid, tenantUUID := range rs.subnets {
		if err := s.SyncClusterSubnet(&model.ClusterSubnet{UUID: uuid}, TenantUUID(tenantUUID)); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceSubnet))
			Reserve(clusterID, ResourceSubnet, uuid, TenantUUID(tenantUUID))
			logger.Warnf("[reservedSync] Could not sync cluster(%d) subnet(%s). Cause: %+v", clusterID, uuid, err)
		}
	}

	for uuid, tenantUUID := range rs.routers {
		if err := s.SyncClusterRouter(&model.ClusterRouter{UUID: uuid}, TenantUUID(tenantUUID)); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceRouter))
			Reserve(clusterID, ResourceRouter, uuid, TenantUUID(tenantUUID))
			logger.Warnf("[reservedSync] Could not sync cluster(%d) router(%s). Cause: %+v", clusterID, uuid, err)
		}
	}

	for uuid, tenantUUID := range rs.floatingIPs {
		if err := s.SyncClusterFloatingIP(&model.ClusterFloatingIP{UUID: uuid}, TenantUUID(tenantUUID)); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceFloatingIP))
			Reserve(clusterID, ResourceFloatingIP, uuid, TenantUUID(tenantUUID))
			logger.Warnf("[reservedSync] Could not sync cluster(%d) floating IP(%s). Cause: %+v", clusterID, uuid, err)
		}
	}

	for uuid := range rs.volumeTypes {
		if err := s.SyncClusterStorage(&model.ClusterStorage{UUID: uuid}); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceStorage))
			Reserve(clusterID, ResourceStorage, uuid)
			logger.Warnf("[reservedSync] Could not sync cluster(%d) storage(%s). Cause: %+v", clusterID, uuid, err)
		}
	}

	for uuid, tenantUUID := range rs.volumes {
		if err := waitVolumeStatus(s.Cli, uuid); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceVolume))
			Reserve(clusterID, ResourceVolume, uuid, TenantUUID(tenantUUID))
			logger.Warnf("[reservedSync] Could not sync cluster(%d) volume(%s). Cause: %+v", clusterID, uuid, err)
			continue
		}

		if err := s.SyncClusterVolume(&model.ClusterVolume{UUID: uuid}, TenantUUID(tenantUUID)); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceVolume))
			Reserve(clusterID, ResourceVolume, uuid, TenantUUID(tenantUUID))
			logger.Warnf("[reservedSync] Could not sync cluster(%d) volume(%s). Cause: %+v", clusterID, uuid, err)
		}
	}

	for uuid, tenantUUID := range rs.volumeSnapshots {
		if err := s.SyncClusterVolumeSnapshot(&model.ClusterVolumeSnapshot{UUID: uuid}, TenantUUID(tenantUUID)); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceVolumeSnapshot))
			Reserve(clusterID, ResourceVolumeSnapshot, uuid, TenantUUID(tenantUUID))
			logger.Warnf("[reservedSync] Could not sync cluster(%d) volume snapshot(%s). Cause: %+v", clusterID, uuid, err)
		}
	}

	for uuid := range rs.keypairs {
		if err := s.SyncClusterKeypair(&model.ClusterKeypair{Name: uuid}); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceKeypair))
			Reserve(clusterID, ResourceKeypair, uuid)
			logger.Warnf("[reservedSync] Could not sync cluster(%d) keypair(%s). Cause: %+v", clusterID, uuid, err)
		}
	}

	for uuid, tenantUUID := range rs.instances {
		if err = waitInstanceStatus(s.Cli, uuid); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceInstance))
			Reserve(clusterID, ResourceInstance, uuid, TenantUUID(tenantUUID))
			logger.Warnf("[reservedSync] Could not sync cluster(%d) instance(%s). Cause: %+v", clusterID, uuid, err)
			continue
		}

		if err := s.SyncClusterInstance(&model.ClusterInstance{UUID: uuid}, TenantUUID(tenantUUID)); err != nil {
			failedReason = append(failedReason, c.MakeFailedMessage(ResourceInstance))
			Reserve(clusterID, ResourceInstance, uuid, TenantUUID(tenantUUID))
			logger.Warnf("[reservedSync] Could not sync cluster(%d) instance(%s). Cause: %+v", clusterID, uuid, err)
		}
	}

	if len(failedReason) != 0 {
		if err := c.SetStatus(clusterID, c.SyncStatus{StatusCode: constant.ClusterSyncStateCodeFailed, Reasons: failedReason}); err != nil {
			logger.Warnf("[reservedSync] Could not set sync status(failed): cluster(%d). Cause: %+v", clusterID, err)
		}
	} else {
		if err := c.SetStatus(clusterID, c.SyncStatus{StatusCode: constant.ClusterSyncStateCodeDone}); err != nil {
			logger.Warnf("[reservedSync] Could not set sync status(done): cluster(%d). Cause: %+v", clusterID, err)
		}
	}

	logger.Debugf("reserved synchronization of cluster(%d) end", clusterID)

	return nil
}

// Reserve 동기화 예약을 위한 함수
// 삭제 이벤트에 대한 동기화 예약시에는 tenantUUID 를 같이 입력(project, volume type 은 제외)
// 삭제 이벤트가 아닌 경우에는 tenantUUID ""로 입력
func Reserve(clusterID uint64, resourceType, uuid string, opts ...Option) {
	reservedSynchronizerMapLock.Lock()
	defer reservedSynchronizerMapLock.Unlock()

	if reservedSynchronizerMap == nil {
		reservedSynchronizerMap = make(map[uint64]*ReservedSynchronizer)
	}

	if _, ok := reservedSynchronizerMap[clusterID]; !ok {
		reservedSynchronizerMap[clusterID] = &ReservedSynchronizer{
			availabilityZoneList: false,
			hypervisorList:       false,
			projects:             make(map[string]bool),
			instances:            make(map[string]string),
			keypairs:             make(map[string]bool),
			volumes:              make(map[string]string),
			volumeSnapshots:      make(map[string]string),
			volumeTypes:          make(map[string]bool),
			networks:             make(map[string]string),
			subnets:              make(map[string]string),
			securityGroups:       make(map[string]string),
			securityGroupRules:   make(map[string]string),
			routers:              make(map[string]string),
			floatingIPs:          make(map[string]string),
		}
	}

	reservedSynchronizerMap[clusterID].reserve(resourceType, uuid, opts...)
}

func (rs *ReservedSynchronizer) reserve(resourceType, uuid string, opts ...Option) {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	if uuid == "" {
		// resourceType 이 ResourceAvailabilityZoneList, ResourceHypervisorList 경우에는 uuid 값이 ""인데
		// 그 외의 경우에는 비정상적인 요청으로 판단
		if resourceType != ResourceAvailabilityZoneList && resourceType != ResourceHypervisorList {
			return
		}
	}

	switch resourceType {
	case ResourceAvailabilityZoneList:
		rs.availabilityZoneList = true

	case ResourceHypervisorList:
		rs.hypervisorList = true

	case ResourceTenant:
		rs.projects[uuid] = true

	case ResourceInstance:
		rs.instances[uuid] = options.TenantUUID

	case ResourceKeypair:
		rs.keypairs[uuid] = true

	case ResourceVolume:
		rs.volumes[uuid] = options.TenantUUID

	case ResourceVolumeSnapshot:
		rs.volumeSnapshots[uuid] = options.TenantUUID

	case ResourceStorage:
		rs.volumeTypes[uuid] = true

	case ResourceNetwork:
		rs.networks[uuid] = options.TenantUUID

	case ResourceSubnet:
		rs.subnets[uuid] = options.TenantUUID

	case ResourceSecurityGroup:
		rs.securityGroups[uuid] = options.TenantUUID

	case ResourceSecurityGroupRule:
		rs.securityGroupRules[uuid] = options.TenantUUID

	case ResourceRouter:
		rs.routers[uuid] = options.TenantUUID

	case ResourceFloatingIP:
		rs.floatingIPs[uuid] = options.TenantUUID
	}
}

// Close 동기화에 사용되는 오픈스택 연결 종료 함수
func (s *Synchronizer) Close() error {
	return s.Cli.Close()
}

// NewSynchronizer 동기화 구조체 생성 함수
func NewSynchronizer(clusterID uint64) (*Synchronizer, error) {
	var cluster model.Cluster
	err := database.Execute(func(db *gorm.DB) error {
		return db.First(&cluster, &model.Cluster{ID: clusterID}).Error
	})
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundCluster(clusterID)
	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	cluster.Credential, err = client.DecryptCredentialPassword(cluster.Credential)
	if err != nil {
		return nil, err
	}

	cli, err := client.New(cluster.TypeCode, cluster.APIServerURL, cluster.Credential, "")
	if err != nil {
		return nil, err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[NewSynchronizer] Could not connect. Cause: %v", err)
		return nil, err
	}

	if err = cli.CheckAuthority(); err != nil {
		return nil, err
	}

	return &Synchronizer{
		Cli:         cli,
		clusterID:   clusterID,
		clusterName: cluster.Name,
	}, nil
}
