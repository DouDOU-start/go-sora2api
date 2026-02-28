import { motion } from 'framer-motion'
import type { ReactNode, CSSProperties } from 'react'

interface Props {
  children: ReactNode
  className?: string
  hover?: boolean
  style?: CSSProperties
  delay?: number
}

export default function GlassCard({ children, className = '', hover = false, style, delay = 0 }: Props) {
  return (
    <motion.div
      className={className}
      style={{
        background: 'var(--bg-surface)',
        border: '1px solid var(--border-default)',
        borderRadius: 'var(--radius-lg)',
        boxShadow: 'var(--shadow-card)',
        transition: 'background 0.3s, border-color 0.3s, box-shadow 0.3s',
        ...style,
      }}
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{
        duration: 0.4,
        delay: delay * 0.06,
        ease: [0.16, 1, 0.3, 1],
      }}
      whileHover={
        hover
          ? {
              y: -2,
              boxShadow: 'var(--shadow-md)',
              transition: { duration: 0.2 },
            }
          : undefined
      }
    >
      {children}
    </motion.div>
  )
}
