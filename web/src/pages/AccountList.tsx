import { useCallback, useEffect, useState } from 'react'
import { listAllAccounts, createAccount, updateAccount, deleteAccount, refreshAccountToken, getAccountStatus } from '../api/account'
import { listGroups } from '../api/group'
import type { SoraAccount, CreateAccountRequest, SoraAccountGroup } from '../types/account'
import GlassCard from '../components/ui/GlassCard'
import StatusBadge from '../components/ui/StatusBadge'
import LoadingState from '../components/ui/LoadingState'
import { AnimatePresence, motion } from 'framer-motion'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

function timeAgo(ts: string | null) {
  if (!ts) return '-'
  return formatDistanceToNow(new Date(ts), { addSuffix: true, locale: zhCN })
}

const emptyForm: CreateAccountRequest = { name: '', access_token: '', refresh_token: '', group_id: null }

export default function AccountList() {
  const [accounts, setAccounts] = useState<SoraAccount[]>([])
  const [groups, setGroups] = useState<SoraAccountGroup[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editId, setEditId] = useState<number | null>(null)
  const [form, setForm] = useState<CreateAccountRequest>({ ...emptyForm })
  const [submitting, setSubmitting] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)

  const reload = useCallback(() => setRefreshKey((k) => k + 1), [])

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
      } else {
        await createAccount(form)
      }
      setShowForm(false)
      setEditId(null)
      setForm({ ...emptyForm })
      reload()
    } catch { /* ignore */ }
    setSubmitting(false)
  }

  const handleEdit = (acc: SoraAccount) => {
    setEditId(acc.id)
    setForm({
      name: acc.name,
      access_token: '',
      refresh_token: '',
      group_id: acc.group_id,
    })
    setShowForm(true)
  }

  const handleDelete = async (id: number) => {
    if (!confirm('确定删除此账号？')) return
    try {
      await deleteAccount(id)
      reload()
    } catch { /* ignore */ }
  }

  const handleRefresh = async (id: number) => {
    try {
      await refreshAccountToken(id)
      reload()
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : '刷新失败')
    }
  }

  const handleSync = async (id: number) => {
    try {
      await getAccountStatus(id)
      reload()
    } catch { /* ignore */ }
  }

  if (loading) return <LoadingState />

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">账号管理</h2>
        <button
          onClick={() => {
            setShowForm(!showForm)
            setEditId(null)
            setForm({ ...emptyForm })
          }}
          className="px-4 py-2 rounded-lg bg-gradient-to-r from-indigo-500 to-purple-500 text-white text-sm font-medium hover:from-indigo-600 hover:to-purple-600 transition-all"
        >
          {showForm ? '取消' : '添加账号'}
        </button>
      </div>

      {/* 添加/编辑表单 */}
      <AnimatePresence>
        {showForm && (
          <motion.div initial={{ opacity: 0, height: 0 }} animate={{ opacity: 1, height: 'auto' }} exit={{ opacity: 0, height: 0 }}>
            <GlassCard className="p-5 mb-6">
              <form onSubmit={handleSubmit} className="space-y-4">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">名称</label>
                    <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required
                      placeholder="备注名称"
                      className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-indigo-500" />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">分组 <span className="text-gray-400 font-normal">（可选）</span></label>
                    <select value={form.group_id ?? ''} onChange={(e) => setForm({ ...form, group_id: e.target.value ? Number(e.target.value) : null })}
                      className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-indigo-500">
                      <option value="">未分组</option>
                      {groups.map(g => (
                        <option key={g.id} value={g.id}>{g.name}</option>
                      ))}
                    </select>
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Access Token</label>
                  <input value={form.access_token} onChange={(e) => setForm({ ...form, access_token: e.target.value })}
                    placeholder={editId ? '留空则不修改' : 'eyJhbGci...'}
                    className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-sm font-mono text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Refresh Token <span className="text-gray-400 font-normal">（可选）</span></label>
                  <input value={form.refresh_token} onChange={(e) => setForm({ ...form, refresh_token: e.target.value })}
                    placeholder={editId ? '留空则不修改' : 'v1.rt-...'}
                    className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 text-sm font-mono text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-indigo-500" />
                </div>
                <div className="flex justify-end">
                  <button type="submit" disabled={submitting}
                    className="px-4 py-2 rounded-lg bg-indigo-500 text-white text-sm font-medium hover:bg-indigo-600 disabled:opacity-50">
                    {submitting ? '保存中...' : editId ? '更新' : '添加'}
                  </button>
                </div>
              </form>
            </GlassCard>
          </motion.div>
        )}
      </AnimatePresence>

      {/* 账号列表 */}
      {accounts.length === 0 ? (
        <div className="text-center text-gray-400 py-16">暂无账号</div>
      ) : (
        <div className="space-y-3">
          {accounts.map((acc) => (
            <GlassCard key={acc.id} className="p-4">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="font-medium text-sm text-gray-900 dark:text-gray-100">{acc.name}</span>
                  <StatusBadge status={acc.status} />
                  {acc.rate_limit_reached && (
                    <span className="text-xs px-2 py-0.5 rounded-full bg-orange-100 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400">限流中</span>
                  )}
                  {acc.group_name && (
                    <span className="text-xs px-2 py-0.5 rounded-full bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400">{acc.group_name}</span>
                  )}
                </div>
                <div className="flex items-center gap-1 flex-shrink-0">
                  <button onClick={() => handleRefresh(acc.id)}
                    className="px-2 py-1 text-xs rounded text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-950">刷新 Token</button>
                  <button onClick={() => handleSync(acc.id)}
                    className="px-2 py-1 text-xs rounded text-green-600 dark:text-green-400 hover:bg-green-50 dark:hover:bg-green-950">同步状态</button>
                  <button onClick={() => handleEdit(acc)}
                    className="px-2 py-1 text-xs rounded text-indigo-600 dark:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-950">编辑</button>
                  <button onClick={() => handleDelete(acc.id)}
                    className="px-2 py-1 text-xs rounded text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-950">删除</button>
                </div>
              </div>
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 text-xs text-gray-500 dark:text-gray-400">
                <div>套餐: <span className="text-gray-700 dark:text-gray-300">{acc.plan_title || '-'}</span></div>
                <div>额度: <span className={`font-medium ${acc.remaining_count === 0 ? 'text-red-500' : 'text-gray-700 dark:text-gray-300'}`}>
                  {acc.remaining_count === -1 ? '未知' : acc.remaining_count}
                </span></div>
                <div>最后使用: <span className="text-gray-700 dark:text-gray-300">{timeAgo(acc.last_used_at)}</span></div>
              </div>
              {acc.last_error && (
                <p className="mt-1 text-xs text-red-500 truncate">错误: {acc.last_error}</p>
              )}
            </GlassCard>
          ))}
        </div>
      )}
    </div>
  )
}
