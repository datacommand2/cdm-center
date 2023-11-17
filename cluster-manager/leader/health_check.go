package leader

import (
	"github.com/datacommand2/cdm-center/cluster-manager/config"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/internal"
	"github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"

	"context"
	"time"
)

func updateClusterState(cluster *model.Cluster, stateCode string) error {
	unlock, err := internal.ClusterDBLock(cluster.ID)
	if err != nil {
		return err
	}

	defer unlock()

	if err = database.Transaction(func(db *gorm.DB) error {
		// 클러스터 유무 확인
		err = db.First(&model.Cluster{}, &model.Cluster{ID: cluster.ID}).Error
		switch {
		case err == gorm.ErrRecordNotFound:
			return internal.NotFoundCluster(cluster.ID)
		case err != nil:
			return errors.UnusableDatabase(err)
		}

		cluster.StateCode = stateCode
		if err = db.Save(cluster).Error; err != nil {
			return errors.UnusableDatabase(err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (l *Leader) check(c *model.Cluster) {
	logger.Infof("[health_check] Start checking the status of nova, cinder, neutron in the cluster(%d:%s).", c.ID, c.Name)
	state := l.CheckCluster(c.ID)

	switch state {
	case constant.ClusterStateInactive:
		if err := updateClusterState(c, constant.ClusterStateInactive); err != nil {
			logger.Errorf("[health_check] Could not update state(inactive) of cluster(%d:%s). Cause: %+v", c.ID, c.Name, err)
		} else {
			logger.Infof("[health_check] Cluster(%d:%s) state update to inactive.", c.ID, c.Name)
		}

	case constant.ClusterStateWarning:
		if err := updateClusterState(c, constant.ClusterStateWarning); err != nil {
			logger.Errorf("[health_check] Could not update state(warning) of cluster(%d:%s). Cause: %+v", c.ID, c.Name, err)
			return
		}
		logger.Infof("[health_check] Cluster(%d:%s) state update to warning.", c.ID, c.Name)

	case constant.ClusterStateActive:
		if err := updateClusterState(c, constant.ClusterStateActive); err != nil {
			logger.Errorf("[health_check] Could not update state(active) of cluster(%d:%s). Cause: %+v", c.ID, c.Name, err)
		} else {
			logger.Infof("[health_check] Cluster(%d:%s) state update to active.", c.ID, c.Name)
		}
	}
}

// healthCheckClusterInit 초기 실행시 cluster 상태 조회 및 주기 설정값 메모리 데이터에 추가
func (l *Leader) healthCheckClusterInit() {
	var err error
	l.c, _ = getClusterList()
	if l.c == nil {
		return
	}

	for _, c := range l.c {
		l.wg.Add(1)
		go func(c *model.Cluster) {
			defer l.wg.Done()

			// etcd 에 저장했던 제외 서비스 초기설정
			if l.excluded[c.ID], err = cluster.GetException(c.ID); err != nil && !errors.Equal(err, internal.ErrNotFoundClusterException) {
				logger.Errorf("[healthCheckClusterInit] Could not get exception: clusters/%d/check/exception. Cause: %+v", c.ID, err)
			}

			l.check(c)

		}(c)
	}

	l.wg.Wait()
	logger.Infof("[healthCheckClusterInit] Initial cluster health check completed")
}

func (l *Leader) leaderInit() {
	var err error
	if l.c == nil {
		return
	}

	for _, c := range l.c {
		if l.cfg[c.ID], err = config.GetConfig(c.ID); errors.Equal(err, config.ErrNotFoundConfig) {
			// 처음 Cluster 를 생성할 때 etcd 에 config 값이 기본값으로 저장
			l.cfg[c.ID] = &config.Config{
				ClusterID:            c.ID,
				TimestampInterval:    5,
				ReservedSyncInterval: 1,
			}

			if err = l.cfg[c.ID].Put(); err != nil {
				logger.Warnf("[leaderInit] Unable to put the config. Cause : ", err)
			}

		} else if err != nil {
			logger.Errorf("[leaderInit] Could not get config: centers/%d/config. Cause: %+v", c.ID, err)
		}

		// 메모리에 클러스터 저장 및 초기설정
		l.clusters[c.ID] = c
		l.isRunningC[c.ID] = false
	}
}

// healthCheckClusterList 는 클러스터 목록의 상태를 확인하는 함수이다.
func (l *Leader) healthCheckClusterList(ctx context.Context) {
	for {
		if l.clusters == nil {
			time.Sleep(time.Minute * 1)
			continue
		}

		for _, c := range l.clusters {
			// isRunningC 데이터에 특정 클러스터 ID로 조회시 해당 클러스터로 CheckCluster 작동중인지 아닌지 확인하고 작동중이지 않다면
			// config 에 저장된 interval 값으로 특정 시간 뒤에 goroutine 이 시작 하도록 하고, config 값이 만약에 없다면 기본값을 다시 저장한다.
			if v, ok := l.isRunningC[c.ID]; ok && !v {
				if d, ok := l.cfg[c.ID]; ok {
					go l.checkCluster(c, time.NewTicker(time.Minute*time.Duration(d.TimestampInterval)))
				} else {
					l.Lock()
					l.cfg[c.ID] = &config.Config{ClusterID: c.ID, TimestampInterval: 5, ReservedSyncInterval: 1}
					l.Unlock()
				}
			}
		}

		select {
		case <-ctx.Done():
			l.done <- true
			return

		case <-time.After(time.Second * 1): // health check interval
		}
	}
}

// checkCluster 함수는 ticker 로 언제 시작할지 cluster 마다 설정 주기 값에 따라 시작 시간이 되면 채널로 값이 들어온다.
func (l *Leader) checkCluster(c *model.Cluster, ticker *time.Ticker) {
	// 특정 클러스터가 실행중
	l.Lock()
	l.isRunningC[c.ID] = true
	l.Unlock()

	// ticker 로 정한 시간이 되면 실행 및 대기 중에 done 채널로 값이 오면 끝
	select {
	case <-l.done:
		return
	case <-ticker.C:
	}

	l.check(c)

	ticker.Stop()
	l.Lock()
	l.isRunningC[c.ID] = false
	l.Unlock()
}
