package consumer

import "github.com/streadway/amqp"

type ConsumerHandler interface {
	HandleMessage(msg amqp.Delivery) error
}
