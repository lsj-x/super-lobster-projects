/**
 * Secret 管理 API 客户端
 * 提供 Secret 的增删改查操作
 */
import apiClient from './client'

// 类型定义
export interface Secret {
  name: string
  description?: string
  data: Record<string, string>
  encrypted: boolean
  created_at: string
  updated_at: string
}

export interface SecretData {
  name: string
  description?: string
  data: Record<string, string>
}

export interface SecretListItem {
  name: string
  description?: string
  encrypted: boolean
  created_at: string
  updated_at: string
}

/**
 * Secret API 方法集合
 */
export const secretsApi = {
  /**
   * 列出所有 Secret
   * GET /api/v1/secrets
   */
  async listSecrets(): Promise<SecretListItem[]> {
    return apiClient.get('/api/v1/secrets')
  },

  /**
   * 获取单个 Secret
   * GET /api/v1/secrets/:name
   */
  async getSecret(name: string): Promise<Secret> {
    return apiClient.get(`/api/v1/secrets/${encodeURIComponent(name)}`)
  },

  /**
   * 创建 Secret
   * POST /api/v1/secrets
   */
  async createSecret(data: SecretData): Promise<Secret> {
    return apiClient.post('/api/v1/secrets', data)
  },

  /**
   * 更新 Secret
   * PUT /api/v1/secrets/:name
   */
  async updateSecret(name: string, data: SecretData): Promise<Secret> {
    return apiClient.put(`/api/v1/secrets/${encodeURIComponent(name)}`, data)
  },

  /**
   * 删除 Secret
   * DELETE /api/v1/secrets/:name
   */
  async deleteSecret(name: string): Promise<{ message: string }> {
    return apiClient.delete(`/api/v1/secrets/${encodeURIComponent(name)}`)
  },

  /**
   * 解密 Secret
   * POST /api/v1/secrets/:name/decrypt
   */
  async decryptSecret(name: string): Promise<Record<string, string>> {
    return apiClient.post(`/api/v1/secrets/${encodeURIComponent(name)}/decrypt`)
  },
}

export default secretsApi
