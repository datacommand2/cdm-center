package migrator

import (
	"fmt"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-disaster-recovery/common/constant"
)

// JSONRef 다른 JSON 스키마 참조를 위한 구조체
type JSONRef struct {
	Ref string `json:"$ref,omitempty"`
}

// NewJSONRef 다른 JSON 스키마 참조를 위한 구조체를 생성한다.
func NewJSONRef(taskID, path string) *JSONRef {
	return &JSONRef{Ref: fmt.Sprintf("%s#%s", taskID, path)}
}

// TenantCreateTaskInputData 테넌트 생성 Task 의 Input Data
type TenantCreateTaskInputData struct {
	Tenant *cms.ClusterTenant `json:"tenant,omitempty"`
}

// TenantCreateTaskInput 테넌트 생성 Task 의 Input
type TenantCreateTaskInput struct {
	Tenant *cms.ClusterTenant `json:"tenant,omitempty"`
}

// TenantCreateTaskOutput 테넌트 생성 Task 의 Output
type TenantCreateTaskOutput struct {
	Tenant  *cms.ClusterTenant `json:"tenant,omitempty"`
	Message *Message           `json:"message"`
}

// SecurityGroupCreateTaskInputData 보안그룹 생성 Task 의 Input Data
type SecurityGroupCreateTaskInputData struct {
	Tenant        *JSONRef                  `json:"tenant,omitempty"`
	SecurityGroup *cms.ClusterSecurityGroup `json:"security_group,omitempty"`
}

// SecurityGroupCreateTaskInput 보안그룹 생성 Task 의 Input
type SecurityGroupCreateTaskInput struct {
	Tenant        *cms.ClusterTenant        `json:"tenant,omitempty"`
	SecurityGroup *cms.ClusterSecurityGroup `json:"security_group,omitempty"`
}

// SecurityGroupCreateTaskOutput 보안그룹 생성 Task 의 Output
type SecurityGroupCreateTaskOutput struct {
	SecurityGroup *cms.ClusterSecurityGroup `json:"security_group,omitempty"`
	Message       *Message                  `json:"message"`
}

// SecurityGroupRuleInputData 보안그룹 규칙 입력을 위한 구조체
type SecurityGroupRuleInputData struct {
	ID                  uint64   `json:"id,omitempty"`
	UUID                string   `json:"uuid,omitempty"`
	Direction           string   `json:"direction,omitempty"`
	Description         string   `json:"description,omitempty"`
	EtherType           uint32   `json:"ether_type,omitempty"`
	NetworkCidr         string   `json:"network_cidr,omitempty"`
	PortRangeMax        uint32   `json:"port_range_max,omitempty"`
	PortRangeMin        uint32   `json:"port_range_min,omitempty"`
	Protocol            string   `json:"protocol,omitempty"`
	RemoteSecurityGroup *JSONRef `json:"remote_security_group,omitempty"`
}

// SecurityGroupRuleCreateTaskInputData 보안그룹 생성 규칙 Task 의 Input Data
type SecurityGroupRuleCreateTaskInputData struct {
	Tenant            *JSONRef                    `json:"tenant,omitempty"`
	SecurityGroup     *JSONRef                    `json:"security_group,omitempty"`
	SecurityGroupRule *SecurityGroupRuleInputData `json:"security_group_rule,omitempty"`
}

// SecurityGroupRuleCreateTaskInput 보안그룹 규칙 생성 Task 의 Input
type SecurityGroupRuleCreateTaskInput struct {
	Tenant            *cms.ClusterTenant            `json:"tenant,omitempty"`
	SecurityGroup     *cms.ClusterSecurityGroup     `json:"security_group,omitempty"`
	SecurityGroupRule *cms.ClusterSecurityGroupRule `json:"security_group_rule,omitempty"`
}

// SecurityGroupRuleCreateTaskOutput 보안그룹 규칙 생성 Task 의 Output
type SecurityGroupRuleCreateTaskOutput struct {
	SecurityGroupRule *cms.ClusterSecurityGroupRule `json:"security_group_rule,omitempty"`
	Message           *Message                      `json:"message"`
}

// NetworkCreateTaskInputData 네트워크 생성 Task 의 Input Data
type NetworkCreateTaskInputData struct {
	Tenant  *JSONRef            `json:"tenant,omitempty"`
	Network *cms.ClusterNetwork `json:"network,omitempty"`
}

// NetworkCreateTaskInput 네트워크 생성 Task 의 Input
type NetworkCreateTaskInput struct {
	Tenant  *cms.ClusterTenant  `json:"tenant,omitempty"`
	Network *cms.ClusterNetwork `json:"network,omitempty"`
}

// NetworkCreateTaskOutput 네트워크 생성 Task 의 Output
type NetworkCreateTaskOutput struct {
	Network *cms.ClusterNetwork `json:"network,omitempty"`
	Message *Message            `json:"message"`
}

// SubnetCreateTaskInputData 서브넷 생성 Task 의 Input Data
type SubnetCreateTaskInputData struct {
	Tenant  *JSONRef           `json:"tenant,omitempty"`
	Network *JSONRef           `json:"network,omitempty"`
	Subnet  *cms.ClusterSubnet `json:"subnet,omitempty"`
}

// SubnetCreateTaskInput 서브넷 생성 Task 의 Input
type SubnetCreateTaskInput struct {
	Tenant  *cms.ClusterTenant  `json:"tenant,omitempty"`
	Network *cms.ClusterNetwork `json:"network,omitempty"`
	Subnet  *cms.ClusterSubnet  `json:"subnet,omitempty"`
}

// SubnetCreateTaskOutput 서브넷 생성 Task 의 Output
type SubnetCreateTaskOutput struct {
	Subnet  *cms.ClusterSubnet `json:"subnet,omitempty"`
	Message *Message           `json:"message"`
}

// InternalRoutingInterfaceInputData 내부 라우팅 인터페이스 입력을 위한 구조체
type InternalRoutingInterfaceInputData struct {
	Network   *JSONRef `json:"network,omitempty"`
	Subnet    *JSONRef `json:"subnet,omitempty"`
	IPAddress string   `json:"ip_address,omitempty"`
}

// RouterInputData 라우터 입력을 위한 구조체
type RouterInputData struct {
	ID                        uint64                                `json:"id,omitempty"`
	Name                      string                                `json:"name,omitempty"`
	Description               string                                `json:"description,omitempty"`
	InternalRoutingInterfaces []*InternalRoutingInterfaceInputData  `json:"internal_routing_interfaces,omitempty"`
	ExternalRoutingInterfaces []*cms.ClusterNetworkRoutingInterface `json:"external_routing_interfaces,omitempty"`
	ExtraRoutes               []*cms.ClusterRouterExtraRoute        `json:"extra_routes,omitempty"`
	State                     string                                `json:"state,omitempty"`
}

// RouterCreateTaskInputData 라우터 생성 Task 의 Input Data
type RouterCreateTaskInputData struct {
	Tenant *JSONRef         `json:"tenant,omitempty"`
	Router *RouterInputData `json:"router,omitempty"`
}

// RouterCreateTaskInput 라우터 생성 Task 의 Input
type RouterCreateTaskInput struct {
	Tenant *cms.ClusterTenant `json:"tenant,omitempty"`
	Router *cms.ClusterRouter `json:"router,omitempty"`
}

// RouterCreateTaskOutput 라우터 생성 Task 의 Output
type RouterCreateTaskOutput struct {
	Router  *cms.ClusterRouter `json:"router,omitempty"`
	Message *Message           `json:"message"`
}

// FloatingIPCreateTaskInputData Floating IP 생성 Task 의 Input Data
type FloatingIPCreateTaskInputData struct {
	Tenant     *JSONRef               `json:"tenant,omitempty"`
	Network    *cms.ClusterNetwork    `json:"network,omitempty"`
	FloatingIP *cms.ClusterFloatingIP `json:"floating_ip,omitempty"`
}

// FloatingIPCreateTaskInput Floating IP 생성 Task 의 Input
type FloatingIPCreateTaskInput struct {
	Tenant     *cms.ClusterTenant     `json:"tenant,omitempty"`
	Network    *cms.ClusterNetwork    `json:"network,omitempty"`
	FloatingIP *cms.ClusterFloatingIP `json:"floating_ip,omitempty"`
}

// FloatingIPCreateTaskOutput Floating IP 생성 Task 의 Output
type FloatingIPCreateTaskOutput struct {
	FloatingIP *cms.ClusterFloatingIP `json:"floating_ip,omitempty"`
	Message    *Message               `json:"message"`
}

// InstanceSpecCreateTaskInputData Instance Spec 생성 Task 의 Input Data
type InstanceSpecCreateTaskInputData struct {
	Spec       *cms.ClusterInstanceSpec        `json:"spec,omitempty"`
	ExtraSpecs []*cms.ClusterInstanceExtraSpec `json:"extra_specs,omitempty"`
}

// InstanceSpecCreateTaskInput Instance Spec 생성 Task 의 Input
type InstanceSpecCreateTaskInput struct {
	Spec       *cms.ClusterInstanceSpec        `json:"spec,omitempty"`
	ExtraSpecs []*cms.ClusterInstanceExtraSpec `json:"extra_specs,omitempty"`
}

// InstanceSpecCreateTaskOutput Instance Spec 생성 Task 의 Output
type InstanceSpecCreateTaskOutput struct {
	Spec       *cms.ClusterInstanceSpec        `json:"spec,omitempty"`
	ExtraSpecs []*cms.ClusterInstanceExtraSpec `json:"extra_specs,omitempty"`
	Message    *Message                        `json:"message"`
}

// KeypairCreateTaskInputData Keypair 생성 Task 의 Input Data
type KeypairCreateTaskInputData struct {
	Keypair *cms.ClusterKeypair `json:"keypair,omitempty"`
}

// KeypairCreateTaskInput Keypair 생성 Task 의 Input
type KeypairCreateTaskInput struct {
	Keypair *cms.ClusterKeypair `json:"keypair,omitempty"`
}

// KeypairCreateTaskOutput Keypair 생성 Task 의 Output
type KeypairCreateTaskOutput struct {
	Keypair *cms.ClusterKeypair `json:"keypair,omitempty"`
	Message *Message            `json:"message"`
}

// VolumeCopyTaskInputData 볼륨 Copy Task 의 Input Data
type VolumeCopyTaskInputData struct {
	Tenant         *JSONRef                     `json:"tenant,omitempty"`
	SourceStorage  *cms.ClusterStorage          `json:"source_storage,omitempty"`
	TargetStorage  *cms.ClusterStorage          `json:"target_storage,omitempty"`
	TargetMetadata map[string]string            `json:"target_metadata,omitempty"`
	Volume         *cms.ClusterVolume           `json:"volume,omitempty"`
	VolumeID       uint64                       `json:"volume_id,omitempty"`
	Snapshots      []*cms.ClusterVolumeSnapshot `json:"snapshots,omitempty"`
}

// VolumeCopyTaskInput 볼륨 copy Task 의 input
type VolumeCopyTaskInput struct {
	Tenant         *cms.ClusterTenant           `json:"tenant,omitempty"`
	SourceStorage  *cms.ClusterStorage          `json:"source_storage,omitempty"`
	TargetStorage  *cms.ClusterStorage          `json:"target_storage,omitempty"`
	TargetMetadata map[string]string            `json:"target_metadata,omitempty"`
	Volume         *cms.ClusterVolume           `json:"volume,omitempty"`
	VolumeID       uint64                       `json:"volume_id,omitempty"`
	Snapshots      []*cms.ClusterVolumeSnapshot `json:"snapshots,omitempty"`
}

// VolumeCopyTaskOutput 볼륨 copy Task 의 Output
type VolumeCopyTaskOutput struct {
	SourceStorage *cms.ClusterStorage          `json:"source_storage,omitempty"`
	TargetStorage *cms.ClusterStorage          `json:"target_storage,omitempty"`
	Volume        *cms.ClusterVolume           `json:"volume,omitempty"`
	Snapshots     []*cms.ClusterVolumeSnapshot `json:"snapshots"`
	Message       *Message                     `json:"message"`
}

// VolumeImportTaskInputData 볼륨 Import Task 의 Input Data
type VolumeImportTaskInputData struct {
	Tenant         *JSONRef          `json:"tenant,omitempty"`
	SourceStorage  *JSONRef          `json:"source_storage,omitempty"`
	TargetStorage  *JSONRef          `json:"target_storage,omitempty"`
	TargetMetadata map[string]string `json:"target_metadata,omitempty"`
	VolumeID       uint64            `json:"volume_id,omitempty"`
	Volume         *JSONRef          `json:"volume,omitempty"`
	Snapshots      *JSONRef          `json:"snapshots,omitempty"`
}

// VolumeImportTaskInput 볼륨 import Task 의 input
type VolumeImportTaskInput struct {
	Tenant         *cms.ClusterTenant           `json:"tenant,omitempty"`
	SourceStorage  *cms.ClusterStorage          `json:"source_storage,omitempty"`
	TargetStorage  *cms.ClusterStorage          `json:"target_storage,omitempty"`
	TargetMetadata map[string]string            `json:"target_metadata,omitempty"`
	Volume         *cms.ClusterVolume           `json:"volume,omitempty"`
	VolumeID       uint64                       `json:"volume_id,omitempty"`
	Snapshots      []*cms.ClusterVolumeSnapshot `json:"snapshots,omitempty"`
}

// VolumeImportTaskOutput 볼륨 import Task 의 Output
type VolumeImportTaskOutput struct {
	SourceStorage *cms.ClusterStorage `json:"source_storage,omitempty"`
	TargetStorage *cms.ClusterStorage `json:"target_storage,omitempty"`
	VolumePair    *cms.VolumePair     `json:"volume_pair,omitempty"`
	SnapshotPairs []*cms.SnapshotPair `json:"snapshot_pairs"`
	Message       *Message            `json:"message"`
}

// InstanceNetworkInputData 인스턴스 네트워크 입력을 위한 구조체
type InstanceNetworkInputData struct {
	Network    *JSONRef `json:"network,omitempty"`
	Subnet     *JSONRef `json:"subnet,omitempty"`
	FloatingIP *JSONRef `json:"floating_ip,omitempty"`
	DhcpFlag   bool     `json:"dhcp_flag,omitempty"`
	IPAddress  string   `json:"ip_address,omitempty"`
}

// InstanceVolumeInputData 인스턴스 볼륨 입력을 위한 구조체
type InstanceVolumeInputData struct {
	Volume     *JSONRef `json:"volume,omitempty"`
	DevicePath string   `json:"device_path,omitempty"`
	BootIndex  int64    `json:"boot_index,omitempty"`
}

// InstanceCreateTaskInputData 인스턴스 생성 Task 의 Input Data
type InstanceCreateTaskInputData struct {
	Tenant           *JSONRef                     `json:"tenant,omitempty"`
	AvailabilityZone *cms.ClusterAvailabilityZone `json:"availability_zone,omitempty"`
	Hypervisor       *cms.ClusterHypervisor       `json:"hypervisor,omitempty"`
	Spec             *JSONRef                     `json:"spec,omitempty"`
	Keypair          *JSONRef                     `json:"keypair,omitempty"`
	Instance         *cms.ClusterInstance         `json:"instance,omitempty"`
	InstanceID       uint64                       `json:"instance_id,omitempty"`
	Networks         []*InstanceNetworkInputData  `json:"networks,omitempty"`
	SecurityGroups   []*JSONRef                   `json:"security_groups,omitempty"`
	Volumes          []*InstanceVolumeInputData   `json:"volumes,omitempty"`
}

// InstanceCreateTaskInput 인스턴스 생성 Task 의 Input
type InstanceCreateTaskInput struct {
	Tenant           *cms.ClusterTenant            `json:"tenant,omitempty"`
	AvailabilityZone *cms.ClusterAvailabilityZone  `json:"availability_zone,omitempty"`
	Hypervisor       *cms.ClusterHypervisor        `json:"hypervisor,omitempty"`
	Spec             *cms.ClusterInstanceSpec      `json:"spec,omitempty"`
	Keypair          *cms.ClusterKeypair           `json:"keypair,omitempty"`
	Instance         *cms.ClusterInstance          `json:"instance,omitempty"`
	InstanceID       uint64                        `json:"instance_id,omitempty"`
	Networks         []*cms.ClusterInstanceNetwork `json:"networks,omitempty"`
	SecurityGroups   []*cms.ClusterSecurityGroup   `json:"security_groups,omitempty"`
	Volumes          []*cms.ClusterInstanceVolume  `json:"volumes,omitempty"`
}

// InstanceCreateOutput 인스턴스 생성 Task 의 Output
type InstanceCreateOutput struct {
	Instance *cms.ClusterInstance `json:"instance,omitempty"`
	Message  *Message             `json:"message"`
}

// TenantDeleteTaskInputData 테넌트 삭제 Task 의 Input Data
type TenantDeleteTaskInputData struct {
	Tenant *JSONRef `json:"tenant,omitempty"`
}

// TenantDeleteTaskInput 테넌트 삭제 Task 의 Input
type TenantDeleteTaskInput struct {
	Tenant *cms.ClusterTenant `json:"tenant,omitempty"`
}

// TenantDeleteTaskOutput 테넌트 삭제 Task 의 Output
type TenantDeleteTaskOutput struct {
	Message *Message `json:"message"`
}

// SecurityGroupDeleteTaskInputData 보안그룹 삭제 Task 의 Input Data
type SecurityGroupDeleteTaskInputData struct {
	Tenant        *JSONRef `json:"tenant,omitempty"`
	SecurityGroup *JSONRef `json:"security_group,omitempty"`
}

// SecurityGroupDeleteTaskInput 보안그룹 삭제 Task 의 Input
type SecurityGroupDeleteTaskInput struct {
	Tenant        *cms.ClusterTenant        `json:"tenant,omitempty"`
	SecurityGroup *cms.ClusterSecurityGroup `json:"security_group,omitempty"`
}

// SecurityGroupDeleteTaskOutput 보안그룹 삭제 Task 의 Output
type SecurityGroupDeleteTaskOutput struct {
	Message *Message `json:"message"`
}

// VolumeCopyDeleteTaskInputData 볼륨 copy 삭제 Task 의 Input Data
type VolumeCopyDeleteTaskInputData struct {
	Tenant         *JSONRef          `json:"tenant,omitempty"`
	SourceStorage  *JSONRef          `json:"source_storage,omitempty"`
	TargetStorage  *JSONRef          `json:"target_storage,omitempty"`
	TargetMetadata map[string]string `json:"target_metadata,omitempty"`
	Volume         *JSONRef          `json:"volume,omitempty"`
	VolumeID       uint64            `json:"volume_id,omitempty"`
	Snapshots      *JSONRef          `json:"snapshots,omitempty"`
}

// VolumeCopyDeleteTaskInput 볼륨 copy 삭제 Task 의 Input
type VolumeCopyDeleteTaskInput struct {
	Tenant         *cms.ClusterTenant           `json:"tenant,omitempty"`
	SourceStorage  *cms.ClusterStorage          `json:"source_storage,omitempty"`
	TargetStorage  *cms.ClusterStorage          `json:"target_storage,omitempty"`
	TargetMetadata map[string]string            `json:"target_metadata,omitempty"`
	Volume         *cms.ClusterVolume           `json:"volume,omitempty"`
	VolumeID       uint64                       `json:"volume_id,omitempty"`
	Snapshots      []*cms.ClusterVolumeSnapshot `json:"snapshots,omitempty"`
}

// VolumeCopyDeleteTaskOutput 볼륨 copy 삭제 Task 의 Output
type VolumeCopyDeleteTaskOutput struct {
	Message *Message `json:"message"`
}

// VolumeUnmanageTaskInputData 볼륨 unmanage Task 의 Input Data
type VolumeUnmanageTaskInputData struct {
	Tenant         *JSONRef          `json:"tenant,omitempty"`
	SourceStorage  *JSONRef          `json:"source_storage,omitempty"`
	TargetStorage  *JSONRef          `json:"target_storage,omitempty"`
	TargetMetadata map[string]string `json:"target_metadata,omitempty"`
	VolumeID       uint64            `json:"volume_id,omitempty"`
	VolumePair     *JSONRef          `json:"volume_pair,omitempty"`
	SnapshotPairs  *JSONRef          `json:"snapshot_pairs,omitempty"`
}

// VolumeUnmanageTaskInput 볼륨 unmanage Task 의 Input
type VolumeUnmanageTaskInput struct {
	Tenant         *cms.ClusterTenant  `json:"tenant,omitempty"`
	SourceStorage  *cms.ClusterStorage `json:"source_storage,omitempty"`
	TargetStorage  *cms.ClusterStorage `json:"target_storage,omitempty"`
	TargetMetadata map[string]string   `json:"target_metadata,omitempty"`
	VolumeID       uint64              `json:"volume_id,omitempty"`
	VolumePair     *cms.VolumePair     `json:"volume_pair,omitempty"`
	SnapshotPairs  []*cms.SnapshotPair `json:"snapshot_pairs,omitempty"`
}

// VolumeUnmanageTaskOutput 볼륨 unmanage Task 의 Output
type VolumeUnmanageTaskOutput struct {
	Message *Message `json:"message"`
}

// InstanceSpecDeleteTaskInputData Instance Spec 삭제 Task 의 Input Data
type InstanceSpecDeleteTaskInputData struct {
	Spec *JSONRef `json:"spec,omitempty"`
}

// InstanceSpecDeleteTaskInput 인스턴스 스펙 삭제 Task 의 Input
type InstanceSpecDeleteTaskInput struct {
	Spec *cms.ClusterInstanceSpec `json:"spec,omitempty"`
}

// InstanceSpecDeleteTaskOutput 인스턴스 스펙 삭제 Task 의 Output
type InstanceSpecDeleteTaskOutput struct {
	Message *Message `json:"message"`
}

// KeypairDeleteTaskInputData Keypair 삭제 Task 의 Input Data
type KeypairDeleteTaskInputData struct {
	Keypair *JSONRef `json:"keypair,omitempty"`
}

// KeypairDeleteTaskInput Keypair 삭제 Task 의 Input
type KeypairDeleteTaskInput struct {
	Keypair *cms.ClusterKeypair `json:"keypair,omitempty"`
}

// KeypairDeleteTaskOutput Keypair 삭제 Task 의 Output
type KeypairDeleteTaskOutput struct {
	Message *Message `json:"message"`
}

// FloatingIPDeleteTaskInputData Floating IP 삭제 Task 의 Input Data
type FloatingIPDeleteTaskInputData struct {
	Tenant     *JSONRef `json:"tenant,omitempty"`
	FloatingIP *JSONRef `json:"floating_ip,omitempty"`
}

// FloatingIPDeleteTaskInput Floating IP 삭제 Task 의 Input
type FloatingIPDeleteTaskInput struct {
	Tenant     *cms.ClusterTenant     `json:"tenant,omitempty"`
	FloatingIP *cms.ClusterFloatingIP `json:"floating_ip,omitempty"`
}

// FloatingIPDeleteTaskOutput Floating IP 삭제 Task 의 Output
type FloatingIPDeleteTaskOutput struct {
	Message *Message `json:"message"`
}

// NetworkDeleteTaskInputData 네트워크 삭제 Task 의 Input Data
type NetworkDeleteTaskInputData struct {
	Tenant  *JSONRef `json:"tenant,omitempty"`
	Network *JSONRef `json:"network,omitempty"`
}

// NetworkDeleteTaskInput 네트워크 삭제 Task 의 Input
type NetworkDeleteTaskInput struct {
	Tenant  *cms.ClusterTenant  `json:"tenant,omitempty"`
	Network *cms.ClusterNetwork `json:"network,omitempty"`
}

// NetworkDeleteTaskOutput 네트워크 삭제 Task 의 Output
type NetworkDeleteTaskOutput struct {
	Message *Message `json:"message"`
}

// RouterDeleteTaskInputData 라우터 삭제 Task 의 Input Data
type RouterDeleteTaskInputData struct {
	Tenant *JSONRef `json:"tenant,omitempty"`
	Router *JSONRef `json:"router,omitempty"`
}

// RouterDeleteTaskInput 라우터 삭제 Task 의 Input
type RouterDeleteTaskInput struct {
	Tenant *cms.ClusterTenant `json:"tenant,omitempty"`
	Router *cms.ClusterRouter `json:"router,omitempty"`
}

// RouterDeleteTaskOutput 라우터 삭제 Task 의 Output
type RouterDeleteTaskOutput struct {
	Message *Message `json:"message"`
}

// InstanceStopTaskInputData 인스턴스 중지 Task 의 Input Data
type InstanceStopTaskInputData struct {
	Tenant   *JSONRef `json:"tenant,omitempty"`
	Instance *JSONRef `json:"instance,omitempty"`
}

// InstanceStopTaskInput 인스턴스 중지 Task 의 Input
type InstanceStopTaskInput struct {
	Tenant   *cms.ClusterTenant   `json:"tenant,omitempty"`
	Instance *cms.ClusterInstance `json:"instance,omitempty"`
}

// InstanceStopTaskOutput 인스턴스 중지 Task 의 Output
type InstanceStopTaskOutput struct {
	Message *Message `json:"message"`
}

// InstanceDeleteTaskInputData 인스턴스 삭제 Task 의 Input Data
type InstanceDeleteTaskInputData struct {
	Tenant   *JSONRef `json:"tenant,omitempty"`
	Instance *JSONRef `json:"instance,omitempty"`
}

// InstanceDeleteTaskInput 인스턴스 삭제 Task 의 Input
type InstanceDeleteTaskInput struct {
	Tenant   *cms.ClusterTenant   `json:"tenant,omitempty"`
	Instance *cms.ClusterInstance `json:"instance,omitempty"`
}

// InstanceDeleteTaskOutput 인스턴스 삭제 Task 의 Output
type InstanceDeleteTaskOutput struct {
	Message *Message `json:"message"`
}

type taskInputData interface {
	IsTaskType(t string) bool
}
type taskInput interface {
	IsTaskType(t string) bool
}
type taskOutput interface {
	IsTaskType(t string) bool
}

// IsTaskType TenantCreateTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *TenantCreateTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateTenant
}

// IsTaskType TenantCreateTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *TenantCreateTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateTenant
}

// IsTaskType TenantCreateTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *TenantCreateTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateTenant
}

// IsTaskType SecurityGroupCreateTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *SecurityGroupCreateTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateSecurityGroup
}

// IsTaskType SecurityGroupCreateTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *SecurityGroupCreateTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateSecurityGroup
}

// IsTaskType SecurityGroupCreateTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *SecurityGroupCreateTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateSecurityGroup
}

// IsTaskType SecurityGroupRuleCreateTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *SecurityGroupRuleCreateTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateSecurityGroupRule
}

// IsTaskType SecurityGroupRuleCreateTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *SecurityGroupRuleCreateTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateSecurityGroupRule
}

// IsTaskType SecurityGroupRuleCreateTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *SecurityGroupRuleCreateTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateSecurityGroupRule
}

// IsTaskType NetworkCreateTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *NetworkCreateTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateNetwork
}

// IsTaskType NetworkCreateTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *NetworkCreateTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateNetwork
}

// IsTaskType NetworkCreateTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *NetworkCreateTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateNetwork
}

// IsTaskType SubnetCreateTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *SubnetCreateTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateSubnet
}

// IsTaskType SubnetCreateTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *SubnetCreateTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateSubnet
}

// IsTaskType SubnetCreateTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *SubnetCreateTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateSubnet
}

// IsTaskType RouterCreateTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *RouterCreateTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateRouter
}

// IsTaskType RouterCreateTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *RouterCreateTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateRouter
}

// IsTaskType RouterCreateTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *RouterCreateTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateRouter
}

// IsTaskType FloatingIPCreateTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *FloatingIPCreateTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateFloatingIP
}

// IsTaskType FloatingIPCreateTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *FloatingIPCreateTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateFloatingIP
}

// IsTaskType FloatingIPCreateTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *FloatingIPCreateTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateFloatingIP
}

// IsTaskType InstanceSpecCreateTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *InstanceSpecCreateTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateSpec
}

// IsTaskType InstanceSpecCreateTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *InstanceSpecCreateTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateSpec
}

// IsTaskType InstanceSpecCreateTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *InstanceSpecCreateTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateSpec
}

// IsTaskType KeypairCreateTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *KeypairCreateTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateKeypair
}

// IsTaskType KeypairCreateTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *KeypairCreateTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateKeypair
}

// IsTaskType KeypairCreateTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *KeypairCreateTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateKeypair
}

// IsTaskType VolumeCopyTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *VolumeCopyTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCopyVolume
}

// IsTaskType VolumeCopyTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *VolumeCopyTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCopyVolume
}

// IsTaskType VolumeCopyTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *VolumeCopyTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCopyVolume
}

// IsTaskType VolumeImportTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *VolumeImportTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeImportVolume
}

// IsTaskType VolumeImportTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *VolumeImportTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeImportVolume
}

// IsTaskType VolumeImportTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *VolumeImportTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeImportVolume
}

// IsTaskType InstanceCreateTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *InstanceCreateTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateAndDiagnosisInstance ||
		t == constant.MigrationTaskTypeCreateAndStopInstance
}

// IsTaskType InstanceCreateTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *InstanceCreateTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateAndDiagnosisInstance ||
		t == constant.MigrationTaskTypeCreateAndStopInstance
}

// IsTaskType InstanceCreateOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *InstanceCreateOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeCreateAndDiagnosisInstance ||
		t == constant.MigrationTaskTypeCreateAndStopInstance
}

// IsTaskType TenantDeleteTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *TenantDeleteTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteTenant
}

// IsTaskType TenantDeleteTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *TenantDeleteTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteTenant
}

// IsTaskType TenantDeleteTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *TenantDeleteTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteTenant
}

// IsTaskType SecurityGroupDeleteTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *SecurityGroupDeleteTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteSecurityGroup
}

// IsTaskType SecurityGroupDeleteTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *SecurityGroupDeleteTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteSecurityGroup
}

// IsTaskType SecurityGroupDeleteTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *SecurityGroupDeleteTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteSecurityGroup
}

// IsTaskType VolumeCopyDeleteTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *VolumeCopyDeleteTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteVolumeCopy
}

// IsTaskType VolumeCopyDeleteTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *VolumeCopyDeleteTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteVolumeCopy
}

// IsTaskType VolumeCopyDeleteTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *VolumeCopyDeleteTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteVolumeCopy
}

// IsTaskType VolumeUnmanageTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *VolumeUnmanageTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeUnmanageVolume
}

// IsTaskType VolumeUnmanageTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *VolumeUnmanageTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeUnmanageVolume
}

// IsTaskType VolumeUnmanageTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *VolumeUnmanageTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeUnmanageVolume
}

// IsTaskType InstanceSpecDeleteTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *InstanceSpecDeleteTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteSpec
}

// IsTaskType InstanceSpecDeleteTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *InstanceSpecDeleteTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteSpec
}

// IsTaskType InstanceSpecDeleteTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *InstanceSpecDeleteTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteSpec
}

// IsTaskType KeypairDeleteTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *KeypairDeleteTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteKeypair
}

// IsTaskType KeypairDeleteTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *KeypairDeleteTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteKeypair
}

// IsTaskType KeypairDeleteTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *KeypairDeleteTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteKeypair
}

// IsTaskType FloatingIPDeleteTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *FloatingIPDeleteTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteFloatingIP
}

// IsTaskType FloatingIPDeleteTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *FloatingIPDeleteTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteFloatingIP
}

// IsTaskType FloatingIPDeleteTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *FloatingIPDeleteTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteFloatingIP
}

// IsTaskType NetworkDeleteTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *NetworkDeleteTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteNetwork
}

// IsTaskType NetworkDeleteTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *NetworkDeleteTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteNetwork
}

// IsTaskType NetworkDeleteTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *NetworkDeleteTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteNetwork
}

// IsTaskType RouterDeleteTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *RouterDeleteTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteRouter
}

// IsTaskType RouterDeleteTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *RouterDeleteTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteRouter
}

// IsTaskType RouterDeleteTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *RouterDeleteTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteRouter
}

// IsTaskType InstanceStopTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *InstanceStopTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeStopInstance
}

// IsTaskType InstanceStopTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *InstanceStopTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeStopInstance
}

// IsTaskType InstanceStopTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *InstanceStopTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeStopInstance
}

// IsTaskType InstanceDeleteTaskInputData 가 해당 타입에 해당하는 Input Data 인지 확인한다.
func (x *InstanceDeleteTaskInputData) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteInstance || t == constant.MigrationTaskTypeDeleteProtectionClusterInstance
}

// IsTaskType InstanceDeleteTaskInput 이 해당 타입에 해당하는 Input 인지 확인한다.
func (x *InstanceDeleteTaskInput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteInstance || t == constant.MigrationTaskTypeDeleteProtectionClusterInstance
}

// IsTaskType InstanceDeleteTaskOutput 이 해당 타입에 해당하는 Output 인지 확인한다.
func (x *InstanceDeleteTaskOutput) IsTaskType(t string) bool {
	return t == constant.MigrationTaskTypeDeleteInstance || t == constant.MigrationTaskTypeDeleteProtectionClusterInstance
}
