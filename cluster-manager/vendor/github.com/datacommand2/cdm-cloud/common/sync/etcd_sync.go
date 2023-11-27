package sync

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/google/uuid"
	"time"

	cc "github.com/coreos/etcd/clientv3/concurrency"
	"github.com/datacommand2/cdm-cloud/common/sync/internal/etcd"
)

const defaultHeartbeatTimeout = 10 * time.Second
const defaultTTL = 60

type etcdSync struct {
	client *clientv3.Client
	ttl    int
}

// NewEtcdSync etcd backend 에 대한 sync 생성
func NewEtcdSync(serviceName string, opts ...Option) (Sync, error) {
	var o = options{
		heartbeatInterval: defaultHeartbeatTimeout,
		ttl:               defaultTTL,
	}
	for _, opt := range opts {
		opt(&o)
	}

	c, err := etcd.NewClient(serviceName, etcd.Auth(o.user, o.password), etcd.Registry(o.registry), etcd.HeartbeatInterval(o.heartbeatInterval))
	if err != nil {
		return nil, err
	}

	return &etcdSync{client: c, ttl: o.ttl}, nil
}

// CampaignLeader etcd backend 에 대한 leader election 함수
func (e *etcdSync) CampaignLeader(ctx context.Context, path string) (Leader, error) {
	s, err := cc.NewSession(e.client, cc.WithTTL(e.ttl))
	if err != nil {
		return nil, err
	}

	l := cc.NewElection(s, path)
	uid := uuid.New().String()

	if err = l.Campaign(ctx, uid); err != nil {
		return &etcdLeader{s: s}, err
	}

	return &etcdLeader{e: l, s: s, uuid: uid}, nil
}

// Lock etcd backend 에 대한 lock acquire 함수
func (e *etcdSync) Lock(ctx context.Context, path string) (Mutex, error) {
	s, err := cc.NewSession(e.client, cc.WithTTL(e.ttl))
	if err != nil {
		return nil, err
	}

	m := cc.NewMutex(s, path)

	if err = m.Lock(ctx); err != nil {
		return &etcdLocker{s: s}, err
	}

	return &etcdLocker{s: s, m: m}, nil
}

// Close etcd backend 연결 종료
func (e *etcdSync) Close() error {
	return e.client.Close()
}

type etcdLeader struct {
	e    *cc.Election
	s    *cc.Session
	uuid string
}

// Status leader 확인 함수
func (l *etcdLeader) Status() chan bool {
	ch := make(chan bool, 1)
	ech := l.e.Observe(context.Background())

	go func() {
		for r := range ech {
			if string(r.Kvs[0].Value) != l.uuid {
				ch <- true
				close(ch)
				return
			}
		}
	}()

	return ch
}

// Resign leader 사임 함수
func (l *etcdLeader) Resign(ctx context.Context) error {
	return l.e.Resign(ctx)
}

// Close etcd backend 에 session 종료 함수
func (l *etcdLeader) Close() error {
	return l.s.Close()
}

type etcdLocker struct {
	s *cc.Session
	m *cc.Mutex
}

// Close etcd backend 에 lock release 함수
func (l *etcdLocker) Unlock(ctx context.Context) error {
	return l.m.Unlock(ctx)
}

// Close etcd backend 에 session 종료 함수
func (l *etcdLocker) Close() (err error) {
	return l.s.Close()
}
