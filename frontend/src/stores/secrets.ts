/**
 * Secret 管理 Store
 * 使用 Pinia 管理 Secret 状态
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { secretsApi, Secret, SecretListItem, SecretData } from '@/api/secrets'
import { ElMessage, ElMessageBox } from 'element-plus'

export interface SecretState {
  list: SecretListItem[]
  currentSecret: Secret | null
  loading: boolean
  detailLoading: boolean
}

export const useSecretsStore = defineStore('secrets', () => {
  // State
  const list = ref<SecretListItem[]>([])
  const currentSecret = ref<Secret | null>(null)
  const loading = ref(false)
  const detailLoading = ref(false)

  // Getters
  const secretCount = computed(() => list.value.length)

  // Actions
  /**
   * 加载 Secret 列表
   */
  async fetchList() {
    loading.value = true
    try {
      list.value = await secretsApi.listSecrets()
    } catch (error) {
      ElMessage.error('加载 Secret 列表失败')
      console.error('Failed to fetch secrets:', error)
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取单个 Secret
   */
  async fetchSecret(name: string) {
    detailLoading.value = true
    try {
      currentSecret.value = await secretsApi.getSecret(name)
    } catch (error) {
      ElMessage.error('获取 Secret 详情失败')
      console.error('Failed to fetch secret:', error)
      throw error
    } finally {
      detailLoading.value = false
    }
  }

  /**
   * 创建 Secret
   */
  async createSecret(data: SecretData) {
    try {
      await secretsApi.createSecret(data)
      ElMessage.success('Secret 创建成功')
      await fetchList()
    } catch (error) {
      ElMessage.error('创建 Secret 失败')
      console.error('Failed to create secret:', error)
      throw error
    }
  }

  /**
   * 更新 Secret
   */
  async updateSecret(name: string, data: SecretData) {
    try {
      await secretsApi.updateSecret(name, data)
      ElMessage.success('Secret 更新成功')
      await fetchList()
      // 如果当前查看的是这个 Secret，更新当前 Secret
      if (currentSecret.value?.name === name) {
        await fetchSecret(name)
      }
    } catch (error) {
      ElMessage.error('更新 Secret 失败')
      console.error('Failed to update secret:', error)
      throw error
    }
  }

  /**
   * 删除 Secret
   */
  async deleteSecret(name: string) {
    try {
      await ElMessageBox.confirm(`确定要删除 Secret "${name}" 吗？此操作不可逆！`, '警告', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      })

      await secretsApi.deleteSecret(name)
      ElMessage.success('Secret 删除成功')

      // 如果删除的是当前 Secret，清空
      if (currentSecret.value?.name === name) {
        currentSecret.value = null
      }

      await fetchList()
    } catch (error) {
      if (error !== 'cancel') {
        ElMessage.error('删除 Secret 失败')
        console.error('Failed to delete secret:', error)
      }
    }
  }

  /**
   * 解密 Secret
   */
  async decryptSecret(name: string) {
    try {
      return await secretsApi.decryptSecret(name)
    } catch (error) {
      ElMessage.error('解密失败')
      console.error('Failed to decrypt secret:', error)
      throw error
    }
  }

  /**
   * 清空当前 Secret
   */
  clearCurrentSecret() {
    currentSecret.value = null
  }

  return {
    // State
    list,
    currentSecret,
    loading,
    detailLoading,
    // Getters
    secretCount,
    // Actions
    fetchList,
    fetchSecret,
    createSecret,
    updateSecret,
    deleteSecret,
    decryptSecret,
    clearCurrentSecret,
  }
})
