import { motion } from 'framer-motion'
import type { ReactNode } from 'react'

interface Props {
  children: ReactNode
  className?: string
  hover?: boolean
}

export default function GlassCard({ children, className = '', hover = false }: Props) {
  return (
    <motion.div
      className={`rounded-2xl bg-white/70 dark:bg-gray-900/70 backdrop-blur-xl border border-white/20 dark:border-white/5 shadow-lg shadow-black/[0.03] dark:shadow-black/[0.2] ${className}`}
      whileHover={hover ? { y: -4, transition: { duration: 0.2 } } : undefined}
    >
      {children}
    </motion.div>
  )
}
