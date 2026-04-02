#!/bin/bash

# ModelMagic Deploy Console - mmctl.sh
# 用于管理 ModelMagic 部署的脚本

set -e

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="${BASE_DIR}/workspaces"
VALUES_DIR="${BASE_DIR}/charts"

# 日志函数
log_info() {
    echo "[INFO] $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo "[ERROR] $(date '+%Y-%m-%d %H:%M:%S') - $1" >&2
}

# 01 - 解压安装包
extract_packages() {
    log_info "开始解压安装包..."
    # 实际实现：解压安装包到指定目录
    log_info "解压完成"
}

# 02 - 生成新命名空间配置
new_namespace() {
    local namespace=$1
    log_info "生成命名空间配置：$namespace"
    mkdir -p "${WORK_DIR}/${namespace}"
    # 创建默认配置文件
    cat > "${WORK_DIR}/${namespace}/hosts" << EOF
all:
  hosts:
    node1:
      ansible_host: 127.0.0.1
EOF
    cat > "${WORK_DIR}/${namespace}/main.yml" << EOF
namespace: ${namespace}
version: latest
EOF
    log_info "配置生成完成"
}

# 03 - 检查环境
check_environment() {
    local namespace=$1
    log_info "检查环境：$namespace"
    # 检查 kubectl, helm 等工具
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装"
        exit 1
    fi
    if ! command -v helm &> /dev/null; then
        log_error "helm 未安装"
        exit 1
    fi
    log_info "环境检查通过"
}

# 04 - 安装
install() {
    local namespace=$1
    local step=${2:-0}
    log_info "开始安装：$namespace (step: $step)"
    # 实际安装逻辑
    log_info "安装完成"
}

# 05 - 获取访问地址
get_access() {
    local namespace=$1
    log_info "获取访问地址：$namespace"
    # 实际实现：获取服务访问地址
    echo "http://${namespace}.example.com"
}

# 88 - 卸载
uninstall() {
    local namespace=$1
    local step=${2:-0}
    log_info "开始卸载：$namespace (step: $step)"
    # 实际卸载逻辑
    log_info "卸载完成"
}


# 91 - 扩缩容
scale() {
  local namespace=$1
  local replicas=$2
  log_info "开始扩缩容：$namespace -> $replicas 副本"

  # 参数验证
  if [ -z "$namespace" ]; then
    log_error "命名空间不能为空"
    exit 1
  fi

  if [ -z "$replicas" ]; then
    log_error "副本数不能为空"
    exit 1
  fi

  # 验证 replicas 为非负整数
  if ! echo "$replicas" | grep -qE '^[0-9]+$'; then
    log_error "无效的副本数，必须是非负整数"
    exit 1
  fi

  # 检查 kubectl 是否可用
  if ! command -v kubectl &> /dev/null; then
    log_error "kubectl 未安装"
    exit 1
  fi

  # 检查命名空间是否存在
  if ! kubectl get namespace "$namespace" &> /dev/null; then
    log_error "命名空间不存在：$namespace"
    exit 1
  fi

  # 执行 kubectl scale
  if kubectl scale deployment/"$namespace" --replicas="$replicas" -n "$namespace"; then
    log_info "扩缩容成功：$namespace -> $replicas 副本"
    echo "SUCCESS: Scaled deployment/$namespace to $replicas replicas"
  else
    log_error "扩缩容失败"
    exit 1
  fi
}

# 89 - 升级
upgrade() {
    local namespace=$1
    local version=$2
    
    log_info "开始升级：$namespace -> $version"
    
    # 参数验证
    if [ -z "$namespace" ]; then
        log_error "命名空间不能为空"
        exit 1
    fi
    
    if [ -z "$version" ]; then
        log_error "版本号不能为空"
        exit 1
    fi
    
    # 命名空间格式验证
    if ! echo "$namespace" | grep -qE '^[a-z0-9-]+$'; then
        log_error "无效的命名空间名称"
        exit 1
    fi
    
    # 版本号格式验证 (简单验证)
    if ! echo "$version" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
        log_error "无效的版本号格式 (期望：x.y.z)"
        exit 1
    fi
    
    local config_dir="${WORK_DIR}/${namespace}"
    
    # 检查命名空间是否存在
    if [ ! -d "$config_dir" ]; then
        log_error "命名空间配置不存在：$namespace"
        exit 1
    fi
    
    # 步骤 1: 备份当前配置
    log_info "备份当前配置..."
    local backup_dir="${config_dir}/backup_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$backup_dir"
    cp -r "${config_dir}"/* "$backup_dir/" 2>/dev/null || true
    log_info "配置已备份到：$backup_dir"
    
    # 步骤 2: 更新 values.yaml
    log_info "更新版本配置到 $version..."
    local main_yml="${config_dir}/main.yml"
    if [ -f "$main_yml" ]; then
        # 更新版本号
        if grep -q "^version:" "$main_yml"; then
            sed -i "s/^version:.*/version: ${version}/" "$main_yml"
        else
            echo "version: ${version}" >> "$main_yml"
        fi
        log_info "版本已更新为 $version"
    else
        log_error "配置文件不存在：$main_yml"
        exit 1
    fi
    
    # 步骤 3: 执行 helm upgrade
    log_info "执行 helm upgrade..."
    
    # 检查 helm 是否可用
    if ! command -v helm &> /dev/null; then
        log_error "helm 未安装"
        exit 1
    fi
    
    # 检查 kubectl 是否可用
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装"
        exit 1
    fi
    
    # 获取 chart 路径
    local chart_path="${VALUES_DIR}/modelmagic"
    if [ ! -d "$chart_path" ]; then
        log_error "Chart 目录不存在：$chart_path"
        exit 1
    fi
    
    # 执行 helm upgrade
    if helm upgrade "$namespace" "$chart_path" \
        --namespace "$namespace" \
        --create-namespace \
        --values "$main_yml" \
        --set image.tag="$version" \
        --wait \
        --timeout 300s; then
        log_info "helm upgrade 成功"
    else
        log_error "helm upgrade 失败"
        # 尝试回滚
        log_info "尝试回滚到上一个版本..."
        helm rollback "$namespace" --namespace "$namespace" || log_error "回滚失败"
        exit 1
    fi
    
    # 步骤 4: 等待部署完成
    log_info "等待部署完成..."
    if kubectl rollout status deployment/"$namespace" -n "$namespace" --timeout=300s; then
        log_info "部署状态检查通过"
    else
        log_error "部署状态检查失败"
        exit 1
    fi
    
    log_info "升级完成：$namespace -> $version"
    echo "SUCCESS: Upgrade completed successfully"
}

# 主函数 - 路由命令
main() {
    local command=$1
    shift
    
    case $command in
        "01")
            extract_packages "$@"
            ;;
        "02")
            new_namespace "$@"
            ;;
        "03")
            check_environment "$@"
            ;;
        "04")
            install "$@"
            ;;
        "05")
            get_access "$@"
            ;;
        "88")
            uninstall "$@"
            ;;
        "91") scale "$@" ;;
 "89") upgrade "$@" ;;
	"90") rollback "$@" ;;
        *)
            echo "Usage: $0 <command> [args...]"
            echo "Commands:"
            echo "  01 - Extract packages"
            echo "  02 - Create namespace config"
            echo "  03 - Check environment"
            echo "  04 - Install"
            echo "  05 - Get access URL"
            echo "  88 - Uninstall"
            echo "  89 - Upgrade"
            exit 1
            ;;
    esac
}

# 只有当脚本直接运行时才执行 main
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
