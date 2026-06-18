# GSAP Landing-Page Animations — Design

**Date:** 2026-06-18
**Scope:** Homepage only (`src/lib/pages/LandingPage.svelte`, served at `/` on suluk.site).
**Status:** Approved design → implementation plan next.

## Goal

Add tasteful motion to the public landing page to make it feel more premium, using
GSAP. "Medium" intensity: subtle on-scroll reveals everywhere, plus two signature
moments (animated stat counters, light hero parallax). Must stay fast, accessible,
and safe on a prerendered page.

## Out of scope (YAGNI)

- The signed-in app / dashboard (CRM, invoices, modals, tables).
- Other marketing pages (feature pillar, `/fitur/*`, About, panduan guides).
- Scroll-scrubbed / pinned sections, large cinematic motion, route transitions.

## Stack constraints

- Svelte 5 (runes) + SvelteKit 2 + `adapter-static`. The homepage is **prerendered**.
- GSAP is not yet installed → add `gsap` (core + `ScrollTrigger`).
- GSAP manipulates the DOM, so it must run **client-only** (in `onMount`, browser side).

## Architecture

- New module `src/lib/anim/landingMotion.js` exporting
  `initLandingMotion(root)` → returns a `cleanup()` function.
  - `root` is the landing page's root element (a `bind:this` ref).
  - All timelines/triggers are created inside `gsap.context(() => { ... }, root)`
    so teardown is a single `ctx.revert()`; element selection is scoped to `root`.
  - The module registers `ScrollTrigger` once.
- `LandingPage.svelte`:
  - Add `import { onMount } from 'svelte'` and a `let rootEl` ref (`bind:this` on
    the existing top-level `.lp` div).
  - In `onMount`, if `gsap`/browser available, call `initLandingMotion(rootEl)` and
    return its cleanup from `onMount` (runs on destroy).
  - No other structural changes; the module targets existing classes
    (`.lp-hero`, `.lp-kicker`, `.lp-h1`, `.lp-lead`, `.lp-hero-cta`, `.lp-hero-note`,
    `.lp-mock`, `.lp-trust-item`, `.lp-trust-item .v`, `.lp-sec`, `.lp-mod`,
    `.lp-uc`, `.lp-pricing > *`, testimonial cards, `.lp-cta-box`, `.lp-float`).

*Alternative considered:* a `use:reveal` Svelte action per element. More
declarative but spreads changes across many template spots. The scoped-module
approach keeps `LandingPage.svelte` edits minimal and teardown trivial. Chosen.

## The animations (medium tier)

1. **Hero entrance (on load).** Stagger fade + slide-up of kicker → h1 → lead →
   CTA row → trust-note; the `.lp-mock` card scales (0.96→1) + fades in slightly
   after. ~0.5–0.7s, soft ease (e.g. `power2.out`). Runs once on mount.

2. **Scroll reveals.** For each content group, a `ScrollTrigger` (start ~`top 85%`,
   `once: true`) fades + rises (y ≈ 24px → 0) the group's items with a small
   stagger:
   - trust items, AI-feature block + list, the 9 module cards, how-it-works steps,
     use-case cards, pricing cards, testimonial cards, app section, final CTA box.

3. **Signature — animated counters.** When the trust bar enters view, the 4 stats
   count from 0 to their targets. Values are formatted strings; parse each into
   `{ value, suffix, decimals, locale }` and tween a number, re-rendering with the
   suffix preserved:
   - `500+` → 0→500 then `+`
   - `120rb+` → 0→120 then `rb+`
   - `99,9%` → 0→99.9 (1 decimal, comma) then `%`
   - `4,9/5` → 0→4.9 (1 decimal, comma) then `/5`
   Uses Indonesian decimal comma. Fires once.

4. **Signature — light hero parallax.** `.lp-mock` and the two `.lp-float` badges
   drift a few px:
   - on scroll (a `ScrollTrigger` with `scrub`) — small `y` offset, different rates
     per layer (mock subtle, floats a bit more);
   - on mouse-move over the hero — a gentle `x/y` translate (a few px), eased.
   Intentionally small, not cinematic.

   **Transform ownership (avoid conflicts on `.lp-mock`/`.lp-float`):** the hero
   entrance (#1) animates only `opacity` + `scale`; the parallax (#4) animates only
   `x`/`y`. GSAP tracks these transform components independently, so the one-shot
   entrance and the continuous parallax never fight over the same property.

## Accessibility & safety (prerendered public page)

- **`prefers-reduced-motion: reduce` → bail out entirely.** `initLandingMotion`
  checks `window.matchMedia('(prefers-reduced-motion: reduce)').matches`; if true,
  it does nothing (no hidden states, no triggers, no listeners). Content stays as
  rendered.
- **No FOUC / no-JS safe.** Elements are **visible by default** in the prerendered
  HTML — no CSS hides them. GSAP sets the "hidden" start state in JS (via
  `gsap.set`/`from`) only when motion is enabled. If JS is disabled, fails, or
  reduced-motion is on, nothing is ever invisible.
- `ScrollTrigger.refresh()` after init (correct positions post-hydration).
- Full cleanup on destroy via `ctx.revert()` + remove the mouse-move listener.

## Files

- `frontend-svelte/package.json` — add `gsap` dependency.
- `frontend-svelte/src/lib/anim/landingMotion.js` — new motion module.
- `frontend-svelte/src/lib/pages/LandingPage.svelte` — `onMount` + `rootEl` ref +
  cleanup (minimal).

## Verification

- `npm run check` (svelte-check) → 0 errors.
- `npm run build` → success (confirms GSAP bundles fine with adapter-static).
- Manual: deploy and watch on suluk.site — hero entrance, scroll reveals,
  counters, parallax; and confirm with OS "reduce motion" on that the page is
  fully visible and static. (Animations are verified by observation, not unit
  tests.)
