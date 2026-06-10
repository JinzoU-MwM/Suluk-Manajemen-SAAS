<script>
  let { to = 0, dur = 900, format = null } = $props();
  let val = $state(0);
  $effect(() => {
    let raf, start;
    const target = to;
    const from = 0;
    function tick(t) {
      if (!start) start = t;
      const p = Math.min(1, (t - start) / dur);
      const e = 1 - Math.pow(1 - p, 3); // easeOutCubic
      val = from + (target - from) * e;
      if (p < 1) raf = requestAnimationFrame(tick);
      else val = target;
    }
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  });
</script>

<span class="tnum">{format ? format(val) : Math.round(val).toLocaleString("id-ID")}</span>
