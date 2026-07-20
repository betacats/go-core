package kafkax

import (
	"context"
	"os"
	"testing"

	"github.com/segmentio/kafka-go"
)

// skipInCI 在 CI 环境中跳过需要真实 Kafka 连接的集成测试
func skipInCI(t *testing.T) {
	if os.Getenv("CI") == "true" {
		t.Skip("跳过 CI 环境中的 Kafka 集成测试")
	}
}

func TestKafkaProducerPublish(t *testing.T) {
	skipInCI(t)
	cfg := &KafkaConfig{
		Username: "test",
		Password: "test",
		Brokers:  "localhost:9092",
	}
	topic := "test-topic"
	ctx := context.Background()
	// 先初始化 producer
	InitProducerForTopics(ctx, cfg, []string{topic})

	// 获取指定 topic 的 producer
	producer, err := GetProducerByTopic(topic)
	if err != nil {
		t.Fatalf("获取 producer 失败: %v", err)
	}
	// 关闭连接的方法（建议在程序优雅关闭后调用）
	defer CloseAllProducers()

	msg := kafka.Message{Value: []byte("test message")}
	err = producer.Publish(ctx, []kafka.Message{msg})
	if err != nil {
		t.Fatalf("消息发送失败: %v", err)
	} else {
		t.Log("消息发送成功")
	}
}

func TestGetProducerByTopicCompatibility(t *testing.T) {
	topic := "test-topic"
	kafkaName := "cluster-a"

	legacyProducer := &KafkaProducer{}
	namedProducer := &KafkaProducer{}

	producerPool.Store(getPoolKey("", topic), legacyProducer)
	producerPool.Store(getPoolKey(kafkaName, topic), namedProducer)
	defer producerPool.Delete(getPoolKey("", topic))
	defer producerPool.Delete(getPoolKey(kafkaName, topic))

	gotLegacy, err := GetProducerByTopic(topic)
	if err != nil {
		t.Fatalf("旧方法获取 producer 失败: %v", err)
	}
	if gotLegacy != legacyProducer {
		t.Fatalf("旧方法返回了错误的 producer")
	}

	gotNamed, err := GetProducerByTopicWithKafkaName(kafkaName, topic)
	if err != nil {
		t.Fatalf("新方法获取 producer 失败: %v", err)
	}
	if gotNamed != namedProducer {
		t.Fatalf("新方法返回了错误的 producer")
	}
}
