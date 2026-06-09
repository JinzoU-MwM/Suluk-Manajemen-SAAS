<!--
  RoomingPage.svelte — Suluk design Rooming List
  Drag jamaah between rooms. Real data wiring preserved (fetch, auto-rooming,
  reset, drag/drop, add/delete room, remove member, PDF, WhatsApp blast).
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
    GripVertical,
  } from "lucide-svelte";
  import { ApiService } from "../services/api.js";
  import PageHeader from "../components/PageHeader.svelte";
  import StatCard from "../components/StatCard.svelte";
  import Avatar from "../components/Avatar.svelte";
  import ProgressBar from "../components/ui/ProgressBar.svelte";
  import Button from "../components/ui/Button.svelte";
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
  let dragOverRoomId = $state(null);

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

  function genderLabel(g) {
    return g === "male" ? "Laki-laki" : g === "female" ? "Perempuan" : "Keluarga";
  }
  function genderToken(g) {
    return g === "male"
      ? "var(--c-info)"
      : g === "female"
        ? "#c2298a"
        : "var(--c-accent)";
  }

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
      const gLabel = genderLabel(room.gender_type);
      if (members.length === 0) {
        tableRows += `<tr><td>${room.room_number}</td><td>${gLabel}</td><td>${room.capacity || 4}</td><td>-</td><td>-</td></tr>`;
      } else {
        members.forEach((m, idx) => {
          tableRows += `<tr>
            ${idx === 0 ? `<td rowspan="${members.length}">${room.room_number}</td><td rowspan="${members.length}">${gLabel}</td><td rowspan="${members.length}">${members.length}/${room.capacity || 4}</td>` : ""}
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

  function handleDragEnd() {
    draggedMember = null;
    dragSourceRoomId = null;
    dragOverRoomId = null;
  }

  function handleDragOver(e) {
    // Allow drop only from same room type
    e.preventDefault();
  }

  async function handleDrop(targetRoomId) {
    dragOverRoomId = null;
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
  <div class="page-enter" style="min-height:100vh;background:var(--c-bg);padding:24px">
    <PageHeader
      kicker="Operasional"
      title="Rooming List"
      subtitle="Alokasi kamar hotel otomatis — seret jamaah antar kamar untuk mengatur penempatan."
    >
      {#snippet actions()}
        <select
          bind:value={selectedGroupId}
          onchange={loadData}
          aria-label="Grup Keberangkatan"
          style="padding:9px 14px;font-size:13.5px;font-weight:600;color:var(--c-ink-soft);background:var(--c-surface);border:1px solid var(--c-line);border-radius:var(--radius);outline:none;min-width:200px;max-width:280px"
        >
          <option value="">Pilih Grup</option>
          {#each groups as group}
            <option value={group.id}>{group.name} ({group.member_count})</option>
          {/each}
        </select>

        {#if selectedGroupId}
          <Button variant="ghost" icon={RefreshCw} onclick={loadData} disabled={isLoading}>
            Muat Ulang
          </Button>
        {/if}

        <Button
          variant="primary"
          icon={isAutoRooming ? Loader2 : Sparkles}
          onclick={runAutoRooming}
          disabled={isAutoRooming || unassignedCount === 0}
        >
          Auto-Generate
        </Button>

        {#if summary?.total_rooms > 0}
          <Button variant="ghost" icon={FileDown} onclick={exportRoomingPDF}>PDF</Button>
          <Button variant="soft" icon={MessageCircle} onclick={() => (waBlastOpen = true)}>
            WA
          </Button>
          <Button variant="danger" icon={Trash2} onclick={clearRooms} disabled={isLoading}>
            Reset
          </Button>
        {/if}
      {/snippet}
    </PageHeader>

    <!-- Messages -->
    {#if error}
      <div
        style="margin-bottom:14px;display:flex;align-items:center;gap:10px;padding:12px 16px;border-radius:var(--radius);background:var(--c-danger-soft);color:var(--c-danger);font-size:13.5px"
      >
        <AlertTriangle size={17} />
        <span style="flex:1">{error}</span>
        <button type="button" onclick={() => (error = null)} style="color:var(--c-danger);display:inline-flex">
          <X size={16} />
        </button>
      </div>
    {/if}

    {#if success}
      <div
        style="margin-bottom:14px;display:flex;align-items:center;gap:10px;padding:12px 16px;border-radius:var(--radius);background:var(--c-success-soft);color:var(--c-success);font-size:13.5px"
      >
        <CheckCircle size={17} />
        <span style="flex:1">{success}</span>
        <button type="button" onclick={() => (success = null)} style="color:var(--c-success);display:inline-flex">
          <X size={16} />
        </button>
      </div>
    {/if}

    <!-- Content -->
    {#if isLoading}
      <div
        style="display:flex;align-items:center;justify-content:center;padding:64px 0;background:var(--c-surface);border:1px solid var(--c-line);border-radius:var(--radius-lg);box-shadow:var(--shadow-sm)"
      >
        <Loader2 size={26} class="animate-spin" style="color:var(--c-primary)" />
      </div>
    {:else if !selectedGroupId}
      <div
        style="display:flex;flex-direction:column;align-items:center;justify-content:center;padding:64px 0;color:var(--c-faint);background:var(--c-surface);border:1px solid var(--c-line);border-radius:var(--radius-lg);box-shadow:var(--shadow-sm)"
      >
        <Hotel size={40} style="margin-bottom:10px" />
        <p style="font-size:13.5px">Pilih grup untuk memulai</p>
      </div>
    {:else if summary}
      <!-- Stats -->
      <div
        style="display:grid;grid-template-columns:repeat(auto-fit,minmax(200px,1fr));gap:var(--gap);margin-bottom:var(--gap)"
      >
        <StatCard
          icon={Users}
          label="Total Jamaah"
          value={String(summary.total_members ?? "—")}
          accent="var(--c-primary)"
        />
        <StatCard
          icon={CheckCircle}
          label="Sudah Ditempatkan"
          value={String(summary.assigned_count ?? 0)}
          accent="var(--c-success)"
        />
        <StatCard
          icon={Users}
          label="Belum Ditempatkan"
          value={String(unassignedCount)}
          accent="var(--c-warning)"
        />
        <StatCard
          icon={TrendingUp}
          label="Okupansi"
          value={`${occupancyPct}%`}
          accent="var(--c-accent)"
          sub={`${summary.total_rooms ?? rooms.length} kamar`}
        />
      </div>

      <!-- Unassigned notice -->
      {#if unassignedCount > 0}
        <div
          style="margin-bottom:14px;display:flex;align-items:center;gap:10px;padding:12px 16px;border-radius:var(--radius);background:var(--c-warning-soft);color:var(--c-warning);font-size:13.5px"
        >
          <Users size={17} />
          <span><strong>{unassignedCount}</strong> jamaah belum ditempatkan — seret ke kamar.</span>
        </div>
      {/if}

      <!-- Rooms grid -->
      {#if rooms.length > 0}
        <div
          style="display:grid;grid-template-columns:repeat(auto-fill,minmax(240px,1fr));gap:14px"
        >
          {#each rooms as room}
            {@const isOver = dragOverRoomId === room.id}
            <div
              role="region"
              aria-label="Kamar {room.room_number}"
              ondragover={(e) => {
                handleDragOver(e);
                dragOverRoomId = room.id;
              }}
              ondragleave={() => {
                if (dragOverRoomId === room.id) dragOverRoomId = null;
              }}
              ondrop={(e) => {
                e.preventDefault();
                handleDrop(room.id);
              }}
              style="background:var(--c-surface);border:1px solid {isOver
                ? 'var(--c-primary)'
                : 'var(--c-line)'};border-radius:var(--radius-lg);padding:14px;box-shadow:{isOver
                ? 'var(--shadow)'
                : 'var(--shadow-sm)'};transition:all .15s"
            >
              <!-- Room header -->
              <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
                <div style="display:flex;align-items:center;gap:9px;min-width:0">
                  <div
                    style="width:34px;height:34px;flex-shrink:0;border-radius:var(--radius-sm);background:var(--c-primary-soft);color:var(--c-primary-deep);display:flex;align-items:center;justify-content:center"
                  >
                    <Bed size={17} />
                  </div>
                  <div style="min-width:0">
                    <div style="font-size:14px;font-weight:800;color:var(--c-ink)">
                      {room.room_number}
                    </div>
                    <div style="font-size:11.5px;font-weight:600;color:{genderToken(room.gender_type)}">
                      {genderLabel(room.gender_type)}
                    </div>
                  </div>
                </div>
                <div style="display:flex;align-items:center;gap:6px;flex-shrink:0">
                  {#if room.is_auto_assigned}
                    <span
                      style="font-size:10px;font-weight:800;color:var(--c-primary);background:var(--c-primary-soft);padding:2px 7px;border-radius:999px"
                      >Auto</span
                    >
                  {/if}
                  <span
                    style="font-size:11.5px;font-weight:700;padding:3px 9px;border-radius:999px;color:{room.is_full
                      ? 'var(--c-success)'
                      : 'var(--c-muted)'};background:{room.is_full
                      ? 'var(--c-success-soft)'
                      : 'var(--c-bg-2)'}"
                  >
                    {room.occupied}/{room.capacity}
                  </span>
                  <button
                    type="button"
                    onclick={() => deleteRoom(room.id)}
                    title="Hapus kamar"
                    style="color:var(--c-faint);display:inline-flex;padding:2px"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>

              <!-- Capacity bar -->
              <div style="margin-bottom:12px">
                <ProgressBar
                  value={room.occupied}
                  max={room.capacity}
                  color={room.is_full ? "var(--c-success)" : "var(--c-accent)"}
                  height={6}
                />
              </div>

              <!-- Members + empty slots -->
              <div style="display:flex;flex-direction:column;gap:8px;min-height:{room.capacity * 46}px">
                {#each room.members as member}
                  {@const dragging =
                    draggedMember &&
                    draggedMember.id === member.id &&
                    dragSourceRoomId === room.id}
                  <div
                    role="listitem"
                    draggable="true"
                    ondragstart={(e) => handleDragStart(e, { id: member.id }, room.id)}
                    ondragend={handleDragEnd}
                    style="display:flex;align-items:center;gap:9px;padding:8px 10px;background:var(--c-surface);border:1px solid var(--c-line);border-radius:var(--radius);cursor:grab;box-shadow:var(--shadow-sm);opacity:{dragging
                      ? 0.4
                      : 1};transition:opacity .15s"
                  >
                    <Avatar name={member.nama || "Tanpa Nama"} size={28} />
                    <span
                      style="font-size:12.5px;font-weight:600;flex:1;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;color:var(--c-ink)"
                      >{member.nama || "Tanpa Nama"}</span
                    >
                    <button
                      type="button"
                      onclick={() => removeFromRoom(member.id)}
                      title="Keluarkan dari kamar"
                      style="color:var(--c-faint);display:inline-flex"
                    >
                      <X size={14} />
                    </button>
                    <GripVertical size={15} style="color:var(--c-faint);flex-shrink:0" />
                  </div>
                {/each}

                {#each Array(Math.max(0, room.capacity - room.occupied)) as _, i}
                  <div
                    style="border:1.5px dashed var(--c-line);border-radius:var(--radius);height:44px;display:flex;align-items:center;justify-content:center;font-size:12px;color:var(--c-faint)"
                  >
                    Kosong
                  </div>
                {/each}
              </div>
            </div>
          {/each}

          <!-- Add room card -->
          {#if showAddRoom}
            <div
              style="background:var(--c-primary-tint);border:1.5px dashed var(--c-primary);border-radius:var(--radius-lg);padding:14px"
            >
              <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:12px">
                <span style="font-size:14px;font-weight:800;color:var(--c-primary-deep)">Kamar Baru</span>
                <button
                  type="button"
                  onclick={() => (showAddRoom = false)}
                  style="color:var(--c-faint);display:inline-flex"
                >
                  <X size={16} />
                </button>
              </div>
              <div style="display:flex;flex-direction:column;gap:8px">
                <input
                  type="text"
                  bind:value={newRoomNumber}
                  placeholder="No. kamar (cth: 301)"
                  style="width:100%;padding:9px 12px;font-size:13px;background:var(--c-surface);border:1px solid var(--c-line);border-radius:var(--radius);outline:none"
                />
                <select
                  bind:value={newRoomGender}
                  style="width:100%;padding:9px 12px;font-size:13px;background:var(--c-surface);border:1px solid var(--c-line);border-radius:var(--radius);outline:none"
                >
                  <option value="male">Laki-laki</option>
                  <option value="female">Perempuan</option>
                  <option value="family">Keluarga</option>
                </select>
                <Button
                  variant="primary"
                  icon={isAddingRoom ? Loader2 : Plus}
                  full
                  onclick={addRoom}
                  disabled={isAddingRoom || !newRoomNumber.trim()}
                >
                  Buat Kamar
                </Button>
              </div>
            </div>
          {:else}
            <button
              type="button"
              onclick={() => (showAddRoom = true)}
              style="min-height:120px;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:8px;border:1.5px dashed var(--c-line);border-radius:var(--radius-lg);color:var(--c-faint);background:transparent;cursor:pointer;transition:all .15s"
            >
              <Plus size={24} />
              <span style="font-size:13.5px;font-weight:700">Tambah Kamar</span>
            </button>
          {/if}
        </div>
      {:else}
        <div
          style="padding:48px 0;text-align:center;background:var(--c-surface);border:1px solid var(--c-line);border-radius:var(--radius-lg);box-shadow:var(--shadow-sm)"
        >
          <Hotel size={40} style="margin:0 auto 12px;color:var(--c-faint)" />
          <p style="font-weight:600;color:var(--c-muted)">Belum ada kamar</p>
          <p style="margin-top:4px;font-size:12.5px;color:var(--c-faint)">
            Klik "Auto-Generate" atau tambah manual
          </p>
          <div style="margin-top:16px;display:inline-flex">
            <Button variant="soft" icon={Plus} onclick={() => (showAddRoom = true)}>
              Tambah Kamar Manual
            </Button>
          </div>
        </div>
      {/if}
    {:else}
      <div
        style="padding:48px 0;text-align:center;font-size:13.5px;color:var(--c-faint);background:var(--c-surface);border:1px solid var(--c-line);border-radius:var(--radius-lg);box-shadow:var(--shadow-sm)"
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
