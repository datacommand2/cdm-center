package openstack

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack/train"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"strings"
	"time"
)

// copyCEPH
func copyCEPH(_ *Client, req *client.CopyVolumeRequest) (*client.CopyVolumeResponse, error) {
	rsp := &client.CopyVolumeResponse{
		Volume: req.Volume,
	}

	logger.Infof("[copyCEPH] Success: %+v", *rsp)
	return rsp, nil
}

// importCEPH
func importCEPH(cli *Client, req *client.ImportVolumeRequest) (*client.ImportVolumeResponse, error) {
	ceph, err := newDirectCEPH(req.TargetStorage, req.TargetMetadata)
	if err != nil {
		logger.Errorf("[importCEPH] newDirectCEPH: source storage(%d) target storage(%d) volume(%d). Cause: %+v",
			req.SourceStorage.ID, req.TargetStorage.ID, req.Volume.ID, err)
		return nil, err
	}

	backendName := ceph.getBackendName()
	serviceHostname, err := getServiceHostname(cli, backendName)
	if err != nil {
		return nil, err
	}

	var retVolPair *client.VolumePair
	var retSnpPairs []client.SnapshotPair

	volName := cephVolumeName(req.Volume.UUID)

	volumeRsp, err := train.ManageExistingVolume(cli, train.ManageExistingVolumeRequest{
		Volume: train.ManageVolumeRequest{
			Name:           req.Volume.Name,
			Description:    req.Volume.Description,
			VolumeTypeUUID: req.TargetStorage.UUID,
			Bootable:       req.Volume.Bootable,
			Host:           serviceHostname,
			Ref: train.VolumeReference{
				SourceName: volName,
			},
		},
	})
	if err != nil {
		logger.Errorf("[importCEPH] Could not manage existing volume(uuid:%s). Cause:%+v", req.Volume.UUID, err)
		return nil, err
	}

	// 볼륨 정보가 nil 이 아닌 경우, return 값으로 볼륨 pair 정보를 채운다.
	if volumeRsp != nil {
		retVolPair = &client.VolumePair{
			Source: req.Volume,
			Target: model.ClusterVolume{
				UUID:        volumeRsp.Volume.UUID,
				Name:        volumeRsp.Volume.Name,
				Description: volumeRsp.Volume.Description,
				SizeBytes:   volumeRsp.Volume.SizeGibiBytes * 1024 * 1024 * 1024,
				Multiattach: volumeRsp.Volume.Multiattach,
				Bootable:    req.Volume.Bootable,
				Readonly:    req.Volume.Readonly,
				Raw:         volumeRsp.Raw,
			},
		}
	} else {
		logger.Infof("[importCEPH] ManageExistingVolume response nil: target VolumeTypeUUID(%s) source cephVolumeName(%s)", req.TargetStorage.UUID, volName)
	}

	// 볼륨 import 가 202(Accepted) 지만 실제 동작은 실패했을 경우 해당 볼륨의 이미지는 존재하지 않음
	// import 된 볼륨의 실제 이미지 존재 여부 확인 후 timeout(30초) 동안 이미지가 존재하지 않으면 볼륨 롤백하고 실패 처리
	managedVolumeName := cephVolumeName(volumeRsp.Volume.UUID)

	logger.Infof("[importCEPH] ManagedImage Done: volName(%s)", managedVolumeName)

	if err := updateVolumeMode(cli, updateVolumeModeRequest{
		VolumeUUID: volumeRsp.Volume.UUID,
		Bootable:   req.Volume.Bootable,
		Readonly:   req.Volume.Readonly,
	}); err != nil {
		logger.Errorf("[importCEPH] Could not update volume(uuid:%s) mode. Cause:%+v", req.Volume.UUID, err)
		return nil, err
	}

	logger.Infof("[importCEPH] updateVolumeMode Done: VolumeUUID(%s)", volumeRsp.Volume.UUID)

	for i, snapshot := range req.VolumeSnapshots {
		snapshotRsp, err := train.ManageExistingVolumeSnapshot(cli, train.ManageExistingVolumeSnapshotRequest{
			Snapshot: train.ManageVolumeSnapshotRequest{
				Name:        snapshot.Name,
				Description: snapshot.Description,
				VolumeID:    volumeRsp.Volume.UUID,
				Ref: train.VolumeSnapshotReference{
					SourceName: cephSnapshotName(snapshot.UUID),
				},
			},
		})
		// volume snapshot import 실패 시 생성했던 볼륨 및 볼륨 스냅샷을 원래 상태로 롤백한다.
		if err != nil {
			logger.Errorf("[importCEPH] Could not manage existing volume snapshot(uuid:%s). Cause:%+v", snapshot.UUID, err)
			return nil, err
		}

		// 볼륨 스냅샷 정보가 nil 이 아닌 경우, return 값으로 볼륨 스냅샷 pair 정보를 채운다.
		if snapshotRsp != nil {
			retSnpPairs = append(retSnpPairs, client.SnapshotPair{
				Source: snapshot,
				Target: model.ClusterVolumeSnapshot{
					UUID:        snapshotRsp.Snapshot.UUID,
					Name:        snapshotRsp.Snapshot.Name,
					Description: snapshotRsp.Snapshot.Description,
					SizeBytes:   snapshotRsp.Snapshot.SizeGibiBytes * 1024 * 1024 * 1024,
					Raw:         snapshotRsp.Raw,
				},
			})
		} else {
			logger.Infof("[importCEPH] ManageExistingVolumeSnapshot response nil: VolumeID(%s) source cephSnapshotName(%s)", volumeRsp.Volume.UUID, cephSnapshotName(snapshot.UUID))
		}

		createdAt, err := time.Parse("2006-01-02T15:04:05.000000", snapshotRsp.Snapshot.CreatedAt)
		if err != nil {
			return nil, errors.Unknown(err)
		}

		retSnpPairs[i].Target.CreatedAt = createdAt.Unix()
	}

	rsp := &client.ImportVolumeResponse{
		VolumePair:    *retVolPair,
		SnapshotPairs: retSnpPairs,
	}

	logger.Infof("[importCEPH] Success: %+v", *rsp)
	return rsp, nil
}

func unmanageCEPH(cli *Client, req *client.UnmanageVolumeRequest) error {
	// unmanage snapshots and volume
	for _, pair := range req.SnapshotPairs {
		if err := train.UnmanageVolumeSnapshot(cli, train.UnmanageVolumeSnapshotRequest{
			SnapshotUUID: pair.Target.UUID,
		}); err != nil {
			logger.Errorf("[unmanageCEPH] UnmanageVolumeSnapshot: SnapshotUUID(%s). Cause: %+v", pair.Target.UUID, err)
			return err
		}
		logger.Infof("[unmanageCEPH] UnmanageVolumeSnapshot Done: SnapshotUUID(%s)", pair.Target.UUID)
	}

	if err := train.UnmanageVolume(cli, train.UnmanageVolumeRequest{
		VolumeUUID: req.VolumePair.Target.UUID,
	}); err != nil {
		logger.Errorf("[unmanageCEPH] UnmanageVolume: VolumeUUID(%s). Cause: %+v", req.VolumePair.Target.UUID, err)
		return err
	}
	logger.Infof("[unmanageCEPH] UnmanageVolume Done: VolumeUUID(%s)", req.VolumePair.Target.UUID)

	return nil
}

// deleteCopyCEPH
func deleteCopyCEPH(_ *Client, req *client.DeleteVolumeCopyRequest) error {

	return nil
}

func getServiceHostname(c client.Client, backendName string) (string, error) {
	servicesRsp, err := train.ListAllCinderServices(c)
	if err != nil {
		return "", err
	}

	var host string
	for _, service := range servicesRsp.Services {
		arr := strings.Split(service.Host, "@")
		if len(arr) != 2 {
			continue
		}

		if arr[1] == backendName {
			if service.State != "up" || service.Status != "enabled" {
				logger.Warnf("service[%s] is unavailable. state: (%s), status: (%s)", service.Host, service.State, service.Status)
				continue
			}

			host = service.Host
			break
		}
	}

	if host == "" {
		return "", NotFoundAvailableServiceHostName(backendName)
	}

	return host, nil
}
