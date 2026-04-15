/**
 * 部署管理 API 客户端
 * 提供安装、升级、回滚、卸载等操作的 API 调用
 */
import apiClient from './client'

// 类型定义
export interface InstallRequest {
  namespace: string
  model_name: string
  version?: string
  replicas?: number
  config?: Record<string, any>
}

export interface UpgradeRequest {
  namespace: string
  model_name: string
  target_version: string
  keep_config?: boolean
}

export interface RollbackRequest {
  namespace: string
  model_name: string
  target_version: string
}

export interface UninstallRequest {
  namespace: string
  model_name: string
  force?: boolean
}

export interface DeploymentHistory {
  version: string
  status: string
  created_at: string
  message?: string
}

export interface OperationResult {
  success: boolean
  message: string
  operation_id?: string
}

/**
 * 部署管理 API 方法集合
 */
export const deploymentApi = {
  /**
   * 开始安装
   * @param data 安装请求数据
   * @returns 操作结果
   */
  async startInstall(data: InstallRequest): Promise<OperationResult> {
    return apiClient.post('/api/v1/install', data)
  },

  /**
   * 开始升级
   * @param data 升级请求数据
   * @returns 操作结果
   */
  async startUpgrade(data: UpgradeRequest): Promise<OperationResult> {
    return apiClient.post('/api/v1/upgrade', data)
  },

  /**
   * 开始回滚
   * @param data 回滚请求数据
   * @returns 操作结果
   */
  async startRollback(data: RollbackRequest): Promise<OperationResult> {
    return apiClient.post('/api/v1/rollback', data)
  },

  /**
   * 开始卸载
   * @param data 卸载请求数据
   * @returns 操作结果
   */
  async startUninstall(data: UninstallRequest): Promise<OperationResult> {
    return apiClient.post('/api/v1/uninstall', data)
  },

  /**
   * 获取部署历史
   * @param namespace 命名空间
   * @param modelName 模型名称
   * @returns 部署历史记录
   */
  async getDeploymentHistory(
    namespace: string,
    modelName: string
  ): Promise<DeploymentHistory[]> {
    return apiClient.get(`/api/v1/deployments/${namespace}/${modelName}/history`)
  },

  /**
   * 获取操作状态
   * @param operationId 操作 ID
   * @returns 操作状态
   */
  async getOperationStatus(operationId: string): Promise<any> {
    return apiClient.get(`/api/v1/operations/${operationId}/status`)
  },
}

export default deploymentApi
