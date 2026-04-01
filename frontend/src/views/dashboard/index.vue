<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-icon"><el-icon><FolderOpened /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.namespaces }}</div>
            <div class="stat-label">环境数量</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-icon success"><el-icon><CircleCheck /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.running }}</div>
            <div class="stat-label">运行中</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-icon warning"><el-icon><Clock /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.deploying }}</div>
            <div class="stat-label">部署中</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-icon danger"><el-icon><CircleClose /></el-icon></div>
          <div class="stat-info">
            <div class="stat-value">{{ stats.failed }}</div>
            <div class="stat-label">失败</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近部署记录</span>
              <el-button type="primary" size="small" @click="$router.push('/namespaces')">查看所有</el-button>
            </div>
          </template>
          <el-table :data="recentDeployments" style="width: 100%">
            <el-table-column prop="namespace" label="环境名称" />
            <el-table-column prop="status" label="状态">
              <template #default="{ row }">
                <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="startTime" label="开始时间" />
            <el-table-column prop="duration" label="耗时" />
            <el-table-column label="操作">
              <template #default="{ row }">
                <el-button type="primary" size="small" @click="viewDetails(row)">详情</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header>
            <span>系统状态</span>
          </template>
          <div class="system-status">
            <div class="status-item">
              <span>mmctl 服务</span>
              <el-tag type="success">正常</el-tag>
            </div>
            <div class="status-item">
              <span>K8s 集群</span>
              <el-tag type="success">连接中</el-tag>
            </div>
            <div class="status-item">
              <span>NFS 存储</span>
              <el-tag type="warning">检查中</el-tag>
            </div>
            <div class="status-item">
              <span>Harbor 仓库</span>
              <el-tag type="success">可达</el-tag>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const stats = ref({
  namespaces: 3,
  running: 2,
  deploying: 1,
  failed: 0
})

const recentDeployments = ref([
  { namespace: 'magic-prod', status: '成功', startTime: '2026-04-01 10:30', duration: '15 分钟' },
  { namespace: 'magic-test', status: '部署中', startTime: '2026-04-01 14:00', duration: '进行中' },
  { namespace: 'magic-dev', status: '成功', startTime: '2026-03-31 16:20', duration: '12 分钟' }
])

const getStatusType = (status: string) => {
  const map: Record<string, any> = {
    '成功': 'success',
    '部署中': 'warning',
    '失败': 'danger'
  }
  return map[status] || 'info'
}

const viewDetails = (row: any) => {
  console.log('查看详情:', row)
}
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.stat-card {
  display: flex;
  align-items: center;
  padding: 20px;
}

.stat-icon {
  font-size: 48px;
  color: #409EFF;
  margin-right: 20px;
}

.stat-icon.success {
  color: #67C23A;
}

.stat-icon.warning {
  color: #E6A23C;
}

.stat-icon.danger {
  color: #F56C6C;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.system-status {
  padding: 10px 0;
}

.status-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.status-item:last-child {
  border-bottom: none;
}
</style>
