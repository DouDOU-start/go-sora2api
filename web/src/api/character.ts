import client from './client'
import type { SoraCharacter } from '../types/character'
import type { PageResponse } from '../types/api'

export function listCharacters(params: { status?: string; is_public?: boolean; page?: number; page_size?: number }) {
  return client.get<PageResponse<SoraCharacter>>('/admin/characters', { params })
}

export function getCharacter(id: string) {
  return client.get<SoraCharacter>(`/admin/characters/${id}`)
}

export function deleteCharacter(id: string) {
  return client.delete(`/admin/characters/${id}`)
}

export function toggleCharacterVisibility(id: string) {
  return client.post<{ message: string; is_public: boolean }>(`/admin/characters/${id}/visibility`)
}

// 构造角色头像图片 URL（通过内部接口加载，附带 JWT 认证）
export function getCharacterImageUrl(id: string): string {
  const token = localStorage.getItem('token')
  return `/admin/characters/${id}/image?token=${encodeURIComponent(token || '')}`
}
