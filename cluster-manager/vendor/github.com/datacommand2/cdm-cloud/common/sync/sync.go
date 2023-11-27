package sync

import (
	"context"
)

// DefaultSync sync 의 기본 sync
var DefaultSync Sync

// Sync 는 분산 동기화 대한 인터 페이스
type Sync interface {
	CampaignLeader(ctx context.Context, path string) (Leader, error)
	Lock(ctx context.Context, path string) (Mutex, error)
	Close() error
}

// Leader 는 Leader 인터 페이스
type Leader interface {
	Status() chan bool
	Resign(ctx context.Context) error
	Close() error
}

// Mutex 는 Mutex 인터 페이스
type Mutex interface {
	Unlock(ctx context.Context) error
	Close() error
}

// CampaignLeader 는 DefaultSync 로 리더를 선출한다.
// 리더로 선출될 때까지 blocking 되며, context timeout 이나 cancel 로 리더 선출 대기를 중지할 수 있다.
func CampaignLeader(ctx context.Context, path string) (Leader, error) {
	return DefaultSync.CampaignLeader(ctx, path)
}

// Lock 는 DefaultSync 의 Lock 로 Global Mutex Lock 을 획득한다.
// 획득할 때까지 blocking 되며, context timeout 이나 cancel 로 lock 획득 대기를 중지할 수 있다.
func Lock(ctx context.Context, path string) (Mutex, error) {
	return DefaultSync.Lock(ctx, path)
}

// Init 는 DefaultSync 를 etcdSync 로 설정 하는 함수
func Init(serviceName string, opts ...Option) error {
	var err error
	DefaultSync, err = NewEtcdSync(serviceName, opts...)
	if err != nil {
		return err
	}

	return nil
}

// Close 는 DefaultSync 의 연결 종료 함수
func Close() error {
	return DefaultSync.Close()
}
