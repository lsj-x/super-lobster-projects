#!/bin/bash
# ModelMagic Deploy Console - 完成上报脚本
# 超级龙虾队汇报助手 v1.0
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

info()    { log "INFO" "$1"; }
warn()    { log "WARN" "$1"; }
error()   { log "ERROR" "$1"; }
success() { log "SUCCESS" "$1"; }

# 获取构建状态
get_build_state() {
  if [ ! -f "$STATE_FILE" ]; then
    echo "unknown|unknown|0"
    return
  fi
  cat "$STATE_FILE"
}

# 解析状态字段
parse_state() {
  local state="$1"
  local field="$2"
  
  case $field in
    step)   echo "$state" | cut -d'|' -f1 ;;
    status) echo "$state" | cut -d'|' -f2 ;;
    time)   echo "$state" | cut -d'|' -f3 ;;
  esac
}

# 统计构建次数
update_stats() {
  local result="$1"  # success or failed
  local duration="$2"
  
  # 初始化统计文件
  if [ ! -f "$STATS_FILE" ]; then
    echo '{"total":0,"success":0,"failed":0,"last_build":null}' > "$STATS_FILE"
  fi
  
  # 更新统计（简化版，实际应该用 jq）
  local total=$(grep -o '"total":[0-9]*' "$STATS_FILE" | cut -d: -f2)
  local success_count=$(grep -o '"success":[0-9]*' "$STATS_FILE" | cut -d: -f2)
  local failed_count=$(grep -o '"failed":[0-9]*' "$STATS_FILE" | cut -d: -f2)
  
  total=$((total + 1))
  
  if [ "$result" = "success" ]; then
    success_count=$((success_count + 1))
  else
    failed_count=$((failed_count + 1))
  fi
  
  local last_build=$(date -Iseconds)
  
  cat > "$STATS_FILE" << EOF
{
  "total": $total,
  "success": $success_count,
  "failed": $failed_count,
  "last_build": "$last_build",
  "last_duration": "$duration"
}
EOF
  
  info "📊 统计信息已更新：总构建 $total 次，成功 $success_count 次，失败 $failed_count 次"
}

# 生成构建报告
generate_report() {
  local status="$1"
  local duration="$2"
  local step="$3"
  
  local report=""
  report+="🦞 **超级龙虾队构建报告**\n\n"
  report+="**项目**: ModelMagic Deploy Console\n"
  report+="**时间**: $(date '+%Y-%m-%d %H:%M:%S')\n"
  report+="**状态**: "
  
  if [ "$status" = "success" ]; then
    report+="✅ 成功\n"
  else
    report+="❌ 失败 (步骤：$step)\n"
  fi
  
  report+="**耗时**: ${duration}\n\n"
  
  # 添加构建摘要
  report+="📋 **构建摘要**:\n"
  
  if [ -f "$LOG_FILE" ]; then
    # 提取最近一次构建的关键步骤
    local backend_time=$(grep "后端编译" "$LOG_FILE" | tail -1 | grep -oE '\[.*\]' | tail -1)
    local frontend_time=$(grep "前端编译" "$LOG_FILE" | tail -1 | grep -oE '\[.*\]' | tail -1)
    local test_time=$(grep "测试" "$LOG_FILE" | tail -1 | grep -oE '\[.*\]' | tail -1)
    
    report+="- 后端编译：${backend_time:-未知}\n"
    report+="- 前端编译：${frontend_time:-未知}\n"
    report+="- 集成测试：${test_time:-未知}\n"
  fi
  
  report+="\n📊 **统计信息**:\n"
  if [ -f "$STATS_FILE" ]; then
    local total=$(grep -o '"total":[0-9]*' "$STATS_FILE" | cut -d: -f2)
    local success_count=$(grep -o '"success":[0-9]*' "$STATS_FILE" | cut -d: -f2)
    local success_rate=0
    if [ -n "$total" ] && [ "$total" -gt 0 ] 2>/dev/null; then
      success_rate=$((success_count * 100 / total))
    fi
    report+="- 总构建次数：$total\n"
    report+="- 成功次数：$success_count\n"
    report+="- 失败次数：$(grep -o '"failed":[0-9]*' "$STATS_FILE" | cut -d: -f2)\n"
    report+="- 成功率：${success_rate}%\n"
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
      -d "{
        \"title\": \"$title\",
        \"content\": \"$report\",
        \"level\": \"$( [ \"$status\" = 'success' ] && echo 'success' || echo 'error')\",
        \"user\": \"$NOTIFY_USER\"
      }"
  fi
  
  # 打印报告
  echo -e "\n$report"
}

# ============ 主流程 ============

main() {
  local status="${1:-unknown}"
  local step="${2:-unknown}"
  
  # 计算耗时
  local start_time=$(parse_state "$(get_build_state)" "time")
  local now=$(date +%s)
  local duration=$((now - start_time))
  local duration_str=""
  
  if [ $duration -gt 3600 ]; then
    duration_str="$((duration / 3600))小时$((duration % 3600 / 60))分钟"
  elif [ $duration -gt 60 ]; then
    duration_str="$((duration / 60))分钟$((duration % 60))秒"
  else
    duration_str="${duration}秒"
  fi
  
  info "🎉 构建流程结束：$status (耗时：$duration_str)"
  
  # 生成并发送报告
  send_complete_notification "$status" "$duration_str" "$step"
  
  # 更新统计
  update_stats "$status" "$duration_str"
  
  return 0
}

# 执行上报
main "$@"
