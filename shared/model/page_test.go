package model

import "testing"

func TestPageParamsGetOffset(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		want     int
	}{
		{name: "第一页", page: 1, pageSize: 20, want: 0},
		{name: "第二页", page: 2, pageSize: 20, want: 20},
		{name: "第三页每页10条", page: 3, pageSize: 10, want: 20},
		{name: "页码非法当作1", page: 0, pageSize: 20, want: 0},
		{name: "每页条数非法当作20", page: 3, pageSize: 0, want: 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PageParams{Page: tt.page, PageSize: tt.pageSize}
			got := p.GetOffset()
			if got != tt.want {
				t.Fatalf("GetOffset() = %d, want %d", got, tt.want)
			}
		})
	}
}
