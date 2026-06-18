<script>
  import { X } from 'lucide-svelte';

  // Renders a centered popup modal (overlay + centered panel). Despite the
  // "SlideDrawer" name (kept so the ~12 pages using it need no import changes),
  // it pops up in the center rather than sliding in from the right — nicer for
  // add/edit forms. `width` is used as the modal's max-width; tall content
  // scrolls inside the panel (capped at 90vh).
  let {
    open = false,
    title = '',
    width = '540px',
    onClose,
    children,
  } = $props();
</script>

{#if open}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6">
    <!-- Backdrop -->
    <button
      type="button"
      class="absolute inset-0 bg-slate-900/40 backdrop-blur-[2px]"
      onclick={onClose}
      aria-label="Tutup"
    ></button>

    <!-- Modal panel -->
    <div
      class="suluk-modal-panel relative z-10 flex max-h-[90vh] w-full flex-col overflow-hidden rounded-2xl bg-white shadow-2xl"
      style="max-width: min({width}, 100%);"
      role="dialog"
      aria-modal="true"
      aria-label={title}
    >
      <!-- Header -->
      {#if title}
        <div class="flex flex-shrink-0 items-center justify-between border-b border-slate-100 px-6 py-4">
          <h2 class="text-base font-semibold text-slate-800">{title}</h2>
          <button
            type="button"
            onclick={onClose}
            class="flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
            aria-label="Tutup"
          >
            <X class="h-5 w-5" />
          </button>
        </div>
      {:else}
        <div class="flex flex-shrink-0 items-center justify-end border-b border-slate-100 px-6 py-3">
          <button
            type="button"
            onclick={onClose}
            class="flex h-8 w-8 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
            aria-label="Tutup"
          >
            <X class="h-5 w-5" />
          </button>
        </div>
      {/if}

      <!-- Scrollable content -->
      <div class="flex-1 overflow-y-auto">
        {@render children?.()}
      </div>
    </div>
  </div>
{/if}

<style>
  .suluk-modal-panel {
    animation: suluk-modal-in 0.18s ease-out;
  }
  @keyframes suluk-modal-in {
    from {
      opacity: 0;
      transform: translateY(8px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .suluk-modal-panel {
      animation: none;
    }
  }
</style>
