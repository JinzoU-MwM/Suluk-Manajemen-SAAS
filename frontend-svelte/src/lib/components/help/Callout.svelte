<script>
  // Kotak sorotan untuk panduan: info | tip | warning. Dipakai oleh GuideBody.
  import { Info, Lightbulb, TriangleAlert } from "lucide-svelte";

  let { variant = "info", text = "" } = $props();

  const META = {
    info: { Icon: Info, label: "Info" },
    tip: { Icon: Lightbulb, label: "Tips" },
    warning: { Icon: TriangleAlert, label: "Perhatian" },
  };

  let meta = $derived(META[variant] ?? META.info);
  let Icon = $derived(meta.Icon);
</script>

<div class="callout callout-{variant}" role="note" aria-label={meta.label}>
  <Icon class="callout-ico" size={18} aria-hidden="true" />
  <p>{text}</p>
</div>

<style>
  .callout {
    display: flex;
    gap: 10px;
    align-items: flex-start;
    padding: 12px 14px;
    border: 1px solid;
    border-radius: var(--radius-md);
    margin: 14px 0;
  }
  .callout :global(.callout-ico) {
    flex-shrink: 0;
    margin-top: 1px;
  }
  .callout p {
    margin: 0;
    font-size: 14px;
    line-height: 1.6;
    color: var(--c-ink-soft);
  }
  .callout-info {
    background: var(--c-info-soft);
    border-color: color-mix(in srgb, var(--c-info) 25%, transparent);
  }
  .callout-info :global(.callout-ico) {
    color: var(--c-info);
  }
  .callout-tip {
    background: var(--c-primary-soft);
    border-color: color-mix(in srgb, var(--c-primary) 25%, transparent);
  }
  .callout-tip :global(.callout-ico) {
    color: var(--c-primary);
  }
  .callout-warning {
    background: var(--c-warning-soft);
    border-color: color-mix(in srgb, var(--c-warning) 30%, transparent);
  }
  .callout-warning :global(.callout-ico) {
    color: var(--c-warning);
  }
</style>
