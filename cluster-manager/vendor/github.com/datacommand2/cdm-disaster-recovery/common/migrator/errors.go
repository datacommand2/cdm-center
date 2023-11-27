package migrator

import "github.com/datacommand2/cdm-cloud/common/errors"

var (
	// ErrUnknownJobStateCode 잘못된 job 상태 값
	ErrUnknownJobStateCode = errors.New("invalid job state")

	// ErrUnknownResultCode 잘못된 job 결과 값
	ErrUnknownResultCode = errors.New("invalid job result")

	// ErrUnknownJobOperation 잘못된 job operation 값
	ErrUnknownJobOperation = errors.New("invalid job operation")

	// ErrNotFoundJob job 을 찾을 수 없음
	ErrNotFoundJob = errors.New("not found job")

	// ErrNotFoundJobResult job 의 결과 값을 찾을 수 없음
	ErrNotFoundJobResult = errors.New("not found job result")

	// ErrNotFoundJobStatus job 의 해당 결과를 찾을 수 없음
	ErrNotFoundJobStatus = errors.New("not found job status")

	// ErrNotFoundJobLog job 의 log 를 찾을 수 없음
	ErrNotFoundJobLog = errors.New("not found job log")

	// ErrNotFoundTask task 을 찾을 수 없음
	ErrNotFoundTask = errors.New("not found task")

	// ErrNotFoundClearTask clear task 을 찾을 수 없음
	ErrNotFoundClearTask = errors.New("not found clear task")

	// ErrNotFoundTaskResult task 의 결과 값을 찾을 수 없음
	ErrNotFoundTaskResult = errors.New("not found task result")

	// ErrNotSharedTaskType 공유 task 의 유형이 아님
	ErrNotSharedTaskType = errors.New("not shared task type")

	// ErrNotFoundSharedTask 공유 task 를 찾을 수 없음
	ErrNotFoundSharedTask = errors.New("not found shared task")

	// ErrNotFoundJobDetail job 의 detail 을 찾을 수 없음
	ErrNotFoundJobDetail = errors.New("not found job detail")

	// ErrNotFoundKey key 에 대한 정보를 찾을 수 없음
	ErrNotFoundKey = errors.New("not found key")
)

// NotFoundJobDetail job 의 detail 을 찾을 수 없음
func NotFoundJobDetail(jid uint64) error {
	return errors.Wrap(
		ErrNotFoundJobDetail,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"job_id": jid,
		}),
	)
}

// NotFoundSharedTask 공유 task 를 찾을 수 없음
func NotFoundSharedTask(sharedTaskKey string) error {
	return errors.Wrap(
		ErrNotFoundSharedTask,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"shared_task_key": sharedTaskKey,
		}),
	)
}

// NotSharedTaskType 공유 task 의 유형이 아님
func NotSharedTaskType(taskType string) error {
	return errors.Wrap(
		ErrNotSharedTaskType,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"task_type": taskType,
		}),
	)
}

// UnknownJobStateCode 잘못된 job 상태 값
func UnknownJobStateCode(code string) error {
	return errors.Wrap(
		ErrUnknownJobStateCode,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"state_code": code,
		}),
	)
}

// UnknownResultCode 잘못된 job 결과 값
func UnknownResultCode(code string) error {
	return errors.Wrap(
		ErrUnknownResultCode,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"result_code": code,
		}),
	)
}

// UnknownJobOperation 잘못된 job operation 값
func UnknownJobOperation(s string) error {
	return errors.Wrap(
		ErrUnknownJobOperation,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"operation": s,
		}),
	)
}

// NotFoundJob job 을 찾을 수 없음
func NotFoundJob(jid uint64) error {
	return errors.Wrap(
		ErrNotFoundJob,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"job_id": jid,
		}),
	)
}

// NotFoundJobResult job 의 결과 값을 찾을 수 없음
func NotFoundJobResult(jid uint64) error {
	return errors.Wrap(
		ErrNotFoundJobResult,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"job_id": jid,
		}),
	)
}

// NotFoundJobStatus 해당 값을 찾을 수 없음
func NotFoundJobStatus(jid, id uint64, param string) error {
	return errors.Wrap(
		ErrNotFoundJobStatus,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"job_id": jid,
			"id":     id,
			"param":  param,
		}),
	)
}

// NotFoundJobLog job 의 log 를 찾을 수 없음
func NotFoundJobLog(jid, seq uint64) error {
	return errors.Wrap(
		ErrNotFoundJobLog,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"job_id":  jid,
			"log_seq": seq,
		}),
	)
}

// NotFoundTask task 를 찾을 수 없음
func NotFoundTask(jid uint64, tid string) error {
	return errors.Wrap(
		ErrNotFoundTask,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"job_id":  jid,
			"task_id": tid,
		}),
	)
}

// NotFoundClearTask task 를 찾을 수 없음
func NotFoundClearTask(jid uint64, tid string) error {
	return errors.Wrap(
		ErrNotFoundClearTask,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"job_id":  jid,
			"task_id": tid,
		}),
	)
}

// NotFoundTaskResult task 의 결과 값을 찾을 수 없음
func NotFoundTaskResult(jid uint64, tid string) error {
	return errors.Wrap(
		ErrNotFoundTaskResult,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"job_id":  jid,
			"task_id": tid,
		}),
	)
}

// NotFoundKey task 의 결과 값을 찾을 수 없음
func NotFoundKey(key string) error {
	return errors.Wrap(
		ErrNotFoundKey,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"key": key,
		}),
	)
}
