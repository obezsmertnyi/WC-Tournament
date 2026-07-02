import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

/** The hashed main bundle this tab loaded (prod build only; null in dev). */
function loadedBundle(): string | null {
  const s = document.querySelector('script[type="module"][src*="/assets/index-"]')
  return s?.getAttribute('src') ?? null
}

/**
 * Detects a new deploy. An SPA keeps running the JS it first loaded, so after a
 * deploy an open tab stays on the old bundle (why "your changes don't show"). We
 * poll index.html (no-store) on focus + every few minutes; if it references a
 * different hashed bundle, we surface a one-tap reload. Prod-only (dev has no
 * hashed bundle, so this no-ops).
 */
export default function VersionWatcher() {
  const { t } = useTranslation()
  const [stale, setStale] = useState(false)

  useEffect(() => {
    const mine = loadedBundle()
    if (!mine) return
    let alive = true
    const check = async () => {
      try {
        const res = await fetch('/', { cache: 'no-store' })
        const html = await res.text()
        const m = html.match(/\/assets\/index-[A-Za-z0-9_-]+\.js/)
        if (alive && m && m[0] !== mine) setStale(true)
      } catch {
        /* offline / transient — ignore */
      }
    }
    check()
    const id = window.setInterval(check, 5 * 60_000)
    const onFocus = () => check()
    window.addEventListener('focus', onFocus)
    return () => {
      alive = false
      window.clearInterval(id)
      window.removeEventListener('focus', onFocus)
    }
  }, [])

  if (!stale) return null
  return (
    <button
      onClick={() => window.location.reload()}
      className="fixed bottom-24 left-1/2 z-50 -translate-x-1/2 whitespace-nowrap rounded-full bg-accent px-4 py-2 text-sm font-semibold text-bg shadow-lg transition-opacity hover:opacity-90 sm:bottom-6"
    >
      {t('app.newVersion')}
    </button>
  )
}
