package queue

import (
	"github.com/datacommand2/cdm-center/cluster-manager/config"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
)

// CreateClusterTenant 테넌트 생성 메시지 구조체
type CreateClusterTenant struct {
	Cluster *model.Cluster       `json:"cluster"`
	Tenant  *model.ClusterTenant `json:"tenant"`
}

// UpdateClusterTenant 테넌트 수정 메시지 구조체
type UpdateClusterTenant struct {
	Cluster    *model.Cluster       `json:"cluster"`
	OrigTenant *model.ClusterTenant `json:"orig_tenant"`
	Tenant     *model.ClusterTenant `json:"tenant"`
}

// DeleteClusterTenant 테넌트 삭제 메시지 구조체
type DeleteClusterTenant struct {
	Cluster    *model.Cluster       `json:"cluster"`
	OrigTenant *model.ClusterTenant `json:"orig_tenant"`
}

// CreateClusterAvailabilityZone 가용구역 생성 메시지 구조체
type CreateClusterAvailabilityZone struct {
	Cluster          *model.Cluster                 `json:"cluster"`
	AvailabilityZone *model.ClusterAvailabilityZone `json:"availability_zone"`
}

// UpdateClusterAvailabilityZone 가용구역 수정 메시지 구조체
type UpdateClusterAvailabilityZone struct {
	Cluster              *model.Cluster                 `json:"cluster"`
	OrigAvailabilityZone *model.ClusterAvailabilityZone `json:"orig_availability_zone"`
	AvailabilityZone     *model.ClusterAvailabilityZone `json:"availability_zone"`
}

// DeleteClusterAvailabilityZone 가용구역 삭제 메시지 구조체
type DeleteClusterAvailabilityZone struct {
	Cluster              *model.Cluster                 `json:"cluster"`
	OrigAvailabilityZone *model.ClusterAvailabilityZone `json:"orig_availability_zone"`
}

// CreateClusterHypervisor 하이퍼바이저 생성 메시지 구조체
type CreateClusterHypervisor struct {
	Cluster    *model.Cluster           `json:"cluster"`
	Hypervisor *model.ClusterHypervisor `json:"hypervisor"`
}

// UpdateClusterHypervisor 하이퍼바이저 수정 메시지 구조체
type UpdateClusterHypervisor struct {
	Cluster        *model.Cluster           `json:"cluster"`
	OrigHypervisor *model.ClusterHypervisor `json:"orig_hypervisor"`
	Hypervisor     *model.ClusterHypervisor `json:"hypervisor"`
}

// DeleteClusterHypervisor 하이퍼바이저 삭제 메시지 구조체
type DeleteClusterHypervisor struct {
	Cluster        *model.Cluster           `json:"cluster"`
	OrigHypervisor *model.ClusterHypervisor `json:"orig_hypervisor"`
}

// CreateClusterSecurityGroup 보안 그룹 생성 메시지 구조체
type CreateClusterSecurityGroup struct {
	Cluster       *model.Cluster              `json:"cluster"`
	SecurityGroup *model.ClusterSecurityGroup `json:"security_group"`
}

// UpdateClusterSecurityGroup 보안 그룹 수정 메시지 구조체
type UpdateClusterSecurityGroup struct {
	Cluster           *model.Cluster              `json:"cluster"`
	OrigSecurityGroup *model.ClusterSecurityGroup `json:"orig_security_group"`
	SecurityGroup     *model.ClusterSecurityGroup `json:"security_group"`
}

// DeleteClusterSecurityGroup 보안 그룹 삭제 메시지 구조체
type DeleteClusterSecurityGroup struct {
	Cluster           *model.Cluster              `json:"cluster"`
	OrigSecurityGroup *model.ClusterSecurityGroup `json:"orig_security_group"`
}

// CreateClusterSecurityGroupRule 보안 그룹 규칙 생성 메시지 구조체
type CreateClusterSecurityGroupRule struct {
	Cluster           *model.Cluster                  `json:"cluster"`
	SecurityGroupRule *model.ClusterSecurityGroupRule `json:"security_group_rule"`
}

// UpdateClusterSecurityGroupRule 보안 그룹 규칙 수정 메시지 구조체
type UpdateClusterSecurityGroupRule struct {
	Cluster               *model.Cluster                  `json:"cluster"`
	OrigSecurityGroupRule *model.ClusterSecurityGroupRule `json:"orig_security_group_rule"`
	SecurityGroupRule     *model.ClusterSecurityGroupRule `json:"security_group_rule"`
}

// DeleteClusterSecurityGroupRule 보안 그룹 규칙 삭제 메시지 구조체
type DeleteClusterSecurityGroupRule struct {
	Cluster               *model.Cluster                  `json:"cluster"`
	OrigSecurityGroupRule *model.ClusterSecurityGroupRule `json:"orig_security_group_rule"`
}

// CreateClusterNetwork 네트워크 생성 메시지 구조체
type CreateClusterNetwork struct {
	Cluster *model.Cluster        `json:"cluster"`
	Network *model.ClusterNetwork `json:"network"`
}

// UpdateClusterNetwork 네트워크 수정 메시지 구조체
type UpdateClusterNetwork struct {
	Cluster     *model.Cluster        `json:"cluster"`
	OrigNetwork *model.ClusterNetwork `json:"orig_network"`
	Network     *model.ClusterNetwork `json:"network"`
}

// DeleteClusterNetwork 네트워크 삭제 메시지 구조체
type DeleteClusterNetwork struct {
	Cluster     *model.Cluster        `json:"cluster"`
	OrigNetwork *model.ClusterNetwork `json:"orig_network"`
}

// CreateClusterSubnet 서브넷 생성 메시지 구조체
type CreateClusterSubnet struct {
	Cluster *model.Cluster       `json:"cluster"`
	Subnet  *model.ClusterSubnet `json:"subnet"`
}

// UpdateClusterSubnet 서브넷 수정 메시지 구조체
type UpdateClusterSubnet struct {
	Cluster    *model.Cluster       `json:"cluster"`
	OrigSubnet *model.ClusterSubnet `json:"orig_subnet"`
	Subnet     *model.ClusterSubnet `json:"subnet"`
}

// DeleteClusterSubnet 서브넷 삭제 메시지 구조체
type DeleteClusterSubnet struct {
	Cluster    *model.Cluster       `json:"cluster"`
	OrigSubnet *model.ClusterSubnet `json:"orig_subnet"`
}

// CreateClusterFloatingIP floating ip 생성 메시지 구조체
type CreateClusterFloatingIP struct {
	Cluster    *model.Cluster           `json:"cluster"`
	FloatingIP *model.ClusterFloatingIP `json:"floating_ip"`
}

// UpdateClusterFloatingIP floating ip 수정 메시지 구조체
type UpdateClusterFloatingIP struct {
	Cluster        *model.Cluster           `json:"cluster"`
	OrigFloatingIP *model.ClusterFloatingIP `json:"orig_floating_ip"`
	FloatingIP     *model.ClusterFloatingIP `json:"floating_ip"`
}

// DeleteClusterFloatingIP floating ip 삭제 메시지 구조체
type DeleteClusterFloatingIP struct {
	Cluster        *model.Cluster           `json:"cluster"`
	OrigFloatingIP *model.ClusterFloatingIP `json:"orig_floating_ip"`
}

// CreateClusterRouter 라우터 생성 메시지 구조체
type CreateClusterRouter struct {
	Cluster *model.Cluster       `json:"cluster"`
	Router  *model.ClusterRouter `json:"router"`
}

// UpdateClusterRouter 라우터 수정 메시지 구조체
type UpdateClusterRouter struct {
	Cluster    *model.Cluster       `json:"cluster"`
	OrigRouter *model.ClusterRouter `json:"orig_router"`
	Router     *model.ClusterRouter `json:"router"`
}

// DeleteClusterRouter 라우터 삭제 메시지 구조체
type DeleteClusterRouter struct {
	Cluster    *model.Cluster       `json:"cluster"`
	OrigRouter *model.ClusterRouter `json:"orig_router"`
}

// CreateClusterRoutingInterface 라우터 생성 메시지 구조체
type CreateClusterRoutingInterface struct {
	Cluster          *model.Cluster                        `json:"cluster"`
	RoutingInterface *model.ClusterNetworkRoutingInterface `json:"routing_interface"`
}

// UpdateClusterRoutingInterface 라우터 수정 메시지 구조체
type UpdateClusterRoutingInterface struct {
	Cluster              *model.Cluster                        `json:"cluster"`
	OrigRoutingInterface *model.ClusterNetworkRoutingInterface `json:"orig_routing_interface"`
	RoutingInterface     *model.ClusterNetworkRoutingInterface `json:"routing_interface"`
}

// DeleteClusterRoutingInterface 라우터 삭제 메시지 구조체
type DeleteClusterRoutingInterface struct {
	Cluster              *model.Cluster                        `json:"cluster"`
	OrigRoutingInterface *model.ClusterNetworkRoutingInterface `json:"orig_routing_interface"`
}

// CreateClusterStorage 스토리지 생성 메시지 구조체
type CreateClusterStorage struct {
	Cluster *model.Cluster        `json:"cluster"`
	Storage *model.ClusterStorage `json:"storage"`
}

// UpdateClusterStorage 스토리지 수정 메시지 구조체
type UpdateClusterStorage struct {
	Cluster               *model.Cluster        `json:"cluster"`
	OrigStorage           *model.ClusterStorage `json:"orig_storage"`
	Storage               *model.ClusterStorage `json:"storage"`
	OrigVolumeBackendName string                `json:"orig_volume_backend_name"`
	VolumeBackendName     string                `json:"volume_backend_name"`
}

// DeleteClusterStorage 스토리지 삭제 메시지 구조체
type DeleteClusterStorage struct {
	Cluster     *model.Cluster        `json:"cluster"`
	OrigStorage *model.ClusterStorage `json:"orig_storage"`
}

// CreateClusterVolume 볼륨 생성 메시지 구조체
type CreateClusterVolume struct {
	Cluster *model.Cluster       `json:"cluster"`
	Volume  *model.ClusterVolume `json:"volume"`
}

// UpdateClusterVolume 볼륨 수정 메시지 구조체
type UpdateClusterVolume struct {
	Cluster    *model.Cluster       `json:"cluster"`
	OrigVolume *model.ClusterVolume `json:"orig_volume"`
	Volume     *model.ClusterVolume `json:"volume"`
}

// DeleteClusterVolume 볼륨 삭제 메시지 구조체
type DeleteClusterVolume struct {
	Cluster    *model.Cluster       `json:"cluster"`
	OrigVolume *model.ClusterVolume `json:"orig_volume"`
}

// CreateClusterVolumeSnapshot 볼륨 스냅샷 생성 메시지 구조체
type CreateClusterVolumeSnapshot struct {
	Cluster        *model.Cluster               `json:"cluster"`
	Volume         *model.ClusterVolume         `json:"volume"`
	VolumeSnapshot *model.ClusterVolumeSnapshot `json:"volume_snapshot"`
}

// UpdateClusterVolumeSnapshot 볼륨 스냅샷 수정 메시지 구조체
type UpdateClusterVolumeSnapshot struct {
	Cluster            *model.Cluster               `json:"cluster"`
	Volume             *model.ClusterVolume         `json:"volume"`
	OrigVolumeSnapshot *model.ClusterVolumeSnapshot `json:"orig_volume_snapshot"`
	VolumeSnapshot     *model.ClusterVolumeSnapshot `json:"volume_snapshot"`
}

// DeleteClusterVolumeSnapshot 볼륨 스냅샷 삭제 메시지 구조체
type DeleteClusterVolumeSnapshot struct {
	Cluster            *model.Cluster               `json:"cluster"`
	Volume             *model.ClusterVolume         `json:"volume"`
	OrigVolumeSnapshot *model.ClusterVolumeSnapshot `json:"orig_volume_snapshot"`
}

// CreateClusterInstanceExtraSpec 인스턴스 엑스트라 스펙 생성 메시지 구조체
type CreateClusterInstanceExtraSpec struct {
	Cluster   *model.Cluster                  `json:"cluster"`
	ExtraSpec *model.ClusterInstanceExtraSpec `json:"extra_spec"`
}

// UpdateClusterInstanceExtraSpec 인스턴스 엑스트라 스펙 수정 메시지 구조체
type UpdateClusterInstanceExtraSpec struct {
	Cluster       *model.Cluster                  `json:"cluster"`
	OrigExtraSpec *model.ClusterInstanceExtraSpec `json:"orig_extra_spec"`
	ExtraSpec     *model.ClusterInstanceExtraSpec `json:"extra_spec"`
}

// DeleteClusterInstanceExtraSpec 인스턴스 엑스트라 스펙 삭제 메시지 구조체
type DeleteClusterInstanceExtraSpec struct {
	Cluster       *model.Cluster                  `json:"cluster"`
	OrigExtraSpec *model.ClusterInstanceExtraSpec `json:"orig_extra_spec"`
}

// CreateClusterInstanceSpec 인스턴스 스펙 생성 메시지 구조체
type CreateClusterInstanceSpec struct {
	Cluster *model.Cluster             `json:"cluster"`
	Spec    *model.ClusterInstanceSpec `json:"spec"`
}

// UpdateClusterInstanceSpec 인스턴스 스펙 수정 메시지 구조체
type UpdateClusterInstanceSpec struct {
	Cluster  *model.Cluster             `json:"cluster"`
	OrigSpec *model.ClusterInstanceSpec `json:"orig_spec"`
	Spec     *model.ClusterInstanceSpec `json:"spec"`
}

// DeleteClusterInstanceSpec 인스턴스 스펙 삭제 메시지 구조체
type DeleteClusterInstanceSpec struct {
	Cluster  *model.Cluster             `json:"cluster"`
	OrigSpec *model.ClusterInstanceSpec `json:"orig_spec"`
}

// CreateClusterInstance 인스턴스 생성 메시지 구조체
type CreateClusterInstance struct {
	Cluster  *model.Cluster         `json:"cluster"`
	Instance *model.ClusterInstance `json:"instance"`
}

// UpdateClusterInstance 인스턴스 수정 메시지 구조체
type UpdateClusterInstance struct {
	Cluster      *model.Cluster         `json:"cluster"`
	OrigInstance *model.ClusterInstance `json:"orig_instance"`
	Instance     *model.ClusterInstance `json:"instance"`
}

// DeleteClusterInstance 인스턴스 삭제 메시지 구조체
type DeleteClusterInstance struct {
	Cluster      *model.Cluster         `json:"cluster"`
	OrigInstance *model.ClusterInstance `json:"orig_instance"`
}

// AttachClusterInstanceVolume 인스턴스 볼륨 attach 메시지 구조체
type AttachClusterInstanceVolume struct {
	Cluster        *model.Cluster               `json:"cluster"`
	InstanceVolume *model.ClusterInstanceVolume `json:"instance_volume"`
	Volume         *model.ClusterVolume         `json:"volume"`
}

// UpdateClusterInstanceVolume 인스턴스 볼륨 수정 메시지 구조체
type UpdateClusterInstanceVolume struct {
	Cluster        *model.Cluster               `json:"cluster"`
	InstanceVolume *model.ClusterInstanceVolume `json:"instance_volume"`
	Volume         *model.ClusterVolume         `json:"volume"`
}

// DetachClusterInstanceVolume 인스턴스 볼륨 detach 메시지 구조체
type DetachClusterInstanceVolume struct {
	Cluster            *model.Cluster               `json:"cluster"`
	OrigInstanceVolume *model.ClusterInstanceVolume `json:"orig_instance_volume"`
	OrigVolume         *model.ClusterVolume         `json:"orig_volume"`
}

// AttachClusterInstanceNetwork 인스턴스 네트워크 attach 메시지 구조체
type AttachClusterInstanceNetwork struct {
	Cluster         *model.Cluster                `json:"cluster"`
	InstanceNetwork *model.ClusterInstanceNetwork `json:"instance_network"`
}

// UpdateClusterInstanceNetwork 인스턴스 네트워크 수정 메시지 구조체
type UpdateClusterInstanceNetwork struct {
	Cluster             *model.Cluster                `json:"cluster"`
	OrigInstanceNetwork *model.ClusterInstanceNetwork `json:"orig_instance_network"`
	InstanceNetwork     *model.ClusterInstanceNetwork `json:"instance_network"`
}

// DetachClusterInstanceNetwork 인스턴스 네트워크 detach 메시지 구조체
type DetachClusterInstanceNetwork struct {
	Cluster             *model.Cluster                `json:"cluster"`
	OrigInstanceNetwork *model.ClusterInstanceNetwork `json:"orig_instance_network"`
}

// StorageServices Storage 서비스 상태 데이터
type StorageServices struct {
	ID          string `json:"id"`
	Binary      string `json:"binary"`
	BackendName string `json:"backendName"`
	Host        string `json:"host"`
	Zone        string `json:"zone"`
	Status      string `json:"status"`
	UpdatedAt   string `json:"updated_at"`
	Exception   bool   `json:"exception"`
	Deleted     bool   `json:"deleted"`
}

// ComputeServices Compute 서비스 상태 데이터
type ComputeServices struct {
	ID        string `json:"id"`
	Binary    string `json:"binary"`
	Host      string `json:"host"`
	Zone      string `json:"zone"`
	Status    string `json:"status"`
	State     string `json:"state"`
	UpdatedAt string `json:"updated_at"`
	Exception bool   `json:"exception"`
	Deleted   bool   `json:"deleted"`
}

// Agent Network Agent 서비스 상태 데이터
type Agent struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Binary             string `json:"binary"`
	Host               string `json:"host"`
	Status             string `json:"status"`
	HeartbeatTimestamp string `json:"heartbeat_timestamp"`
	Exception          bool   `json:"exception"`
	Deleted            bool   `json:"deleted"`
}

// ClusterServiceStatus 클러스터 서비스 상태 메시지 구조체
type ClusterServiceStatus struct {
	ClusterID    uint64            `json:"cluster_id,omitempty"`
	Storage      []StorageServices `json:"storage,omitempty"`
	Compute      []ComputeServices `json:"compute,omitempty"`
	Network      []Agent           `json:"network,omitempty"`
	Status       string            `json:"status,omitempty"`
	IsNew        bool              `json:"is_new,omitempty"`
	StorageError string            `json:"storage_error,omitempty"`
	ComputeError string            `json:"compute_error,omitempty"`
	NetworkError string            `json:"network_error,omitempty"`
}

// ExcludedCluster 클러스터 상태조회 제외 목록 저장 메시지 구조체
type ExcludedCluster struct {
	ClusterID uint64                        `json:"cluster_id,omitempty"`
	Storage   map[string]*cms.StorageStatus `json:"storage,omitempty"`
	Compute   map[string]*cms.ComputeStatus `json:"compute,omitempty"`
	Network   map[string]*cms.NetworkStatus `json:"network,omitempty"`
}

// SyncClusterStatus 클러스터 동기화 진행상태 메시지 구조체
type SyncClusterStatus struct {
	ClusterID  uint64            `json:"cluster_id,omitempty"`
	Completion map[string]string `json:"completion,omitempty"`
	Progress   int64             `json:"progress,omitempty"`
}

// ClusterConfig 클러스터 설정 주기값 메시지 구조체
type ClusterConfig struct {
	ClusterID uint64         `json:"cluster_id,omitempty"`
	Config    *config.Config `json:"config,omitempty"`
}
