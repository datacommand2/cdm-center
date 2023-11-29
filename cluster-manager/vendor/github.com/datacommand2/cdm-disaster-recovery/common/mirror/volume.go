package mirror

import (
	"github.com/datacommand2/cdm-center/cluster-manager/storage"
	"github.com/datacommand2/cdm-center/cluster-manager/volume"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/store"

	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strconv"
)

const (
	volumeKeyBase         = "mirror_volume"
	volumeKeyRegexp       = "^mirror_volume/storage.\\d+.\\d+/volume.\\d+$"
	volumeKeyFormat       = "mirror_volume/storage.%d.%d/volume.%d"
	volumeKeyPrefixFormat = "mirror_volume/storage.%d.%d/volume.%d/"
	sourceVolumeKeyFormat = "mirror_volume/storage.%d/volume.%d"
)

// Agent -
type Agent struct {
	IP             string `json:"ip,omitempty"`
	Port           uint   `json:"port,omitempty"`
	Version        string `json:"version,omitempty"`
	InstalledAt    int64  `json:"installed_at,omitempty"`
	LastUpgradedAt int64  `json:"last_upgraded_at,omitempty"`
}

// Volume 볼륨 복제 정보에 대한 구조체
type Volume struct {
	SourceClusterStorage *storage.ClusterStorage `json:"source_cluster_storage,omitempty"`
	TargetClusterStorage *storage.ClusterStorage `json:"target_cluster_storage,omitempty"`
	SourceVolume         *volume.ClusterVolume   `json:"source_volume,omitempty"`
	SourceAgent          *Agent                  `json:"source_agent,omitempty"`
}

// VolumeOperation 볼륨 복제 오퍼레이션에 대한 구조체
type VolumeOperation struct {
	Operation string `json:"operation,omitempty"`
}

// VolumeStatus 볼륨 복제 상태에 대한 구조체
type VolumeStatus struct {
	StateCode   string   `json:"state_code,omitempty"`
	ErrorReason *Message `json:"error_reason,omitempty"`
}

// GetVolumeList 볼륨 복제 목록 조회 함수
func GetVolumeList() ([]*Volume, error) {
	keys, err := store.List(volumeKeyBase)
	if err != nil && err != store.ErrNotFoundKey {
		return nil, errors.UnusableStore(err)
	}

	var volumes []*Volume
	for _, k := range keys {
		var matched bool
		if matched, err = regexp.Match(volumeKeyRegexp, []byte(k)); err != nil {
			return nil, errors.Unknown(err)
		}

		if !matched {
			continue
		}

		var ssid, stid, vsid uint64
		if _, err = fmt.Sscanf(k, volumeKeyFormat, &ssid, &stid, &vsid); err != nil {
			return nil, errors.Unknown(err)
		}

		var v *Volume
		if v, err = GetVolume(ssid, stid, vsid); err != nil {
			return nil, err
		}
		volumes = append(volumes, v)
	}

	return volumes, nil
}

// GetVolume 볼륨 복제 단일 조회 함수
func GetVolume(sourceStorageID, targetStorageID, sourceVolumeID uint64) (*Volume, error) {
	s, err := store.Get(fmt.Sprintf(volumeKeyFormat, sourceStorageID, targetStorageID, sourceVolumeID))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundMirrorVolume(sourceStorageID, targetStorageID, sourceVolumeID)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var v Volume
	if err = json.Unmarshal([]byte(s), &v); err != nil {
		return nil, errors.Unknown(err)
	}

	return &v, nil
}

// PutVolume 볼륨 복제 정보를 kv store 에 저장 하는 함수
func PutVolume(v *Volume) error {
	return v.Put()
}

// Put 볼륨 복제 정보를 kv store 에 저장 하는 함수
func (v *Volume) Put() error {
	b, err := json.Marshal(v)
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(v.GetKey(), string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetMirrorEnvironment 볼륨 복제의 복제 환경 조회 함수
func (v *Volume) GetMirrorEnvironment() (*Environment, error) {
	return GetEnvironment(v.SourceClusterStorage.StorageID, v.TargetClusterStorage.StorageID)
}

// GetSourceStorage 볼륨 복제의 source 스토리지 정보 조회 함수
func (v *Volume) GetSourceStorage() (*storage.ClusterStorage, error) {
	return storage.GetStorage(v.SourceClusterStorage.ClusterID, v.SourceClusterStorage.StorageID)
}

// GetTargetStorage 볼륨 복제의 target 스토리지 정보 조회 함수
func (v *Volume) GetTargetStorage() (*storage.ClusterStorage, error) {
	return storage.GetStorage(v.TargetClusterStorage.ClusterID, v.TargetClusterStorage.StorageID)
}

// GetOperation 볼륨 복제에 오퍼 레이션 조회 함수
func (v *Volume) GetOperation() (*VolumeOperation, error) {
	s, err := store.Get(path.Join(v.GetKey(), "operation"))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundMirrorVolumeOperation(v.SourceClusterStorage.StorageID, v.TargetClusterStorage.StorageID, v.SourceVolume.VolumeID)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var op VolumeOperation
	if err = json.Unmarshal([]byte(s), &op); err != nil {
		return nil, errors.Unknown(err)
	}

	return &op, nil
}

// SetOperation 볼륨 에 오퍼 레이션 설정 함수
func (v *Volume) SetOperation(op string) error {
	if !isMirrorVolumeOperation(op) {
		return UnknownMirrorVolumeOperation(op)
	}

	b, err := json.Marshal(&VolumeOperation{Operation: op})
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(path.Join(v.GetKey(), "operation"), string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetStatus 볼륨 에 복제 상태 조회 함수
func (v *Volume) GetStatus() (*VolumeStatus, error) {
	s, err := store.Get(path.Join(v.GetKey(), "status"))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundMirrorVolumeStatus(v.SourceClusterStorage.StorageID, v.TargetClusterStorage.StorageID, v.SourceVolume.VolumeID)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var status VolumeStatus
	if err = json.Unmarshal([]byte(s), &status); err != nil {
		return nil, errors.Unknown(err)
	}

	return &status, nil
}

// SetStatus 볼륨 에 복제 상태 설정 함수
func (v *Volume) SetStatus(state string, code string, contents interface{}) error {
	if !isMirrorVolumeStateCode(state) {
		return UnknownMirrorVolumeState(state)
	}

	var c string
	if contents != nil {
		b, err := json.Marshal(contents)
		if err != nil {
			return errors.Unknown(err)
		}
		c = string(b)
	}

	b, err := json.Marshal(&VolumeStatus{StateCode: state, ErrorReason: &Message{Code: code, Contents: c}})
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(path.Join(v.GetKey(), "status"), string(b)); err != nil {
		return errors.UnusableStore(err)
	}
	return nil
}

// GetTargetMetadata target 볼륨의 메타 데이터 조회 함수
func (v *Volume) GetTargetMetadata() (map[string]interface{}, error) {
	s, err := store.Get(path.Join(v.GetKey(), "target", "metadata"))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundMirrorVolumeTargetMetadata(v.SourceClusterStorage.StorageID, v.TargetClusterStorage.StorageID, v.SourceVolume.VolumeID)
	case err != nil:
		return nil, errors.Unknown(err)
	}

	var m = make(map[string]interface{})
	if err = json.Unmarshal([]byte(s), &m); err != nil {
		return nil, errors.Unknown(err)
	}

	return m, nil
}

// SetTargetMetadata target 볼륨의 메타 데이터 설정 함수
func (v *Volume) SetTargetMetadata(md map[string]interface{}) error {
	b, err := json.Marshal(md)
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(path.Join(v.GetKey(), "target", "metadata"), string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetTargetAgent 볼륨 복제에 설정 된 target agent 를 조회 하는 함수
func (v *Volume) GetTargetAgent() (*Agent, error) {
	s, err := store.Get(path.Join(v.GetKey(), "target", "agent"))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundMirrorVolumeTargetAgent(v.SourceClusterStorage.StorageID, v.TargetClusterStorage.StorageID, v.SourceVolume.VolumeID)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var r Agent
	if err = json.Unmarshal([]byte(s), &r); err != nil {
		return nil, errors.Unknown(err)
	}

	return &r, nil
}

// SetTargetAgent 볼륨 복제를 위해 target agent 를 설정 하는 함수
func (v *Volume) SetTargetAgent(agent *Agent) error {
	b, err := json.Marshal(agent)
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(path.Join(v.GetKey(), "target", "agent"), string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetSourceVolumeFileList source 볼륨의 동기화 파일 목록을 조회 하는 함수
func (v *Volume) GetSourceVolumeFileList() ([]string, error) {
	s, err := store.Get(path.Join(v.GetKey(), "source", "files"))
	switch {
	case err != nil && err == store.ErrNotFoundKey:
		return []string{}, nil
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var snapshots []string
	if err = json.Unmarshal([]byte(s), &snapshots); err != nil {
		return nil, errors.Unknown(err)
	}

	return snapshots, nil
}

// SetSourceVolumeFileList source 볼륨의 동기화 파일 목록을 추가 하는 함수
func (v *Volume) SetSourceVolumeFileList(snapshots []string) error {
	b, err := json.Marshal(&snapshots)
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(path.Join(v.GetKey(), "source", "files"), string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetTargetVolumeFileList target 으로 복제 중인 볼륨의 복제 파일 목록을 조회 하는 함수
func (v *Volume) GetTargetVolumeFileList() ([]string, error) {
	s, err := store.Get(path.Join(v.GetKey(), "target", "files"))
	switch {
	case err != nil && err == store.ErrNotFoundKey:
		return []string{}, nil
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var snapshots []string
	if err = json.Unmarshal([]byte(s), &snapshots); err != nil {
		return nil, errors.Unknown(err)
	}

	return snapshots, nil
}

// SetTargetVolumeFileList target 으로 복제 중인 볼륨의 복제 파일을 추가 하는 함수
func (v *Volume) SetTargetVolumeFileList(snapshots []string) error {
	b, err := json.Marshal(&snapshots)
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(path.Join(v.GetKey(), "target", "files"), string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// Delete 볼륨 복제와 관련 된 모든 정보를 kv store 에서 삭제 하는 함수
func (v *Volume) Delete() error {
	// mirror_volume/storage.%d.%d/volume.%d/ 형태의 key prefix 옵션으로 삭제
	if err := store.Delete(v.GetKeyPrefix(), store.DeletePrefix()); err != nil {
		return errors.UnusableStore(err)
	}

	// mirror_volume/storage.%d.%d/volume.%d 해당 key 삭제
	if err := store.Delete(v.GetKey()); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetKey 볼륨 키 반환 함수
func (v *Volume) GetKey() string {
	return fmt.Sprintf(volumeKeyFormat, v.SourceClusterStorage.StorageID, v.TargetClusterStorage.StorageID, v.SourceVolume.VolumeID)
}

// GetKeyPrefix 볼륨 키 의 prefix 옵션을 사용할 수 있는 키 반환 함수
func (v *Volume) GetKeyPrefix() string {
	return fmt.Sprintf(volumeKeyPrefixFormat, v.SourceClusterStorage.StorageID, v.TargetClusterStorage.StorageID, v.SourceVolume.VolumeID)
}

// IncreaseRefCount source volume 을 mirroring 걸고 있는 target volume 의 수를 증가시키는 함수
// source volume reference 가 존재하지 않는 경우 etcd 에 추가한다.
func (v *Volume) IncreaseRefCount(txn store.Txn) error {
	count, err := v.GetRefCount()
	switch {
	case errors.Equal(err, ErrNotFoundSourceVolumeReference):
		txn.Put(path.Join(v.GetSourceVolumeKey(), "reference"), strconv.FormatUint(1, 10))
		return nil

	case err != nil:
		return errors.UnusableStore(err)
	}

	count++

	txn.Put(path.Join(v.GetSourceVolumeKey(), "reference"), strconv.FormatUint(count, 10))

	return nil
}

// DecreaseRefCount source volume 을 mirroring 걸고 있는 target volume 의 수를 감소시키는 함수
// source volume reference count 가 0 인 경우 etcd 에서 제거한다.
func (v *Volume) DecreaseRefCount(txn store.Txn) error {
	count, err := v.GetRefCount()
	switch {
	// count 가 없는 것으로 간주
	case errors.Equal(err, ErrNotFoundSourceVolumeReference):
		return nil

	case err != nil:
		return errors.UnusableStore(err)
	}

	count--

	if count == 0 {
		txn.Delete(path.Join(v.GetSourceVolumeKey(), "reference"), store.DeletePrefix())
	} else {
		txn.Put(path.Join(v.GetSourceVolumeKey(), "reference"), strconv.FormatUint(count, 10))
	}

	return nil
}

// GetRefCount source volume 을 mirroring 걸고 있는 target volume 의 수를 조회하는 함수
func (v *Volume) GetRefCount() (uint64, error) {
	s, err := store.Get(path.Join(v.GetSourceVolumeKey(), "reference"))
	switch {
	case errors.Equal(err, store.ErrNotFoundKey):
		return 0, NotFoundSourceVolumeReference(v.SourceClusterStorage.StorageID, v.SourceVolume.VolumeID)

	case err != nil:
		return 0, errors.UnusableStore(err)
	}

	count, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.Unknown(err)
	}

	return count, nil
}

// GetPlanOfMirrorVolume mirror volume 의 plan id 값을 가져온다.
func (v *Volume) GetPlanOfMirrorVolume() (uint64, error) {
	s, err := store.Get(path.Join(v.GetKey(), "plan"))
	switch {
	case err != nil && err == store.ErrNotFoundKey:
		return 0, nil
	case err != nil:
		return 0, errors.UnusableStore(err)
	}

	pid, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.Unknown(err)
	}

	return pid, nil
}

// SetPlanOfMirrorVolume mirror volume 의 plan 의 id 값을 etcd 에 저장한다.
func (v *Volume) SetPlanOfMirrorVolume(pid uint64) error {
	if err := store.Put(path.Join(v.GetKey(), "plan"), strconv.FormatUint(pid, 10)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetSourceVolumeKey source volume 키 반환 함수
func (v *Volume) GetSourceVolumeKey() string {
	return fmt.Sprintf(sourceVolumeKeyFormat, v.SourceClusterStorage.StorageID, v.SourceVolume.VolumeID)
}
