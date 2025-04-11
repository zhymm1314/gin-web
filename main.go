package main

import (
	"fmt"
	"gin-web/bootstrap"
	"gin-web/global"
)

func main() {
	a := 1
	b := 2
	fmt.Println(a + b)

	// 初始化配置
	bootstrap.InitializeConfig()

	// 初始化日志
	global.App.Log = bootstrap.InitializeLog()
	global.App.Log.Info("log init success!")

	// 初始化数据库
	global.App.DB = bootstrap.InitializeDB()
	// 程序关闭前，释放数据库连接
	defer func() {
		if global.App.DB != nil {
			db, _ := global.App.DB.DB()
			db.Close()
		}
	}()
	bootstrap.InitializeValidator()

	// 初始化Redis
	global.App.Redis = bootstrap.InitializeRedis()

	//携程去启动消费者，暂时有点问题后续优化一下
	go bootstrap.InitRabbitmq()
	// 启动服务器
	bootstrap.RunServer()

}

//func sendmessage() {
//	cfg := config.RabbitMQ{
//		Host:     "10.10.65.54",
//		Port:     5672,
//		Username: "magento",
//		Password: "123456",
//		Vhost:    "/saas-tenant",
//	}
//	client := bootstrap.NewRabbitClient(cfg)
//	defer client.Close()
//
//	// 等待连接就绪（最多等待15秒）
//	if err := client.WaitForConnection(15 * time.Second); err != nil {
//		log.Fatal("等待连接超时:", err)
//	}
//
//	// 发送消息
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	if err := client.Publish(ctx, "default", "base.multi.new.check.line", []byte(`{"msg":"test"}`)); err != nil {
//		log.Fatal("发送失败:", err)
//	}
//
//	log.Println("发送成功")
//	log.Println("消息发送成功")
//}
//
////func startConsumer(client *bootstrap.RabbitClient) {
////	// 声明队列
////	queue, err := client.DeclareQueue("base.multi.new.check.line", true, false, false, nil)
////	if err != nil {
////		println(err)
////	}
////
////	// 开始消费
////	msgs, err := client.Consume(
////		queue.Name,
////		"",    // consumer
////		false, // auto-ack
////		false, // exclusive
////		nil,
////	)
////	if err != nil {
////		println(err)
////	}
////
////	for msg := range msgs {
////		fmt.Printf("Received message: %s\n", msg.Body)
////		msg.Ack(false)
////	}
////}
//
//func main() {
//
//	// 连接字符串使用 URL 编码的 vhost
//	conn, err := amqp.Dial("amqp://magento:123456@10.10.65.54:5672/%2Fsaas-tenant")
//	if err != nil {
//		log.Fatalf("无法连接到 RabbitMQ: %v", err)
//	}
//	defer conn.Close()
//
//	// 创建 Channel
//	ch, err := conn.Channel()
//	if err != nil {
//		log.Fatalf("无法打开 Channel: %v", err)
//	}
//	defer ch.Close()
//
//	// 声明队列
//	queueName := "test_queue"
//	_, err = ch.QueueDeclare(
//		queueName, // 队列名称
//		true,      // 持久化
//		false,     // 自动删除
//		false,     // 排他性
//		false,     // 不等待
//		nil,       // 参数
//	)
//	if err != nil {
//		log.Fatalf("声明队列失败: %v", err)
//	}
//
//	log.Println("✅ 成功连接到 RabbitMQ 并准备好队列")
//
//	// 初始化 Gin
//	r := gin.Default()
//
//	err = ch.PublishWithContext(
//		context.Background(),
//		"",        // 默认交换机
//		queueName, // 路由键
//		false,     // 强制标志
//		false,     // 立即标志
//		amqp.Publishing{
//			ContentType:  "text/plain",
//			DeliveryMode: amqp.Persistent,
//		},
//	)
//
//	if err != nil {
//		log.Printf("❌ 消息发送失败: %v", err)
//		return
//	}
//
//	// 消息发送接口
//	r.GET("/send", func(c *gin.Context) {
//		msg := "Hello RabbitMQ at " + time.Now().Format(time.RFC3339)
//
//		err = ch.PublishWithContext(
//			context.Background(),
//			"",        // 默认交换机
//			queueName, // 路由键
//			false,     // 强制标志
//			false,     // 立即标志
//			amqp.Publishing{
//				ContentType:  "text/plain",
//				Body:         []byte(msg),
//				DeliveryMode: amqp.Persistent,
//			},
//		)
//
//		if err != nil {
//			log.Printf("❌ 消息发送失败: %v", err)
//			c.JSON(http.StatusInternalServerError, gin.H{"error": "消息发送失败"})
//			return
//		}
//
//		log.Printf("✔️ 消息已发送: %s", msg)
//		c.JSON(http.StatusOK, gin.H{"message": "消息发送成功", "data": msg})
//	})
//
//	log.Println("🚀 启动 Gin 服务在 :8080")
//	r.Run(":8080")
//}
