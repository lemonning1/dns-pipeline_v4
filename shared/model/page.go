package model

type PageParams struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}
type PageResult struct {
	Items    []DNSQuery `json:"items"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
	Total    int        `json:"total"`
}

func (p *PageParams) GetOffset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}

	return (p.Page - 1) * p.PageSize
}
