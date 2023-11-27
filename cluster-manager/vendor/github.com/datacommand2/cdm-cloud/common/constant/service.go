package constant

// SolutionName 솔루션 이름
const SolutionName = "CDM-Cloud"

// services
const (
	// ServiceLicense 사용권 계약을 관리하고, 계약의 유효성을 검증하는 서비스
	ServiceLicense = "cdm-cloud-license"

	// ServiceLicenseDescription 사용권 계약에서 제약사항을 확인하는 서비스
	ServiceLicenseDescription = "License Service"

	// ServiceIdentity 사용자 계정, 권한과 세션을 관리하고, API 요청을 인증하는 서비스
	ServiceIdentity = "cdm-cloud-identity"

	// ServiceIdentityDescription 사용자 계정, 권한과 세션을 관리하고, API 요청을 확인하는 서비스
	ServiceIdentityDescription = "Identity Service"

	// ServiceScheduler 스케쥴을 관리하고 트리거하는 서비스
	ServiceScheduler = "cdm-cloud-scheduler"

	// ServiceSchedulerDescription 스케쥴을 확인하는 서비스
	ServiceSchedulerDescription = "Scheduler Service"

	// ServiceBackup Persistent Data 를 백업하고 복구하는 서비스. 추가로 백업본을 관리
	ServiceBackup = "cdm-cloud-backup"

	// ServiceBackupDescription 백업을 확인하는 서비스
	ServiceBackupDescription = "Backup Service"

	// ServiceMonitor CDM-Cloud 가 전개된 노드 및 서비스 전체의 상태/리소스와 각 서비스의 로그를 모니터링하는 서비스
	ServiceMonitor = "cdm-cloud-monitor"

	// ServiceMonitorDescription CDM-Cloud 가 전개된 노드 및 서비스 전체의 상태/리소스와 각 서비스의 로그를 확인하는 서비스
	ServiceMonitorDescription = "Monitoring Service"

	// ServiceNotification CDM-Cloud 에서 발생하는 이벤트를 관리하고, Notification 을 담당하는 서비스
	ServiceNotification = "cdm-cloud-notification"

	// ServiceNotificationDescription CDM-Cloud 에서 발생하는 이벤트를 관리하고, Notification 을 확인하는 서비스
	ServiceNotificationDescription = "Notification Service"

	// ServiceAPIGateway 외부로 공개된 Restful API를 내부의 서비스로 라우팅하는 기능을 담당하는 서비스
	ServiceAPIGateway = "cdm-cloud-api-gateway"

	// ServiceAPIGatewayDescription 외부로 공개된 Restful API를 내부의 서비스로 라우팅하는 기능을 확인하는 서비스
	ServiceAPIGatewayDescription = "API Gateway Service"
)

// backing services
const (
	// BackingServiceStore CDM-Cloud 의 Key-Value Store 로서 License, Session 등의 Non-Persistent Data Store
	BackingServiceStore = "cdm-cloud-etcd"

	// BackingServiceBroker CDM-Cloud 내 모든 서비스들 간의 비동기 메시지를 처리하기 위한 Message Queue Broker
	BackingServiceBroker = "cdm-cloud-rabbitmq"

	// BackingServiceDatabase CDM-Cloud 및 배포된 솔루션의 Persistent Metadata 를 위한 Database
	BackingServiceDatabase = "cdm-cloud-cockroach"

	// BackingServiceTelemetry CDM-Cloud 가 전개된 노드 및 서비스 전체의 상태/리소스 Metric 데이터를 수집하고, 모니터링 정보를 제공하는 Monitoring System
	BackingServiceTelemetry = "cdm-cloud-prometheus"
)
