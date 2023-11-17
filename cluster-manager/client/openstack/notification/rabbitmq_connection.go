package notification

//
// All credit to Mondo
//

import (
	"errors"
	"github.com/datacommand2/cdm-center/cluster-manager/constant"
	"github.com/micro/go-micro/v2/logger"
	"github.com/streadway/amqp"
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

	openstackExchange = []string{"aodh", "openstack", "nova", "neutron", "keystone", "cinder", "ceilometer", "dns", "glance", "heat", "ironic", "magnum", "sahara", "swift", "trove", "zaqar"}
)

type rabbitMQConn struct {
	Connection     *amqp.Connection
	accountInfo    string
	address        string
	port           string
	prefetchCount  int
	prefetchGlobal bool

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

func newRabbitMQConn(prefetchCount int, prefetchGlobal bool, accountInfo string, address string) *rabbitMQConn {
	ret := &rabbitMQConn{
		prefetchCount:  prefetchCount,
		prefetchGlobal: prefetchGlobal,
		close:          make(chan bool),
		waitConnection: make(chan struct{}),
		accountInfo:    accountInfo,
		address:        address,
	}

	// its bad case of nil == waitConnection, so close it at start
	close(ret.waitConnection)
	return ret
}

func (r *rabbitMQConn) connect(config *amqp.Config) error {
	// try connect
	if err := r.tryConnect(config); err != nil {
		return err
	}

	// connected
	r.Lock()
	r.connected = true
	r.Unlock()

	// create reconnect loop
	go r.reconnect(config)
	return nil
}

func (r *rabbitMQConn) reconnect(config *amqp.Config) {
	// skip first connect
	var connect bool

	for {
		if connect {
			// try reconnect
			if err := r.tryConnect(config); err != nil {
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

func (r *rabbitMQConn) Connect(config *amqp.Config) error {
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

	return r.connect(config)
}

func (r *rabbitMQConn) Close() error {
	r.Lock()
	defer r.Unlock()
	select {
	case <-r.close:
		logger.Warn("Close : -1")
		return nil
	default:
		close(r.close)
		r.connected = false
		logger.Warn("Close : 0")
	}

	//check
	if r.Connection == nil {
		logger.Warn("Close : 1")
		return nil
	}

	return r.Connection.Close()
}

// registry 에서 rabbitmq서비스 정보를 얻어와 연결 시도
func (r *rabbitMQConn) doTryConnect(config *amqp.Config) error {
	var err error
	url := "amqp://" + r.accountInfo + "@" + r.address + "/"

	r.Connection, err = amqp.DialConfig(url, *config)
	if err == nil { //Connection success
		return nil
	}

	return errors.New("could not connect to openstack notification")
}

func (r *rabbitMQConn) deleteQueue() error {
	c, err := r.Connection.Channel()
	if err != nil {
		return err
	}

	_, err = c.QueueDelete(constant.ServiceName, false, false, false)
	if err != nil {
		return err
	}

	return nil
}

func (r *rabbitMQConn) tryConnect(config *amqp.Config) error {
	var err error

	if config == nil {
		config = &defaultAmqpConfig
	}

	if err = r.doTryConnect(config); err != nil {
		return err
	}

	return nil
}

func (r *rabbitMQConn) Consume() (*amqp.Channel, <-chan amqp.Delivery, error) {
	c, err := r.Connection.Channel()
	if err != nil {
		return nil, nil, err
	}

	_, err = c.QueueDeclare(constant.ServiceName, false, true, false, false, nil)
	if err != nil {
		return nil, nil, err
	}

	for _, ex := range openstackExchange {
		if err = c.ExchangeDeclare(ex, "topic", false, false, false, false, nil); err != nil {
			return nil, nil, err
		}

		if err = c.QueueBind(constant.ServiceName, "#", ex, false, nil); err != nil {
			return nil, nil, err
		}
	}

	deliveries, err := c.Consume(constant.ServiceName, constant.ServiceName, false, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}

	return c, deliveries, nil
}
