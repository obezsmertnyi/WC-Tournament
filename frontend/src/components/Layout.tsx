import { useEffect } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import { motion, useReducedMotion } from 'framer-motion'
import AppBar, { BottomNav } from './AppBar'
import Spotlight from './Spotlight'
import Footer from './Footer'
import LoginScreen from './LoginScreen'
import DemoBanner from './DemoBanner'
import { useAuth } from '../auth/AuthContext'

export default function Layout() {
  const location = useLocation()
  const reduce = useReducedMotion()
  const { status } = useAuth()

  // Reset scroll to the top on every route change — otherwise (esp. on mobile)
  // a new page opens still scrolled to wherever the previous one was.
  useEffect(() => {
    window.scrollTo(0, 0)
  }, [location.pathname])

  // Auth wall: the whole app requires login. Anonymous → login screen only;
  // logout flips status to anonymous → this renders the login screen again.
  if (status === 'loading') {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gradient-to-b from-bg to-bg-end">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-hairline border-t-accent" />
      </div>
    )
  }
  if (status !== 'authenticated') {
    return <LoginScreen />
  }

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
          <DemoBanner />
          <Outlet />
        </motion.div>

        <Footer />
      </main>

      <BottomNav />
    </div>
  )
}
