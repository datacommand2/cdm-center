package etcd

import (
	"context"
	cryptotls "crypto/tls"
	"github.com/micro/go-micro/v2/registry"
	"google.golang.org/grpc"
	"time"
)

// Implement all the options from https://pkg.go.dev/github.com/coreos/etcd/clientv3?tab=doc#Config
// Need to use non basic types in context.WithValue

type autoSyncInterval string
type dialTimeout string
type dialKeepAliveTime string
type dialKeepAliveTimeout string
type maxCallSendMsgSize string
type maxCallRecvMsgSize string
type tls string
type rejectOldCluster string
type dialOptions string
type clientContext string
type permitWithoutStream string
type username string
type password string

// AutoSyncInterval is the interval to update endpoints with its latest members.
// 0 disables auto-sync. By default auto-sync is disabled.
func AutoSyncInterval(d time.Duration) Option {
	return func(o *options) {
		o.Context = context.WithValue(o.Context, autoSyncInterval(""), d)
	}
}

// DialTimeout is the timeout for failing to establish a connection.
func DialTimeout(d time.Duration) Option {
	return func(o *options) {
		o.Context = context.WithValue(o.Context, dialTimeout(""), d)
	}
}

// DialKeepAliveTime is the time after which client pings the server to see if
// transport is alive.
func DialKeepAliveTime(d time.Duration) Option {
	return func(o *options) {
		o.Context = context.WithValue(o.Context, dialKeepAliveTime(""), d)
	}
}

// DialKeepAliveTimeout is the time that the client waits for a response for the
// keep-alive probe. If the response is not received in this time, the connection is closed.
func DialKeepAliveTimeout(d time.Duration) Option {
	return func(o *options) {
		o.Context = context.WithValue(o.Context, dialKeepAliveTimeout(""), d)
	}
}

// MaxCallSendMsgSize is the client-side request send limit in bytes.
// If 0, it defaults to 2.0 MiB (2 * 1024 * 1024).
// Make sure that "MaxCallSendMsgSize" < server-side default send/recv limit.
// ("--max-request-bytes" flag to etcdClient or "embed.Config.MaxRequestBytes").
func MaxCallSendMsgSize(size int) Option {
	return func(o *options) {
		o.Context = context.WithValue(o.Context, maxCallSendMsgSize(""), size)
	}
}

// MaxCallRecvMsgSize is the client-side response receive limit.
// If 0, it defaults to "math.MaxInt32", because range response can
// easily exceed request send limits.
// Make sure that "MaxCallRecvMsgSize" >= server-side default send/recv limit.
// ("--max-request-bytes" flag to etcdClient or "embed.Config.MaxRequestBytes").
func MaxCallRecvMsgSize(size int) Option {
	return func(o *options) {
		o.Context = context.WithValue(o.Context, maxCallRecvMsgSize(""), size)
	}
}

// TLS holds the client secure credentials, if any.
func TLS(conf *cryptotls.Config) Option {
	return func(o *options) {
		t := conf.Clone()
		o.Context = context.WithValue(o.Context, tls(""), t)
	}
}

// RejectOldCluster when set will refuse to create a client against an outdated cluster.
func RejectOldCluster(b bool) Option {
	return func(o *options) {
		o.Context = context.WithValue(o.Context, rejectOldCluster(""), b)
	}
}

// DialOptions is a list of dial options for the grpc client (e.g., for interceptors).
// For example, pass "grpc.WithBlock()" to block until the underlying connection is up.
// Without this, Dial returns immediately and connecting the server happens in background.
// "github.com/google.golang.org/grpc/dialoptions.go"
func DialOptions(opts []grpc.DialOption) Option {
	return func(o *options) {
		if len(opts) > 0 {
			ops := make([]grpc.DialOption, len(opts))
			copy(ops, opts)
			o.Context = context.WithValue(o.Context, dialOptions(""), ops)
		}
	}
}

// ClientContext is the default etcd3 client context; it can be used to cancel grpc
// dial out another operations that do not have an explicit context.
func ClientContext(ctx context.Context) Option {
	return func(o *options) {
		o.Context = context.WithValue(o.Context, clientContext(""), ctx)
	}
}

// PermitWithoutStream when set will allow client to send keepalive pings to server without any active streams(RPCs).
func PermitWithoutStream(b bool) Option {
	return func(o *options) {
		o.Context = context.WithValue(o.Context, permitWithoutStream(""), b)
	}
}

// Auth sets username, password.
func Auth(u, p string) Option {
	return func(o *options) {
		o.Context = context.WithValue(o.Context, username(""), u)
		o.Context = context.WithValue(o.Context, password(""), p)
	}
}

func (e *etcdClient) applyConfig() {
	if v := e.options.Context.Value(autoSyncInterval("")); v != nil {
		e.config.AutoSyncInterval = v.(time.Duration)
	}
	if v := e.options.Context.Value(dialTimeout("")); v != nil {
		e.config.DialTimeout = v.(time.Duration)
	}
	if v := e.options.Context.Value(dialKeepAliveTime("")); v != nil {
		e.config.DialKeepAliveTime = v.(time.Duration)
	}
	if v := e.options.Context.Value(dialKeepAliveTimeout("")); v != nil {
		e.config.DialKeepAliveTimeout = v.(time.Duration)
	}
	if v := e.options.Context.Value(maxCallSendMsgSize("")); v != nil {
		e.config.MaxCallSendMsgSize = v.(int)
	}
	if v := e.options.Context.Value(maxCallRecvMsgSize("")); v != nil {
		e.config.MaxCallRecvMsgSize = v.(int)
	}
	if v := e.options.Context.Value(tls("")); v != nil {
		e.config.TLS = v.(*cryptotls.Config)
	}
	if v := e.options.Context.Value(rejectOldCluster("")); v != nil {
		e.config.RejectOldCluster = v.(bool)
	}
	if v := e.options.Context.Value(dialOptions("")); v != nil {
		e.config.DialOptions = v.([]grpc.DialOption)
	}
	if v := e.options.Context.Value(clientContext("")); v != nil {
		e.config.Context = v.(context.Context)
	}
	if v := e.options.Context.Value(permitWithoutStream("")); v != nil {
		e.config.PermitWithoutStream = v.(bool)
	}
	if v := e.options.Context.Value(username("")); v != nil {
		e.config.Username = v.(string)
	}
	if v := e.options.Context.Value(password("")); v != nil {
		e.config.Password = v.(string)
	}
}

type options struct {
	// Registry service registry
	Registry registry.Registry

	// ServiceName store 서비스의 이름
	ServiceName string

	// Context store 의 context 저장
	Context context.Context

	//HeartBeatInterval store 의 접속여부 확인 주기
	HeartBeatInterval time.Duration
}

// Option options 값 설정
type Option func(o *options)

// Registry 는 서비스 레지스트리를 설정하는 함수이다.
// 설정하지 않을 경우 registry.DefaultRegistry 를 사용한다.
func Registry(r registry.Registry) Option {
	return func(o *options) {
		o.Registry = r
	}
}

// HeartbeatInterval store 의 접속 여부 주기를 설정하는 함수 이다.
func HeartbeatInterval(interval time.Duration) Option {
	return func(o *options) {
		o.HeartBeatInterval = interval
	}
}
