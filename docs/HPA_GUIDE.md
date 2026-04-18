# HPA 自动扩缩容使用指南

## 概述

HPA (Horizontal Pod Autoscaler) 是 Kubernetes 的自动扩缩容功能，可以根据 CPU 或内存使用率自动调整 Pod 副本数量。

## 功能命令

### 92 - 启用 HPA 自动扩缩容

```bash
./mmctl.sh 92 <namespace> [min_replicas] [max_replicas] [cpu_threshold]
```

**参数说明：**
- `namespace` (必需): 目标命名空间
- `min_replicas` (可选): 最小副本数，默认 1
- `max_replicas` (可选): 最大副本数，默认 10
- `cpu_threshold` (可选): CPU 使用率阈值，默认 80%

**示例：**
```bash
# 使用默认参数启用 HPA
./mmctl.sh 92 my-namespace

# 自定义参数：最小 2 个副本，最大 20 个副本，CPU 阈值 70%
./mmctl.sh 92 my-namespace 2 20 70
```

**参数验证规则：**
- 最小副本数必须 ≥ 1
- 最大副本数必须 ≥ 最小副本数
- CPU 阈值必须在 1-100 之间

**前置条件：**
1. kubectl 已安装并配置
2. 目标命名空间存在
3. deployment `modelmagic-service` 已存在
4. metrics-server 已安装（用于获取 CPU 指标）

### 93 - 禁用 HPA

```bash
./mmctl.sh 93 <namespace>
```

**参数说明：**
- `namespace` (必需): 目标命名空间

**示例：**
```bash
./mmctl.sh 93 my-namespace
```

**说明：**
- 删除 HPA 配置
- 建议禁用后手动设置副本数
- 如果 HPA 不存在，会显示提示信息

### 94 - 查看 HPA 状态

```bash
./mmctl.sh 94 <namespace>
```

**参数说明：**
- `namespace` (必需): 目标命名空间

**示例：**
```bash
./mmctl.sh 94 my-namespace
```

**输出内容：**
1. HPA 基本信息（当前副本数、目标等）
2. HPA 详细信息（事件、指标等）
3. 相关 Deployment 信息
4. Pod 资源使用情况

## 使用流程

### 1. 环境准备

确保已安装以下工具：
```bash
# 检查 kubectl
kubectl version --client

# 检查 metrics-server
kubectl top nodes
```

### 2. 部署应用

确保 `modelmagic-service` deployment 已部署：
```bash
kubectl get deployment modelmagic-service -n <namespace>
```

### 3. 启用 HPA

```bash
./mmctl.sh 92 <namespace> 2 10 80
```

### 4. 监控 HPA 状态

```bash
./mmctl.sh 94 <namespace>
```

### 5. 禁用 HPA（可选）

```bash
./mmctl.sh 93 <namespace>
```

## 故障排查

### 问题 1: "无法获取指标，请确保 metrics-server 已安装"

**原因：** metrics-server 未安装或未正常运行

**解决方案：**
```bash
# 安装 metrics-server（如果需要）
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# 检查 metrics-server 状态
kubectl get pods -n kube-system | grep metrics-server
```

### 问题 2: "Deployment 不存在：modelmagic-service"

**原因：** 目标命名空间中没有名为 `modelmagic-service` 的 deployment

**解决方案：**
```bash
# 检查 deployment 列表
kubectl get deployments -n <namespace>

# 创建或修复 deployment
kubectl create deployment modelmagic-service --image=<your-image> -n <namespace>
```

### 问题 3: HPA 状态显示 "unable to get metrics"

**原因：** metrics-server 数据未就绪或 deployment 没有设置资源请求

**解决方案：**
```bash
# 检查 deployment 的资源配置
kubectl get deployment modelmagic-service -n <namespace> -o yaml

# 确保设置了 resources.requests.cpu
```

## 最佳实践

### 1. 合理设置副本数范围

- **minReplicas**: 建议设置为至少 2，保证高可用
- **maxReplicas**: 根据实际资源和成本限制设置
- **cpuThreshold**: 一般设置为 70-80%，留出缓冲空间

### 2. 监控和调整

定期检查 HPA 状态：
```bash
# 每 5 分钟检查一次
watch -n 300 ./mmctl.sh 94 <namespace>
```

### 3. 资源限制

确保 deployment 设置了合理的资源请求和限制：
```yaml
resources:
  requests:
    cpu: "100m"
    memory: "128Mi"
  limits:
    cpu: "500m"
    memory: "512Mi"
```

### 4. 测试环境

使用测试脚本验证 HPA 配置：
```bash
cd backend/scripts
chmod +x hpa-test.sh
./hpa-test.sh all
```

## 示例场景

### 场景 1: 生产环境配置

```bash
# 高可用配置：最小 3 个副本，最大 50 个，CPU 阈值 70%
./mmctl.sh 92 production 3 50 70
```

### 场景 2: 开发环境配置

```bash
# 经济配置：最小 1 个副本，最大 5 个，CPU 阈值 80%
./mmctl.sh 92 development 1 5 80
```

### 场景 3: 临时高负载

```bash
# 临时提高上限应对流量高峰
./mmctl.sh 92 production 5 100 60
```

## 相关文档

- [Kubernetes HPA 官方文档](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
- [metrics-server 项目](https://github.com/kubernetes-sigs/metrics-server)
- [HPA 测试脚本](../backend/scripts/hpa-test.sh)
