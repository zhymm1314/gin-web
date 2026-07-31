package kafka

import (
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"gin-web/pkg/mq"

	"go.uber.org/zap"
)

// TestKafkaProduceAndConsume 集成测试：通过 pkg/kafka 投递并消费一条消息，验证端到端链路。
//
// 需要可用的 Kafka：默认 127.0.0.1:9092，可用环境变量 KAFKA_BROKERS 覆盖（逗号分隔）。
// 当 Kafka 不可达时自动 t.Skip，因此 `go test ./...` 在无 Kafka 环境下不会失败。
//
// 运行：
//
//	docker compose up -d kafka
//	go test ./pkg/kafka/ -run TestKafkaProduceAndConsume -v
func TestKafkaProduceAndConsume(t *testing.T) {
	brokers := brokersFromEnv()
	if len(brokers) == 0 || !brokerReachable(brokers[0]) {
		t.Skipf("kafka not reachable at %v, skipping integration test", brokers)
	}

	topic := fmt.Sprintf("test-mq-%d", time.Now().UnixNano())
	group := "test-group-" + topic
	body := []byte(fmt.Sprintf("hello-kafka-%d", time.Now().UnixNano()))

	cfg := &Config{
		Brokers:       brokers,
		InitialOffset: "oldest", // 新消费组从头消费，确保能读到投递前的消息
		GroupID:       group,
	}

	// 1. 投递
	prod, err := NewProducer(cfg, topic)
	if err != nil {
		t.Fatalf("create producer failed: %v", err)
	}
	defer prod.Close()

	if err := prod.Publish(body); err != nil {
		t.Fatalf("publish failed: %v", err)
	}
	t.Logf("produced message to topic %q: %s", topic, body)

	// 2. 消费
	received := make(chan *mq.Message, 1)
	handler := mq.HandlerFunc(func(msg *mq.Message) error {
		received <- msg
		return nil
	})

	manager := NewManager(
		cfg,
		[]mq.ConsumerConfig{{Topic: topic, Group: group, Handler: "test", Concurrency: 1}},
		map[string]mq.Handler{"test": handler},
		zap.NewNop(),
	)
	manager.Start()
	defer manager.Stop()

	select {
	case msg := <-received:
		if string(msg.Body) != string(body) {
			t.Fatalf("consumed body mismatch: got %q, want %q", msg.Body, body)
		}
		t.Logf("consumed message from topic %q (partition=%d offset=%d): %s",
			msg.Topic, msg.Partition, msg.Offset, msg.Body)
	case <-time.After(30 * time.Second):
		t.Fatal("timeout waiting for consumed message")
	}
}

// brokersFromEnv 从 KAFKA_BROKERS 读取 broker 列表，默认 127.0.0.1:9092。
func brokersFromEnv() []string {
	raw := os.Getenv("KAFKA_BROKERS")
	if raw == "" {
		raw = "127.0.0.1:9092"
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// brokerReachable 快速检测 broker 端口是否可达。
func brokerReachable(broker string) bool {
	conn, err := net.DialTimeout("tcp", broker, 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
