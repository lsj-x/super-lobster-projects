#!/bin/bash
# ModelMagic Deploy Console - 自动构建脚本
# 超级龙虾队自动化工作流

set -e

# ============ 配置区 ============
PROJECT_DIR="/root/.openclaw/workspace/repos/modelmagic-deploy-console"
LOG_FILE="${PROJECT_DIR}/auto-build.log"
STATE_FILE="${PROJECT_DIR}/.auto-build-state"
MAX_RETRIES=3
NOTIFY_CHANNEL="feishu"  # 通知渠道

# ============ 工具函数 ============
log() {
  local msg="[$(date '+%Y-%m-%d %H:%M:%S')] $1"
  echo "$msg"
  echo "$msg" >> "$LOG_FILE"
}

error() {
  local msg="[$(date '+%Y-%m-%d %H:%M:%S')] [ERROR] $1"
  echo "$msg" >&2
  echo "$msg" >> "$LOG_FILE"
}

# 保存进度状态
save_state() {
  local step="$1"
  local status="$2"
  echo "${step}|${status}|$(date +%s)" > "$STATE_FILE"
}

# 读取进度状态
load_state() {
  if [ -f "$STATE_FILE" ]; then
    cat "$STATE_FILE"
  else
    echo "none|none|0"
  fi
}

# 通知用户（飞书）
notify_user() {
  local title="$1"
  local content="$2"
  local level="${3:-info}"  # info, success, error
  
  log "📢 通知用户：$title"
  
  # 使用 OpenClaw 内部通知（如果可用）
  # 或者调用飞书 webhook
  # 这里简化为日志输出
  echo "📢 [$title] $content" >> "${LOG_FILE}.notify"
}

# ============ 构建步骤 ============

# 步骤 1: 环境检查
step_check_env() {
  log "🔍 步骤 1: 检查环境..."
  
  # 检查 Go 版本
  if ! command -v go &> /dev/null; then
    error "Go 未安装"
    return 1
  fi
  
  # 检查 Node.js
  if ! command -v node &> /dev/null; then
    error "Node.js 未安装"
    return 1
  fi
  
  # 检查 kubectl/helm（可选）
  if ! command -v kubectl &> /dev/null; then
    log "⚠️  kubectl 未安装（可选）"
  fi
  
  if ! command -v helm &> /dev/null; then
    log "⚠️  helm 未安装（可选）"
  fi
  
  log "✅ 环境检查通过"
  return 0
}

# 步骤 2: 后端编译
step_build_backend() {
  log "🔧 步骤 2: 编译后端..."
  
  cd "${PROJECT_DIR}/backend"
  
  # 清理并重新编译
  go mod tidy
  go build -o bin/server ./cmd/server
  
  if [ $? -eq 0 ]; then
    log "✅ 后端编译成功"
    return 0
  else
    error "❌ 后端编译失败"
    return 1
  fi
}

# 步骤 3: 前端编译
step_build_frontend() {
  log "🎨 步骤 3: 编译前端..."
  
  cd "${PROJECT_DIR}/frontend"
  
  # 安装依赖并编译
  npm install --legacy-peer-deps
  npm run build
  
  if [ $? -eq 0 ]; then
    log "✅ 前端编译成功"
    return 0
  else
    error "❌ 前端编译失败"
    return 1
  fi
}

# 步骤 4: 运行测试
step_run_tests() {
  log "🧪 步骤 4: 运行测试..."
  
  cd "${PROJECT_DIR}/backend"
  go test ./... -v
  
  if [ $? -eq 0 ]; then
    log "✅ 测试全部通过"
    return 0
  else
    error "❌ 测试失败"
    return 1
  fi
}

# 步骤 5: Git 提交
step_git_commit() {
  log "📝 步骤 5: Git 提交..."
  
  cd "${PROJECT_DIR}"
  
  # 检查是否有更改
  if [ -z "$(git status --porcelain)" ]; then
    log "ℹ️  没有更改需要提交"
    return 0
  fi
  
  # 添加所有更改
  git add -A
  
  # 生成提交信息
  local commit_msg="auto: 自动构建完成 $(date '+%Y-%m-%d %H:%M')"
  
  # 提交
  git commit -m "$commit_msg"
  
  log "✅ Git 提交成功：$commit_msg"
  return 0
}

# ============ 主流程 ============
main() {
  local current_step="${1:-all}"
  local retry_count=0
  
  log "🚀 超级龙虾队自动构建开始"
  log "📁 项目目录：$PROJECT_DIR"
  
  # 保存开始状态
  save_state "started" "running"
  
  # 步骤 1: 环境检查
  if ! step_check_env; then
    error "环境检查失败"
    notify_user "构建失败" "环境检查失败" "error"
    save_state "check_env" "failed"
    return 1
  fi
  save_state "check_env" "success"
  
  # 步骤 2: 后端编译
  if [ "$current_step" != "frontend" ] && [ "$current_step" != "test" ]; then
    while [ $retry_count -lt $MAX_RETRIES ]; do
      if step_build_backend; then
        break
      else
        retry_count=$((retry_count + 1))
        log "⚠️  后端编译失败，重试 $retry_count/$MAX_RETRIES"
        sleep 5
      fi
    done
    if [ $retry_count -eq $MAX_RETRIES ]; then
      error "后端编译失败，已达最大重试次数"
      notify_user "构建失败" "后端编译失败" "error"
      save_state "build_backend" "failed"
      return 1
    fi
  fi
  save_state "build_backend" "success"
  
  # 步骤 3: 前端编译
  if [ "$current_step" != "backend" ]; then
    retry_count=0
    while [ $retry_count -lt $MAX_RETRIES ]; do
      if step_build_frontend; then
        break
      else
        retry_count=$((retry_count + 1))
        log "⚠️  前端编译失败，重试 $retry_count/$MAX_RETRIES"
        sleep 5
      fi
    done
    if [ $retry_count -eq $MAX_RETRIES ]; then
      error "前端编译失败，已达最大重试次数"
      notify_user "构建失败" "前端编译失败" "error"
      save_state "build_frontend" "failed"
      return 1
    fi
  fi
  save_state "build_frontend" "success"
  
  # 步骤 4: 运行测试
  if [ "$current_step" != "backend" ] && [ "$current_step" != "frontend" ]; then
    retry_count=0
    while [ $retry_count -lt $MAX_RETRIES ]; do
      if step_run_tests; then
        break
      else
        retry_count=$((retry_count + 1))
        log "⚠️  测试失败，重试 $retry_count/$MAX_RETRIES"
        sleep 5
      fi
    done
    if [ $retry_count -eq $MAX_RETRIES ]; then
      error "测试失败，已达最大重试次数"
      notify_user "构建失败" "测试失败" "error"
      save_state "run_tests" "failed"
      return 1
    fi
  fi
  save_state "run_tests" "success"
  
  # 步骤 5: Git 提交
  retry_count=0
  while [ $retry_count -lt $MAX_RETRIES ]; do
    if step_git_commit; then
      break
    else
      retry_count=$((retry_count + 1))
      log "⚠️  Git 提交失败，重试 $retry_count/$MAX_RETRIES"
      sleep 5
    fi
  done
  if [ $retry_count -eq $MAX_RETRIES ]; then
    error "Git 提交失败"
    notify_user "构建警告" "Git 提交失败但构建成功" "error"
    # 不返回失败，因为构建本身成功了
  fi
  save_state "git_commit" "success"
  
  # 完成通知
  log "🎉 构建完成！"
  notify_user "构建成功" "ModelMagic 项目自动构建完成" "success"
  
  return 0
}

# 执行主流程
main "$@"
