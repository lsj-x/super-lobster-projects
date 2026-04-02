/**
 * API 客户端 - 连接后端 API
 * 提供类型安全的 API 调用方法
 */

import axios, { AxiosInstance, AxiosError } from 'axios'

// API 基础配置
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

// 创建 Axios 实例
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000, // 30 秒超时
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器 - 添加认证令牌等
apiClient.interceptors.request.use(
  (config) => {
    // TODO: 如果需要认证，在这里添加 token
    // const token = localStorage.getItem('token')
    // if (token) {
    //   config.headers.Authorization = `Bearer ${token}`
    // }
    console.log(`[API] ${config.method?.toUpperCase()} ${config.url}`)
    return config
  },
  (error) => {
    console.error('[API] 请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器 - 统一处理错误
apiClient.interceptors.response.use(
  (response) => {
    return response.data
  },
  (error: AxiosError) => {
    console.error('[API] 响应错误:', error.message)
    
    // 处理常见错误
    if (error.response) {
      // 服务器返回错误状态码
      const status = error.response.status
      switch (status) {
        case 401:
          console.error('未授权，请重新登录')
          break
        case 403:
          console.error('禁止访问')
          break
        case 404:
          console.error('资源未找到')
          break
        case 500:
          console.error('服务器内部错误')
          break
        default:
          console.error(`API 错误：${status}`)
      }
    } else if (error.request) {
      console.error('网络错误，请检查连接')
    }
    
    return Promise.reject(error)
  }
)

// 类型定义
export interface Deployment {
  namespace: string
  model_name: string
  version: string
  status: string
  replicas: number
  ready_replicas: number
  created_at: string
}

export interface DeploymentStatus extends Deployment {
  message?: string
}

export interface Namespace {
  name: string
  description: string
  created_at: string
}

export interface DeployRequest {
  namespace: string
  model_name: string
  version?: string
  replicas?: number
}

export interface ScaleRequest {
  replicas: number
}

/**
 * API 方法集合
 */
export const api = {
  /**
   * 健康检查
   */
  async healthCheck(): Promise<{ status: string; version: string }> {
    return apiClient.get('/health')
  },

  /**
   * 列出所有命名空间
   */
  async listNamespaces(): Promise<Namespace[]> {
    return apiClient.get('/api/namespaces')
  },

  /**
   * 部署模型
   */
  async deploy(request: DeployRequest): Promise<DeploymentStatus> {
    return apiClient.post('/api/deploy', request)
  },

  /**
   * 获取部署状态
   */
  async getDeploymentStatus(namespace: string, modelName: string): Promise<DeploymentStatus> {
    return apiClient.get(`/api/deployments/${namespace}/${modelName}/status`)
  },

  /**
   * 列出所有部署
   */
  async listDeployments(namespace?: string): Promise<Deployment[]> {
    const params = namespace ? { namespace } : undefined
    return apiClient.get('/api/deployments', { params })
  },

  /**
   * 扩容部署
   */
  async scaleDeployment(namespace: string, modelName: string, replicas: number): Promise<{ message: string }> {
    return apiClient.put(`/api/deployments/${namespace}/${modelName}/scale`, { replicas })
  },

  /**
   * 删除部署
   */
  async deleteDeployment(namespace: string, modelName: string): Promise<{ message: string }> {
    return apiClient.delete(`/api/deployments/${namespace}/${modelName}`)
  },
}

export default api
