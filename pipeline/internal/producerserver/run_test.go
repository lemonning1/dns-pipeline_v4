package producerserver

import (
	"context"
	"errors"
	"testing"
	"time"

	"shared/model"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type fakeCap struct {
	ch chan gopacket.Packet
}

func (f fakeCap) Packets() <-chan gopacket.Packet { return f.ch }

type fakeProd struct {
	err   error
	n     int
	last  string
}

func (f *fakeProd) Send(q *model.DNSQuery) error {
	f.n++
	if q != nil {
		f.last = q.Domain
	}
	return f.err
}

func TestRun_cancel(t *testing.T) {
	ch := make(chan gopacket.Packet)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	if err := Run(ctx, fakeCap{ch}, &fakeProd{}); err != nil {
		t.Fatal(err)
	}
}

func TestRun_channelClosed(t *testing.T) {
	ch := make(chan gopacket.Packet)
	close(ch)
	if err := Run(context.Background(), fakeCap{ch}, &fakeProd{}); err != nil {
		t.Fatal(err)
	}
}

func TestRun_sendAndSkipEmpty(t *testing.T) {
	ch := make(chan gopacket.Packet, 2)
	empty := gopacket.NewPacket(nil, layers.LayerTypeEthernet, gopacket.Default)
	dns := &layers.DNS{
		QR: false,
		Questions: []layers.DNSQuestion{
			{Name: []byte("hit.example.com"), Type: layers.DNSTypeA, Class: layers.DNSClassIN},
		},
	}
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf, gopacket.SerializeOptions{FixLengths: true}, dns); err != nil {
		t.Fatal(err)
	}
	dnsPkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeDNS, gopacket.Default)

	ch <- empty
	ch <- dnsPkt
	close(ch)

	prod := &fakeProd{err: errors.New("kafka down")}
	if err := Run(context.Background(), fakeCap{ch}, prod); err != nil {
		t.Fatal(err)
	}
	if prod.n != 1 || prod.last != "hit.example.com" {
		t.Fatalf("send calls=%d last=%s", prod.n, prod.last)
	}
}

func TestRun_sendOK(t *testing.T) {
	ch := make(chan gopacket.Packet, 1)
	dns := &layers.DNS{
		QR: false,
		Questions: []layers.DNSQuestion{
			{Name: []byte("ok.example.com"), Type: layers.DNSTypeA, Class: layers.DNSClassIN},
		},
	}
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf, gopacket.SerializeOptions{FixLengths: true}, dns); err != nil {
		t.Fatal(err)
	}
	ch <- gopacket.NewPacket(buf.Bytes(), layers.LayerTypeDNS, gopacket.Default)
	close(ch)
	prod := &fakeProd{}
	if err := Run(context.Background(), fakeCap{ch}, prod); err != nil {
		t.Fatal(err)
	}
	if prod.n != 1 {
		t.Fatalf("n=%d", prod.n)
	}
}
