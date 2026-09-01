package consumerserver

import (
	"context"
	"shared/logger"
	"shared/model"
	"time"
)

type reader interface {
	Read(timeout time.Duration) (*model.DNSQuery, error)
}

type inserter interface {
	Insert(query *model.DNSQuery) error
	InsertBatch(queries []*model.DNSQuery) error
}

func Run(ctx context.Context, cons reader, svc inserter) error {
	const (
		batchSize     = 100
		flushInterval = time.Second
	)

	batch := make([]*model.DNSQuery, 0, batchSize)
	ticker := time.NewTicker(5 * flushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := svc.InsertBatch(batch); err != nil {
			logger.Error("批量入库失败: n=%d err=%v", len(batch), err)
			return
		}
		logger.Debug("批量入库成功: n=%d", len(batch))
		batch = batch[:0]

	}

	for {
		select {
		case <-ctx.Done():
			flush()
			logger.Info("消费者停止")
			return nil

		case <-ticker.C:
			flush()

		default:
			record, err := cons.Read(200 * time.Millisecond)
			if err != nil {
				logger.Error("读取kafka失败: %v", err)
				continue
			}
			if record == nil {
				continue
			}

			batch = append(batch, record)
			if len(batch) >= batchSize {
				flush()
			}
		}
	}
}
