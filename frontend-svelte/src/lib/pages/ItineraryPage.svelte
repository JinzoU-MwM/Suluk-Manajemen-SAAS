<!--
  ItineraryPage.svelte - Trip schedule/itinerary manager.
  Timeline view grouped by date with color-coded categories.
-->
<script>
    import { onMount } from "svelte";
    import {
        CalendarDays,
        Plane,
        Hotel,
        Bus,
        MapPin,
        MoreHorizontal,
        Plus,
        Trash2,
        Edit3,
        X,
        Loader2,
        AlertCircle,
        Clock,
        Save,
    } from "lucide-svelte";
    import { ApiService } from "../services/api";
    import PageHeader from "../components/PageHeader.svelte";

    let { groups = [], isPro = false } = $props();

    let selectedGroupId = $state(null);
    let items = $state([]);
    let isLoading = $state(false);
    let error = $state("");
    let showForm = $state(false);
    let editingItem = $state(null);

    // Form state
    let formDate = $state("");
    let formTimeStart = $state("");
    let formTimeEnd = $state("");
    let formActivity = $state("");
    let formLocation = $state("");
    let formNotes = $state("");
    let formCategory = $state("activity");
    let isSaving = $state(false);

    // Grouped by date
    let groupedItems = $derived(() => {
        const grouped = {};
        for (const item of items) {
            if (!grouped[item.date]) grouped[item.date] = [];
            grouped[item.date].push(item);
        }
        // Sort dates
        return Object.entries(grouped).sort(([a], [b]) => a.localeCompare(b));
    });

    const categoryConfig = {
        flight: {
            icon: Plane,
            label: "Penerbangan",
            color: "bg-blue-50 text-blue-600 border-blue-200",
        },
        hotel: {
            icon: Hotel,
            label: "Hotel",
            color: "bg-purple-50 text-purple-600 border-purple-200",
        },
        transport: {
            icon: Bus,
            label: "Transportasi",
            color: "bg-amber-50 text-amber-600 border-amber-200",
        },
        activity: {
            icon: MapPin,
            label: "Aktivitas",
            color: "bg-emerald-50 text-emerald-600 border-emerald-200",
        },
        other: {
            icon: MoreHorizontal,
            label: "Lainnya",
            color: "bg-slate-50 text-slate-600 border-slate-200",
        },
    };

    async function loadItinerary() {
        if (!selectedGroupId) return;
        isLoading = true;
        error = "";
        try {
            const res = await ApiService.getItinerary(selectedGroupId);
            items = res.items || [];
        } catch (e) {
            error = e.message;
        } finally {
            isLoading = false;
        }
    }

    function openAddForm() {
        editingItem = null;
        formDate = "";
        formTimeStart = "";
        formTimeEnd = "";
        formActivity = "";
        formLocation = "";
        formNotes = "";
        formCategory = "activity";
        showForm = true;
    }

    function openEditForm(item) {
        editingItem = item;
        formDate = item.date;
        formTimeStart = item.time_start;
        formTimeEnd = item.time_end;
        formActivity = item.activity;
        formLocation = item.location;
        formNotes = item.notes;
        formCategory = item.category;
        showForm = true;
    }

    async function handleSubmit() {
        if (!formActivity.trim() || !formDate.trim()) return;
        isSaving = true;
        error = "";
        try {
            const data = {
                date: formDate,
                time_start: formTimeStart,
                time_end: formTimeEnd,
                activity: formActivity,
                location: formLocation,
                notes: formNotes,
                category: formCategory,
            };
            if (editingItem) {
                await ApiService.updateItinerary(
                    selectedGroupId,
                    editingItem.id,
                    data,
                );
            } else {
                await ApiService.createItinerary(selectedGroupId, data);
            }
            showForm = false;
            await loadItinerary();
        } catch (e) {
            error = e.message;
        } finally {
            isSaving = false;
        }
    }

    async function deleteItem(itemId) {
        if (!confirm("Hapus item jadwal ini?")) return;
        try {
            await ApiService.deleteItinerary(selectedGroupId, itemId);
            items = items.filter((i) => i.id !== itemId);
        } catch (e) {
            error = e.message;
        }
    }

    function formatDate(dateStr) {
        try {
            return new Date(dateStr).toLocaleDateString("id-ID", {
                weekday: "long",
                day: "numeric",
                month: "long",
                year: "numeric",
            });
        } catch {
            return dateStr;
        }
    }
</script>

<div class="min-h-screen bg-slate-50/70 p-4 lg:p-8">
    <!-- Header -->
    <PageHeader
        kicker="Operasional"
        title="Itinerary"
        subtitle="Kelola jadwal perjalanan per grup keberangkatan."
    >
        {#snippet actions()}
            {#if selectedGroupId}
                <button
                    type="button"
                    onclick={openAddForm}
                    class="inline-flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm shadow-primary-600/30 transition hover:bg-primary-700"
                >
                    <Plus class="w-4 h-4" /> Tambah
                </button>
            {/if}
        {/snippet}
    </PageHeader>

    <!-- Error -->
    {#if error}
        <div
            class="mb-5 flex items-center gap-2 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600"
        >
            <AlertCircle class="w-4 h-4 flex-shrink-0" />
            {error}
            <button onclick={() => (error = "")} class="ml-auto"
                ><X class="w-4 h-4" /></button
            >
        </div>
    {/if}

    <!-- Group Selector -->
    <div
        class="mb-6 rounded-2xl border border-slate-200/70 bg-white p-4 shadow-sm"
    >
        <label
            for="itinerary-group"
            class="mb-2 block text-xs font-bold uppercase tracking-[0.08em] text-primary-600"
            >Pilih Grup</label
        >
        <select
            id="itinerary-group"
            class="w-full rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-medium text-slate-700 outline-none transition focus:border-primary-400 focus:bg-white focus:ring-2 focus:ring-primary-100"
            onchange={(e) => {
                selectedGroupId = parseInt(
                    /** @type {HTMLSelectElement} */ (e.target).value,
                );
                loadItinerary();
            }}
        >
            <option value="">-- Pilih grup --</option>
            {#each groups as g}
                <option value={g.id}>{g.name}</option>
            {/each}
        </select>
    </div>

    {#if !selectedGroupId}
        <div
            class="rounded-2xl border border-slate-200/70 bg-white py-16 text-center text-slate-400 shadow-sm"
        >
            <CalendarDays class="w-12 h-12 mx-auto mb-3 text-slate-300" />
            <p class="text-sm">
                Pilih grup terlebih dahulu untuk melihat jadwal
            </p>
        </div>
    {:else if isLoading}
        <div
            class="flex items-center justify-center rounded-2xl border border-slate-200/70 bg-white py-16 text-slate-400 shadow-sm"
        >
            <Loader2 class="w-6 h-6 animate-spin mr-2" /> Memuat jadwal...
        </div>
    {:else if items.length === 0}
        <div
            class="rounded-2xl border border-slate-200/70 bg-white py-16 text-center text-slate-400 shadow-sm"
        >
            <CalendarDays class="w-12 h-12 mx-auto mb-3 text-slate-300" />
            <p class="font-medium text-slate-500">Belum ada jadwal</p>
            <p class="text-sm mt-1">
                Klik "Tambah" untuk menambahkan item jadwal pertama
            </p>
        </div>
    {:else}
        <!-- Timeline -->
        <div
            class="rounded-2xl border border-slate-200/70 bg-white p-5 shadow-sm lg:p-6"
        >
            <div class="relative pl-1">
                {#each groupedItems() as [date, dateItems], dayIdx}
                    <div
                        class="relative flex gap-4 sm:gap-5 {dayIdx <
                        groupedItems().length - 1
                            ? 'pb-7'
                            : ''}"
                    >
                        <!-- Connector line -->
                        {#if dayIdx < groupedItems().length - 1}
                            <div
                                class="absolute left-[21px] top-12 bottom-0 w-0.5 bg-slate-200"
                            ></div>
                        {/if}

                        <!-- Day marker -->
                        <div
                            class="z-10 flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-full border-[3px] border-white bg-primary-50 text-primary-800 shadow-sm"
                        >
                            <CalendarDays class="h-5 w-5" />
                        </div>

                        <!-- Day content -->
                        <div class="min-w-0 flex-1 pt-1">
                            <p
                                class="text-[11.5px] font-bold uppercase tracking-[0.05em] text-[#C99A2E]"
                            >
                                {formatDate(date)}
                            </p>

                            <!-- Items -->
                            <div class="mt-3 space-y-2.5">
                                {#each dateItems as item}
                                    {@const cat =
                                        categoryConfig[item.category] ||
                                        categoryConfig.other}
                                    <div
                                        class="group rounded-xl border border-slate-200/70 bg-white p-3.5 transition-shadow hover:shadow-sm"
                                    >
                                        <div
                                            class="flex items-start justify-between"
                                        >
                                            <div
                                                class="flex min-w-0 flex-1 items-start gap-3"
                                            >
                                                <div
                                                    class="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-lg border {cat.color}"
                                                >
                                                    <cat.icon class="w-4 h-4" />
                                                </div>
                                                <div class="min-w-0 flex-1">
                                                    <p
                                                        class="text-sm font-semibold text-slate-800"
                                                    >
                                                        {item.activity}
                                                    </p>
                                                    {#if item.location}
                                                        <p
                                                            class="mt-0.5 text-xs text-slate-500"
                                                        >
                                                            ?? {item.location}
                                                        </p>
                                                    {/if}
                                                    {#if item.time_start}
                                                        <p
                                                            class="mt-0.5 flex items-center gap-1 text-xs text-slate-400"
                                                        >
                                                            <Clock
                                                                class="w-3 h-3"
                                                            />
                                                            {item.time_start}{item.time_end
                                                                ? ` - ${item.time_end}`
                                                                : ""}
                                                        </p>
                                                    {/if}
                                                    {#if item.notes}
                                                        <p
                                                            class="mt-1 text-xs italic text-slate-400"
                                                        >
                                                            {item.notes}
                                                        </p>
                                                    {/if}
                                                </div>
                                            </div>
                                            <div
                                                class="flex gap-1 opacity-0 transition-opacity group-hover:opacity-100"
                                            >
                                                <button
                                                    type="button"
                                                    onclick={() =>
                                                        openEditForm(item)}
                                                    class="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-primary-50 hover:text-primary-600"
                                                >
                                                    <Edit3 class="w-3.5 h-3.5" />
                                                </button>
                                                <button
                                                    type="button"
                                                    onclick={() =>
                                                        deleteItem(item.id)}
                                                    class="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-red-50 hover:text-red-500"
                                                >
                                                    <Trash2
                                                        class="w-3.5 h-3.5"
                                                    />
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                {/each}
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        </div>
    {/if}
</div>

<!-- Add/Edit Modal -->
{#if showForm}
    <div
        class="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4"
        role="dialog"
    >
        <div class="bg-white rounded-2xl shadow-2xl max-w-md w-full p-5">
            <div class="flex items-center justify-between mb-4">
                <h3 class="text-lg font-semibold text-slate-800">
                    {editingItem ? "Edit Jadwal" : "Tambah Jadwal"}
                </h3>
                <button
                    type="button"
                    onclick={() => (showForm = false)}
                    class="p-1.5 hover:bg-slate-100 rounded-lg"
                >
                    <X class="w-4 h-4" />
                </button>
            </div>

            <div class="space-y-3">
                <!-- Category -->
                <div>
                    <label
                        for="itinerary-category"
                        class="block text-xs font-medium text-slate-600 mb-1"
                        >Kategori</label
                    >
                    <select
                        id="itinerary-category"
                        bind:value={formCategory}
                        class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary-400"
                    >
                        <option value="flight">?? Penerbangan</option>
                        <option value="hotel">?? Hotel</option>
                        <option value="transport">?? Transportasi</option>
                        <option value="activity">?? Aktivitas</option>
                        <option value="other">?? Lainnya</option>
                    </select>
                </div>

                <!-- Date -->
                <div>
                    <label
                        for="itinerary-date"
                        class="block text-xs font-medium text-slate-600 mb-1"
                        >Tanggal *</label
                    >
                    <input
                        id="itinerary-date"
                        type="date"
                        bind:value={formDate}
                        class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary-400"
                    />
                </div>

                <!-- Time -->
                <div class="grid grid-cols-2 gap-3">
                    <div>
                        <label
                            for="itinerary-time-start"
                            class="block text-xs font-medium text-slate-600 mb-1"
                            >Jam Mulai</label
                        >
                        <input
                            id="itinerary-time-start"
                            type="time"
                            bind:value={formTimeStart}
                            class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary-400"
                        />
                    </div>
                    <div>
                        <label
                            for="itinerary-time-end"
                            class="block text-xs font-medium text-slate-600 mb-1"
                            >Jam Selesai</label
                        >
                        <input
                            id="itinerary-time-end"
                            type="time"
                            bind:value={formTimeEnd}
                            class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary-400"
                        />
                    </div>
                </div>

                <!-- Activity -->
                <div>
                    <label
                        for="itinerary-activity"
                        class="block text-xs font-medium text-slate-600 mb-1"
                        >Aktivitas *</label
                    >
                    <input
                        id="itinerary-activity"
                        type="text"
                        bind:value={formActivity}
                        placeholder="e.g. Check-in Hotel Al Safwah"
                        class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary-400"
                    />
                </div>

                <!-- Location -->
                <div>
                    <label
                        for="itinerary-location"
                        class="block text-xs font-medium text-slate-600 mb-1"
                        >Lokasi</label
                    >
                    <input
                        id="itinerary-location"
                        type="text"
                        bind:value={formLocation}
                        placeholder="e.g. Makkah"
                        class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary-400"
                    />
                </div>

                <!-- Notes -->
                <div>
                    <label
                        for="itinerary-notes"
                        class="block text-xs font-medium text-slate-600 mb-1"
                        >Catatan</label
                    >
                    <textarea
                        id="itinerary-notes"
                        bind:value={formNotes}
                        rows="2"
                        placeholder="Catatan tambahan..."
                        class="w-full px-3 py-2 border border-slate-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-primary-400 resize-none"
                    ></textarea>
                </div>
            </div>

            <div class="flex gap-2 mt-4">
                <button
                    type="button"
                    onclick={() => (showForm = false)}
                    class="flex-1 px-4 py-2.5 border border-slate-200 text-slate-600 rounded-xl text-sm font-medium hover:bg-slate-50 transition-colors"
                >
                    Batal
                </button>
                <button
                    type="button"
                    onclick={handleSubmit}
                    disabled={isSaving ||
                        !formActivity.trim() ||
                        !formDate.trim()}
                    class="flex-1 px-4 py-2.5 bg-primary-600 hover:bg-primary-700 text-white rounded-xl text-sm font-medium transition-colors disabled:opacity-50 flex items-center justify-center gap-1.5"
                >
                    {#if isSaving}
                        <Loader2 class="w-4 h-4 animate-spin" />
                    {:else}
                        <Save class="w-4 h-4" />
                    {/if}
                    {editingItem ? "Simpan" : "Tambah"}
                </button>
            </div>
        </div>
    </div>
{/if}
