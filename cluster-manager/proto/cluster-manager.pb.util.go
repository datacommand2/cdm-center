package cms

import (
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	commonModel "github.com/datacommand2/cdm-cloud/common/database/model"
	"github.com/datacommand2/cdm-cloud/common/errors"
)

// Model cms.Group -> model.Group
func (x *Group) Model() (*commonModel.Group, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m commonModel.Group
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	return &m, nil
}

// SetFromModel model.Group -> cms.Group
func (x *Group) SetFromModel(m *commonModel.Group) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterPermission -> model.ClusterPermission
func (x *ClusterPermission) Model() (*model.ClusterPermission, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterPermission
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	// permission group
	g, err := x.Group.Model()
	if err != nil {
		return nil, errors.Unknown(err)
	}

	m.GroupID = g.ID
	return &m, nil
}

// SetFromModel model.ClusterPermission -> cms.ClusterPermission
func (x *ClusterPermission) SetFromModel(m *model.ClusterPermission) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.Cluster 을 model.Cluster 로 변환하여 반환
func (x *Cluster) Model() (*model.Cluster, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.Cluster
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	// owner group
	g, err := x.OwnerGroup.Model()
	if err != nil {
		return nil, errors.Unknown(err)
	}

	m.OwnerGroupID = g.ID
	return &m, nil
}

// Model cms.ClusterQuota -> model.ClusterQuota
func (x *ClusterQuota) Model() (*model.ClusterQuota, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterQuota
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}
	return &m, nil
}

// Model cms.ClusterRouter -> model.ClusterRouter
func (x *ClusterRouter) Model() (*model.ClusterRouter, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterRouter
	_ = json.Unmarshal(b, &m)
	return &m, nil
}

// Model cms.ClusterRouterExtraRoute -> model.ClusterRouterExtraRoute
func (x *ClusterRouterExtraRoute) Model() (*model.ClusterRouterExtraRoute, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterRouterExtraRoute
	_ = json.Unmarshal(b, &m)
	return &m, nil
}

// SetFromModel model.ClusterQuota -> cms.ClusterQuota
func (x *ClusterQuota) SetFromModel(m *model.ClusterQuota) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.Cluster -> cms.Cluster
func (x *Cluster) SetFromModel(m *model.Cluster) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterNode 을 model.ClusterNode 로 변환하여 반환
func (x *ClusterHypervisor) Model() (*model.ClusterHypervisor, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterHypervisor
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	// cluster
	c, err := x.Cluster.Model()
	if err != nil {
		return nil, errors.Unknown(err)
	}

	m.ClusterID = c.ID
	return &m, nil
}

// SetFromModel model.ClusterHypervisor -> cms.ClusterHypervisor
func (x *ClusterHypervisor) SetFromModel(m *model.ClusterHypervisor) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.ClusterInstanceSpec -> cms.ClusterInstanceSpec
func (x *ClusterInstanceSpec) SetFromModel(m *model.ClusterInstanceSpec) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterInstanceSpec -> model.ClusterInstanceSpec
func (x *ClusterInstanceSpec) Model() (*model.ClusterInstanceSpec, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterInstanceSpec
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}
	return &m, nil
}

// SetFromModel model.ClusterInstanceExtraSpec -> cms.ClusterInstanceExtraSpec
func (x *ClusterInstanceExtraSpec) SetFromModel(m *model.ClusterInstanceExtraSpec) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.ClusterKeypair -> cms.ClusterKeypair
func (x *ClusterKeypair) SetFromModel(m *model.ClusterKeypair) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterKeypair -> model.ClusterKeypair
func (x *ClusterKeypair) Model() (*model.ClusterKeypair, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterKeypair
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}
	return &m, nil
}

// SetFromModel model.ClusterNetwork -> cms.ClusterNetwork
func (x *ClusterNetwork) SetFromModel(m *model.ClusterNetwork) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterNetwork -> model.ClusterNetwork
func (x *ClusterNetwork) Model() (*model.ClusterNetwork, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterNetwork
	_ = json.Unmarshal(b, &m)
	return &m, nil
}

// SetFromModel model.ClusterInstanceNetwork -> cms.ClusterInstanceNetwork
func (x *ClusterInstanceNetwork) SetFromModel(m *model.ClusterInstanceNetwork) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterSubnet -> model.ClusterSubnet
func (x *ClusterSubnet) Model() (*model.ClusterSubnet, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterSubnet
	_ = json.Unmarshal(b, &m)
	return &m, nil
}

// Model cms.ClusterSubnetDHCPPool -> model.ClusterSubnetDHCPPool
func (x *ClusterSubnetDHCPPool) Model() (*model.ClusterSubnetDHCPPool, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterSubnetDHCPPool
	_ = json.Unmarshal(b, &m)
	return &m, nil
}

// Model cms.ClusterSubnetNameserver -> model.ClusterSubnetNameserver
func (x *ClusterSubnetNameserver) Model() (*model.ClusterSubnetNameserver, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterSubnetNameserver
	_ = json.Unmarshal(b, &m)
	return &m, nil
}

// SetFromModel model.ClusterSubnet -> cms.ClusterSubnet
func (x *ClusterSubnet) SetFromModel(m *model.ClusterSubnet) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.ClusterSubnetDHCPPool -> cms.ClusterDHCPPool
func (x *ClusterSubnetDHCPPool) SetFromModel(m *model.ClusterSubnetDHCPPool) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.ClusterSubnet -> cms.ClusterSubnet
func (x *ClusterSubnetNameserver) SetFromModel(m *model.ClusterSubnetNameserver) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.ClusterAvailabilityZone -> cms.ClusterAvailabilityZone
func (x *ClusterAvailabilityZone) SetFromModel(m *model.ClusterAvailabilityZone) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterTenant -> model.ClusterTenant
func (x *ClusterTenant) Model() (*model.ClusterTenant, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterTenant
	_ = json.Unmarshal(b, &m)
	return &m, nil
}

// SetFromModel model.ClusterTenant -> cms.ClusterTenant
func (x *ClusterTenant) SetFromModel(m *model.ClusterTenant) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.ClusterInstance -> cms.ClusterInstance
func (x *ClusterInstance) SetFromModel(m *model.ClusterInstance) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterInstance -> model.ClusterInstance
func (x *ClusterInstance) Model() (*model.ClusterInstance, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterInstance
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}
	return &m, nil
}

// SetFromModel model.ClusterInstanceVolume -> cms.ClusterInstanceVolume
func (x *ClusterInstanceVolume) SetFromModel(m *model.ClusterInstanceVolume) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.ClusterVolume -> cms.ClusterVolume
func (x *ClusterVolume) SetFromModel(m *model.ClusterVolume) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterVolume -> model.ClusterVolume
func (x *ClusterVolume) Model() (*model.ClusterVolume, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterVolume
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}
	return &m, nil
}

// SetFromModel model.ClusterSecurityGroup -> cms.ClusterSecurityGroup
func (x *ClusterSecurityGroup) SetFromModel(m *model.ClusterSecurityGroup) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.ClusterSecurityGroupRule -> cms.ClusterSecurityGroupRule
func (x *ClusterSecurityGroupRule) SetFromModel(m *model.ClusterSecurityGroupRule) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterSecurityGroup -> model.ClusterSecurityGroup
func (x *ClusterSecurityGroup) Model() (*model.ClusterSecurityGroup, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterSecurityGroup

	_ = json.Unmarshal(b, &m)
	return &m, nil
}

// Model cms.ClusterSecurityGroupRule -> model.ClusterSecurityGroupRule
func (x *ClusterSecurityGroupRule) Model() (*model.ClusterSecurityGroupRule, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterSecurityGroupRule

	_ = json.Unmarshal(b, &m)
	return &m, nil
}

// SetFromModel model.ClusterStorage -> cms.ClusterStorage
func (x *ClusterStorage) SetFromModel(m *model.ClusterStorage) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterStorage -> model.ClusterStorage
func (x *ClusterStorage) Model() (*model.ClusterStorage, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterStorage
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	m.ClusterID = x.GetCluster().GetId()

	return &m, nil
}

// SetFromModel model.ClusterVolumeSnapshot -> cms.ClusterVolumeSnapshot
func (x *ClusterVolumeSnapshot) SetFromModel(m *model.ClusterVolumeSnapshot) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterVolumeSnapshot -> model.ClusterVolumeSnapshot
func (x *ClusterVolumeSnapshot) Model() (*model.ClusterVolumeSnapshot, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterVolumeSnapshot
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}
	return &m, nil
}

// SetFromModel model.ClusterRouter -> cms.ClusterRouter
func (x *ClusterRouter) SetFromModel(m *model.ClusterRouter) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.ClusterNetworkRoutingInterface -> cms.ClusterNetworkRoutingInterface
func (x *ClusterNetworkRoutingInterface) SetFromModel(m *model.ClusterNetworkRoutingInterface) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.ClusterRouterExtraRoute -> cms.ClusterNetworkRoutingInterface
func (x *ClusterRouterExtraRoute) SetFromModel(m *model.ClusterRouterExtraRoute) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.ClusterFloatingIP -> cms.ClusterFloatingIP
func (x *ClusterFloatingIP) SetFromModel(m *model.ClusterFloatingIP) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model cms.ClusterFloatingIP -> model.ClusterFloatingIP
func (x *ClusterFloatingIP) Model() (*model.ClusterFloatingIP, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ClusterFloatingIP
	_ = json.Unmarshal(b, &m)
	return &m, nil
}
