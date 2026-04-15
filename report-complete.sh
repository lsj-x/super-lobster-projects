#!/bin/bash
# ModelMagic Deploy Console - 完成上报脚本
# 超级龙虾队汇报助手 v1.1 (修复统计信息空值问题)
#
# 功能:
# - 构建完成后触发
# - 生成详细报告（成功/失败/警告）
# - 发送飞书通知
# - 更新统计信息

set -e

# ============ 配置区 ============
PROJECT_DIR="/root/.openclaw/workspace/repos/modelmagic-deploy-console"
LOG_FILE="${PROJECT_DIR}/auto-build.log"
STATE_FILE="${PROJECT_DIR}/.auto-build-state"
REPORT_LOG="${PROJECT_DIR}/report.log"
STATS_FILE="${PROJECT_DIR}/.build-stats.json"

# 通知配置
NOTIFY_USER="大爪"
NOTIFY_CHANNEL="feishu"

# ============ 工具函数 ============

log() {
 local level="$1"
 local msg="[$(date '+%Y-%m-%d %H:%M:%S')] [$level] $2"
 echo -e "$msg"
 echo -e "$msg" >> "$REPORT_LOG"
}

info() { log "INFO" "$1"; }
warn() { log "WARN" "$1"; }
error() { log "ERROR" "$1"; }
success() { log "SUCCESS" "$1"; }

# 获取构建状态
get_build_state() {
 if [ -f "$STATE_FILE" ]; then
  cat "$STATE_FILE"
 else
  echo "unknown"
 fi
}

# 生成构建报告
generate_report() {
 local status="$1"
 local duration="$2"
 local step="$3"

 local report=""
 report+="🦞 **超级龙虾队构建报告**\n"
 report+="**项目**: ModelMagic Deploy Console\n"
 report+="**时间**: $(date '+%Y-%m-%d %H:%M:%S')\n"
 report+="**状态**: "
 if [ "$status" = "success" ]; then
  report+="✅ 成功\n"
 else
  report+="❌ 失败\n"
 fi
 report+="**耗时**: ${duration}秒\n"

 report+="\n📋 **构建摘要**:\n"
 if [ -n "$step" ]; then
  # 从日志中提取时间戳
  local backend_time=$(grep "✅ 后端编译成功" "$LOG_FILE" 2>/dev/null | tail -1 | sed 's/\[\([^]]*\)\].*/\1/' || echo "未知")
  local frontend_time=$(grep "✅ 前端编译成功" "$LOG_FILE" 2>/dev/null | tail -1 | sed 's/\[\([^]]*\)\].*/\1/' || echo "未知")
  local test_time=$(grep "✅ 测试全部通过" "$LOG_FILE" 2>/dev/null | tail -1 | sed 's/\[\([^]]*\)\].*/\1/' || echo "未知")
  report+=" - 后端编译：${backend_time:-未知}\n"
  report+=" - 前端编译：${frontend_time:-未知}\n"
  report+=" - 集成测试：${test_time:-未知}\n"
 fi

 report+="\n📊 **统计信息**:\n"
 if [ -f "$STATS_FILE" ]; then
  local total=$(grep -o '"total":[0-9]*' "$STATS_FILE" | cut -d: -f2)
  local success_count=$(grep -o '"success":[0-9]*' "$STATS_FILE" | cut -d: -f2)
  local failed_count=$(grep -o '"failed":[0-9]*' "$STATS_FILE" | cut -d: -f2)
  local success_rate=0

  # 处理空值
  total=${total:-0}
  success_count=${success_count:-0}
  failed_count=${failed_count:-0}

  if [ "$total" -gt 0 ] 2>/dev/null; then
   success_rate=$((success_count * 100 / total))
  fi

  report+=" - 总构建次数：$total\n"
  report+=" - 成功次数：$success_count\n"
  report+=" - 失败次数：$failed_count\n"
  report+=" - 成功率：${success_rate}%\n"
 fi

 echo -e "$report"
}

# 发送完成通知
send_complete_notification() {
 local status="$1"
 local duration="$2"
 local step="$3"

 local report=$(generate_report "$status" "$duration" "$step")
 info "📢 发送完成通知：$status"

 # 发送到飞书
 local title=""
 if [ "$status" = "success" ]; then
  title="✅ 构建成功"
 else
  title="❌ 构建失败"
 fi

 # 调用通知 API
 echo -e "$report" >> "${LOG_FILE}.notify"

 # 飞书 webhook（简化实现）
 if [ -n "$OPENCLAW_NOTIFY_URL" ]; then
  curl -s -X POST "$OPENCLAW_NOTIFY_URL" \
   -H "Content-Type: application/json" \
   -d "{ \"title\": \"$title\", \"content\": \"$(echo -e "$report" | sed 's/"/\\"/g')\" }" || true
 fi

 # 如果配置了飞书机器人，也可以通过环境变量发送
 if [ -n "$FEISHU_WEBHOOK_URL" ]; then
  curl -s -X POST "$FEISHU_WEBHOOK_URL" \
   -H "Content-Type: application/json" \
   -d "{ \"msg_type\": \"text\", \"content\": { \"text\": \"$(echo -e "$report" | sed 's/"/\\"/g')\" } }" || true
 fi
}

# 更新统计信息
update_stats() {
 local status="$1"

 if [ ! -f "$STATS_FILE" ]; then
  echo '{"total":0,"success":0,"failed":0,"last_update":""}' > "$STATS_FILE"
 fi

 local total=$(grep -o '"total":[0-9]*' "$STATS_FILE" | cut -d: -f2)
 local success_count=$(grep -o '"success":[0-9]*' "$STATS_FILE" | cut -d: -f2)
 local failed_count=$(grep -o '"failed":[0-9]*' "$STATS_FILE" | cut -d: -f2)

 total=${total:-0}
 success_count=${success_count:-0}
 failed_count=${failed_count:-0}

 total=$((total + 1))
 if [ "$status" = "success" ]; then
  success_count=$((success_count + 1))
 else
  failed_count=$((failed_count + 1))
 fi

 cat > "$STATS_FILE" << EOF
{"total":$total,"success":$success_count,"failed":$failed_count,"last_update":"$(date -Iseconds)"}
EOF

 info "📊 统计信息已更新：总构建 $total 次，成功 $success_count 次，失败 $failed_count 次"
}

# ============ 主流程 ============

main() {
 local status="${1:-unknown}"
 local duration="${2:-0}"
 local step="${3:-}"

 info "📊 生成构建报告..."

 # 更新统计
 update_stats "$status"

 # 发送通知
 send_complete_notification "$status" "$duration" "$step"

 info "🎉 构建流程结束：$status (耗时：${duration}秒)"
}

# 执行主流程
main "$@"
