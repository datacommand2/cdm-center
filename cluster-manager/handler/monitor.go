package handler

import (
	"context"
	"encoding/json"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/broker"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/google/uuid"
)

// StatusMonitor 클러스터 동기화 진행상태와 서비스 상태 메시지 받는 함수
// 메시지를 받으면 메모리 저장
func (h *ClusterManagerHandler) StatusMonitor() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())

	go func(ctx context.Context) {
		syncSub, err := broker.SubscribeTempQueue(constant.QueueNoticeClusterSyncStatus, h.syncClusterStatusMonitor)
		if err != nil {
			logger.Warnf("[SyncStatusMonitor] QueueNoticeClusterSyncStatus Broker error. Cause: ", err)
		}

		serviceSub, err := broker.SubscribeTempQueue(constant.QueueClusterServiceGetStatus, h.getClusterStatusMonitor)
		if err != nil {
			logger.Warnf("[ServiceStatusMonitor] QueueClusterServiceGetStatus Broker error. Cause: ", err)
		}

		deleteSub, err := broker.SubscribeTempQueue(constant.QueueClusterServiceDelete, h.deleteClusterMonitor)
		if err != nil {
			logger.Warnf("[ServiceStatusMonitor] QueueClusterServiceException Broker error. Cause: ", err)
		}

		// wait until context is done
		<-ctx.Done()

		_ = syncSub.Unsubscribe()
		_ = serviceSub.Unsubscribe()
		_ = deleteSub.Unsubscribe()
	}(ctx)

	return cancel
}

// syncClusterStatus 클러스터의 동기화 진행 상태를 저장하는 함수
func (h *ClusterManagerHandler) syncClusterStatusMonitor(p broker.Event) error {
	var msg queue.SyncClusterStatus
	if err := json.Unmarshal(p.Message().Body, &msg); err != nil {
		logger.Errorf("[syncClusterStatusMonitor] Error occurred during unmarshal the message. Cause: %+v", err)
		return err
	}

	h.syncStatus[msg.ClusterID] = queue.SyncClusterStatus{
		Completion: msg.Completion,
		Progress:   msg.Progress,
	}

	return nil
}

// getClusterStatus 클러스터의 compute, cinder, network 상태 메시지를 메모리에 저장하는 함수
func (h *ClusterManagerHandler) getClusterStatusMonitor(e broker.Event) error {
	var cnt int
	var msg queue.ClusterServiceStatus
	if err := json.Unmarshal(e.Message().Body, &msg); err != nil {
		logger.Errorf("[getClusterStatusMonitor] Error occurred during unmarshal the message. Cause: %+v", err)
		return errors.Unknown(err)
	}

	// 메모리 데이터를 변수에 저장
	networks := h.checkStatus[msg.ClusterID].Network
	computes := h.checkStatus[msg.ClusterID].Compute
	storages := h.checkStatus[msg.ClusterID].Storage

	// 이 코드는 Network, Compute, Storage 모두 동일한 방식입니다.
	if networks != nil && msg.Network != nil {
		for _, m := range msg.Network {
			found := false
			for i, n := range networks {
				// 두 개의 리스트를 반복문을 통해 비교하고 같은 값을 가지는 것은 msg 로 들어온 데이터로 대체합니다.
				// 대체 할때 Exception 은 이곳에서 바뀌지 않아야 하므로 넣지않는다.
				if n.ID == m.ID {
					networks[i] = m
					// OpenStack 에서 새롭게 불러온 Data 라면 Deleted 값을 false 로 준다. 사라졌던 데이터가 새롭게 다시 나타났을때 필요
					if msg.IsNew {
						networks[i].Deleted = false
					}
					found = true
					break
				}
			}
			// networks 메모리 데이터에는 없고 msg.Network 있는 값은 networks 메모리 데이터에 추가합니다.
			if !found {
				networks = append(networks, m)
			}
		}

		// networks 메모리 데이터에만 있고 msg.Network 에는 없는 값은 메모리 데이터에 사라진 데이터라고 표시합니다.
		for i := 0; i < len(networks); i++ {
			found := false
			for _, n := range msg.Network {
				if networks[i].ID == n.ID {
					found = true
					break
				}
			}
			if !found {
				networks[i].Deleted = true
				msg.NetworkError = "check if there is a deleted service in the network"
				cnt++
			}
		}
	}

	if computes != nil && msg.Compute != nil {
		for _, m := range msg.Compute {
			found := false
			for i, n := range computes {
				// 두 개의 리스트를 반복문을 통해 비교하고 같은 값을 가지는 것은 msg 로 들어온 데이터로 대체합니다.
				if n.ID == m.ID {
					computes[i] = m
					if msg.IsNew {
						computes[i].Deleted = false
					}
					found = true
					break
				}
			}
			// computes 메모리 데이터에는 없고 msg.Compute 있는 값은 networks 메모리 데이터에 추가합니다.
			if !found {
				computes = append(computes, m)
			}
		}

		// computes 메모리 데이터에만 있고 msg.Compute 에는 없는 값은 메모리 데이터에 사라진 데이터라고 표시합니다.
		for i := 0; i < len(computes); i++ {
			found := false
			for _, n := range msg.Compute {
				if computes[i].ID == n.ID {
					found = true
					break
				}
			}
			if !found {
				computes[i].Deleted = true
				msg.ComputeError = "check if there is a deleted service in the compute"
				cnt++
			}
		}
	}

	if storages != nil && msg.Storage != nil {
		for _, m := range msg.Storage {
			found := false
			for i, n := range storages {
				// 두 개의 리스트를 반복문을 통해 비교하고 같은 값을 가지는 것은 msg 로 들어온 데이터로 대체합니다.
				// Storage 는 값이 들어올때 ID가 없기때문에 기존값을 넣어줘야 합니다.
				if n.Binary == m.Binary && n.Host == m.Host && n.Zone == m.Zone {
					storages[i] = m
					storages[i].ID = n.ID
					if msg.IsNew {
						storages[i].Deleted = false
					}
					found = true
					break
				}
			}
			// storages 메모리 데이터에는 없고 msg.Storage 에 있는 값은 storages 메모리 데이터에 추가합니다.
			if !found {
				id := uuid.New()
				storages = append(storages, queue.StorageServices{
					ID:          id.String(),
					Binary:      m.Binary,
					BackendName: m.BackendName,
					Host:        m.Host,
					Zone:        m.Zone,
					Status:      m.Status,
					UpdatedAt:   m.UpdatedAt,
				})
			}
		}

		// storages 메모리 데이터에만 있고 msg.Storage 에는 없는 값은 메모리 데이터에 사라진 데이터라고 표시합니다.
		for i := 0; i < len(storages); i++ {
			found := false
			for _, n := range msg.Storage {
				if n.Binary == storages[i].Binary && n.Host == storages[i].Host && n.Zone == storages[i].Zone {
					found = true
					break
				}
			}
			if !found {
				storages[i].Deleted = true
				msg.StorageError = "check if there is a deleted service in the storage"
				cnt++
			}
		}
	}

	// 기존 메모리 데이터에 값이 없으면 값을 그냥 넣는다. 하지만 Storage 는 처음에 ID 값이 없기 때문에 새로 만들어서 넣어준다.
	if networks == nil {
		networks = msg.Network
	}
	if computes == nil {
		computes = msg.Compute
	}
	if storages == nil {
		for _, m := range msg.Storage {
			id := uuid.New()
			storages = append(storages, queue.StorageServices{
				ID:          id.String(),
				Binary:      m.Binary,
				BackendName: m.BackendName,
				Host:        m.Host,
				Zone:        m.Zone,
				Status:      m.Status,
				UpdatedAt:   m.UpdatedAt,
			})
		}
	}

	h.updatedAt[msg.ClusterID] = e.Message().TimeStamp

	// cnt 가 0 보다 크면 기존 데이터에서 삭제된 서비스가 있기 때문에 확인하는 차원에서 상태를 Warning 으로 준다.
	if cnt > 0 {
		msg.Status = constant.ClusterStateWarning
	}

	h.checkStatus[msg.ClusterID] = queue.ClusterServiceStatus{
		ClusterID:    msg.ClusterID,
		Storage:      storages,
		Compute:      computes,
		Network:      networks,
		Status:       msg.Status,
		StorageError: msg.StorageError,
		ComputeError: msg.ComputeError,
		NetworkError: msg.NetworkError,
	}

	logger.Infof("[Handler-Monitor] Completed saving the cluster(%d) status to memory.", msg.ClusterID)

	return nil
}

// deleteClusterMonitor 클러스터가 API 를 통해 삭제가 되면 메모리에서 삭제를 하는 함수
func (h *ClusterManagerHandler) deleteClusterMonitor(e broker.Event) error {
	var clusterID uint64
	if err := json.Unmarshal(e.Message().Body, &clusterID); err != nil {
		logger.Errorf("[deleteClusterMonitor] Error occurred during unmarshal the message. Cause: %+v", err)
		return errors.Unknown(err)
	}

	// 클러스터 삭제시 메모리 데이터도 삭제
	if _, ok := h.checkStatus[clusterID]; ok {
		delete(h.checkStatus, clusterID)
	}
	if _, ok := h.updatedAt[clusterID]; ok {
		delete(h.updatedAt, clusterID)
	}
	if _, ok := h.syncStatus[clusterID]; ok {
		delete(h.syncStatus, clusterID)
	}

	logger.Infof("[Handler-Monitor] Completed deleting the cluster(%d) data from memory.", clusterID)

	return nil
}
