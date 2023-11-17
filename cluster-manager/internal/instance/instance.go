package instance

import (
	"context"
	"encoding/base64"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/network"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/router"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"
)

// KB16 16kb 크기의 상수
const KB16 = 16 * 1024

type instanceRecord struct {
	Instance         model.ClusterInstance         `gorm:"embedded"`
	Tenant           model.ClusterTenant           `gorm:"embedded"`
	Hypervisor       model.ClusterHypervisor       `gorm:"embedded"`
	AvailabilityZone model.ClusterAvailabilityZone `gorm:"embedded"`
	InstanceSpec     model.ClusterInstanceSpec     `gorm:"embedded"`
	Keypair          model.ClusterKeypair          `gorm:"embedded"`
}

type networkRecord struct {
	InstanceNetwork model.ClusterInstanceNetwork `gorm:"embedded"`
	Network         model.ClusterNetwork         `gorm:"embedded"`
	Subnet          model.ClusterSubnet          `gorm:"embedded"`
	FloatingIP      model.ClusterFloatingIP      `gorm:"embedded"`
}

type instanceVolumeRecord struct {
	InstanceVolume model.ClusterInstanceVolume `gorm:"embedded"`
	Volume         model.ClusterVolume         `gorm:"embedded"`
	Storage        model.ClusterStorage        `gorm:"embedded"`
}

func getCluster(ctx context.Context, db *gorm.DB, id uint64) (*model.Cluster, error) {
	tid, _ := metadata.GetTenantID(ctx)

	var m model.Cluster
	err := db.Model(&m).
		Select("id, name, type_code, remarks").
		Where(&model.Cluster{ID: id, TenantID: tid}).
		First(&m).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundCluster(id)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &m, nil
}

func getNetworkRecords(db *gorm.DB, instanceID uint64) ([]networkRecord, error) {
	var records []networkRecord

	err := db.Table(model.ClusterInstanceNetwork{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_network on cdm_cluster_network.id = cdm_cluster_instance_network.cluster_network_id").
		Joins("join cdm_cluster_subnet on cdm_cluster_subnet.id = cdm_cluster_instance_network.cluster_subnet_id").
		Joins("left join cdm_cluster_floating_ip on cdm_cluster_floating_ip.id = cdm_cluster_instance_network.cluster_floating_ip_id").
		Where(&model.ClusterInstanceNetwork{ClusterInstanceID: instanceID}).
		Find(&records).Error
	if err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return records, nil
}

func getInstanceVolumeRecords(db *gorm.DB, instanceID uint64) ([]instanceVolumeRecord, error) {
	var records []instanceVolumeRecord

	err := db.Table(model.ClusterInstanceVolume{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_volume on cdm_cluster_volume.id = cdm_cluster_instance_volume.cluster_volume_id").
		Joins("join cdm_cluster_storage on cdm_cluster_storage.id = cdm_cluster_volume.cluster_storage_id").
		Where(&model.ClusterInstanceVolume{ClusterInstanceID: instanceID}).
		Find(&records).Error
	if err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	return records, nil
}

func getInstanceRecord(db *gorm.DB, clusterID, instanceID uint64) (*instanceRecord, error) {
	var record instanceRecord
	err := db.Table(model.ClusterInstance{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_instance.cluster_tenant_id").
		Joins("join cdm_cluster_hypervisor on cdm_cluster_hypervisor.id = cdm_cluster_instance.cluster_hypervisor_id").
		Joins("join cdm_cluster_availability_zone on cdm_cluster_availability_zone.id = cdm_cluster_instance.cluster_availability_zone_id").
		Joins("join cdm_cluster_instance_spec on cdm_cluster_instance_spec.id = cdm_cluster_instance.cluster_instance_spec_id").
		Joins("left join cdm_cluster_keypair on cdm_cluster_keypair.id = cdm_cluster_instance.cluster_keypair_id").
		Where(&model.ClusterTenant{ClusterID: clusterID}).
		Where(&model.ClusterHypervisor{ClusterID: clusterID}).
		Where(&model.ClusterAvailabilityZone{ClusterID: clusterID}).
		Where(&model.ClusterInstance{ID: instanceID}).
		First(&record).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterInstance(instanceID)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &record, nil
}

func getInstanceRecordByUUID(db *gorm.DB, clusterID uint64, uuid string) (*instanceRecord, error) {
	var record instanceRecord
	err := db.Table(model.ClusterInstance{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_instance.cluster_tenant_id").
		Joins("join cdm_cluster_hypervisor on cdm_cluster_hypervisor.id = cdm_cluster_instance.cluster_hypervisor_id").
		Joins("join cdm_cluster_availability_zone on cdm_cluster_availability_zone.id = cdm_cluster_instance.cluster_availability_zone_id").
		Joins("join cdm_cluster_instance_spec on cdm_cluster_instance_spec.id = cdm_cluster_instance.cluster_instance_spec_id").
		Joins("left join cdm_cluster_keypair on cdm_cluster_keypair.id = cdm_cluster_instance.cluster_keypair_id").
		Where(&model.ClusterTenant{ClusterID: clusterID}).
		Where(&model.ClusterHypervisor{ClusterID: clusterID}).
		Where(&model.ClusterAvailabilityZone{ClusterID: clusterID}).
		Where(&model.ClusterInstance{UUID: uuid}).
		First(&record).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterInstanceByUUID(uuid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &record, nil
}

// Get 클러스터 인스턴스 조회
func Get(ctx context.Context, req *cms.ClusterInstanceRequest) (*cms.ClusterInstance, error) {
	var err error
	var c *model.Cluster
	var ins *instanceRecord

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateInstanceRequest(ctx, db, req); err != nil {
			logger.Errorf("[Instance-Get] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Instance-Get] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		if ins, err = getInstanceRecord(db, req.ClusterId, req.ClusterInstanceId); err != nil {
			logger.Errorf("[Instance-Get] Could not get the cluster(%d) instance(%d) record. Cause: %+v", req.ClusterId, req.ClusterInstanceId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var (
		keypair        *model.ClusterKeypair
		extraSpecs     []model.ClusterInstanceExtraSpec
		networks       []networkRecord
		volumes        []instanceVolumeRecord
		securityGroups []model.ClusterSecurityGroup
		routerMap      = make(map[uint64]*cms.ClusterRouter)
	)

	if err = database.Execute(func(db *gorm.DB) error {
		if keypair, err = ins.Instance.GetKeypair(db); err != nil && err != gorm.ErrRecordNotFound {
			logger.Errorf("[Instance-Get] Could not get the cluster(%d) keypair. Cause: %+v", req.ClusterId, err)
			return errors.UnusableDatabase(err)
		}

		if extraSpecs, err = ins.InstanceSpec.GetExtraSpecs(db); err != nil {
			logger.Errorf("[Instance-Get] Could not get the cluster(%d) extra specs. Cause: %+v", req.ClusterId, err)
			return errors.UnusableDatabase(err)
		}

		if networks, err = getNetworkRecords(db, ins.Instance.ID); err != nil {
			logger.Errorf("[Instance-Get] Could not get the cluster(%d) network records. Cause: %+v", req.ClusterId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	for _, n := range networks {
		routers, _, err := router.GetList(ctx, &cms.ClusterRouterListRequest{ClusterId: req.ClusterId, ClusterNetworkId: n.Network.ID})
		if err != nil {
			logger.Errorf("[Instance-Get] Could not get the cluster(%d) router list. Cause: %+v", req.ClusterId, err)
			return nil, err
		}
		for _, r := range routers {
			routerMap[r.Id] = r
		}
	}

	if err = database.Execute(func(db *gorm.DB) error {
		if volumes, err = getInstanceVolumeRecords(db, ins.Instance.ID); err != nil {
			logger.Errorf("[Instance-Get] Could not get the cluster(%d) volume records. Cause: %+v", req.ClusterId, err)
			return err
		}

		if securityGroups, err = ins.Instance.GetSecurityGroups(db); err != nil {
			logger.Errorf("[Instance-Get] Could not get the cluster(%d) security groups. Cause: %+v", req.ClusterId, err)
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	instance := cms.ClusterInstance{
		Cluster:          new(cms.Cluster),
		Tenant:           new(cms.ClusterTenant),
		Hypervisor:       new(cms.ClusterHypervisor),
		AvailabilityZone: new(cms.ClusterAvailabilityZone),
		Keypair:          new(cms.ClusterKeypair),
		Spec:             new(cms.ClusterInstanceSpec),
	}

	if err = instance.SetFromModel(&ins.Instance); err != nil {
		return nil, err
	}
	if err = instance.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}
	if err = instance.Tenant.SetFromModel(&ins.Tenant); err != nil {
		return nil, err
	}
	if err = instance.Hypervisor.SetFromModel(&ins.Hypervisor); err != nil {
		return nil, err
	}
	if err = instance.AvailabilityZone.SetFromModel(&ins.AvailabilityZone); err != nil {
		return nil, err
	}
	if err = instance.Keypair.SetFromModel(keypair); err != nil {
		return nil, err
	}
	if err = instance.Spec.SetFromModel(&ins.InstanceSpec); err != nil {
		return nil, err
	}

	for _, e := range extraSpecs {
		var extraSpec cms.ClusterInstanceExtraSpec
		if err = extraSpec.SetFromModel(&e); err != nil {
			return nil, err
		}

		instance.Spec.ExtraSpecs = append(instance.Spec.ExtraSpecs, &extraSpec)
	}

	var networkMap = make(map[uint64]*cms.ClusterNetwork)

	for _, n := range networks {
		if _, ok := networkMap[n.Network.ID]; !ok {
			networkMap[n.Network.ID], err = network.Get(ctx, &cms.ClusterNetworkRequest{
				ClusterId:        req.ClusterId,
				ClusterNetworkId: n.Network.ID,
			})
			if err != nil {
				logger.Errorf("[Instance-Get] Could not get the cluster(%d) network(%d). Cause: %+v", req.ClusterId, n.Network.ID, err)
				return nil, err
			}
		}

		var net = cms.ClusterInstanceNetwork{
			Network: networkMap[n.Network.ID],
			Subnet:  new(cms.ClusterSubnet),
		}

		if err = net.SetFromModel(&n.InstanceNetwork); err != nil {
			return nil, err
		}
		if err = net.Subnet.SetFromModel(&n.Subnet); err != nil {
			return nil, err
		}

		if n.InstanceNetwork.ClusterFloatingIPID != nil {
			net.FloatingIp = new(cms.ClusterFloatingIP)
			if err = net.FloatingIp.SetFromModel(&n.FloatingIP); err != nil {
				return nil, err
			}
		}

		instance.Networks = append(instance.Networks, &net)
	}

	for _, r := range routerMap {
		instance.Routers = append(instance.Routers, r)
	}

	for _, v := range volumes {
		var volume = cms.ClusterInstanceVolume{
			Storage: new(cms.ClusterStorage),
			Volume:  new(cms.ClusterVolume),
		}

		if err = volume.SetFromModel(&v.InstanceVolume); err != nil {
			return nil, err
		}
		if err = volume.Storage.SetFromModel(&v.Storage); err != nil {
			return nil, err
		}
		if err = volume.Volume.SetFromModel(&v.Volume); err != nil {
			return nil, err
		}

		instance.Volumes = append(instance.Volumes, &volume)
	}

	for _, s := range securityGroups {
		var securityGroup cms.ClusterSecurityGroup
		if err = securityGroup.SetFromModel(&s); err != nil {
			return nil, err
		}

		var rules []model.ClusterSecurityGroupRule
		if err = database.Execute(func(db *gorm.DB) error {
			if rules, err = s.GetSecurityGroupRules(db); err != nil {
				logger.Errorf("[Instance-Get] Could not get the cluster(%d) security group rules. Cause: %+v", req.ClusterId, err)
				return errors.UnusableDatabase(err)
			}
			return nil
		}); err != nil {
			return nil, err
		}

		for _, r := range rules {
			var rule cms.ClusterSecurityGroupRule
			if err = rule.SetFromModel(&r); err != nil {
				return nil, err
			}

			if r.RemoteSecurityGroupID != nil {
				if err = database.Execute(func(db *gorm.DB) error {
					rsg, err := r.GetRemoteSecurityGroup(db)
					switch {
					case err == gorm.ErrRecordNotFound:
						return errors.Unknown(err)
					case err != nil:
						return errors.UnusableDatabase(err)
					}

					rule.RemoteSecurityGroup = new(cms.ClusterSecurityGroup)
					if err = rule.RemoteSecurityGroup.SetFromModel(rsg); err != nil {
						return err
					}

					return nil
				}); err != nil {
					return nil, err
				}

			}
			securityGroup.Rules = append(securityGroup.Rules, &rule)
		}

		instance.SecurityGroups = append(instance.SecurityGroups, &securityGroup)
	}

	return &instance, nil
}

// GetByUUID uuid 를 통해 클러스터 인스턴스 조회
func GetByUUID(ctx context.Context, req *cms.ClusterInstanceByUUIDRequest) (*cms.ClusterInstance, error) {
	var err error
	var c *model.Cluster

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateInstanceByUUIDRequest(ctx, db, req); err != nil {
			logger.Errorf("[Instance-GetByUUID] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Instance-GetByUUID] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var (
		ins        *instanceRecord
		keypair    *model.ClusterKeypair
		extraSpecs []model.ClusterInstanceExtraSpec
		networks   []networkRecord
	)

	if err = database.Execute(func(db *gorm.DB) error {
		if ins, err = getInstanceRecordByUUID(db, req.ClusterId, req.Uuid); err != nil {
			return err
		}

		if keypair, err = ins.Instance.GetKeypair(db); err != nil && err != gorm.ErrRecordNotFound {
			return errors.UnusableDatabase(err)
		}

		if extraSpecs, err = ins.InstanceSpec.GetExtraSpecs(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		if networks, err = getNetworkRecords(db, ins.Instance.ID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var routerMap = make(map[uint64]*cms.ClusterRouter)
	for _, n := range networks {
		routers, _, err := router.GetList(ctx, &cms.ClusterRouterListRequest{ClusterId: req.ClusterId, ClusterNetworkId: n.Network.ID})
		if err != nil {
			return nil, err
		}
		for _, r := range routers {
			routerMap[r.Id] = r
		}
	}

	var volumes []instanceVolumeRecord
	var securityGroups []model.ClusterSecurityGroup
	if err = database.Execute(func(db *gorm.DB) error {
		if volumes, err = getInstanceVolumeRecords(db, ins.Instance.ID); err != nil {
			return err
		}

		if securityGroups, err = ins.Instance.GetSecurityGroups(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	instance := cms.ClusterInstance{
		Cluster:          new(cms.Cluster),
		Tenant:           new(cms.ClusterTenant),
		Hypervisor:       new(cms.ClusterHypervisor),
		AvailabilityZone: new(cms.ClusterAvailabilityZone),
		Keypair:          new(cms.ClusterKeypair),
		Spec:             new(cms.ClusterInstanceSpec),
	}

	if err = instance.SetFromModel(&ins.Instance); err != nil {
		return nil, err
	}
	if err = instance.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}
	if err = instance.Tenant.SetFromModel(&ins.Tenant); err != nil {
		return nil, err
	}
	if err = instance.Hypervisor.SetFromModel(&ins.Hypervisor); err != nil {
		return nil, err
	}
	if err = instance.AvailabilityZone.SetFromModel(&ins.AvailabilityZone); err != nil {
		return nil, err
	}
	if err = instance.Keypair.SetFromModel(keypair); err != nil {
		return nil, err
	}
	if err = instance.Spec.SetFromModel(&ins.InstanceSpec); err != nil {
		return nil, err
	}

	for _, e := range extraSpecs {
		var extraSpec cms.ClusterInstanceExtraSpec
		if err = extraSpec.SetFromModel(&e); err != nil {
			return nil, err
		}

		instance.Spec.ExtraSpecs = append(instance.Spec.ExtraSpecs, &extraSpec)
	}

	var networkMap = make(map[uint64]*cms.ClusterNetwork)

	for _, n := range networks {
		if _, ok := networkMap[n.Network.ID]; !ok {
			networkMap[n.Network.ID], err = network.Get(ctx, &cms.ClusterNetworkRequest{
				ClusterId:        req.ClusterId,
				ClusterNetworkId: n.Network.ID,
			})
			if err != nil {
				return nil, err
			}
		}

		var net = cms.ClusterInstanceNetwork{
			Network:    networkMap[n.Network.ID],
			Subnet:     new(cms.ClusterSubnet),
			FloatingIp: new(cms.ClusterFloatingIP),
		}

		if err = net.SetFromModel(&n.InstanceNetwork); err != nil {
			return nil, err
		}
		if err = net.Subnet.SetFromModel(&n.Subnet); err != nil {
			return nil, err
		}
		if err = net.FloatingIp.SetFromModel(&n.FloatingIP); err != nil {
			return nil, err
		}

		instance.Networks = append(instance.Networks, &net)
	}

	for _, r := range routerMap {
		instance.Routers = append(instance.Routers, r)
	}

	for _, v := range volumes {
		var volume = cms.ClusterInstanceVolume{
			Storage: new(cms.ClusterStorage),
			Volume:  new(cms.ClusterVolume),
		}

		if err = volume.SetFromModel(&v.InstanceVolume); err != nil {
			return nil, err
		}
		if err = volume.Storage.SetFromModel(&v.Storage); err != nil {
			return nil, err
		}
		if err = volume.Volume.SetFromModel(&v.Volume); err != nil {
			return nil, err
		}

		instance.Volumes = append(instance.Volumes, &volume)
	}

	for _, s := range securityGroups {
		var securityGroup cms.ClusterSecurityGroup
		if err = securityGroup.SetFromModel(&s); err != nil {
			return nil, err
		}

		var rules []model.ClusterSecurityGroupRule
		if err = database.Execute(func(db *gorm.DB) error {
			if rules, err = s.GetSecurityGroupRules(db); err != nil {
				return errors.UnusableDatabase(err)
			}

			for _, r := range rules {
				var rule cms.ClusterSecurityGroupRule
				if err = rule.SetFromModel(&r); err != nil {
					return err
				}

				if r.RemoteSecurityGroupID != nil {
					rsg, err := r.GetRemoteSecurityGroup(db)
					switch {
					case err == gorm.ErrRecordNotFound:
						return errors.Unknown(err)

					case err != nil:
						return errors.UnusableDatabase(err)
					}

					rule.RemoteSecurityGroup = new(cms.ClusterSecurityGroup)
					if err = rule.RemoteSecurityGroup.SetFromModel(rsg); err != nil {
						return err
					}
				}
				securityGroup.Rules = append(securityGroup.Rules, &rule)
			}

			return nil
		}); err != nil {

		}

		instance.SecurityGroups = append(instance.SecurityGroups, &securityGroup)
	}

	return &instance, nil
}

func getClusterInstanceList(db *gorm.DB, cid uint64, filters ...instanceFilter) ([]instanceRecord, error) {
	var err error

	for _, f := range filters {
		if db, err = f.Apply(db); err != nil {
			return nil, err
		}
	}

	var records []instanceRecord
	err = db.Table(model.ClusterInstance{}.TableName()).
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_instance.cluster_tenant_id").
		Joins("join cdm_cluster_hypervisor on cdm_cluster_hypervisor.id = cdm_cluster_instance.cluster_hypervisor_id").
		Joins("join cdm_cluster_availability_zone on cdm_cluster_availability_zone.id = cdm_cluster_instance.cluster_availability_zone_id").
		Joins("join cdm_cluster_instance_spec on cdm_cluster_instance_spec.id = cdm_cluster_instance.cluster_instance_spec_id").
		Joins("left join cdm_cluster_keypair on cdm_cluster_keypair.id = cdm_cluster_instance.cluster_keypair_id").
		Where(&model.ClusterTenant{ClusterID: cid}).
		Where(&model.ClusterHypervisor{ClusterID: cid}).
		Where(&model.ClusterAvailabilityZone{ClusterID: cid}).
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	return records, nil
}

// GetInstanceNumber 클러스터 인스턴스 수 조회
func GetInstanceNumber(ctx context.Context, db *gorm.DB, req *cms.ClusterInstanceNumberRequest) (*cms.ClusterInstanceNumber, error) {
	var cidList []uint64
	if req.GetClusterId() == 0 {
		cs, _, err := cluster.GetList(ctx, db, &cms.ClusterListRequest{})
		if err != nil {
			logger.Errorf("[GetInstanceNumber] Could not get the cluster list. Cause: %+v", err)
			return nil, err
		}

		for _, c := range cs {
			cidList = append(cidList, c.Id)
		}
	} else {
		c, err := cluster.Get(ctx, db, &cms.ClusterRequest{
			ClusterId: req.GetClusterId(),
		})
		if err != nil {
			logger.Errorf("[GetInstanceNumber] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return nil, err
		}

		cidList = append(cidList, c.Id)
	}

	filters := makeClusterInstanceNumberFilters(cidList)
	total, err := getInstanceCount(db, filters...)
	if err != nil {
		return nil, err
	}

	return &cms.ClusterInstanceNumber{
		Total: total,
	}, nil
}

func getInstanceCount(db *gorm.DB, filters ...instanceFilter) (uint64, error) {
	var err error

	for _, f := range filters {
		if db, err = f.Apply(db); err != nil {
			return 0, err
		}
	}

	var total uint64
	if err := db.Table(model.ClusterInstance{}.TableName()).
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_instance.cluster_tenant_id").
		Count(&total).Error; err != nil {
		return 0, errors.UnusableDatabase(err)
	}

	return total, nil
}

// GetList 클러스터 인스턴스 목록 조회
func GetList(ctx context.Context, req *cms.ClusterInstanceListRequest) ([]*cms.ClusterInstance, *cms.Pagination, error) {
	var err error
	if err = database.Execute(func(db *gorm.DB) error {
		return validateInstanceListRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[Instance-GetList] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	var (
		c       *model.Cluster
		rsp     []*cms.ClusterInstance
		records []instanceRecord
		filters []instanceFilter
	)

	if err = database.Execute(func(db *gorm.DB) error {
		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Instance-GetList] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		filters = makeClusterInstanceListFilters(db, req)

		if records, err = getClusterInstanceList(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[Instance-GetList] Could not get the cluster(%d) instance list. Cause: %+v", req.ClusterId, err)
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	var cmsCluster cms.Cluster
	if err = cmsCluster.SetFromModel(c); err != nil {
		return nil, nil, err
	}

	for _, record := range records {
		var instance cms.ClusterInstance
		if err = instance.SetFromModel(&record.Instance); err != nil {
			return nil, nil, err
		}

		instance.Cluster = &cmsCluster

		instance.Tenant = new(cms.ClusterTenant)
		if err = instance.Tenant.SetFromModel(&record.Tenant); err != nil {
			return nil, nil, err
		}

		instance.AvailabilityZone = new(cms.ClusterAvailabilityZone)
		if err = instance.AvailabilityZone.SetFromModel(&record.AvailabilityZone); err != nil {
			return nil, nil, err
		}

		instance.Hypervisor = new(cms.ClusterHypervisor)
		if err = instance.Hypervisor.SetFromModel(&record.Hypervisor); err != nil {
			return nil, nil, err
		}

		rsp = append(rsp, &instance)
	}

	var pagination *cms.Pagination
	if err = database.Execute(func(db *gorm.DB) error {
		if pagination, err = getInstancePagination(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[Instance-GetList] Could not get the cluster instance list pagination. Cause: %+v", err)
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return rsp, pagination, nil
}

func getInstancePagination(db *gorm.DB, clusterID uint64, filters ...instanceFilter) (*cms.Pagination, error) {
	var err error
	var offset, limit uint64

	var conditions = db

	for _, f := range filters {
		if _, ok := f.(*paginationFilter); ok {
			offset = f.(*paginationFilter).Offset.GetValue()
			limit = f.(*paginationFilter).Limit.GetValue()
			continue
		}

		conditions, err = f.Apply(conditions)
		if err != nil {
			return nil, errors.UnusableDatabase(err)
		}
	}

	var total uint64
	if err = conditions.Table(model.ClusterInstance{}.TableName()).
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_instance.cluster_tenant_id").
		Where(&model.ClusterTenant{ClusterID: clusterID}).
		Count(&total).Error; err != nil {
		return nil, errors.UnusableDatabase(err)
	}

	if limit == 0 {
		return &cms.Pagination{
			Page:       &wrappers.UInt64Value{Value: 1},
			TotalPage:  &wrappers.UInt64Value{Value: 1},
			TotalItems: &wrappers.UInt64Value{Value: total},
		}, nil
	}

	return &cms.Pagination{
		Page:       &wrappers.UInt64Value{Value: offset/limit + 1},
		TotalPage:  &wrappers.UInt64Value{Value: (total + limit - 1) / limit},
		TotalItems: &wrappers.UInt64Value{Value: total},
	}, nil
}

// StartInstance 클러스터 인스턴스 기동
func StartInstance(req *cms.StartClusterInstanceRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[StartInstance] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}
	if err = validateStartClusterInstanceRequest(req); err != nil {
		logger.Errorf("[StartInstance] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[StartInstance] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[StartInstance] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[StartInstance] Could not close client. Cause: %+v", err)
		}
	}()

	err = cli.StartInstance(client.StartInstanceRequest{
		Tenant: model.ClusterTenant{
			UUID: req.GetTenant().GetUuid(),
		},
		Instance: model.ClusterInstance{
			UUID: req.GetInstance().GetUuid(),
		},
	})
	if err != nil {
		logger.Errorf("[StartInstance] Could not start the cluster instance. Cause: %+v", err)
		return err
	}

	return nil
}

// StopInstance 클러스터 인스턴스 중지
func StopInstance(req *cms.StopClusterInstanceRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[StopInstance] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateStopClusterInstanceRequest(req); err != nil {
		logger.Errorf("[StopInstance] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[StopInstance] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[StopInstance] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[StopInstance] Could not close client. Cause: %+v", err)
		}
	}()

	err = cli.StopInstance(client.StopInstanceRequest{
		Tenant: model.ClusterTenant{
			UUID: req.GetTenant().GetUuid(),
		},
		Instance: model.ClusterInstance{
			UUID: req.GetInstance().GetUuid(),
		},
	})
	if err != nil {
		logger.Errorf("[StopInstance] Could not stop the cluster instance. Cause: %+v", err)
		return err
	}

	return nil
}

// CreateInstance 클러스터 인스턴스 생성
func CreateInstance(req *cms.CreateClusterInstanceRequest) (*cms.ClusterInstance, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateInstance] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err := validateCreateClusterInstanceRequest(req); err != nil {
		logger.Errorf("[CreateInstance] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateInstance] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[CreateInstance] Could not connect client. Cause: %+v", err)
		return nil, err
	}

	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateInstance] Could not close client. Cause: %+v", err)
		}
	}()

	var mStorages []*model.ClusterStorage
	for _, v := range req.Volumes {
		s, err := v.GetVolume().GetStorage().Model()
		if err != nil {
			return nil, err
		}

		mStorages = append(mStorages, s)
	}

	mHypervisor, err := req.GetHypervisor().Model()
	if err != nil {
		return nil, err
	}

	if err := cli.PreprocessCreatingInstance(client.PreprocessCreatingInstanceRequest{Storages: mStorages, Hypervisor: mHypervisor}); err != nil {
		logger.Errorf("[CreateInstance] Could not preprocess creating the instance. Cause: %+v", err)
		return nil, err
	}

	mInstance, err := req.GetInstance().Model()
	if err != nil {
		return nil, err
	}

	// 인스턴스에 할당된 네트워크
	var networks []client.AssignedNetwork
	for _, n := range req.GetNetworks() {
		net := client.AssignedNetwork{
			Network: model.ClusterNetwork{
				UUID: n.Network.Uuid,
			},
			Subnet: model.ClusterSubnet{
				UUID: n.Subnet.Uuid,
			},
			DhcpFlag:  n.DhcpFlag,
			IPAddress: n.IpAddress,
		}

		if n.FloatingIp != nil {
			net.FloatingIP = &model.ClusterFloatingIP{
				UUID: n.FloatingIp.Uuid,
			}
		}

		networks = append(networks, net)
	}

	// 인스턴스에 할당된 보안그룹
	var securityGroups []client.AssignedSecurityGroup
	for _, securityGroup := range req.GetSecurityGroups() {
		securityGroups = append(securityGroups, client.AssignedSecurityGroup{
			SecurityGroup: model.ClusterSecurityGroup{
				Name: securityGroup.Name,
				UUID: securityGroup.Uuid,
			},
		})
	}

	// 인스턴스에 attach 된 볼륨
	// TODO 인스턴스 volume boot 순서를 보장 하지 못함
	// openstack API 를 통해 block_device_mapping 정보를 가져올 수가 없음
	// 추후 boot index 지정 하는 방법 연구 및 구현이 필요
	// --user-data option 으로 인스턴스 시작시 실행이 가능하며, 추후 개발
	var bootAbleIndex = int64(0)
	var volumes []client.AttachedVolume
	for _, volume := range req.GetVolumes() {
		var bootIndex = int64(-1)
		if volume.Volume.Bootable {
			bootIndex = bootAbleIndex
			bootAbleIndex++
		}

		volumes = append(volumes, client.AttachedVolume{
			Volume: model.ClusterVolume{
				UUID: volume.Volume.Uuid,
			},
			DevicePath: volume.DevicePath,
			BootIndex:  bootIndex,
		})
	}

	var script model.ClusterInstanceUserScript
	err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterInstanceUserScript{ClusterInstanceID: mInstance.ID}).First(&script).Error
	})
	switch {
	case err == gorm.ErrRecordNotFound:
		// db 에 저장된 값이 없으면 err 를 반환하지 않고 빈 값을 내보낸다.
		script.UserData = ""

	case err != nil:
		logger.Errorf("[CreateInstance] Could not get user script of instance(%d). Cause: %+v", mInstance.ID, err)
		return nil, errors.UnusableDatabase(err)
	}

	res, err := cli.CreateInstance(client.CreateInstanceRequest{
		Tenant: model.ClusterTenant{
			UUID: req.GetTenant().GetUuid(),
		},
		Hypervisor: model.ClusterHypervisor{
			Hostname: req.GetHypervisor().GetHostname(),
		},
		AvailabilityZone: model.ClusterAvailabilityZone{
			Name: req.GetAvailabilityZone().GetName(),
		},
		InstanceSpec: model.ClusterInstanceSpec{
			UUID: req.GetSpec().GetUuid(),
		},
		Keypair: model.ClusterKeypair{
			Name: req.GetKeypair().GetName(),
		},
		InstanceUserScript: model.ClusterInstanceUserScript{
			UserData: script.UserData,
		},
		AssignedNetworks:       networks,
		AssignedSecurityGroups: securityGroups,
		AttachedVolumes:        volumes,
		Instance:               *mInstance,
	})
	if err != nil {
		logger.Errorf("[CreateInstance] Could not create the cluster instance. Cause: %+v", err)
		return nil, err
	}

	var instance cms.ClusterInstance
	if err = instance.SetFromModel(&res.Instance); err != nil {
		return nil, err
	}

	return &instance, nil
}

// DeleteInstance 인스턴스 삭제
func DeleteInstance(req *cms.DeleteClusterInstanceRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteInstance] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteClusterInstanceRequest(req); err != nil {
		logger.Errorf("[DeleteInstance] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteInstance] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[DeleteInstance] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteInstance] Could not close client. Cause: %+v", err)
		}
	}()

	if err = cli.DeleteInstance(client.DeleteInstanceRequest{
		Tenant: model.ClusterTenant{
			UUID: req.GetTenant().GetUuid(),
		},
		Instance: model.ClusterInstance{
			UUID: req.GetInstance().GetUuid(),
		},
	}); err != nil {
		logger.Errorf("[DeleteInstance] Could not delete the cluster instance. Cause: %+v", err)
		return err
	}

	return nil
}

func validateCheckIsExistInstance(req *cms.CheckIsExistClusterInstanceRequest) error {
	if req.GetClusterId() == 0 {
		return errors.RequiredParameter("cluster_id")
	}

	if req.GetUuid() == "" {
		return errors.RequiredParameter("cluster_instance_uuid")
	}

	return nil
}

// CheckIsExistInstance uuid 를 통한 클러스터 instance 존재유무 확인
func CheckIsExistInstance(req *cms.CheckIsExistClusterInstanceRequest) (bool, error) {
	var (
		err  error
		conn *cms.ClusterConnectionInfo
	)

	if err = validateCheckIsExistInstance(req); err != nil {
		logger.Errorf("[CheckIsExistInstance] Errors occurred during validating the request. Cause: %+v", err)
		return false, err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		if conn, err = cluster.GetConnectionInfo(db, req.ClusterId); err != nil {
			logger.Errorf("[CheckIsExistInstance] Error occurred during checking the cluster connection info. Cause: %+v", err)
			return err
		}
		return nil
	}); err != nil {
		return false, err
	}

	conn.Credential, err = client.DecryptCredentialPassword(conn.GetCredential())
	if err != nil {
		logger.Errorf("[CheckIsExistInstance] Error occurred during checking the cluster connection info. Cause: %v", err)
		return false, err
	}

	cli, err := client.New(conn.GetTypeCode(), conn.GetApiServerUrl(), conn.GetCredential(), conn.GetTenantId())
	if err != nil {
		logger.Errorf("[CheckIsExistInstance] Could not create client. Cause: %v", err)
		return false, err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[CheckIsExistInstance] Could not connect client. Cause: %v", err)
		return false, err
	}
	defer func() {
		if err = cli.Close(); err != nil {
			logger.Warnf("[CheckIsExistInstance] Could not close client. Cause: %v", err)
		}
	}()

	_, err = cli.GetInstance(client.GetInstanceRequest{Instance: model.ClusterInstance{UUID: req.Uuid}})
	if errors.Equal(err, client.ErrNotFound) {
		logger.Warnf("[CheckIsExistInstance] Could not found cluster(%d) instance(%s) in openstack.", req.ClusterId, req.Uuid)
		return false, nil
	} else if err != nil {
		logger.Errorf("[CheckIsExistInstance] Could not get cluster(%d) instance(%s) from openstack. Cause: %+v", req.ClusterId, req.Uuid, err)
		return false, err
	}

	return true, nil
}

// GetInstanceUserScript instance 의 user custom script 조회
func GetInstanceUserScript(req *cms.ClusterInstanceUserScriptRequest) (string, error) {
	if req.GetClusterInstanceId() == 0 {
		return "", errors.RequiredParameter("cluster_instance_id")
	}

	var script model.ClusterInstanceUserScript
	err := database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterInstanceUserScript{ClusterInstanceID: req.ClusterInstanceId}).First(&script).Error
	})
	switch {
	case err == gorm.ErrRecordNotFound:
		// db 에 저장된 값이 없어도 err 를 반환하지 않고 빈 값을 내보낸다.
		return "", nil
	case err != nil:
		logger.Errorf("[GetInstanceUserScript] Could not get user script of instance(%d). Cause: %+v", req.ClusterInstanceId, err)
		return "", errors.UnusableDatabase(err)
	}

	return script.UserData, nil
}

// UpdateInstanceUserScript instance 의 user custom script 수정
func UpdateInstanceUserScript(req *cms.UpdateClusterInstanceUserScriptRequest) error {
	if req.GetClusterInstanceId() == 0 {
		return errors.RequiredParameter("cluster_instance_id")
	}

	decoded, err := base64.StdEncoding.DecodeString(req.UserData)
	if err != nil {
		logger.Errorf("[UpdateInstanceUserScript] Could not decoded. Cause: %+v", err)
		return err
	}

	if len(decoded) > KB16 {
		return errors.OutOfRangeParameterValue("cluster.instance.user_data", req.UserData, 0, KB16)
	}

	if err := database.Transaction(func(db *gorm.DB) error {
		return db.Save(&model.ClusterInstanceUserScript{ClusterInstanceID: req.ClusterInstanceId, UserData: req.UserData}).Error
	}); err != nil {
		logger.Errorf("[UpdateInstanceUserScript] Could not update user script of instance(%d). Cause: %+v", req.ClusterInstanceId, err)
		return errors.UnusableDatabase(err)
	}

	return nil
}
