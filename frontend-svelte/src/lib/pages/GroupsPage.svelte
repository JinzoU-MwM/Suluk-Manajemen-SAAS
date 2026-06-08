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

<div class="min-h-screen bg-slate-50/70 p-4 lg:p-8">
  <PageHeader
    kicker="Manajemen"
    title="Grup & Hotel"
    subtitle="Atur grup keberangkatan sebelum masuk ke rooming hotel."
  >
    {#snippet actions()}
      <button
        type="button"
        onclick={() => (showCreate = !showCreate)}
        class="inline-flex items-center justify-center gap-2 rounded-xl bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm shadow-primary-600/30 transition-all hover:bg-primary-700"
      >
        <Plus class="h-4 w-4" />
        Buat Grup Baru
      </button>
    {/snippet}
  </PageHeader>

  <!-- Summary cards (Suluk design) -->
  <div class="mb-6 grid grid-cols-2 gap-3 sm:grid-cols-4">
    <StatCard icon={Layers} label="Total Grup" value={`${summaryStats.totalGroups}`} accent="#1B7F5A" />
    <StatCard icon={UsersRound} label="Total Jamaah" value={`${summaryStats.totalJamaah}`} accent="#C99A2E" />
  </div>

  {#if error}
    <div class="mb-5 flex items-start gap-3 rounded-2xl border border-red-100 bg-red-50 p-4 text-sm text-red-700">
      <AlertCircle class="mt-0.5 h-5 w-5 flex-shrink-0" />
      <span>{error}</span>
    </div>
  {/if}

  {#if showCreate}
    <div class="mb-6 rounded-2xl border border-slate-200/70 bg-white p-5 shadow-sm">
      <div class="grid gap-3 md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
        <input
          bind:value={newGroupName}
          class="rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm outline-none transition focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
          placeholder="Nama grup, mis. Umrah Maret 2026"
        />
        <input
          bind:value={newGroupDescription}
          class="rounded-xl border border-slate-200 bg-white px-4 py-3 text-sm outline-none transition focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
          placeholder="Catatan singkat"
        />
        <button
          type="button"
          onclick={() => createGroup()}
          disabled={isCreating || !newGroupName.trim()}
          class="inline-flex items-center justify-center gap-2 rounded-xl bg-primary-600 px-5 py-3 text-sm font-semibold text-white transition hover:bg-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {#if isCreating}
            <Loader2 class="h-4 w-4 animate-spin" />
          {/if}
          Simpan
        </button>
      </div>
    </div>
  {/if}

  {#if isLoading}
    <div class="flex items-center justify-center gap-3 rounded-2xl border border-slate-200/70 bg-white p-12 text-sm text-slate-500 shadow-sm">
      <Loader2 class="h-5 w-5 animate-spin text-primary-500" />
      Memuat grup...
    </div>
  {:else if groups.length === 0}
    <div class="rounded-2xl border border-slate-200/70 bg-white p-12 text-center shadow-sm">
      <div class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-primary-50">
        <Building2 class="h-7 w-7 text-primary-600" />
      </div>
      <h3 class="text-sm font-bold text-[#10211c]">Belum ada grup keberangkatan</h3>
      <p class="mt-1 text-sm text-slate-500">Buat grup untuk menampung data jamaah, hotel, rooming, dan manifest.</p>
    </div>
  {:else}
    <div class="grid gap-5 sm:grid-cols-2 xl:grid-cols-3">
      {#each groups as group, i}
        {@const stripe = STRIPE_COLORS[i % STRIPE_COLORS.length]}
        <div class="overflow-hidden rounded-2xl border border-slate-200/70 bg-white shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-lg">
          <div class="h-[5px]" style="background:{stripe}"></div>
          <div class="p-5">
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <h2 class="truncate font-serif text-[16.5px] font-extrabold text-[#10211c]">{group.name}</h2>
                <p class="mt-0.5 truncate text-[13px] text-slate-500">{group.description || "Tanpa catatan"}</p>
              </div>
              <div
                class="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl"
                style="background:{stripe}18;color:{stripe}"
              >
                <Building2 class="h-5 w-5" />
              </div>
            </div>

            <div class="mt-5 flex gap-6">
              <div>
                <p class="text-[22px] font-extrabold leading-none text-[#10211c]" style="font-variant-numeric:tabular-nums">{group.member_count || 0}</p>
                <p class="mt-1.5 flex items-center gap-1 text-xs text-slate-400">
                  <UsersRound class="h-3.5 w-3.5" />
                  jamaah
                </p>
              </div>
              <div>
                <p class="text-sm font-bold text-[#10211c]">{formatDate(group.updated_at || group.created_at)}</p>
                <p class="mt-1.5 flex items-center gap-1 text-xs text-slate-400">
                  <CalendarDays class="h-3.5 w-3.5" />
                  update
                </p>
              </div>
            </div>

            <div class="mt-5 flex gap-2 border-t border-slate-100 pt-4">
              <button
                type="button"
                onclick={() => onNavigate("jamaah")}
                class="flex-1 rounded-xl border border-slate-200 px-3 py-2.5 text-sm font-semibold text-slate-700 transition hover:bg-slate-50"
              >
                Lihat Jamaah
              </button>
              <button
                type="button"
                onclick={() => onNavigate("rooming")}
                class="inline-flex flex-1 items-center justify-center gap-2 rounded-xl bg-primary-600 px-3 py-2.5 text-sm font-semibold text-white transition hover:bg-primary-700"
              >
                <Hotel class="h-4 w-4" />
                Rooming
              </button>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>
