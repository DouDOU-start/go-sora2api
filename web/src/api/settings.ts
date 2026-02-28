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
