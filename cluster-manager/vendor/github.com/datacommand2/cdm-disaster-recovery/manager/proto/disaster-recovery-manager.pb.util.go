package drms

import (
	"encoding/json"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-disaster-recovery/common/database/model"
)

// Model drms.ProtectionGroup 을 model.ProtectionGroup 로 변환하여 반환
func (x *ProtectionGroup) Model() (*model.ProtectionGroup, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ProtectionGroup
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	return &m, nil
}

// Model drms.RecoveryJob 을 model.Job 로 변환하여 반환
func (x *RecoveryJob) Model() (*model.Job, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.Job
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	m.RecoveryPlanID = x.Plan.Id
	if x.RecoveryPointSnapshot != nil {
		m.RecoveryPointSnapshotID = &x.RecoveryPointSnapshot.Id
	}

	return &m, nil
}

// SetFromModel model.ProtectionGroup -> drms.ProtectionGroup
func (x *ProtectionGroup) SetFromModel(m *model.ProtectionGroup) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model drms.Snapshot 을 model.ProtectionGroupSnapshot 로 변환하여 반환
func (x *ProtectionGroupSnapshot) Model() (*model.ProtectionGroupSnapshot, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.ProtectionGroupSnapshot
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	return &m, nil
}

// Model drms.RecoveryPlan 을 model.Plan 으로 변환하여 반환
func (x *RecoveryPlan) Model() (*model.Plan, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.Plan
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	m.RecoveryClusterID = x.GetRecoveryCluster().GetId()

	return &m, nil
}

// SetFromModel model.Plan -> drms.RecoveryPlan
func (x *RecoveryPlan) SetFromModel(m *model.Plan) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

func makeMessage(code, contents *string) *Message {
	if code == nil {
		return nil
	}

	msg := Message{Code: *code}
	if contents != nil {
		msg.Contents = *contents
	}

	return &msg
}

// Model drms.TenantRecoveryPlan 을 model.PlanTenant 으로 변환하여 반환
func (x *TenantRecoveryPlan) Model() (*model.PlanTenant, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.PlanTenant
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	m.ProtectionClusterTenantID = x.GetProtectionClusterTenant().GetId()

	if x.GetRecoveryClusterTenant().GetId() != 0 {
		m.RecoveryClusterTenantID = &x.RecoveryClusterTenant.Id
	}

	return &m, nil
}

// SetFromModel model.PlanTenant -> drms.TenantRecoveryPlan
func (x *TenantRecoveryPlan) SetFromModel(m *model.PlanTenant) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)

	x.RecoveryClusterTenantMirrorNameUpdateReason = makeMessage(
		m.RecoveryClusterTenantMirrorNameUpdateReasonCode,
		m.RecoveryClusterTenantMirrorNameUpdateReasonContents,
	)

	return nil
}

// Model drms.AvailabilityZoneRecoveryPlan 을 model.PlanAvailabilityZone 으로 변환하여 반환
func (x *AvailabilityZoneRecoveryPlan) Model() (*model.PlanAvailabilityZone, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.PlanAvailabilityZone
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	m.ProtectionClusterAvailabilityZoneID = x.GetProtectionClusterAvailabilityZone().GetId()

	if x.GetRecoveryClusterAvailabilityZone().GetId() != 0 {
		m.RecoveryClusterAvailabilityZoneID = &x.RecoveryClusterAvailabilityZone.Id
	}

	return &m, nil
}

// SetFromModel model.PlanAvailabilityZone -> drms.AvailabilityZoneRecoveryPlan
func (x *AvailabilityZoneRecoveryPlan) SetFromModel(m *model.PlanAvailabilityZone) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)

	x.RecoveryClusterAvailabilityZoneUpdateReason = makeMessage(
		m.RecoveryClusterAvailabilityZoneUpdateReasonCode,
		m.RecoveryClusterAvailabilityZoneUpdateReasonContents,
	)

	return nil
}

// Model drms.ExternalNetworkRecoveryPlan 을 model.PlanExternalNetwork 으로 변환하여 반환
func (x *ExternalNetworkRecoveryPlan) Model() (*model.PlanExternalNetwork, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.PlanExternalNetwork
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	m.ProtectionClusterExternalNetworkID = x.GetProtectionClusterExternalNetwork().GetId()

	if x.GetRecoveryClusterExternalNetwork().GetId() != 0 {
		m.RecoveryClusterExternalNetworkID = &x.RecoveryClusterExternalNetwork.Id
	}

	return &m, nil
}

// SetFromModel model.PlanExternalNetwork -> drms.ExternalNetworkRecoveryPlan
func (x *ExternalNetworkRecoveryPlan) SetFromModel(m *model.PlanExternalNetwork) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)

	x.RecoveryClusterExternalNetworkUpdateReason = makeMessage(
		m.RecoveryClusterExternalNetworkUpdateReasonCode,
		m.RecoveryClusterExternalNetworkUpdateReasonContents,
	)

	return nil
}

// Model drms.RouterRecoveryPlan 을 model.PlanRouter 으로 변환하여 반환
func (x *RouterRecoveryPlan) Model() (*model.PlanRouter, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.PlanRouter
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	m.ProtectionClusterRouterID = x.GetProtectionClusterRouter().GetId()

	if x.GetRecoveryClusterExternalNetwork().GetId() != 0 {
		m.RecoveryClusterExternalNetworkID = &x.RecoveryClusterExternalNetwork.Id
	}

	return &m, nil
}

// SetFromModel model.PlanRouter -> drms.RouterRecoveryPlan
func (x *RouterRecoveryPlan) SetFromModel(m *model.PlanRouter) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)

	x.RecoveryClusterExternalNetworkUpdateReason = makeMessage(
		m.RecoveryClusterExternalNetworkUpdateReasonCode,
		m.RecoveryClusterExternalNetworkUpdateReasonContents,
	)

	return nil
}

// Model drms.StorageRecoveryPlan 을 model.PlanStorage 으로 변환하여 반환
func (x *StorageRecoveryPlan) Model() (*model.PlanStorage, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.PlanStorage
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	m.ProtectionClusterStorageID = x.GetProtectionClusterStorage().GetId()

	if x.GetRecoveryClusterStorage().GetId() != 0 {
		m.RecoveryClusterStorageID = &x.RecoveryClusterStorage.Id
	}

	return &m, nil
}

// SetFromModel model.PlanStorage -> drms.StorageRecoveryPlan
func (x *StorageRecoveryPlan) SetFromModel(m *model.PlanStorage) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)

	x.RecoveryClusterStorageUpdateReason = makeMessage(
		m.RecoveryClusterStorageUpdateReasonCode,
		m.RecoveryClusterStorageUpdateReasonContents,
	)
	x.UnavailableReason = makeMessage(
		m.UnavailableReasonCode,
		m.UnavailableReasonContents,
	)

	return nil
}

// Model drms.InstanceRecoveryPlan 을 model.PlanInstance 으로 변환하여 반환
func (x *InstanceRecoveryPlan) Model() (*model.PlanInstance, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.PlanInstance
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	m.ProtectionClusterInstanceID = x.GetProtectionClusterInstance().GetId()

	if x.GetRecoveryClusterAvailabilityZone().GetId() != 0 {
		m.RecoveryClusterAvailabilityZoneID = &x.RecoveryClusterAvailabilityZone.Id
	}
	if x.GetRecoveryClusterHypervisor().GetId() != 0 {
		m.RecoveryClusterHypervisorID = &x.RecoveryClusterHypervisor.Id
	}

	return &m, nil
}

// SetFromModel model.PlanInstance -> drms.InstanceRecoveryPlan
func (x *InstanceRecoveryPlan) SetFromModel(m *model.PlanInstance) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)

	x.RecoveryClusterAvailabilityZoneUpdateReason = makeMessage(
		m.RecoveryClusterAvailabilityZoneUpdateReasonCode,
		m.RecoveryClusterAvailabilityZoneUpdateReasonContents,
	)

	return nil
}

// Model drms.VolumeRecoveryPlan 을 model.PlanVolume 으로 변환하여 반환
func (x *VolumeRecoveryPlan) Model() (*model.PlanVolume, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.PlanVolume
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	m.ProtectionClusterVolumeID = x.GetProtectionClusterVolume().GetId()

	if x.GetRecoveryClusterStorage().GetId() != 0 {
		m.RecoveryClusterStorageID = &x.RecoveryClusterStorage.Id
	}

	return &m, nil
}

// SetFromModel model.PlanVolume -> drms.VolumeRecoveryPlan
func (x *VolumeRecoveryPlan) SetFromModel(m *model.PlanVolume) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)

	x.RecoveryClusterStorageUpdateReason = makeMessage(
		m.RecoveryClusterStorageUpdateReasonCode,
		m.RecoveryClusterStorageUpdateReasonContents,
	)
	x.UnavailableReason = makeMessage(
		m.UnavailableReasonCode,
		m.UnavailableReasonContents,
	)

	return nil
}

// SetFromModel model.PlanInstanceFloatingIP -> drms.FloatingIPRecoveryPlan
func (x *FloatingIPRecoveryPlan) SetFromModel(m *model.PlanInstanceFloatingIP) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)

	x.UnavailableReason = makeMessage(
		m.UnavailableReasonCode,
		m.UnavailableReasonContents,
	)

	return nil
}

// SetFromModel model.RecoveryJob -> drms.RecoveryJob
func (x *RecoveryJob) SetFromModel(m *model.Job) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.PlanSnapshot -> drms.Snapshot
func (x *ProtectionGroupSnapshot) SetFromModel(m *model.ProtectionGroupSnapshot) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.RecoveryResult -> drms.RecoveryResult
func (x *RecoveryResult) SetFromModel(m *model.RecoveryResult) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// SetFromModel model.InstanceTemplate -> drms.InstanceTemplate
func (x *InstanceTemplate) SetFromModel(m *model.InstanceTemplate) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model drms.InstanceTemplate 을 model.Plan 으로 변환하여 반환
func (x *InstanceTemplate) Model() (*model.InstanceTemplate, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.InstanceTemplate
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	return &m, nil
}

func (x *InstanceTemplateInstance) SetFromModel(m *model.InstanceTemplateInstance) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model drms.InstanceTemplateInstance 을 model.Plan 으로 변환하여 반환
func (x *InstanceTemplateInstance) Model() (*model.InstanceTemplateInstance, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.InstanceTemplateInstance
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	return &m, nil
}

func (x *InstanceTemplateInstanceDependency) SetFromModel(m *model.InstanceTemplateInstanceDependency) error {
	b, err := json.Marshal(m)
	if err != nil {
		return errors.Unknown(err)
	}

	_ = json.Unmarshal(b, x)
	return nil
}

// Model drms.InstanceTemplateInstanceDependency 을 model.Plan 으로 변환하여 반환
func (x *InstanceTemplateInstanceDependency) Model() (*model.InstanceTemplateInstanceDependency, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	var m model.InstanceTemplateInstanceDependency
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errors.Unknown(err)
	}

	return &m, nil
}
