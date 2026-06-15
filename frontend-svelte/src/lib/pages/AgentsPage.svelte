<script>
  import { onMount } from 'svelte';
  import { Users, DollarSign, CheckCheck, Plus, Pencil, Search, UserCheck, TrendingUp, Wallet, Network, X, KeyRound } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import PageHeader from '../components/PageHeader.svelte';
  import StatCard from '../components/StatCard.svelte';
  import Avatar from '../components/Avatar.svelte';
  import Card from '../components/ui/Card.svelte';
  import Button from '../components/ui/Button.svelte';
  import FilterTabs from '../components/ui/FilterTabs.svelte';
  import { showToast } from '../services/toast.svelte.js';
  import { ApiService } from '../services/api.js';

  let { onNavigate, user = null } = $props();

  let agents = $state([]);
  let commissions = $state([]);
  let loading = $state(true);
  let tab = $state('agents');
  let searchQuery = $state('');
  let statusFilter = $state('all');

  let tierFilter = $state('');

  let showAgentDrawer = $state(false);
  let editingAgent = $state(null);
  const emptyAgentForm = { name: '', phone: '', email: '', address: '', commission_rate: 5, bank_name: '', bank_account_number: '', bank_account_name: '', notes: '', parent_id: '', type: 'agent' };
  let agentForm = $state({ ...emptyAgentForm });
  let savingAgent = $state(false);

  let showCommDrawer = $state(false);
  let commForm = $state({ agent_id: '', jamaah_name: '', package_name: '', commission_amount: 0, commission_rate: 5, notes: '' });
  let savingComm = $state(false);

  let selectedAgent = $state(null);
  let showAgentCommissions = $state(false);
  let agentCommissions = $state([]);
  let agentDownline = $state([]);
  let agentUpline = $state([]);

  // Tier configuration
  let showTierDrawer = $state(false);
  let tiers = $state([]);
  let savingTiers = $state(false);

  // Portal credential provisioning
  let showPortalDrawer = $state(false);
  let portalAgent = $state(null);
  let portalForm = $state({ email: '', name: '', password: '' });
  let savingPortal = $state(false);

  function openPortal(a) { portalAgent = a; portalForm = { email: a.email || '', name: a.name, password: '' }; showPortalDrawer = true; }
  async function savePortal() {
    if (!portalForm.email || !portalForm.password) { showToast('Email dan password wajib diisi', 'warning'); return; }
    savingPortal = true;
    try {
      await ApiService.provisionAgentPortal({ agent_id: portalAgent.id, email: portalForm.email, name: portalForm.name, password: portalForm.password });
      showToast('Akun portal agen dibuat');
      showPortalDrawer = false;
    } catch (e) { showToast(e.message, 'error'); } finally { savingPortal = false; }
  }

  function formatIDR(n) { return n ? `Rp ${Number(n).toLocaleString('id-ID')}` : 'Rp 0'; }

  const AGENT_TYPES = [
    { value: 'agent', label: 'Agen' },
    { value: 'agency', label: 'Agency / Travel' },
    { value: 'koordinator', label: 'Koordinator' },
  ];

  // Parent options for the agent form: every other agent except the one being
  // edited (a self-parent is rejected server-side anyway).
  let parentOptions = $derived(agents.filter((a) => !editingAgent || a.id !== editingAgent.id));

  function tierLabel(t) { return t > 1 ? `Tier ${t}` : 'Langsung'; }
  function tierColor(t) { return t > 1 ? 'var(--c-info)' : 'var(--c-primary)'; }

  // Summary stats (Suluk design) derived from existing data.
  let summaryStats = $derived({
    agentCount: agents.length,
    totalJamaah: agents.reduce((s, a) => s + (a.total_jamaah || 0), 0),
    totalOmzet: agents.reduce((s, a) => s + (a.total_commissions || 0), 0),
    outstanding: agents.reduce((s, a) => s + (a.total_outstanding || 0), 0),
  });

  async function loadData() {
    loading = true;
    try {
      const [agentData, commData] = await Promise.all([
        ApiService.listAgents({ search: searchQuery }),
        ApiService.listCommissions({ status: statusFilter, tier_level: tierFilter }),
      ]);
      agents = agentData.agents || [];
      commissions = commData.commissions || [];
    } catch (e) { showToast(e.message, 'error'); } finally { loading = false; }
  }
  onMount(() => { loadData(); });

  function openNewAgent() { editingAgent = null; agentForm = { ...emptyAgentForm }; showAgentDrawer = true; }
  function editAgent(a) { editingAgent = a; agentForm = { name: a.name, phone: a.phone, email: a.email, address: a.address, commission_rate: a.commission_rate, bank_name: a.bank_name, bank_account_number: a.bank_account_number, bank_account_name: a.bank_account_name, notes: a.notes, parent_id: a.parent_id || '', type: a.type || 'agent' }; showAgentDrawer = true; }

  async function openTierConfig() {
    try { const data = await ApiService.getTiers(); tiers = (data.tiers || []).map((t) => ({ ...t })); }
    catch (e) { tiers = []; showToast(e.message, 'error'); }
    if (tiers.length === 0) tiers = [{ level: 2, rate_pct: 2 }, { level: 3, rate_pct: 1 }];
    showTierDrawer = true;
  }
  function addTier() { const next = (tiers.at(-1)?.level || 1) + 1; tiers = [...tiers, { level: next, rate_pct: 0 }]; }
  function removeTier(i) { tiers = tiers.filter((_, idx) => idx !== i); }
  async function saveTiers() {
    savingTiers = true;
    try {
      await ApiService.setTiers(tiers.map((t) => ({ level: Number(t.level), rate_pct: Number(t.rate_pct) })));
      showToast('Pengaturan tier disimpan');
      showTierDrawer = false;
    } catch (e) { showToast(e.message, 'error'); } finally { savingTiers = false; }
  }

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
    agentCommissions = []; agentDownline = []; agentUpline = [];
    showAgentCommissions = true;
    try {
      const [comm, down, up] = await Promise.all([
        ApiService.getAgentCommissions(a.id),
        ApiService.getDownline(a.id).catch(() => ({ downline: [] })),
        ApiService.getUpline(a.id).catch(() => ({ upline: [] })),
      ]);
      agentCommissions = comm.commissions || [];
      agentDownline = down.downline || [];
      agentUpline = up.upline || [];
    } catch (e) { /* best-effort */ }
  }
</script>

<div class="flex flex-col gap-6 p-4 lg:p-8">
  <PageHeader kicker="Komisi Agen" title="Agen &amp; Mitra" subtitle="Pantau performa agen penjualan dan komisi yang harus dibayarkan.">
    {#snippet actions()}
      {#if tab === 'agents'}
        <Button variant="ghost" icon={Network} onclick={openTierConfig}>Tier Komisi</Button>
        <Button variant="primary" icon={Plus} onclick={openNewAgent}>Tambah Agen</Button>
      {:else}
        <Button variant="primary" icon={Plus} onclick={openNewCommission}>Catat Komisi</Button>
      {/if}
    {/snippet}
  </PageHeader>

  {#if loading}
    <div class="flex items-center justify-center py-16"><div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-600 border-t-transparent"></div></div>
  {:else}
    <!-- Summary cards (Suluk design) -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard icon={UserCheck} label="Agen Aktif" value={`${summaryStats.agentCount}`} accent="var(--c-primary)" />
      <StatCard icon={Users} label="Jamaah dari Agen" value={`${summaryStats.totalJamaah}`} accent="var(--c-info)" />
      <StatCard icon={TrendingUp} label="Total Omzet Agen" value={formatIDR(summaryStats.totalOmzet)} accent="var(--c-accent)" />
      <StatCard icon={Wallet} label="Komisi Terutang" value={formatIDR(summaryStats.outstanding)} accent="var(--c-warning)" />
    </div>

    <!-- Tabs + filters -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <FilterTabs
        tabs={[
          { value: 'agents', label: 'Agen', count: agents.length },
          { value: 'commissions', label: 'Komisi', count: commissions.length },
        ]}
        value={tab}
        onChange={(v) => (tab = v)}
      />
      {#if tab === 'agents'}
        <div class="relative">
          <Search class="absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2" style="color:var(--c-faint)" />
          <input
            type="text"
            bind:value={searchQuery}
            oninput={loadData}
            placeholder="Cari agen..."
            class="w-48 rounded-[var(--radius)] py-2 pl-9 pr-3 text-[13.5px] outline-none focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
            style="border:1px solid var(--c-line);color:var(--c-ink)"
          />
        </div>
      {:else}
        <div class="flex items-center gap-2">
          <select
            bind:value={tierFilter}
            onchange={loadData}
            class="rounded-[var(--radius)] bg-white px-3 py-2 text-[13.5px] font-medium outline-none focus:border-primary-400"
            style="border:1px solid var(--c-line);color:var(--c-ink)"
          >
            <option value="">Semua Tier</option><option value="1">Langsung</option><option value="2">Tier 2</option><option value="3">Tier 3</option>
          </select>
          <select
            bind:value={statusFilter}
            onchange={loadData}
            class="rounded-[var(--radius)] bg-white px-3 py-2 text-[13.5px] font-medium outline-none focus:border-primary-400"
            style="border:1px solid var(--c-line);color:var(--c-ink)"
          >
            <option value="all">Semua</option><option value="pending">Pending</option><option value="paid">Dibayar</option>
          </select>
        </div>
      {/if}
    </div>

    <!-- Agents Tab -->
    {#if tab === 'agents'}
      <Card pad={false} style="padding:8px 4px">
        {#if agents.length === 0}
          <div class="flex flex-col items-center justify-center py-16" style="color:var(--c-faint)"><Users class="mb-2 h-10 w-10" /><p class="text-sm">Belum ada agen</p></div>
        {:else}
          <div class="overflow-x-auto">
            <table class="w-full" style="border-collapse:collapse;font-size:13.5px">
              <thead>
                <tr>
                  <th style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Agen / Mitra</th>
                  <th style="text-align:center;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Komisi</th>
                  <th class="hidden md:table-cell" style="text-align:right;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Total Komisi</th>
                  <th class="hidden lg:table-cell" style="text-align:right;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Dibayar</th>
                  <th style="text-align:right;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Outstanding</th>
                  <th style="text-align:center;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Jamaah</th>
                  <th style="padding:0 16px 12px;border-bottom:1px solid var(--c-line)"></th>
                </tr>
              </thead>
              <tbody>
                {#each agents as a}
                  <tr class="suluk-row" style="transition:background .12s">
                    <td style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;white-space:nowrap">
                      <div class="flex items-center gap-3">
                        <Avatar name={a.name} size={40} />
                        <div class="min-w-0">
                          <div class="flex items-center gap-1.5">
                            <p class="truncate font-bold" style="color:var(--c-ink)">{a.name}</p>
                            {#if a.level > 1}
                              <span class="rounded-full px-1.5 py-0.5 text-[10px] font-bold" style="background:var(--c-info-soft, #e0f2fe);color:var(--c-info)" title="Kedalaman hierarki">L{a.level}</span>
                            {/if}
                          </div>
                          <p class="truncate" style="font-size:12px;color:var(--c-faint);margin-top:2px">
                            {#if a.parent_name}↳ {a.parent_name}{:else}{a.phone || '-'}{/if}
                          </p>
                        </div>
                      </div>
                    </td>
                    <td style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;text-align:center;white-space:nowrap">
                      <span style="font-weight:600;color:var(--c-ink-soft)">{a.commission_rate}%</span>
                    </td>
                    <td class="hidden md:table-cell" style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;text-align:right;font-weight:700;color:var(--c-ink);font-variant-numeric:tabular-nums;white-space:nowrap">{formatIDR(a.total_commissions)}</td>
                    <td class="hidden lg:table-cell" style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;text-align:right;font-weight:800;color:var(--c-success);font-variant-numeric:tabular-nums;white-space:nowrap">{formatIDR(a.total_paid)}</td>
                    <td style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;text-align:right;font-weight:700;color:var(--c-warning);font-variant-numeric:tabular-nums;white-space:nowrap">{formatIDR(a.total_outstanding)}</td>
                    <td style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;text-align:center;font-weight:700;color:var(--c-ink);font-variant-numeric:tabular-nums;white-space:nowrap">{a.total_jamaah}</td>
                    <td style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;text-align:right;white-space:nowrap">
                      <div class="flex justify-end gap-2">
                        <button type="button" onclick={() => editAgent(a)} class="inline-flex items-center gap-1 rounded-lg px-2.5 py-1.5 text-xs font-semibold" style="border:1px solid var(--c-line);color:var(--c-ink-soft)"><Pencil class="h-3 w-3" />Edit</button>
                        <button type="button" onclick={() => viewAgentCommissions(a)} class="rounded-lg px-2.5 py-1.5 text-xs font-semibold" style="background:var(--c-primary-soft);color:var(--c-primary-deep)">Komisi</button>
                        <button type="button" onclick={() => openPortal(a)} class="inline-flex items-center gap-1 rounded-lg px-2.5 py-1.5 text-xs font-semibold" style="border:1px solid var(--c-line);color:var(--c-ink-soft)" title="Buat akun login portal agen"><KeyRound class="h-3 w-3" />Portal</button>
                      </div>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </Card>
    {/if}

    <!-- Commissions Tab -->
    {#if tab === 'commissions'}
      <Card pad={false} style="padding:8px 4px">
        {#if commissions.length === 0}
          <div class="flex flex-col items-center justify-center py-16" style="color:var(--c-faint)"><DollarSign class="mb-2 h-10 w-10" /><p class="text-sm">Belum ada komisi</p></div>
        {:else}
          <div class="overflow-x-auto">
            <table class="w-full" style="border-collapse:collapse;font-size:13.5px">
              <thead>
                <tr>
                  <th style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Agen</th>
                  <th style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Jamaah</th>
                  <th class="hidden md:table-cell" style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Paket</th>
                  <th style="text-align:right;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Jumlah</th>
                  <th style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Status</th>
                  <th style="padding:0 16px 12px;border-bottom:1px solid var(--c-line)"></th>
                </tr>
              </thead>
              <tbody>
                {#each commissions as c}
                  <tr class="suluk-row" style="transition:background .12s">
                    <td style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;white-space:nowrap">
                      <div class="flex items-center gap-3">
                        <Avatar name={c.agent_name} size={36} />
                        <div class="min-w-0">
                          <span class="font-bold" style="color:var(--c-ink)">{c.agent_name}</span>
                          {#if c.tier_level > 1}
                            <span class="ml-1 rounded-full px-1.5 py-0.5 text-[10px] font-bold" style="background:var(--c-info-soft, #e0f2fe);color:var(--c-info)">{tierLabel(c.tier_level)}</span>
                          {/if}
                        </div>
                      </div>
                    </td>
                    <td style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;color:var(--c-ink-soft);white-space:nowrap">{c.jamaah_name || '-'}</td>
                    <td class="hidden md:table-cell" style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;font-size:12px;color:var(--c-muted);white-space:nowrap">{c.package_name || '-'}</td>
                    <td style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;text-align:right;font-weight:700;color:var(--c-ink);font-variant-numeric:tabular-nums;white-space:nowrap">{formatIDR(c.commission_amount)}</td>
                    <td style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;white-space:nowrap"><StatusBadge status={c.status} size="xs" /></td>
                    <td style="padding:calc(var(--row-h) / 4.2) 16px;border-bottom:1px solid var(--c-line-soft);vertical-align:middle;text-align:right;white-space:nowrap">{#if c.status === 'pending'}<button type="button" onclick={() => payCommission(c.id)} class="inline-flex items-center gap-1 rounded-lg px-2.5 py-1.5 text-xs font-semibold" style="background:var(--c-success-soft);color:var(--c-success)"><CheckCheck class="h-3 w-3" />Bayar</button>{/if}</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </Card>
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
    <div class="grid grid-cols-2 gap-3">
      <div class="flex flex-col gap-1">
        <label for="a-parent" class="text-xs font-medium text-slate-700">Upline (Atasan)</label>
        <select id="a-parent" bind:value={agentForm.parent_id} class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none focus:border-primary-400">
          <option value="">— Tanpa upline —</option>
          {#each parentOptions as p}<option value={p.id}>{p.name}</option>{/each}
        </select>
      </div>
      <div class="flex flex-col gap-1">
        <label for="a-type" class="text-xs font-medium text-slate-700">Tipe</label>
        <select id="a-type" bind:value={agentForm.type} class="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none focus:border-primary-400">
          {#each AGENT_TYPES as t}<option value={t.value}>{t.label}</option>{/each}
        </select>
      </div>
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
    {#if agentUpline.length}
      <div class="mb-4">
        <p class="mb-1.5 text-[11px] font-bold uppercase tracking-wider text-slate-400">Upline</p>
        <div class="flex flex-wrap items-center gap-1.5">
          {#each agentUpline as u, i}
            {#if i > 0}<span class="text-slate-300">›</span>{/if}
            <span class="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-semibold text-slate-600">{u.name}</span>
          {/each}
        </div>
      </div>
    {/if}

    {#if agentDownline.length}
      <div class="mb-4">
        <p class="mb-1.5 text-[11px] font-bold uppercase tracking-wider text-slate-400">Downline ({agentDownline.length})</p>
        <div class="space-y-1">
          {#each agentDownline as d}
            <div class="flex items-center justify-between rounded-lg border border-slate-100 px-3 py-2" style="margin-left:{(d.depth - 1) * 16}px">
              <div class="flex items-center gap-2">
                <span class="text-slate-300">↳</span>
                <span class="text-sm font-medium text-slate-700">{d.name}</span>
                {#if !d.is_active}<span class="rounded-full bg-slate-100 px-1.5 py-0.5 text-[10px] text-slate-400">nonaktif</span>{/if}
              </div>
              <span class="text-xs font-semibold text-slate-500">{d.total_jamaah} jamaah · {formatIDR(d.total_commissions)}</span>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <p class="mb-1.5 text-[11px] font-bold uppercase tracking-wider text-slate-400">Komisi</p>
    {#if agentCommissions.length === 0}
      <p class="text-sm text-slate-400 text-center py-8">Belum ada komisi</p>
    {:else}
      <div class="space-y-2">
        {#each agentCommissions as c}
          <div class="rounded-xl border border-slate-100 p-3 flex items-center justify-between">
            <div>
              <div class="flex items-center gap-1.5">
                <p class="text-sm font-semibold text-slate-800">{c.jamaah_name || '-'}</p>
                {#if c.tier_level > 1}<span class="rounded-full px-1.5 py-0.5 text-[10px] font-bold" style="background:var(--c-info-soft, #e0f2fe);color:var(--c-info)">{tierLabel(c.tier_level)}</span>{/if}
              </div>
              <p class="text-xs text-slate-500">{c.package_name || '-'}</p>
            </div>
            <div class="text-right"><p class="text-sm font-bold text-slate-800">{formatIDR(c.commission_amount)}</p><StatusBadge status={c.status} size="xs" /></div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</SlideDrawer>

<!-- Tier configuration drawer -->
<SlideDrawer open={showTierDrawer} onClose={() => showTierDrawer = false} title="Pengaturan Tier Komisi" width="460px">
  <div class="flex flex-col gap-4 p-4">
    <p class="text-sm text-slate-500">Rate komisi berjenjang untuk upline penjual. Tier 1 (penjual langsung) memakai komisi yang dicatat; tier 2+ dihitung dari nilai penjualan yang sama.</p>
    <div class="space-y-2">
      {#each tiers as t, i}
        <div class="flex items-center gap-2">
          <span class="w-16 text-sm font-semibold text-slate-600">Tier {t.level}</span>
          <input type="number" bind:value={t.rate_pct} step="0.01" min="0" max="100" class="w-24 rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" />
          <span class="text-sm text-slate-400">%</span>
          <button type="button" onclick={() => removeTier(i)} class="ml-auto rounded-lg p-1.5 text-slate-400 hover:bg-slate-50 hover:text-red-500"><X class="h-4 w-4" /></button>
        </div>
      {/each}
    </div>
    <button type="button" onclick={addTier} class="self-start rounded-lg border border-dashed border-slate-300 px-3 py-1.5 text-xs font-semibold text-slate-500 hover:border-primary-400 hover:text-primary-600">+ Tambah Tier</button>
    <div class="flex gap-2 pt-2">
      <button type="button" onclick={() => showTierDrawer = false} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600">Batal</button>
      <button type="button" onclick={saveTiers} disabled={savingTiers} class="flex-1 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">{savingTiers ? '...' : 'Simpan'}</button>
    </div>
  </div>
</SlideDrawer>

<!-- Portal credential drawer -->
<SlideDrawer open={showPortalDrawer} onClose={() => showPortalDrawer = false} title={portalAgent ? `Akun Portal: ${portalAgent.name}` : 'Akun Portal'} width="440px">
  <div class="flex flex-col gap-4 p-4">
    <p class="text-sm text-slate-500">Buat akun login agar agen ini bisa masuk ke portal dan melihat komisi & jaringannya sendiri.</p>
    <div class="flex flex-col gap-1"><label for="p-email" class="text-xs font-medium text-slate-700">Email Login <span class="text-red-500">*</span></label><input id="p-email" type="email" bind:value={portalForm.email} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex flex-col gap-1"><label for="p-name" class="text-xs font-medium text-slate-700">Nama Tampilan</label><input id="p-name" type="text" bind:value={portalForm.name} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <div class="flex flex-col gap-1"><label for="p-pass" class="text-xs font-medium text-slate-700">Password <span class="text-red-500">*</span></label><input id="p-pass" type="text" bind:value={portalForm.password} placeholder="min. 6 karakter" class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" /></div>
    <p class="text-xs text-slate-400">Bagikan email & password ini ke agen. Mereka login di halaman login yang sama dan otomatis diarahkan ke portal agen.</p>
    <div class="flex gap-2 pt-2">
      <button type="button" onclick={() => showPortalDrawer = false} class="flex-1 rounded-xl border border-slate-200 py-2.5 text-sm font-semibold text-slate-600">Batal</button>
      <button type="button" onclick={savePortal} disabled={savingPortal} class="flex-1 rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">{savingPortal ? '...' : 'Buat Akun'}</button>
    </div>
  </div>
</SlideDrawer>

<style>
  .suluk-row:hover { background: var(--c-bg); }
</style>
