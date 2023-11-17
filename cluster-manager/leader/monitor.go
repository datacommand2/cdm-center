package leader

import (
	"context"
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/config"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/broker"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/jinzhu/gorm"
	"sync"

	// Connect for openstack
	_ "github.com/datacommand2/cdm-center/cluster-manager/client/openstack/notification"
)

var clusterEventMonitorLock = sync.Mutex{}
var clusterEventMonitorMap map[uint64]client.Monitor

func (l *Leader) createClusterEventMonitor(p broker.Event) error {
	var c model.Cluster
	var clusterID uint64
	if err := json.Unmarshal(p.Message().Body, &clusterID); err != nil {
		logger.Errorf("[createClusterEventMonitor] Error occurred during unmarshal the message. Cause: %+v", err)
		return errors.Unknown(err)
	}

	if err := database.Transaction(func(db *gorm.DB) error {
		return db.First(&c, model.Cluster{ID: clusterID}).Error
	}); err != nil {
		logger.Errorf("[createClusterEventMonitor] Could not get cluster. Cause: %+v", err)
		return errors.UnusableDatabase(err)
	}

	logger.Infof("[createClusterEventMonitor] Starting cluster(%d:%s) event monitor.", c.ID, c.Name)

	l.Lock()
	// 메모리 데이터에 조회된 cluster 저장
	l.clusters[c.ID] = &c
	l.isRunningC[c.ID] = false
	l.cfg[c.ID] = &config.Config{
		ClusterID:            c.ID,
		TimestampInterval:    5,
		ReservedSyncInterval: 1,
	}
	l.Unlock()

	// 처음 생성된 클러스터를 check
	l.check(&c)

	logger.Infof("[createClusterEventMonitor] Cluster(%d:%s) event monitor started.", c.ID, c.Name)

	return nil
}

func (l *Leader) deleteClusterEventMonitor(p broker.Event) error {
	var clusterID uint64
	if err := json.Unmarshal(p.Message().Body, &clusterID); err != nil {
		logger.Errorf("[deleteClusterEventMonitor] Error occurred during unmarshal the message. Cause: %+v", err)
		return errors.Unknown(err)
	}

	l.Lock()
	defer l.Unlock()

	// 메모리 클러스터 리스트에서 특정 클러스터 삭제
	if _, ok := l.isRunningC[clusterID]; ok {
		delete(l.isRunningC, clusterID)
	}
	if _, ok := l.clusters[clusterID]; ok {
		delete(l.clusters, clusterID)
	}
	if _, ok := l.cfg[clusterID]; ok {
		delete(l.cfg, clusterID)
	}
	if _, ok := l.excluded[clusterID]; ok {
		delete(l.excluded, clusterID)
	}

	logger.Infof("[deleteClusterEventMonitor] Stopping cluster(%d) event monitor.", clusterID)

	stopClusterEventMonitor(&model.Cluster{ID: clusterID})

	logger.Infof("[deleteClusterEventMonitor] Cluster(%d) event monitor stopped.", clusterID)

	return nil
}

func (l *Leader) updateConfigEventMonitor(p broker.Event) (err error) {
	var msg queue.ClusterConfig
	if err = json.Unmarshal(p.Message().Body, &msg); err != nil {
		logger.Errorf("[updateConfigEventMonitor] Error occurred during unmarshal the message. Cause: %+v", err)
		return errors.Unknown(err)
	}

	l.Lock()
	defer l.Unlock()

	if msg.Config.ReservedSyncInterval != 0 {
		l.cfg[msg.ClusterID].ReservedSyncInterval = msg.Config.ReservedSyncInterval
	}
	if msg.Config.TimestampInterval != 0 {
		l.cfg[msg.ClusterID].TimestampInterval = msg.Config.TimestampInterval
	}

	logger.Infof("[updateConfigEventMonitor] Updated the config values for clusters(%d).", msg.ClusterID)

	return nil
}

// clusterExceptionMonitor 클러스터의 compute, cinder, network 상태중 제외된 데이터 저장
func (l *Leader) clusterExceptionMonitor(e broker.Event) error {
	var msg queue.ExcludedCluster
	if err := json.Unmarshal(e.Message().Body, &msg); err != nil {
		logger.Errorf("[clusterExceptionMonitor] Error occurred during unmarshal the message. Cause: %+v", err)
		return errors.Unknown(err)
	}

	if msg.ClusterID == 0 {
		return errors.InvalidParameterValue("cluster.id", msg.ClusterID, "not found cluster_id")
	}

	l.Lock()
	defer l.Unlock()

	if l.excluded[msg.ClusterID] != nil {
		// 메모리에 있는 데이터 삭제 후 모두 다시 저장. storage 같은 경우 id가 계속 변하고, 변한 데이터를
		// 찾아서 바꾸는 것보다 리소스 사용측에서 낫다고 판단.
		delete(l.excluded, msg.ClusterID)
	}

	// 메시지로 들어온 데이터 모두 저장
	l.excluded[msg.ClusterID] = &queue.ExcludedCluster{
		ClusterID: msg.ClusterID,
		Storage:   msg.Storage,
		Compute:   msg.Compute,
		Network:   msg.Network,
	}

	logger.Infof("[clusterExceptionMonitor] Updated the status exclusions for clusters(%d)", msg.ClusterID)

	return nil
}

func startClusterEventMonitor(c *model.Cluster) error {
	clusterEventMonitorLock.Lock()
	defer clusterEventMonitorLock.Unlock()

	var err error

	if _, ok := clusterEventMonitorMap[c.ID]; !ok {
		clusterEventMonitorMap[c.ID], err = client.NewMonitor(c.TypeCode, c.APIServerURL)
		if err != nil {
			return err
		}
	}

	return clusterEventMonitorMap[c.ID].Start(c)
}

func stopClusterEventMonitor(c *model.Cluster) {
	clusterEventMonitorLock.Lock()
	defer clusterEventMonitorLock.Unlock()

	if _, ok := clusterEventMonitorMap[c.ID]; !ok {
		return
	}

	clusterEventMonitorMap[c.ID].Stop()

	delete(clusterEventMonitorMap, c.ID)
}

func startClustersEventMonitor(ctx context.Context) {
	clusterEventMonitorLock.Lock()

	clusterEventMonitorMap = make(map[uint64]client.Monitor)

	clusterEventMonitorLock.Unlock()

	clusters, err := getClusterList()
	if err != nil {
		logger.Fatalf("[startClustersEventMonitor] Could not start clusters event monitor. Cause: %+v", err)
	}

	for _, c := range clusters {
		logger.Infof("[startClustersEventMonitor] Starting cluster(%d:%s) event monitor.", c.ID, c.Name)

		if err = startClusterEventMonitor(c); err != nil {
			logger.Warnf("[startClustersEventMonitor] Failed to start cluster(%d:%s) event monitor. Sync cluster manually. Cause: %+v.", c.ID, c.Name, err)
		}

		logger.Infof("[startClustersEventMonitor] Cluster(%d:%s) event monitor started.", c.ID, c.Name)
	}

	<-ctx.Done()

	for id := range clusterEventMonitorMap {
		logger.Infof("[startClustersEventMonitor] Stopping cluster(%d) event monitor.", id)

		stopClusterEventMonitor(&model.Cluster{ID: id})

		logger.Infof("[startClustersEventMonitor] Cluster(%d) event monitor stopped.", id)
	}
}

func (l *Leader) startClusterNoticeSubscriber(ctx context.Context) {
	eSub, err := broker.SubscribeTempQueue(constant.TopicNoticeClusterCreated, l.createClusterEventMonitor)
	if err != nil {
		logger.Fatalf("[startClusterNoticeSubscriber] Failed to subscribe cluster created notice. Cause: %+v", errors.UnusableBroker(err))
	}

	dSub, err := broker.SubscribeTempQueue(constant.TopicNoticeClusterDeleted, l.deleteClusterEventMonitor)
	if err != nil {
		logger.Fatalf("[startClusterNoticeSubscriber] Failed to subscribe cluster deleted notice. Cause: %+v", errors.UnusableBroker(err))
	}

	nSub, err := broker.SubscribeTempQueue(constant.TopicNoticeCenterConfigUpdated, l.updateConfigEventMonitor)
	if err != nil {
		logger.Fatalf("[startClusterNoticeSubscriber] Failed to subscribe config updated notice. Cause: %+v", errors.UnusableBroker(err))
	}

	exceptionSub, err := broker.SubscribeTempQueue(constant.QueueClusterServiceException, l.clusterExceptionMonitor)
	if err != nil {
		logger.Warnf("[startClusterNoticeSubscriber] Failed to subscribe cluster exception notice. Cause: ", err)
	}

	<-ctx.Done()

	_ = eSub.Unsubscribe()
	_ = dSub.Unsubscribe()
	_ = nSub.Unsubscribe()
	_ = exceptionSub.Unsubscribe()
}
