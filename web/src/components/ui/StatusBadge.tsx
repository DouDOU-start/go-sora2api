interface Props {
  status: string
  className?: string
}

const statusConfig: Record<string, { bg: string; text: string; dot: string; label: string }> = {
  active: { bg: 'bg-green-100 dark:bg-green-900/30', text: 'text-green-700 dark:text-green-400', dot: 'bg-green-500', label: '正常' },
  token_expired: { bg: 'bg-red-100 dark:bg-red-900/30', text: 'text-red-700 dark:text-red-400', dot: 'bg-red-500', label: 'Token 过期' },
  quota_exhausted: { bg: 'bg-gray-100 dark:bg-gray-800/30', text: 'text-gray-700 dark:text-gray-400', dot: 'bg-gray-500', label: '额度耗尽' },
  queued: { bg: 'bg-blue-100 dark:bg-blue-900/30', text: 'text-blue-700 dark:text-blue-400', dot: 'bg-blue-500', label: '排队中' },
  in_progress: { bg: 'bg-yellow-100 dark:bg-yellow-900/30', text: 'text-yellow-700 dark:text-yellow-400', dot: 'bg-yellow-500', label: '进行中' },
  completed: { bg: 'bg-green-100 dark:bg-green-900/30', text: 'text-green-700 dark:text-green-400', dot: 'bg-green-500', label: '已完成' },
  failed: { bg: 'bg-red-100 dark:bg-red-900/30', text: 'text-red-700 dark:text-red-400', dot: 'bg-red-500', label: '失败' },
}

export default function StatusBadge({ status, className = '' }: Props) {
  const config = statusConfig[status] || { bg: 'bg-gray-100', text: 'text-gray-700', dot: 'bg-gray-500', label: status }

  return (
    <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${config.bg} ${config.text} ${className}`}>
      <span className={`w-1.5 h-1.5 rounded-full ${config.dot} ${status === 'in_progress' ? 'animate-pulse' : ''}`} />
      {config.label}
    </span>
  )
}
