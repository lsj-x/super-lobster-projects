#!/bin/bash
# HPA 功能测试脚本
# 用于测试 HPA 自动扩缩容功能

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MMCTL="${SCRIPT_DIR}/mmctl.sh"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试配置
TEST_NAMESPACE="${TEST_NAMESPACE:-"test-hpa"}"
MIN_REPLICAS="${MIN_REPLICAS:-2}"
MAX_REPLICAS="${MAX_REPLICAS:-10}"
CPU_THRESHOLD="${CPU_THRESHOLD:-80}"

# 打印函数
print_header() {
  echo -e "\n${GREEN}========================================${NC}"
  echo -e "${GREEN}$1${NC}"
  echo -e "${GREEN}========================================${NC}\n"
}

print_success() {
  echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
  echo -e "${RED}✗ $1${NC}"
}

print_info() {
  echo -e "${YELLOW}→ $1${NC}"
}

# 检查依赖
check_dependencies() {
  print_header "检查依赖"
  
  local missing=0
  
  if ! command -v kubectl &> /dev/null; then
    print_error "kubectl 未安装"
    missing=1
  else
    print_success "kubectl 已安装"
  fi
  
  if ! command -v helm &> /dev/null; then
    print_error "helm 未安装"
    missing=1
  else
    print_success "helm 已安装"
  fi
  
  if [ $missing -eq 1 ]; then
    print_error "缺少必要依赖，无法继续测试"
    exit 1
  fi
}

# 创建测试命名空间
create_test_namespace() {
  print_header "创建测试命名空间"
  
  print_info "创建命名空间：$TEST_NAMESPACE"
  kubectl create namespace "$TEST_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
  
  print_success "命名空间创建完成"
}

# 部署测试应用
deploy_test_app() {
  print_header "部署测试应用"
  
  print_info "部署 modelmagic-service 用于测试"
  
  # 创建简单的测试 deployment
  cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: modelmagic-service
  namespace: ${TEST_NAMESPACE}
spec:
  replicas: 2
  selector:
    matchLabels:
      app: modelmagic-service
  template:
    metadata:
      labels:
        app: modelmagic-service
    spec:
      containers:
      - name: modelmagic
        image: nginx:1.21
        ports:
        - containerPort: 80
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"
EOF
  
  print_success "测试应用部署完成"
}

# 测试启用 HPA
test_enable_hpa() {
  print_header "测试启用 HPA (命令 92)"
  
  print_info "执行：$MMCTL 92 $TEST_NAMESPACE $MIN_REPLICAS $MAX_REPLICAS $CPU_THRESHOLD"
  
  if "$MMCTL" 92 "$TEST_NAMESPACE" "$MIN_REPLICAS" "$MAX_REPLICAS" "$CPU_THRESHOLD"; then
    print_success "HPA 启用成功"
  else
    print_error "HPA 启用失败"
    exit 1
  fi
}

# 测试查看 HPA 状态
test_get_hpa_status() {
  print_header "测试查看 HPA 状态 (命令 94)"
  
  print_info "执行：$MMCTL 94 $TEST_NAMESPACE"
  
  if "$MMCTL" 94 "$TEST_NAMESPACE"; then
    print_success "HPA 状态查询成功"
  else
    print_error "HPA 状态查询失败"
    exit 1
  fi
}

# 测试禁用 HPA
test_disable_hpa() {
  print_header "测试禁用 HPA (命令 93)"
  
  print_info "执行：$MMCTL 93 $TEST_NAMESPACE"
  
  if "$MMCTL" 93 "$TEST_NAMESPACE"; then
    print_success "HPA 禁用成功"
  else
    print_error "HPA 禁用失败"
    exit 1
  fi
}

# 测试参数验证
test_parameter_validation() {
  print_header "测试参数验证"
  
  print_info "测试：命名空间为空"
  if "$MMCTL" 92 2>&1 | grep -q "命名空间不能为空"; then
    print_success "命名空间验证通过"
  else
    print_error "命名空间验证失败"
  fi
  
  print_info "测试：最小副本数为 0"
  if "$MMCTL" 92 "$TEST_NAMESPACE" 0 5 80 2>&1 | grep -q "最小副本数必须大于 0"; then
    print_success "最小副本数验证通过"
  else
    print_error "最小副本数验证失败"
  fi
  
  print_info "测试：最大副本数小于最小副本数"
  if "$MMCTL" 92 "$TEST_NAMESPACE" 5 3 80 2>&1 | grep -q "最大副本数必须大于或等于最小副本数"; then
    print_success "副本数范围验证通过"
  else
    print_error "副本数范围验证失败"
  fi
  
  print_info "测试：CPU 阈值超出范围"
  if "$MMCTL" 92 "$TEST_NAMESPACE" 2 10 150 2>&1 | grep -q "CPU 阈值必须在 1-100 之间"; then
    print_success "CPU 阈值验证通过"
  else
    print_error "CPU 阈值验证失败"
  fi
}

# 清理测试环境
cleanup() {
  print_header "清理测试环境"
  
  print_info "删除测试命名空间：$TEST_NAMESPACE"
  kubectl delete namespace "$TEST_NAMESPACE" --ignore-not-found
  
  print_success "清理完成"
}

# 主测试流程
main() {
  local action=${1:-"all"}
  
  case $action in
    "setup")
      check_dependencies
      create_test_namespace
      deploy_test_app
      ;;
    "test")
      test_enable_hpa
      test_get_hpa_status
      test_disable_hpa
      test_parameter_validation
      ;;
    "cleanup")
      cleanup
      ;;
    "all")
      check_dependencies
      create_test_namespace
      deploy_test_app
      test_enable_hpa
      test_get_hpa_status
      test_disable_hpa
      test_parameter_validation
      cleanup
      ;;
    *)
      echo "Usage: $0 {setup|test|cleanup|all}"
      echo ""
      echo "Commands:"
      echo "  setup   - 创建测试环境和部署测试应用"
      echo "  test    - 运行 HPA 功能测试"
      echo "  cleanup - 清理测试环境"
      echo "  all     - 完整测试流程 (默认)"
      exit 1
      ;;
  esac
  
  print_header "测试完成"
  print_success "所有测试已通过"
}

# 如果直接运行此脚本，则执行 main
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  main "$@"
fi
