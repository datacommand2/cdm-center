package broker

type durableQueueKey struct{}
type headersKey struct{}
type queueArgumentsKey struct{}
type exchangeKey struct{}
type requeueOnErrorKey struct{}
type deliveryMode struct{}
type priorityKey struct{}
type durableExchange struct{}

// DurableQueue creates a durable queue when subscribing.
func DurableQueue() SubscribeOption {
	return setSubscribeOption(durableQueueKey{}, true)
}

// DurableExchange is an option to set the Exchange to be durable
func DurableExchange() Option {
	return setBrokerOption(durableExchange{}, true)
}

// Headers adds headers used by the headers exchange
func Headers(h map[string]interface{}) SubscribeOption {
	return setSubscribeOption(headersKey{}, h)
}

// QueueArguments sets arguments for queue creation
func QueueArguments(h map[string]interface{}) SubscribeOption {
	return setSubscribeOption(queueArgumentsKey{}, h)
}

// RequeueOnError calls Nack(muliple:false, requeue:true) on amqp delivery when handler returns error
func RequeueOnError() SubscribeOption {
	return setSubscribeOption(requeueOnErrorKey{}, true)
}

// ExchangeName is an option to set the ExchangeName
func ExchangeName(e string) Option {
	return setBrokerOption(exchangeKey{}, e)
}

// DeliveryMode sets a delivery mode for publishing
func DeliveryMode(value uint8) PublishOption {
	return setPublishOption(deliveryMode{}, value)
}

// Priority sets a priority level for publishing
func Priority(value uint8) PublishOption {
	return setPublishOption(priorityKey{}, value)
}
