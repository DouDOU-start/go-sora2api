import { useCallback, useEffect, useRef, useState } from 'react'
import { listAccounts, createAccount, updateAccount, deleteAccount, refreshAccountToken, getAccountStatus, revealAccountTokens, batchImportAccounts } from '../api/account'
import { listGroups } from '../api/group'
import type { SoraAccount, CreateAccountRequest, SoraAccountGroup, BatchImportResult } from '../types/account'
import GlassCard from '../components/ui/GlassCard'
import StatusBadge from '../components/ui/StatusBadge'
import LoadingState from '../components/ui/LoadingState'
import ConfirmDialog from '../components/ui/ConfirmDialog'
import FormModal from '../components/ui/FormModal'
import { toast } from '../components/ui/toastStore'
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
const inputFocus = (e: React.FocusEvent<HTMLElement>) => {
  (e.target as HTMLElement).style.borderColor = 'var(--accent)';
  (e.target as HTMLElement).style.boxShadow = '0 0 0 3px var(--accent-soft)'
}
const inputBlur = (e: React.FocusEvent<HTMLElement>) => {
  (e.target as HTMLElement).style.borderColor = 'var(--border-default)';
  (e.target as HTMLElement).style.boxShadow = 'none'
}

const statusFilters = [
  { label: '全部', value: '' },
  { label: '正常', value: 'active' },
  { label: 'Token 过期', value: 'token_expired' },
  { label: '额度耗尽', value: 'quota_exhausted' },
]

const PAGE_SIZE = 20

export default function AccountList() {
  const [accounts, setAccounts] = useState<SoraAccount[]>([])
  const [total, setTotal] = useState(0)
  const [groups, setGroups] = useState<SoraAccountGroup[]>([])
  const [loading, setLoading] = useState(true)

  // 筛选 & 分页
  const [status, setStatus] = useState('')
  const [groupId, setGroupId] = useState<string>('')
  const [keyword, setKeyword] = useState('')
  const [inputKeyword, setInputKeyword] = useState('')
  const [page, setPage] = useState(1)

  // 表单
  const [showForm, setShowForm] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [form, setForm] = useState<CreateAccountRequest>({ ...emptyForm })
  const [submitting, setSubmitting] = useState(false)

  const [actionLoading, setActionLoading] = useState<Record<string, boolean>>({})
  const [confirmState, setConfirmState] = useState<{ open: boolean; id: number }>({ open: false, id: 0 })
  const [revealedTokens, setRevealedTokens] = useState<Record<number, { access_token: string; refresh_token: string }>>({})

  // 批量导入状态
  const [showBatch, setShowBatch] = useState(false)
  const [batchTokens, setBatchTokens] = useState('')
  const [batchGroupId, setBatchGroupId] = useState<number | null>(null)
  const [batchImporting, setBatchImporting] = useState(false)
  const [batchResult, setBatchResult] = useState<BatchImportResult | null>(null)

  const mountedRef = useRef(true)

  const fetchData = useCallback(async () => {
    const [aRes, gRes] = await Promise.all([
      listAccounts({ page, page_size: PAGE_SIZE, status: status || undefined, group_id: groupId ? Number(groupId) : undefined, keyword: keyword || undefined }),
      listGroups(),
    ])
    return {
      accounts: aRes.data.list ?? [],
      total: aRes.data.total,
      groups: gRes.data ?? [],
    }
  }, [page, status, groupId, keyword])

  useEffect(() => {
    mountedRef.current = true
    let canceled = false
    void (async () => {
      try {
        const data = await fetchData()
        if (!canceled && mountedRef.current) {
          setAccounts(data.accounts)
          setTotal(data.total)
          setGroups(data.groups)
        }
      } catch {
        // ignore
      } finally {
        if (!canceled && mountedRef.current) setLoading(false)
      }
    })()
    return () => { canceled = true; mountedRef.current = false }
  }, [fetchData])

  const reload = useCallback(async () => {
    setLoading(true)
    try {
      const data = await fetchData()
      if (mountedRef.current) {
        setAccounts(data.accounts)
        setTotal(data.total)
        setGroups(data.groups)
      }
    } catch {
      // ignore
    } finally {
      if (mountedRef.current) setLoading(false)
    }
  }, [fetchData])

  const closeForm = () => {
    setShowForm(false)
    setEditId(null)
    setForm({ ...emptyForm })
  }

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

  const handleRevealTokens = async (id: number) => {
    try {
      const res = await revealAccountTokens(id)
      setRevealedTokens((prev) => ({ ...prev, [id]: res.data }))
    } catch {
      toast.error('获取 Token 失败')
    }
  }

  const handleHideTokens = (id: number) => {
    setRevealedTokens((prev) => {
      const next = { ...prev }
      delete next[id]
      return next
    })
  }

  const copyText = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text)
    } catch {
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    toast.success(`${label} 已复制`)
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

  const handleBatchImport = async () => {
    const tokens = batchTokens.split('\n').map(t => t.trim()).filter(Boolean)
    if (tokens.length === 0) {
      toast.error('请输入至少一个 Token')
      return
    }
    setBatchImporting(true)
    try {
      const res = await batchImportAccounts({ tokens, group_id: batchGroupId })
      setBatchResult(res.data)
      reload()
    } catch (err) {
      toast.error(getErrorMessage(err, '批量导入失败'))
    }
    setBatchImporting(false)
  }

  const closeBatch = () => {
    setShowBatch(false)
    setBatchTokens('')
    setBatchGroupId(null)
    setBatchResult(null)
  }

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setKeyword(inputKeyword)
    setPage(1)
  }

  const totalPages = Math.ceil(total / PAGE_SIZE)

  if (loading) return <LoadingState />

  return (
    <div>
      {/* 页头 */}
      <motion.div
        className="flex items-center justify-between mb-4"
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <div>
          <h1 className="text-2xl font-semibold tracking-tight" style={{ color: 'var(--text-primary)' }}>
            账号管理
          </h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--text-tertiary)' }}>
            共 {total} 个账号
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => { setBatchResult(null); setShowBatch(true) }}
            className="px-4 py-2 rounded-xl text-sm font-medium transition-all cursor-pointer"
            style={{ background: 'var(--bg-elevated)', color: 'var(--text-secondary)', border: '1px solid var(--border-default)' }}
            onMouseEnter={(e) => e.currentTarget.style.background = 'var(--bg-inset)'}
            onMouseLeave={(e) => e.currentTarget.style.background = 'var(--bg-elevated)'}
          >
            批量导入
          </button>
          <button
            onClick={() => { setEditId(null); setForm({ ...emptyForm }); setShowForm(true) }}
            className="px-4 py-2 rounded-xl text-sm font-medium text-white transition-all cursor-pointer"
            style={{ background: 'var(--accent)' }}
            onMouseEnter={(e) => e.currentTarget.style.background = 'var(--accent-hover)'}
            onMouseLeave={(e) => e.currentTarget.style.background = 'var(--accent)'}
          >
            + 添加账号
          </button>
        </div>
      </motion.div>

      {/* 筛选栏 */}
      <div className="flex flex-col sm:flex-row gap-3 mb-5 flex-wrap">
        <div
          className="flex items-center gap-0.5 p-1 rounded-xl"
          style={{ background: 'var(--bg-inset)' }}
        >
          {statusFilters.map((f) => (
            <button
              key={f.value}
              onClick={() => { setStatus(f.value); setPage(1) }}
              className="px-3 py-1.5 rounded-lg text-[13px] font-medium transition-all cursor-pointer"
              style={{
                background: status === f.value ? 'var(--bg-surface)' : 'transparent',
                color: status === f.value ? 'var(--text-primary)' : 'var(--text-tertiary)',
                boxShadow: status === f.value ? 'var(--shadow-sm)' : 'none',
              }}
            >
              {f.label}
            </button>
          ))}
        </div>
        <select
          value={groupId}
          onChange={(e) => { setGroupId(e.target.value); setPage(1) }}
          className="px-3 py-1.5 rounded-xl text-sm outline-none transition-all cursor-pointer"
          style={{ ...inputStyle, minWidth: 120 }}
          onFocus={inputFocus}
          onBlur={inputBlur}
        >
          <option value="">全部分组</option>
          {groups.map(g => (
            <option key={g.id} value={g.id}>{g.name}</option>
          ))}
        </select>
        <form onSubmit={handleSearch} className="flex items-center gap-2 flex-1">
          <input
            value={inputKeyword}
            onChange={(e) => setInputKeyword(e.target.value)}
            placeholder="搜索邮箱或备注名"
            className="flex-1 px-3 py-1.5 text-sm outline-none transition-all"
            style={{ ...inputStyle, minWidth: 160 }}
            onFocus={inputFocus}
            onBlur={inputBlur}
          />
          <button
            type="submit"
            className="px-3 py-1.5 rounded-xl text-sm font-medium transition-all cursor-pointer"
            style={{ background: 'var(--accent)', color: '#fff' }}
            onMouseEnter={(e) => e.currentTarget.style.background = 'var(--accent-hover)'}
            onMouseLeave={(e) => e.currentTarget.style.background = 'var(--accent)'}
          >
            搜索
          </button>
          {(keyword || status || groupId) && (
            <button
              type="button"
              onClick={() => { setInputKeyword(''); setKeyword(''); setStatus(''); setGroupId(''); setPage(1) }}
              className="px-3 py-1.5 rounded-xl text-sm font-medium transition-all cursor-pointer"
              style={{ background: 'var(--bg-inset)', color: 'var(--text-tertiary)' }}
            >
              清除
            </button>
          )}
        </form>
      </div>

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

                {/* Token 详情 */}
                {revealedTokens[acc.id] && (
                  <div
                    className="mb-2 px-3 py-2.5 rounded-lg space-y-2"
                    style={{ background: 'var(--bg-inset)', border: '1px solid var(--border-default)' }}
                  >
                    <TokenRow
                      label="AT"
                      value={revealedTokens[acc.id].access_token}
                      onCopy={() => copyText(revealedTokens[acc.id].access_token, 'Access Token')}
                    />
                    <TokenRow
                      label="RT"
                      value={revealedTokens[acc.id].refresh_token}
                      onCopy={() => copyText(revealedTokens[acc.id].refresh_token, 'Refresh Token')}
                    />
                    <div className="flex justify-end">
                      <button
                        onClick={() => handleHideTokens(acc.id)}
                        className="text-[11px] px-2 py-0.5 rounded transition-colors cursor-pointer"
                        style={{ color: 'var(--text-tertiary)' }}
                        onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-elevated)' }}
                        onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
                      >
                        收起
                      </button>
                    </div>
                  </div>
                )}

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
                  <ActionBtn
                    label={revealedTokens[acc.id] ? '隐藏 Token' : '查看 Token'}
                    onClick={() => revealedTokens[acc.id] ? handleHideTokens(acc.id) : handleRevealTokens(acc.id)}
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

      {/* 分页 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-3 mt-8">
          <button
            onClick={() => setPage(Math.max(1, page - 1))}
            disabled={page === 1}
            className="px-4 py-2 rounded-xl text-sm font-medium transition-all cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
            style={{ background: 'var(--bg-surface)', color: 'var(--text-secondary)', border: '1px solid var(--border-default)' }}
          >
            上一页
          </button>
          <span className="text-sm tabular-nums px-2" style={{ color: 'var(--text-tertiary)' }}>
            {page} / {totalPages}
          </span>
          <button
            onClick={() => setPage(Math.min(totalPages, page + 1))}
            disabled={page === totalPages}
            className="px-4 py-2 rounded-xl text-sm font-medium transition-all cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
            style={{ background: 'var(--bg-surface)', color: 'var(--text-secondary)', border: '1px solid var(--border-default)' }}
          >
            下一页
          </button>
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

      {/* 批量导入弹窗 */}
      <FormModal open={showBatch} title="批量导入账号" onClose={closeBatch}>
        {batchResult ? (
          <div className="space-y-4">
            {/* 汇总 */}
            <div className="grid grid-cols-3 gap-3 text-center">
              <div className="py-3 rounded-xl" style={{ background: 'var(--success-soft)' }}>
                <div className="text-2xl font-bold" style={{ color: 'var(--success)' }}>{batchResult.created}</div>
                <div className="text-xs mt-0.5" style={{ color: 'var(--text-tertiary)' }}>新建</div>
              </div>
              <div className="py-3 rounded-xl" style={{ background: 'var(--info-soft)' }}>
                <div className="text-2xl font-bold" style={{ color: 'var(--info)' }}>{batchResult.updated}</div>
                <div className="text-xs mt-0.5" style={{ color: 'var(--text-tertiary)' }}>更新</div>
              </div>
              <div className="py-3 rounded-xl" style={{ background: batchResult.failed > 0 ? 'var(--danger-soft)' : 'var(--bg-inset)' }}>
                <div className="text-2xl font-bold" style={{ color: batchResult.failed > 0 ? 'var(--danger)' : 'var(--text-tertiary)' }}>{batchResult.failed}</div>
                <div className="text-xs mt-0.5" style={{ color: 'var(--text-tertiary)' }}>失败</div>
              </div>
            </div>
            {/* 明细 */}
            <div className="max-h-60 overflow-y-auto rounded-xl" style={{ border: '1px solid var(--border-default)' }}>
              {batchResult.details.map((item, i) => (
                <div
                  key={i}
                  className="flex items-start gap-3 px-3 py-2.5 text-xs"
                  style={{
                    borderBottom: i < batchResult.details.length - 1 ? '1px solid var(--border-default)' : undefined,
                    background: i % 2 === 0 ? 'transparent' : 'var(--bg-inset)',
                  }}
                >
                  <span
                    className="px-1.5 py-0.5 rounded font-medium flex-shrink-0"
                    style={{
                      background: item.action === 'created' ? 'var(--success-soft)' : item.action === 'updated' ? 'var(--info-soft)' : 'var(--danger-soft)',
                      color: item.action === 'created' ? 'var(--success)' : item.action === 'updated' ? 'var(--info)' : 'var(--danger)',
                    }}
                  >
                    {item.action === 'created' ? '新建' : item.action === 'updated' ? '更新' : '失败'}
                  </span>
                  <div className="flex-1 min-w-0">
                    <code className="block truncate" style={{ color: 'var(--text-secondary)', fontFamily: 'var(--font-mono)' }}>{item.token}</code>
                    {item.email && <span style={{ color: 'var(--text-tertiary)' }}>{item.email}</span>}
                    {item.error && <span style={{ color: 'var(--danger)' }}>{item.error}</span>}
                  </div>
                </div>
              ))}
            </div>
            <div className="flex justify-end gap-2 pt-1">
              <button
                onClick={() => setBatchResult(null)}
                className="px-4 py-2 rounded-xl text-sm font-medium cursor-pointer"
                style={{ color: 'var(--text-secondary)', background: 'var(--bg-inset)' }}
              >
                继续导入
              </button>
              <button
                onClick={closeBatch}
                className="px-5 py-2 rounded-xl text-sm font-medium text-white cursor-pointer"
                style={{ background: 'var(--accent)' }}
              >
                完成
              </button>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            <div>
              <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>
                Token 列表
                <span className="ml-1 font-normal" style={{ color: 'var(--text-tertiary)' }}>（每行一个，自动识别 AT 或 RT）</span>
              </label>
              <textarea
                value={batchTokens}
                onChange={(e) => setBatchTokens(e.target.value)}
                rows={8}
                placeholder={'rt_xxxxxxxx（Refresh Token，rt_ 开头）\neyJhbGci...（Access Token，JWT 格式）\n...'}
                className="w-full px-3 py-2.5 text-sm outline-none transition-all resize-none"
                style={{ ...inputStyle, fontFamily: 'var(--font-mono)', fontSize: '12px', lineHeight: '1.6' }}
                onFocus={inputFocus}
                onBlur={inputBlur}
              />
              <p className="text-[11px] mt-1" style={{ color: 'var(--text-tertiary)' }}>
                以 <code style={{ fontFamily: 'var(--font-mono)' }}>rt_</code> 开头识别为 RT，否则视为 AT。以邮箱为唯一标识，已存在则更新 Token。
              </p>
            </div>
            <div>
              <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>
                分组 <span style={{ color: 'var(--text-tertiary)', fontWeight: 400 }}>（可选）</span>
              </label>
              <select
                value={batchGroupId ?? ''}
                onChange={(e) => setBatchGroupId(e.target.value ? Number(e.target.value) : null)}
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
            <div className="flex justify-end gap-2 pt-1">
              <button
                type="button"
                onClick={closeBatch}
                className="px-4 py-2 rounded-xl text-sm font-medium cursor-pointer"
                style={{ color: 'var(--text-secondary)', background: 'var(--bg-inset)' }}
              >
                取消
              </button>
              <button
                onClick={handleBatchImport}
                disabled={batchImporting || !batchTokens.trim()}
                className="px-5 py-2 rounded-xl text-sm font-medium text-white disabled:opacity-50 transition-all cursor-pointer"
                style={{ background: 'var(--accent)' }}
                onMouseEnter={(e) => { if (!batchImporting) e.currentTarget.style.background = 'var(--accent-hover)' }}
                onMouseLeave={(e) => { e.currentTarget.style.background = 'var(--accent)' }}
              >
                {batchImporting ? '导入中...' : '开始导入'}
              </button>
            </div>
          </div>
        )}
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

function TokenRow({ label, value, onCopy }: { label: string; value: string; onCopy: () => void }) {
  if (!value) return (
    <div className="flex items-center gap-2">
      <span className="text-[11px] font-medium w-6 flex-shrink-0" style={{ color: 'var(--text-tertiary)' }}>{label}</span>
      <span className="text-[11px]" style={{ color: 'var(--text-tertiary)' }}>-</span>
    </div>
  )
  return (
    <div className="flex items-start gap-2">
      <span className="text-[11px] font-medium w-6 flex-shrink-0 pt-0.5" style={{ color: 'var(--text-tertiary)' }}>{label}</span>
      <code
        className="flex-1 text-[11px] break-all leading-relaxed"
        style={{ color: 'var(--text-secondary)', fontFamily: 'var(--font-mono)' }}
      >
        {value}
      </code>
      <button
        onClick={onCopy}
        className="p-1 rounded transition-colors cursor-pointer flex-shrink-0"
        style={{ color: 'var(--accent)' }}
        title="复制"
      >
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
          <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
        </svg>
      </button>
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
