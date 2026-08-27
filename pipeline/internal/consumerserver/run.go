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
}

func Run(ctx context.Context, cons reader, svc inserter) error {
	for {
		select {
		case <-ctx.Done():
			logger.Info("消费者停止")
			return nil
		default:
		}
		record, err := cons.Read(time.Second)
		if err != nil {
			logger.Error("读取kafka失败: %v", err)
			continue
		}
		if record == nil {
			continue
		}
		if err := svc.Insert(record); err != nil {
			logger.Error("入库失败: domain=%s err=%v", record.Domain, err)
			continue
		}
		logger.Debug("入库成功: domain=%s", record.Domain)
	}
}
