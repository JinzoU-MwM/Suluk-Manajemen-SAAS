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
  import EmptyState from '../components/EmptyState.svelte';
  import Card from '../components/ui/Card.svelte';
  import Button from '../components/ui/Button.svelte';
  import FilterTabs from '../components/ui/FilterTabs.svelte';
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

  const payrollTabs = [
    { value: 'employees', label: 'Karyawan' },
    { value: 'slips', label: 'Slip Gaji' },
    { value: 'advances', label: 'Kasbon' },
  ];
</script>

<div class="payroll-page">
  <PageHeader kicker="Penggajian" title="Penggajian" subtitle="Kelola karyawan, gaji, dan kasbon." />

  {#if loading}
    <div class="payroll-loading"><div class="payroll-spinner"></div></div>
  {:else}
    <!-- Summary -->
    <div class="payroll-stats">
      <StatCard icon={Users} label="Karyawan Aktif" value={`${summary.active_employees}/${summary.total_employees}`} accent="var(--c-primary)" />
      <StatCard icon={Banknote} label="Gaji Bulan Ini" value={formatIDR(summary.monthly_payroll)} accent="var(--c-info)" />
      <StatCard icon={Clock} label="Kasbon Outstanding" value={formatIDR(summary.outstanding_advances)} accent="var(--c-accent)" />
      <StatCard icon={Wallet} label="Total Kasbon" value={formatIDR(summary.total_advances)} accent="var(--c-primary)" />
    </div>

    <!-- Tabs + actions -->
    <div class="payroll-toolbar">
      <FilterTabs tabs={payrollTabs} value={tab} onChange={(v) => tab = v} />

      {#if tab === 'employees'}
        <Button variant="primary" icon={Plus} onclick={openNewEmployee}>Tambah Karyawan</Button>
      {:else if tab === 'slips'}
        <Button variant="primary" icon={FileText} onclick={() => { slipForm = { employee_id: '', period: new Date().toISOString().slice(0, 7), package_id: '', notes: '' }; showSlipDrawer = true; }}>Buat Slip</Button>
      {:else}
        <Button variant="primary" icon={Wallet} onclick={() => { advanceForm = { employee_id: '', amount: 0, reason: '' }; showAdvanceDrawer = true; }}>Catat Kasbon</Button>
      {/if}
    </div>

    <!-- Employees Tab -->
    {#if tab === 'employees'}
      {#if employees.length === 0}
        <Card><EmptyState icon={Users} title="Belum ada karyawan" text="Tambahkan karyawan untuk mulai mengelola penggajian." /></Card>
      {:else}
        <Card pad={false} style="padding:8px 4px">
          <div class="payroll-table-wrap">
            <table class="payroll-table">
              <thead>
                <tr>
                  <th>Nama</th>
                  <th>Jabatan</th>
                  <th>Tipe</th>
                  <th class="ta-right">Gaji Pokok</th>
                  <th class="ta-right">Tunjangan</th>
                  <th class="ta-right"></th>
                </tr>
              </thead>
              <tbody>
                {#each employees as e}
                  <tr class="payroll-row">
                    <td>
                      <div class="cell-id">
                        <Avatar name={e.name} size={38} />
                        <span class="cell-name">{e.name}</span>
                      </div>
                    </td>
                    <td>{e.position}</td>
                    <td>
                      <span class="type-pill" style={e.type === 'tetap' ? 'background:var(--c-primary-soft);color:var(--c-primary-deep)' : 'background:var(--c-info-soft);color:var(--c-info)'}>{e.type === 'tetap' ? 'Tetap' : 'Freelance'}</span>
                    </td>
                    <td class="ta-right tabular cell-strong">{formatIDR(e.base_salary)}</td>
                    <td class="ta-right tabular">{formatIDR(e.allowance)}</td>
                    <td class="ta-right">
                      <button type="button" onclick={() => editEmployee(e)} class="icon-btn" aria-label="Edit karyawan"><Pencil size={16} /></button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </Card>
      {/if}
    {/if}

    <!-- Slips Tab -->
    {#if tab === 'slips'}
      {#if slips.length === 0}
        <Card><EmptyState icon={Receipt} title="Belum ada slip gaji" text="Buat slip gaji untuk karyawan pada periode terpilih." /></Card>
      {:else}
        <Card pad={false} style="padding:8px 4px">
          <div class="payroll-table-wrap">
            <table class="payroll-table">
              <thead>
                <tr>
                  <th>Karyawan</th>
                  <th>Periode</th>
                  <th class="ta-right">Gaji Kotor</th>
                  <th class="ta-right">PPh21</th>
                  <th class="ta-right">BPJS</th>
                  <th class="ta-right">Bersih</th>
                  <th>Status</th>
                  <th class="ta-right"></th>
                </tr>
              </thead>
              <tbody>
                {#each slips as s}
                  <tr class="payroll-row">
                    <td>
                      <div class="cell-id">
                        <Avatar name={s.employee_name} size={38} />
                        <span class="cell-name">{s.employee_name}</span>
                      </div>
                    </td>
                    <td>{s.period}</td>
                    <td class="ta-right tabular">{formatIDR(s.base_salary + s.allowance)}</td>
                    <td class="ta-right tabular cell-danger">{formatIDR(s.pph21_amount)}</td>
                    <td class="ta-right tabular cell-danger">{formatIDR(s.bpjs_amount)}</td>
                    <td class="ta-right tabular cell-net">{formatIDR(s.net_salary)}</td>
                    <td><StatusBadge status={s.status} size="xs" /></td>
                    <td class="ta-right">
                      {#if s.status === 'draft'}
                        <Button size="sm" variant="soft" onclick={() => finalizeSlip(s.id)}>Final</Button>
                      {:else if s.status === 'final'}
                        <Button size="sm" variant="ghost" onclick={() => downloadSlipPDF(s)}>PDF</Button>
                      {/if}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </Card>
      {/if}
    {/if}

    <!-- Advances Tab -->
    {#if tab === 'advances'}
      {#if advances.length === 0}
        <Card><EmptyState icon={Wallet} title="Belum ada kasbon" text="Catat kasbon karyawan untuk melacak sisa pinjaman." /></Card>
      {:else}
        <Card pad={false} style="padding:8px 4px">
          <div class="payroll-table-wrap">
            <table class="payroll-table">
              <thead>
                <tr>
                  <th>Karyawan</th>
                  <th class="ta-right">Jumlah</th>
                  <th class="ta-right">Sisa</th>
                  <th>Alasan</th>
                  <th>Status</th>
                  <th class="ta-right"></th>
                </tr>
              </thead>
              <tbody>
                {#each advances as a}
                  <tr class="payroll-row">
                    <td>
                      <div class="cell-id">
                        <Avatar name={a.employee_name} size={38} />
                        <span class="cell-name">{a.employee_name}</span>
                      </div>
                    </td>
                    <td class="ta-right tabular cell-strong">{formatIDR(a.amount)}</td>
                    <td class="ta-right tabular" style={a.remaining > 0 ? 'color:var(--c-warning);font-weight:700' : 'color:var(--c-faint)'}>{formatIDR(a.remaining)}</td>
                    <td class="cell-reason">{a.reason || '-'}</td>
                    <td><StatusBadge status={a.status} size="xs" /></td>
                    <td class="ta-right">
                      {#if a.remaining > 0}
                        <Button size="sm" variant="soft" onclick={() => openRepay(a)}>Bayar</Button>
                      {/if}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </Card>
      {/if}
    {/if}
  {/if}
</div>

<!-- Employee Drawer -->
<SlideDrawer open={showEmployeeDrawer} onClose={() => showEmployeeDrawer = false} title={editingEmployee ? 'Edit Karyawan' : 'Tambah Karyawan'} width="520px">
  <div class="drawer-form">
    <div class="field"><label for="emp-name">Nama</label><input id="emp-name" type="text" bind:value={employeeForm.name} /></div>
    <div class="field"><label for="emp-pos">Jabatan</label><input id="emp-pos" type="text" bind:value={employeeForm.position} /></div>
    <div class="field"><label for="emp-type">Tipe</label><select id="emp-type" bind:value={employeeForm.type}><option value="tetap">Tetap</option><option value="freelance">Freelance</option></select></div>
    <div class="field-grid">
      <div class="field"><label for="emp-base">Gaji Pokok</label><input id="emp-base" type="number" bind:value={employeeForm.base_salary} /></div>
      <div class="field"><label for="emp-allow">Tunjangan</label><input id="emp-allow" type="number" bind:value={employeeForm.allowance} /></div>
    </div>
    <div class="field-grid">
      <div class="field"><label for="emp-bpjs-tk">BPJS TK</label><input id="emp-bpjs-tk" type="number" bind:value={employeeForm.bpjs_tk} /></div>
      <div class="field"><label for="emp-bpjs-kes">BPJS Kes</label><input id="emp-bpjs-kes" type="number" bind:value={employeeForm.bpjs_kes} /></div>
    </div>
    <div class="field"><label for="emp-pph21">PPh21 (%)</label><input id="emp-pph21" type="number" bind:value={employeeForm.pph21_rate} step="0.01" /></div>
    <div class="field-grid">
      <div class="field"><label for="emp-phone">Telepon</label><input id="emp-phone" type="text" bind:value={employeeForm.phone} /></div>
      <div class="field"><label for="emp-email">Email</label><input id="emp-email" type="email" bind:value={employeeForm.email} /></div>
    </div>
    <div class="drawer-actions">
      <Button variant="ghost" full onclick={() => showEmployeeDrawer = false}>Batal</Button>
      <Button variant="primary" full disabled={savingEmployee} onclick={saveEmployee}>{savingEmployee ? '...' : 'Simpan'}</Button>
    </div>
  </div>
</SlideDrawer>

<!-- Slip Drawer -->
<SlideDrawer open={showSlipDrawer} onClose={() => showSlipDrawer = false} title="Buat Slip Gaji" width="480px">
  <div class="drawer-form">
    <div class="field"><label for="slip-emp">Karyawan</label><select id="slip-emp" bind:value={slipForm.employee_id}><option value="">Pilih Karyawan</option>{#each employees as e}<option value={e.id}>{e.name} — {e.position}</option>{/each}</select></div>
    <div class="field"><label for="slip-period">Periode</label><input id="slip-period" type="month" bind:value={slipForm.period} /></div>
    <div class="field"><label for="slip-notes">Catatan</label><input id="slip-notes" type="text" bind:value={slipForm.notes} /></div>
    <div class="drawer-actions">
      <Button variant="ghost" full onclick={() => showSlipDrawer = false}>Batal</Button>
      <Button variant="primary" full disabled={savingSlip || !slipForm.employee_id} onclick={generateSlip}>{savingSlip ? '...' : 'Generate'}</Button>
    </div>
  </div>
</SlideDrawer>

<!-- Advance Drawer -->
<SlideDrawer open={showAdvanceDrawer} onClose={() => showAdvanceDrawer = false} title="Catat Kasbon" width="480px">
  <div class="drawer-form">
    <div class="field"><label for="adv-emp">Karyawan</label><select id="adv-emp" bind:value={advanceForm.employee_id}><option value="">Pilih Karyawan</option>{#each employees as e}<option value={e.id}>{e.name}</option>{/each}</select></div>
    <div class="field"><label for="adv-amt">Jumlah</label><input id="adv-amt" type="number" bind:value={advanceForm.amount} /></div>
    <div class="field"><label for="adv-reason">Alasan</label><input id="adv-reason" type="text" bind:value={advanceForm.reason} /></div>
    <div class="drawer-actions">
      <Button variant="ghost" full onclick={() => showAdvanceDrawer = false}>Batal</Button>
      <Button variant="primary" full disabled={savingAdvance || !advanceForm.employee_id || advanceForm.amount <= 0} onclick={createAdvance}>{savingAdvance ? '...' : 'Catat'}</Button>
    </div>
  </div>
</SlideDrawer>

<!-- Repay Drawer -->
<SlideDrawer open={showRepayDrawer} onClose={() => showRepayDrawer = false} title="Bayar Kasbon" width="480px">
  <div class="drawer-form">
    {#if repayingAdvance}
      <div class="repay-summary">
        <p>Karyawan: <span class="rs-name">{repayingAdvance.employee_name}</span></p>
        <p>Sisa: <span class="rs-amount">{formatIDR(repayingAdvance.remaining)}</span></p>
      </div>
    {/if}
    <div class="field"><label for="repay-amt">Jumlah Bayar</label><input id="repay-amt" type="number" bind:value={repayForm.amount} max={repayingAdvance?.remaining || 0} /></div>
    <div class="drawer-actions">
      <Button variant="ghost" full onclick={() => showRepayDrawer = false}>Batal</Button>
      <Button variant="primary" full disabled={savingRepay || repayForm.amount <= 0} onclick={submitRepay}>{savingRepay ? '...' : 'Bayar'}</Button>
    </div>
  </div>
</SlideDrawer>

<style>
  .payroll-page {
    display: flex;
    flex-direction: column;
    gap: var(--gap, 24px);
    padding: 16px;
  }
  @media (min-width: 1024px) {
    .payroll-page { padding: 32px; }
  }

  .payroll-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 64px 0;
  }
  .payroll-spinner {
    height: 32px;
    width: 32px;
    border-radius: 999px;
    border: 2px solid var(--c-primary);
    border-top-color: transparent;
    animation: payroll-spin 0.7s linear infinite;
  }
  @keyframes payroll-spin { to { transform: rotate(360deg); } }

  .payroll-stats {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 16px;
  }
  @media (min-width: 1024px) {
    .payroll-stats { grid-template-columns: repeat(4, minmax(0, 1fr)); }
  }

  .payroll-toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
  }

  .payroll-table-wrap { overflow-x: auto; }
  .payroll-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13.5px;
  }
  .payroll-table thead th {
    text-align: left;
    padding: 0 16px 12px;
    font-size: 11.5px;
    font-weight: 700;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: var(--c-faint);
    white-space: nowrap;
    border-bottom: 1px solid var(--c-line);
  }
  .payroll-table tbody td {
    padding: 14px 16px;
    text-align: left;
    border-bottom: 1px solid var(--c-line-soft);
    color: var(--c-ink-soft);
    white-space: nowrap;
    vertical-align: middle;
  }
  .payroll-row { transition: background 0.12s; }
  .payroll-row:hover { background: var(--c-primary-tint); }

  .ta-right { text-align: right !important; }
  .tabular { font-variant-numeric: tabular-nums; }

  .cell-id { display: flex; align-items: center; gap: 12px; }
  .cell-name { font-weight: 700; color: var(--c-ink); }
  .cell-strong { font-weight: 600; color: var(--c-ink-soft); }
  .cell-danger { color: var(--c-danger); }
  .cell-net { font-weight: 800; color: var(--c-primary); }
  .cell-reason {
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 12px;
    color: var(--c-muted);
  }

  .type-pill {
    display: inline-block;
    border-radius: 999px;
    padding: 3px 10px;
    font-size: 12px;
    font-weight: 600;
    white-space: nowrap;
  }

  .icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    padding: 6px;
    color: var(--c-faint);
    transition: background 0.12s, color 0.12s;
  }
  .icon-btn:hover { background: var(--c-bg-2); color: var(--c-ink-soft); }

  /* Drawer forms */
  .drawer-form {
    display: flex;
    flex-direction: column;
    gap: 16px;
    padding: 16px;
  }
  .field { display: flex; flex-direction: column; gap: 4px; }
  .field-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
  }
  .field label {
    font-size: 12px;
    font-weight: 500;
    color: var(--c-ink-soft);
  }
  .field input,
  .field select {
    border-radius: var(--radius, 12px);
    border: 1px solid var(--c-line);
    background: var(--c-surface);
    padding: 9px 12px;
    font-size: 13px;
    color: var(--c-ink);
    outline: none;
    transition: border-color 0.12s;
  }
  .field input:focus,
  .field select:focus { border-color: var(--c-primary); }

  .drawer-actions { display: flex; gap: 8px; padding-top: 8px; }

  .repay-summary {
    border-radius: var(--radius, 12px);
    background: var(--c-bg-2);
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .repay-summary p { font-size: 12px; color: var(--c-muted); }
  .rs-name { font-weight: 600; color: var(--c-ink); }
  .rs-amount { font-weight: 600; color: var(--c-warning); }
</style>
