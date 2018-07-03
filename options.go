package smplmsg

import (
	"time"
)

// ClientOption is a type of function that allows users to set options for the client
type ClientOption func(c *client)

// SetDeliveryMode sets the delivery mode
func SetDeliveryMode(d uint8) ClientOption {
	return func(c *client) {
		c.deliveryMode = DeliveryMode(d)
	}
}

// SetContentType sets the content type
func SetContentType(t string) ClientOption {
	return func(c *client) {
		c.contentType = ContentType(t)
	}
}

// SetRetryTimeout sets the retry timeout
func SetRetryTimeout(d time.Duration) ClientOption {
	return func(c *client) {
		c.retryTimeout = d
	}
}
