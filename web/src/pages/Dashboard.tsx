import { useEffect, useRef, useState } from 'react'
import { getDashboard } from '../api/task'
import type { DashboardStats } from '../types/task'
import GlassCard from '../components/ui/GlassCard'
import LoadingState from '../components/ui/LoadingState'
import { motion } from 'framer-motion'

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
  if (!stats) return (
    <div className="text-center py-20" style={{ color: 'var(--text-tertiary)' }}>加载失败</div>
  )

  const accountCards = [
    { label: '总账号', value: stats.total_accounts, color: 'var(--info)', softColor: 'var(--info-soft)', icon: UsersIcon },
    { label: '活跃账号', value: stats.active_accounts, color: 'var(--success)', softColor: 'var(--success-soft)', icon: CheckCircleIcon },
    { label: 'Token 过期', value: stats.expired_accounts, color: 'var(--danger)', softColor: 'var(--danger-soft)', icon: AlertIcon },
    { label: '额度耗尽', value: stats.exhausted_accounts, color: 'var(--warning)', softColor: 'var(--warning-soft)', icon: BanIcon },
  ]

  const taskCards = [
    { label: '总任务', value: stats.total_tasks, color: 'var(--info)', softColor: 'var(--info-soft)', icon: ClipboardIcon },
    { label: '进行中', value: stats.pending_tasks, color: 'var(--warning)', softColor: 'var(--warning-soft)', icon: ClockIcon },
    { label: '已完成', value: stats.completed_tasks, color: 'var(--success)', softColor: 'var(--success-soft)', icon: SparkleIcon },
    { label: '失败', value: stats.failed_tasks, color: 'var(--danger)', softColor: 'var(--danger-soft)', icon: XCircleIcon },
  ]

  const characterCards = [
    { label: '总角色', value: stats.total_characters, color: 'var(--info)', softColor: 'var(--info-soft)', icon: CharacterTotalIcon },
    { label: '就绪', value: stats.ready_characters, color: 'var(--success)', softColor: 'var(--success-soft)', icon: CheckCircleIcon },
    { label: '处理中', value: stats.processing_characters, color: 'var(--warning)', softColor: 'var(--warning-soft)', icon: ClockIcon },
    { label: '失败', value: stats.failed_characters, color: 'var(--danger)', softColor: 'var(--danger-soft)', icon: XCircleIcon },
  ]

  return (
    <div>
      {/* 页头 */}
      <motion.div
        className="mb-8"
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4 }}
      >
        <h1 className="text-2xl font-semibold tracking-tight" style={{ color: 'var(--text-primary)' }}>
          系统概览
        </h1>
        <p className="text-sm mt-1" style={{ color: 'var(--text-tertiary)' }}>
          实时监控账号与任务状态
        </p>
      </motion.div>

      {/* 账号统计 */}
      <div className="mb-6">
        <h3 className="text-xs font-semibold uppercase tracking-wider mb-3" style={{ color: 'var(--text-tertiary)' }}>
          账号状态
        </h3>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
          {accountCards.map((card, i) => (
            <GlassCard key={card.label} hover delay={i} className="p-4 sm:p-5">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-xs font-medium mb-2" style={{ color: 'var(--text-tertiary)' }}>
                    {card.label}
                  </p>
                  <p className="text-2xl sm:text-3xl font-semibold tabular-nums" style={{ color: 'var(--text-primary)' }}>
                    {card.value}
                  </p>
                </div>
                <div
                  className="w-9 h-9 rounded-xl flex items-center justify-center flex-shrink-0"
                  style={{ background: card.softColor }}
                >
                  <card.icon color={card.color} />
                </div>
              </div>
            </GlassCard>
          ))}
        </div>
      </div>

      {/* 任务统计 */}
      <div className="mb-6">
        <h3 className="text-xs font-semibold uppercase tracking-wider mb-3" style={{ color: 'var(--text-tertiary)' }}>
          任务状态
        </h3>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
          {taskCards.map((card, i) => (
            <GlassCard key={card.label} hover delay={i + 4} className="p-4 sm:p-5">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-xs font-medium mb-2" style={{ color: 'var(--text-tertiary)' }}>
                    {card.label}
                  </p>
                  <p className="text-2xl sm:text-3xl font-semibold tabular-nums" style={{ color: 'var(--text-primary)' }}>
                    {card.value}
                  </p>
                </div>
                <div
                  className="w-9 h-9 rounded-xl flex items-center justify-center flex-shrink-0"
                  style={{ background: card.softColor }}
                >
                  <card.icon color={card.color} />
                </div>
              </div>
            </GlassCard>
          ))}
        </div>
      </div>

      {/* 角色统计 */}
      <div>
        <h3 className="text-xs font-semibold uppercase tracking-wider mb-3" style={{ color: 'var(--text-tertiary)' }}>
          角色状态
        </h3>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
          {characterCards.map((card, i) => (
            <GlassCard key={card.label} hover delay={i + 8} className="p-4 sm:p-5">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-xs font-medium mb-2" style={{ color: 'var(--text-tertiary)' }}>
                    {card.label}
                  </p>
                  <p className="text-2xl sm:text-3xl font-semibold tabular-nums" style={{ color: 'var(--text-primary)' }}>
                    {card.value}
                  </p>
                </div>
                <div
                  className="w-9 h-9 rounded-xl flex items-center justify-center flex-shrink-0"
                  style={{ background: card.softColor }}
                >
                  <card.icon color={card.color} />
                </div>
              </div>
            </GlassCard>
          ))}
        </div>
      </div>
    </div>
  )
}

/* ── 图标组件 ── */
function UsersIcon({ color }: { color: string }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="9" cy="7" r="4" /><path d="M3 21v-2a4 4 0 014-4h4a4 4 0 014 4v2" />
      <circle cx="17" cy="7" r="3" /><path d="M21 21v-2a3 3 0 00-2-2.83" />
    </svg>
  )
}

function CheckCircleIcon({ color }: { color: string }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" /><path d="M8 12l3 3 5-5" />
    </svg>
  )
}

function AlertIcon({ color }: { color: string }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z" />
      <line x1="12" y1="9" x2="12" y2="13" /><line x1="12" y1="17" x2="12.01" y2="17" />
    </svg>
  )
}

function BanIcon({ color }: { color: string }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" /><line x1="4.93" y1="4.93" x2="19.07" y2="19.07" />
    </svg>
  )
}

function ClipboardIcon({ color }: { color: string }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <rect x="8" y="2" width="8" height="4" rx="1" /><path d="M16 4h2a2 2 0 012 2v14a2 2 0 01-2 2H6a2 2 0 01-2-2V6a2 2 0 012-2h2" />
    </svg>
  )
}

function ClockIcon({ color }: { color: string }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" /><polyline points="12 6 12 12 16 14" />
    </svg>
  )
}

function SparkleIcon({ color }: { color: string }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 2l2.4 7.2L22 12l-7.6 2.8L12 22l-2.4-7.2L2 12l7.6-2.8L12 2z" />
    </svg>
  )
}

function XCircleIcon({ color }: { color: string }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" /><line x1="15" y1="9" x2="9" y2="15" /><line x1="9" y1="9" x2="15" y2="15" />
    </svg>
  )
}

function CharacterTotalIcon({ color }: { color: string }) {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="8" r="4" /><path d="M6 21v-1a6 6 0 0112 0v1" />
    </svg>
  )
}
