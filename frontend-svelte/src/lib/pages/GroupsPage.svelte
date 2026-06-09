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
  import PageHeader from "../components/PageHeader.svelte";
  import StatCard from "../components/StatCard.svelte";
  import EmptyState from "../components/EmptyState.svelte";
  import Card from "../components/ui/Card.svelte";
  import Button from "../components/ui/Button.svelte";
  import { ApiService } from "../services/api.js";

  let { onNavigate = () => {} } = $props();

  let groups = $state([]);
  let isLoading = $state(true);
  let isCreating = $state(false);
  let showCreate = $state(false);
  let newGroupName = $state("");
  let newGroupDescription = $state("");
  let error = $state("");

  // Card stripe palette (Suluk design).
  const STRIPE_COLORS = ["#a9842f", "#c79a3e", "#0f7a5a", "#2563c9", "#1B7F5A", "#7a5ae0"];

  // Summary tiles derived from the loaded groups.
  let summaryStats = $derived({
    totalGroups: groups.length,
    totalJamaah: groups.reduce((s, g) => s + (g.member_count || 0), 0),
  });

  onMount(loadGroups);

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
                <h2 class="group-name">{group.name}</h2>
                <p class="group-desc">{group.description || "Tanpa catatan"}</p>
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
              <Button variant="ghost" full onclick={() => onNavigate("jamaah")}>
                Lihat Jamaah
              </Button>
              <Button variant="primary" full icon={Hotel} onclick={() => onNavigate("rooming")}>
                Rooming
              </Button>
            </div>
          </div>
        </Card>
      {/each}
    </div>
  {/if}
</div>

<style>
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
