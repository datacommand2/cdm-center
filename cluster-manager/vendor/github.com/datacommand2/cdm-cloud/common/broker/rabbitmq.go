package broker

// Package rabbitmq provides a RabbitMQ broker

import (
	"context"
	"errors"
	"github.com/datacommand2/cdm-cloud/common/logger"
	"github.com/streadway/amqp"
	"sync"
	"time"
)

type rbroker struct {
	conn           *rabbitMQConn
	opts           Options
	prefetchCount  int
	prefetchGlobal bool
	mtx            sync.Mutex
	wg             sync.WaitGroup
	auth           string
}

type subscriber struct {
	mtx          sync.Mutex
	mayRun       bool
	opts         SubscribeOptions
	topic        string
	ch           *rabbitMQChannel
	durableQueue bool
	queueArgs    map[string]interface{}
	r            *rbroker
	fn           func(msg amqp.Delivery)
	headers      map[string]interface{}
}

type publication struct {
	d   amqp.Delivery
	m   *Message
	t   string
	err error
}

func (p *publication) Topic() string {
	return p.t
}

func (p *publication) Message() *Message {
	return p.m
}

func (s *subscriber) Options() SubscribeOptions {
	return s.opts
}

func (s *subscriber) Topic() string {
	return s.topic
}

func (s *subscriber) Message() func(msg amqp.Delivery) {
	return s.fn
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
		case <-s.r.conn.close:
			//yep, its shutdown case
			return
			//wait until we reconect to rabbit
		case <-s.r.conn.waitConnection:
		}

		// it may crash (panic) in case of Consume without connection, so recheck it
		s.r.mtx.Lock()
		if !s.r.conn.connected {
			s.r.mtx.Unlock()
			continue
		}

		ch, sub, err := s.r.conn.Consume(
			s.opts.Queue,
			s.topic,
			s.headers,
			s.queueArgs,
			s.opts.AutoAck,
			s.durableQueue,
		)

		s.r.mtx.Unlock()
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
			s.r.wg.Add(1)
			s.fn(d)
			s.r.wg.Done()
		}
	}
}

func (r *rbroker) Publish(topic string, msg *Message, opts ...PublishOption) error {
	if r.conn == nil {
		return errors.New("not connected broker")
	}

	m := amqp.Publishing{
		Body:      msg.Body,
		Headers:   amqp.Table{},
		Timestamp: time.Now(),
	}

	options := PublishOptions{}
	for _, o := range opts {
		o(&options)
	}

	if options.Context != nil {
		if value, ok := options.Context.Value(deliveryMode{}).(uint8); ok {
			m.DeliveryMode = value
		}

		if value, ok := options.Context.Value(priorityKey{}).(uint8); ok {
			m.Priority = value
		}
	}

	for k, v := range msg.Header {
		m.Headers[k] = v
	}

	return r.conn.Publish(r.conn.exchange.Name, topic, m)
}

// SubscribePersistentQueue 큐를 지속적으로 유지하기위해 DurableQueue옵션, 데이터 안정성을 위해 DisableAutoAck, AckOnSuccess를 사용한다
// 특히 핸들러 에러 시 메시지의 requeue 여부를 플래그로 전달하며 이 플래그에 따라 RequeueOnError() 옵션을 호출한다
// 이 함수는 topic 을 queue name 으로도 지정한다.
func (r *rbroker) SubscribePersistentQueue(topic string, handler Handler, requeueOnError bool, opts ...SubscribeOption) (Subscriber, error) {
	opts = append(opts, Queue(topic), DurableQueue(), DisableAutoAck())
	if requeueOnError == true {
		opts = append(opts, RequeueOnError())
	}

	return r.subscribe(topic, handler, opts...)
}

// SubscribeTempQueue 임시 큐를 통해 데이터를 수신하기위한 함수
func (r *rbroker) SubscribeTempQueue(topic string, handler Handler, opts ...SubscribeOption) (Subscriber, error) {
	return r.subscribe(topic, handler, opts...)
}

func (r *rbroker) subscribe(topic string, handler Handler, opts ...SubscribeOption) (Subscriber, error) {
	if r.conn == nil {
		return nil, errors.New("not connected broker")
	}

	options := SubscribeOptions{
		Context: context.Background(),
		AutoAck: true,
	}
	for _, o := range opts {
		o(&options)
	}

	var requeueOnError bool
	requeueOnError, _ = options.Context.Value(requeueOnErrorKey{}).(bool)

	var durableQueue bool
	durableQueue, _ = options.Context.Value(durableQueueKey{}).(bool)

	var qArgs map[string]interface{}
	if qa, ok := options.Context.Value(queueArgumentsKey{}).(map[string]interface{}); ok {
		qArgs = qa
	}

	var headers map[string]interface{}
	if h, ok := options.Context.Value(headersKey{}).(map[string]interface{}); ok {
		headers = h
	}

	fn := func(msg amqp.Delivery) {
		header := make(map[string]string)
		for k, v := range msg.Headers {
			header[k], _ = v.(string)
		}
		m := &Message{
			Header:    header,
			Body:      msg.Body,
			TimeStamp: msg.Timestamp,
		}

		p := &publication{d: msg, m: m, t: msg.RoutingKey}

		p.err = handler(p)
		if p.err == nil && !options.AutoAck {
			if err := msg.Ack(false); err != nil {
				logger.Warn("could not acknowledge on a delivery. cause: %v", err)
			}
		} else if p.err != nil && !options.AutoAck {
			if err := msg.Nack(false, requeueOnError); err != nil {
				logger.Warn("could not negatively acknowledge on a delivery. cause: %v", err)
			}
		}
	}

	sret := &subscriber{topic: topic, opts: options, mayRun: true, r: r,
		durableQueue: durableQueue, fn: fn, headers: headers, queueArgs: qArgs}

	go sret.resubscribe()

	return sret, nil
}

func (r *rbroker) Options() Options {
	return r.opts
}

func (r *rbroker) Connect() error {
	conn := r.conn

	if r.conn == nil {
		conn = newRabbitMQConn(r.getExchange(),
			0,
			false,
			r.opts.Registry,
			r.opts.ServiceName,
			r.opts.PersistentQueues,
			r.auth)
	}

	conf := defaultAmqpConfig
	conf.TLSClientConfig = r.opts.TLSConfig

	if err := conn.Connect(r.opts.Secure, &conf); err != nil {
		logger.Errorf("Could not connect to broker(%s) service", r.opts.ServiceName)
		return err
	}

	r.conn = conn
	return nil
}

func (r *rbroker) Disconnect() error {
	if r.conn == nil {
		return errors.New("connection is nil")
	}
	r.mtx.Lock()
	ret := r.conn.Close()
	r.mtx.Unlock()

	r.wg.Wait() // wait all goroutines
	return ret
}

// NewBroker 함수는 브로커 구조체를 생성한다.
func NewBroker(name, id, passwd string, opts ...Option) Broker {
	options := Options{
		Context:     context.Background(),
		ServiceName: name,
	}

	for _, o := range opts {
		o(&options)
	}

	return &rbroker{
		opts: options,
		auth: id + ":" + passwd,
	}
}

func (r *rbroker) getExchange() Exchange {
	ex := Exchange{Name: "micro"}

	if e, ok := r.opts.Context.Value(exchangeKey{}).(string); ok {
		ex.Name = e
	}

	if d, ok := r.opts.Context.Value(durableExchange{}).(bool); ok {
		ex.Durable = d
	}

	return ex
}
