package kafka

import (
	"encoding/json"
	"fmt"
	"strings"

	"shared/config"
	"shared/model"

	confluent "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Producer struct {
	p     *confluent.Producer
	topic string
}

func NewProducer(cfg config.KafkaConfig) (*Producer, error) {
	p, err := confluent.NewProducer(&confluent.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.Brokers, ","),
		"acks":              "all",
	})
	if err != nil {
		return nil, fmt.Errorf("创建生产者失败：%w", err)
	}
	return &Producer{p: p, topic: cfg.Topic}, nil

}

func (p *Producer) Send(query *model.DNSQuery) error {
	data, err := json.Marshal(query)
	if err != nil {
		return err
	}
	return p.p.Produce(&confluent.Message{
		TopicPartition: confluent.TopicPartition{
			Topic:     &p.topic,
			Partition: confluent.PartitionAny,
		},
		Value: data,
	}, nil)
}
func (prod *Producer) Close() {
	prod.p.Flush(5000) // 等待未发送完的消息
	prod.p.Close()
}
