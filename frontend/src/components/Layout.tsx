import { Outlet, useLocation } from 'react-router-dom'
import { motion, useReducedMotion } from 'framer-motion'
import AppBar, { BottomNav } from './AppBar'
import Spotlight from './Spotlight'

export default function Layout() {
  const location = useLocation()
  const reduce = useReducedMotion()

  return (
    <div className="relative min-h-screen bg-gradient-to-b from-bg to-bg-end">
      <Spotlight />
      <AppBar />

      <main className="px-4 pb-28 pt-8 sm:px-6 sm:pb-16">
        {/* keyed by path → remounts & plays an enter animation on each route.
            No AnimatePresence/exit: the incoming page must never wait on an
            outgoing animation (that caused pages to need a manual refresh). */}
        <motion.div
          key={location.pathname}
          initial={reduce ? false : { opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: reduce ? 0 : 0.28, ease: [0.22, 1, 0.36, 1] }}
        >
          <Outlet />
        </motion.div>
      </main>

      <BottomNav />
    </div>
  )
}
