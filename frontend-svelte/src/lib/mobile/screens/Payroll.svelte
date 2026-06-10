<script>
  import { onMount } from "svelte";
  import { Plus } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRp, fmtRpShort } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MGroup from "../ui/MGroup.svelte";
  import MAvatar from "../ui/MAvatar.svelte";
  import MFormSheet from "../ui/MFormSheet.svelte";

  let { nav } = $props();
  let employees = $state([]);
  let summary = $state(null);
  let loading = $state(true);
  let paid = $state({});
  let formOpen = $state(false);

  const FIELDS = [
    { key: "name", label: "Nama Karyawan", required: true },
    { key: "position", label: "Jabatan" },
    { key: "type", label: "Status", type: "select", options: [{ value: "tetap", label: "Tetap" }, { value: "kontrak", label: "Kontrak" }, { value: "harian", label: "Harian" }] },
    { key: "base_salary", label: "Gaji Pokok (Rp)", type: "number" },
    { key: "allowance", label: "Tunjangan (Rp)", type: "number" },
    { key: "phone", label: "No. HP", type: "tel" },
    { key: "email", label: "Email", type: "email" },
  ];

  const nm = (e) => e.name || e.employee_name || "Karyawan";
  const gross = (e) => Number(e.base_salary ?? e.gaji ?? 0) + Number(e.allowance ?? e.tunjangan ?? 0);

  async function load() {
    try {
      const [emps, sum] = await Promise.all([ApiService.listEmployees().catch(() => []), ApiService.getSummary().catch(() => null)]);
      employees = emps?.employees || emps?.data || (Array.isArray(emps) ? emps : []) || [];
      summary = sum;
    } catch {
      employees = [];
    } finally {
      loading = false;
    }
  }
  onMount(load);

  async function submit(data) {
    const payload = { ...data, base_salary: Number(data.base_salary) || 0, allowance: Number(data.allowance) || 0 };
    await ApiService.createEmployee(payload);
    nav.toast("Karyawan ditambahkan");
    await load();
  }

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

{#snippet headerRight()}
  <button type="button" class="m-nav-btn" onclick={() => (formOpen = true)} aria-label="Tambah karyawan"><Plus size={22} /></button>
{/snippet}

<MScreen title="Payroll" onBack={nav.back} right={headerRight}>
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

<MFormSheet open={formOpen} title="Karyawan Baru" fields={FIELDS} submitLabel="Tambah Karyawan" onClose={() => (formOpen = false)} onSubmit={submit} />
