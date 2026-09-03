package main

import (
	"api/internal/apiserver"
	"api/internal/handler"
	"api/internal/repository"
	"api/internal/service"
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shared/config"
	"shared/logger"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "../config/api.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := logger.Init(cfg.API.Log); err != nil {
		log.Printf("初始化日志失败 %v", err)
	}

	db, err := sql.Open("clickhouse", cfg.API.Database.DSN())
	if err != nil {
		log.Fatal("DSN格式错误: ", err)
	}
	defer db.Close()
	for {
		if err := db.Ping(); err != nil {
			logger.Warnf("无法连接到ClickHouse服务器:%v", err)
			time.Sleep(5 * time.Second)
		} else {
			break
		}

	}
	repo := repository.NewDNSRepo(db)
	svc := service.NewDNSService(repo)
	h := handler.NewDNSHandler(svc)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	host := cfg.API.API.Host
	port := cfg.API.API.Port
	if host == "" {
		host = "0.0.0.0"
	}
	if port == "" {
		port = "8080"
	}
	addr := host + ":" + port
	if err := apiserver.Run(ctx, h, addr); err != nil {
		log.Fatal(err)
	}
}
