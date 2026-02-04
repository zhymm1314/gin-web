package rabbitmq

import (
	"errors"

	"gin-web/app/amqp/consumer"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// Consumer RabbitMQ 消费者
type Consumer struct {
	queueName string
	handler   consumer.ConsumerHandler
	conn      *amqp.Connection
	done      chan struct{}
	log       *zap.Logger
}

// NewConsumer 创建消费者
func NewConsumer(conn *amqp.Connection, queueName string, handler consumer.ConsumerHandler, log *zap.Logger) *Consumer {
	return &Consumer{
		queueName: queueName,
		handler:   handler,
		conn:      conn,
		done:      make(chan struct{}),
		log:       log,
	}
}

// Start 启动消费者
func (c *Consumer) Start() {
	for {
		select {
		case <-c.done:
			return
		default:
			err := c.consume()
			if err != nil {
				c.log.Error("consume error", zap.String("queue", c.queueName), zap.Error(err))
				continue
			}
		}
	}
}

// consume 消费消息
func (c *Consumer) consume() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// 设置预取值
	err = ch.Qos(
		120,   // prefetchCount
		0,     // prefetchSize
		false, // global
	)
	if err != nil {
		c.log.Error("failed to set QoS", zap.Error(err))
		return err
	}

	q, err := ch.QueueDeclare(
		c.queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		q.Name,
		"",    // consumer
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	c.log.Debug("consuming from queue", zap.String("queue", c.queueName))

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("message channel closed")
			}
			if err := c.handler.HandleMessage(msg); err != nil {
				c.log.Error("handle message error",
					zap.String("queue", c.queueName),
					zap.Error(err))
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
			}
		case <-c.done:
			return nil
		}
	}
}

// Stop 停止消费者
func (c *Consumer) Stop() {
	close(c.done)
}
