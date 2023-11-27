package mirror

import "github.com/datacommand2/cdm-cloud/common/errors"

var (
	// ErrNotFoundMirrorEnvironment 스토어에서 복제 환경 구성를 찾을 수 없음
	ErrNotFoundMirrorEnvironment = errors.New("not found mirror environment")

	// ErrNotFoundMirrorEnvironmentStatus 스토어에서 복제 환경의 상태를 찾을 수 없음
	ErrNotFoundMirrorEnvironmentStatus = errors.New("not found mirror environment status")

	// ErrNotFoundMirrorEnvironmentOperation 스토어에서 복제 환경의 operation 을 찾을 수 없음
	ErrNotFoundMirrorEnvironmentOperation = errors.New("not found mirror environment operation")

	// ErrUnknownMirrorEnvironmentOperation 복제 환경의 잘못된 operation 값
	ErrUnknownMirrorEnvironmentOperation = errors.New("invalid mirror environment operation value")

	// ErrUnknownMirrorEnvironmentState 복제 환경의 잘못된 상태 값
	ErrUnknownMirrorEnvironmentState = errors.New("invalid mirror environment state value")

	// ErrNotFoundMirrorVolume 스토어에서 볼륨 복제 정보를 찾을 수 없음
	ErrNotFoundMirrorVolume = errors.New("not found mirror volume")

	// ErrNotFoundMirrorVolumeOperation 스토어에서 볼륨 복제의 operation 을 찾을 수 없음
	ErrNotFoundMirrorVolumeOperation = errors.New("not found mirror volume operation")

	// ErrNotFoundMirrorVolumeStatus 스토어에서 볼륨 복제의 status 을 찾을 수 없음
	ErrNotFoundMirrorVolumeStatus = errors.New("not found mirror volume status")

	// ErrUnknownMirrorVolumeOperation 볼륨 복제의 잘못된 operation 값
	ErrUnknownMirrorVolumeOperation = errors.New("invalid mirror volume operation value")

	// ErrUnknownMirrorVolumeState 볼륨 복제의 잘못된 상태 값
	ErrUnknownMirrorVolumeState = errors.New("invalid mirror volume state value")

	// ErrNotFoundMirrorVolumeTargetMetadata 볼륨 복제의 설정 된 target 메타 데이터를 찾을 수 없음
	ErrNotFoundMirrorVolumeTargetMetadata = errors.New("not found mirror volume target metadata")

	// ErrNotFoundMirrorVolumeTargetAgent 볼륨 복제의 설정된 target agent 를 찾을 수 없음
	ErrNotFoundMirrorVolumeTargetAgent = errors.New("not found mirror volume target agent")

	// ErrVolumeExisted 볼륨 복제 환경의 복제 볼륨이 존재 함
	ErrVolumeExisted = errors.New("mirror volume exists")

	// ErrNotFoundSourceVolumeReference 스토어에서 source volume 의 reference 정보를 찾을 수 없음
	ErrNotFoundSourceVolumeReference = errors.New("not found source volume reference")
)

// NotFoundMirrorEnvironment 스토어에서 복제 환경 구성를 찾을 수 없음
func NotFoundMirrorEnvironment(sid, tid uint64) error {
	return errors.Wrap(
		ErrNotFoundMirrorEnvironment,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"source_storage_id": sid,
			"target_storage_id": tid,
		}),
	)
}

// NotFoundMirrorEnvironmentStatus 스토어에서 복제 환경의 상태를 찾을 수 없음
func NotFoundMirrorEnvironmentStatus(sid, tid uint64) error {
	return errors.Wrap(
		ErrNotFoundMirrorEnvironmentStatus,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"source_storage_id": sid,
			"target_storage_id": tid,
		}),
	)
}

// NotFoundMirrorEnvironmentOperation 스토어에서 복제 환경의 operation 을 찾을 수 없음
func NotFoundMirrorEnvironmentOperation(sid, tid uint64) error {
	return errors.Wrap(
		ErrNotFoundMirrorEnvironmentOperation,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"source_storage_id": sid,
			"target_storage_id": tid,
		}),
	)
}

// UnknownMirrorEnvironmentOperation 복제 환경의 잘못된 operation 값
func UnknownMirrorEnvironmentOperation(op string) error {
	return errors.Wrap(
		ErrUnknownMirrorEnvironmentOperation,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"operation": op,
		}),
	)
}

// UnknownMirrorEnvironmentState 복제 환경의 잘못된 상태 값
func UnknownMirrorEnvironmentState(state string) error {
	return errors.Wrap(
		ErrUnknownMirrorEnvironmentState,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"state": state,
		}),
	)
}

// NotFoundMirrorVolume 스토어에서 볼륨 정보를 찾을 수 없음
func NotFoundMirrorVolume(ssid, tsid, svid uint64) error {
	return errors.Wrap(
		ErrNotFoundMirrorVolume,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"source_storage_id": ssid,
			"target_storage_id": tsid,
			"source_volume_id":  svid,
		}),
	)
}

// NotFoundMirrorVolumeOperation 스토어에서 볼륨 복제의 operation 을 찾을 수 없음
func NotFoundMirrorVolumeOperation(ssid, tsid, svid uint64) error {
	return errors.Wrap(
		ErrNotFoundMirrorVolumeOperation,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"source_storage_id": ssid,
			"target_storage_id": tsid,
			"source_volume_id":  svid,
		}),
	)
}

// NotFoundMirrorVolumeStatus 스토어에서 볼륨 복제의 status 를 찾을 수 없음
func NotFoundMirrorVolumeStatus(ssid, tsid, svid uint64) error {
	return errors.Wrap(
		ErrNotFoundMirrorVolumeStatus,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"source_storage_id": ssid,
			"target_storage_id": tsid,
			"source_volume_id":  svid,
		}),
	)
}

// UnknownMirrorVolumeOperation 볼륨 복제의 잘못된 operation 값
func UnknownMirrorVolumeOperation(op string) error {
	return errors.Wrap(
		ErrUnknownMirrorVolumeOperation,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"operation": op,
		}),
	)
}

// UnknownMirrorVolumeState 볼륨 복제의 잘못된 state 값
func UnknownMirrorVolumeState(state string) error {
	return errors.Wrap(
		ErrUnknownMirrorVolumeState,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"state": state,
		}),
	)
}

// NotFoundMirrorVolumeTargetMetadata 스토어에서 볼륨의 target 메타 데이터를 찾을 수 없음
func NotFoundMirrorVolumeTargetMetadata(ssid, tsid, svid uint64) error {
	return errors.Wrap(
		ErrNotFoundMirrorVolumeTargetMetadata,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"source_storage_id": ssid,
			"target_storage_id": tsid,
			"source_volume_id":  svid,
		}),
	)
}

// NotFoundMirrorVolumeTargetAgent 스토어에서 볼륨의 target agent 정보를 찾을 수 없음
func NotFoundMirrorVolumeTargetAgent(ssid, tsid, svid uint64) error {
	return errors.Wrap(
		ErrNotFoundMirrorVolumeTargetAgent,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"source_storage_id": ssid,
			"target_storage_id": tsid,
			"source_volume_id":  svid,
		}),
	)
}

// VolumeExisted 볼륨 복제 환경의 복제 볼륨이 존재 함
func VolumeExisted() error {
	return errors.Wrap(
		ErrVolumeExisted,
		errors.CallerSkipCount(1),
	)
}

// NotFoundSourceVolumeReference 스토어에서 source volume 의 reference 정보를 찾을 수 없음
func NotFoundSourceVolumeReference(ssid, svid uint64) error {
	return errors.Wrap(
		ErrNotFoundSourceVolumeReference,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"source_storage_id": ssid,
			"source_volume_id":  svid,
		}),
	)
}
