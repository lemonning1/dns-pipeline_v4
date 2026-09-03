package kafka

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"shared/config"
	"shared/model"

	confluent "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Consumer struct {
	c     *confluent.Consumer
	topic string
}

// ConsumedMessage 把「业务数据」和「Kafka 原消息」绑在一起。
// 防重复入库时：先用 Record 写库，成功后再用 KafkaMsg 去 commit offset。
type ConsumedMessage struct {
	Record   *model.DNSQuery
	KafkaMsg *confluent.Message
}

func NewConsumer(cfg *config.KafkaConfig) (*Consumer, error) {
	c, err := confluent.NewConsumer(&confluent.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.Brokers, ","),
		"group.id":          cfg.GroupID,
		"auto.offset.reset": cfg.AutoOffsetReset,
		// 关闭自动 commit：offset 不再由客户端后台悄悄提交，
		// 改由我们在 ClickHouse 批量写入成功后，显式 CommitMessage。
		"enable.auto.commit": false,
	})

	if err != nil {
		return nil, fmt.Errorf("创建消费者失败:%w", err)
	}

	if err := c.SubscribeTopics([]string{cfg.Topic}, nil); err != nil {
		c.Close()
		return nil, fmt.Errorf("订阅 topic 失败:%w", err)
	}
	return &Consumer{c: c, topic: cfg.Topic}, nil
}

func (cons *Consumer) Read(timeout time.Duration) (*ConsumedMessage, error) {
	msg, err := cons.c.ReadMessage(timeout)
	if err != nil {
		if kafkaErr, ok := err.(confluent.Error); ok && kafkaErr.Code() == confluent.ErrTimedOut {
			return nil, nil
		}
		return nil, err
	}

	var record model.DNSQuery
	if err := json.Unmarshal(msg.Value, &record); err != nil {
		return nil, fmt.Errorf("反序列化失败:%w", err)
	}

	tp := msg.TopicPartition
	record.ID = idFromKafkaPosition(tp.Partition, int64(tp.Offset))
	return &ConsumedMessage{Record: &record, KafkaMsg: msg}, nil
}

func (cons *Consumer) Commit(msg *confluent.Message) error {
	_, err := cons.c.CommitMessage(msg)
	return err
}

func (cons *Consumer) Close() {
	if cons.c != nil {
		cons.c.Close()
	}
}

func idFromKafkaPosition(partition int32, offset int64) int {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%d:%d", partition, offset)
	return int(h.Sum64() & 0x7fffffffffffffff)
}
