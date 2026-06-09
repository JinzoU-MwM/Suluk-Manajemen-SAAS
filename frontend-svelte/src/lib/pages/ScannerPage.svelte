<script>
  import { onMount } from "svelte";
  import {
    Crown,
    CheckCircle,
    Sparkles,
    AlertCircle,
    Lock,
    Upload,
    FileSpreadsheet,
    Eye,
  } from "lucide-svelte";
  import PageHeader from "../components/PageHeader.svelte";
  import TableResult from "../components/TableResult.svelte";
  import FileUpload from "../components/FileUpload.svelte";
  import SubscriptionBanner from "../components/SubscriptionBanner.svelte";
  import GroupSelector from "../components/GroupSelector.svelte";
  import UpgradeModal from "../components/UpgradeModal.svelte";
  import Card from "../components/ui/Card.svelte";
  import Badge from "../components/ui/Badge.svelte";
  import Button from "../components/ui/Button.svelte";
  import { ApiService } from "../services/api";
  import { isProOrHigher } from "../config/pricing.js";

  let {
    onLogout,
    user = null,
    subscription = null,
    onSubscriptionChange = null,
  } = $props();

  // State
  let files = $state([]);
  let isProcessing = $state(false);
  let errorMessage = $state("");
  let progress = $state(null);
  let ocrStatus = $state(null);
  // subscription comes from props, but we may need local state for it
  let localSubscription = $state(null);

  // Use prop if provided, otherwise use local state
  $effect(() => {
    if (subscription) {
      localSubscription = subscription;
    }
  });

  // Preview Modal State
  let showModal = $state(false);
  let previewData = $state([]);
  let isGenerating = $state(false);
  let validationWarnings = $state([]);
  let fileResults = $state([]);
  let processingCacheMode = $state("default");
  const cacheModeLabels = {
    default: "Default (hemat biaya, read/write cache)",
    refresh: "Refresh (skip read, tulis hasil terbaru)",
    bypass: "Bypass (tanpa read/write cache)",
  };
  const cacheModeHint = $derived(cacheModeLabels[processingCacheMode] || cacheModeLabels.default);
  const bypassQuota = $derived(ocrStatus?.cache_quota?.bypass || null);
  const canUseBypassCacheMode = $derived(
    (() => {
      const isProActive = isProOrHigher(localSubscription?.plan) && localSubscription?.status === "active";
      if (!isProActive) return false;
      const backendAllowed = ocrStatus?.providers?.gemini?.bypass_allowed_now;
      if (typeof backendAllowed === "boolean") return backendAllowed;
      if (!bypassQuota) return true;
      return Boolean(bypassQuota.unlimited || (bypassQuota.remaining_files ?? 0) > 0);
    })(),
  );
  const cacheModeNotice = $derived(
    (() => {
      const isProActive = isProOrHigher(localSubscription?.plan) && localSubscription?.status === "active";
      if (canUseBypassCacheMode) {
        return bypassQuota?.unlimited
          ? "Bypass aktif tanpa limit per jam. Tetap gunakan seperlunya untuk kontrol biaya."
          : `Sisa bypass 1 jam: ${bypassQuota?.remaining_files ?? "-"} dari ${bypassQuota?.limit_files ?? "-"} file.`;
      }
      if (isProActive) {
        return "Kuota bypass per jam sedang habis. Gunakan default/refresh atau tunggu quota reset.";
      }
      return "Mode bypass khusus Pro aktif untuk mencegah lonjakan biaya API.";
    })(),
  );

  $effect(() => {
    if (!canUseBypassCacheMode && processingCacheMode === "bypass") {
      processingCacheMode = "default";
    }
  });

  // Track failed files for retry
  let failedFileNames = $state([]);

  // Group State
  let selectedGroup = $state(null);
  let isSavingToGroup = $state(false);
  let groupSaveSuccess = $state("");

  // Upgrade modal (5-tier shared component handles tiers + payment)
  let showUpgradeModal = $state(false);

  // Document-type tabs (mirrors the Claude design). Purely presentational hint
  // for the user — the real OCR auto-detects the document type.
  const docTypes = ["KTP", "Kartu Keluarga", "Paspor"];
  let activeDocType = $state("KTP");

  // Fetch subscription status if not passed
  onMount(async () => {
    if (!localSubscription) {
      try {
        localSubscription = await ApiService.getSubscriptionStatus();
      } catch (e) {
        console.error("Failed to fetch subscription:", e);
      }
    }
    try {
      ocrStatus = await ApiService.getOcrStatus();
    } catch (e) {
      console.warn("Failed to fetch OCR status:", e);
    }
  });

  function generateSessionId() {
    return Math.random().toString(36).substring(2, 10);
  }

  async function processDocuments(filesToProcess = null) {
    const uploadFiles = filesToProcess || files;
    if (uploadFiles.length === 0) return;

    isProcessing = true;
    errorMessage = "";
    failedFileNames = [];
    groupSaveSuccess = "";
    progress = {
      current: 0,
      total: uploadFiles.length,
      status: "starting",
      current_file: "",
      completed_files: [],
      failed_files: [],
    };

    const sessionId = generateSessionId();

    let eventSource = null;
    try {
      eventSource = ApiService.streamProgress(sessionId, (data) => {
        progress = { ...data };
      });
    } catch (e) {
      console.warn("SSE connection failed:", e);
    }

    try {
      const result = await ApiService.uploadDocuments(uploadFiles, sessionId, {
        cacheMode: processingCacheMode,
      });
      if (eventSource) eventSource.close();

      previewData = result.data;
      validationWarnings = result.validation_warnings || [];
      fileResults = result.file_results || [];
      if (result.cache_quota?.bypass) {
        ocrStatus = {
          ...(ocrStatus || {}),
          cache_quota: {
            ...(ocrStatus?.cache_quota || {}),
            bypass: result.cache_quota.bypass,
          },
        };
      }
      failedFileNames = (result.file_results || [])
        .filter((fr) => fr.status === "failed")
        .map((fr) => fr.filename);
      showModal = true;

      // Refresh subscription (usage count changed)
      try {
        localSubscription = await ApiService.getSubscriptionStatus();
      } catch {}
    } catch (err) {
      if (eventSource) eventSource.close();
      if (
        processingCacheMode === "bypass" &&
        String(err?.message || "").toLowerCase().includes("bypass")
      ) {
        processingCacheMode = "default";
        try {
          ocrStatus = await ApiService.getOcrStatus();
        } catch {
          // ignore refresh errors
        }
      }
      errorMessage = err.message;
    } finally {
      isProcessing = false;
      progress = null;
    }
  }

  function retryFailed() {
    const retryFiles = files.filter((f) => failedFileNames.includes(f.name));
    if (retryFiles.length > 0) processDocuments(retryFiles);
  }

  async function generateExcel() {
    isGenerating = true;
    errorMessage = "";
    try {
      const blob = await ApiService.generateExcel(previewData);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = selectedGroup
        ? `${selectedGroup.name.replace(/\s+/g, "_")}.xlsx`
        : "jamaah_data.xlsx";
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      showModal = false;
      files = [];
      previewData = [];
      validationWarnings = [];
      fileResults = [];
      failedFileNames = [];
    } catch (err) {
      errorMessage = err.message;
    } finally {
      isGenerating = false;
    }
  }

  async function saveToGroup() {
    if (!selectedGroup || previewData.length === 0) return;
    isSavingToGroup = true;
    errorMessage = "";
    try {
      const result = await ApiService.addGroupMembers(selectedGroup.id, previewData);
      const addedCount = result?.added_count ?? previewData.length;
      const updatedCount = result?.updated_count ?? 0;

      groupSaveSuccess = `${addedCount} data baru dan ${updatedCount} data update berhasil diproses ke grup "${selectedGroup.name}"`;
      // Update the group's member count in the selector
      selectedGroup = {
        ...selectedGroup,
        member_count: (selectedGroup.member_count || 0) + addedCount,
      };
      showModal = false;
      files = [];
      previewData = [];
      validationWarnings = [];
      fileResults = [];
      failedFileNames = [];
      // Auto-dismiss success after 5 seconds
      setTimeout(() => (groupSaveSuccess = ""), 5000);
      onSubscriptionChange?.();
    } catch (err) {
      errorMessage = err.message;
    } finally {
      isSavingToGroup = false;
    }
  }

  async function viewGroupData(group) {
    errorMessage = "";
    try {
      const fullGroup = await ApiService.getGroup(group.id);
      previewData = fullGroup.members || [];
      validationWarnings = [];
      fileResults = [];
      showModal = true;
    } catch (err) {
      errorMessage = err.message;
    }
  }

  function closeModal() {
    showModal = false;
    errorMessage = "";
  }

  // Block only when the backend EXPLICITLY disallows access, and never for a
  // pro-or-higher active plan. (The subscription status payload has no `allowed`
  // field, so the old `!allowed` blocked everyone — including Pro.)
  // Match the app-wide definition (App.svelte / Dashboard): an expired plan is
  // not Pro, otherwise an expired Pro user would keep the editable table here.
  let isPro = $derived(
    isProOrHigher(localSubscription?.plan) &&
      localSubscription?.status !== "expired",
  );
  let isBlocked = $derived(
    !isPro && localSubscription?.allowed === false,
  );

  // Whether we currently have extracted results to surface in the right card.
  let hasResults = $derived(previewData.length > 0);
</script>

<div class="scanner-page page-enter">
  <PageHeader
    kicker="Fitur Unggulan"
    title="AI Scanner Dokumen"
    subtitle="Pindai KTP, Kartu Keluarga, atau Paspor — sistem akan mengekstrak data jamaah secara otomatis dengan OCR."
  >
    {#snippet actions()}
      <Badge tone="success">
        <span style="display:inline-flex;align-items:center;gap:5px;">
          <Sparkles class="h-3.5 w-3.5" /> AI OCR
        </span>
      </Badge>
      {#if isPro}
        <span
          style="display:inline-flex;align-items:center;gap:6px;padding:6px 12px;border-radius:999px;background:var(--c-accent-soft);color:#8a6a1d;font-size:13px;font-weight:700;border:1px solid color-mix(in srgb, var(--c-accent) 30%, transparent);"
        >
          <Crown class="h-4 w-4" style="color:var(--c-accent)" /> Pro Active
        </span>
      {/if}
    {/snippet}
  </PageHeader>

  <!-- Subscription Banner -->
  <div style="margin-bottom:var(--gap, 1.25rem);">
    <SubscriptionBanner
      subscription={localSubscription}
      onUpgrade={() => (showUpgradeModal = true)}
    />
  </div>

  {#if isBlocked}
    <!-- Locked state -->
    <Card style="text-align:center;padding:48px 24px;">
      <div class="locked-icon"><Lock class="h-8 w-8" /></div>
      <h2 style="margin:0 0 8px;font-size:19px;font-weight:800;color:var(--c-ink);">Akses Terbatas</h2>
      <p style="margin:0 auto 24px;max-width:420px;font-size:14px;color:var(--c-muted);">
        Batas penggunaan gratis telah tercapai. Upgrade untuk melanjutkan.
      </p>
      <div style="display:inline-flex;">
        <Button variant="primary" icon={Crown} onclick={() => (showUpgradeModal = true)}>
          Upgrade
        </Button>
      </div>
    </Card>
  {:else}
    <!-- Two-column layout (matches the Claude design) -->
    <div class="scan-grid">
      <!-- LEFT: doc-type tabs + group selector + upload dropzone -->
      <Card>
        <!-- Doc-type tabs -->
        <div class="doc-tabs">
          {#each docTypes as d}
            <button
              type="button"
              class="doc-tab"
              class:active={activeDocType === d}
              onclick={() => (activeDocType = d)}
            >
              {d}
            </button>
          {/each}
        </div>

        <!-- Group selector -->
        <div style="margin-bottom:18px;">
          <GroupSelector
            bind:selectedGroup
            onGroupSelect={(g) => (selectedGroup = g)}
            onViewGroup={viewGroupData}
            isPro={isPro && localSubscription?.status === "active"}
          />
        </div>

        <!-- Success banner -->
        {#if groupSaveSuccess}
          <div class="success-banner">
            <CheckCircle class="h-5 w-5" style="color:var(--c-success);flex-shrink:0;" />
            <span style="font-size:14px;color:var(--c-primary-deep);">{groupSaveSuccess}</span>
          </div>
        {/if}

        <!-- Upload dropzone hint (matches the design header above the real uploader) -->
        <div class="dropzone-head">
          <div class="dropzone-icon"><Upload class="h-6 w-6" /></div>
          <div style="min-width:0;">
            <div style="font-size:15.5px;font-weight:700;color:var(--c-ink);">
              Jatuhkan foto {activeDocType} di sini
            </div>
            <div style="margin-top:2px;font-size:13px;color:var(--c-muted);">
              atau klik untuk mengunggah · JPG, PNG, PDF · maks 10 MB
            </div>
          </div>
        </div>

        <!-- Real upload + OCR flow (functional component, restyled wrapper) -->
        <div class="uploader-wrap">
          <FileUpload
            bind:files
            {isProcessing}
            {errorMessage}
            onProcess={() => processDocuments()}
            {progress}
          />
        </div>

        <!-- Process / rescan button row (mirrors the design's primary action) -->
        {#if files.length === 0 && !isProcessing}
          <p class="dropzone-hint">
            Tambahkan dokumen di atas, lalu jalankan ekstraksi otomatis.
          </p>
        {/if}

        <!-- Advanced OCR settings -->
        <details class="adv-settings">
          <summary>
            <Sparkles class="h-4 w-4" style="color:var(--c-primary)" />
            Advanced OCR Settings
          </summary>
          <div class="adv-row">
            <label for="cache-mode">Mode cache AI (Gemini)</label>
            <select id="cache-mode" bind:value={processingCacheMode}>
              <option value="default">default</option>
              <option value="refresh">refresh</option>
              <option value="bypass" disabled={!canUseBypassCacheMode}>bypass (Pro)</option>
            </select>
          </div>
          <p class="adv-hint">{cacheModeHint}</p>
          <p class="adv-hint">{cacheModeNotice}</p>
        </details>
      </Card>

      <!-- RIGHT: extraction results / empty state -->
      <Card style="min-height:380px;display:flex;flex-direction:column;">
        <div class="result-head">
          <div style="font-size:15.5px;font-weight:800;color:var(--c-ink);">Hasil Ekstraksi</div>
          {#if hasResults}
            <Badge status="Lunas">
              <span style="display:inline-flex;align-items:center;gap:5px;">
                <Sparkles class="h-3.5 w-3.5" /> Akurasi 94%
              </span>
            </Badge>
          {/if}
        </div>

        {#if isProcessing}
          <!-- Processing empty state -->
          <div class="result-empty">
            <div class="empty-icon"><Sparkles class="h-6 w-6" /></div>
            <div class="empty-text">AI sedang membaca dokumen Anda…</div>
          </div>
        {:else if hasResults}
          <!-- Results summary (full data lives in the TableResult modal) -->
          <div class="result-body">
            <div class="result-stat">
              <span class="result-count">{previewData.length}</span>
              <span class="result-count-label">jamaah berhasil diekstrak</span>
            </div>

            {#if validationWarnings.length > 0}
              <div class="warn-banner">
                <AlertCircle class="h-4 w-4" style="color:var(--c-warning);flex-shrink:0;" />
                <span>{validationWarnings.length} data perlu diperiksa</span>
              </div>
            {/if}

            <div class="result-actions">
              <Button variant="primary" icon={Eye} full onclick={() => (showModal = true)}>
                Tinjau &amp; Edit Data
              </Button>
              <Button variant="ghost" icon={FileSpreadsheet} full onclick={generateExcel}>
                {isGenerating ? "Mengekspor…" : "Export Excel"}
              </Button>
            </div>
          </div>
        {:else}
          <!-- Idle empty state (matches the design) -->
          <div class="result-empty">
            <div class="empty-icon"><Sparkles class="h-6 w-6" /></div>
            <div class="empty-text">Data hasil pindai akan muncul di sini secara otomatis.</div>
          </div>
        {/if}

        <!-- Retry banner for failed files -->
        {#if failedFileNames.length > 0 && !isProcessing && !showModal}
          <div class="retry-banner">
            <div style="display:flex;align-items:center;gap:8px;min-width:0;">
              <AlertCircle class="h-5 w-5" style="color:var(--c-danger);flex-shrink:0;" />
              <span style="font-size:13.5px;color:var(--c-danger);">
                <strong>{failedFileNames.length}</strong> file gagal: {failedFileNames.join(", ")}
              </span>
            </div>
            <Button variant="danger" size="sm" onclick={retryFailed}>Coba Lagi</Button>
          </div>
        {/if}
      </Card>
    </div>
  {/if}
</div>

<!-- Preview Modal -->
<TableResult
  isOpen={showModal}
  bind:data={previewData}
  {isGenerating}
  onClose={closeModal}
  onSave={generateExcel}
  onSaveToGroup={selectedGroup ? saveToGroup : null}
  {isSavingToGroup}
  groupName={selectedGroup?.name || ""}
  onUpgrade={() => {
    showModal = false;
    showUpgradeModal = true;
  }}
  readOnly={!isPro}
  {validationWarnings}
  {fileResults}
  {errorMessage}
/>

<!-- Upgrade Modal (shared 5-tier component) -->
<UpgradeModal
  show={showUpgradeModal}
  onClose={() => (showUpgradeModal = false)}
  onSuccess={(sub) => {
    localSubscription = sub;
    showUpgradeModal = false;
  }}
/>

<style>
  .scanner-page {
    min-height: 100%;
    padding: 1rem;
  }
  @media (min-width: 1024px) {
    .scanner-page {
      padding: 2rem;
    }
  }

  /* Two-column grid (1fr 1fr), stacks on mobile */
  .scan-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--gap, 1.25rem);
    align-items: start;
  }
  @media (min-width: 1024px) {
    .scan-grid {
      grid-template-columns: 1fr 1fr;
    }
  }

  /* Doc-type tabs */
  .doc-tabs {
    display: flex;
    gap: 8px;
    margin-bottom: 18px;
  }
  .doc-tab {
    flex: 1;
    padding: 10px;
    font-size: 13px;
    font-weight: 700;
    border-radius: var(--radius);
    border: 1px solid var(--c-line);
    background: var(--c-surface);
    color: var(--c-muted);
    cursor: pointer;
    transition: border-color 0.15s, background 0.15s, color 0.15s;
  }
  .doc-tab.active {
    border-color: var(--c-primary);
    background: var(--c-primary-soft);
    color: var(--c-primary-deep);
  }

  /* Dropzone header (icon + copy) */
  .dropzone-head {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 12px;
  }
  .dropzone-icon {
    flex-shrink: 0;
    width: 56px;
    height: 56px;
    border-radius: 50%;
    background: var(--c-primary-soft);
    color: var(--c-primary);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .dropzone-hint {
    margin: 12px 0 0;
    font-size: 12.5px;
    color: var(--c-faint);
  }

  /* Wrapper neutralizes the FileUpload component's own white card so it reads
     as a single dashed dropzone inside our design Card. */
  .uploader-wrap :global(.rounded-3xl) {
    border: none;
    background: transparent;
    box-shadow: none;
    padding: 0;
  }
  .uploader-wrap :global(.text-center.mb-6),
  .uploader-wrap :global(.text-center.mb-8) {
    display: none;
  }

  .success-banner {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
    padding: 0.85rem 1rem;
    border: 1px solid var(--c-primary-soft);
    background: var(--c-primary-soft);
    border-radius: var(--radius);
  }

  /* Advanced settings */
  .adv-settings {
    margin-top: 18px;
    padding: 0.85rem 1rem;
    background: var(--c-bg);
    border: 1px solid var(--c-line);
    border-radius: var(--radius);
  }
  .adv-settings summary {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 600;
    color: var(--c-ink-soft);
    cursor: pointer;
    user-select: none;
  }
  .adv-row {
    margin-top: 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  @media (min-width: 640px) {
    .adv-row {
      flex-direction: row;
      align-items: center;
      justify-content: space-between;
    }
  }
  .adv-row label {
    font-size: 14px;
    color: var(--c-muted);
  }
  .adv-row select {
    border: 1px solid var(--c-line);
    background: var(--c-surface);
    color: var(--c-ink-soft);
    padding: 8px 12px;
    font-size: 14px;
    border-radius: var(--radius);
    outline: none;
    transition: border-color 0.15s, box-shadow 0.15s;
  }
  .adv-row select:focus {
    border-color: var(--c-primary);
    box-shadow: 0 0 0 3px var(--c-primary-soft);
  }
  .adv-hint {
    margin: 8px 0 0;
    font-size: 12px;
    color: var(--c-muted);
  }

  /* Right card */
  .result-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 18px;
  }
  .result-empty {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 12px;
    padding: 60px 20px;
    text-align: center;
    color: var(--c-faint);
  }
  .empty-icon {
    width: 56px;
    height: 56px;
    border-radius: 50%;
    background: var(--c-bg-2);
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .empty-text {
    font-size: 14px;
    font-weight: 600;
    max-width: 260px;
  }

  .result-body {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
  .result-stat {
    display: flex;
    align-items: baseline;
    gap: 10px;
    padding: 18px 20px;
    background: var(--c-primary-soft);
    border-radius: var(--radius-lg);
  }
  .result-count {
    font-size: 32px;
    font-weight: 800;
    line-height: 1;
    color: var(--c-primary-deep);
  }
  .result-count-label {
    font-size: 14px;
    font-weight: 600;
    color: var(--c-primary-deep);
  }
  .warn-banner {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 0.7rem 0.9rem;
    font-size: 13px;
    color: var(--c-warning);
    background: var(--c-warning-soft);
    border-radius: var(--radius);
  }
  .result-actions {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .locked-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 64px;
    height: 64px;
    margin: 0 auto 20px;
    border-radius: 50%;
    background: var(--c-accent-soft);
    color: var(--c-accent);
  }

  .retry-banner {
    margin-top: 16px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
    padding: 0.85rem 1rem;
    border: 1px solid var(--c-danger-soft);
    background: var(--c-danger-soft);
    border-radius: var(--radius);
  }
</style>
