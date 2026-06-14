<script>
  import { onMount } from "svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRp } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MSegmented from "../ui/MSegmented.svelte";
  import MSection from "../ui/MSection.svelte";

  let { nav } = $props();

  const TABS = [
    { value: "neraca", label: "Neraca" },
    { value: "laba", label: "Laba Rugi" },
    { value: "jurnal", label: "Jurnal" },
  ];
  const MODULE_LABEL = {
    invoice: "Invoice/Bayar", vendor: "Vendor", payroll: "Payroll",
    agent: "Komisi", tabungan: "Tabungan", manual: "Manual", opening: "Saldo Awal",
  };

  let tab = $state("neraca");
  let loading = $state(true);
  let neraca = $state(null);
  let laba = $state(null);
  let journals = $state([]);

  async function load(t) {
    loading = true;
    try {
      if (t === "neraca") neraca = await ApiService.getNeraca();
      else if (t === "laba") laba = await ApiService.getLabaRugi();
      else if (t === "jurnal") {
        const r = await ApiService.listJournals({ page: 1, limit: 50 });
        journals = r.items || [];
      }
    } catch (e) {
      nav.toast?.(e?.message || "Gagal memuat akuntansi", "error");
    } finally {
      loading = false;
    }
  }
  function setTab(v) { tab = v; load(v); }
  onMount(() => load(tab));
</script>

<MScreen title="Akuntansi" onBack={nav.back}>
  <div style="padding:14px 18px 0">
    <MSegmented tabs={TABS} value={tab} onChange={setTab} />
  </div>

  {#if loading}
    <div style="padding:40px;text-align:center;color:var(--c-ink-soft)">Memuat…</div>

  {:else if tab === "neraca" && neraca}
    <div style="padding:16px 18px 0">
      <div class="m-card m-card-pad" style="background:linear-gradient(150deg,var(--c-primary-deep),var(--c-primary));border:none;color:#fff">
        <div style="font-size:12.5px;opacity:.85;font-weight:600">Total Aset</div>
        <div class="tnum" style="font-size:28px;font-weight:800;margin-top:4px">{fmtRp(neraca.total_assets)}</div>
        <div style="margin-top:10px;font-size:12px;opacity:.9">
          {neraca.balanced ? "✓ Seimbang" : "⚠ Tidak seimbang"} · per {neraca.as_of}
        </div>
      </div>
    </div>
    <MSection label="Aset" style="padding-top:18px">
      <div class="m-card m-card-pad" style="display:flex;flex-direction:column;gap:9px">
        {#each neraca.assets || [] as a}
          <div style="display:flex;justify-content:space-between;font-size:13px"><span style="color:var(--c-ink-soft)">{a.name}</span><span class="tnum" style="font-weight:700">{fmtRp(a.amount)}</span></div>
        {/each}
      </div>
    </MSection>
    <MSection label="Liabilitas & Ekuitas" style="padding-top:14px">
      <div class="m-card m-card-pad" style="display:flex;flex-direction:column;gap:9px">
        {#each [...(neraca.liabilities || []), ...(neraca.equity || [])] as l}
          <div style="display:flex;justify-content:space-between;font-size:13px"><span style="color:var(--c-ink-soft)">{l.name}</span><span class="tnum" style="font-weight:700">{fmtRp(l.amount)}</span></div>
        {/each}
      </div>
    </MSection>

  {:else if tab === "laba" && laba}
    <div style="padding:16px 18px 0">
      <div class="m-card m-card-pad" style="background:linear-gradient(150deg,var(--c-primary-deep),var(--c-primary));border:none;color:#fff">
        <div style="font-size:12.5px;opacity:.85;font-weight:600">Laba (Rugi) Bersih</div>
        <div class="tnum" style="font-size:28px;font-weight:800;margin-top:4px">{fmtRp(laba.net_income)}</div>
        <div style="display:flex;gap:18px;margin-top:12px">
          <div><div style="font-size:11px;opacity:.8">Pendapatan</div><div class="tnum" style="font-size:14px;font-weight:700">{fmtRp(laba.total_revenue)}</div></div>
          <div><div style="font-size:11px;opacity:.8">Beban</div><div class="tnum" style="font-size:14px;font-weight:700">{fmtRp(laba.total_expense)}</div></div>
        </div>
      </div>
    </div>
    <MSection label="Beban" style="padding-top:18px">
      <div class="m-card m-card-pad" style="display:flex;flex-direction:column;gap:9px">
        {#each laba.expenses || [] as e}
          <div style="display:flex;justify-content:space-between;font-size:13px"><span style="color:var(--c-ink-soft)">{e.name}</span><span class="tnum" style="font-weight:700">{fmtRp(e.amount)}</span></div>
        {:else}
          <div style="font-size:13px;color:var(--c-ink-soft)">Belum ada beban pada periode ini.</div>
        {/each}
      </div>
    </MSection>

  {:else if tab === "jurnal"}
    <MSection label="Jurnal Terbaru" style="padding-top:16px">
      <div class="m-card m-card-pad" style="display:flex;flex-direction:column;gap:11px">
        {#each journals as j}
          <div style="border-bottom:1px solid var(--c-line);padding-bottom:9px">
            <div style="display:flex;justify-content:space-between;font-size:12px;color:var(--c-ink-soft)"><span>{MODULE_LABEL[j.source_module] || j.source_module}</span><span class="tnum">{j.journal_date}</span></div>
            <div style="font-size:13px;font-weight:600;margin-top:2px">{j.description}</div>
          </div>
        {:else}
          <div style="font-size:13px;color:var(--c-ink-soft)">Belum ada jurnal.</div>
        {/each}
      </div>
    </MSection>
  {/if}
</MScreen>
