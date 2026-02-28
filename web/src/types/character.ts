export type CharacterStatus = 'processing' | 'ready' | 'failed'

export interface SoraCharacter {
  id: string
  account_id: number
  cameo_id: string
  character_id: string
  status: CharacterStatus
  display_name: string
  username: string
  profile_url: string
  error_message: string
  account_email: string
  created_at: string
  updated_at: string
  completed_at: string | null
}
