export default function LoadingState({ text = '加载中...' }: { text?: string }) {
  return (
    <div className="flex flex-col items-center justify-center py-20" style={{ color: 'var(--text-tertiary)' }}>
      <div className="relative w-10 h-10 mb-4">
        <div
          className="absolute inset-0 rounded-full border-2"
          style={{
            borderColor: 'var(--border-default)',
          }}
        />
        <div
          className="absolute inset-0 rounded-full border-2 border-transparent"
          style={{
            borderTopColor: 'var(--accent)',
            animation: 'spin 0.8s linear infinite',
          }}
        />
      </div>
      <span className="text-sm font-medium">{text}</span>
    </div>
  )
}
