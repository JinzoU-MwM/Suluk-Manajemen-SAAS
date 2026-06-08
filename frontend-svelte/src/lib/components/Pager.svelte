<!--
  Pager.svelte — simple prev/next pager with range + page count.
  Props: page (1-based), pageSize, total, onchange(nextPage). Hidden when it fits one page.
-->
<script>
  import { ChevronLeft, ChevronRight } from "lucide-svelte";

  let { page = 1, pageSize = 25, total = 0, onchange = null } = $props();

  let totalPages = $derived(Math.max(1, Math.ceil(total / pageSize)));
  let from = $derived(total === 0 ? 0 : (page - 1) * pageSize + 1);
  let to = $derived(Math.min(total, page * pageSize));

  function go(p) {
    const next = Math.min(totalPages, Math.max(1, p));
    if (next !== page) onchange?.(next);
  }
</script>

{#if total > pageSize}
  <div class="flex items-center justify-between gap-3 px-1 py-3 text-sm text-slate-500">
    <span>Menampilkan {from}–{to} dari {total}</span>
    <div class="flex items-center gap-1">
      <button
        type="button"
        class="flex h-8 w-8 items-center justify-center rounded-lg border border-slate-200 text-slate-600 transition-colors hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
        onclick={() => go(page - 1)}
        disabled={page <= 1}
        aria-label="Halaman sebelumnya"
      >
        <ChevronLeft class="h-4 w-4" />
      </button>
      <span class="px-2 font-medium text-slate-700">{page} / {totalPages}</span>
      <button
        type="button"
        class="flex h-8 w-8 items-center justify-center rounded-lg border border-slate-200 text-slate-600 transition-colors hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-40"
        onclick={() => go(page + 1)}
        disabled={page >= totalPages}
        aria-label="Halaman berikutnya"
      >
        <ChevronRight class="h-4 w-4" />
      </button>
    </div>
  </div>
{/if}
