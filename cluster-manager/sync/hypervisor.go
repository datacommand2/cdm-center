package sync

import (
	"github.com/datacommand2/cdm-center/cluster-manager/client"
	"github.com/datacommand2/cdm-center/cluster-manager/client/openstack/train"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/datacommand2/cdm-center/cluster-manager/database/model"
	"github.com/datacommand2/cdm-center/cluster-manager/queue"
	"github.com/datacommand2/cdm-cloud/common/database"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"

	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
)

func compareHypervisor(dbHypervisor *model.ClusterHypervisor, result *client.GetHypervisorResult) bool {
	if dbHypervisor.UUID != result.Hypervisor.UUID {
		return false
	}

	if dbHypervisor.TypeCode != result.Hypervisor.TypeCode {
		return false
	}

	if dbHypervisor.Hostname != result.Hypervisor.Hostname {
		return false
	}

	if dbHypervisor.IPAddress != result.Hypervisor.IPAddress {
		return false
	}

	if dbHypervisor.VcpuTotalCnt != result.Hypervisor.VcpuTotalCnt {
		return false
	}

	if dbHypervisor.VcpuUsedCnt != result.Hypervisor.VcpuUsedCnt {
		return false
	}

	if dbHypervisor.MemTotalBytes != result.Hypervisor.MemTotalBytes {
		return false
	}

	if dbHypervisor.MemUsedBytes != result.Hypervisor.MemUsedBytes {
		return false
	}

	if dbHypervisor.DiskTotalBytes != result.Hypervisor.DiskTotalBytes {
		return false
	}

	if dbHypervisor.DiskUsedBytes != result.Hypervisor.DiskUsedBytes {
		return false
	}

	if dbHypervisor.State != result.Hypervisor.State {
		return false
	}

	if dbHypervisor.Status != result.Hypervisor.Status {
		return false
	}

	var az model.ClusterAvailabilityZone
	if err := database.Execute(func(db *gorm.DB) error {
		return db.Find(&az, &model.ClusterAvailabilityZone{ID: dbHypervisor.ClusterAvailabilityZoneID}).Error
	}); err != nil {
		return false
	}

	if az.Name != result.AvailabilityZone.Name {
		return false
	}

	return true
}

// GetComputeServiceList 시스템 정보의 Compute 목록 가져오기
func GetComputeServiceList(c client.Client) (map[string]queue.ComputeServices, error) {
	// get cinder service info
	rsp, err := train.ListAllComputeServices(c)
	if err != nil {
		logger.Errorf("[GetComputeServices] Could not get all compute service list. Cause: %+v", err)
		return nil, err
	}

	computeMap := make(map[string]queue.ComputeServices)
	if rsp.ComputeServices != nil {
		for _, service := range rsp.ComputeServices {
			computeMap[service.ID] = queue.ComputeServices{
				Binary:    service.Binary,
				Host:      service.Host,
				Zone:      service.ZoneName,
				Status:    service.Status,
				State:     service.State,
				UpdatedAt: service.UpdatedAt,
			}
		}
	}

	return computeMap, nil
}

// SetUnknownClusterHypervisorStatus 클러스터의 compute 모두 unknown 으로 설정
func SetUnknownClusterHypervisorStatus(id uint64) (err error) {
	// get hypervisor list
	var hypervisorList []*model.ClusterHypervisor
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterHypervisor{ClusterID: id}).Find(&hypervisorList).Error
	}); err != nil {
		logger.Errorf("[SetUnknownClusterHypervisorStatus] Unable to get hypervisor list for cluster(%d). Cause: %+v", id, err)
		return errors.UnusableDatabase(err)
	}

	if len(hypervisorList) == 0 {
		logger.Infof("[SetUnknownClusterHypervisorStatus] storageList is empty: cluster(%d)", id)
		return nil
	}

	for _, hypervisor := range hypervisorList {
		if hypervisor.Status == "unknown" && hypervisor.State == "unknown" {
			continue
		}

		hypervisor.Status = "unknown"
		hypervisor.State = "unknown"

		if err = database.Transaction(func(db *gorm.DB) error {
			return db.Save(hypervisor).Error
		}); err != nil {
			logger.Warnf("[SetUnknownClusterHypervisorStatus] The hypervisor(%d) status of the cluster(%d) cannot be updated to 'Unknown'. Cause: %+v",
				hypervisor.Hostname, id, err)
			continue
		}

		logger.Infof("[SetUnknownClusterHypervisorStatus] The hypervisor(%d) status of the cluster(%d) has changed to 'Unknown'.", hypervisor.Hostname, id)
	}

	return nil
}

// UpdateClusterHypervisorStatus Compute status update
func (s *Synchronizer) UpdateClusterHypervisorStatus(computes map[string]queue.ComputeServices) (err error) {
	// get hypervisor list
	var hypervisorList []*model.ClusterHypervisor
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterHypervisor{ClusterID: s.clusterID}).Find(&hypervisorList).Error
	}); err != nil {
		logger.Errorf("[UpdateClusterHypervisorStatus] Could not get hypervisor list: cluster(%d). Cause: %+v", s.clusterID, err)
		return errors.UnusableDatabase(err)
	}

	if len(hypervisorList) == 0 {
		logger.Infof("[UpdateClusterComputeStatus] storageList is empty: cluster(%d)", s.clusterID)
		return nil
	}

	for _, dbHyper := range hypervisorList {
		if computes != nil {
			value := computes[dbHyper.UUID]
			if value.State != "" && value.Status != "" {
				if value.Status == dbHyper.Status && value.State == dbHyper.State {
					continue
				}

				dbHyper.Status = value.Status
				dbHyper.State = value.State

				if err = database.Transaction(func(db *gorm.DB) error {
					return db.Save(dbHyper).Error
				}); err != nil {
					logger.Warnf("[UpdateClusterComputeStatus] Could not update hypervisor(%d) : cluster(%d). Cause: %+v", dbHyper.Hostname, s.clusterID, err)
					continue
				}

				logger.Infof("[UpdateClusterComputeStatus] The hypervisor(%s) in the cluster(%d) has changed status(%s) and state(%s).",
					dbHyper.Hostname, s.clusterID, dbHyper.State, dbHyper.Status)
			}
		}
	}

	return nil
}

func (s *Synchronizer) getHypervisor(hypervisor *model.ClusterHypervisor, opts ...Option) (*model.ClusterHypervisor, error) {
	var err error
	var h model.ClusterHypervisor

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterHypervisor{ClusterID: s.clusterID, UUID: hypervisor.UUID}).First(&h).Error
	}); err == nil {
		return &h, nil
	}

	if err = s.SyncClusterHypervisor(&model.ClusterHypervisor{UUID: hypervisor.UUID}, opts...); err != nil {
		return nil, err
	}

	err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterHypervisor{ClusterID: s.clusterID, UUID: hypervisor.UUID}).First(&h).Error
	})
	switch {
	case err == gorm.ErrRecordNotFound:
		return nil, nil

	case err != nil:
		return nil, errors.UnusableDatabase(err)
	}

	return &h, nil
}

func (s *Synchronizer) deleteHypervisor(hypervisor *model.ClusterHypervisor) (err error) {
	logger.Infof("[Sync-deleteHypervisor] Start: cluster(%d) hypervisor(%s)", s.clusterID, hypervisor.UUID)

	var instanceList []model.ClusterInstance
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterInstance{ClusterHypervisorID: hypervisor.ID}).Find(&instanceList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// delete instance
	for _, instance := range instanceList {
		if err = s.deleteInstance(&instance); err != nil {
			return err
		}
	}

	// delete hypervisor
	if err = database.GormTransaction(func(db *gorm.DB) error {
		return db.Delete(hypervisor).Error
	}); err != nil {
		logger.Errorf("[Sync-deleteHypervisor] Could not delete cluster(%d) hypervisor(%s). Cause: %+v", s.clusterID, hypervisor.UUID, err)
		return errors.UnusableDatabase(err)
	}

	if err = publishMessage(constant.QueueNoticeClusterHypervisorDeleted, &queue.DeleteClusterHypervisor{
		Cluster:        &model.Cluster{ID: s.clusterID},
		OrigHypervisor: hypervisor,
	}); err != nil {
		logger.Warnf("[Sync-deleteHypervisor] Could not publish cluster(%d) hypervisor(%s) deleted message. Cause: %+v", s.clusterID, hypervisor.UUID, err)
	}

	logger.Infof("[Sync-deleteHypervisor] Success: cluster(%d) hypervisor(%s)", s.clusterID, hypervisor.UUID)
	return nil
}

func (s *Synchronizer) syncDeletedHypervisorList(rsp *client.GetHypervisorListResponse) (err error) {
	var hypervisorList []*model.ClusterHypervisor

	// 동기화되어있는 하이퍼바이저 목록 조회
	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterHypervisor{ClusterID: s.clusterID}).Find(&hypervisorList).Error
	}); err != nil {
		return errors.UnusableDatabase(err)
	}

	// 클러스터에서 삭제된 하이퍼바이저 제거
	for _, hypervisor := range hypervisorList {
		exists := false
		for _, clusterHypervisor := range rsp.ResultList {
			exists = exists || (clusterHypervisor.Hypervisor.UUID == hypervisor.UUID)
		}

		if exists {
			continue
		}

		if err = s.deleteHypervisor(hypervisor); err != nil {
			return err
		}
	}

	return nil
}

// SyncClusterHypervisorList 하이퍼바이저 목록 동기화
func (s *Synchronizer) SyncClusterHypervisorList(opts ...Option) error {
	logger.Infof("[SyncClusterHypervisorList] Start: cluster(%d)", s.clusterID)

	// 클러스터의 하이퍼바이저 목록 조회
	rsp, err := s.Cli.GetHypervisorList()
	if err != nil {
		return err
	}

	// 클러스터에서 삭제된 하이퍼바이저 제거
	if err = s.syncDeletedHypervisorList(rsp); err != nil {
		return err
	}

	// 하이퍼바이저 상세 동기화
	for _, clusterHypervisor := range rsp.ResultList {
		if e := s.SyncClusterHypervisor(&model.ClusterHypervisor{UUID: clusterHypervisor.Hypervisor.UUID}, opts...); e != nil {
			logger.Warnf("[SyncClusterHypervisorList] Failed to sync cluster(%d) hypervisor(%s). Cause: %+v", s.clusterID, clusterHypervisor.Hypervisor.UUID, e)
			err = e
		}
	}

	logger.Infof("[SyncClusterHypervisorList] Success: cluster(%d)", s.clusterID)
	return err
}

// SyncClusterHypervisor 하이퍼바이저 상세 동기화
func (s *Synchronizer) SyncClusterHypervisor(hypervisor *model.ClusterHypervisor, opts ...Option) error {
	logger.Infof("[SyncClusterHypervisor] Start: cluster(%d) hypervisor(%s)", s.clusterID, hypervisor.UUID)

	var options Options
	for _, o := range opts {
		o(&options)
	}

	rsp, err := s.Cli.GetHypervisor(client.GetHypervisorRequest{Hypervisor: model.ClusterHypervisor{UUID: hypervisor.UUID}})
	if err != nil && !errors.Equal(err, client.ErrNotFound) {
		return err
	}

	if err = database.Execute(func(db *gorm.DB) error {
		return db.Where(&model.ClusterHypervisor{ClusterID: s.clusterID, UUID: hypervisor.UUID}).First(hypervisor).Error
	}); err != nil && err != gorm.ErrRecordNotFound {
		return errors.UnusableDatabase(err)
	}

	origHypervisor := new(model.ClusterHypervisor)
	if err = copier.CopyWithOption(origHypervisor, hypervisor, copier.Option{DeepCopy: true}); err != nil {
		return errors.Unknown(err)
	}

	switch {
	case rsp == nil && err == gorm.ErrRecordNotFound: // 디비에도 없고, 클러스터에도 없음
		logger.Infof("[SyncClusterHypervisor] Success - nothing was updated: cluster(%d) hypervisor(%s)", s.clusterID, hypervisor.UUID)
		return nil

	case rsp == nil: // 하이퍼바이저 삭제
		if err = s.deleteHypervisor(hypervisor); err != nil {
			return err
		}

	case err == gorm.ErrRecordNotFound: // 하이퍼바이저 추가
		if rsp.Result.AvailabilityZone.Name == "" {
			logger.Warnf("[SyncClusterHypervisor] Could not sync cluster(%d) hypervisor(%s). Cause: empty availability zone", s.clusterID, rsp.Result.Hypervisor.UUID)
			return nil
		}

		var az *model.ClusterAvailabilityZone
		if az, err = s.getAvailabilityZone(&model.ClusterAvailabilityZone{Name: rsp.Result.AvailabilityZone.Name}, opts...); err != nil {
			return err
		} else if az == nil {
			logger.Warnf("[SyncClusterHypervisor] Could not sync cluster(%d) hypervisor(%s). Cause: not found availability zone(%s)",
				s.clusterID, rsp.Result.Hypervisor.UUID, rsp.Result.AvailabilityZone.Name)
			return nil
		}

		rsp.Result.Hypervisor.ClusterID = s.clusterID
		rsp.Result.Hypervisor.ClusterAvailabilityZoneID = az.ID
		*hypervisor = rsp.Result.Hypervisor

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(hypervisor).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterHypervisor] Done - added: cluster(%d) hypervisor(%s)", s.clusterID, hypervisor.UUID)

		if err = publishMessage(constant.QueueNoticeClusterHypervisorCreated, &queue.CreateClusterHypervisor{
			Cluster:    &model.Cluster{ID: s.clusterID},
			Hypervisor: hypervisor,
		}); err != nil {
			logger.Warnf("[SyncClusterHypervisor] Could not publish cluster(%d) hypervisor(%s) created message. Cause: %+v", s.clusterID, rsp.Result.Hypervisor.UUID, err)
		}

	case options.Force || !compareHypervisor(hypervisor, &rsp.Result): // 하이퍼바이저 정보 수정
		if rsp.Result.AvailabilityZone.Name == "" {
			logger.Warnf("[SyncClusterHypervisor] Could not sync cluster(%d) hypervisor(%s). Cause: empty availability zone", s.clusterID, rsp.Result.Hypervisor.UUID)
			return nil
		}

		var az *model.ClusterAvailabilityZone
		if az, err = s.getAvailabilityZone(&model.ClusterAvailabilityZone{Name: rsp.Result.AvailabilityZone.Name}, opts...); err != nil {
			return err
		} else if az == nil {
			logger.Warnf("[SyncClusterHypervisor] Could not sync cluster(%d) hypervisor(%s). Cause: not found availability zone(%s)",
				s.clusterID, rsp.Result.Hypervisor.UUID, rsp.Result.AvailabilityZone.Name)
			return nil
		}

		rsp.Result.Hypervisor.ID = hypervisor.ID
		rsp.Result.Hypervisor.ClusterID = s.clusterID
		rsp.Result.Hypervisor.ClusterAvailabilityZoneID = az.ID
		rsp.Result.Hypervisor.SSHPort = hypervisor.SSHPort
		rsp.Result.Hypervisor.SSHAccount = hypervisor.SSHAccount
		rsp.Result.Hypervisor.SSHPassword = hypervisor.SSHPassword
		rsp.Result.Hypervisor.AgentPort = hypervisor.AgentPort
		rsp.Result.Hypervisor.AgentVersion = hypervisor.AgentVersion
		rsp.Result.Hypervisor.AgentInstalledAt = hypervisor.AgentInstalledAt
		rsp.Result.Hypervisor.AgentLastUpgradedAt = hypervisor.AgentLastUpgradedAt
		*hypervisor = rsp.Result.Hypervisor

		if err = database.GormTransaction(func(db *gorm.DB) error {
			return db.Save(hypervisor).Error
		}); err != nil {
			return errors.UnusableDatabase(err)
		}

		logger.Infof("[SyncClusterHypervisor] Done - updated: cluster(%d) hypervisor(%s)", s.clusterID, hypervisor.UUID)
		fallthrough

	default:
		// rsp != nil && err == nil (hypervisor 존재. 변화여부 상관없이 publish)
		if err = publishMessage(constant.QueueNoticeClusterHypervisorUpdated, &queue.UpdateClusterHypervisor{
			Cluster:        &model.Cluster{ID: s.clusterID},
			OrigHypervisor: origHypervisor,
			Hypervisor:     hypervisor,
		}); err != nil {
			logger.Warnf("[SyncClusterHypervisor] Could not publish cluster(%d) hypervisor(%s) updated message. Cause: %+v", s.clusterID, hypervisor.UUID, err)
		}
	}

	logger.Infof("[SyncClusterHypervisor] Success: cluster(%d) hypervisor(%s)", s.clusterID, hypervisor.UUID)
	return nil
}
