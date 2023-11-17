package notification

// Package rabbitmq provides a RabbitMQ Openstack Monitor

import (
	"errors"
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

// OsloMessage notification event message body 구조체
type OsloMessage struct {
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
}

type notification struct {
	conn           *rabbitMQConn
	prefetchCount  int
	prefetchGlobal bool
	mtx            sync.Mutex
	wg             sync.WaitGroup
	address        string
	auth           string
}

type subscriber struct {
	mtx       sync.Mutex
	mayRun    bool
	ch        *amqp.Channel
	queueArgs map[string]interface{}
	notifier  *notification
	fn        func(msg amqp.Delivery)
	headers   map[string]interface{}
}

type publication struct {
	d   amqp.Delivery
	m   *Message
	t   string
	err error
}

// Handler is used to process messages via a subscription of a topic.
// The handler is passed a publication interface which contains the
// message and optional Ack method to acknowledge receipt of the message.
type Handler func(Event) error

// Message structure for monitor
type Message struct {
	Header map[string]string
	Body   []byte
}

// Event is given to a subscription handler for processing
type Event interface {
	Topic() string
	Message() *Message
}

// Subscriber is a convenience return type for the Subscribe method
type Subscriber interface {
	Unsubscribe() error
}

func (p *publication) Topic() string {
	return p.t
}

func (p *publication) Message() *Message {
	return p.m
}

func (s *subscriber) Unsubscribe() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.mayRun = false
	if s.ch != nil {
		return s.ch.Close()
	}
	return nil
}

func (s *subscriber) resubscribe() {
	minResubscribeDelay := 100 * time.Millisecond
	maxResubscribeDelay := 30 * time.Second
	expFactor := time.Duration(2)
	reSubscribeDelay := minResubscribeDelay

	//loop until unsubscribe
	for {
		s.mtx.Lock()
		mayRun := s.mayRun
		s.mtx.Unlock()
		if !mayRun {
			// we are unsubscribed, showdown routine
			return
		}

		select {
		//check shutdown case
		case <-s.notifier.conn.close:
			//yep, its shutdown case
			return
			//wait until we reconect to rabbit
		case <-s.notifier.conn.waitConnection:
		}

		// it may crash (panic) in case of Consume without connection, so recheck it
		s.notifier.mtx.Lock()
		if !s.notifier.conn.connected {
			s.notifier.mtx.Unlock()
			continue
		}
		ch, sub, err := s.notifier.conn.Consume()

		s.notifier.mtx.Unlock()
		switch err {
		case nil:
			reSubscribeDelay = minResubscribeDelay
			s.mtx.Lock()
			s.ch = ch
			s.mtx.Unlock()
		default:
			if reSubscribeDelay > maxResubscribeDelay {
				reSubscribeDelay = maxResubscribeDelay
			}
			time.Sleep(reSubscribeDelay)
			reSubscribeDelay *= expFactor
			continue
		}
		for d := range sub {
			s.notifier.wg.Add(1)
			s.fn(d)
			s.notifier.wg.Done()
		}
	}
}

func (n *notification) Subscribe(handler Handler) (Subscriber, error) {
	if n.conn == nil {
		return nil, errors.New("not connected openstack notification")
	}

	fn := func(msg amqp.Delivery) {
		header := make(map[string]string)
		for k, v := range msg.Headers {
			header[k], _ = v.(string)
		}
		m := &Message{
			Header: header,
			Body:   msg.Body,
		}

		p := &publication{d: msg, m: m, t: msg.RoutingKey}
		p.err = handler(p)
		if p.err != nil {
			if err := msg.Nack(false, false); err != nil {
				log.Printf("could not negatively acknowledge on a delivery. Cause: %+v\n", err)
			}
		}
	}

	sret := &subscriber{mayRun: true, notifier: n,
		fn: fn}

	go sret.resubscribe()

	return sret, nil
}

func (n *notification) Connect() error {
	conn := n.conn

	if n.conn == nil {
		conn = newRabbitMQConn(
			0,
			false,
			n.auth,
			n.address)
	}

	conf := defaultAmqpConfig

	if err := conn.Connect(&conf); err != nil {
		return err
	}

	n.conn = conn
	return nil
}

func (n *notification) DeleteQueue() error {
	return n.conn.deleteQueue()
}

func (n *notification) Disconnect() error {
	if n.conn == nil {
		return errors.New("connection is nil")
	}
	n.mtx.Lock()
	ret := n.conn.Close()
	n.mtx.Unlock()

	n.wg.Wait() // wait all goroutines
	return ret
}
