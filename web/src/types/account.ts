export interface SoraAccountGroup {
  id: number
  name: string
  description: string
  enabled: boolean
  account_count: number
  created_at: string
  updated_at: string
}

export interface SoraAccount {
  id: number
  group_id: number | null
  group_name: string
  name: string
  at_hint: string
  rt_hint: string
  token_expires_at: string | null
  plan_title: string
  plan_expires_at: string | null
  remaining_count: number
  rate_limit_reached: boolean
  rate_limit_resets_at: string | null
  enabled: boolean
  status: 'active' | 'token_expired' | 'quota_exhausted'
  last_used_at: string | null
  last_error: string
  last_sync_at: string | null
  created_at: string
  updated_at: string
}

export interface CreateAccountRequest {
  name: string
  access_token?: string
  refresh_token?: string
  group_id?: number | null
  enabled?: boolean
}

export interface CreateGroupRequest {
  name: string
  description?: string
  enabled?: boolean
}
