# dns_pipeline_v4

## 项目简介
DNS流量抓取 → kafka → clickhouse → API/网页查询

## 技术栈
- GO(producer / consumer / api)
- Kafka(KRaft,3brokers)
- Clickhouse(2分片*2副本+Keeper)
- Docker Compose、Nginx、静态Web

## 架构
capture → parse → Kafka → consumer → ClickHouse → Gin API → Nginx → web

