import { useEffect, useSyncExternalStore } from 'react'
import { AnimatePresence, motion } from 'framer-motion'
import { type ToastItem, subscribe, getToasts, removeToast } from './toastStore'

type ToastType = ToastItem['type']

function useToastStore() {
  return useSyncExternalStore(subscribe, getToasts)
}

const colorMap: Record<ToastType, { bg: string; color: string; icon: string }> = {
  success: { bg: 'var(--success-soft)', color: 'var(--success)', icon: 'M8 12l3 3 5-5' },
  error:   { bg: 'var(--danger-soft)',  color: 'var(--danger)',  icon: 'M15 9l-6 6M9 9l6 6' },
  info:    { bg: 'var(--info-soft)',    color: 'var(--info)',    icon: 'M12 8v4M12 16h.01' },
}

function ToastItem({ item, onDismiss }: { item: ToastItem; onDismiss: () => void }) {
  useEffect(() => {
    const t = setTimeout(onDismiss, 3500)
    return () => clearTimeout(t)
  }, [onDismiss])

  const c = colorMap[item.type]

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: -8, scale: 0.96 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, y: -8, scale: 0.96 }}
      transition={{ duration: 0.25, ease: [0.16, 1, 0.3, 1] }}
      className="flex items-center gap-2.5 px-4 py-3 rounded-xl text-sm font-medium shadow-lg cursor-pointer max-w-[400px]"
      style={{
        background: 'var(--bg-elevated)',
        border: '1px solid var(--border-default)',
        color: 'var(--text-primary)',
        boxShadow: 'var(--shadow-lg)',
      }}
      onClick={onDismiss}
    >
      <div
        className="w-6 h-6 rounded-lg flex items-center justify-center flex-shrink-0"
        style={{ background: c.bg }}
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke={c.color} strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
          <circle cx="12" cy="12" r="10" strokeWidth="2" />
          <path d={c.icon} />
        </svg>
      </div>
      <span className="line-clamp-2">{item.message}</span>
    </motion.div>
  )
}

export default function ToastContainer() {
  const items = useToastStore()

  return (
    <div className="fixed top-4 right-4 z-[100] flex flex-col items-end gap-2 pointer-events-none">
      <AnimatePresence mode="popLayout">
        {items.map((item) => (
          <div key={item.id} className="pointer-events-auto">
            <ToastItem item={item} onDismiss={() => removeToast(item.id)} />
          </div>
        ))}
      </AnimatePresence>
    </div>
  )
}
