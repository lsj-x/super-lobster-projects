/**
 * 命名空间 API 客户端
 * 提供命名空间相关的 API 调用方法
 */
import apiClient from './client'

// 类型定义
export interface Namespace {
  name: string
  description?: string
  version?: string
  status: string
  resource_limits?: {
    cpu?: string
    memory?: string
    gpu?: string
  }
  created_at: string
  updated_at?: string
}

export interface CreateNamespaceRequest {
  name: string
  description?: string
  version?: string
  resource_limits?: {
    cpu?: string
    memory?: string
    gpu?: string
  }
}

export interface UpdateNamespaceRequest {
  description?: string
  version?: string
  resource_limits?: {
    cpu?: string
    memory?: string
    gpu?: string
  }
}

export interface NamespaceListResponse {
  namespaces: Namespace[]
  total: number
}

/**
 * 命名空间 API 方法集合
 */
export const namespacesApi = {
  /**
   * 列出所有命名空间
   * GET /api/v1/namespaces
   */
  async listNamespaces(): Promise<NamespaceListResponse> {
    return apiClient.get('/api/v1/namespaces')
  },

  /**
   * 获取单个命名空间详情
   * GET /api/v1/namespaces/:name
   */
  async getNamespace(name: string): Promise<Namespace> {
    return apiClient.get(`/api/v1/namespaces/${name}`)
  },

  /**
   * 创建命名空间
   * POST /api/v1/namespaces
   */
  async createNamespace(data: CreateNamespaceRequest): Promise<Namespace> {
    return apiClient.post('/api/v1/namespaces', data)
  },

  /**
   * 更新命名空间
   * PUT /api/v1/namespaces/:name
   */
  async updateNamespace(name: string, data: UpdateNamespaceRequest): Promise<Namespace> {
    return apiClient.put(`/api/v1/namespaces/${name}`, data)
  },

  /**
   * 删除命名空间
   * DELETE /api/v1/namespaces/:name
   */
  async deleteNamespace(name: string): Promise<{ message: string }> {
    return apiClient.delete(`/api/v1/namespaces/${name}`)
  },
}

export default namespacesApi
