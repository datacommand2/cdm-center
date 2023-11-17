package volume

import "github.com/datacommand2/cdm-cloud/common/errors"

var (
	// ErrNotFoundVolume 스토어에서 볼륨 정보를 찾을 수 없음
	ErrNotFoundVolume = errors.New("not found volume")

	// ErrNotFoundVolumeMetadata 스토어에서 볼륨 메타데이터를 찾을 수 없음
	ErrNotFoundVolumeMetadata = errors.New("not found volume metadata")
)

// NotFoundVolume 스토어에서 스토리지를 찾을 수 없음
func NotFoundVolume(cid, vid uint64) error {
	return errors.Wrap(
		ErrNotFoundVolume,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_id":        cid,
			"cluster_volume_id": vid,
		}),
	)
}

// NotFoundVolumeMetadata 스토어에서 볼륨 메타데이터를 찾을 수 없음
func NotFoundVolumeMetadata(cid, sid uint64) error {
	return errors.Wrap(
		ErrNotFoundVolumeMetadata,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_id":         cid,
			"cluster_storage_id": sid,
		}),
	)
}
