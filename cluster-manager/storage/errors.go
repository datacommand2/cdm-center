package storage

import "github.com/datacommand2/cdm-cloud/common/errors"

var (
	// ErrNotFoundStorage 스토어에서 스토리지를 찾을 수 없음
	ErrNotFoundStorage = errors.New("not found storage")

	// ErrUnsupportedStorageType 지원하지 않는 스토리지 타입
	ErrUnsupportedStorageType = errors.New("unsupported storage type")

	// ErrNotFoundStorageMetadata 스토어에서 스토리지 메타데이터를 찾을 수 없음
	ErrNotFoundStorageMetadata = errors.New("not found storage metadata")
)

// NotFoundStorage 스토어에서 스토리지를 찾을 수 없음
func NotFoundStorage(cid, sid uint64) error {
	return errors.Wrap(
		ErrNotFoundStorage,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_id":         cid,
			"cluster_storage_id": sid,
		}),
	)
}

// UnsupportedStorageType 지원하지 않는 스토리지 타입
func UnsupportedStorageType(t string) error {
	return errors.Wrap(
		ErrUnsupportedStorageType,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"storage_type": t,
		}),
	)
}

// NotFoundStorageMetadata 스토어에서 스토리지 메타데이터를 찾을 수 없음
func NotFoundStorageMetadata(cid, sid uint64) error {
	return errors.Wrap(
		ErrNotFoundStorageMetadata,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_id":         cid,
			"cluster_storage_id": sid,
		}),
	)
}
