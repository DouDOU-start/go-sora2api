import client from './client'

export interface SystemSettings {
  proxy_url: string
  token_refresh_interval: string
  credit_sync_interval: string
  subscription_sync_interval: string
}

export const getSettings = () => client.get<SystemSettings>('/admin/settings')

export const updateSettings = (data: Partial<Record<string, string>>) =>
  client.put<SystemSettings>('/admin/settings', data)

export interface ProxyTestResult {
  success: boolean
  status_code?: number
  latency?: number
  error?: string
}

export const testProxy = (proxyUrl: string) =>
  client.post<ProxyTestResult>('/admin/proxy-test', { proxy_url: proxyUrl })

export interface VersionInfo {
  current: string
  latest: string
  has_update: boolean
}

export const getVersion = () => client.get<VersionInfo>('/admin/version')

export const triggerUpgrade = () => client.post<{ message?: string; error?: string }>('/admin/upgrade')
