package sync

import (
	"github.com/micro/go-micro/v2/registry"
	"time"
)

type options struct {
	ttl               int
	registry          registry.Registry
	user              string
	password          string
	heartbeatInterval time.Duration
}

// Option sync 옵션 설정
type Option func(o *options)

// TTL 는 ttl 설정
func TTL(ttl int) Option {
	return func(o *options) {
		o.ttl = ttl
	}
}

// Registry 는 서비스 레지 스트리 를 설정
// 설정 하지 않을 경우 registry.DefaultRegistry 를 사용
func Registry(r registry.Registry) Option {
	return func(o *options) {
		o.registry = r
	}
}

// Auth 는 사용자 계정, 비밀 번호 설정
func Auth(u, p string) Option {
	return func(o *options) {
		o.user = u
		o.password = p
	}
}

// HeartbeatInterval 서비스 의 접속 여부 주기를 설정
// 설정 하지 않을 경우 10 초로 설정 됨
func HeartbeatInterval(interval time.Duration) Option {
	return func(o *options) {
		o.heartbeatInterval = interval
	}
}
