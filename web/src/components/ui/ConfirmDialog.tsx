import { motion, AnimatePresence } from 'framer-motion'

interface Props {
  open: boolean
  title: string
  message: string
  confirmLabel?: string
  danger?: boolean
  onConfirm: () => void
  onCancel: () => void
}

export default function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel = '确认',
  danger = false,
  onConfirm,
  onCancel,
}: Props) {
  return (
    <AnimatePresence>
      {open && (
        <>
          <motion.div
            className="fixed inset-0 z-[90]"
            style={{ background: 'rgba(0,0,0,0.4)' }}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={onCancel}
          />
          <div className="fixed inset-0 z-[91] flex items-center justify-center p-4" onClick={onCancel}>
            <motion.div
              className="w-full max-w-[360px] p-6 rounded-2xl"
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
              <h3 className="text-[15px] font-semibold mb-1.5" style={{ color: 'var(--text-primary)' }}>
                {title}
              </h3>
              <p className="text-sm mb-5" style={{ color: 'var(--text-secondary)' }}>
                {message}
              </p>
              <div className="flex items-center justify-end gap-2">
                <button
                  onClick={onCancel}
                  className="px-4 py-2 rounded-xl text-sm font-medium transition-colors cursor-pointer"
                  style={{
                    color: 'var(--text-secondary)',
                    background: 'var(--bg-inset)',
                  }}
                >
                  取消
                </button>
                <button
                  onClick={onConfirm}
                  className="px-4 py-2 rounded-xl text-sm font-medium text-white transition-colors cursor-pointer"
                  style={{
                    background: danger ? 'var(--danger)' : 'var(--accent)',
                  }}
                >
                  {confirmLabel}
                </button>
              </div>
            </motion.div>
          </div>
        </>
      )}
    </AnimatePresence>
  )
}
