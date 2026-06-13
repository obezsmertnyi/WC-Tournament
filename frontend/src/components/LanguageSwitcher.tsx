import { useTranslation } from 'react-i18next'
import { motion } from 'framer-motion'
import { SUPPORTED_LANGUAGES, type Language } from '../i18n'

/**
 * Minimal segmented EN/UA control for the app bar. Champagne-gold active
 * state on glass, matching the rest of the chrome. Persists via i18next's
 * localStorage detector and switches instantly (no reload).
 */
export default function LanguageSwitcher() {
  const { t, i18n } = useTranslation()
  const current = (
    SUPPORTED_LANGUAGES.includes(i18n.resolvedLanguage as Language)
      ? i18n.resolvedLanguage
      : 'uk'
  ) as Language

  return (
    <div
      role="group"
      aria-label={t('language.label')}
      className="inline-flex items-center rounded-full border border-hairline bg-surface p-0.5 backdrop-blur-md"
    >
      {SUPPORTED_LANGUAGES.map((lng) => {
        const active = current === lng
        return (
          <button
            key={lng}
            type="button"
            onClick={() => i18n.changeLanguage(lng)}
            aria-pressed={active}
            className={`relative rounded-full px-2.5 py-1 text-[0.65rem] font-semibold uppercase tracking-[0.14em] transition-colors ${
              active ? 'text-bg' : 'text-muted hover:text-text'
            }`}
          >
            {active && (
              <motion.span
                layoutId="lang-pill"
                className="absolute inset-0 rounded-full bg-accent shadow-[0_0_10px_1px_rgba(201,162,75,0.5)]"
                transition={{ type: 'spring', stiffness: 380, damping: 30 }}
              />
            )}
            <span className="relative">{t(`language.${lng}`)}</span>
          </button>
        )
      })}
    </div>
  )
}
