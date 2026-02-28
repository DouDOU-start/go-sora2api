import { useEffect, useRef, useState } from 'react'
import { listTasks } from '../api/task'
import type { SoraTask } from '../types/task'
import GlassCard from '../components/ui/GlassCard'
import StatusBadge from '../components/ui/StatusBadge'
import LoadingState from '../components/ui/LoadingState'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

const statusFilters: { label: string; value: string }[] = [
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

  // 自动刷新进行中的任务
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
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">任务列表</h2>
        <span className="text-sm text-gray-500 dark:text-gray-400">共 {total} 条</span>
      </div>

      {/* 状态筛选 */}
      <div className="flex items-center gap-2 mb-4">
        {statusFilters.map((f) => (
          <button
            key={f.value}
            onClick={() => { setStatus(f.value); setPage(1) }}
            className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${
              status === f.value
                ? 'bg-indigo-50 dark:bg-indigo-950 text-indigo-600 dark:text-indigo-400'
                : 'text-gray-500 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800'
            }`}
          >
            {f.label}
          </button>
        ))}
      </div>

      {loading ? <LoadingState /> : tasks.length === 0 ? (
        <div className="text-center text-gray-400 py-16">暂无任务</div>
      ) : (
        <div className="space-y-3">
          {tasks.map((task) => (
            <GlassCard key={task.id} className="overflow-hidden">
              <div
                className="p-4 cursor-pointer hover:bg-gray-50/50 dark:hover:bg-gray-800/50 transition-colors"
                onClick={() => setExpandedTask(expandedTask === task.id ? null : task.id)}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <StatusBadge status={task.status} />
                    <div>
                      <p className="text-sm font-mono text-gray-900 dark:text-gray-100">{task.id}</p>
                      <p className="text-xs text-gray-500 dark:text-gray-400">{task.model}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    {task.status === 'in_progress' && (
                      <div className="flex items-center gap-2 mb-1">
                        <div className="w-24 h-1.5 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden">
                          <div className="h-full rounded-full bg-indigo-500 transition-all" style={{ width: `${task.progress}%` }} />
                        </div>
                        <span className="text-xs text-gray-500">{task.progress}%</span>
                      </div>
                    )}
                    <p className="text-xs text-gray-400">
                      {formatDistanceToNow(new Date(task.created_at), { addSuffix: true, locale: zhCN })}
                    </p>
                  </div>
                </div>
              </div>

              {/* 展开详情 */}
              {expandedTask === task.id && (
                <div className="border-t border-gray-200 dark:border-gray-800 px-4 py-3 text-sm">
                  <div className="grid grid-cols-2 gap-2 text-gray-600 dark:text-gray-400">
                    <div>Sora ID: <span className="font-mono text-gray-800 dark:text-gray-200">{task.sora_task_id}</span></div>
                    <div>账号 ID: <span className="text-gray-800 dark:text-gray-200">{task.account_id}</span></div>
                    <div>类型: <span className="text-gray-800 dark:text-gray-200">{task.type}</span></div>
                    <div>状态: <StatusBadge status={task.status} /></div>
                  </div>
                  {task.prompt && (
                    <div className="mt-2">
                      <p className="text-xs text-gray-500 mb-1">Prompt:</p>
                      <p className="text-sm text-gray-800 dark:text-gray-200 bg-gray-50 dark:bg-gray-800/50 rounded-lg p-2 whitespace-pre-wrap break-all">{task.prompt}</p>
                    </div>
                  )}
                  {task.error_message && (
                    <div className="mt-2 p-2 rounded-lg bg-red-50 dark:bg-red-950/30 text-red-600 dark:text-red-400 text-xs">
                      错误: {task.error_message}
                    </div>
                  )}
                </div>
              )}
            </GlassCard>
          ))}
        </div>
      )}

      {/* 分页 */}
      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-2 mt-6">
          <button
            onClick={() => setPage(Math.max(1, page - 1))}
            disabled={page === 1}
            className="px-3 py-1.5 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-30"
          >
            上一页
          </button>
          <span className="text-sm text-gray-500">{page} / {totalPages}</span>
          <button
            onClick={() => setPage(Math.min(totalPages, page + 1))}
            disabled={page === totalPages}
            className="px-3 py-1.5 rounded-lg text-sm text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-30"
          >
            下一页
          </button>
        </div>
      )}
    </div>
  )
}
