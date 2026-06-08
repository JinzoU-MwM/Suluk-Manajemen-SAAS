<!--
  InventoryPage.svelte — Suluk design (Inventaris)

  Header (kicker + title + actions) · StatCard summary row · brand inventory table.
  Presentation only — all data fetching, state, handlers, and labels preserved.
-->
<script>
  import {
    Package,
    Shirt,
    CheckCircle,
    Loader2,
    AlertTriangle,
    RefreshCw,
    X,
    Boxes,
    Clock,
  } from "lucide-svelte";
  import PageHeader from "../components/PageHeader.svelte";
  import StatCard from "../components/StatCard.svelte";
  import Avatar from "../components/Avatar.svelte";
  import { ApiService } from "../services/api.js";

  let { isOpen = false, onClose, groups = [], isPro = false } = $props();

  // State
  let selectedGroupId = $state(null);
  let forecast = $state(null);
  let fulfillment = $state(null);
  let isLoading = $state(false);
  let error = $state(null);
  let selectedMembers = $state(new Set());
  let isMarking = $state(false);

  // Derived values
  let selectedGroup = $derived(groups.find((g) => g.id === selectedGroupId));
  let canMarkSelected = $derived(selectedMembers.size > 0 && !isMarking);

  // Size options
  const sizes = ["", "S", "M", "L", "XL", "XXL"];

  // Fetch forecast when group changes
  async function loadGroupData() {
    if (!selectedGroupId) {
      forecast = null;
      fulfillment = null;
      return;
    }

    isLoading = true;
    error = null;

    try {
      const [forecastData, fulfillmentData] = await Promise.all([
        ApiService.getInventoryForecast(selectedGroupId),
        ApiService.getFulfillmentStatus(selectedGroupId),
      ]);

      forecast = forecastData;
      fulfillment = fulfillmentData;
      selectedMembers = new Set();
    } catch (e) {
      error = e.message;
    } finally {
      isLoading = false;
    }
  }

  // Mark selected members as received
  async function markSelectedAsReceived() {
    if (!selectedGroupId || selectedMembers.size === 0) return;

    isMarking = true;
    try {
      const memberIds = Array.from(selectedMembers);
      await ApiService.markMembersReceived(selectedGroupId, memberIds);

      if (fulfillment) {
        for (const id of memberIds) {
          const pending = fulfillment.pending.find((m) => m.id === id);
          if (pending) {
            pending.is_equipment_received = true;
            fulfillment.received.push(pending);
          }
        }
        fulfillment.pending = fulfillment.pending.filter(
          (m) => !memberIds.includes(m.id),
        );
      }

      selectedMembers = new Set();
    } catch (e) {
      error = e.message;
    } finally {
      isMarking = false;
    }
  }

  function toggleMember(memberId) {
    const newSet = new Set(selectedMembers);
    if (newSet.has(memberId)) {
      newSet.delete(memberId);
    } else {
      newSet.add(memberId);
    }
    selectedMembers = newSet;
  }

  function selectAllPending() {
    if (fulfillment) {
      const newSet = new Set();
      for (const m of fulfillment.pending) {
        newSet.add(m.id);
      }
      selectedMembers = newSet;
    }
  }

  function clearSelection() {
    selectedMembers = new Set();
  }

  async function updateBajuSize(memberId, size) {
    try {
      await ApiService.updateMemberOperational(memberId, size, "");
      if (forecast) {
        const detail = forecast.details.find((d) => d.member_id === memberId);
        if (detail) detail.baju_size = size;
      }
    } catch (e) {
      error = e.message;
    }
  }

  async function updateFamilyId(memberId, familyId) {
    try {
      const detail = forecast?.details.find((d) => d.member_id === memberId);
      await ApiService.updateMemberOperational(
        memberId,
        detail?.baju_size || "",
        familyId,
      );
      if (forecast) {
        const d = forecast.details.find((d) => d.member_id === memberId);
        if (d) d.family_id = familyId;
      }
    } catch (e) {
      error = e.message;
    }
  }
</script>

{#if isOpen}
  <div class="min-h-screen bg-slate-50/70 p-4 lg:p-8">
    <PageHeader
      kicker="Logistik"
      title="Inventory"
      subtitle="Kelola stok koper, ihram, mukena, dan distribusi perlengkapan jamaah."
    >
      {#snippet actions()}
        <select
          id="inv-group-select"
          bind:value={selectedGroupId}
          onchange={loadGroupData}
          class="w-full rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm font-medium text-slate-700 outline-none transition focus:border-primary-400 focus:ring-2 focus:ring-primary-100 sm:w-auto sm:min-w-[14rem]"
        >
          <option value="">Pilih Grup</option>
          {#each groups as group}
            <option value={group.id}>{group.name} ({group.member_count})</option>
          {/each}
        </select>

        {#if selectedGroupId}
          <button
            type="button"
            onclick={loadGroupData}
            disabled={isLoading}
            class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-xl border border-slate-200 text-slate-500 transition hover:bg-slate-100 disabled:opacity-50"
            title="Muat ulang"
          >
            <RefreshCw class="h-4 w-4 {isLoading ? 'animate-spin' : ''}" />
          </button>
        {/if}
      {/snippet}
    </PageHeader>

    <!-- Summary cards (Suluk design) -->
    {#if forecast}
      <div class="mb-5 grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard
          icon={Boxes}
          label="Total"
          value={String(forecast.total_members)}
          accent="#1B7F5A"
        />
        <StatCard
          icon={Package}
          label="Koper"
          value={String(forecast.requirements?.koper || 0)}
          accent="#2563a8"
        />
        <StatCard
          icon={CheckCircle}
          label="Ihram"
          value={String(forecast.requirements?.ihram || 0)}
          accent="#1B7F5A"
        />
        <StatCard
          icon={Shirt}
          label="Mukena"
          value={String(forecast.requirements?.mukena || 0)}
          accent="#C99A2E"
        />
      </div>
    {/if}

    <!-- Bulk actions toolbar -->
    {#if fulfillment && fulfillment.pending?.length > 0}
      <div class="mb-5 flex flex-wrap items-center gap-3 rounded-2xl border border-slate-200/70 bg-white px-4 py-3 shadow-sm">
        <span class="text-xs font-medium text-slate-500"
          >{selectedMembers.size} dipilih</span
        >
        <button
          type="button"
          onclick={selectAllPending}
          class="text-xs font-semibold text-primary-600 hover:underline"
          >Semua</button
        >
        <button
          type="button"
          onclick={clearSelection}
          class="text-xs font-medium text-slate-500 hover:underline">Batal</button
        >
        <button
          type="button"
          onclick={markSelectedAsReceived}
          disabled={!canMarkSelected}
          class="ml-auto inline-flex items-center gap-1.5 rounded-xl bg-primary-600 px-3.5 py-2 text-xs font-semibold text-white shadow-sm shadow-primary-600/30 transition hover:bg-primary-700 disabled:cursor-not-allowed disabled:opacity-50"
        >
          {#if isMarking}<Loader2
              class="h-3.5 w-3.5 animate-spin"
            />{:else}<CheckCircle class="h-3.5 w-3.5" />{/if}
          Tandai Terima
        </button>
      </div>
    {/if}

    <!-- Error -->
    {#if error}
      <div
        class="mb-5 flex items-center gap-2 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm"
      >
        <AlertTriangle class="h-4 w-4 text-red-500" />
        <span class="flex-1 text-red-700">{error}</span>
        <button
          type="button"
          onclick={() => (error = null)}
          class="text-red-500 hover:text-red-700"
        >
          <X class="h-4 w-4" />
        </button>
      </div>
    {/if}

    <!-- Content -->
    <div class="overflow-hidden rounded-2xl border border-slate-200/70 bg-white shadow-sm">
      {#if isLoading}
        <div class="flex items-center justify-center py-16">
          <Loader2 class="h-6 w-6 animate-spin text-primary-500" />
        </div>
      {:else if !selectedGroupId}
        <div
          class="flex flex-col items-center justify-center py-16 text-slate-400"
        >
          <Package class="mb-2 h-10 w-10" />
          <p class="text-sm">Pilih grup untuk memulai</p>
        </div>
      {:else if forecast}
        <!-- Size & Status summary row -->
        <div class="flex flex-wrap gap-3 border-b border-slate-100 p-4">
          <!-- Sizes -->
          <div class="min-w-[220px] flex-1 rounded-xl border border-slate-200/70 bg-slate-50/60 p-3">
            <div class="mb-2 flex items-center gap-1.5">
              <Shirt class="h-3.5 w-3.5 text-slate-500" />
              <span class="text-xs font-semibold text-slate-600">Ukuran Baju</span
              >
            </div>
            <div class="flex flex-wrap gap-1.5">
              {#each Object.entries(forecast.size_breakdown || {}) as [size, count]}
                <span
                  class="rounded-md bg-white px-2 py-0.5 text-xs font-medium text-slate-700 ring-1 ring-slate-200"
                >
                  {size || "N/A"}: {count}
                </span>
              {/each}
            </div>
          </div>
          <!-- Status -->
          {#if fulfillment}
            <div
              class="min-w-[160px] rounded-xl border border-slate-200/70 bg-slate-50/60 p-3"
            >
              <div class="mb-2 flex items-center gap-1.5">
                <Clock class="h-3.5 w-3.5 text-slate-500" />
                <span class="text-xs font-semibold text-slate-600">Status</span>
              </div>
              <div class="flex items-center gap-4 text-xs">
                <span class="flex items-center gap-1">
                  <CheckCircle class="h-3.5 w-3.5 text-primary-500" />
                  <span class="font-semibold text-primary-700"
                    >{fulfillment.received_count || 0}</span
                  >
                </span>
                <span class="flex items-center gap-1">
                  <Package class="h-3.5 w-3.5 text-gold-500" />
                  <span class="font-semibold text-amber-600"
                    >{fulfillment.pending_count || 0}</span
                  >
                </span>
              </div>
            </div>
          {/if}
        </div>

        <!-- Member Table -->
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead>
              <tr class="text-left">
                <th class="w-10 px-4 py-3">
                  <input
                    type="checkbox"
                    checked={fulfillment?.pending?.length > 0 &&
                      selectedMembers.size === fulfillment?.pending?.length}
                    onchange={() => {
                      if (selectedMembers.size === fulfillment?.pending?.length)
                        clearSelection();
                      else selectAllPending();
                    }}
                    class="rounded border-slate-300 text-primary-600 focus:ring-primary-500"
                  />
                </th>
                <th class="px-4 py-3 text-[11.5px] font-semibold uppercase tracking-wider text-slate-400"
                  >Nama</th
                >
                <th
                  class="w-16 px-4 py-3 text-[11.5px] font-semibold uppercase tracking-wider text-slate-400"
                  >Gender</th
                >
                <th
                  class="w-24 px-4 py-3 text-[11.5px] font-semibold uppercase tracking-wider text-slate-400"
                  >Baju</th
                >
                <th
                  class="w-24 px-4 py-3 text-[11.5px] font-semibold uppercase tracking-wider text-slate-400"
                  >Family</th
                >
                <th
                  class="w-24 px-4 py-3 text-[11.5px] font-semibold uppercase tracking-wider text-slate-400"
                  >Status</th
                >
              </tr>
            </thead>
            <tbody>
              {#each forecast.details || [] as member}
                {@const isReceived = member.is_equipment_received}
                <tr
                  class="transition-colors hover:bg-primary-50/30 {isReceived
                    ? 'bg-primary-50/20'
                    : ''}"
                >
                  <td class="border-b border-slate-100 px-4 py-3.5">
                    {#if !isReceived}
                      <input
                        type="checkbox"
                        checked={selectedMembers.has(member.member_id)}
                        onchange={() => toggleMember(member.member_id)}
                        class="rounded border-slate-300 text-primary-600 focus:ring-primary-500"
                      />
                    {:else}
                      <CheckCircle class="h-4 w-4 text-primary-500" />
                    {/if}
                  </td>
                  <td class="border-b border-slate-100 px-4 py-3.5">
                    <div class="flex items-center gap-2.5">
                      <Avatar name={member.nama} size={32} />
                      <span
                        class="max-w-[180px] truncate text-sm font-semibold text-[#10211c]"
                        >{member.nama}</span
                      >
                    </div>
                  </td>
                  <td class="border-b border-slate-100 px-4 py-3.5">
                    <span
                      class="rounded-md px-2 py-0.5 text-[10px] font-semibold {member.gender ===
                      'male'
                        ? 'bg-blue-100 text-blue-700'
                        : 'bg-pink-100 text-pink-700'}"
                    >
                      {member.gender === "male" ? "L" : "P"}
                    </span>
                  </td>
                  <td class="border-b border-slate-100 px-4 py-3.5">
                    <select
                      value={member.baju_size || ""}
                      onchange={(e) =>
                        updateBajuSize(
                          member.member_id,
                          /** @type {HTMLSelectElement} */ (e.target).value,
                        )}
                      class="w-full rounded-lg border border-slate-200 bg-white px-2 py-1.5 text-xs outline-none focus:border-primary-400"
                    >
                      {#each sizes as s}<option value={s}>{s || "-"}</option
                        >{/each}
                    </select>
                  </td>
                  <td class="border-b border-slate-100 px-4 py-3.5">
                    <input
                      type="text"
                      value={member.family_id || ""}
                      onblur={(e) =>
                        updateFamilyId(
                          member.member_id,
                          /** @type {HTMLInputElement} */ (e.target).value,
                        )}
                      placeholder="F00"
                      class="w-full rounded-lg border border-slate-200 bg-white px-2 py-1.5 text-xs outline-none focus:border-primary-400"
                    />
                  </td>
                  <td class="border-b border-slate-100 px-4 py-3.5">
                    {#if isReceived}
                      <span
                        class="inline-flex items-center gap-1 rounded-full bg-primary-50 px-2.5 py-0.5 text-[11px] font-semibold text-primary-700"
                      >
                        <CheckCircle class="h-3 w-3" /> Diterima
                      </span>
                    {:else}
                      <span
                        class="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2.5 py-0.5 text-[11px] font-semibold text-amber-700"
                      >
                        Menunggu
                      </span>
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else}
        <div class="py-8 text-center text-sm text-slate-400">
          Tidak ada data
        </div>
      {/if}
    </div>
  </div>
{/if}
