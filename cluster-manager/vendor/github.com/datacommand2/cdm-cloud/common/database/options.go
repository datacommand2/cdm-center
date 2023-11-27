package database

import (
	"github.com/micro/go-micro/v2/registry"
	"time"
)

// Option 는 Options 를 인자로 받는 함수의 타입이다.
type Option func(*Options)

// Options 는 ConnectionWrapper 의 설정 항목들에 대한 구조체이다.
type Options struct {
	Registry registry.Registry

	ServiceName string
	DBName      string
	Username    string
	Password    string

	SSLEnable bool
	SSLCACert string

	HeartbeatInterval time.Duration
	ReconnectInterval time.Duration

	TestMode bool
}

// Registry 는 서비스 레지스트리를 설정하는 함수이다.
// 설정하지 않을 경우 registry.DefaultRegistry 를 사용한다.
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

// SSLEnable 은 SSL 접속 활성화 여부를 설정하는 함수이다.
func SSLEnable(enable bool) Option {
	return func(o *Options) {
		o.SSLEnable = enable
	}
}

// SSLCACert 는 서버에 적용된 인증서를 발행한 CA 의 인증서를 설정하는 함수이다.
func SSLCACert(ca string) Option {
	return func(o *Options) {
		o.SSLCACert = ca
	}
}

// HeartbeatInterval 은 접속유지 여부 확인 주기를 설정하는 함수이며, 설정시 기본 설정의 확인 주기를 대체한다.
func HeartbeatInterval(i time.Duration) Option {
	return func(o *Options) {
		o.HeartbeatInterval = i
	}
}

// ReconnectInterval 은 재접속 시도 주기를 설정하는 함수이며, 설정시 기본 설정의 시도 주기를 대체한다.
func ReconnectInterval(i time.Duration) Option {
	return func(o *Options) {
		o.ReconnectInterval = i
	}
}

// TestMode 는 테스트 모드로 설정하는 함수이며, 테스트 모드에서의 변경분은 모두 롤백된다.
func TestMode() Option {
	return func(o *Options) {
		o.TestMode = true
	}
}
