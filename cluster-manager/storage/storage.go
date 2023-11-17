package storage

import (
	"encoding/json"
	"fmt"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/store"
	"path"
)

const (
	clusterStorageBase      = "clusters"
	clusterStorageKeyFormat = "clusters/%d/storages/%d"
)

// ClusterStorage 클러스터 스토리지 구조체
type ClusterStorage struct {
	ClusterID   uint64 `json:"cluster_id,omitempty"`
	StorageID   uint64 `json:"storage_id,omitempty"`
	Type        string `json:"type,omitempty"`
	BackendName string `json:"backend_name,omitempty"`
}

// GetStorage 클러스터 ID 와 storage ID 에 일치 하는 클러스터 스토리지를 조회 하는 함수
func GetStorage(cid, sid uint64) (*ClusterStorage, error) {
	s, err := store.Get(fmt.Sprintf(clusterStorageKeyFormat, cid, sid))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundStorage(cid, sid)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var storage ClusterStorage
	if err = json.Unmarshal([]byte(s), &storage); err != nil {
		return nil, errors.Unknown(err)
	}

	return &storage, nil
}

// PutStorage 클러스터 스토리지를 kv store 에 저장 하는 함수
func PutStorage(c *ClusterStorage) error {
	return c.Put()
}

// Put 클러스터 스토리지를 kv store 에 저장 하는 함수
func (c *ClusterStorage) Put() error {
	if !IsClusterStorageTypeCode(c.Type) {
		return UnsupportedStorageType(c.Type)
	}

	b, err := json.Marshal(c)
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(c.GetKey(), string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// Delete 클러스터 스토리지와 관련된 모든 정보를 삭제 하는 함수
func (c *ClusterStorage) Delete() error {
	if err := store.Delete(c.GetKey(), store.DeletePrefix()); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetKey ClusterStorage 의 키 반환 함수
func (c *ClusterStorage) GetKey() string {
	return fmt.Sprintf(clusterStorageKeyFormat, c.ClusterID, c.StorageID)
}

/*
 * NFS Metadata
 * {
 * 	"exports": [""]
 * }
 *
 * Ceph Metadata
 * {
 * 	"pool": "",
 * 	"conf": "",
 * 	"client": "",
 * 	"keyring": "",
 * 	"admin_client": "",
 * 	"admin_keyring": ""
 * }
 */

// GetMetadata 클러스터 스토리지의 metadata 를 조회 하는 함수
func (c *ClusterStorage) GetMetadata() (map[string]interface{}, error) {
	s, err := store.Get(path.Join(c.GetKey(), "metadata"))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundStorageMetadata(c.ClusterID, c.StorageID)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var i = make(map[string]interface{})
	if err = json.Unmarshal([]byte(s), &i); err != nil {
		return nil, errors.Unknown(err)
	}

	return i, nil
}

// SetMetadata 클러스터 스토리지의 metadata 를 설정 하는 함수
func (c *ClusterStorage) SetMetadata(md map[string]interface{}) error {
	b, err := json.Marshal(md)
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(path.Join(c.GetKey(), "metadata"), string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// MergeMetadata 클러스터 스토리지의 metadata 를 병합하는 함수
func (c *ClusterStorage) MergeMetadata(md map[string]interface{}) error {
	var metadata map[string]interface{}

	metadata, err := c.GetMetadata()
	if err != nil && !errors.Equal(err, ErrNotFoundStorageMetadata) {
		return err
	}

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	for k, v := range md {
		metadata[k] = v
	}

	return c.SetMetadata(metadata)
}
