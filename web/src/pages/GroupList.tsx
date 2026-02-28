import { useCallback, useEffect, useState } from 'react'
import { listGroups, createGroup, updateGroup, deleteGroup, type GroupWithCount } from '../api/group'
import GlassCard from '../components/ui/GlassCard'
import LoadingState from '../components/ui/LoadingState'
import { AnimatePresence, motion } from 'framer-motion'

export default function GroupList() {
  const [groups, setGroups] = useState<GroupWithCount[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [form, setForm] = useState({ name: '', description: '', enabled: true })
  const [submitting, setSubmitting] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)

  const reload = useCallback(() => setRefreshKey((k) => k + 1), [])

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
      } else {
        await createGroup(form)
      }
      setShowForm(false)
      setEditId(null)
      setForm({ name: '', description: '', enabled: true })
      reload()
    } catch { /* ignore */ }
    setSubmitting(false)
  }

  const handleDelete = async (id: number) => {
    if (!confirm('确定删除此分组？组内账号将变为未分组状态。')) return
    try {
      await deleteGroup(id)
      reload()
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : '删除失败')
    }
  }

  if (loading) return <LoadingState />

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">账号组管理</h2>
        <button
          onClick={() => {
            setShowForm(!showForm)
            setEditId(null)
            setForm({ name: '', description: '', enabled: true })
          }}
          className="px-4 py-2 rounded-lg bg-gradient-to-r from-indigo-500 to-purple-500 text-white text-sm font-medium hover:from-indigo-600 hover:to-purple-600 transition-all"
        >
          {showForm ? '取消' : '新建分组'}
        </button>
      </div>

      {/* 表单 */}
      <AnimatePresence>
        {showForm && (
          <motion.div initial={{ opacity: 0, height: 0 }} animate={{ opacity: 1, height: 'auto' }} exit={{ opacity: 0, height: 0 }}>
            <GlassCard className="p-5 mb-6">
              <form onSubmit={handleSubmit} className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">组名</label>
                  <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required
                    className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">描述</label>
                  <input value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })}
                    className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div className="sm:col-span-2 flex justify-end">
                  <button type="submit" disabled={submitting}
                    className="px-4 py-2 rounded-lg bg-indigo-500 text-white text-sm font-medium hover:bg-indigo-600 disabled:opacity-50">
                    {submitting ? '保存中...' : editId ? '更新' : '创建'}
                  </button>
                </div>
              </form>
            </GlassCard>
          </motion.div>
        )}
      </AnimatePresence>

      {/* 分组列表 */}
      {groups.length === 0 ? (
        <div className="text-center text-gray-400 py-16">暂无分组</div>
      ) : (
        <div className="space-y-3">
          {groups.map((g) => (
            <GlassCard key={g.id} className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium text-gray-900 dark:text-gray-100">{g.name}</p>
                  <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                    {g.description || '无描述'} · {g.account_count} 个账号
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => { setEditId(g.id); setForm({ name: g.name, description: g.description, enabled: g.enabled }); setShowForm(true) }}
                    className="px-3 py-1 text-xs rounded-lg text-indigo-600 dark:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-950"
                  >编辑</button>
                  <button
                    onClick={() => handleDelete(g.id)}
                    className="px-3 py-1 text-xs rounded-lg text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950"
                  >删除</button>
                </div>
              </div>
            </GlassCard>
          ))}
        </div>
      )}
    </div>
  )
}
