package service

import (
	"errors"
	"testing"

	"shared/model"
)

type stubWriter struct {
	ensureErr error
	insertErr error
}

func (s stubWriter) EnsureTable() error { return s.ensureErr }
func (s stubWriter) InsertDNSQuery(*model.DNSQuery) error {
	return s.insertErr
}

func TestEnsureAndInsert(t *testing.T) {
	svc := &DNSService{repo: stubWriter{}}
	if err := svc.Ensure(); err != nil {
		t.Fatal(err)
	}
	if err := svc.Insert(&model.DNSQuery{Domain: "a.com"}); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureAndInsert_err(t *testing.T) {
	svc := &DNSService{repo: stubWriter{ensureErr: errors.New("e"), insertErr: errors.New("i")}}
	if err := svc.Ensure(); err == nil {
		t.Fatal("期望 Ensure 错误")
	}
	if err := svc.Insert(&model.DNSQuery{}); err == nil {
		t.Fatal("期望 Insert 错误")
	}
}
