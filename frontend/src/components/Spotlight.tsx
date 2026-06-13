import { motion } from 'framer-motion'

/** Ambient gold spotlight glow used across pages. Purely decorative. */
export default function Spotlight() {
  return (
    <motion.div
      aria-hidden
      className="pointer-events-none fixed left-1/2 top-0 -z-10 h-[42rem] w-[42rem] -translate-x-1/2 -translate-y-1/3 rounded-full"
      style={{
        background:
          'radial-gradient(circle, rgba(201,162,75,0.10) 0%, rgba(245,246,247,0.05) 28%, rgba(11,12,14,0) 65%)',
      }}
      animate={{ opacity: [0.6, 0.95, 0.6], scale: [1, 1.04, 1] }}
      transition={{ duration: 7, repeat: Infinity, ease: 'easeInOut' }}
    />
  )
}
