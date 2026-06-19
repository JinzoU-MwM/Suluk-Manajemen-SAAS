<!--
  InventoryPage.svelte — Suluk design (Inventaris)

  PageHeader (kicker + title + group selector/actions) · StatCard forecast row ·
  bulk-action toolbar · member equipment table (design th/td styling, Avatar in the
  member cell, status Badge for Diterima/Menunggu).
  Presentation only — all data fetching, $props, $state/$derived, handlers, and labels preserved.
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
  import EmptyState from "../components/EmptyState.svelte";
  import Card from "../components/ui/Card.svelte";
  import Badge from "../components/ui/Badge.svelte";
  import Button from "../components/ui/Button.svelte";
  import ProgressBar from "../components/ui/ProgressBar.svelte";
  import FilterTabs from "../components/ui/FilterTabs.svelte";
  import StockTab from "./inventory/StockTab.svelte";
  import { ApiService } from "../services/api.js";

  let { isOpen = false, onClose, groups = [], isPro = false } = $props();

  // Tab switcher: "distribusi" | "stok"
  let tab = $state("distribusi");

  // State
  let selectedGroupId = $state(null);
  let forecast = $state(null);
  let fulfillment = $state(null);
  let isLoading = $state(false);
  let error = $state(null);
  let selectedMembers = $state(new Set());
  let isMarking = $state(false);

  // QR handover (Phase 4C)
  let checkpoints = $state([]);
  let scanToken = $state("");
  let scanCheckpoint = $state("equipment");
  let isScanning = $state(false);
  let scanMsg = $state("");
  let scanOk = $state(false);
  let qrUrl = $state(null);
  let qrMember = $state(null);

  let handoverDone = $derived(
    checkpoints.filter((m) => m.is_equipment_received && m.is_luggage_checked).length,
  );

  // Derived values
  let selectedGroup = $derived(groups.find((g) => g.id === selectedGroupId));
  let canMarkSelected = $derived(selectedMembers.size > 0 && !isMarking);
  let receivedCount = $derived(fulfillment?.received_count || 0);
  let pendingCount = $derived(fulfillment?.pending_count || 0);
  let totalFulfillment = $derived(receivedCount + pendingCount);

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
      await loadCheckpoints();
    } catch (e) {
      error = e.message;
    } finally {
      isLoading = false;
    }
  }

  async function loadCheckpoints() {
    if (!selectedGroupId) {
      checkpoints = [];
      return;
    }
    try {
      const r = await ApiService.getHandoverCheckpoints(selectedGroupId);
      checkpoints = r.members || [];
    } catch {
      checkpoints = [];
    }
  }

  async function submitScan() {
    const token = scanToken.trim();
    if (!token || isScanning) return;
    isScanning = true;
    scanMsg = "";
    try {
      const m = await ApiService.scanHandover({
        token,
        checkpoint: scanCheckpoint,
        items: scanCheckpoint === "equipment" ? ["koper", "baju"] : [],
      });
      scanOk = true;
      scanMsg = `✓ ${m.nama} — ${scanCheckpoint === "equipment" ? "perlengkapan" : "koper"} tercatat`;
      scanToken = "";
      await loadCheckpoints();
    } catch (e) {
      scanOk = false;
      scanMsg = `✗ ${e.message}`;
    } finally {
      isScanning = false;
    }
  }

  async function showQr(member) {
    closeQr();
    qrMember = member;
    try {
      qrUrl = await ApiService.getMemberQrUrl(member.member_id);
    } catch (e) {
      scanOk = false;
      scanMsg = e.message;
      qrMember = null;
    }
  }

  function closeQr() {
    if (qrUrl) URL.revokeObjectURL(qrUrl);
    qrUrl = null;
    qrMember = null;
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
  <div class="inv-page">
    <PageHeader
      kicker="Logistik"
      title="Inventaris"
      subtitle="Pantau stok perlengkapan jamaah — koper, seragam, kain ihram, dan distribusi ke jamaah."
    >
      {#snippet actions()}
        <select
          id="inv-group-select"
          bind:value={selectedGroupId}
          onchange={loadGroupData}
          class="inv-select"
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
            class="inv-icon-btn"
            title="Muat ulang"
          >
            <RefreshCw class="h-4 w-4 {isLoading ? 'animate-spin' : ''}" />
          </button>
        {/if}
      {/snippet}
    </PageHeader>

    <FilterTabs
      tabs={[
        { value: "distribusi", label: "Distribusi" },
        { value: "stok", label: "Stok" },
      ]}
      value={tab}
      onChange={(v) => (tab = v)}
    />

    {#if tab === "stok"}
      <StockTab {groups} />
    {/if}

    {#if tab === "distribusi"}

    <!-- Summary / forecast cards -->
    {#if forecast}
      <div class="inv-stats">
        <StatCard
          icon={Boxes}
          label="Total Jamaah"
          value={String(forecast.total_members)}
          accent="var(--c-primary)"
        />
        <StatCard
          icon={Package}
          label="Koper"
          value={String(forecast.requirements?.koper || 0)}
          accent="var(--c-info)"
        />
        <StatCard
          icon={CheckCircle}
          label="Ihram"
          value={String(forecast.requirements?.ihram || 0)}
          accent="var(--c-success)"
        />
        <StatCard
          icon={Shirt}
          label="Mukena"
          value={String(forecast.requirements?.mukena || 0)}
          accent="var(--c-accent)"
        />
      </div>
    {/if}

    <!-- Bulk actions toolbar -->
    {#if fulfillment && fulfillment.pending?.length > 0}
      <Card pad={false} class="inv-toolbar-card" style="margin-bottom:var(--gap)">
        <div class="inv-toolbar">
          <span class="inv-toolbar-count">{selectedMembers.size} dipilih</span>
          <button type="button" onclick={selectAllPending} class="inv-link inv-link-primary"
            >Semua</button
          >
          <button type="button" onclick={clearSelection} class="inv-link"
            >Batal</button
          >
          <div class="inv-toolbar-spacer"></div>
          <Button
            variant="primary"
            size="sm"
            onclick={markSelectedAsReceived}
            disabled={!canMarkSelected}
          >
            {#if isMarking}<Loader2 class="h-3.5 w-3.5 animate-spin" />{:else}<CheckCircle
                class="h-3.5 w-3.5"
              />{/if}
            Tandai Terima
          </Button>
        </div>
      </Card>
    {/if}

    <!-- Error -->
    {#if error}
      <div class="inv-error" role="alert">
        <AlertTriangle class="h-4 w-4" style="color:var(--c-danger);flex-shrink:0" />
        <span class="inv-error-text">{error}</span>
        <button type="button" onclick={() => (error = null)} class="inv-error-close">
          <X class="h-4 w-4" />
        </button>
      </div>
    {/if}

    <!-- Content -->
    <Card pad={false}>
      {#if isLoading}
        <div class="inv-center">
          <Loader2 class="h-6 w-6 animate-spin" style="color:var(--c-primary)" />
        </div>
      {:else if !selectedGroupId}
        <EmptyState
          icon={Package}
          title="Pilih grup untuk memulai"
          text="Pilih sebuah grup keberangkatan di atas untuk melihat kebutuhan perlengkapan dan status distribusi jamaah."
        />
      {:else if forecast}
        <!-- Size & Status summary row -->
        <div class="inv-meta-row">
          <!-- Sizes -->
          <div class="inv-meta-box inv-meta-grow">
            <div class="inv-meta-head">
              <Shirt class="h-3.5 w-3.5" style="color:var(--c-faint)" />
              <span>Ukuran Baju</span>
            </div>
            <div class="inv-chips">
              {#each Object.entries(forecast.size_breakdown || {}) as [size, count]}
                <span class="inv-chip">{size || "N/A"}: {count}</span>
              {/each}
            </div>
          </div>
          <!-- Status -->
          {#if fulfillment}
            <div class="inv-meta-box">
              <div class="inv-meta-head">
                <Clock class="h-3.5 w-3.5" style="color:var(--c-faint)" />
                <span>Status Distribusi</span>
              </div>
              <div class="inv-status-figures">
                <span class="inv-status-fig" style="color:var(--c-success)">
                  <CheckCircle class="h-3.5 w-3.5" />
                  <strong>{receivedCount}</strong> diterima
                </span>
                <span class="inv-status-fig" style="color:var(--c-warning)">
                  <Package class="h-3.5 w-3.5" />
                  <strong>{pendingCount}</strong> menunggu
                </span>
              </div>
              <div class="inv-progress">
                <ProgressBar
                  value={receivedCount}
                  max={Math.max(totalFulfillment, 1)}
                  color="var(--c-success)"
                />
              </div>
            </div>
          {/if}
        </div>

        <!-- Member Table -->
        <div class="inv-table-wrap">
          <table class="inv-table">
            <thead>
              <tr>
                <th class="inv-th inv-th-check">
                  <input
                    type="checkbox"
                    checked={fulfillment?.pending?.length > 0 &&
                      selectedMembers.size === fulfillment?.pending?.length}
                    onchange={() => {
                      if (selectedMembers.size === fulfillment?.pending?.length)
                        clearSelection();
                      else selectAllPending();
                    }}
                    class="inv-checkbox"
                  />
                </th>
                <th class="inv-th">Nama</th>
                <th class="inv-th inv-th-center">Gender</th>
                <th class="inv-th">Baju</th>
                <th class="inv-th">Family</th>
                <th class="inv-th inv-th-center">Status</th>
              </tr>
            </thead>
            <tbody>
              {#each forecast.details || [] as member}
                {@const isReceived = member.is_equipment_received}
                <tr class="inv-row {isReceived ? 'inv-row-received' : ''}">
                  <td class="inv-td">
                    {#if !isReceived}
                      <input
                        type="checkbox"
                        checked={selectedMembers.has(member.member_id)}
                        onchange={() => toggleMember(member.member_id)}
                        class="inv-checkbox"
                      />
                    {:else}
                      <CheckCircle class="h-4 w-4" style="color:var(--c-success)" />
                    {/if}
                  </td>
                  <td class="inv-td">
                    <div class="inv-member">
                      <Avatar name={member.nama} size={34} />
                      <span class="inv-member-name">{member.nama}</span>
                    </div>
                  </td>
                  <td class="inv-td inv-td-center">
                    <Badge
                      label={member.gender === "male" ? "L" : "P"}
                      tone={member.gender === "male" ? "info" : "danger"}
                    />
                  </td>
                  <td class="inv-td">
                    <select
                      value={member.baju_size || ""}
                      onchange={(e) =>
                        updateBajuSize(
                          member.member_id,
                          /** @type {HTMLSelectElement} */ (e.target).value,
                        )}
                      class="inv-field"
                    >
                      {#each sizes as s}<option value={s}>{s || "-"}</option
                        >{/each}
                    </select>
                  </td>
                  <td class="inv-td">
                    <input
                      type="text"
                      value={member.family_id || ""}
                      onblur={(e) =>
                        updateFamilyId(
                          member.member_id,
                          /** @type {HTMLInputElement} */ (e.target).value,
                        )}
                      placeholder="F00"
                      class="inv-field"
                    />
                  </td>
                  <td class="inv-td inv-td-center">
                    {#if isReceived}
                      <Badge status="Selesai" label="Diterima" dot />
                    {:else}
                      <Badge tone="warning" label="Menunggu" dot />
                    {/if}
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {:else}
        <EmptyState icon={Package} title="Tidak ada data" />
      {/if}
    </Card>

    {#if selectedGroupId}
      <Card style="margin-top:16px">
        <div class="hand-head">
          <div>
            <h3 class="hand-title">Serah Terima (QR)</h3>
            <p class="hand-sub">Pindai QR jamaah untuk mencatat penyerahan perlengkapan & koper. {handoverDone}/{checkpoints.length} lengkap.</p>
          </div>
        </div>

        <div class="hand-scan">
          <div class="hand-toggle">
            <button type="button" class="hand-tog" class:hand-tog-on={scanCheckpoint === 'equipment'} onclick={() => (scanCheckpoint = 'equipment')}>Perlengkapan</button>
            <button type="button" class="hand-tog" class:hand-tog-on={scanCheckpoint === 'luggage'} onclick={() => (scanCheckpoint = 'luggage')}>Koper</button>
          </div>
          <input
            type="text"
            bind:value={scanToken}
            onkeydown={(e) => e.key === 'Enter' && submitScan()}
            placeholder="Pindai / ketik kode QR lalu Enter…"
            class="hand-input"
          />
          <button type="button" class="hand-btn" onclick={submitScan} disabled={isScanning}>Catat</button>
        </div>
        {#if scanMsg}
          <p class="hand-msg" style="color:{scanOk ? 'var(--c-success)' : 'var(--c-danger)'}">{scanMsg}</p>
        {/if}

        {#if checkpoints.length}
          <div class="inv-table-wrap" style="margin-top:12px">
            <table class="inv-table">
              <thead>
                <tr><th>Jamaah</th><th>Perlengkapan</th><th>Koper</th><th>QR</th></tr>
              </thead>
              <tbody>
                {#each checkpoints as m (m.member_id)}
                  <tr>
                    <td>{m.nama}</td>
                    <td>{#if m.is_equipment_received}<span class="hand-ok">✓</span>{:else}<span class="hand-no">—</span>{/if}</td>
                    <td>{#if m.is_luggage_checked}<span class="hand-ok">✓</span>{:else}<span class="hand-no">—</span>{/if}</td>
                    <td><button type="button" class="hand-qrbtn" onclick={() => showQr(m)}>Lihat QR</button></td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </Card>
    {/if}

    {/if}<!-- end tab === "distribusi" -->
  </div>
{/if}

{#if qrMember}
  <div class="hand-modal" role="button" tabindex="-1" onclick={closeQr} onkeydown={(e) => e.key === 'Escape' && closeQr()}>
    <div class="hand-modal-card" role="dialog" tabindex="-1" onclick={(e) => e.stopPropagation()} onkeydown={() => {}}>
      <p class="hand-modal-name">{qrMember.nama}</p>
      {#if qrUrl}
        <img src={qrUrl} alt="QR {qrMember.nama}" width="240" height="240" />
        <p class="hand-modal-tok">{qrMember.handover_token}</p>
      {:else}
        <p class="hand-sub">Memuat QR…</p>
      {/if}
      <button type="button" class="hand-btn" style="margin-top:12px" onclick={closeQr}>Tutup</button>
    </div>
  </div>
{/if}

<style>
  /* QR handover (Phase 4C) */
  .hand-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 14px; }
  .hand-title { font-size: 15px; font-weight: 800; color: var(--c-ink); }
  .hand-sub { font-size: 12.5px; color: var(--c-faint); margin-top: 2px; }
  .hand-scan { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; }
  .hand-toggle { display: inline-flex; gap: 4px; background: var(--c-bg-2); padding: 4px; border-radius: var(--radius); }
  .hand-tog { padding: 7px 12px; font-size: 12.5px; font-weight: 700; border-radius: var(--radius-sm); color: var(--c-muted); }
  .hand-tog-on { background: var(--c-surface); color: var(--c-primary); box-shadow: var(--shadow-sm); }
  .hand-input { flex: 1; min-width: 200px; border: 1px solid var(--c-line); background: var(--c-surface); border-radius: var(--radius); padding: 9px 12px; font-size: 13.5px; color: var(--c-ink); outline: none; }
  .hand-input:focus { border-color: var(--c-primary); box-shadow: 0 0 0 3px var(--c-primary-soft); }
  .hand-btn { background: var(--c-primary); color: #fff; font-weight: 700; font-size: 13px; border-radius: var(--radius); padding: 9px 16px; }
  .hand-btn:disabled { opacity: 0.5; }
  .hand-msg { margin-top: 8px; font-size: 12.5px; font-weight: 600; }
  .hand-ok { color: var(--c-success); font-weight: 800; }
  .hand-no { color: var(--c-faint); }
  .hand-qrbtn { font-size: 12px; font-weight: 700; color: var(--c-primary); }
  .hand-modal { position: fixed; inset: 0; z-index: 50; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.45); }
  .hand-modal-card { background: var(--c-surface); border-radius: var(--radius-lg, 16px); padding: 20px; text-align: center; box-shadow: var(--shadow-lg, 0 10px 40px rgba(0,0,0,0.2)); }
  .hand-modal-name { font-size: 14px; font-weight: 800; color: var(--c-ink); margin-bottom: 12px; }
  .hand-modal-tok { margin-top: 8px; font-family: monospace; font-size: 13px; letter-spacing: 1px; color: var(--c-muted); }

  .inv-page {
    min-height: 100vh;
    background: var(--c-bg);
    padding: 16px;
  }
  @media (min-width: 1024px) {
    .inv-page {
      padding: 32px;
    }
  }

  /* Header actions */
  .inv-select {
    width: 100%;
    border: 1px solid var(--c-line);
    background: var(--c-surface);
    border-radius: var(--radius);
    padding: 9px 14px;
    font-size: 13.5px;
    font-weight: 600;
    color: var(--c-ink-soft);
    outline: none;
    transition: border-color 0.15s, box-shadow 0.15s;
  }
  .inv-select:focus {
    border-color: var(--c-primary);
    box-shadow: 0 0 0 3px var(--c-primary-soft);
  }
  @media (min-width: 640px) {
    .inv-select {
      width: auto;
      min-width: 14rem;
    }
  }
  .inv-icon-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    flex-shrink: 0;
    border: 1px solid var(--c-line);
    border-radius: var(--radius);
    color: var(--c-muted);
    background: var(--c-surface);
    transition: background 0.15s;
  }
  .inv-icon-btn:hover {
    background: var(--c-bg-2);
  }
  .inv-icon-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  /* Stat row */
  .inv-stats {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: var(--gap);
    margin-bottom: var(--gap);
  }
  @media (min-width: 1024px) {
    .inv-stats {
      grid-template-columns: repeat(4, 1fr);
    }
  }

  /* Toolbar */
  .inv-toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 14px;
    padding: 12px 16px;
  }
  .inv-toolbar-count {
    font-size: 12.5px;
    font-weight: 600;
    color: var(--c-muted);
  }
  .inv-toolbar-spacer {
    margin-left: auto;
  }
  .inv-link {
    font-size: 12.5px;
    font-weight: 600;
    color: var(--c-muted);
    background: none;
    border: none;
    cursor: pointer;
  }
  .inv-link:hover {
    text-decoration: underline;
  }
  .inv-link-primary {
    color: var(--c-primary);
    font-weight: 700;
  }

  /* Error */
  .inv-error {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: var(--gap);
    border: 1px solid var(--c-danger);
    background: var(--c-danger-soft);
    border-radius: var(--radius-lg);
    padding: 12px 16px;
    font-size: 13.5px;
  }
  .inv-error-text {
    flex: 1;
    color: var(--c-danger);
    font-weight: 600;
  }
  .inv-error-close {
    color: var(--c-danger);
    background: none;
    border: none;
    cursor: pointer;
    line-height: 0;
  }

  .inv-center {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 64px 0;
  }

  /* Meta row (sizes + status) */
  .inv-meta-row {
    display: flex;
    flex-wrap: wrap;
    gap: 14px;
    padding: 16px;
    border-bottom: 1px solid var(--c-line);
  }
  .inv-meta-box {
    min-width: 180px;
    border: 1px solid var(--c-line);
    background: var(--c-primary-tint);
    border-radius: var(--radius);
    padding: 12px 14px;
  }
  .inv-meta-grow {
    flex: 1;
    min-width: 220px;
  }
  .inv-meta-head {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 8px;
    font-size: 11.5px;
    font-weight: 700;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--c-faint);
  }
  .inv-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }
  .inv-chip {
    background: var(--c-surface);
    border: 1px solid var(--c-line);
    border-radius: var(--radius-sm);
    padding: 2px 8px;
    font-size: 12px;
    font-weight: 600;
    color: var(--c-ink-soft);
  }
  .inv-status-figures {
    display: flex;
    gap: 16px;
    font-size: 12.5px;
  }
  .inv-status-fig {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-weight: 600;
  }
  .inv-status-fig strong {
    font-weight: 800;
  }
  .inv-progress {
    margin-top: 10px;
  }

  /* Table (matches design Table th/td styling) */
  .inv-table-wrap {
    overflow-x: auto;
  }
  .inv-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13.5px;
  }
  .inv-th {
    text-align: left;
    padding: 14px 16px 12px;
    font-size: 11.5px;
    font-weight: 700;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: var(--c-faint);
    white-space: nowrap;
    border-bottom: 1px solid var(--c-line);
  }
  .inv-th-center {
    text-align: center;
  }
  .inv-th-check {
    width: 44px;
  }
  .inv-td {
    padding: calc(var(--row-h, 56px) / 4.2) 16px;
    border-bottom: 1px solid var(--c-line-soft);
    color: var(--c-ink-soft);
    vertical-align: middle;
    white-space: nowrap;
  }
  .inv-td-center {
    text-align: center;
  }
  .inv-row {
    transition: background 0.12s;
  }
  .inv-row:hover {
    background: var(--c-primary-tint);
  }
  .inv-row-received {
    background: var(--c-success-soft);
  }
  .inv-row-received:hover {
    background: var(--c-success-soft);
  }
  .inv-member {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .inv-member-name {
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 14px;
    font-weight: 700;
    color: var(--c-ink);
  }
  .inv-checkbox {
    width: 16px;
    height: 16px;
    border-radius: 4px;
    border: 1px solid var(--c-line);
    accent-color: var(--c-primary);
    cursor: pointer;
  }
  .inv-field {
    width: 100%;
    min-width: 72px;
    border: 1px solid var(--c-line);
    background: var(--c-surface);
    border-radius: var(--radius-sm);
    padding: 6px 8px;
    font-size: 12.5px;
    color: var(--c-ink);
    outline: none;
    transition: border-color 0.15s;
  }
  .inv-field:focus {
    border-color: var(--c-primary);
  }
</style>
