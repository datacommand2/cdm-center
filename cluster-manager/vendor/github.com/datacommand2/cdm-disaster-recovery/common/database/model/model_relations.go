package model

import (
	"github.com/jinzhu/gorm"
)

// GetProtectionInstanceIDs 는 보호 그룹의 '보호 인스턴스 ID' 목록을 조회 하는 함수 이다.
func (g *ProtectionGroup) GetProtectionInstanceIDs(db *gorm.DB) ([]uint64, error) {
	var protectionInstances []ProtectionInstance
	if err := db.Find(&protectionInstances, &ProtectionInstance{ProtectionGroupID: g.ID}).Error; err != nil {
		return nil, err
	}

	var ids []uint64
	for _, protectionInstance := range protectionInstances {
		ids = append(ids, protectionInstance.ProtectionClusterInstanceID)
	}

	return ids, nil
}

// GetPlans 는 보호 그룹의 '재해 복구 계획' 목록을 조회 하는 함수 이다.
func (g *ProtectionGroup) GetPlans(db *gorm.DB) ([]Plan, error) {
	var plans []Plan
	if err := db.Find(&plans, &Plan{ProtectionGroupID: g.ID}).Error; err != nil {
		return nil, err
	}

	return plans, nil
}

// GetPlanTenants 는 재해 복구 계획의 '테넌트 재해 복구 계획' 목록을 조회하는 함수 이다.
func (p *Plan) GetPlanTenants(db *gorm.DB) ([]PlanTenant, error) {
	var planTenants []PlanTenant
	if err := db.Find(&planTenants, &PlanTenant{RecoveryPlanID: p.ID}).Error; err != nil {
		return nil, err
	}

	return planTenants, nil
}

// GetPlanAvailabilityZones 는 재해 복구 계획의 '가용 구역 재해 복구 계획' 목록을 조회하는 함수 이다.
func (p *Plan) GetPlanAvailabilityZones(db *gorm.DB) ([]PlanAvailabilityZone, error) {
	var planAvailabilityZones []PlanAvailabilityZone
	if err := db.Find(&planAvailabilityZones, &PlanAvailabilityZone{RecoveryPlanID: p.ID}).Error; err != nil {
		return nil, err
	}

	return planAvailabilityZones, nil
}

// GetPlanStorages 는 재해 복구 계획의 '스토리지 재해 복구 계획' 목록을 조회하는 함수 이다.
func (p *Plan) GetPlanStorages(db *gorm.DB) ([]PlanStorage, error) {
	var planStorages []PlanStorage
	if err := db.Find(&planStorages, &PlanStorage{RecoveryPlanID: p.ID}).Error; err != nil {
		return nil, err
	}

	return planStorages, nil
}

// GetPlanVolumes 는 재해 복구 계획의 '볼륨 재해 복구 계획' 목록을 조회하는 함수 이다.
func (p *Plan) GetPlanVolumes(db *gorm.DB) ([]PlanVolume, error) {
	var planVolumes []PlanVolume
	if err := db.Find(&planVolumes, &PlanVolume{RecoveryPlanID: p.ID}).Error; err != nil {
		return nil, err
	}

	return planVolumes, nil
}

// GetPlanExternalNetworks 는 재해 복구 계획의 '외부 네트워크 재해 복구 계획' 목록을 조회하는 함수 이다.
func (p *Plan) GetPlanExternalNetworks(db *gorm.DB) ([]PlanExternalNetwork, error) {
	var planNetworks []PlanExternalNetwork
	if err := db.Find(&planNetworks, &PlanExternalNetwork{RecoveryPlanID: p.ID}).Error; err != nil {
		return nil, err
	}

	return planNetworks, nil
}

// GetPlanRouters 는 재해 복구 계획의 '라우터 재해 복구 계획' 목록을 조회하는 함수 이다.
func (p *Plan) GetPlanRouters(db *gorm.DB) ([]PlanRouter, error) {
	var planRouters []PlanRouter
	if err := db.Find(&planRouters, &PlanRouter{RecoveryPlanID: p.ID}).Error; err != nil {
		return nil, err
	}

	return planRouters, nil
}

// GetPlanExternalRoutingInterfaces 는 라우터 재해 복구 계획의 '외부 라우팅 인터페이스' 목록을 조회하는 함수 이다.
func (p *PlanRouter) GetPlanExternalRoutingInterfaces(db *gorm.DB) ([]PlanExternalRoutingInterface, error) {
	var ifaceList []PlanExternalRoutingInterface
	if err := db.Find(&ifaceList, &PlanExternalRoutingInterface{RecoveryPlanID: p.RecoveryPlanID, ProtectionClusterRouterID: p.ProtectionClusterRouterID}).Error; err != nil {
		return nil, err
	}

	return ifaceList, nil
}

// GetPlanInstances 는 재해 복구 계획의 '인스턴스 재해 복구 계획' 목록을 조회하는 함수 이다.
func (p *Plan) GetPlanInstances(db *gorm.DB) ([]PlanInstance, error) {
	var planInstances []PlanInstance
	if err := db.Find(&planInstances, &PlanInstance{RecoveryPlanID: p.ID}).Error; err != nil {
		return nil, err
	}

	return planInstances, nil
}

// GetJobs 는 재해 복구 계획의 '재해 복구 작업 목록' 을 조회 하는 함수 이다.
func (p *Plan) GetJobs(db *gorm.DB) ([]Job, error) {
	var jobs []Job
	if err := db.Find(&jobs, &Job{RecoveryPlanID: p.ID}).Error; err != nil {
		return nil, err
	}

	return jobs, nil
}

// GetPlanInstanceFloatingIPs 는 재해 복구 계획의 '인스턴스 floating IP 재해 복구 계획' 목록을 조회하는 함수 이다.
func (p *Plan) GetPlanInstanceFloatingIPs(db *gorm.DB) ([]PlanInstanceFloatingIP, error) {
	var planInstanceFloatingIPs []PlanInstanceFloatingIP
	if err := db.Find(&planInstanceFloatingIPs, &PlanInstanceFloatingIP{RecoveryPlanID: p.ID}).Error; err != nil {
		return nil, err
	}

	return planInstanceFloatingIPs, nil
}

// GetPlanInstanceFloatingIPs 는 인스턴스 재해 복구 계획의 '인스턴스 floating IP 재해 복구 계획' 목록을 조회하는 함수 이다.
func (i *PlanInstance) GetPlanInstanceFloatingIPs(db *gorm.DB) ([]PlanInstanceFloatingIP, error) {
	var planInstanceFloatingIPs []PlanInstanceFloatingIP
	if err := db.Find(&planInstanceFloatingIPs,
		&PlanInstanceFloatingIP{
			RecoveryPlanID:              i.RecoveryPlanID,
			ProtectionClusterInstanceID: i.ProtectionClusterInstanceID,
		}).Error; err != nil {
		return nil, err
	}

	return planInstanceFloatingIPs, nil
}

// GetDependProtectionInstanceIDs 는 인스턴스 재해 복구 계획의 '의존 인스턴스 ID' 목록을 조회 하는 함수 이다.
func (i *PlanInstance) GetDependProtectionInstanceIDs(db *gorm.DB) ([]uint64, error) {
	var planInstanceDependencies []PlanInstanceDependency
	if err := db.Find(&planInstanceDependencies,
		&PlanInstanceDependency{
			RecoveryPlanID:              i.RecoveryPlanID,
			ProtectionClusterInstanceID: i.ProtectionClusterInstanceID,
		}).Error; err != nil {
		return nil, err
	}

	var ids []uint64
	for _, planInstanceDependency := range planInstanceDependencies {
		ids = append(ids, planInstanceDependency.DependProtectionClusterInstanceID)
	}

	return ids, nil
}
