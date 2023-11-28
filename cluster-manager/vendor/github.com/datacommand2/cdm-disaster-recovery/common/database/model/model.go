package model

import "time"

// ProtectionGroup 보호 그룹
type ProtectionGroup struct {
	// ID 는 자동 증가 되는 보호 그룹의 식별자 이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// TenantID 는 테넌트의 ID 이다.
	TenantID uint64 `gorm:"not null" json:"tenant_id,omitempty"`

	// OwnerGroupID 는 보호 그룹의 소유 그룹 ID 이다
	OwnerGroupID uint64 `gorm:"not null" json:"owner_group_id,omitempty"`

	// ProtectionClusterID 는 원본 클러스터의 ID 이다.
	ProtectionClusterID uint64 `gorm:"not null" json:"protection_cluster_id,omitempty"`

	// Name 는 보호 그룹의 이름 이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Remarks 는 보호 그룹의 비고 이다.
	Remarks *string `json:"remarks,omitempty"`

	// RecoveryPointObjectiveType 는 목표 복구 시점의 시간 단위 이다.
	RecoveryPointObjectiveType string `gorm:"not null" json:"recovery_point_objective_type,omitempty"`

	// RecoveryPointObjective 는 목표 복구 시점 이다.
	RecoveryPointObjective uint32 `gorm:"not null" json:"recovery_point_objective,omitempty"`

	// RecoveryTimeObjective 는 목표 복구 시간 이다.
	RecoveryTimeObjective uint32 `gorm:"not null" json:"recovery_time_objective,omitempty"`

	// SnapshotIntervalType 는 스냅샷 생성 주기의 시간 단위 이다.
	SnapshotIntervalType string `gorm:"not null" json:"snapshot_interval_type,omitempty"`

	// SnapshotInterval 는 스냅샷 생성 주기 이다.
	SnapshotInterval uint32 `gorm:"not null" json:"snapshot_interval,omitempty"`

	// SnapshotScheduleID 는 스냅샷 생성 스케쥴 ID 이다.
	SnapshotScheduleID uint64 `gorm:"not null" json:"snapshot_schedule_id,omitempty"`

	// CreatedAt 는 보호 그룹 정보가 생성된 (UnixTime)시간 이다.
	CreatedAt int64 `gorm:"not null" json:"created_at,omitempty"`

	// UpdatedAt 는 보호 그룹 정보가 수정된 (UnixTime)시간 이다.
	UpdatedAt int64 `gorm:"not null" json:"updated_at,omitempty"`
}

// ProtectionInstance 보호 그룹의 인스턴스 목록
type ProtectionInstance struct {
	// ProtectionGroupID 는 보호 그룹의 ID 이다.
	ProtectionGroupID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_group_id,omitempty"`

	// ProtectionClusterInstanceID 는 보호 그룹의 보호 인스턴스 ID 이다.
	ProtectionClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_id,omitempty"`
}

// ProtectionGroupSnapshot 보호그룹의 스냅샷
type ProtectionGroupSnapshot struct {
	// ID 는 자동 증가 되는 스냅샷의 식별자 이다
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ProtectionGroupID 는 보호 그룹 ID 이다.
	ProtectionGroupID uint64 `gorm:"not null" json:"protection_group_id,omitempty"`

	// Name 는 스냅샷의 이름 이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// CreatedAt 는 스냅샷이 생성된 시간(UnixTime) 이다.
	CreatedAt int64 `gorm:"not null" json:"created_at,omitempty"`
}

// ConsistencyGroup 일관성 그룹
// FIXME: openstack 한정. 다른 클러스터 넣을 때 구조를 바꿔야 될 수도 있음
type ConsistencyGroup struct {
	// ProtectionGroupID 는 보호 그룹의 ID 이다.
	ProtectionGroupID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_group_id,omitempty"`

	// ProtectionClusterStorageID 는 보호 cluster 의 스토리지 ID 이다
	ProtectionClusterStorageID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_storage_id,omitempty"`

	// ProtectionClusterTenantID 는 보호 cluster 의 테넌트 ID 이다
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterVolumeGroupUUID 는 보호 cluster 의 볼륨그룹 ID 이다
	ProtectionClusterVolumeGroupUUID string `gorm:"not null" json:"protection_cluster_volume_group_uuid,omitempty"`
}

// ConsistencyGroupSnapshot 일관성 그룹 스냅샷
// FIXME: openstack 한정. 다른 클러스터 넣을 때 구조를 바꿔야 될 수도 있음
type ConsistencyGroupSnapshot struct {
	ProtectionGroupSnapshotID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_group_snapshot_id,omitempty"`

	// ProtectionGroupID 는 보호 그룹의 ID 이다.
	ProtectionGroupID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_group_id,omitempty"`

	// ProtectionClusterStorageID 는 보호 cluster 의 스토리지 ID 이다
	ProtectionClusterStorageID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_storage_id,omitempty"`

	// ProtectionClusterTenantID 는 보호 cluster 의 테넌트 ID 이다
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterVolumeGroupSnapshotUUID 는 보호 cluster 볼륨그룹의 스냅샷 ID 이다
	ProtectionClusterVolumeGroupSnapshotUUID string `gorm:"not null" json:"protection_cluster_volume_group_snapshot_uuid,omitempty"`
}

// Plan 재해 복구 계획
type Plan struct {
	// ID 는 자동 증가 되는 재해 복구 계획의 식별자 이다
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ProtectionGroupID 는 보호 그룹 ID 이다.
	ProtectionGroupID uint64 `gorm:"not null" json:"protection_group_id,omitempty"`

	// RecoveryClusterID 는 복구 cluster 의 ID 이다.
	RecoveryClusterID uint64 `gorm:"not null" json:"recovery_cluster_id,omitempty"`

	// ActivationFlag 는 재해 복구 계획의 활성화 유무 이다.
	ActivationFlag bool `gorm:"not null" json:"activation_flag,omitempty"`

	// Name 는 재해 복구 계획의 이름 이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// SnapshotRetentionCount 는 스냅샷 유지 갯수 이다.
	SnapshotRetentionCount uint32 `gorm:"not null" json:"snapshot_retention_count,omitempty"`

	// Remarks 는 재해 복구 계획의 비고 이다.
	Remarks *string `json:"remarks,omitempty"`

	// DirectionCode 는 실행 작업의 direction code 이다.
	DirectionCode string `gorm:"not null" json:"direction_code,omitempty"`

	// CreatedAt 는 재해 복구 계획 정보가 생성된 시간(UnixTime) 이다.
	CreatedAt int64 `gorm:"not null" json:"created_at,omitempty"`

	// UpdatedAt 는 재해 복구 계획 정보가 수정된 시간(UnixTime) 이다.
	UpdatedAt int64 `gorm:"not null" json:"updated_at,omitempty"`
}

// PlanTenant 테넌트 재해 복구 계획
type PlanTenant struct {
	// RecoveryPlanID 는 재해 복구 계획의 ID 이다.
	RecoveryPlanID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_plan_id,omitempty"`

	// ProtectionClusterTenantID 는 보호 cluster 의 테넌트 ID 이다.
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	//RecoveryTypeCode 는 복구 유형 이다.
	RecoveryTypeCode string `gorm:"not null" json:"recovery_type_code,omitempty"`

	//RecoveryClusterTenantID 는 복구 cluster 의 테넌트 ID 이다.
	RecoveryClusterTenantID *uint64 `json:"recovery_cluster_tenant_id,omitempty"`

	// RecoveryClusterTenantMirrorName 는 복구 cluster 의 미러링 테넌트 이름 이다.
	RecoveryClusterTenantMirrorName *string `json:"recovery_cluster_tenant_mirror_name,omitempty"`

	// RecoveryClusterTenantMirrorNameUpdateFlag 는 복구 cluster 의 미러링 테넌트 이름 변경 여부 이다.
	RecoveryClusterTenantMirrorNameUpdateFlag *bool `json:"recovery_cluster_tenant_mirror_name_update_flag,omitempty"`

	// RecoveryClusterTenantMirrorNameUpdateReasonCode 는 복구 cluster 의 미러링 테넌트 이름 변경 사유 메세지 코드 이다.
	RecoveryClusterTenantMirrorNameUpdateReasonCode *string `json:"recovery_cluster_tenant_mirror_name_update_reason_code,omitempty"`

	// RecoveryClusterTenantMirrorNameUpdateReasonContents 는 복구 cluster 의 미러링 테넌트 이름 변경 사유 메세지 상세 내용 이다.
	RecoveryClusterTenantMirrorNameUpdateReasonContents *string `json:"recovery_cluster_tenant_mirror_name_update_reason_contents,omitempty"`
}

// PlanAvailabilityZone 가용 구역 재해 복구 계획
type PlanAvailabilityZone struct {
	// RecoveryPlanID 는 재해 복구 계획의 ID 이다.
	RecoveryPlanID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_plan_id,omitempty"`

	// ProtectionClusterAvailabilityZoneID 는 보호 cluster 의 가용 구역 ID 이다.
	ProtectionClusterAvailabilityZoneID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_availability_zone_id,omitempty"`

	// RecoveryTypeCode 는 복구 유형 이다.
	RecoveryTypeCode string `gorm:"not null" json:"recovery_type_code,omitempty"`

	// RecoveryClusterAvailabilityZoneID 는 복구 cluster 의 가용 구역 ID 이다.
	RecoveryClusterAvailabilityZoneID *uint64 `json:"recovery_cluster_availability_zone_id,omitempty"`

	// RecoveryClusterAvailabilityZoneUpdateFlag 는 복구 cluster 의 가용 구역 변경 여부 이다.
	RecoveryClusterAvailabilityZoneUpdateFlag *bool `json:"recovery_cluster_availability_zone_update_flag,omitempty"`

	// RecoveryClusterAvailabilityZoneUpdateReasonCode 는 복구 cluster 의 가용 구역 변경 사유 메세지 코드 이다.
	RecoveryClusterAvailabilityZoneUpdateReasonCode *string `json:"recovery_cluster_availability_zone_update_reason_code,omitempty"`

	// RecoveryClusterAvailabilityZoneUpdateReasonContents 는 복구 cluster 의 가용 구역 변경 사유 메세지 상세 내용 이다.
	RecoveryClusterAvailabilityZoneUpdateReasonContents *string `json:"recovery_cluster_availability_zone_update_reason_contents,omitempty"`
}

// PlanExternalNetwork 외부 네트워크 재해 복구 계획
type PlanExternalNetwork struct {
	// RecoveryPlanID 는 재해 복구 계획의 ID 이다.
	RecoveryPlanID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_plan_id,omitempty"`

	// ProtectionClusterExternalNetworkID 는 보호 cluster 의 네트워크 ID 이다.
	ProtectionClusterExternalNetworkID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_external_network_id,omitempty"`

	// RecoveryTypeCode 는 복구 유형 이다.
	RecoveryTypeCode string `gorm:"not null" json:"recovery_type_code,omitempty"`

	// RecoveryClusterExternalNetworkID 는 복구 cluster 의 외부 네트워크 ID 이다.
	RecoveryClusterExternalNetworkID *uint64 `json:"recovery_cluster_external_network_id,omitempty"`

	// RecoveryClusterExternalNetworkUpdateFlag 는 복구 cluster 의 외부 네트워크 변경 여부 이다.
	RecoveryClusterExternalNetworkUpdateFlag *bool `json:"recovery_cluster_external_network_update_flag,omitempty"`

	// RecoveryClusterExternalNetworkUpdateReasonCode 는 복구 cluster 의 외부 네트워크 변경 사유 메시지 코드 이다.
	RecoveryClusterExternalNetworkUpdateReasonCode *string `json:"recovery_cluster_external_network_update_reason_code,omitempty"`

	//  RecoveryClusterExternalNetworkUpdateReasonContents 는 복구 cluster 의 외부 네트워크 변경 사유 메세지 상세 내용 이다.
	RecoveryClusterExternalNetworkUpdateReasonContents *string `json:"recovery_cluster_external_network_update_reason_contents,omitempty"`
}

// PlanRouter 라우터 재해 복구 계획
type PlanRouter struct {
	// RecoveryPlanID 는 재해 복구 계획의 ID 이다.
	RecoveryPlanID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_plan_id,omitempty"`

	// ProtectionClusterRouterID 는 보호 cluster 의 라우터 ID 이다.
	ProtectionClusterRouterID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_router_id,omitempty"`

	// RecoveryTypeCode 는 복구 유형 이다.
	RecoveryTypeCode string `gorm:"not null" json:"recovery_type_code,omitempty"`

	// RecoveryClusterExternalNetworkID 는 복구 cluster 에서 라우팅할 외부 네트워크 ID 이다.
	RecoveryClusterExternalNetworkID *uint64 `json:"recovery_cluster_external_network_id,omitempty"`

	// RecoveryClusterExternalNetworkUpdateFlag 는 복구 cluster 에서 라우팅할 외부 네트워크 변경 여부 이다.
	RecoveryClusterExternalNetworkUpdateFlag *bool `json:"recovery_cluster_external_network_update_flag,omitempty"`

	// RecoveryClusterExternalNetworkUpdateReasonCode 는 복구 cluster 에서 라우팅할 외부 네트워크 변경 사유 메시지 코드 이다.
	RecoveryClusterExternalNetworkUpdateReasonCode *string `json:"recovery_cluster_external_network_update_reason_code,omitempty"`

	//  RecoveryClusterExternalNetworkUpdateReasonContents 는 복구 cluster 에서 라우팅할 외부 네트워크 변경 사유 메세지 상세 내용 이다.
	RecoveryClusterExternalNetworkUpdateReasonContents *string `json:"recovery_cluster_external_network_update_reason_contents,omitempty"`
}

// PlanExternalRoutingInterface 라우터 재해복구계획의 외부 라우팅 인터페이스 설정
type PlanExternalRoutingInterface struct {
	// RecoveryPlanID 는 재해 복구 계획의 ID 이다.
	RecoveryPlanID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_plan_id,omitempty"`

	// ProtectionClusterRouterID 는 보호 cluster 의 라우터 ID 이다.
	ProtectionClusterRouterID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_router_id,omitempty"`

	// RecoveryClusterExternalSubnetID 는 라우터의 외부 라우팅 인터페이스를 연결할 복구 cluster 외부 네트워크의 서브넷 ID 이다.
	RecoveryClusterExternalSubnetID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_cluster_external_subnet_id,omitempty"`

	// IPAddress 는 라우터의 외부 라우팅 인터페이스가 사용할 IP 주소이다.
	IPAddress *string `json:"ip_address,omitempty"`
}

// PlanStorage 스토리지 재해 복구 계획
type PlanStorage struct {
	// RecoveryPlanID 는 재해 복구 계획의 ID 이다.
	RecoveryPlanID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_plan_id,omitempty"`

	// ProtectionClusterStorageID 는 보호 cluster 의 스토리지 ID 이다
	ProtectionClusterStorageID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_storage_id,omitempty"`

	// RecoveryTypeCode 는 복구 유형 이다.
	RecoveryTypeCode string `gorm:"not null" json:"recovery_type_code,omitempty"`

	// RecoveryClusterStorageID 는 복구 cluster 의 스토리지 ID 이다
	RecoveryClusterStorageID *uint64 `json:"recovery_cluster_storage_id,omitempty"`

	// RecoveryClusterStorageUpdateFlag 는 복구 cluster 의 스토리지 변경 여부 이다.
	RecoveryClusterStorageUpdateFlag *bool `json:"recovery_cluster_storage_update_flag,omitempty"`

	// RecoveryClusterStorageUpdateReasonCode 는 복구 cluster 의 스토리지 변경 사유 메세지 코드 이다.
	RecoveryClusterStorageUpdateReasonCode *string `json:"recovery_cluster_storage_update_reason_code,omitempty"`

	// RecoveryClusterStorageUpdateReasonContents 는 복구 cluster 의 스토리지 사유 메세지 상세 내용 이다.
	RecoveryClusterStorageUpdateReasonContents *string `json:"recovery_cluster_storage_update_reason_contents,omitempty"`

	// UnavailableFlag 는 복구 불가 여부 이다.
	UnavailableFlag *bool `json:"unavailable_flag,omitempty"`

	// UnavailableReasonCode 는 복구 불가 메세지 코드 이다.
	UnavailableReasonCode *string `json:"unavailable_reason_code,omitempty"`

	// UnavailableReasonContents 는 복구 불가 메세지 상세 내용 이다.
	UnavailableReasonContents *string `json:"unavailable_reason_contents,omitempty"`
}

// PlanInstance 인스턴스 재해 복구
type PlanInstance struct {
	// RecoveryPlanID 는 재해 복구 계획의 ID 이다
	RecoveryPlanID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_plan_id,omitempty"`

	// ProtectionClusterInstanceID 는 보호 cluster 의 인스턴스 ID 이다
	ProtectionClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_id,omitempty"`

	// RecoveryTypeCode 는 복구 유형 이다.
	RecoveryTypeCode string `gorm:"not null" json:"recovery_type_code,omitempty"`

	// RecoveryClusterAvailabilityZoneID 는 복구 cluster 의 가용 구역 ID 이다
	RecoveryClusterAvailabilityZoneID *uint64 `json:"recovery_cluster_availability_zone_id,omitempty"`

	// RecoveryClusterAvailabilityZoneUpdateFlag 는 복구 cluster 의 가용 구역 변경 여부 이다
	RecoveryClusterAvailabilityZoneUpdateFlag *bool `json:"recovery_cluster_availability_zone_update_flag,omitempty"`

	// RecoveryClusterAvailabilityZoneUpdateReasonCode 는 복구 cluster 의 가용 구역 변경 사유 메세지 코드 이다
	RecoveryClusterAvailabilityZoneUpdateReasonCode *string `json:"recovery_cluster_availability_zone_update_reason_code,omitempty"`

	// RecoveryClusterAvailabilityZoneUpdateReasonContents 는 복구 cluster 의 가용 구역 변경 사유 메세지 상세 내용 이다
	RecoveryClusterAvailabilityZoneUpdateReasonContents *string `json:"recovery_cluster_availability_zone_update_reason_contents,omitempty"`

	// RecoveryClusterHypervisorID 는 복구 cluster 의 하이퍼 바이저의 ID 이다.
	RecoveryClusterHypervisorID *uint64 `json:"recovery_cluster_hypervisor_id,omitempty"`

	// AutoStartFlag 는 인스턴스 자동 실행 여부 이다.
	AutoStartFlag *bool `json:"auto_start_flag,omitempty"`

	// DiagnosisFlag 는 인스턴스 정상 작동 확인 여부 이다.
	DiagnosisFlag *bool `json:"diagnosis_flag,omitempty"`

	// DiagnosisMethodCode 는 인스턴스 정상 작동 확인 방법 이다.
	DiagnosisMethodCode *string `json:"diagnosis_method_code,omitempty"`

	// DiagnosisMethodData 는 인스턴스 정상 작동 확인 방법 데이터 이다.
	DiagnosisMethodData *string `json:"diagnosis_method_data,omitempty"`

	// DiagnosisTimeout 는 인스턴스 정상 작동 확인 대기 시간 이다.
	DiagnosisTimeout *uint32 `json:"diagnosis_timeout,omitempty"`
}

// PlanInstanceDependency 인스턴스 재해 복구 계획의 의존 인스턴스 목록
type PlanInstanceDependency struct {
	// RecoveryPlanID 재해 복구 계획의 ID 이다.
	RecoveryPlanID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_plan_id,omitempty"`

	// ProtectionClusterInstanceID 는 보호 cluster 의 인스턴스 ID 이다.
	ProtectionClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_id,omitempty"`

	// DependProtectionClusterInstanceID 는 복구 cluster 의 인스턴스 의존 인스턴스 ID 이다.
	DependProtectionClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"depend_protection_cluster_instance_id,omitempty"`
}

// PlanInstanceFloatingIP 인스턴스 floating IP 재해 복구 계획
type PlanInstanceFloatingIP struct {
	// RecoveryPlanID 재해 복구 계획의 ID 이다.
	RecoveryPlanID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_plan_id,omitempty"`

	// ProtectionClusterInstanceID 는 보호 cluster 의 인스턴스 ID 이다.
	ProtectionClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_id,omitempty"`

	// ProtectionClusterInstanceFloatingIPID 는 보호 cluster 의 인스턴스 floating IP ID 이다.
	ProtectionClusterInstanceFloatingIPID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_floating_ip_id,omitempty"`

	// UnavailableFlag 는 복구 cluster 의 인스턴스 floating IP 사용 불가능 유무 이다.
	UnavailableFlag *bool `json:"unavailable_flag,omitempty"`

	// UnavailableReasonCode 는 복구 cluster 의 인스턴스 floating IP 사용 불가능 메세지 코드 이다.
	UnavailableReasonCode *string `json:"unavailable_reason_code,omitempty"`

	// UnavailableReasonContents 는 복구 cluster 의 인스턴스 floating IP 사용 불가능 메세지 상세 내용 이다.
	UnavailableReasonContents *string `json:"unavailable_reason_contents,omitempty"`
}

// PlanVolume 스토리지 볼륨 재해 복구 계획
type PlanVolume struct {
	// RecoveryPlanID 재해 복구 계획의 ID 이다.
	RecoveryPlanID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_plan_id,omitempty"`

	// ProtectionClusterVolumeID 는 보호 cluster 의 볼륨 ID 이다.
	ProtectionClusterVolumeID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_volume_id,omitempty"`

	// RecoveryTypeCode 는 복구 유형 이다.
	RecoveryTypeCode string `gorm:"not null" json:"recovery_type_code,omitempty"`

	// RecoveryClusterStorageID 는 복구 cluster 의 storage ID 이다.
	RecoveryClusterStorageID *uint64 `json:"recovery_cluster_storage_id,omitempty"`

	// RecoveryClusterStorageUpdateFlag 는 복구 cluster 의 storage 변경 여부 이다.
	RecoveryClusterStorageUpdateFlag *bool `json:"recovery_cluster_storage_update_flag,omitempty"`

	// RecoveryClusterStorageUpdateReasonCode 는 복구 cluster 의 storage 변경 사유 메세지 코드 이다.
	RecoveryClusterStorageUpdateReasonCode *string `json:"recovery_cluster_storage_update_reason_code,omitempty"`

	// RecoveryClusterStorageUpdateReasonContents 는 복구 cluster 의 storage 변경 사유 메세지 상세 내용 이다.
	RecoveryClusterStorageUpdateReasonContents *string `json:"recovery_cluster_storage_update_reason_contents,omitempty"`

	// UnavailableFlag 는 복구 불가 여부 이다.
	UnavailableFlag *bool `json:"unavailable_flag,omitempty"`

	// UnavailableReasonCode 는 복구 불가 메세지 코드 이다.
	UnavailableReasonCode *string `json:"unavailable_reason_code,omitempty"`

	// UnavailableReasonContents 는 복구 불가 메세지 상세 내용 이다.
	UnavailableReasonContents *string `json:"unavailable_reason_contents,omitempty"`
}

// Job 재해 복구 작업
type Job struct {
	// ID 는 자동 증가 되는 재해 복구 작업 식별자 이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// RecoveryPlanID 는 재해 복구 계획의 ID 이다.
	RecoveryPlanID uint64 `gorm:"not null" json:"recovery_plan_id,omitempty"`

	// TypeCode 는 실행 작업 유형 이다.
	TypeCode string `gorm:"not null" json:"type_code,omitempty"`

	// RecoveryPointTypeCode 는 Job 실행 시 사용 할 데이터 시점 유형 이다.
	RecoveryPointTypeCode string `gorm:"not null" json:"recovery_point_type_code,omitempty"`

	// ScheduleID 는 작업의 스케쥴 ID 이다.
	ScheduleID *uint64 `json:"schedule_id,omitempty"`

	// NextRuntime 는 작업의 다음 실행시간이다.
	NextRuntime int64 `gorm:"not null" json:"next_runtime,omitempty"`

	// RecoveryPointSnapshotID 는 Job 실행 시 사용 할 보호그룹의 스냅샷 ID 이다.
	RecoveryPointSnapshotID *uint64 `json:"recovery_point_snapshot_id,omitempty"`

	// OperatorID 는 job 을 등록한 operator ID 이다.
	OperatorID uint64 `gorm:"not null" json:"operator_id,omitempty"`
}

// RecoveryResult 재해복구결과
type RecoveryResult struct {
	// ID 재해복구결과 ID
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// OwnerGroupID 는 보호 그룹의 소유 그룹 ID 이다
	OwnerGroupID uint64 `gorm:"not null" json:"owner_group_id,omitempty"`

	// OperatorID 재해복구작업 실행 혹은 예약한 사용자 ID
	OperatorID uint64 `gorm:"not null" json:"operator_id,omitempty"`

	// OperatorAccount 재해복구작업 실행 혹은 예약한 사용자 계정
	OperatorAccount string `gorm:"not null" json:"operator_account,omitempty"`

	// OperatorName 재해복구작업 실행 혹은 예약한 사용자 이름
	OperatorName string `gorm:"not null" json:"operator_name,omitempty"`

	// OperatorDepartment 재해복구작업 실행 혹은 예약한 사용자 부서
	OperatorDepartment *string `json:"operator_department,omitempty"`

	// OperatorPosition 재해복구작업 실행 혹은 예약한 사용자 직책
	OperatorPosition *string `json:"operator_position,omitempty"`

	// ApproverID 재해복구작업 확정/취소한 사용자 ID
	ApproverID *uint64 `json:"approver_id,omitempty"`

	// ApproverAccount 재해복구작업 확정/취소한 사용자 계정
	ApproverAccount *string `json:"approver_account,omitempty"`

	// ApproverName 재해복구작업 확정/취소한 사용자 이름
	ApproverName *string `json:"approver_name,omitempty"`

	// ApproverDepartment 재해복구작업 확정/취소한 사용자 부서
	ApproverDepartment *string `json:"approver_department,omitempty"`

	// ApproverPosition 재해복구작업 확정/취소한 사용자 직책
	ApproverPosition *string `json:"approver_position,omitempty"`

	// ProtectionGroupID 보호그룹 ID
	ProtectionGroupID uint64 `gorm:"not null" json:"protection_group_id,omitempty"`

	// ProtectionGroupName 보호그룹 이름
	ProtectionGroupName string `gorm:"not null" json:"protection_group_name,omitempty"`

	// ProtectionGroupRemarks 보호그룹 설명
	ProtectionGroupRemarks *string `json:"protection_group_remarks,omitempty"`

	// ProtectionClusterID 보호대상 클러스터 ID
	ProtectionClusterID uint64 `gorm:"not null" json:"protection_cluster_id,omitempty"`

	// ProtectionClusterTypeCode 보호대상 클러스터 종류
	ProtectionClusterTypeCode string `gorm:"not null" json:"protection_cluster_type_code,omitempty"`

	// ProtectionClusterName 보호대상 클러스터 이름
	ProtectionClusterName string `gorm:"not null" json:"protection_cluster_name,omitempty"`

	// ProtectionClusterRemarks 보호대상 클러스터 설명
	ProtectionClusterRemarks *string `json:"protection_cluster_remarks,omitempty"`

	// RecoveryPlanID 재해복구계획 ID
	RecoveryPlanID uint64 `gorm:"not null" json:"recovery_plan_id,omitempty"`

	// RecoveryPlanName 재해복구계획 이름
	RecoveryPlanName string `gorm:"not null" json:"recovery_plan_name,omitempty"`

	// RecoveryPlanRemarks 재해복구계획 설명
	RecoveryPlanRemarks *string `json:"recovery_plan_remarks,omitempty"`

	// RecoveryClusterID 복구대상 클러스터 ID
	RecoveryClusterID uint64 `gorm:"not null" json:"recovery_cluster_id,omitempty"`

	// RecoveryClusterTypeCode 복구대상 클러스터 종류
	RecoveryClusterTypeCode string `gorm:"not null" json:"recovery_cluster_type_code,omitempty"`

	// RecoveryClusterName 복구대상 클러스터 이름
	RecoveryClusterName string `gorm:"not null" json:"recovery_cluster_name,omitempty"`

	// RecoveryClusterRemarks 복구대상 클러스터 설명
	RecoveryClusterRemarks *string `json:"recovery_cluster_remarks,omitempty"`

	// RecoveryTypeCode 복구 유형
	RecoveryTypeCode string `gorm:"not null" json:"recovery_type_code,omitempty"`

	// RecoveryDirectionCode 복구 방향
	RecoveryDirectionCode string `gorm:"not null" json:"recovery_direction_code,omitempty"`

	// RecoveryPointObjectiveType 목표 복구 시점의 시간 단위
	RecoveryPointObjectiveType string `gorm:"not null" json:"recovery_point_objective_type,omitempty"`

	// RecoveryPointObjective 목표복구시점
	RecoveryPointObjective uint32 `gorm:"not null" json:"recovery_point_objective,omitempty"`

	// RecoveryTimeObjective 목표복구시간 (min)
	RecoveryTimeObjective uint32 `gorm:"not null" json:"recovery_time_objective,omitempty"`

	// RecoveryPointTypeCode 데이터 시점 유형
	RecoveryPointTypeCode string `gorm:"not null" json:"recovery_point_type_code,omitempty"`

	// RecoveryPoint 복구 데이터 시점
	RecoveryPoint int64 `gorm:"not null" json:"recovery_point,omitempty"`

	// ScheduleType 복구 실행타입
	ScheduleType *string `json:"schedule_type,omitempty"`

	// StartedAt 복구 시작일시
	StartedAt int64 `gorm:"not null" json:"started_at,omitempty"`

	// FinishedAt 복구 종료일시
	FinishedAt int64 `gorm:"not null" json:"finished_at,omitempty"`

	// ElapsedTime 복구 소요시간
	ElapsedTime int64 `gorm:"not null" json:"elapsed_time,omitempty"`

	// ResultCode 실행 결과
	ResultCode string `gorm:"not null" json:"result_code,omitempty"`

	// WarningFlag 경고 여부
	WarningFlag bool `gorm:"not null" json:"warning_flag,omitempty"`
}

// RecoveryResultRaw 재해복구결과의 raw 데이터
type RecoveryResultRaw struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// RecoveryJobID 재해복구 ID
	RecoveryJobID uint64 `gorm:"autoIncrement:false" json:"recovery_job_id,omitempty"`

	// Contents 재해복구에 사용된 정보
	Contents *string `json:"contents,omitempty"`

	// SharedTask shared task 정보
	SharedTask *string `json:"shared_task,omitempty"`
}

// RecoveryResultWarningReason 재해복구결과의 경고 사유
type RecoveryResultWarningReason struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ReasonSeq 재해복구결과 경고 사유 seq
	ReasonSeq uint64 `gorm:"primary_key;autoIncrement:false" json:"reason_seq,omitempty"`

	// Code 재해복구결과의 경고 사유 메시지 코드
	Code string `gorm:"not null" json:"code,omitempty"`

	// Contents 경고 사유 메시지 데이터
	Contents *string `json:"contents,omitempty"`
}

// RecoveryResultFailedReason 재해복구결과의 실패 사유
type RecoveryResultFailedReason struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ReasonSeq 재해복구결과 실패 사유 seq
	ReasonSeq uint64 `gorm:"primary_key;autoIncrement:false" json:"reason_seq,omitempty"`

	// Code 재해복구결과의 실패 사유 메시지 코드
	Code string `gorm:"not null" json:"code,omitempty"`

	// Contents 실패 사유 메시지 데이터
	Contents *string `json:"contents,omitempty"`
}

// RecoveryResultTaskLog 재해복구결과 Task 로그
type RecoveryResultTaskLog struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// LogSeq 재해복구결과 Task 로그 seq
	LogSeq uint64 `gorm:"primary_key;autoIncrement:false" json:"log_seq,omitempty"`

	// Code Task 로그 메시지 코드
	Code string `gorm:"not null" json:"code,omitempty"`

	// Contents Task 로그 메시지 데이터
	Contents *string `json:"contents,omitempty"`

	// LogDt Task 로그 일시
	LogDt int64 `gorm:"not null" json:"log_dt,omitempty"`
}

// RecoveryResultTenant 테넌트 복구결과
type RecoveryResultTenant struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterTenantUUID 보호대상 클러스터 테넌트 UUID
	ProtectionClusterTenantUUID string `gorm:"not null" json:"protection_cluster_tenant_uuid,omitempty"`

	// ProtectionClusterTenantName 보호대상 클러스터 테넌트명
	ProtectionClusterTenantName string `gorm:"not null" json:"protection_cluster_tenant_name,omitempty"`

	// ProtectionClusterTenantDescription 보호대상 클러스터 테넌트 설명
	ProtectionClusterTenantDescription *string `json:"protection_cluster_tenant_description,omitempty"`

	// ProtectionClusterTenantEnabled 보호대상 클러스터 테넌트 활성화 여부
	ProtectionClusterTenantEnabled bool `gorm:"not null" json:"protection_cluster_tenant_enabled,omitempty"`

	// RecoveryClusterTenantID 복구대상 클러스터 테넌트 ID
	RecoveryClusterTenantID *uint64 `json:"recovery_cluster_tenant_id,omitempty"`

	// RecoveryClusterTenantUUID 복구대상 클러스터 테넌트 UUID
	RecoveryClusterTenantUUID *string `json:"recovery_cluster_tenant_uuid,omitempty"`

	// RecoveryClusterTenantName 복구대상 클러스터 테넌트명
	RecoveryClusterTenantName *string `json:"recovery_cluster_tenant_name,omitempty"`

	// RecoveryClusterTenantDescription 복구대상 클러스터 테넌트 설명
	RecoveryClusterTenantDescription *string `json:"recovery_cluster_tenant_description,omitempty"`

	// RecoveryClusterTenantEnabled 복구대상 클러스터 테넌트 활성화 여부
	RecoveryClusterTenantEnabled *bool `json:"recovery_cluster_tenant_enabled,omitempty"`
}

// RecoveryResultInstanceSpec 인스턴스 Spec 복구결과
type RecoveryResultInstanceSpec struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterInstanceSpecID 보호대상 클러스터 인스턴스 Spec ID
	ProtectionClusterInstanceSpecID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_spec_id,omitempty"`

	// ProtectionClusterInstanceSpecUUID 보호대상 클러스터 인스턴스 Spec UUID
	ProtectionClusterInstanceSpecUUID string `gorm:"not null" json:"protection_cluster_instance_spec_uuid,omitempty"`

	// ProtectionClusterInstanceSpecName 보호대상 클러스터 인스턴스 Spec 이름
	ProtectionClusterInstanceSpecName string `gorm:"not null" json:"protection_cluster_instance_spec_name,omitempty"`

	// ProtectionClusterInstanceSpecDescription 보호대상 클러스터 인스턴스 Spec 설명
	ProtectionClusterInstanceSpecDescription *string `json:"protection_cluster_instance_spec_description,omitempty"`

	// ProtectionClusterInstanceSpecVcpuTotalCnt 보호대상 클러스터 인스턴스 Spec VCPU core count
	ProtectionClusterInstanceSpecVcpuTotalCnt uint64 `gorm:"not null" json:"protection_cluster_instance_spec_vcpu_total_cnt,omitempty"`

	// ProtectionClusterInstanceSpecMemTotalBytes 보호대상 클러스터 인스턴스 Spec total memory bytes
	ProtectionClusterInstanceSpecMemTotalBytes uint64 `gorm:"not null" json:"protection_cluster_instance_spec_mem_total_bytes,omitempty"`

	// ProtectionClusterInstanceSpecDiskTotalBytes 보호대상 클러스터 인스턴스 Spec total disk bytes
	ProtectionClusterInstanceSpecDiskTotalBytes uint64 `gorm:"not null" json:"protection_cluster_instance_spec_disk_total_bytes,omitempty"`

	// ProtectionClusterInstanceSpecSwapTotalBytes 보호대상 클러스터 인스턴스 Spec total swap bytes
	ProtectionClusterInstanceSpecSwapTotalBytes uint64 `gorm:"not null" json:"protection_cluster_instance_spec_swap_total_bytes,omitempty"`

	// ProtectionClusterInstanceSpecEphemeralTotalBytes 보호대상 클러스터 인스턴스 Spec total ephemeral bytes
	ProtectionClusterInstanceSpecEphemeralTotalBytes uint64 `gorm:"not null" json:"protection_cluster_instance_spec_ephemeral_total_bytes,omitempty"`

	// RecoveryClusterInstanceSpecID 복구대상 클러스터 인스턴스 Spec ID
	RecoveryClusterInstanceSpecID *uint64 `json:"recovery_cluster_instance_spec_id,omitempty"`

	// RecoveryClusterInstanceSpecUUID 복구대상 클러스터 인스턴스 Spec UUID
	RecoveryClusterInstanceSpecUUID *string `json:"recovery_cluster_instance_spec_uuid,omitempty"`

	// RecoveryClusterInstanceSpecName 복구대상 클러스터 인스턴스 Spec 이름
	RecoveryClusterInstanceSpecName *string `json:"recovery_cluster_instance_spec_name,omitempty"`

	// RecoveryClusterInstanceSpecDescription 복구대상 클러스터 인스턴스 Spec 설명
	RecoveryClusterInstanceSpecDescription *string `json:"recovery_cluster_instance_spec_description,omitempty"`

	// RecoveryClusterInstanceSpecVcpuTotalCnt 복구대상 클러스터 인스턴스 Spec VCPU core count
	RecoveryClusterInstanceSpecVcpuTotalCnt *uint64 `json:"recovery_cluster_instance_spec_vcpu_total_cnt,omitempty"`

	// RecoveryClusterInstanceSpecMemTotalBytes 복구대상 클러스터 인스턴스 Spec total memory bytes
	RecoveryClusterInstanceSpecMemTotalBytes *uint64 `json:"recovery_cluster_instance_spec_mem_total_bytes,omitempty"`

	// RecoveryClusterInstanceSpecDiskTotalBytes 복구대상 클러스터 인스턴스 Spec total disk bytes
	RecoveryClusterInstanceSpecDiskTotalBytes *uint64 `json:"recovery_cluster_instance_spec_disk_total_bytes,omitempty"`

	// RecoveryClusterInstanceSpecSwapTotalBytes 복구대상 클러스터 인스턴스 Spec total swap bytes
	RecoveryClusterInstanceSpecSwapTotalBytes *uint64 `json:"recovery_cluster_instance_spec_swap_total_bytes,omitempty"`

	// RecoveryClusterInstanceSpecEphemeralTotalBytes 복구대상 클러스터 인스턴스 Spec total ephemeral bytes
	RecoveryClusterInstanceSpecEphemeralTotalBytes *uint64 `json:"recovery_cluster_instance_spec_ephemeral_total_bytes,omitempty"`
}

// RecoveryResultInstanceExtraSpec 인스턴스 Extra Spec 복구결과
type RecoveryResultInstanceExtraSpec struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterInstanceSpecID 보호대상 클러스터 인스턴스 Spec ID
	ProtectionClusterInstanceSpecID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_spec_id,omitempty"`

	// ProtectionClusterInstanceExtraSpecID 보호대상 클러스터 인스턴스 Extra Spec ID
	ProtectionClusterInstanceExtraSpecID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_extra_spec_id,omitempty"`

	// ProtectionClusterInstanceExtraSpecKey 보호대상 클러스터 인스턴스 Extra Spec Key
	ProtectionClusterInstanceExtraSpecKey string `gorm:"not null" json:"protection_cluster_instance_extra_spec_key,omitempty"`

	// ProtectionClusterInstanceExtraSpecValue 보호대상 클러스터 인스턴스 Extra Spec Value
	ProtectionClusterInstanceExtraSpecValue string `gorm:"not null" json:"protection_cluster_instance_extra_spec_value,omitempty"`

	// RecoveryClusterInstanceExtraSpecID 복구대상 클러스터 인스턴스 Extra Spec ID
	RecoveryClusterInstanceExtraSpecID *uint64 `json:"recovery_cluster_instance_extra_spec_id,omitempty"`

	// RecoveryClusterInstanceExtraSpecKey 복구대상 클러스터 인스턴스 Extra Spec Key
	RecoveryClusterInstanceExtraSpecKey *string `json:"recovery_cluster_instance_extra_spec_key,omitempty"`

	// RecoveryClusterInstanceExtraSpecValue 복구대상 클러스터 인스턴스 Extra Spec Value
	RecoveryClusterInstanceExtraSpecValue *string `json:"recovery_cluster_instance_extra_spec_value,omitempty"`
}

// RecoveryResultKeypair 키페어 복구 결과
type RecoveryResultKeypair struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterKeypairID 보호대상 클러스터 Keypair ID
	ProtectionClusterKeypairID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_keypair_id,omitempty"`

	// ProtectionClusterKeypairName 보호대상 클러스터 Keypair 이름
	ProtectionClusterKeypairName string `gorm:"not null" json:"protection_cluster_keypair_name,omitempty"`

	// ProtectionClusterKeypairFingerprint 보호대상 클러스터 Keypair Fingerprint
	ProtectionClusterKeypairFingerprint string `gorm:"not null" json:"protection_cluster_keypair_fingerprint,omitempty"`

	// ProtectionClusterKeypairPublicKey 보호대상 클러스터 Keypair Public key
	ProtectionClusterKeypairPublicKey string `gorm:"not null" json:"protection_cluster_keypair_public_key,omitempty"`

	// ProtectionClusterKeypairTypeCode 보호대상 클러스터 Keypair Type code
	ProtectionClusterKeypairTypeCode string `gorm:"not null" json:"protection_cluster_keypair_type_code,omitempty"`

	// RecoveryClusterKeypairID 복구대상 클러스터 Keypair ID
	RecoveryClusterKeypairID *uint64 `json:"recovery_cluster_keypair_id,omitempty"`

	// RecoveryClusterKeypairName 복구대상 클러스터 Keypair 이름
	RecoveryClusterKeypairName *string `json:"recovery_cluster_keypair_name,omitempty"`

	// RecoveryClusterKeypairFingerprint 복구대상 클러스터 Keypair Fingerprint
	RecoveryClusterKeypairFingerprint *string `json:"recovery_cluster_keypair_fingerprint,omitempty"`

	// RecoveryClusterKeypairPublicKey 복구대상 클러스터 Keypair Public key
	RecoveryClusterKeypairPublicKey *string `json:"recovery_cluster_keypair_public_key,omitempty"`

	// RecoveryClusterKeypairTypeCode 복구대상 클러스터 Keypair Type code
	RecoveryClusterKeypairTypeCode *string `json:"recovery_cluster_keypair_type_code,omitempty"`
}

// RecoveryResultNetwork 네트워크 복구결과
type RecoveryResultNetwork struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterNetworkID 보호대상 클러스터 네트워크 ID
	ProtectionClusterNetworkID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_network_id,omitempty"`

	// ProtectionClusterNetworkUUID 보호대상 클러스터 네트워크 UUID
	ProtectionClusterNetworkUUID string `gorm:"not null" json:"protection_cluster_network_uuid,omitempty"`

	// ProtectionClusterNetworkName 보호대상 클러스터 네트워크명
	ProtectionClusterNetworkName string `gorm:"not null" json:"protection_cluster_network_name,omitempty"`

	// ProtectionClusterNetworkDescription 보호대상 클러스터 네트워크 설명
	ProtectionClusterNetworkDescription *string `json:"protection_cluster_network_description,omitempty"`

	// ProtectionClusterNetworkTypeCode 보호대상 클러스터 네트워크 종류
	ProtectionClusterNetworkTypeCode string `gorm:"not null" json:"protection_cluster_network_type_code,omitempty"`

	// ProtectionClusterNetworkState 보호대상 클러스터 네트워크 상태
	ProtectionClusterNetworkState string `gorm:"not null" json:"protection_cluster_network_state,omitempty"`

	// RecoveryClusterNetworkID 복구대상 클러스터 네트워크 ID
	RecoveryClusterNetworkID *uint64 `json:"recovery_cluster_network_id,omitempty"`

	// RecoveryClusterNetworkUUID 복구대상 클러스터 네트워크 UUID
	RecoveryClusterNetworkUUID *string `json:"recovery_cluster_network_uuid,omitempty"`

	// RecoveryClusterNetworkName 복구대상 클러스터 네트워크명
	RecoveryClusterNetworkName *string `json:"recovery_cluster_network_name,omitempty"`

	// RecoveryClusterNetworkDescription 복구대상 클러스터 네트워크 설명
	RecoveryClusterNetworkDescription *string `json:"recovery_cluster_network_description,omitempty"`

	// RecoveryClusterNetworkTypeCode 복구대상 클러스터 네트워크 종류
	RecoveryClusterNetworkTypeCode *string `json:"recovery_cluster_network_type_code,omitempty"`

	// RecoveryClusterNetworkState 복구대상 클러스터 네트워크 상태
	RecoveryClusterNetworkState *string `json:"recovery_cluster_network_state,omitempty"`
}

// RecoveryResultSubnet 서브넷 복구결과
type RecoveryResultSubnet struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterNetworkID 보호대상 클러스터 네트워크 ID
	ProtectionClusterNetworkID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_network_id,omitempty"`

	// ProtectionClusterSubnetID 보호대상 클러스터 서브넷 ID
	ProtectionClusterSubnetID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_subnet_id,omitempty"`

	// ProtectionClusterSubnetUUID 보호대상 클러스터 서브넷 UUID
	ProtectionClusterSubnetUUID string `gorm:"not null" json:"protection_cluster_subnet_uuid,omitempty"`

	// ProtectionClusterSubnetName 보호대상 클러스터 서브넷 이름
	ProtectionClusterSubnetName string `gorm:"not null" json:"protection_cluster_subnet_name,omitempty"`

	// ProtectionClusterSubnetDescription 보호대상 클러스터 서브넷 설명
	ProtectionClusterSubnetDescription *string `json:"protection_cluster_subnet_description,omitempty"`

	// ProtectionClusterSubnetNetworkCidr 보호대상 클러스터 서브넷 네트워크 CIDR
	ProtectionClusterSubnetNetworkCidr string `gorm:"not null" json:"protection_cluster_subnet_network_cidr,omitempty"`

	// ProtectionClusterSubnetDhcpEnabled 보호대상 클러스터 서브넷 DHCP 여부
	ProtectionClusterSubnetDhcpEnabled bool `gorm:"not null" json:"protection_cluster_subnet_dhcp_enabled,omitempty"`

	// ProtectionClusterSubnetGatewayEnabled 보호대상 클러스터 서브넷 게이트웨이 여부
	ProtectionClusterSubnetGatewayEnabled bool `gorm:"not null" json:"protection_cluster_subnet_gateway_enabled,omitempty"`

	// ProtectionClusterSubnetGatewayIPAddress 보호대상 클러스터 서브넷 게이트웨이 IP 주소
	ProtectionClusterSubnetGatewayIPAddress *string `json:"protection_cluster_subnet_gateway_ip_address,omitempty"`

	// ProtectionClusterSubnetIpv6AddressModeCode 보호대상 클러스터 서브넷 IPv6 address mode
	ProtectionClusterSubnetIpv6AddressModeCode *string `json:"protection_cluster_subnet_ipv6_address_mode_code,omitempty"`

	// ProtectionClusterSubnetIpv6RaModeCode 보호대상 클러스터 서브넷 IPv6 ra mode
	ProtectionClusterSubnetIpv6RaModeCode *string `json:"protection_cluster_subnet_ipv6_ra_mode_code,omitempty"`

	// RecoveryClusterSubnetID 복구대상 클러스터 서브넷 ID
	RecoveryClusterSubnetID *uint64 `json:"recovery_cluster_subnet_id,omitempty"`

	// RecoveryClusterSubnetUUID 복구대상 클러스터 서브넷 UUID
	RecoveryClusterSubnetUUID *string `json:"recovery_cluster_subnet_uuid,omitempty"`

	// RecoveryClusterSubnetName 복구대상 클러스터 서브넷 이름
	RecoveryClusterSubnetName *string `json:"recovery_cluster_subnet_name,omitempty"`

	// RecoveryClusterSubnetDescription 복구대상 클러스터 서브넷 설명
	RecoveryClusterSubnetDescription *string `json:"recovery_cluster_subnet_description,omitempty"`

	// RecoveryClusterSubnetNetworkCidr 복구대상 클러스터 서브넷 네트워크 CIDR
	RecoveryClusterSubnetNetworkCidr *string `json:"recovery_cluster_subnet_network_cidr,omitempty"`

	// RecoveryClusterSubnetDhcpEnabled 복구대상 클러스터 서브넷 DHCP 여부
	RecoveryClusterSubnetDhcpEnabled *bool `json:"recovery_cluster_subnet_dhcp_enabled,omitempty"`

	// RecoveryClusterSubnetGatewayEnabled 복구대상 클러스터 서브넷 게이트웨이 여부
	RecoveryClusterSubnetGatewayEnabled *bool `json:"recovery_cluster_subnet_gateway_enabled,omitempty"`

	// RecoveryClusterSubnetGatewayIPAddress 복구대상 클러스터 서브넷 게이트웨이 IP 주소
	RecoveryClusterSubnetGatewayIPAddress *string `json:"recovery_cluster_subnet_gateway_ip_address,omitempty"`

	// RecoveryClusterSubnetIpv6AddressModeCode 복구대상 클러스터 서브넷 IPv6 address mode
	RecoveryClusterSubnetIpv6AddressModeCode *string `json:"recovery_cluster_subnet_ipv6_address_mode_code,omitempty"`

	// RecoveryClusterSubnetIpv6RaModeCode 복구대상 클러스터 서브넷 IPv6 ra mode
	RecoveryClusterSubnetIpv6RaModeCode *string `json:"recovery_cluster_subnet_ipv6_ra_mode_code,omitempty"`
}

// RecoveryResultSubnetDHCPPool 서브넷 DHCP Pool 복구결과
type RecoveryResultSubnetDHCPPool struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterNetworkID 보호대상 클러스터 네트워크 ID
	ProtectionClusterNetworkID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_network_id,omitempty"`

	// ProtectionClusterSubnetID 보호대상 클러스터 서브넷 ID
	ProtectionClusterSubnetID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_subnet_id,omitempty"`

	// DhcpPoolSeq 서브넷 DHCP Pool seq
	DhcpPoolSeq uint64 `gorm:"primary_key;autoIncrement:false" json:"dhcp_pool_seq,omitempty"`

	// ProtectionClusterStartIPAddress 보호대상 클러스터 서브넷 DHCP Pool IP 주소 범위 시작
	ProtectionClusterStartIPAddress string `gorm:"not null" json:"protection_cluster_start_ip_address,omitempty"`

	// ProtectionClusterEndIPAddress 보호대상 클러스터 서브넷 DHCP Pool IP 주소 범위 끝
	ProtectionClusterEndIPAddress string `gorm:"not null" json:"protection_cluster_end_ip_address,omitempty"`

	// RecoveryClusterStartIPAddress 복구대상 클러스터 서브넷 DHCP Pool IP 주소 범위 시작
	RecoveryClusterStartIPAddress *string `json:"recovery_cluster_start_ip_address,omitempty"`

	// RecoveryClusterEndIPAddress 복구대상 클러스터 서브넷 DHCP Pool IP 주소 범위 끝
	RecoveryClusterEndIPAddress *string `json:"recovery_cluster_end_ip_address,omitempty"`
}

// RecoveryResultSubnetNameserver 서브넷 Nameserver 복구결과
type RecoveryResultSubnetNameserver struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterNetworkID 보호대상 클러스터 네트워크 ID
	ProtectionClusterNetworkID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_network_id,omitempty"`

	// ProtectionClusterSubnetID 보호대상 클러스터 서브넷 ID
	ProtectionClusterSubnetID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_subnet_id,omitempty"`

	// NameserverSeq 서브넷 Nameserver seq
	NameserverSeq uint64 `gorm:"primary_key;autoIncrement:false" json:"nameserver_seq,omitempty"`

	// ProtectionClusterNameserver 보호대상 클러스터 서브넷 Nameserver
	ProtectionClusterNameserver string `gorm:"not null" json:"protection_cluster_nameserver,omitempty"`

	// RecoveryClusterNameserver 복구대상 클러스터 서브넷 Nameserver
	RecoveryClusterNameserver *string `json:"recovery_cluster_nameserver,omitempty"`
}

// RecoveryResultFloatingIP Floating IP 복구결과
type RecoveryResultFloatingIP struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterFloatingIPID 보호대상 클러스터 Floating IP ID
	ProtectionClusterFloatingIPID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_floating_ip_id,omitempty"`

	// ProtectionClusterFloatingIPUUID 보호대상 클러스터 Floating IP UUID
	ProtectionClusterFloatingIPUUID string `gorm:"not null" json:"protection_cluster_floating_ip_uuid,omitempty"`

	// ProtectionClusterFloatingIPDescription 보호대상 클러스터 Floating IP 설명
	ProtectionClusterFloatingIPDescription *string `json:"protection_cluster_floating_ip_description,omitempty"`

	// ProtectionClusterFloatingIPIPAddress 보호대상 클러스터 Floating IP IP 주소
	ProtectionClusterFloatingIPIPAddress string `gorm:"not null" json:"protection_cluster_floating_ip_ip_address,omitempty"`

	// ProtectionClusterNetworkID 보호대상 클러스터 네트워크 ID
	ProtectionClusterNetworkID uint64 `gorm:"not null" json:"protection_cluster_network_id,omitempty"`

	// ProtectionClusterNetworkUUID 보호대상 클러스터 네트워크 UUID
	ProtectionClusterNetworkUUID string `gorm:"not null" json:"protection_cluster_network_uuid,omitempty"`

	// ProtectionClusterNetworkName 보호대상 클러스터 네트워크명
	ProtectionClusterNetworkName string `gorm:"not null" json:"protection_cluster_network_name,omitempty"`

	// ProtectionClusterNetworkDescription 보호대상 클러스터 네트워크 설명
	ProtectionClusterNetworkDescription *string `json:"protection_cluster_network_description,omitempty"`

	// ProtectionClusterNetworkTypeCode 보호대상 클러스터 네트워크 종류
	ProtectionClusterNetworkTypeCode string `gorm:"not null" json:"protection_cluster_network_type_code,omitempty"`

	// ProtectionClusterNetworkState 보호대상 클러스터 네트워크 상태
	ProtectionClusterNetworkState string `gorm:"not null" json:"protection_cluster_network_state,omitempty"`

	// RecoveryClusterFloatingIPID 복구대상 클러스터 Floating IP ID
	RecoveryClusterFloatingIPID *uint64 `json:"recovery_cluster_floating_ip_id,omitempty"`

	// RecoveryClusterFloatingIPUUID 복구대상 클러스터 Floating IP UUID
	RecoveryClusterFloatingIPUUID *string `json:"recovery_cluster_floating_ip_uuid,omitempty"`

	// RecoveryClusterFloatingIPDescription 복구대상 클러스터 Floating IP 설명
	RecoveryClusterFloatingIPDescription *string `json:"recovery_cluster_floating_ip_description,omitempty"`

	// RecoveryClusterFloatingIPIPAddress 보호대상 복구스터 Floating IP IP 주소
	RecoveryClusterFloatingIPIPAddress *string `json:"recovery_cluster_floating_ip_ip_address,omitempty"`

	// RecoveryClusterNetworkID 복구대상 클러스터 네트워크 ID
	RecoveryClusterNetworkID *uint64 `json:"recovery_cluster_network_id,omitempty"`

	// RecoveryClusterNetworkUUID 복구대상 클러스터 네트워크 UUID
	RecoveryClusterNetworkUUID *string `json:"recovery_cluster_network_uuid,omitempty"`

	// RecoveryClusterNetworkName 복구대상 클러스터 네트워크명
	RecoveryClusterNetworkName *string `json:"recovery_cluster_network_name,omitempty"`

	// RecoveryClusterNetworkDescription 복구대상 클러스터 네트워크 설명
	RecoveryClusterNetworkDescription *string `json:"recovery_cluster_network_description,omitempty"`

	// RecoveryClusterNetworkTypeCode 복구대상 클러스터 네트워크 종류
	RecoveryClusterNetworkTypeCode *string `json:"recovery_cluster_network_type_code,omitempty"`

	// RecoveryClusterNetworkState 복구대상 클러스터 네트워크 상태
	RecoveryClusterNetworkState *string `json:"recovery_cluster_network_state,omitempty"`
}

// RecoveryResultRouter 라우터 복구결과
type RecoveryResultRouter struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterRouterID 보호대상 클러스터 라우터 ID
	ProtectionClusterRouterID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_router_id,omitempty"`

	// ProtectionClusterRouterUUID 보호대상 클러스터 라우터 UUID
	ProtectionClusterRouterUUID string `gorm:"not null" json:"protection_cluster_router_uuid,omitempty"`

	// ProtectionClusterRouterName 보호대상 클러스터 라우터 이름
	ProtectionClusterRouterName string `gorm:"not null" json:"protection_cluster_router_name,omitempty"`

	// ProtectionClusterRouterDescription 보호대상 클러스터 라우터 설명
	ProtectionClusterRouterDescription *string `json:"protection_cluster_router_description,omitempty"`

	// ProtectionClusterRouterState 보호대상 클러스터 라우터 상태
	ProtectionClusterRouterState string `gorm:"not null" json:"protection_cluster_router_state,omitempty"`

	// RecoveryClusterRouterID 복구대상 클러스터 라우터 ID
	RecoveryClusterRouterID *uint64 `json:"recovery_cluster_router_id,omitempty"`

	// RecoveryClusterRouterUUID 복구대상 클러스터 라우터 UUID
	RecoveryClusterRouterUUID *string `json:"recovery_cluster_router_uuid,omitempty"`

	// RecoveryClusterRouterName 복구대상 클러스터 라우터 이름
	RecoveryClusterRouterName *string `json:"recovery_cluster_router_name,omitempty"`

	// RecoveryClusterRouterDescription 복구대상 클러스터 라우터 설명
	RecoveryClusterRouterDescription *string `json:"recovery_cluster_router_description,omitempty"`

	// RecoveryClusterRouterState 복구대상 클러스터 라우터 상태
	RecoveryClusterRouterState *string `json:"recovery_cluster_router_state,omitempty"`
}

// RecoveryResultInternalRoutingInterface 내부 네트워크 라우팅 인터페이스 복구결과
type RecoveryResultInternalRoutingInterface struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterRouterID 보호대상 클러스터 라우터 ID
	ProtectionClusterRouterID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_router_id,omitempty"`

	// RoutingInterfaceSeq 라우팅 인터페이스 seq
	RoutingInterfaceSeq uint64 `gorm:"primary_key;autoIncrement:false" json:"routing_interface_seq,omitempty"`

	// ProtectionClusterNetworkID 보호대상 클러스터 네트워크 ID
	ProtectionClusterNetworkID uint64 `gorm:"not null" json:"protection_cluster_network_id,omitempty"`

	// ProtectionClusterSubnetID 보호대상 클러스터 서브넷 ID
	ProtectionClusterSubnetID uint64 `gorm:"not null" json:"protection_cluster_subnet_id,omitempty"`

	// ProtectionClusterIPAddress 보호대상 클러스터 라우팅 인터페이스의 IP 주소
	ProtectionClusterIPAddress string `gorm:"not null" json:"protection_cluster_ip_address,omitempty"`

	// RecoveryClusterNetworkID 복구대상 클러스터 네트워크 ID
	RecoveryClusterNetworkID *uint64 `json:"recovery_cluster_network_id,omitempty"`

	// RecoveryClusterSubnetID 복구대상 클러스터 서브넷 ID
	RecoveryClusterSubnetID *uint64 `json:"recovery_cluster_subnet_id,omitempty"`

	// RecoveryClusterIPAddress 복구대상 클러스터 라우팅 인터페이스의 IP 주소
	RecoveryClusterIPAddress *string `json:"recovery_cluster_ip_address,omitempty"`
}

// RecoveryResultExternalRoutingInterface 외부 네트워크 라우팅 인터페이스 복구결과
type RecoveryResultExternalRoutingInterface struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterRouterID 보호대상 클러스터 라우터 ID
	ProtectionClusterRouterID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_router_id,omitempty"`

	// RoutingInterfaceSeq 라우팅 인터페이스 seq
	RoutingInterfaceSeq uint64 `gorm:"primary_key;autoIncrement:false" json:"routing_interface_seq,omitempty"`

	// ClusterNetworkID 보호대상 클러스터 네트워크 ID
	ClusterNetworkID uint64 `gorm:"not null" json:"cluster_network_id,omitempty"`

	// ClusterNetworkUUID 보호대상 클러스터 네트워크 UUID
	ClusterNetworkUUID string `gorm:"not null" json:"cluster_network_uuid,omitempty"`

	// ClusterNetworkName 보호대상 클러스터 네트워크명
	ClusterNetworkName string `gorm:"not null" json:"cluster_network_name,omitempty"`

	// ClusterNetworkDescription 보호대상 클러스터 네트워크 설명
	ClusterNetworkDescription *string `json:"cluster_network_description,omitempty"`

	// ClusterNetworkTypeCode 보호대상 클러스터 네트워크 종류
	ClusterNetworkTypeCode string `gorm:"not null" json:"cluster_network_type_code,omitempty"`

	// ClusterNetworkState 보호대상 클러스터 네트워크 상태
	ClusterNetworkState string `gorm:"not null" json:"cluster_network_state,omitempty"`

	// ClusterSubnetID 보호대상 클러스터 서브넷 ID
	ClusterSubnetID uint64 `gorm:"not null" json:"cluster_subnet_id,omitempty"`

	// ClusterSubnetUUID 보호대상 클러스터 서브넷 UUID
	ClusterSubnetUUID string `gorm:"not null" json:"cluster_subnet_uuid,omitempty"`

	// ClusterSubnetName 보호대상 클러스터 서브넷 이름
	ClusterSubnetName string `gorm:"not null" json:"cluster_subnet_name,omitempty"`

	// ClusterSubnetDescription 보호대상 클러스터 서브넷 설명
	ClusterSubnetDescription *string `json:"cluster_subnet_description,omitempty"`

	// ClusterSubnetNetworkCidr 보호대상 클러스터 서브넷 네트워크 CIDR
	ClusterSubnetNetworkCidr string `gorm:"not null" json:"cluster_subnet_network_cidr,omitempty"`

	// ClusterSubnetDhcpEnabled 보호대상 클러스터 서브넷 DHCP 여부
	ClusterSubnetDhcpEnabled bool `gorm:"not null" json:"cluster_subnet_dhcp_enabled,omitempty"`

	// ClusterSubnetGatewayEnabled 보호대상 클러스터 서브넷 게이트웨이 여부
	ClusterSubnetGatewayEnabled bool `gorm:"not null" json:"cluster_subnet_gateway_enabled,omitempty"`

	// ClusterSubnetGatewayIPAddress 보호대상 클러스터 서브넷 게이트웨이 IP 주소
	ClusterSubnetGatewayIPAddress *string `json:"cluster_subnet_gateway_ip_address,omitempty"`

	// ClusterSubnetIpv6AddressModeCode 보호대상 클러스터 서브넷 IPv6 address mode
	ClusterSubnetIpv6AddressModeCode *string `json:"cluster_subnet_ipv6_address_mode_code,omitempty"`

	// ClusterSubnetIpv6RaModeCode 보호대상 클러스터 서브넷 IPv6 ra mode
	ClusterSubnetIpv6RaModeCode *string `json:"cluster_subnet_ipv6_ra_mode_code,omitempty"`

	// ClusterIPAddress 보호대상 클러스터 라우팅 인터페이스의 IP 주소
	ClusterIPAddress string `gorm:"not null" json:"cluster_ip_address,omitempty"`

	// ProtectionFlag 보호대상 클러스터인지 여부
	ProtectionFlag bool `gorm:"not null" json:"protection_flag,omitempty"`
}

// RecoveryResultExtraRoute Extra Route 복구결과
type RecoveryResultExtraRoute struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterRouterID 보호대상 클러스터 라우터 ID
	ProtectionClusterRouterID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_router_id,omitempty"`

	// ExtraRouteSeq Extra Route seq
	ExtraRouteSeq uint64 `gorm:"primary_key;autoIncrement:false" json:"extra_route_seq,omitempty"`

	// ProtectionClusterDestination 보호대상 클러스터에서의 Route Destination
	ProtectionClusterDestination string `gorm:"not null" json:"protection_cluster_destination,omitempty"`

	// ProtectionClusterNexthop 보호대상 클러스터에서의 Route Nexthop
	ProtectionClusterNexthop string `gorm:"not null" json:"protection_cluster_nexthop,omitempty"`

	// RecoveryClusterDestination 복구대상 클러스터에서의 Route Destination
	RecoveryClusterDestination *string `json:"recovery_cluster_destination,omitempty"`

	// RecoveryClusterNexthop 복구대상 클러스터에서의 Route Nexthop
	RecoveryClusterNexthop *string `json:"recovery_cluster_nexthop,omitempty"`
}

// RecoveryResultSecurityGroup 보안그룹 복구결과
type RecoveryResultSecurityGroup struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterSecurityGroupID 보호대상 클러스터 보안그룹 ID
	ProtectionClusterSecurityGroupID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_security_group_id,omitempty"`

	// ProtectionClusterSecurityGroupUUID 보호대상 클러스터 보안그룹 UUID
	ProtectionClusterSecurityGroupUUID string `gorm:"not null" json:"protection_cluster_security_group_uuid,omitempty"`

	// ProtectionClusterSecurityGroupName 보호대상 클러스터 보안그룹 이름
	ProtectionClusterSecurityGroupName string `gorm:"not null" json:"protection_cluster_security_group_name,omitempty"`

	// ProtectionClusterSecurityGroupDescription 보호대상 클러스터 보안그룹 설명
	ProtectionClusterSecurityGroupDescription *string `json:"protection_cluster_security_group_description,omitempty"`

	// RecoveryClusterSecurityGroupID 복구대상 클러스터 보안그룹 ID
	RecoveryClusterSecurityGroupID *uint64 `json:"recovery_cluster_security_group_id,omitempty"`

	// RecoveryClusterSecurityGroupUUID 복구대상 클러스터 보안그룹 UUID
	RecoveryClusterSecurityGroupUUID *string `json:"recovery_cluster_security_group_uuid,omitempty"`

	// RecoveryClusterSecurityGroupName 복구대상 클러스터 보안그룹 이름
	RecoveryClusterSecurityGroupName *string `json:"recovery_cluster_security_group_name,omitempty"`

	// RecoveryClusterSecurityGroupDescription 복구대상 클러스터 보안그룹 설명
	RecoveryClusterSecurityGroupDescription *string `json:"recovery_cluster_security_group_description,omitempty"`
}

// RecoveryResultSecurityGroupRule 보안그룹 규칙 복구결과
type RecoveryResultSecurityGroupRule struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterSecurityGroupID 보호대상 클러스터 보안그룹 ID
	ProtectionClusterSecurityGroupID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_security_group_id,omitempty"`

	// ProtectionClusterSecurityGroupRuleID 보호대상 클러스터 보안그룹 규칙 ID
	ProtectionClusterSecurityGroupRuleID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_security_group_rule_id,omitempty"`

	// ProtectionClusterSecurityGroupRuleRemoteSecurityGroupID 보호대상 클러스터 보안그룹 규칙 대상 보안그룹
	ProtectionClusterSecurityGroupRuleRemoteSecurityGroupID *uint64 `json:"protection_cluster_security_group_rule_remote_security_group_id,omitempty"`

	// ProtectionClusterSecurityGroupRuleUUID 보호대상 클러스터 보안그룹 규칙 UUID
	ProtectionClusterSecurityGroupRuleUUID string `gorm:"not null" json:"protection_cluster_security_group_rule_uuid,omitempty"`

	// ProtectionClusterSecurityGroupRuleDescription 보호대상 클러스터 보안그룹 규칙 설명
	ProtectionClusterSecurityGroupRuleDescription *string `json:"protection_cluster_security_group_rule_description,omitempty"`

	// ProtectionClusterSecurityGroupRuleNetworkCidr 보호대상 클러스터 보안그룹 규칙 네트워크 CIDR
	ProtectionClusterSecurityGroupRuleNetworkCidr *string `json:"protection_cluster_security_group_rule_network_cidr,omitempty"`

	// ProtectionClusterSecurityGroupRuleDirection 보호대상 클러스터 보안그룹 규칙 방향
	ProtectionClusterSecurityGroupRuleDirection string `gorm:"not null" json:"protection_cluster_security_group_rule_direction,omitempty"`

	// ProtectionClusterSecurityGroupRulePortRangeMin 보호대상 클러스터 보안그룹 규칙 허용 포트 범위 시작
	ProtectionClusterSecurityGroupRulePortRangeMin *uint64 `json:"protection_cluster_security_group_rule_port_range_min,omitempty"`

	// ProtectionClusterSecurityGroupRulePortRangeMax 보호대상 클러스터 보안그룹 규칙 허용 포트 범위 끝
	ProtectionClusterSecurityGroupRulePortRangeMax *uint64 `json:"protection_cluster_security_group_rule_port_range_max,omitempty"`

	// ProtectionClusterSecurityGroupRuleProtocol 보호대상 클러스터 보안그룹 규칙 프로토콜
	ProtectionClusterSecurityGroupRuleProtocol *string `json:"protection_cluster_security_group_rule_protocol,omitempty"`

	// ProtectionClusterSecurityGroupRuleEtherType 보호대상 클러스터 이더넷 타입
	ProtectionClusterSecurityGroupRuleEtherType uint32 `gorm:"not null" json:"protection_cluster_security_group_rule_ether_type,omitempty"`

	// RecoveryClusterSecurityGroupRuleID 복구대상 클러스터 보안그룹 규칙 ID
	RecoveryClusterSecurityGroupRuleID *uint64 `json:"recovery_cluster_security_group_rule_id,omitempty"`

	// RecoveryClusterSecurityGroupRuleRemoteSecurityGroupID 보호대상 복구스터 보안그룹 규칙 대상 보안그룹
	RecoveryClusterSecurityGroupRuleRemoteSecurityGroupID *uint64 `json:"recovery_cluster_security_group_rule_remote_security_group_id,omitempty"`

	// RecoveryClusterSecurityGroupRuleUUID 복구대상 클러스터 보안그룹 규칙 UUID
	RecoveryClusterSecurityGroupRuleUUID *string `json:"recovery_cluster_security_group_rule_uuid,omitempty"`

	// RecoveryClusterSecurityGroupRuleDescription 복구대상 클러스터 보안그룹 규칙 설명
	RecoveryClusterSecurityGroupRuleDescription *string `json:"recovery_cluster_security_group_rule_description,omitempty"`

	// RecoveryClusterSecurityGroupRuleNetworkCidr 보호대상 복구스터 보안그룹 규칙 네트워크 CIDR
	RecoveryClusterSecurityGroupRuleNetworkCidr *string `json:"recovery_cluster_security_group_rule_network_cidr,omitempty"`

	// RecoveryClusterSecurityGroupRuleDirection 복구대상 클러스터 보안그룹 규칙 방향
	RecoveryClusterSecurityGroupRuleDirection *string `json:"recovery_cluster_security_group_rule_direction,omitempty"`

	// RecoveryClusterSecurityGroupRulePortRangeMin 보호대상 클러스터 보안그룹 복구 허용 포트 범위 시작
	RecoveryClusterSecurityGroupRulePortRangeMin *uint64 `json:"recovery_cluster_security_group_rule_port_range_min,omitempty"`

	// RecoveryClusterSecurityGroupRulePortRangeMax 보호대상 클러스터 보안그룹 복구 허용 포트 범위 끝
	RecoveryClusterSecurityGroupRulePortRangeMax *uint64 `json:"recovery_cluster_security_group_rule_port_range_max,omitempty"`

	// RecoveryClusterSecurityGroupRuleProtocol 복구대상 클러스터 보안그룹 규칙 프로토콜
	RecoveryClusterSecurityGroupRuleProtocol *string `json:"recovery_cluster_security_group_rule_protocol,omitempty"`

	// RecoveryClusterSecurityGroupRuleEtherType 복구대상 클러스터 이더넷 타입
	RecoveryClusterSecurityGroupRuleEtherType *uint32 `json:"recovery_cluster_security_group_rule_ether_type,omitempty"`
}

// RecoveryResultInstance 인스턴스 복구결과
type RecoveryResultInstance struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterInstanceID 보호대상 클러스터 인스턴스 ID
	ProtectionClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_id,omitempty"`

	// ProtectionClusterInstanceUUID 보호대상 클러스터 인스턴스 UUID
	ProtectionClusterInstanceUUID string `gorm:"not null" json:"protection_cluster_instance_uuid,omitempty"`

	// ProtectionClusterInstanceName 보호대상 클러스터 인스턴스 이름
	ProtectionClusterInstanceName string `gorm:"not null" json:"protection_cluster_instance_name,omitempty"`

	// ProtectionClusterInstanceDescription 보호대상 클러스터 인스턴스 설명
	ProtectionClusterInstanceDescription *string `json:"protection_cluster_instance_description,omitempty"`

	// ProtectionClusterInstanceSpecID 보호대상 클러스터 인스턴스 Spec ID
	ProtectionClusterInstanceSpecID uint64 `gorm:"not null" json:"protection_cluster_instance_spec_id,omitempty"`

	// ProtectionClusterKeypairID 보호대상 클러스터 keypair ID
	ProtectionClusterKeypairID *uint64 `json:"protection_cluster_keypair_id,omitempty"`

	// ProtectionClusterAvailabilityZoneID 보호대상 클러스터 가용구역 ID
	ProtectionClusterAvailabilityZoneID uint64 `gorm:"not null" json:"protection_cluster_availability_zone_id,omitempty"`

	// ProtectionClusterAvailabilityZoneName 보호대상 클러스터 가용구역 이름
	ProtectionClusterAvailabilityZoneName string `gorm:"not null" json:"protection_cluster_availability_zone_name,omitempty"`

	// ProtectionClusterHypervisorID 보호대상 클러스터 Hypervisor ID
	ProtectionClusterHypervisorID uint64 `gorm:"not null" json:"protection_cluster_hypervisor_id,omitempty"`

	// ProtectionClusterHypervisorTypeCode 보호대상 클러스터 Hypervisor 종류
	ProtectionClusterHypervisorTypeCode string `gorm:"not null" json:"protection_cluster_hypervisor_type_code,omitempty"`

	// ProtectionClusterHypervisorHostname 보호대상 클러스터 Hypervisor Hostname
	ProtectionClusterHypervisorHostname string `gorm:"not null" json:"protection_cluster_hypervisor_hostname,omitempty"`

	// ProtectionClusterHypervisorIPAddress 보호대상 클러스터 Hypervisor IP 주소
	ProtectionClusterHypervisorIPAddress string `gorm:"not null" json:"protection_cluster_hypervisor_ip_address,omitempty"`

	// RecoveryClusterInstanceID 복구대상 클러스터 인스턴스 ID
	RecoveryClusterInstanceID *uint64 `json:"recovery_cluster_instance_id,omitempty"`

	// RecoveryClusterInstanceUUID 복구대상 클러스터 인스턴스 UUID
	RecoveryClusterInstanceUUID *string `json:"recovery_cluster_instance_uuid,omitempty"`

	// RecoveryClusterInstanceName 복구대상 클러스터 인스턴스 이름
	RecoveryClusterInstanceName *string `json:"recovery_cluster_instance_name,omitempty"`

	// RecoveryClusterInstanceDescription 복구대상 클러스터 인스턴스 설명
	RecoveryClusterInstanceDescription *string `json:"recovery_cluster_instance_description,omitempty"`

	// RecoveryClusterAvailabilityZoneID 복구대상 클러스터 가용구역 ID
	RecoveryClusterAvailabilityZoneID *uint64 `json:"recovery_cluster_availability_zone_id,omitempty"`

	// RecoveryClusterAvailabilityZoneName 복구대상 클러스터 가용구역 이름
	RecoveryClusterAvailabilityZoneName *string `json:"recovery_cluster_availability_zone_name,omitempty"`

	// RecoveryClusterHypervisorID 복구대상 클러스터 Hypervisor ID
	RecoveryClusterHypervisorID *uint64 `json:"recovery_cluster_hypervisor_id,omitempty"`

	// RecoveryClusterHypervisorTypeCode 복구대상 클러스터 Hypervisor 종류
	RecoveryClusterHypervisorTypeCode *string `json:"recovery_cluster_hypervisor_type_code,omitempty"`

	// RecoveryClusterHypervisorHostname 복구대상 클러스터 Hypervisor Hostname
	RecoveryClusterHypervisorHostname *string `json:"recovery_cluster_hypervisor_hostname,omitempty"`

	// RecoveryClusterHypervisorIPAddress 복구대상 클러스터 Hypervisor IP 주소
	RecoveryClusterHypervisorIPAddress *string `json:"recovery_cluster_hypervisor_ip_address,omitempty"`

	// RecoveryPointTypeCode 데이터 시점 유형
	RecoveryPointTypeCode string `gorm:"not null" json:"recovery_point_type_code,omitempty"`

	// RecoveryPoint 복구 데이터 시점
	RecoveryPoint int64 `gorm:"not null" json:"recovery_point,omitempty"`

	// AutoStartFlag 재해복구시 기동 여부
	AutoStartFlag bool `gorm:"not null" json:"auto_start_flag,omitempty"`

	// DiagnosisFlag 재해복구 기동 후 정상동작 확인 여부
	DiagnosisFlag bool `gorm:"not null" json:"diagnosis_flag,omitempty"`

	// DiagnosisMethodCode 재해복구 기동 후 정상동작 확인 방법
	DiagnosisMethodCode *string `json:"diagnosis_method_code,omitempty"`

	// DiagnosisMethodData 재해복구 기동 후 정상동작 확인 방법 데이터
	DiagnosisMethodData *string `json:"diagnosis_method_data,omitempty"`

	// DiagnosisTimeout 재해복구 기동 후 정상동작 확인 대기 시간
	DiagnosisTimeout *uint32 `json:"diagnosis_timeout,omitempty"`

	// StartedAt 복구 시작 일시
	StartedAt *int64 `json:"started_at,omitempty"`

	// FinishedAt 복구 종료 일시
	FinishedAt *int64 `json:"finished_at,omitempty"`

	// ResultCode 복구 결과
	ResultCode string `gorm:"not null" json:"result_code,omitempty"`

	// FailedReasonCode 실패 사유 메시지 코드
	FailedReasonCode *string `json:"failed_reason_code,omitempty"`

	// FailedReasonContents 실패 사유 메시지 데이터
	FailedReasonContents *string `json:"failed_reason_contents,omitempty"`
}

// RecoveryResultInstanceDependency 인스턴스 기동 의존성
type RecoveryResultInstanceDependency struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterInstanceID 보호대상 클러스터 인스턴스 ID
	ProtectionClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_id,omitempty"`

	// DependencyProtectionClusterInstanceID 보호대상 클러스터 인스턴스의 선행 인스턴스 ID
	DependencyProtectionClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"dependency_protection_cluster_instance_id,omitempty"`
}

// RecoveryResultInstanceNetwork 인스턴스 네트워크 복구결과
type RecoveryResultInstanceNetwork struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterInstanceID 보호대상 클러스터 인스턴스 ID
	ProtectionClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_id,omitempty"`

	// InstanceNetworkSeq 인스턴스 네트워크 seq
	InstanceNetworkSeq uint64 `gorm:"primary_key;autoIncrement:false" json:"instance_network_seq,omitempty"`

	// ProtectionClusterNetworkID 보호대상 클러스터 네트워크 ID
	ProtectionClusterNetworkID uint64 `gorm:"not null" json:"protection_cluster_network_id,omitempty"`

	// ProtectionClusterSubnetID 보호대상 클러스터 서브넷 ID
	ProtectionClusterSubnetID uint64 `gorm:"not null" json:"protection_cluster_subnet_id,omitempty"`

	// ProtectionClusterFloatingIPID 보호대상 클러스터 Floating IP ID
	ProtectionClusterFloatingIPID *uint64 `json:"protection_cluster_floating_ip_id,omitempty"`

	// ProtectionClusterDhcpFlag 보호대상 클러스터 DHCP 여부
	ProtectionClusterDhcpFlag bool `gorm:"not null" json:"protection_cluster_dhcp_flag,omitempty"`

	// ProtectionClusterIPAddress 보호대상 클러스터 IP 주소
	ProtectionClusterIPAddress string `gorm:"not null" json:"protection_cluster_ip_address,omitempty"`

	// RecoveryClusterDhcpFlag 복구대상 클러스터 DHCP 여부
	RecoveryClusterDhcpFlag *bool `json:"recovery_cluster_dhcp_flag,omitempty"`

	// RecoveryClusterIPAddress 복구대상 클러스터 IP 주소
	RecoveryClusterIPAddress *string `json:"recovery_cluster_ip_address,omitempty"`
}

// RecoveryResultInstanceSecurityGroup 인스턴스 보안그룹 복구결과
type RecoveryResultInstanceSecurityGroup struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterInstanceID 보호대상 클러스터 인스턴스 ID
	ProtectionClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_id,omitempty"`

	// ProtectionClusterSecurityGroupID 보호대상 클러스터 보안그룹 ID
	ProtectionClusterSecurityGroupID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_security_group_id,omitempty"`
}

// RecoveryResultVolume 볼륨 복구결과
type RecoveryResultVolume struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterVolumeID 보호대상 클러스터 볼륨 ID
	ProtectionClusterVolumeID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_volume_id,omitempty"`

	// ProtectionClusterVolumeUUID 보호대상 클러스터 볼륨 UUID
	ProtectionClusterVolumeUUID string `gorm:"not null" json:"protection_cluster_volume_uuid,omitempty"`

	// ProtectionClusterVolumeName 보호대상 클러스터 볼륨 이름
	ProtectionClusterVolumeName *string `json:"protection_cluster_volume_name,omitempty"`

	// ProtectionClusterVolumeDescription 보호대상 클러스터 볼륨 설명
	ProtectionClusterVolumeDescription *string `json:"protection_cluster_volume_description,omitempty"`

	// ProtectionClusterVolumeSizeBytes 보호대상 클러스터 볼륨 Size
	ProtectionClusterVolumeSizeBytes uint64 `gorm:"not null" json:"protection_cluster_volume_size_bytes,omitempty"`

	// ProtectionClusterVolumeMultiattach 보호대상 클러스터 볼륨 Multiattach 여부
	ProtectionClusterVolumeMultiattach bool `gorm:"not null" json:"protection_cluster_volume_multiattach,omitempty"`

	// ProtectionClusterVolumeBootable 보호대상 클러스터 볼륨 Bootable 여부
	ProtectionClusterVolumeBootable bool `gorm:"not null" json:"protection_cluster_volume_bootable,omitempty"`

	// ProtectionClusterVolumeReadonly 보호대상 클러스터 볼륨 Readonly 여부
	ProtectionClusterVolumeReadonly bool `gorm:"not null" json:"protection_cluster_volume_readonly,omitempty"`

	// ProtectionClusterStorageID 보호대상 클러스터 스토리지 ID
	ProtectionClusterStorageID uint64 `gorm:"not null" json:"protection_cluster_storage_id,omitempty"`

	// ProtectionClusterStorageUUID 보호대상 클러스터 스토리지 UUID
	ProtectionClusterStorageUUID string `gorm:"not null" json:"protection_cluster_storage_uuid,omitempty"`

	// ProtectionClusterStorageName 보호대상 클러스터 스토리지 이름
	ProtectionClusterStorageName string `gorm:"not null" json:"protection_cluster_storage_name,omitempty"`

	// ProtectionClusterStorageDescription 보호대상 클러스터 스토리지 설명
	ProtectionClusterStorageDescription *string `json:"protection_cluster_storage_description,omitempty"`

	// ProtectionClusterStorageTypeCode 보호대상 클러스터 스토리지 종류
	ProtectionClusterStorageTypeCode string `gorm:"not null" json:"protection_cluster_storage_type_code,omitempty"`

	// RecoveryClusterVolumeID 복구대상 클러스터 볼륨 ID
	RecoveryClusterVolumeID *uint64 `json:"recovery_cluster_volume_id,omitempty"`

	// RecoveryClusterVolumeUUID 복구대상 클러스터 볼륨 UUID
	RecoveryClusterVolumeUUID *string `json:"recovery_cluster_volume_uuid,omitempty"`

	// RecoveryClusterVolumeName 복구대상 클러스터 볼륨 이름
	RecoveryClusterVolumeName *string `json:"recovery_cluster_volume_name,omitempty"`

	// RecoveryClusterVolumeDescription 복구대상 클러스터 볼륨 설명
	RecoveryClusterVolumeDescription *string `json:"recovery_cluster_volume_description,omitempty"`

	// RecoveryClusterVolumeSizeBytes 복구대상 클러스터 볼륨 Size
	RecoveryClusterVolumeSizeBytes *uint64 `json:"recovery_cluster_volume_size_bytes,omitempty"`

	// RecoveryClusterVolumeMultiattach 복구대상 클러스터 볼륨 Multiattach 여부
	RecoveryClusterVolumeMultiattach *bool `json:"recovery_cluster_volume_multiattach,omitempty"`

	// RecoveryClusterVolumeBootable 복구대상 클러스터 볼륨 Bootable 여부
	RecoveryClusterVolumeBootable *bool `json:"recovery_cluster_volume_bootable,omitempty"`

	// RecoveryClusterVolumeReadonly 복구대상 클러스터 볼륨 Readonly 여부
	RecoveryClusterVolumeReadonly *bool `json:"recovery_cluster_volume_readonly,omitempty"`

	// RecoveryClusterStorageID 복구대상 클러스터 스토리지 ID
	RecoveryClusterStorageID *uint64 `json:"recovery_cluster_storage_id,omitempty"`

	// RecoveryClusterStorageUUID 복구대상 클러스터 스토리지 UUID
	RecoveryClusterStorageUUID *string `json:"recovery_cluster_storage_uuid,omitempty"`

	// RecoveryClusterStorageName 복구대상 클러스터 스토리지 이름
	RecoveryClusterStorageName *string `json:"recovery_cluster_storage_name,omitempty"`

	// RecoveryClusterStorageDescription 복구대상 클러스터 스토리지 설명
	RecoveryClusterStorageDescription *string `json:"recovery_cluster_storage_description,omitempty"`

	// RecoveryClusterStorageTypeCode 복구대상 클러스터 스토리지 종류
	RecoveryClusterStorageTypeCode *string `json:"recovery_cluster_storage_type_code,omitempty"`

	// RecoveryPointTypeCode 데이터 시점 유형
	RecoveryPointTypeCode string `gorm:"not null" json:"recovery_point_type_code,omitempty"`

	// RecoveryPoint 복구 데이터 시점
	RecoveryPoint int64 `gorm:"not null" json:"recovery_point,omitempty"`

	// StartedAt 복구 시작 일시
	StartedAt *int64 `json:"started_at,omitempty"`

	// FinishedAt 복구 종료 일시
	FinishedAt *int64 `json:"finished_at,omitempty"`

	// ResultCode 복구 결과
	ResultCode string `gorm:"not null" json:"result_code,omitempty"`

	// FailedReasonCode 복구 실패 사유 메시지 코드
	FailedReasonCode *string `json:"failed_reason_code,omitempty"`

	// FailedReasonContents 복구 실패 사유 메시지 데이터
	FailedReasonContents *string `json:"failed_reason_contents,omitempty"`

	// RollbackFlag 롤백 여부
	RollbackFlag bool `gorm:"not null" json:"rollback_flag,omitempty"`
}

// RecoveryResultInstanceVolume 인스턴스 볼륨 복구결과
type RecoveryResultInstanceVolume struct {
	// RecoveryResultID 재해복구결과 ID
	RecoveryResultID uint64 `gorm:"primary_key;autoIncrement:false" json:"recovery_result_id,omitempty"`

	// ProtectionClusterTenantID 보호대상 클러스터 테넌트 ID
	ProtectionClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_tenant_id,omitempty"`

	// ProtectionClusterInstanceID 보호대상 클러스터 인스턴스 ID
	ProtectionClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_instance_id,omitempty"`

	// ProtectionClusterVolumeID 보호대상 클러스터 볼륨 ID
	ProtectionClusterVolumeID uint64 `gorm:"primary_key;autoIncrement:false" json:"protection_cluster_volume_id,omitempty"`

	// ProtectionClusterDevicePath 보호대상 클러스터 디바이스 경로
	ProtectionClusterDevicePath string `gorm:"not null" json:"protection_cluster_device_path,omitempty"`

	// ProtectionClusterBootIndex 보호대상 클러스터 부팅 순서
	ProtectionClusterBootIndex int64 `gorm:"not null" json:"protection_cluster_boot_index,omitempty"`

	// RecoveryClusterDevicePath 복구대상 클러스터 디바이스 경로
	RecoveryClusterDevicePath *string `json:"recovery_cluster_device_path,omitempty"`

	// RecoveryClusterBootIndex 복구대상 클러스터 부팅 순서
	RecoveryClusterBootIndex *int64 `json:"recovery_cluster_boot_index,omitempty"`
}

// InstanceTemplate 인스턴스 템플릿
type InstanceTemplate struct {
	// ID 는 자동으로 증가되는 템플릿의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// OwnerGroupID 는 인스턴스 템플릿의 소유 그룹 ID 이다
	OwnerGroupID uint64 `gorm:"not null" json:"owner_group_id,omitempty"`

	// Name 는 템플릿의 이름이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Remarks 는 템플릿의 비고이다.
	Remarks *string `json:"remarks,omitempty"`

	// CreatedAt 는 자동으로 생성되는 템플릿의 생성날짜이다.
	CreatedAt int64 `gorm:"not null" json:"created_at,omitempty"`
}

// InstanceTemplateInstance 인스턴스 템플릿의 인스턴스
type InstanceTemplateInstance struct {
	// InstanceTemplateID 는 인스턴스 템플릿의 식별자이다.
	InstanceTemplateID uint64 `gorm:"primary_key;autoIncrement:false" json:"instance_template_id,omitempty"`

	// ProtectionClusterInstanceName 는 인스턴스의 이름이다.
	ProtectionClusterInstanceName string `gorm:"primary_key;not null" json:"protection_cluster_instance_name,omitempty"`

	// AutoStartFlag 는 인스턴스 자동 실행 여부 이다.
	AutoStartFlag *bool `json:"auto_start_flag,omitempty"`

	// DiagnosisFlag 는 인스턴스 정상 작동 확인 여부 이다.
	DiagnosisFlag *bool `json:"diagnosis_flag,omitempty"`

	// DiagnosisMethodCode 는 인스턴스 정상 작동 확인 방법 이다.
	DiagnosisMethodCode *string `json:"diagnosis_method_code,omitempty"`

	// DiagnosisMethodData 는 인스턴스 정상 작동 확인 방법 데이터 이다.
	DiagnosisMethodData *string `json:"diagnosis_method_data,omitempty"`

	// DiagnosisTimeout 는 인스턴스 정상 작동 확인 대기 시간 이다.
	DiagnosisTimeout *uint32 `json:"diagnosis_timeout,omitempty"`
}

type InstanceTemplateInstanceDependency struct {
	// InstanceTemplateID 는 인스턴스 템플릿의 식별자이다.
	InstanceTemplateID uint64 `gorm:"primary_key;not null" json:"instance_template_id,omitempty"`

	// InstanceTemplateInstanceName 는 인스턴스 템플릿에 저장 인스턴스의 이름이다.
	ProtectionClusterInstanceName string `gorm:"primary_key;not null" json:"protection_cluster_instance_name,omitempty"`

	// DependInstanceTemplateInstanceName 는 인스턴스 템플릿에 저장된 인스턴스의 의존 인스턴스 이름이다.
	DependProtectionClusterInstanceName string `gorm:"primary_key;not null" json:"depend_protection_cluster_instance_name,omitempty"`
}

// BeforeCreate 는 Record 수정 전 호출 되는 함수
func (g *ProtectionGroup) BeforeCreate() error {
	g.CreatedAt = time.Now().Unix()
	g.UpdatedAt = time.Now().Unix()
	return nil
}

// BeforeUpdate 는 Record 수정 전 호출 되는 함수
func (g *ProtectionGroup) BeforeUpdate() error {
	g.UpdatedAt = time.Now().Unix()
	return nil
}

// BeforeCreate 는 Record 수정 전 호출 되는 함수
func (p *Plan) BeforeCreate() error {
	p.CreatedAt = time.Now().Unix()
	p.UpdatedAt = time.Now().Unix()
	return nil
}

// BeforeUpdate 는 Record 수정 전 호출 되는 함수
func (p *Plan) BeforeUpdate() error {
	p.UpdatedAt = time.Now().Unix()
	return nil
}

// BeforeCreate 는 Record 생성 전 호출 되는 함수
func (t *InstanceTemplate) BeforeCreate() error {
	t.CreatedAt = time.Now().Unix()
	return nil
}

// TableName ProtectionGroup 테이블 명 반환 함수
func (ProtectionGroup) TableName() string {
	return "cdm_disaster_recovery_protection_group"
}

// TableName ProtectionInstance 테이블 명 반환 함수
func (ProtectionInstance) TableName() string {
	return "cdm_disaster_recovery_protection_instance"
}

// TableName ProtectionGroupSnapshot 테이블 명 반환 함수
func (ProtectionGroupSnapshot) TableName() string {
	return "cdm_disaster_recovery_protection_group_snapshot"
}

// TableName ConsistencyGroup 테이블 명 반환 함수
func (ConsistencyGroup) TableName() string {
	return "cdm_disaster_recovery_consistency_group"
}

// TableName ConsistencyGroupSnapshot 테이블 명 반환 함수
func (ConsistencyGroupSnapshot) TableName() string {
	return "cdm_disaster_recovery_consistency_group_snapshot"
}

// TableName Plan 테이블 명 반환 함수
func (Plan) TableName() string {
	return "cdm_disaster_recovery_plan"
}

// TableName PlanTenant 테이블 명 반환 함수
func (PlanTenant) TableName() string {
	return "cdm_disaster_recovery_plan_tenant"
}

// TableName PlanAvailabilityZone 테이블 명 반환 함수
func (PlanAvailabilityZone) TableName() string {
	return "cdm_disaster_recovery_plan_availability_zone"
}

// TableName PlanExternalNetwork 테이블 명 반환 함수
func (PlanExternalNetwork) TableName() string {
	return "cdm_disaster_recovery_plan_external_network"
}

// TableName PlanRouter 테이블 명 반환 함수
func (PlanRouter) TableName() string {
	return "cdm_disaster_recovery_plan_router"
}

// TableName PlanExternalRoutingInterface 테이블 명 반환 함수
func (PlanExternalRoutingInterface) TableName() string {
	return "cdm_disaster_recovery_plan_external_routing_interface"
}

// TableName PlanStorage 테이블 명 반환 함수
func (PlanStorage) TableName() string {
	return "cdm_disaster_recovery_plan_storage"
}

// TableName PlanInstance 테이블 명 반환 함수
func (PlanInstance) TableName() string {
	return "cdm_disaster_recovery_plan_instance"
}

// TableName PlanInstanceDependency 테이블 명 반환 함수
func (PlanInstanceDependency) TableName() string {
	return "cdm_disaster_recovery_plan_instance_dependency"
}

// TableName PlanInstanceFloatingIP 테이블 명 반환 함수
func (PlanInstanceFloatingIP) TableName() string {
	return "cdm_disaster_recovery_plan_instance_floating_ip"
}

// TableName PlanVolume 테이블 명 반환 함수
func (PlanVolume) TableName() string {
	return "cdm_disaster_recovery_plan_volume"
}

// TableName Job 테이블 명 반환 함수
func (Job) TableName() string {
	return "cdm_disaster_recovery_job"
}

// TableName RecoveryResult 테이블 명 반환 함수
func (RecoveryResult) TableName() string {
	return "cdm_disaster_recovery_result"
}

// TableName RecoveryResultRaw 테이블 명 반환 함수
func (RecoveryResultRaw) TableName() string {
	return "cdm_disaster_recovery_result_raw"
}

// TableName RecoveryResultWarningReason 테이블 명 반환 함수
func (RecoveryResultWarningReason) TableName() string {
	return "cdm_disaster_recovery_result_warning_reason"
}

// TableName RecoveryResultFailedReason 테이블 명 반환 함수
func (RecoveryResultFailedReason) TableName() string {
	return "cdm_disaster_recovery_result_failed_reason"
}

// TableName RecoveryResultTaskLog 테이블 명 반환 함수
func (RecoveryResultTaskLog) TableName() string {
	return "cdm_disaster_recovery_result_task_log"
}

// TableName RecoveryResultTenant 테이블 명 반환 함수
func (RecoveryResultTenant) TableName() string {
	return "cdm_disaster_recovery_result_tenant"
}

// TableName RecoveryResultInstanceSpec 테이블 명 반환 함수
func (RecoveryResultInstanceSpec) TableName() string {
	return "cdm_disaster_recovery_result_instance_spec"
}

// TableName RecoveryResultInstanceExtraSpec 테이블 명 반환 함수
func (RecoveryResultInstanceExtraSpec) TableName() string {
	return "cdm_disaster_recovery_result_instance_extra_spec"
}

// TableName RecoveryResultKeypair 테이블 명 반환 함수
func (RecoveryResultKeypair) TableName() string {
	return "cdm_disaster_recovery_result_keypair"
}

// TableName RecoveryResultNetwork 테이블 명 반환 함수
func (RecoveryResultNetwork) TableName() string {
	return "cdm_disaster_recovery_result_network"
}

// TableName RecoveryResultSubnet 테이블 명 반환 함수
func (RecoveryResultSubnet) TableName() string {
	return "cdm_disaster_recovery_result_subnet"
}

// TableName RecoveryResultSubnetDHCPPool 테이블 명 반환 함수
func (RecoveryResultSubnetDHCPPool) TableName() string {
	return "cdm_disaster_recovery_result_subnet_dhcp_pool"
}

// TableName RecoveryResultSubnetNameserver 테이블 명 반환 함수
func (RecoveryResultSubnetNameserver) TableName() string {
	return "cdm_disaster_recovery_result_subnet_nameserver"
}

// TableName RecoveryResultFloatingIP 테이블 명 반환 함수
func (RecoveryResultFloatingIP) TableName() string {
	return "cdm_disaster_recovery_result_floating_ip"
}

// TableName RecoveryResultRouter 테이블 명 반환 함수
func (RecoveryResultRouter) TableName() string {
	return "cdm_disaster_recovery_result_router"
}

// TableName RecoveryResultInternalRoutingInterface 테이블 명 반환 함수
func (RecoveryResultInternalRoutingInterface) TableName() string {
	return "cdm_disaster_recovery_result_internal_routing_interface"
}

// TableName RecoveryResultExternalRoutingInterface 테이블 명 반환 함수
func (RecoveryResultExternalRoutingInterface) TableName() string {
	return "cdm_disaster_recovery_result_external_routing_interface"
}

// TableName RecoveryResultExtraRoute 테이블 명 반환 함수
func (RecoveryResultExtraRoute) TableName() string {
	return "cdm_disaster_recovery_result_extra_route"
}

// TableName RecoveryResultSecurityGroup 테이블 명 반환 함수
func (RecoveryResultSecurityGroup) TableName() string {
	return "cdm_disaster_recovery_result_security_group"
}

// TableName RecoveryResultSecurityGroupRule 테이블 명 반환 함수
func (RecoveryResultSecurityGroupRule) TableName() string {
	return "cdm_disaster_recovery_result_security_group_rule"
}

// TableName RecoveryResultInstance 테이블 명 반환 함수
func (RecoveryResultInstance) TableName() string {
	return "cdm_disaster_recovery_result_instance"
}

// TableName RecoveryResultInstanceDependency 테이블 명 반환 함수
func (RecoveryResultInstanceDependency) TableName() string {
	return "cdm_disaster_recovery_result_instance_dependency"
}

// TableName RecoveryResultInstanceNetwork 테이블 명 반환 함수
func (RecoveryResultInstanceNetwork) TableName() string {
	return "cdm_disaster_recovery_result_instance_network"
}

// TableName RecoveryResultInstanceSecurityGroup 테이블 명 반환 함수
func (RecoveryResultInstanceSecurityGroup) TableName() string {
	return "cdm_disaster_recovery_result_instance_security_group"
}

// TableName RecoveryResultVolume 테이블 명 반환 함수
func (RecoveryResultVolume) TableName() string {
	return "cdm_disaster_recovery_result_volume"
}

// TableName RecoveryResultInstanceVolume 테이블 명 반환 함수
func (RecoveryResultInstanceVolume) TableName() string {
	return "cdm_disaster_recovery_result_instance_volume"
}

func (InstanceTemplate) TableName() string {
	return "cdm_disaster_recovery_instance_template"
}

func (InstanceTemplateInstance) TableName() string {
	return "cdm_disaster_recovery_instance_template_instance"
}

func (InstanceTemplateInstanceDependency) TableName() string {
	return "cdm_disaster_recovery_instance_template_instance_dependency"
}
