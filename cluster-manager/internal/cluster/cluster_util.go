package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/metadata"
	"github.com/datacommand2/cdm-cloud/common/store"
	"github.com/datacommand2/cdm-disaster-recovery/common/mirror"
	"github.com/jinzhu/gorm"
	"time"
)

// Message 클러스터 동기화 중 실패 상태를 저장히기위한 구조체
type Message struct {
	Code     string `json:"code"`
	Contents string `json:"contents"`
}

// SyncStatus 클러스터의 동기화 상태 정보를 저장하기위한 구조체
type SyncStatus struct {
	StatusCode string     `json:"status_code,omitempty"`
	Reasons    []*Message `json:"reasons,omitempty"`
	StartedAt  time.Time  `json:"created_at,omitempty"`
}

const (
	clusterBase            = "clusters/%d"
	clusterKeyPrefixFormat = "clusters/%d/"
	// clusterSyncStatusKeyFormat 클러스터 동기화 진행상태를 저장하기위한 키 값
	clusterSyncStatusKeyFormat = "clusters/%d/sync/status"
	clusterAgentKeyFormat      = "clusters/%d/agents/cinder"

	clusterExceptionFormat = "clusters/%d/check/exception"
)

// IsClusterOwner 사용자의 그룹이 클러스터의 조회권한을 가지고 있는지 확인
func IsClusterOwner(ctx context.Context, db *gorm.DB, id uint64) error {
	c := model.Cluster{}
	tid, _ := metadata.GetTenantID(ctx)
	user, _ := metadata.GetAuthenticatedUser(ctx)

	var cond *model.Cluster
	if internal.IsAdminUser(user) {
		cond = &model.Cluster{ID: id}
	} else {
		cond = &model.Cluster{ID: id, TenantID: tid}
	}

	// 클러스터 유무 확인
	err := db.Select("owner_group_id").
		Where(cond).
		First(&c).Error

	if err == gorm.ErrRecordNotFound {
		return internal.NotFoundCluster(id)
	} else if err != nil {
		return errors.UnusableDatabase(err)
	}

	// admin 의 경우 모든 클러스터를 조회할 수 있다.
	if !internal.IsAdminUser(user) && !internal.IsGroupUser(user, c.OwnerGroupID) {
		return errors.UnauthorizedRequest(ctx)
	}

	return nil
}

// GetConnectionInfo 클러스터 ID 로 클러스터 접속정보 조회.
// 테넌트 확인을 하지 않기 때문에 서비스간 요청에서만 사용해야 한다.
func GetConnectionInfo(db *gorm.DB, id uint64) (*cms.ClusterConnectionInfo, error) {
	var c model.Cluster

	err := db.First(&c, &model.Cluster{ID: id}).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, internal.NotFoundCluster(id)
	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &cms.ClusterConnectionInfo{
		TypeCode:     c.TypeCode,
		ApiServerUrl: c.APIServerURL,
		Credential:   c.Credential,
	}, nil
}

// SetStatus 클러스터 동기화 상태 저장
func SetStatus(cid uint64, state SyncStatus) error {
	b, err := json.Marshal(&state)
	if err != nil {
		return errors.Unknown(err)
	}

	k := fmt.Sprintf(clusterSyncStatusKeyFormat, cid)
	if err := store.Put(k, string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetStatus 클러스터 동기화 상태 조회
func GetStatus(cid uint64) (*SyncStatus, error) {
	k := fmt.Sprintf(clusterSyncStatusKeyFormat, cid)
	v, err := store.Get(k)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, internal.NotFoundClusterSyncStatus(cid)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var status SyncStatus
	if err = json.Unmarshal([]byte(v), &status); err != nil {
		return nil, errors.Unknown(err)
	}

	return &status, nil
}

// DeleteStatus 클러스터 동기화 상태 삭제
func DeleteStatus(cid uint64) error {
	k := fmt.Sprintf(clusterSyncStatusKeyFormat, cid)
	if err := store.Delete(k, store.DeletePrefix()); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// IsAbleToSync 동기화 가능 상태인지 확인
func IsAbleToSync(cid uint64) (bool, *SyncStatus) {
	unknown := &SyncStatus{StatusCode: constant.ClusterSyncStateCodeUnknown}
	status, err := GetStatus(cid)
	switch {
	case errors.Equal(err, internal.ErrNotFoundClusterSyncStatus):
		return true, unknown

	case err != nil:
		return false, unknown

	case status.StatusCode == constant.ClusterSyncStateCodeInit:
		syncTimeout := time.Minute * 1 // 1분
		// 시작후, 1분 > 지금: false - 동기화 불가능
		// 시작후, 1분 <= 지금: true - 동기화 가능
		if !time.Time.IsZero(status.StartedAt) && time.Since(status.StartedAt) >= syncTimeout {
			return false, status
		}
		fallthrough

	case status.StatusCode == constant.ClusterSyncStateCodeRunning:
		syncTimeout := time.Minute * 5 // 5분
		// 시작후, 5분 > 지금: false - 동기화 불가능
		// 시작후, 5분 <= 지금: true - 동기화 가능
		if !time.Time.IsZero(status.StartedAt) && time.Since(status.StartedAt) >= syncTimeout {
			return false, status
		}
		fallthrough

	default:
		// status.StatusCode == constant.ClusterSyncStateCodeDone, constant.ClusterSyncStateCodeFailed
		// status.StatusCode == constant.ClusterSyncStateCodeRunning && status.StartedAt + 30분 <= time.Now()
		return true, status
	}
}

// SetAgent cinder 의 agent 접속정보 저장
// TODO:
func SetAgent(cid uint64, agent mirror.Agent) error {
	b, _ := json.Marshal(&agent)
	key := fmt.Sprintf(clusterAgentKeyFormat, cid)
	if err := store.Put(key, string(b)); err != nil {
		return errors.UnusableStore(err)
	}
	return nil
}

// GetAgent cinder 의 agent 접속정보 조회
// TODO:
func GetAgent(cid uint64) (*mirror.Agent, error) {
	key := fmt.Sprintf(clusterAgentKeyFormat, cid)

	a, err := store.Get(key)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, internal.NotFoundClusterSyncStatus(cid)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	// TODO:
	var agent mirror.Agent
	if err = json.Unmarshal([]byte(a), &agent); err != nil {
		return nil, errors.Unknown(err)
	}

	return &agent, nil
}

// DeleteClusters 클러스터 정보 삭제
// 클러스터 삭제시 관련된 모든 정보를 지우기위해서만 사용
func DeleteClusters(cid uint64) error {
	// clusters/{clusterId} 로 시작하는 모든 정보 삭제
	if err := store.Delete(fmt.Sprintf(clusterKeyPrefixFormat, cid), store.DeletePrefix()); err != nil {
		return errors.UnusableStore(err)
	}

	if err := store.Delete(fmt.Sprintf(clusterBase, cid)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// SetException 클러스터 상태 제외 항목 저장
func SetException(cid uint64, excluded queue.ExcludedCluster) error {
	b, err := json.Marshal(&excluded)
	if err != nil {
		return errors.Unknown(err)
	}

	k := fmt.Sprintf(clusterExceptionFormat, cid)
	if err := store.Put(k, string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetException 클러스터 상태 제외 항목 조회
func GetException(cid uint64) (*queue.ExcludedCluster, error) {
	k := fmt.Sprintf(clusterExceptionFormat, cid)
	v, err := store.Get(k)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, internal.NotFoundClusterException(cid)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var status queue.ExcludedCluster
	if err = json.Unmarshal([]byte(v), &status); err != nil {
		return nil, errors.Unknown(err)
	}

	return &status, nil
}

// DeleteException 클러스터 상태 제외 항목 삭제
func DeleteException(cid uint64) error {
	k := fmt.Sprintf(clusterExceptionFormat, cid)
	if err := store.Delete(k, store.DeletePrefix()); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// MakeFailedMessage failed message 생성
func MakeFailedMessage(resource string) *Message {
	code := "cluster-manager." + resource + ".failed"
	contents := "contents not defined"
	return &Message{Code: code, Contents: contents}
}
