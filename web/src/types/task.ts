export type TaskStatus = 'queued' | 'in_progress' | 'completed' | 'failed'

export interface SoraTask {
  id: string
  sora_task_id: string
  account_id: number
  type: string
  model: string
  prompt: string
  status: TaskStatus
  progress: number
  error_message: string
  image_url: string
  created_at: string
  updated_at: string
  completed_at: string | null
}

export interface DashboardStats {
  total_accounts: number
  active_accounts: number
  expired_accounts: number
  exhausted_accounts: number
  total_tasks: number
  pending_tasks: number
  completed_tasks: number
  failed_tasks: number
}
