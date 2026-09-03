package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"pipeline/internal/consumerserver"
	"pipeline/internal/kafka"
	"pipeline/internal/repository"
	"pipeline/internal/service"
	"shared/config"
	"shared/logger"
	"syscall"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "../config/pipeline.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		logger.Fatalf("加载配置文件失败:%v", err)
		return
	}
	if err := logger.Init(cfg.Pipeline.ConsumerLog); err != nil {
		log.Printf("初始化日志失败: %v", err)
	}

	db, err := sql.Open("clickhouse", cfg.Pipeline.Database.DSN())
	if err != nil {
		logger.Fatalf("DSN格式错误:%v", err)
	}
	defer db.Close()
	for {
		if err := db.Ping(); err != nil {
			logger.Warnf("无法连接到ClickHouse服务器:%v", err)
		} else {
			break
		}

	}
	repo := repository.NewDNSRepo(db)
	err = repo.EnsureTable()
	if err != nil {
		log.Fatal("创建表失败:", err)
	}

	cons, err := kafka.NewConsumer(&cfg.Pipeline.Kafka)
	if err != nil {
		log.Fatal(err)
	}
	logger.Infof("消费者已创建，开始消费...")
	defer cons.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	svc := service.NewDNSService(repo)
	if err := consumerserver.Run(ctx, cons, svc); err != nil {
		log.Fatal(err)
	}

}
