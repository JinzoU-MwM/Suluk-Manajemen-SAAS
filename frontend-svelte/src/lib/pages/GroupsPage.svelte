<script>
  import { onMount } from "svelte";
  import {
    AlertCircle,
    Building2,
    CalendarDays,
    Hotel,
    Layers,
    Loader2,
    Plus,
    UsersRound,
  } from "lucide-svelte";
  import { Plane } from "lucide-svelte";
  import PageHeader from "../components/PageHeader.svelte";
  import StatCard from "../components/StatCard.svelte";
  import EmptyState from "../components/EmptyState.svelte";
  import Card from "../components/ui/Card.svelte";
  import Button from "../components/ui/Button.svelte";
  import SlideDrawer from "../components/SlideDrawer.svelte";
  import { showToast, mapError } from "../services/toast.svelte.js";
  import { ApiService } from "../services/api.js";

  let { onNavigate = () => {} } = $props();

  let groups = $state([]);
  let isLoading = $state(true);
  let isCreating = $state(false);
  let showCreate = $state(false);
  let newGroupName = $state("");
  let newGroupDescription = $state("");
  let error = $state("");

  // Kloter / departure ops (Phase 5A)
  let packages = $state([]);
  let pkgName = $derived(new Map(packages.map((p) => [p.id, p.name])));
  let showDep = $state(false);
  let depGroup = $state(null);
  let depForm = $state({ package_id: "", departure_date: "" });
  let savingDep = $state(false);
  let manifest = $state(null);
  // Guide assignment (Phase 5B)
  let groupGuides = $state([]);
  let allGuides = $state([]);
  let assignGuideId = $state("");
  let assignRole = $state("leader");
  const GUIDE_ROLES = [["leader", "Ketua"], ["co_leader", "Wakil"], ["kesehatan", "Kesehatan"]];

  const DEP_STATUS = {
    draft:     { label: "Draf",        color: "var(--c-muted)" },
    siap:      { label: "Siap",        color: "var(--c-info)" },
    berangkat: { label: "Berangkat",   color: "var(--c-success)" },
    selesai:   { label: "Selesai",     color: "var(--c-primary)" },
    batal:     { label: "Batal",       color: "var(--c-danger)" },
  };
  function depMeta(s) { return DEP_STATUS[s] || DEP_STATUS.draft; }
  // Allowed next states per the server state machine.
  const NEXT = {
    draft:     [["siap", "Finalkan Manifest"], ["batal", "Batalkan"]],
    siap:      [["berangkat", "Berangkatkan"], ["draft", "Buka Lagi"], ["batal", "Batalkan"]],
    berangkat: [["selesai", "Selesaikan"]],
    selesai:   [],
    batal:     [["draft", "Aktifkan"]],
  };

  async function openDeparture(group) {
    depGroup = group;
    depForm = { package_id: group.package_id || "", departure_date: (group.departure_date || "").slice(0, 10) };
    manifest = null;
    groupGuides = [];
    assignGuideId = "";
    showDep = true;
    try { manifest = await ApiService.getGroupManifest(group.id); } catch (e) { /* best-effort */ }
    try { const gg = await ApiService.listGroupGuides(group.id); groupGuides = gg.assignments || []; } catch (e) {}
    try { const ag = await ApiService.listGuides(); allGuides = ag.guides || []; } catch (e) {}
  }

  async function assignGuide() {
    if (!assignGuideId) return;
    try {
      await ApiService.assignGuide({ guide_id: assignGuideId, group_id: depGroup.id, role: assignRole });
      const gg = await ApiService.listGroupGuides(depGroup.id);
      groupGuides = gg.assignments || [];
      assignGuideId = "";
      showToast("Pembimbing ditugaskan", "success");
    } catch (e) { showToast(mapError(e.message), "error"); }
  }

  async function unassignGuide(guideId) {
    try {
      await ApiService.unassignGuide(depGroup.id, guideId);
      groupGuides = groupGuides.filter((a) => a.guide_id !== guideId);
    } catch (e) { showToast(mapError(e.message), "error"); }
  }

  function roleLabel(r) { return GUIDE_ROLES.find((x) => x[0] === r)?.[1] || r; }

  async function saveDeparture() {
    savingDep = true;
    try {
      const g = await ApiService.setGroupDeparture(depGroup.id, depForm);
      patchGroup(g);
      depGroup = g;
      showToast("Detail keberangkatan disimpan", "success");
    } catch (e) { showToast(mapError(e.message), "error"); } finally { savingDep = false; }
  }

  async function transitionDep(status) {
    try {
      const g = await ApiService.transitionGroupDeparture(depGroup.id, status);
      patchGroup(g);
      depGroup = g;
      showToast(`Status: ${depMeta(g.departure_status).label}`, "success");
    } catch (e) { showToast(mapError(e.message), "error"); }
  }

  function patchGroup(g) {
    groups = groups.map((x) => (x.id === g.id ? { ...x, ...g } : x));
  }

  // Card stripe palette (Suluk design).
  const STRIPE_COLORS = ["#a9842f", "#c79a3e", "#0f7a5a", "#2563c9", "#1B7F5A", "#7a5ae0"];

  // Summary tiles derived from the loaded groups.
  let summaryStats = $derived({
    totalGroups: groups.length,
    totalJamaah: groups.reduce((s, g) => s + (g.member_count || 0), 0),
  });

  onMount(() => {
    loadGroups();
    ApiService.listPackages({ pageSize: 200 })
      .then((p) => { packages = p?.packages || p?.data || p || []; })
      .catch(() => {});
  });

  async function loadGroups() {
    isLoading = true;
    error = "";
    try {
      const result = await ApiService.listGroups();
      groups = result.groups || [];
    } catch (e) {
      error = e.message;
    } finally {
      isLoading = false;
    }
  }

  async function createGroup() {
    if (!newGroupName.trim()) return;
    isCreating = true;
    error = "";
    try {
      const group = await ApiService.createGroup(newGroupName.trim(), newGroupDescription.trim());
      groups = [group, ...groups];
      newGroupName = "";
      newGroupDescription = "";
      showCreate = false;
    } catch (e) {
      error = e.message;
    } finally {
      isCreating = false;
    }
  }

  function formatDate(value) {
    if (!value) return "-";
    return new Date(value).toLocaleDateString("id-ID", {
      day: "numeric",
      month: "short",
      year: "numeric",
    });
  }
</script>

<div class="grup-page">
  <PageHeader
    kicker="Manajemen"
    title="Grup & Hotel"
    subtitle="Atur grup keberangkatan sebelum masuk ke rooming hotel."
  >
    {#snippet actions()}
      <Button variant="primary" icon={Plus} onclick={() => (showCreate = !showCreate)}>
        Buat Grup Baru
      </Button>
    {/snippet}
  </PageHeader>

  <!-- Summary cards (Suluk design) -->
  <div class="stat-grid">
    <StatCard icon={Layers} label="Total Grup" value={`${summaryStats.totalGroups}`} accent="var(--c-primary)" />
    <StatCard icon={UsersRound} label="Total Jamaah" value={`${summaryStats.totalJamaah}`} accent="var(--c-accent)" />
  </div>

  {#if error}
    <div class="alert">
      <AlertCircle size={20} style="flex-shrink:0;margin-top:1px" />
      <span>{error}</span>
    </div>
  {/if}

  {#if showCreate}
    <Card style="margin-bottom:var(--gap)">
      <div class="create-form">
        <input
          bind:value={newGroupName}
          class="field"
          placeholder="Nama grup, mis. Umrah Maret 2026"
        />
        <input
          bind:value={newGroupDescription}
          class="field"
          placeholder="Catatan singkat"
        />
        <Button
          variant="primary"
          onclick={() => createGroup()}
          disabled={isCreating || !newGroupName.trim()}
          icon={isCreating ? Loader2 : null}
        >
          Simpan
        </Button>
      </div>
    </Card>
  {/if}

  {#if isLoading}
    <Card>
      <div class="loading">
        <Loader2 size={20} class="spin" style="color:var(--c-primary)" />
        Memuat grup...
      </div>
    </Card>
  {:else if groups.length === 0}
    <Card pad={false}>
      <EmptyState
        icon={Building2}
        title="Belum ada grup keberangkatan"
        text="Buat grup untuk menampung data jamaah, hotel, rooming, dan manifest."
      />
    </Card>
  {:else}
    <div class="group-grid">
      {#each groups as group, i}
        {@const stripe = STRIPE_COLORS[i % STRIPE_COLORS.length]}
        <Card hover pad={false} style="overflow:hidden">
          <div style="height:5px;background:{stripe}"></div>
          <div style="padding:var(--pad)">
            <div class="card-head">
              <div style="flex:1;min-width:0">
                <div style="display:flex;align-items:center;gap:8px">
                  <h2 class="group-name">{group.name}</h2>
                  <span class="dep-badge" style="background:{depMeta(group.departure_status).color}1a;color:{depMeta(group.departure_status).color}">{depMeta(group.departure_status).label}</span>
                </div>
                <p class="group-desc">{group.package_id ? (pkgName.get(group.package_id) || "Paket") + (group.departure_date ? " · " + formatDate(group.departure_date) : "") : (group.description || "Tanpa catatan")}</p>
              </div>
              <div class="group-icon" style="background:{stripe}18;color:{stripe}">
                <Building2 size={21} />
              </div>
            </div>

            <div class="stats-row">
              <div>
                <p class="stat-value tabular">{group.member_count || 0}</p>
                <p class="stat-label"><UsersRound size={13} /> jamaah</p>
              </div>
              <div>
                <p class="stat-date">{formatDate(group.updated_at || group.created_at)}</p>
                <p class="stat-label"><CalendarDays size={13} /> update</p>
              </div>
            </div>

            <div class="card-actions">
              <Button variant="ghost" full icon={Hotel} onclick={() => onNavigate("rooming")}>
                Rooming
              </Button>
              <Button variant="primary" full icon={Plane} onclick={() => openDeparture(group)}>
                Keberangkatan
              </Button>
            </div>
          </div>
        </Card>
      {/each}
    </div>
  {/if}
</div>

<!-- Departure (kloter) drawer -->
<SlideDrawer open={showDep} title={depGroup ? `Keberangkatan: ${depGroup.name}` : "Keberangkatan"} width="480px" onClose={() => (showDep = false)}>
  {#if depGroup}
    <div class="space-y-4 p-4">
      <div class="flex items-center gap-2">
        <span class="dep-badge" style="background:{depMeta(depGroup.departure_status).color}1a;color:{depMeta(depGroup.departure_status).color}">{depMeta(depGroup.departure_status).label}</span>
        <span class="text-xs" style="color:var(--c-faint)">{depGroup.member_count || 0} jamaah</span>
      </div>

      {#if depGroup.departure_status === "draft" || depGroup.departure_status === "batal"}
        <div class="flex flex-col gap-1">
          <label for="d-pkg" class="text-xs font-medium text-slate-700">Paket</label>
          <select id="d-pkg" bind:value={depForm.package_id} class="rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm outline-none focus:border-primary-400">
            <option value="">— Pilih paket —</option>
            {#each packages as p}<option value={p.id}>{p.name}</option>{/each}
          </select>
        </div>
        <div class="flex flex-col gap-1">
          <label for="d-date" class="text-xs font-medium text-slate-700">Tanggal Berangkat</label>
          <input id="d-date" type="date" bind:value={depForm.departure_date} class="rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-primary-400" />
        </div>
        <button type="button" onclick={saveDeparture} disabled={savingDep} class="w-full rounded-xl bg-primary-600 py-2.5 text-sm font-semibold text-white hover:bg-primary-700 disabled:opacity-50">{savingDep ? "..." : "Simpan Detail"}</button>
      {:else}
        <div class="rounded-xl p-3 text-sm" style="background:var(--c-bg-2)">
          <p style="color:var(--c-ink)">{depGroup.package_id ? (pkgName.get(depGroup.package_id) || "Paket") : "—"}</p>
          <p class="text-xs" style="color:var(--c-faint)">Berangkat: {formatDate(depGroup.departure_date)}</p>
        </div>
      {/if}

      <div class="flex flex-wrap gap-2 border-t border-slate-100 pt-3">
        {#each NEXT[depGroup.departure_status] || [] as [st, label]}
          <button type="button" onclick={() => transitionDep(st)} class="rounded-xl px-3 py-2 text-xs font-semibold" style="border:1px solid var(--c-line);color:var(--c-ink-soft)">{label}</button>
        {/each}
      </div>

      <div>
        <p class="mb-1.5 text-[11px] font-bold uppercase tracking-wider" style="color:var(--c-faint)">Pembimbing ({groupGuides.length})</p>
        {#each groupGuides as a (a.guide_id)}
          <div class="mb-1 flex items-center justify-between rounded-lg px-3 py-1.5 text-sm" style="background:var(--c-bg-2)">
            <span style="color:var(--c-ink)">{a.guide_name} <span class="text-xs" style="color:var(--c-faint)">· {roleLabel(a.role)}</span></span>
            <button type="button" onclick={() => unassignGuide(a.guide_id)} class="text-xs font-semibold" style="color:var(--c-danger)">Lepas</button>
          </div>
        {/each}
        <div class="mt-2 flex gap-2">
          <select bind:value={assignGuideId} class="flex-1 rounded-lg border border-slate-200 bg-white px-2 py-1.5 text-xs outline-none">
            <option value="">Pilih pembimbing…</option>
            {#each allGuides as g}<option value={g.id}>{g.name}</option>{/each}
          </select>
          <select bind:value={assignRole} class="rounded-lg border border-slate-200 bg-white px-2 py-1.5 text-xs outline-none">
            {#each GUIDE_ROLES as [val, lbl]}<option value={val}>{lbl}</option>{/each}
          </select>
          <button type="button" onclick={assignGuide} class="rounded-lg px-3 py-1.5 text-xs font-semibold text-white" style="background:var(--c-primary)">Tugaskan</button>
        </div>
      </div>

      {#if manifest?.members?.length}
        <div>
          <p class="mb-1.5 text-[11px] font-bold uppercase tracking-wider" style="color:var(--c-faint)">Manifest ({manifest.members.length})</p>
          <div class="max-h-60 space-y-1 overflow-y-auto">
            {#each manifest.members as m}
              <div class="flex items-center justify-between rounded-lg px-3 py-1.5 text-sm" style="background:var(--c-bg-2)">
                <span style="color:var(--c-ink)">{m.name}</span>
                <span class="text-xs" style="color:var(--c-faint)">{m.phone || ""}</span>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  {/if}
</SlideDrawer>

<style>
  .dep-badge {
    flex-shrink: 0;
    padding: 2px 8px;
    border-radius: 999px;
    font-size: 10.5px;
    font-weight: 800;
  }
  .grup-page {
    min-height: 100vh;
    background: var(--c-bg);
    padding: 16px;
  }
  @media (min-width: 1024px) {
    .grup-page {
      padding: 32px;
    }
  }

  .stat-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: var(--gap);
    margin-bottom: var(--gap);
  }

  .alert {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: var(--gap);
    padding: 16px;
    border: 1px solid var(--c-danger-soft);
    background: var(--c-danger-soft);
    border-radius: var(--radius-lg);
    font-size: 14px;
    color: var(--c-danger);
  }

  .create-form {
    display: grid;
    gap: 12px;
    grid-template-columns: 1fr;
  }
  @media (min-width: 768px) {
    .create-form {
      grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) auto;
    }
  }

  .field {
    width: 100%;
    padding: 12px 16px;
    font-size: 14px;
    color: var(--c-ink);
    background: var(--c-surface);
    border: 1px solid var(--c-line);
    border-radius: var(--radius);
    outline: none;
    transition: border-color 0.15s, box-shadow 0.15s;
  }
  .field:focus {
    border-color: var(--c-primary);
    box-shadow: 0 0 0 3px var(--c-primary-soft);
  }

  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 32px;
    font-size: 14px;
    color: var(--c-muted);
  }

  .group-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: var(--gap);
  }

  .card-head {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 12px;
  }

  .group-name {
    margin: 0;
    font-family: var(--font-display, "Playfair Display", serif);
    font-size: 16.5px;
    font-weight: 800;
    line-height: 1.25;
    color: var(--c-ink);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .group-desc {
    margin: 4px 0 0;
    font-size: 13px;
    color: var(--c-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .group-icon {
    width: 44px;
    height: 44px;
    flex-shrink: 0;
    border-radius: var(--radius);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .stats-row {
    display: flex;
    gap: 24px;
    margin-top: 18px;
  }

  .stat-value {
    margin: 0;
    font-size: 22px;
    font-weight: 800;
    line-height: 1;
    color: var(--c-ink);
  }

  .stat-date {
    margin: 0;
    font-size: 14px;
    font-weight: 700;
    color: var(--c-ink);
  }

  .stat-label {
    display: flex;
    align-items: center;
    gap: 4px;
    margin: 6px 0 0;
    font-size: 12px;
    color: var(--c-faint);
  }

  .card-actions {
    display: flex;
    gap: 8px;
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid var(--c-line-soft);
  }

  .tabular {
    font-variant-numeric: tabular-nums;
  }

  :global(.spin) {
    animation: grup-spin 1s linear infinite;
  }
  @keyframes grup-spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
