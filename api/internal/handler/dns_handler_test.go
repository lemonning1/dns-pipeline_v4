package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"shared/model"

	"github.com/gin-gonic/gin"
)

type stubService struct {
	result *model.PageResult
	err    error
}

func (s stubService) GetDNSRecords(string, *int, model.PageParams) (*model.PageResult, error) {
	return s.result, s.err
}

func TestGetDNSRecords_200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &DNSHandler{service: stubService{
		result: &model.PageResult{Items: []model.DNSQuery{}, Page: 1, PageSize: 20, Total: 0},
	}}

	r := gin.New()
	r.GET("/api/dns", h.GetDNSRecords)

	req := httptest.NewRequest(http.MethodGet, "/api/dns?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var body model.PageResult
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Page != 1 {
		t.Fatalf("%+v", body)
	}
}

func TestGetDNSRecords_400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &DNSHandler{service: stubService{}}
	r := gin.New()
	r.GET("/api/dns", h.GetDNSRecords)

	req := httptest.NewRequest(http.MethodGet, "/api/dns?page=0&page_size=20", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestGetDNSRecords_500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &DNSHandler{service: stubService{err: errors.New("db down")}}
	r := gin.New()
	r.GET("/api/dns", h.GetDNSRecords)

	req := httptest.NewRequest(http.MethodGet, "/api/dns?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestGetDNSRecords_defaultPage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &DNSHandler{service: stubService{
		result: &model.PageResult{Items: []model.DNSQuery{}, Page: 1, PageSize: 20},
	}}
	r := gin.New()
	r.GET("/api/dns", h.GetDNSRecords)

	req := httptest.NewRequest(http.MethodGet, "/api/dns", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
