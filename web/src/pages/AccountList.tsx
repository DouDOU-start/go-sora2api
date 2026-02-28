import { useCallback, useEffect, useState } from 'react'
import { listAllAccounts, createAccount, updateAccount, deleteAccount, refreshAccountToken, getAccountStatus } from '../api/account'
import { listGroups } from '../api/group'
import type { SoraAccount, CreateAccountRequest, SoraAccountGroup } from '../types/account'
import GlassCard from '../components/ui/GlassCard'
import StatusBadge from '../components/ui/StatusBadge'
import LoadingState from '../components/ui/LoadingState'
import ConfirmDialog from '../components/ui/ConfirmDialog'
import FormModal from '../components/ui/FormModal'
import { toast } from '../components/ui/Toast'
import { getErrorMessage } from '../api/client'
import { motion } from 'framer-motion'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

function timeAgo(ts: string | null) {
  if (!ts) return '-'
  return formatDistanceToNow(new Date(ts), { addSuffix: true, locale: zhCN })
}

const emptyForm: CreateAccountRequest = { name: '', access_token: '', refresh_token: '', group_id: null }

const inputStyle = {
  background: 'var(--bg-inset)',
  border: '1px solid var(--border-default)',
  color: 'var(--text-primary)',
  borderRadius: 'var(--radius-md)',
}
const inputFocus = (e: React.FocusEvent<HTMLInputElement | HTMLSelectElement>) => {
  e.target.style.borderColor = 'var(--accent)'
  e.target.style.boxShadow = '0 0 0 3px var(--accent-soft)'
}
const inputBlur = (e: React.FocusEvent<HTMLInputElement | HTMLSelectElement>) => {
  e.target.style.borderColor = 'var(--border-default)'
  e.target.style.boxShadow = 'none'
}

export default function AccountList() {
  const [accounts, setAccounts] = useState<SoraAccount[]>([])
  const [groups, setGroups] = useState<SoraAccountGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [form, setForm] = useState<CreateAccountRequest>({ ...emptyForm })
  const [submitting, setSubmitting] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)
  const [actionLoading, setActionLoading] = useState<Record<string, boolean>>({})
  const [confirmState, setConfirmState] = useState<{ open: boolean; id: number }>({ open: false, id: 0 })

  const reload = useCallback(() => setRefreshKey((k) => k + 1), [])

  const closeForm = () => {
    setShowForm(false)
    setEditId(null)
    setForm({ ...emptyForm })
  }

  useEffect(() => {
    const load = async () => {
      try {
        const [aRes, gRes] = await Promise.all([listAllAccounts(), listGroups()])
        setAccounts(aRes.data ?? [])
        setGroups(gRes.data ?? [])
      } catch { /* ignore */ }
      setLoading(false)
    }
    load()
  }, [refreshKey])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      if (editId) {
        await updateAccount(editId, form)
        toast.success('账号已更新')
      } else {
        await createAccount(form)
        toast.success('账号已添加')
      }
      closeForm()
      reload()
    } catch (err) {
      toast.error(getErrorMessage(err, editId ? '更新失败' : '添加失败'))
    }
    setSubmitting(false)
  }

  const handleEdit = (acc: SoraAccount) => {
    setEditId(acc.id)
    setForm({ name: acc.name, access_token: '', refresh_token: '', group_id: acc.group_id })
    setShowForm(true)
  }

  const handleDelete = (id: number) => {
    setConfirmState({ open: true, id })
  }

  const confirmDelete = async () => {
    const id = confirmState.id
    setConfirmState({ open: false, id: 0 })
    try {
      await deleteAccount(id)
      toast.success('账号已删除')
      reload()
    } catch (err) {
      toast.error(getErrorMessage(err, '删除失败'))
    }
  }

  const handleRefresh = async (id: number) => {
    setActionLoading(prev => ({ ...prev, [`refresh-${id}`]: true }))
    try {
      await refreshAccountToken(id)
      toast.success('Token 续期成功')
      reload()
    } catch (err) {
      toast.error(getErrorMessage(err, 'Token 续期失败'))
    }
    setActionLoading(prev => ({ ...prev, [`refresh-${id}`]: false }))
  }

  const handleSync = async (id: number) => {
    setActionLoading(prev => ({ ...prev, [`sync-${id}`]: true }))
    try {
      await getAccountStatus(id)
      toast.success('刷新成功')
      reload()
    } catch (err) {
      toast.error(getErrorMessage(err, '刷新失败'))
    }
    setActionLoading(prev => ({ ...prev, [`sync-${id}`]: false }))
  }

  if (loading) return <LoadingState />

  return (
    <div>
      {/* 页头 */}
      <motion.div
        className="flex items-center justify-between mb-6"
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <div>
          <h1 className="text-2xl font-semibold tracking-tight" style={{ color: 'var(--text-primary)' }}>
            账号管理
          </h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--text-tertiary)' }}>
            共 {accounts.length} 个账号
          </p>
        </div>
        <button
          onClick={() => { setEditId(null); setForm({ ...emptyForm }); setShowForm(true) }}
          className="px-4 py-2 rounded-xl text-sm font-medium text-white transition-all cursor-pointer"
          style={{ background: 'var(--accent)' }}
          onMouseEnter={(e) => e.currentTarget.style.background = 'var(--accent-hover)'}
          onMouseLeave={(e) => e.currentTarget.style.background = 'var(--accent)'}
        >
          + 添加账号
        </button>
      </motion.div>

      {/* 账号列表 */}
      {accounts.length === 0 ? (
        <div className="text-center py-20" style={{ color: 'var(--text-tertiary)' }}>
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" className="mx-auto mb-3 opacity-40">
            <circle cx="12" cy="8" r="4" /><path d="M6 21v-1a6 6 0 0112 0v1" />
          </svg>
          暂无账号，点击上方按钮添加
        </div>
      ) : (
        <div className="space-y-3">
          {accounts.map((acc, i) => (
            <GlassCard key={acc.id} delay={i} className="overflow-hidden">
              <div className="p-4 sm:p-5">
                {/* 第一行：标识 + 状态标签 */}
                <div className="flex items-center gap-2 flex-wrap mb-1">
                  <span className="font-semibold text-[15px] truncate" style={{ color: 'var(--text-primary)' }}>
                    {acc.name || acc.email || acc.at_hint || `#${acc.id}`}
                  </span>
                  <StatusBadge status={acc.status} />
                  {acc.rate_limit_reached && (
                    <span
                      className="text-[11px] px-2 py-0.5 rounded-full font-medium"
                      style={{ background: 'var(--warning-soft)', color: 'var(--warning)' }}
                    >
                      限流中
                    </span>
                  )}
                  {acc.group_name && (
                    <span
                      className="text-[11px] px-2 py-0.5 rounded-full font-medium"
                      style={{ background: 'var(--info-soft)', color: 'var(--info)' }}
                    >
                      {acc.group_name}
                    </span>
                  )}
                </div>

                {/* 第二行：副标题 */}
                {acc.name && acc.email && (
                  <p className="text-[12px] mb-2" style={{ color: 'var(--text-tertiary)' }}>{acc.email}</p>
                )}
                {!acc.name && acc.email && (
                  <p className="text-[12px] mb-2" style={{ color: 'var(--text-tertiary)' }}>AT: {acc.at_hint || '-'}</p>
                )}
                {!acc.name && !acc.email && acc.at_hint && (
                  <p className="text-[12px] mb-2" style={{ color: 'var(--text-tertiary)' }}>AT: {acc.at_hint}</p>
                )}

                {/* 信息指标行 */}
                <div
                  className="flex items-center flex-wrap gap-x-5 gap-y-1 text-[13px] py-2 mb-2"
                  style={{ borderTop: '1px solid var(--border-default)' }}
                >
                  <InfoItem label="套餐" value={acc.plan_title || '-'} />
                  <InfoItem
                    label="额度"
                    value={acc.remaining_count === -1 ? '未知' : String(acc.remaining_count)}
                    color={acc.remaining_count === 0 ? 'var(--danger)' : undefined}
                    bold
                  />
                  <InfoItem label="最后使用" value={timeAgo(acc.last_used_at)} />
                </div>

                {/* 错误信息 */}
                {acc.last_error && (
                  <div
                    className="mb-2 px-3 py-2 rounded-lg text-xs truncate"
                    style={{ background: 'var(--danger-soft)', color: 'var(--danger)' }}
                  >
                    {acc.last_error}
                  </div>
                )}

                {/* 操作按钮行 */}
                <div
                  className="flex items-center gap-1 pt-2 flex-wrap"
                  style={{ borderTop: '1px solid var(--border-default)' }}
                >
                  {acc.rt_hint && (
                    <ActionBtn
                      label="续期 Token"
                      title="使用 Refresh Token 获取新的 Access Token"
                      loading={actionLoading[`refresh-${acc.id}`]}
                      onClick={() => handleRefresh(acc.id)}
                    />
                  )}
                  <ActionBtn
                    label="刷新"
                    title="同步配额和订阅信息"
                    loading={actionLoading[`sync-${acc.id}`]}
                    onClick={() => handleSync(acc.id)}
                  />
                  <ActionBtn label="编辑" onClick={() => handleEdit(acc)} />
                  <div className="flex-1" />
                  <ActionBtn label="删除" danger onClick={() => handleDelete(acc.id)} />
                </div>
              </div>
            </GlassCard>
          ))}
        </div>
      )}

      {/* 添加/编辑弹窗 */}
      <FormModal
        open={showForm}
        title={editId ? '编辑账号' : '添加账号'}
        onClose={closeForm}
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>
                备注名称 <span style={{ color: 'var(--text-tertiary)', fontWeight: 400 }}>（可选）</span>
              </label>
              <input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                placeholder="留空则通过 Token 自动识别"
                className="w-full px-3 py-2.5 text-sm outline-none transition-all"
                style={inputStyle}
                onFocus={inputFocus}
                onBlur={inputBlur}
              />
            </div>
            <div>
              <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>
                分组 <span style={{ color: 'var(--text-tertiary)', fontWeight: 400 }}>（可选）</span>
              </label>
              <select
                value={form.group_id ?? ''}
                onChange={(e) => setForm({ ...form, group_id: e.target.value ? Number(e.target.value) : null })}
                className="w-full px-3 py-2.5 text-sm outline-none transition-all"
                style={inputStyle}
                onFocus={inputFocus}
                onBlur={inputBlur}
              >
                <option value="">未分组</option>
                {groups.map(g => (
                  <option key={g.id} value={g.id}>{g.name}</option>
                ))}
              </select>
            </div>
          </div>
          <div>
            <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>Access Token</label>
            <input
              value={form.access_token}
              onChange={(e) => setForm({ ...form, access_token: e.target.value })}
              placeholder={editId ? '留空则不修改' : 'eyJhbGci...'}
              className="w-full px-3 py-2.5 text-sm outline-none transition-all"
              style={{ ...inputStyle, fontFamily: 'var(--font-mono)' }}
              onFocus={inputFocus}
              onBlur={inputBlur}
            />
          </div>
          <div>
            <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>
              Refresh Token <span style={{ color: 'var(--text-tertiary)', fontWeight: 400 }}>（可选）</span>
            </label>
            <input
              value={form.refresh_token}
              onChange={(e) => setForm({ ...form, refresh_token: e.target.value })}
              placeholder={editId ? '留空则不修改' : 'v1.rt-...'}
              className="w-full px-3 py-2.5 text-sm outline-none transition-all"
              style={{ ...inputStyle, fontFamily: 'var(--font-mono)' }}
              onFocus={inputFocus}
              onBlur={inputBlur}
            />
          </div>
          <div className="flex justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={closeForm}
              className="px-4 py-2 rounded-xl text-sm font-medium transition-colors cursor-pointer"
              style={{ color: 'var(--text-secondary)', background: 'var(--bg-inset)' }}
            >
              取消
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="px-5 py-2 rounded-xl text-sm font-medium text-white disabled:opacity-50 transition-all cursor-pointer"
              style={{ background: 'var(--accent)' }}
              onMouseEnter={(e) => e.currentTarget.style.background = 'var(--accent-hover)'}
              onMouseLeave={(e) => e.currentTarget.style.background = 'var(--accent)'}
            >
              {submitting ? '保存中...' : editId ? '更新' : '添加'}
            </button>
          </div>
        </form>
      </FormModal>

      {/* 删除确认对话框 */}
      <ConfirmDialog
        open={confirmState.open}
        title="删除账号"
        message="确定删除此账号？此操作不可恢复。"
        confirmLabel="删除"
        danger
        onConfirm={confirmDelete}
        onCancel={() => setConfirmState({ open: false, id: 0 })}
      />
    </div>
  )
}

function InfoItem({ label, value, color, bold }: { label: string; value: string; color?: string; bold?: boolean }) {
  return (
    <div style={{ color: 'var(--text-tertiary)' }}>
      {label}{' '}
      <span style={{ color: color || 'var(--text-secondary)', fontWeight: bold ? 500 : 400 }}>{value}</span>
    </div>
  )
}

function ActionBtn({ label, title, onClick, danger, loading: isLoading }: {
  label: string
  title?: string
  onClick: () => void
  danger?: boolean
  loading?: boolean
}) {
  return (
    <button
      onClick={onClick}
      disabled={isLoading}
      title={title}
      className="px-2.5 py-1 text-[12px] font-medium rounded-lg transition-colors cursor-pointer disabled:opacity-50"
      style={{
        color: danger ? 'var(--danger)' : 'var(--text-tertiary)',
        background: 'transparent',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.background = danger ? 'var(--danger-soft)' : 'var(--bg-inset)'
        e.currentTarget.style.color = danger ? 'var(--danger)' : 'var(--text-secondary)'
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.background = 'transparent'
        e.currentTarget.style.color = danger ? 'var(--danger)' : 'var(--text-tertiary)'
      }}
    >
      {isLoading ? '...' : label}
    </button>
  )
}
