package leader

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-center/cluster-manager/sync"
	"github.com/datacommand2/cdm-cloud/common/broker"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	mu "sync"

	"encoding/json"
	"time"
)

//func syncClusterList() error {
//	clusters, err := getClusterList()
//	if err != nil {
//		return err
//	}
//
//	for _, c := range clusters {
//		unlock, err := internal.ClusterDBLock(c.ID)
//		if err != nil {
//			logger.Warnf("[syncClusterList] Could not cluster lock (%d). Cause: %+v", c.ID, err)
//			continue
//		}
//
//		if err := sync.Cluster(c); err != nil {
//			logger.Warnf("[syncClusterList] Some errors occurred during synchronizing the cluster(%d:%s).", c.ID, c.Name)
//		}
//
//		unlock()
//	}
//
//	logger.Info("[syncClusterList] success to sync.")
//
//	return nil
//}
//
//func startPeriodicSynchronization(ctx context.Context) {
//	logger.Info("starting periodic synchronization")
//
//	for {
//		ticker := nextMidnightTicker()
//
//		select {
//		case <-ctx.Done():
//			ticker.Stop()
//			return
//
//		case <-ticker.C:
//			ticker.Stop()
//		}
//		logger.Infof("Run midnight sync start")
//		if err := syncClusterList(); err != nil {
//			logger.Warnf("Could not sync cluster list. Cause: %+v", err)
//		}
//	}
//}

// startReservedSynchronization 함수는 예약 된 시간에 동기화 실행
func (l *Leader) startReservedSynchronization(ticker *time.Ticker) {
	l.Lock()
	l.isRunningR = true
	l.Unlock()

	select {
	case <-l.done:
		return
	case <-ticker.C:
	}

	sync.ReservedSync()

	ticker.Stop()
	l.Lock()
	l.isRunningR = false
	l.Unlock()
}

/*
CheckCluster Cluster 의 nova, cinder, neutron 상태조회
각 서비스마다 API 를 통해 조회를 했을때 만약 API 자체 에러가 발생 시 Inactive 상태가 되고, 불러온 서비스 상태가 Down, disabled 상태이면 Warning 으로 된다.
만약 Inactive, Warning 상태가 나오면 각 서비스마다 최대 2번씩 재시도를 하게 되고 2번 재시도에도 불구하고 Warning 이나 Inactive 면 그대로 반환한다.
*/
func (l *Leader) CheckCluster(id uint64) string {
	var wg mu.WaitGroup
	s, err := sync.NewSynchronizer(id)
	if err != nil {
		logger.Errorf("[Leader-CheckCluster] Could not check cluster(%d). Cause: %+v", id, err)
		if err = publishToMessage(id, nil, nil, nil, false, constant.ClusterStateInactive, err.Error(), err.Error(), err.Error()); err != nil {
			logger.Warnf("[Leader-CheckCluster] Could not publish check cluster(%d) message. Cause: %+v", id, err)
		}

		return constant.ClusterStateInactive
	}

	// 옵션값이 true 라면 기존 데이터를 사용하고, false 이면 최신 데이터를 API 에서 불러와 사용한다. 기존 데이터는 옵션값을 통해 넘어온다.
	if l.opts.OpenStack {
		l.Lock()
		defer l.Unlock()

		logger.Infof("[Leader-CheckCluster] Starting to check the status of the cluster(%d) using existing data.", id)

		if l.opts.CheckCluster == nil {
			logger.Errorf("[Leader-CheckCluster] Execution failed because there is no value in CheckCluster")
			return ""
		}

		// OpenStack 옵션 값이 true 일떄 기존 값을 사용하는데 이때 option 에 PutCheckCluster 함수로 값을 넣어준 값을 사용한다.
		status := l.checkStatus()

		if err = publishToMessage(id, l.opts.CheckCluster.Storage, l.opts.CheckCluster.Compute, l.opts.CheckCluster.Network,
			false, status, l.opts.CheckCluster.StorageError, l.opts.CheckCluster.ComputeError, l.opts.CheckCluster.NetworkError); err != nil {
			logger.Warnf("[Leader-CheckCluster] Could not publish check cluster(%d) message. Cause: %+v", id, err)
		}

		logger.Infof("[Leader-CheckCluster] Cluster(%d) Nova, Cinder and Network status check has been completed.", id)
		return status
	}

	var (
		// 메시지에 보낼 데이터
		storages []queue.StorageServices
		networks []queue.Agent
		computes []queue.ComputeServices
		sErr     error
		cErr     error
		nErr     error

		computesMap = make(map[string]queue.ComputeServices)

		// 각 서비스마다의 상태
		storageStatus = constant.ClusterStateActive
		computeStatus = constant.ClusterStateActive
		networkStatus = constant.ClusterStateActive
	)

	wg.Add(3)
	go func() {
		logger.Debugf("[Leader-CheckCluster] Starting to check the Cinder volume status of the cluster(%d).", id)
		defer wg.Done()
		for i := 0; i < 2; i++ {
			// Cinder volume 상태 확인
			storages, _, sErr = sync.GetStorageServiceList(s.Cli)
			if sErr != nil {
				logger.Warnf("[Leader-CheckCluster] Error occurred during get storage data. Cause: %+v", sErr)
				l.Lock()
				storageStatus = constant.ClusterStateWarning
				l.Unlock()

				if errors.Equal(sErr, client.ErrRemoteServerError) {
					l.Lock()
					storageStatus = constant.ClusterStateInactive
					l.Unlock()
				}

				// cinder 데이터를 가져오지 못해 DB에 상태값을 Unknown 으로 변경
				if err = sync.SetUnknownClusterStoragesStatus(id); err != nil {
					// for 문이 총 2번 돌아가게 되는데 처음에만 continue 를 작동하고 두번째에는 그냥 지나가도록 한다.
					if i == 0 {
						logger.Infof("[Leader-CheckCluster] Retry service lookup in the Storage service. Cause: %+v", err)
						continue
					}
				}
				break
			}

			if err = s.UpdateClusterStoragesStatus(storages); err != nil {
				// for 문이 총 2번 돌아가게 되는데 처음에만 continue 를 작동하고 두번째에는 그냥 지나가도록 한다.
				if i == 0 {
					logger.Infof("[Leader-CheckCluster] Retry service lookup in the Storage service. Cause: %+v", err)
					continue
				}
			}

			l.Lock()
			if storageStatus, sErr = checkStorageWarning(storages, l.excluded[id], i); storageStatus == constant.ClusterStateWarning {
				// for 문이 총 2번 돌아가게 되는데 처음에만 continue 를 작동하고 두번째에는 그냥 지나가도록 한다.
				if i == 0 {
					l.Unlock()
					logger.Infof("[Leader-CheckCluster] Retry service lookup because there is an unavailable status in the Storage service.")
					continue
				}
			}
			l.Unlock()

			break
		}
		logger.Debugf("[Leader-CheckCluster] Finished checking the status of the cluster(%d) Cinder volume.", id)
	}()

	go func() {
		logger.Debugf("[Leader-CheckCluster] Starting to check the Compute status of the cluster(%d).", id)
		defer wg.Done()

		for i := 0; i < 2; i++ {
			// Compute 서비스 상태 확인
			computesMap, cErr = sync.GetComputeServiceList(s.Cli)
			if cErr != nil {
				logger.Warnf("[Leader-CheckCluster] Error occurred during get compute data. Cause: %+v", cErr)
				l.Lock()
				computeStatus = constant.ClusterStateWarning
				l.Unlock()

				if errors.Equal(cErr, client.ErrRemoteServerError) {
					l.Lock()
					computeStatus = constant.ClusterStateInactive
					l.Unlock()
				}

				// compute 데이터를 가져오지 못해 DB에 상태값을 Unknown 으로 변경
				if err = sync.SetUnknownClusterHypervisorStatus(id); err != nil {
					if i == 0 {
						logger.Infof("[Leader-CheckCluster] Retry service lookup in the Compute service. Cause: %+v", err)
						continue
					}
				}
				break
			}

			if err = s.UpdateClusterHypervisorStatus(computesMap); err != nil {
				if i == 0 {
					logger.Infof("[Leader-CheckCluster] Retry service lookup in the Compute service. Cause: %+v", err)
					continue
				}
			}

			l.Lock()
			if computes, computeStatus, cErr = checkComputeWarning(computesMap, l.excluded[id], i); computeStatus == constant.ClusterStateWarning {
				if i == 0 {
					l.Unlock()
					logger.Infof("[Leader-CheckCluster] Retry service lookup because there is an unavailable status in the Compute service.")
					continue
				}
			}
			l.Unlock()

			break
		}
		logger.Debugf("[Leader-CheckCluster] Finished checking the status of the cluster(%d) Compute.", id)
	}()

	go func() {
		logger.Debugf("[Leader-CheckCluster] Starting to check the Network agent status of the cluster(%d).", id)
		defer wg.Done()

		for i := 0; i < 2; i++ {
			// Network 서비스 상태 확인
			networks, nErr = sync.GetNetworkAgentServices(s.Cli)
			if nErr != nil {
				logger.Warnf("[Leader-CheckCluster] Error occurred during get network data. Cause: %+v", nErr)
				l.Lock()
				networkStatus = constant.ClusterStateWarning
				l.Unlock()

				if errors.Equal(nErr, client.ErrRemoteServerError) {
					l.Lock()
					networkStatus = constant.ClusterStateInactive
					l.Unlock()
				}
				break
			}

			l.Lock()
			if networkStatus, nErr = checkNetworkWarning(networks, l.excluded[id], i); networkStatus == constant.ClusterStateWarning {
				if i == 0 {
					l.Unlock()
					logger.Infof("[Leader-CheckCluster] Retry service lookup because there is an unavailable status in the Network service.")
					continue
				}
			}
			l.Unlock()

			break
		}
		logger.Debugf("[Leader-CheckCluster] Finished checking the status of the cluster(%d) Network agent.", id)
	}()
	wg.Wait()

	l.Lock()
	defer l.Unlock()

	status, se, ce, ne := validateStatusAndErr(storageStatus, computeStatus, networkStatus, sErr, cErr, nErr)

	//status = l.checkStatus(storages, computes, networks, status, l.excluded[id])

	if err = publishToMessage(id, storages, computes, networks, true, status, se, ce, ne); err != nil {
		logger.Warnf("[Leader-CheckCluster] Could not publish check cluster(%d) message. Cause: %+v", id, err)
	}

	logger.Infof("[Leader-CheckCluster] Cluster(%d) Nova, Cinder and Network status check has been completed.", id)
	return status
}

func checkStorageWarning(storages []queue.StorageServices, excluded *queue.ExcludedCluster, i int) (string, error) {
	var err error
	var state string
	if storages != nil {
		for _, value := range storages {
			if value.Status == "unavailable" {
				state = constant.ClusterStateWarning

				if excluded != nil {
					for _, ex := range excluded.Storage {
						// Storage 는 처음 받아올때 ID가 없기때문에 다른걸로 비교한다.
						if value.Binary == ex.Binary && value.Host == ex.Host && value.Zone == ex.Zone && ex.Exception {
							state = constant.ClusterStateActive
						}
					}
				}

				if i == 0 {
					logger.Warnf("[Sync-CheckCluster] This Storage service(name: %s Host: %s) is not available.", value.Binary, value.Host)
				}
			}
		}
	}

	if state == constant.ClusterStateWarning {
		err = errors.New("Inoperative due to 'unavailable' condition in the storage")
	}

	return state, err
}

func checkComputeWarning(computesMap map[string]queue.ComputeServices, excluded *queue.ExcludedCluster, i int) ([]queue.ComputeServices, string, error) {
	var err error
	var state string
	var computes []queue.ComputeServices

	// Hypervisor 는 Update 할때 state, status 둘다 필요해서 업데이트 후에 status 하나로 통일
	// 이유는 storage 와 동일하게 api 를 내보내기 위해서이다.
	if computesMap != nil {
		for id, value := range computesMap {
			if value.State == "up" && value.Status == "enabled" {
				value.Status = "available"
			} else {
				value.Status = "unavailable"
				state = constant.ClusterStateWarning

				if excluded != nil {
					if ex, ok := excluded.Compute[value.ID]; ok && ex.Exception {
						state = constant.ClusterStateActive
					}
				}

				if i == 0 {
					logger.Warnf("[Leader-CheckCluster] This Compute service(name: %s Host: %s) is not available.", value.Binary, value.Host)
				}
			}

			computes = append(computes, queue.ComputeServices{
				ID:        id,
				Binary:    value.Binary,
				Host:      value.Host,
				Zone:      value.Zone,
				Status:    value.Status,
				UpdatedAt: value.UpdatedAt,
			})
		}
	}

	if state == constant.ClusterStateWarning {
		err = errors.New("Inoperative due to 'unavailable' condition in the compute")
	}

	return computes, state, err
}

func checkNetworkWarning(networks []queue.Agent, excluded *queue.ExcludedCluster, i int) (string, error) {
	var err error
	var state string
	if networks != nil {
		for _, value := range networks {
			if value.Status == "unavailable" {
				state = constant.ClusterStateWarning

				if excluded != nil {
					if ex, ok := excluded.Network[value.ID]; ok && ex.Exception {
						state = constant.ClusterStateActive
					}
				}

				if i == 0 {
					logger.Warnf("[Sync-CheckCluster] This Network service(name: %s Host: %s) is not available.", value.Binary, value.Host)
				}
			}
		}
	}

	if state == constant.ClusterStateWarning {
		err = errors.New("Inoperative due to 'unavailable' condition in the network")
	}

	return state, err
}

/*
validateStatusAndErr 함수는 Storage, Compute, Network os-service 들의 상태를 통해 클러스터 상태를 도출하고 에러를 스트링으로 변환하는 함수입니다.
Inactive 상태로 들어온 서비스가 있다면 클러스터 상태도 Inactive 로 출력하고 Warning 상태인 서비스가 있다면 Warning 으로 출력하고, 모두 정상일때만 Active 입니다.
*/
func validateStatusAndErr(storageStatus, computeStatus, networkStatus string, sErr, cErr, nErr error) (string, string, string, string) {
	var status, se, ce, ne string
	if storageStatus == constant.ClusterStateInactive || computeStatus == constant.ClusterStateInactive || networkStatus == constant.ClusterStateInactive {
		status = constant.ClusterStateInactive

	} else if storageStatus == constant.ClusterStateWarning || computeStatus == constant.ClusterStateWarning || networkStatus == constant.ClusterStateWarning {
		status = constant.ClusterStateWarning
	} else {
		status = constant.ClusterStateActive
	}

	// error 타입은 json.marshaling 이 되지 않아 string 으로 변경
	if sErr != nil {
		se = sErr.Error()
	}
	if cErr != nil {
		ce = cErr.Error()
	}
	if nErr != nil {
		ne = nErr.Error()
	}

	return status, se, ce, ne
}

/*
checkStatus 함수는 Storage, Compute, Network os-service 들의 상태를 통해 클러스터 상태를 도출하는 함수입니다.
excluded 변수에는 SyncException API 를 통해 사용자가 제외 하고픈 서비스를 받아온 정보입니다. 만약 서비스의 상태가
unavailable 상태이면 unCnt 값을 올리고 만약 Exception 값이 true 이면 이 서비스를 전체 상태결과에서 제외 하는 것이기 때문에
unCntExd 값을 올립니다. 그러면 unCnt 가 `1`, unCntExd 가 `1` 이기 때문에 unavailable 이어도 서로가 값이 같다면 전체 상태값은 Active 가 됩니다.

Unavailable 상태일때 unCnt 를 카운트하고 그때 제외된 항목이면 unCntExd 도 카운트하여 두 변수가 같다면 Active 상태를 출력합니다.
*/
func (l *Leader) checkStatus() string {
	if l.opts.Excluded != nil {
		var unCnt int    // unCnt 는 서비스 상태가 Unavailable 일때 카운트하는 변수
		var unCntExd int // unCntExd 는 상태가 Unavailable 일때 제외된 항목이면 카운트하는 변수

		if l.opts.CheckCluster.Storage != nil && l.opts.Excluded.Storage != nil {
			for _, st := range l.opts.CheckCluster.Storage {
				if st.Status == "unavailable" {
					unCnt++
					l.opts.CheckCluster.StorageError = "Inoperative due to 'unavailable' condition in the storage"
					for _, ex := range l.opts.Excluded.Storage {
						// Storage 는 처음 받아올때 ID가 없기때문에 다른걸로 비교한다.
						if st.Binary == ex.Binary && st.Host == ex.Host && st.Zone == ex.Zone && ex.Exception {
							unCntExd++
							// unavailable 에서 제외가 되었기때문에 에러메시지를 초기화한다.
							l.opts.CheckCluster.StorageError = ""
						}
					}
				}
			}
		}

		if l.opts.CheckCluster.Compute != nil && l.opts.Excluded.Compute != nil {
			for _, cp := range l.opts.CheckCluster.Compute {
				if cp.Status == "unavailable" {
					unCnt++
					l.opts.CheckCluster.ComputeError = "Inoperative due to 'unavailable' condition in the compute"
					if ex, ok := l.opts.Excluded.Compute[cp.ID]; ok && ex.Exception {
						unCntExd++
						l.opts.CheckCluster.ComputeError = ""
					}
				}
			}
		}

		if l.opts.CheckCluster.Network != nil && l.opts.Excluded.Network != nil {
			for _, n := range l.opts.CheckCluster.Network {
				if n.Status == "unavailable" {
					unCnt++
					l.opts.CheckCluster.NetworkError = "Inoperative due to 'unavailable' condition in the network"
					if ex, ok := l.opts.Excluded.Network[n.ID]; ok && ex.Exception {
						unCntExd++
						l.opts.CheckCluster.NetworkError = ""
					}
				}
			}
		}

		// unCnt 와 unCntExd 가 같다는 것은 unavailable 로 인한 Warning 을 다 제 외 했다는 의미이기 때문에 Active 로 출력
		if unCnt > 0 {
			if unCnt == unCntExd {
				return constant.ClusterStateActive
			}

			return constant.ClusterStateWarning
		}
	}
	return l.opts.CheckCluster.Status
}

func publishToMessage(cid uint64, storages []queue.StorageServices, computes []queue.ComputeServices, networks []queue.Agent, isNew bool, status, sErr, cErr, nErr string) error {
	b, err := json.Marshal(&queue.ClusterServiceStatus{
		ClusterID:    cid,
		Storage:      storages,
		Compute:      computes,
		Network:      networks,
		Status:       status,
		IsNew:        isNew,
		StorageError: sErr,
		ComputeError: cErr,
		NetworkError: nErr,
	})
	if err != nil {
		logger.Warnf("[Sync-CheckCluster] Could not unmarshal message. Cause: %+v", cid, err)
		return err
	}

	return broker.Publish(constant.QueueClusterServiceGetStatus, &broker.Message{Body: b})
}
