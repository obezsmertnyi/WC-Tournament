import { useReducedMotion } from 'framer-motion'
import type { Transition } from 'framer-motion'

/**
 * Animation helpers that stay robust in headless renders, screenshots, and on
 * slow devices.
 *
 * The core rule: never leave content stuck at `opacity: 0`. When the user (or a
 * headless capture) prefers reduced motion we render the *final* state
 * immediately — no enter offset, no fade-in. Otherwise we use short,
 * self-completing mount animations.
 */

const ENTER_EASE: [number, number, number, number] = [0.22, 1, 0.36, 1]

/** A plain animation target (the shape <motion.*> `initial`/`animate` accept). */
type Target = { opacity?: number; y?: number }

interface MountAnimation {
  initial: Target
  animate: Target
  transition: Transition
}

/**
 * Props for a self-completing "rise + fade" mount animation. Pass the result
 * straight into a <motion.*> element.
 *
 * Under reduced motion the `initial` state IS the final, fully-visible state
 * (and the transition is instant), so the element renders correctly even when
 * framer-motion never gets a chance to run an enter animation — exactly the
 * case in headless captures and on slow devices. We deliberately return a final
 * `TargetAndTransition` (rather than `false`) so the visible-from-mount
 * guarantee holds without widening the motion prop union.
 *
 * @param y      enter offset in px (default 12)
 * @param delay  start delay in seconds (default 0)
 */
export function useMountAnimation(y = 12, delay = 0): MountAnimation {
  const reduce = useReducedMotion()
  const settled: Target = { opacity: 1, y: 0 }
  if (reduce) {
    return { initial: settled, animate: settled, transition: { duration: 0 } }
  }
  return {
    initial: { opacity: 0, y },
    animate: settled,
    transition: { duration: 0.32, delay, ease: ENTER_EASE },
  }
}
