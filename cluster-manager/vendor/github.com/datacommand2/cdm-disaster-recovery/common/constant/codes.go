package constant

// 클러스터 관계의 plan 상태
const (
	// ClusterRelationshipStateNormal 클러스터 관계의 plan normal 상태
	ClusterRelationshipStateNormal = "dr.cluster.relationship.state.normal"

	// ClusterRelationshipStateWarning 클러스터 관계의 plan warning 상태
	ClusterRelationshipStateWarning = "dr.cluster.relationship.state.warning"

	// ClusterRelationshipStateEmergency 클러스터 관계의 plan emergency 상태
	ClusterRelationshipStateEmergency = "dr.cluster.relationship.state.emergency"

	// ClusterRelationshipStateCritical 클러스터 관계의 plan critical 상태
	ClusterRelationshipStateCritical = "dr.cluster.relationship.state.critical"
)

// 보호그룹의 상태
const (
	// ProtectionGroupStateNormal 보호그룹의 normal 상태
	ProtectionGroupStateNormal = "dr.protection.group.state.normal"

	// ProtectionGroupStateWarning 보호그룹의 warning 상태
	ProtectionGroupStateWarning = "dr.protection.group.state.warning"

	// ProtectionGroupStateEmergency 보호그룹의 emergency 상태
	ProtectionGroupStateEmergency = "dr.protection.group.state.emergency"

	// ProtectionGroupStateCritical 보호그룹의 critical 상태
	ProtectionGroupStateCritical = "dr.protection.group.state.critical"
)

// 재해 복구 계획의 상태
const (
	// RecoveryPlanStateNormal 재해 복구 계획의 normal 상태
	RecoveryPlanStateNormal = "dr.recovery.plan.state.normal"

	// RecoveryPlanStateWarning 재해 복구 계획의 warning 상태
	RecoveryPlanStateWarning = "dr.recovery.plan.state.warning"

	// RecoveryPlanStateEmergency 재해 복구 계획의 emergency 상태
	RecoveryPlanStateEmergency = "dr.recovery.plan.state.emergency"

	// RecoveryPlanStateCritical 재해 복구 계획의 critical 상태
	RecoveryPlanStateCritical = "dr.recovery.plan.state.critical"
)

// 재해 복구 계획의 mirroring 상태
const (
	// RecoveryPlanMirroringStatePrepare 재해 복구 계획 내 volume 의 mirroring prepare 상태
	RecoveryPlanMirroringStatePrepare = "dr.recovery.plan.state.prepare"

	// RecoveryPlanMirroringStateMirroring 재해 복구 계획 내 volume 의 mirroring prepare 상태
	RecoveryPlanMirroringStateMirroring = "dr.recovery.plan.state.mirroring"

	// RecoveryPlanMirroringStatePaused 재해 복구 계획 내 volume 의 mirroring paused 상태
	RecoveryPlanMirroringStatePaused = "dr.recovery.plan.state.paused"

	// RecoveryPlanMirroringStateStopped 재해 복구 계획 내 volume 의 mirroring stopped 상태
	RecoveryPlanMirroringStateStopped = "dr.recovery.plan.state.stopped"

	// RecoveryPlanMirroringStateWarning 재해 복구 계획 내 volume 의 mirroring warning 상태
	RecoveryPlanMirroringStateWarning = "dr.recovery.plan.state.warning"
)

// 재해 복구 계획 update reason code
const (
	// Tenant
	RecoveryPlanTenantMirrorNameDuplicated = "cdm-dr.recovery.plan.tenant.mirror_name.duplicated"
	RecoveryPlanTenantUpdateFlag           = "cdm-dr.recovery.get_plan.tenants.recovery_cluster_tenant.need_to_be_updated"

	// AvailabilityZone
	RecoveryPlanAvailabilityZoneTargetHypervisorInsufficientResource = "cdm-dr.recovery.plan.availability_zone.recovery_cluster_hypervisor.insufficient_resources"
	RecoveryPlanAvailabilityZoneTargetRequired                       = "cdm-dr.recovery.plan.availability_zone.recovery_cluster_availability_zone.required"
	RecoveryPlanAvailabilityZoneTargetDeleted                        = "cdm-dr.recovery.plan.availability_zone.recovery_cluster_availability_zone.deleted"
	RecoveryPlanAvailabilityZoneUpdateFlag                           = "cdm-dr.recovery.get_plan.availability_zones.recovery_availability_zone.need_to_be_updated"

	// Volume
	RecoveryPlanVolumeSourceInitCopyRequired   = "cdm-dr.recovery.plan.volume.protection_cluster_volume.init_copy.required"
	RecoveryPlanVolumeTargetBackendNameChanged = "cdm-dr.recovery.plan.volume.recovery_cluster_storage.backend_name_changed"
	RecoveryPlanVolumeTargetRequired           = "cdm-dr.recovery.plan.volume.recovery_cluster_storage.required"
	RecoveryPlanVolumeTargetDeleted            = "cdm-dr.recovery.plan.volume.recovery_cluster_storage.deleted"
	RecoveryPlanVolumeUpdateFlag               = "cdm-dr.recovery.get_plan.volumes.cluster_storage.need_to_be_updated"
	RecoveryPlanVolumeUnavailableFlag          = "cdm-dr.recovery.get_plan.volumes.cluster_storage.unavailable_state"

	// Storage
	RecoveryPlanStorageSourceTypeUnknown        = "cdm-dr.recovery.plan.storage.protection_cluster_storage.type.unknown"
	RecoveryPlanStorageTargetBackendNameChanged = "cdm-dr.recovery.plan.storage.recovery_cluster_storage.backend_name_changed"
	RecoveryPlanStorageTargetRequired           = "cdm-dr.recovery.plan.storage.recovery_cluster_storage.required"
	RecoveryPlanStorageTargetDeleted            = "cdm-dr.recovery.plan.storage.recovery_cluster_storage.deleted"
	RecoveryPlanStorageUpdateFlag               = "cdm-dr.recovery.get_plan.storages.cluster_storage.need_to_be_updated"
	RecoveryPlanStorageUnavailableFlag          = "cdm-dr.recovery.get_plan.storages.cluster_storage.unavailable_state"

	// Network
	RecoveryPlanNetworkTargetExternalFlagChanged = "cdm-dr.recovery.plan.external_network.recovery_cluster_external_network.external_flag_changed"
	RecoveryPlanNetworkTargetRequired            = "cdm-dr.recovery.plan.external_network.recovery_cluster_external_network.required"
	RecoveryPlanNetworkTargetDeleted             = "cdm-dr.recovery.plan.external_network.recovery_cluster_external_network.deleted"
	RecoveryPlanNetworkUpdateFlag                = "cdm-dr.recovery.get_plan.external_networks.recovery_cluster_external_network.need_to_be_updated"

	// Router
	RecoveryPlanRouterTargetExternalFlagChanged = "cdm-dr.recovery.plan.router.recovery_cluster_external_network.external_flag_changed"
	RecoveryPlanRouterTargetDeleted             = "cdm-dr.recovery.plan.router.recovery_cluster_external_network.deleted"
	RecoveryPlanRouterUpdateFlag                = "cdm-dr.recovery.get_plan.routers.recovery_cluster_external_network.need_to_be_updated"

	// Instance
	RecoveryPlanInstanceTargetRequired                       = "cdm-dr.recovery.plan.instance.recovery_cluster_availability_zone.required"
	RecoveryPlanInstanceTargetDeleted                        = "cdm-dr.recovery.plan.instance.recovery_cluster_availability_zone.deleted"
	RecoveryPlanInstanceTargetHypervisorInsufficientResource = "cdm-dr.recovery.plan.instance.recovery_cluster_hypervisor.insufficient_resources"
	RecoveryPlanInstanceTargetHypervisorUnusableState        = "cdm-dr.recovery.plan.instance.recovery_cluster_hypervisor.unusable"
	RecoveryPlanInstanceTargetHypervisorRequired             = "cdm-dr.recovery.plan.instance.recovery_cluster_hypervisor.required"
	RecoveryPlanInstanceTargetHypervisorDeleted              = "cdm-dr.recovery.plan.instance.recovery_cluster_hypervisor.deleted"
	RecoveryPlanInstanceUpdateFlag                           = "cdm-dr.recovery.get_plan.instances.recovery_cluster_hypervisor.need_to_be_updated"

	// FloatingIP
	RecoveryPlanFloatingIPTargetFloatingIPDuplicatedIPAddress                    = "cdm-dr.recovery.plan.instance_floating_ip.ip_address.duplicated_with_recovery_floating_ip"
	RecoveryPlanFloatingIPTargetRoutingInterfaceDuplicatedIPAddress              = "cdm-dr.recovery.plan.instance_floating_ip.ip_address.duplicated_with_recovery_routing_interface"
	RecoveryPlanFloatingIPTargetFloatingIPAndRoutingInterfaceDuplicatedIPAddress = "cdm-dr.recovery.plan.instance_floating_ip.ip_address.duplicated_with_recovery_floating_ip_and_recovery_routing_interface"
	RecoveryPlanFloatingIPUpdateFlag                                             = "cdm-dr.recovery.get_plan.instance_floating_ips.recovery_cluster_ip_address.need_to_be_updated"
)

// 테넌트 복구유형
const (
	// TenantRecoveryPlanTypeMirroring 테넌트 미러링
	TenantRecoveryPlanTypeMirroring = "dr.recovery.plan.tenant.recovery.type.mirroring"
)

// 가용구역 복구유형
const (
	// AvailabilityZoneRecoveryPlanTypeMapping 가용구역 매핑
	AvailabilityZoneRecoveryPlanTypeMapping = "dr.recovery.plan.availability-zone.recovery.type.mapping"
)

// 외부 네트워크 복구유형
const (
	// ExternalNetworkRecoveryPlanTypeMapping 외부 네트워크 매핑
	ExternalNetworkRecoveryPlanTypeMapping = "dr.recovery.plan.network.recovery.type.mapping"
)

// 라우터 복구유형
const (
	// RouterRecoveryPlanTypeMirroring 라우터 미러링
	RouterRecoveryPlanTypeMirroring = "dr.recovery.plan.network.recovery.type.mirroring"
)

// 스토리지 복구유형
const (
	// StorageRecoveryPlanTypeMapping 스토리지 매핑
	StorageRecoveryPlanTypeMapping = "dr.recovery.plan.storage.recovery.type.mapping"
)

// 인스턴스 복구유형
const (
	// InstanceRecoveryPlanTypeMirroring 인스턴스 미러링
	InstanceRecoveryPlanTypeMirroring = "dr.recovery.plan.instance.recovery.type.mirroring"
)

// 볼륨 복구유형
const (
	// VolumeRecoveryPlanTypeMirroring 볼륨 미러링
	VolumeRecoveryPlanTypeMirroring = "dr.recovery.plan.volume.recovery.type.mirroring"
)

// 인스턴스 진단방식
const (
	// InstanceDiagnosisMethodShellScript 인스턴스 내부에서 스크립트를 수행하여 진단하는 방식. 스크립트의 종료 상태코드가 0이면 성공
	InstanceDiagnosisMethodShellScript = "dr.recovery.plan.instance.diagnosis.method.shell-script"

	// InstanceDiagnosisMethodPortScan 활성화 포트 스캔 방식. 포트가 활성화되어 있다면 성공
	InstanceDiagnosisMethodPortScan = "dr.recovery.plan.instance.diagnosis.method.port-scan"

	// InstanceDiagnosisMethodHTTPGet HTTP Get 요청 방식. 응답의 상태코드가 200보다 크고 400보다 작으면 성공
	InstanceDiagnosisMethodHTTPGet = "dr.recovery.plan.instance.diagnosis.method.http-get"
)

// 복구 유형
const (
	// RecoveryTypeCodeSimulation 재해 복구 시뮬레이션
	RecoveryTypeCodeSimulation = "dr.recovery.type.simulation"

	// RecoveryTypeCodeMigration 재해 복구 실행
	RecoveryTypeCodeMigration = "dr.recovery.type.migration"
)

// 재해 복구 방향
const (
	// RecoveryJobDirectionCodeFailover 재해 복구 failover
	RecoveryJobDirectionCodeFailover = "dr.recovery.direction.failover"

	// RecoveryJobDirectionCodeFailback 재해 복구 failback
	RecoveryJobDirectionCodeFailback = "dr.recovery.direction.failback"
)

// 재해 복구 데이터 시점 유형
const (
	// RecoveryPointTypeCodeLatest 최신 데이터로 복구
	RecoveryPointTypeCodeLatest = "dr.recovery.recovery_point.type.latest"
)

// 재해복구작업 operation
const (
	// RecoveryJobOperationRun 재해복구작업을 실행
	RecoveryJobOperationRun = "run"

	// RecoveryJobOperationPause 재해복구작업을 일시중지
	RecoveryJobOperationPause = "pause"

	// RecoveryJobOperationCancel 재해복구작업을 취소
	RecoveryJobOperationCancel = "cancel"

	// RecoveryJobOperationRetry 재해복구작업을 재시도
	RecoveryJobOperationRetry = "retry"

	// RecoveryJobOperationRollback 재해복구작업을 롤백
	RecoveryJobOperationRollback = "rollback"

	// RecoveryJobOperationRetryRollback 재해복구작업 롤백 재시도
	RecoveryJobOperationRetryRollback = "retry-rollback"

	// RecoveryJobOperationIgnoreRollback 재해복구작업의 롤백을 무시
	RecoveryJobOperationIgnoreRollback = "ignore-rollback"

	// RecoveryJobOperationConfirm 재해복구작업을 확정
	RecoveryJobOperationConfirm = "confirm"

	// RecoveryJobOperationRetryConfirm 재해복구작업 확정 재시도
	RecoveryJobOperationRetryConfirm = "retry-confirm"

	// RecoveryJobOperationCancelConfirm 재해복구작업의 확정을 취소
	RecoveryJobOperationCancelConfirm = "cancel-confirm"
)

// 재해복구작업 상태
const (
	// RecoveryJobStateCodeWaiting 재해복구작업 대기 중
	RecoveryJobStateCodeWaiting = "dr.recovery.job.state.waiting"

	// RecoveryJobStateCodePending 재해복구작업 보류 중
	RecoveryJobStateCodePending = "dr.recovery.job.state.pending"

	// RecoveryJobStateCodeRunning 재해복구작업 실행 중
	RecoveryJobStateCodeRunning = "dr.recovery.job.state.running"

	// RecoveryJobStateCodeCanceling 재해복구작업 취소 중
	RecoveryJobStateCodeCanceling = "dr.recovery.job.state.canceling"

	// RecoveryJobStateCodePaused 재해복구작업 일시 중지
	RecoveryJobStateCodePaused = "dr.recovery.job.state.paused"

	// RecoveryJobStateCodeCompleted 재해복구작업 완료
	RecoveryJobStateCodeCompleted = "dr.recovery.job.state.completed"

	// RecoveryJobStateCodeClearing 재해복구작업 정리 중
	RecoveryJobStateCodeClearing = "dr.recovery.job.state.clearing"

	// RecoveryJobStateCodeClearFailed 재해복구작업 정리 실패
	RecoveryJobStateCodeClearFailed = "dr.recovery.job.state.clear-failed"

	// RecoveryJobStateCodeReporting 재해복구작업 리포팅 중
	RecoveryJobStateCodeReporting = "dr.recovery.job.state.reporting"

	// RecoveryJobStateCodeFinished 재해복구작업 종료
	RecoveryJobStateCodeFinished = "dr.recovery.job.state.finished"
)

// 재해복구작업 보호그룹 상태
const (
	// RecoveryJobSecurityGroupStateCodePreparing 재해복구작업 보호그룹 복구 준비중
	RecoveryJobSecurityGroupStateCodePreparing = "dr.recovery.job.security_group.state.preparing"

	// RecoveryJobSecurityGroupStateCodeExcepted 재해복구작업 보호그룹 제외됨
	RecoveryJobSecurityGroupStateCodeExcepted = "dr.recovery.job.security_group.state.excepted"

	// RecoveryJobSecurityGroupStateCodeSuccess 재해복구작업 보호그룹 복구 성공
	RecoveryJobSecurityGroupStateCodeSuccess = "dr.recovery.job.security_group.state.success"

	// RecoveryJobSecurityGroupStateCodeFailed 재해복구작업 보호그룹 복구 실패 (오류)
	RecoveryJobSecurityGroupStateCodeFailed = "dr.recovery.job.security_group.state.failed"
)

// 재해복구작업 보호그룹 결과
const (
	// SecurityGroupRecoveryResultCodeSuccess 보호그룹 복구 성공
	SecurityGroupRecoveryResultCodeSuccess = "dr.recovery.security_group.result.success"

	// SecurityGroupRecoveryResultCodeFailed 보호그룹 복구 실패
	SecurityGroupRecoveryResultCodeFailed = "dr.recovery.security_group.result.failed"

	// SecurityGroupRecoveryResultCodeCanceled 보호그룹 복구 취소
	SecurityGroupRecoveryResultCodeCanceled = "dr.recovery.security_group.result.canceled"
)

// 재해복구작업 볼륨 상태
const (
	// RecoveryJobVolumeStateCodePreparing 재해복구작업 볼륨 복구 준비중
	RecoveryJobVolumeStateCodePreparing = "dr.recovery.job.volume.state.preparing"

	// RecoveryJobVolumeStateCodeExcepted 재해복구작업 볼륨 제외됨
	RecoveryJobVolumeStateCodeExcepted = "dr.recovery.job.volume.state.excepted"

	// RecoveryJobVolumeStateCodeSuccess 재해복구작업 볼륨 복구 성공
	RecoveryJobVolumeStateCodeSuccess = "dr.recovery.job.volume.state.success"

	// RecoveryJobVolumeStateCodeFailed 재해복구작업 볼륨 복구 실패
	RecoveryJobVolumeStateCodeFailed = "dr.recovery.job.volume.state.failed"
)

// 재해복구작업 볼륨 결과
const (
	// VolumeRecoveryResultCodeSuccess 볼륨 복구 성공
	VolumeRecoveryResultCodeSuccess = "dr.recovery.volume.result.success"

	// VolumeRecoveryResultCodeFailed 볼륨 복구 실패
	VolumeRecoveryResultCodeFailed = "dr.recovery.volume.result.failed"

	// VolumeRecoveryResultCodeCanceled 볼륨 복구 취소
	VolumeRecoveryResultCodeCanceled = "dr.recovery.volume.result.canceled"
)

// 재해복구작업 인스턴스 상태
const (
	// RecoveryJobInstanceStateCodeExcepted 재해복구작업 인스턴스 제외됨
	RecoveryJobInstanceStateCodeExcepted = "dr.recovery.job.instance.state.excepted"

	// RecoveryJobInstanceStateCodeIgnored 재해복구작업 인스턴스 무시됨
	RecoveryJobInstanceStateCodeIgnored = "dr.recovery.job.instance.state.ignored"

	// RecoveryJobInstanceStateCodePreparing 재해복구작업 인스턴스 복구 준비중
	RecoveryJobInstanceStateCodePreparing = "dr.recovery.job.instance.state.preparing"

	// RecoveryJobInstanceStateCodeBooting 재해복구작업 인스턴스 생성중(+부팅중)
	RecoveryJobInstanceStateCodeBooting = "dr.recovery.job.instance.state.booting"

	// RecoveryJobInstanceStateCodeDiagnosing 재해복구작업 인스턴스 정상동작 여부 확인 중
	RecoveryJobInstanceStateCodeDiagnosing = "dr.recovery.job.instance.state.diagnosing"

	// RecoveryJobInstanceStateCodeSuccess 재해복구작업 인스턴스 복구 성공
	RecoveryJobInstanceStateCodeSuccess = "dr.recovery.job.instance.state.success"

	// RecoveryJobInstanceStateCodeFailed 재해복구작업 인스턴스 복구 실패 (오류)
	RecoveryJobInstanceStateCodeFailed = "dr.recovery.job.instance.state.failed"
)

// 재해복구작업 인스턴스 결과
const (
	// InstanceRecoveryResultCodeSuccess 인스턴스 복구 성공
	InstanceRecoveryResultCodeSuccess = "dr.recovery.instance.result.success"

	// InstanceRecoveryResultCodeFailed 인스턴스 복구 실패
	InstanceRecoveryResultCodeFailed = "dr.recovery.instance.result.failed"

	// InstanceRecoveryResultCodeCanceled 인스턴스 복구 취소
	InstanceRecoveryResultCodeCanceled = "dr.recovery.instance.result.canceled"
)

// 재해복구 결과
const (
	// RecoveryResultCodeSuccess 재해복구 성공
	RecoveryResultCodeSuccess = "dr.recovery.result.success"

	// RecoveryResultCodePartialSuccess 재해복구 부분 성공
	RecoveryResultCodePartialSuccess = "dr.recovery.result.partial_success"

	// RecoveryResultCodeFailed 재해복구 실패
	RecoveryResultCodeFailed = "dr.recovery.result.failed"

	// RecoveryResultCodeCanceled 재해복구 취소
	RecoveryResultCodeCanceled = "dr.recovery.result.canceled"

	// RecoveryResultCodeTimePassedCanceled 재해복구 대기시간 초과 취소
	RecoveryResultCodeTimePassedCanceled = "dr.recovery.result.passed_waiting_time_limit_canceled"

	// RecoveryResultCodeDuplicatedCanceled 재해복구 실행 예정시간 중복 취소
	RecoveryResultCodeDuplicatedCanceled = "dr.recovery.result.duplicated_next_run_time_canceled"
)

// 복제 환경 operation
const (
	// MirrorEnvironmentOperationStart 복제 환경을 구성
	MirrorEnvironmentOperationStart = "start"

	// MirrorEnvironmentOperationStop 복제 환경을 삭제
	MirrorEnvironmentOperationStop = "stop"
)

// 복제 환경의 상태
const (
	// MirrorEnvironmentStateCodeMirroring 복제 환경 구성 성공에 대한 상태 값
	MirrorEnvironmentStateCodeMirroring = "dr.mirror.environment.state.mirroring"

	// MirrorEnvironmentStateCodeStopping 복제 환경 삭제 진행에 대한 상태 값
	MirrorEnvironmentStateCodeStopping = "dr.mirror.environment.state.stopping"

	// MirrorEnvironmentStateCodeError 복제 환경 구성/삭제에 대한 에러 상태 값
	MirrorEnvironmentStateCodeError = "dr.mirror.environment.state.error"
)

// 볼륨 복제 operation
const (
	// MirrorVolumeOperationStart 볼륨 복제 시작
	MirrorVolumeOperationStart = "start"

	// MirrorVolumeOperationStop 볼륨 복제 중지
	MirrorVolumeOperationStop = "stop"

	// MirrorVolumeOperationStopAndDelete 볼륨 복제를 중지 및 데이터 삭제
	MirrorVolumeOperationStopAndDelete = "stop_and_delete"

	// MirrorVolumeOperationPause 볼륨 복제 일시 중지
	MirrorVolumeOperationPause = "pause"

	// MirrorVolumeOperationResume 볼륨 복제 재개
	MirrorVolumeOperationResume = "resume"

	// MirrorVolumeOperationDestroy 볼륨 복제 메타데이터 제거
	MirrorVolumeOperationDestroy = "destroy"
)

// 복제 환경의 상태
const (
	// MirrorVolumeStateCodeWaiting 볼륨 복제의 실시간 복제 대기 상태 값
	MirrorVolumeStateCodeWaiting = "dr.volume.mirror.state.waiting"

	// MirrorVolumeStateCodeInitializing 볼륨 복제의 실시간 복제 이전 동기화 상태 값
	MirrorVolumeStateCodeInitializing = "dr.volume.mirror.state.initializing"

	// MirrorVolumeStateCodeMirroring 볼륨 복제의 실시간 복제 상태 값
	MirrorVolumeStateCodeMirroring = "dr.volume.mirror.state.mirroring"

	// MirrorVolumeStateCodePaused 볼륨 복제의 일시 정지 상태 값
	MirrorVolumeStateCodePaused = "dr.volume.mirror.state.paused"

	// MirrorVolumeStateCodeStopping 볼륨 복제의 중지 상태 값
	MirrorVolumeStateCodeStopping = "dr.volume.mirror.state.stopping"

	// MirrorVolumeStateCodeDeleting 볼륨 복제의 실시간 복제 대기 데이터 삭제 상태 값
	MirrorVolumeStateCodeDeleting = "dr.volume.mirror.state.deleting"

	// MirrorVolumeStateCodeStopped 볼륨 복제의 중지 완료 상태 값
	MirrorVolumeStateCodeStopped = "dr.volume.mirror.state.stopped"

	// MirrorVolumeStateCodeError 볼륨 복제의 에러 상태 값
	MirrorVolumeStateCodeError = "dr.volume.mirror.state.error"
)

// task 진행 상태 코드
const (
	// MigrationTaskStateCodeWaiting task 가 진행 이전의 상태 값
	MigrationTaskStateCodeWaiting = "waiting"

	// MigrationTaskStateCodeRunning task 가 진행 중의 상태 값
	MigrationTaskStateCodeRunning = "running"

	// MigrationTaskStateCodeDone task 진행 완료 상태 값
	MigrationTaskStateCodeDone = "done"
)

// task 타입 코드
const (
	// MigrationTaskTypeCreateTenant 테넌트 생성
	MigrationTaskTypeCreateTenant = "dr.migration.task.create.tenant"

	// MigrationTaskTypeCreateSecurityGroup 보안 그룹 생성
	MigrationTaskTypeCreateSecurityGroup = "dr.migration.task.create.security_group"

	// MigrationTaskTypeCreateSecurityGroupRule 보안 그룹 규칙 생성
	MigrationTaskTypeCreateSecurityGroupRule = "dr.migration.task.create.security_group_rule"

	// MigrationTaskTypeCopyVolume 볼륨 copy
	MigrationTaskTypeCopyVolume = "dr.migration.task.copy.volume"

	// MigrationTaskTypeImportVolume 볼륨 import
	MigrationTaskTypeImportVolume = "dr.migration.task.import.volume"

	// MigrationTaskTypeCreateSpec 스팩 생성
	MigrationTaskTypeCreateSpec = "dr.migration.task.create.spec"

	// MigrationTaskTypeCreateKeypair 키페어 생성
	MigrationTaskTypeCreateKeypair = "dr.migration.task.create.keypair"

	// MigrationTaskTypeCreateFloatingIP floatingIP 생성
	MigrationTaskTypeCreateFloatingIP = "dr.migration.task.create.floatingip"

	// MigrationTaskTypeCreateNetwork 네트워크 생성
	MigrationTaskTypeCreateNetwork = "dr.migration.task.create.network"

	// MigrationTaskTypeCreateSubnet 서브넷 생성
	MigrationTaskTypeCreateSubnet = "dr.migration.task.create.subnet"

	// MigrationTaskTypeCreateRouter 라우터 생성
	MigrationTaskTypeCreateRouter = "dr.migration.task.create.router"

	// MigrationTaskTypeCreateAndStopInstance 인스턴스 생성 & 중지(비기동)
	MigrationTaskTypeCreateAndStopInstance = "dr.migration.task.create-and-stop.instance"

	// MigrationTaskTypeCreateAndDiagnosisInstance 인스턴스 생성 & 정상 동작 여부 확인(기동)
	MigrationTaskTypeCreateAndDiagnosisInstance = "dr.migration.task.create-and-diagnosis.instance"

	// MigrationTaskTypeStopInstance 인스턴스 중지
	MigrationTaskTypeStopInstance = "dr.migration.task.stop.instance"

	// MigrationTaskTypeDeleteInstance 인스턴스 삭제
	MigrationTaskTypeDeleteInstance = "dr.migration.task.delete.instance"

	// MigrationTaskTypeDeleteSpec 스팩 삭제
	MigrationTaskTypeDeleteSpec = "dr.migration.task.delete.spec"

	// MigrationTaskTypeDeleteKeypair 키페어 삭제
	MigrationTaskTypeDeleteKeypair = "dr.migration.task.delete.keypair"

	// MigrationTaskTypeUnmanageVolume 볼륨 unmanage
	MigrationTaskTypeUnmanageVolume = "dr.migration.task.unmanage.volume"

	// MigrationTaskTypeDeleteVolumeCopy 볼륨 copy 삭제
	MigrationTaskTypeDeleteVolumeCopy = "dr.migration.task.delete.volume.copy"

	// MigrationTaskTypeDeleteSecurityGroup 보안 그룹 삭제
	MigrationTaskTypeDeleteSecurityGroup = "dr.migration.task.delete.security_group"

	// MigrationTaskTypeDeleteFloatingIP floatingIP 삭제
	MigrationTaskTypeDeleteFloatingIP = "dr.migration.task.delete.floatingip"

	// MigrationTaskTypeDeleteRouter 라우터 삭제
	MigrationTaskTypeDeleteRouter = "dr.migration.task.delete.router"

	// MigrationTaskTypeDeleteNetwork 네트워크 삭제
	MigrationTaskTypeDeleteNetwork = "dr.migration.task.delete.network"

	// MigrationTaskTypeDeleteTenant 테넌트 삭제
	MigrationTaskTypeDeleteTenant = "dr.migration.task.delete.tenant"

	// MigrationTaskTypeDeleteProtectionClusterInstance protection cluster 의 인스턴스 삭제
	MigrationTaskTypeDeleteProtectionClusterInstance = "dr.migration.task.delete.protection.cluster.instance"
)

// task result 코드
const (
	// MigrationTaskResultCodeSuccess task 성공
	MigrationTaskResultCodeSuccess = "success"

	// MigrationTaskResultCodeExcepted task 제외
	MigrationTaskResultCodeExcepted = "excepted"

	// MigrationTaskResultCodeIgnored task 무시
	MigrationTaskResultCodeIgnored = "ignored"

	// MigrationTaskResultCodeCanceled task 취소
	MigrationTaskResultCodeCanceled = "canceled"

	// MigrationTaskResultCodeFailed task 실패
	MigrationTaskResultCodeFailed = "failed"

	// MigrationTaskResultCodeUnknown task 알수 없음
	// task status 는 done 이지만 ErrNotFoundTaskResult 일때 사용
	MigrationTaskResultCodeUnknown = "unknown"
)

// 목표복구시점 시간 단위
const (
	// RecoveryPointObjectiveTypeMinute 분 단위
	RecoveryPointObjectiveTypeMinute = "recovery.point.objective.type.minute"

	// RecoveryPointObjectiveTypeHour 시 단위
	RecoveryPointObjectiveTypeHour = "recovery.point.objective.type.hour"

	// RecoveryPointObjectiveTypeDay 일 단위
	RecoveryPointObjectiveTypeDay = "recovery.point.objective.type.day"
)

// 클러스터 스토리지 상태
const (
	// ClusterStorageStateAvailable 사용 가능한 클러스터 스토리지 상태
	ClusterStorageStateAvailable = "available"

	// ClusterStorageStateUnavailable 사용 불가능한 클러스터 스토리지 상태
	ClusterStorageStateUnavailable = "unavailable"

	// ClusterStorageStateUnknown 알 수 없는 클러스터 스토리지 상태
	ClusterStorageStateUnknown = "unknown"
)
