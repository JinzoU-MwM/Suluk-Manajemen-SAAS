<!--
  ItineraryPage.svelte - Trip schedule/itinerary manager.
  Suluk design: PageHeader + group selector + day-by-day vertical timeline
  (green circle day markers + connector line, gold date labels, category-coloured
  activity items with time/location/notes). All data wiring preserved.
-->
<script>
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
    import EmptyState from "../components/EmptyState.svelte";
    import Card from "../components/ui/Card.svelte";
    import Button from "../components/ui/Button.svelte";

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
            fg: "--c-info",
            bg: "--c-info-soft",
        },
        hotel: {
            icon: Hotel,
            label: "Hotel",
            fg: "--c-primary",
            bg: "--c-primary-soft",
        },
        transport: {
            icon: Bus,
            label: "Transportasi",
            fg: "--c-accent",
            bg: "--c-accent-soft",
        },
        activity: {
            icon: MapPin,
            label: "Aktivitas",
            fg: "--c-success",
            bg: "--c-success-soft",
        },
        other: {
            icon: MoreHorizontal,
            label: "Lainnya",
            fg: "--c-muted",
            bg: "--c-bg-2",
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

<div class="itinerary-page">
    <!-- Header -->
    <PageHeader
        kicker="Operasional"
        title="Itinerary Perjalanan"
        subtitle="Kelola jadwal perjalanan per grup keberangkatan."
    >
        {#snippet actions()}
            {#if selectedGroupId}
                <Button variant="primary" icon={Plus} onclick={openAddForm}>
                    Tambah
                </Button>
            {/if}
        {/snippet}
    </PageHeader>

    <!-- Error -->
    {#if error}
        <div class="itinerary-alert">
            <AlertCircle size={16} style="flex-shrink:0" />
            <span>{error}</span>
            <button
                type="button"
                onclick={() => (error = "")}
                class="itinerary-alert-close"
                aria-label="Tutup"
            >
                <X size={16} />
            </button>
        </div>
    {/if}

    <!-- Group Selector -->
    <Card style="margin-bottom:var(--gap)">
        <label for="itinerary-group" class="itinerary-field-label">
            Pilih Grup
        </label>
        <select
            id="itinerary-group"
            class="itinerary-select"
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
    </Card>

    {#if !selectedGroupId}
        <Card>
            <EmptyState
                icon={CalendarDays}
                title="Pilih grup keberangkatan"
                text="Pilih grup terlebih dahulu untuk melihat jadwal perjalanan."
            />
        </Card>
    {:else if isLoading}
        <Card>
            <div class="itinerary-loading">
                <Loader2 size={22} class="spin" />
                <span>Memuat jadwal...</span>
            </div>
        </Card>
    {:else if items.length === 0}
        <Card>
            <EmptyState
                icon={CalendarDays}
                title="Belum ada jadwal"
                text={'Klik "Tambah" untuk menambahkan item jadwal pertama.'}
            />
        </Card>
    {:else}
        <!-- Timeline -->
        <Card>
            <div class="timeline">
                {#each groupedItems() as [date, dateItems], dayIdx}
                    {@const isLast = dayIdx === groupedItems().length - 1}
                    <div class="timeline-day" class:is-last={isLast}>
                        <!-- Connector line -->
                        {#if !isLast}
                            <div class="timeline-connector"></div>
                        {/if}

                        <!-- Day marker -->
                        <div class="timeline-marker">
                            <CalendarDays size={20} />
                        </div>

                        <!-- Day content -->
                        <div class="timeline-content">
                            <p class="timeline-date">{formatDate(date)}</p>

                            <!-- Items -->
                            <div class="timeline-items">
                                {#each dateItems as item}
                                    {@const cat =
                                        categoryConfig[item.category] ||
                                        categoryConfig.other}
                                    <div class="timeline-item">
                                        <div
                                            class="timeline-item-icon"
                                            style="background:var({cat.bg});color:var({cat.fg})"
                                        >
                                            <cat.icon size={16} />
                                        </div>
                                        <div class="timeline-item-body">
                                            <p class="timeline-item-title">
                                                {item.activity}
                                            </p>
                                            {#if item.location}
                                                <p class="timeline-item-meta">
                                                    <MapPin size={12} />
                                                    {item.location}
                                                </p>
                                            {/if}
                                            {#if item.time_start}
                                                <p class="timeline-item-meta">
                                                    <Clock size={12} />
                                                    {item.time_start}{item.time_end
                                                        ? ` - ${item.time_end}`
                                                        : ""}
                                                </p>
                                            {/if}
                                            {#if item.notes}
                                                <p class="timeline-item-notes">
                                                    {item.notes}
                                                </p>
                                            {/if}
                                        </div>
                                        <div class="timeline-item-actions">
                                            <button
                                                type="button"
                                                onclick={() =>
                                                    openEditForm(item)}
                                                class="timeline-action edit"
                                                aria-label="Edit jadwal"
                                            >
                                                <Edit3 size={14} />
                                            </button>
                                            <button
                                                type="button"
                                                onclick={() =>
                                                    deleteItem(item.id)}
                                                class="timeline-action delete"
                                                aria-label="Hapus jadwal"
                                            >
                                                <Trash2 size={14} />
                                            </button>
                                        </div>
                                    </div>
                                {/each}
                            </div>
                        </div>
                    </div>
                {/each}
            </div>
        </Card>
    {/if}
</div>

<!-- Add/Edit Modal -->
{#if showForm}
    <div class="modal-overlay" role="dialog" aria-modal="true">
        <div class="modal-panel">
            <div class="modal-head">
                <h3 class="modal-title">
                    {editingItem ? "Edit Jadwal" : "Tambah Jadwal"}
                </h3>
                <button
                    type="button"
                    onclick={() => (showForm = false)}
                    class="modal-close"
                    aria-label="Tutup"
                >
                    <X size={16} />
                </button>
            </div>

            <div class="modal-body">
                <!-- Category -->
                <div>
                    <label for="itinerary-category" class="modal-label">
                        Kategori
                    </label>
                    <select
                        id="itinerary-category"
                        bind:value={formCategory}
                        class="modal-input"
                    >
                        <option value="flight">Penerbangan</option>
                        <option value="hotel">Hotel</option>
                        <option value="transport">Transportasi</option>
                        <option value="activity">Aktivitas</option>
                        <option value="other">Lainnya</option>
                    </select>
                </div>

                <!-- Date -->
                <div>
                    <label for="itinerary-date" class="modal-label">
                        Tanggal *
                    </label>
                    <input
                        id="itinerary-date"
                        type="date"
                        bind:value={formDate}
                        class="modal-input"
                    />
                </div>

                <!-- Time -->
                <div class="modal-grid-2">
                    <div>
                        <label
                            for="itinerary-time-start"
                            class="modal-label"
                        >
                            Jam Mulai
                        </label>
                        <input
                            id="itinerary-time-start"
                            type="time"
                            bind:value={formTimeStart}
                            class="modal-input"
                        />
                    </div>
                    <div>
                        <label for="itinerary-time-end" class="modal-label">
                            Jam Selesai
                        </label>
                        <input
                            id="itinerary-time-end"
                            type="time"
                            bind:value={formTimeEnd}
                            class="modal-input"
                        />
                    </div>
                </div>

                <!-- Activity -->
                <div>
                    <label for="itinerary-activity" class="modal-label">
                        Aktivitas *
                    </label>
                    <input
                        id="itinerary-activity"
                        type="text"
                        bind:value={formActivity}
                        placeholder="e.g. Check-in Hotel Al Safwah"
                        class="modal-input"
                    />
                </div>

                <!-- Location -->
                <div>
                    <label for="itinerary-location" class="modal-label">
                        Lokasi
                    </label>
                    <input
                        id="itinerary-location"
                        type="text"
                        bind:value={formLocation}
                        placeholder="e.g. Makkah"
                        class="modal-input"
                    />
                </div>

                <!-- Notes -->
                <div>
                    <label for="itinerary-notes" class="modal-label">
                        Catatan
                    </label>
                    <textarea
                        id="itinerary-notes"
                        bind:value={formNotes}
                        rows="2"
                        placeholder="Catatan tambahan..."
                        class="modal-input modal-textarea"
                    ></textarea>
                </div>
            </div>

            <div class="modal-actions">
                <Button
                    variant="ghost"
                    full
                    onclick={() => (showForm = false)}
                >
                    Batal
                </Button>
                <Button
                    variant="primary"
                    full
                    icon={isSaving ? Loader2 : Save}
                    disabled={isSaving ||
                        !formActivity.trim() ||
                        !formDate.trim()}
                    onclick={handleSubmit}
                >
                    {editingItem ? "Simpan" : "Tambah"}
                </Button>
            </div>
        </div>
    </div>
{/if}

<style>
    .itinerary-page {
        min-height: 100vh;
        background: var(--c-bg, #f6f8f7);
        padding: 16px;
    }
    @media (min-width: 1024px) {
        .itinerary-page {
            padding: 32px;
        }
    }

    /* Error alert */
    .itinerary-alert {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-bottom: var(--gap);
        padding: 12px 16px;
        border: 1px solid var(--c-danger);
        background: var(--c-danger-soft);
        color: var(--c-danger);
        border-radius: var(--radius-lg);
        font-size: 14px;
    }
    .itinerary-alert-close {
        margin-left: auto;
        display: inline-flex;
        color: inherit;
    }

    /* Group selector */
    .itinerary-field-label {
        display: block;
        margin-bottom: 8px;
        font-size: 12px;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.08em;
        color: var(--c-primary);
    }
    .itinerary-select {
        width: 100%;
        padding: 12px 16px;
        border: 1px solid var(--c-line);
        background: var(--c-bg-2, #f2f6f4);
        border-radius: var(--radius-md);
        font-size: 14px;
        font-weight: 500;
        color: var(--c-ink-soft);
        outline: none;
        transition: border-color 0.15s, background 0.15s;
    }
    .itinerary-select:focus {
        border-color: var(--c-primary);
        background: var(--c-surface);
    }

    /* Loading */
    .itinerary-loading {
        display: flex;
        align-items: center;
        justify-content: center;
        gap: 8px;
        padding: 40px 0;
        color: var(--c-faint);
        font-size: 14px;
    }
    :global(.itinerary-loading .spin) {
        animation: itinerary-spin 1s linear infinite;
    }
    @keyframes itinerary-spin {
        to {
            transform: rotate(360deg);
        }
    }

    /* Timeline */
    .timeline {
        position: relative;
        padding-left: 8px;
    }
    .timeline-day {
        position: relative;
        display: flex;
        gap: 18px;
        padding-bottom: 24px;
    }
    .timeline-day.is-last {
        padding-bottom: 0;
    }
    .timeline-connector {
        position: absolute;
        left: 21px;
        top: 44px;
        bottom: 0;
        width: 2px;
        background: var(--c-line);
    }
    .timeline-marker {
        width: 44px;
        height: 44px;
        border-radius: 50%;
        background: var(--c-primary-soft);
        color: var(--c-primary-deep);
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
        z-index: 1;
        border: 3px solid var(--c-surface);
    }
    .timeline-content {
        flex: 1;
        min-width: 0;
        padding-top: 2px;
    }
    .timeline-date {
        font-size: 11.5px;
        font-weight: 700;
        color: var(--c-accent);
        letter-spacing: 0.05em;
        text-transform: uppercase;
    }
    .timeline-items {
        margin-top: 12px;
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    .timeline-item {
        display: flex;
        align-items: flex-start;
        gap: 12px;
        padding: 14px;
        border: 1px solid var(--c-line);
        border-radius: var(--radius-md);
        background: var(--c-surface);
        transition: box-shadow 0.15s;
    }
    .timeline-item:hover {
        box-shadow: var(--shadow-sm);
    }
    .timeline-item:hover .timeline-item-actions {
        opacity: 1;
    }
    .timeline-item-icon {
        width: 32px;
        height: 32px;
        flex-shrink: 0;
        display: flex;
        align-items: center;
        justify-content: center;
        border-radius: var(--radius-md);
    }
    .timeline-item-body {
        min-width: 0;
        flex: 1;
    }
    .timeline-item-title {
        font-size: 14px;
        font-weight: 700;
        color: var(--c-ink);
    }
    .timeline-item-meta {
        margin-top: 2px;
        display: flex;
        align-items: center;
        gap: 4px;
        font-size: 12px;
        color: var(--c-muted);
    }
    .timeline-item-notes {
        margin-top: 4px;
        font-size: 12px;
        font-style: italic;
        color: var(--c-faint);
    }
    .timeline-item-actions {
        display: flex;
        gap: 4px;
        opacity: 0;
        transition: opacity 0.15s;
    }
    .timeline-action {
        display: inline-flex;
        padding: 6px;
        border-radius: var(--radius-sm);
        color: var(--c-faint);
        transition: background 0.15s, color 0.15s;
    }
    .timeline-action.edit:hover {
        background: var(--c-primary-soft);
        color: var(--c-primary);
    }
    .timeline-action.delete:hover {
        background: var(--c-danger-soft);
        color: var(--c-danger);
    }
    @media (max-width: 640px) {
        .timeline-item-actions {
            opacity: 1;
        }
    }

    /* Modal */
    .modal-overlay {
        position: fixed;
        inset: 0;
        z-index: 50;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 16px;
        background: rgba(0, 0, 0, 0.4);
        backdrop-filter: blur(4px);
    }
    .modal-panel {
        width: 100%;
        max-width: 28rem;
        background: var(--c-surface);
        border-radius: var(--radius-lg);
        box-shadow: var(--shadow-lg, 0 20px 50px rgba(0, 0, 0, 0.25));
        padding: 20px;
    }
    .modal-head {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 16px;
    }
    .modal-title {
        font-size: 18px;
        font-weight: 700;
        color: var(--c-ink);
    }
    .modal-close {
        display: inline-flex;
        padding: 6px;
        border-radius: var(--radius-sm);
        color: var(--c-muted);
    }
    .modal-close:hover {
        background: var(--c-bg-2, #f2f6f4);
    }
    .modal-body {
        display: flex;
        flex-direction: column;
        gap: 12px;
    }
    .modal-grid-2 {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 12px;
    }
    .modal-label {
        display: block;
        margin-bottom: 4px;
        font-size: 12px;
        font-weight: 600;
        color: var(--c-ink-soft);
    }
    .modal-input {
        width: 100%;
        padding: 8px 12px;
        border: 1px solid var(--c-line);
        border-radius: var(--radius-md);
        font-size: 14px;
        color: var(--c-ink);
        background: var(--c-surface);
        outline: none;
        transition: border-color 0.15s, box-shadow 0.15s;
    }
    .modal-input:focus {
        border-color: var(--c-primary);
        box-shadow: 0 0 0 2px var(--c-primary-soft);
    }
    .modal-textarea {
        resize: none;
    }
    .modal-actions {
        display: flex;
        gap: 8px;
        margin-top: 16px;
    }
</style>
