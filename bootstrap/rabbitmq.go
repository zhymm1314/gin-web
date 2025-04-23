package bootstrap

import (
	"fmt"
	"gin-web/app/ampq/consumer"
	"gin-web/config"
	"gin-web/global"
	"log"
	"reflect"
)

var consumers = make([]consumer.Consumer, 0)

func InitRabbitmq() {

	// 注册所有消费者
	registerConsumers()

	// 启动所有消费者
	for _, c := range consumers {
		if err := c.Start(); err != nil {
			log.Fatalf("启动消费者失败: %v", err)
		}
		log.Printf("消费者已启动: %s", reflect.TypeOf(c).String())
	}

	// 保持主线程运行
	select {} // 启动消费者协程
}

// 注册消费者（通过导入包触发init）
func registerConsumers() {
	config, _ := config.LoadQueueConfig()
	for typeName, queueConfig := range config.Queues {
		for a := queueConfig.WorkerNum; a > 0; a-- {
			consumer, err := getConsumerTypes(typeName, queueConfig.Name)
			fmt.Println(err)
			consumers = append(consumers, consumer)
		}

	}
}

// 中间方法
func getConsumerTypes(ct string, queueName string) (consumer.Consumer, error) {
	cfg := global.App.Config.RabbitMQ
	if ct == "log" {
		return consumer.NewLogConsumer(cfg, queueName)
	}

	return nil, fmt.Errorf("consumer type %s not found", ct)
}
