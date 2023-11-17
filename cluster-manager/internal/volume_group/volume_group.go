package volumegroup

import (
	"context"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/tenant"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
)

// GetVolumeGroupList 클러스터 볼륨 그룹 목록을 조회한다.
func GetVolumeGroupList(db *gorm.DB, req *cms.GetClusterVolumeGroupListRequest) ([]*cms.ClusterVolumeGroup, error) {
	var (
		err     error
		conn    *cms.ClusterConnectionInfo
		listRsp *client.GetVolumeGroupListResponse
		rsp     *client.GetVolumeGroupResponse
		groups  []*cms.ClusterVolumeGroup
	)

	if req.GetClusterId() == 0 {
		return nil, errors.RequiredParameter("cluster.id")
	}

	if conn, err = cluster.GetConnectionInfo(db, req.ClusterId); err != nil {
		logger.Errorf("[GetVolumeGroupList] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	conn.Credential, err = client.DecryptCredentialPassword(conn.GetCredential())
	if err != nil {
		logger.Errorf("[GetVolumeGroupList] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(conn.GetTypeCode(), conn.GetApiServerUrl(), conn.GetCredential(), conn.GetTenantId())
	if err != nil {
		logger.Errorf("[GetVolumeGroupList] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[GetVolumeGroupList] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[GetVolumeGroupList] Could not close client. Cause: %+v", err)
		}
	}()

	if listRsp, err = cli.GetVolumeGroupList(); err != nil {
		logger.Errorf("[GetVolumeGroupList] Could not get the volume group list. Cause: %+v", err)
		return nil, err
	}

	for _, vg := range listRsp.VolumeGroupList {
		if rsp, err = cli.GetVolumeGroup(client.GetVolumeGroupRequest{
			VolumeGroup: client.VolumeGroup{UUID: vg.UUID},
		}); err != nil {
			logger.Errorf("[GetVolumeGroupList] Could not get the volume group. Cause: %+v", err)
			return nil, err
		}

		var group = &cms.ClusterVolumeGroup{
			Uuid:        rsp.VolumeGroup.UUID,
			Name:        rsp.VolumeGroup.Name,
			Description: rsp.VolumeGroup.Description,
		}

		for _, v := range rsp.VolumeGroup.VolumeList {
			var m model.ClusterVolume
			if err = db.Where(&model.ClusterVolume{UUID: v.UUID}).First(&m).Error; err != nil {
				logger.Errorf("[GetVolumeGroupList] Could not get the volume. Cause: %+v", err)
				return nil, errors.UnusableDatabase(err)
			}

			var volume cms.ClusterVolume
			if err := volume.SetFromModel(&m); err != nil {
				return nil, err
			}

			group.Volumes = append(group.Volumes, &volume)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// GetVolumeGroupByUUID 클러스터 볼륨 그룹을 조회한다.
// volume group 을 따로 저장하고 있지 않기 때문에, 클러스터에서 바로 조회한다.
func GetVolumeGroupByUUID(ctx context.Context, db *gorm.DB, req *cms.GetClusterVolumeGroupByUUIDRequest) (*cms.ClusterVolumeGroup, error) {
	var (
		err  error
		conn *cms.ClusterConnectionInfo
		rsp  *client.GetVolumeGroupResponse
	)

	if err = validateGetClusterVolumeGroupRequest(db, req); err != nil {
		return nil, err
	}

	if conn, err = cluster.GetConnectionInfo(db, req.ClusterId); err != nil {
		logger.Errorf("[GetVolumeGroupByUUID] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	conn.Credential, err = client.DecryptCredentialPassword(conn.GetCredential())
	if err != nil {
		logger.Errorf("[GetVolumeGroupByUUID] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(conn.GetTypeCode(), conn.GetApiServerUrl(), conn.GetCredential(), conn.GetTenantId())
	if err != nil {
		logger.Errorf("[GetVolumeGroupByUUID] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[GetVolumeGroupByUUID] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[GetVolumeGroupByUUID] Could not close client. Cause: %+v", err)
		}
	}()

	if rsp, err = cli.GetVolumeGroup(client.GetVolumeGroupRequest{
		Tenant:      model.ClusterTenant{ID: req.ClusterTenantId, ClusterID: req.ClusterId},
		VolumeGroup: client.VolumeGroup{UUID: req.ClusterVolumeGroupUuid},
	}); err != nil {
		logger.Errorf("[GetVolumeGroupByUUID] Could not get the volume group. Cause: %+v", err)
		return nil, err
	}

	var group = cms.ClusterVolumeGroup{
		Uuid:        rsp.VolumeGroup.UUID,
		Name:        rsp.VolumeGroup.Name,
		Description: rsp.VolumeGroup.Description,
		Tenant:      new(cms.ClusterTenant),
	}

	if group.Tenant, err = tenant.Get(ctx, &cms.ClusterTenantRequest{ClusterId: req.ClusterId, ClusterTenantId: req.ClusterTenantId}); err != nil {
		return nil, err
	}

	for _, s := range rsp.VolumeGroup.StorageList {
		var m model.ClusterStorage
		if err := db.Where(&model.ClusterStorage{ClusterID: req.ClusterId, UUID: s.UUID}).First(&m).Error; err != nil {
			return nil, errors.UnusableDatabase(err)
		}

		var storage cms.ClusterStorage
		if err := storage.SetFromModel(&m); err != nil {
			return nil, err
		}

		group.Storages = append(group.Storages, &storage)
	}

	for _, v := range rsp.VolumeGroup.VolumeList {
		var m model.ClusterVolume
		if err := db.Where(&model.ClusterVolume{ClusterTenantID: req.ClusterTenantId, UUID: v.UUID}).First(&m).Error; err != nil {
			return nil, errors.UnusableDatabase(err)
		}

		var volume cms.ClusterVolume
		if err := volume.SetFromModel(&m); err != nil {
			return nil, err
		}

		group.Volumes = append(group.Volumes, &volume)
	}

	return &group, nil
}

// CreateVolumeGroup 볼륨 그룹 생성
func CreateVolumeGroup(req *cms.CreateClusterVolumeGroupRequest) (*cms.ClusterVolumeGroup, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateVolumeGroup] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err = validateCreateClusterVolumeGroupRequest(req); err != nil {
		logger.Errorf("[CreateVolumeGroup] Errors occurred during validating the request. Cause: %+v", err)

		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateVolumeGroup] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[CreateVolumeGroup] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateVolumeGroup] Could not close client. Cause: %+v", err)
		}
	}()

	tenant, err := req.GetTenant().Model()
	if err != nil {
		return nil, err
	}

	var mStorages []model.ClusterStorage
	for _, storage := range req.GetVolumeGroup().GetStorages() {
		mStorage, err := storage.Model()
		if err != nil {
			return nil, err
		}
		mStorages = append(mStorages, *mStorage)
	}

	group, err := cli.CreateVolumeGroup(client.CreateVolumeGroupRequest{
		Tenant: *tenant,
		VolumeGroup: client.VolumeGroup{
			Name:        req.GetVolumeGroup().GetName(),
			Description: req.GetVolumeGroup().GetDescription(),
			StorageList: mStorages,
		},
	})
	if err != nil {
		logger.Errorf("[CreateVolumeGroup] Could not create the volume group. Cause: %+v", err)
		return nil, err
	}

	var rollback = true
	defer func() {
		if rollback {
			if err := cli.DeleteVolumeGroup(client.DeleteVolumeGroupRequest{
				Tenant: *tenant,
				VolumeGroup: client.VolumeGroup{
					UUID:        group.VolumeGroup.UUID,
					Name:        group.VolumeGroup.Name,
					Description: group.VolumeGroup.Description,
					StorageList: mStorages,
				},
			}); err != nil {
				logger.Warnf("[CreateVolumeGroup] Could not delete cluster volume group. Cause: %+v", err)
			}
		}
	}()

	var cStorages []*cms.ClusterStorage
	for _, storage := range group.VolumeGroup.StorageList {
		var cStorage cms.ClusterStorage
		if err := cStorage.SetFromModel(&storage); err != nil {
			return nil, err
		}
		cStorages = append(cStorages, &cStorage)
	}

	rollback = false
	return &cms.ClusterVolumeGroup{
		Uuid:        group.VolumeGroup.UUID,
		Name:        group.VolumeGroup.Name,
		Description: group.VolumeGroup.Description,
		Storages:    cStorages,
	}, nil
}

// DeleteVolumeGroup 볼륨 그룹 삭제
func DeleteVolumeGroup(req *cms.DeleteClusterVolumeGroupRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteVolumeGroup] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteClusterVolumeGroupRequest(req); err != nil {
		logger.Errorf("[DeleteVolumeGroup] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteVolumeGroup] Could not create client. Cause: %+v", err)
		return err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[DeleteVolumeGroup] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteVolumeGroup] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return err
	}

	var mStorages []model.ClusterStorage
	for _, storage := range req.GetVolumeGroup().GetStorages() {
		mStorage, err := storage.Model()
		if err != nil {
			return err
		}
		mStorages = append(mStorages, *mStorage)
	}

	return cli.DeleteVolumeGroup(client.DeleteVolumeGroupRequest{
		Tenant: *mTenant,
		VolumeGroup: client.VolumeGroup{
			UUID:        req.GetVolumeGroup().GetUuid(),
			Name:        req.GetVolumeGroup().GetName(),
			Description: req.GetVolumeGroup().GetDescription(),
			StorageList: mStorages,
		},
	})
}

// UpdateVolumeGroup 볼륨 그룹 수정
func UpdateVolumeGroup(req *cms.UpdateClusterVolumeGroupRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[UpdateVolumeGroup] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateUpdateClusterVolumeGroupRequest(req); err != nil {
		logger.Errorf("[UpdateVolumeGroup] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[UpdateVolumeGroup] Could not create client. Cause: %+v", err)
		return err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[UpdateVolumeGroup] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[UpdateVolumeGroup] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return err
	}

	var mStorages []model.ClusterStorage
	for _, storage := range req.GetVolumeGroup().GetStorages() {
		mStorage, err := storage.Model()
		if err != nil {
			return err
		}
		mStorages = append(mStorages, *mStorage)
	}

	var mAddVolumes []model.ClusterVolume
	for _, volume := range req.GetAddVolumes() {
		mAddVolume, err := volume.Model()
		if err != nil {
			return err
		}

		mAddVolumes = append(mAddVolumes, *mAddVolume)
	}

	var mDelVolumes []model.ClusterVolume
	for _, volume := range req.GetDeleteVolumes() {
		mDelVolume, err := volume.Model()
		if err != nil {
			return err
		}

		mDelVolumes = append(mDelVolumes, *mDelVolume)
	}

	return cli.UpdateVolumeGroup(client.UpdateVolumeGroupRequest{
		Tenant: *mTenant,
		VolumeGroup: client.VolumeGroup{
			UUID:        req.GetVolumeGroup().GetUuid(),
			Name:        req.GetVolumeGroup().GetName(),
			Description: req.GetVolumeGroup().GetDescription(),
			StorageList: mStorages,
		},
		AddVolumes: mAddVolumes,
		DelVolumes: mDelVolumes,
	})
}

// CreateVolumeGroupSnapshot 볼륨 그룹 스냅샷 생성
func CreateVolumeGroupSnapshot(req *cms.CreateClusterVolumeGroupSnapshotRequest) (*cms.ClusterVolumeGroupSnapshot, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[CreateVolumeGroupSnapshot] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err = validateCreateClusterVolumeGroupSnapshotRequest(req); err != nil {
		logger.Errorf("[CreateVolumeGroupSnapshot] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[CreateVolumeGroupSnapshot] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[CreateVolumeGroupSnapshot] Could not connect client. Cause: %+v", err)
		return nil, err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[CreateVolumeGroupSnapshot] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return nil, err
	}

	rsp, err := cli.CreateVolumeGroupSnapshot(client.CreateVolumeGroupSnapshotRequest{
		Tenant: *mTenant,
		VolumeGroup: client.VolumeGroup{
			UUID: req.GetVolumeGroup().GetUuid(),
		},
		VolumeGroupSnapshot: client.VolumeGroupSnapshot{
			Name:        req.GetVolumeGroupSnapshot().GetName(),
			Description: req.GetVolumeGroupSnapshot().GetDescription(),
		},
	})
	if err != nil {
		logger.Errorf("[CreateVolumeGroupSnapshot] Could not create the volume group snapshot. Cause: %+v", err)
		return nil, err
	}

	return &cms.ClusterVolumeGroupSnapshot{
		Uuid:        rsp.VolumeGroupSnapshot.UUID,
		Name:        rsp.VolumeGroupSnapshot.Name,
		Description: rsp.VolumeGroupSnapshot.Description,
	}, nil
}

// DeleteVolumeGroupSnapshot 볼륨 그룹 스냅샷 삭제
func DeleteVolumeGroupSnapshot(req *cms.DeleteClusterVolumeGroupSnapshotRequest) error {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[DeleteVolumeGroupSnapshot] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return err
	}

	if err = validateDeleteClusterVolumeGroupSnapshotRequest(req); err != nil {
		logger.Errorf("[DeleteVolumeGroupSnapshot] Errors occurred during validating the request. Cause: %+v", err)
		return err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[DeleteVolumeGroupSnapshot] Could not create client. Cause: %+v", err)
		return err
	}

	if err := cli.Connect(); err != nil {
		logger.Errorf("[DeleteVolumeGroupSnapshot] Could not connect client. Cause: %+v", err)
		return err
	}
	defer func() {
		if err := cli.Close(); err != nil {
			logger.Warnf("[DeleteVolumeGroupSnapshot] Could not close client. Cause: %+v", err)
		}
	}()

	mTenant, err := req.GetTenant().Model()
	if err != nil {
		return err
	}

	return cli.DeleteVolumeGroupSnapshot(client.DeleteVolumeGroupSnapshotRequest{
		Tenant:              *mTenant,
		VolumeGroup:         client.VolumeGroup{UUID: req.GetVolumeGroup().GetUuid()},
		VolumeGroupSnapshot: client.VolumeGroupSnapshot{UUID: req.GetVolumeGroupSnapshot().GetUuid()},
	})
}

// GetClusterVolumeGroupSnapshotList 클러스터 볼륨 그룹 목록 조회
func GetClusterVolumeGroupSnapshotList(req *cms.GetClusterVolumeGroupSnapshotListRequest) ([]*cms.GetClusterVolumeGroupSnapshotListResult, error) {
	var err error
	req.Conn.Credential, err = client.DecryptCredentialPassword(req.Conn.GetCredential())
	if err != nil {
		logger.Errorf("[GetClusterVolumeGroupSnapshotList] Error occurred during checking the cluster connection info. Cause: %+v", err)
		return nil, err
	}

	if err = validateGetClusterVolumeGroupSnapshotListRequest(req); err != nil {
		logger.Errorf("[GetClusterVolumeGroupSnapshotList] Errors occurred during validating the request. Cause: %+v", err)
		return nil, err
	}

	cli, err := client.New(req.GetConn().GetTypeCode(), req.GetConn().GetApiServerUrl(), req.GetConn().GetCredential(), req.GetConn().GetTenantId())
	if err != nil {
		logger.Errorf("[GetClusterVolumeGroupSnapshotList] Could not create client. Cause: %+v", err)
		return nil, err
	}

	if err = cli.Connect(); err != nil {
		logger.Errorf("[GetClusterVolumeGroupSnapshotList] Could not connect client. Cause: %+v", err)
		return nil, err
	}

	rsp, err := cli.GetVolumeGroupSnapshotList()
	if err != nil {
		logger.Errorf("[GetClusterVolumeGroupSnapshotList] Could not get the volume group snapshot list. Cause: %+v", err)
		return nil, err
	}

	var r []*cms.GetClusterVolumeGroupSnapshotListResult
	for _, vgs := range rsp.VolumeGroupSnapshots {
		r = append(r, &cms.GetClusterVolumeGroupSnapshotListResult{
			VolumeGroup: &cms.ClusterVolumeGroup{
				Uuid: vgs.VolumeGroup.UUID,
			},
			VolumeGroupSnapshot: &cms.ClusterVolumeGroupSnapshot{
				Uuid:        vgs.VolumeGroupSnapshot.UUID,
				Name:        vgs.VolumeGroupSnapshot.Name,
				Description: vgs.VolumeGroupSnapshot.Description,
			},
		})
	}

	return r, nil
}
