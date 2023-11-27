package broker

import (
	"errors"
	"github.com/streadway/amqp"
	"time"
)

// Broker is an interface used for asynchronous messaging.
type Broker interface {
	Options() Options
	Connect() error
	Disconnect() error
	Publish(topic string, m *Message, opts ...PublishOption) error
	// SubscribePersistentQueue cdm-cloud 에서는 Persistent Queueu와 Temp Queue를 사용함에 따라 기존의 Subscribe인터페이스는 삭제하고
	// 아래 두 인터페이스를 추가하였다
	SubscribePersistentQueue(topic string, h Handler, requeueOnError bool, opts ...SubscribeOption) (Subscriber, error)
	SubscribeTempQueue(topic string, h Handler, opts ...SubscribeOption) (Subscriber, error)
}

// Handler is used to process messages via a subscription of a topic.
// The handler is passed a publication interface which contains the
// message and optional Ack method to acknowledge receipt of the message.
type Handler func(Event) error

// Message structure for broker
type Message struct {
	Header    map[string]string
	Body      []byte
	TimeStamp time.Time
}

// Event is given to a subscription handler for processing
type Event interface {
	Topic() string
	Message() *Message
}

// Subscriber is a convenience return type for the Subscribe method
type Subscriber interface {
	Options() SubscribeOptions
	Topic() string
	Message() func(msg amqp.Delivery)
	Unsubscribe() error
}

// DefaultBroker 는 기본으로 사용되는 브로커를 저장하기위한 변수이다
var DefaultBroker Broker

// Init 함수는 Default Broker를 사용하기위한 초기화 함수
func Init(name, id, passwd string, opts ...Option) error {
	DefaultBroker = NewBroker(name, id, passwd, opts...)
	if DefaultBroker == nil {
		return errors.New("could not create default broker")
	}
	return nil
}

// Connect 함수는 Default Broker의 연결 함수
func Connect() error {
	return DefaultBroker.Connect()
}

// Disconnect 함수는 Default Broker의 연결 종료 함수
func Disconnect() error {
	err := DefaultBroker.Disconnect()
	DefaultBroker = nil
	return err
}

// Publish 함수는 Default Broker의 데이터 전송
func Publish(topic string, msg *Message, opts ...PublishOption) error {
	return DefaultBroker.Publish(topic, msg, opts...)
}

// SubscribePersistentQueue 함수는 지속 가능한 큐를 통해 데이터를 전송하기위한 함수. (cdm-cloud를 위해 추가)
// 지속 가능한 큐는 Subscriber가 존재하지 않을 경우 메시지를 큐에 저장하고 Subscriber가 연결되면 쌓인 메시지도 함께 전달한다.
func SubscribePersistentQueue(topic string, handler Handler, requeueOnError bool, opts ...SubscribeOption) (Subscriber, error) {
	return DefaultBroker.SubscribePersistentQueue(topic, handler, requeueOnError, opts...)
}

// SubscribeTempQueue 함수는 임시 큐를 통해 데이터를 전송하기 위한 함수. (cdm-cloud를 위해 추가)
// 임시 큐는 Subscriber가 존재하지 않을 경우 메시지를 쌓지않고 버리며, Subscriber가 연결된 이후부터의 메시지를 전달한다.
func SubscribeTempQueue(topic string, handler Handler, opts ...SubscribeOption) (Subscriber, error) {
	return DefaultBroker.SubscribeTempQueue(topic, handler, opts...)
}
