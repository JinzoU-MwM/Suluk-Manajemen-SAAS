<script>
  import { onMount } from "svelte";
  import { PiggyBank, Plus, X, Check, ArrowDownCircle, Wallet } from "lucide-svelte";
  import PageHeader from "$lib/components/PageHeader.svelte";
  import Card from "$lib/components/ui/Card.svelte";
  import Button from "$lib/components/ui/Button.svelte";
  import IDRInput from "$lib/components/IDRInput.svelte";
  import EmptyState from "$lib/components/EmptyState.svelte";
  import { ApiService } from "$lib/services/api";
  import { showToast } from "$lib/services/toast.svelte.js";
  import { formatRupiah, formatDate } from "$lib/utils/formatting.js";

  let isLoading = $state(true);
  let accounts = $state([]);
  let jamaahList = $state([]);

  // create
  let showCreate = $state(false);
  let cJamaahId = $state("");
  let cTarget = $state(0);
  let cNotes = $state("");
  let saving = $state(false);

  // deposit
  let showDeposit = $state(false);
  let active = $state(null);
  let dAmount = $state(0);
  let dMethod = $state("cash");
  let dRef = $state("");
  let dNotes = $state("");
  let dIdempotencyKey = $state("");

  const STATUS_LABEL = { aktif: "Aktif", converted: "Terkonversi", closed: "Ditutup" };
  const METHODS = [
    { id: "cash", label: "Tunai" },
    { id: "transfer", label: "Transfer Bank" },
    { id: "qris", label: "QRIS" },
  ];

  function progress(a) {
    if (!a.target_amount || a.target_amount <= 0) return 0;
    return Math.min(100, Math.round((a.balance / a.target_amount) * 100));
  }

  async function load() {
    isLoading = true;
    try {
      const r = await ApiService.listTabungan({ page: 1, limit: 100 });
      accounts = r.items || [];
    } catch (e) {
      showToast(e?.message || "Gagal memuat tabungan", "error");
    } finally {
      isLoading = false;
    }
  }

  async function openCreate() {
    showCreate = true;
    if (jamaahList.length === 0) {
      try {
        const data = await ApiService.listJamaah({ pageSize: 200 });
        jamaahList = Array.isArray(data) ? data : data?.data || [];
      } catch {
        /* picker stays empty — name typed manually is not supported, jamaah required */
      }
    }
  }

  async function submitCreate() {
    if (!cJamaahId) return showToast("Pilih jamaah terlebih dahulu", "error");
    saving = true;
    try {
      const j = jamaahList.find((x) => x.id === cJamaahId);
      await ApiService.createTabungan({
        jamaah_id: cJamaahId,
        jamaah_name: j ? j.full_name || j.name || "" : "",
        target_amount: cTarget,
        notes: cNotes,
      });
      showToast("Tabungan dibuat", "success");
      showCreate = false;
      cJamaahId = ""; cTarget = 0; cNotes = "";
      await load();
    } catch (e) {
      showToast(e?.message || "Gagal membuat tabungan", "error");
    } finally {
      saving = false;
    }
  }

  function openDeposit(a) {
    active = a;
    dAmount = 0; dMethod = "cash"; dRef = ""; dNotes = "";
    dIdempotencyKey = crypto.randomUUID();
    showDeposit = true;
  }

  async function submitDeposit() {
    if (dAmount < 1) return showToast("Jumlah setoran minimal Rp 1", "error");
    saving = true;
    try {
      await ApiService.depositTabungan(active.id, {
        amount: dAmount, method: dMethod, reference: dRef, notes: dNotes, idempotency_key: dIdempotencyKey,
      });
      showToast("Setoran tercatat", "success");
      showDeposit = false;
      await load();
    } catch (e) {
      showToast(e?.message || "Gagal mencatat setoran", "error");
    } finally {
      saving = false;
    }
  }

  onMount(load);
</script>

<div class="min-h-screen p-6 lg:p-8" style="background:var(--c-bg)">
  <PageHeader
    kicker="Keuangan"
    title="Tabungan Umrah"
    subtitle="Setoran bertahap jamaah menuju target paket. Setiap setoran otomatis tercatat sebagai jurnal (Dr Kas / Cr Hutang Tabungan)."
  >
    {#snippet actions()}
      <Button variant="primary" icon={Plus} onclick={openCreate}>Buka Tabungan</Button>
    {/snippet}
  </PageHeader>

  {#if isLoading}
    <div class="h-48 animate-pulse rounded-2xl" style="background:var(--c-bg-2,#eef2f0)"></div>
  {:else if accounts.length === 0}
    <Card><EmptyState icon={PiggyBank} title="Belum ada tabungan" text="Buka tabungan untuk jamaah agar bisa menyetor bertahap." /></Card>
  {:else}
    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      {#each accounts as a}
        <Card>
          <div class="flex items-start justify-between gap-2">
            <div>
              <div class="font-serif text-base font-bold">{a.jamaah_name || "—"}</div>
              <span class="mt-1 inline-block rounded-full px-2 py-0.5 text-xs font-semibold"
                style="background:{a.status === 'aktif' ? 'var(--c-primary-soft)' : 'var(--c-bg-2)'};color:{a.status === 'aktif' ? 'var(--c-primary-deep)' : 'var(--c-muted)'}">
                {STATUS_LABEL[a.status] || a.status}
              </span>
            </div>
            <PiggyBank size={22} style="color:var(--c-accent)" />
          </div>

          <div class="mt-4">
            <div class="flex items-baseline justify-between">
              <span class="text-2xl font-bold tabular-nums">{formatRupiah(a.balance)}</span>
              {#if a.target_amount > 0}
                <span class="text-xs text-slate-500">target {formatRupiah(a.target_amount)}</span>
              {/if}
            </div>
            {#if a.target_amount > 0}
              <div class="mt-2 h-2 overflow-hidden rounded-full" style="background:var(--c-bg-2)">
                <div class="h-full rounded-full" style="width:{progress(a)}%;background:var(--c-primary)"></div>
              </div>
              <div class="mt-1 text-right text-xs text-slate-500">{progress(a)}%</div>
            {/if}
          </div>

          {#if a.status === "aktif"}
            <div class="mt-4">
              <Button variant="soft" size="sm" icon={ArrowDownCircle} full onclick={() => openDeposit(a)}>Setor</Button>
            </div>
          {/if}
        </Card>
      {/each}
    </div>
  {/if}
</div>

<!-- Buka Tabungan -->
{#if showCreate}
  <button type="button" class="modal-backdrop" aria-label="Tutup" onclick={() => (showCreate = false)}></button>
  <div class="modal-panel" role="dialog" aria-modal="true" aria-label="Buka Tabungan">
    <div class="modal-head">
      <h3 class="modal-title">Buka Tabungan</h3>
      <button type="button" class="modal-close" onclick={() => (showCreate = false)} aria-label="Tutup"><X size={18} /></button>
    </div>
    <div class="modal-body">
      <div class="field">
        <label for="t-jamaah">Jamaah</label>
        <select id="t-jamaah" bind:value={cJamaahId}>
          <option value="">— Pilih jamaah —</option>
          {#each jamaahList as j}
            <option value={j.id}>{j.full_name || j.name || j.id}</option>
          {/each}
        </select>
      </div>
      <IDRInput label="Target (opsional)" bind:value={cTarget} />
      <div class="field">
        <label for="t-notes">Catatan (opsional)</label>
        <input id="t-notes" type="text" bind:value={cNotes} placeholder="Misal: target keberangkatan 2027…" />
      </div>
    </div>
    <div class="modal-foot">
      <Button variant="ghost" onclick={() => (showCreate = false)}>Batal</Button>
      <Button variant="primary" icon={Check} disabled={saving} onclick={submitCreate}>Simpan</Button>
    </div>
  </div>
{/if}

<!-- Setor -->
{#if showDeposit && active}
  <button type="button" class="modal-backdrop" aria-label="Tutup" onclick={() => (showDeposit = false)}></button>
  <div class="modal-panel" role="dialog" aria-modal="true" aria-label="Setoran Tabungan">
    <div class="modal-head">
      <h3 class="modal-title">Setoran Tabungan</h3>
      <button type="button" class="modal-close" onclick={() => (showDeposit = false)} aria-label="Tutup"><X size={18} /></button>
    </div>
    <div class="modal-body">
      <p class="modal-hint">
        <strong>{active.jamaah_name}</strong> · saldo saat ini <strong>{formatRupiah(active.balance)}</strong>
      </p>
      <IDRInput label="Jumlah Setoran" bind:value={dAmount} required />
      <div class="field">
        <label for="d-method">Metode</label>
        <select id="d-method" bind:value={dMethod}>
          {#each METHODS as m}<option value={m.id}>{m.label}</option>{/each}
        </select>
      </div>
      <div class="field">
        <label for="d-ref">No. Referensi (opsional)</label>
        <input id="d-ref" type="text" bind:value={dRef} placeholder="Nomor bukti transfer…" />
      </div>
      <div class="field">
        <label for="d-notes">Catatan (opsional)</label>
        <input id="d-notes" type="text" bind:value={dNotes} />
      </div>
    </div>
    <div class="modal-foot">
      <Button variant="ghost" onclick={() => (showDeposit = false)}>Batal</Button>
      <Button variant="primary" icon={Check} disabled={saving} onclick={submitDeposit}>Setor</Button>
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
  .field input, .field select {
    width: 100%; padding: 11px 13px; font-size: 13.5px; color: var(--c-ink);
    background: var(--c-surface); border: 1px solid var(--c-line);
    border-radius: var(--radius); outline: none; transition: border-color .15s, box-shadow .15s;
  }
  .field input:focus, .field select:focus {
    border-color: var(--c-primary); box-shadow: 0 0 0 3px var(--c-primary-soft);
  }
</style>
