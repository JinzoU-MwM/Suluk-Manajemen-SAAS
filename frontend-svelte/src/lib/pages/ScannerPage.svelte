<script>
  import { onMount } from "svelte";
  import {
    Crown,
    CheckCircle,
    ScanLine,
    Sparkles,
    ShieldCheck,
    FileCheck,
    AlertCircle,
    Lock,
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
  let isPro = $derived(isProOrHigher(localSubscription?.plan));
  let isBlocked = $derived(
    !isPro && localSubscription?.allowed === false,
  );
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

  <!-- AI flow / info panel -->
  <div class="info-grid">
    <Card style="grid-column: span 1;">
      <div style="display:flex;align-items:flex-start;gap:16px;">
        <div class="scan-tile">
          <ScanLine class="h-6 w-6" />
          <div class="scan-beam"></div>
        </div>
        <div style="min-width:0;">
          <div style="display:flex;align-items:center;gap:8px;flex-wrap:wrap;">
            <h2 style="margin:0;font-size:15.5px;font-weight:800;color:var(--c-ink);">Alur Scan Dokumen</h2>
          </div>
          <p style="margin:6px 0 0;font-size:14px;line-height:1.55;color:var(--c-muted);">
            Pilih grup, upload dokumen, review hasil AI, lalu simpan ke grup atau export Excel.
          </p>
          <div class="flow-chips">
            <span class="flow-chip"><FileCheck class="h-3.5 w-3.5" style="color:var(--c-primary)" /> Upload</span>
            <span class="flow-chip"><Sparkles class="h-3.5 w-3.5" style="color:var(--c-primary)" /> Ekstrak AI</span>
            <span class="flow-chip"><ShieldCheck class="h-3.5 w-3.5" style="color:var(--c-primary)" /> Review &amp; Simpan</span>
          </div>
        </div>
      </div>
    </Card>

    <Card>
      <div style="display:flex;align-items:center;gap:8px;">
        <div class="mode-icon"><Sparkles class="h-4 w-4" /></div>
        <p style="margin:0;font-size:11.5px;font-weight:700;text-transform:uppercase;letter-spacing:.04em;color:var(--c-faint);">Mode AI</p>
      </div>
      <p style="margin:8px 0 0;font-size:14px;font-weight:700;color:var(--c-ink);">{cacheModeLabels[processingCacheMode]}</p>
      <p style="margin:4px 0 0;font-size:12px;color:var(--c-muted);">{canUseBypassCacheMode ? "Bypass tersedia untuk Pro." : "Default aman untuk pemrosesan rutin."}</p>
    </Card>
  </div>

  <!-- Subscription Banner -->
  <div style="margin-bottom:1.5rem;">
    <SubscriptionBanner
      subscription={localSubscription}
      onUpgrade={() => (showUpgradeModal = true)}
    />
  </div>

  <!-- Main Content -->
  {#if isBlocked}
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
    <!-- Group Selector -->
    <Card style="margin-bottom:1.5rem;">
      <GroupSelector
        bind:selectedGroup
        onGroupSelect={(g) => (selectedGroup = g)}
        onViewGroup={viewGroupData}
        isPro={isPro && localSubscription?.status === "active"}
      />
    </Card>

    <!-- Success Banner -->
    {#if groupSaveSuccess}
      <div class="success-banner">
        <CheckCircle class="h-5 w-5" style="color:var(--c-success);flex-shrink:0;" />
        <span style="font-size:14px;color:var(--c-primary-deep);">{groupSaveSuccess}</span>
      </div>
    {/if}

    <!-- Advanced OCR Settings -->
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

    <FileUpload
      bind:files
      {isProcessing}
      {errorMessage}
      onProcess={() => processDocuments()}
      {progress}
    />
  {/if}

  <!-- Retry Banner -->
  {#if failedFileNames.length > 0 && !isProcessing && !showModal}
    <div class="retry-banner">
      <div style="display:flex;align-items:center;gap:8px;min-width:0;">
        <AlertCircle class="h-5 w-5" style="color:var(--c-danger);flex-shrink:0;" />
        <span style="font-size:14px;color:var(--c-danger);">
          <strong>{failedFileNames.length}</strong> file gagal: {failedFileNames.join(", ")}
        </span>
      </div>
      <Button variant="danger" onclick={retryFailed}>Coba Lagi</Button>
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

  /* info grid */
  .info-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--gap, 1.25rem);
    margin-bottom: 1.5rem;
  }
  @media (min-width: 1024px) {
    .info-grid {
      grid-template-columns: 2fr 1fr;
    }
  }

  /* scan tile with animated beam */
  .scan-tile {
    position: relative;
    flex-shrink: 0;
    width: 48px;
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    border-radius: var(--radius, 12px);
    background: var(--c-primary-deep);
    color: #fff;
  }
  .scan-beam {
    pointer-events: none;
    position: absolute;
    left: 0;
    right: 0;
    top: 0;
    height: 2px;
    background: var(--c-accent);
    box-shadow: 0 0 10px 2px var(--c-accent);
    animation: scan-sweep 1.8s ease-in-out infinite;
  }
  @keyframes scan-sweep {
    0% { top: 0; opacity: 0; }
    15% { opacity: 1; }
    85% { opacity: 1; }
    100% { top: 100%; opacity: 0; }
  }
  @media (prefers-reduced-motion: reduce) {
    .scan-beam { animation: none; opacity: 0; }
  }

  .flow-chips {
    margin-top: 12px;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .flow-chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 5px 10px;
    font-size: 12px;
    font-weight: 600;
    color: var(--c-muted);
    background: var(--c-bg-2);
    border: 1px solid var(--c-line);
    border-radius: var(--radius-sm, 8px);
  }

  .mode-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: var(--radius-sm, 8px);
    background: var(--c-accent-soft);
    color: var(--c-accent);
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

  .success-banner {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 1.25rem;
    padding: 1rem;
    border: 1px solid var(--c-primary-soft);
    background: var(--c-primary-soft);
    border-radius: var(--radius-lg, 16px);
  }

  /* advanced settings */
  .adv-settings {
    margin-bottom: 1.5rem;
    padding: 1rem 1.25rem;
    background: var(--c-surface);
    border: 1px solid var(--c-line);
    border-radius: var(--radius-lg, 16px);
    box-shadow: var(--shadow-sm);
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
    background: var(--c-bg);
    color: var(--c-ink-soft);
    padding: 8px 12px;
    font-size: 14px;
    border-radius: var(--radius, 12px);
    outline: none;
    transition: border-color .15s, box-shadow .15s;
  }
  .adv-row select:focus {
    border-color: var(--c-primary);
    background: var(--c-surface);
    box-shadow: 0 0 0 3px var(--c-primary-soft);
  }
  .adv-hint {
    margin: 8px 0 0;
    font-size: 12px;
    color: var(--c-muted);
  }

  .retry-banner {
    margin-top: 1.25rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
    padding: 1rem;
    border: 1px solid var(--c-danger-soft);
    background: var(--c-danger-soft);
    border-radius: var(--radius-lg, 16px);
  }
</style>
