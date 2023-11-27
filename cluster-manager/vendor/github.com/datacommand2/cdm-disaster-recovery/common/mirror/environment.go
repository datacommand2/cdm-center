package mirror

import (
	"encoding/json"
	"fmt"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/store"
	"path"
	"regexp"
)

const (
	mirrorEnvironmentKeyBase         = "mirror_environment"
	mirrorEnvironmentKeyRegexp       = "^mirror_environment/storage.\\d+.\\d+$"
	mirrorEnvironmentKeyFormat       = "mirror_environment/storage.%d.%d"
	mirrorEnvironmentKeyPrefixFormat = "mirror_environment/storage.%d.%d/"
)

// Message 에러 사유에 대한 메세지 코드 구조체
type Message struct {
	Code     string `json:"code,omitempty"`
	Contents string `json:"contents,omitempty"`
}

// Environment 복제 환경 구성/삭제 에서 사용 하며 오퍼 레이션, 상태 값, 에러 사유 메시지 정보로 구성된 구조체
type Environment struct {
	//SourceClusterStorage *storage.ClusterStorage `json:"source_cluster_storage,omitempty"`
	//TargetClusterStorage *storage.ClusterStorage `json:"target_cluster_storage,omitempty"`
}

// EnvironmentOperation 복제 환경의 operation 정보 구조체
type EnvironmentOperation struct {
	Operation string `json:"operation,omitempty"`
}

// EnvironmentStatus 복제 환경의 상태 정보 구조체
type EnvironmentStatus struct {
	StateCode   string   `json:"state_code,omitempty"`
	ErrorReason *Message `json:"error_reason,omitempty"`
}

// GetEnvironmentList 복제 환경 목록을 조회 하는 함수
func GetEnvironmentList() ([]*Environment, error) {
	keys, err := store.List(mirrorEnvironmentKeyBase)
	if err != nil && err != store.ErrNotFoundKey {
		return nil, errors.UnusableStore(err)
	}

	var envs []*Environment
	for _, k := range keys {
		var matched bool
		if matched, err = regexp.Match(mirrorEnvironmentKeyRegexp, []byte(k)); err != nil {
			return nil, errors.Unknown(err)
		}
		if !matched {
			continue
		}

		var sid, tid uint64
		if _, err = fmt.Sscanf(k, mirrorEnvironmentKeyFormat, &sid, &tid); err != nil {
			return nil, errors.Unknown(err)
		}

		var e *Environment
		if e, err = GetEnvironment(sid, tid); err != nil {
			return nil, err
		}

		envs = append(envs, e)
	}

	return envs, nil
}

// GetEnvironment 복제 환경 조회 하는 함수
func GetEnvironment(sid, tid uint64) (*Environment, error) {
	v, err := store.Get(fmt.Sprintf(mirrorEnvironmentKeyFormat, sid, tid))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundMirrorEnvironment(sid, tid)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var e Environment
	if err = json.Unmarshal([]byte(v), &e); err != nil {
		return nil, errors.Unknown(err)
	}

	return &e, nil
}

// PutEnvironment 스토어에 복제 환경 정보를 저장하는 함수
func PutEnvironment(e *Environment) error {
	return e.Put()
}

// Put 스토어에 복제 환경 정보를 저장하는 함수
func (e *Environment) Put() error {
	b, err := json.Marshal(e)
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(e.GetKey(), string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetKey Environment 의 키 반환 함수
func (e *Environment) GetKey() string {
	return fmt.Sprintf(mirrorEnvironmentKeyFormat, e.SourceClusterStorage.StorageID, e.TargetClusterStorage.StorageID)
}

// GetKeyPrefix Environment 의 prefix 옵션을 사용할 수 있는 키 반환 함수
func (e *Environment) GetKeyPrefix() string {
	return fmt.Sprintf(mirrorEnvironmentKeyPrefixFormat, e.SourceClusterStorage.StorageID, e.TargetClusterStorage.StorageID)
}

//// GetSourceStorage 복제 환경 으로 설정 된 소스 클러스터 스토리지를 조회하는 함수
//func (e *Environment) GetSourceStorage() (*storage.ClusterStorage, error) {
//	return storage.GetStorage(e.SourceClusterStorage.ClusterID, e.SourceClusterStorage.StorageID)
//}
//
//// GetTargetStorage 복제 환경 으로 설정 된 타겟 클러스터 스토리지를 조회하는 함수
//func (e *Environment) GetTargetStorage() (*storage.ClusterStorage, error) {
//	return storage.GetStorage(e.TargetClusterStorage.ClusterID, e.TargetClusterStorage.StorageID)
//}

// GetOperation 복제 환경의 operation 을 조회 하는 함수
func (e *Environment) GetOperation() (*EnvironmentOperation, error) {
	k := path.Join(fmt.Sprintf(mirrorEnvironmentKeyFormat, e.SourceClusterStorage.StorageID, e.TargetClusterStorage.StorageID), "operation")
	v, err := store.Get(k)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundMirrorEnvironmentOperation(e.SourceClusterStorage.StorageID, e.TargetClusterStorage.StorageID)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var operation EnvironmentOperation
	if err = json.Unmarshal([]byte(v), &operation); err != nil {
		return nil, errors.Unknown(err)
	}

	return &operation, nil
}

// SetOperation 복제 환경의 operation 을 설정 하는 함수
func (e *Environment) SetOperation(op string) error {
	if !isMirrorEnvironmentOperation(op) {
		return UnknownMirrorEnvironmentOperation(op)
	}

	b, err := json.Marshal(&EnvironmentOperation{Operation: op})
	if err != nil {
		return errors.Unknown(err)
	}

	k := path.Join(fmt.Sprintf(mirrorEnvironmentKeyFormat, e.SourceClusterStorage.StorageID, e.TargetClusterStorage.StorageID), "operation")
	if err = store.Put(k, string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetStatus 복제 환경의 상태를 조회 하는 함수
func (e *Environment) GetStatus() (*EnvironmentStatus, error) {
	k := path.Join(fmt.Sprintf(mirrorEnvironmentKeyFormat, e.SourceClusterStorage.StorageID, e.TargetClusterStorage.StorageID), "status")
	v, err := store.Get(k)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundMirrorEnvironmentStatus(e.SourceClusterStorage.StorageID, e.TargetClusterStorage.StorageID)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var state EnvironmentStatus
	if err = json.Unmarshal([]byte(v), &state); err != nil {
		return nil, errors.Unknown(err)
	}

	return &state, nil
}

// SetStatus 복제 환경의 상태를 설정 하는 함수
func (e *Environment) SetStatus(state string, code string, contents interface{}) error {
	if !isMirrorEnvironmentStateCode(state) {
		return UnknownMirrorEnvironmentState(state)
	}

	var c string
	if contents != nil {
		b, err := json.Marshal(contents)
		if err != nil {
			return errors.Unknown(err)
		}
		c = string(b)
	}

	b, err := json.Marshal(&EnvironmentStatus{StateCode: state, ErrorReason: &Message{Code: code, Contents: c}})
	if err != nil {
		return errors.Unknown(err)
	}

	k := path.Join(fmt.Sprintf(mirrorEnvironmentKeyFormat, e.SourceClusterStorage.StorageID, e.TargetClusterStorage.StorageID), "status")
	if err := store.Put(k, string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// Delete 복제 환경과 관련된 모든 정보를 kv store 에서 삭제 하는 함수
func (e *Environment) Delete() error {
	// mirror_environment/storage.%d.%d/ 형태의 key prefix 옵션으로 삭제
	if err := store.Delete(e.GetKeyPrefix(), store.DeletePrefix()); err != nil {
		return errors.UnusableStore(err)
	}
	// mirror_environment/storage.%d.%d 해당 key 삭제
	return store.Delete(e.GetKey())
}

// IsMirrorVolumeExisted 복제 환경의 볼륨 복제가 존재 하는지 확인 하는 함수
func (e *Environment) IsMirrorVolumeExisted() error {
	_, err := store.List(fmt.Sprintf(path.Join(volumeKeyBase, "storage.%d.%d"), e.SourceClusterStorage.StorageID, e.TargetClusterStorage.StorageID))
	switch {
	case err == store.ErrNotFoundKey:
		return nil
	case err != nil:
		return errors.UnusableStore(err)
	}

	return VolumeExisted()
}
