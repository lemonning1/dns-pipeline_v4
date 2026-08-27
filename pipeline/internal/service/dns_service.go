package service

import (
	"pipeline/internal/repository"
	"shared/model"
)

type DNSWriter interface {
	EnsureTable() error
	InsertDNSQuery(query *model.DNSQuery) error
}

type DNSService struct {
	repo DNSWriter
}

func NewDNSService(repo *repository.DNSRepo) *DNSService {
	return &DNSService{repo: repo}
}

func (s *DNSService) Ensure() error {
	return s.repo.EnsureTable()
}

func (s *DNSService) Insert(query *model.DNSQuery) error {
	return s.repo.InsertDNSQuery(query)
}
