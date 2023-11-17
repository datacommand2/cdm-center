package internal

import "github.com/datacommand2/cdm-cloud/common/errors"

var (
	// ErrNotFoundGroup 사용자 그룹을 찾을 수 없음
	ErrNotFoundGroup = errors.New("not found group")

	// ErrNotFoundCluster 클러스터를 찾을 수 없음
	ErrNotFoundCluster = errors.New("not found cluster")

	// ErrNotFoundClusterHypervisor 클러스터 노드를 찾을 수 없음
	ErrNotFoundClusterHypervisor = errors.New("not found cluster hypervisor")

	// ErrNotFoundClusterNetwork id 를 이용해 클러스터 네트워크를 찾을 수 없음
	ErrNotFoundClusterNetwork = errors.New("not found cluster network")

	// ErrNotFoundClusterNetworkByUUID uuid 를 이용해 클러스터 네트워크를 찾을 수 없음
	ErrNotFoundClusterNetworkByUUID = errors.New("not found cluster network by uuid")

	// ErrNotFoundClusterSubnet 클러스터 서브넷을 찾을 수 없음
	ErrNotFoundClusterSubnet = errors.New("not found cluster subnet")

	// ErrNotFoundClusterFloatingIP 클러스터 Floating IP 를 찾을 수 없음
	ErrNotFoundClusterFloatingIP = errors.New("not found cluster floating ip")

	// ErrNotFoundClusterTenant id 를 이용해 클러스터 테넌트를 찾을 수 없음
	ErrNotFoundClusterTenant = errors.New("not found cluster tenant")

	// ErrNotFoundClusterTenantByUUID uuid 를 이용해 클러스터 테넌트를 찾을 수 없음
	ErrNotFoundClusterTenantByUUID = errors.New("not found cluster tenant by uuid")

	// ErrNotFoundClusterInstance id 를 이용해 클러스터 인스턴스를 찾을 수 없음
	ErrNotFoundClusterInstance = errors.New("not found cluster instance")

	// ErrNotFoundClusterInstanceByUUID uuid 를 이용해 클러스터 인스턴스를 찾을 수 없음
	ErrNotFoundClusterInstanceByUUID = errors.New("not found cluster instance by uuid")

	// ErrNotFoundClusterInstanceSpec id 를 이용해 클러스터 인스턴스 Spec 을 찾을 수 없음
	ErrNotFoundClusterInstanceSpec = errors.New("not found cluster instance spec")

	// ErrNotFoundClusterInstanceSpecByUUID uuid 를 이용해 클러스터 인스턴스 Spec 을 찾을 수 없음
	ErrNotFoundClusterInstanceSpecByUUID = errors.New("not found cluster instance spec by uuid")

	// ErrNotFoundClusterKeyPair 클러스터 KeyPair 를 찾을 수 없음
	ErrNotFoundClusterKeyPair = errors.New("not found cluster keypair")

	// ErrUnsupportedClusterType 지원하지 않는 클러스터 유형
	ErrUnsupportedClusterType = errors.New("unsupported cluster type")

	// ErrNotFoundClusterAvailabilityZone 클러스터 가용 구역을 찾을 수 없음
	ErrNotFoundClusterAvailabilityZone = errors.New("not found cluster availability zone")

	// ErrNotFoundClusterVolume id 를 이용해 클러스터 볼륨을 찾을 수 없음
	ErrNotFoundClusterVolume = errors.New("not found cluster volume")

	// ErrNotFoundClusterVolumeByUUID uuid 를 이용해 클러스터 볼륨을 찾을 수 없음
	ErrNotFoundClusterVolumeByUUID = errors.New("not found cluster volume by uuid")

	// ErrNotFoundClusterStorage 클러스터 볼륨타입을 찾을 수 없음
	ErrNotFoundClusterStorage = errors.New("not found cluster storage")

	// ErrNotFoundRouterExternalNetwork 클러스터 라우터의 연관된 외부 네트워크를 찾을 수 없음
	ErrNotFoundRouterExternalNetwork = errors.New("not found relational external network")

	// ErrNotFoundClusterRouter id 를 이용해 클러스터 라우터를 찾을 수 없음
	ErrNotFoundClusterRouter = errors.New("not found cluster router")

	// ErrNotFoundClusterRouterByUUID uuid 를 이용해 클러스터 라우터를 찾을 수 없음
	ErrNotFoundClusterRouterByUUID = errors.New("not found cluster router by uuid")

	// ErrNotFoundClusterSecurityGroup id 를 이용해 클러스터 보안그룹을 찾을 수 없음
	ErrNotFoundClusterSecurityGroup = errors.New("not found cluster security group")

	// ErrNotFoundClusterSecurityGroupByUUID uuid 를 이용해 클러스터 보안그룹을 찾을 수 없음
	ErrNotFoundClusterSecurityGroupByUUID = errors.New("not found cluster security group by uuid")

	// ErrNotFoundClusterSyncStatus 클러스터 동기화 상태를 찾을 수 없음
	ErrNotFoundClusterSyncStatus = errors.New("not found cluster sync status")

	// ErrCurrentPasswordMismatch 클러스터의 현재 패스워드가 일치하지 않음
	ErrCurrentPasswordMismatch = errors.New("current password mismatch")

	// ErrUnableSynchronizeStatus 클러스터가 동기화 할 수 없는 상태
	ErrUnableSynchronizeStatus = errors.New("unable to synchronize because of status")

	// ErrNotFoundClusterException Network, Storage, Compute 상태중에 제외할 값을 찾을 수 없음
	ErrNotFoundClusterException = errors.New("not found cluster sync exception")
)

// NotFoundGroup 사용자 그룹을 찾을 수 없음
func NotFoundGroup(id uint64) error {
	return errors.Wrap(
		ErrNotFoundGroup,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"group_id": id,
		}),
	)
}

// NotFoundCluster 클러스터를 찾을 수 없음
func NotFoundCluster(id uint64) error {
	return errors.Wrap(
		ErrNotFoundCluster,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_id": id,
		}),
	)
}

// NotFoundClusterHypervisor 클러스터 노드를 찾을 수 없음
func NotFoundClusterHypervisor(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterHypervisor,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_hypervisor_id": id,
		}),
	)
}

// NotFoundClusterTenant ID를 이용해 클러스터 테넌트를 찾을 수 없음
func NotFoundClusterTenant(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterTenant,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_tenant_id": id,
		}),
	)
}

// NotFoundClusterTenantByUUID UUID 를 이용해 클러스터 테넌트를 찾을 수 없음
func NotFoundClusterTenantByUUID(uuid string) error {
	return errors.Wrap(
		ErrNotFoundClusterTenantByUUID,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_tenant_uuid": uuid,
		}),
	)
}

// NotFoundClusterInstance ID를 이용해 클러스터 인스턴스를 찾을 수 없음
func NotFoundClusterInstance(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterInstance,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_instance_id": id,
		}),
	)
}

// NotFoundClusterInstanceByUUID UUID 를 이용해 클러스터 인스턴스를 찾을 수 없음
func NotFoundClusterInstanceByUUID(uuid string) error {
	return errors.Wrap(
		ErrNotFoundClusterInstanceByUUID,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_instance_uuid": uuid,
		}),
	)
}

// NotFoundClusterInstanceSpec ID를 이용해 클러스터 인스턴스 Spec 을 찾을 수 없음
func NotFoundClusterInstanceSpec(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterInstanceSpec,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_instance_spec_id": id,
		}),
	)
}

// NotFoundClusterInstanceSpecByUUID UUID 를 이용해 클러스터 인스턴스 Spec 을 찾을 수 없음
func NotFoundClusterInstanceSpecByUUID(uuid string) error {
	return errors.Wrap(
		ErrNotFoundClusterInstanceSpecByUUID,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_instance_spec_uuid": uuid,
		}),
	)
}

// NotFoundClusterKeyPair 클러스터 KeyPair 를 찾을 수 없음
func NotFoundClusterKeyPair(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterKeyPair,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_keypair_id": id,
		}),
	)
}

// UnsupportedClusterType 지원하지 않는 클러스터 유형
func UnsupportedClusterType(code string) error {
	return errors.Wrap(
		ErrUnsupportedClusterType,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"type_code": code,
		}),
	)
}

// NotFoundRelationalExternalNetwork 클러스터 라우터의 외부 네트워크를 찾을 수 없음
func NotFoundRelationalExternalNetwork(id uint64) error {
	return errors.Wrap(
		ErrNotFoundRouterExternalNetwork,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_network_id": id,
		}),
	)
}

// NotFoundClusterNetwork ID 를 이용해 클러스터 네트워크를 찾을 수 없음
func NotFoundClusterNetwork(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterNetwork,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_network_id": id,
		}),
	)
}

// NotFoundClusterNetworkByUUID UUID 를 이용해 클러스터 네트워크를 찾을 수 없음
func NotFoundClusterNetworkByUUID(uuid string) error {
	return errors.Wrap(
		ErrNotFoundClusterNetworkByUUID,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_network_uuid": uuid,
		}),
	)
}

// NotFoundClusterSubnet 클러스터 서브넷을 찾을 수 없음
func NotFoundClusterSubnet(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterSubnet,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_subnet_id": id,
		}),
	)
}

// NotFoundClusterFloatingIP 클러스터 Floating IP 를 찾을 수 없음
func NotFoundClusterFloatingIP(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterFloatingIP,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_floating_ip_id": id,
		}),
	)
}

// NotFoundClusterAvailabilityZone 클러스터 가용 구역을 찾을 수 없음
func NotFoundClusterAvailabilityZone(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterAvailabilityZone,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_availability_zone_id": id,
		}),
	)
}

// NotFoundClusterVolume ID 를 이용해 볼륨을 찾을 수 없음
func NotFoundClusterVolume(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterVolume,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_volume_id": id,
		}),
	)
}

// NotFoundClusterVolumeByUUID UUID 를 이용해 볼륨을 찾을 수 없음
func NotFoundClusterVolumeByUUID(uuid string) error {
	return errors.Wrap(
		ErrNotFoundClusterVolumeByUUID,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_volume_uuid": uuid,
		}),
	)
}

// NotFoundClusterStorage 볼륨타입을 찾을 수 없음
func NotFoundClusterStorage(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterStorage,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_storage_id": id,
		}),
	)
}

// NotFoundClusterRouter ID 를 이용해 클러스터 라우터를 찾을 수 없음
func NotFoundClusterRouter(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterRouter,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_router_id": id,
		}),
	)
}

// NotFoundClusterRouterByUUID UUID 를 이용해 클러스터 라우터를 찾을 수 없음
func NotFoundClusterRouterByUUID(uuid string) error {
	return errors.Wrap(
		ErrNotFoundClusterRouterByUUID,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_router_uuid": uuid,
		}),
	)
}

// NotFoundClusterSecurityGroup ID 를 이용해 클러스터 보안그룹을 찾을 수 없음
func NotFoundClusterSecurityGroup(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterSecurityGroup,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_security_group_id": id,
		}),
	)
}

// NotFoundClusterSecurityGroupByUUID UUID 를 이용해 클러스터 보안그룹을 찾을 수 없음
func NotFoundClusterSecurityGroupByUUID(uuid string) error {
	return errors.Wrap(
		ErrNotFoundClusterSecurityGroupByUUID,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_security_group_uuid": uuid,
		}),
	)
}

// NotFoundClusterSyncStatus 클러스터 동기화 상태를 찾을 수 없음
func NotFoundClusterSyncStatus(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterSyncStatus,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_id": id,
		}),
	)
}

// CurrentPasswordMismatch 클러스터의 현재 패스워드가 일치하지 않음
func CurrentPasswordMismatch(id uint64) error {
	return errors.Wrap(
		ErrCurrentPasswordMismatch,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_id": id,
		}),
	)
}

// UnableSynchronizeStatus 클러스터 동기화할 수 없는 상태
func UnableSynchronizeStatus(id uint64, status string) error {
	return errors.Wrap(
		ErrUnableSynchronizeStatus,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_id": id,
			"status":     status,
		}),
	)
}

// NotFoundClusterException Network, Storage, Compute 상태중에 제외할 값을 찾을 수 없음
func NotFoundClusterException(id uint64) error {
	return errors.Wrap(
		ErrNotFoundClusterException,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_id": id,
		}),
	)
}
