package smplmsg

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Consume will consume message with a given routing key on the client's exchange
func (c *client) Consume(routingKey string) (<-chan amqp.Delivery, <-chan error) {
	var (
		errCh = make(chan error)
		queue = fmt.Sprintf("%s.%s.%s", c.exchange, c.clientID, routingKey)
	)

	err := c.declareQueue(queue, routingKey)
	if err != nil {
		errCh <- errors.Wrap(err, "failed to consume")
		return nil, errCh
	}

	ch, err := c.amqpChannel().Consume(
		queue,
		uuid(),
		false,
		false,
		false,
		false,
		nil,
	)

	go validateMessage(ch, errCh)

	return ch, errCh
}

func validateMessage(delivCh <-chan amqp.Delivery, errCh chan error) {
	for msg := range delivCh {
		err := msg.Ack(false)
		if err != nil {
			errCh <- err
		}

		err = msg.Reject(true)
		if err != nil {
			errCh <- err
		}
	}
}

func (c *client) declareQueue(queue, routingKey string) error {
	q, err := c.amqpChannel().QueueDeclare(
		queue, // name of the queue
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to declare queue (%s)", queue))
	}

	err = c.amqpChannel().QueueBind(
		q.Name,
		routingKey,
		c.exchange,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to bind queue (%s)", queue))

	}

	return nil
}
