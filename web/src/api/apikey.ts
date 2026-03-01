import client from './client'
import type { SoraAPIKey, CreateAPIKeyRequest } from '../types/account'
import type { PageResponse } from '../types/api'

export function listAPIKeys(params?: { page?: number; page_size?: number; keyword?: string; enabled?: boolean; group_id?: number | 'null' }) {
  return client.get<PageResponse<SoraAPIKey>>('/admin/api-keys', { params })
}

export function createAPIKey(data: CreateAPIKeyRequest) {
  return client.post<SoraAPIKey>('/admin/api-keys', data)
}

export function updateAPIKey(id: number, data: CreateAPIKeyRequest) {
  return client.put<SoraAPIKey>(`/admin/api-keys/${id}`, data)
}

export function deleteAPIKey(id: number) {
  return client.delete(`/admin/api-keys/${id}`)
}

export function revealAPIKey(id: number) {
  return client.get<{ key: string }>(`/admin/api-keys/${id}/reveal`)
}
