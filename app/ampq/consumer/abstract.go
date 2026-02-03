package consumer

import amqp "github.com/rabbitmq/amqp091-go"

type ConsumerHandler interface {
	HandleMessage(msg amqp.Delivery) error
}
