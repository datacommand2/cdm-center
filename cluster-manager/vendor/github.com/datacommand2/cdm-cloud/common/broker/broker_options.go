package broker

import (
	"context"
	"crypto/tls"

	"github.com/micro/go-micro/v2/codec"
	"github.com/micro/go-micro/v2/registry"
)

// Options structure for broker
type Options struct {
	Secure    bool
	Codec     codec.Marshaler
	TLSConfig *tls.Config

	// Registry used for clustering
	Registry registry.Registry

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context

	//ServiceName 을 통해 etcd 에서 서비스 정보(IP)를 받아 옴
	ServiceName string

	// PersistentQueue is an option to set QueueName(for cdm-cloud)
	PersistentQueues []string
}

// PublishOptions structure for broker
type PublishOptions struct {
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

// SubscribeOptions structure for broker
type SubscribeOptions struct {
	// AutoAck defaults to true. When a handler returns
	// with a nil error the message is acked.
	AutoAck bool
	// Subscribers with the same queue name
	// will create a shared subscription where each
	// receives a subset of messages.
	Queue string

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

// Option 은 broker option 설정 함수이다.
type Option func(*Options)

// PersistentQueue is an option to set QueueName(for cdm-cloud)
func PersistentQueue(values ...string) Option {
	return func(o *Options) {
		o.PersistentQueues = append(o.PersistentQueues, values...)
	}
}

// PublishOption 은 publish option 설정 함수이다.
type PublishOption func(*PublishOptions)

// PublishContext set context
func PublishContext(ctx context.Context) PublishOption {
	return func(o *PublishOptions) {
		o.Context = ctx
	}
}

// SubscribeOption 은 subscribe option 설정 함수이다.
type SubscribeOption func(*SubscribeOptions)

// Codec sets the codec used for encoding/decoding used where
// a broker does not support headers
func Codec(c codec.Marshaler) Option {
	return func(o *Options) {
		o.Codec = c
	}
}

// DisableAutoAck will disable auto acking of messages
// after they have been handled.
func DisableAutoAck() SubscribeOption {
	return func(o *SubscribeOptions) {
		o.AutoAck = false
	}
}

// Queue sets the name of the queue to share messages on
func Queue(name string) SubscribeOption {
	return func(o *SubscribeOptions) {
		o.Queue = name
	}
}

// Registry sets the service registry
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

// Secure communication with the broker
func Secure(b bool) Option {
	return func(o *Options) {
		o.Secure = b
	}
}

// TLSConfig sets Specify TLS Config
func TLSConfig(t *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = t
	}
}
