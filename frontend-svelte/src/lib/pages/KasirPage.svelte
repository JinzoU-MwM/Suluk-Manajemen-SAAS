<script>
  import { onMount } from "svelte";
  import { Wallet, X, Check, LockOpen, Lock, History } from "lucide-svelte";
  import PageHeader from "$lib/components/PageHeader.svelte";
  import Card from "$lib/components/ui/Card.svelte";
  import Button from "$lib/components/ui/Button.svelte";
  import IDRInput from "$lib/components/IDRInput.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { ApiService } from "$lib/services/api";
  import { showToast } from "$lib/services/toast.svelte.js";
  import { formatRupiah, formatDateTime } from "$lib/utils/formatting.js";

  let isLoading = $state(true);
  let activeSession = $state(null);
  let history = $state([]);
  let saving = $state(false);

  // open
  let showOpen = $state(false);
  let openingFloat = $state(0);
  let openNotes = $state("");

  // close
  let showClose = $state(false);
  let countedCash = $state(0);

  const STATUS_LABEL = { open: "Terbuka", closed: "Tertutup" };

  async function load() {
    isLoading = true;
    try {
      const [act, hist] = await Promise.all([
        ApiService.getActiveCashSession().catch(() => null),
        ApiService.listCashSessions({ limit: 30 }).catch(() => []),
      ]);
      activeSession = act && act.id ? act : null;
      history = hist || [];
    } catch (e) {
      showToast(e?.message || "Gagal memuat sesi kas", "error");
    } finally {
      isLoading = false;
    }
  }

  async function submitOpen() {
    saving = true;
    try {
      await ApiService.openCashSession({ opening_float: openingFloat, notes: openNotes });
      showToast("Kas dibuka", "success");
      showOpen = false; openingFloat = 0; openNotes = "";
      await load();
    } catch (e) {
      showToast(e?.message || "Gagal membuka kas", "error");
    } finally {
      saving = false;
    }
  }

  function openClose() {
    countedCash = activeSession?.expected_cash ?? 0;
    showClose = true;
  }

  // Live preview of over/short before the user confirms.
  let previewDiff = $derived(
    activeSession ? countedCash - (activeSession.expected_cash ?? activeSession.opening_float ?? 0) : 0,
  );
  let previewDiffLabel = $derived(diffLabel(previewDiff));

  async function submitClose() {
    saving = true;
    try {
      await ApiService.closeCashSession(activeSession.id, { counted_cash: countedCash });
      showToast("Kas ditutup", "success");
      showClose = false;
      await load();
    } catch (e) {
      showToast(e?.message || "Gagal menutup kas", "error");
    } finally {
      saving = false;
    }
  }

  function diffLabel(d) {
    if (d === 0 || d == null) return { text: "Pas", color: "var(--c-primary-deep)", bg: "var(--c-primary-soft)" };
    if (d > 0) return { text: "Lebih " + formatRupiah(d), color: "var(--c-primary-deep)", bg: "var(--c-primary-soft)" };
    return { text: "Kurang " + formatRupiah(-d), color: "var(--c-danger)", bg: "var(--c-danger-soft)" };
  }

  onMount(load);
</script>

<div class="min-h-screen p-6 lg:p-8" style="background:var(--c-bg)">
  <PageHeader
    kicker="Keuangan"
    title="Kasir (POS)"
    subtitle="Sesi laci kas harian — buka kas dengan modal awal, tutup kas dengan hitungan fisik. Selisih lebih/kurang otomatis dijurnal."
  >
    {#snippet actions()}
      {#if activeSession}
        <Button variant="danger" icon={Lock} onclick={openClose}>Tutup Kas</Button>
      {:else}
        <Button variant="primary" icon={LockOpen} onclick={() => (showOpen = true)}>Buka Kas</Button>
      {/if}
    {/snippet}
  </PageHeader>

  {#if isLoading}
    <div class="h-40 animate-pulse rounded-2xl" style="background:var(--c-bg-2,#eef2f0)"></div>
  {:else}
    {#if activeSession}
      <Card class="mb-6">
        <div class="flex items-center justify-between">
          <div>
            <div class="flex items-center gap-2">
              <span class="inline-block h-2.5 w-2.5 rounded-full" style="background:var(--c-primary)"></span>
              <span class="font-serif text-lg font-bold">Sesi kas terbuka</span>
            </div>
            <p class="mt-1 text-xs text-slate-500">Dibuka {formatDateTime(activeSession.opened_at)}</p>
          </div>
          <Wallet size={28} style="color:var(--c-accent)" />
        </div>
        <div class="mt-4 grid grid-cols-2 gap-4">
          <div>
            <div class="text-xs text-slate-500">Modal Awal</div>
            <div class="text-xl font-bold tabular-nums">{formatRupiah(activeSession.opening_float)}</div>
          </div>
          <div>
            <div class="text-xs text-slate-500">Perkiraan Kas</div>
            <div class="text-xl font-bold tabular-nums">{formatRupiah(activeSession.expected_cash ?? activeSession.opening_float)}</div>
          </div>
        </div>
      </Card>
    {/if}

    <h3 class="mb-3 flex items-center gap-2 font-serif text-base font-bold">
      <History size={18} /> Riwayat Sesi
    </h3>
    {#if history.length === 0}
      <Card><EmptyState icon={Wallet} title="Belum ada sesi kas" text="Buka kas untuk memulai sesi penjualan hari ini." /></Card>
    {:else}
      <Card pad={false} class="overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-slate-500" style="border-bottom:1px solid var(--c-line)">
                <th class="px-4 py-3 font-semibold">Dibuka</th>
                <th class="px-4 py-3 font-semibold">Ditutup</th>
                <th class="px-4 py-3 font-semibold text-right">Modal</th>
                <th class="px-4 py-3 font-semibold text-right">Perkiraan</th>
                <th class="px-4 py-3 font-semibold text-right">Dihitung</th>
                <th class="px-4 py-3 font-semibold">Selisih</th>
                <th class="px-4 py-3 font-semibold">Status</th>
              </tr>
            </thead>
            <tbody>
              {#each history as s}
                {@const d = diffLabel(s.difference)}
                <tr style="border-top:1px solid var(--c-line)">
                  <td class="px-4 py-3">{formatDateTime(s.opened_at)}</td>
                  <td class="px-4 py-3">{s.closed_at ? formatDateTime(s.closed_at) : "—"}</td>
                  <td class="px-4 py-3 text-right tabular-nums">{formatRupiah(s.opening_float)}</td>
                  <td class="px-4 py-3 text-right tabular-nums">{s.expected_cash != null ? formatRupiah(s.expected_cash) : "—"}</td>
                  <td class="px-4 py-3 text-right tabular-nums">{s.counted_cash != null ? formatRupiah(s.counted_cash) : "—"}</td>
                  <td class="px-4 py-3">
                    {#if s.status === "closed"}
                      <span class="rounded-full px-2 py-0.5 text-xs font-semibold" style="background:{d.bg};color:{d.color}">{d.text}</span>
                    {:else}—{/if}
                  </td>
                  <td class="px-4 py-3">{STATUS_LABEL[s.status] || s.status}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </Card>
    {/if}
  {/if}
</div>

<!-- Buka Kas -->
{#if showOpen}
  <button type="button" class="modal-backdrop" aria-label="Tutup" onclick={() => (showOpen = false)}></button>
  <div class="modal-panel" role="dialog" aria-modal="true" aria-label="Buka Kas">
    <div class="modal-head">
      <h3 class="modal-title">Buka Kas</h3>
      <button type="button" class="modal-close" onclick={() => (showOpen = false)} aria-label="Tutup"><X size={18} /></button>
    </div>
    <div class="modal-body">
      <p class="modal-hint">Masukkan modal awal (uang di laci saat mulai shift).</p>
      <IDRInput label="Modal Awal" bind:value={openingFloat} required />
      <div class="field">
        <label for="o-notes">Catatan (opsional)</label>
        <input id="o-notes" type="text" bind:value={openNotes} placeholder="Misal: shift pagi…" />
      </div>
    </div>
    <div class="modal-foot">
      <Button variant="ghost" onclick={() => (showOpen = false)}>Batal</Button>
      <Button variant="primary" icon={Check} disabled={saving} onclick={submitOpen}>Buka</Button>
    </div>
  </div>
{/if}

<!-- Tutup Kas -->
{#if showClose && activeSession}
  <button type="button" class="modal-backdrop" aria-label="Tutup" onclick={() => (showClose = false)}></button>
  <div class="modal-panel" role="dialog" aria-modal="true" aria-label="Tutup Kas">
    <div class="modal-head">
      <h3 class="modal-title">Tutup Kas</h3>
      <button type="button" class="modal-close" onclick={() => (showClose = false)} aria-label="Tutup"><X size={18} /></button>
    </div>
    <div class="modal-body">
      <p class="modal-hint">
        Perkiraan kas di laci: <strong>{formatRupiah(activeSession.expected_cash ?? activeSession.opening_float)}</strong>.
        Masukkan jumlah uang fisik yang dihitung.
      </p>
      <IDRInput label="Kas Dihitung" bind:value={countedCash} required />
      <div class="flex items-center justify-between rounded-xl px-3 py-2.5" style="background:{previewDiffLabel.bg}">
        <span class="text-sm font-semibold" style="color:{previewDiffLabel.color}">Selisih</span>
        <span class="text-sm font-bold tabular-nums" style="color:{previewDiffLabel.color}">{previewDiffLabel.text}</span>
      </div>
    </div>
    <div class="modal-foot">
      <Button variant="ghost" onclick={() => (showClose = false)}>Batal</Button>
      <Button variant="primary" icon={Check} disabled={saving} onclick={submitClose}>Tutup Kas</Button>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed; inset: 0; z-index: 95;
    background: rgba(16,33,28,.4); backdrop-filter: blur(2px);
    border: none; cursor: pointer; animation: suluk-fade .2s ease;
  }
  .modal-panel {
    position: fixed; z-index: 96; top: 50%; left: 50%; transform: translate(-50%, -50%);
    width: 460px; max-width: 94vw; background: var(--c-surface);
    border-radius: var(--radius-xl); box-shadow: var(--shadow-lg); overflow: hidden;
    animation: suluk-scale .25s cubic-bezier(.2,.7,.3,1) both;
  }
  @keyframes suluk-fade { from { opacity: 0; } to { opacity: 1; } }
  @keyframes suluk-scale {
    from { opacity: 0; transform: translate(-50%, -50%) scale(.96); }
    to { opacity: 1; transform: translate(-50%, -50%) scale(1); }
  }
  .modal-head { padding: 22px 26px 0; display: flex; justify-content: space-between; align-items: flex-start; }
  .modal-title { font-size: 18px; font-weight: 800; color: var(--c-ink); }
  .modal-close {
    display: flex; align-items: center; justify-content: center;
    width: 32px; height: 32px; border-radius: 8px; color: var(--c-faint);
    transition: background .15s, color .15s;
  }
  .modal-close:hover { background: var(--c-bg-2); color: var(--c-ink); }
  .modal-body { padding: 14px 26px 24px; display: flex; flex-direction: column; gap: 16px; }
  .modal-hint { font-size: 13.5px; color: var(--c-muted); }
  .modal-hint strong { color: var(--c-ink); font-variant-numeric: tabular-nums; }
  .modal-foot {
    padding: 16px 26px; border-top: 1px solid var(--c-line);
    display: flex; gap: 10px; justify-content: flex-end; background: var(--c-bg);
  }
  .field { display: flex; flex-direction: column; gap: 6px; }
  .field label { font-size: 12.5px; font-weight: 700; color: var(--c-ink-soft); }
  .field input {
    width: 100%; padding: 11px 13px; font-size: 13.5px; color: var(--c-ink);
    background: var(--c-surface); border: 1px solid var(--c-line);
    border-radius: var(--radius); outline: none; transition: border-color .15s, box-shadow .15s;
  }
  .field input:focus { border-color: var(--c-primary); box-shadow: 0 0 0 3px var(--c-primary-soft); }
</style>
