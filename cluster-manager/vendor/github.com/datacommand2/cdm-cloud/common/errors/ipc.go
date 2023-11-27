package errors

import (
	"encoding/json"
	microErr "github.com/micro/go-micro/v2/errors"
)

// ErrIPCFailed IPC Failed
var ErrIPCFailed = New("inter process communication failed")

// ErrIPCSuccess IPC Success
var ErrIPCSuccess = New("inter process communication success")

// IPCStatusCodes
const (
	IPCStatusUnknown int32 = 0 // ipc status code 를 알 수 없음

	IPCStatusOK                   int32 = 200 // RFC 7231, 6.3.1
	IPCStatusCreated              int32 = 201 // RFC 7231, 6.3.2
	IPCStatusAccepted             int32 = 202 // RFC 7231, 6.3.3
	IPCStatusNonAuthoritativeInfo int32 = 203 // RFC 7231, 6.3.4
	IPCStatusNoContent            int32 = 204 // RFC 7231, 6.3.5
	IPCStatusResetContent         int32 = 205 // RFC 7231, 6.3.6
	IPCStatusPartialContent       int32 = 206 // RFC 7233, 4.1
	IPCStatusMultiStatus          int32 = 207 // RFC 4918, 11.1
	IPCStatusAlreadyReported      int32 = 208 // RFC 5842, 7.1
	IPCStatusIMUsed               int32 = 226 // RFC 3229, 10.4.1

	IPCStatusBadRequest                   int32 = 400 // RFC 7231, 6.5.1
	IPCStatusUnauthenticated              int32 = 401 // RFC 7235, 3.1
	IPCStatusPaymentRequired              int32 = 402 // RFC 7231, 6.5.2
	IPCStatusUnauthorized                 int32 = 403 // RFC 7231, 6.5.3
	IPCStatusNotFound                     int32 = 404 // RFC 7231, 6.5.4
	IPCStatusMethodNotAllowed             int32 = 405 // RFC 7231, 6.5.5
	IPCStatusNotAcceptable                int32 = 406 // RFC 7231, 6.5.6
	IPCStatusProxyAuthRequired            int32 = 407 // RFC 7235, 3.2
	IPCStatusRequestTimeout               int32 = 408 // RFC 7231, 6.5.7
	IPCStatusConflict                     int32 = 409 // RFC 7231, 6.5.8
	IPCStatusGone                         int32 = 410 // RFC 7231, 6.5.9
	IPCStatusLengthRequired               int32 = 411 // RFC 7231, 6.5.10
	IPCStatusPreconditionFailed           int32 = 412 // RFC 7232, 4.2
	IPCStatusRequestEntityTooLarge        int32 = 413 // RFC 7231, 6.5.11
	IPCStatusRequestURITooLong            int32 = 414 // RFC 7231, 6.5.12
	IPCStatusUnsupportedMediaType         int32 = 415 // RFC 7231, 6.5.13
	IPCStatusRequestedRangeNotSatisfiable int32 = 416 // RFC 7233, 4.4
	IPCStatusExpectationFailed            int32 = 417 // RFC 7231, 6.5.14
	IPCStatusTeapot                       int32 = 418 // RFC 7168, 2.3.3
	IPCStatusMisdirectedRequest           int32 = 421 // RFC 7540, 9.1.2
	IPCStatusUnprocessableEntity          int32 = 422 // RFC 4918, 11.2
	IPCStatusLocked                       int32 = 423 // RFC 4918, 11.3
	IPCStatusFailedDependency             int32 = 424 // RFC 4918, 11.4
	IPCStatusTooEarly                     int32 = 425 // RFC 8470, 5.2.
	IPCStatusUpgradeRequired              int32 = 426 // RFC 7231, 6.5.15
	IPCStatusPreconditionRequired         int32 = 428 // RFC 6585, 3
	IPCStatusTooManyRequests              int32 = 429 // RFC 6585, 4
	IPCStatusRequestHeaderFieldsTooLarge  int32 = 431 // RFC 6585, 5
	IPCStatusUnavailableForLegalReasons   int32 = 451 // RFC 7725, 3

	IPCStatusInternalServerError           int32 = 500 // RFC 7231, 6.6.1
	IPCStatusNotImplemented                int32 = 501 // RFC 7231, 6.6.2
	IPCStatusBadGateway                    int32 = 502 // RFC 7231, 6.6.3
	IPCStatusServiceUnavailable            int32 = 503 // RFC 7231, 6.6.4
	IPCStatusGatewayTimeout                int32 = 504 // RFC 7231, 6.6.5
	IPCStatusHTTPVersionNotSupported       int32 = 505 // RFC 7231, 6.6.6
	IPCStatusVariantAlsoNegotiates         int32 = 506 // RFC 2295, 8.1
	IPCStatusInsufficientStorage           int32 = 507 // RFC 4918, 11.5
	IPCStatusLoopDetected                  int32 = 508 // RFC 5842, 7.2
	IPCStatusNotExtended                   int32 = 510 // RFC 2774, 7
	IPCStatusNetworkAuthenticationRequired int32 = 511 // RFC 6585, 6
)

// IsIPCFailed IPC Failed 여부 확인
func IsIPCFailed(cause error) bool {
	return cause != nil && microErr.Parse(cause.Error()).Code >= 400
}

// IsIPCSuccess IPC Success 여부 확인
func IsIPCSuccess(cause error) bool {
	return cause != nil && microErr.Parse(cause.Error()).Code >= 200 && microErr.Parse(cause.Error()).Code < 300
}

// IPCFailed IPC Failed
func IPCFailed(cause error) error {
	err := microErr.Parse(cause.Error())

	var msg Message
	if json.Unmarshal([]byte(err.Detail), &msg) != nil || msg.Code == "" {
		msg.Code = err.Detail
	}

	return Wrap(
		ErrIPCFailed,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"code":    err.Code,
			"message": &msg,
		}),
	)
}

// IPCSuccess IPC Success
func IPCSuccess(cause error) error {
	err := microErr.Parse(cause.Error())

	var msg Message
	if json.Unmarshal([]byte(err.Detail), &msg) != nil || msg.Code == "" {
		msg.Code = err.Detail
	}

	return Wrap(
		ErrIPCSuccess,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"code":    err.Code,
			"message": &msg,
		}),
	)
}

// GetIPCMessage 는 IPCFailed error 의 message 를 반환하는 함수
func GetIPCMessage(err error) *Message {
	if err == nil {
		return nil
	}

	var e *Error
	switch {
	case Equal(err, ErrIPCFailed):
		e = err.(*Error)

	case IsIPCFailed(err):
		e = IPCFailed(err).(*Error)

	default:
		return nil
	}

	if msg, ok := e.value["message"]; ok {
		if m, ok := msg.(*Message); ok {
			return m
		}
	}

	return nil
}

// GetIPCStatusCode 은 IPCFailed error 의 code 를 반환하는 함수
func GetIPCStatusCode(err error) int32 {
	if err == nil {
		return IPCStatusUnknown
	}

	var e *Error
	switch {
	case Equal(err, ErrIPCFailed), Equal(err, ErrIPCSuccess):
		e = err.(*Error)

	case IsIPCFailed(err):
		e = IPCFailed(err).(*Error)

	case IsIPCSuccess(err):
		e = IPCSuccess(err).(*Error)

	default:
		return IPCStatusUnknown
	}

	if code, ok := e.value["code"]; ok {
		if c, ok := code.(int32); ok {
			return c
		}
	}

	return IPCStatusUnknown
}

// IsIPCStatusSuccess 은 IPCFailed error 의 code 가 success 인지 확인하는 함수
func IsIPCStatusSuccess(err error) bool {
	code := GetIPCStatusCode(err)
	return code >= 200 && code < 300
}

// IsIPCStatusClientErrors 은 IPCFailed error 의 code 가 client error 인지 확인하는 함수
func IsIPCStatusClientErrors(err error) bool {
	code := GetIPCStatusCode(err)
	return code >= 400 && code < 500
}

// IsIPCStatusServerErrors 은 IPCFailed error 의 code 가 server error 인지 확인하는 함수
func IsIPCStatusServerErrors(err error) bool {
	code := GetIPCStatusCode(err)
	return code >= 500 && code < 600
}
