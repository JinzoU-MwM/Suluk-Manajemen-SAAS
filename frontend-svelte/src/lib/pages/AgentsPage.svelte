<script>
  import { onMount } from 'svelte';
  import { Users, DollarSign, CheckCheck, Plus, Pencil, Search } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import { showToast } from '../services/toast.svelte.js';
  import { ApiService } from '../services/api.js';

  let { onNavigate, user = null } = $props();

  let agents = $state([]);
  let commissions = $state([]);
  let loading = $state(true);
  let tab = $state('agents');
  let searchQuery = $state('');
  let statusFilter = $state('all');

  let showAgentDrawer = $state(false);
  let editingAgent = $state(null);
  let agentForm = $state({ name: '', phone: '', email: '', address: '', commission_rate: 5, bank_name: '', bank_account_number: '', bank_account_name: '', notes: '' });
  let savingAgent = $state(false);

  let showCommDrawer = $state(false);
  let commForm = $state({ agent_id: '', jamaah_name: '', package_name: '', commission_amount: 0, commission_rate: 5, notes: '' });
  let savingComm = $state(false);

  let selectedAgent = $state(null);
  let showAgentCommissions = $state(false);
  let agentCommissions = $state([]);

  function formatIDR(n) { return n ? `Rp ${Number(n).toLocaleString('id-ID')}` : 'Rp 0'; }

  async function loadData() {
    loading = true;
    try {
      const [agentData, commData] = await Promise.all([
        ApiService.listAgents({ search: searchQuery }),
        ApiService.listCommissions({ status: statusFilter }),
      ]);
      agents = agentData.agents || [];
      commissions = commData.commissions || [];
    } catch (e) { showToast(e.message, 'error'); } finally { loading = false; }
  }
  onMount(() => { loadData(); });

  function openNewAgent() { editingAgent = null; agentForm = { name: '', phone: '', email: '', address: '', commission_rate: 5, bank_name: '', bank_account_number: '', bank_account_name: '', notes: '' }; showAgentDrawer = true; }
  function editAgent(a) { editingAgent = a; agentForm = { name: a.name, phone: a.phone, email: a.email, address: a.address, commission_rate: a.commission_rate, bank_name: a.bank_name, bank_account_number: a.bank_account_number, bank_account_name: a.bank_account_name, notes: a.notes }; showAgentDrawer = true; }

  async function saveAgent() {
    savingAgent = true;
    try {
      if (editingAgent) { await ApiService.updateAgent(editingAgent.id, agentForm); showToast('Agen diperbarui'); }
      else { await ApiService.createAgent(agentForm); showToast('Agen ditambahkan'); }
      showAgentDrawer = false; await loadData();
    } catch (e) { showToast(e.message, 'error'); } finally { savingAgent = false; }
  }

  async function payCommission(id) {
    try { await ApiService.payCommission(id); showToast('Komisi dibayar'); await loadData(); }
    catch (e) { showToast(e.message, 'error'); }
  }

  function openNewCommission() { commForm = { agent_id: '', jamaah_name: '', package_name: '', commission_amount: 0, commission_rate: 5, notes: '' }; showCommDrawer = true; }

  async function saveCommission() {
    savingComm = true;
    try {
      await ApiService.createCommission(commForm); showToast('Komisi dicatat');
      showCommDrawer = false; await loadData();
    } catch (e) { showToast(e.message, 'error'); } finally { savingComm = false; }
  }

  async function viewAgentCommissions(a) {
    selectedAgent = a;
    try { const data = await ApiService.getAgentCommissions(a.id); agentCommissions = data.commissions || []; }
    catch (e) { agentCommissions = []; }
    showAgentCommissions = true;
  }
</script>

<div class="flex flex-col gap-6 p-4 lg:p-8">
  <header class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
    <div><h1 class="font-serif text-xl font-bold text-slate-900">Komisi Agen</h1><p class="text-sm text-slate-500">Kelola jaringan agen dan komisi referral.</p></div>
  </header>

  {#if loading}
    <div class="flex items-center justify-center py-16"><div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-600 border-t-transparent"></div></div>
  {:else}
    <!-- Tabs -->
    <div class="flex items-center justify-between">
      <div class="flex gap-1 rounded-xl bg-slate-100 p-1 w-fit">
        <button type="button" onclick={() => tab = 'agents'} class="rounded-lg px-4 py-2 text-xs font-semibold {tab === 'agents' ? 'bg-white text-slate-800 shadow-sm' : 'text-slate-500'}">Agen ({agents.length})</button>
        <button type="button" onclick={() => tab = 'commissions'} class="rounded-lg px-4 py-2 text-xs font-semibold {tab === 'commissions' ? 'bg-white text-slate-800 shadow-sm' : 'text-slate-500'}">Komisi ({commissions.length})</button>
      </div>
      {#if tab === 'agents'}
        <div class="flex items-center gap-2">
          <div class="relative"><Search class="absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-400" /><input type="text" bind:value={searchQuery} oninput={loadData} placeholder="Cari..." class="rounded-xl border border-slate-200 py-2 pl-9 pr-3 text-xs outline-none focus:border-primary-400 w-40" /></div>
          <button type="button" onclick={openNewAgent} class="flex items-center gap-2 rounded-xl bg-primary-600 px-3 py-2 text-xs font-semibold text-white hover:bg-primary-700"><Plus class="h-3.5 w-3.5" /> Agen</button>
        </div>
      {:else}
        <div class="flex items-center gap-2">
          <select bind:value={statusFilter} onchange={loadData} class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-xs font-medium outline-none">
            <option value="all">Semua</option><option value="pending">Pending</option><option value="paid">Dibayar</option>
          </select>
          <button type="button" onclick={openNewCommission} class="flex items-center gap-2 rounded-xl bg-primary-600 px-3 py-2 text-xs font-semibold text-white hover:bg-primary-700"><Plus class="h-3.5 w-3.5" /> Komisi</button>
        </div>
      {/if}
    </div>

    <!-- Agents Tab -->
    {#if tab === 'agents'}
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        {#each agents as a}
          <div class="rounded-2xl border border-slate-100 bg-white p-5 shadow-sm hover:shadow-md transition-shadow">
            <div class="flex items-start justify-between mb-3">
              <div>
                <h3 class="font-semibold text-slate-800">{a.name}</h3>
                <p class="text-xs text-slate-500">{a.phone || '-'}</p>
              </div>
              <span class="rounded-full bg-primary-50 px-2.5 py-1 text-xs font-bold text-primary-700">{a.commission_rate}%</span>
            </div>
            <div class="grid grid-cols-2 gap-2 mb-3 text-xs">
              <div><span class="text-slate-400">Total Komisi</span><p class="font-bold text-slate-800">{formatIDR(a.total_commissions)}</p></div>
              <div><span class="text-slate-400">Dibayar</span><p class="font-bold text-emerald-600">{formatIDR(a.total_paid)}</p></div>
              <div><span class="text-slate-400">Outstanding</span><p class="font-bold text-amber-600">{formatIDR(a.total_outstanding)}</p></div>
              <div><span class="text-slate-400">Jamaah</span><p class="font-bold text-slate-800">{a.total_jamaah}</p></div>
            </div>
            <div class="flex gap-2">
              <button type="button" onclick={() => editAgent(a)} class="flex-1 rounded-lg border border-slate-200 py-1.5 text-xs font-semibold text-slate-600 hover:bg-slate-50"><Pencil class="h-3 w-3 inline mr-1" />Edit</button>
              <button type="button" onclick={() => viewAgentCommissions(a)} class="flex-1 rounded-lg bg-primary-50 py-1.5 text-xs font-semibold text-primary-700 hover:bg-primary-100">Komisi</button>
            </div>
          </div>
        {/each}
      </div>
      {#if agents.length === 0}<div class="flex flex-col items-center justify-center py-16 text-slate-400"><Users class="h-10 w-10 mb-2" /><p class="text-sm">Belum ada agen</p></div>{/if}
    {/if}

    <!-- Commissions Tab -->
    {#if tab === 'commissions'}
      <div class="overflow-hidden rounded-2xl border border-slate-100 bg-white shadow-sm">
        {#if commissions.length === 0}
          <div class="flex flex-col items-center justify-center py-16 text-slate-400"><DollarSign class="h-10 w-10 mb-2" /><p class="text-sm">Belum ada komisi</p></div>
        {:else}
          <table class="w-full text-sm">
            <thead class="border-b border-slate-100 bg-slate-50"><tr><th class="px-4 py-3 text-left text-xs font-semibold text-slate-500">Agen</th><th class="px-4 py-3 text-left text-xs font-semibold text-slate-500">Jamaah</th><th class="px-4 py-3 text-left text-xs font-semibold text-slate-500">Paket</th><th class="px-4 py-3 text-right text-xs font-semibold text-slate-500">Jumlah</th><th class="px-4 py-3 text-left text-xs font-semibold text-slate-500">Status</th><th class="px-4 py-3 text-right text-xs font-semibold text-slate-500"></th></tr></thead>
            <tbody class="divide-y divide-slate-50">
              {#each commissions as c}
                <tr class="hover:bg-slate-50"><td class="px-4 py-3 font-medium text-slate-800">{c.agent_name}</td><td class="px-4 py-3 text-slate-600">{c.jamaah_name || '-'}</td><td class="px-4 py-3 text-xs text-slate-500">{c.package_name || '-'}</td><td class="px-4 py-3 text-right font-bold text-slate-700">{formatIDR(c.commission_amount)}</td><td class="px-4 py-3"><StatusBadge status={c.status} size="xs" /></td><td class="px-4 py-3 text-right">{#if c.status === 'pending'}<button type="button" onclick={() => payCommission(c.id)} class="rounded-lg bg-emerald-100 px-2.5 py-1 text-xs font-semibold text-emerald-700 hover:bg-emerald-200"><CheckCheck class="h-3 w-3 inline mr-1" />Bayar</button>{/if}</td></tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </div>
    {/if}
  {/if}
</div>

<!-- Agent Drawer -->
<SlideDrawer open={showAgentDrawer} onClose={() => showAgentDrawer = false} title={editingAgent ? 'Edit Agen' : 'Tambah Agen'} width="520px">
  <div class="flex flex-col gap-4 p-4">
    <div class="flex flex-col gap-1"><label for="a-name" class="text-xs font-medium text-slate-700">Nama <span class="text-red-500">*</span></label><input id="a-name" type="text" bind:value={agentForm.name} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1"><label for="a-phone" class="text-xs font-medium text-slate-700">Telepon</label><input id="a-phone" type="text" bind:value={agentForm.phone} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
      <div class="flex flex-col gap-1"><label for="a-rate" class="text-xs font-medium text-slate-700">Komisi (%)</label><input id="a-rate" type="number" bind:value={agentForm.commission_rate} step="0.01" class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    </div>
    <div class="flex flex-col gap-1"><label for="a-email" class="text-xs font-medium text-slate-700">Email</label><input id="a-email" type="email" bind:value={agentForm.email} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex flex-col gap-1"><label for="a-address" class="text-xs font-medium text-slate-700">Alamat</label><textarea id="a-address" bind:value={agentForm.address} rows="2" class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400"></textarea></div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1"><label for="a-bank-name" class="text-xs font-medium text-slate-700">Nama Bank</label><input id="a-bank-name" type="text" bind:value={agentForm.bank_name} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
      <div class="flex flex-col gap-1"><label for="a-bank-ac" class="text-xs font-medium text-slate-700">No. Rekening</label><input id="a-bank-ac" type="text" bind:value={agentForm.bank_account_number} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    </div>
    <div class="flex flex-col gap-1"><label for="a-bank-an" class="text-xs font-medium text-slate-700">Atas Nama</label><input id="a-bank-an" type="text" bind:value={agentForm.bank_account_name} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex flex-col gap-1"><label for="a-notes" class="text-xs font-medium text-slate-700">Catatan</label><input id="a-notes" type="text" bind:value={agentForm.notes} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex gap-2 pt-2">
      <button type="button" onclick={() => showAgentDrawer = false} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600">Batal</button>
      <button type="button" onclick={saveAgent} disabled={savingAgent || !agentForm.name} class="flex-1 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">{savingAgent ? '...' : 'Simpan'}</button>
    </div>
  </div>
</SlideDrawer>

<!-- Commission Drawer -->
<SlideDrawer open={showCommDrawer} onClose={() => showCommDrawer = false} title="Catat Komisi" width="480px">
  <div class="flex flex-col gap-4 p-4">
    <div class="flex flex-col gap-1"><label for="c-agent" class="text-xs font-medium text-slate-700">Agen</label><select id="c-agent" bind:value={commForm.agent_id} class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none"><option value="">Pilih Agen</option>{#each agents as a}<option value={a.id}>{a.name}</option>{/each}</select></div>
    <div class="flex flex-col gap-1"><label for="c-jamaah" class="text-xs font-medium text-slate-700">Nama Jamaah</label><input id="c-jamaah" type="text" bind:value={commForm.jamaah_name} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex flex-col gap-1"><label for="c-pkg" class="text-xs font-medium text-slate-700">Nama Paket</label><input id="c-pkg" type="text" bind:value={commForm.package_name} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1"><label for="c-amt" class="text-xs font-medium text-slate-700">Jumlah</label><input id="c-amt" type="number" bind:value={commForm.commission_amount} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
      <div class="flex flex-col gap-1"><label for="c-rate" class="text-xs font-medium text-slate-700">Rate (%)</label><input id="c-rate" type="number" bind:value={commForm.commission_rate} step="0.01" class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    </div>
    <div class="flex flex-col gap-1"><label for="c-notes" class="text-xs font-medium text-slate-700">Catatan</label><input id="c-notes" type="text" bind:value={commForm.notes} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex gap-2 pt-2">
      <button type="button" onclick={() => showCommDrawer = false} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600">Batal</button>
      <button type="button" onclick={saveCommission} disabled={savingComm || !commForm.agent_id || commForm.commission_amount <= 0} class="flex-1 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">{savingComm ? '...' : 'Catat'}</button>
    </div>
  </div>
</SlideDrawer>

<!-- Agent Commissions Detail -->
<SlideDrawer open={showAgentCommissions} onClose={() => showAgentCommissions = false} title={selectedAgent ? `Komisi: ${selectedAgent.name}` : 'Komisi'} width="640px">
  <div class="p-4">
    {#if selectedAgent}
      <div class="grid grid-cols-3 gap-3 mb-4">
        <div class="rounded-xl bg-slate-50 p-3 text-center"><p class="text-lg font-bold text-slate-800">{formatIDR(selectedAgent.total_commissions)}</p><p class="text-xs text-slate-500">Total</p></div>
        <div class="rounded-xl bg-emerald-50 p-3 text-center"><p class="text-lg font-bold text-emerald-600">{formatIDR(selectedAgent.total_paid)}</p><p class="text-xs text-slate-500">Dibayar</p></div>
        <div class="rounded-xl bg-amber-50 p-3 text-center"><p class="text-lg font-bold text-amber-600">{formatIDR(selectedAgent.total_outstanding)}</p><p class="text-xs text-slate-500">Outstanding</p></div>
      </div>
    {/if}
    {#if agentCommissions.length === 0}
      <p class="text-sm text-slate-400 text-center py-8">Belum ada komisi</p>
    {:else}
      <div class="space-y-2">
        {#each agentCommissions as c}
          <div class="rounded-xl border border-slate-100 p-3 flex items-center justify-between">
            <div><p class="text-sm font-semibold text-slate-800">{c.jamaah_name || '-'}</p><p class="text-xs text-slate-500">{c.package_name || '-'}</p></div>
            <div class="text-right"><p class="text-sm font-bold text-slate-800">{formatIDR(c.commission_amount)}</p><StatusBadge status={c.status} size="xs" /></div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</SlideDrawer>
