package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	c "github.com/datacommand2/cdm-center/cluster-manager/internal/cluster"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/broker"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"

	"encoding/json"
	"time"

	// Connect for openstack
	_ "github.com/datacommand2/cdm-center/cluster-manager/client/openstack"
)

func checkFailedSyncFunc(syncFunc map[string]bool) bool {
	for _, v := range syncFunc {
		if !v {
			return true
		}
	}
	return false
}

func publishToStatus(cid uint64, completion map[string]string) error {
	var progress int64
	count := 0
	for _, value := range completion {
		if value == Completed || value == Failed {
			count++
		}
	}

	if count <= 2 {
		progress = int64(count * 5)
	} else if count > 2 {
		progress = int64((count - 1) * 10)
	}

	b, err := json.Marshal(queue.SyncClusterStatus{
		ClusterID:  cid,
		Completion: completion,
		Progress:   progress,
	})
	if err != nil {
		return errors.Unknown(err)
	}

	return broker.Publish(constant.QueueNoticeClusterSyncStatus, &broker.Message{Body: b})
}

// Cluster 클러스터 상세 동기화
func Cluster(cluster *model.Cluster) error {
	var err error
	now := time.Now()
	if err = c.SetStatus(cluster.ID, c.SyncStatus{StatusCode: constant.ClusterSyncStateCodeInit, StartedAt: now}); err != nil {
		return err
	}

	var s *Synchronizer
	s, err = NewSynchronizer(cluster.ID)
	if err != nil {
		return err
	}

	var syncFunc = map[string]bool{
		ResourceTenantList:           false,
		ResourceAvailabilityZoneList: false,
		ResourceHypervisorList:       false,
		ResourceSecurityGroupList:    false,
		ResourceNetworkList:          false,
		ResourceRouterList:           false,
		ResourceStorageList:          false,
		ResourceVolumeList:           false,
		ResourceVolumeSnapshotList:   false,
		ResourceKeypairList:          false,
		ResourceInstanceList:         false,
	}
	// syncStatusFunc 은 동기화 상태의 진행 상황을 UI에 전달하기 위한 데이터
	var syncStatusFunc = map[string]string{
		ResourceTenantList:           "waiting",
		ResourceAvailabilityZoneList: "waiting",
		ResourceHypervisorList:       "waiting",
		ResourceSecurityGroupList:    "waiting",
		ResourceNetworkList:          "waiting",
		ResourceRouterList:           "waiting",
		ResourceStorageList:          "waiting",
		ResourceVolumeList:           "waiting",
		ResourceVolumeSnapshotList:   "waiting",
		ResourceKeypairList:          "waiting",
		ResourceInstanceList:         "waiting",
	}

	defer func() {
		if err != nil {
			if e := c.SetStatus(s.clusterID, c.SyncStatus{StatusCode: constant.ClusterSyncStateCodeFailed}); err != nil {
				logger.Warnf("[Sync-Cluster] Could not set cluster sync status. Cause: %+v", e)
			}
		}

		if e := s.Close(); e != nil {
			logger.Warnf("[Sync-Cluster] Could not close synchronizer client. Cause: %+v", e)
		}
	}()

	logger.Infof("[Sync-Cluster] Start to sync cluster(%d:%s)", s.clusterID, s.clusterName)
	if err = c.SetStatus(s.clusterID, c.SyncStatus{StatusCode: constant.ClusterSyncStateCodeRunning}); err != nil {
		return err
	}

	for i := 0; i < 3 && checkFailedSyncFunc(syncFunc); i++ {
		if i > 0 {
			logger.Infof("[Sync-Cluster] Synchronization service encountered an error while running and is being restarted for the (%d)nd time.", i)
		}
		time.Sleep(time.Duration(i) * 2 * time.Minute)

		// sync cluster tenant listRetry service lookup because
		if syncStatusFunc[ResourceTenantList] != Completed {
			syncStatusFunc[ResourceTenantList] = Running
			if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
				logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) tenant list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
			}
		}

		if !syncFunc[ResourceTenantList] {
			if err = s.SyncClusterTenantList(); err == nil {
				syncFunc[ResourceTenantList] = true
				syncStatusFunc[ResourceTenantList] = Completed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) tenant list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}

			} else {
				syncStatusFunc[ResourceTenantList] = Failed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) tenant list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}
				logger.Warnf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d:%s) tenant list. Cause: %+v", s.clusterID, s.clusterName, err)
			}
		}

		// sync cluster az list
		if syncStatusFunc[ResourceAvailabilityZoneList] != Completed {
			syncStatusFunc[ResourceAvailabilityZoneList] = Running
			if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
				logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) availability zone list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
			}
		}

		if !syncFunc[ResourceAvailabilityZoneList] {
			if err = s.SyncClusterAvailabilityZoneList(); err == nil {
				syncFunc[ResourceAvailabilityZoneList] = true
				syncStatusFunc[ResourceAvailabilityZoneList] = Completed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) availability zone list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}

			} else {
				syncStatusFunc[ResourceAvailabilityZoneList] = Failed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) availability zone list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}
				logger.Warnf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d:%s) availability zone list. Cause: %+v", s.clusterID, s.clusterName, err)
			}
		}

		// sync cluster hypervisor list
		if syncStatusFunc[ResourceHypervisorList] != Completed {
			syncStatusFunc[ResourceHypervisorList] = Running
			if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
				logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) hypervisor list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
			}
		}

		if !syncFunc[ResourceHypervisorList] {
			if err = s.SyncClusterHypervisorList(); err == nil {
				syncFunc[ResourceHypervisorList] = true
				syncStatusFunc[ResourceHypervisorList] = Completed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) hypervisor list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}

			} else {
				syncStatusFunc[ResourceHypervisorList] = Failed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) hypervisor list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}
				logger.Warnf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d:%s) hypervisor list. Cause: %+v", s.clusterID, s.clusterName, err)
			}
		}

		// sync cluster security group list
		if syncStatusFunc[ResourceSecurityGroupList] != Completed {
			syncStatusFunc[ResourceSecurityGroupList] = Running
			if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
				logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) security group list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
			}
		}

		if !syncFunc[ResourceSecurityGroupList] {
			if err = s.SyncClusterSecurityGroupList(); err == nil {
				syncFunc[ResourceSecurityGroupList] = true
				syncStatusFunc[ResourceSecurityGroupList] = Completed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) security group list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}

			} else {
				syncStatusFunc[ResourceSecurityGroupList] = Failed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) security group list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}
				logger.Warnf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d:%s) security group list. Cause: %+v", s.clusterID, s.clusterName, err)
			}
		}

		// sync cluster network list
		if syncStatusFunc[ResourceNetworkList] != Completed {
			syncStatusFunc[ResourceNetworkList] = Running
			if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
				logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) network list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
			}
		}

		if !syncFunc[ResourceNetworkList] {
			if err = s.SyncClusterNetworkList(); err == nil {
				syncFunc[ResourceNetworkList] = true
				syncStatusFunc[ResourceNetworkList] = Completed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) network list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}

			} else {
				syncStatusFunc[ResourceNetworkList] = Failed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) network list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}
				logger.Warnf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d:%s) network list. Cause: %+v", s.clusterID, s.clusterName, err)
			}
		}

		// sync cluster router list
		if syncStatusFunc[ResourceRouterList] != Completed {
			syncStatusFunc[ResourceRouterList] = Running
			if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
				logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) router list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
			}
		}

		if !syncFunc[ResourceRouterList] {
			if err = s.SyncClusterRouterList(); err == nil {
				syncFunc[ResourceRouterList] = true
				syncStatusFunc[ResourceRouterList] = Completed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) router list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}

			} else {
				syncStatusFunc[ResourceRouterList] = Failed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) router list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}
				logger.Warnf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d:%s) router list. Cause: %+v", s.clusterID, s.clusterName, err)
			}
		}

		// sync cluster storage list
		if syncStatusFunc[ResourceStorageList] != Completed {
			syncStatusFunc[ResourceStorageList] = Running
			if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
				logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) storage list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
			}
		}

		if !syncFunc[ResourceStorageList] {
			if err = s.SyncClusterStorageList(); err == nil {
				syncFunc[ResourceStorageList] = true
				syncStatusFunc[ResourceStorageList] = Completed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) storage list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}

			} else {
				syncStatusFunc[ResourceStorageList] = Failed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) storage list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}
				logger.Warnf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d:%s) storage list. Cause: %+v", s.clusterID, s.clusterName, err)
			}
		}

		// sync cluster volume list
		if syncStatusFunc[ResourceVolumeList] != Completed {
			syncStatusFunc[ResourceVolumeList] = Running
			if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
				logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) volume list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
			}
		}

		if !syncFunc[ResourceVolumeList] {
			if err = s.SyncClusterVolumeList(); err == nil {
				syncFunc[ResourceVolumeList] = true
				syncStatusFunc[ResourceVolumeList] = Completed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) volume list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}

			} else {
				syncStatusFunc[ResourceVolumeList] = Failed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) volume list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}
				logger.Warnf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d:%s) volume list. Cause: %+v", s.clusterID, s.clusterName, err)
			}
		}

		// sync cluster volume snapshot list
		if syncStatusFunc[ResourceVolumeSnapshotList] != Completed {
			syncStatusFunc[ResourceVolumeSnapshotList] = Running
			if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
				logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) volume snapshot list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
			}
		}

		if !syncFunc[ResourceVolumeSnapshotList] {
			if err = s.SyncClusterVolumeSnapshotList(); err == nil {
				syncFunc[ResourceVolumeSnapshotList] = true
				syncStatusFunc[ResourceVolumeSnapshotList] = Completed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) volume snapshot list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}

			} else {
				syncStatusFunc[ResourceVolumeSnapshotList] = Failed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) volume snapshot list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}
				logger.Warnf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d:%s) volume snapshot list. Cause: %+v", s.clusterID, s.clusterName, err)
			}
		}

		// sync cluster keypair list
		if syncStatusFunc[ResourceKeypairList] != Completed {
			syncStatusFunc[ResourceKeypairList] = Running
			if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
				logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) keypair list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
			}
		}

		if !syncFunc[ResourceKeypairList] {
			if err = s.SyncClusterKeypairList(); err == nil {
				syncFunc[ResourceKeypairList] = true
				syncStatusFunc[ResourceKeypairList] = Completed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) keypair list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}

			} else {
				syncStatusFunc[ResourceKeypairList] = Failed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) keypair list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}
				logger.Warnf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d:%s) keypair list. Cause: %+v", s.clusterID, s.clusterName, err)
			}
		}

		// sync cluster instance list
		if syncStatusFunc[ResourceInstanceList] != Completed {
			syncStatusFunc[ResourceInstanceList] = Running
			if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
				logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) instance list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
			}
		}

		if !syncFunc[ResourceInstanceList] {
			if err = s.SyncClusterInstanceList(); err == nil {
				syncFunc[ResourceInstanceList] = true
				syncStatusFunc[ResourceInstanceList] = Completed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) instance list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}

			} else {
				syncStatusFunc[ResourceInstanceList] = Failed
				if pubErr := publishToStatus(cluster.ID, syncStatusFunc); pubErr != nil {
					logger.Warnf("[Sync-Cluster] Error occurred during publishing the cluster(%d:%s) instance list. Cause: %+v", cluster.ID, cluster.Name, pubErr)
				}
				logger.Warnf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d:%s) instance list. Cause: %+v", s.clusterID, s.clusterName, err)
			}
		}

		if !syncFunc[ResourceInstanceSpecList] {
			if err = s.SyncClusterInstanceSpecList(); err != nil {
				syncStatusFunc[ResourceInstanceSpecList] = Failed
				logger.Warnf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d:%s) instance spec list. Cause: %+v", s.clusterID, s.clusterName, err)
			} else {
				syncStatusFunc[ResourceInstanceSpecList] = Completed
			}
		}

	}

	var failedReason []*c.Message
	for k, v := range syncFunc {
		if !v {
			failedReason = append(failedReason, c.MakeFailedMessage(k))
		}
	}

	logger.Infof("[Sync-Cluster] Cluster(%d:%s) synchronization completed.", s.clusterID, s.clusterName)
	if len(failedReason) != 0 {
		logger.Errorf("[Sync-Cluster] Error occurred during synchronizing the cluster(%d). Sync manually.", s.clusterID)
		if err = c.SetStatus(s.clusterID, c.SyncStatus{StatusCode: constant.ClusterSyncStateCodeFailed, Reasons: failedReason}); err != nil {
			return err
		}
	} else {
		if err = c.SetStatus(s.clusterID, c.SyncStatus{StatusCode: constant.ClusterSyncStateCodeDone}); err != nil {
			return err
		}
	}

	return nil
}
