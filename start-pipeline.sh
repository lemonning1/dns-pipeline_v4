#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT/pipeline"

if [ -z "${DB_PASSWORD:-}" ]; then
  echo "请先执行: export DB_PASSWORD='你的数据库密码'"
  exit 1
fi

mkdir -p logs

echo "启动消费者..."
nohup go run ./cmd/consumer/ &
echo $! > logs/consumer.pid
echo "消费者 PID=$(cat logs/consumer.pid)  日志: pipeline/logs/consumer.log"

echo "启动生产者..."
sudo -v
sudo env PATH="$PATH" DB_PASSWORD="$DB_PASSWORD" \
  nohup go run ./cmd/producer/ &
echo $! > logs/producer.pid
echo "生产者 PID=$(cat logs/producer.pid)  日志: pipeline/logs/producer.log"

echo "pipeline 已在后台运行"
