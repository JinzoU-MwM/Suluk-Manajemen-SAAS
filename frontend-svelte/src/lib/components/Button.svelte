<!--
  Button.svelte — consistent button built on the app.css .btn classes.
  Props: variant (primary|secondary|danger|ghost), size (sm|md), loading, disabled.
  Any extra props (onclick, title, aria-*) pass through to the <button>.
-->
<script>
  import { Loader2 } from "lucide-svelte";

  /**
   * @type {{
   *   variant?: 'primary'|'secondary'|'danger'|'ghost',
   *   size?: 'sm'|'md',
   *   type?: 'button'|'submit'|'reset',
   *   loading?: boolean,
   *   disabled?: boolean,
   *   class?: string,
   *   children?: import('svelte').Snippet,
   *   [key: string]: any
   * }}
   */
  let {
    variant = "primary",
    size = "md",
    type = "button",
    loading = false,
    disabled = false,
    class: klass = "",
    children,
    ...rest
  } = $props();

  const variantClass = {
    primary: "btn-primary",
    secondary: "btn-secondary",
    danger: "btn-danger",
    ghost: "text-slate-600 hover:bg-slate-100",
  };
</script>

<button
  {type}
  class={`btn ${variantClass[variant] ?? "btn-primary"} ${size === "sm" ? "btn-sm" : ""} disabled:cursor-not-allowed disabled:opacity-50 ${klass}`}
  disabled={disabled || loading}
  {...rest}
>
  {#if loading}
    <Loader2 class="h-4 w-4 animate-spin" />
  {/if}
  {@render children?.()}
</button>
