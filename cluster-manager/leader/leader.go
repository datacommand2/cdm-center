package leader

import (
	"github.com/datacommand2/cdm-center/cluster-manager/config"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/sync"

	"context"
	"fmt"
	mu "sync"
)

// Leader 리더로 선출된 컨테이너의 메모리 데이터
type Leader struct {
	opts *Options

	mu.Mutex
	wg mu.WaitGroup

	// c 초기 실행시 클러스터가 있을때 저장하고 healthCheckClusterInit(), leaderInit() 함수에서 사용 후 끝
	c []*model.Cluster

	excluded map[uint64]*queue.ExcludedCluster
	clusters map[uint64]*model.Cluster
	cfg      map[uint64]*config.Config

	isRunningC map[uint64]bool
	isRunningR bool

	done chan bool
}

// NewLeader new Leader
func NewLeader(opts ...Option) Leader {
	options := newOptions(opts...)

	return Leader{
		opts:       options,
		isRunningR: false,
		isRunningC: make(map[uint64]bool),
		excluded:   make(map[uint64]*queue.ExcludedCluster),
		clusters:   make(map[uint64]*model.Cluster),
		cfg:        make(map[uint64]*config.Config),
		done:       make(chan bool),
	}
}

// Init initialize leader
func Init() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	l := NewLeader()

	l.healthCheckClusterInit() // cluster-manager 를 처음 시작할 때 클러스터가 있다면 초기 체크

	go func() {
		cl, err := sync.CampaignLeader(ctx, fmt.Sprintf("%s/clusters", constant.ServiceName))
		if err != nil {
			logger.Fatalf("[Leader-Init] Could not campaign leader. Cause: %+v", err)
		}
		defer func() {
			_ = cl.Resign(context.Background())
			_ = cl.Close()
		}()

		logger.Info("[Leader-Init] Initializing leader.")

		l.leaderInit()
		go l.healthCheckClusterList(ctx)
		//go startClustersEventMonitor(ctx)
		go l.startClusterNoticeSubscriber(ctx)
		//--go startPeriodicSynchronization(ctx)

		logger.Info("[Leader-Init] Leader initialization complete.")

		// wait until context is done
		// wait leader to changes
		select {
		case <-cl.Status():
			cancel()
			logger.Fatal("[Leader-Init] Leader has changed")

		case <-ctx.Done():
		}

	}()

	return cancel
}
