package client

import (
	"github.com/datacommand2/cdm-cloud/common/errors"
)

var (
	// ErrNotFound 요청된 리소스를 찾을 수 없음
	ErrNotFound = errors.New("not found request resource")
	// ErrUnAuthenticated 인증 되지 않은 사용자
	ErrUnAuthenticated = errors.New("unauthenticated user")
	// ErrBadRequest 잘못된 요청
	ErrBadRequest = errors.New("bad request")
	// ErrUnAuthorized 권한이 없는 사용자
	ErrUnAuthorized = errors.New("unauthorized user")
	// ErrUnAuthorizedUser 필요한 api 요청 권한이 없는 사용자
	ErrUnAuthorizedUser = errors.New("unauthorized user")
	// ErrNotConnected 클러스터에 연결되어 있지 않음
	ErrNotConnected = errors.New("not connected to cluster")
	//ErrConflict 자원 충돌
	ErrConflict = errors.New("conflict error")
	// ErrNotFoundEndpoint 엔드포인트를 찾을 수 없음
	ErrNotFoundEndpoint = errors.New("not found endpoint")
	// ErrUnknown 알 수 없는 에러
	ErrUnknown = errors.New("unknown error")
	// ErrRemoteServerError 리모트 서버 오류
	ErrRemoteServerError = errors.New("cluster server error")
)

// Unknown 알 수 없는 에러
func Unknown(cause error) error {
	return errors.Wrap(
		ErrUnknown,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cause": cause.Error(),
		}),
	)
}

// UnAuthorized 권한이 없는 사용자
func UnAuthorized(cause error) error {
	return errors.Wrap(
		ErrUnAuthorized,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cause": cause.Error(),
		}),
	)
}

// UnAuthorizedUser 필요한 api 요청 권한을 가진 사용자가 아님
func UnAuthorizedUser(projectName, userName, requiredRoleName string) error {
	return errors.Wrap(
		ErrUnAuthorizedUser,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"project_name":       projectName,
			"user_name":          userName,
			"required_role_name": requiredRoleName,
		}),
	)
}

// NotConnected 클러스터에 연결되어 있지 않음
func NotConnected(typeCode string) error {
	return errors.Wrap(
		ErrNotConnected,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"type_code": typeCode,
		}),
	)
}

// BadRequest 잘못된 요청
func BadRequest(cause error) error {
	return errors.Wrap(
		ErrBadRequest,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cause": cause.Error(),
		}),
	)
}

// UnAuthenticated 승인되지 않은 사용자
func UnAuthenticated(cause error) error {
	return errors.Wrap(
		ErrUnAuthenticated,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cause": cause.Error(),
		}),
	)
}

// NotFound 요청된 리소스를 찾을 수 없음
func NotFound(cause error) error {
	return errors.Wrap(
		ErrNotFound,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cause": cause.Error(),
		}),
	)
}

// RemoteServerError 리모트 서버 오류
func RemoteServerError(cause error) error {
	return errors.Wrap(
		ErrRemoteServerError,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cause": cause.Error(),
		}),
	)
}

// NotFoundEndpoint 엔드포인트를 찾을 수 없음
func NotFoundEndpoint(serviceType string) error {
	return errors.Wrap(
		ErrNotFoundEndpoint,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"service_type": serviceType,
		}),
	)
}

// Conflict 자원 충돌
func Conflict(cause error) error {
	return errors.Wrap(
		ErrConflict,
		errors.CallerSkipCount(1),
		errors.WithValue(map[string]interface{}{
			"cause": cause.Error(),
		}),
	)
}
