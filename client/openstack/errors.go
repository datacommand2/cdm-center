package openstack

import (
	"fmt"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-disaster-recovery/common/mirror"
)

var (
	//ErrNotFoundVolumeType 볼륨 타입을 찾을 수 없음
	ErrNotFoundVolumeType = errors.New("not found volume type")
	// ErrNotFoundVolumeBackendName 볼륨 백엔드 네임을 찾을 수 없음
	ErrNotFoundVolumeBackendName = errors.New("not found volume backend name")
	// ErrNotFoundAvailableServiceHostName 사용 가능한 서비스 호스트 네임을 찾을 수 없음
	ErrNotFoundAvailableServiceHostName = errors.New("not found available service host name")
	// ErrNotFoundVolumeSourceName 볼륨 소스 네임을 찾을 수 없음
	ErrNotFoundVolumeSourceName = errors.New("not found volume source name")
	// ErrNotFoundSnapshotSourceName 스냅샷 소스 네임을 찾을 수 없음
	ErrNotFoundSnapshotSourceName = errors.New("not found snapshot source name")
	//ErrNotFoundInstanceHypervisor 인스턴스의 하이퍼바이저를 찾을 수 없음
	ErrNotFoundInstanceHypervisor = errors.New("not found instance's hypervisor")
	//ErrNotFoundInstanceSpec 인스턴스의 Spec 을 찾을 수 없음
	ErrNotFoundInstanceSpec = errors.New("not found instance's spec")
	//ErrUnknownInstanceState 알 수 없는 인스턴스 상태
	ErrUnknownInstanceState = errors.New("unknown instance's state")
	// ErrImportVolumeRollbackFailed import volume 의 rollback 실패
	ErrImportVolumeRollbackFailed = errors.New("fail to rollback import volume")
	// ErrNotFoundNFSExportPath NFS 의 Export 정보를 찾을 수 없음
	ErrNotFoundNFSExportPath = errors.New("not found NFS export path")
	// ErrUnusableMirrorVolume 사용 할 수 없는 복제 볼륨
	ErrUnusableMirrorVolume = errors.New("unusable mirror volume")
	// ErrNotFoundManagedVolume managed 볼륨을 찾을 수 없음
	ErrNotFoundManagedVolume = errors.New("not found managed volume")
)

// NotFoundManagedVolume managed 볼륨을 찾을 수 없음
func NotFoundManagedVolume(volumePair UUIDPair) error {
	return errors.Wrap(
		ErrNotFoundManagedVolume,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"volume_pair": volumePair,
		}),
	)
}

// UnusableMirrorVolume 사용 할 수 없는 복제 볼륨
func UnusableMirrorVolume(volume *mirror.Volume, err error) error {
	return errors.Wrap(
		ErrUnusableMirrorVolume,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"mirror_volume": volume,
			"cause":         fmt.Sprintf("%+v", err),
		}),
	)
}

// NotFoundNFSExportPath NFS 의 export path 를 찾을 수 없음
func NotFoundNFSExportPath(volumeUUID string) error {
	return errors.Wrap(
		ErrNotFoundNFSExportPath,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"volume_uuid": volumeUUID,
		}),
	)
}

// NotFoundVolumeType 볼륨 타입을 찾을 수 없음
func NotFoundVolumeType(volumeName, volumeTypeName string) error {
	return errors.Wrap(
		ErrNotFoundVolumeType,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"volume_name":      volumeName,
			"volume_type_name": volumeTypeName,
		}),
	)
}

// NotFoundVolumeBackendName 볼륨 백엔드 네임을 찾을 수 없음
func NotFoundVolumeBackendName(volumeTypeName string) error {
	return errors.Wrap(
		ErrNotFoundVolumeBackendName,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"volume_type_name": volumeTypeName,
		}),
	)
}

// NotFoundAvailableServiceHostName 사용 가능한 서비스 호스트 네임을 찾을 수 없음
func NotFoundAvailableServiceHostName(backendName string) error {
	return errors.Wrap(
		ErrNotFoundAvailableServiceHostName,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"backend_name": backendName,
		}),
	)
}

// NotFoundInstanceHypervisor 인스턴스의 하이퍼바이저를 찾을 수 없음
func NotFoundInstanceHypervisor(instanceName string) error {
	return errors.Wrap(
		ErrNotFoundInstanceHypervisor,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"instance_name": instanceName,
		}),
	)
}

// NotFoundInstanceSpec 인스턴스의 Spec 을 찾을 수 없음
func NotFoundInstanceSpec(instanceName, flavorName string) error {
	return errors.Wrap(
		ErrNotFoundInstanceSpec,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"instance_name": instanceName,
			"flavor_name":   flavorName,
		}),
	)
}

// UnknownInstanceState 알 수 없는 인스턴스 상태
func UnknownInstanceState(instanceName string, state int64) error {
	return errors.Wrap(
		ErrUnknownInstanceState,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"instance_name": instanceName,
			"state":         state,
		}),
	)
}

// ImportVolumeRollbackFailed import volume 의 rollback 실패
func ImportVolumeRollbackFailed(volumePair UUIDPair, snapshotPairList ...UUIDPair) error {
	return errors.Wrap(
		ErrImportVolumeRollbackFailed,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"volume_pair":        volumePair,
			"snapshot_pair_list": snapshotPairList,
		}),
	)
}
