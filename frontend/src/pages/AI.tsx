import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { aiStatus, streamChat, fetchCard, type AiTurn, type AiCard as Card } from '../lib/aiApi'
import AiCard from '../components/AiCard'
import StarHero from '../components/StarHero'

type Msg =
  | { id: number; role: 'user'; text: string }
  | { id: number; role: 'assistant'; text: string; streaming?: boolean }
  | { id: number; role: 'assistant'; card: Card }

let seq = 0
const nextId = () => ++seq

/** Three staggered gold dots — the "assistant is working" indicator. Pure CSS
 *  (opacity/transform), only mounted while a reply is pending. */
function TypingDots() {
  return (
    <span className="inline-flex items-center gap-1 py-0.5" aria-label="…">
      {[0, 1, 2].map((i) => (
        <span
          key={i}
          className="ai-typing-dot h-1.5 w-1.5 rounded-full bg-accent/80"
          style={{ animationDelay: `${i * 0.16}s` }}
        />
      ))}
    </span>
  )
}

/** Small gold-ringed football mark that tags every assistant turn (SVG, not an
 *  emoji) — gives the thread a clear "who's speaking" rhythm. */
function Crest() {
  return (
    <span className="ai-crest-ring mt-0.5 grid h-6 w-6 shrink-0 place-items-center rounded-full p-[1.5px]">
      <span className="grid h-full w-full place-items-center rounded-full bg-bg">
        <svg viewBox="0 0 24 24" className="h-3.5 w-3.5 text-accent" fill="none" stroke="currentColor" strokeWidth="1.6">
          <circle cx="12" cy="12" r="8.5" />
          <path d="M12 8.2l3.4 2.5-1.3 4h-4.2l-1.3-4L12 8.2z" strokeLinejoin="round" />
        </svg>
      </span>
    </span>
  )
}

/**
 * "Pitchside" — the football AI assistant (ADR-0017, refreshed per ADR-0021).
 * Available to any logged-in user (incl. demo `none`). Streams chat replies; a
 * card button fetches a structured club/player card. Guardrail + master prompt
 * live server-side.
 */
export default function AI() {
  const { t } = useTranslation()
  const [available, setAvailable] = useState<boolean | null>(null)
  const [messages, setMessages] = useState<Msg[]>([])
  const [input, setInput] = useState('')
  const [busy, setBusy] = useState(false)
  const threadRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const c = new AbortController()
    aiStatus(c.signal).then((ok) => !c.signal.aborted && setAvailable(ok))
    return () => c.abort()
  }, [])

  useEffect(() => {
    threadRef.current?.scrollTo({ top: threadRef.current.scrollHeight, behavior: 'smooth' })
  }, [messages])

  function historyFor(): AiTurn[] {
    return messages
      .filter((m): m is Extract<Msg, { text: string }> => 'text' in m && !!m.text)
      .map((m) => ({ role: m.role === 'user' ? 'user' : 'model', text: m.text }))
  }

  // preset lets a suggestion chip send its label directly (without touching the
  // input field); a plain submit reads and clears the input.
  async function onChat(preset?: string) {
    const q = (preset ?? input).trim()
    if (!q || busy) return
    if (preset === undefined) setInput('')
    setBusy(true)
    const history = historyFor()
    setMessages((m) => [...m, { id: nextId(), role: 'user', text: q }])
    const aid = nextId()
    setMessages((m) => [...m, { id: aid, role: 'assistant', text: '', streaming: true }])
    try {
      await streamChat(q, history, (tok) => {
        setMessages((m) => m.map((x) => (x.id === aid && 'text' in x ? { ...x, text: x.text + tok } : x)))
      })
    } catch {
      setMessages((m) => m.map((x) => (x.id === aid && 'text' in x ? { ...x, text: t('ai.error') } : x)))
    } finally {
      setMessages((m) => m.map((x) => (x.id === aid && 'streaming' in x ? { ...x, streaming: false } : x)))
      setBusy(false)
    }
  }

  async function onCard() {
    const q = input.trim()
    if (!q || busy) return
    setInput('')
    setBusy(true)
    setMessages((m) => [...m, { id: nextId(), role: 'user', text: q }])
    try {
      const res = await fetchCard(q)
      // Refusal or an ambiguity clarification both carry a `message` (shown as a
      // text bubble); a real card is rendered by <AiCard>.
      if ('message' in res) {
        setMessages((m) => [...m, { id: nextId(), role: 'assistant', text: res.message }])
      } else {
        setMessages((m) => [...m, { id: nextId(), role: 'assistant', card: res }])
      }
    } catch {
      setMessages((m) => [...m, { id: nextId(), role: 'assistant', text: t('ai.error') }])
    } finally {
      setBusy(false)
    }
  }

  const canSend = !!input.trim() && !busy

  return (
    <div className="mx-auto flex w-full max-w-2xl flex-col">
      <header className="relative mb-4 -mx-4 overflow-hidden rounded-b-3xl px-4 pb-5 pt-4 sm:-mx-6 sm:px-6">
        <StarHero variant="band" />
        <div className="relative">
          <p className="mb-1 text-[0.62rem] font-bold uppercase tracking-[0.22em] text-accent">{t('ai.eyebrow')}</p>
          <h1 className="text-2xl font-bold tracking-tight text-text sm:text-3xl">{t('ai.title')}</h1>
          <p className="mt-1 text-sm text-muted">{t('ai.subtitle')}</p>
        </div>
      </header>

      {available === false ? (
        <p className="rounded-2xl border border-hairline bg-surface px-6 py-12 text-center text-sm text-muted backdrop-blur-md">
          {t('ai.unavailable')}
        </p>
      ) : (
        <>
          <div ref={threadRef} className="min-h-[40vh] space-y-3 overflow-y-auto pb-4">
            {messages.length === 0 && (
              <div className="ai-msg-in px-1 py-8">
                <p className="mb-3.5 text-center text-sm text-muted/70">{t('ai.empty')}</p>
                <div className="flex flex-wrap justify-center gap-2">
                  {(['results', 'scorers', 'leader'] as const).map((k) => {
                    const label = t(`ai.suggest.${k}`)
                    return (
                      <button
                        key={k}
                        type="button"
                        onClick={() => void onChat(label)}
                        disabled={busy}
                        className="rounded-full border border-hairline bg-white/[0.04] px-3.5 py-1.5 text-sm text-text/90 transition hover:-translate-y-0.5 hover:border-accent/50 hover:bg-accent/[0.08] disabled:opacity-40"
                      >
                        {label}
                      </button>
                    )
                  })}
                </div>
              </div>
            )}
            {messages.map((m) => {
              if ('card' in m) {
                return (
                  <div key={m.id} className="ai-msg-in flex items-start gap-2">
                    <Crest />
                    <div className="min-w-0 max-w-[88%] flex-1">
                      <AiCard card={m.card} />
                    </div>
                  </div>
                )
              }
              if (m.role === 'user') {
                return (
                  <div key={m.id} className="ai-msg-in flex justify-end">
                    <div className="max-w-[85%] rounded-2xl rounded-tr-md bg-accent/15 px-3.5 py-2.5 text-sm leading-relaxed text-text">
                      {m.text}
                    </div>
                  </div>
                )
              }
              return (
                <div key={m.id} className="ai-msg-in flex items-start gap-2">
                  <Crest />
                  <div className="max-w-[85%] rounded-2xl rounded-tl-md border border-hairline bg-white/[0.04] px-3.5 py-2.5 text-sm leading-relaxed text-text/90">
                    {m.text ? m.text : 'streaming' in m && m.streaming ? <TypingDots /> : ''}
                  </div>
                </div>
              )
            })}
            {/* Card generation has no streaming placeholder — show the indicator
                for the 3-7s wait so the click has immediate feedback. */}
            {busy && !messages.some((m) => 'streaming' in m && m.streaming) && (
              <div className="ai-msg-in flex items-start gap-2">
                <Crest />
                <div className="rounded-2xl rounded-tl-md border border-hairline bg-white/[0.04] px-3.5 py-2.5">
                  <TypingDots />
                </div>
              </div>
            )}
          </div>

          <form
            onSubmit={(e) => {
              e.preventDefault()
              void onChat()
            }}
            className="sticky bottom-24 mt-2 sm:bottom-4"
          >
            <div className="flex items-center gap-1 rounded-2xl border border-hairline bg-white/[0.05] py-1.5 pl-4 pr-1.5 backdrop-blur-md transition-colors focus-within:border-accent/60">
              <input
                value={input}
                onChange={(e) => setInput(e.target.value)}
                maxLength={2000}
                placeholder={t('ai.placeholder')}
                disabled={busy}
                className="min-w-0 flex-1 bg-transparent text-sm text-text outline-none placeholder:text-muted/50 disabled:opacity-50"
              />
              <button
                type="button"
                onClick={() => void onCard()}
                disabled={!canSend}
                title={t('ai.cardHint')}
                aria-label={t('ai.cardHint')}
                className="grid h-9 w-9 shrink-0 place-items-center rounded-full text-muted transition hover:bg-white/[0.07] hover:text-accent active:scale-90 disabled:opacity-30"
              >
                <svg viewBox="0 0 24 24" className="h-[1.15rem] w-[1.15rem]" fill="none" stroke="currentColor" strokeWidth="1.7">
                  <rect x="3" y="5" width="18" height="14" rx="2" />
                  <circle cx="8.5" cy="11" r="1.8" />
                  <path d="M13 9.5h5M13 13h5M5.5 15.5h8" strokeLinecap="round" />
                </svg>
              </button>
              <button
                type="submit"
                disabled={!canSend}
                aria-label={t('ai.send')}
                title={t('ai.send')}
                className={`grid h-9 w-9 shrink-0 place-items-center rounded-full bg-accent text-bg transition hover:opacity-90 active:scale-90 disabled:opacity-30 ${canSend ? 'ai-send-active' : ''}`}
              >
                <svg viewBox="0 0 24 24" className="h-[1.1rem] w-[1.1rem]" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M22 2 11 13" />
                  <path d="M22 2 15 22l-4-9-9-4 20-7z" />
                </svg>
              </button>
            </div>
          </form>
          <p className="mt-1.5 px-1 text-center text-[0.62rem] text-muted/50">{t('ai.disclaimer')}</p>
        </>
      )}
    </div>
  )
}
