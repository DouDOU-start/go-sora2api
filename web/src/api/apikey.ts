import client from './client'
import type { SoraAPIKey, CreateAPIKeyRequest } from '../types/account'

export function listAPIKeys() {
  return client.get<SoraAPIKey[]>('/admin/api-keys')
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
