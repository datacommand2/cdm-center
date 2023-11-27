package store

import (
	"context"
	"github.com/micro/go-micro/v2/registry"
	"time"
)

// Options store 설정을 위한 옵션
type Options struct {
	// Registry service registry
	Registry registry.Registry

	// ServiceName store 서비스의 이름
	ServiceName string

	// Context store 의 context 저장
	Context context.Context

	//HeartBeatInterval store 의 접속여부 확인 주기
	HeartBeatInterval time.Duration
}

// Option Options 값 설정
type Option func(o *Options)

// Registry 는 서비스 레지스트리를 설정하는 함수이다.
// 설정하지 않을 경우 registry.DefaultRegistry 를 사용한다.
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

// HeartbeatInterval store 의 접속 여부 주기를 설정하는 함수 이다.
func HeartbeatInterval(interval time.Duration) Option {
	return func(o *Options) {
		o.HeartBeatInterval = interval
	}
}

// GetOptions Get 함수 옵션
type GetOptions struct {
	// Timeout 요청 timeout
	Timeout time.Duration
}

// GetOption GetOptions 값 설정
type GetOption func(r *GetOptions)

// GetTimeout Get 요청 타임아웃 설정
func GetTimeout(d time.Duration) GetOption {
	return func(r *GetOptions) {
		r.Timeout = d
	}
}

// PutOptions Put 함수 옵션
type PutOptions struct {
	// TTL is the time until the record expires
	TTL time.Duration

	// Timeout 요청 timeout
	Timeout time.Duration
}

// PutOption PutOptions 값 설정
type PutOption func(w *PutOptions)

// PutTTL 데이터 만료 시간 설정
func PutTTL(d time.Duration) PutOption {
	return func(w *PutOptions) {
		w.TTL = d
	}
}

// PutTimeout Put 요청 타임아웃 설정
func PutTimeout(d time.Duration) PutOption {
	return func(w *PutOptions) {
		w.Timeout = d
	}
}

// DeleteOptions Delete 함수 옵션
type DeleteOptions struct {
	// Timeout 요청 timeout
	Timeout time.Duration

	// Prefix prefix 와 일치하는 key 삭제 여부
	Prefix bool
}

// DeleteOption DeleteOptions 값 설정
type DeleteOption func(d *DeleteOptions)

// DeleteTimeout Delete 요청 타임아웃 설정
func DeleteTimeout(t time.Duration) DeleteOption {
	return func(d *DeleteOptions) {
		d.Timeout = t
	}
}

// DeletePrefix 삭제 할 prefix key 설정
func DeletePrefix() DeleteOption {
	return func(d *DeleteOptions) {
		d.Prefix = true
	}
}

// ListOptions List 함수 옵션
type ListOptions struct {
	// Timeout 요청 timeout
	Timeout time.Duration
}

// ListOption ListOptions 값 설정
type ListOption func(l *ListOptions)

// ListTimeout List 요청 타임아웃 설정
func ListTimeout(d time.Duration) ListOption {
	return func(l *ListOptions) {
		l.Timeout = d
	}
}

// TxnOptions Txn 함수 옵션
type TxnOptions struct {
	// Timeout 요청 timeout
	Timeout time.Duration
}

// TxnOption TxnOption 값 설정
type TxnOption func(l *TxnOptions)

// TxnCommitTimeout TxnCommit 요청 타임아웃 설정
func TxnCommitTimeout(d time.Duration) TxnOption {
	return func(o *TxnOptions) {
		o.Timeout = d
	}
}
