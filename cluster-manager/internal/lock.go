package internal

import (
	"context"
	"fmt"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/sync"
)

const (
	clusterBrokerLockPathFormat              = "cluster.broker.locks/%d"
	clusterDBLockPathFormat                  = "cluster.db.locks/%d"
	clusterVolumeGroupSnapshotLockPathFormat = "cluster.volume.group.locks/%s"
)

// ClusterDBLock 은 cluster 의 데이터 베이스 lock
func ClusterDBLock(id uint64) (func(), error) {
	p := fmt.Sprintf(clusterDBLockPathFormat, id)
	lock, err := sync.Lock(context.Background(), p)
	if err != nil {
		_ = lock.Close()
		return nil, errors.Unknown(err)
	}

	return func() {
		if err = lock.Unlock(context.Background()); err != nil {
			logger.Warnf("[ClusterDBLock] Could not unlock (%s). Cause: %+v", p, err)
		}
		if err = lock.Close(); err != nil {
			logger.Warnf("[ClusterBrokerLock] Could not close the session (%s). Cause: %+v", p, err)
		}
	}, nil
}

// ClusterBrokerLock 은 cluster 의 브로커 lock
func ClusterBrokerLock(id uint64) (func(), error) {
	p := fmt.Sprintf(clusterBrokerLockPathFormat, id)
	lock, err := sync.Lock(context.Background(), p)
	if err != nil {
		_ = lock.Close()
		return nil, errors.Unknown(err)
	}

	return func() {
		if err = lock.Unlock(context.Background()); err != nil {
			logger.Warnf("[ClusterBrokerLock] Could not unlock (%s). Cause: %+v", p, err)
		}
		if err = lock.Close(); err != nil {
			logger.Warnf("[ClusterBrokerLock] Could not close the session (%s). Cause: %+v", p, err)
		}
	}, nil
}

// ClusterVolumeGroupSnapshotLock 볼륨 그룹 스냅샷 lock
func ClusterVolumeGroupSnapshotLock(vgUUID string) (func(), error) {
	p := fmt.Sprintf(clusterVolumeGroupSnapshotLockPathFormat, vgUUID)
	lock, err := sync.Lock(context.Background(), p)
	if err != nil {
		_ = lock.Close()
		return nil, errors.Unknown(err)
	}

	return func() {
		if err = lock.Unlock(context.Background()); err != nil {
			logger.Warnf("[ClusterVolumeGroupSnapshotLock] Could not unlock (%s). Cause: %+v", p, err)
		}
		if err = lock.Close(); err != nil {
			logger.Warnf("[ClusterBrokerLock] Could not close the session (%s). Cause: %+v", p, err)
		}
	}, nil
}
