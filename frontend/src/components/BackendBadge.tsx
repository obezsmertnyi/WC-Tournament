import { useEffect, useState } from 'react'
import { fetchHealth, type HealthState } from '../lib/api'

/**
 * Small, unobtrusive backend status badge for the app bar corner.
 * Shows a dot only on mobile; expands to a label on larger screens.
 */
export default function BackendBadge() {
  const [state, setState] = useState<HealthState>('loading')

  useEffect(() => {
    const controller = new AbortController()
    fetchHealth(controller.signal)
      .then((ok) => setState(ok ? 'online' : 'offline'))
      .catch(() => {
        if (!controller.signal.aborted) setState('offline')
      })
    return () => controller.abort()
  }, [])

  const label =
    state === 'loading' ? 'Перевірка' : state === 'online' ? 'Online' : 'Offline'

  const dotClass =
    state === 'online'
      ? 'bg-accent shadow-[0_0_10px_2px_rgba(201,162,75,0.55)]'
      : state === 'offline'
        ? 'bg-red-500/70'
        : 'bg-muted/50'

  return (
    <div
      className="inline-flex items-center gap-2 rounded-full border border-hairline bg-surface px-2.5 py-1.5 backdrop-blur-md"
      title={`Backend ${label}`}
    >
      <span className="relative flex h-1.5 w-1.5">
        {state === 'online' && (
          <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-accent/60" />
        )}
        <span className={`relative inline-flex h-1.5 w-1.5 rounded-full ${dotClass}`} />
      </span>
      <span className="hidden text-[0.65rem] font-medium uppercase tracking-[0.14em] text-muted sm:inline">
        {label}
      </span>
    </div>
  )
}
