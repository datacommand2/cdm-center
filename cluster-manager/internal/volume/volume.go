package volume

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-center/cluster-manager/sync"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/jinzhu/gorm"

	// for openstack client
	_ "github.com/datacommand2/cdm-center/cluster-manager/client/openstack"
)

type volumeRecord struct {
	Volume  model.ClusterVolume  `gorm:"embedded"`
	Tenant  model.ClusterTenant  `gorm:"embedded"`
	Storage model.ClusterStorage `gorm:"embedded"`
}

func getVolumeRecord(db *gorm.DB, clusterID, volumeID uint64) (*volumeRecord, error) {
	var result volumeRecord
	err := db.Table("cdm_cluster_volume").
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_volume.cluster_tenant_id = cdm_cluster_tenant.id").
		Joins("join cdm_cluster_storage on cdm_cluster_volume.cluster_storage_id = cdm_cluster_storage.id").
		Where(&model.ClusterTenant{ClusterID: clusterID}).
		Where(&model.ClusterStorage{ClusterID: clusterID}).
		Where(&model.ClusterVolume{ID: volumeID}).
		First(&result).Error

	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterVolume(volumeID)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &result, nil
}

func getVolumeRecordByUUID(db *gorm.DB, clusterID uint64, uuid string) (*volumeRecord, error) {
	var result volumeRecord
	err := db.Table("cdm_cluster_volume").
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_volume.cluster_tenant_id = cdm_cluster_tenant.id").
		Joins("join cdm_cluster_storage on cdm_cluster_volume.cluster_storage_id = cdm_cluster_storage.id").
		Where(&model.ClusterTenant{ClusterID: clusterID}).
		Where(&model.ClusterStorage{ClusterID: clusterID}).
		Where(&model.ClusterVolume{UUID: uuid}).
		First(&result).Error

	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundClusterVolumeByUUID(uuid)

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &result, nil
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

func getVolumeRecords(db *gorm.DB, clusterID uint64, filters ...volumeFilter) ([]volumeRecord, error) {
	var err error
	cond := db
	for _, f := range filters {
		cond, err = f.Apply(cond)
		if err != nil {
			return nil, err
		}
	}

	var result []volumeRecord
	err = cond.Table("cdm_cluster_volume").
		Select("*").
		Joins("join cdm_cluster_tenant on cdm_cluster_volume.cluster_tenant_id = cdm_cluster_tenant.id").
		Joins("join cdm_cluster_storage on cdm_cluster_volume.cluster_storage_id = cdm_cluster_storage.id").
		Where(&model.ClusterTenant{ClusterID: clusterID}).
		Where(&model.ClusterStorage{ClusterID: clusterID}).
		Find(&result).Error

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetList 클러스터 볼륨 목록을 조회한다.
func GetList(ctx context.Context, req *cms.ClusterVolumeListRequest) ([]*cms.ClusterVolume, *cms.Pagination, error) {
	var err error

	if err = database.Execute(func(db *gorm.DB) error {
		return validateVolumeListRequest(ctx, db, req)
	}); err != nil {
		logger.Errorf("[Volume-GetList] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	var c *model.Cluster
	var cmsCluster cms.Cluster
	var vrs []volumeRecord
	var filters []volumeFilter
	if err = database.Execute(func(db *gorm.DB) error {
		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Volume-GetList] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		if err = cmsCluster.SetFromModel(c); err != nil {
			return err
		}

		filters = makeVolumeFilter(db, req)

		if vrs, err = getVolumeRecords(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[Volume-GetList] Could not get the cluster(%d) volume records. Cause: %+v", req.ClusterId, err)
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	var rsp []*cms.ClusterVolume
	for _, v := range vrs {
		var volume cms.ClusterVolume
		if err = volume.SetFromModel(&v.Volume); err != nil {
			return nil, nil, err
		}

		volume.Cluster = &cmsCluster

		volume.Tenant = new(cms.ClusterTenant)
		if err = volume.Tenant.SetFromModel(&v.Tenant); err != nil {
			return nil, nil, err
		}

		volume.Storage = new(cms.ClusterStorage)
		if err = volume.Storage.SetFromModel(&v.Storage); err != nil {
			return nil, nil, err
		}

		var snapshots []model.ClusterVolumeSnapshot
		if err = database.Execute(func(db *gorm.DB) error {
			if snapshots, err = v.Volume.GetVolumeSnapshots(db); err != nil {
				return errors.UnusableDatabase(err)
			}

			return nil
		}); err != nil {
			return nil, nil, err
		}

		for _, s := range snapshots {
			snapshot := new(cms.ClusterVolumeSnapshot)
			if err = snapshot.SetFromModel(&s); err != nil {
				return nil, nil, err
			}
			volume.Snapshots = append(volume.Snapshots, snapshot)
		}

		rsp = append(rsp, &volume)
	}

	var pagination *cms.Pagination
	if err = database.Execute(func(db *gorm.DB) error {
		if pagination, err = getVolumePagination(db, req.ClusterId, filters...); err != nil {
			logger.Errorf("[Volume-GetList] Could not get the cluster(%d) volume list pagination. Cause: %+v", req.ClusterId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, nil, err
	}

	return rsp, pagination, nil
}

func getVolumePagination(db *gorm.DB, clusterID uint64, filters ...volumeFilter) (*cms.Pagination, error) {
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
	if err = conditions.Table(model.ClusterVolume{}.TableName()).
		Joins("join cdm_cluster_tenant on cdm_cluster_tenant.id = cdm_cluster_volume.cluster_tenant_id").
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

// Get 클러스터 볼륨을 조회한다.
func Get(ctx context.Context, req *cms.ClusterVolumeRequest) (*cms.ClusterVolume, error) {
	var err error
	var c *model.Cluster
	var v *volumeRecord

	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateVolumeRequest(ctx, db, req); err != nil {
			logger.Errorf("[Volume-Get] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Volume-Get] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		if v, err = getVolumeRecord(db, req.ClusterId, req.ClusterVolumeId); err != nil {
			logger.Errorf("[Volume-Get] Could not get the cluster(%d) volume(%d) record. Cause: %+v", req.ClusterId, req.ClusterVolumeId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var rsp cms.ClusterVolume
	if err = rsp.SetFromModel(&v.Volume); err != nil {
		return nil, err
	}

	rsp.Cluster = new(cms.Cluster)
	if err = rsp.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}

	rsp.Tenant = new(cms.ClusterTenant)
	if err = rsp.Tenant.SetFromModel(&v.Tenant); err != nil {
		return nil, err
	}

	rsp.Storage = new(cms.ClusterStorage)
	if err = rsp.Storage.SetFromModel(&v.Storage); err != nil {
		return nil, err
	}

	var snapshots []model.ClusterVolumeSnapshot
	if err = database.Execute(func(db *gorm.DB) error {
		if snapshots, err = v.Volume.GetVolumeSnapshots(db); err != nil {
			return errors.UnusableDatabase(err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	for _, s := range snapshots {
		snapshot := new(cms.ClusterVolumeSnapshot)
		if err = snapshot.SetFromModel(&s); err != nil {
			return nil, err
		}

		rsp.Snapshots = append(rsp.Snapshots, snapshot)
	}
	return &rsp, nil
}

// GetByUUID uuid 를 통해 클러스터 볼륨을 조회
func GetByUUID(ctx context.Context, req *cms.ClusterVolumeByUUIDRequest) (*cms.ClusterVolume, error) {
	var err error
	var c *model.Cluster
	if err = database.Execute(func(db *gorm.DB) error {
		if err = validateVolumeByUUIDRequest(ctx, db, req); err != nil {
			logger.Errorf("[Volume-GetByUUID] Errors occurred during validating the request. Cause: %+v", err)
			return err
		}

		if c, err = getCluster(ctx, db, req.ClusterId); err != nil {
			logger.Errorf("[Volume-GetByUUID] Could not get the cluster(%d). Cause: %+v", req.ClusterId, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var v *volumeRecord
	if err = database.Execute(func(db *gorm.DB) error {
		if v, err = getVolumeRecordByUUID(db, req.ClusterId, req.Uuid); err != nil {
			logger.Errorf("[Volume-GetByUUID] Could not get the cluster(%d) volume record by uuid(%s). Cause: %+v", req.ClusterId, req.Uuid, err)
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var rsp cms.ClusterVolume
	if err = rsp.SetFromModel(&v.Volume); err != nil {
		return nil, err
	}

	rsp.Cluster = new(cms.Cluster)
	if err = rsp.Cluster.SetFromModel(c); err != nil {
		return nil, err
	}

	rsp.Tenant = new(cms.ClusterTenant)
	if err = rsp.Tenant.SetFromModel(&v.Tenant); err != nil {
		return nil, err
	}

	rsp.Storage = new(cms.ClusterStorage)
	if err = rsp.Storage.SetFromModel(&v.Storage); err != nil {
		return nil, err
	}

	var snapshots []model.ClusterVolumeSnapshot
	if err = database.Execute(func(db *gorm.DB) error {
		if snapshots, err = v.Volume.GetVolumeSnapshots(db); err != nil {
			return errors.UnusableDatabase(err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	for _, s := range snapshots {
		snapshot := new(cms.ClusterVolumeSnapshot)
		if err = snapshot.SetFromModel(&s); err != nil {
			return nil, err
		}

		rsp.Snapshots = append(rsp.Snapshots, snapshot)
	}
	return &rsp, nil
}

// CreateVolume 클러스터 볼륨 생성
func CreateVolume(req *cms.CreateClusterVolumeRequest) (*cms.ClusterVolume, error) {
	if err := validateCreateClusterVolumeRequest(req); err != nil {
		logger.Errorf("[CreateVolume] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateVolume] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[CreateVolume] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateVolume] Could not close client. Cause: %+v", err)
		}
	}()

	v, err := req.GetVolume().Model()
	if err != nil {
		return nil, err
	}

	res, err := cli.CreateVolume(client.CreateVolumeRequest{
		Tenant: model.ClusterTenant{
			UUID: req.GetTenant().GetUuid(),
		},
		Storage: model.ClusterStorage{
			Name: req.GetStorage().GetName(),
		},
		Volume: *v,
	})
	if err != nil {
		logger.Errorf("[CreateVolume] Could not create the cluster volume. Cause: %+v", err)
		return nil, err
	}

	var volume cms.ClusterVolume
	if err = volume.SetFromModel(&res.Volume); err != nil {
		return nil, err
	}

	return &volume, nil
}

// ImportVolume 클러스터 볼륨 import
func ImportVolume(req *cms.ImportClusterVolumeRequest) (*cms.VolumePair, []*cms.SnapshotPair, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[ImportVolume] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, nil, err
	}

	if err = validateImportClusterVolumeRequest(req); err != nil {
		logger.Errorf("[ImportVolume] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[ImportVolume] Could not create client. Cause: %+v", err)
		return nil, nil, err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[ImportVolume] Could not connect client. Cause: %+v", err)
		return nil, nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[ImportVolume] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return nil, nil, err
	}

	mSourceStorage, err := req.GetSourceStorage().Model()
	if err != nil {
		return nil, nil, err
	}

	mTargetStorage, err := req.GetTargetStorage().Model()
	if err != nil {
		return nil, nil, err
	}

	mTargetMetadata := req.GetTargetMetadata()
	if len(mTargetMetadata) == 0 {
		return nil, nil, err
	}

	mVolume, err := req.GetVolume().Model()
	if err != nil {
		return nil, nil, err
	}

	var mSnapshots []model.ClusterVolumeSnapshot
	for _, snapshot := range req.GetSnapshots() {
		mSnapshot, err := snapshot.Model()
		if err != nil {
			return nil, nil, err
		}
		mSnapshots = append(mSnapshots, *mSnapshot)
	}

	res, err := cli.ImportVolume(client.ImportVolumeRequest{
		Tenant:          *mTenant,
		SourceStorage:   *mSourceStorage,
		TargetStorage:   *mTargetStorage,
		TargetMetadata:  mTargetMetadata,
		Volume:          *mVolume,
		VolumeSnapshots: mSnapshots,
	})
	if err != nil {
		logger.Errorf("[ImportVolume] Could not import the cluster volume. Cause: %+v", err)
		return nil, nil, err
	}

	volPair := cms.VolumePair{
		Source:     &cms.ClusterVolume{},
		Target:     &cms.ClusterVolume{},
		SourceFile: res.VolumePair.SourceFile,
		TargetFile: res.VolumePair.TargetFile,
	}
	if err = volPair.Source.SetFromModel(&res.VolumePair.Source); err != nil {
		return nil, nil, err
	}
	if err = volPair.Target.SetFromModel(&res.VolumePair.Target); err != nil {
		return nil, nil, err
	}

	var snpPairs []*cms.SnapshotPair
	for _, pair := range res.SnapshotPairs {
		snpPair := cms.SnapshotPair{
			Source:     &cms.ClusterVolumeSnapshot{},
			Target:     &cms.ClusterVolumeSnapshot{},
			SourceFile: pair.SourceFile,
			TargetFile: pair.TargetFile,
		}
		if err = snpPair.Source.SetFromModel(&pair.Source); err != nil {
			return nil, nil, err
		}
		if err = snpPair.Target.SetFromModel(&pair.Target); err != nil {
			return nil, nil, err
		}
		snpPairs = append(snpPairs, &snpPair)
	}

	return &volPair, snpPairs, nil
}

// DeleteVolume 클러스터 볼륨 삭제
func DeleteVolume(req *cms.DeleteClusterVolumeRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteVolume] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteClusterVolumeRequest(req); err != nil {
		logger.Errorf("[DeleteVolume] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteVolume] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[DeleteVolume] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteVolume] Could not close client. Cause: %+v", err)
		}
	}()

	v, err := req.GetVolume().Model()
	if err != nil {
		return err
	}

	err = cli.DeleteVolume(client.DeleteVolumeRequest{
		Tenant: model.ClusterTenant{
			UUID: req.GetTenant().GetUuid(),
		},
		Volume: *v,
	})
	if err != nil {
		logger.Errorf("[DeleteVolume] Could not delete the cluster volume. Cause: %+v", err)
		return err
	}

	return nil
}

// SyncVolumeSnapshotList 클러스터 볼륨 스냅샷 목록 동기화
func SyncVolumeSnapshotList(req *cms.SyncClusterVolumeSnapshotListRequest) error {
	s, err := sync.NewSynchronizer(req.Cluster.Id)
	if err != nil {
		logger.Errorf("[SyncVolumeSnapshotList] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	defer func() {
		if err = s.Close(); err != nil {
			logger.Warnf("[SyncVolumeSnapshotList] Could not close synchronizer client. Cause: %+v", err)
		}
	}()

	return s.SyncClusterVolumeSnapshotList(sync.Force(true))
}

// CreateVolumeSnapshot 클러스터 볼륨 스냅샷 생성
func CreateVolumeSnapshot(req *cms.CreateClusterVolumeSnapshotRequest) (*cms.ClusterVolumeSnapshot, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateVolumeSnapshot] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err = validateCreateClusterVolumeSnapshotRequest(req); err != nil {
		logger.Errorf("[CreateVolumeSnapshot] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateVolumeSnapshot] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[CreateVolumeSnapshot] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateVolumeSnapshot] Could not close client. Cause: %+v", err)
		}
	}()

	s, err := req.GetSnapshot().Model()
	if err != nil {
		return nil, err
	}

	res, err := cli.CreateVolumeSnapshot(client.CreateVolumeSnapshotRequest{
		Tenant: model.ClusterTenant{
			UUID: req.GetTenant().GetUuid(),
		},
		Volume: model.ClusterVolume{
			UUID: req.GetVolume().GetUuid(),
		},
		VolumeSnapshot: *s,
	})
	if err != nil {
		logger.Errorf("[CreateVolumeSnapshot] Could not create the volume snapshot. Cause: %+v", err)
		return nil, err
	}

	var volumeSnapshot cms.ClusterVolumeSnapshot
	if err = volumeSnapshot.SetFromModel(&res.VolumeSnapshot); err != nil {
		return nil, err
	}

	return &volumeSnapshot, nil
}

// DeleteVolumeSnapshot 클러스터 볼륨 스냅샷 삭제
func DeleteVolumeSnapshot(req *cms.DeleteClusterVolumeSnapshotRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteVolumeSnapshot] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteClusterVolumeSnapshotRequest(req); err != nil {
		logger.Errorf("[DeleteVolumeSnapshot] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteVolumeSnapshot] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[DeleteVolumeSnapshot] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteVolumeSnapshot] Could not close client. Cause: %+v", err)
		}
	}()

	s, err := req.GetSnapshot().Model()
	if err != nil {
		return err
	}

	err = cli.DeleteVolumeSnapshot(client.DeleteVolumeSnapshotRequest{
		Tenant: model.ClusterTenant{
			UUID: req.GetTenant().GetUuid(),
		},
		VolumeSnapshot: *s,
	})
	if err != nil {
		logger.Errorf("[DeleteVolumeSnapshot] Could not delete the volume snapshot. Cause: %+v", err)
		return err
	}

	return nil
}

// UnmanageVolume 볼륨 unmanage
func UnmanageVolume(req *cms.UnmanageClusterVolumeRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[UnmanageVolume] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateUnmanageClusterVolumeRequest(req); err != nil {
		logger.Errorf("[UnmanageVolume] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[UnmanageVolume] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[UnmanageVolume] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[UnmanageVolume] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return err
	}
	mSourceStorage, err := req.GetSourceStorage().Model()
	if err != nil {
		return err
	}
	mTargetStorage, err := req.GetTargetStorage().Model()
	if err != nil {
		return err
	}

	mTargetMetadata := req.GetTargetMetadata()
	if len(mTargetMetadata) == 0 {
		return err
	}

	mSourceVolume, err := req.GetVolumePair().GetSource().Model()
	if err != nil {
		return err
	}

	mTargetVolume, err := req.GetVolumePair().GetTarget().Model()
	if err != nil {
		return err
	}

	var mSnapshotPairs []client.SnapshotPair
	for _, pair := range req.GetSnapshotPairs() {
		mSnapshotSource, err := pair.GetSource().Model()
		if err != nil {
			return err
		}

		mSnapshotTarget, err := pair.GetTarget().Model()
		if err != nil {
			return err
		}

		mSnapshotPairs = append(mSnapshotPairs, client.SnapshotPair{
			Source:     *mSnapshotSource,
			Target:     *mSnapshotTarget,
			SourceFile: pair.GetSourceFile(),
			TargetFile: pair.GetTargetFile(),
		})
	}

	return cli.UnmanageVolume(client.UnmanageVolumeRequest{
		Tenant:         *mTenant,
		SourceStorage:  *mSourceStorage,
		TargetStorage:  *mTargetStorage,
		TargetMetadata: mTargetMetadata,
		VolumePair: client.VolumePair{
			Source:     *mSourceVolume,
			Target:     *mTargetVolume,
			SourceFile: req.VolumePair.GetSourceFile(),
			TargetFile: req.VolumePair.GetTargetFile(),
		},
		SnapshotPairs: mSnapshotPairs,
		VolumeID:      req.VolumeId,
	})
}

// CopyVolume 볼륨 copy
func CopyVolume(req *cms.CopyClusterVolumeRequest) (*cms.ClusterVolume, []*cms.ClusterVolumeSnapshot, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CopyVolume] Error occurred during checking the cluster connection info. Cause: %+v", err)

		return nil, nil, err
	}

	if err = validateCopyClusterVolumeRequest(req); err != nil {
		logger.Errorf("[CopyVolume] Errors occurred during validating the request. Cause: %+v", err)
		return nil, nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CopyVolume] Could not create client. Cause: %+v", err)
		return nil, nil, err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[CopyVolume] Could not connect client. Cause: %+v", err)
		return nil, nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CopyVolume] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return nil, nil, err
	}

	mSourceStorage, err := req.GetSourceStorage().Model()
	if err != nil {
		return nil, nil, err
	}

	mTargetStorage, err := req.GetTargetStorage().Model()
	if err != nil {
		return nil, nil, err
	}

	mTargetMetadata := req.GetTargetMetadata()
	if len(mTargetMetadata) == 0 {
		return nil, nil, err
	}

	mVolume, err := req.GetVolume().Model()
	if err != nil {
		return nil, nil, err
	}

	var mSnapshots []model.ClusterVolumeSnapshot
	for _, snapshot := range req.GetSnapshots() {
		mSnapshot, err := snapshot.Model()
		if err != nil {
			return nil, nil, err
		}

		mSnapshots = append(mSnapshots, *mSnapshot)
	}

	rsp, err := cli.CopyVolume(client.CopyVolumeRequest{
		Tenant:         *mTenant,
		SourceStorage:  *mSourceStorage,
		TargetStorage:  *mTargetStorage,
		TargetMetadata: mTargetMetadata,
		Volume:         *mVolume,
		Snapshots:      mSnapshots,
	})
	if err != nil {
		logger.Errorf("[CopyVolume] Could not copy the cluster volume. Cause: %+v", err)
		return nil, nil, err
	}

	var retVol cms.ClusterVolume
	var retSnpList []*cms.ClusterVolumeSnapshot

	if err := retVol.SetFromModel(&rsp.Volume); err != nil {
		return nil, nil, err
	}

	//for _, snapshot := range rsp.Snapshots {
	//	var cSnp cms.ClusterVolumeSnapshot
	//	if err := cSnp.SetFromModel(&snapshot); err != nil {
	//		return nil, nil, err
	//	}
	//
	//	retSnpList = append(retSnpList, &cSnp)
	//}

	return &retVol, retSnpList, nil
}

// DeleteVolumeCopy 볼륨 copy 삭제
func DeleteVolumeCopy(req *cms.DeleteClusterVolumeCopyRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteVolumeCopy] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteClusterVolumeCopyRequest(req); err != nil {
		logger.Errorf("[DeleteVolumeCopy] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteVolumeCopy] Could not create client. Cause: %+v", err)
		return err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[DeleteVolumeCopy] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteVolumeCopy] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return err
	}

	mSourceStorage, err := req.GetSourceStorage().Model()
	if err != nil {
		return err
	}

	mTargetStorage, err := req.GetTargetStorage().Model()
	if err != nil {
		return err
	}

	mTargetMetadata := req.GetTargetMetadata()
	if len(mTargetMetadata) == 0 {
		return err
	}

	mVolume, err := req.GetVolume().Model()
	if err != nil {
		return err
	}

	var mSnapshots []model.ClusterVolumeSnapshot
	for _, snapshot := range req.GetSnapshots() {
		mSnapshot, err := snapshot.Model()
		if err != nil {
			return err
		}

		mSnapshots = append(mSnapshots, *mSnapshot)
	}

	return cli.DeleteVolumeCopy(client.DeleteVolumeCopyRequest{
		Tenant:         *mTenant,
		SourceStorage:  *mSourceStorage,
		TargetStorage:  *mTargetStorage,
		TargetMetadata: mTargetMetadata,
		Volume:         *mVolume,
		VolumeID:       req.VolumeId,
		Snapshots:      mSnapshots,
	})
}
