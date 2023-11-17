package sync

import "github.com/datacommand2/cdm-cloud/common/errors"

var (
	// ErrVolumeStatusTimeout 볼륨 상태가 시간내에 변하지 않음
	ErrVolumeStatusTimeout = errors.New("volume status timeout")
	// ErrInstanceStatusTimeout 인스턴스 상태가 시간내에 변하지 않음
	ErrInstanceStatusTimeout = errors.New("instance status timeout")
)

// VolumeStatusTimeout 볼륨 상태가 시간내에 변하지 않음
func VolumeStatusTimeout(status, uuid string) error {
	return errors.Wrap(
		ErrVolumeStatusTimeout,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"status":      status,
			"volume_uuid": uuid,
		}),
	)
}

// InstanceStatusTimeout 인스턴스 상태가 시간내에 변하지 않음
func InstanceStatusTimeout(status, uuid string) error {
	return errors.Wrap(
		ErrInstanceStatusTimeout,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"status":        status,
			"instance_uuid": uuid,
		}),
	)
}
