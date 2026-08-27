package apiserver

import (
	"api/internal/handler"
	"context"
	"net/http"
	"time"

	"shared/logger"

	"github.com/gin-gonic/gin"
)

func setupRouter(h *handler.DNSHandler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
	r.GET("/api/dns", h.GetDNSRecords)
	return r
}

func Run(ctx context.Context, handler *handler.DNSHandler, addr string) error {
	r := setupRouter(handler)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}
	errCh := make(chan error, 1)
	go func() {
		logger.Info("API 已启动，监听 %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	select {
	case <-ctx.Done():
		logger.Info("API 收到退出信号,正在关闭...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
