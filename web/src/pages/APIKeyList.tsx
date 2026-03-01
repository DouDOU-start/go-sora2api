import { useCallback, useEffect, useRef, useState } from 'react'
import { listAPIKeys, createAPIKey, updateAPIKey, deleteAPIKey, revealAPIKey } from '../api/apikey'
import { listGroups, type GroupWithCount } from '../api/group'
import GlassCard from '../components/ui/GlassCard'
import LoadingState from '../components/ui/LoadingState'
import ConfirmDialog from '../components/ui/ConfirmDialog'
import FormModal from '../components/ui/FormModal'
import { toast } from '../components/ui/toastStore'
import { getErrorMessage } from '../api/client'
import { motion } from 'framer-motion'
import { format } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import type { SoraAPIKey } from '../types/account'

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

const enabledFilters = [
  { label: '全部', value: '' },
  { label: '已启用', value: 'true' },
  { label: '已禁用', value: 'false' },
]

const PAGE_SIZE = 20

export default function APIKeyList() {
  const [keys, setKeys] = useState<SoraAPIKey[]>([])
  const [total, setTotal] = useState(0)
  const [groups, setGroups] = useState<GroupWithCount[]>([])
  const [loading, setLoading] = useState(true)

  // 筛选 & 分页
  const [enabled, setEnabled] = useState('')
  const [groupId, setGroupId] = useState('')
  const [keyword, setKeyword] = useState('')
  const [inputKeyword, setInputKeyword] = useState('')
  const [page, setPage] = useState(1)

  // 表单
  const [showForm, setShowForm] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [form, setForm] = useState({ name: '', key: '', group_id: '' as string, enabled: true })
  const [submitting, setSubmitting] = useState(false)
  const [confirmState, setConfirmState] = useState<{ open: boolean; id: number }>({ open: false, id: 0 })
  const [revealedKeys, setRevealedKeys] = useState<Record<number, string>>({})

  const mountedRef = useRef(true)

  const fetchData = useCallback(async () => {
    const params: Record<string, unknown> = { page, page_size: PAGE_SIZE }
    if (keyword) params.keyword = keyword
    if (enabled === 'true') params.enabled = true
    if (enabled === 'false') params.enabled = false
    if (groupId === 'null') params.group_id = 'null'
    else if (groupId) params.group_id = Number(groupId)
    const [keysRes, groupsRes] = await Promise.all([listAPIKeys(params), listGroups()])
    return {
      list: keysRes.data.list ?? [],
      total: keysRes.data.total,
      groups: groupsRes.data ?? [],
    }
  }, [page, keyword, enabled, groupId])

  useEffect(() => {
    mountedRef.current = true
    let canceled = false
    void (async () => {
      try {
        const data = await fetchData()
        if (!canceled && mountedRef.current) {
          setKeys(data.list)
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
        setKeys(data.list)
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
    setForm({ name: '', key: '', group_id: '', enabled: true })
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      const data = {
        name: form.name,
        key: form.key || undefined,
        group_id: form.group_id ? Number(form.group_id) : null,
        enabled: form.enabled,
      }
      if (editId) {
        await updateAPIKey(editId, data)
        toast.success('API Key 已更新')
        closeForm()
      } else {
        await createAPIKey(data)
        toast.success('API Key 已创建')
        closeForm()
      }
      reload()
    } catch (err) {
      toast.error(getErrorMessage(err, editId ? '更新失败' : '创建失败'))
    }
    setSubmitting(false)
  }

  const handleDelete = (id: number) => {
    setConfirmState({ open: true, id })
  }

  const confirmDelete = async () => {
    const id = confirmState.id
    setConfirmState({ open: false, id: 0 })
    try {
      await deleteAPIKey(id)
      toast.success('API Key 已删除')
      reload()
    } catch (err) {
      toast.error(getErrorMessage(err, '删除失败'))
    }
  }

  const handleReveal = async (id: number) => {
    try {
      const res = await revealAPIKey(id)
      setRevealedKeys((prev) => ({ ...prev, [id]: res.data.key }))
    } catch {
      toast.error('获取密钥失败')
    }
  }

  const handleHide = (id: number) => {
    setRevealedKeys((prev) => {
      const next = { ...prev }
      delete next[id]
      return next
    })
  }

  const copyKey = async (key: string) => {
    try {
      await navigator.clipboard.writeText(key)
    } catch {
      const textarea = document.createElement('textarea')
      textarea.value = key
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
    }
    toast.success('已复制到剪贴板')
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
            API Keys
          </h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--text-tertiary)' }}>
            共 {total} 个密钥，用于 /v1/ 接口认证
          </p>
        </div>
        <button
          onClick={() => { setEditId(null); setForm({ name: '', key: '', group_id: '', enabled: true }); setShowForm(true) }}
          className="px-4 py-2 rounded-xl text-sm font-medium text-white transition-all cursor-pointer"
          style={{ background: 'var(--accent)' }}
          onMouseEnter={(e) => e.currentTarget.style.background = 'var(--accent-hover)'}
          onMouseLeave={(e) => e.currentTarget.style.background = 'var(--accent)'}
        >
          + 新建密钥
        </button>
      </motion.div>

      {/* 筛选栏 */}
      <div className="flex flex-col sm:flex-row gap-3 mb-5 flex-wrap">
        <div
          className="flex items-center gap-0.5 p-1 rounded-xl"
          style={{ background: 'var(--bg-inset)' }}
        >
          {enabledFilters.map((f) => (
            <button
              key={f.value}
              onClick={() => { setEnabled(f.value); setPage(1) }}
              className="px-3 py-1.5 rounded-lg text-[13px] font-medium transition-all cursor-pointer"
              style={{
                background: enabled === f.value ? 'var(--bg-surface)' : 'transparent',
                color: enabled === f.value ? 'var(--text-primary)' : 'var(--text-tertiary)',
                boxShadow: enabled === f.value ? 'var(--shadow-sm)' : 'none',
              }}
            >
              {f.label}
            </button>
          ))}
        </div>
        {/* 分组筛选 */}
        <select
          value={groupId}
          onChange={(e) => { setGroupId(e.target.value); setPage(1) }}
          className="px-3 py-1.5 text-sm outline-none transition-all cursor-pointer rounded-xl"
          style={{ ...inputStyle, minWidth: 120 }}
          onFocus={inputFocus}
          onBlur={inputBlur}
        >
          <option value="">全部分组</option>
          <option value="null">未绑定分组</option>
          {groups.map((g) => (
            <option key={g.id} value={g.id}>{g.name}</option>
          ))}
        </select>
        <form onSubmit={handleSearch} className="flex items-center gap-2 flex-1">
          <input
            value={inputKeyword}
            onChange={(e) => setInputKeyword(e.target.value)}
            placeholder="搜索密钥名称"
            className="flex-1 px-3 py-1.5 text-sm outline-none transition-all"
            style={{ ...inputStyle, minWidth: 160 }}
            onFocus={inputFocus}
            onBlur={inputBlur}
          />
          <button
            type="submit"
            className="px-3 py-1.5 rounded-xl text-sm font-medium transition-all cursor-pointer"
            style={{ background: 'var(--accent)', color: '#fff' }}
            onMouseEnter={(e) => (e.currentTarget as HTMLButtonElement).style.background = 'var(--accent-hover)'}
            onMouseLeave={(e) => (e.currentTarget as HTMLButtonElement).style.background = 'var(--accent)'}
          >
            搜索
          </button>
          {(keyword || enabled || groupId) && (
            <button
              type="button"
              onClick={() => { setInputKeyword(''); setKeyword(''); setEnabled(''); setGroupId(''); setPage(1) }}
              className="px-3 py-1.5 rounded-xl text-sm font-medium transition-all cursor-pointer"
              style={{ background: 'var(--bg-inset)', color: 'var(--text-tertiary)' }}
            >
              清除
            </button>
          )}
        </form>
      </div>

      {/* 列表 */}
      {keys.length === 0 ? (
        <div className="text-center py-20" style={{ color: 'var(--text-tertiary)' }}>
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" className="mx-auto mb-3 opacity-40">
            <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />
          </svg>
          <p>暂无 API Key</p>
          {!keyword && !enabled && <p className="text-xs mt-1">点击上方按钮创建密钥以启用认证</p>}
        </div>
      ) : (
        <div className="space-y-3">
          {keys.map((k, i) => (
            <GlassCard key={k.id} hover delay={i} className="overflow-hidden">
              <div className="p-5 flex items-center gap-4">
                {/* 图标 */}
                <div
                  className="w-10 h-10 rounded-xl flex items-center justify-center flex-shrink-0"
                  style={{ background: k.enabled ? 'var(--accent-soft)' : 'var(--bg-inset)', color: k.enabled ? 'var(--accent)' : 'var(--text-tertiary)' }}
                >
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                    <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />
                  </svg>
                </div>

                {/* 信息 */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-0.5">
                    <span className="text-sm font-semibold truncate" style={{ color: 'var(--text-primary)' }}>{k.name}</span>
                    <span
                      className="text-[11px] font-medium px-2 py-0.5 rounded-full flex-shrink-0"
                      style={{
                        background: k.enabled ? 'var(--success-soft)' : 'var(--bg-inset)',
                        color: k.enabled ? 'var(--success)' : 'var(--text-tertiary)',
                      }}
                    >
                      {k.enabled ? '启用' : '禁用'}
                    </span>
                  </div>
                  <div className="flex items-center gap-3 text-xs" style={{ color: 'var(--text-tertiary)' }}>
                    <span className="inline-flex items-center gap-1.5">
                      <code className="font-mono" style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>
                        {revealedKeys[k.id] || k.key_hint}
                      </code>
                      {revealedKeys[k.id] ? (
                        <>
                          <button onClick={() => copyKey(revealedKeys[k.id])} className="p-0.5 rounded transition-colors cursor-pointer" style={{ color: 'var(--accent)' }} title="复制">
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                              <rect x="9" y="9" width="13" height="13" rx="2" ry="2" /><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
                            </svg>
                          </button>
                          <button onClick={() => handleHide(k.id)} className="p-0.5 rounded transition-colors cursor-pointer" style={{ color: 'var(--text-tertiary)' }} title="隐藏">
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                              <path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24" />
                              <line x1="1" y1="1" x2="23" y2="23" />
                            </svg>
                          </button>
                        </>
                      ) : (
                        <button onClick={() => handleReveal(k.id)} className="p-0.5 rounded transition-colors cursor-pointer" style={{ color: 'var(--text-tertiary)' }} title="显示完整密钥">
                          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" /><circle cx="12" cy="12" r="3" />
                          </svg>
                        </button>
                      )}
                    </span>
                    {k.group_name ? (
                      <span className="inline-flex items-center gap-1">
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z" /></svg>
                        {k.group_name}
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1" style={{ color: 'var(--warning, #e6a700)' }}>
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                          <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
                          <line x1="12" y1="9" x2="12" y2="13" /><line x1="12" y1="17" x2="12.01" y2="17" />
                        </svg>
                        未绑定分组
                      </span>
                    )}
                    <span>调用 {k.usage_count} 次</span>
                    {k.last_used_at && (
                      <span className="hidden sm:inline">
                        最后使用 {format(new Date(k.last_used_at), 'MM-dd HH:mm', { locale: zhCN })}
                      </span>
                    )}
                  </div>
                </div>

                {/* 操作 */}
                <div className="flex items-center gap-1 flex-shrink-0">
                  <button
                    onClick={() => { setEditId(k.id); setForm({ name: k.name, key: '', group_id: k.group_id ? String(k.group_id) : '', enabled: k.enabled }); setShowForm(true) }}
                    className="p-1.5 rounded-lg transition-colors cursor-pointer" style={{ color: 'var(--text-tertiary)' }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-inset)' }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
                    title="编辑"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                      <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" /><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" />
                    </svg>
                  </button>
                  <button
                    onClick={() => handleDelete(k.id)}
                    className="p-1.5 rounded-lg transition-colors cursor-pointer" style={{ color: 'var(--text-tertiary)' }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--danger-soft)'; e.currentTarget.style.color = 'var(--danger)' }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--text-tertiary)' }}
                    title="删除"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                      <polyline points="3 6 5 6 21 6" /><path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6" />
                      <path d="M10 11v6" /><path d="M14 11v6" />
                    </svg>
                  </button>
                </div>
              </div>
            </GlassCard>
          ))}
        </div>
      )}

      {/* 分页 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-3 mt-8">
          <button onClick={() => setPage(Math.max(1, page - 1))} disabled={page === 1}
            className="px-4 py-2 rounded-xl text-sm font-medium transition-all cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
            style={{ background: 'var(--bg-surface)', color: 'var(--text-secondary)', border: '1px solid var(--border-default)' }}>
            上一页
          </button>
          <span className="text-sm tabular-nums px-2" style={{ color: 'var(--text-tertiary)' }}>{page} / {totalPages}</span>
          <button onClick={() => setPage(Math.min(totalPages, page + 1))} disabled={page === totalPages}
            className="px-4 py-2 rounded-xl text-sm font-medium transition-all cursor-pointer disabled:opacity-30 disabled:cursor-not-allowed"
            style={{ background: 'var(--bg-surface)', color: 'var(--text-secondary)', border: '1px solid var(--border-default)' }}>
            下一页
          </button>
        </div>
      )}

      {/* 添加/编辑弹窗 */}
      <FormModal open={showForm} title={editId ? '编辑 API Key' : '新建 API Key'} onClose={closeForm}>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>名称</label>
            <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required
              placeholder="例：生产环境、测试用" className="w-full px-3 py-2.5 text-sm outline-none transition-all"
              style={inputStyle} onFocus={inputFocus} onBlur={inputBlur} />
          </div>
          <div>
            <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>
              {editId ? 'Key（留空则不修改）' : 'Key（留空自动生成）'}
            </label>
            <input value={form.key} onChange={(e) => setForm({ ...form, key: e.target.value })}
              placeholder="sk-... 或留空自动生成" className="w-full px-3 py-2.5 text-sm outline-none transition-all font-mono"
              style={{ ...inputStyle, fontSize: '13px' }} onFocus={inputFocus} onBlur={inputBlur} />
          </div>
          <div>
            <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>绑定分组</label>
            <select value={form.group_id} onChange={(e) => setForm({ ...form, group_id: e.target.value })} required
              className="w-full px-3 py-2.5 text-sm outline-none transition-all cursor-pointer"
              style={inputStyle} onFocus={inputFocus} onBlur={inputBlur}>
              <option value="" disabled>请选择分组</option>
              {groups.map((g) => (
                <option key={g.id} value={g.id}>{g.name}（{g.account_count} 个账号）</option>
              ))}
            </select>
            <p className="text-[11px] mt-1" style={{ color: 'var(--text-tertiary)' }}>API Key 必须绑定分组，仅可调度该分组内的账号</p>
          </div>
          {editId && (
            <div className="flex items-center gap-2">
              <input type="checkbox" id="enabled" checked={form.enabled} onChange={(e) => setForm({ ...form, enabled: e.target.checked })} className="cursor-pointer" />
              <label htmlFor="enabled" className="text-sm cursor-pointer" style={{ color: 'var(--text-secondary)' }}>启用</label>
            </div>
          )}
          <div className="flex justify-end gap-2 pt-2">
            <button type="button" onClick={closeForm} className="px-4 py-2 rounded-xl text-sm font-medium transition-colors cursor-pointer"
              style={{ color: 'var(--text-secondary)', background: 'var(--bg-inset)' }}>取消</button>
            <button type="submit" disabled={submitting} className="px-5 py-2 rounded-xl text-sm font-medium text-white disabled:opacity-50 transition-all cursor-pointer"
              style={{ background: 'var(--accent)' }}
              onMouseEnter={(e) => e.currentTarget.style.background = 'var(--accent-hover)'}
              onMouseLeave={(e) => e.currentTarget.style.background = 'var(--accent)'}>
              {submitting ? '保存中...' : editId ? '更新' : '创建'}
            </button>
          </div>
        </form>
      </FormModal>

      {/* 删除确认 */}
      <ConfirmDialog open={confirmState.open} title="删除 API Key"
        message="确定删除此 API Key？使用此 Key 的客户端将无法访问。"
        confirmLabel="删除" danger onConfirm={confirmDelete}
        onCancel={() => setConfirmState({ open: false, id: 0 })} />
    </div>
  )
}
