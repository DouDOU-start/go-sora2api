import client from './client'
import type { SoraAccount, CreateAccountRequest } from '../types/account'

export function listAllAccounts() {
  return client.get<SoraAccount[]>('/admin/accounts')
}

export function createAccount(data: CreateAccountRequest) {
  return client.post<SoraAccount>('/admin/accounts', data)
}

export function updateAccount(accountId: number, data: CreateAccountRequest) {
  return client.put<SoraAccount>(`/admin/accounts/${accountId}`, data)
}

export function deleteAccount(accountId: number) {
  return client.delete(`/admin/accounts/${accountId}`)
}

export function refreshAccountToken(accountId: number) {
  return client.post<SoraAccount>(`/admin/accounts/${accountId}/refresh`)
}

export function getAccountStatus(accountId: number) {
  return client.get<SoraAccount>(`/admin/accounts/${accountId}/status`)
}

export function revealAccountTokens(accountId: number) {
  return client.get<{ access_token: string; refresh_token: string }>(`/admin/accounts/${accountId}/tokens`)
}
