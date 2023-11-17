package model

import (
	"time"
)

// Cluster 클러스터
type Cluster struct {
	// ID 는 자동 증가 되는 클러스터의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// TenantID 는 cdm-cloud 테넌트 ID 이다.
	TenantID uint64 `gorm:"not null" json:"tenant_id,omitempty"`

	// OwnerGroupID 는 클러스터의 소유 권이 있는 cdm-cloud 의 사용자 그룹 ID 이다.
	OwnerGroupID uint64 `gorm:"not null" json:"owner_group_id,omitempty"`

	// Name 는 클러스터의 이름 이다.
	Name string `gorm:"not null;unique" json:"name,omitempty"`

	// Remarks 는 클러스터의 비고 내용 이다.
	Remarks *string `json:"remarks,omitempty"`

	// TypeCode 는 클러스터의 종류 이다.
	TypeCode string `gorm:"not null" json:"type_code,omitempty"`

	// APIServerURL 는 클러스터의 API Server URL 이다.
	APIServerURL string `gorm:"not null" json:"api_server_url,omitempty"`

	// Credential 는 클러스터의 API Credential 이다.
	Credential string `gorm:"not null" json:"credential,omitempty"`

	// StateCode 는 클러스터 상태 코드이다.
	StateCode string `gorm:"not null" json:"state_code,omitempty"`

	// CreatedAt 클러스터가 생성된 (UnixTime)시간 이다.
	CreatedAt int64 `gorm:"not null" json:"created_at,omitempty"`

	// UpdatedAt 클러스터가 수정된 (UnixTime)시간 이다.
	UpdatedAt int64 `gorm:"not null" json:"updated_at,omitempty"`

	// SynchronizedAt 클러스터가 동기화된 (UnixTime)시간 이다.
	SynchronizedAt int64 `gorm:"not null" json:"synchronized_at,omitempty"`
}

// ClusterPermission 클러스터 권한
type ClusterPermission struct {
	// ClusterID 는 클러스터의 ID 이다.
	ClusterID uint64 `gorm:"primary_key;autoIncrement:false" json:"cluster_id,omitempty"`

	// GroupID 는 cdm-cloud 의 사용자 그룹 ID 이다.
	GroupID uint64 `gorm:"primary_key;autoIncrement:false" json:"group_id,omitempty"`

	// ModeCode 는 클러스터에 대한 그룹의 권한이다.
	ModeCode string `gorm:"not null" json:"mode_code,omitempty"`
}

// ClusterHypervisor 클러스터의 하이퍼바이저
type ClusterHypervisor struct {
	// ID 는 자동 증가 되는 하이퍼바이저의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterID 는 클러스터의 ID 이다.
	ClusterID uint64 `gorm:"not null" json:"cluster_id,omitempty"`

	// ClusterAvailabilityZoneID 는 가용 구역의 ID 이다.
	ClusterAvailabilityZoneID uint64 `gorm:"not null" json:"cluster_availability_zone_id,omitempty"`

	// UUID 는 하이퍼바이저 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// TypeCode 는 하이퍼바이저 종류 코드이다.
	TypeCode string `gorm:"not null" json:"type_code,omitempty"`

	// Hostname 는 하이퍼바이저의 호스트 이름.
	Hostname string `gorm:"not null" json:"hostname,omitempty"`

	// IPAddress 는 하이퍼바이저의 IP 주소이다.
	IPAddress string `gorm:"not null" json:"ip_address,omitempty"`

	// VcpuTotalCnt 는 하이퍼바이저의 CPU 코어 갯수 이다.
	VcpuTotalCnt uint32 `gorm:"not null" json:"vcpu_total_cnt,omitempty"`

	// VcpuTotalCnt 는 하이퍼바이저의 CPU 코어 갯수 이다.
	VcpuUsedCnt uint32 `gorm:"not null" json:"vcpu_used_cnt,omitempty"`

	// MemTotalBytes 는 하이퍼바이저의 총 메모리 용량이다.
	MemTotalBytes uint64 `gorm:"not null" json:"mem_total_bytes,omitempty"`

	// MemUsedBytes 는 하이퍼바이저의 총 메모리 용량이다.
	MemUsedBytes uint64 `gorm:"not null" json:"mem_used_bytes,omitempty"`

	// DiskTotalBytes 는 하이퍼바이저의 총 디스크 용량이다.
	DiskTotalBytes uint64 `gorm:"not null" json:"disk_total_bytes,omitempty"`

	// DiskUsedBytes 는 하이퍼바이저의 총 디스크 용량이다.
	DiskUsedBytes uint64 `gorm:"not null" json:"disk_used_bytes,omitempty"`

	// State 는 하이퍼바이저의 state 이다.
	State string `gorm:"not null" json:"state,omitempty"`

	// Status 는 하이퍼바이저의 status 이다.
	Status string `gorm:"not null" json:"status,omitempty"`

	// SSHPort 는 SSH 접속을 위한 포트이다.
	SSHPort *uint32 `json:"ssh_port,omitempty"`

	// SSHAccount 는 SSH 접속을 위한 계정이다.
	SSHAccount *string `json:"ssh_account,omitempty"`

	// SSHPassword 는 SSH 접속을 위한 비밀 번호이다.
	SSHPassword *string `json:"ssh_password,omitempty"`

	// AgentPort 는 Agent 가 사용하는 PORT 이다.
	AgentPort *uint32 `json:"agent_port,omitempty"`

	// AgentVersion 는 Agent 의 버전 정보 이다.
	AgentVersion *string `json:"agent_version,omitempty"`

	// AgentInstalledAt 는 Agent 가 최초 설지 된 (UnixTime)시간 이다.
	AgentInstalledAt *int64 `json:"agent_installed_at,omitempty"`

	// AgentLastUpgradedAt 는 Agent 가 최종 업데이트 가 된 (UnixTime)시간 이다.
	AgentLastUpgradedAt *int64 `json:"agent_last_upgraded_at,omitempty"`

	// Raw 는 하이퍼바이저의 raw 이다.
	Raw *string `json:"raw,omitempty"`
}

// ClusterAvailabilityZone 는 클러스의 가용 구역이다.
type ClusterAvailabilityZone struct {
	// ID 는 자동 증가되는 가용 구역의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterID 는 클러스터의 ID 이다
	ClusterID uint64 `gorm:"not null" json:"cluster_id,omitempty"`

	// Name 는 가용 구역의 이름 이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Available 는 가용 구역의 활성화 여부이다.
	Available bool `gorm:"not null" json:"available,omitempty"`

	// Raw 는 가용구역의 raw 이다.
	Raw *string `json:"raw,omitempty"`
}

// ClusterTenant 는 클러스터의 테넌트 이다.
type ClusterTenant struct {
	// ID 는 자동 증가 되는 클러스터의 테넌트 식별자 이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterID 는 클러스터의 ID 이다
	ClusterID uint64 `gorm:"not null" json:"cluster_id,omitempty"`

	// UUID 는 클러스터 테넌트의 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// Name 는 테넌트의 이름이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Description 테넌트의 설명이다.
	Description *string `json:"description,omitempty"`

	// Enabled 테넌트의 사용여부이다.
	Enabled bool `gorm:"not null" json:"enabled,omitempty"`

	// Raw 는 테넌트의 raw 이다.
	Raw *string `json:"raw,omitempty"`
}

// ClusterQuota 는 클러스터의 쿼타이다.
type ClusterQuota struct {
	// ClusterTenantID
	ClusterTenantID uint64 `gorm:"primary_key;autoIncrement:false" json:"cluster_tenant_id,omitempty"`

	// Key 쿼타의 키 값이다.
	Key string `gorm:"primary_key" json:"key,omitempty"`

	// Value 쿼타의 값이다.
	Value int64 `gorm:"not null" json:"value,omitempty"`
}

// ClusterNetwork 는 클러스터의 네트워크 이다.
type ClusterNetwork struct {
	// ID 는 자동 증가 되는 네트워크의 식별자 이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterTenantID 는 테넌트의 ID 이다.
	ClusterTenantID uint64 `gorm:"not null" json:"cluster_tenant_id,omitempty"`

	// UUID 는 네트워크의 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// Name 는 네트워크의 이름이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Description 는 네트워크 설명이다.
	Description *string `json:"description,omitempty"`

	// TypeCode 는 네트워크 종류이다.
	TypeCode string `gorm:"not null" json:"type_code,omitempty"`

	// ExternalFlag 는 외부 네트여크 여부 이다.
	ExternalFlag bool `gorm:"not null" json:"external_flag,omitempty"`

	// State 는 네트워크의 state 이다.
	State string `gorm:"not null" json:"state,omitempty"`

	// Status 는 네트워크의 status 이다.
	Status string `gorm:"not null" json:"status,omitempty"`

	// Raw 는 네트워크의 raw 이다.
	Raw *string `json:"raw,omitempty"`
}

// ClusterFloatingIP 는 클러스터의 플로팅 아이피이다.
type ClusterFloatingIP struct {
	// ID 는 자동 증가되는 플로팅 아이피의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	//ClusterTenantID 클러스터 테넌트의 ID 이다.
	ClusterTenantID uint64 `gorm:"not null" json:"cluster_tenant_id,omitempty"`

	//ClusterNetworkID 클러스터 네트워크의 ID 이다.
	ClusterNetworkID uint64 `gorm:"not null" json:"cluster_network_id,omitempty"`

	// UUID 는 플로팅 아이피의 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// Description 는 플로팅 아이피의 설명이다.
	Description *string `json:"description,omitempty"`

	// IPAddress 플로팅 아이피의 아이피주소이다.
	IPAddress string `gorm:"not null" json:"ip_address,omitempty"`

	// Status 는 플로팅 아이피의 status 이다.
	Status string `gorm:"not null" json:"status,omitempty"`

	// Raw 는 플로팅 아이피의 raw 이다.
	Raw *string `json:"raw,omitempty"`
}

// ClusterSubnet 는 클러스터 네트워크의 Subnet 이다.
type ClusterSubnet struct {
	// ID 는 자동 증가 되는 네트워크의 식별자 이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterNetworkID 는 네트워크의 ID 이다.
	ClusterNetworkID uint64 `gorm:"not null" json:"cluster_network_id,omitempty"`

	// UUID 는 네트워크 서브넷의 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// Name 는 네트워크 서브넷의 이름 이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Description 는 네트워크 서브넷의 설명이다.
	Description *string `json:"description,omitempty"`

	// NetworkCidr 는 네트워크 서브넷의 CIDR 이다.
	NetworkCidr string `gorm:"not null" json:"network_cidr,omitempty"`

	// DHCPEnabled 는 DHCP 사용 여부이다.
	DHCPEnabled bool `gorm:"not null" json:"dhcp_enabled,omitempty"`

	// GatewayFlag 는 게이트 웨이 유무이다.
	GatewayEnabled bool `gorm:"not null" json:"gateway_enabled,omitempty"`

	// GatewayIPAddress 는 게이트 웨이 주소이다.
	GatewayIPAddress *string `json:"gateway_ip_address,omitempty"`

	// IPV6AddressModeCode 는 ipv6 의 ip 주소 할당 방법이다.
	Ipv6AddressModeCode *string `json:"ipv6_address_mode_code,omitempty"`

	// IPV6RaModeCode 는 router advertisement 방법이다.
	Ipv6RaModeCode *string `json:"ipv6_ra_mode_code,omitempty"`

	// Raw 는 네트워크 서브넷의 raw 이다.
	Raw *string `json:"raw,omitempty"`
}

// ClusterSubnetNameserver 는 네임서버이다.
type ClusterSubnetNameserver struct {
	// ID 는 자동 증가되는 서브넷 네임서버의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterSubnetID 는 클러스터 서브넷의 ID 이다.
	ClusterSubnetID uint64 `gorm:"not null" json:"cluster_subnet_id,omitempty"`

	// Nameserver 는 클러스터 서브넷의 네임서버이다.
	Nameserver string `gorm:"not null" json:"nameserver,omitempty"`
}

// ClusterRouter 는 클러스터 네트워크 Router 이다.
type ClusterRouter struct {
	// ID 는 자동 증가되는 네트워크 Router 의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterTenantID 는 클러스터의 테넌트 ID 이다.
	ClusterTenantID uint64 `gorm:"not null" json:"cluster_tenant_id,omitempty"`

	// UUID Router의 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// Name Router의 이름 이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Description Router 의 설명이다.
	Description *string `json:"description,omitempty"`

	// State 는 라우터의 state 이다.
	State string `gorm:"not null" json:"state,omitempty"`

	// Status 는 라우터의 status 이다.
	Status string `gorm:"not null" json:"status,omitempty"`

	// Raw 는 라우터의 raw 이다.
	Raw *string `json:"raw,omitempty"`
}

// ClusterNetworkRoutingInterface 는 클러스터 네트워크 Router 의 인터페이스 목록 이다.
type ClusterNetworkRoutingInterface struct {
	// ClusterRouterID 는 Router 의 ID 이다.
	ClusterRouterID uint64 `gorm:"primary_key;autoIncrement:false" json:"cluster_router_id,omitempty"`

	// ClusterSubnetID 는 네트워크 서브넷 ID 이다.
	ClusterSubnetID uint64 `gorm:"primary_key;autoIncrement:false" json:"cluster_subnet_id,omitempty"`

	// IPAddress 는 Router 인터페이스의 IP 주소 이다.
	IPAddress string `gorm:"not null" json:"ip_address,omitempty"`

	// ExternalFlag 는 외부 라우팅 인터페이스 여부이다.
	ExternalFlag bool `gorm:"not null" json:"-"`
}

// ClusterRouterExtraRoute 는 클러스터 네트워크 Router 의 추가 라우트 정보이다.
type ClusterRouterExtraRoute struct {
	// ID 는 자동 증가되는 네트워크 추가 라우트 정보의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterRouterID 는 Router 의 ID 이다.
	ClusterRouterID uint64 `gorm:"primary_key;autoIncrement:false" json:"cluster_router_id,omitempty"`

	// Destination 는 목적지 CIDR 이다.
	Destination string `gorm:"not null" json:"destination,omitempty"`

	// ExternalFlag 는 목적지에 대한 Nexthop IP 주소이다.
	Nexthop string `gorm:"not null" json:"nexthop,omitempty"`
}

// ClusterSubnetDHCPPool 는 사용 가능한 DHCP의 범위이다.
type ClusterSubnetDHCPPool struct {
	// ID 는 자동 증가되는 DHCP Pool 의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterSubnetID 는 클러스터 네트워크 서브넷의 아이디이다.
	ClusterSubnetID uint64 `gorm:"not null" json:"cluster_subnet_id,omitempty"`

	// StartIPAddress 는 DHCP Pool 의 시작 주소이다.
	StartIPAddress string `gorm:"not null" json:"start_ip_address,omitempty"`

	// EndIPAddress 는 DHCP Pool 의 끝 주소이다.
	EndIPAddress string `gorm:"not null" json:"end_ip_address,omitempty"`
}

// ClusterSecurityGroup 은 보안 그룹이다.
type ClusterSecurityGroup struct {
	// ID 는 자동 증가되는 보안 그룹의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterTenantID 는 클러스터 테넌트의 ID 이다.
	ClusterTenantID uint64 `gorm:"not null" json:"cluster_tenant_id,omitempty"`

	// UUID 는 보안 그룹의 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// Name 는 보안 그룹의 이름이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Description 는 보안 그룹의 설명이다.
	Description *string `json:"description,omitempty"`

	// Raw 는 보안 그룹의 raw 이다.
	Raw *string `json:"raw,omitempty"`
}

// ClusterSecurityGroupRule 은 보안 그룹 역할이다.
type ClusterSecurityGroupRule struct {
	// ID 는 자동 증가되는 보안 그룹 규칙의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// SecurityGroupID 보안 그룹의 ID 이다.
	SecurityGroupID uint64 `gorm:"not null" json:"security_group_id,omitempty"`

	// RemoteSecurityGroupID 원격 보안 그룹의 ID 이다.
	RemoteSecurityGroupID *uint64 `json:"remote_security_group_id,omitempty"`

	// UUID 보안 그룹 규칙의 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// Description 보안 그룹 규칙의 설명이다.
	Description *string `json:"description,omitempty"`

	// NetworkCIDR 보안 그룹 규칙의 네트워크 CIDR 이다.
	NetworkCidr *string `json:"network_cidr,omitempty"`

	// Direction 보안 그룹 규칙의 다이렉션이다.
	Direction string `gorm:"not null" json:"direction,omitempty"`

	// PortRangeMin 보안 그룹 규칙 포트의 최소값이다.
	PortRangeMin *uint64 `json:"port_range_min,omitempty"`

	// PortRangeMax 보안 그룹 규칙 포트의 최대값이다.
	PortRangeMax *uint64 `json:"port_range_max,omitempty"`

	// Protocol 보안 그룹 규칙의 프로토콜이다.
	Protocol *string `json:"protocol,omitempty"`

	// EtherType 보안 그룹 규칙의 이더넷 타입이다.
	EtherType uint32 `gorm:"not null" json:"ether_type,omitempty"`

	// Raw 는 보안 그룹 규칙의 raw 이다.
	Raw *string `json:"raw,omitempty"`
}

// ClusterStorage 는 클러스터의 storage 볼륨 타입 이다.
type ClusterStorage struct {
	// ID 는 자동 증가 되는 storage 볼륨 타입 의 식별자 이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterID 는 클러스터의 ID 이다.
	ClusterID uint64 `gorm:"not null" json:"cluster_id,omitempty"`

	// UUID 는 storage 볼륨 타입 의 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// Name 는 storage 볼륨 타입 의 이름 이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Description 스토리지의 설명이다.
	Description *string `json:"description,omitempty"`

	// TypeCode 는 storage 볼륨 타입의 종류 이다.
	TypeCode string `gorm:"not null" json:"type_code,omitempty"`

	// CapacityBytes 는 볼륨 타입의 용량 이다.
	CapacityBytes *uint64 `json:"capacity_bytes,omitempty"`

	// UsedBytes 는 볼륨 타입의 사용량 이다.
	UsedBytes *uint64 `json:"used_bytes,omitempty"`

	// Status 는 볼륨 타입 의 status 이다.
	Status string `gorm:"not null" json:"status,omitempty"`

	// Credential 볼륨 타입의 Credential 이다.
	Credential *string `json:"credential,omitempty"`

	// Raw 는 볼륨 타입의 raw 이다.
	Raw *string `json:"raw,omitempty"`
}

// ClusterVolume 클러스터의 volume 이다.
type ClusterVolume struct {
	// ID 는 자동 증가 되는 volume의 식별자 이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterTenantID 는 테넌트의 ID 이다.
	ClusterTenantID uint64 `gorm:"not null" json:"cluster_tenant_id,omitempty"`

	// ClusterStorageID 는 storage 볼륨 타입의 ID 이다.
	ClusterStorageID uint64 `gorm:"not null" json:"cluster_storage_id,omitempty"`

	// UUID 는 volume 의 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// Name 는 volume 의 이름 이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Description 는 volume 의 설명이다.
	Description *string `json:"description,omitempty"`

	// SizeBytes volume 의 용량 이다.
	SizeBytes uint64 `gorm:"not null" json:"size_bytes,omitempty"`

	// Multiattach 는 volume 의 multiattach 여부이다.
	Multiattach bool `gorm:"not null" json:"multiattach,omitempty"`

	// Bootable 는 volume 의 부팅 가능 여부이다.
	Bootable bool `gorm:"not null" json:"bootable,omitempty"`

	// Readonly 는 volume 의 read only 여부이다.
	Readonly bool `gorm:"not null" json:"readonly,omitempty"`

	// Status 는 volume 의 status 이다.
	Status string `gorm:"not null" json:"status,omitempty"`

	// Raw 는 volume의 raw 이다.
	Raw *string `json:"raw,omitempty"`
}

// ClusterVolumeSnapshot 는 볼륨의 스냅샷이다.
type ClusterVolumeSnapshot struct {
	// ID 는 자동 증가되는 ClusterVolumeSnapshot 의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterVolumeID 는 클러스터 볼륨의 아이디이다.
	ClusterVolumeID uint64 `gorm:"not null" json:"cluster_volume_id,omitempty"`

	// UUID 는 클러스터 볼륨 스냅샷의 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// ClusterVolumeGroupSnapshotUUID 는 클러스터 볼륨 그룹의 스냅샷 UUID 이다.
	ClusterVolumeGroupSnapshotUUID *string `json:"cluster_volume_group_snapshot_uuid,omitempty"`

	// Name 은 클러스터 볼륨 스냅샷의 아이디이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Description 은 클러스터 볼륨 스냅샷의 설명이다.
	Description *string `json:"description,omitempty"`

	// SizeBytes 은 클러스터 볼륨 스냅샷의 용량이다.
	SizeBytes uint64 `gorm:"not null" json:"size_bytes,omitempty"`

	// Status 는 클러스터 볼륨 스냅샷의 Status 이다.
	Status string `gorm:"not null" json:"status,omitempty"`

	// CreatedAt 는 클러스터 볼륨 스냅샷의 생성날짜이다.
	CreatedAt int64 `gorm:"not null" json:"created_at,omitempty"`

	// Raw 는 클러스터 볼륨 스냅샷의 Raw 이다.
	Raw *string `json:"raw,omitempty"`
}

// ClusterInstanceSpec 는 인스턴스의 사양이다.
type ClusterInstanceSpec struct {
	// ID 는 자동 증가되는 ClusterInstanceSpec 의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterID 는 클러스터의 ID 이다.
	ClusterID uint64 `gorm:"not null" json:"cluster_id,omitempty"`

	// UUID 는 ClusterInstanceSpec 의 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// Name 은 ClusterInstanceSpec 의 이름이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Description 은 ClusterInstanceSpec 의 설명이다.
	Description *string `json:"description,omitempty"`

	// VcpuTotalCnt 는 인스턴스의 가상 CPU 의 개수이다.
	VcpuTotalCnt uint64 `gorm:"not null" json:"vcpu_total_cnt,omitempty"`

	// MemTotalBytes 는 인스턴스의 메모리의 용량이다.
	MemTotalBytes uint64 `gorm:"not null" json:"mem_total_bytes,omitempty"`

	// DiskTotalBytes 는 인스턴스의 디스크 용량이다.
	DiskTotalBytes uint64 `gorm:"not null" json:"disk_total_bytes,omitempty"`

	// SwapTotalBytes 는 인스턴스의 Swap 영역 용량이다.
	SwapTotalBytes uint64 `gorm:"not null" json:"swap_total_bytes,omitempty"`

	// EphemeralTotalBytes 는 인스턴스의 임시 볼륨 용량이다.
	EphemeralTotalBytes uint64 `gorm:"not null" json:"ephemeral_total_bytes,omitempty"`
}

// ClusterInstanceExtraSpec 는 인스턴스의 추가 사양이다.
type ClusterInstanceExtraSpec struct {
	// ID 는 자동 증가되는 ClusterInstanceExtraSpec 의 식별자이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterInstanceSpecID 는 인스턴스 사양의 ID 이다.
	ClusterInstanceSpecID uint64 `gorm:"not null" json:"cluster_instance_spec_id,omitempty"`

	// Key 는 InstanceExtraSpec 의 키값이다.
	Key string `gorm:"not null" json:"key,omitempty"`

	// Value 는 InstanceExtraSpec 의 값이다.
	Value string `gorm:"not null" json:"value,omitempty"`
}

// ClusterKeypair 는 클러스터 키 페어이다.
type ClusterKeypair struct {
	// ID 는 자동 증가 되는 클러스터 키 페어의 식별자 이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterID 는 클러스터의 ID 이다.
	ClusterID uint64 `gorm:"not null" json:"cluster_id,omitempty"`

	// Name 은 클러스터 키 페어의 이름이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Fingerprint 는 클러스터 키 페어의 FingerPrint 이다.
	Fingerprint string `gorm:"not null" json:"fingerprint,omitempty"`

	// PublicKey 은 클러스터 키 페어의 공개 키이다.
	PublicKey string `gorm:"not null" json:"public_key,omitempty"`

	// TypeCode 은 클러스터 키 페어의 종류이다.
	TypeCode string `gorm:"not null" json:"type_code,omitempty"`
}

// ClusterInstance 는 클러스터 instance 이다.
type ClusterInstance struct {
	// ID 는 자동 증가 되는 인스턴스의 식별자 이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterTenantID 테넌트의 ID 이다.
	ClusterTenantID uint64 `gorm:"not null" json:"cluster_tenant_id,omitempty"`

	// ClusterAvailabilityZoneID 는 가용 구역의 ID 이다.
	ClusterAvailabilityZoneID uint64 `gorm:"not null" json:"cluster_availability_zone_id,omitempty"`

	// ClusterHypervisorID 는 하이퍼바이저의 ID 이다.
	ClusterHypervisorID uint64 `gorm:"not null" json:"cluster_hypervisor_id,omitempty"`

	// ClusterInstanceSpecID 는 인스턴트 스팩의 ID 이다.
	ClusterInstanceSpecID uint64 `gorm:"not null" json:"cluster_instance_spec_id,omitempty"`

	// ClusterKeypairID 는 클러스터 키 페어의 ID 이다.
	ClusterKeypairID *uint64 `json:"cluster_keypair_id,omitempty"`

	// UUID 는 인스턴스의 UUID 이다.
	UUID string `gorm:"not null" json:"uuid,omitempty"`

	// Name 는 인스턴스의 이름 이다.
	Name string `gorm:"not null" json:"name,omitempty"`

	// Description 는 인스턴스의 설명이다.
	Description *string `json:"description,omitempty"`

	// State 는 인스턴스의 state 이다.
	State string `gorm:"not null" json:"state,omitempty"`

	//Status 는 인스턴스의 status 이다.
	Status string `gorm:"not null" json:"status,omitempty"`

	// Raw 는 인스턴스의 Raw 이다.
	Raw *string
}

// ClusterInstanceNetwork 는 클러스터 인스턴스의 네트워크 목록 이다.
type ClusterInstanceNetwork struct {
	// ID 는 자동 증가 되는 인스턴스 네트워크의 식별자 이다.
	ID uint64 `gorm:"primary_key;autoIncrement" json:"id,omitempty"`

	// ClusterInstanceID 는 인스턴스의 ID 이다.
	ClusterInstanceID uint64 `gorm:"not null" json:"cluster_instance_id,omitempty"`

	// ClusterNetworkID 는 테넌트의 ID 이다.
	ClusterNetworkID uint64 `gorm:"not null" json:"cluster_network_id,omitempty"`

	// ClusterSubnetID 는 네트워크 서브넷의 ID 이다.
	ClusterSubnetID uint64 `gorm:"not null" json:"cluster_subnet_id,omitempty"`

	// ClusterFloatingIPID 는 플로팅 아이피의 ID 이다.
	ClusterFloatingIPID *uint64 `json:"cluster_floating_ip_id,omitempty"`

	// DhcpFlag 는 DHCP 사용여부이다.
	DhcpFlag bool `gorm:"not null" json:"dhcp_flag,omitempty"`

	// IPAddress 는 인스턴스 네트워크의 아이피주소이다.
	IPAddress string `gorm:"not null" json:"ip_address,omitempty"`
}

// ClusterInstanceSecurityGroup 는 인스턴스 보안 그룹이다.
type ClusterInstanceSecurityGroup struct {
	// ClusterInstanceID 는 클러스터 인스턴스의 ID 이다.
	ClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"cluster_instance_id,omitempty"`

	// ClusterSecurityGroupID 는 클러스터 보안 그룹의 아이디이다.
	ClusterSecurityGroupID uint64 `gorm:"primary_key;autoIncrement:false" json:"cluster_security_group_id,omitempty"`
}

// ClusterInstanceVolume 는 인스턴스의 볼륨 목록 이다.
type ClusterInstanceVolume struct {
	// ClusterInstanceID 는 인스턴스의 ID 이다.
	ClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"cluster_instance_id,omitempty"`

	// ClusterVolumeID 는 볼륨의 ID 이다.
	ClusterVolumeID uint64 `gorm:"primary_key;autoIncrement:false" json:"cluster_volume_id,omitempty"`

	// DevicePath 는 인스턴스 볼륨의 경로이다.
	DevicePath string `gorm:"not null" json:"device_path,omitempty"`

	// BootIndex 는 인스턴스 볼륨의 부트 인덱스이다.
	BootIndex int64 `gorm:"not null" json:"boot_index,omitempty"`
}

// ClusterInstanceUserScript 는 인스턴스의 user script 데이터이다.
type ClusterInstanceUserScript struct {
	// ClusterInstanceID 는 인스턴스의 ID 이다.
	ClusterInstanceID uint64 `gorm:"primary_key;autoIncrement:false" json:"cluster_instance_id,omitempty"`

	// DevicePath 는 인스턴스 볼륨의 경로이다.
	UserData string `gorm:"not null" json:"user_data,omitempty"`
}

// TableName Cluster 테이블 명 반환 함수
func (Cluster) TableName() string {
	return "cdm_cluster"
}

// BeforeCreate BeforeCreate Record 수정 전 호출 되는 함수
func (c *Cluster) BeforeCreate() error {
	c.CreatedAt = time.Now().Unix()
	c.UpdatedAt = time.Now().Unix()
	return nil
}

// BeforeUpdate BeforeUpdate Record 수정 전 호출 되는 함수
func (c *Cluster) BeforeUpdate() error {
	c.UpdatedAt = time.Now().Unix()
	return nil
}

// TableName ClusterPermission 테이블 명 반환 함수
func (ClusterPermission) TableName() string {
	return "cdm_cluster_permission"
}

// TableName ClusterHypervisor 테이블 명 반환 함수
func (ClusterHypervisor) TableName() string {
	return "cdm_cluster_hypervisor"
}

// TableName ClusterAvailabilityZone 테이블 명 반환 함수
func (ClusterAvailabilityZone) TableName() string {
	return "cdm_cluster_availability_zone"
}

// TableName ClusterTenant 테이블 명 반환 함수
func (ClusterTenant) TableName() string {
	return "cdm_cluster_tenant"
}

// TableName ClusterQuota 테이블 명 반환 함수
func (ClusterQuota) TableName() string {
	return "cdm_cluster_quota"
}

// TableName ClusterNetwork 테이블 명 반환 함수
func (ClusterNetwork) TableName() string {
	return "cdm_cluster_network"
}

// TableName ClusterFloatingIP 테이블 명 반환 함수
func (ClusterFloatingIP) TableName() string {
	return "cdm_cluster_floating_ip"
}

// TableName ClusterSubnet 테이블 명 반환 함수
func (ClusterSubnet) TableName() string {
	return "cdm_cluster_subnet"
}

// TableName ClusterSubnetNameserver 테이블 명 반환 함수
func (ClusterSubnetNameserver) TableName() string {
	return "cdm_cluster_subnet_nameserver"
}

// TableName ClusterRouter 테이블 명 반환 함수
func (ClusterRouter) TableName() string {
	return "cdm_cluster_router"
}

// TableName NetworkRoutingInterface 테이블 명 반환 함수
func (ClusterNetworkRoutingInterface) TableName() string {
	return "cdm_cluster_routing_interface"
}

// TableName ClusterRouterExtraRoute 테이블 명 반환 함수
func (ClusterRouterExtraRoute) TableName() string {
	return "cdm_cluster_router_extra_route"
}

// TableName SubnetDHCPPool 테이블 명 반환 함수
func (ClusterSubnetDHCPPool) TableName() string {
	return "cdm_cluster_subnet_dhcp_pool"
}

// TableName ClusterSecurityGroup 테이블 명 반환 함수
func (ClusterSecurityGroup) TableName() string {
	return "cdm_cluster_security_group"
}

// TableName ClusterSecurityGroup 테이블 명 반환 함수
func (ClusterSecurityGroupRule) TableName() string {
	return "cdm_cluster_security_group_rule"
}

// TableName ClusterStorage 테이블 명 반환 함수
func (ClusterStorage) TableName() string {
	return "cdm_cluster_storage"
}

// TableName ClusterVolumeSnapshot 테이블 명 반환 함수
func (ClusterVolumeSnapshot) TableName() string {
	return "cdm_cluster_volume_snapshot"
}

// TableName ClusterVolume 테이블 명 반환 함수
func (ClusterVolume) TableName() string {
	return "cdm_cluster_volume"
}

// TableName ClusterInstanceSpec 테이블 명 반환 함수
func (ClusterInstanceSpec) TableName() string {
	return "cdm_cluster_instance_spec"
}

// TableName ClusterInstanceExtraSpec 테이블 명 반환 함수
func (ClusterInstanceExtraSpec) TableName() string {
	return "cdm_cluster_instance_extra_spec"
}

// TableName ClusterInstanceExtraSpec 테이블 명 반환 함수
func (ClusterKeypair) TableName() string {
	return "cdm_cluster_keypair"
}

// TableName ClusterInstance 테이블 명 반환 함수
func (ClusterInstance) TableName() string {
	return "cdm_cluster_instance"
}

// TableName ClusterInstanceNetwork 테이블 명 반환 함수
func (ClusterInstanceNetwork) TableName() string {
	return "cdm_cluster_instance_network"
}

// TableName ClusterInstanceVolume 테이블 명 반환 함수
func (ClusterInstanceSecurityGroup) TableName() string {
	return "cdm_cluster_instance_security_group"
}

// TableName ClusterInstanceVolume 테이블 명 반환 함수
func (ClusterInstanceVolume) TableName() string {
	return "cdm_cluster_instance_volume"
}

// TableName ClusterInstanceUserScript 테이블 명 반환 함수
func (ClusterInstanceUserScript) TableName() string {
	return "cdm_cluster_instance_user_script"
}
