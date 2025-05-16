// 这是一个rabbitmq的管理类
package bootstrap

import (
	"fmt"
	"gin-web/app/ampq/consumer"
	"gin-web/config"
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

type ConsumerManager struct {
	cfg            *config.AppConfig
	handlers       map[string]consumer.ConsumerHandler
	conn           *amqp.Connection
	consumers      []*Consumer
	wg             sync.WaitGroup
	done           chan struct{}
	reconnectTimer *time.Ticker
}

func NewConsumerManager(cfg *config.AppConfig, handlers map[string]consumer.ConsumerHandler) *ConsumerManager {
	return &ConsumerManager{
		cfg:      cfg,
		handlers: handlers,
		done:     make(chan struct{}),
	}
}

func (cm *ConsumerManager) Start() {
	for {
		select {
		case <-cm.done:
			return
		default:
			if err := cm.connect(); err != nil {
				log.Printf("Connection failed: %v", err)
				cm.reconnect()
				continue
			}

			cm.startConsumers()
			cm.monitorConnection()
			return
		}
	}
}

func (cm *ConsumerManager) connect() error {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cm.cfg.RabbitMQ.Username,
		cm.cfg.RabbitMQ.Password,
		cm.cfg.RabbitMQ.Host,
		cm.cfg.RabbitMQ.Port,
		cm.cfg.RabbitMQ.Vhost,
	))
	if err != nil {
		return err
	}
	cm.conn = conn
	return nil
}

func (cm *ConsumerManager) startConsumers() {
	for _, consumerCfg := range cm.cfg.Consumers {
		handler, ok := cm.handlers[consumerCfg.Handler]
		if !ok {
			log.Printf("Handler %s not registered", consumerCfg.Handler)
			continue
		}

		for i := 0; i < consumerCfg.Concurrency; i++ {
			log.Printf("Starting consumer"+consumerCfg.Handler, i)
			consumer := NewConsumer(cm.conn, consumerCfg.Queue, handler)
			cm.consumers = append(cm.consumers, consumer)
			cm.wg.Add(1)
			go func(c *Consumer) {
				defer cm.wg.Done()
				c.Start()
			}(consumer)
		}
	}
}

func (cm *ConsumerManager) monitorConnection() {
	closeChan := make(chan *amqp.Error)
	cm.conn.NotifyClose(closeChan)

	select {
	case err := <-closeChan:
		log.Printf("Connection closed: %v", err)
		cm.reconnect()
	case <-cm.done:
		return
	}
}

func (cm *ConsumerManager) reconnect() {
	cm.stopConsumers()
	interval := time.Duration(cm.cfg.RabbitMQ.ReconnectInterval) * time.Second
	log.Printf("Reconnecting in %v...", interval)
	time.Sleep(interval)
	cm.Start()
}

func (cm *ConsumerManager) stopConsumers() {
	for _, consumer := range cm.consumers {
		consumer.Stop()
	}
	cm.wg.Wait()
	cm.consumers = nil
}

func (cm *ConsumerManager) Stop() {
	close(cm.done)
	cm.stopConsumers()
	if cm.conn != nil {
		cm.conn.Close()
	}
}
