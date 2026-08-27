package service

import (
	"api/internal/repository"
	"shared/model"
)

type DNSFinder interface {
	FindByDomain(domain string, qr *int, page model.PageParams) ([]model.DNSQuery, int, error)
}

type DNSService struct {
	repo DNSFinder
}

func NewDNSService(repo *repository.DNSRepo) *DNSService {
	return &DNSService{repo: repo}
}

func (s *DNSService) GetDNSRecords(domain string, qr *int, page model.PageParams) (*model.PageResult, error) {
	items, total, err := s.repo.FindByDomain(domain, qr, page)
	if err != nil {
		return nil, err
	}
	return &model.PageResult{
		Items:    items,
		Page:     page.Page,
		PageSize: page.PageSize,
		Total:    total,
	}, nil
}
