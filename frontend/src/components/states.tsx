import { motion } from 'framer-motion'

/** Single shimmering placeholder card. */
function SkeletonCard() {
  return (
    <div className="rounded-2xl border border-hairline bg-surface p-4 backdrop-blur-md">
      <div className="mb-3 flex items-center justify-between">
        <div className="h-3 w-12 animate-pulse rounded-full bg-white/10" />
        <div className="h-4 w-14 animate-pulse rounded-full bg-white/10" />
      </div>
      <div className="space-y-2.5">
        {[0, 1].map((i) => (
          <div key={i} className="flex items-center gap-3">
            <div className="h-4 w-6 animate-pulse rounded-[3px] bg-white/10" />
            <div className="h-3.5 flex-1 animate-pulse rounded-full bg-white/10" />
          </div>
        ))}
      </div>
      <div className="mt-3.5 border-t border-hairline pt-2.5">
        <div className="h-2.5 w-2/3 animate-pulse rounded-full bg-white/10" />
      </div>
    </div>
  )
}

export function FixturesSkeleton() {
  return (
    <div className="space-y-8">
      {[0, 1].map((section) => (
        <div key={section}>
          <div className="mb-3 h-3 w-24 animate-pulse rounded-full bg-white/10" />
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 4 }).map((_, i) => (
              <SkeletonCard key={i} />
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

export function EmptyState() {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
      className="flex flex-col items-center rounded-2xl border border-hairline bg-surface px-6 py-16 text-center backdrop-blur-md"
    >
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full border border-hairline bg-white/[0.03]">
        <span className="h-2 w-2 animate-pulse rounded-full bg-accent shadow-[0_0_10px_2px_rgba(201,162,75,0.55)]" />
      </div>
      <p className="text-sm font-medium text-text">Календар синхронізується…</p>
      <p className="mt-1.5 max-w-xs text-xs leading-relaxed text-muted">
        Матчі з’являться, щойно дані надійдуть із FIFA. Завітайте трохи згодом.
      </p>
    </motion.div>
  )
}

export function ErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
      className="flex flex-col items-center rounded-2xl border border-hairline bg-surface px-6 py-16 text-center backdrop-blur-md"
    >
      <div className="mb-4 flex h-12 w-12 items-center justify-center rounded-full border border-red-500/30 bg-red-500/[0.06]">
        <span className="h-2 w-2 rounded-full bg-red-500/70" />
      </div>
      <p className="text-sm font-medium text-text">Не вдалося завантажити календар</p>
      <p className="mt-1.5 max-w-xs text-xs leading-relaxed text-muted">
        Перевірте з’єднання та спробуйте ще раз.
      </p>
      <button
        type="button"
        onClick={onRetry}
        className="mt-5 rounded-full border border-hairline bg-white/[0.04] px-5 py-2 text-xs font-medium uppercase tracking-[0.14em] text-text transition-colors hover:border-accent/40 hover:text-accent"
      >
        Повторити
      </button>
    </motion.div>
  )
}
