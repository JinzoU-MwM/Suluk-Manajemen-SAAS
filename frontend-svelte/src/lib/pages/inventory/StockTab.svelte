<!--
  StockTab.svelte — Stok (inventory level) view for InventoryPage.
  Provides: summary stat cards, low-stock alert, search + item table,
  add-item modal, restock/adjust modals, movement history modal,
  and "Kit per Paket" configuration section.
-->
<script>
  import { onMount } from "svelte";
  import {
    Plus,
    Search,
    AlertTriangle,
    Package,
    TrendingDown,
    RefreshCw,
    History,
    Settings,
    Loader2,
  } from "lucide-svelte";
  import StatCard from "../../components/StatCard.svelte";
  import EmptyState from "../../components/EmptyState.svelte";
  import SlideDrawer from "../../components/SlideDrawer.svelte";
  import { ApiService } from "../../services/api.js";
  import { showToast } from "../../services/toast.svelte.js";

  let { groups = [] } = $props();

  // ── Item list state ────────────────────────────────────────────────────────
  let items = $state([]);
  let loading = $state(true);
  let q = $state("");

  let lowCount = $derived(
    items.filter((i) => i.stock <= 0 || i.stock < i.min_stock).length,
  );
  let filtered = $derived(
    q.trim()
      ? items.filter((i) =>
          i.name.toLowerCase().includes(q.trim().toLowerCase()),
        )
      : items,
  );

  onMount(load);

  async function load() {
    loading = true;
    try {
      const res = await ApiService.listStockItems();
      items = res?.items ?? (Array.isArray(res) ? res : []);
    } catch (e) {
      showToast(e.message || "Gagal memuat stok", "error");
    } finally {
      loading = false;
    }
  }

  // ── Add item modal ─────────────────────────────────────────────────────────
  let addOpen = $state(false);
  let addForm = $state({
    name: "",
    category: "",
    unit: "",
    stock: 0,
    min_stock: 0,
  });
  let addBusy = $state(false);

  function openAdd() {
    addForm = { name: "", category: "", unit: "", stock: 0, min_stock: 0 };
    addOpen = true;
  }

  async function submitAdd() {
    if (!addForm.name.trim()) {
      showToast("Nama item wajib diisi", "warning");
      return;
    }
    addBusy = true;
    try {
      await ApiService.createStockItem({
        name: addForm.name.trim(),
        category: addForm.category.trim(),
        unit: addForm.unit.trim(),
        stock: Number(addForm.stock),
        min_stock: Number(addForm.min_stock),
      });
      showToast("Item berhasil ditambahkan", "success");
      addOpen = false;
      await load();
    } catch (e) {
      showToast(e.message || "Gagal menambah item", "error");
    } finally {
      addBusy = false;
    }
  }

  // ── Restock modal ──────────────────────────────────────────────────────────
  let restockOpen = $state(false);
  let restockTarget = $state(null);
  let restockQty = $state(0);
  let restockNote = $state("");
  let restockBusy = $state(false);

  function openRestock(item) {
    restockTarget = item;
    restockQty = 0;
    restockNote = "";
    restockOpen = true;
  }

  async function submitRestock() {
    if (!restockTarget || restockQty <= 0) {
      showToast("Jumlah harus lebih dari 0", "warning");
      return;
    }
    restockBusy = true;
    try {
      await ApiService.restockItem(restockTarget.id, {
        qty: Number(restockQty),
        note: restockNote.trim(),
      });
      showToast(`Stok ${restockTarget.name} berhasil ditambah`, "success");
      restockOpen = false;
      await load();
    } catch (e) {
      showToast(e.message || "Gagal menambah stok", "error");
    } finally {
      restockBusy = false;
    }
  }

  // ── Adjust modal ───────────────────────────────────────────────────────────
  let adjustOpen = $state(false);
  let adjustTarget = $state(null);
  let adjustQty = $state(0);
  let adjustNote = $state("");
  let adjustBusy = $state(false);

  function openAdjust(item) {
    adjustTarget = item;
    adjustQty = item.stock;
    adjustNote = "";
    adjustOpen = true;
  }

  async function submitAdjust() {
    if (!adjustTarget) return;
    adjustBusy = true;
    try {
      await ApiService.adjustItem(adjustTarget.id, {
        stock: Number(adjustQty),
        note: adjustNote.trim(),
      });
      showToast(`Stok ${adjustTarget.name} berhasil disesuaikan`, "success");
      adjustOpen = false;
      await load();
    } catch (e) {
      showToast(e.message || "Gagal menyesuaikan stok", "error");
    } finally {
      adjustBusy = false;
    }
  }

  // ── History drawer ─────────────────────────────────────────────────────────
  let histOpen = $state(false);
  let histTarget = $state(null);
  let histMovements = $state([]);
  let histLoading = $state(false);

  async function openHistory(item) {
    histTarget = item;
    histMovements = [];
    histOpen = true;
    histLoading = true;
    try {
      const res = await ApiService.getItemMovements(item.id);
      histMovements = res?.movements ?? (Array.isArray(res) ? res : []);
    } catch (e) {
      showToast(e.message || "Gagal memuat riwayat", "error");
    } finally {
      histLoading = false;
    }
  }

  // ── Kit per Paket ──────────────────────────────────────────────────────────
  let packages = $state([]);
  let pkgsLoading = $state(false);
  let selectedPkgId = $state("");
  let kitLines = $state([]); // { item_id, item_name, qty_per_jamaah }
  let kitLoading = $state(false);
  let kitBusy = $state(false);

  onMount(loadPackages);

  async function loadPackages() {
    pkgsLoading = true;
    try {
      const res = await ApiService.listPackages();
      packages = res?.packages ?? res?.data ?? (Array.isArray(res) ? res : []);
    } catch {
      // Fall back to deriving distinct packages from groups prop
      packages = [];
    } finally {
      pkgsLoading = false;
    }
  }

  async function loadKit(pkgId) {
    if (!pkgId) {
      kitLines = [];
      return;
    }
    kitLoading = true;
    try {
      const res = await ApiService.getPackageKit(pkgId);
      kitLines = res?.items ?? (Array.isArray(res) ? res : []);
    } catch {
      kitLines = [];
    } finally {
      kitLoading = false;
    }
  }

  function onPkgChange(pkgId) {
    selectedPkgId = pkgId;
    loadKit(pkgId);
  }

  function addKitLine() {
    kitLines = [...kitLines, { item_id: "", item_name: "", qty_per_jamaah: 1 }];
  }

  function removeKitLine(idx) {
    kitLines = kitLines.filter((_, i) => i !== idx);
  }

  async function saveKit() {
    if (!selectedPkgId) return;
    kitBusy = true;
    try {
      await ApiService.setPackageKit(
        selectedPkgId,
        kitLines.map((l) => ({
          item_id: l.item_id,
          qty_per_jamaah: Number(l.qty_per_jamaah),
        })),
      );
      showToast("Kit per paket berhasil disimpan", "success");
    } catch (e) {
      showToast(e.message || "Gagal menyimpan kit", "error");
    } finally {
      kitBusy = false;
    }
  }

  // ── Helpers ────────────────────────────────────────────────────────────────
  function isLow(i) {
    return i.stock <= 0 || i.stock < i.min_stock;
  }
</script>

<div class="flex flex-col gap-5">
  <!-- Low-stock alert banner -->
  {#if lowCount > 0}
    <div
      class="flex items-center gap-2 rounded-xl bg-amber-50 p-3 text-[13.5px] text-amber-900 ring-1 ring-amber-200"
    >
      <AlertTriangle class="h-[18px] w-[18px] flex-shrink-0 text-amber-600" />
      <span
        ><strong>{lowCount}</strong> item stok menipis atau habis — periksa dan tambah
        stok.</span
      >
    </div>
  {/if}

  <!-- Summary stat cards -->
  <div class="grid grid-cols-2 gap-4 lg:grid-cols-3">
    <StatCard
      icon={Package}
      label="Jenis Item"
      value={String(items.length)}
      accent="var(--c-primary)"
    />
    <StatCard
      icon={TrendingDown}
      label="Stok Menipis"
      value={String(lowCount)}
      accent={lowCount > 0 ? "var(--c-danger)" : "var(--c-muted)"}
    />
  </div>

  <!-- Search + add-item bar -->
  <div class="flex flex-wrap items-center gap-3">
    <div class="relative flex-1" style="min-width:220px;max-width:400px">
      <Search
        class="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400"
      />
      <input
        bind:value={q}
        placeholder="Cari item…"
        class="w-full rounded-xl border border-slate-200 bg-white py-2.5 pl-9 pr-3 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
      />
    </div>
    <button
      type="button"
      onclick={openAdd}
      class="inline-flex items-center gap-1.5 rounded-xl px-4 py-2.5 text-[13px] font-bold text-white transition-opacity hover:opacity-90"
      style="background:var(--c-primary)"
    >
      <Plus class="h-4 w-4" />
      Tambah Item
    </button>
  </div>

  <!-- Item table -->
  {#if loading}
    <div class="h-32 animate-pulse rounded-xl bg-slate-100"></div>
  {:else if filtered.length === 0}
    <EmptyState
      icon={Package}
      title="Belum ada item"
      text={q.trim()
        ? "Tidak ada item yang cocok dengan pencarian."
        : "Tambahkan stok perlengkapan untuk mulai memantau."}
    />
  {:else}
    <div
      class="overflow-hidden rounded-xl bg-white shadow-sm ring-1 ring-slate-200/60"
    >
      <div class="overflow-x-auto">
        <table class="w-full text-sm">
          <thead
            class="border-b border-slate-100 text-left text-[12px] uppercase tracking-wide text-slate-500"
          >
            <tr>
              <th class="p-3 pl-4">Item</th>
              <th class="p-3">Kategori</th>
              <th class="p-3 text-right">Stok</th>
              <th class="p-3 text-right">Min</th>
              <th class="p-3 text-right">Satuan</th>
              <th class="p-3 pr-4"></th>
            </tr>
          </thead>
          <tbody>
            {#each filtered as it (it.id)}
              <tr
                class="border-b border-slate-50 transition-colors hover:bg-slate-50/60"
              >
                <td class="p-3 pl-4 font-medium text-slate-800">{it.name}</td>
                <td class="p-3 text-slate-500">{it.category || "—"}</td>
                <td
                  class="p-3 text-right font-semibold {isLow(it)
                    ? 'text-red-600'
                    : 'text-slate-800'}"
                >
                  {it.stock}
                </td>
                <td class="p-3 text-right text-slate-400">{it.min_stock}</td>
                <td class="p-3 text-right text-slate-400">{it.unit || "—"}</td>
                <td class="p-3 pr-4 text-right">
                  <div class="inline-flex items-center gap-1">
                    <button
                      type="button"
                      onclick={() => openRestock(it)}
                      title="Tambah Stok"
                      class="inline-flex items-center gap-1 rounded-lg px-2.5 py-1.5 text-[12px] font-semibold text-green-700 transition-colors hover:bg-green-50"
                    >
                      <Plus class="h-3.5 w-3.5" />
                      Tambah
                    </button>
                    <button
                      type="button"
                      onclick={() => openAdjust(it)}
                      title="Sesuaikan Stok"
                      class="inline-flex items-center gap-1 rounded-lg px-2.5 py-1.5 text-[12px] font-semibold text-blue-700 transition-colors hover:bg-blue-50"
                    >
                      <Settings class="h-3.5 w-3.5" />
                      Sesuaikan
                    </button>
                    <button
                      type="button"
                      onclick={() => openHistory(it)}
                      title="Riwayat"
                      class="inline-flex items-center gap-1 rounded-lg px-2.5 py-1.5 text-[12px] font-semibold text-slate-600 transition-colors hover:bg-slate-100"
                    >
                      <History class="h-3.5 w-3.5" />
                      Riwayat
                    </button>
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}

  <!-- Kit per Paket section -->
  <div class="rounded-xl bg-white p-4 shadow-sm ring-1 ring-slate-200/60">
    <div class="mb-4 flex items-center justify-between gap-3">
      <div>
        <h2 class="font-serif text-[15px] font-semibold text-slate-700">
          Kit per Paket
        </h2>
        <p class="mt-0.5 text-[12.5px] text-slate-400">
          Tentukan item dan jumlah per jamaah untuk setiap paket.
        </p>
      </div>
      {#if selectedPkgId}
        <button
          type="button"
          onclick={saveKit}
          disabled={kitBusy}
          class="inline-flex items-center gap-1.5 rounded-xl px-4 py-2 text-[13px] font-bold text-white disabled:opacity-60"
          style="background:var(--c-primary)"
        >
          {#if kitBusy}
            <Loader2 class="h-3.5 w-3.5 animate-spin" />
          {/if}
          Simpan
        </button>
      {/if}
    </div>

    <!-- Package selector -->
    <div class="mb-4">
      <label
        for="kit-pkg-select"
        class="mb-1.5 block text-[12px] font-semibold uppercase tracking-wide text-slate-400"
        >Pilih Paket</label
      >
      {#if pkgsLoading}
        <div class="h-10 animate-pulse rounded-xl bg-slate-100"></div>
      {:else}
        <select
          id="kit-pkg-select"
          value={selectedPkgId}
          onchange={(e) =>
            onPkgChange(/** @type {HTMLSelectElement} */ (e.target).value)}
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
          style="max-width:360px"
        >
          <option value="">— Pilih paket —</option>
          {#each packages as pkg}
            <option value={pkg.id}>{pkg.name}</option>
          {/each}
        </select>
      {/if}
    </div>

    <!-- Kit lines -->
    {#if selectedPkgId}
      {#if kitLoading}
        <div class="h-16 animate-pulse rounded-xl bg-slate-100"></div>
      {:else}
        <div class="overflow-hidden rounded-xl ring-1 ring-slate-200/60">
          <table class="w-full text-sm">
            <thead
              class="border-b border-slate-100 text-left text-[12px] uppercase tracking-wide text-slate-500"
            >
              <tr>
                <th class="p-3 pl-4">Item</th>
                <th class="p-3 text-right" style="width:140px"
                  >Qty / Jamaah</th
                >
                <th class="p-3 pr-4" style="width:48px"></th>
              </tr>
            </thead>
            <tbody>
              {#each kitLines as line, idx}
                <tr class="border-b border-slate-50">
                  <td class="p-2 pl-4">
                    <select
                      value={line.item_id}
                      onchange={(e) => {
                        const id =
                          /** @type {HTMLSelectElement} */ (e.target).value;
                        const found = items.find((i) => i.id === id);
                        kitLines = kitLines.map((l, i) =>
                          i === idx
                            ? {
                                ...l,
                                item_id: id,
                                item_name: found?.name ?? "",
                              }
                            : l,
                        );
                      }}
                      class="w-full rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-[13px] text-slate-800 outline-none focus:border-primary-400"
                    >
                      <option value="">— Pilih item —</option>
                      {#each items as it}
                        <option value={it.id}>{it.name}</option>
                      {/each}
                    </select>
                  </td>
                  <td class="p-2 text-right">
                    <input
                      type="number"
                      min="0"
                      value={line.qty_per_jamaah}
                      oninput={(e) => {
                        kitLines = kitLines.map((l, i) =>
                          i === idx
                            ? {
                                ...l,
                                qty_per_jamaah: Number(
                                  /** @type {HTMLInputElement} */ (e.target)
                                    .value,
                                ),
                              }
                            : l,
                        );
                      }}
                      class="w-full rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-right text-[13px] text-slate-800 outline-none focus:border-primary-400"
                    />
                  </td>
                  <td class="p-2 pr-4 text-right">
                    <button
                      type="button"
                      onclick={() => removeKitLine(idx)}
                      class="rounded p-1 text-slate-400 transition-colors hover:bg-red-50 hover:text-red-500"
                      aria-label="Hapus baris"
                    >
                      ×
                    </button>
                  </td>
                </tr>
              {/each}
              {#if kitLines.length === 0}
                <tr>
                  <td
                    colspan="3"
                    class="py-6 text-center text-[13px] text-slate-400"
                    >Belum ada item di kit ini.</td
                  >
                </tr>
              {/if}
            </tbody>
          </table>
        </div>
        <button
          type="button"
          onclick={addKitLine}
          class="mt-3 inline-flex items-center gap-1.5 text-[13px] font-semibold"
          style="color:var(--c-primary)"
        >
          <Plus class="h-4 w-4" />
          Tambah Baris
        </button>
      {/if}
    {/if}
  </div>
</div>

<!-- Add Item Modal -->
<SlideDrawer open={addOpen} title="Tambah Item Stok" onClose={() => (addOpen = false)}>
  <div class="flex flex-col gap-4 p-6">
    <div>
      <label for="add-name" class="mb-1 block text-[13px] font-medium text-slate-600">Nama Item *</label>
      <input
        id="add-name"
        bind:value={addForm.name}
        placeholder="cth. Koper 20 inch"
        class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
      />
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div>
        <label for="add-category" class="mb-1 block text-[13px] font-medium text-slate-600">Kategori</label>
        <input
          id="add-category"
          bind:value={addForm.category}
          placeholder="cth. Perlengkapan"
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>
      <div>
        <label for="add-unit" class="mb-1 block text-[13px] font-medium text-slate-600">Satuan</label>
        <input
          id="add-unit"
          bind:value={addForm.unit}
          placeholder="cth. pcs"
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>
    </div>
    <div class="grid grid-cols-2 gap-4">
      <div>
        <label for="add-stock" class="mb-1 block text-[13px] font-medium text-slate-600">Stok Awal</label>
        <input
          id="add-stock"
          type="number"
          min="0"
          bind:value={addForm.stock}
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>
      <div>
        <label for="add-min-stock" class="mb-1 block text-[13px] font-medium text-slate-600">Stok Minimum</label>
        <input
          id="add-min-stock"
          type="number"
          min="0"
          bind:value={addForm.min_stock}
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>
    </div>
    <div class="flex justify-end gap-3 pt-2">
      <button
        type="button"
        onclick={() => (addOpen = false)}
        class="rounded-xl border border-slate-200 px-4 py-2 text-[13px] font-semibold text-slate-600 transition-colors hover:bg-slate-50"
      >
        Batal
      </button>
      <button
        type="button"
        onclick={submitAdd}
        disabled={addBusy}
        class="inline-flex items-center gap-1.5 rounded-xl px-5 py-2 text-[13px] font-bold text-white disabled:opacity-60"
        style="background:var(--c-primary)"
      >
        {#if addBusy}<Loader2 class="h-3.5 w-3.5 animate-spin" />{/if}
        Simpan
      </button>
    </div>
  </div>
</SlideDrawer>

<!-- Restock Modal -->
<SlideDrawer
  open={restockOpen}
  title="Tambah Stok — {restockTarget?.name ?? ''}"
  width="420px"
  onClose={() => (restockOpen = false)}
>
  <div class="flex flex-col gap-4 p-6">
    <div>
      <label for="restock-qty" class="mb-1 block text-[13px] font-medium text-slate-600">Jumlah Ditambahkan *</label>
      <input
        id="restock-qty"
        type="number"
        min="1"
        bind:value={restockQty}
        class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
      />
    </div>
    <div>
      <label for="restock-note" class="mb-1 block text-[13px] font-medium text-slate-600">Catatan</label>
      <input
        id="restock-note"
        bind:value={restockNote}
        placeholder="Opsional"
        class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
      />
    </div>
    <div class="flex justify-end gap-3 pt-2">
      <button
        type="button"
        onclick={() => (restockOpen = false)}
        class="rounded-xl border border-slate-200 px-4 py-2 text-[13px] font-semibold text-slate-600 transition-colors hover:bg-slate-50"
      >
        Batal
      </button>
      <button
        type="button"
        onclick={submitRestock}
        disabled={restockBusy}
        class="inline-flex items-center gap-1.5 rounded-xl px-5 py-2 text-[13px] font-bold text-white disabled:opacity-60"
        style="background:var(--c-primary)"
      >
        {#if restockBusy}<Loader2 class="h-3.5 w-3.5 animate-spin" />{/if}
        Tambah Stok
      </button>
    </div>
  </div>
</SlideDrawer>

<!-- Adjust Modal -->
<SlideDrawer
  open={adjustOpen}
  title="Sesuaikan Stok — {adjustTarget?.name ?? ''}"
  width="420px"
  onClose={() => (adjustOpen = false)}
>
  <div class="flex flex-col gap-4 p-6">
    <p class="text-[13px] text-slate-500">
      Set stok aktual ke jumlah di bawah (stok saat ini: <strong
        >{adjustTarget?.stock ?? 0} {adjustTarget?.unit ?? ""}</strong
      >).
    </p>
    <div>
      <label for="adjust-qty" class="mb-1 block text-[13px] font-medium text-slate-600">Stok Aktual *</label>
      <input
        id="adjust-qty"
        type="number"
        min="0"
        bind:value={adjustQty}
        class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
      />
    </div>
    <div>
      <label for="adjust-note" class="mb-1 block text-[13px] font-medium text-slate-600">Alasan Penyesuaian</label>
      <input
        id="adjust-note"
        bind:value={adjustNote}
        placeholder="cth. Rusak / kehilangan"
        class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
      />
    </div>
    <div class="flex justify-end gap-3 pt-2">
      <button
        type="button"
        onclick={() => (adjustOpen = false)}
        class="rounded-xl border border-slate-200 px-4 py-2 text-[13px] font-semibold text-slate-600 transition-colors hover:bg-slate-50"
      >
        Batal
      </button>
      <button
        type="button"
        onclick={submitAdjust}
        disabled={adjustBusy}
        class="inline-flex items-center gap-1.5 rounded-xl px-5 py-2 text-[13px] font-bold text-white disabled:opacity-60"
        style="background:var(--c-primary)"
      >
        {#if adjustBusy}<Loader2 class="h-3.5 w-3.5 animate-spin" />{/if}
        Simpan
      </button>
    </div>
  </div>
</SlideDrawer>

<!-- History Modal -->
<SlideDrawer
  open={histOpen}
  title="Riwayat — {histTarget?.name ?? ''}"
  width="560px"
  onClose={() => (histOpen = false)}
>
  <div class="p-6">
    {#if histLoading}
      <div class="space-y-2">
        {#each [1, 2, 3] as _}
          <div class="h-10 animate-pulse rounded-xl bg-slate-100"></div>
        {/each}
      </div>
    {:else if histMovements.length === 0}
      <p class="py-8 text-center text-[13px] text-slate-400">
        Belum ada riwayat pergerakan stok.
      </p>
    {:else}
      <div class="overflow-hidden rounded-xl ring-1 ring-slate-200/60">
        <table class="w-full text-sm">
          <thead
            class="border-b border-slate-100 text-left text-[12px] uppercase tracking-wide text-slate-500"
          >
            <tr>
              <th class="p-3 pl-4">Tanggal</th>
              <th class="p-3">Alasan</th>
              <th class="p-3 text-right">Delta</th>
              <th class="p-3 pr-4">Catatan</th>
            </tr>
          </thead>
          <tbody>
            {#each histMovements as mv}
              <tr class="border-b border-slate-50">
                <td class="p-3 pl-4 text-slate-500"
                  >{new Date(mv.created_at ?? mv.date).toLocaleDateString(
                    "id-ID",
                  )}</td
                >
                <td class="p-3 text-slate-700">{mv.reason ?? mv.type ?? "—"}</td>
                <td
                  class="p-3 text-right font-semibold {(mv.delta ??
                    mv.qty) >= 0
                    ? 'text-green-600'
                    : 'text-red-600'}"
                >
                  {(mv.delta ?? mv.qty) >= 0 ? "+" : ""}{mv.delta ?? mv.qty}
                </td>
                <td class="p-3 pr-4 text-slate-400">{mv.note || "—"}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </div>
</SlideDrawer>
