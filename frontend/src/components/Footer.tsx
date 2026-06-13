import { useTranslation } from 'react-i18next'

/**
 * Page footer — a quiet sign-off so pages don't end on a bare edge above the
 * bottom nav. Centres the tournament trophy mark with a hairline divider and a
 * small bilingual caption. Purely decorative; no links to chase.
 */
export default function Footer() {
  const { t } = useTranslation()
  return (
    <footer className="mx-auto mt-14 w-full max-w-5xl px-2 text-center">
      <div className="flex items-center justify-center gap-3 text-muted/40">
        <span className="h-px w-12 bg-gradient-to-r from-transparent to-hairline" />
        <img
          src="/img/trophy.png"
          alt=""
          aria-hidden
          className="h-9 w-auto opacity-70"
          style={{ filter: 'grayscale(0.3) brightness(0.95)' }}
        />
        <span className="h-px w-12 bg-gradient-to-l from-transparent to-hairline" />
      </div>
      <p className="mt-3 text-[0.62rem] font-semibold uppercase tracking-[0.28em] text-muted/55">
        World Cup <span className="text-accent/70 tabular-nums">2026</span>
      </p>
      <p className="mt-1 text-[0.62rem] tracking-wide text-muted/40">{t('footer.tagline')}</p>
    </footer>
  )
}
