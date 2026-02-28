import { useEffect, useRef, useState } from 'react'
import { listCharacters, deleteCharacter, toggleCharacterVisibility, getCharacterImageUrl } from '../api/character'
import type { SoraCharacter } from '../types/character'
import GlassCard from '../components/ui/GlassCard'
import StatusBadge from '../components/ui/StatusBadge'
import LoadingState from '../components/ui/LoadingState'
import ConfirmDialog from '../components/ui/ConfirmDialog'
import { toast } from '../components/ui/Toast'
import { getErrorMessage } from '../api/client'
import { motion, AnimatePresence } from 'framer-motion'
import { formatDistanceToNow } from 'date-fns'
import { zhCN } from 'date-fns/locale'

const statusFilters = [
  { label: '全部', value: '' },
  { label: '就绪', value: 'ready' },
  { label: '处理中', value: 'processing' },
  { label: '失败', value: 'failed' },
]

// 角色状态映射到 StatusBadge 可识别的值
const statusMap: Record<string, string> = {
  processing: 'in_progress',
  ready: 'completed',
  failed: 'failed',
}

const statusLabel: Record<string, string> = {
  processing: '处理中',
  ready: '就绪',
  failed: '失败',
}

export default function CharacterList() {
  const [characters, setCharacters] = useState<SoraCharacter[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)
  const pageSize = 20
  const mountedRef = useRef(true)

  // 详情弹窗
  const [selectedChar, setSelectedChar] = useState<SoraCharacter | null>(null)

  // 删除确认
  const [deleteTarget, setDeleteTarget] = useState<SoraCharacter | null>(null)
  const [deleting, setDeleting] = useState(false)

  // 可见性切换
  const [toggling, setToggling] = useState(false)

  // 复制反馈
  const [copied, setCopied] = useState(false)

  const load = async () => {
    try {
      const res = await listCharacters({ status: status || undefined, page, page_size: pageSize })
      if (mountedRef.current) {
        setCharacters(res.data.list ?? [])
        setTotal(res.data.total)
      }
    } catch { /* ignore */ }
  }

  useEffect(() => {
    mountedRef.current = true
    setLoading(true)
    load().finally(() => { if (mountedRef.current) setLoading(false) })
    return () => { mountedRef.current = false }
  }, [status, page])

  // 自动刷新（有 processing 角色或查看全部时）
  useEffect(() => {
    if (status === '' || status === 'processing') {
      const timer = setInterval(load, 10000)
      return () => clearInterval(timer)
    }
  }, [status, page])

  const totalPages = Math.ceil(total / pageSize)

  const handleDelete = async () => {
    if (!deleteTarget) return
    setDeleting(true)
    try {
      await deleteCharacter(deleteTarget.id)
      toast.success('角色已删除')
      setDeleteTarget(null)
      // 如果删除的是当前查看的角色，关闭详情
      if (selectedChar?.id === deleteTarget.id) setSelectedChar(null)
      load()
    } catch (err) {
      toast.error(getErrorMessage(err, '删除失败'))
    } finally {
      setDeleting(false)
    }
  }

  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  const handleToggleVisibility = async (char: SoraCharacter) => {
    setToggling(true)
    try {
      const res = await toggleCharacterVisibility(char.id)
      toast.success(res.data.message)
      // 更新本地状态
      const updated = { ...char, is_public: res.data.is_public }
      setCharacters((prev) => prev.map((c) => c.id === char.id ? updated : c))
      if (selectedChar?.id === char.id) setSelectedChar(updated)
    } catch (err) {
      toast.error(getErrorMessage(err, '切换可见性失败'))
    } finally {
      setToggling(false)
    }
  }

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
            角色库
          </h1>
          <p className="text-sm mt-0.5" style={{ color: 'var(--text-tertiary)' }}>
            共 {total} 个角色
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

      {loading ? <LoadingState /> : characters.length === 0 ? (
        <div className="text-center py-20" style={{ color: 'var(--text-tertiary)' }}>
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" className="mx-auto mb-3 opacity-40">
            <circle cx="12" cy="8" r="4" />
            <path d="M6 21v-1a6 6 0 0112 0v1" />
          </svg>
          暂无角色
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {characters.map((char, i) => (
            <GlassCard key={char.id} delay={i} hover className="overflow-hidden">
              <div
                className="cursor-pointer"
                onClick={() => setSelectedChar(char)}
              >
                {/* 角色头像 */}
                <div
                  className="aspect-square w-full overflow-hidden flex items-center justify-center"
                  style={{ background: 'var(--bg-inset)' }}
                >
                  {(char.status === 'ready' || char.profile_url) ? (
                    <img
                      src={char.status === 'ready' ? getCharacterImageUrl(char.id) : char.profile_url}
                      alt={char.display_name || char.username}
                      className="w-full h-full object-cover"
                      loading="lazy"
                    />
                  ) : (
                    <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1" style={{ color: 'var(--text-tertiary)', opacity: 0.3 }}>
                      <circle cx="12" cy="8" r="4" />
                      <path d="M6 21v-1a6 6 0 0112 0v1" />
                    </svg>
                  )}
                </div>

                {/* 角色信息 */}
                <div className="p-3.5">
                  <div className="flex items-center justify-between gap-2 mb-1.5">
                    <h3
                      className="text-sm font-semibold truncate"
                      style={{ color: 'var(--text-primary)' }}
                    >
                      {char.display_name || '未命名'}
                    </h3>
                    <StatusBadge status={statusMap[char.status] || char.status} />
                  </div>
                  <div className="flex items-center gap-1.5 mb-1">
                    {char.username && (
                      <p className="text-xs truncate" style={{ color: 'var(--text-tertiary)' }}>
                        @{char.username}
                      </p>
                    )}
                    {char.status === 'ready' && (
                      <span
                        className="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium flex-shrink-0"
                        style={{
                          background: char.is_public ? 'var(--success-soft)' : 'var(--bg-inset)',
                          color: char.is_public ? 'var(--success)' : 'var(--text-tertiary)',
                        }}
                      >
                        {char.is_public ? '公开' : '私密'}
                      </span>
                    )}
                  </div>
                  <p className="text-xs" style={{ color: 'var(--text-tertiary)' }}>
                    {formatDistanceToNow(new Date(char.created_at), { addSuffix: true, locale: zhCN })}
                    {char.account_email && ` · ${char.account_email}`}
                  </p>
                </div>
              </div>
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

      {/* 角色详情弹窗 */}
      <AnimatePresence>
        {selectedChar && (
          <>
            <motion.div
              className="fixed inset-0 z-[90]"
              style={{ background: 'rgba(0,0,0,0.5)' }}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              onClick={() => setSelectedChar(null)}
            />
            <div className="fixed inset-0 z-[91] flex items-center justify-center p-4" onClick={() => setSelectedChar(null)}>
              <motion.div
                className="w-full max-w-[480px] rounded-2xl overflow-hidden"
                style={{
                  background: 'var(--bg-elevated)',
                  border: '1px solid var(--border-default)',
                  boxShadow: 'var(--shadow-lg)',
                }}
                initial={{ opacity: 0, scale: 0.95, y: 8 }}
                animate={{ opacity: 1, scale: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.95, y: 8 }}
                transition={{ duration: 0.2, ease: [0.16, 1, 0.3, 1] }}
                onClick={(e) => e.stopPropagation()}
              >
                {/* 头像区域 */}
                {(selectedChar.status === 'ready' || selectedChar.profile_url) && (
                  <div
                    className="w-full aspect-video overflow-hidden flex items-center justify-center"
                    style={{ background: 'var(--bg-inset)' }}
                  >
                    <img
                      src={selectedChar.status === 'ready' ? getCharacterImageUrl(selectedChar.id) : selectedChar.profile_url}
                      alt={selectedChar.display_name}
                      className="w-full h-full object-cover"
                    />
                  </div>
                )}

                {/* 详情内容 */}
                <div className="p-6">
                  <div className="flex items-center justify-between gap-3 mb-4">
                    <div className="min-w-0">
                      <h2 className="text-lg font-semibold truncate" style={{ color: 'var(--text-primary)' }}>
                        {selectedChar.display_name || '未命名'}
                      </h2>
                      {selectedChar.username && (
                        <p className="text-sm" style={{ color: 'var(--text-tertiary)' }}>@{selectedChar.username}</p>
                      )}
                    </div>
                    <StatusBadge status={statusMap[selectedChar.status] || selectedChar.status} />
                  </div>

                  <div className="space-y-3">
                    <DetailRow label="角色 ID" value={selectedChar.id} mono />
                    {selectedChar.character_id && (
                      <DetailRow label="Sora Character ID" value={selectedChar.character_id} mono copyable onCopy={handleCopy} copied={copied} />
                    )}
                    {selectedChar.account_email && (
                      <DetailRow label="关联账号" value={selectedChar.account_email} />
                    )}
                    <DetailRow label="状态" value={statusLabel[selectedChar.status] || selectedChar.status} />
                    {selectedChar.status === 'ready' && (
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-xs font-medium flex-shrink-0" style={{ color: 'var(--text-tertiary)' }}>可见性</span>
                        <button
                          onClick={() => handleToggleVisibility(selectedChar)}
                          disabled={toggling}
                          className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium transition-colors cursor-pointer disabled:opacity-50"
                          style={{
                            background: selectedChar.is_public ? 'var(--success-soft)' : 'var(--bg-inset)',
                            color: selectedChar.is_public ? 'var(--success)' : 'var(--text-tertiary)',
                          }}
                        >
                          {selectedChar.is_public ? (
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                              <circle cx="12" cy="12" r="10" /><line x1="2" y1="12" x2="22" y2="12" /><path d="M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z" />
                            </svg>
                          ) : (
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                              <rect x="3" y="11" width="18" height="11" rx="2" ry="2" /><path d="M7 11V7a5 5 0 0110 0v4" />
                            </svg>
                          )}
                          {toggling ? '切换中...' : selectedChar.is_public ? '公开 · 点击设为私密' : '私密 · 点击设为公开'}
                        </button>
                      </div>
                    )}
                    <DetailRow label="创建时间" value={new Date(selectedChar.created_at).toLocaleString('zh-CN')} />
                    {selectedChar.completed_at && (
                      <DetailRow label="完成时间" value={new Date(selectedChar.completed_at).toLocaleString('zh-CN')} />
                    )}
                    {selectedChar.error_message && (
                      <div>
                        <span className="text-xs font-medium" style={{ color: 'var(--text-tertiary)' }}>错误信息</span>
                        <p className="text-sm mt-0.5" style={{ color: 'var(--danger)' }}>{selectedChar.error_message}</p>
                      </div>
                    )}
                  </div>

                  {/* 操作按钮 */}
                  <div className="flex items-center justify-end gap-2 mt-6 pt-4" style={{ borderTop: '1px solid var(--border-default)' }}>
                    <button
                      onClick={() => setSelectedChar(null)}
                      className="px-4 py-2 rounded-xl text-sm font-medium transition-colors cursor-pointer"
                      style={{ color: 'var(--text-secondary)', background: 'var(--bg-inset)' }}
                    >
                      关闭
                    </button>
                    <button
                      onClick={() => { setDeleteTarget(selectedChar); }}
                      className="px-4 py-2 rounded-xl text-sm font-medium text-white transition-colors cursor-pointer"
                      style={{ background: 'var(--danger)' }}
                    >
                      删除
                    </button>
                  </div>
                </div>
              </motion.div>
            </div>
          </>
        )}
      </AnimatePresence>

      {/* 删除确认弹窗 */}
      <ConfirmDialog
        open={!!deleteTarget}
        title="删除角色"
        message={`确定要删除角色「${deleteTarget?.display_name || deleteTarget?.id}」吗？此操作不可撤销。`}
        confirmLabel={deleting ? '删除中...' : '删除'}
        danger
        onConfirm={handleDelete}
        onCancel={() => setDeleteTarget(null)}
      />
    </div>
  )
}

/* ── 详情行组件 ── */

function DetailRow({ label, value, mono, copyable, onCopy, copied }: {
  label: string
  value: string
  mono?: boolean
  copyable?: boolean
  onCopy?: (v: string) => void
  copied?: boolean
}) {
  return (
    <div className="flex items-start justify-between gap-3">
      <span className="text-xs font-medium flex-shrink-0 pt-0.5" style={{ color: 'var(--text-tertiary)' }}>{label}</span>
      <div className="flex items-center gap-1.5 min-w-0">
        <span
          className="text-sm text-right truncate"
          style={{
            color: 'var(--text-primary)',
            fontFamily: mono ? 'var(--font-mono)' : undefined,
          }}
        >
          {value}
        </span>
        {copyable && onCopy && (
          <button
            onClick={() => onCopy(value)}
            className="flex-shrink-0 p-1 rounded-md transition-colors cursor-pointer"
            style={{ color: copied ? 'var(--success)' : 'var(--text-tertiary)' }}
            title="复制"
          >
            {copied ? (
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
            ) : (
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                <path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" />
              </svg>
            )}
          </button>
        )}
      </div>
    </div>
  )
}
