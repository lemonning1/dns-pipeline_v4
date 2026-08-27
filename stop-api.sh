#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")" && pwd)"

if [ -f "$ROOT/api/logs/api.pid" ]; then
  kill "$(cat "$ROOT/api/logs/api.pid")" 2>/dev/null || true
  rm -f "$ROOT/api/logs/api.pid"
  echo "已停止 API"
else
  echo "未找到 API PID 文件"
fi
