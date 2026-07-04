# Design-first UI workflow (Claude Design / Artifacts → implement)

A reusable playbook for building UI with a coding agent so the result looks
designed, not generated. It applies to **greenfield** (no UI yet) and
**brownfield** (an app already exists) work. Portable — copy this file into any
repo and adapt the token/command specifics.

## The principle

**Separate the design step from the implementation step.** Produce a precise,
reviewable design artifact *first*; then implement the code against that exact
artifact. Never hand the agent a raw screenshot/mockup plus "build this, but a
bit different" — that asks it to parse pixels, guess intent, and reconcile the
delta all at once, and it does all three poorly.

This is the visual analog of spec-driven development: the design artifact is the
spec, and code is generated against a spec, not a vibe.

```
  raw idea / reference  ──▶  DESIGN step  ──▶  reviewable artifact  ──▶  IMPLEMENT step  ──▶  shipped UI
                             (Claude Design           (tokens, states,        (code the artifact
                              or an Artifact)          real copy)              1:1, verify)
```

## Why "make it like this, but different" fails

- The agent must simultaneously **interpret** the reference, **infer** the
  unstated intent behind "different", and **produce** code — three lossy steps
  compounded into one.
- Visual intent is ambiguous in prose. "Cleaner", "more modern", "more us" mean
  nothing an agent can verify against.
- There is no reviewable checkpoint between "idea" and "code", so the first
  thing you can react to is already an implementation — expensive to change.

Splitting it fixes all three: the design step resolves the ambiguity into a
concrete artifact you approve; the implementation step has an unambiguous target.

## The two routes

### Route A — Artifact (no setup, use by default)

The coding agent authors the design itself as a **self-contained HTML Artifact**
(hosted page you can open, review, and iterate on), then implements it.

1. Load the `artifact-design` skill (calibrates design quality; it also lists the
   AI-generated-design tells to avoid).
2. Author the design as one self-contained HTML file honoring the target's real
   design tokens and **real copy** (never lorem). Show *all states* the surface
   has (empty, loading, populated, error), and design **both light and dark**.
3. Publish it (the `Artifact` tool). Review together; iterate on the artifact
   until approved.
4. Implement the approved artifact in the app framework, 1:1.

Best for: a single screen or a focused refresh; fastest path; zero auth.

### Route B — Claude Design + DesignSync (a real, synced design system)

Use the claude.ai/design surface as the design tool and sync a component library
into the repo.

1. **One-time:** run `/design-login` to authorize design-system access for your
   claude.ai account (required before the `DesignSync` tool can reach
   claude.ai/design; without it the tool errors "needs design-system
   authorization").
2. Create/populate a design-system **project** (via the `/design-sync` skill +
   `DesignSync` tool) — **component by component, never a wholesale replace**.
3. Iterate visually in claude.ai/design; pull updates back into the repo's local
   component library via DesignSync.
4. Implement against the synced components.

Best for: a design system reused across many surfaces or repos; ongoing visual
iteration; a team that wants the design system as a first-class, versioned asset.

## Greenfield playbook (no UI yet)

1. **Pin the subject.** One sentence: what it is, who uses it, the page's single
   job. Distinctive design comes from the subject's own world, not a template.
2. **Establish the design language in the design step**, before any code:
   - Color: 4–6 named hex values (one accent; neutrals chosen, not defaulted).
   - Type: 2+ roles (a characterful display face used with restraint, a body
     face, optionally a mono/utility face). Set a type scale and stay on it.
   - Layout: one or two sentences describing the composition.
3. **Author the artifact** (Route A) or **seed the design system** (Route B),
   covering every state and both themes.
4. **Review and approve** the design. This is the checkpoint — change it here,
   not in code.
5. **Implement** against the approved artifact/system; wire real data.
6. **Verify** (see Definition of done) and, for a system, keep it synced.

## Brownfield playbook (an app already exists)

The extra discipline over greenfield: **honor what's already there.** Precedence
is always `user's explicit words > the project's existing system > the agent's
choices`.

1. **Extract the ground truth first** — do not invent tokens:
   - Design tokens: the theme/config file (e.g. `frontend/tailwind.config.js`),
     CSS custom properties, existing component styles.
   - Real copy: the i18n/locale files (all languages), not placeholder text.
   - Existing components/anatomy you must match (e.g. a card component's fields).
   - Icon convention already in use (e.g. inline SVG) — match it.
2. **Design the new/refreshed surface honoring those tokens.** The refresh adds
   information design and polish; it does not restyle the whole app. Keep the
   existing accent and type system unless the user explicitly asks to change
   them.
3. **Constrain for the existing runtime.** If the app avoids heavy assets, the
   design must too — e.g. do not introduce a large webfont (especially wide
   scripts like Cyrillic/CJK) when the app already commits to a system/loaded
   face. Get personality from scale, spacing, and layout instead.
4. **Author the artifact** showing the refreshed surface *inside* the app's
   chrome (so it reads as the same product), all states, both themes.
5. **Review and approve.**
6. **Implement 1:1** into the existing components; add only the new i18n keys and
   styles the design needs. Reuse existing helpers/components.
7. **Verify** and deploy per the project's deploy workflow.

## Design guardrails (apply to every design and its docs)

- **Honor the existing system.** Existing tokens/components/copy win over new
  invention. Fill gaps; don't override.
- **No emoji.** Emoji as icons, section markers, or bullets read as unpolished /
  AI-generated. Use inline SVG icons (match the app's existing icon style) or
  restrained typographic marks. This applies to the app **and** to
  documentation.
- **Design both themes.** Drive everything through tokens; redefine tokens under
  `@media (prefers-color-scheme)` and the explicit theme attribute. Give the
  second theme the same care — don't naively invert.
- **One accent, spent once.** Keep the brand accent singular; semantic colors
  (good/warning/critical) are separate from the accent and don't count as a
  second accent.
- **Real content, never lorem.** Pull real copy and plausible data so the design
  reveals real overflow/empty/long-string behavior.
- **Performance is a design constraint.** No heavy fonts/assets the runtime
  wouldn't ship. State this in the design plan.
- **Accessibility.** Visible keyboard focus, legible contrast in both themes,
  `prefers-reduced-motion` disables non-essential motion.
- **Motion is deliberate and light.** One orchestrated moment beats scattered
  effects; too much animation itself reads as AI-generated.
- **Don't let the coding agent infer the design.** Everything the code needs
  (tokens, states, spacing, copy) must be explicit in the artifact — the same
  reason our data layers hand the model explicit values instead of letting it
  infer (see `AGENTS.md` correctness rules).

## Definition of done

- The design artifact was reviewed and approved before implementation began.
- The implementation matches the approved artifact (both themes, all states).
- Real copy/data wired; no lorem, no placeholder emoji.
- No new heavy assets vs. the app baseline; motion respects reduced-motion.
- Verified by driving the real surface (see the repo's `verify` flow), not just
  a build/typecheck.

## Tools & commands

| Purpose | Tool / command |
| --- | --- |
| Design-quality guidance before authoring a page | `artifact-design` skill |
| Publish a self-contained design as a hosted page | `Artifact` tool (Route A) |
| Authorize claude.ai/design access (one-time) | `/design-login` (Route B) |
| Create/sync a design-system project, component-by-component | `/design-sync` skill + `DesignSync` tool (Route B) |
| Read this app's tokens | `frontend/tailwind.config.js` (colors, fonts) |
| Read this app's copy | `frontend/src/i18n/locales/*.json` |

## This repo's applied example

The **AI assistant page** (`frontend/src/pages/AI.tsx`, `components/AiCard.tsx`)
was the first surface refreshed with Route A: extract tokens (gold `#C9A24B` on
near-black `#0B0C0E→#15171B`, Inter) + real UA/EN copy → author an emoji-free
Artifact showing empty/populated/loading states in both themes → review →
implement 1:1. See ADR-0021.
