<!-- Button — Suluk design primitive (mirrors components.jsx Button variants/sizes). -->
<script>
  import * as Icons from "lucide-svelte";
  let {
    variant = "primary", // primary | deep | accent | ghost | soft | danger
    size = "md", // sm | md | lg
    icon = null, // lucide component
    full = false,
    type = "button",
    onclick = null,
    disabled = false,
    style = "",
    children,
  } = $props();

  const SIZES = { sm: ["7px 12px", 13], md: ["10px 16px", 14], lg: ["13px 22px", 15] };
  const VARIANTS = {
    primary: "background:var(--c-primary);color:#fff;border:1px solid var(--c-primary)",
    deep: "background:var(--c-primary-deep);color:#fff;border:1px solid var(--c-primary-deep)",
    accent: "background:var(--c-accent);color:#fff;border:1px solid var(--c-accent)",
    ghost: "background:transparent;color:var(--c-ink-soft);border:1px solid var(--c-line)",
    soft: "background:var(--c-primary-soft);color:var(--c-primary-deep);border:1px solid transparent",
    danger: "background:var(--c-danger-soft);color:var(--c-danger);border:1px solid transparent",
  };
  let pad = $derived(SIZES[size]?.[0] || SIZES.md[0]);
  let fs = $derived(SIZES[size]?.[1] || SIZES.md[1]);
  let Icon = $derived(icon);
</script>

<button
  type={type === "submit" ? "submit" : type === "reset" ? "reset" : "button"}
  {onclick}
  {disabled}
  class="suluk-btn"
  style="display:inline-flex;align-items:center;justify-content:center;gap:8px;padding:{pad};font-size:{fs}px;font-weight:600;border-radius:var(--radius);{full ? 'width:100%;' : ''}{VARIANTS[variant] || VARIANTS.primary};{disabled ? 'opacity:.55;cursor:not-allowed;' : ''}{style}"
>
  {#if Icon}<Icon size={fs + 3} />{/if}
  {@render children?.()}
</button>
