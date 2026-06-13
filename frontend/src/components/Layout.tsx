import { Outlet, useLocation } from 'react-router-dom'
import { AnimatePresence, motion, useReducedMotion } from 'framer-motion'
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
        <AnimatePresence mode="wait">
          <motion.div
            key={location.pathname}
            initial={reduce ? false : { opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={reduce ? undefined : { opacity: 0, y: -8 }}
            transition={{ duration: reduce ? 0 : 0.28, ease: [0.22, 1, 0.36, 1] }}
          >
            <Outlet />
          </motion.div>
        </AnimatePresence>
      </main>

      <BottomNav />
    </div>
  )
}
