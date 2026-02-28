import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { getTask, downloadTaskContent } from '../api/task'
import type { SoraTask } from '../types/task'
import GlassCard from '../components/ui/GlassCard'
import StatusBadge from '../components/ui/StatusBadge'
import LoadingState from '../components/ui/LoadingState'
import { toast } from '../components/ui/Toast'
import { motion } from 'framer-motion'
import { format } from 'date-fns'
import { zhCN } from 'date-fns/locale'

export default function TaskDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [task, setTask] = useState<SoraTask | null>(null)
  const [loading, setLoading] = useState(true)
  const [downloading, setDownloading] = useState(false)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [videoUrl, setVideoUrl] = useState<string | null>(null)

  useEffect(() => {
    if (!id) return
    let timer: ReturnType<typeof setInterval> | null = null

    const load = async () => {
      try {
        const res = await getTask(id)
        setTask(res.data)
        // 进行中的任务自动轮询
        if (res.data.status === 'queued' || res.data.status === 'in_progress') {
          timer = setInterval(async () => {
            try {
              const r = await getTask(id)
              setTask(r.data)
              if (r.data.status === 'completed' || r.data.status === 'failed') {
                if (timer) clearInterval(timer)
              }
            } catch { /* ignore */ }
          }, 5000)
        }
      } catch {
        toast.error('任务不存在')
        navigate('/tasks')
      }
      setLoading(false)
    }
    load()
    return () => { if (timer) clearInterval(timer) }
  }, [id, navigate])

  const handleDownload = async () => {
    if (!task) return
    setDownloading(true)
    try {
      const res = await downloadTaskContent(task.id)
      const blob = res.data
      const contentType = res.headers['content-type'] || ''
      const isImage = task.type === 'image' || contentType.startsWith('image/')

      if (isImage) {
        setPreviewUrl(URL.createObjectURL(blob))
      } else {
        const url = URL.createObjectURL(blob)
        setVideoUrl(url)
      }
      toast.success(isImage ? '加载完成' : '加载完成')
    } catch {
      toast.error('下载失败')
    }
    setDownloading(false)
  }

  const handleSaveFile = () => {
    if (!task) return
    const url = task.type === 'image' ? previewUrl : videoUrl
    if (!url) return
    const a = document.createElement('a')
    a.href = url
    a.download = `${task.id}.${task.type === 'image' ? 'png' : 'mp4'}`
    a.click()
  }

  if (loading) return <LoadingState />
  if (!task) return null

  const isCompleted = task.status === 'completed'
  const hasPreview = previewUrl || videoUrl

  return (
    <div>
      {/* 页头 */}
      <motion.div
        className="flex items-center gap-3 mb-6"
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <button
          onClick={() => navigate('/tasks')}
          className="p-2 rounded-xl transition-colors cursor-pointer flex-shrink-0"
          style={{ color: 'var(--text-tertiary)' }}
          onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--bg-inset)'; e.currentTarget.style.color = 'var(--text-primary)' }}
          onMouseLeave={(e) => { e.currentTarget.style.background = 'transparent'; e.currentTarget.style.color = 'var(--text-tertiary)' }}
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="15 18 9 12 15 6" />
          </svg>
        </button>
        <div className="min-w-0">
          <h1 className="text-2xl font-semibold tracking-tight truncate" style={{ color: 'var(--text-primary)' }}>
            {task.id}
          </h1>
          <div className="flex items-center gap-2 mt-0.5">
            <StatusBadge status={task.status} />
            <span className="text-xs" style={{ color: 'var(--text-tertiary)' }}>
              {task.model} · {task.type}
            </span>
          </div>
        </div>
      </motion.div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-5">
        {/* 左侧：预览区 */}
        <div className="lg:col-span-2">
          <GlassCard delay={0} className="overflow-hidden">
            <div className="p-5">
              <h3 className="text-xs font-semibold uppercase tracking-wider mb-3" style={{ color: 'var(--text-tertiary)' }}>
                预览
              </h3>

              {hasPreview ? (
                <div className="space-y-3">
                  <div className="rounded-xl overflow-hidden" style={{ border: '1px solid var(--border-default)', background: 'var(--bg-inset)' }}>
                    {previewUrl && (
                      <img
                        src={previewUrl}
                        alt="生成的图片"
                        className="w-full h-auto"
                        style={{ display: 'block' }}
                      />
                    )}
                    {videoUrl && (
                      <video
                        src={videoUrl}
                        controls
                        className="w-full h-auto"
                        style={{ display: 'block' }}
                      />
                    )}
                  </div>
                  <button
                    onClick={handleSaveFile}
                    className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-colors cursor-pointer"
                    style={{ background: 'var(--accent-soft)', color: 'var(--accent)' }}
                    onMouseEnter={(e) => { e.currentTarget.style.background = 'var(--accent)'; e.currentTarget.style.color = '#fff' }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'var(--accent-soft)'; e.currentTarget.style.color = 'var(--accent)' }}
                  >
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4" />
                      <polyline points="7 10 12 15 17 10" />
                      <line x1="12" y1="15" x2="12" y2="3" />
                    </svg>
                    保存到本地
                  </button>
                </div>
              ) : isCompleted ? (
                <div
                  className="flex flex-col items-center justify-center py-12 rounded-xl"
                  style={{ background: 'var(--bg-inset)', border: '1px dashed var(--border-default)' }}
                >
                  <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" className="mb-3" style={{ color: 'var(--text-tertiary)', opacity: 0.5 }}>
                    {task.type === 'image' ? (
                      <>
                        <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
                        <circle cx="8.5" cy="8.5" r="1.5" />
                        <polyline points="21 15 16 10 5 21" />
                      </>
                    ) : (
                      <>
                        <polygon points="5 3 19 12 5 21 5 3" />
                      </>
                    )}
                  </svg>
                  <p className="text-sm mb-3" style={{ color: 'var(--text-tertiary)' }}>
                    点击加载{task.type === 'image' ? '图片' : '视频'}预览
                  </p>
                  <button
                    onClick={handleDownload}
                    disabled={downloading}
                    className="inline-flex items-center gap-1.5 px-4 py-2 rounded-xl text-sm font-medium text-white transition-all cursor-pointer disabled:opacity-50"
                    style={{ background: 'var(--accent)' }}
                    onMouseEnter={(e) => { if (!downloading) e.currentTarget.style.background = 'var(--accent-hover)' }}
                    onMouseLeave={(e) => { e.currentTarget.style.background = 'var(--accent)' }}
                  >
                    {downloading ? (
                      <>
                        <svg className="w-4 h-4" style={{ animation: 'spin 0.6s linear infinite' }} viewBox="0 0 24 24" fill="none">
                          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                        </svg>
                        加载中...
                      </>
                    ) : (
                      <>
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                          <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4" />
                          <polyline points="7 10 12 15 17 10" />
                          <line x1="12" y1="15" x2="12" y2="3" />
                        </svg>
                        加载预览
                      </>
                    )}
                  </button>
                </div>
              ) : task.status === 'in_progress' ? (
                <div
                  className="flex flex-col items-center justify-center py-12 rounded-xl"
                  style={{ background: 'var(--bg-inset)', border: '1px dashed var(--border-default)' }}
                >
                  <div className="flex items-center gap-3 mb-2">
                    <div
                      className="w-32 h-2 rounded-full overflow-hidden"
                      style={{ background: 'var(--border-default)' }}
                    >
                      <motion.div
                        className="h-full rounded-full"
                        style={{ background: 'var(--accent)' }}
                        initial={{ width: 0 }}
                        animate={{ width: `${task.progress}%` }}
                        transition={{ duration: 0.6, ease: 'easeOut' }}
                      />
                    </div>
                    <span className="text-sm font-medium tabular-nums" style={{ color: 'var(--text-secondary)' }}>
                      {task.progress}%
                    </span>
                  </div>
                  <p className="text-xs" style={{ color: 'var(--text-tertiary)' }}>
                    生成中，请稍候...
                  </p>
                </div>
              ) : task.status === 'failed' ? (
                <div
                  className="py-12 text-center rounded-xl"
                  style={{ background: 'var(--danger-soft)', border: '1px dashed var(--danger)' }}
                >
                  <p className="text-sm" style={{ color: 'var(--danger)' }}>
                    {task.error_message || '任务失败'}
                  </p>
                </div>
              ) : (
                <div
                  className="py-12 text-center rounded-xl"
                  style={{ background: 'var(--bg-inset)', border: '1px dashed var(--border-default)' }}
                >
                  <p className="text-sm" style={{ color: 'var(--text-tertiary)' }}>排队中...</p>
                </div>
              )}
            </div>
          </GlassCard>
        </div>

        {/* 右侧：任务信息 */}
        <div className="space-y-5">
          <GlassCard delay={1} className="overflow-hidden">
            <div className="p-5 space-y-3">
              <h3 className="text-xs font-semibold uppercase tracking-wider" style={{ color: 'var(--text-tertiary)' }}>
                任务信息
              </h3>
              <InfoRow label="任务 ID" value={task.id} mono />
              <InfoRow label="Sora ID" value={task.sora_task_id} mono />
              <InfoRow label="账号 ID" value={String(task.account_id)} />
              <InfoRow label="类型" value={task.type} />
              <InfoRow label="模型" value={task.model} />
              <InfoRow
                label="创建时间"
                value={format(new Date(task.created_at), 'yyyy-MM-dd HH:mm:ss', { locale: zhCN })}
              />
              {task.completed_at && (
                <InfoRow
                  label="完成时间"
                  value={format(new Date(task.completed_at), 'yyyy-MM-dd HH:mm:ss', { locale: zhCN })}
                />
              )}
            </div>
          </GlassCard>

          {task.prompt && (
            <GlassCard delay={2} className="overflow-hidden">
              <div className="p-5">
                <h3 className="text-xs font-semibold uppercase tracking-wider mb-3" style={{ color: 'var(--text-tertiary)' }}>
                  Prompt
                </h3>
                <div
                  className="text-sm p-3 rounded-xl whitespace-pre-wrap break-all leading-relaxed max-h-80 overflow-y-auto"
                  style={{
                    background: 'var(--bg-inset)',
                    color: 'var(--text-secondary)',
                    border: '1px solid var(--border-subtle)',
                  }}
                >
                  {task.prompt}
                </div>
              </div>
            </GlassCard>
          )}

          {task.error_message && (
            <GlassCard delay={3} className="overflow-hidden">
              <div className="p-5">
                <h3 className="text-xs font-semibold uppercase tracking-wider mb-3" style={{ color: 'var(--danger)' }}>
                  错误信息
                </h3>
                <div
                  className="text-sm p-3 rounded-xl"
                  style={{ background: 'var(--danger-soft)', color: 'var(--danger)' }}
                >
                  {task.error_message}
                </div>
              </div>
            </GlassCard>
          )}
        </div>
      </div>
    </div>
  )
}

function InfoRow({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex items-start gap-2 text-[13px]">
      <span className="flex-shrink-0" style={{ color: 'var(--text-tertiary)', minWidth: '5em' }}>{label}</span>
      <span
        className={mono ? 'break-all' : 'truncate'}
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
