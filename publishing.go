package smplmsg

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Publish will send a message with a given routing key on the client's exchange
func (c *client) Publish(routing string, headers amqp.Table, payload []byte) error {
	pub := amqp.Publishing{
		DeliveryMode: uint8(c.deliveryMode),
		ContentType:  string(c.contentType),
		Body:         payload,
		Headers:      headers,
		MessageId:    uuid(),
		Timestamp:    time.Now(),
	}
	err := c.amqpCh.ch.Publish(
		c.exchange,
		routing,
		false,
		false,
		pub,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to publish message (exchange: %s, routing: %s)", c.exchange, routing))
	}

	return nil
}
