import client from './client'
import type { SoraTask, DashboardStats } from '../types/task'
import type { PageResponse } from '../types/api'

export function getDashboard() {
  return client.get<DashboardStats>('/admin/dashboard')
}

export function listTasks(params: { status?: string; type?: string; page?: number; page_size?: number }) {
  return client.get<PageResponse<SoraTask>>('/admin/tasks', { params })
}

export function getTask(id: string) {
  return client.get<SoraTask>(`/admin/tasks/${id}`)
}

export function downloadTaskContent(id: string) {
  return client.get<Blob>(`/admin/tasks/${id}/content`, { responseType: 'blob', timeout: 120000 })
}
