package capture

import (
	"strings"
	"testing"

	"shared/config"
)

func TestNewPacketCapture_badDevice(t *testing.T) {
	_, err := NewPacketCapture(&config.CollectorConfig{
		Device:  "___no_such_iface___",
		SnapLen: 1600,
	})
	if err == nil {
		t.Fatal("期望打开网卡失败")
	}
	if !strings.Contains(err.Error(), "打开网卡") && !strings.Contains(err.Error(), "___no_such_iface___") {
		t.Fatalf("err=%v", err)
	}
}
