#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"
cd "$ROOT/api"

if [ -z "${DB_PASSWORD:-}" ]; then
  echo "请先执行: export DB_PASSWORD='你的数据库密码'"
  exit 1
fi

mkdir -p logs/

echo "启动 API..."
nohup go run ./cmd/api/  &
echo $! > logs/api.pid
echo "API PID=$(cat logs/api.pid)  日志: api/logs/api.log"
echo "API 已在后台运行"
