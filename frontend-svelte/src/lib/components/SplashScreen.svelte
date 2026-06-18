<!--
  SplashScreen.svelte — full-screen branded loading splash (the Suluk mark + a
  loading indicator + "Memuat…"). Used for the initial session/app-load guards
  instead of a bare "Memuat…" line. Reduced-motion safe (animations disabled,
  content stays visible).
-->
<script>
  let { label = "Memuat…" } = $props();
</script>

<div class="splash" role="status" aria-live="polite">
  <div class="splash-inner">
    <div class="splash-mark">
      <img src="/brand/suluk-mark-white.png" alt="Suluk" />
    </div>
    <div class="splash-dots" aria-hidden="true">
      <span></span><span></span><span></span>
    </div>
    <p class="splash-label">{label}</p>
  </div>
</div>

<style>
  .splash {
    position: fixed;
    inset: 0;
    z-index: 9999;
    display: flex;
    align-items: center;
    justify-content: center;
    background: radial-gradient(125% 125% at 50% 28%, #1b7f5a 0%, #0f3d2e 68%);
  }
  .splash-inner {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 24px;
    animation: splashIn 0.4s ease-out;
  }
  .splash-mark img {
    height: 64px;
    width: auto;
    display: block;
    animation: splashPulse 1.8s ease-in-out infinite;
  }
  .splash-dots {
    display: flex;
    gap: 8px;
  }
  .splash-dots span {
    width: 9px;
    height: 9px;
    border-radius: 50%;
    background: rgba(255, 255, 255, 0.85);
    animation: splashDot 1.2s ease-in-out infinite;
  }
  .splash-dots span:nth-child(2) {
    animation-delay: 0.16s;
  }
  .splash-dots span:nth-child(3) {
    animation-delay: 0.32s;
  }
  .splash-label {
    margin: 0;
    color: rgba(255, 255, 255, 0.78);
    font-size: 13.5px;
    font-weight: 600;
    letter-spacing: 0.04em;
  }

  @keyframes splashIn {
    from {
      opacity: 0;
      transform: translateY(8px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  @keyframes splashPulse {
    0%,
    100% {
      transform: scale(1);
      opacity: 0.92;
    }
    50% {
      transform: scale(1.06);
      opacity: 1;
    }
  }
  @keyframes splashDot {
    0%,
    100% {
      transform: translateY(0);
      opacity: 0.5;
    }
    50% {
      transform: translateY(-6px);
      opacity: 1;
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .splash-inner,
    .splash-mark img,
    .splash-dots span {
      animation: none;
    }
  }
</style>
