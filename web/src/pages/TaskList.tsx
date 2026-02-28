import { useEffect, useRef, useState } from 'react'
import { listTasks } from '../api/task'
import type { SoraTask } from '../types/task'
import GlassCard from '../components/ui/GlassCard'
import StatusBadge from '../components/ui/StatusBadge'
import LoadingState from '../components/ui/LoadingState'
import { AnimatePresence, motion } from 'framer-motion'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

const statusFilters = [
  { label: '全部', value: '' },
  { label: '进行中', value: 'in_progress' },
  { label: '已完成', value: 'completed' },
  { label: '失败', value: 'failed' },
]

export default function TaskList() {
  const [tasks, setTasks] = useState<SoraTask[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)
  const [expandedTask, setExpandedTask] = useState<string | null>(null)
  const pageSize = 20
  const mountedRef = useRef(true)

  useEffect(() => {
    mountedRef.current = true
    const load = async () => {
      setLoading(true)
      try {
        const res = await listTasks({ status: status || undefined, page, page_size: pageSize })
        if (mountedRef.current) {
          setTasks(res.data.list ?? [])
          setTotal(res.data.total)
        }
      } catch { /* ignore */ }
      if (mountedRef.current) setLoading(false)
    }
    load()
    return () => { mountedRef.current = false }
  }, [status, page])

  // 自动刷新
  useEffect(() => {
    if (status === '' || status === 'in_progress') {
      const timer = setInterval(async () => {
        try {
          const res = await listTasks({ status: status || undefined, page, page_size: pageSize })
          setTasks(res.data.list ?? [])
          setTotal(res.data.total)
        } catch { /* ignore */ }
      }, 10000)
      return () => clearInterval(timer)
    }
  }, [status, page])

  const totalPages = Math.ceil(total / pageSize)

  return (
    <div>
      {/* 页头 */}
      <motion.div
        className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6"
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <div>
          <h1 className="text-2xl font-semibold tracking-tight" style={{ color: 'var(--text-primary)' }}>
            任务列表
          </h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--text-tertiary)' }}>
            共 {total} 条任务记录
          </p>
        </div>

        {/* 状态筛选 */}
        <div
          className="flex items-center gap-0.5 p-1 rounded-xl self-start"
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
      </motion.div>

      {loading ? <LoadingState /> : tasks.length === 0 ? (
        <div className="text-center py-20" style={{ color: 'var(--text-tertiary)' }}>
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" className="mx-auto mb-3 opacity-40">
            <rect x="3" y="4" width="18" height="4" rx="1" />
            <rect x="3" y="10" width="18" height="4" rx="1" />
            <rect x="3" y="16" width="18" height="4" rx="1" />
          </svg>
          暂无任务记录
        </div>
      ) : (
        <div className="space-y-2">
          {tasks.map((task, i) => (
            <GlassCard key={task.id} delay={i} className="overflow-hidden">
              <div
                className="p-4 sm:p-5 cursor-pointer transition-colors"
                onClick={() => setExpandedTask(expandedTask === task.id ? null : task.id)}
                style={{}}
                onMouseEnter={(e) => e.currentTarget.style.background = 'var(--bg-surface-hover)'}
                onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
              >
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-3 min-w-0">
                    <StatusBadge status={task.status} />
                    <div className="min-w-0">
                      <p
                        className="text-sm font-medium truncate"
                        style={{ fontFamily: 'var(--font-mono)', color: 'var(--text-primary)' }}
                      >
                        {task.id}
                      </p>
                      <p className="text-xs mt-0.5" style={{ color: 'var(--text-tertiary)' }}>
                        {task.model} · {task.type}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3 flex-shrink-0">
                    {task.status === 'in_progress' && (
                      <div className="flex items-center gap-2">
                        <div
                          className="w-20 h-1.5 rounded-full overflow-hidden"
                          style={{ background: 'var(--bg-inset)' }}
                        >
                          <motion.div
                            className="h-full rounded-full"
                            style={{ background: 'var(--accent)' }}
                            initial={{ width: 0 }}
                            animate={{ width: `${task.progress}%` }}
                            transition={{ duration: 0.6, ease: 'easeOut' }}
                          />
                        </div>
                        <span className="text-xs tabular-nums" style={{ color: 'var(--text-tertiary)' }}>
                          {task.progress}%
                        </span>
                      </div>
                    )}
                    <span className="text-xs hidden sm:inline" style={{ color: 'var(--text-tertiary)' }}>
                      {formatDistanceToNow(new Date(task.created_at), { addSuffix: true, locale: zhCN })}
                    </span>
                    <svg
                      width="16"
                      height="16"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="2"
                      strokeLinecap="round"
                      className="transition-transform duration-200"
                      style={{
                        color: 'var(--text-tertiary)',
                        transform: expandedTask === task.id ? 'rotate(180deg)' : 'rotate(0)',
                      }}
                    >
                      <polyline points="6 9 12 15 18 9" />
                    </svg>
                  </div>
                </div>
              </div>

              {/* 展开详情 */}
              <AnimatePresence>
                {expandedTask === task.id && (
                  <motion.div
                    initial={{ height: 0, opacity: 0 }}
                    animate={{ height: 'auto', opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    transition={{ duration: 0.2 }}
                    className="overflow-hidden"
                  >
                    <div
                      className="px-4 sm:px-5 py-4 text-sm space-y-3"
                      style={{ borderTop: '1px solid var(--border-subtle)' }}
                    >
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                        <InfoRow label="Sora ID" value={task.sora_task_id} mono />
                        <InfoRow label="账号 ID" value={String(task.account_id)} />
                        <InfoRow label="类型" value={task.type} />
                        <InfoRow
                          label="创建时间"
                          value={formatDistanceToNow(new Date(task.created_at), { addSuffix: true, locale: zhCN })}
                        />
                      </div>

                      {task.prompt && (
                        <div>
                          <p className="text-xs font-medium mb-1.5" style={{ color: 'var(--text-tertiary)' }}>Prompt</p>
                          <div
                            className="text-sm p-3 rounded-xl whitespace-pre-wrap break-all leading-relaxed"
                            style={{
                              background: 'var(--bg-inset)',
                              color: 'var(--text-secondary)',
                              border: '1px solid var(--border-subtle)',
                            }}
                          >
                            {task.prompt}
                          </div>
                        </div>
                      )}

                      {task.error_message && (
                        <div
                          className="px-3 py-2.5 rounded-xl text-xs"
                          style={{ background: 'var(--danger-soft)', color: 'var(--danger)' }}
                        >
                          {task.error_message}
                        </div>
                      )}
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>
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
            style={{
              background: 'var(--bg-surface)',
              color: 'var(--text-secondary)',
              border: '1px solid var(--border-default)',
            }}
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
            style={{
              background: 'var(--bg-surface)',
              color: 'var(--text-secondary)',
              border: '1px solid var(--border-default)',
            }}
          >
            下一页
          </button>
        </div>
      )}
    </div>
  )
}

function InfoRow({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex items-center gap-2 text-[13px]">
      <span style={{ color: 'var(--text-tertiary)' }}>{label}</span>
      <span
        className="truncate"
        style={{
          color: 'var(--text-secondary)',
          fontFamily: mono ? 'var(--font-mono)' : undefined,
          fontSize: mono ? '12px' : undefined,
        }}
      >
        {value}
      </span>
    </div>
  )
}
