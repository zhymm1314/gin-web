// app/consumer/log_consumer.go
package consumer

import (
	"gin-web/config"
	"github.com/rabbitmq/amqp091-go"
	"log"
)

type LogConsumer struct {
	*BaseConsumer
}

// NewLogConsumer 创建一个日志消费者构造方法
func NewLogConsumer(cfg config.RabbitMQ, queueName string) (Consumer, error) {
	base, err := NewBaseConsumer(cfg, queueName)
	if err != nil {
		return nil, err
	}
	return &LogConsumer{base}, nil
}

func (c *LogConsumer) Start() error {
	return c.BaseConsumer.Start(func(msg amqp091.Delivery) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("消息处理崩溃: %v", r)
				msg.Nack(false, true) // 重新入队
			}
		}()
		//todo
		log.Printf("收到日志消息: %s", msg.Body)
		msg.Ack(false)
		//if err := msg.Ack(false); err != nil {
		//	log.Printf("确认消息失败: %v，触发重连", err)
		//	c.reconnect()
		//	return
		//}
	})
}

func (c *LogConsumer) Stop() {
	close(c.stopChan)
	c.channel.Close()
	c.conn.Close()
}
