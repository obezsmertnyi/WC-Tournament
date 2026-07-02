import { Component, type ReactNode } from 'react'

/**
 * Isolates a subtree: if a child throws during render, show `fallback` (default:
 * nothing) instead of crashing the parent. Used to make the AI recap a strictly
 * additive layer — a recap failure can never take down the surrounding UI (e.g. the
 * DB-backed predictions grid).
 */
export default class ErrorBoundary extends Component<
  { children: ReactNode; fallback?: ReactNode },
  { failed: boolean }
> {
  state = { failed: false }
  static getDerivedStateFromError() {
    return { failed: true }
  }
  componentDidCatch() {
    /* swallow — the fallback is intentional; the rest of the UI carries on */
  }
  render() {
    return this.state.failed ? (this.props.fallback ?? null) : this.props.children
  }
}
