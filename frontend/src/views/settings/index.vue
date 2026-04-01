<template>
  <div class="settings-page">
    <h2>⚙️ 系统设置</h2>
    
    <el-card class="setting-card">
      <template #header>
        <div class="card-header">
          <span>部署配置</span>
        </div>
      </template>
      <el-form :model="settings" label-width="150px">
        <el-form-item label="默认命名空间">
          <el-select v-model="settings.defaultNamespace" placeholder="选择默认命名空间">
            <el-option label="production" value="production" />
            <el-option label="staging" value="staging" />
            <el-option label="development" value="development" />
          </el-select>
        </el-form-item>
        <el-form-item label="默认副本数">
          <el-input-number v-model="settings.defaultReplicas" :min="1" :max="10" />
        </el-form-item>
        <el-form-item label="自动扩缩容">
          <el-switch v-model="settings.autoScale" />
        </el-form-item>
        <el-form-item label="最大副本数 (当自动扩缩容启用)">
          <el-input-number v-model="settings.maxReplicas" :min="1" :max="50" :disabled="!settings.autoScale" />
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="setting-card">
      <template #header>
        <div class="card-header">
          <span>通知设置</span>
        </div>
      </template>
      <el-form :model="settings" label-width="150px">
        <el-form-item label="部署成功通知">
          <el-switch v-model="settings.notifySuccess" />
        </el-form-item>
        <el-form-item label="部署失败通知">
          <el-switch v-model="settings.notifyFailure" />
        </el-form-item>
        <el-form-item label="通知方式">
          <el-checkbox-group v-model="settings.notifyMethods">
            <el-checkbox label="email">邮件</el-checkbox>
            <el-checkbox label="slack">Slack</el-checkbox>
            <el-checkbox label="webhook">Webhook</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="setting-card">
      <template #header>
        <div class="card-header">
          <span>关于</span>
        </div>
      </template>
      <div class="about-info">
        <p><strong>版本:</strong> {{ version }}</p>
        <p><strong>构建时间:</strong> {{ buildTime }}</p>
        <p><strong>API 版本:</strong> v1</p>
      </div>
    </el-card>

    <div class="action-buttons">
      <el-button type="primary" @click="saveSettings">保存设置</el-button>
      <el-button @click="resetSettings">重置</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const version = ref('1.0.0')
const buildTime = ref('2026-04-01')

const settings = ref({
  defaultNamespace: 'production',
  defaultReplicas: 1,
  autoScale: false,
  maxReplicas: 10,
  notifySuccess: true,
  notifyFailure: true,
  notifyMethods: ['email']
})

const saveSettings = () => {
  // TODO: 保存到后端 API
  console.log('保存设置:', settings.value)
  ElMessage.success('设置已保存')
}

const resetSettings = () => {
  settings.value = {
    defaultNamespace: 'production',
    defaultReplicas: 1,
    autoScale: false,
    maxReplicas: 10,
    notifySuccess: true,
    notifyFailure: true,
    notifyMethods: ['email']
  }
  ElMessage.info('已重置为默认设置')
}

onMounted(() => {
  // TODO: 从后端加载设置
})
</script>

<style scoped>
.settings-page { padding: 20px; }
.setting-card { margin-bottom: 20px; }
.card-header { font-weight: bold; }
.about-info { line-height: 2; }
.action-buttons { text-align: right; margin-top: 20px; }
</style>
