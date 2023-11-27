package broker

//
// All credit to Mondo
//

import (
	"crypto/tls"
	"errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"
	"github.com/streadway/amqp"
	"strings"
	"sync"
	"time"
)

var (
	// The amqp library does not seem to set these when using amqp.DialConfig
	// (even though it says so in the comments) so we set them manually to make
	// sure to not brake any existing functionality
	defaultAmqpConfig = amqp.Config{
		Heartbeat: 10 * time.Second,
		Locale:    "en_US",
	}
)

type rabbitMQConn struct {
	Registry        registry.Registry
	Connection      *amqp.Connection
	Channel         *rabbitMQChannel
	ExchangeChannel *rabbitMQChannel
	exchange        Exchange
	serviceName     string
	qNames          []string
	accountInfo     string
	prefetchCount   int
	prefetchGlobal  bool

	sync.Mutex
	connected bool
	close     chan bool

	waitConnection chan struct{}
}

// Exchange is the rabbitmq exchange
type Exchange struct {
	// Name of the exchange
	Name string
	// Whether its persistent
	Durable bool
}

func newRabbitMQConn(ex Exchange, prefetchCount int, prefetchGlobal bool, reg registry.Registry, serviceName string, qNames []string, accountInfo string) *rabbitMQConn {
	ret := &rabbitMQConn{
		Registry:       reg,
		exchange:       ex,
		prefetchCount:  prefetchCount,
		prefetchGlobal: prefetchGlobal,
		close:          make(chan bool),
		waitConnection: make(chan struct{}),
		qNames:         qNames,
		serviceName:    serviceName,
		accountInfo:    accountInfo,
	}

	// its bad case of nil == waitConnection, so close it at start
	close(ret.waitConnection)
	return ret
}

func (r *rabbitMQConn) connect(secure bool, config *amqp.Config) error {
	// try connect
	if err := r.tryConnect(secure, config); err != nil {
		return err
	}

	// connected
	r.Lock()
	r.connected = true
	r.Unlock()

	// create reconnect loop
	go r.reconnect(secure, config)
	return nil
}

func (r *rabbitMQConn) reconnect(secure bool, config *amqp.Config) {
	// skip first connect
	var connect bool

	for {
		if connect {
			// try reconnect
			if err := r.tryConnect(secure, config); err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			// connected
			r.Lock()
			r.connected = true
			r.Unlock()
			//unblock resubscribe cycle - close channel
			//at this point channel is created and unclosed - close it without any additional checks
			close(r.waitConnection)
		}

		connect = true
		notifyClose := make(chan *amqp.Error)
		r.Connection.NotifyClose(notifyClose)

		// block until closed
		select {
		case <-notifyClose:
			// block all resubscribe attempt - they are useless because there is no connection to rabbitmq
			// create channel 'waitConnection' (at this point channel is nil or closed, create it without unnecessary checks)
			r.Lock()
			r.connected = false
			r.waitConnection = make(chan struct{})
			r.Unlock()
		case <-r.close:
			return
		}
	}
}

func (r *rabbitMQConn) Connect(secure bool, config *amqp.Config) error {
	r.Lock()
	// already connected
	if r.connected {
		r.Unlock()
		return nil
	}

	// check it was closed
	select {
	case <-r.close:
		r.close = make(chan bool)
	default:
		// no op
		// new conn
	}

	r.Unlock()

	return r.connect(secure, config)
}

func (r *rabbitMQConn) Close() error {
	r.Lock()
	defer r.Unlock()
	select {
	case <-r.close:
		return nil
	default:
		close(r.close)
		r.connected = false
	}

	return r.Connection.Close()
}

// PersistrentQueue는 RabbitMQ 연결 단계에서 Queue와 Exchange를 미리 할당하고 바인딩 (cdm-cloud를 위해 추가)
func (r *rabbitMQConn) declarePersistentQueue(qNames, key, exchange string) error {
	consumerChannel, err := newRabbitChannel(r.Connection, r.prefetchCount, r.prefetchGlobal)
	if err != nil {
		return err
	}
	err = consumerChannel.DeclareDurableQueue(qNames, nil)
	if err != nil {
		return err
	}

	err = consumerChannel.BindQueue(qNames, key, exchange, nil)
	if err != nil {
		return err
	}

	return nil
}

// registry 에서 rabbitmq서비스 정보를 얻어와 연결 시도
func (r *rabbitMQConn) doTryConnect(secure bool, config *amqp.Config) error {
	reg := registry.DefaultRegistry
	if r.Registry != nil {
		reg = r.Registry
	}

	s, err := reg.GetService(r.serviceName)
	switch {
	case err != nil && err != registry.ErrNotFound:
		logger.Errorf("Could not discover broker(%s) service. cause: %v", r.serviceName, err)
		return errors.New("could not discover broker service")

	case err == registry.ErrNotFound || s == nil || len(s) == 0:
		logger.Errorf("Not found broker(%s) service.", r.serviceName)
		return errors.New("not found broker service")
	}

	next := selector.RoundRobin(s)

	logger.Infof("Try to connect broker(%s) service", r.serviceName)
	for i := 0; i < len(s[0].Nodes); i++ {
		n, _ := next()

		url := "amqp://" + r.accountInfo + "@" + n.Address + "/"

		if secure || config.TLSClientConfig != nil || strings.HasPrefix(url, "amqps://") {
			if config.TLSClientConfig == nil {
				config.TLSClientConfig = &tls.Config{
					InsecureSkipVerify: true,
				}
			}

			url = strings.Replace(url, "amqp://", "amqps://", 1)
		}

		logger.Debugf("Try to connect %s", url)
		r.Connection, err = amqp.DialConfig(url, *config)
		if err == nil { //Connection success
			return nil
		}

		logger.Debugf("Failed to connect. cause: %v", err)
	}
	return errors.New("could not connect to broker")
}

func (r *rabbitMQConn) tryConnect(secure bool, config *amqp.Config) error {
	var err error

	if config == nil {
		config = &defaultAmqpConfig
	}

	if err = r.doTryConnect(secure, config); err != nil {
		return err
	}

	if r.Channel, err = newRabbitChannel(r.Connection, r.prefetchCount, r.prefetchGlobal); err != nil {
		_ = r.Close()
		return err
	}

	if r.exchange.Durable {
		err = r.Channel.DeclareDurableExchange(r.exchange.Name)
	} else {
		err = r.Channel.DeclareExchange(r.exchange.Name)
	}
	if err != nil {
		_ = r.Close()
		return err
	}

	if r.ExchangeChannel, err = newRabbitChannel(r.Connection, r.prefetchCount, r.prefetchGlobal); err != nil {
		_ = r.Close()
		return err
	}

	// 연결 단계에서 Queue가 할당되어야 하는 경우를 위해 추가
	// Queue 생성이 실패할 경우 이후 각 컴포넌트간의 통신이 불가능하기 때문에 생성될 때까지 반복한다.
	for _, qName := range r.qNames {
		for {
			// PreBindQueue함수의 두번째 매개변수는 topic이지만 이 경우 Queue이름과 topic을 동일하게 처리
			err := r.declarePersistentQueue(qName, qName, r.exchange.Name)
			if err == nil {
				break
			}
			logger.Errorf("Failed to PreBind [Qname : %s <-> Exchange : %s]",
				qName, r.exchange.Name)
			time.Sleep(time.Second * 1)
		}
	}
	return nil
}

func (r *rabbitMQConn) Consume(queue, key string, headers amqp.Table, qArgs amqp.Table, autoAck, durableQueue bool) (*rabbitMQChannel, <-chan amqp.Delivery, error) {
	consumerChannel, err := newRabbitChannel(r.Connection, r.prefetchCount, r.prefetchGlobal)
	if err != nil {
		return nil, nil, err
	}

	if durableQueue {
		err = consumerChannel.DeclareDurableQueue(queue, qArgs)
	} else {
		err = consumerChannel.DeclareQueue(queue, qArgs)
	}

	if err != nil {
		return nil, nil, err
	}

	deliveries, err := consumerChannel.ConsumeQueue(queue, autoAck)
	if err != nil {
		return nil, nil, err
	}

	err = consumerChannel.BindQueue(queue, key, r.exchange.Name, headers)
	if err != nil {
		return nil, nil, err
	}

	return consumerChannel, deliveries, nil
}

func (r *rabbitMQConn) Publish(exchange, key string, msg amqp.Publishing) error {
	return r.ExchangeChannel.Publish(exchange, key, msg)
}
