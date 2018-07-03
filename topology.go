package smplmsg

import "github.com/streadway/amqp"

// Subscriber is a message topology which can only consume messages on a given route
type Subscriber interface {
	Consume(routingKey string) (<-chan amqp.Delivery, <-chan error)
	Close() error
}

// Publisher is a message topology which can only publish messages on a given route
type Publisher interface {
	Publish(routingKey string, headers amqp.Table, payload []byte) error
	Close() error
}

// PublisherSubscriber is a message topology which can both publish and consume messages on a given route
type PublisherSubscriber interface {
	Consume(routingKey string) (<-chan amqp.Delivery, <-chan error)
	Publish(routingKey string, headers amqp.Table, payload []byte) error
	Close() error
}
