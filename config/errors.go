package config

import "github.com/datacommand2/cdm-cloud/common/errors"

var (
	// ErrNotFoundConfig 스토어에서 config 를 찾을 수 없음
	ErrNotFoundConfig = errors.New("not found config")
)

// NotFoundConfig 스토어에서 config 를 찾을 수 없음
func NotFoundConfig(cid uint64) error {
	return errors.Wrap(
		ErrNotFoundConfig,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cluster_id": cid,
		}),
	)
}
