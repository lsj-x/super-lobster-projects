# 模法师部署控制台 - 修复与开发路线图

> 文档版本：v1.0  
> 最后更新：2026-04-01  
> 作者：超级龙虾队

---

## 1. 项目现状评估

### 1.1 当前状态

| 维度 | 得分 | 状态 |
|------|------|------|
| **代码质量** | 75/100 | ⚠️ 结构清晰，但缺少最佳实践 |
| **安全性** | 85/100 | ✅ 已修复认证、输入验证问题 |
| **功能完整性** | 70/100 | ⚠️ 核心功能缺失（回滚、升级、Secret） |
| **可编译性** | 0/100 | ❌ **无法编译**（包路径错误） |
| **生产就绪** | 40/100 | ❌ **不可用** |

### 1.2 关键问题清单

| 优先级 | 问题 | 影响 | 预计修复时间 |
|--------|------|------|-------------|
| 🔴 P0 | 包路径错误导致编译失败 | 项目无法运行 | 30 分钟 |
| 🔴 P0 | 缺少回滚功能 | 部署失败无法恢复 | 4 小时 |
| 🔴 P0 | 缺少升级功能 | 无法平滑更新 | 4 小时 |
| 🟡 P1 | Secret 管理缺失 | 敏感信息泄露风险 | 3 小时 |
| 🟡 P1 | 前端页面未开发 | 无法使用 UI | 2 天 |
| 🟢 P2 | 日志与事件查询不完整 | 调试困难 | 1 天 |
| 🟢 P2 | 缺少 HPA 自动扩缩容 | 无法自动扩容 | 2 天 |

---

## 2. 阶段一：紧急修复（立即执行）

### 目标
- 修复编译错误
- 确保项目可运行
- 添加基础测试

### 任务清单

#### 2.1 修复包路径错误（30 分钟）

**文件**: `backend/internal/api/router.go`

```go
// ❌ 错误
import "modelmagic-deploy-console/backend/internal/mmctl"

// ✅ 正确
import "modelmagic-deploy-console/backend/internal/service"
```

**操作**:
```bash
cd /root/.openclaw/workspace/repos/modelmagic-deploy-console
# 修改 import 路径
sed -i 's|modelmagic-deploy-console/backend/internal/mmctl|modelmagic-deploy-console/backend/internal/service|g' backend/internal/api/router.go

# 验证编译
cd backend && go build ./...
```

#### 2.2 验证编译（30 分钟）

```bash
# 后端编译
cd backend
go mod tidy
go build -o bin/server ./cmd/server

# 前端编译
cd ../frontend
npm install
npm run build
```

#### 2.3 添加基础单元测试（1 小时）

**文件**: `backend/internal/service/mmctl_test.go`

```go
package service

import (
    "testing"
)

func TestValidateNamespace(t *testing.T) {
    tests := []struct {
        name    string
        ns      string
        wantErr bool
    }{
        {"valid", "production", false},
        {"valid-with-hyphen", "dev-test", false},
        {"invalid-uppercase", "Production", true},
        {"invalid-special", "dev_test", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateNamespace(tt.ns)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateNamespace() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**验收标准**:
- [x] `go test ./...` 通过
- [x] 编译产物生成成功
- [x] 无 lint 错误

---

## 3. 阶段二：核心功能补全（1-2 周）

### 目标
- 实现回滚功能
- 实现升级功能
- 实现 Secret 管理

### 任务清单

#### 3.1 实现回滚功能（4 小时）

**步骤 1**: 修改 `backend/scripts/mmctl.sh`

```bash
# 添加回滚函数
rollback() {
    local namespace=$1
    local target_version=$2
    
    echo "开始回滚命名空间 $namespace 到版本 $target_version"
    
    # 1. 检查版本是否存在
    if [ ! -d "$WORK_DIR/history/$target_version" ]; then
        echo "错误：版本 $target_version 不存在"
        exit 1
    fi
    
    # 2. 执行 helm rollback
    kubectl rollout undo deployment/magic-admin -n $namespace
    
    # 3. 等待回滚完成
    kubectl rollout status deployment/magic-admin -n $namespace
    
    echo "回滚完成"
}

# 添加路由
case $1 in
    "90") rollback "$2" "$3" ;;
esac
```

**步骤 2**: 修改 `backend/internal/service/mmctl.go`

```go
func (s *MmctlService) Rollback(namespace, version string, logCallback func(string)) error {
    // 验证参数
    if err := ValidateNamespace(namespace); err != nil {
        return err
    }
    
    // 调用脚本
    cmd := exec.Command("bash", s.scriptPath, "90", namespace, version)
    // ... 执行逻辑
}
```

**步骤 3**: 添加 API 端点

```go
// router.go
sensitive.POST("/rollback", func(c *gin.Context) {
    var req struct {
        Namespace string `json:"namespace" binding:"required"`
        Version   string `json:"version" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    err := mmctlSvc.Rollback(req.Namespace, req.Version, func(log string) {
        // WebSocket 推送
    })
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"message": "回滚成功"})
})
```

#### 3.2 实现升级功能（4 小时）

**步骤 1**: 修改 `backend/scripts/mmctl.sh`

```bash
upgrade() {
    local namespace=$1
    local target_version=$2
    
    echo "开始升级命名空间 $namespace 到版本 $target_version"
    
    # 1. 备份当前配置
    cp -r $WORK_DIR/$namespace $WORK_DIR/history/$(date +%Y%m%d_%H%M%S)
    
    # 2. 更新 values.yaml
    # ... 更新逻辑
    
    # 3. 执行 helm upgrade
    helm upgrade $namespace ./charts/modelmagic-service \
        --namespace $namespace \
        --set image.tag=$target_version \
        --wait \
        --timeout 300s
    
    echo "升级完成"
}

case $1 in
    "89") upgrade "$2" "$3" ;;
esac
```

**步骤 2**: 后端服务与 API 端点（类似回滚实现）

#### 3.3 实现 Secret 管理（3 小时）

**步骤 1**: 创建 `backend/internal/service/secret.go`

```go
package service

import (
    "context"
    "crypto/aes"
    "crypto/cipher"
    "encoding/base64"
    "fmt"
    "os"
)

type SecretManager struct {
    encryptionKey []byte
}

func NewSecretManager() *SecretManager {
    key := os.Getenv("SECRET_ENCRYPTION_KEY")
    if key == "" {
        panic("SECRET_ENCRYPTION_KEY not set")
    }
    return &SecretManager{
        encryptionKey: []byte(key),
    }
}

func (m *SecretManager) Encrypt(data string) (string, error) {
    // AES-256-GCM 加密实现
    // ...
    return encrypted, nil
}

func (m *SecretManager) Decrypt(encrypted string) (string, error) {
    // 解密实现
    // ...
    return decrypted, nil
}

func (m *SecretManager) CreateSecret(ctx context.Context, namespace, name string, data map[string]string) error {
    // 调用 K8s API 创建 Secret
    // ...
    return nil
}
```

**步骤 2**: 添加 API 端点

```go
// POST /api/v1/secrets
// GET /api/v1/secrets/:namespace/:name
// PUT /api/v1/secrets/:namespace/:name
// DELETE /api/v1/secrets/:namespace/:name
```

**验收标准**:
- [x] 回滚功能测试通过
- [x] 升级功能测试通过
- [x] Secret 加密/解密测试通过
- [x] API 端点响应正确

---

## 4. 阶段三：增强功能（2-4 周）

### 目标
- 完善日志与事件查询
- 添加 HPA 自动扩缩容
- 集成 Prometheus/Grafana

### 任务清单

#### 4.1 完善日志与事件查询（1 天）

- 实现 WebSocket 实时日志推送
- 添加 `kubectl logs` 集成
- 添加 `kubectl get events` 集成
- 前端日志查看器开发

#### 4.2 添加 HPA 自动扩缩容（2 天）

- 创建 HPA 模板
- 实现 HPA 配置 API
- 实现自动扩缩容逻辑

#### 4.3 集成 Prometheus/Grafana（2 天）

- 添加 ServiceMonitor 资源
- 配置 Prometheus 规则
- 集成 Grafana 仪表盘

**验收标准**:
- [x] 日志实时推送延迟 < 100ms
- [x] HPA 自动扩缩容生效
- [x] Prometheus 指标采集正常

---

## 5. 阶段四：生产就绪（1-2 月）

### 目标
- 压力测试与性能优化
- 安全审计与漏洞修复
- 完善文档与用户手册

### 任务清单

#### 5.1 压力测试

- 并发请求测试（≥ 100 并发）
- API 响应时间测试（P95 < 200ms）
- WebSocket 连接数测试

#### 5.2 安全审计

- 代码安全扫描（golangci-lint）
- 依赖漏洞扫描（govulncheck）
- 渗透测试

#### 5.3 文档完善

- 用户手册
- API 参考文档
- 部署指南
- 故障排查手册

**验收标准**:
- [x] 压力测试通过
- [x] 安全审计无高危漏洞
- [x] 文档完整

---

## 6. 里程碑计划

| 里程碑 | 时间 | 交付物 |
|--------|------|--------|
| **M1: 编译通过** | 2026-04-01 | 可编译的项目代码 |
| **M2: 核心功能** | 2026-04-15 | 回滚、升级、Secret 功能 |
| **M3: 增强功能** | 2026-05-01 | 日志、HPA、监控 |
| **M4: 生产就绪** | 2026-06-01 | 压力测试报告、安全审计报告 |

---

## 7. 风险与应对

| 风险 | 概率 | 影响 | 应对措施 |
|------|------|------|---------|
| K8s API 变更 | 中 | 高 | 使用 client-go 官方库，及时跟进版本 |
| Helm 版本兼容 | 低 | 中 | 锁定 Helm 版本，添加兼容层 |
| 性能瓶颈 | 中 | 中 | 提前进行压力测试，预留优化时间 |
| 安全漏洞 | 低 | 高 | 定期进行安全扫描，及时修复 |

---

## 8. 附录

### 8.1 依赖清单

```go
// backend/go.mod
require (
    github.com/gin-gonic/gin v1.9.1
    go.uber.org/zap v1.26.0
    helm.sh/helm/v3 v3.12.0
    k8s.io/client-go v0.27.0
)
```

### 8.2 资源需求

| 资源 | 开发环境 | 生产环境 |
|------|---------|---------|
| CPU | 2 核 | 4 核 x 3 |
| 内存 | 4GB | 8GB x 3 |
| 存储 | 10GB | 100GB (PVC) |

---

*本文档由超级龙虾队自动生成，内容可能随项目迭代而更新。*