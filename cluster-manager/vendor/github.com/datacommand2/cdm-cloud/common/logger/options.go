package logger

import (
	"context"
	"github.com/micro/go-micro/v2/logger"
)

type serviceNameKey struct{}

// WithServiceName 은 서비스 이름 옵션을 설정하는 함수를 반환한다.
// 서비스 이름이 설정되면 서비스 이름이 로그 텍스트의 prefix 에 붙는다.
func WithServiceName(name string) logger.Option {
	return func(o *logger.Options) {
		o.Context = context.WithValue(o.Context, serviceNameKey{}, name)
	}
}

// WithLevel 은 로그 레벨을 수정하는 logger.Option 타입의 함수를 반환한다.
// Init 함수의 인자로 넘겨 로그 레벨을 수정할 수 있다.
// TraceLevel(-2) ~ FatalLevel(3)
func WithLevel(level int) logger.Option {
	return func(args *logger.Options) {
		args.Level = logger.Level(level)
	}
}
