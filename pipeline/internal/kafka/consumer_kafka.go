package kafka

import (
	"encoding/json"
	"fmt"
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

func NewConsumer(cfg *config.KafkaConfig) (*Consumer, error) {
	c, err := confluent.NewConsumer(&confluent.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.Brokers, ","),
		"group.id":          cfg.GroupID,
		"auto.offset.reset": cfg.AutoOffsetReset,
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

func (cons *Consumer) Read(timeout time.Duration) (*model.DNSQuery, error) {
	msg, err := cons.c.ReadMessage(timeout)
	if err != nil {
		if kafkaErr, ok := err.(confluent.Error); ok && kafkaErr.Code() == confluent.ErrTimedOut {
			return nil, nil
		} //超时检验
		return nil, err
	}
	var record model.DNSQuery
	if err := json.Unmarshal(msg.Value, &record); err != nil {
		return nil, fmt.Errorf("反序列化失败:%w", err)
	}
	return &record, nil
}

func (cons *Consumer) Close() {
	if cons.c != nil {
		cons.c.Close()
	}
}
