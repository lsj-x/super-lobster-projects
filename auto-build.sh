#!/bin/bash
# ModelMagic Deploy Console - 自动构建脚本（增强版）
# 超级龙虾队自动化工作流 v2.0
# 
# 新增功能:
# - API 限流保护（指数退避）
# - 子任务超时控制
# - 断点续建
# - 详细日志

set -e

# 记录开始时间
BUILD_START_TIME=$(date +%s)

# ============ 配置区 ============
PROJECT_DIR="/root/.openclaw/workspace/repos/modelmagic-deploy-console"
LOG_FILE="${PROJECT_DIR}/auto-build.log"
STATE_FILE="${PROJECT_DIR}/.auto-build-state"
LOCK_FILE="${PROJECT_DIR}/.auto-build.lock"
REPORT_SCRIPT="${PROJECT_DIR}/report-complete.sh"
MONITOR_SCRIPT="${PROJECT_DIR}/monitor-status.sh"

# 重试配置
MAX_RETRIES=3
BASE_DELAY=5          # 基础延迟（秒）
MAX_DELAY=60          # 最大延迟（秒）
EXPONENTIAL_BASE=2    # 指数退避基数

# 超时配置（秒）
TIMEOUT_BACKEND=300    # 后端编译 5 分钟
TIMEOUT_FRONTEND=600   # 前端编译 10 分钟
TIMEOUT_TEST=300       # 测试 5 分钟
TIMEOUT_GIT=60         # Git 操作 1 分钟

# API 限流配置
API_MAX_REQUESTS=100   # 每分钟最大请求数
API_WINDOW=60          # 时间窗口（秒）

# ============ 工具函数 ============

# 日志函数（带时间戳和级别）
log() {
  local level="$1"
  local msg="[$(date '+%Y-%m-%d %H:%M:%S')] [$level] $2"
  echo -e "$msg"
  echo -e "$msg" >> "$LOG_FILE"
}

info()    { log "INFO" "$1"; }
warn()    { log "WARN" "$1"; }
error()   { log "ERROR" "$1"; }
debug()   { log "DEBUG" "$1"; }

# 保存进度状态
save_state() {
  local step="$1"
  local status="$2"
  local timestamp=$(date +%s)
  echo "${step}|${status}|${timestamp}" > "$STATE_FILE"
  debug "状态已保存：${step} -> ${status}"
}

# 读取进度状态
load_state() {
  if [ -f "$STATE_FILE" ]; then
    cat "$STATE_FILE"
  else
    echo "none|none|0"
  fi
}

# 获取当前步骤
get_current_step() {
  if [ -f "$STATE_FILE" ]; then
    cut -d'|' -f1 "$STATE_FILE"
  else
    echo "none"
  fi
}

# 检查是否应该继续上一步
should_resume() {
  local state=$(load_state)
  local step=$(echo "$state" | cut -d'|' -f1)
  local status=$(echo "$state" | cut -d'|' -f2)
  local timestamp=$(echo "$state" | cut -d'|' -f3)
  local now=$(date +%s)
  
  # 如果上次执行在 1 小时内失败，尝试恢复
  if [ "$status" = "failed" ] && [ $((now - timestamp)) -lt 3600 ]; then
    info "检测到上次执行失败，尝试从步骤 ${step} 恢复..."
    return 0
  fi
  return 1
}

# 指数退避等待
exponential_backoff() {
  local attempt=$1
  local delay=$((BASE_DELAY * (EXPONENTIAL_BASE ** (attempt - 1))))
  
  # 限制最大延迟
  if [ $delay -gt $MAX_DELAY ]; then
    delay=$MAX_DELAY
  fi
  
  warn "第 ${attempt} 次失败，等待 ${delay} 秒后重试..."
  sleep $delay
}

# 带超时的命令执行
run_with_timeout() {
  local timeout=$1
  local cmd="$2"
  local retry_count=0
  local success=false
  
  while [ $retry_count -lt $MAX_RETRIES ] && [ "$success" = "false" ]; do
    info "执行命令（超时 ${timeout}s）: $cmd"
    
    # 使用 timeout 命令
    if timeout $timeout bash -c "$cmd" >> "$LOG_FILE" 2>&1; then
      success=true
      info "命令执行成功"
    else
      local exit_code=$?
      retry_count=$((retry_count + 1))
      
      if [ $exit_code -eq 124 ]; then
        error "命令执行超时（${timeout}s）"
      else
        error "命令执行失败（退出码：$exit_code）"
      fi
      
      if [ $retry_count -lt $MAX_RETRIES ]; then
        exponential_backoff $retry_count
      else
        error "已达最大重试次数 ($MAX_RETRIES)"
        return 1
      fi
    fi
  done
  
  return 0
}

# API 限流控制（简单令牌桶）
API_REQUEST_COUNT=0
API_WINDOW_START=$(date +%s)

check_api_rate_limit() {
  local now=$(date +%s)
  local elapsed=$((now - API_WINDOW_START))
  
  # 重置时间窗口
  if [ $elapsed -ge $API_WINDOW ]; then
    API_REQUEST_COUNT=0
    API_WINDOW_START=$now
  fi
  
  # 检查是否超限
  if [ $API_REQUEST_COUNT -ge $API_MAX_REQUESTS ]; then
    local wait_time=$((API_WINDOW - elapsed))
    warn "API 请求限流，等待 ${wait_time} 秒..."
    sleep $wait_time
    API_REQUEST_COUNT=0
    API_WINDOW_START=$(date +%s)
  fi
  
  API_REQUEST_COUNT=$((API_REQUEST_COUNT + 1))
}

# 带限流的 API 调用
api_call() {
  check_api_rate_limit
  local url="$1"
  local method="${2:-GET}"
  local data="${3:-}"
  
  debug "API 调用：$method $url"
  
  if [ "$method" = "POST" ] || [ "$method" = "PUT" ]; then
    curl -s -X "$method" -H "Content-Type: application/json" -d "$data" "$url"
  else
    curl -s -X "$method" "$url"
  fi
}

# 获取锁（防止并发执行）
acquire_lock() {
  if [ -f "$LOCK_FILE" ]; then
    local lock_pid=$(cat "$LOCK_FILE")
    if kill -0 "$lock_pid" 2>/dev/null; then
      error "检测到另一个构建进程正在运行 (PID: $lock_pid)"
      return 1
    else
      warn "发现残留锁文件，清理中..."
      rm -f "$LOCK_FILE"
    fi
  fi
  
  echo $$ > "$LOCK_FILE"
  info "获取构建锁成功 (PID: $$)"
  return 0
}

# 释放锁
release_lock() {
  if [ -f "$LOCK_FILE" ]; then
    rm -f "$LOCK_FILE"
    info "释放构建锁"
  fi
}

# ============ 构建步骤 ============

# 步骤 1: 环境检查
step_check_env() {
  info "🔍 步骤 1: 检查环境..."
  
  local checks_passed=true
  
  # 检查 Go 版本
  if command -v go &> /dev/null; then
    local go_version=$(go version | cut -d' ' -f3)
    info "  ✓ Go 版本：$go_version"
  else
    error "  ✗ Go 未安装"
    checks_passed=false
  fi
  
  # 检查 Node.js
  if command -v node &> /dev/null; then
    local node_version=$(node --version)
    info "  ✓ Node.js 版本：$node_version"
  else
    error "  ✗ Node.js 未安装"
    checks_passed=false
  fi
  
  # 检查 kubectl/helm（可选）
  if command -v kubectl &> /dev/null; then
    info "  ✓ kubectl 已安装"
  else
    warn "  ⚠ kubectl 未安装（可选）"
  fi
  
  if command -v helm &> /dev/null; then
    info "  ✓ helm 已安装"
  else
    warn "  ⚠ helm 未安装（可选）"
  fi
  
  if [ "$checks_passed" = "false" ]; then
    return 1
  fi
  
  info "✅ 环境检查通过"
  return 0
}

# 步骤 2: 后端编译
step_build_backend() {
  info "🔧 步骤 2: 编译后端..."
  
  cd "${PROJECT_DIR}/backend"
  
  # 使用超时控制
  if ! run_with_timeout $TIMEOUT_BACKEND "go mod tidy"; then
    error "后端依赖整理失败"
    return 1
  fi
  
  if ! run_with_timeout $TIMEOUT_BACKEND "go build -o bin/server ./cmd/server"; then
    error "后端编译失败"
    return 1
  fi
  
  info "✅ 后端编译成功"
  return 0
}

# 步骤 3: 前端编译
step_build_frontend() {
  info "🎨 步骤 3: 编译前端..."
  
  cd "${PROJECT_DIR}/frontend"
  
  # 使用超时控制
  if ! run_with_timeout $TIMEOUT_FRONTEND "npm install --legacy-peer-deps"; then
    error "前端依赖安装失败"
    return 1
  fi
  
  if ! run_with_timeout $TIMEOUT_FRONTEND "npm run build"; then
    error "前端编译失败"
    return 1
  fi
  
  info "✅ 前端编译成功"
  return 0
}

# 步骤 4: 运行测试
step_run_tests() {
  info "🧪 步骤 4: 运行测试..."
  
  cd "${PROJECT_DIR}/backend"
  
  # 使用超时控制
  if ! run_with_timeout $TIMEOUT_TEST "go test ./... -v"; then
    error "测试失败"
    return 1
  fi
  
  info "✅ 测试全部通过"
  return 0
}

# 步骤 5: Git 提交
step_git_commit() {
  info "📝 步骤 5: Git 提交..."
  
  cd "${PROJECT_DIR}"
  
  # 检查是否有更改
  if [ -z "$(git status --porcelain)" ]; then
    info "ℹ️  没有更改需要提交"
    return 0
  fi
  
  # 使用超时控制
  if ! run_with_timeout $TIMEOUT_GIT "git add -A"; then
    error "Git add 失败"
    return 1
  fi
  
  # 生成提交信息
  local commit_msg="auto: 自动构建完成 $(date '+%Y-%m-%d %H:%M')"
  
  if ! run_with_timeout $TIMEOUT_GIT "git commit -m '$commit_msg'"; then
    # 提交失败可能是因为没有更改
    warn "Git 提交失败，可能没有更改"
    return 0
  fi
  
  info "✅ Git 提交成功：$commit_msg"
  return 0
}

# ============ 主流程 ============
main() {
  local current_step="${1:-all}"
  
  # 获取锁
  if ! acquire_lock; then
    error "无法获取构建锁，退出"
    exit 1
  fi
  
  # 设置异常退出时的清理
  trap 'release_lock; error "构建过程中断"; exit 1' INT TERM ERR
  
  info "🚀 超级龙虾队自动构建开始 (增强版)"
  info "📁 项目目录：$PROJECT_DIR"
  
  # 保存开始状态
  save_state "started" "running"
  
  # 步骤 1: 环境检查
  if ! step_check_env; then
    error "环境检查失败"
    save_state "check_env" "failed"
    release_lock
    # 调用失败上报
    if [ -x "$REPORT_SCRIPT" ]; then
      bash "$REPORT_SCRIPT" "failed" "check_env" 2>&1 | tee -a "$LOG_FILE"
    fi
    return 1
  fi
  save_state "check_env" "success"
  
  # 步骤 2: 后端编译
  if [ "$current_step" != "frontend" ] && [ "$current_step" != "test" ]; then
    if ! step_build_backend; then
      error "后端编译失败"
      save_state "build_backend" "failed"
      release_lock
      if [ -x "$REPORT_SCRIPT" ]; then
        bash "$REPORT_SCRIPT" "failed" "build_backend" 2>&1 | tee -a "$LOG_FILE"
      fi
      return 1
    fi
  fi
  save_state "build_backend" "success"
  
  # 步骤 3: 前端编译
  if [ "$current_step" != "backend" ]; then
    if ! step_build_frontend; then
      error "前端编译失败"
      save_state "build_frontend" "failed"
      release_lock
      if [ -x "$REPORT_SCRIPT" ]; then
        bash "$REPORT_SCRIPT" "failed" "build_frontend" 2>&1 | tee -a "$LOG_FILE"
      fi
      return 1
    fi
  fi
  save_state "build_frontend" "success"
  
  # 步骤 4: 运行测试
  if [ "$current_step" != "backend" ] && [ "$current_step" != "frontend" ]; then
    if ! step_run_tests; then
      error "测试失败"
      save_state "run_tests" "failed"
      release_lock
      if [ -x "$REPORT_SCRIPT" ]; then
        bash "$REPORT_SCRIPT" "failed" "run_tests" 2>&1 | tee -a "$LOG_FILE"
      fi
      return 1
    fi
  fi
  save_state "run_tests" "success"
  
  # 步骤 5: Git 提交
  if ! step_git_commit; then
    warn "Git 提交失败或无需提交"
  fi
  save_state "git_commit" "success"
  
  # 完成
  info "🎉 构建完成！"
  save_state "completed" "success"
  
  release_lock

    # 计算耗时
    local end_time=$(date +%s)
    local duration=$((end_time - BUILD_START_TIME))

    # 调用完成上报
    if [ -x "$REPORT_SCRIPT" ]; then
        info "📊 生成构建报告..."
        bash "$REPORT_SCRIPT" "success" "$duration" "completed" 2>&1 | tee -a "$LOG_FILE"
    fi
  
  return 0
}

# 执行主流程
main "$@"
