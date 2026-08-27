package parse

import (
	"net"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func TestParseDNSQuestion_query(t *testing.T) {
	dns := &layers.DNS{QR: false}
	q := &layers.DNSQuestion{
		Name: []byte("example.com."),
		Type: layers.DNSTypeA,
	}

	got := ParseDNSQuestion(dns, q)
	if got == nil {
		t.Fatal("期望得到record,得到了nil")
	}
	if got.Domain != "example.com" {
		t.Fatalf("Domain解析异常=%q", got.Domain)
	}
	if got.QR != 0 {
		t.Fatalf("QR解析异常=%d", got.QR)
	}
	if got.QType != int(layers.DNSTypeA) {
		t.Fatalf("QType解析异常=%d", got.QR)
	}
}

func TestParseDNSQuestion_response(t *testing.T) {
	dns := &layers.DNS{
		QR:           true,
		ResponseCode: layers.DNSResponseCodeNoErr,
		Answers: []layers.DNSResourceRecord{
			{Type: layers.DNSTypeCNAME, CNAME: []byte("cdn1.example.com.")},
			{Type: layers.DNSTypeCNAME, CNAME: []byte("cdn2.example.com.")},
			{Type: layers.DNSTypeA, IP: net.ParseIP("1.2.3.4"), TTL: 300},
			{Type: layers.DNSTypeA, IP: net.ParseIP("1.2.3.43"), TTL: 600},
		},
	}
	q := &layers.DNSQuestion{
		Name: []byte("example.com."),
		Type: layers.DNSTypeA,
	}
	got := ParseDNSQuestion(dns, q)
	if got == nil {
		t.Fatal("期望得到record,得到了nil")
	}
	if got.Domain != "example.com" {
		t.Fatalf("Domain解析异常=%q", got.Domain)
	}
	if got.QR != 1 {
		t.Fatalf("QR解析异常=%d", got.QR)
	}
	if got.QType != int(layers.DNSTypeA) {
		t.Fatalf("QType解析异常=%d", got.QR)
	}
	if got.Cnamechain != "cdn1.example.com,cdn2.example.com" {
		t.Fatalf("Cnamechain=%q", got.Cnamechain)
	}
	if got.ResponseIPs != "1.2.3.4,1.2.3.43" {
		t.Fatalf("ResponseIPs=%q", got.ResponseIPs)
	}
	if got.TTL != 300 {
		t.Fatalf("TTL=%d", got.TTL)
	}
}

func TestParseDNSQuestion_aaaa(t *testing.T) {
	ip := net.ParseIP("2001:db8::1")
	dns := &layers.DNS{
		QR: true,
		Answers: []layers.DNSResourceRecord{
			{Type: layers.DNSTypeAAAA, IP: ip, TTL: 60},
			{Type: layers.DNSTypeA, TTL: 10},
		},
	}
	q := &layers.DNSQuestion{Name: []byte("v6.example.com."), Type: layers.DNSTypeAAAA}
	got := ParseDNSQuestion(dns, q)
	if got.ResponseIPs != ip.String() {
		t.Fatalf("ResponseIPs=%q", got.ResponseIPs)
	}
	if got.TTL != 60 {
		t.Fatalf("TTL=%d", got.TTL)
	}
}

func TestFromPacket_noDNS(t *testing.T) {
	pkt := gopacket.NewPacket(nil, layers.LayerTypeEthernet, gopacket.Default)
	if records := FromPacket(pkt); records != nil {
		t.Fatalf("无 DNS 时应为 nil, got %#v", records)
	}
}

func TestFromPacket_withDNS(t *testing.T) {
	dns := &layers.DNS{
		QR: false,
		Questions: []layers.DNSQuestion{
			{Name: []byte("a.example.com"), Type: layers.DNSTypeA, Class: layers.DNSClassIN},
		},
	}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	if err := gopacket.SerializeLayers(buf, opts, dns); err != nil {
		t.Fatal(err)
	}
	pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeDNS, gopacket.Default)
	records := FromPacket(pkt)
	if len(records) != 1 {
		t.Fatalf("len=%d", len(records))
	}
	if records[0].Domain != "a.example.com" || records[0].QR != 0 {
		t.Fatalf("%+v", records[0])
	}
}

func TestFromPacket_noQuestions(t *testing.T) {
	dns := &layers.DNS{QR: false}
	buf := gopacket.NewSerializeBuffer()
	if err := gopacket.SerializeLayers(buf, gopacket.SerializeOptions{FixLengths: true}, dns); err != nil {
		t.Fatal(err)
	}
	pkt := gopacket.NewPacket(buf.Bytes(), layers.LayerTypeDNS, gopacket.Default)
	records := FromPacket(pkt)
	if records != nil {
		t.Fatalf("got %#v", records)
	}
}
