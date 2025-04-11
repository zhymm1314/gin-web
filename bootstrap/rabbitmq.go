package bootstrap

import (
	"gin-web/app/ampq/consumer"
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
	cfg := global.App.Config.RabbitMQ
	// 示例消费者
	if c, err := consumer.NewLogConsumer(cfg); err == nil {
		consumers = append(consumers, c)
	}

	// 在此添加其他消费者...
}
