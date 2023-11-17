package constant

const (
	// TopicNoticeClusterCreated 클러스터 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <ClusterID>
	TopicNoticeClusterCreated = "cdm.center.notice.cluster.created"

	// TopicNoticeClusterUpdated 클러스터 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <ClusterID>
	TopicNoticeClusterUpdated = "cdm.center.notice.cluster.updated"

	// TopicNoticeClusterDeleted 클러스터 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <ClusterID>
	TopicNoticeClusterDeleted = "cdm.center.notice.cluster.deleted"

	// TopicClusterBecomeActivated 클러스터의 활성화를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <ClusterID>
	TopicClusterBecomeActivated = "cdm.center.notice.cluster.activated"

	// TopicClusterBecomeInactivated 클러스터의 비활성화를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <ClusterID>
	TopicClusterBecomeInactivated = "cdm.center.notice.cluster.inactivated"

	// TopicNoticeCenterConfigUpdated center 의 config 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <ClusterID>
	TopicNoticeCenterConfigUpdated = "cdm.center.notice.config.updated"

	// TopicClusterServiceStatus 클러스터 서비스들의 상태를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.ClusterServiceStatus>
	// `status.` 뒤에 cluster id를 붙여줘야한다.
	TopicClusterServiceStatus = "cdm.center.notice.status.%d"
)

const (
	// QueueClusterServiceGetStatus service 의 status 를 가져오는 Queue 의 이름
	QueueClusterServiceGetStatus = "cdm.center.service.cluster.status"
	// QueueClusterServiceException service 의 상태를 제외하는 Queue 의 이름
	QueueClusterServiceException = "cdm.center.service.cluster.exception"
	// QueueClusterServiceDelete 클러스터 삭제됐을때 메모리 데이터도 삭제를 하는 Queue 의 이름
	QueueClusterServiceDelete = "cdm.center.service.cluster.delete"
)

const (
	// QueueNoticeClusterSyncStatus 클러스터 동기화 상태를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterTenant>
	QueueNoticeClusterSyncStatus = "cdm.center.notice.cluster.sync.status"

	// QueueNoticeClusterTenantCreated 클러스터 테넌트 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterTenant>
	QueueNoticeClusterTenantCreated = "cdm.center.notice.cluster.tenant.created"

	// QueueNoticeClusterTenantUpdated 클러스터 테넌트 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterTenant>
	QueueNoticeClusterTenantUpdated = "cdm.center.notice.cluster.tenant.updated"

	// QueueNoticeClusterTenantDeleted 클러스터 테넌트 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterTenant>
	QueueNoticeClusterTenantDeleted = "cdm.center.notice.cluster.tenant.deleted"

	// QueueNoticeClusterAvailabilityZoneCreated 클러스터 가용 구역 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterAvailabilityZone>
	QueueNoticeClusterAvailabilityZoneCreated = "cdm.center.notice.cluster.availability-zone.created"

	// QueueNoticeClusterAvailabilityZoneUpdated 클러스터 가용 구역 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterAvailabilityZone>
	QueueNoticeClusterAvailabilityZoneUpdated = "cdm.center.notice.cluster.availability-zone.updated"

	// QueueNoticeClusterAvailabilityZoneDeleted 클러스터 가용 구역 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterAvailabilityZone>
	QueueNoticeClusterAvailabilityZoneDeleted = "cdm.center.notice.cluster.availability-zone.deleted"

	// QueueNoticeClusterHypervisorCreated 클러스터 하이퍼바이저 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterHypervisor>
	QueueNoticeClusterHypervisorCreated = "cdm.center.notice.cluster.hypervisor.created"

	// QueueNoticeClusterHypervisorUpdated 클러스터 하이퍼바이저 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterHypervisor>
	QueueNoticeClusterHypervisorUpdated = "cdm.center.notice.cluster.hypervisor.updated"

	// QueueNoticeClusterHypervisorDeleted 클러스터 하이퍼바이저 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterHypervisor>
	QueueNoticeClusterHypervisorDeleted = "cdm.center.notice.cluster.hypervisor.deleted"

	// QueueNoticeClusterSecurityGroupCreated 클러스터 보안 그룹 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterSecurityGroup>
	QueueNoticeClusterSecurityGroupCreated = "cdm.center.notice.cluster.security-group.created"

	// QueueNoticeClusterSecurityGroupUpdated 클러스터 보안 그룹 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterSecurityGroup>
	QueueNoticeClusterSecurityGroupUpdated = "cdm.center.notice.cluster.security-group.updated"

	// QueueNoticeClusterSecurityGroupDeleted 클러스터 보안 그룹 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterSecurityGroup>
	QueueNoticeClusterSecurityGroupDeleted = "cdm.center.notice.cluster.security-group.deleted"

	// QueueNoticeClusterSecurityGroupRuleCreated 클러스터 보안 그룹 규칙 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterSecurityGroupRule>
	QueueNoticeClusterSecurityGroupRuleCreated = "cdm.center.notice.cluster.security-group-rule.created"

	// QueueNoticeClusterSecurityGroupRuleUpdated 클러스터 보안 그룹 규칙 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterSecurityGroupRule>
	QueueNoticeClusterSecurityGroupRuleUpdated = "cdm.center.notice.cluster.security-group-rule.updated"

	// QueueNoticeClusterSecurityGroupRuleDeleted 클러스터 보안 그룹 규칙 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterSecurityGroupRule>
	QueueNoticeClusterSecurityGroupRuleDeleted = "cdm.center.notice.cluster.security-group-rule.deleted"

	// QueueNoticeClusterNetworkCreated 클러스터 네트워크 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterNetwork>
	QueueNoticeClusterNetworkCreated = "cdm.center.notice.cluster.network.created"

	// QueueNoticeClusterNetworkUpdated 클러스터 네트워크 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterNetwork>
	QueueNoticeClusterNetworkUpdated = "cdm.center.notice.cluster.network.updated"

	// QueueNoticeClusterNetworkDeleted 클러스터 네트워크 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterNetwork>
	QueueNoticeClusterNetworkDeleted = "cdm.center.notice.cluster.network.deleted"

	// QueueNoticeClusterFloatingIPCreated 클러스터 FloatingIP 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterFloatingIP>
	QueueNoticeClusterFloatingIPCreated = "cdm.center.notice.cluster.floating-ip.created"

	// QueueNoticeClusterFloatingIPUpdated 클러스터 FloatingIP 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterFloatingIP>
	QueueNoticeClusterFloatingIPUpdated = "cdm.center.notice.cluster.floating-ip.updated"

	// QueueNoticeClusterFloatingIPDeleted 클러스터 FloatingIP 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterFloatingIP>
	QueueNoticeClusterFloatingIPDeleted = "cdm.center.notice.cluster.floating-ip.deleted"

	// QueueNoticeClusterSubnetCreated 클러스터 Subnet 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterSubnet>
	QueueNoticeClusterSubnetCreated = "cdm.center.notice.cluster.subnet.created"

	// QueueNoticeClusterSubnetUpdated 클러스터 Subnet 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterSubnet>
	QueueNoticeClusterSubnetUpdated = "cdm.center.notice.cluster.subnet.updated"

	// QueueNoticeClusterSubnetDeleted 클러스터 Subnet 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterSubnet>
	QueueNoticeClusterSubnetDeleted = "cdm.center.notice.cluster.subnet.deleted"

	// QueueNoticeClusterRouterCreated 클러스터 라우터 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterRouter>
	QueueNoticeClusterRouterCreated = "cdm.center.notice.cluster.router.created"

	// QueueNoticeClusterRouterUpdated 클러스터 라우터 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterRouter>
	QueueNoticeClusterRouterUpdated = "cdm.center.notice.cluster.router.updated"

	// QueueNoticeClusterRouterDeleted 클러스터 라우터 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterRouter>
	QueueNoticeClusterRouterDeleted = "cdm.center.notice.cluster.router.deleted"

	// QueueNoticeClusterRoutingInterfaceCreated 클러스터 라우팅 인터페이스 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterRoutingInterface>
	QueueNoticeClusterRoutingInterfaceCreated = "cdm.center.notice.cluster.routing-interface.created"

	// QueueNoticeClusterRoutingInterfaceUpdated 클러스터 라우팅 인터페이스 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterRoutingInterface>
	QueueNoticeClusterRoutingInterfaceUpdated = "cdm.center.notice.cluster.routing-interface.updated"

	// QueueNoticeClusterRoutingInterfaceDeleted 클러스터 라우팅 인터페이스 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterRoutingInterface>
	QueueNoticeClusterRoutingInterfaceDeleted = "cdm.center.notice.cluster.routing-interface.deleted"

	// QueueNoticeClusterStorageCreated 클러스터 스토리지 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterStorage>
	QueueNoticeClusterStorageCreated = "cdm.center.notice.cluster.storage.created"

	// QueueNoticeClusterStorageUpdated 클러스터 스토리지 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterStorage>
	QueueNoticeClusterStorageUpdated = "cdm.center.notice.cluster.storage.updated"

	// QueueNoticeClusterStorageDeleted 클러스터 스토리지 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterStorage>
	QueueNoticeClusterStorageDeleted = "cdm.center.notice.cluster.storage.deleted"

	// QueueNoticeClusterVolumeCreated 클러스터 볼륨 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterVolume>
	QueueNoticeClusterVolumeCreated = "cdm.center.notice.cluster.volume.created"

	// QueueNoticeClusterVolumeUpdated 클러스터 볼륨 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterVolume>
	QueueNoticeClusterVolumeUpdated = "cdm.center.notice.cluster.volume.updated"

	// QueueNoticeClusterVolumeDeleted 클러스터 볼륨 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterVolume>
	QueueNoticeClusterVolumeDeleted = "cdm.center.notice.cluster.volume.deleted"

	// QueueNoticeClusterVolumeSnapshotCreated 클러스터 볼륨 스냅샷 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterVolumeSnapshot>
	QueueNoticeClusterVolumeSnapshotCreated = "cdm.center.notice.cluster.volume-snapshot.created"

	// QueueNoticeClusterVolumeSnapshotUpdated 클러스터 볼륨 스냅샷 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterVolumeSnapshot>
	QueueNoticeClusterVolumeSnapshotUpdated = "cdm.center.notice.cluster.volume-snapshot.updated"

	// QueueNoticeClusterVolumeSnapshotDeleted 클러스터 볼륨 스냅샷 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterVolumeSnapshot>
	QueueNoticeClusterVolumeSnapshotDeleted = "cdm.center.notice.cluster.volume-snapshot.deleted"

	// QueueNoticeClusterInstanceSpecCreated 클러스터 인스턴스 스펙 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterInstanceSpec>
	QueueNoticeClusterInstanceSpecCreated = "cdm.center.notice.cluster.instance-spec.created"

	// QueueNoticeClusterInstanceSpecUpdated 클러스터 인스턴스 스펙 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterInstanceSpec>
	QueueNoticeClusterInstanceSpecUpdated = "cdm.center.notice.cluster.instance-spec.updated"

	// QueueNoticeClusterInstanceSpecDeleted 클러스터 인스턴스 스펙 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterInstanceSpec>
	QueueNoticeClusterInstanceSpecDeleted = "cdm.center.notice.cluster.instance-spec.deleted"

	// QueueNoticeClusterInstanceExtraSpecCreated 클러스터 인스턴스 엑스트라 스펙 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterInstanceExtraSpec>
	QueueNoticeClusterInstanceExtraSpecCreated = "cdm.center.notice.cluster.instance-extra-spec.created"

	// QueueNoticeClusterInstanceExtraSpecUpdated 클러스터 인스턴스 엑스트라 스펙 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterInstanceExtraSpec>
	QueueNoticeClusterInstanceExtraSpecUpdated = "cdm.center.notice.cluster.instance-extra-spec.updated"

	// QueueNoticeClusterInstanceExtraSpecDeleted 클러스터 인스턴스 엑스트라 스펙 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterInstanceExtraSpec>
	QueueNoticeClusterInstanceExtraSpecDeleted = "cdm.center.notice.cluster.instance-extra-spec.deleted"

	// QueueNoticeClusterInstanceCreated 클러스터 인스턴스 생성을 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.CreateClusterInstance>
	QueueNoticeClusterInstanceCreated = "cdm.center.notice.cluster.instance.created"

	// QueueNoticeClusterInstanceUpdated 클러스터 인스턴스 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterInstance>
	QueueNoticeClusterInstanceUpdated = "cdm.center.notice.cluster.instance.updated"

	// QueueNoticeClusterInstanceDeleted 클러스터 인스턴스 삭제를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DeleteClusterInstance>
	QueueNoticeClusterInstanceDeleted = "cdm.center.notice.cluster.instance.deleted"

	// QueueNoticeClusterInstanceVolumeAttached 클러스터 인스턴스 볼륨 attached 를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.AttachClusterInstanceVolume>
	QueueNoticeClusterInstanceVolumeAttached = "cdm.center.notice.cluster.instance.volume.attached"

	// QueueNoticeClusterInstanceVolumeUpdated 클러스터 인스턴스 볼륨 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterInstanceVolume>
	QueueNoticeClusterInstanceVolumeUpdated = "cdm.center.notice.cluster.instance.volume.updated"

	// QueueNoticeClusterInstanceVolumeDetached 클러스터 인스턴스 볼륨 detached 를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DetachClusterInstanceVolume>
	QueueNoticeClusterInstanceVolumeDetached = "cdm.center.notice.cluster.instance.volume.detached"

	// QueueNoticeClusterInstanceNetworkAttached 클러스터 인스턴스 네트워크 attached 를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.AttachClusterInstanceNetwork>
	QueueNoticeClusterInstanceNetworkAttached = "cdm.center.notice.cluster.instance.network.attached"

	// QueueNoticeClusterInstanceNetworkUpdated 클러스터 인스턴스 네트워크 업데이트를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.UpdateClusterInstanceNetwork>
	QueueNoticeClusterInstanceNetworkUpdated = "cdm.center.notice.cluster.instance.network.updated"

	// QueueNoticeClusterInstanceNetworkDetached 클러스터 인스턴스 네트워크 detached 를 공지하는 Topic
	// message:
	//   header: {}
	//   body: <queue.DetachClusterInstanceNetwork>
	QueueNoticeClusterInstanceNetworkDetached = "cdm.center.notice.cluster.instance.network.detached"
)
