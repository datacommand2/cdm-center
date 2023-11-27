package errors

import (
	"fmt"
	"github.com/pkg/errors"
)

var (
	// ErrUnusableDatabase 데이터베이스 에러
	ErrUnusableDatabase = New("unusable database")

	// ErrUnusableStore 키-밸류 스토어 에러
	ErrUnusableStore = New("unusable key-value store")

	// ErrUnusableBroker 메시지 브로커 에러
	ErrUnusableBroker = New("unusable broker")

	// ErrUnknown 알 수 없는 에러
	ErrUnknown = New("unknown error")
)

// UnwrapUnusableDatabase error 가 데이터 베이스 에러 일 경우 unwrap 하는 함수
func UnwrapUnusableDatabase(err error) error {
	if !Equal(err, ErrUnusableDatabase) {
		return err
	}

	if e, ok := err.(*Error); ok {
		if dbError, ok := e.value["db_error"].(error); ok {
			return dbError
		}
	}

	return err
}

// UnusableDatabase 데이터베이스 에러
func UnusableDatabase(cause error) error {
	return Wrap(
		ErrUnusableDatabase,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"cause":    fmt.Sprintf("%+v", errors.WithStack(cause)),
			"db_error": cause,
		}),
	)
}

// UnusableStore 키-밸류 스토어 에러
func UnusableStore(cause error) error {
	return Wrap(
		ErrUnusableStore,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"cause": fmt.Sprintf("%+v", errors.WithStack(cause)),
		}),
	)
}

// UnusableBroker 메시지 브로커 에러
func UnusableBroker(cause error) error {
	return Wrap(
		ErrUnusableBroker,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"cause": fmt.Sprintf("%+v", errors.WithStack(cause)),
		}),
	)
}

// Unknown 알 수 없는 에러
func Unknown(cause error) error {
	return Wrap(
		ErrUnknown,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"cause": fmt.Sprintf("%+v", errors.WithStack(cause)),
		}),
	)
}
