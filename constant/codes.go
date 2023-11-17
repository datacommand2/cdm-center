package constant

var (
	// ClusterTypeOpenstack 오픈스택
	ClusterTypeOpenstack = "cluster.type.openstack"

	// ClusterTypeOpenshift 오픈쉬프트
	ClusterTypeOpenshift = "cluster.type.openshift"

	// ClusterTypeKubernetes 쿠버네티스
	ClusterTypeKubernetes = "cluster.type.kubernetes"

	// ClusterTypeVMWare VMWare
	ClusterTypeVMWare = "cluster.type.vmware"
)

var (
	// OpenstackHypervisorTypeKVM KVM 타입
	OpenstackHypervisorTypeKVM = "openstack.hypervisor.type.kvm"

	// OpenstackHypervisorTypeLXC LXC 타입
	OpenstackHypervisorTypeLXC = "openstack.hypervisor.type.lxc"

	// OpenstackHypervisorTypeQemu Qemu 타입
	OpenstackHypervisorTypeQemu = "openstack.hypervisor.type.qemu"

	// OpenstackHypervisorTypeUML UML 타입
	OpenstackHypervisorTypeUML = "openstack.hypervisor.type.uml"

	// OpenstackHypervisorTypeVmware Vmware 타입
	OpenstackHypervisorTypeVmware = "openstack.hypervisor.type.vmware"

	// OpenstackHypervisorTypeXen Xen 타입
	OpenstackHypervisorTypeXen = "openstack.hypervisor.type.xen"

	// OpenstackHypervisorTypeXenserver Xenserver 타입
	OpenstackHypervisorTypeXenserver = "openstack.hypervisor.type.xenserver"

	// OpenstackHypervisorTypeHyperv Hyperv 타입
	OpenstackHypervisorTypeHyperv = "openstack.hypervisor.type.hyperv"

	// OpenstackHypervisorTypeVirtuozzo Virtuozzo 타입
	OpenstackHypervisorTypeVirtuozzo = "openstack.hypervisor.type.virtuozzo"
)

var (
	// ClusterStateActive 클러스터 활성화 상태
	ClusterStateActive = "cluster.state.active"

	// ClusterStateInactive 클러스터 비활성화 상태
	ClusterStateInactive = "cluster.state.inactive"

	// ClusterStateWarning 클러스터 워닝 상태
	ClusterStateWarning = "cluster.state.warning"

	// ClusterStateLoading 클러스터 로딩 상태
	ClusterStateLoading = "cluster.state.loading"
)

var (
	// ClusterSyncStateCodeInit 클러스터 동기화를위해 초기화를 진행 중
	ClusterSyncStateCodeInit = "cluster.sync.state.init"

	// ClusterSyncStateCodeRunning 클러스터 동기화 진행 중
	ClusterSyncStateCodeRunning = "cluster.sync.state.running"

	// ClusterSyncStateCodeDone 클러스터 동기화 완료
	ClusterSyncStateCodeDone = "cluster.sync.state.done"

	// ClusterSyncStateCodeFailed 클러스터 동기화 실패
	ClusterSyncStateCodeFailed = "cluster.sync.state.failed"

	// ClusterSyncStateCodeUnknown 클러스터 동기화 상태 알 수 없음
	ClusterSyncStateCodeUnknown = "cluster.sync.state.unknown"
)

var (
	// ClusterPermissionModeReadOnly 읽기 권한
	ClusterPermissionModeReadOnly = "cluster.permission.mode.readonly"

	// ClusterPermissionModeReadWrite 읽기/쓰기 권한
	ClusterPermissionModeReadWrite = "cluster.permission.mode.readwrite"
)

var (
	// OpenstackKeypairTypeSSH SSH 타입
	OpenstackKeypairTypeSSH = "ssh"
	// OpenstackKeypairTypeX509 X509 타입
	OpenstackKeypairTypeX509 = "x509"
)
