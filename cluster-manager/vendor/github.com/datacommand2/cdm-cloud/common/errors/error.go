package errors

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"runtime"
)

// Error 는 event.ReportEvent() 전달할 Contents 를 담기 위한 error 와 에러 추적에 사용되는 stack 이 선언된 구조체
type Error struct {
	error error
	stack *stack
	value
}

func (e *Error) Error() string {
	return e.error.Error()
}

// MarshalJSON json marshal function
func (e *Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.value)
}

// UnmarshalJSON json unmarshal function
func (e *Error) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &e.value)
}

// Format 은 Error 를 통한 에러 추적시 출력 형식을 정하는 함수
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, e.Error())
			e.value.format(s, verb)
			e.stack.format(s, verb)
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	}
}

type value map[string]interface{}

func (v *value) len() int {
	b, err := json.Marshal(v)
	if err != nil {
		return 0
	}

	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)

	return len(m)
}

func (v *value) format(s fmt.State, verb rune) {
	if v.len() == 0 {
		return
	}

	switch verb {
	case 'v', 's':
		b, _ := json.MarshalIndent(v, "", "  ")
		_, _ = fmt.Fprintf(s, "\n%s", string(b))
	}
}

// stack represents a stack of program counters.
type stack []uintptr

func (s *stack) format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case st.Flag('+'):
			for _, pc := range *s {
				f := errors.Frame(pc)
				_, _ = fmt.Fprintf(st, "\n%+v", f)
			}
		}
	}
}

func callers(skipCount int) *stack {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	var st stack = pcs[skipCount:n]
	return &st
}

// New 는 메시지를 Wrapper 구조체로 wrapping 하는 함수
func New(message string) error {
	return &Error{error: errors.New(message), stack: callers(0)}
}

// Wrap 은 에러를 Wrapper 구조체로 wrapping 하는 함수
func Wrap(err error, opts ...Option) *Error {
	var options Options
	for _, o := range opts {
		o(&options)
	}

	return &Error{
		error: err,
		stack: callers(options.CallerSkipCount),
		value: options.Value,
	}
}

// Equal 은 동일한 에러인지 확인하는 함수
func Equal(a, b error) bool {
	if a == nil {
		return false
	}

	return a.Error() == b.Error()
}
