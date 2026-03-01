import { useCallback, useEffect, useState } from 'react'
import { listGroups, createGroup, updateGroup, deleteGroup, type GroupWithCount } from '../api/group'
import GlassCard from '../components/ui/GlassCard'
import LoadingState from '../components/ui/LoadingState'
import ConfirmDialog from '../components/ui/ConfirmDialog'
import FormModal from '../components/ui/FormModal'
import { toast } from '../components/ui/toastStore'
import { getErrorMessage } from '../api/client'
import { motion } from 'framer-motion'

const inputStyle = {
  background: 'var(--bg-inset)',
  border: '1px solid var(--border-default)',
  color: 'var(--text-primary)',
  borderRadius: 'var(--radius-md)',
}
const inputFocus = (e: React.FocusEvent<HTMLInputElement>) => {
  e.target.style.borderColor = 'var(--accent)'
  e.target.style.boxShadow = '0 0 0 3px var(--accent-soft)'
}
const inputBlur = (e: React.FocusEvent<HTMLInputElement>) => {
  e.target.style.borderColor = 'var(--border-default)'
  e.target.style.boxShadow = 'none'
}

export default function GroupList() {
  const [groups, setGroups] = useState<GroupWithCount[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [form, setForm] = useState({ name: '', description: '', enabled: true })
  const [submitting, setSubmitting] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)
  const [confirmState, setConfirmState] = useState<{ open: boolean; id: number }>({ open: false, id: 0 })

  const reload = useCallback(() => setRefreshKey((k) => k + 1), [])

  const closeForm = () => {
    setShowForm(false)
    setEditId(null)
    setForm({ name: '', description: '', enabled: true })
  }

  useEffect(() => {
    const load = async () => {
      try {
        const res = await listGroups()
        setGroups(res.data ?? [])
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
        await updateGroup(editId, form)
        toast.success('分组已更新')
      } else {
        await createGroup(form)
        toast.success('分组已创建')
      }
      closeForm()
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
      await deleteGroup(id)
      toast.success('分组已删除')
      reload()
    } catch (err) {
      toast.error(getErrorMessage(err, '删除失败'))
    }
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
            分组管理
          </h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--text-tertiary)' }}>
            共 {groups.length} 个分组
          </p>
        </div>
        <button
          onClick={() => { setEditId(null); setForm({ name: '', description: '', enabled: true }); setShowForm(true) }}
          className="px-4 py-2 rounded-xl text-sm font-medium text-white transition-all cursor-pointer"
          style={{ background: 'var(--accent)' }}
          onMouseEnter={(e) => e.currentTarget.style.background = 'var(--accent-hover)'}
          onMouseLeave={(e) => e.currentTarget.style.background = 'var(--accent)'}
        >
          + 新建分组
        </button>
      </motion.div>

      {/* 分组列表 */}
      {groups.length === 0 ? (
        <div className="text-center py-20" style={{ color: 'var(--text-tertiary)' }}>
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" className="mx-auto mb-3 opacity-40">
            <path d="M3 7a2 2 0 012-2h4l2 2h8a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2V7z" />
          </svg>
          暂无分组，点击上方按钮创建
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3 sm:gap-4">
          {groups.map((g, i) => (
            <GlassCard key={g.id} hover delay={i} className="p-5">
              <div className="flex items-start justify-between mb-3">
                <div
                  className="w-10 h-10 rounded-xl flex items-center justify-center text-sm font-semibold"
                  style={{ background: 'var(--accent-soft)', color: 'var(--accent)' }}
                >
                  {g.name.charAt(0).toUpperCase()}
                </div>
                <div className="flex items-center gap-1">
                  <button
                    onClick={() => {
                      setEditId(g.id)
                      setForm({ name: g.name, description: g.description, enabled: g.enabled })
                      setShowForm(true)
                    }}
                    className="p-1.5 rounded-lg transition-colors cursor-pointer"
                    style={{ color: 'var(--text-tertiary)' }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-inset)' }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent' }}
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                      <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7" />
                      <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z" />
                    </svg>
                  </button>
                  <button
                    onClick={() => handleDelete(g.id)}
                    className="p-1.5 rounded-lg transition-colors cursor-pointer"
                    style={{ color: 'var(--text-tertiary)' }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--danger-soft)'; e.currentTarget.style.color = 'var(--danger)' }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--text-tertiary)' }}
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
                      <polyline points="3 6 5 6 21 6" />
                      <path d="M19 6l-1 14a2 2 0 01-2 2H8a2 2 0 01-2-2L5 6" />
                      <path d="M10 11v6" /><path d="M14 11v6" />
                    </svg>
                  </button>
                </div>
              </div>
              <h3 className="text-sm font-semibold mb-0.5" style={{ color: 'var(--text-primary)' }}>{g.name}</h3>
              <p className="text-xs mb-3 line-clamp-1" style={{ color: 'var(--text-tertiary)' }}>
                {g.description || '无描述'}
              </p>
              <div
                className="inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-full"
                style={{ background: 'var(--bg-inset)', color: 'var(--text-secondary)' }}
              >
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <circle cx="12" cy="8" r="4" /><path d="M6 21v-1a6 6 0 0112 0v1" />
                </svg>
                {g.account_count} 个账号
              </div>
            </GlassCard>
          ))}
        </div>
      )}

      {/* 添加/编辑弹窗 */}
      <FormModal
        open={showForm}
        title={editId ? '编辑分组' : '新建分组'}
        onClose={closeForm}
      >
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>组名</label>
            <input
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              required
              placeholder="分组名称"
              className="w-full px-3 py-2.5 text-sm outline-none transition-all"
              style={inputStyle}
              onFocus={inputFocus}
              onBlur={inputBlur}
            />
          </div>
          <div>
            <label className="block text-[13px] font-medium mb-1.5" style={{ color: 'var(--text-secondary)' }}>描述</label>
            <input
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              placeholder="可选描述"
              className="w-full px-3 py-2.5 text-sm outline-none transition-all"
              style={inputStyle}
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
              {submitting ? '保存中...' : editId ? '更新' : '创建'}
            </button>
          </div>
        </form>
      </FormModal>

      {/* 删除确认对话框 */}
      <ConfirmDialog
        open={confirmState.open}
        title="删除分组"
        message="确定删除此分组？组内账号将变为未分组状态。"
        confirmLabel="删除"
        danger
        onConfirm={confirmDelete}
        onCancel={() => setConfirmState({ open: false, id: 0 })}
      />
    </div>
  )
}
