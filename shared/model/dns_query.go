package model

import "time"

type DNSQuery struct {
	ID          int       `json:"id"`
	Domain      string    `json:"domain"`
	ClientIP    string    `json:"clientip"`
	QType       int       `json:"qtype"`
	QR          int       `json:"qr"`
	RCode       int       `json:"rcode"`
	Cnamechain  string    `json:"cnamechain"`
	ResponseIPs string    `json:"responseips"`
	TTL         int       `json:"ttl"`
	CreatedAt   time.Time `json:"created_at"`
}
