package service

import (
	"errors"
	"shared/model"
	"testing"
)

type stubRepo struct {
	items []model.DNSQuery
	total int
	err   error
	top   []model.TopResult
}

func (s stubRepo) FindByDomain(string, *int, model.PageParams) ([]model.DNSQuery, int, error) {
	return s.items, s.total, s.err
}

func (s stubRepo) TopDomains(string) ([]model.TopResult, error) {
	return s.top, s.err
}

func TestGetDNSRecords_ok(t *testing.T) {
	svc := &DNSService{repo: stubRepo{
		items: []model.DNSQuery{{Domain: "a.com"}},
		total: 1,
	}}
	page := model.PageParams{Page: 1, PageSize: 20}

	got, err := svc.GetDNSRecords("a.com", nil, page)
	if err != nil {
		t.Fatal(err)
	}
	if got.Total != 1 || len(got.Items) != 1 || got.Page != 1 {
		t.Fatalf("%+v", got)
	}
}

func TestGetDNSRecords_err(t *testing.T) {
	svc := &DNSService{repo: stubRepo{err: errors.New("db down")}}
	_, err := svc.GetDNSRecords("", nil, model.PageParams{Page: 1, PageSize: 20})
	if err == nil {
		t.Fatal("期望错误")
	}
}

func TestNewDNSService(t *testing.T) {
	if NewDNSService(nil) == nil {
		t.Fatal("nil service")
	}
}
