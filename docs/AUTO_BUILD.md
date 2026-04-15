# 自动构建系统文档

## 📋 概述

ModelMagic 项目已配置全自动构建系统，每小时自动执行编译、测试、提交流程。

## 🚀 核心功能

### 1. API 限流保护
防止频繁调用外部 API 导致被封锁。

**配置参数:**
```bash
API_MAX_REQUESTS=100   # 每分钟最大请求数
API_WINDOW=60          # 时间窗口（秒）
```

**工作原理:**
- 令牌桶算法控制请求频率
- 超限时自动等待，直到时间窗口重置
- 适用于 npm install、go mod tidy 等操作

### 2. 子任务超时控制
防止单个任务卡死导致整个构建流程阻塞。

**超时配置:**
| 步骤 | 超时时间 | 说明 |
|------|---------|------|
| 后端编译 | 300s (5 分钟) | `go build` |
| 前端编译 | 600s (10 分钟) | `npm run build` |
| 测试执行 | 300s (5 分钟) | `go test ./...` |
| Git 操作 | 60s (1 分钟) | `git add/commit` |

**超时行为:**
- 自动终止超时任务
- 记录详细错误日志
- 根据重试策略决定是否重试

### 3. 指数退避重试
网络波动或临时故障时自动重试，避免雪崩效应。

**配置参数:**
```bash
MAX_RETRIES=3           # 最大重试次数
BASE_DELAY=5            # 基础延迟（秒）
EXPONENTIAL_BASE=2      # 指数基数
MAX_DELAY=60            # 最大延迟（秒）
```

**退避策略:**
- 第 1 次失败：等待 5 秒
- 第 2 次失败：等待 10 秒
- 第 3 次失败：等待 20 秒
- 超过最大延迟：统一等待 60 秒

### 4. 断点续建
构建失败后，1 小时内可从上一步骤恢复。

**状态文件:** `.auto-build-state`
```
<步骤>|<状态>|<时间戳>
```

**状态流转:**
```
none -> started -> check_env -> build_backend -> build_frontend -> run_tests -> git_commit -> completed
                              ↓
                         failed (可恢复)
```

### 5. 并发锁
防止多个构建任务同时执行导致冲突。

**锁文件:** `.auto-build.lock`
- 进程启动时获取锁
- 进程退出时释放锁
- 检测残留锁自动清理

## 📁 文件结构

```
modelmagic-deploy-console/
├── auto-build.sh              # 主构建脚本
├── .auto-build-state          # 状态文件（运行时生成）
├── .auto-build.lock           # 锁文件（运行时生成）
├── auto-build.log             # 构建日志
├── docs/
│   └── AUTO_BUILD.md         # 本文档
└── backend/
    ├── cmd/server/           # 后端入口
    ├── internal/             # 内部包
    └── test/integration/     # 集成测试
```

## 🔧 配置说明

### 环境变量
```bash
# 开发模式（可选）
export ENV="development"

# Secret 加密密钥（生产环境必须）
export SECRET_ENCRYPTION_KEY="your-32-byte-key-here"
```

### 超时调整
根据项目规模调整超时时间：
```bash
TIMEOUT_BACKEND=300      # 大型项目可增至 600
TIMEOUT_FRONTEND=600     # 大型前端可增至 900
TIMEOUT_TEST=300         # 测试多可增至 600
```

### 限流调整
根据 API 提供商限制调整：
```bash
API_MAX_REQUESTS=50      # 严格限制时调小
API_WINDOW=60            # 保持 60 秒窗口
```

## 📊 监控与日志

### 实时日志
```bash
tail -f /root/.openclaw/workspace/repos/modelmagic-deploy-console/auto-build.log
```

### 状态查询
```bash
cat /root/.openclaw/workspace/repos/modelmagic-deploy-console/.auto-build-state
```

### 构建历史
```bash
grep "构建完成\|失败\|中断" auto-build.log
```

## 🛠️ 故障排查

### 问题 1: 构建卡在 npm install
**现象:** 前端编译超时
**原因:** 网络波动或依赖过大
**解决:** 
```bash
# 检查网络
curl -I https://registry.npmjs.org

# 清理缓存
cd frontend && rm -rf node_modules package-lock.json
npm cache clean --force

# 使用国内镜像
npm config set registry https://registry.npmmirror.com
```

### 问题 2: Go 编译内存不足
**现象:** `go build` 被 OOM Killer 终止
**原因:** 项目过大或内存不足
**解决:**
```bash
# 限制编译内存
export GOGC=50

# 增加 swap
sudo fallocate -l 2G /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

### 问题 3: Git 提交冲突
**现象:** `git commit` 失败
**原因:** 手动修改与自动提交冲突
**解决:**
```bash
cd /root/.openclaw/workspace/repos/modelmagic-deploy-console

# 查看冲突
git status

# 手动解决后重置状态
git reset --hard HEAD
rm -f .auto-build-state
```

## 📈 性能优化建议

1. **使用构建缓存**
   - Go: `go build -cache`
   - npm: 使用 `npm ci` 代替 `npm install`

2. **增量编译**
   - 只编译变更的模块
   - 使用 `webpack --watch` 模式

3. **并行执行**
   - 后端和前端可并行编译
   - 测试分组件并行执行

4. **减少依赖**
   - 定期清理未使用依赖
   - 使用轻量级替代库

## 🎯 最佳实践

1. **定期审查日志**
   - 每天检查 `auto-build.log`
   - 关注重复失败的步骤

2. **监控资源使用**
   - CPU/内存使用率
   - 磁盘空间（特别是 `node_modules`）

3. **定期清理**
   - 删除旧的构建产物
   - 清理过期的状态文件

4. **备份配置**
   - 定期备份 `auto-build.sh`
   - 保存关键配置参数

## 📞 支持

遇到问题请查看：
- 构建日志：`auto-build.log`
- 状态文件：`.auto-build-state`
- 系统日志：`journalctl -u openclaw`

---

*文档版本：v2.0*
*最后更新：2026-04-16*
*维护者：超级龙虾队*
