package migrator

import (
	"context"
	"encoding/json"
	"fmt"
	cms "github.com/datacommand2/cdm-center/cluster-manager/proto"
	"github.com/datacommand2/cdm-cloud/common/broker"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/store"
	"github.com/datacommand2/cdm-cloud/common/sync"
	identity "github.com/datacommand2/cdm-cloud/services/identity/proto"
	"github.com/datacommand2/cdm-disaster-recovery/common/constant"
	drms "github.com/datacommand2/cdm-disaster-recovery/manager/proto"
	"github.com/google/uuid"
	"path"
	"regexp"
	"sort"
	"strconv"
	"time"
)

const (
	recoveryJobKeyBase   = "dr.recovery.job"
	recoveryJobKeyRegexp = "^dr.recovery.job/\\d+$"
	recoveryJobKeyFormat = "dr.recovery.job/%d"

	recoveryJobKeyPrefixFormat = "dr.recovery.job/%d/"

	recoveryJobLogKeyBase           = "dr.recovery.job/%d/log"
	recoveryJobLogKeyRegexp         = "^dr.recovery.job/\\d+/log/\\d+$"
	recoveryJobLogKeyFormat         = "dr.recovery.job/%d/log/%d"
	recoveryJobLastLogSeqKeyFormat  = "dr.recovery.job/%d/log/last_log_seq"
	recoveryJobLastLogSeqLockFormat = "dr.recovery.job/%d/log/last_log_seq_lock"

	recoveryJobDetailKeyFormat    = "dr.recovery.job/%d/detail"
	recoveryJobOperationKeyFormat = "dr.recovery.job/%d/operation"
	recoveryJobStatusKeyFormat    = "dr.recovery.job/%d/status"
	recoveryJobResultKeyFormat    = "dr.recovery.job/%d/result"
	recoveryJobMetadataKeyFormat  = "dr.recovery.job/%d/metadata/%s"

	recoveryJobVolumeStatusKeyFormat   = "dr.recovery.job/%d/result/volume/%d"
	recoveryJobInstanceStatusKeyFormat = "dr.recovery.job/%d/result/instance/%d"

	recoveryJobClearingFormat = "dr.recovery.job/cluster/%d/clear"
	recoveryJobRunningFormat  = "dr.recovery.job/cluster/%d/run"

	recoveryJobFailbackProtectionGroupFormat = "dr.protection_group/%d/recovery_cluster/%d/failback_protection_group"
	recoveryJobFailbackProtectionGroupPrefix = "dr.protection_group/%d/recovery_cluster/"
)

// RecoveryJob 특정 job 의 정보
type RecoveryJob struct {
	RecoveryJobID         uint64                        `json:"recovery_job_id,omitempty"`
	RecoveryJobHash       string                        `json:"recovery_job_hash,omitempty"`
	ProtectionGroupID     uint64                        `json:"protection_group_id,omitempty"`
	ProtectionCluster     *cms.Cluster                  `json:"protection_cluster,omitempty"`
	RecoveryCluster       *cms.Cluster                  `json:"recovery_cluster,omitempty"`
	RecoveryJobTypeCode   string                        `json:"recovery_job_type_code,omitempty"`
	RecoveryTimeObjective uint32                        `json:"recovery_time_objective,omitempty"`
	RecoveryPointTypeCode string                        `json:"recovery_point_type_code,omitempty"`
	RecoveryPointSnapshot *drms.ProtectionGroupSnapshot `json:"recovery_point_snapshot,omitempty"`
	RecoveryPoint         int64                         `json:"recovery_point,omitempty"`
	TriggeredAt           int64                         `json:"triggered_at,omitempty"`
	Approver              *identity.User                `json:"approver,omitempty"`
	RecoveryResultID      uint64                        `json:"recovery_result_id,omitempty"`
}

// RecoveryJobStatus 특정 job 의 status
type RecoveryJobStatus struct {
	StateCode string `json:"state_code,omitempty"`

	StartedAt   int64 `json:"started_at,omitempty"`
	PausedTime  int64 `json:"paused_time,omitempty"`
	FinishedAt  int64 `json:"finished_at,omitempty"`
	ElapsedTime int64 `json:"-"`

	PausedAt   int64 `json:"paused_at,omitempty"`
	ResumeAt   int64 `json:"resume_at,omitempty"`
	RollbackAt int64 `json:"rollback_at,omitempty"`
}

// RecoveryJobOperation 특정 job 의 operation
type RecoveryJobOperation struct {
	Operation string `json:"operation,omitempty"`
}

// RecoveryJobResult 특정 job 의 result
type RecoveryJobResult struct {
	ResultCode string `json:"result_code,omitempty"`

	WarningReasons []*Message `json:"warning_reasons,omitempty"`
	FailedReasons  []*Message `json:"failed_reasons,omitempty"`
}

// RecoveryJobVolumeStatus 특정 job 의 볼륨 status
type RecoveryJobVolumeStatus struct {
	RecoveryJobTaskID string `json:"recovery_job_task_id,omitempty"`

	RecoveryPointTypeCode string `json:"recovery_point_type_code,omitempty"`
	RecoveryPoint         int64  `json:"recovery_point,omitempty"`

	StateCode   string `json:"state_code,omitempty"`
	ResultCode  string `json:"result_code,omitempty"`
	StartedAt   int64  `json:"started_at,omitempty"`
	FinishedAt  int64  `json:"finished_at,omitempty"`
	ElapsedTime int64  `json:"-"`

	FailedReason *Message `json:"failed_reason,omitempty"`
	RollbackFlag bool     `json:"rollback_flag,omitempty"`
}

// RecoveryJobInstanceStatus 특정 job 의 인스턴스 status
type RecoveryJobInstanceStatus struct {
	RecoveryJobTaskID string `json:"recovery_job_task_id,omitempty"`

	RecoveryPointTypeCode string `json:"recovery_point_type_code,omitempty"`
	RecoveryPoint         int64  `json:"recovery_point,omitempty"`

	StateCode   string `json:"state_code,omitempty"`
	ResultCode  string `json:"result_code,omitempty"`
	StartedAt   int64  `json:"started_at,omitempty"`
	FinishedAt  int64  `json:"finished_at,omitempty"`
	ElapsedTime int64  `json:"-"`

	FailedReason *Message `json:"failed_reason,omitempty"`
}

// RecoveryJobLog 특정 job 의 log
type RecoveryJobLog struct {
	*Message

	LogSeq uint64 `json:"log_seq,omitempty"`
	LogDt  int64  `json:"log_dt,omitempty"`
}

// GetJobList job list 조회
func GetJobList() ([]*RecoveryJob, error) {
	keys, err := store.List(recoveryJobKeyBase)
	if err != nil && err != store.ErrNotFoundKey {
		return nil, errors.UnusableStore(err)
	}

	var jobs []*RecoveryJob
	for _, k := range keys {
		var matched bool
		if matched, err = regexp.Match(recoveryJobKeyRegexp, []byte(k)); err != nil {
			return nil, errors.Unknown(err)
		}
		if !matched {
			continue
		}

		var jid uint64
		if _, err = fmt.Sscanf(k, recoveryJobKeyFormat, &jid); err != nil {
			return nil, errors.Unknown(err)
		}

		var j *RecoveryJob
		if j, err = GetJob(jid); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	return jobs, nil
}

// GetJob job 조회
func GetJob(jid uint64) (*RecoveryJob, error) {
	v, err := store.Get(fmt.Sprintf(recoveryJobKeyFormat, jid))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundJob(jid)

	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var j RecoveryJob
	if err = json.Unmarshal([]byte(v), &j); err != nil {
		return nil, errors.Unknown(err)
	}

	return &j, nil
}

// GetAllJob job id를 갖는 모든 데이터 조회
func GetAllJob(jobID uint64) (string, error) {
	key := fmt.Sprintf(recoveryJobKeyFormat, jobID)
	job, err := store.Get(key)
	switch {
	case err == store.ErrNotFoundKey:
		logger.Warnf("[job-GetAllJob] Not found key: dr.recovery.job/%d", jobID)
		//return "", nil

	case err != nil:
		logger.Errorf("[job-GetAllJob] Could not get job(%d) data. Cause: %v", jobID, err)
		//return "", errors.UnusableStore(err)
	}

	var rsp string
	if job != "" {
		data := make(map[string]string)
		data[key] = job
		v, err := json.Marshal(data)
		if err == nil {
			rsp = string(v)
		} else {
			logger.Errorf("[job-GetAllJob] Could not get marshal data. Cause: %+v", err)
		}
	}

	allJob, err := store.GetAll(fmt.Sprintf(recoveryJobKeyPrefixFormat, jobID))
	switch {
	case err == store.ErrNotFoundKey:
		logger.Warnf("[job-GetAllJob] Not found key with prefix: dr.recovery.job/%d/", jobID)
		return rsp, nil

	case err != nil:
		logger.Errorf("[job-GetAllJob] Could not get job(%d) data. Cause: %v", jobID, err)
		return rsp, errors.UnusableStore(err)
	}

	if allJob != "" {
		if rsp == "" {
			rsp = allJob
		} else {
			rsp = fmt.Sprintf("%s , %s", rsp, allJob)
		}
	}

	return rsp, nil
}

// GetKey job 의 키 반환
func (j *RecoveryJob) GetKey() string {
	return fmt.Sprintf(recoveryJobKeyFormat, j.RecoveryJobID)
}

// Put job 추가
func (j *RecoveryJob) Put(txn store.Txn) error {
	var err error

	b, err := json.Marshal(j)
	if err != nil {
		return errors.Unknown(err)
	}

	k := fmt.Sprintf(recoveryJobKeyFormat, j.RecoveryJobID)
	txn.Put(k, string(b))

	return nil
}

// Delete job 삭제
func (j *RecoveryJob) Delete(txn store.Txn) {
	// dr.recovery.job/%d/ 형태의 key prefix 옵션으로 삭제
	txn.Delete(fmt.Sprintf(recoveryJobKeyPrefixFormat, j.RecoveryJobID), store.DeletePrefix())
	// dr.recovery.job/%d 해당 key 삭제
	txn.Delete(fmt.Sprintf(recoveryJobKeyFormat, j.RecoveryJobID))
}

// GetDetail job 의 detail 정보 조회
func (j *RecoveryJob) GetDetail(detail interface{}) error {
	v, err := store.Get(fmt.Sprintf(recoveryJobDetailKeyFormat, j.RecoveryJobID))
	switch {
	case errors.Equal(err, store.ErrNotFoundKey):
		return NotFoundJobDetail(j.RecoveryJobID)

	case err != nil:
		return errors.UnusableStore(err)
	}

	if err = json.Unmarshal([]byte(v), detail); err != nil {
		return errors.Unknown(err)
	}

	return nil
}

// SetDetail job 의 detail 정보 설정
func (j *RecoveryJob) SetDetail(txn store.Txn, detail interface{}) error {
	b, err := json.Marshal(detail)
	if err != nil {
		return errors.Unknown(err)
	}

	txn.Put(fmt.Sprintf(recoveryJobDetailKeyFormat, j.RecoveryJobID), string(b))

	return nil
}

// GetMetadata job 의 metadata 를 조회
func (j *RecoveryJob) GetMetadata(key string, out interface{}) error {
	if out == nil {
		return nil
	}

	k := fmt.Sprintf(recoveryJobMetadataKeyFormat, j.RecoveryJobID, key)
	v, err := store.Get(k)
	switch {
	case err == store.ErrNotFoundKey:
		return nil

	case err != nil:
		return errors.UnusableStore(err)
	}

	if err = json.Unmarshal([]byte(v), out); err != nil {
		return errors.Unknown(err)
	}

	return nil
}

// PutMetadata job 의 metadata 를 저장
func (j *RecoveryJob) PutMetadata(txn store.Txn, key string, value interface{}) error {
	b, err := json.Marshal(value)
	if err != nil {
		return errors.Unknown(err)
	}

	k := fmt.Sprintf(recoveryJobMetadataKeyFormat, j.RecoveryJobID, key)
	txn.Put(k, string(b))

	return nil
}

// GetStatus job 의 상태를 조회
func (j *RecoveryJob) GetStatus() (*RecoveryJobStatus, error) {
	k := fmt.Sprintf(recoveryJobStatusKeyFormat, j.RecoveryJobID)
	v, err := store.Get(k)

	switch {
	case err == store.ErrNotFoundKey:
		//return &RecoveryJobStatus{StateCode: constant.RecoveryJobStateCodeWaiting}, nil
		return nil, NotFoundKey(k)

	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var status RecoveryJobStatus
	if err = json.Unmarshal([]byte(v), &status); err != nil {
		return nil, errors.Unknown(err)
	}

	if status.StartedAt != 0 {
		now := time.Now().Unix() - status.PausedAt // PausedAt 은 일시중지 중에만 값이 있다
		if status.FinishedAt != 0 {
			now = status.FinishedAt
		}

		status.ElapsedTime = now - status.StartedAt - status.PausedTime
	}

	return &status, nil
}

// SetStatus 작업의 상태를 변경한다.
func (j *RecoveryJob) SetStatus(txn store.Txn, status *RecoveryJobStatus) error {
	if !isRecoveryJobStateCodes(status.StateCode) {
		return UnknownJobStateCode(status.StateCode)
	}

	b, err := json.Marshal(status)
	if err != nil {
		return errors.Unknown(err)
	}

	k := fmt.Sprintf(recoveryJobStatusKeyFormat, j.RecoveryJobID)
	txn.Put(k, string(b))

	return nil
}

// GetOperation job 의 operation 을 조회
func (j *RecoveryJob) GetOperation() (*RecoveryJobOperation, error) {
	k := fmt.Sprintf(recoveryJobOperationKeyFormat, j.RecoveryJobID)
	v, err := store.Get(k)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundKey(k)

	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var op RecoveryJobOperation
	if err = json.Unmarshal([]byte(v), &op); err != nil {
		return nil, errors.Unknown(err)
	}

	return &op, nil
}

// SetOperation job 의 operation 을 설정
func (j *RecoveryJob) SetOperation(txn store.Txn, op string) error {
	if !isRecoveryJobOperation(op) {
		return UnknownJobOperation(op)
	}

	b, err := json.Marshal(&RecoveryJobOperation{Operation: op})
	if err != nil {
		return errors.Unknown(err)
	}

	k := fmt.Sprintf(recoveryJobOperationKeyFormat, j.RecoveryJobID)
	txn.Put(k, string(b))

	return nil
}

// GetResult job 의 result 조회
func (j *RecoveryJob) GetResult() (*RecoveryJobResult, error) {
	k := fmt.Sprintf(recoveryJobResultKeyFormat, j.RecoveryJobID)
	v, err := store.Get(k)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundJobResult(j.RecoveryJobID)

	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var result RecoveryJobResult
	if err = json.Unmarshal([]byte(v), &result); err != nil {
		return nil, errors.Unknown(err)
	}

	return &result, nil
}

// SetResult job 의 result 설정
func (j *RecoveryJob) SetResult(txn store.Txn, resultCode string, warningReasons, failedReasons []*Message) error {
	if !isRecoveryResultCode(resultCode) {
		return UnknownResultCode(resultCode)
	}

	b, err := json.Marshal(&RecoveryJobResult{
		ResultCode:     resultCode,
		WarningReasons: warningReasons,
		FailedReasons:  failedReasons,
	})
	if err != nil {
		return errors.Unknown(err)
	}

	k := fmt.Sprintf(recoveryJobResultKeyFormat, j.RecoveryJobID)
	txn.Put(k, string(b))

	return nil
}

// GetVolumeStatus 볼륨 복구상태 조회
func (j *RecoveryJob) GetVolumeStatus(id uint64) (*RecoveryJobVolumeStatus, error) {
	k := fmt.Sprintf(recoveryJobVolumeStatusKeyFormat, j.RecoveryJobID, id)
	v, err := store.Get(k)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundJobStatus(j.RecoveryJobID, id, "volume")

	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var result RecoveryJobVolumeStatus
	if err = json.Unmarshal([]byte(v), &result); err != nil {
		return nil, errors.Unknown(err)
	}

	return &result, nil
}

// SetVolumeStatus 볼륨 복구상태 변경
func (j *RecoveryJob) SetVolumeStatus(txn store.Txn, id uint64, status *RecoveryJobVolumeStatus) error {
	b, err := json.Marshal(status)
	if err != nil {
		return errors.Unknown(err)
	}

	k := fmt.Sprintf(recoveryJobVolumeStatusKeyFormat, j.RecoveryJobID, id)
	txn.Put(k, string(b))

	return nil
}

// GetInstanceStatus 인스턴스 복구상태 조회
func (j *RecoveryJob) GetInstanceStatus(id uint64) (*RecoveryJobInstanceStatus, error) {
	k := fmt.Sprintf(recoveryJobInstanceStatusKeyFormat, j.RecoveryJobID, id)
	v, err := store.Get(k)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundJobStatus(j.RecoveryJobID, id, "instance")

	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var result RecoveryJobInstanceStatus
	if err = json.Unmarshal([]byte(v), &result); err != nil {
		return nil, errors.Unknown(err)
	}

	return &result, nil
}

// SetInstanceStatus 인스턴스 복구상태 변경
func (j *RecoveryJob) SetInstanceStatus(txn store.Txn, id uint64, status *RecoveryJobInstanceStatus) error {
	b, err := json.Marshal(status)
	if err != nil {
		return errors.Unknown(err)
	}

	k := fmt.Sprintf(recoveryJobInstanceStatusKeyFormat, j.RecoveryJobID, id)
	txn.Put(k, string(b))

	return nil
}

// GetLogs 특정 job 의 log 목록 조회
func (j *RecoveryJob) GetLogs() ([]*RecoveryJobLog, error) {
	keys, err := store.List(fmt.Sprintf(recoveryJobLogKeyBase, j.RecoveryJobID))
	if err != nil && err != store.ErrNotFoundKey {
		return nil, errors.UnusableStore(err)
	}

	var logs []*RecoveryJobLog
	for _, k := range keys {
		var matched bool
		if matched, err = regexp.Match(recoveryJobLogKeyRegexp, []byte(k)); err != nil {
			return nil, errors.Unknown(err)
		}
		if !matched {
			continue
		}

		var jid uint64
		var sid uint64

		if _, err = fmt.Sscanf(k, recoveryJobLogKeyFormat, &jid, &sid); err != nil {
			return nil, errors.Unknown(err)
		}

		var l *RecoveryJobLog
		if l, err = j.GetLog(sid); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].LogSeq < logs[j].LogSeq
	})

	return logs, nil
}

// GetLogsAfter sequence 이후의 log 목록 조회
func (j *RecoveryJob) GetLogsAfter(seq uint64) ([]*RecoveryJobLog, error) {
	var lastSeq uint64
	var log *RecoveryJobLog
	var logs []*RecoveryJobLog

	s, err := store.Get(fmt.Sprintf(recoveryJobLastLogSeqKeyFormat, j.RecoveryJobID))
	if err != nil && err != store.ErrNotFoundKey {
		return nil, errors.UnusableStore(err)
	}

	if err != store.ErrNotFoundKey {
		lastSeq, _ = strconv.ParseUint(s, 10, 64)
	}

	seq = seq + 1
	for seq <= lastSeq {
		if log, err = j.GetLog(seq); err != nil {
			return nil, err
		}
		logs = append(logs, log)

		seq = seq + 1
	}

	return logs, nil
}

// GetLog 특정 job 의 로그 조회
func (j *RecoveryJob) GetLog(sid uint64) (*RecoveryJobLog, error) {
	v, err := store.Get(fmt.Sprintf(recoveryJobLogKeyFormat, j.RecoveryJobID, sid))
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundJobLog(j.RecoveryJobID, sid)

	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var l RecoveryJobLog
	if err = json.Unmarshal([]byte(v), &l); err != nil {
		return nil, errors.Unknown(err)
	}

	return &l, nil
}

func (j *RecoveryJob) GetLastLogSeq() (uint64, error) {
	s, err := store.Get(fmt.Sprintf(recoveryJobLastLogSeqKeyFormat, j.RecoveryJobID))
	if errors.Equal(err, store.ErrNotFoundKey) {
		return 0, nil
	} else if err != nil {
		return 0, errors.UnusableStore(err)
	}

	lastSeq, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return lastSeq, nil
}

func (j *RecoveryJob) SetLastLogSeq(txn store.Txn, seq uint64) error {
	ctx := context.Background()
	l, err := sync.Lock(ctx, fmt.Sprintf(recoveryJobLastLogSeqLockFormat, j.RecoveryJobID))
	if err != nil {
		return errors.Unknown(err)
	}

	defer func() {
		if err := l.Unlock(ctx); err != nil {
			logger.Warnf("Could not unlock recovery job log seq lock. cause: %v", err)
		}
	}()

	txn.Put(fmt.Sprintf(recoveryJobLastLogSeqKeyFormat, j.RecoveryJobID), strconv.FormatUint(seq, 10))

	return nil
}

// AddLog 로그 추가
// 해당 함수는 동시에 호출시 error 가 발생하지 않아 log 가 누락되는 현상이 있을 수 있음
// 함수 사용시 앞뒤로 sync.Mutex lock, unlock 하여 log 누락을 방지해야함
func (j *RecoveryJob) AddLog(txn store.Txn, sid uint64, code string, contents interface{}) error {
	b, err := json.Marshal(contents)
	if err != nil {
		return errors.Unknown(err)
	}

	var msg = Message{
		Code:     code,
		Contents: string(b),
	}

	log := RecoveryJobLog{
		Message: &msg,
		LogDt:   time.Now().Unix(),
		LogSeq:  sid,
	}

	b, err = json.Marshal(&log)
	if err != nil {
		return errors.Unknown(err)
	}

	txn.Put(fmt.Sprintf(recoveryJobLogKeyFormat, j.RecoveryJobID, log.LogSeq), string(b))

	j.SetLastLogSeq(txn, sid)

	return nil
}

// GetTaskList task list 조회
func (j *RecoveryJob) GetTaskList() ([]*RecoveryJobTask, error) {
	keys, err := store.List(fmt.Sprintf(recoveryJobTaskKeyBase, j.RecoveryJobID))
	if err != nil && err != store.ErrNotFoundKey {
		return nil, errors.UnusableStore(err)
	}

	var tasks []*RecoveryJobTask
	for _, k := range keys {
		var matched bool
		if matched, err = regexp.Match(recoveryJobTaskKeyRegexp, []byte(k)); err != nil {
			return nil, errors.Unknown(err)
		}
		if !matched {
			continue
		}

		var jid uint64
		var tid string

		if _, err = fmt.Sscanf(k, recoveryJobTaskKeyFormat, &jid, &tid); err != nil {
			return nil, errors.Unknown(err)
		}

		var t *RecoveryJobTask
		if t, err = j.GetTask(tid); err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

// GetTask task 조회
func (j *RecoveryJob) GetTask(id string) (*RecoveryJobTask, error) {
	k := fmt.Sprintf(recoveryJobTaskKeyFormat, j.RecoveryJobID, id)
	v, err := store.Get(k)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundTask(j.RecoveryJobID, id)

	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var t RecoveryJobTask
	if err = json.Unmarshal([]byte(v), &t); err != nil {
		return nil, errors.Unknown(err)
	}

	t.RecoveryJobTaskKey = k

	return &t, nil
}

// PutTask task 추가
// 공유 resource task 일 경우 공유 resource task put 도 처리함
func (j *RecoveryJob) PutTask(txn store.Txn, task *RecoveryJobTask) (*RecoveryJobTask, error) {
	task.RecoveryJobTaskID = uuid.New().String()
	if v, ok := sharedTaskMap[task.TypeCode]; ok {
		task.SharedTaskKey = path.Join(v.sharedKey, task.RecoveryJobTaskID)
		if err := task.PutSharedTask(txn); err != nil {
			return nil, err
		}

	}

	task.RecoveryJobID = j.RecoveryJobID
	task.SetRecoveryTaskKey()

	if err := task.Put(txn); err != nil {
		return nil, err
	}

	t := RecoveryJobTask{
		RecoveryJobTaskKey: task.RecoveryJobTaskKey,
		RecoveryJobID:      j.RecoveryJobID,
		RecoveryJobTaskID:  task.RecoveryJobTaskID,
		ReverseTaskID:      task.ReverseTaskID,
		ResourceID:         task.ResourceID,
		ResourceName:       task.ResourceName,
		TypeCode:           task.TypeCode,
		Input:              task.Input,
	}

	b, err := json.Marshal(t)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	if err = broker.Publish(constant.QueueRecoveryJobTaskMonitor, &broker.Message{Body: b}); err != nil {
		return nil, errors.UnusableBroker(err)
	}

	if err = task.SetStatusWaiting(txn, WithSharedTask(true)); err != nil {
		return nil, err
	}

	return task, nil
}

// GetClearTaskList clear task list 조회
func (j *RecoveryJob) GetClearTaskList() ([]*RecoveryJobTask, error) {
	keys, err := store.List(fmt.Sprintf(recoveryJobClearTaskKeyBase, j.RecoveryJobID))
	if err != nil && err != store.ErrNotFoundKey {
		return nil, errors.UnusableStore(err)
	}

	var tasks []*RecoveryJobTask
	for _, k := range keys {
		var matched bool
		if matched, err = regexp.Match(recoveryJobClearTaskKeyRegexp, []byte(k)); err != nil {
			return nil, errors.Unknown(err)
		}
		if !matched {
			continue
		}

		var jid uint64
		var tid string

		if _, err = fmt.Sscanf(k, recoveryJobClearTaskKeyFormat, &jid, &tid); err != nil {
			return nil, errors.Unknown(err)
		}

		var t *RecoveryJobTask
		if t, err = j.GetClearTask(tid); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

// GetClearTask clear task 조회
func (j *RecoveryJob) GetClearTask(id string) (*RecoveryJobTask, error) {
	k := fmt.Sprintf(recoveryJobClearTaskKeyFormat, j.RecoveryJobID, id)
	v, err := store.Get(k)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundClearTask(j.RecoveryJobID, id)

	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var t RecoveryJobTask
	if err = json.Unmarshal([]byte(v), &t); err != nil {
		return nil, errors.Unknown(err)
	}

	t.RecoveryJobTaskKey = k

	return &t, nil
}

// PutClearTask clear task 추가
func (j *RecoveryJob) PutClearTask(txn store.Txn, task *RecoveryJobTask) (*RecoveryJobTask, error) {
	task.RecoveryJobID = j.RecoveryJobID
	task.RecoveryJobTaskID = uuid.New().String()
	task.RecoveryJobTaskKey = fmt.Sprintf(recoveryJobClearTaskKeyFormat, task.RecoveryJobID, task.RecoveryJobTaskID)

	if err := task.Put(txn); err != nil {
		return nil, err
	}

	t := RecoveryJobTask{
		RecoveryJobTaskKey: task.RecoveryJobTaskKey,
		RecoveryJobID:      j.RecoveryJobID,
		RecoveryJobTaskID:  task.RecoveryJobTaskID,
		ReverseTaskID:      task.ReverseTaskID,
		ResourceID:         task.ResourceID,
		ResourceName:       task.ResourceName,
		TypeCode:           task.TypeCode,
		Input:              task.Input,
	}

	b, err := json.Marshal(t)
	if err != nil {
		return nil, errors.Unknown(err)
	}

	if err = broker.Publish(constant.QueueRecoveryJobClearTaskMonitor, &broker.Message{Body: b}); err != nil {
		return nil, errors.UnusableBroker(err)
	}

	return task, nil
}

// IncreaseClearJobCount clear 중인 job 의 수를 증가시킨다.
func (j *RecoveryJob) IncreaseClearJobCount(txn store.Txn) error {
	k := fmt.Sprintf(recoveryJobClearingFormat, j.RecoveryCluster.Id)

	count, err := j.GetClearJobCount()
	if err != nil {
		return errors.UnusableStore(err)
	}

	count++

	txn.Put(k, strconv.FormatUint(count, 10))

	return nil
}

// DecreaseClearJobCount clear 중인 job 의 수를 감소시킨다.
func (j *RecoveryJob) DecreaseClearJobCount(txn store.Txn) error {
	k := fmt.Sprintf(recoveryJobClearingFormat, j.RecoveryCluster.Id)

	count, err := j.GetClearJobCount()
	if err != nil {
		return err
	}

	if count != 0 {
		count--
	}

	if count == 0 {
		txn.Delete(k, store.DeletePrefix())
	} else {
		txn.Put(k, strconv.FormatUint(count, 10))
	}

	return nil
}

// GetClearJobCount clear 중인 job 의 수를 가져온다.
func (j *RecoveryJob) GetClearJobCount() (uint64, error) {
	s, err := store.Get(fmt.Sprintf(recoveryJobClearingFormat, j.RecoveryCluster.Id))
	switch {
	case errors.Equal(err, store.ErrNotFoundKey):
		return 0, nil

	case err != nil:
		return 0, errors.UnusableStore(err)
	}

	count, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.Unknown(err)
	}

	return count, nil
}

// IncreaseRunJobCount run 중인 job 의 수를 증가시킨다.
func (j *RecoveryJob) IncreaseRunJobCount(txn store.Txn) error {
	k := fmt.Sprintf(recoveryJobRunningFormat, j.RecoveryCluster.Id)

	count, err := j.GetRunJobCount()
	if err != nil {
		return errors.UnusableStore(err)
	}

	count++

	txn.Put(k, strconv.FormatUint(count, 10))

	return nil
}

// DecreaseRunJobCount run 중인 job 의 수를 감소시킨다.
func (j *RecoveryJob) DecreaseRunJobCount(txn store.Txn) error {
	k := fmt.Sprintf(recoveryJobRunningFormat, j.RecoveryCluster.Id)

	count, err := j.GetRunJobCount()
	if err != nil {
		return err
	}

	if count != 0 {
		count--
	}

	if count == 0 {
		txn.Delete(k, store.DeletePrefix())
	} else {
		txn.Put(k, strconv.FormatUint(count, 10))
	}

	return nil
}

// GetRunJobCount run 중인 job 의 수를 가져온다.
func (j *RecoveryJob) GetRunJobCount() (uint64, error) {
	s, err := store.Get(fmt.Sprintf(recoveryJobRunningFormat, j.RecoveryCluster.Id))
	switch {
	case errors.Equal(err, store.ErrNotFoundKey):
		return 0, nil

	case err != nil:
		return 0, errors.UnusableStore(err)
	}

	count, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.Unknown(err)
	}

	return count, nil
}

func GetFailbackProtectionGroup(pgid, cid uint64) (uint64, error) {
	v, err := store.Get(fmt.Sprintf(recoveryJobFailbackProtectionGroupFormat, pgid, cid))
	switch {
	case errors.Equal(err, store.ErrNotFoundKey):
		return 0, nil

	case err != nil:
		return 0, errors.UnusableStore(err)
	}

	id, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, errors.Unknown(err)
	}

	return id, nil
}

func PutFailbackProtectionGroup(pgid, cid, newPgid uint64) error {
	k := fmt.Sprintf(recoveryJobFailbackProtectionGroupFormat, pgid, cid)

	b, err := json.Marshal(newPgid)
	if err != nil {
		return errors.Unknown(err)
	}

	if err = store.Put(k, string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

func DeleteFailbackProtectionGroup(pgid, cid uint64) error {
	k := fmt.Sprintf(recoveryJobFailbackProtectionGroupFormat, pgid, cid)

	if err := store.Delete(k); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

func DeleteAllFailbackProtectionGroup(pgid uint64) error {
	k := fmt.Sprintf(recoveryJobFailbackProtectionGroupPrefix, pgid)

	if err := store.Delete(k, store.DeletePrefix()); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}
