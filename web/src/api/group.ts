import client from './client'
import type { SoraAccountGroup, CreateGroupRequest } from '../types/account'

export interface GroupWithCount extends SoraAccountGroup {
  account_count: number
}

export function listGroups() {
  return client.get<GroupWithCount[]>('/admin/groups')
}

export function createGroup(data: CreateGroupRequest) {
  return client.post<SoraAccountGroup>('/admin/groups', data)
}

export function updateGroup(id: number, data: CreateGroupRequest) {
  return client.put<SoraAccountGroup>(`/admin/groups/${id}`, data)
}

export function deleteGroup(id: number) {
  return client.delete(`/admin/groups/${id}`)
}
