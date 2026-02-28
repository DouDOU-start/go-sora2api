import { useCallback, useEffect, useState } from 'react'
import { listAPIKeys, createAPIKey, updateAPIKey, deleteAPIKey, revealAPIKey } from '../api/apikey'
import { listGroups, type GroupWithCount } from '../api/group'
import GlassCard from '../components/ui/GlassCard'
import LoadingState from '../components/ui/LoadingState'
import ConfirmDialog from '../components/ui/ConfirmDialog'
import FormModal from '../components/ui/FormModal'
import { toast } from '../components/ui/Toast'
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

export default function APIKeyList() {
  const [keys, setKeys] = useState<SoraAPIKey[]>([])
  const [groups, setGroups] = useState<GroupWithCount[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [form, setForm] = useState({ name: '', key: '', group_id: '' as string, enabled: true })
  const [submitting, setSubmitting] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)
  const [confirmState, setConfirmState] = useState<{ open: boolean; id: number }>({ open: false, id: 0 })
  const [newKeyVisible, setNewKeyVisible] = useState<{ id: number; key: string } | null>(null)
  const [revealedKeys, setRevealedKeys] = useState<Record<number, string>>({})

  const reload = useCallback(() => setRefreshKey((k) => k + 1), [])

  const closeForm = () => {
    setShowForm(false)
    setEditId(null)
    setForm({ name: '', key: '', group_id: '', enabled: true })
  }

  useEffect(() => {
    const load = async () => {
      try {
        const [keysRes, groupsRes] = await Promise.all([listAPIKeys(), listGroups()])
        setKeys(keysRes.data ?? [])
        setGroups(groupsRes.data ?? [])
      } catch { /* ignore */ }
      setLoading(false)
    }
    load()
  }, [refreshKey])

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
        const res = await createAPIKey(data)
        toast.success('API Key 已创建')
        // 显示新创建的完整 Key
        if (res.data?.key) {
          setNewKeyVisible({ id: res.data.id, key: res.data.key })
        }
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

  const copyKey = (key: string) => {
    navigator.clipboard.writeText(key)
    toast.success('已复制到剪贴板')
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
            API Keys
          </h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--text-tertiary)' }}>
            共 {keys.length} 个密钥，用于 /v1/ 接口认证
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

      {/* 无 Key 提示 */}
      {keys.length === 0 ? (
        <div className="text-center py-20" style={{ color: 'var(--text-tertiary)' }}>
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" className="mx-auto mb-3 opacity-40">
            <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 11-7.778 7.778 5.5 5.5 0 017.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />
          </svg>
          <p>暂无 API Key，所有 /v1/ 接口无需认证</p>
          <p className="text-xs mt-1">点击上方按钮创建密钥以启用认证</p>
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
                          <button
                            onClick={() => copyKey(revealedKeys[k.id])}
                            className="p-0.5 rounded transition-colors cursor-pointer"
                            style={{ color: 'var(--accent)' }}
                            title="复制"
                          >
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                              <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                              <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
                            </svg>
                          </button>
                          <button
                            onClick={() => handleHide(k.id)}
                            className="p-0.5 rounded transition-colors cursor-pointer"
                            style={{ color: 'var(--text-tertiary)' }}
                            title="隐藏"
                          >
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                              <path d="M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19m-6.72-1.07a3 3 0 11-4.24-4.24" />
                              <line x1="1" y1="1" x2="23" y2="23" />
                            </svg>
                          </button>
                        </>
                      ) : (
                        <button
                          onClick={() => handleReveal(k.id)}
                          className="p-0.5 rounded transition-colors cursor-pointer"
                          style={{ color: 'var(--text-tertiary)' }}
                          title="显示完整密钥"
                        >
                          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                            <circle cx="12" cy="12" r="3" />
                          </svg>
                        </button>
                      )}
                    </span>
                    {k.group_name && (
                      <span className="inline-flex items-center gap-1">
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                          <path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z" />
                        </svg>
                        {k.group_name}
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
                    onClick={() => {
                      setEditId(k.id)
                      setForm({
                        name: k.name,
                        key: '',
                        group_id: k.group_id ? String(k.group_id) : '',
                        enabled: k.enabled,
                      })
                      setShowForm(true)
                    }}
                    className="p-1.5 rounded-lg transition-colors cursor-pointer"
                    style={{ color: 'var(--text-tertiary)' }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-inset)' }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
                    title="编辑"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                      <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" />
                      <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" />
                    </svg>
                  </button>
                  <button
                    onClick={() => handleDelete(k.id)}
                    className="p-1.5 rounded-lg transition-colors cursor-pointer"
                    style={{ color: 'var(--text-tertiary)' }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--danger-soft)'; e.currentTarget.style.color = 'var(--danger)' }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--text-tertiary)' }}
                    title="删除"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                      <polyline points="3 6 5 6 21 6" />
                      <path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6" />
                      <path d="M10 11v6" /><path d="M14 11v6" />
                    </svg>
                  </button>
                </div>
              </div>
            </GlassCard>
          ))}
        </div>
      )}

      {/* 新建 Key 后显示完整密钥 */}
      <FormModal
        open={!!newKeyVisible}
        title="密钥已创建"
        onClose={() => setNewKeyVisible(null)}
      >
        <div className="space-y-3">
          <p className="text-sm" style={{ color: 'var(--text-secondary)' }}>
            请立即复制保存，关闭后将无法再次查看完整密钥。
          </p>
          <div
            className="flex items-center gap-2 px-3.5 py-2.5 rounded-lg"
            style={{ background: 'var(--bg-inset)', border: '1px solid var(--border-default)' }}
          >
            <code className="flex-1 text-sm font-mono break-all" style={{ color: 'var(--text-primary)', fontSize: '13px' }}>
              {newKeyVisible?.key}
            </code>
            <button
              onClick={() => newKeyVisible && copyKey(newKeyVisible.key)}
              className="p-1.5 rounded-lg transition-colors cursor-pointer flex-shrink-0"
              style={{ color: 'var(--accent)' }}
              title="复制"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
              </svg>
            </button>
          </div>
          <div className="flex justify-end pt-1">
            <button
              onClick={() => setNewKeyVisible(null)}
              className="px-5 py-2 rounded-xl text-sm font-medium text-white transition-all cursor-pointer"
              style={{ background: 'var(--accent)' }}
              onMouseEnter={(e) => e.currentTarget.style.background = 'var(--accent-hover)'}
              onMouseLeave={(e) => e.currentTarget.style.background = 'var(--accent)'}
            >
              已复制，关闭
            </button>
          </div>
        </div>
      </FormModal>

      {/* 添加/编辑弹窗 */}
      <FormModal
        open={showForm}
        title={editId ? '编辑 API Key' : '新建 API Key'}
        onClose={closeForm}
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>名称</label>
            <input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              required
              placeholder="例：生产环境、测试用"
              className="w-full px-3 py-2.5 text-sm outline-none transition-all"
              style={inputStyle}
              onFocus={inputFocus}
              onBlur={inputBlur}
            />
          </div>
          <div>
            <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>
              {editId ? 'Key（留空则不修改）' : 'Key（留空自动生成）'}
            </label>
            <input
              value={form.key}
              onChange={(e) => setForm({ ...form, key: e.target.value })}
              placeholder="sk-... 或留空自动生成"
              className="w-full px-3 py-2.5 text-sm outline-none transition-all font-mono"
              style={{ ...inputStyle, fontSize: '13px' }}
              onFocus={inputFocus}
              onBlur={inputBlur}
            />
          </div>
          <div>
            <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>绑定分组</label>
            <select
              value={form.group_id}
              onChange={(e) => setForm({ ...form, group_id: e.target.value })}
              className="w-full px-3 py-2.5 text-sm outline-none transition-all cursor-pointer"
              style={inputStyle}
              onFocus={inputFocus}
              onBlur={inputBlur}
            >
              <option value="">不绑定（可使用所有账号）</option>
              {groups.map((g) => (
                <option key={g.id} value={g.id}>{g.name}（{g.account_count} 个账号）</option>
              ))}
            </select>
            <p className="text-[11px] mt-1" style={{ color: 'var(--text-tertiary)' }}>
              绑定后，此 Key 仅可使用该分组内的账号
            </p>
          </div>
          {editId && (
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="enabled"
                checked={form.enabled}
                onChange={(e) => setForm({ ...form, enabled: e.target.checked })}
                className="cursor-pointer"
              />
              <label htmlFor="enabled" className="text-sm cursor-pointer" style={{ color: 'var(--text-secondary)' }}>
                启用
              </label>
            </div>
          )}
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
              {submitting ? '保存中...' : editId ? '更新' : '创建'}
            </button>
          </div>
        </form>
      </FormModal>

      {/* 删除确认 */}
      <ConfirmDialog
        open={confirmState.open}
        title="删除 API Key"
        message="确定删除此 API Key？使用此 Key 的客户端将无法访问。"
        confirmLabel="删除"
        danger
        onConfirm={confirmDelete}
        onCancel={() => setConfirmState({ open: false, id: 0 })}
      />
    </div>
  )
}
