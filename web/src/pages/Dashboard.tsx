import { useEffect, useRef, useState } from 'react'
import { getDashboard } from '../api/task'
import type { DashboardStats } from '../types/task'
import GlassCard from '../components/ui/GlassCard'
import LoadingState from '../components/ui/LoadingState'

export default function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const mountedRef = useRef(true)

  useEffect(() => {
    mountedRef.current = true
    const load = async () => {
      try {
        const res = await getDashboard()
        if (mountedRef.current) setStats(res.data)
      } catch { /* ignore */ }
      if (mountedRef.current) setLoading(false)
    }
    load()
    const timer = setInterval(load, 15000)
    return () => { mountedRef.current = false; clearInterval(timer) }
  }, [])

  if (loading) return <LoadingState />
  if (!stats) return <div className="text-center text-gray-500 py-16">加载失败</div>

  const cards = [
    { label: '总账号', value: stats.total_accounts, color: 'from-blue-500 to-blue-600', icon: '👥' },
    { label: '活跃账号', value: stats.active_accounts, color: 'from-green-500 to-green-600', icon: '✅' },
    { label: 'Token 过期', value: stats.expired_accounts, color: 'from-red-500 to-red-600', icon: '⚠️' },
    { label: '额度耗尽', value: stats.exhausted_accounts, color: 'from-gray-500 to-gray-600', icon: '🚫' },
    { label: '总任务', value: stats.total_tasks, color: 'from-indigo-500 to-indigo-600', icon: '📋' },
    { label: '进行中', value: stats.pending_tasks, color: 'from-yellow-500 to-yellow-600', icon: '⏳' },
    { label: '已完成', value: stats.completed_tasks, color: 'from-emerald-500 to-emerald-600', icon: '🎉' },
    { label: '失败', value: stats.failed_tasks, color: 'from-rose-500 to-rose-600', icon: '❌' },
  ]

  return (
    <div>
      <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100 mb-6">系统概览</h2>

      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
        {cards.map((card) => (
          <GlassCard key={card.label} hover className="p-4">
            <div className="flex items-center gap-3">
              <span className="text-2xl">{card.icon}</span>
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400">{card.label}</p>
                <p className={`text-2xl font-bold bg-gradient-to-r ${card.color} bg-clip-text text-transparent`}>
                  {card.value}
                </p>
              </div>
            </div>
          </GlassCard>
        ))}
      </div>
    </div>
  )
}
