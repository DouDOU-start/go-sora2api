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

  // 计算比率
  const accountHealthRate = stats.total_accounts > 0
    ? Math.round((stats.active_accounts / stats.total_accounts) * 100) : 0
  const taskSuccessRate = stats.total_tasks > 0
    ? Math.round((stats.completed_tasks / stats.total_tasks) * 100) : 0
  const characterReadyRate = stats.total_characters > 0
    ? Math.round((stats.ready_characters / stats.total_characters) * 100) : 0

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
          实时监控账号、任务与角色状态
        </p>
      </motion.div>

      {/* 概览指标 — 三个环形卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
        <RingCard
          title="账号健康率"
          rate={accountHealthRate}
          color="var(--success)"
          softColor="var(--success-soft)"
          subtitle={`${stats.active_accounts} / ${stats.total_accounts} 活跃`}
          delay={0}
        />
        <RingCard
          title="任务成功率"
          rate={taskSuccessRate}
          color="var(--info)"
          softColor="var(--info-soft)"
          subtitle={`${stats.completed_tasks} / ${stats.total_tasks} 完成`}
          delay={1}
        />
        <RingCard
          title="角色就绪率"
          rate={characterReadyRate}
          color="var(--accent)"
          softColor="var(--accent-soft)"
          subtitle={`${stats.ready_characters} / ${stats.total_characters} 就绪`}
          delay={2}
        />
      </div>

      {/* 账号状态 */}
      <div className="mb-6">
        <h3 className="text-xs font-semibold uppercase tracking-wider mb-3" style={{ color: 'var(--text-tertiary)' }}>
          账号状态
        </h3>
        <GlassCard delay={3} className="p-5">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6 mb-5">
            <StatItem label="总账号" value={stats.total_accounts} icon={UsersIcon} color="var(--info)" />
            <StatItem label="活跃" value={stats.active_accounts} icon={CheckCircleIcon} color="var(--success)" />
            <StatItem label="Token 过期" value={stats.expired_accounts} icon={AlertIcon} color="var(--danger)" />
            <StatItem label="额度耗尽" value={stats.exhausted_accounts} icon={BanIcon} color="var(--warning)" />
          </div>
          {/* 横向比例条 */}
          <SegmentBar
            segments={[
              { value: stats.active_accounts, color: 'var(--success)', label: '活跃' },
              { value: stats.expired_accounts, color: 'var(--danger)', label: '过期' },
              { value: stats.exhausted_accounts, color: 'var(--warning)', label: '耗尽' },
            ]}
            total={stats.total_accounts}
          />
        </GlassCard>
      </div>

      {/* 任务状态 */}
      <div className="mb-6">
        <h3 className="text-xs font-semibold uppercase tracking-wider mb-3" style={{ color: 'var(--text-tertiary)' }}>
          任务状态
        </h3>
        <GlassCard delay={4} className="p-5">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6 mb-5">
            <StatItem label="总任务" value={stats.total_tasks} icon={ClipboardIcon} color="var(--info)" />
            <StatItem label="进行中" value={stats.pending_tasks} icon={ClockIcon} color="var(--warning)" />
            <StatItem label="已完成" value={stats.completed_tasks} icon={SparkleIcon} color="var(--success)" />
            <StatItem label="失败" value={stats.failed_tasks} icon={XCircleIcon} color="var(--danger)" />
          </div>
          <SegmentBar
            segments={[
              { value: stats.completed_tasks, color: 'var(--success)', label: '完成' },
              { value: stats.pending_tasks, color: 'var(--warning)', label: '进行中' },
              { value: stats.failed_tasks, color: 'var(--danger)', label: '失败' },
            ]}
            total={stats.total_tasks}
          />
        </GlassCard>
      </div>

      {/* 角色状态 */}
      <div>
        <h3 className="text-xs font-semibold uppercase tracking-wider mb-3" style={{ color: 'var(--text-tertiary)' }}>
          角色状态
        </h3>
        <GlassCard delay={5} className="p-5">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6 mb-5">
            <StatItem label="总角色" value={stats.total_characters} icon={CharacterTotalIcon} color="var(--info)" />
            <StatItem label="就绪" value={stats.ready_characters} icon={CheckCircleIcon} color="var(--success)" />
            <StatItem label="处理中" value={stats.processing_characters} icon={ClockIcon} color="var(--warning)" />
            <StatItem label="失败" value={stats.failed_characters} icon={XCircleIcon} color="var(--danger)" />
          </div>
          <SegmentBar
            segments={[
              { value: stats.ready_characters, color: 'var(--success)', label: '就绪' },
              { value: stats.processing_characters, color: 'var(--warning)', label: '处理中' },
              { value: stats.failed_characters, color: 'var(--danger)', label: '失败' },
            ]}
            total={stats.total_characters}
          />
        </GlassCard>
      </div>
    </div>
  )
}

/* ═══════════════════════════════════════════
   子组件
   ═══════════════════════════════════════════ */

/* ── 环形进度卡片 ── */
function RingCard({ title, rate, color, softColor, subtitle, delay }: {
  title: string
  rate: number
  color: string
  softColor: string
  subtitle: string
  delay: number
}) {
  const radius = 40
  const stroke = 6
  const circumference = 2 * Math.PI * radius
  const progress = circumference - (rate / 100) * circumference

  return (
    <GlassCard hover delay={delay} className="p-5 flex items-center gap-5">
      <div className="relative flex-shrink-0">
        <svg width="96" height="96" viewBox="0 0 96 96">
          {/* 背景环 */}
          <circle
            cx="48" cy="48" r={radius}
            fill="none" stroke={softColor} strokeWidth={stroke}
          />
          {/* 进度环 */}
          <motion.circle
            cx="48" cy="48" r={radius}
            fill="none" stroke={color} strokeWidth={stroke}
            strokeLinecap="round"
            strokeDasharray={circumference}
            strokeDashoffset={circumference}
            animate={{ strokeDashoffset: progress }}
            transition={{ duration: 1.2, delay: delay * 0.15, ease: [0.16, 1, 0.3, 1] }}
            style={{ transformOrigin: '48px 48px', transform: 'rotate(-90deg)' }}
          />
        </svg>
        <div className="absolute inset-0 flex items-center justify-center">
          <span className="text-xl font-bold tabular-nums" style={{ color }}>{rate}%</span>
        </div>
      </div>
      <div className="min-w-0">
        <p className="text-sm font-semibold" style={{ color: 'var(--text-primary)' }}>{title}</p>
        <p className="text-xs mt-1 truncate" style={{ color: 'var(--text-tertiary)' }}>{subtitle}</p>
      </div>
    </GlassCard>
  )
}

/* ── 统计数字项 ── */
function StatItem({ label, value, icon: Icon, color }: {
  label: string
  value: number
  icon: React.FC<{ color: string }>
  color: string
}) {
  return (
    <div className="flex items-center gap-3">
      <div
        className="w-9 h-9 rounded-xl flex items-center justify-center flex-shrink-0"
        style={{ background: color, opacity: 0.12 }}
      >
        <Icon color={color} />
      </div>
      <div>
        <p className="text-xs" style={{ color: 'var(--text-tertiary)' }}>{label}</p>
        <p className="text-lg font-semibold tabular-nums" style={{ color: 'var(--text-primary)' }}>{value}</p>
      </div>
    </div>
  )
}

/* ── 分段比例条 ── */
function SegmentBar({ segments, total }: {
  segments: { value: number; color: string; label: string }[]
  total: number
}) {
  if (total === 0) {
    return (
      <div>
        <div className="h-2.5 rounded-full overflow-hidden" style={{ background: 'var(--bg-inset)' }} />
        <p className="text-xs mt-2 text-center" style={{ color: 'var(--text-tertiary)' }}>暂无数据</p>
      </div>
    )
  }

  return (
    <div>
      {/* 条形图 */}
      <div className="h-2.5 rounded-full overflow-hidden flex" style={{ background: 'var(--bg-inset)' }}>
        {segments.map((seg) => {
          const pct = (seg.value / total) * 100
          if (pct === 0) return null
          return (
            <motion.div
              key={seg.label}
              initial={{ width: 0 }}
              animate={{ width: `${pct}%` }}
              transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
              style={{ background: seg.color }}
            />
          )
        })}
      </div>
      {/* 图例 */}
      <div className="flex flex-wrap gap-x-4 gap-y-1 mt-2.5">
        {segments.map((seg) => (
          <div key={seg.label} className="flex items-center gap-1.5">
            <div className="w-2 h-2 rounded-full flex-shrink-0" style={{ background: seg.color }} />
            <span className="text-xs" style={{ color: 'var(--text-tertiary)' }}>
              {seg.label} {seg.value}
              {total > 0 && <span className="ml-0.5">({Math.round((seg.value / total) * 100)}%)</span>}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}

/* ═══════════════════════════════════════════
   图标组件
   ═══════════════════════════════════════════ */
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
