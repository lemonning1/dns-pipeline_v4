# dns_pipeline_v4

## 项目简介
DNS流量抓取 → kafka → clickhouse → API/网页查询

## 技术栈
- GO(producer / consumer / api)
- Kafka(KRaft,3brokers)
- Clickhouse(2分片*2副本+Keeper)
- Docker Compose、Nginx、静态Web

## 架构
基于 Go 的 DNS 流量管道：pcap 抓包并经 BPF 过滤，gopacket 解析 Question/Answer 后由 Producer 序列化写入 Kafka；Consumer 按消费组拉取并写入 ClickHouse 集群（ReplicatedMergeTree + Distributed）。Gin 暴露 GET /api/dns 分页查询，Nginx 提供静态页面并将 /api 转发至 API，前端通过 fetch 渲染列表与分页。

## 快速启动
1. 环境变量:`.env`里自行设置CLICKHOUSE_PASSWORD;运行时 DB_PASSWORD
2. 起 Kafka:`docker compose -f compose.kafka.yaml up -d`
3. 起 Clickhouse:`docker compose -p ch-cluster -f compose.clickhouse.yaml up -d`
4. 起producer / comsumer / api(根据实际使用的脚本或docker命令)

## 配置说明
- `confg/pipeline.yaml`、`config/api.yaml`
- Kafka 多地址、Clickhouse DSN / 多 host

## 集群说明
- Kafka:3 分区、RF=3、消费组并发
- Clickhouse:分片/副本、Distributed、Keeper、secret

## 常见问题
- Keeper Connection refused → listen_host + restart
- Distributed 认证失败 → cluster secret
- 表在 default 不在dnsdb → 库名对齐

## 后续可做
- consumer 批量写入
- parse 函数优化 record 处理DNS多问题字段方式 
- Grafana 统计图
- 生产者 worker 池