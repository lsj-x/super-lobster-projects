#!/bin/bash
# ModelMagic Deploy Console - 状态监控脚本
# 超级龙虾队监控助手 v1.0
#
# 功能:
# - 定时检查构建状态
# - 检测异常情况（超时/失败/卡死）
# - 自动恢复机制
# - 完成上报（飞书通知）

set -e

# ============ 配置区 ============
PROJECT_DIR="/root/.openclaw/workspace/repos/modelmagic-deploy-console"
LOG_FILE="${PROJECT_DIR}/auto-build.log"
STATE_FILE="${PROJECT_DIR}/.auto-build-state"
LOCK_FILE="${PROJECT_DIR}/.auto-build.lock"
MONITOR_LOG="${PROJECT_DIR}/monitor.log"

# 监控阈值（秒）
MAX_BUILD_TIME=3600      # 最大构建时间 1 小时
MAX_STALE_TIME=7200      # 最大陈旧时间 2 小时
CHECK_INTERVAL=300       # 检查间隔 5 分钟

# 通知配置
NOTIFY_CHANNEL="feishu"
NOTIFY_USER="大爪"

# ============ 工具函数 ============

log() {
  local level="$1"
  local msg="[$(date '+%Y-%m-%d %H:%M:%S')] [$level] $2"
  echo -e "$msg"
  echo -e "$msg" >> "$MONITOR_LOG"
}

info()    { log "INFO" "$1"; }
warn()    { log "WARN" "$1"; }
error()   { log "ERROR" "$1"; }

# 获取当前状态
get_build_state() {
  if [ ! -f "$STATE_FILE" ]; then
    echo "none|none|0"
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

# 计算时间差（秒）
time_diff() {
  local now=$(date +%s)
  local past="$1"
  echo $((now - past))
}

# 格式化时间差
format_time_diff() {
  local diff=$1
  local hours=$((diff / 3600))
  local minutes=$(((diff % 3600) / 60))
  local seconds=$((diff % 60))
  
  if [ $hours -gt 0 ]; then
    echo "${hours}小时${minutes}分钟"
  elif [ $minutes -gt 0 ]; then
    echo "${minutes}分钟${seconds}秒"
  else
    echo "${seconds}秒"
  fi
}

# 发送通知（飞书）
send_notification() {
  local title="$1"
  local content="$2"
  local level="${3:-info}"  # info, success, warning, error
  
  info "📢 发送通知：$title"
  
  # 记录到通知日志
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] [$level] $title: $content" >> "${LOG_FILE}.notify"
  
  # 飞书 webhook 通知（简化版，实际应调用飞书 API）
  # 这里使用 OpenClaw 内部通知机制
  if [ -n "$OPENCLAW_NOTIFY_URL" ]; then
    curl -s -X POST "$OPENCLAW_NOTIFY_URL" \
      -H "Content-Type: application/json" \
      -d "{
        \"title\": \"$title\",
        \"content\": \"$content\",
        \"level\": \"$level\",
        \"user\": \"$NOTIFY_USER\"
      }"
  fi
}

# ============ 监控检查 ============

# 检查 1: 构建是否卡死
check_build_stuck() {
  local state=$(get_build_state)
  local status=$(parse_state "$state" "status")
  local timestamp=$(parse_state "$state" "time")
  local now=$(date +%s)
  
  # 只检查运行中的任务
  if [ "$status" != "running" ]; then
    return 0
  fi
  
  # 检查是否超时
  local diff=$(time_diff "$timestamp")
  if [ $diff -gt $MAX_BUILD_TIME ]; then
    local time_str=$(format_time_diff $diff)
    error "❌ 构建任务已运行 ${time_str}，可能已卡死"
    
    send_notification "⚠️ 构建任务卡死警告" \
      "构建任务已运行 ${time_str} 未完成，可能已卡死。\n步骤：$(parse_state "$state" "step")\n建议：检查日志或手动干预" \
      "warning"
    
    return 1
  fi
  
  info "✓ 构建任务正常（已运行 $(format_time_diff $diff)）"
  return 0
}

# 检查 2: 状态文件是否过期
check_state_stale() {
  if [ ! -f "$STATE_FILE" ]; then
    return 0  # 无状态文件，正常
  fi
  
  local state=$(get_build_state)
  local timestamp=$(parse_state "$state" "time")
  local diff=$(time_diff "$timestamp")
  
  if [ $diff -gt $MAX_STALE_TIME ]; then
    local time_str=$(format_time_diff $diff)
    warn "⚠️ 状态文件已过陈旧 ${time_str}，可能需要清理"
    
    send_notification "📝 状态文件陈旧警告" \
      "状态文件已 ${time_str} 未更新，可能是上次构建异常退出。\n建议：检查日志或清理状态文件" \
      "warning"
    
    return 1
  fi
  
  return 0
}

# 检查 3: 锁文件是否残留
check_lock_stale() {
  if [ ! -f "$LOCK_FILE" ]; then
    return 0  # 无锁文件，正常
  fi
  
  local lock_pid=$(cat "$LOCK_FILE" 2>/dev/null)
  if [ -z "$lock_pid" ]; then
    return 0
  fi
  
  # 检查进程是否存在
  if ! kill -0 "$lock_pid" 2>/dev/null; then
    warn "发现残留锁文件（PID: $lock_pid 不存在），清理中..."
    rm -f "$LOCK_FILE"
    
    send_notification "🔓 清理残留锁文件" \
      "发现构建锁文件残留（原 PID: $lock_pid），已自动清理。" \
      "info"
  fi
  
  return 0
}

# 检查 4: 磁盘空间
check_disk_space() {
  local usage=$(df -h "$PROJECT_DIR" | tail -1 | awk '{print $5}' | sed 's/%//')
  
  if [ "$usage" -gt 90 ]; then
    error "❌ 磁盘空间不足 ${usage}%"
    
    send_notification "💾 磁盘空间警告" \
      "项目磁盘使用率已达 ${usage}%，可能导致构建失败。\n建议：清理旧构建产物或扩容" \
      "error"
    
    return 1
  elif [ "$usage" -gt 80 ]; then
    warn "⚠️ 磁盘使用率 ${usage}%"
  else
    info "✓ 磁盘空间正常 (${usage}%)"
  fi
  
  return 0
}

# 检查 5: 最近构建结果
check_last_build() {
  if [ ! -f "$LOG_FILE" ]; then
    return 0
  fi
  
  # 检查最近一次构建是否成功
  local last_result=$(grep "构建完成\|构建失败\|中断" "$LOG_FILE" | tail -1)
  
  if echo "$last_result" | grep -q "失败\|中断"; then
    local time_ago=$(echo "$last_result" | grep -oE '\[.*\]' | head -1)
    warn "最近一次构建失败：$last_result"
    
    send_notification "❌ 构建失败通知" \
      "最近一次构建失败：$last_result\n请检查日志排查原因" \
      "error"
    
    return 1
  elif echo "$last_result" | grep -q "完成"; then
    info "✓ 最近一次构建成功"
  fi
  
  return 0
}

# ============ 自动恢复 ============

# 恢复卡死的构建
recover_stuck_build() {
  local state=$(get_build_state)
  local step=$(parse_state "$state" "step")
  
  warn "尝试恢复卡死的构建任务（步骤：$step）..."
  
  # 清理锁文件
  rm -f "$LOCK_FILE"
  
  # 重置状态为失败，允许重新构建
  echo "${step}|failed|$(date +%s)" > "$STATE_FILE"
  
  send_notification "🔄 自动恢复构建任务" \
    "已检测到构建卡死，自动重置状态并清理锁文件。\n步骤：$step\n下次构建将自动重试" \
    "success"
}

# ============ 主流程 ============

main() {
  info "🔍 开始状态监控检查..."
  
  local issues=0
  
  # 执行所有检查
  check_build_stuck || ((issues++))
  check_state_stale || ((issues++))
  check_lock_stale || ((issues++))
  check_disk_space || ((issues++))
  check_last_build || ((issues++))
  
  # 汇总报告
  if [ $issues -gt 0 ]; then
    warn "监控发现 $issues 个问题，已发送通知"
    
    # 如果有卡死问题，尝试恢复
    if [ $issues -gt 2 ]; then
      recover_stuck_build
    fi
  else
    info "✅ 所有检查通过，系统运行正常"
  fi
  
  return 0
}

# 执行监控
main "$@"
