interface Props {
  status: string
  className?: string
}

const statusConfig: Record<string, { bg: string; color: string; dotColor: string; label: string }> = {
  active:          { bg: 'var(--success-soft)', color: 'var(--success)', dotColor: 'var(--success)', label: '正常' },
  token_expired:   { bg: 'var(--danger-soft)',  color: 'var(--danger)',  dotColor: 'var(--danger)',  label: 'Token 过期' },
  quota_exhausted: { bg: 'var(--warning-soft)', color: 'var(--warning)', dotColor: 'var(--warning)', label: '额度耗尽' },
  queued:          { bg: 'var(--info-soft)',    color: 'var(--info)',    dotColor: 'var(--info)',    label: '排队中' },
  in_progress:     { bg: 'var(--warning-soft)', color: 'var(--warning)', dotColor: 'var(--warning)', label: '进行中' },
  completed:       { bg: 'var(--success-soft)', color: 'var(--success)', dotColor: 'var(--success)', label: '已完成' },
  failed:          { bg: 'var(--danger-soft)',  color: 'var(--danger)',  dotColor: 'var(--danger)',  label: '失败' },
}

export default function StatusBadge({ status, className = '' }: Props) {
  const config = statusConfig[status] || {
    bg: 'var(--bg-inset)',
    color: 'var(--text-tertiary)',
    dotColor: 'var(--text-tertiary)',
    label: status,
  }

  return (
    <span
      className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${className}`}
      style={{ background: config.bg, color: config.color }}
    >
      <span
        className="relative flex h-1.5 w-1.5"
      >
        {status === 'in_progress' && (
          <span
            className="absolute inset-0 rounded-full opacity-75"
            style={{
              background: config.dotColor,
              animation: 'pulse-ring 1.5s ease-out infinite',
            }}
          />
        )}
        <span
          className="relative inline-flex rounded-full h-1.5 w-1.5"
          style={{ background: config.dotColor }}
        />
      </span>
      {config.label}
    </span>
  )
}
