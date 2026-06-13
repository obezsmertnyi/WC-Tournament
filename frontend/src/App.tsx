import { useEffect, useState } from 'react'
import { motion } from 'framer-motion'

type HealthState = 'loading' | 'online' | 'offline'

interface HealthResponse {
  status?: string
}

const container = {
  hidden: { opacity: 0 },
  show: {
    opacity: 1,
    transition: { staggerChildren: 0.12, delayChildren: 0.08 },
  },
}

const item = {
  hidden: { opacity: 0, y: 12 },
  show: {
    opacity: 1,
    y: 0,
    transition: { duration: 0.6, ease: [0.22, 1, 0.36, 1] as const },
  },
}

function StatusPill({ state }: { state: HealthState }) {
  const label =
    state === 'loading'
      ? 'checking backend'
      : state === 'online'
        ? 'backend online'
        : 'backend offline'

  const dotClass =
    state === 'online'
      ? 'bg-accent shadow-[0_0_10px_2px_rgba(201,162,75,0.55)]'
      : state === 'offline'
        ? 'bg-red-500/70'
        : 'bg-muted/50'

  return (
    <div className="inline-flex items-center gap-2.5 rounded-full border border-hairline bg-surface px-4 py-2 backdrop-blur-md">
      <span className="relative flex h-2 w-2">
        {state === 'online' && (
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-accent/60" />
        )}
        <span className={`relative inline-flex h-2 w-2 rounded-full ${dotClass}`} />
      </span>
      <span className="text-xs font-medium uppercase tracking-[0.14em] text-muted">
        {label}
      </span>
    </div>
  )
}

export default function App() {
  const [health, setHealth] = useState<HealthState>('loading')

  useEffect(() => {
    const controller = new AbortController()

    async function check() {
      try {
        const res = await fetch('/api/healthz', { signal: controller.signal })
        if (!res.ok) throw new Error(`status ${res.status}`)
        const data = (await res.json().catch(() => ({}))) as HealthResponse
        const ok =
          !data.status || ['ok', 'healthy', 'up'].includes(data.status.toLowerCase())
        setHealth(ok ? 'online' : 'offline')
      } catch (err) {
        if (controller.signal.aborted) return
        setHealth('offline')
      }
    }

    void check()
    return () => controller.abort()
  }, [])

  return (
    <main className="relative min-h-screen overflow-hidden bg-gradient-to-b from-bg to-bg-end">
      {/* Spotlight glow */}
      <motion.div
        aria-hidden
        className="pointer-events-none absolute left-1/2 top-0 h-[42rem] w-[42rem] -translate-x-1/2 -translate-y-1/3 rounded-full"
        style={{
          background:
            'radial-gradient(circle, rgba(201,162,75,0.10) 0%, rgba(245,246,247,0.05) 28%, rgba(11,12,14,0) 65%)',
        }}
        animate={{ opacity: [0.65, 1, 0.65], scale: [1, 1.04, 1] }}
        transition={{ duration: 7, repeat: Infinity, ease: 'easeInOut' }}
      />

      <div className="relative z-10 flex min-h-screen flex-col items-center justify-between px-6 py-12 sm:py-16">
        <div className="flex flex-1 items-center justify-center">
          <motion.div
            variants={container}
            initial="hidden"
            animate="show"
            className="flex w-full max-w-md flex-col items-center text-center"
          >
            <motion.p
              variants={item}
              className="text-[0.7rem] font-semibold uppercase tracking-[0.32em] text-muted"
            >
              Friends Prediction Pool
            </motion.p>

            <motion.h1
              variants={item}
              className="mt-5 text-balance text-5xl font-bold leading-[0.95] tracking-tight text-text sm:text-6xl"
            >
              <span className="block">World Cup</span>
              <span className="mt-1 block text-accent tabular-nums">2026</span>
            </motion.h1>

            <motion.p
              variants={item}
              className="mt-5 max-w-xs text-balance text-sm leading-relaxed text-muted"
            >
              Predict every match. Outscore your friends. Settle it once and for all.
            </motion.p>

            <motion.div
              variants={item}
              whileHover={{ y: -3 }}
              transition={{ type: 'spring', stiffness: 300, damping: 22 }}
              className="mt-10 w-full"
            >
              <div className="flex items-center justify-center rounded-2xl border border-hairline bg-surface px-6 py-7 backdrop-blur-md">
                <StatusPill state={health} />
              </div>
            </motion.div>
          </motion.div>
        </div>

        <motion.footer
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.9, duration: 0.8 }}
          className="pt-10 text-[0.7rem] uppercase tracking-[0.22em] text-muted/60"
        >
          WC-Tournament · PoC
        </motion.footer>
      </div>
    </main>
  )
}
