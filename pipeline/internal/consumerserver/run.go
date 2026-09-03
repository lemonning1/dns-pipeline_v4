package consumerserver

import (
	"context"
	"pipeline/internal/kafka"
	"shared/logger"
	"shared/model"
	"time"

	confluent "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// kafkaClient：从 Kafka 读消息，并在入库成功后提交 offset。
type kafkaClient interface {
	Read(timeout time.Duration) (*kafka.ConsumedMessage, error)
	Commit(msg *confluent.Message) error
}

type inserter interface {
	Insert(query *model.DNSQuery) error
	InsertBatch(queries []*model.DNSQuery) error
}

type batchItem struct {
	record *model.DNSQuery
	msg    *confluent.Message
}

func Run(ctx context.Context, cons kafkaClient, svc inserter) error {
	const (
		batchSize     = 100
		flushInterval = time.Second
	)

	batch := make([]batchItem, 0, batchSize)
	ticker := time.NewTicker(5 * flushInterval)
	defer ticker.Stop()

	// flush：先批量写 ClickHouse，成功后再按分区 commit Kafka offset。
	flush := func() {
		if len(batch) == 0 {
			return
		}
		records := make([]*model.DNSQuery, 0, len(batch))
		for _, it := range batch {
			records = append(records, it.record)
		}
		if err := svc.InsertBatch(records); err != nil {
			logger.Errorf("批量入库失败: n=%d err=%v", len(batch), err)
			return // 不写库就不 commit，下次重启会重读这批消息
		}
		// 一个 batch 可能混多个分区；每个分区 commit 本批里该分区「最后一条」的 offset。
		lastByPartition := make(map[int32]*confluent.Message)
		for _, it := range batch {
			lastByPartition[it.msg.TopicPartition.Partition] = it.msg
		}
		for _, msg := range lastByPartition {
			if err := cons.Commit(msg); err != nil {
				logger.Errorf("commit offset 失败: %v", err)
				return // commit 失败也不清空 batch，便于重试
			}
		}
		logger.Debugf("批量入库并 commit 成功: n=%d", len(batch))
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			logger.Infof("消费者停止")
			return nil

		case <-ticker.C:
			flush()

		default:
			cm, err := cons.Read(200 * time.Millisecond)
			if err != nil {
				logger.Errorf("读取kafka失败: %v", err)
				continue
			}
			if cm == nil {
				continue
			}

			batch = append(batch, batchItem{record: cm.Record, msg: cm.KafkaMsg})
			if len(batch) >= batchSize {
				flush()
			}
		}
	}
}
