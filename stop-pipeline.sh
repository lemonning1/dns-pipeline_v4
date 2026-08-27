#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"

if [ -f "$ROOT/pipeline/logs/producer.pid" ]; then
  sudo kill "$(cat "$ROOT/pipeline/logs/producer.pid")" 2>/dev/null || true
  rm -f "$ROOT/pipeline/logs/producer.pid"
  echo "已停止生产者"
else
  echo "未找到生产者 PID 文件"
fi

if [ -f "$ROOT/pipeline/logs/consumer.pid" ]; then
  kill "$(cat "$ROOT/pipeline/logs/consumer.pid")" 2>/dev/null || true
  rm -f "$ROOT/pipeline/logs/consumer.pid"
  echo "已停止消费者"
else
  echo "未找到消费者 PID 文件"
fi
