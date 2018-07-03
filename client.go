package smplmsg

import (
	"crypto/rand"
	"io"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// DeliveryMode is the type of MIME type used to delivery of messages
type DeliveryMode uint8

// ContentType is the type of MIME type used for the content of the messages
type ContentType string

const (
	// Transient means higher throughput but messages will not be restored on broker restart
	Transient DeliveryMode = 1

	// Persistent means lower throughput but messages will be restored on broker restart
	Persistent DeliveryMode = 2

	// OctetStream is a MIME type used as a content type
	OctetStream ContentType = "application/octeet-stream"

	defaultRetryTimeout = 5 * time.Second
)

type client struct {
	uri      string
	exchange string
	clientID string

	errorCh       chan error
	endMonitoring chan struct{}
	timeout       time.Duration
	retryTimeout  time.Duration
	amqpCh        atomic.Value
	contentType   ContentType
	deliveryMode  DeliveryMode
}

type amqpCh struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	errCh chan *amqp.Error
}

// NewSubscriber initializes and returns a Client that implements the Subscriber interface
func NewSubscriber(msgURI, exchange, routingKey, clientID string, opts ...ClientOption) (Subscriber, error) {
	c, err := newClient(msgURI, exchange, clientID, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initilize new subscriber")
	}

	return c, nil
}

// NewPublisher initializes and returns a Client that implements the Publisher interface
func NewPublisher(msgURI, exchange, routingKey, clientID string, opts ...ClientOption) (Publisher, error) {
	c, err := newClient(msgURI, exchange, clientID, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initilize new publisher")
	}

	return c, nil
}

// NewPubSub initializes and returns a Client that implements the PublisherSubscriber interface
func NewPubSub(msgURI, exchange, routingKey, clientID string, opts ...ClientOption) (PublisherSubscriber, error) {
	c, err := newClient(msgURI, exchange, clientID, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initilize new pub/sub")
	}

	return c, nil
}

func newClient(msgqURI, exchange, clientID string, opts ...ClientOption) (*client, error) {
	conn, err := amqp.Dial(msgqURI)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	c := &client{
		uri:           msgqURI,
		clientID:      clientID,
		exchange:      exchange,
		contentType:   OctetStream,
		deliveryMode:  Transient,
		retryTimeout:  defaultRetryTimeout,
		endMonitoring: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.amqpCh.Store(&amqpCh{
		conn:  conn,
		ch:    ch,
		errCh: conn.NotifyClose(make(chan *amqp.Error)),
	})

	err = declareExchange(ch, exchange)
	if err != nil {
		return nil, errors.Wrap(err, "failed to declare exchange")
	}

	go c.monitorConnection()

	return c, nil
}

func (c *client) amqpCon() *amqp.Connection {
	return c.amqpCh.Load().(amqpCh).conn
}

func (c *client) amqpChannel() *amqp.Channel {
	return c.amqpCh.Load().(amqpCh).ch
}

func (c *client) amqpErrChan() chan *amqp.Error {
	return c.amqpCh.Load().(amqpCh).errCh
}

// Close closes the channels
func (c *client) Close() error {
	close(c.endMonitoring)

	if err := c.amqpChannel().Close(); err != nil {
		return errors.Wrap(err, "failed to close channel")
	}

	if err := c.amqpCon().Close(); err != nil {
		return errors.Wrap(err, "failed to close connection")
	}

	return nil
}

func declareExchange(ch *amqp.Channel, exchange string) error {
	err := ch.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		true,     // durable
		true,     // auto-deleted
		false,    // internal
		false,    // noWait
		nil,      // arguments
	)
	if err != nil {
		return errors.Wrap(err, "failed to declare exchange")
	}

	return nil
}

func uuid() string {
	fillWithRandomBits := func(b []byte) {
		if _, err := io.ReadFull(rand.Reader, b); err != nil {
			panic(err)
		}
	}

	uuid := make([]byte, 16)
	fillWithRandomBits(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10
	return string(uuid)
}
