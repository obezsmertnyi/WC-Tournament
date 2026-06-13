import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'

type ComingSoonSection = 'leaderboard' | 'bracket' | 'notFound'

interface ComingSoonProps {
  section: ComingSoonSection
}

export default function ComingSoon({ section }: ComingSoonProps) {
  const { t } = useTranslation()
  const title = t(`comingSoon.${section}.title`)
  const description = t(`comingSoon.${section}.description`)

  return (
    <div className="mx-auto w-full max-w-5xl">
      <header className="mb-8">
        <h1 className="text-2xl font-bold tracking-tight text-text sm:text-3xl">{title}</h1>
      </header>

      <motion.div
        initial={{ opacity: 0, y: 14 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.55, ease: [0.22, 1, 0.36, 1] }}
        className="flex flex-col items-center rounded-2xl border border-hairline bg-surface px-6 py-20 text-center backdrop-blur-md"
      >
        <p className="text-[0.65rem] font-semibold uppercase tracking-[0.28em] text-accent">
          {t('comingSoon.eyebrow')}
        </p>
        <p className="mt-4 text-lg font-semibold text-text">{title}</p>
        <p className="mt-2 max-w-sm text-sm leading-relaxed text-muted">{description}</p>
      </motion.div>
    </div>
  )
}
