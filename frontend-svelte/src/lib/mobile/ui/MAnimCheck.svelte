<script>
  let { size = 60, color = "var(--c-primary)", bg = "#fff" } = $props();
</script>

<div
  class="m-pop"
  style="width:{size}px;height:{size}px;border-radius:50%;background:{bg};display:flex;align-items:center;justify-content:center"
>
  <svg width={size * 0.56} height={size * 0.56} viewBox="0 0 60 60" fill="none">
    <circle class="m-check-c" cx="30" cy="30" r="26" stroke={color} stroke-width="4" stroke-linecap="round" />
    <path class="m-check-p" d="M18 31l8 8 16-17" stroke={color} stroke-width="5" stroke-linecap="round" stroke-linejoin="round" />
  </svg>
</div>
