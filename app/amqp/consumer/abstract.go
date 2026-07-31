package consumer

import "gin-web/pkg/mq"

// ConsumerHandler 消息处理器接口（与底层 MQ 实现解耦）。
// 类型别名指向 mq.Handler，历史代码可继续使用 consumer.ConsumerHandler。
type ConsumerHandler = mq.Handler
