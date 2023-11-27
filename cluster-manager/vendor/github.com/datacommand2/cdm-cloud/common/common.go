package common

import (
	"github.com/micro/cli/v2"
	mCmd "github.com/micro/go-micro/v2/config/cmd"
)

// Option 는 Options 를 인자로 받는 함수 타입이다.
type Option func(*Options)

// Options 는 command line parser 의 설정 항목들에 대한 구조체이다.
type Options struct {
	Flags            []cli.Flag
	PersistentQueues []string
}

// 이 패키지를 import 하면 자동으로 go-micro framework 서비스 초기화시 수행하는
// command line parser 를 CDM Cloud framework 를 초기화하는 기능의 parser 로 대체한다.
//
// 대체되는 parser 는 CDM 서비스들이 공통으로 사용할 database connection config 와
// message broker, key-value store, logger 를 초기화하는 기능을 수행한다.
func init() {
	mCmd.DefaultCmd = newCommand()
}

// Init 은 CDM Cloud framework 초기화 command line parser 에 옵션을 추가해주는 함수이다.
// go-micro framework 서비스 초기화 후 호출하는 경우 옵션이 적영되지 않는다.
func Init(opts ...Option) {
	mCmd.DefaultCmd = newCommand(opts...)
}

// Destroy 는 CDM Cloud framework 에서 제공하는 message broker, key-value store, logger 등의 기능을 중지한다.
func Destroy() {
	if cmd, ok := mCmd.DefaultCmd.(command); ok {
		cmd.Destroy()
	}
}

// WithFlags 는 서비스에서 필요로하는 특정 Flag 를 추가해주는 함수이다.
// 공통 arguments 가 아닌 다른 arguments 가 필요한 경우 사용한다.
func WithFlags(flags ...cli.Flag) Option {
	return func(o *Options) {
		o.Flags = append(o.Flags, flags...)
	}
}

// WithPersistentQueue 는 서비스에서 필요로하는 Persistent Queue 를 추가해주는 함수이다.
// Persistent Queue 에 message 를 publish 하거나 subscribe 해야 하는 경우 사용한다.
func WithPersistentQueue(name ...string) Option {
	return func(o *Options) {
		o.PersistentQueues = append(o.PersistentQueues, name...)
	}
}
