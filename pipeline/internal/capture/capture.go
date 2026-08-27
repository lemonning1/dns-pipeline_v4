package capture

import (
	"fmt"
	"sync"
	"time"

	"shared/config"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type PacketCapture struct {
	handle *pcap.Handle
	source *gopacket.PacketSource
	once   sync.Once
}

func NewPacketCapture(cfg *config.CollectorConfig) (*PacketCapture, error) {
	device := cfg.Device
	if device == "" {
		fmt.Println("未配置网卡，开始自动配置")
		devices, err := pcap.FindAllDevs()
		if err != nil {
			fmt.Println("查找网卡失败:", err)
			return nil, err
		}
		for _, d := range devices {
			if d.Name != "lo" && len(d.Addresses) > 0 {
				device = d.Name
				break
			}
		}
		if device == "" {
			return nil, fmt.Errorf("未找到可用的网卡")
		}
	}
	fmt.Println("配置网卡:", device)

	snaplen := cfg.SnapLen
	if snaplen <= 0 {
		snaplen = 1600
	}
	handle, err := pcap.OpenLive(device, snaplen, true, 100*time.Millisecond)
	if err != nil {
		return nil, fmt.Errorf("打开网卡%s失败: %w", device, err)
	}

	if cfg.Filter != "" {
		if err := handle.SetBPFFilter(cfg.Filter); err != nil {
			handle.Close()
			return nil, fmt.Errorf("设置 BPF 过滤器失败: %w", err)
		}
	}

	return &PacketCapture{
		handle: handle,
		source: gopacket.NewPacketSource(handle, handle.LinkType()),
	}, nil
}
func (pc *PacketCapture) Packets() <-chan gopacket.Packet {
	return pc.source.Packets()
}

func (pc *PacketCapture) Close() {
	pc.once.Do(func() {
		if pc.handle != nil {
			pc.handle.Close()
			pc.handle = nil
		}
	})
}
