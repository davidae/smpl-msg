package smplmsg

import (
	"crypto/rand"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type DeliveryMode uint8
type ContentType string

const (
	Transient  DeliveryMode = 1
	Persistent DeliveryMode = 2

	OctetStream ContentType = "application/octeet-stream"
)

type Client struct {
	uri            string
	exchange       string
	clientID       string
	timeout        time.Duration
	reconnectRetry time.Duration
	amqpCh         atomic.Value
	contentType    ContentType
	deliveryMode   DeliveryMode
}

type amqpCh struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	errCh chan *amqp.Error
}

type clientOption func(c *Client)

func NewSubscriber(msgURI, exchange, routingKey, clientID string, opts ...clientOption) (Subscriber, error) {
	return nil, nil
}

func NewPublisher(msgURI, exchange, routingKey, clientID string, opts ...clientOption) (Publisher, error) {
	c, err := newClient(msgURI, exchange, clientID, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initilize new publisher")
	}

	return c, nil
}

func NewPubSub(msgURI, exchange, routingKey, clientID string, opts ...clientOption) (PublisherSubscriber, error) {
	return nil, nil
}

func newClient(msgqURI, exchange, clientID string, opts ...clientOption) (*Client, error) {
	conn, err := amqp.Dial(msgqURI)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	c := &Client{
		uri:          msgqURI,
		clientID:     clientID,
		exchange:     exchange,
		contentType:  OctetStream,
		deliveryMode: Transient,
	}

	c.amqpCh.Store(&amqpCh{
		conn:  conn,
		ch:    ch,
		errCh: conn.NotifyClose(make(chan *amqp.Error)),
	})

	for _, opt := range opts {
		opt(c)
	}

	err = declareExchange(ch, exchange)
	if err != nil {
		return nil, errors.Wrap(err, "failed to declare exchange")
	}

	return c, nil
}

func (c *Client) amqpCon() *amqp.Connection {
	return c.amqpCh.Load().(amqpCh).conn
}

func (c *Client) amqpChannel() *amqp.Channel {
	return c.amqpCh.Load().(amqpCh).ch
}

func (c *Client) amqpErrChan() chan *amqp.Error {
	return c.amqpCh.Load().(amqpCh).errCh
}

func (c *Client) Close() error {
	if err := c.amqpChannel().Close(); err != nil && !isAmqpErrClosed(err) {
		return fmt.Errorf("smplmsg: closing send channel failed: %s", err)
	}

	if err := c.amqpCon().Close(); err != nil && !isAmqpErrClosed(err) {
		return fmt.Errorf("msgq: closing connection failed: %s", err)
	}

	return nil
}

func isAmqpErrClosed(err error) bool {
	if err == nil {
		return false
	}

	if amqpErr := err.(*amqp.Error); amqpErr == amqp.ErrClosed {
		return true
	}

	return false
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
