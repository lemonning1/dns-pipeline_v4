package producerserver

import (
	"context"
	"pipeline/internal/parse"
	"shared/logger"
	"shared/model"

	"github.com/google/gopacket"
)

type packetSource interface {
	Packets() <-chan gopacket.Packet
}

type sender interface {
	Send(query *model.DNSQuery) error
}

func Run(ctx context.Context, cap packetSource, prod sender) error {
	packets := cap.Packets()
	jobs := make(chan gopacket.Packet, 128)
	go func() {

		for packet := range jobs {
			processPacket(prod, packet)
		}
	}()

	defer close(jobs)
	for {
		select {
		case <-ctx.Done():
			logger.Info("采集器暂停")
			return nil

		case packet, ok := <-packets:
			if !ok {
				logger.Info("抓包通道已关闭")
				return nil
			}

			select {
			case jobs <- packet:
			case <-ctx.Done():
				logger.Info("采集器暂停")
				return nil
			}
		}
	}
}
func processPacket(prod sender, packet gopacket.Packet) {
	records := parse.FromPacket(packet)
	for _, record := range records {
		if record == nil {
			continue
		}
		if err := prod.Send(record); err != nil {
			logger.Error("发送至kafka失败: domain=%s err=%v", record.Domain, err)
			continue
		}
		logger.Debug("已发送 record: domain=%s", record.Domain)
	}
}
