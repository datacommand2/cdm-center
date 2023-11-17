package model

import (
	cloudModel "github.com/datacommand2/cdm-cloud/common/database/model"
	"github.com/jinzhu/gorm"
)

// GetGroup 은 클러스터 소유 사용자 그룹을 조회 하는 함수 이다.
func (c *Cluster) GetGroup(db *gorm.DB) (*cloudModel.Group, error) {
	var group cloudModel.Group
	if err := db.Where("id = ?", c.OwnerGroupID).Find(&group).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

// GetPermissions 는 클러스터의 권한 목록을 조회 하는 함수 이다.
func (c *Cluster) GetPermissions(db *gorm.DB) ([]ClusterPermission, error) {
	var clusterPermissions []ClusterPermission
	if err := db.Where("cluster_id = ?", c.ID).Find(&clusterPermissions).Error; err != nil {
		return nil, err
	}

	return clusterPermissions, nil
}

// GetHypervisors 는 클러스터의 하이퍼바이저 목록을 조회 하는 함수 이다.
func (c *Cluster) GetHypervisors(db *gorm.DB) ([]ClusterHypervisor, error) {
	var clusterHypervisors []ClusterHypervisor
	if err := db.Where("cluster_id = ?", c.ID).Find(&clusterHypervisors).Error; err != nil {
		return nil, err
	}

	return clusterHypervisors, nil
}

// GetAvailabilityZones 는 클러스터의 가용 구역 목록을 조회 하는 함수 이다.
func (c *Cluster) GetAvailabilityZones(db *gorm.DB) ([]ClusterAvailabilityZone, error) {
	var clusterAvailabilityZones []ClusterAvailabilityZone
	if err := db.Where("cluster_id = ?", c.ID).Find(&clusterAvailabilityZones).Error; err != nil {
		return nil, err
	}

	return clusterAvailabilityZones, nil
}

// GetTenants 는 클러스터의 테넌트 목록을 조회 하는 함수 이다.
func (c *Cluster) GetTenants(db *gorm.DB) ([]ClusterTenant, error) {
	var clusterTenants []ClusterTenant
	if err := db.Where("cluster_id = ?", c.ID).Find(&clusterTenants).Error; err != nil {
		return nil, err
	}

	return clusterTenants, nil
}

// GetQuotas 는 클러스터 테넌트의 쿼타 목록을 조회하는 함수이다.
func (t *ClusterTenant) GetQuotas(db *gorm.DB) ([]ClusterQuota, error) {
	var clusterQuotas []ClusterQuota
	if err := db.Where("cluster_tenant_id = ?", t.ID).Find(&clusterQuotas).Error; err != nil {
		return nil, err
	}

	return clusterQuotas, nil
}

// GetStorages 는 클러스터의 스토리지 볼륨 타입 목록을 조회 하는 함수 이다.
func (c *Cluster) GetStorages(db *gorm.DB) ([]ClusterStorage, error) {
	var clusterStorages []ClusterStorage
	if err := db.Where("cluster_id = ?", c.ID).Find(&clusterStorages).Error; err != nil {
		return nil, err
	}

	return clusterStorages, nil
}

// GetGroup 은 클러스터에 권한이 있는 사용자 그룹을 조회 하는 함수 이다.
func (c *ClusterPermission) GetGroup(db *gorm.DB) (*cloudModel.Group, error) {
	var group cloudModel.Group
	if err := db.Where("id = ?", c.GroupID).Find(&group).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

// GetInstances 클러스터 하이퍼바이저의 인스턴스 목록을 조회 하는 함수 이다.
func (n *ClusterHypervisor) GetInstances(db *gorm.DB) ([]ClusterInstance, error) {
	var clusterInstances []ClusterInstance
	if err := db.Where("cluster_hypervisor_id = ?", n.ID).Find(&clusterInstances).Error; err != nil {
		return nil, err
	}

	return clusterInstances, nil
}

// GetInstances 는 클러스터 가용 구역의 인스턴스 목록을 조회 하는 함수 이다.
func (z *ClusterAvailabilityZone) GetInstances(db *gorm.DB) ([]ClusterInstance, error) {
	var clusterInstances []ClusterInstance
	if err := db.Where("cluster_availability_zone_id = ?", z.ID).Find(&clusterInstances).Error; err != nil {
		return nil, err
	}

	return clusterInstances, nil
}

// GetNetworks 는 클러스터 테넌트의 네트워크 목록을 조회 하는 함수 이다.
func (t *ClusterTenant) GetNetworks(db *gorm.DB) ([]ClusterNetwork, error) {
	var clusterNetworks []ClusterNetwork
	if err := db.Where("cluster_tenant_id = ?", t.ID).Find(&clusterNetworks).Error; err != nil {
		return nil, err
	}

	return clusterNetworks, nil
}

// GetInstances 는 클러스터 테넌트의 인스턴스 목록을 조회 하는 함수 이다.
func (t *ClusterTenant) GetInstances(db *gorm.DB) ([]ClusterInstance, error) {
	var clusterInstances []ClusterInstance
	if err := db.Where("cluster_tenant_id = ?", t.ID).Find(&clusterInstances).Error; err != nil {
		return nil, err
	}

	return clusterInstances, nil
}

// GetVolumes 는 클러스터 테넌트의 볼륨 목록을 조회 하는 함수 이다.
func (t *ClusterTenant) GetVolumes(db *gorm.DB) ([]ClusterVolume, error) {
	var clusterVolume []ClusterVolume
	if err := db.Where("cluster_tenant_id = ?", t.ID).Find(&clusterVolume).Error; err != nil {
		return nil, err
	}

	return clusterVolume, nil
}

// GetVolumes 는 클러스터 스토리지 타입의 볼륨 목록을 조회 하는 함수 이다.
func (s *ClusterStorage) GetVolumes(db *gorm.DB) ([]ClusterVolume, error) {
	var clusterVolumes []ClusterVolume
	if err := db.Where("cluster_storage_id = ?", s.ID).Find(&clusterVolumes).Error; err != nil {
		return nil, err
	}

	return clusterVolumes, nil
}

// GetInstances 는 클러스터 볼륨의 인스턴스 목록을 조회 하는 함수 이다.
func (v *ClusterVolume) GetInstances(db *gorm.DB) ([]ClusterInstance, error) {
	var clusterInstanceVolumes []ClusterInstanceVolume
	if err := db.Where("cluster_volume_id = ?", v.ID).Find(&clusterInstanceVolumes).Error; err != nil {
		return nil, err
	}

	var id []uint64
	for _, instanceVolume := range clusterInstanceVolumes {
		id = append(id, instanceVolume.ClusterInstanceID)
	}

	var clusterInstances []ClusterInstance
	if err := db.Where("id IN (?)", id).Find(&clusterInstances).Error; err != nil {
		return nil, err
	}

	return clusterInstances, nil
}

// GetSubnets 는 클러스터 네트워크의 서브넷 목록을 조회 하는 함수 이다.
func (n *ClusterNetwork) GetSubnets(db *gorm.DB) ([]ClusterSubnet, error) {
	var clusterSubnets []ClusterSubnet
	if err := db.Where("cluster_network_id = ?", n.ID).Find(&clusterSubnets).Error; err != nil {
		return nil, err
	}

	return clusterSubnets, nil
}

// GetSubnetNameservers 는 클러스터 서브넷의 네임서버를 조회하는 함수이다.
func (s *ClusterSubnet) GetSubnetNameservers(db *gorm.DB) ([]ClusterSubnetNameserver, error) {
	var clusterSubnetNameservers []ClusterSubnetNameserver
	if err := db.Where("cluster_subnet_id = ?", s.ID).Find(&clusterSubnetNameservers).Error; err != nil {
		return nil, err
	}

	return clusterSubnetNameservers, nil
}

// GetDHCPPools 는 클러스터 서브넷의 DHCP Pool 을 목록 조회하는 함수이다.
func (s *ClusterSubnet) GetDHCPPools(db *gorm.DB) ([]ClusterSubnetDHCPPool, error) {
	var clusterSubnetDHCPools []ClusterSubnetDHCPPool
	if err := db.Where("cluster_subnet_id = ?", s.ID).Find(&clusterSubnetDHCPools).Error; err != nil {
		return nil, err
	}

	return clusterSubnetDHCPools, nil
}

// GetFloatingIPs 는 클러스터 네트워크의 FloatingIP 목록을 조회 하는 함수 이다.
func (n *ClusterNetwork) GetFloatingIPs(db *gorm.DB) ([]ClusterFloatingIP, error) {
	var floatingIPs []ClusterFloatingIP
	if err := db.Where("cluster_network_id = ?", n.ID).Find(&floatingIPs).Error; err != nil {
		return nil, err
	}

	return floatingIPs, nil
}

// GetExtraRoutes 는 클러스터 라우터의 추가 라우트 정보 목록을 조회 하는 함수 이다.
func (r *ClusterRouter) GetExtraRoutes(db *gorm.DB) ([]ClusterRouterExtraRoute, error) {
	var routes []ClusterRouterExtraRoute
	if err := db.
		Where(&ClusterRouterExtraRoute{ClusterRouterID: r.ID}).
		Find(&routes).Error; err != nil {
		return nil, err
	}

	return routes, nil
}

// GetInstanceKeypair 인스턴스의 키페어를 조회하는 함수이다.
func (i *ClusterInstance) GetInstanceKeypair(db *gorm.DB) (*ClusterKeypair, error) {
	var clusterKeypair ClusterKeypair
	if err := db.Where("id = ?", i.ClusterKeypairID).Find(&clusterKeypair).Error; err != nil {
		return nil, err
	}

	return &clusterKeypair, nil
}

// GetInstanceSpec 인스턴스의 사양을 조회하는 함수이다.
func (i *ClusterInstance) GetInstanceSpec(db *gorm.DB) (*ClusterInstanceSpec, error) {
	var clusterInstanceSpec ClusterInstanceSpec
	if err := db.Where("id = ?", i.ClusterInstanceSpecID).Find(&clusterInstanceSpec).Error; err != nil {
		return nil, err
	}

	return &clusterInstanceSpec, nil
}

// GetExtraSpecs 는 추가 사양을 조회하는 함수이다.
func (s *ClusterInstanceSpec) GetExtraSpecs(db *gorm.DB) ([]ClusterInstanceExtraSpec, error) {
	var extraSpecs []ClusterInstanceExtraSpec
	if err := db.Where("cluster_instance_spec_id = ?", s.ID).Find(&extraSpecs).Error; err != nil {
		return nil, err
	}

	return extraSpecs, nil
}

// GetInstanceExtraSpecs 는 인스턴스의 추가 사양을 조회하는 함수이다.
func (i *ClusterInstance) GetInstanceExtraSpecs(db *gorm.DB) ([]ClusterInstanceExtraSpec, error) {
	var err error
	var instanceSpec *ClusterInstanceSpec
	if instanceSpec, err = i.GetInstanceSpec(db); err != nil {
		return nil, err
	}

	var extraSpecs []ClusterInstanceExtraSpec
	if extraSpecs, err = instanceSpec.GetExtraSpecs(db); err != nil {
		return nil, err
	}

	return extraSpecs, nil
}

// GetInstanceNetworks 는 클러스터 인스턴스 네트워크 목록을 조회 하는 함수 이다.
func (i *ClusterInstance) GetInstanceNetworks(db *gorm.DB) ([]ClusterInstanceNetwork, error) {
	var instanceNetworks []ClusterInstanceNetwork
	if err := db.Where("cluster_instance_id = ?", i.ID).Find(&instanceNetworks).Error; err != nil {
		return nil, err
	}

	return instanceNetworks, nil
}

// GetSecurityGroups 인스턴스의 보안 그룹을 조회하는 함수이다.
func (i *ClusterInstance) GetSecurityGroups(db *gorm.DB) ([]ClusterSecurityGroup, error) {
	var clusterInstanceSecurityGroups []ClusterInstanceSecurityGroup
	if err := db.Where("cluster_instance_id = ?", i.ID).Find(&clusterInstanceSecurityGroups).Error; err != nil {
		return nil, err
	}

	var id []uint64
	for _, instanceSecurityGroup := range clusterInstanceSecurityGroups {
		id = append(id, instanceSecurityGroup.ClusterSecurityGroupID)
	}

	var clusterSecurityGroup []ClusterSecurityGroup
	if err := db.Where("id IN (?)", id).Find(&clusterSecurityGroup).Error; err != nil {
		return nil, err
	}

	return clusterSecurityGroup, nil
}

// GetSecurityGroupRules 는 보안그룹의 역할을 조회하는 함수이다.
func (g *ClusterSecurityGroup) GetSecurityGroupRules(db *gorm.DB) ([]ClusterSecurityGroupRule, error) {
	var clusterSecurityGroupRules []ClusterSecurityGroupRule
	if err := db.Where("security_group_id = ?", g.ID).Find(&clusterSecurityGroupRules).Error; err != nil {
		return nil, err
	}

	return clusterSecurityGroupRules, nil
}

// GetRemoteSecurityGroup 는 보안그룹의 역할을 조회하는 함수이다.
func (r *ClusterSecurityGroupRule) GetRemoteSecurityGroup(db *gorm.DB) (*ClusterSecurityGroup, error) {
	var remoteSecurityGroup ClusterSecurityGroup
	if err := db.Where("id = ?", r.RemoteSecurityGroupID).Find(&remoteSecurityGroup).Error; err != nil {
		return nil, err
	}

	return &remoteSecurityGroup, nil
}

// GetKeypair 는 클러스터 인스턴스의 키페어를 조회하는 함수이다.
func (i *ClusterInstance) GetKeypair(db *gorm.DB) (*ClusterKeypair, error) {
	var clusterKeypair ClusterKeypair
	if err := db.Where("id = ?", i.ClusterKeypairID).Find(&clusterKeypair).Error; err != nil {
		return nil, err
	}
	return &clusterKeypair, nil
}

// GetVolumes 는 클러스터 인스턴스의 볼륨 목록을 조회 하는 함수 이다.
func (i *ClusterInstance) GetVolumes(db *gorm.DB) ([]ClusterVolume, error) {
	var clusterInstanceVolumes []ClusterInstanceVolume
	if err := db.Where("cluster_instance_id = ?", i.ID).Find(&clusterInstanceVolumes).Error; err != nil {
		return nil, err
	}

	var id []uint64
	for _, instanceVolume := range clusterInstanceVolumes {
		id = append(id, instanceVolume.ClusterVolumeID)
	}

	var clusterVolumes []ClusterVolume
	if err := db.Where("id IN (?)", id).Find(&clusterVolumes).Error; err != nil {
		return nil, err
	}

	return clusterVolumes, nil
}

// GetInstanceVolumes 는 클러스터 인스턴스의 볼륨 목록을 조회 하는 함수이다.
func (i *ClusterInstance) GetInstanceVolumes(db *gorm.DB) ([]ClusterInstanceVolume, error) {
	var clusterInstanceVolumes []ClusterInstanceVolume
	if err := db.Where("cluster_instance_id = ?", i.ID).Find(&clusterInstanceVolumes).Error; err != nil {
		return nil, err
	}

	return clusterInstanceVolumes, nil
}

// GetVolumes 는 볼륨을 조회하는 함수이다.
func (iv *ClusterInstanceVolume) GetVolumes(db *gorm.DB) (*ClusterVolume, error) {
	var clusterVolume ClusterVolume
	if err := db.Where("id = ?", iv.ClusterVolumeID).Find(&clusterVolume).Error; err != nil {
		return nil, err
	}
	return &clusterVolume, nil
}

// GetExternalNetwork 는 Floating IP 의 외부 네트워크를 조회하는 함수이다.
func (f *ClusterFloatingIP) GetExternalNetwork(db *gorm.DB) (*ClusterNetwork, error) {
	var clusterNetwork ClusterNetwork
	if err := db.Where("id = ?", f.ClusterNetworkID).Find(&clusterNetwork).Error; err != nil {
		return nil, err
	}

	return &clusterNetwork, nil
}

// GetFloatingIP 는 Floating IP 를 조회하는 함수이다.
func (t *ClusterInstanceNetwork) GetFloatingIP(db *gorm.DB) (*ClusterFloatingIP, error) {
	var floatingIP ClusterFloatingIP
	if err := db.Where("id = ?", t.ClusterFloatingIPID).Find(&floatingIP).Error; err != nil {
		return nil, err
	}

	return &floatingIP, nil
}

// GetVolumeSnapshots 는 볼륨 스냅샷을 조회하는 함수이다.
func (v *ClusterVolume) GetVolumeSnapshots(db *gorm.DB) ([]ClusterVolumeSnapshot, error) {
	var volumeSnapshot []ClusterVolumeSnapshot
	if err := db.Where("cluster_volume_id = ?", v.ID).Find(&volumeSnapshot).Order("created_at").Error; err != nil {
		return nil, err
	}

	return volumeSnapshot, nil
}
