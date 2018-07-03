package smplmsg

import "github.com/streadway/amqp"

type Subscriber interface {
	Close() error
	Consume(routingKey string) (<-chan amqp.Delivery, <-chan error)
}

type Publisher interface {
	Close() error
	Publish(routingKey string, headers amqp.Table, payload []byte) error
}

type PublisherSubscriber interface {
	Close() error
	Consume(routingKey string) (<-chan amqp.Delivery, <-chan error)
	Publish(routingKey string, headers amqp.Table, payload []byte) error
}
