<script>
  import { onMount } from "svelte";
  import { ReceiptText } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRp, fmtRpShort } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MChips from "../ui/MChips.svelte";
  import MGroup from "../ui/MGroup.svelte";
  import MBadge from "../ui/MBadge.svelte";

  let { nav } = $props();
  let all = $state([]);
  let loading = $state(true);
  let tab = $state("Semua");
  let totalTagih = $state(0);
  let totalBayar = $state(0);
  const tabs = ["Semua", "Lunas", "Sebagian", "Belum Bayar", "Jatuh Tempo"];

  // backend status -> display label
  const LABEL = { lunas: "Lunas", sebagian: "Sebagian", belum_bayar: "Belum Bayar", jatuh_tempo: "Jatuh Tempo", batal: "Batal" };
  const label = (s) => LABEL[s] || s || "—";

  onMount(async () => {
    try {
      const [res, summary] = await Promise.all([
        ApiService.listInvoices({ pageSize: 50 }),
        ApiService.getInvoiceSummary().catch(() => null),
      ]);
      all = res?.invoices || res?.data || (Array.isArray(res) ? res : []) || [];
      if (summary) {
        totalTagih = Number(summary.total_amount ?? 0);
        totalBayar = Number(summary.total_paid ?? 0);
      } else {
        // Summary endpoint unreachable — fall back to the old (capped-at-50) sum
        // rather than showing blank KPI tiles.
        totalTagih = all.reduce((s, i) => s + Number(i.total_amount ?? i.jumlah ?? 0), 0);
        totalBayar = all.reduce((s, i) => s + Number(i.amount_paid ?? i.dibayar ?? 0), 0);
      }
    } catch {
      all = [];
    } finally {
      loading = false;
    }
  });
  function isOverdue(iv) {
    return (iv.status === "belum_bayar" || iv.status === "sebagian") && iv.due_date && new Date(iv.due_date) < new Date();
  }
  let rows = $derived(
    all.filter((iv) => {
      if (tab === "Semua") return true;
      if (tab === "Jatuh Tempo") return isOverdue(iv);
      return label(iv.status) === tab;
    })
  );
</script>

<MScreen title="Invoice" onBack={nav.back}>
  <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;padding:16px 18px 0">
    <div class="m-card m-card-pad"><div class="tnum" style="font-size:20px;font-weight:800;letter-spacing:-.02em">{fmtRpShort(totalTagih)}</div><div style="font-size:12px;color:var(--c-muted);margin-top:2px">Total tagihan</div></div>
    <div class="m-card m-card-pad"><div class="tnum" style="font-size:20px;font-weight:800;letter-spacing:-.02em">{fmtRpShort(totalBayar)}</div><div style="font-size:12px;color:var(--c-muted);margin-top:2px">Diterima</div></div>
  </div>
  <div style="padding:14px 0 8px"><MChips {tabs} value={tab} onChange={(v) => (tab = v)} /></div>
  <div style="padding:0 18px">
    {#if loading}
      <div class="m-loading" style="padding:50px 0">Memuat…</div>
    {:else if rows.length}
      <MGroup>
        {#each rows as iv (iv.id)}
          <div class="m-row" role="button" tabindex="0" onclick={() => nav.toast("Detail " + (iv.invoice_number || iv.id))} onkeydown={() => {}}>
            <div class="m-row-ic" style="background:var(--c-bg-2);color:var(--c-ink-soft)"><ReceiptText size={18} /></div>
            <div class="m-row-main">
              <div class="m-row-title">{iv.jamaah_name || iv.jamaah || "Jamaah"}</div>
              <div class="m-row-sub tnum">{(iv.invoice_number || iv.id) + " · " + fmtRp(iv.total_amount ?? iv.jumlah ?? 0)}</div>
            </div>
            <MBadge status={label(iv.status)} />
          </div>
        {/each}
      </MGroup>
    {:else}
      <div class="m-empty">Tidak ada invoice</div>
    {/if}
    <div style="height:24px"></div>
  </div>
</MScreen>
