<template>
  <div class="deploy-page">
    <div class="page-header">
      <h2>🚀 模型部署</h2>
      <el-button type="primary" @click="showDeployDialog = true">
        <el-icon><Plus /></el-icon> 新部署
      </el-button>
    </div>

    <!-- 部署列表 -->
    <el-table :data="deployments" style="width: 100%" v-loading="loading">
      <el-table-column prop="namespace" label="命名空间" width="120" />
      <el-table-column prop="model_name" label="模型名称" width="200" />
      <el-table-column prop="version" label="版本" width="100" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="getStatusType(row.status)">
            {{ getStatusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="replicas" label="副本数" width="80" />
      <el-table-column prop="ready_replicas" label="就绪" width="80" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="handleScale(row)">扩容</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 部署对话框 -->
    <el-dialog v-model="showDeployDialog" title="部署模型" width="500px">
      <el-form :model="deployForm" label-width="100px">
        <el-form-item label="命名空间">
          <el-select v-model="deployForm.namespace" placeholder="选择命名空间">
            <el-option v-for="ns in namespaces" :key="ns.name" :label="ns.name" :value="ns.name" />
          </el-select>
        </el-form-item>
        <el-form-item label="模型名称" required>
          <el-input v-model="deployForm.modelName" placeholder="例如: llama-2-7b" />
        </el-form-item>
        <el-form-item label="版本">
          <el-input v-model="deployForm.version" placeholder="可选，例如: v1.0" />
        </el-form-item>
        <el-form-item label="副本数">
          <el-input-number v-model="deployForm.replicas" :min="0" :max="10" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showDeployDialog = false">取消</el-button>
        <el-button type="primary" @click="handleDeploy">部署</el-button>
      </template>
    </el-dialog>

    <!-- 扩容对话框 -->
    <el-dialog v-model="showScaleDialog" title="扩容部署" width="400px">
      <el-form :model="scaleForm" label-width="100px">
        <el-form-item label="副本数">
          <el-input-number v-model="scaleForm.replicas" :min="0" :max="10" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showScaleDialog = false">取消</el-button>
        <el-button type="primary" @click="handleScaleConfirm">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'

interface Deployment {
  namespace: string
  model_name: string
  version: string
  status: string
  replicas: number
  ready_replicas: number
  created_at: string
}

interface Namespace {
  name: string
  description: string
}

const loading = ref(false)
const showDeployDialog = ref(false)
const showScaleDialog = ref(false)
const deployments = ref<Deployment[]>([])
const namespaces = ref<Namespace[]>([])
const currentDeployment = ref<Deployment | null>(null)

const deployForm = ref({
  namespace: '',
  modelName: '',
  version: '',
  replicas: 1
})

const scaleForm = ref({
  replicas: 1
})

const getStatusType = (status: string) => {
  const types: Record<string, any> = {
    running: 'success',
    pending: 'warning',
    failed: 'danger',
    succeeded: 'success'
  }
  return types[status] || 'info'
}

const getStatusText = (status: string) => {
  const texts: Record<string, string> = {
    running: '运行中',
    pending: '等待中',
    failed: '失败',
    succeeded: '已完成'
  }
  return texts[status] || status
}

const fetchDeployments = async () => {
  loading.value = true
  try {
    deployments.value = [
      { namespace: 'production', model_name: 'llama-2-7b', version: 'v1.0', status: 'running', replicas: 2, ready_replicas: 2, created_at: new Date().toISOString() },
      { namespace: 'staging', model_name: 'mistral-7b', version: 'v1.2', status: 'pending', replicas: 1, ready_replicas: 0, created_at: new Date().toISOString() }
    ]
  } catch (error) {
    ElMessage.error('获取部署列表失败')
  } finally {
    loading.value = false
  }
}

const fetchNamespaces = async () => {
  try {
    namespaces.value = [
      { name: 'production', description: '生产环境' },
      { name: 'staging', description: '测试环境' },
      { name: 'development', description: '开发环境' }
    ]
  } catch (error) {
    ElMessage.error('获取命名空间失败')
  }
}

const handleDeploy = async () => {
  if (!deployForm.value.namespace || !deployForm.value.modelName) {
    ElMessage.warning('请填写完整信息')
    return
  }
  try {
    ElMessage.success('部署已启动')
    showDeployDialog.value = false
    fetchDeployments()
  } catch (error) {
    ElMessage.error('部署失败')
  }
}

const handleScale = (deployment: Deployment) => {
  currentDeployment.value = deployment
  scaleForm.value.replicas = deployment.replicas
  showScaleDialog.value = true
}

const handleScaleConfirm = async () => {
  if (!currentDeployment.value) return
  try {
    ElMessage.success('扩容已启动')
    showScaleDialog.value = false
    fetchDeployments()
  } catch (error) {
    ElMessage.error('扩容失败')
  }
}

const handleDelete = async (deployment: Deployment) => {
  try {
    await ElMessageBox.confirm(`确定要删除 ${deployment.namespace}/${deployment.model_name} 吗？`, '警告', { type: 'warning' })
    ElMessage.success('删除已启动')
    fetchDeployments()
  } catch {
    // 用户取消
  }
}

onMounted(() => {
  fetchDeployments()
  fetchNamespaces()
})
</script>

<style scoped>
.deploy-page { padding: 20px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.page-header h2 { margin: 0; }
</style>
