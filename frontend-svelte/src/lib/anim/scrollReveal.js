// Generic on-scroll reveal for content pages (marketing site). Fades + rises
// matched elements as they enter the viewport, staggered per group.
//
// Same safety contract as the landing-page motion:
// - Client-only (call from onMount/afterNavigate).
// - Respects `prefers-reduced-motion: reduce`: bails out, content stays visible.
// - No FOUC / no-JS / SEO-safe: nothing is hidden via CSS; start states are set
//   in JS only when motion is enabled, so the prerendered content is fully
//   visible (and in the DOM, opacity-only — readable by crawlers) regardless.
// - Scoped to `root` via gsap.context → teardown is one ctx.revert().
import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";

gsap.registerPlugin(ScrollTrigger);

// Distinct content blocks on the marketing pages (gp-* classes). Section
// headings/eyebrows/leads reveal per section; card grids stagger their items.
const DEFAULT_SELECTORS = [
  ".gp-eyebrow",
  ".gp-h",
  ".gp-lead",
  ".gp-prose",
  ".gp-value",
  ".gp-cap",
  ".gp-step",
  ".gp-benefits > *",
  ".gp-related > *",
  ".gp-cta-box",
];

/**
 * Initialise scroll reveals within `root`. Returns a cleanup function.
 * @param {HTMLElement} root
 * @param {string[]} [selectors]
 */
export function initScrollReveals(root, selectors = DEFAULT_SELECTORS) {
  if (typeof window === "undefined" || !root) return () => {};
  if (window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
    return () => {};
  }

  const ctx = gsap.context(() => {
    selectors.forEach((sel) => {
      const els = Array.from(root.querySelectorAll(sel));
      if (!els.length) return;
      gsap.set(els, { opacity: 0, y: 22 });
      ScrollTrigger.batch(els, {
        start: "top 88%",
        once: true,
        onEnter: (batch) =>
          gsap.to(batch, {
            opacity: 1,
            y: 0,
            duration: 0.5,
            ease: "power2.out",
            stagger: 0.07,
            overwrite: true,
          }),
      });
    });
  }, root);

  ScrollTrigger.refresh();
  return () => ctx.revert();
}
