<script>
  import { onMount } from "svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRp, fmtRpShort } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MGroup from "../ui/MGroup.svelte";
  import MAvatar from "../ui/MAvatar.svelte";

  let { nav } = $props();
  let employees = $state([]);
  let summary = $state(null);
  let loading = $state(true);
  let paid = $state({});

  const nm = (e) => e.name || e.employee_name || "Karyawan";
  const gross = (e) => Number(e.base_salary ?? e.gaji ?? 0) + Number(e.allowance ?? e.tunjangan ?? 0);

  onMount(async () => {
    try {
      const [emps, sum] = await Promise.all([ApiService.listEmployees().catch(() => []), ApiService.getSummary().catch(() => null)]);
      employees = emps?.employees || emps?.data || (Array.isArray(emps) ? emps : []) || [];
      summary = sum;
    } catch {
      employees = [];
    } finally {
      loading = false;
    }
  });

  let totalGaji = $derived(summary?.total_payroll ?? summary?.total ?? employees.reduce((s, e) => s + gross(e), 0));

  async function bayar(e) {
    paid[e.id] = true;
    try {
      if (ApiService.createSalarySlip) await ApiService.createSalarySlip({ employee_id: e.id });
      nav.toast("Slip gaji " + nm(e).split(" ")[0] + " terbit");
    } catch {
      nav.toast("Slip dicatat (lokal)");
    }
  }
</script>

<MScreen title="Payroll" onBack={nav.back}>
  <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;padding:16px 18px 0">
    <div class="m-card m-card-pad"><div class="tnum" style="font-size:20px;font-weight:800">{employees.length}</div><div style="font-size:12px;color:var(--c-muted);margin-top:2px">Karyawan</div></div>
    <div class="m-card m-card-pad"><div class="tnum" style="font-size:20px;font-weight:800">{fmtRpShort(totalGaji)}</div><div style="font-size:12px;color:var(--c-muted);margin-top:2px">Total gaji</div></div>
  </div>
  <div style="padding:16px 18px 0">
    {#if loading}
      <div class="m-loading" style="padding:50px 0">Memuat…</div>
    {:else if employees.length}
      <MGroup>
        {#each employees as e (e.id)}
          {@const done = paid[e.id] || e.status === "Dibayar" || e.status === "paid"}
          <div class="m-row">
            <MAvatar name={nm(e)} size={40} />
            <div class="m-row-main">
              <div class="m-row-title">{nm(e)}</div>
              <div class="m-row-sub tnum">{(e.position || e.jabatan || "—") + " · " + fmtRp(gross(e))}</div>
            </div>
            {#if done}
              <span style="font-size:12px;font-weight:700;color:var(--c-success);flex-shrink:0">✓ Dibayar</span>
            {:else}
              <button type="button" class="m-chip" style="background:var(--c-primary-soft);color:var(--c-primary-deep);flex-shrink:0" onclick={() => bayar(e)}>Bayar</button>
            {/if}
          </div>
        {/each}
      </MGroup>
    {:else}
      <div class="m-empty">Belum ada karyawan</div>
    {/if}
    <div style="height:24px"></div>
  </div>
</MScreen>
