package apiserver

import (
	"api/internal/handler"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"shared/model"

	"github.com/gin-gonic/gin"
)

type stubService struct{}

func (stubService) GetDNSRecords(string, *int, model.PageParams) (*model.PageResult, error) {
	return &model.PageResult{Items: []model.DNSQuery{}, Page: 1, PageSize: 20, Total: 0}, nil
}
func (stubService) GetTopDomains(string) ([]model.TopResult, error) {
	items := []model.TopResult{}
	return items, nil
}
func TestSetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewDNSHandler(stubService{})
	r := setupRouter(h)

	req := httptest.NewRequest(http.MethodOptions, "/api/dns", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS status=%d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/dns?page=1&page_size=20", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestRun_listenError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	err := Run(context.Background(), handler.NewDNSHandler(stubService{}), "127.0.0.1:notaport")
	if err == nil {
		t.Fatal("期望监听失败")
	}
}

func TestRun_shutdown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(80 * time.Millisecond)
		cancel()
	}()
	if err := Run(ctx, handler.NewDNSHandler(stubService{}), "127.0.0.1:0"); err != nil {
		t.Fatal(err)
	}
}
