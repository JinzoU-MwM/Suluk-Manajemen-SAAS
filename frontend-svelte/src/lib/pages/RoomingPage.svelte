<!--
  RoomingPage.svelte — Professional Room Allocation Dashboard
  
  Design: Compact card-based layout, efficient use of space
  Tone: Clean, data-dense, professional travel agency tool
-->
<script>
  import {
    Hotel,
    Bed,
    RefreshCw,
    Loader2,
    AlertTriangle,
    CheckCircle,
    Trash2,
    Sparkles,
    Users,
    TrendingUp,
    X,
    Plus,
    FileDown,
    MessageCircle,
  } from "lucide-svelte";
  import { ApiService } from "../services/api.js";
  import PageHeader from "../components/PageHeader.svelte";
  import StatCard from "../components/StatCard.svelte";
  import Avatar from "../components/Avatar.svelte";
  import WhatsAppBlast from "../components/WhatsAppBlast.svelte";

  let { isOpen = false, onClose, groups = [], isPro = false } = $props();

  // State
  let selectedGroupId = $state(null);
  let summary = $state(null);
  let rooms = $state([]);
  let isLoading = $state(false);
  let isAutoRooming = $state(false);
  let error = $state(null);
  let success = $state(null);

  // Drag state
  let draggedMember = $state(null);
  let dragSourceRoomId = $state(null);

  // Add room state
  let showAddRoom = $state(false);
  let newRoomNumber = $state("");
  let newRoomGender = $state("male");
  let isAddingRoom = $state(false);

  // WA Blast state
  let waBlastOpen = $state(false);
  let waBlastGroup = $derived(
    selectedGroupId
      ? groups.find((g) => g.id == selectedGroupId) || null
      : null,
  );

  function exportRoomingPDF() {
    const groupName =
      groups.find((g) => g.id == selectedGroupId)?.name || "Rooming List";
    const now = new Date().toLocaleDateString("id-ID", {
      day: "numeric",
      month: "long",
      year: "numeric",
    });

    let tableRows = "";
    rooms.forEach((room) => {
      const members = room.members || [];
      const genderLabel =
        room.gender_type === "male"
          ? "Laki-laki"
          : room.gender_type === "female"
            ? "Perempuan"
            : "Keluarga";
      if (members.length === 0) {
        tableRows += `<tr><td>${room.room_number}</td><td>${genderLabel}</td><td>${room.capacity || 4}</td><td>-</td><td>-</td></tr>`;
      } else {
        members.forEach((m, idx) => {
          tableRows += `<tr>
            ${idx === 0 ? `<td rowspan="${members.length}">${room.room_number}</td><td rowspan="${members.length}">${genderLabel}</td><td rowspan="${members.length}">${members.length}/${room.capacity || 4}</td>` : ""}
            <td>${m.nama || "-"}</td>
            <td>${m.no_paspor || "-"}</td>
          </tr>`;
        });
      }
    });

    const html = `<!DOCTYPE html><html><head><title>Rooming - ${groupName}</title>
      <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; padding: 2rem; color: #1e293b; }
        h1 { font-size: 1.25rem; margin: 0 0 0.25rem; }
        .subtitle { color: #64748b; font-size: 0.875rem; margin-bottom: 1.5rem; }
        table { width: 100%; border-collapse: collapse; font-size: 0.8125rem; }
        th { background: #f1f5f9; padding: 0.5rem 0.75rem; text-align: left; font-weight: 600; border: 1px solid #e2e8f0; }
        td { padding: 0.5rem 0.75rem; border: 1px solid #e2e8f0; }
        .summary { display: flex; gap: 2rem; margin-bottom: 1rem; font-size: 0.8125rem; color: #475569; }
        @media print { body { padding: 0.5rem; } }
      </style>
    </head><body>
      <h1>🏨 ${groupName} — Rooming List</h1>
      <p class="subtitle">${now}</p>
      <div class="summary">
        <span>Total Kamar: <strong>${summary?.total_rooms || 0}</strong></span>
        <span>Assigned: <strong>${summary?.assigned_count || 0}</strong></span>
        <span>Unassigned: <strong>${summary?.unassigned_count || 0}</strong></span>
      </div>
      <table><thead><tr><th>No. Kamar</th><th>Gender</th><th>Kapasitas</th><th>Nama Jamaah</th><th>No. Paspor</th></tr></thead>
      <tbody>${tableRows}</tbody></table>
    </body></html>`;

    const printWindow = window.open("", "_blank");
    printWindow.document.write(html);
    printWindow.document.close();
    printWindow.focus();
    setTimeout(() => {
      printWindow.print();
    }, 500);
  }

  // Derived
  let unassignedCount = $derived(summary?.unassigned_count || 0);
  let occupancyPct = $derived(
    summary && summary.total_members > 0
      ? Math.round(((summary.assigned_count || 0) / summary.total_members) * 100)
      : 0,
  );

  async function loadData() {
    if (!selectedGroupId) {
      summary = null;
      rooms = [];
      return;
    }

    isLoading = true;
    error = null;

    try {
      const [summaryData, roomsData] = await Promise.all([
        ApiService.getRoomingSummary(selectedGroupId),
        ApiService.getGroupRooms(selectedGroupId),
      ]);

      summary = summaryData;
      rooms = roomsData.rooms || [];
    } catch (e) {
      error = e.message;
    } finally {
      isLoading = false;
    }
  }

  async function runAutoRooming() {
    if (!selectedGroupId) return;

    isAutoRooming = true;
    error = null;
    success = null;

    try {
      const result = await ApiService.autoRooming(selectedGroupId, 4);
      success = result.summary;
      await loadData();
    } catch (e) {
      error = e.message;
    } finally {
      isAutoRooming = false;
    }
  }

  async function clearRooms() {
    if (!selectedGroupId) return;
    if (
      !confirm(
        "Reset semua kamar dan assignment? Tindakan ini tidak bisa dibatalkan.",
      )
    )
      return;

    isLoading = true;
    error = null;
    try {
      await ApiService.clearAutoRooming(selectedGroupId);
      await loadData();
    } catch (e) {
      error = e.message;
    } finally {
      isLoading = false;
    }
  }

  function handleDragStart(e, member, roomId) {
    draggedMember = member;
    dragSourceRoomId = roomId;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = "move";
      e.dataTransfer.setData("text/plain", member.id);
    }
  }

  function handleDragOver(e) {
    // Allow drop only from same room type
    e.preventDefault();
  }

  async function handleDrop(targetRoomId) {
    if (!draggedMember || dragSourceRoomId === targetRoomId) {
      draggedMember = null;
      dragSourceRoomId = null;
      return;
    }

    // Find target room and check capacity
    const targetRoom = rooms.find((r) => r.id === targetRoomId);
    if (targetRoom && targetRoom.is_full) {
      error = "Kamar sudah penuh";
      draggedMember = null;
      dragSourceRoomId = null;
      return;
    }

    // --- OPTIMISTIC UI UPDATE --- move member instantly
    const memberId = draggedMember.id;
    const sourceRoomId = dragSourceRoomId;

    // Snapshot for rollback
    const prevRooms = JSON.parse(JSON.stringify(rooms));
    const prevSummary = summary ? JSON.parse(JSON.stringify(summary)) : null;

    // Find member data from source room
    let memberData = null;
    if (sourceRoomId !== null) {
      const srcRoom = rooms.find((r) => r.id === sourceRoomId);
      if (srcRoom) {
        memberData = srcRoom.members.find((m) => m.id === memberId);
        // Remove from source room
        srcRoom.members = srcRoom.members.filter((m) => m.id !== memberId);
        srcRoom.occupied = srcRoom.members.length;
        srcRoom.is_full = srcRoom.occupied >= srcRoom.capacity;
      }
    }

    // Add to target room
    if (targetRoom && memberData) {
      targetRoom.members = [...targetRoom.members, memberData];
      targetRoom.occupied = targetRoom.members.length;
      targetRoom.is_full = targetRoom.occupied >= targetRoom.capacity;
    }

    // Auto-delete empty source room
    if (sourceRoomId !== null) {
      const srcRoom = rooms.find((r) => r.id === sourceRoomId);
      if (srcRoom && srcRoom.members.length === 0) {
        rooms = rooms.filter((r) => r.id !== sourceRoomId);
        if (summary) {
          summary = { ...summary, total_rooms: summary.total_rooms - 1 };
        }
      }
    }

    // Force Svelte reactivity
    rooms = [...rooms];

    // Update summary counts optimistically
    if (summary && sourceRoomId === null) {
      // Moving from unassigned to room
      summary = {
        ...summary,
        assigned_count: summary.assigned_count + 1,
        unassigned_count: summary.unassigned_count - 1,
      };
    }

    // Clear drag state immediately
    draggedMember = null;
    dragSourceRoomId = null;

    // --- BACKGROUND SYNC ---
    try {
      if (sourceRoomId !== null) {
        await ApiService.unassignMember(memberId);
      }
      await ApiService.assignMemberToRoom(memberId, targetRoomId);
    } catch (e) {
      // Revert on failure
      rooms = prevRooms;
      summary = prevSummary;
      error = e.message;
    }
  }

  async function removeFromRoom(memberId) {
    // --- OPTIMISTIC UI UPDATE ---
    const prevRooms = JSON.parse(JSON.stringify(rooms));
    const prevSummary = summary ? JSON.parse(JSON.stringify(summary)) : null;

    // Find and remove the member from their room locally
    for (const room of rooms) {
      const idx = room.members.findIndex((m) => m.id === memberId);
      if (idx !== -1) {
        room.members.splice(idx, 1);
        room.occupied = room.members.length;
        room.is_full = room.occupied >= room.capacity;
        break;
      }
    }
    // Auto-delete empty rooms
    const emptyCount = rooms.filter((r) => r.members.length === 0).length;
    rooms = rooms.filter((r) => r.members.length > 0);

    if (summary) {
      summary = {
        ...summary,
        assigned_count: summary.assigned_count - 1,
        unassigned_count: summary.unassigned_count + 1,
        total_rooms: summary.total_rooms - emptyCount,
      };
    }

    // --- BACKGROUND SYNC ---
    try {
      await ApiService.unassignMember(memberId);
    } catch (e) {
      rooms = prevRooms;
      summary = prevSummary;
      error = e.message;
    }
  }

  async function addRoom() {
    if (!selectedGroupId || !newRoomNumber.trim()) return;

    isAddingRoom = true;
    error = null;

    try {
      const newRoom = await ApiService.createRoom(
        selectedGroupId,
        newRoomNumber.trim(),
        newRoomGender,
        "quad",
        4,
      );
      // Optimistic: add to local state
      rooms = [
        ...rooms,
        { ...newRoom, members: [], occupied: 0, is_full: false },
      ];
      if (summary) {
        summary = { ...summary, total_rooms: summary.total_rooms + 1 };
      }
      // Reset form
      newRoomNumber = "";
      newRoomGender = "male";
      showAddRoom = false;
    } catch (e) {
      error = e.message;
    } finally {
      isAddingRoom = false;
    }
  }

  async function deleteRoom(roomId) {
    // --- OPTIMISTIC UI ---
    const prevRooms = JSON.parse(JSON.stringify(rooms));
    const prevSummary = summary ? JSON.parse(JSON.stringify(summary)) : null;

    const room = rooms.find((r) => r.id === roomId);
    const memberCount = room ? room.members.length : 0;

    rooms = rooms.filter((r) => r.id !== roomId);
    if (summary) {
      summary = {
        ...summary,
        total_rooms: summary.total_rooms - 1,
        assigned_count: summary.assigned_count - memberCount,
        unassigned_count: summary.unassigned_count + memberCount,
      };
    }

    // --- BACKGROUND SYNC ---
    try {
      await ApiService.deleteRoom(roomId);
    } catch (e) {
      rooms = prevRooms;
      summary = prevSummary;
      error = e.message;
    }
  }
</script>

{#if isOpen}
  <div class="min-h-screen bg-slate-50/70 p-4 lg:p-8">
    <PageHeader
      kicker="Operasional"
      title="Auto-Rooming"
      subtitle="Alokasi kamar hotel otomatis, tetap bisa disesuaikan manual."
    >
      {#snippet actions()}
        {#if selectedGroupId}
          <button
            type="button"
            onclick={loadData}
            disabled={isLoading}
            class="flex h-10 w-10 items-center justify-center rounded-xl border border-slate-200/70 bg-white text-slate-500 shadow-sm transition hover:bg-slate-50 disabled:opacity-50"
            title="Muat ulang"
          >
            <RefreshCw class="w-4 h-4 {isLoading ? 'animate-spin' : ''}" />
          </button>
        {/if}
        <button
          type="button"
          onclick={runAutoRooming}
          disabled={isAutoRooming || unassignedCount === 0}
          class="inline-flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white shadow-sm shadow-primary-600/30 transition hover:bg-primary-700 disabled:opacity-50"
        >
          {#if isAutoRooming}<Loader2
              class="w-4 h-4 animate-spin"
            />{:else}<Sparkles class="w-4 h-4" />{/if}
          Auto-Generate
        </button>
        {#if summary?.total_rooms > 0}
          <button
            type="button"
            onclick={exportRoomingPDF}
            class="inline-flex items-center gap-2 rounded-xl border border-slate-200/70 bg-white px-4 py-2.5 text-sm font-semibold text-slate-600 shadow-sm transition hover:bg-slate-50"
          >
            <FileDown class="w-4 h-4" />
            PDF
          </button>
          <button
            type="button"
            onclick={() => (waBlastOpen = true)}
            class="inline-flex items-center gap-2 rounded-xl border border-slate-200/70 bg-white px-4 py-2.5 text-sm font-semibold text-emerald-600 shadow-sm transition hover:bg-emerald-50"
          >
            <MessageCircle class="w-4 h-4" />
            WA
          </button>
          <button
            type="button"
            onclick={clearRooms}
            disabled={isLoading}
            class="inline-flex items-center gap-2 rounded-xl border border-slate-200/70 bg-white px-4 py-2.5 text-sm font-semibold text-red-600 shadow-sm transition hover:bg-red-50 disabled:opacity-50"
          >
            <Trash2 class="w-4 h-4" />
            Reset
          </button>
        {/if}
      {/snippet}
    </PageHeader>

    <!-- Group selector -->
    <div class="mb-5 rounded-2xl border border-slate-200/70 bg-white p-4 shadow-sm">
      <label
        for="room-group-select"
        class="mb-2 block text-[11.5px] font-semibold uppercase tracking-wide text-slate-400"
        >Grup Keberangkatan</label
      >
      <select
        id="room-group-select"
        bind:value={selectedGroupId}
        onchange={loadData}
        class="w-full rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-medium text-slate-700 outline-none transition focus:border-primary-400 focus:bg-white focus:ring-2 focus:ring-primary-100 lg:max-w-sm"
      >
        <option value="">Pilih Grup</option>
        {#each groups as group}
          <option value={group.id}>{group.name} ({group.member_count})</option>
        {/each}
      </select>
    </div>

    <!-- Messages -->
    {#if error}
      <div
        class="mb-3 flex items-center gap-2 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm"
      >
        <AlertTriangle class="w-4 h-4 text-red-500" />
        <span class="text-red-700 flex-1">{error}</span>
        <button
          type="button"
          onclick={() => (error = null)}
          class="text-red-500"><X class="w-4 h-4" /></button
        >
      </div>
    {/if}

    {#if success}
      <div
        class="mb-3 flex items-center gap-2 rounded-2xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm"
      >
        <CheckCircle class="w-4 h-4 text-emerald-500" />
        <span class="text-emerald-700 flex-1">{success}</span>
        <button
          type="button"
          onclick={() => (success = null)}
          class="text-emerald-500"><X class="w-4 h-4" /></button
        >
      </div>
    {/if}

    <!-- Content -->
    {#if isLoading}
      <div
        class="flex items-center justify-center rounded-2xl border border-slate-200/70 bg-white py-16 shadow-sm"
      >
        <Loader2 class="w-6 h-6 animate-spin text-primary-600" />
      </div>
    {:else if !selectedGroupId}
      <div
        class="flex flex-col items-center justify-center rounded-2xl border border-slate-200/70 bg-white py-16 text-slate-400 shadow-sm"
      >
        <Hotel class="w-10 h-10 mb-2" />
        <p class="text-sm">Pilih grup untuk memulai</p>
      </div>
    {:else if summary}
      <!-- Stats -->
      <div class="mb-5 grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard
          icon={Users}
          label="Total Jamaah"
          value={String(summary.total_members)}
          accent="#1B7F5A"
        />
        <StatCard
          icon={CheckCircle}
          label="Sudah Ditempatkan"
          value={String(summary.assigned_count)}
          accent="#1B7F5A"
        />
        <StatCard
          icon={Users}
          label="Belum Ditempatkan"
          value={String(unassignedCount)}
          accent="#C99A2E"
        />
        <StatCard
          icon={TrendingUp}
          label="Okupansi"
          value={`${occupancyPct}%`}
          accent="#2563a8"
          sub={`${summary.total_rooms} kamar`}
        />
      </div>

      <div
        class="overflow-hidden rounded-2xl border border-slate-200/70 bg-white shadow-sm"
      >
        <!-- Unassigned Notice -->
        {#if unassignedCount > 0}
          <div
            class="m-4 flex items-center gap-2 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm"
          >
            <Users class="w-4 h-4 text-amber-600" />
            <span class="text-amber-700">
              <strong>{unassignedCount}</strong> jamaah belum ditempatkan
            </span>
          </div>
        {/if}

        <!-- Rooms Grid -->
        {#if rooms.length > 0}
          <div
            class="grid grid-cols-1 gap-3.5 p-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4"
          >
            {#each rooms as room}
              <article
                class="rounded-2xl border border-slate-200/70 bg-white p-4 shadow-sm transition hover:shadow-md"
                ondragover={handleDragOver}
                ondrop={(e) => {
                  e.preventDefault();
                  handleDrop(room.id);
                }}
                role="region"
                aria-label="Kamar {room.room_number}"
              >
                <!-- Room Header -->
                <div class="mb-3 flex items-center justify-between">
                  <div class="flex items-center gap-2.5">
                    <div
                      class="flex h-9 w-9 items-center justify-center rounded-xl bg-primary-50 text-primary-700"
                    >
                      <Bed class="h-4 w-4" />
                    </div>
                    <div>
                      <p class="text-sm font-bold text-[#10211c]">
                        {room.room_number}
                      </p>
                      <span
                        class="text-[11.5px] font-medium {room.gender_type ===
                        'male'
                          ? 'text-blue-600'
                          : room.gender_type === 'female'
                            ? 'text-pink-600'
                            : 'text-purple-600'}"
                      >
                        {room.gender_type === "male"
                          ? "Laki-laki"
                          : room.gender_type === "female"
                            ? "Perempuan"
                            : "Keluarga"}
                      </span>
                    </div>
                  </div>
                  <div class="flex items-center gap-1.5">
                    {#if room.is_auto_assigned}
                      <span
                        class="rounded-full bg-primary-50 px-2 py-0.5 text-[10px] font-bold text-primary-700"
                        >Auto</span
                      >
                    {/if}
                    <span
                      class="rounded-full px-2.5 py-0.5 text-[11.5px] font-bold {room.is_full
                        ? 'bg-primary-50 text-primary-700'
                        : 'bg-slate-100 text-slate-500'}"
                    >
                      {room.occupied}/{room.capacity}
                    </span>
                    <button
                      type="button"
                      onclick={() => deleteRoom(room.id)}
                      class="p-1 text-slate-300 transition-colors hover:text-red-500"
                      title="Hapus kamar"
                    >
                      <Trash2 class="h-3.5 w-3.5" />
                    </button>
                  </div>
                </div>

                <!-- Capacity Bar -->
                <div class="mb-3 h-1.5 overflow-hidden rounded-full bg-slate-100">
                  <div
                    class="h-full rounded-full transition-all {room.is_full
                      ? 'bg-primary-600'
                      : 'bg-gold-500'}"
                    style="width: {(room.occupied / room.capacity) * 100}%"
                  ></div>
                </div>

                <!-- Members -->
                <div class="flex flex-col gap-2">
                  {#each room.members as member}
                    <div
                      class="flex cursor-move items-center gap-2.5 rounded-xl border border-slate-200/70 bg-white px-2.5 py-2 shadow-sm"
                      draggable="true"
                      role="listitem"
                      ondragstart={(e) =>
                        handleDragStart(e, { id: member.id }, room.id)}
                    >
                      <Avatar name={member.nama || "Tanpa Nama"} size={28} />
                      <span
                        class="min-w-0 flex-1 truncate text-[12.5px] font-semibold text-[#10211c]"
                        >{member.nama || "Tanpa Nama"}</span
                      >
                      <button
                        type="button"
                        onclick={() => removeFromRoom(member.id)}
                        class="text-slate-300 transition-colors hover:text-red-500"
                      >
                        <X class="h-3.5 w-3.5" />
                      </button>
                    </div>
                  {/each}

                  <!-- Empty Slots -->
                  {#each Array(Math.max(0, room.capacity - room.occupied)) as _, i}
                    <div
                      class="flex h-11 items-center justify-center rounded-xl border border-dashed border-slate-200 text-[11.5px] text-slate-400"
                    >
                      Kosong
                    </div>
                  {/each}
                </div>
              </article>
            {/each}

            <!-- Add Room Card -->
            {#if showAddRoom}
              <article
                class="rounded-2xl border border-dashed border-primary-300 bg-primary-50/50 p-4"
              >
                <div class="mb-3 flex items-center justify-between">
                  <span class="text-sm font-bold text-primary-700">Kamar Baru</span>
                  <button
                    type="button"
                    onclick={() => {
                      showAddRoom = false;
                    }}
                    class="text-slate-400 hover:text-red-500"
                  >
                    <X class="h-4 w-4" />
                  </button>
                </div>
                <div class="space-y-2">
                  <input
                    type="text"
                    bind:value={newRoomNumber}
                    placeholder="No. kamar (cth: 301)"
                    class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
                  />
                  <select
                    bind:value={newRoomGender}
                    class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm outline-none focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
                  >
                    <option value="male">Laki-laki</option>
                    <option value="female">Perempuan</option>
                    <option value="family">Keluarga</option>
                  </select>
                  <button
                    type="button"
                    onclick={addRoom}
                    disabled={isAddingRoom || !newRoomNumber.trim()}
                    class="flex w-full items-center justify-center gap-1.5 rounded-xl bg-primary-600 px-3 py-2 text-sm font-semibold text-white transition hover:bg-primary-700 disabled:opacity-50"
                  >
                    {#if isAddingRoom}<Loader2
                        class="h-4 w-4 animate-spin"
                      />{:else}<Plus class="h-4 w-4" />{/if}
                    Buat Kamar
                  </button>
                </div>
              </article>
            {:else}
              <button
                type="button"
                onclick={() => {
                  showAddRoom = true;
                }}
                class="flex min-h-[120px] cursor-pointer flex-col items-center justify-center gap-2 rounded-2xl border border-dashed border-slate-300 text-slate-400 transition-colors hover:border-primary-400 hover:bg-primary-50/30 hover:text-primary-600"
              >
                <Plus class="h-6 w-6" />
                <span class="text-sm font-semibold">Tambah Kamar</span>
              </button>
            {/if}
          </div>
        {:else}
          <div class="py-12 text-center text-sm text-slate-400">
            <Hotel class="mx-auto mb-3 h-10 w-10" />
            <p class="font-medium text-slate-500">Belum ada kamar</p>
            <p class="mt-1 text-xs">Klik "Auto-Generate" atau tambah manual</p>
            <button
              type="button"
              onclick={() => {
                showAddRoom = true;
              }}
              class="mt-4 inline-flex items-center gap-1.5 rounded-xl border border-dashed border-primary-300 px-4 py-2 text-sm font-semibold text-primary-600 transition-colors hover:bg-primary-50"
            >
              <Plus class="h-4 w-4" />
              Tambah Kamar Manual
            </button>
          </div>
        {/if}
      </div>
    {:else}
      <div
        class="rounded-2xl border border-slate-200/70 bg-white py-12 text-center text-sm text-slate-400 shadow-sm"
      >
        Tidak ada data
      </div>
    {/if}
  </div>
{/if}

<!-- WhatsApp Blast Modal -->
<WhatsAppBlast
  isOpen={waBlastOpen}
  onClose={() => (waBlastOpen = false)}
  group={waBlastGroup}
/>
