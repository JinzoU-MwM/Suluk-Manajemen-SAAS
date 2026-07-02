<script>
  import { onMount } from "svelte";
  import { Plus } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRp } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MSection from "../ui/MSection.svelte";
  import MFormSheet from "../ui/MFormSheet.svelte";

  let { nav } = $props();

  const STATUS_LABEL = { active: "Aktif", converted: "Terkonversi", closed: "Ditutup" };

  let loading = $state(true);
  let accounts = $state([]);
  let jamaahOpts = $state([]);

  let showCreate = $state(false);
  let showDeposit = $state(false);
  let active = $state(null);
  let dIdempotencyKey = $state("");

  function progress(a) {
    if (!a.target_amount || a.target_amount <= 0) return 0;
    return Math.min(100, Math.round((a.balance / a.target_amount) * 100));
  }

  async function load() {
    loading = true;
    try {
      const r = await ApiService.listTabungan({ page: 1, limit: 100 });
      accounts = r.items || [];
    } catch (e) {
      nav.toast?.(e?.message || "Gagal memuat tabungan", "error");
    } finally {
      loading = false;
    }
  }

  async function openCreate() {
    if (jamaahOpts.length === 0) {
      try {
        const data = await ApiService.listJamaah({ pageSize: 200 });
        const list = Array.isArray(data) ? data : data?.data || [];
        jamaahOpts = list.map((j) => ({ value: j.id, label: j.full_name || j.name || j.id }));
      } catch { /* ignore */ }
    }
    showCreate = true;
  }

  async function submitCreate(v) {
    const j = jamaahOpts.find((o) => o.value === v.jamaah_id);
    await ApiService.createTabungan({
      jamaah_id: v.jamaah_id,
      jamaah_name: j ? j.label : "",
      target_amount: parseInt(v.target_amount || "0", 10) || 0,
      notes: v.notes || "",
    });
    nav.toast?.("Tabungan dibuka", "success");
    await load();
  }

  function openDeposit(a) {
    active = a;
    dIdempotencyKey = crypto.randomUUID();
    showDeposit = true;
  }

  async function submitDeposit(v) {
    const amount = parseInt(v.amount || "0", 10) || 0;
    if (amount < 1) throw new Error("Jumlah setoran minimal Rp 1");
    await ApiService.depositTabungan(active.id, {
      amount, method: v.method || "cash", reference: v.reference || "", notes: v.notes || "", idempotency_key: dIdempotencyKey,
    });
    nav.toast?.("Setoran tercatat", "success");
    await load();
  }

  onMount(load);
</script>

<MScreen title="Tabungan" onBack={nav.back}>
  {#snippet right()}
    <button type="button" class="m-nav-btn" onclick={openCreate}><Plus size={18} /></button>
  {/snippet}

  {#if loading}
    <div style="padding:40px;text-align:center;color:var(--c-ink-soft)">Memuat…</div>
  {:else if accounts.length === 0}
    <div style="padding:48px 24px;text-align:center;color:var(--c-ink-soft)">
      Belum ada tabungan.<br />Ketuk + untuk membuka tabungan jamaah.
    </div>
  {:else}
    <MSection label="Tabungan Jamaah" style="padding-top:16px">
      <div style="display:flex;flex-direction:column;gap:11px">
        {#each accounts as a}
          <button type="button" class="m-card m-card-pad" style="text-align:left;width:100%"
            onclick={() => a.status === "active" && openDeposit(a)}>
            <div style="display:flex;justify-content:space-between;align-items:flex-start">
              <span style="font-weight:700;font-size:14px">{a.jamaah_name || "—"}</span>
              <span style="font-size:11px;font-weight:700;color:var(--c-ink-soft)">{STATUS_LABEL[a.status] || a.status}</span>
            </div>
            <div class="tnum" style="font-size:20px;font-weight:800;margin-top:6px">{fmtRp(a.balance)}</div>
            {#if a.target_amount > 0}
              <div style="height:7px;border-radius:99px;background:var(--c-bg-2);overflow:hidden;margin-top:8px">
                <div style="height:100%;border-radius:99px;background:var(--c-primary);width:{progress(a)}%"></div>
              </div>
              <div style="display:flex;justify-content:space-between;font-size:11px;color:var(--c-ink-soft);margin-top:4px">
                <span>{progress(a)}%</span><span>target {fmtRp(a.target_amount)}</span>
              </div>
            {/if}
          </button>
        {/each}
      </div>
    </MSection>
  {/if}
</MScreen>

<MFormSheet
  open={showCreate}
  title="Buka Tabungan"
  submitLabel="Simpan"
  fields={[
    { key: "jamaah_id", label: "Jamaah", type: "select", required: true, options: jamaahOpts },
    { key: "target_amount", label: "Target (opsional)", type: "number", placeholder: "0" },
    { key: "notes", label: "Catatan (opsional)", type: "text" },
  ]}
  onClose={() => (showCreate = false)}
  onSubmit={submitCreate}
/>

<MFormSheet
  open={showDeposit}
  title={active ? "Setor — " + (active.jamaah_name || "") : "Setor"}
  submitLabel="Setor"
  fields={[
    { key: "amount", label: "Jumlah Setoran", type: "number", required: true, placeholder: "0" },
    { key: "method", label: "Metode", type: "select", options: [
      { value: "cash", label: "Tunai" }, { value: "transfer", label: "Transfer Bank" }, { value: "qris", label: "QRIS" },
    ] },
    { key: "reference", label: "No. Referensi (opsional)", type: "text" },
    { key: "notes", label: "Catatan (opsional)", type: "text" },
  ]}
  onClose={() => (showDeposit = false)}
  onSubmit={submitDeposit}
/>
