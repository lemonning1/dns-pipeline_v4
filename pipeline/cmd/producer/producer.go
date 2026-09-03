package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"pipeline/internal/capture"
	"pipeline/internal/kafka"
	"pipeline/internal/producerserver"
	"shared/config"
	"shared/logger"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "../config/pipeline.yaml"
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal("加载配置文件失败:", err)
		return
	}
	if err := logger.Init(cfg.Pipeline.ProducerLog); err != nil {
		log.Printf("初始化日志失败: %v", err)
	}

	cap, err := capture.NewPacketCapture(&cfg.Pipeline.Collector)
	if err != nil {
		log.Fatal(err)
	}
	defer cap.Close()

	prod, err := kafka.NewProducer(cfg.Pipeline.Kafka)
	if err != nil {
		log.Fatal(err)
	}
	logger.Infof("生产者已创建,开始生产...")
	defer prod.Close()
	//更干净的优雅退出
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		cap.Close()
	}()

	if err := producerserver.Run(ctx, cap, prod); err != nil {
		log.Fatal(err)
	}
}
