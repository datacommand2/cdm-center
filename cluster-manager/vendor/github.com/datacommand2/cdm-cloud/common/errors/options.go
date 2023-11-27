package errors

// Option 는 Options 를 인자로 받는 함수의 타입이다.
type Option func(*Options)

// Options 는 error 를 wrapping 할 때, skip 할 frame 갯수를 설정하기 위한 구조체이다.
type Options struct {
	CallerSkipCount int
	Value           value
}

// CallerSkipCount set frame count to skip
func CallerSkipCount(c int) Option {
	return func(args *Options) {
		args.CallerSkipCount = c
	}
}

// WithValue set frame count to skip
func WithValue(v map[string]interface{}) Option {
	return func(args *Options) {
		args.Value = v
	}
}
