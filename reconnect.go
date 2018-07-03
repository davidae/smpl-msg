package smplmsg

import (
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

func (c *client) monitorConnection() {
	for {
		select {
		case <-c.endMonitoring:
			return
		case <-c.amqpCh.errCh:
			err := c.reconnect()
			if err != nil {
				time.Sleep(c.retryTimeout)

				err := c.reconnect()
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func (c *client) reconnect() error {
	conn, err := amqp.Dial(c.uri)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return errors.Wrap(err, "failed to open channel")
	}

	c.amqpCh = &amqpCh{
		conn:  conn,
		ch:    ch,
		errCh: conn.NotifyClose(make(chan *amqp.Error)),
	}

	err = declareExchange(ch, c.exchange)
	if err != nil {
		return errors.Wrap(err, "failed to redeclare exchange")
	}

	return nil
}
