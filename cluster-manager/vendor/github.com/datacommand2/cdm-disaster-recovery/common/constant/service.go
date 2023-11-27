package constant

// SolutionName 솔루션 이름
const SolutionName = "CDM-DisasterRecovery"

const (
	// ServiceLicenseName 서비스 이름
	ServiceLicenseName = "cdm-dr-license"

	// ServiceLicenseDescription 사용권 계약에서 제약사항을 확인하는 서비스
	ServiceLicenseDescription = "Disaster Recovery License Service"

	// ServiceManagerName 서비스 이름
	ServiceManagerName = "cdm-dr-manager"

	// ServiceManagerDescription 재해복구계획 및 모의훈련, 재해복구 작업등 솔루션을 전반적으로 관리하는 서비스
	ServiceManagerDescription = "Disaster Recovery Management Service"

	// ServiceStatisticsName 서비스 이름
	ServiceStatisticsName = "cdm-dr-statistics"

	// ServiceStatisticsDescription 클러스터의 통계보고서를 생성하고 관리하는 서비스
	ServiceStatisticsDescription = "Disaster Recovery Statistics Service"

	// ServiceMonitorName 서비스 이름
	ServiceMonitorName = "cdm-dr-monitor"

	// ServiceMonitorDescription 클러스터의 보호현황 실시간 모니터링 기능을 제공하는 서비스
	ServiceMonitorDescription = "Disaster Recovery Monitoring Service"

	// ServiceSnapshotName 서비스 이름
	ServiceSnapshotName = "cdm-target-snapshot"

	// ServiceSnapshotDescription 보호 그룹 스냅샷, 재해 복구 스냅샷, 볼륨 스냅샷 등 스냅샷을 관리 하는 서비스
	ServiceSnapshotDescription = "Disaster Recovery Target Snapshot Service"
)

const (
	// DaemonMigratorName 데몬 이름
	DaemonMigratorName = "cdm-dr-migrator"

	// DaemonMigratorDescription 재해복구계획의 모의훈련 및 재해복구 작업을 수행하는 데몬
	DaemonMigratorDescription = "Disaster Recovery Migration Daemon"

	// DaemonMirrorName 데몬 이름
	DaemonMirrorName = "cdm-dr-mirror"

	// DaemonMirrorDescription 재해복구계획의 복제환경 구성 및 볼륨 복제, 스냅샷 생성 등을 수행하는 데몬
	DaemonMirrorDescription = "Disaster Recovery Mirroring Daemon"
)
