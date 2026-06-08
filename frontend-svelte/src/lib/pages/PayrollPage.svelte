<script>
  import { onMount } from 'svelte';
  import {
    Users, Receipt, Wallet, Plus,
    Pencil, FileText, Banknote, Clock,
  } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import PageHeader from '../components/PageHeader.svelte';
  import StatCard from '../components/StatCard.svelte';
  import Avatar from '../components/Avatar.svelte';
  import { showToast } from '../services/toast.svelte.js';
  import { ApiService, authHeaders } from '../services/api.js';

  let { onNavigate, user = null } = $props();

  let employees = $state([]);
  let slips = $state([]);
  let advances = $state([]);
  let summary = $state({ total_employees: 0, active_employees: 0, total_advances: 0, outstanding_advances: 0, monthly_payroll: 0 });
  let loading = $state(true);
  let tab = $state('employees');

  let showEmployeeDrawer = $state(false);
  let editingEmployee = $state(null);
  let employeeForm = $state({ name: '', position: '', type: 'tetap', base_salary: 0, allowance: 0, bpjs_tk: 0, bpjs_kes: 0, pph21_rate: 0, phone: '', email: '' });
  let savingEmployee = $state(false);

  let showSlipDrawer = $state(false);
  let slipForm = $state({ employee_id: '', period: new Date().toISOString().slice(0, 7), package_id: '', notes: '' });
  let savingSlip = $state(false);

  let showAdvanceDrawer = $state(false);
  let advanceForm = $state({ employee_id: '', amount: 0, reason: '' });
  let savingAdvance = $state(false);

  let showRepayDrawer = $state(false);
  let repayingAdvance = $state(null);
  let repayForm = $state({ amount: 0, salary_slip_id: '' });
  let savingRepay = $state(false);

  function formatIDR(n) { return n ? `Rp ${Number(n).toLocaleString('id-ID')}` : 'Rp 0'; }
  function formatDate(d) { return d ? new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) : '-'; }

  async function loadData() {
    loading = true;
    try {
      const [empData, slipData, advData, sumData] = await Promise.all([
        ApiService.listEmployees(),
        ApiService.listSalarySlips(),
        ApiService.listAdvances(),
        ApiService.getSummary(),
      ]);
      employees = empData.employees || [];
      slips = slipData.slips || [];
      advances = advData.advances || [];
      summary = sumData;
    } catch (e) {
      showToast(e.message, 'error');
    } finally {
      loading = false;
    }
  }

  onMount(() => { loadData(); });

  function openNewEmployee() { editingEmployee = null; employeeForm = { name: '', position: '', type: 'tetap', base_salary: 0, allowance: 0, bpjs_tk: 0, bpjs_kes: 0, pph21_rate: 0, phone: '', email: '' }; showEmployeeDrawer = true; }
  function editEmployee(e) { editingEmployee = e; employeeForm = { name: e.name, position: e.position, type: e.type, base_salary: e.base_salary, allowance: e.allowance, bpjs_tk: e.bpjs_tk, bpjs_kes: e.bpjs_kes, pph21_rate: e.pph21_rate, phone: e.phone, email: e.email }; showEmployeeDrawer = true; }

  async function saveEmployee() {
    savingEmployee = true;
    try {
      if (editingEmployee) { await ApiService.updateEmployee(editingEmployee.id, employeeForm); showToast('Karyawan diperbarui'); }
      else { await ApiService.createEmployee(employeeForm); showToast('Karyawan ditambahkan'); }
      showEmployeeDrawer = false; await loadData();
    } catch (e) { showToast(e.message, 'error'); } finally { savingEmployee = false; }
  }

  async function generateSlip() {
    savingSlip = true;
    try {
      await ApiService.createSalarySlip(slipForm);
      showToast('Slip gaji dibuat'); showSlipDrawer = false; await loadData();
    } catch (e) { showToast(e.message, 'error'); } finally { savingSlip = false; }
  }

  async function finalizeSlip(id) {
    try { await ApiService.finalizeSlip(id); showToast('Slip difinalisasi'); await loadData(); }
    catch (e) { showToast(e.message, 'error'); }
  }

  async function createAdvance() {
    savingAdvance = true;
    try {
      await ApiService.createAdvance(advanceForm);
      showToast('Kasbon dicatat'); showAdvanceDrawer = false; await loadData();
    } catch (e) { showToast(e.message, 'error'); } finally { savingAdvance = false; }
  }

  function openRepay(a) { repayingAdvance = a; repayForm = { amount: a.remaining, salary_slip_id: '' }; showRepayDrawer = true; }

  async function submitRepay() {
    savingRepay = true;
    try {
      await ApiService.repayAdvance(repayingAdvance.id, repayForm);
      showToast('Pembayaran kasbon dicatat'); showRepayDrawer = false; await loadData();
    } catch (e) { showToast(e.message, 'error'); } finally { savingRepay = false; }
  }

  async function downloadSlipPDF(slip) {
    if (!slip) return;
    try {
      showToast('Menyiapkan slip gaji PDF...');
      const url = `/api/payroll/slips/${slip.id}/pdf`;
      const res = await fetch(url, { headers: authHeaders() });
      if (!res.ok) {
        const err = await res.text();
        showToast(err || 'Gagal mengunduh PDF', 'error');
        return;
      }
      const blob = await res.blob();
      const blobUrl = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = blobUrl;
      a.download = `slip_${slip.employee_name.replace(/\s+/g, '_')}_${slip.period}.pdf`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(blobUrl);
      showToast('Slip gaji PDF berhasil diunduh');
    } catch (e) {
      showToast('Gagal mengunduh: ' + e.message, 'error');
    }
  }
</script>

<div class="flex flex-col gap-6 p-4 lg:p-8">
  <PageHeader kicker="Penggajian" title="Penggajian" subtitle="Kelola karyawan, gaji, dan kasbon." />

  {#if loading}
    <div class="flex items-center justify-center py-16"><div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-600 border-t-transparent"></div></div>
  {:else}
    <!-- Summary -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard icon={Users} label="Karyawan Aktif" value={`${summary.active_employees}/${summary.total_employees}`} accent="#1B7F5A" />
      <StatCard icon={Banknote} label="Gaji Bulan Ini" value={formatIDR(summary.monthly_payroll)} accent="#2563a8" />
      <StatCard icon={Clock} label="Kasbon Outstanding" value={formatIDR(summary.outstanding_advances)} accent="#C99A2E" />
      <StatCard icon={Wallet} label="Total Kasbon" value={formatIDR(summary.total_advances)} accent="#1B7F5A" />
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 rounded-xl bg-slate-100 p-1 w-fit">
      <button type="button" onclick={() => tab = 'employees'} class="rounded-lg px-4 py-2 text-xs font-semibold {tab === 'employees' ? 'bg-white text-slate-800 shadow-sm' : 'text-slate-500'}">Karyawan</button>
      <button type="button" onclick={() => tab = 'slips'} class="rounded-lg px-4 py-2 text-xs font-semibold {tab === 'slips' ? 'bg-white text-slate-800 shadow-sm' : 'text-slate-500'}">Slip Gaji</button>
      <button type="button" onclick={() => tab = 'advances'} class="rounded-lg px-4 py-2 text-xs font-semibold {tab === 'advances' ? 'bg-white text-slate-800 shadow-sm' : 'text-slate-500'}">Kasbon</button>
    </div>

    <!-- Employees Tab -->
    {#if tab === 'employees'}
      <div class="flex justify-end"><button type="button" onclick={openNewEmployee} class="flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700"><Plus class="h-4 w-4" /> Tambah Karyawan</button></div>
      <div class="overflow-hidden rounded-2xl border border-slate-200/70 bg-white shadow-sm">
        {#if employees.length === 0}
          <div class="flex flex-col items-center justify-center py-16 text-slate-400"><Users class="h-10 w-10 mb-2" /><p class="text-sm">Belum ada karyawan</p></div>
        {:else}
          <table class="w-full text-sm">
            <thead><tr class="border-b border-slate-100"><th class="px-4 py-3 text-left text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Nama</th><th class="px-4 py-3 text-left text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Jabatan</th><th class="px-4 py-3 text-left text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Tipe</th><th class="px-4 py-3 text-right text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Gaji Pokok</th><th class="px-4 py-3 text-right text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Tunjangan</th><th class="px-4 py-3 text-right text-[11.5px] font-semibold uppercase tracking-wider text-slate-400"></th></tr></thead>
            <tbody>
              {#each employees as e}
                <tr class="transition-colors hover:bg-primary-50/30">
                  <td class="border-b border-slate-100 px-4 py-3.5"><div class="flex items-center gap-3"><Avatar name={e.name} size={38} /><span class="font-bold text-[#10211c]">{e.name}</span></div></td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-slate-600">{e.position}</td>
                  <td class="border-b border-slate-100 px-4 py-3.5"><span class="rounded-full px-2 py-0.5 text-xs font-medium {e.type === 'tetap' ? 'bg-primary-50 text-primary-700' : 'bg-purple-50 text-purple-700'}">{e.type === 'tetap' ? 'Tetap' : 'Freelance'}</span></td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-right font-semibold text-slate-700" style="font-variant-numeric:tabular-nums">{formatIDR(e.base_salary)}</td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-right text-slate-600" style="font-variant-numeric:tabular-nums">{formatIDR(e.allowance)}</td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-right"><button type="button" onclick={() => editEmployee(e)} class="rounded-lg p-1.5 text-slate-400 hover:bg-slate-100"><Pencil class="h-4 w-4" /></button></td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </div>
    {/if}

    <!-- Slips Tab -->
    {#if tab === 'slips'}
      <div class="flex justify-end"><button type="button" onclick={() => { slipForm = { employee_id: '', period: new Date().toISOString().slice(0, 7), package_id: '', notes: '' }; showSlipDrawer = true; }} class="flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700"><FileText class="h-4 w-4" /> Buat Slip</button></div>
      <div class="overflow-hidden rounded-2xl border border-slate-200/70 bg-white shadow-sm">
        {#if slips.length === 0}
          <div class="flex flex-col items-center justify-center py-16 text-slate-400"><Receipt class="h-10 w-10 mb-2" /><p class="text-sm">Belum ada slip gaji</p></div>
        {:else}
          <table class="w-full text-sm">
            <thead><tr class="border-b border-slate-100"><th class="px-4 py-3 text-left text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Karyawan</th><th class="px-4 py-3 text-left text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Periode</th><th class="px-4 py-3 text-right text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Gaji Kotor</th><th class="px-4 py-3 text-right text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">PPh21</th><th class="px-4 py-3 text-right text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">BPJS</th><th class="px-4 py-3 text-right text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Bersih</th><th class="px-4 py-3 text-left text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Status</th><th class="px-4 py-3 text-right text-[11.5px] font-semibold uppercase tracking-wider text-slate-400"></th></tr></thead>
            <tbody>
              {#each slips as s}
                <tr class="transition-colors hover:bg-primary-50/30">
                  <td class="border-b border-slate-100 px-4 py-3.5"><div class="flex items-center gap-3"><Avatar name={s.employee_name} size={38} /><span class="font-bold text-[#10211c]">{s.employee_name}</span></div></td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-slate-600">{s.period}</td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-right text-slate-700" style="font-variant-numeric:tabular-nums">{formatIDR(s.base_salary + s.allowance)}</td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-right text-red-600" style="font-variant-numeric:tabular-nums">{formatIDR(s.pph21_amount)}</td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-right text-red-600" style="font-variant-numeric:tabular-nums">{formatIDR(s.bpjs_amount)}</td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-right font-bold text-primary-600" style="font-variant-numeric:tabular-nums">{formatIDR(s.net_salary)}</td>
                  <td class="border-b border-slate-100 px-4 py-3.5"><StatusBadge status={s.status} size="xs" /></td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-right">
                    {#if s.status === 'draft'}
                      <button type="button" onclick={() => finalizeSlip(s.id)} class="rounded-lg bg-emerald-100 px-2.5 py-1 text-xs font-semibold text-emerald-700 hover:bg-emerald-200">Final</button>
                    {:else if s.status === 'final'}
                      <button type="button" onclick={() => downloadSlipPDF(s)} class="rounded-lg bg-blue-100 px-2.5 py-1 text-xs font-semibold text-blue-700 hover:bg-blue-200">PDF</button>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </div>
    {/if}

    <!-- Advances Tab -->
    {#if tab === 'advances'}
      <div class="flex justify-end"><button type="button" onclick={() => { advanceForm = { employee_id: '', amount: 0, reason: '' }; showAdvanceDrawer = true; }} class="flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2 text-sm font-semibold text-white hover:bg-primary-700"><Wallet class="h-4 w-4" /> Catat Kasbon</button></div>
      <div class="overflow-hidden rounded-2xl border border-slate-200/70 bg-white shadow-sm">
        {#if advances.length === 0}
          <div class="flex flex-col items-center justify-center py-16 text-slate-400"><Wallet class="h-10 w-10 mb-2" /><p class="text-sm">Belum ada kasbon</p></div>
        {:else}
          <table class="w-full text-sm">
            <thead><tr class="border-b border-slate-100"><th class="px-4 py-3 text-left text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Karyawan</th><th class="px-4 py-3 text-right text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Jumlah</th><th class="px-4 py-3 text-right text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Sisa</th><th class="px-4 py-3 text-left text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Alasan</th><th class="px-4 py-3 text-left text-[11.5px] font-semibold uppercase tracking-wider text-slate-400">Status</th><th class="px-4 py-3 text-right text-[11.5px] font-semibold uppercase tracking-wider text-slate-400"></th></tr></thead>
            <tbody>
              {#each advances as a}
                <tr class="transition-colors hover:bg-primary-50/30">
                  <td class="border-b border-slate-100 px-4 py-3.5"><div class="flex items-center gap-3"><Avatar name={a.employee_name} size={38} /><span class="font-bold text-[#10211c]">{a.employee_name}</span></div></td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-right font-semibold text-slate-700" style="font-variant-numeric:tabular-nums">{formatIDR(a.amount)}</td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-right {a.remaining > 0 ? 'text-amber-600 font-semibold' : 'text-slate-400'}" style="font-variant-numeric:tabular-nums">{formatIDR(a.remaining)}</td>
                  <td class="border-b border-slate-100 px-4 py-3.5 max-w-[200px] truncate text-xs text-slate-500">{a.reason || '-'}</td>
                  <td class="border-b border-slate-100 px-4 py-3.5"><StatusBadge status={a.status} size="xs" /></td>
                  <td class="border-b border-slate-100 px-4 py-3.5 text-right">{#if a.remaining > 0}<button type="button" onclick={() => openRepay(a)} class="rounded-lg bg-primary-100 px-2.5 py-1 text-xs font-semibold text-primary-700 hover:bg-primary-200">Bayar</button>{/if}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </div>
    {/if}
  {/if}
</div>

<!-- Employee Drawer -->
<SlideDrawer open={showEmployeeDrawer} onClose={() => showEmployeeDrawer = false} title={editingEmployee ? 'Edit Karyawan' : 'Tambah Karyawan'} width="520px">
  <div class="flex flex-col gap-4 p-4">
    <div class="flex flex-col gap-1"><label for="emp-name" class="text-xs font-medium text-slate-700">Nama</label><input id="emp-name" type="text" bind:value={employeeForm.name} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex flex-col gap-1"><label for="emp-pos" class="text-xs font-medium text-slate-700">Jabatan</label><input id="emp-pos" type="text" bind:value={employeeForm.position} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex flex-col gap-1"><label for="emp-type" class="text-xs font-medium text-slate-700">Tipe</label><select id="emp-type" bind:value={employeeForm.type} class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none"><option value="tetap">Tetap</option><option value="freelance">Freelance</option></select></div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1"><label for="emp-base" class="text-xs font-medium text-slate-700">Gaji Pokok</label><input id="emp-base" type="number" bind:value={employeeForm.base_salary} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
      <div class="flex flex-col gap-1"><label for="emp-allow" class="text-xs font-medium text-slate-700">Tunjangan</label><input id="emp-allow" type="number" bind:value={employeeForm.allowance} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    </div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1"><label for="emp-bpjs-tk" class="text-xs font-medium text-slate-700">BPJS TK</label><input id="emp-bpjs-tk" type="number" bind:value={employeeForm.bpjs_tk} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
      <div class="flex flex-col gap-1"><label for="emp-bpjs-kes" class="text-xs font-medium text-slate-700">BPJS Kes</label><input id="emp-bpjs-kes" type="number" bind:value={employeeForm.bpjs_kes} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    </div>
    <div class="flex flex-col gap-1"><label for="emp-pph21" class="text-xs font-medium text-slate-700">PPh21 (%)</label><input id="emp-pph21" type="number" bind:value={employeeForm.pph21_rate} step="0.01" class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1"><label for="emp-phone" class="text-xs font-medium text-slate-700">Telepon</label><input id="emp-phone" type="text" bind:value={employeeForm.phone} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
      <div class="flex flex-col gap-1"><label for="emp-email" class="text-xs font-medium text-slate-700">Email</label><input id="emp-email" type="email" bind:value={employeeForm.email} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    </div>
    <div class="flex gap-2 pt-2">
      <button type="button" onclick={() => showEmployeeDrawer = false} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600">Batal</button>
      <button type="button" onclick={saveEmployee} disabled={savingEmployee} class="flex-1 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">{savingEmployee ? '...' : 'Simpan'}</button>
    </div>
  </div>
</SlideDrawer>

<!-- Slip Drawer -->
<SlideDrawer open={showSlipDrawer} onClose={() => showSlipDrawer = false} title="Buat Slip Gaji" width="480px">
  <div class="flex flex-col gap-4 p-4">
    <div class="flex flex-col gap-1"><label for="slip-emp" class="text-xs font-medium text-slate-700">Karyawan</label><select id="slip-emp" bind:value={slipForm.employee_id} class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none"><option value="">Pilih Karyawan</option>{#each employees as e}<option value={e.id}>{e.name} — {e.position}</option>{/each}</select></div>
    <div class="flex flex-col gap-1"><label for="slip-period" class="text-xs font-medium text-slate-700">Periode</label><input id="slip-period" type="month" bind:value={slipForm.period} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex flex-col gap-1"><label for="slip-notes" class="text-xs font-medium text-slate-700">Catatan</label><input id="slip-notes" type="text" bind:value={slipForm.notes} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex gap-2 pt-2">
      <button type="button" onclick={() => showSlipDrawer = false} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600">Batal</button>
      <button type="button" onclick={generateSlip} disabled={savingSlip || !slipForm.employee_id} class="flex-1 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">{savingSlip ? '...' : 'Generate'}</button>
    </div>
  </div>
</SlideDrawer>

<!-- Advance Drawer -->
<SlideDrawer open={showAdvanceDrawer} onClose={() => showAdvanceDrawer = false} title="Catat Kasbon" width="480px">
  <div class="flex flex-col gap-4 p-4">
    <div class="flex flex-col gap-1"><label for="adv-emp" class="text-xs font-medium text-slate-700">Karyawan</label><select id="adv-emp" bind:value={advanceForm.employee_id} class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none"><option value="">Pilih Karyawan</option>{#each employees as e}<option value={e.id}>{e.name}</option>{/each}</select></div>
    <div class="flex flex-col gap-1"><label for="adv-amt" class="text-xs font-medium text-slate-700">Jumlah</label><input id="adv-amt" type="number" bind:value={advanceForm.amount} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex flex-col gap-1"><label for="adv-reason" class="text-xs font-medium text-slate-700">Alasan</label><input id="adv-reason" type="text" bind:value={advanceForm.reason} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex gap-2 pt-2">
      <button type="button" onclick={() => showAdvanceDrawer = false} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600">Batal</button>
      <button type="button" onclick={createAdvance} disabled={savingAdvance || !advanceForm.employee_id || advanceForm.amount <= 0} class="flex-1 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">{savingAdvance ? '...' : 'Catat'}</button>
    </div>
  </div>
</SlideDrawer>

<!-- Repay Drawer -->
<SlideDrawer open={showRepayDrawer} onClose={() => showRepayDrawer = false} title="Bayar Kasbon" width="480px">
  <div class="flex flex-col gap-4 p-4">
    {#if repayingAdvance}
      <div class="rounded-xl bg-slate-50 p-3"><p class="text-xs text-slate-500">Karyawan: <span class="font-semibold text-slate-800">{repayingAdvance.employee_name}</span></p><p class="text-xs text-slate-500">Sisa: <span class="font-semibold text-amber-600">{formatIDR(repayingAdvance.remaining)}</span></p></div>
    {/if}
    <div class="flex flex-col gap-1"><label for="repay-amt" class="text-xs font-medium text-slate-700">Jumlah Bayar</label><input id="repay-amt" type="number" bind:value={repayForm.amount} max={repayingAdvance?.remaining || 0} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex gap-2 pt-2">
      <button type="button" onclick={() => showRepayDrawer = false} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600">Batal</button>
      <button type="button" onclick={submitRepay} disabled={savingRepay || repayForm.amount <= 0} class="flex-1 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">{savingRepay ? '...' : 'Bayar'}</button>
    </div>
  </div>
</SlideDrawer>
