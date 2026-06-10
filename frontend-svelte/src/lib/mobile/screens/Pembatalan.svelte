<script>
  import { onMount } from "svelte";
  import { Plus } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRp } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MBadge from "../ui/MBadge.svelte";
  import MFormSheet from "../ui/MFormSheet.svelte";

  let { nav } = $props();
  let refunds = $state([]);
  let invoices = $state([]);
  let loading = $state(true);
  let formOpen = $state(false);

  const LABEL = { pending: "Menunggu", approved: "Disetujui", processing: "Diproses", completed: "Dibayar", rejected: "Ditolak", partial: "Sebagian" };
  const label = (s) => LABEL[s] || s || "—";

  async function load() {
    try {
      const [rf, inv] = await Promise.all([ApiService.listRefunds({ limit: 50 }), ApiService.listInvoices({ pageSize: 50 }).catch(() => null)]);
      refunds = rf?.data || rf?.refunds || (Array.isArray(rf) ? rf : []) || [];
      const list = inv?.invoices || inv?.data || (Array.isArray(inv) ? inv : []) || [];
      // refundable = not cancelled, has some value
      invoices = list.filter((i) => i.status !== "batal");
    } catch {
      refunds = [];
    } finally {
      loading = false;
    }
  }
  onMount(load);

  let formFields = $derived([
    { key: "invoice_id", label: "Invoice", type: "select", required: true, options: invoices.map((i) => ({ value: i.id, label: (i.jamaah_name || i.invoice_number || i.id) + " · " + fmtRp(i.total_amount ?? 0) })) },
    { key: "reason", label: "Alasan Pembatalan", type: "textarea", required: true, placeholder: "Sakit / force majeure / kendala dokumen…" },
    { key: "refund_pct", label: "Persentase Refund (%)", type: "number", placeholder: "100" },
  ]);

  async function submit(data) {
    const { invoice_id, reason } = data;
    const body = { reason };
    if (data.refund_pct !== "" && data.refund_pct != null) body.refund_pct = Number(data.refund_pct);
    await ApiService.initiateRefund(invoice_id, body);
    nav.toast("Pengajuan refund dibuat");
    await load();
  }
</script>

{#snippet headerRight()}
  <button type="button" class="m-nav-btn" onclick={() => (formOpen = true)} aria-label="Ajukan refund"><Plus size={22} /></button>
{/snippet}

<MScreen title="Pembatalan & Refund" onBack={nav.back} right={headerRight}>
  <div style="padding:16px 18px 0;display:flex;flex-direction:column;gap:12px">
    {#if loading}
      <div class="m-loading" style="padding:50px 0">Memuat…</div>
    {:else if refunds.length}
      {#each refunds as r (r.id)}
        <div class="m-card m-card-pad">
          <div style="display:flex;justify-content:space-between;align-items:flex-start;margin-bottom:10px">
            <div>
              <div style="font-size:15px;font-weight:700">{r.jamaah_name || r.jamaah || "Jamaah"}</div>
              <div style="font-size:12px;color:var(--c-muted);margin-top:1px">{(r.id || "").toString().slice(0, 12) + (r.package_name ? " · " + r.package_name : "")}</div>
            </div>
            <MBadge status={label(r.status)} />
          </div>
          <div style="background:var(--c-bg);border-radius:11px;padding:11px 13px;display:flex;justify-content:space-between">
            <div><div style="font-size:11px;color:var(--c-faint)">Alasan</div><div style="font-size:13px;font-weight:600;margin-top:1px">{r.reason || r.alasan || "—"}</div></div>
            <div style="text-align:right"><div style="font-size:11px;color:var(--c-faint)">Refund</div><div class="tnum" style="font-size:14px;font-weight:800">{fmtRp(r.amount ?? r.refund_amount ?? 0)}</div></div>
          </div>
        </div>
      {/each}
    {:else}
      <div class="m-empty" style="padding:50px 20px">Tidak ada pembatalan</div>
    {/if}
  </div>
</MScreen>

<MFormSheet open={formOpen} title="Ajukan Refund" fields={formFields} submitLabel="Ajukan" onClose={() => (formOpen = false)} onSubmit={submit} />
