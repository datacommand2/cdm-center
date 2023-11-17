package config

import (
	"encoding/json"
	"fmt"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/store"
)

const clusterConfigFormat = "clusters/%d/config"

// Config 클러스터의 config 설정 정보를 저장하기위한 구조체
type Config struct {
	ClusterID            uint64 `json:"cluster_id,omitempty"`
	TimestampInterval    int64  `json:"timestamp_interval,omitempty"`
	ReservedSyncInterval int64  `json:"reserved_sync_interval,omitempty"`
}

// GetConfig 클러스터 config 정보 조회
func GetConfig(cid uint64) (*Config, error) {
	v, err := store.Get(fmt.Sprintf(clusterConfigFormat, cid))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundConfig(cid)

	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var conf Config
	if err = json.Unmarshal([]byte(v), &conf); err != nil {
		return nil, errors.Unknown(err)
	}

	return &conf, nil
}

// Put 클러스터 config 정보 저장
func (c *Config) Put() error {
	b, err := json.Marshal(c)
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(c.GetKey(), string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// DeleteConfig 클러스터 config 정보 삭제
func DeleteConfig(cid uint64) error {
	k := fmt.Sprintf(clusterConfigFormat, cid)
	if err := store.Delete(k); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetKey 클러스터 볼륨의 키 반환 함수
func (c *Config) GetKey() string {
	return fmt.Sprintf(clusterConfigFormat, c.ClusterID)
}
