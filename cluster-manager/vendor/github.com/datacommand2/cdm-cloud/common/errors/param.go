package errors

import (
	"regexp"
)

var (
	// ErrRequiredParameter 필수 파라메터 누락
	ErrRequiredParameter = New("required parameter")

	// ErrUnchangeableParameter 수정이 불가능한 파라메터
	ErrUnchangeableParameter = New("unchangeable parameter")

	// ErrConflictParameterValue 충돌(중복)이 발생한 파라메터 값
	ErrConflictParameterValue = New("conflict parameter value")

	// ErrInvalidParameterValue 유효하지 않은 파라메터 값
	ErrInvalidParameterValue = New("invalid parameter value")

	// ErrLengthOverflowParameterValue 파라메터 값 문자열 길이 오버플로우
	ErrLengthOverflowParameterValue = New("length overflow parameter value")

	// ErrOutOfRangeParameterValue 범위를 벗어난 파라메터 값
	ErrOutOfRangeParameterValue = New("out of range parameter value")

	// ErrUnavailableParameterValue 사용할 수 없는 파라메터 값
	ErrUnavailableParameterValue = New("unavailable parameter value")

	// ErrPatternMismatchParameterValue 파라메터 값 패턴 불일치(정규표현식)
	ErrPatternMismatchParameterValue = New("pattern mismatch parameter value")

	// ErrFormatMismatchParameterValue 파라메터 값 형식 불일치
	ErrFormatMismatchParameterValue = New("format mismatch parameter value")
)

// RequiredParameter 필수 파라메터 누락
func RequiredParameter(param string) error {
	return Wrap(
		ErrRequiredParameter,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"param": param,
		}),
	)
}

// UnchangeableParameter 수정이 불가능한 파라메터
func UnchangeableParameter(param string) error {
	return Wrap(
		ErrUnchangeableParameter,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"param": param,
		}),
	)
}

// ConflictParameterValue 충돌(중복)이 발생한 파라메터 값
func ConflictParameterValue(param string, value interface{}) error {
	return Wrap(
		ErrConflictParameterValue,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"param": param,
			"value": value,
		}),
	)
}

// InvalidParameterValue 유효하지 않은 파라메터 값
func InvalidParameterValue(param string, value interface{}, cause string) error {
	return Wrap(
		ErrInvalidParameterValue,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"param": param,
			"value": value,
			"cause": cause,
		}),
	)
}

// LengthOverflowParameterValue 파라메터 값 문자열 길이 오버플로우
func LengthOverflowParameterValue(param, value string, maxLength uint64) error {
	return Wrap(
		ErrLengthOverflowParameterValue,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"param":      param,
			"value":      value,
			"max_length": maxLength,
		}),
	)
}

// OutOfRangeParameterValue 범위를 벗어난 파라메터 값
func OutOfRangeParameterValue(param string, value interface{}, min, max interface{}) error {
	return Wrap(
		ErrOutOfRangeParameterValue,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"param": param,
			"value": value,
			"min":   min,
			"max":   max,
		}),
	)
}

// UnavailableParameterValue 사용할 수 없는 파라메터 값
func UnavailableParameterValue(param string, value interface{}, availableValues []interface{}) error {
	return Wrap(
		ErrUnavailableParameterValue,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"param":            param,
			"value":            value,
			"available_values": availableValues,
		}),
	)
}

// PatternMismatchParameterValue 파라메터 값 패턴 불일치(정규표현식)
func PatternMismatchParameterValue(param string, value interface{}, pattern regexp.Regexp) error {
	return Wrap(
		ErrPatternMismatchParameterValue,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"param":   param,
			"value":   value,
			"pattern": pattern,
		}),
	)
}

// FormatMismatchParameterValue 파라메터 값 형식 불일치
func FormatMismatchParameterValue(param string, value interface{}, format string) error {
	return Wrap(
		ErrFormatMismatchParameterValue,
		CallerSkipCount(1),
		WithValue(map[string]interface{}{
			"param":  param,
			"value":  value,
			"format": format,
		}),
	)
}
