// GSAP motion for the public landing page (homepage only).
//
// Design notes:
// - Runs client-only (the homepage is prerendered) — call from onMount.
// - Respects `prefers-reduced-motion: reduce`: bails out entirely, leaving the
//   prerendered content fully visible and static.
// - No FOUC / no-JS safety: nothing is hidden via CSS. Start ("hidden") states
//   are applied here in JS only when motion is enabled, so if JS is disabled,
//   fails, or reduced-motion is on, content is never invisible.
// - Everything is created inside a `gsap.context()` scoped to the page root, so
//   teardown is a single `ctx.revert()`.
//
// Transform ownership (so tweens on the same element never fight):
//   hero entrance  → opacity + scale
//   scroll parallax → y
//   mouse parallax  → x
import { gsap } from "gsap";
import { ScrollTrigger } from "gsap/ScrollTrigger";

gsap.registerPlugin(ScrollTrigger);

const EASE = "power2.out";

// Parse a formatted stat ("120rb+", "99,9%", "4,9/5") into a numeric target +
// suffix so we can count up to it while preserving the original formatting.
function parseStat(text) {
  const m = String(text).trim().match(/^(\d+(?:[.,]\d+)?)(.*)$/s);
  if (!m) return null;
  const numStr = m[1].replace(",", ".");
  const decimals = numStr.includes(".") ? numStr.split(".")[1].length : 0;
  return { target: parseFloat(numStr), decimals, suffix: m[2] };
}

// Indonesian decimal comma, no thousands grouping (matches the source values).
function formatStat(value, decimals) {
  return value.toFixed(decimals).replace(".", ",");
}

/**
 * Initialise landing-page motion. Returns a cleanup function.
 * @param {HTMLElement} root - the landing page root element.
 */
export function initLandingMotion(root) {
  if (typeof window === "undefined" || !root) return () => {};
  if (window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
    return () => {};
  }

  let removeMouse = () => {};

  const ctx = gsap.context(() => {
    const q = (sel) => Array.from(root.querySelectorAll(sel));

    // 1) Hero entrance (on load) -------------------------------------------
    const heroBits = [".lp-kicker", ".lp-h1", ".lp-lead", ".lp-hero-cta", ".lp-hero-note"]
      .map((s) => root.querySelector(s))
      .filter(Boolean);
    if (heroBits.length) {
      gsap.from(heroBits, { opacity: 0, y: 22, duration: 0.6, ease: EASE, stagger: 0.09 });
    }
    const mock = root.querySelector(".lp-mock");
    if (mock) {
      gsap.from(mock, { opacity: 0, scale: 0.96, duration: 0.7, ease: EASE, delay: 0.15 });
    }

    // 2) Scroll reveals — each group fades + rises as it scrolls into view.
    //    ScrollTrigger.batch staggers items that enter the viewport together.
    const revealSelectors = [
      ".lp-sec-head",
      ".lp-trust-item",
      ".lp-feature-list > *",
      ".lp-scan-card",
      ".lp-mod",
      ".lp-step",
      ".lp-uc",
      ".lp-pricing > *",
      ".lp-tcard",
      ".lp-phone",
      ".lp-cta-box",
    ];
    revealSelectors.forEach((sel) => {
      const els = q(sel);
      if (!els.length) return;
      gsap.set(els, { opacity: 0, y: 24 });
      ScrollTrigger.batch(els, {
        start: "top 86%",
        once: true,
        onEnter: (batch) =>
          gsap.to(batch, {
            opacity: 1,
            y: 0,
            duration: 0.55,
            ease: EASE,
            stagger: 0.08,
            overwrite: true,
          }),
      });
    });

    // 3) Signature — animated stat counters in the trust bar ----------------
    const trust = root.querySelector(".lp-trust");
    if (trust) {
      ScrollTrigger.create({
        trigger: trust,
        start: "top 80%",
        once: true,
        onEnter: () => {
          q(".lp-trust-item .v").forEach((el) => {
            const parsed = parseStat(el.textContent);
            if (!parsed) return;
            const proxy = { val: 0 };
            gsap.to(proxy, {
              val: parsed.target,
              duration: 1.4,
              ease: "power1.out",
              onUpdate: () => {
                el.textContent = formatStat(proxy.val, parsed.decimals) + parsed.suffix;
              },
            });
          });
        },
      });
    }

    // 4) Signature — light hero parallax (scroll: y, mouse: x) --------------
    const hero = root.querySelector(".lp-hero");
    const floats = q(".lp-float");
    if (hero && mock) {
      const scrollTrigger = { trigger: hero, start: "top top", end: "bottom top", scrub: true };
      gsap.to(mock, { y: 40, ease: "none", scrollTrigger });
      floats.forEach((f, i) => {
        gsap.to(f, { y: i === 0 ? 64 : -52, ease: "none", scrollTrigger });
      });

      const onMove = (e) => {
        const r = hero.getBoundingClientRect();
        const cx = (e.clientX - r.left) / r.width - 0.5; // -0.5..0.5
        gsap.to(mock, { x: cx * 14, duration: 0.5, ease: EASE, overwrite: "auto" });
        floats.forEach((f, i) =>
          gsap.to(f, { x: cx * (i === 0 ? 24 : -20), duration: 0.6, ease: EASE, overwrite: "auto" }),
        );
      };
      hero.addEventListener("mousemove", onMove);
      removeMouse = () => hero.removeEventListener("mousemove", onMove);
    }
  }, root);

  // Recompute trigger positions once layout has settled after hydration.
  ScrollTrigger.refresh();

  return () => {
    removeMouse();
    ctx.revert();
  };
}
