<!-- Avatar.svelte — initials in a hashed brand-tinted circle (Suluk design). -->
<script>
  let { name = "?", size = 38 } = $props();
  const COLORS = ["#1B7F5A", "#2563c9", "#c79a3e", "#7a5ae0", "#0f7a5a", "#b87708", "#15564a", "#a9842f"];
  let initials = $derived(
    (name || "?").split(" ").filter((w) => !/^(H\.|Hj\.)$/.test(w)).slice(0, 2).map((w) => w[0]).join("").toUpperCase(),
  );
  let color = $derived(
    COLORS[Math.abs([...(name || "?")].reduce((acc, ch) => ch.charCodeAt(0) + ((acc << 5) - acc), 0)) % COLORS.length],
  );
</script>

<div
  class="flex flex-shrink-0 items-center justify-center rounded-full font-bold"
  style="width:{size}px;height:{size}px;background:{color}1f;color:{color};font-size:{Math.round(size * 0.36)}px"
>
  {initials}
</div>
