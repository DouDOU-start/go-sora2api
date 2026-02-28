import client from './client'
import type { SoraCharacter } from '../types/character'
import type { PageResponse } from '../types/api'

export function listCharacters(params: { status?: string; page?: number; page_size?: number }) {
  return client.get<PageResponse<SoraCharacter>>('/admin/characters', { params })
}

export function getCharacter(id: string) {
  return client.get<SoraCharacter>(`/admin/characters/${id}`)
}

export function deleteCharacter(id: string) {
  return client.delete(`/admin/characters/${id}`)
}
