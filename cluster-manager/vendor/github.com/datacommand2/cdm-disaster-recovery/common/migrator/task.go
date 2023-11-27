package migrator

import (
	"github.com/datacommand2/cdm-cloud/common/broker"
	"github.com/datacommand2/cdm-cloud/common/errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/datacommand2/cdm-cloud/common/store"
	cloudSync "github.com/datacommand2/cdm-cloud/common/sync"
	"github.com/datacommand2/cdm-disaster-recovery/common/constant"
	"github.com/lestrrat-go/jsref"
	"github.com/lestrrat-go/jsref/provider"

	"context"
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"time"
)

const (
	recoveryJobTaskKeyBase   = "dr.recovery.job/%d/task"
	recoveryJobTaskKeyRegexp = "^dr.recovery.job/\\d+/task/[A-Za-z0-9\\-]+$"
	recoveryJobTaskKeyFormat = "dr.recovery.job/%d/task/%s"

	recoveryJobClearTaskKeyBase   = "dr.recovery.job/%d/clear_task"
	recoveryJobClearTaskKeyRegexp = "^dr.recovery.job/\\d+/clear_task/[A-Za-z0-9\\-]+$"
	recoveryJobClearTaskKeyFormat = "dr.recovery.job/%d/clear_task/%s"

	sharedTaskBuilderPrefixKey                   = "cdm.dr.manager.recovery_job.builder.shared_"
	sharedTaskBuilderTenantTaskMapKey            = "cdm.dr.manager.recovery_job.builder.shared_tenant_task_map"
	sharedTaskBuilderSecurityGroupTaskMapKey     = "cdm.dr.manager.recovery_job.builder.shared_security_group_task_map"
	sharedTaskBuilderSecurityGroupRuleTaskMapKey = "cdm.dr.manager.recovery_job.builder.shared_security_group_rule_task_map"
	sharedTaskBuilderInstanceSpecTaskMapKey      = "cdm.dr.manager.recovery_job.builder.shared_instance_spec_task_map"
	sharedTaskBuilderNetworkTaskMapKey           = "cdm.dr.manager.recovery_job.builder.shared_network_task_map"
	sharedTaskBuilderSubnetTaskMapKey            = "cdm.dr.manager.recovery_job.builder.shared_subnet_task_map"
	sharedTaskBuilderRouterTaskMapKey            = "cdm.dr.manager.recovery_job.builder.shared_router_task_map"
	sharedTaskBuilderKeypairTaskMapKey           = "cdm.dr.manager.recovery_job.builder.shared_keypair_task_map"
)

var (
	sharedTaskMap = map[string]*sharedTask{
		constant.MigrationTaskTypeCreateTenant:            {sharedKey: sharedTaskBuilderTenantTaskMapKey},
		constant.MigrationTaskTypeCreateSecurityGroup:     {sharedKey: sharedTaskBuilderSecurityGroupTaskMapKey},
		constant.MigrationTaskTypeCreateSecurityGroupRule: {sharedKey: sharedTaskBuilderSecurityGroupRuleTaskMapKey},
		constant.MigrationTaskTypeCreateSpec:              {sharedKey: sharedTaskBuilderInstanceSpecTaskMapKey},
		constant.MigrationTaskTypeCreateNetwork:           {sharedKey: sharedTaskBuilderNetworkTaskMapKey},
		constant.MigrationTaskTypeCreateSubnet:            {sharedKey: sharedTaskBuilderSubnetTaskMapKey},
		constant.MigrationTaskTypeCreateRouter:            {sharedKey: sharedTaskBuilderRouterTaskMapKey},
		constant.MigrationTaskTypeCreateKeypair:           {sharedKey: sharedTaskBuilderKeypairTaskMapKey},
	}
	reversedTaskTypeMap = map[string]string{
		constant.MigrationTaskTypeDeleteTenant:        constant.MigrationTaskTypeCreateTenant,
		constant.MigrationTaskTypeDeleteSecurityGroup: constant.MigrationTaskTypeCreateSecurityGroup,
		constant.MigrationTaskTypeDeleteSpec:          constant.MigrationTaskTypeCreateSpec,
		constant.MigrationTaskTypeDeleteNetwork:       constant.MigrationTaskTypeCreateNetwork,
		constant.MigrationTaskTypeDeleteRouter:        constant.MigrationTaskTypeCreateRouter,
		constant.MigrationTaskTypeDeleteKeypair:       constant.MigrationTaskTypeCreateKeypair,
	}
)

/*
constant.MigrationTaskTypeCreateTenant
constant.MigrationTaskTypeCreateSecurityGroup
constant.MigrationTaskTypeCreateSecurityGroupRule
constant.MigrationTaskTypeCreateSpec
constant.MigrationTaskTypeCreateNetwork
constant.MigrationTaskTypeCreateSubnet
constant.MigrationTaskTypeCreateRouter
위에 task 들은 kv store 에 cdm.dr.manager.recovery_job.builder.shared.xxx_task_map 의 형태로 저장됨

taskMap 의 구조는 다음과 같음
{
  "protection_cluster_resource_id": {"recovery_cluster_id": task.uuid}
}

cdm.dr.manager.recovery_job.builder.shared.xxx_task_map/taskID 에는 기존 복구 작업 task input 값,
cdm.dr.manager.recovery_job.builder.shared.xxx_task_map/taskID/status 에는 기존 복구 작업 task 상태값,
cdm.dr.manager.recovery_job.builder.shared.xxx_task_map/taskID/result 에는 기존 복구 작업 task 결과 값이 저장됨
protection_cluster_resource_id, recovery_cluster_id 같은 task 들은 task.uuid, status, result 를 공유 하며
기존에 작업 수행에 대한 결과 및 상태 값이 존재 하고 새로운 작업이 수행될 경우 작업 생성시 cdm.dr.manager.recovery_job.builder.shared.xxx_task_map/taskID 에 저장된 값을 사용
작업이 수행 될 때마다 refcount 가 증가 하며 롤백 실행 시 refcount 가 감소 하며, refcount 값이 0 일 경우 실제 delete task 가 수행 되며 그 외에는 상태 값과 result 값만 변경함
TODO multi attach 되는 볼륨에 대해서도 공유 task로 처리할 필요가 있으며, 재해복구 확정 시 공유 task 는 그대로 kv에 저장 되며, resource 삭제 시 공유 task 를 kv에 지우는 루틴이 필요함
*/

type sharedTask struct {
	sharedKey string
}

// getDependencySharedTask subnet 삭제, security group rule 삭제에 대한 task 가 별도로 존재 하지 않음 으로
// network 삭제, security group 삭제시 subnet, security group rule 에 대한 task 삭제 처리가 들어감
func (s *sharedTask) getDependencySharedTask() *sharedTask {
	switch s.sharedKey {
	case sharedTaskBuilderNetworkTaskMapKey:
		return &sharedTask{sharedKey: sharedTaskBuilderSubnetTaskMapKey}

	case sharedTaskBuilderSecurityGroupTaskMapKey:
		return &sharedTask{sharedKey: sharedTaskBuilderSecurityGroupRuleTaskMapKey}
	}

	return nil
}

func (s *sharedTask) getSharedTaskMap() (map[uint64]map[uint64]string, error) {
	var taskIDMap = make(map[uint64]map[uint64]string)

	val, err := store.Get(s.sharedKey)
	switch {
	case errors.Equal(err, store.ErrNotFoundKey):
		return taskIDMap, nil

	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	if err := json.Unmarshal([]byte(val), &taskIDMap); err != nil {
		return nil, err
	}

	return taskIDMap, nil
}

func (s *sharedTask) getSharedTask(taskID string) (*RecoveryJobTask, error) {
	val, err := store.Get(path.Join(s.sharedKey, taskID))
	switch {
	case errors.Equal(err, store.ErrNotFoundKey):
		return nil, NotFoundSharedTask(path.Join(s.sharedKey, taskID))

	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var t RecoveryJobTask
	if err := json.Unmarshal([]byte(val), &t); err != nil {
		return nil, errors.Unknown(err)
	}

	return &t, nil
}

// 기존 shared task map 을 가져와서 새로운 shared task 정보를 추가한 후 다시 etcd 에 저장하는 함수
func (s *sharedTask) mergeSharedTaskMap(txn store.Txn, m map[uint64]map[uint64]string) error {
	taskMap, err := s.getSharedTaskMap()
	if err != nil {
		return err
	}

	// k : protection cluster 의 shared resource 의 id 값
	// v : recovery cluster id 값을 key, 해당 resource 을 복구하기 위한 task id 값을 value 로 가진 map
	for k, v := range m {
		taskMap[k] = v
	}

	b, err := json.Marshal(&taskMap)
	if err != nil {
		return errors.Unknown(err)
	}

	txn.Put(s.sharedKey, string(b))

	return nil
}

func (s *sharedTask) deleteSharedTask(txn store.Txn, taskID string, recoveryCLusterID uint64) error {
	taskMap, err := s.getSharedTaskMap()
	if err != nil {
		return err
	}

	var exist = false
	for k, v := range taskMap {
		if v2, ok := v[recoveryCLusterID]; ok && v2 == taskID {
			delete(taskMap[k], recoveryCLusterID)

			if len(taskMap[k]) == 0 {
				delete(taskMap, k)
			}

			exist = true
			break
		}
	}

	if !exist {
		return nil
	}

	b, err := json.Marshal(&taskMap)
	if err != nil {
		return errors.Unknown(err)
	}

	txn.Put(s.sharedKey, string(b))
	txn.Delete(path.Join(s.sharedKey, taskID), store.DeletePrefix())

	return nil
}

func (s *sharedTask) deleteDependencyTask(txn store.Txn, parentTaskID string, recoveryClusterID uint64) error {
	// dependency 삭제가 필요한 task 목록
	// - network 삭제 task 시, subnet
	// - securitry group 삭제 task 시, security group rule
	depSharedTask := s.getDependencySharedTask()
	if depSharedTask == nil {
		return nil
	}

	// etcd 에서 dependency task 의 key 값을 가지는 task map 목록을 가져온다.
	// 예를 들어,
	// k : "cdm.dr.manager.recovery_job.builder.shared_security_group_rule_task_map"
	// v : {"169":{"5":"319e0019-fede-45e4-919c-06ba8edff236"}} 인 경우
	taskIDMap, err := depSharedTask.getSharedTaskMap()
	if err != nil {
		return err
	}

	var dependencyTasksMap = make(map[string]bool)
	// k : protection cluster 의 resource id 값
	// v : recovery cluster id 를 key, resource 생성 task id 값을 value 로 가지는 map
	// 예를 들어,
	// {"169":{"5":"319e0019-fede-45e4-919c-06ba8edff236"}} 인 경우
	// 169 : protection cluster 의 security_group_rule id
	// 5: recovery cluster id
	// 319e0019-fede-45e4-919c-06ba8edff236 : security_group_rule(169) 을 생성하기 위한 task id
	for k, v := range taskIDMap {
		// recovery cluster id 값을 key 값으로 가진 resource 생성 task id 값을 가져온다.
		taskID, ok := v[recoveryClusterID]
		if !ok {
			continue
		}

		// etcd 에서 dependency task 의 key 값을 가지는 task 정보 가져온다.
		// 예를 들어,
		// k : "cdm.dr.manager.recovery_job.builder.shared_security_group_rule_task_map/319e0019-fede-45e4-919c-06ba8edff236"
		// v : {"recovery_job_task_key":"","shared_task_key":"cdm.dr.manager.recovery_job.builder.shared_security_group_rule_task_map/e0e3586f-375e-47af-8eb4-7820002aa8f4",
		//"recovery_job_task_id":"e0e3586f-375e-47af-8eb4-7820002aa8f4","type_code":"dr.migration.task.create.security_group_rule",
		// "dependencies":["319e0019-fede-45e4-919c-06ba8edff236","9d277032-33e3-464c-98fd-7ee645ee9ecd"],
		// "input":{"tenant":{"$ref":"319e0019-fede-45e4-919c-06ba8edff236#/tenant"},
		// "security_group":{"$ref":"9d277032-33e3-464c-98fd-7ee645ee9ecd#/security_group"},
		//"security_group_rule":{"id":170,"uuid":"3a1c59fc-ba36-4fd3-b291-49bb4f207fc6","direction":"egress","ether_type":6}}}
		t, err := depSharedTask.getSharedTask(taskID)
		switch {
		// 이미 다른 작업에서 shared task 정보를 삭제한 경우
		case errors.Equal(err, ErrNotFoundSharedTask):
			continue

		case err != nil:
			return err
		}

		// 해당 task 의 dependencies 를 확인하여, parent task 가 해당 task 의 dependency 관계이면,
		// dependencyTasksMap 에 추가한다.
		for _, d := range t.Dependencies {
			if d == parentTaskID && !dependencyTasksMap[taskID] {
				dependencyTasksMap[taskID] = true
				// 특정 resource 의 id(k) 를 key 로 가지는 map 에서, recoveryClusterID 값을 key 로 가지는 map value 를 삭제한다.
				delete(taskIDMap[k], recoveryClusterID)
				if len(taskIDMap[k]) == 0 {
					// 특정 resource 의 id(k) 를 key 로 가지는 map 의 value 가 없는 경우, taskIDMap 을 삭제한다.
					delete(taskIDMap, k)
				}
			}
		}
	}

	for k := range dependencyTasksMap {
		txn.Delete(path.Join(depSharedTask.sharedKey, k), store.DeletePrefix())
	}

	b, err := json.Marshal(&taskIDMap)
	if err != nil {
		return err
	}

	txn.Put(depSharedTask.sharedKey, string(b))

	return nil
}

// GetSharedTaskIDMap 공유 taskIDMap 조회 함수
func GetSharedTaskIDMap(typeCode string) (map[uint64]map[uint64]string, error) {
	st, ok := sharedTaskMap[typeCode]
	if !ok {
		return nil, NotSharedTaskType(typeCode)
	}

	return st.getSharedTaskMap()
}

// MergeSharedTaskIDMap 공유 taskIDMap 머지 함수
func MergeSharedTaskIDMap(txn store.Txn, typeCode string, input map[uint64]map[uint64]string) error {
	st, ok := sharedTaskMap[typeCode]
	if !ok {
		return NotSharedTaskType(typeCode)
	}

	return st.mergeSharedTaskMap(txn, input)
}

// DeleteReverseSharedTask 공유 task 삭제(reverse) 함수
func DeleteReverseSharedTask(txn store.Txn, typeCode, taskID string, recoveryClusterID uint64) error {
	return DeleteSharedTask(txn, reversedTaskTypeMap[typeCode], taskID, recoveryClusterID)
}

// DeleteSharedTask 공유 task 삭제 함수
func DeleteSharedTask(txn store.Txn, typeCode, taskID string, recoveryClusterID uint64) error {
	st, ok := sharedTaskMap[typeCode]
	if !ok {
		return NotSharedTaskType(typeCode)
	}

	if err := st.deleteSharedTask(txn, taskID, recoveryClusterID); err != nil {
		return err
	}

	return st.deleteDependencyTask(txn, taskID, recoveryClusterID)
}

// GetSharedTask 공유 task 조회 함수
func GetSharedTask(typeCode, taskID string) (*RecoveryJobTask, error) {
	st, ok := sharedTaskMap[typeCode]
	if !ok {
		return nil, NotSharedTaskType(typeCode)
	}

	return st.getSharedTask(taskID)
}

// GetReverseSharedTask 공유 task 조회(reverse) 함수
func GetReverseSharedTask(typeCode, taskID string) (*RecoveryJobTask, error) {
	return GetSharedTask(reversedTaskTypeMap[typeCode], taskID)
}

// Message 에러 사유에 대한 메세지 코드 구조체
type Message struct {
	Code     string `json:"code,omitempty"`
	Contents string `json:"contents,omitempty"`
}

// RecoveryJobTask 특정 job 의 특정 task 정보
type RecoveryJobTask struct {
	RecoveryJobTaskKey string      `json:"recovery_job_task_key"`
	RecoveryJobID      uint64      `json:"recovery_job_id,omitempty"`
	SharedTaskKey      string      `json:"shared_task_key"`
	RecoveryJobTaskID  string      `json:"recovery_job_task_id,omitempty"`
	ReverseTaskID      string      `json:"reverse_task_id,omitempty"`
	TypeCode           string      `json:"type_code,omitempty"`
	ResourceID         uint64      `json:"resource_id,omitempty"`
	ResourceName       string      `json:"resource_name,omitempty"`
	Dependencies       []string    `json:"dependencies,omitempty"`
	Input              interface{} `json:"input,omitempty"`
	ForceRetryFlag     bool        `json:"force_retry_flag,omitempty"`
}

// RecoveryJobTaskStatus 특정 job 의 특정 task 의 status
type RecoveryJobTaskStatus struct {
	StateCode   string `json:"state_code,omitempty"`
	StartedAt   int64  `json:"started_at,omitempty"`
	FinishedAt  int64  `json:"finished_at,omitempty"`
	ElapsedTime int64  `json:"-"`
}

// RecoveryJobTaskResult 특정 job 의 특정 task 의 result
type RecoveryJobTaskResult struct {
	ResultCode   string      `json:"result_code,omitempty"`
	FailedReason *Message    `json:"failed_reason,omitempty"`
	Output       interface{} `json:"output,omitempty"`
}

// RecoveryJobTaskConfirm 특정 job 의 특정 task 의 confirm 정보
type RecoveryJobTaskConfirm struct {
	RecoveryJobID uint64 `json:"recovery_job_id,omitempty"`
	ConfirmedAt   int64  `json:"confirmed_at,omitempty"`
}

// GetKey task 의 키 반환 함수
func (t *RecoveryJobTask) GetKey() string {
	return t.RecoveryJobTaskKey
}

// SetSharedTaskKey 공유 task key 설정 함수
func (t *RecoveryJobTask) SetSharedTaskKey() {
	st, ok := sharedTaskMap[t.TypeCode]
	if ok {
		t.SharedTaskKey = path.Join(st.sharedKey, t.RecoveryJobTaskID)
	}
}

// SetRecoveryTaskKey task key 설정 함수
func (t *RecoveryJobTask) SetRecoveryTaskKey() {
	t.RecoveryJobTaskKey = fmt.Sprintf(recoveryJobTaskKeyFormat, t.RecoveryJobID, t.RecoveryJobTaskID)
}

// PutSharedTask 공유 task 를 kv store 에 저장 함수
func (t *RecoveryJobTask) PutSharedTask(txn store.Txn) error {
	b, err := json.Marshal(t)
	if err != nil {
		return errors.Unknown(err)
	}

	txn.Put(t.SharedTaskKey, string(b))

	return nil
}

// Put task 를 kv store 에 저장 하는 함수
func (t *RecoveryJobTask) Put(txn store.Txn) error {
	b, err := json.Marshal(t)
	if err != nil {
		return errors.Unknown(err)
	}

	txn.Put(t.RecoveryJobTaskKey, string(b))

	return nil
}

// IncreaseRefCount 공유 task ref count 증가 함수
func (t *RecoveryJobTask) IncreaseRefCount(txn store.Txn) error {
	if t.SharedTaskKey == "" {
		return nil
	}

	count, err := t.GetRefCount()
	if err != nil {
		return err
	}

	count++

	txn.Put(path.Join(t.SharedTaskKey, "reference"), strconv.FormatUint(count, 10))

	return nil
}

// GetRefCount 공유 task ref count 조회 함수
func (t *RecoveryJobTask) GetRefCount() (uint64, error) {
	if t.SharedTaskKey == "" {
		return 0, nil
	}

	val, err := store.Get(path.Join(t.SharedTaskKey, "reference"))
	switch {
	case errors.Equal(err, store.ErrNotFoundKey):
		return 0, nil

	case err != nil:
		return 0, errors.UnusableStore(err)
	}

	count, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, errors.Unknown(err)
	}

	return count, nil
}

// DecreaseRefCount 공유 task ref count 감소 함수
func (t *RecoveryJobTask) DecreaseRefCount(txn store.Txn) (uint64, error) {
	if t.SharedTaskKey == "" {
		return 0, nil
	}

	count, err := t.GetRefCount()
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}

	count--

	txn.Put(path.Join(t.SharedTaskKey, "reference"), strconv.FormatUint(count, 10))

	return count, nil
}

// Delete task 삭제
func (t *RecoveryJobTask) Delete(txn store.Txn) {
	txn.Delete(t.RecoveryJobTaskKey, store.DeletePrefix())
}

// GetInputData task 의 input data 를 가져온다.
func (t *RecoveryJobTask) GetInputData(out interface{}) error {
	if v, ok := out.(taskInputData); !ok {
		return errors.Unknown(errors.New("parameter type is must be implements an taskInputData"))

	} else if !v.IsTaskType(t.TypeCode) {
		return errors.Unknown(errors.New("parameter type is mismatched on task type"))
	}

	b, err := json.Marshal(t.Input)
	if err != nil {
		return errors.Unknown(err)
	}

	if err := json.Unmarshal(b, out); err != nil {
		return errors.Unknown(err)
	}

	return nil
}

// ResolveInput task 의 input data 의 json ref 를 resolve 하여 input 을 가져온다.
func (t *RecoveryJobTask) ResolveInput(in interface{}) error {
	if v, ok := in.(taskInput); !ok {
		return errors.Unknown(errors.New("parameter type is must be implements an taskInput"))

	} else if !v.IsTaskType(t.TypeCode) {
		return errors.Unknown(errors.New("parameter type is mismatched on task type"))
	}

	getJobTask := func(job *RecoveryJob, taskID string) (*RecoveryJobTask, error) {
		if clearTask, err := job.GetClearTask(taskID); err == nil {
			return clearTask, nil

		} else if !errors.Equal(err, ErrNotFoundClearTask) {
			return nil, err
		}

		return job.GetTask(taskID)
	}

	job, err := GetJob(t.RecoveryJobID)
	if err != nil {
		return err
	}

	res := jsref.New()

	for _, d := range t.Dependencies {
		depTask, err := getJobTask(job, d)
		if err != nil {
			return err
		}

		result, err := depTask.GetResult()
		if err != nil {
			return err
		}

		mp := provider.NewMap()
		if err := mp.Set(d, result.Output); err != nil {
			return errors.Unknown(err)
		}

		if err := res.AddProvider(mp); err != nil {
			return errors.Unknown(err)
		}
	}
	tempTask, err := getJobTask(job, t.RecoveryJobTaskID)
	if err != nil {
		return err
	}

	r, err := res.Resolve(tempTask, "#/input", jsref.WithRecursiveResolution(true))
	if err != nil {
		return errors.Unknown(err)
	}

	b, err := json.Marshal(r)
	if err != nil {
		return errors.Unknown(err)
	}

	if err := json.Unmarshal(b, in); err != nil {
		return errors.Unknown(err)
	}

	return nil
}

type sharedTaskOption struct {
	sharedTaskFlag bool
}

// SharedTaskOptions sharedTaskOption 옵션 함수
type SharedTaskOptions func(o *sharedTaskOption)

// WithSharedTask 공유 task 일 경우 공유 task result, status 등을 조회 하도록 설정 하는 함수
func WithSharedTask(flag bool) SharedTaskOptions {
	return func(o *sharedTaskOption) {
		o.sharedTaskFlag = flag
	}
}

// GetOutput task result 의 output 을 가져온다.
func (t *RecoveryJobTask) GetOutput(out interface{}, options ...SharedTaskOptions) error {
	if v, ok := out.(taskOutput); !ok {
		return errors.Unknown(errors.New("parameter type is must be implements an taskOutput"))

	} else if !v.IsTaskType(t.TypeCode) {
		return errors.Unknown(errors.New("parameter type is mismatched on task type"))
	}

	r, err := t.GetResult(options...)
	if err != nil {
		return err
	}

	b, err := json.Marshal(r.Output)
	if err != nil {
		return errors.Unknown(err)
	}

	if err := json.Unmarshal(b, out); err != nil {
		return errors.Unknown(err)
	}

	return nil
}

// GetStatus task 의 상태를 조회
func (t *RecoveryJobTask) GetStatus(options ...SharedTaskOptions) (*RecoveryJobTaskStatus, error) {
	var option = sharedTaskOption{
		sharedTaskFlag: false,
	}

	for _, o := range options {
		o(&option)
	}

	var key = path.Join(t.RecoveryJobTaskKey, "status")
	if _, ok := sharedTaskMap[t.TypeCode]; ok && option.sharedTaskFlag && t.SharedTaskKey != "" {
		key = path.Join(t.SharedTaskKey, "status")
	}
	v, err := store.Get(key)
	switch {
	case err == store.ErrNotFoundKey:
		return &RecoveryJobTaskStatus{StateCode: constant.MigrationTaskStateCodeWaiting}, nil
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var state RecoveryJobTaskStatus
	if err = json.Unmarshal([]byte(v), &state); err != nil {
		return nil, errors.Unknown(err)
	}

	return &state, nil
}

// SetStatus task status 설정 함수
func (t *RecoveryJobTask) SetStatus(txn store.Txn, state *RecoveryJobTaskStatus) error {
	b, err := json.Marshal(state)
	if err != nil {
		return errors.Unknown(err)
	}

	key := path.Join(t.RecoveryJobTaskKey, "status")
	txn.Put(key, string(b))

	return nil
}

// SetStatusWaiting task 의 상태를 waiting 으로 설정
func (t *RecoveryJobTask) SetStatusWaiting(txn store.Txn, options ...SharedTaskOptions) error {
	status := &RecoveryJobTaskStatus{StateCode: constant.MigrationTaskStateCodeWaiting}
	b1, err := json.Marshal(status)
	if err != nil {
		return errors.Unknown(err)
	}

	var option = sharedTaskOption{
		sharedTaskFlag: false,
	}

	for _, o := range options {
		o(&option)
	}

	var key string
	if _, ok := sharedTaskMap[t.TypeCode]; ok && option.sharedTaskFlag && t.SharedTaskKey != "" {
		key = path.Join(t.SharedTaskKey, "status")
		txn.Put(key, string(b1))
	}

	key = path.Join(t.RecoveryJobTaskKey, "status")
	txn.Put(key, string(b1))

	b2, err := json.Marshal(&RecoveryJobTaskStatusMessage{
		RecoveryJobTaskStatusKey: key,
		TypeCode:                 t.TypeCode,
		RecoveryJobTaskStatus:    status,
	})
	if err != nil {
		return errors.Unknown(err)
	}

	if err = broker.Publish(constant.QueueRecoveryJobTaskStatusMonitor, &broker.Message{Body: b2}); err != nil {
		logger.Errorf("[Task-SetStatusWaiting] Could not publish task status message. Cause: %+v", errors.UnusableBroker(err))
	}

	return nil
}

// SetStatusStarted task 의 상태를 running 으로 설정
func (t *RecoveryJobTask) SetStatusStarted(txn store.Txn, options ...SharedTaskOptions) error {
	now := time.Now().Unix()
	status := &RecoveryJobTaskStatus{StateCode: constant.MigrationTaskStateCodeRunning, StartedAt: now}
	b1, err := json.Marshal(status)
	if err != nil {
		return errors.Unknown(err)
	}

	var option = sharedTaskOption{
		sharedTaskFlag: false,
	}

	for _, o := range options {
		o(&option)
	}

	var key string
	if _, ok := sharedTaskMap[t.TypeCode]; ok && option.sharedTaskFlag && t.SharedTaskKey != "" {
		key = path.Join(t.SharedTaskKey, "status")
		txn.Put(key, string(b1))
	}

	key = path.Join(t.RecoveryJobTaskKey, "status")
	txn.Put(key, string(b1))

	b2, err := json.Marshal(&RecoveryJobTaskStatusMessage{
		RecoveryJobTaskStatusKey: key,
		TypeCode:                 t.TypeCode,
		RecoveryJobTaskStatus:    status,
	})
	if err != nil {
		return errors.Unknown(err)
	}

	if err = broker.Publish(constant.QueueRecoveryJobTaskStatusMonitor, &broker.Message{Body: b2}); err != nil {
		logger.Errorf("[Task-SetStatusStarted] Could not publish task status message. Cause: %+v", errors.UnusableBroker(err))
	}

	return nil
}

// SetStatusFinished task 의 상태를 done 으로 설정
func (t *RecoveryJobTask) SetStatusFinished(txn store.Txn, options ...SharedTaskOptions) error {
	s, err := t.GetStatus()
	if err != nil {
		return err
	}

	s.StateCode = constant.MigrationTaskStateCodeDone
	s.FinishedAt = time.Now().Unix()

	b, err := json.Marshal(s)
	if err != nil {
		return errors.Unknown(err)
	}

	var option = sharedTaskOption{
		sharedTaskFlag: false,
	}

	for _, o := range options {
		o(&option)
	}

	var key string
	if _, ok := sharedTaskMap[t.TypeCode]; ok && option.sharedTaskFlag && t.SharedTaskKey != "" {
		key = path.Join(t.SharedTaskKey, "status")
		txn.Put(key, string(b))
	}

	key = path.Join(t.RecoveryJobTaskKey, "status")
	txn.Put(key, string(b))

	b2, err := json.Marshal(&RecoveryJobTaskStatusMessage{
		RecoveryJobTaskStatusKey: key,
		TypeCode:                 t.TypeCode,
		RecoveryJobTaskStatus:    s})
	if err != nil {
		return errors.Unknown(err)
	}

	if err = broker.Publish(constant.QueueRecoveryJobTaskStatusMonitor, &broker.Message{Body: b2}); err != nil {
		logger.Errorf("[Task-SetStatusFinished] Could not publish task status message. Cause: %+v", errors.UnusableBroker(err))
	}

	return nil
}

// DeleteResult task 의 result 를 삭제
func (t *RecoveryJobTask) DeleteResult(txn store.Txn, options ...SharedTaskOptions) error {
	var option = sharedTaskOption{
		sharedTaskFlag: false,
	}

	for _, o := range options {
		o(&option)
	}

	var key string
	if _, ok := sharedTaskMap[t.TypeCode]; ok && option.sharedTaskFlag && t.SharedTaskKey != "" {
		key = path.Join(t.SharedTaskKey, "result")
		txn.Delete(key, store.DeletePrefix())
	}

	key = path.Join(t.RecoveryJobTaskKey, "result")
	txn.Delete(key, store.DeletePrefix())

	b, err := json.Marshal(RecoveryJobTaskResultMessage{
		RecoveryJobTaskResultKey: key,
		TypeCode:                 t.TypeCode,
		IsDeleted:                true,
	})
	if err != nil {
		return errors.Unknown(err)
	}

	if err = broker.Publish(constant.QueueRecoveryJobTaskResultMonitor, &broker.Message{Body: b}); err != nil {
		logger.Errorf("[Task-DeleteResult] Could not publish task result delete message. Cause: %+v", errors.UnusableBroker(err))
	}

	return nil
}

// SetResult task 의 result 설정
func (t *RecoveryJobTask) SetResult(txn store.Txn, result *RecoveryJobTaskResult) error {
	b, err := json.Marshal(result)
	if err != nil {
		return errors.Unknown(err)
	}

	k := path.Join(t.RecoveryJobTaskKey, "result")
	txn.Put(k, string(b))

	return nil
}

// GetResult task 의 result 조회
func (t *RecoveryJobTask) GetResult(options ...SharedTaskOptions) (*RecoveryJobTaskResult, error) {
	var option = sharedTaskOption{
		sharedTaskFlag: false,
	}

	for _, o := range options {
		o(&option)
	}

	var k = path.Join(t.RecoveryJobTaskKey, "result")
	if _, ok := sharedTaskMap[t.TypeCode]; ok && option.sharedTaskFlag && t.SharedTaskKey != "" {
		k = path.Join(t.SharedTaskKey, "result")
	}

	v, err := store.Get(k)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, NotFoundTaskResult(t.RecoveryJobID, t.RecoveryJobTaskID)
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var result RecoveryJobTaskResult
	if err = json.Unmarshal([]byte(v), &result); err != nil {
		return nil, errors.Unknown(err)
	}

	return &result, nil
}

// SetResultSuccess 성공한 task 의 result 설정
func (t *RecoveryJobTask) SetResultSuccess(txn store.Txn, output interface{}, options ...SharedTaskOptions) error {
	result := &RecoveryJobTaskResult{
		ResultCode: constant.MigrationTaskResultCodeSuccess,
		Output:     output,
	}
	b1, err := json.Marshal(result)
	if err != nil {
		return errors.Unknown(err)
	}

	var option = sharedTaskOption{
		sharedTaskFlag: false,
	}

	for _, o := range options {
		o(&option)
	}

	var key string
	if _, ok := sharedTaskMap[t.TypeCode]; ok && option.sharedTaskFlag && t.SharedTaskKey != "" {
		key = path.Join(t.SharedTaskKey, "result")
		txn.Put(key, string(b1))
	}

	key = path.Join(t.RecoveryJobTaskKey, "result")
	txn.Put(key, string(b1))

	b2, err := json.Marshal(&RecoveryJobTaskResultMessage{
		RecoveryJobTaskResultKey: key,
		TypeCode:                 t.TypeCode,
		RecoveryJobTaskResult:    result,
	})
	if err != nil {
		return errors.Unknown(err)
	}

	if err = broker.Publish(constant.QueueRecoveryJobTaskResultMonitor, &broker.Message{Body: b2}); err != nil {
		logger.Errorf("[Task-SetResultSuccess] Could not publish task result message. Cause: %+v", errors.UnusableBroker(err))
	}

	return nil
}

// SetResultExcepted 제외된 task 의 result 설정
func (t *RecoveryJobTask) SetResultExcepted(txn store.Txn, reasonCode string, reasonContents interface{}) error {
	b, err := json.Marshal(reasonContents)
	if err != nil {
		return errors.Unknown(err)
	}

	result := &RecoveryJobTaskResult{
		ResultCode: constant.MigrationTaskResultCodeExcepted,
		FailedReason: &Message{
			Code:     reasonCode,
			Contents: string(b),
		},
	}

	b, err = json.Marshal(result)
	if err != nil {
		return errors.Unknown(err)
	}

	k := path.Join(t.RecoveryJobTaskKey, "result")
	txn.Put(k, string(b))

	b2, err := json.Marshal(&RecoveryJobTaskResultMessage{
		RecoveryJobTaskResultKey: k,
		TypeCode:                 t.TypeCode,
		RecoveryJobTaskResult:    result,
	})
	if err != nil {
		return errors.Unknown(err)
	}

	if err = broker.Publish(constant.QueueRecoveryJobTaskResultMonitor, &broker.Message{Body: b2}); err != nil {
		logger.Errorf("[Task-SetResultExcepted] Could not publish task result message. Cause: %+v", errors.UnusableBroker(err))
	}

	return nil
}

// SetResultIgnored 무시된 task 의 result 설정
func (t *RecoveryJobTask) SetResultIgnored(txn store.Txn, reasonCode string, reasonContents interface{}, options ...SharedTaskOptions) error {
	b, err := json.Marshal(reasonContents)
	if err != nil {
		return errors.Unknown(err)
	}

	result := &RecoveryJobTaskResult{
		ResultCode: constant.MigrationTaskResultCodeIgnored,
		FailedReason: &Message{
			Code:     reasonCode,
			Contents: string(b),
		},
	}

	b, err = json.Marshal(result)
	if err != nil {
		return errors.Unknown(err)
	}

	var option = sharedTaskOption{
		sharedTaskFlag: false,
	}

	for _, o := range options {
		o(&option)
	}

	var key string
	if _, ok := sharedTaskMap[t.TypeCode]; ok && option.sharedTaskFlag && t.SharedTaskKey != "" {
		key = path.Join(t.SharedTaskKey, "result")
		txn.Put(key, string(b))
	}

	key = path.Join(t.RecoveryJobTaskKey, "result")
	txn.Put(key, string(b))

	b2, err := json.Marshal(RecoveryJobTaskResultMessage{
		RecoveryJobTaskResultKey: key,
		TypeCode:                 t.TypeCode,
		RecoveryJobTaskResult:    result,
	})
	if err != nil {
		return errors.Unknown(err)
	}

	if err = broker.Publish(constant.QueueRecoveryJobTaskResultMonitor, &broker.Message{Body: b2}); err != nil {
		logger.Errorf("[Task-SetResultIgnored] Could not publish task result message. Cause: %+v", errors.UnusableBroker(err))
	}

	return nil
}

// SetResultCanceled 취소된 task 의 result 설정
func (t *RecoveryJobTask) SetResultCanceled(txn store.Txn, reasonCode string) error {
	result := &RecoveryJobTaskResult{
		ResultCode: constant.MigrationTaskResultCodeCanceled,
		FailedReason: &Message{
			Code: reasonCode,
		},
	}

	b, err := json.Marshal(result)
	if err != nil {
		return errors.Unknown(err)
	}

	k := path.Join(t.RecoveryJobTaskKey, "result")
	txn.Put(k, string(b))

	b2, err := json.Marshal(&RecoveryJobTaskResultMessage{
		RecoveryJobTaskResultKey: k,
		TypeCode:                 t.TypeCode,
		RecoveryJobTaskResult:    result,
	})
	if err != nil {
		return errors.Unknown(err)
	}

	if err = broker.Publish(constant.QueueRecoveryJobTaskResultMonitor, &broker.Message{Body: b2}); err != nil {
		logger.Errorf("[Task-SetResultCanceled] Could not publish task result message. Cause: %+v", errors.UnusableBroker(err))
	}

	return nil
}

// SetResultFailed 실패한 task 의 result 설정
func (t *RecoveryJobTask) SetResultFailed(txn store.Txn, reasonCode, reasonContents string, options ...SharedTaskOptions) error {
	result := &RecoveryJobTaskResult{
		ResultCode: constant.MigrationTaskResultCodeFailed,
		FailedReason: &Message{
			Code:     reasonCode,
			Contents: reasonContents,
		},
	}

	b, err := json.Marshal(result)
	if err != nil {
		return errors.Unknown(err)
	}

	var option = sharedTaskOption{
		sharedTaskFlag: false,
	}

	for _, o := range options {
		o(&option)
	}

	var key string
	if _, ok := sharedTaskMap[t.TypeCode]; ok && option.sharedTaskFlag && t.SharedTaskKey != "" {
		key = path.Join(t.SharedTaskKey, "result")
		txn.Put(key, string(b))
	}

	key = path.Join(t.RecoveryJobTaskKey, "result")
	txn.Put(key, string(b))

	b2, err := json.Marshal(&RecoveryJobTaskResultMessage{
		RecoveryJobTaskResultKey: key,
		TypeCode:                 t.TypeCode,
		RecoveryJobTaskResult:    result,
	})
	if err != nil {
		return errors.Unknown(err)
	}

	if err = broker.Publish(constant.QueueRecoveryJobTaskResultMonitor, &broker.Message{Body: b2}); err != nil {
		logger.Errorf("[Task-SetResultFailed] Could not publish task result message. Cause: %+v", errors.UnusableBroker(err))
	}

	return nil
}

// SetResultUnknown 알 수 없는 task 의 result 설정
func (t *RecoveryJobTask) SetResultUnknown(txn store.Txn, reasonCode string) error {
	result := &RecoveryJobTaskResult{
		ResultCode: constant.MigrationTaskResultCodeUnknown,
		FailedReason: &Message{
			Code: reasonCode,
		},
	}

	b, err := json.Marshal(result)
	if err != nil {
		return errors.Unknown(err)
	}

	k := path.Join(t.RecoveryJobTaskKey, "result")
	txn.Put(k, string(b))

	b2, err := json.Marshal(RecoveryJobTaskResultMessage{
		RecoveryJobTaskResultKey: k,
		TypeCode:                 t.TypeCode,
		RecoveryJobTaskResult:    result,
	})
	if err != nil {
		return errors.Unknown(err)
	}

	if err = broker.Publish(constant.QueueRecoveryJobTaskResultMonitor, &broker.Message{Body: b2}); err != nil {
		logger.Errorf("[Task-SetResultUnknown] Could not publish task result message. Cause: %+v", errors.UnusableBroker(err))
	}

	return nil
}

// Lock 공유 task 의 중복 실행을 막기 위한 함수
func (t *RecoveryJobTask) Lock() (func(), error) {
	if t.SharedTaskKey == "" {
		return func() {}, nil
	}

	mutex, err := cloudSync.Lock(context.Background(), path.Join(t.SharedTaskKey, "lock"))
	if err != nil {
		return nil, err
	}

	return func() {
		if err := mutex.Unlock(context.Background()); err != nil {
			logger.Errorf("Could not unlock(%s). Cause:%+v", t.SharedTaskKey, err)
		}
	}, nil
}

// GetConfirm task 의 confirm 정보를 조회
func (t *RecoveryJobTask) GetConfirm() (*RecoveryJobTaskConfirm, error) {
	var key = path.Join(t.SharedTaskKey, "confirm")
	v, err := store.Get(key)
	switch {
	case err == store.ErrNotFoundKey:
		return nil, nil
	case err != nil:
		return nil, errors.UnusableStore(err)
	}

	var confirm RecoveryJobTaskConfirm
	if err = json.Unmarshal([]byte(v), &confirm); err != nil {
		return nil, errors.Unknown(err)
	}

	return &confirm, nil
}

// SetConfirm task confirm 설정 함수
func (t *RecoveryJobTask) SetConfirm(confirm *RecoveryJobTaskConfirm) error {
	b, err := json.Marshal(confirm)
	if err != nil {
		return errors.Unknown(err)
	}

	key := path.Join(t.SharedTaskKey, "confirm")

	if err = store.Put(key, string(b)); err != nil {
		return errors.UnusableStore(err)
	}

	return nil
}

// GetAllSharedTaskMapToString string 으로 SharedTaskMap 모든 데이터를 조회
func GetAllSharedTaskMapToString() (string, error) {
	rsp, err := store.GetAll(sharedTaskBuilderPrefixKey)
	switch {
	case err == store.ErrNotFoundKey:
		logger.Warnf("[task-GetAllSharedTaskMapToString] Not found key: shared task builder.")
		return "", nil

	case err != nil:
		logger.Errorf("[task-GetAllSharedTaskMapToString] Could not get shared task builder prefix data. Cause: %v", err)
		return "", errors.UnusableStore(err)
	}

	return rsp, nil
}

func deleteSharedTaskByClusterId(cid uint64, typeCode string) {
	st, _ := sharedTaskMap[typeCode]
	taskMap, err := st.getSharedTaskMap()
	if err != nil {
		logger.Warnf("[deleteSharedTaskByClusterId] Could not get shared task(%s) map. Cause: %+v", typeCode, err)
		return
	}

	if taskMap == nil {
		return
	}

	for _, val := range taskMap {
		if taskID, ok := val[cid]; ok {
			if err = store.Transaction(func(txn store.Txn) error {
				return DeleteSharedTask(txn, typeCode, taskID, cid)
			}); err != nil {
				logger.Warnf("[deleteSharedTaskByClusterId] Could not delete shared task(%s:%s) cluster(%d). Cause: %+v", typeCode, taskID, cid, err)
				continue
			}
			logger.Infof("[deleteSharedTaskByClusterId] Done - deleted shared task(%s:%s) cluster(%d)", typeCode, taskID, cid)
		}
	}

}

func ClearSharedTaskByClusterId(cid uint64) {
	deleteSharedTaskByClusterId(cid, constant.MigrationTaskTypeCreateSpec)
	deleteSharedTaskByClusterId(cid, constant.MigrationTaskTypeCreateKeypair)
	deleteSharedTaskByClusterId(cid, constant.MigrationTaskTypeCreateRouter)
	deleteSharedTaskByClusterId(cid, constant.MigrationTaskTypeCreateSubnet)
	deleteSharedTaskByClusterId(cid, constant.MigrationTaskTypeCreateNetwork)
	deleteSharedTaskByClusterId(cid, constant.MigrationTaskTypeCreateSecurityGroupRule)
	deleteSharedTaskByClusterId(cid, constant.MigrationTaskTypeCreateSecurityGroup)
	deleteSharedTaskByClusterId(cid, constant.MigrationTaskTypeCreateTenant)
}

func ClearSharedTaskById(id uint64, typeCode string) map[uint64]map[uint64]string {
	var (
		err     error
		st      *sharedTask
		taskMap = make(map[uint64]map[uint64]string)
	)

	st, _ = sharedTaskMap[typeCode]
	taskMap, err = st.getSharedTaskMap()
	if err != nil {
		logger.Warnf("[ClearSharedTaskById] Could not get shared task(%s) map. Cause: %+v", typeCode, err)
		return nil
	}

	if taskMap == nil {
		return nil
	}

	// 해당 id가 보호대상으로 등록된 이하 task 모두 삭제
	if tasks, ok := taskMap[id]; ok {
		for c, taskID := range tasks {
			if err = store.Transaction(func(txn store.Txn) error {
				return DeleteSharedTask(txn, typeCode, taskID, c)
			}); err != nil {
				logger.Warnf("[ClearSharedTaskById] Could not delete shared task(%s:%s) cluster(%d) id(%s). Cause: %+v", typeCode, taskID, c, id, err)
				continue
			}
			logger.Infof("[ClearSharedTaskById] Done - deleted shared task(%s:%s) cluster(%d) id(%s)", typeCode, taskID, c, id)
			delete(tasks, c)
		}
		if len(tasks) == 0 {
			delete(taskMap, id)
		}
	}

	return taskMap
}

// ClearSharedTask 일부 항목이 삭제되었을 떄 사용
func ClearSharedTask(cid, id uint64, delUuid, typeCode string) {
	var (
		taskMap = make(map[uint64]map[uint64]string)
		taskIDs []string
	)

	taskMap = ClearSharedTaskById(id, typeCode)
	if taskMap == nil {
		return
	}

	// 해당 cid 가 복구대상으로 등록된 taskMap
	for _, val := range taskMap {
		if t, ok := val[cid]; ok {
			taskIDs = append(taskIDs, t)
		}
	}

	for _, taskID := range taskIDs {
		task, err := GetSharedTask(typeCode, taskID)
		if err != nil {
			logger.Warnf("[ClearSharedTask] Could not get shared task(%s:%s). Cause: %+v", typeCode, taskID, err)
			continue
		}

		// result 의 out 에 해당 uuid 가 등록된 task 삭제
		taskUUID, err := getOutputUUID(task)
		if err != nil {
			logger.Warnf("[ClearSharedTask] Could not get task output: task(%s:%s). Cause: %+v", typeCode, taskID, err)
			continue
		}

		if taskUUID == delUuid {
			store.Transaction(func(txn store.Txn) error {
				if err := DeleteSharedTask(txn, task.TypeCode, taskID, cid); err != nil {
					logger.Warnf("[ClearSharedTask] Could not delete shared task(%s:%s) cluster(%d) uuid(%s). Cause: %+v", task.TypeCode, taskID, cid, delUuid, err)
				} else {
					logger.Infof("[ClearSharedTask] Done - deleted shared task(%s:%s) cluster(%d) uuid(%s)", task.TypeCode, taskID, cid, delUuid)
				}
				return nil
			})
			break
		}
	}
}

func getOutputUUID(task *RecoveryJobTask) (string, error) {
	var uuid string

	switch task.TypeCode {
	case constant.MigrationTaskTypeCreateTenant:
		var out TenantCreateTaskOutput
		if err := task.GetOutput(&out, WithSharedTask(true)); err != nil {
			return "", err
		}
		uuid = out.Tenant.Uuid

	case constant.MigrationTaskTypeCreateSecurityGroup:
		var out SecurityGroupCreateTaskOutput
		if err := task.GetOutput(&out, WithSharedTask(true)); err != nil {
			return "", err
		}
		uuid = out.SecurityGroup.Uuid

	case constant.MigrationTaskTypeCreateSecurityGroupRule:
		var out SecurityGroupRuleCreateTaskOutput
		if err := task.GetOutput(&out, WithSharedTask(true)); err != nil {
			return "", err
		}
		uuid = out.SecurityGroupRule.Uuid

	case constant.MigrationTaskTypeCreateNetwork:
		var out NetworkCreateTaskOutput
		if err := task.GetOutput(&out, WithSharedTask(true)); err != nil {
			return "", err
		}
		uuid = out.Network.Uuid

	case constant.MigrationTaskTypeCreateSubnet:
		var out SubnetCreateTaskOutput
		if err := task.GetOutput(&out, WithSharedTask(true)); err != nil {
			return "", err
		}
		uuid = out.Subnet.Uuid

	case constant.MigrationTaskTypeCreateRouter:
		var out RouterCreateTaskOutput
		if err := task.GetOutput(&out, WithSharedTask(true)); err != nil {
			return "", err
		}
		uuid = out.Router.Uuid

	case constant.MigrationTaskTypeCreateSpec:
		var out InstanceSpecCreateTaskOutput
		if err := task.GetOutput(&out, WithSharedTask(true)); err != nil {
			return "", err
		}
		uuid = out.Spec.Uuid

	case constant.MigrationTaskTypeCreateKeypair:
		var out KeypairCreateTaskOutput
		if err := task.GetOutput(&out, WithSharedTask(true)); err != nil {
			return "", err
		}
		uuid = out.Keypair.Name
	}

	return uuid, nil
}
