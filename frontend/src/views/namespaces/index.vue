<template>
  <div class="namespaces">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>环境列表</span>
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            新建环境
          </el-button>
        </div>
      </template>

      <el-table :data="namespaces" style="width: 100%" v-loading="loading">
        <el-table-column prop="name" label="环境名称" width="180" />
        <el-table-column prop="namespace" label="命名空间" width="180" />
        <el-table-column prop="isHa" label="高可用" width="100">
          <template #default="{ row }">
            <el-tag :type="row.isHa ? 'success' : 'info'">{{ row.isHa ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="use_resource" label="资源规格" width="120" />
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" @click="deploy(row)">部署</el-button>
            <el-button type="success" size="small" @click="access(row)">访问</el-button>
            <el-button type="danger" size="small" @click="uninstall(row)">卸载</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建环境对话框 -->
    <el-dialog v-model="dialogVisible" title="新建环境" width="600px">
      <el-form :model="formData" label-width="120px">
        <el-form-item label="环境名称" required>
          <el-input v-model="formData.name" placeholder="例如：magic-prod" />
        </el-form-item>
        <el-form-item label="命名空间" required>
          <el-input v-model="formData.namespace" placeholder="k8s 命名空间名称" />
        </el-form-item>
        <el-form-item label="部署模式">
          <el-radio-group v-model="formData.mode">
            <el-radio label="single">单节点</el-radio>
            <el-radio label="multi">多节点</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="资源规格">
          <el-select v-model="formData.resource" placeholder="请选择">
            <el-option label="低配 (low)" value="low" />
            <el-option label="中配 (medium)" value="medium" />
            <el-option label="高配 (high)" value="high" />
          </el-select>
        </el-form-item>
        <el-form-item label="高可用">
          <el-switch v-model="formData.isHa" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="createNamespace">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const dialogVisible = ref(false)
const namespaces = ref<any[]>([])

const formData = ref({
  name: '',
  namespace: '',
  mode: 'single',
  resource: 'low',
  isHa: false
})

onMounted(() => {
  fetchNamespaces()
})

const fetchNamespaces = async () => {
  loading.value = true
  try {
    // 实际项目中调用 API
    // const res = await axios.get('/api/v1/namespaces')
    // namespaces.value = res.data.namespaces
    
    // 模拟数据
    namespaces.value = [
      { name: 'magic-prod', namespace: 'magic-prod', isHa: true, resource: 'high', status: '运行中' },
      { name: 'magic-test', namespace: 'magic-test', isHa: false, resource: 'low', status: '运行中' },
      { name: 'magic-dev', namespace: 'magic-dev', isHa: false, resource: 'low', status: '未部署' }
    ]
  } finally {
    loading.value = false
  }
}

const getStatusType = (status: string) => {
  const map: Record<string, any> = {
    '运行中': 'success',
    '部署中': 'warning',
    '未部署': 'info',
    '失败': 'danger'
  }
  return map[status] || 'info'
}

const showCreateDialog = () => {
  formData.value = {
    name: '',
    namespace: '',
    mode: 'single',
    resource: 'low',
    isHa: false
  }
  dialogVisible.value = true
}

const createNamespace = async () => {
  if (!formData.value.name || !formData.value.namespace) {
    ElMessage.warning('请填写完整信息')
    return
  }

  try {
    // 调用 API 创建
    // await axios.post('/api/v1/namespaces', formData.value)
    ElMessage.success('环境创建成功')
    dialogVisible.value = false
    fetchNamespaces()
  } catch (error) {
    ElMessage.error('创建失败')
  }
}

const deploy = (row: any) => {
  ElMessage.info(`开始部署环境：${row.name}`)
  // 跳转到部署页面
}

const access = (row: any) => {
  ElMessage.info(`获取访问地址：${row.name}`)
  // 获取并显示访问地址
}

const uninstall = (row: any) => {
  ElMessageBox.confirm(`确定要卸载环境 ${row.name} 吗？此操作不可逆！`, '警告', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    ElMessage.success('卸载任务已启动')
  })
}
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
