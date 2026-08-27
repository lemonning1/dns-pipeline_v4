package parse

import (
	"shared/model"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

func FromPacket(pkt gopacket.Packet) []*model.DNSQuery {
	dnslayer := pkt.Layer(layers.LayerTypeDNS)
	if dnslayer == nil {
		return nil
	}
	dns, ok := dnslayer.(*layers.DNS)
	if !ok {
		return nil
	}
	var records []*model.DNSQuery
	for i := range dns.Questions {
		record := ParseDNSQuestion(dns, &dns.Questions[i])
		if record != nil {
			records = append(records, record)
		}
	}
	return records
}

func ParseDNSQuestion(dns *layers.DNS, q *layers.DNSQuestion) *model.DNSQuery {
	domain := strings.TrimSuffix(string(q.Name), ".")

	record := &model.DNSQuery{
		Domain:    domain,
		QType:     int(q.Type),
		CreatedAt: time.Now(),
	}

	if !dns.QR {
		record.QR = 0
		return record
	}
	record.QR = 1
	record.RCode = int(dns.ResponseCode)

	var cnames, ips []string
	var ttl uint32

	for _, a := range dns.Answers {
		switch a.Type {
		case layers.DNSTypeCNAME:
			cnames = append(cnames, strings.TrimSuffix(string(a.CNAME), "."))
		case layers.DNSTypeA, layers.DNSTypeAAAA:
			if a.IP != nil {
				ips = append(ips, a.IP.String())
				if ttl == 0 {
					ttl = a.TTL
				}
			}
		}
	}
	record.Cnamechain = strings.Join(cnames, ",")
	record.ResponseIPs = strings.Join(ips, ",")
	record.TTL = int(ttl)

	return record

}
