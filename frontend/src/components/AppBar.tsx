import { NavLink } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useTranslation } from 'react-i18next'
import BackendBadge from './BackendBadge'
import LanguageSwitcher from './LanguageSwitcher'
import AuthControl from './AuthControl'
import Trophy from './Trophy'

const TABS = [
  { to: '/', labelKey: 'nav.calendar', end: true },
  { to: '/groups', labelKey: 'nav.groups', end: false },
  { to: '/competition', labelKey: 'nav.competition', end: false },
  { to: '/history', labelKey: 'nav.history', end: false },
  { to: '/bracket', labelKey: 'nav.bracket', end: false },
  { to: '/ai', labelKey: 'nav.ai', end: false },
  { to: '/audit', labelKey: 'nav.audit', end: false },
] as const

function Wordmark() {
  return (
    <NavLink to="/" className="flex items-center gap-2 select-none">
      <Trophy className="h-[22px] w-[22px] rounded-lg" />
      <span className="text-sm font-semibold uppercase tracking-[0.22em] text-text">
        WC
      </span>
      <span className="text-sm font-semibold uppercase tracking-[0.22em] text-muted/50">
        ·
      </span>
      <span className="text-sm font-semibold uppercase tracking-[0.22em] text-accent tabular-nums">
        2026
      </span>
    </NavLink>
  )
}

function Tab({ to, label, end }: { to: string; label: string; end: boolean }) {
  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        `relative px-1 py-2 text-sm font-medium transition-colors ${
          isActive ? 'text-text' : 'text-muted hover:text-text'
        }`
      }
    >
      {({ isActive }) => (
        <>
          {label}
          {isActive && (
            <motion.span
              layoutId="tab-underline"
              className="absolute -bottom-px left-0 right-0 h-[2px] rounded-full bg-accent shadow-[0_0_8px_1px_rgba(201,162,75,0.6)]"
              transition={{ type: 'spring', stiffness: 380, damping: 30 }}
            />
          )}
        </>
      )}
    </NavLink>
  )
}

export default function AppBar() {
  const { t } = useTranslation()

  return (
    <header className="sticky top-0 z-30 border-b border-hairline bg-bg/70 backdrop-blur-xl">
      <div className="mx-auto flex h-14 max-w-5xl items-center justify-between px-4 sm:px-6">
        <Wordmark />

        {/* Desktop / tablet tabs */}
        <nav className="hidden items-center gap-5 lg:flex">
          {TABS.map((tab) => (
            <Tab key={tab.to} to={tab.to} end={tab.end} label={t(tab.labelKey)} />
          ))}
        </nav>

        <div className="flex items-center gap-2 sm:gap-3">
          <LanguageSwitcher />
          <BackendBadge />
          <AuthControl />
        </div>
      </div>
    </header>
  )
}

/** Bottom navigation shown on mobile + tablet (< lg). Five compact items fit
 *  within 375px because each cell is flex-1 with tight padding/type. */
export function BottomNav() {
  const { t } = useTranslation()

  return (
    <nav className="fixed inset-x-0 bottom-0 z-30 border-t border-hairline bg-bg/80 backdrop-blur-xl lg:hidden">
      <div className="mx-auto flex max-w-5xl items-stretch justify-around px-1 pb-[env(safe-area-inset-bottom)]">
        {TABS.map((tab) => (
          <NavLink
            key={tab.to}
            to={tab.to}
            end={tab.end}
            className={({ isActive }) =>
              `relative flex flex-1 flex-col items-center gap-1 px-0.5 py-3 text-center text-[0.62rem] font-medium leading-tight transition-colors ${
                isActive ? 'text-accent' : 'text-muted'
              }`
            }
          >
            {({ isActive }) => (
              <>
                {isActive && (
                  <motion.span
                    layoutId="bottomnav-indicator"
                    className="absolute top-0 h-[2px] w-8 rounded-full bg-accent shadow-[0_0_8px_1px_rgba(201,162,75,0.6)]"
                    transition={{ type: 'spring', stiffness: 380, damping: 30 }}
                  />
                )}
                <span className="truncate">{t(tab.labelKey)}</span>
              </>
            )}
          </NavLink>
        ))}
      </div>
    </nav>
  )
}
