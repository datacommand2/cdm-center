package volume

import (
	"encoding/json"
	"fmt"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/store"
	"path"
)

const (
	clusterVolumeKeyFormat = "clusters/%d/volumes/%d"
)

// ClusterVolume 클러스터 볼륨 구조체
type ClusterVolume struct {
	ClusterID  uint64 `json:"cluster_id,omitempty"`
	VolumeID   uint64 `json:"volume_id,omitempty"`
	VolumeUUID string `json:"volume_uuid,omitempty"`
}

// GetVolume 클러스터 볼륨을 kv store 에서 조회 하는 함수
func GetVolume(cid, vid uint64) (*ClusterVolume, error) {
	s, err := store.Get(fmt.Sprintf(clusterVolumeKeyFormat, cid, vid))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundVolume(cid, vid)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var v ClusterVolume
	if err = json.Unmarshal([]byte(s), &v); err != nil {
		return nil, errors.Unknown(err)
	}

	return &v, nil
}

// PutVolume 클러스터 볼륨을 kv store 에 저장 하는 함수
func PutVolume(v *ClusterVolume) error {
	return v.Put()
}

// Put 클러스터 볼륨을 kv store 에 저장 하는 함수
func (v *ClusterVolume) Put() error {
	b, err := json.Marshal(v)
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(v.GetKey(), string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetKey 클러스터 볼륨의 키 반환 함수
func (v *ClusterVolume) GetKey() string {
	return fmt.Sprintf(clusterVolumeKeyFormat, v.ClusterID, v.VolumeID)
}

// GetMetadata 클러스터 볼륨의 메타 데이터 조회 함수
func (v *ClusterVolume) GetMetadata() (map[string]interface{}, error) {
	s, err := store.Get(path.Join(v.GetKey(), "metadata"))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundVolumeMetadata(v.ClusterID, v.VolumeID)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var i = make(map[string]interface{})
	if err = json.Unmarshal([]byte(s), &i); err != nil {
		return nil, errors.Unknown(err)
	}

	return i, nil
}

// SetMetadata 클러스터 볼륨의 메타 데이터 설정 함수
func (v *ClusterVolume) SetMetadata(md map[string]interface{}) error {
	b, err := json.Marshal(md)
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(path.Join(v.GetKey(), "metadata"), string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// Delete 클러스터 볼륨과 관련 된 모든 정보를 삭제 하는 함수
func (v *ClusterVolume) Delete() error {
	if err := store.Delete(v.GetKey(), store.DeletePrefix()); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// MergeMetadata 클러스터 볼륨의 metadata 를 병합하는 함수
func (v *ClusterVolume) MergeMetadata(md map[string]interface{}) error {
	var metadata map[string]interface{}

	metadata, err := v.GetMetadata()
	if err != nil && !errors.Equal(err, ErrNotFoundVolumeMetadata) {
		return err
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	for k, v := range md {
		metadata[k] = v
	}

	return v.SetMetadata(metadata)
}
