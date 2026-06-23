<script>
  import { onMount } from "svelte";
  import {
    Crown,
    CheckCircle,
    Sparkles,
    AlertCircle,
    Lock,
    UploadCloud,
    FileSpreadsheet,
    Table,
    Users,
    ScanLine,
    FileText,
    Image as ImageIcon,
    ShieldCheck,
    Clipboard,
    X,
  } from "lucide-svelte";
  import PageHeader from "../components/PageHeader.svelte";
  import TableResult from "../components/TableResult.svelte";
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

  // Upload + processing state
  let files = $state([]);
  let dragOver = $state(false);
  let isProcessing = $state(false);
  let errorMessage = $state("");
  let localSubscription = $state(null);
  $effect(() => {
    if (subscription) localSubscription = subscription;
  });

  // Result / preview state
  let showModal = $state(false);
  let previewData = $state([]);
  let isGenerating = $state(false);
  let validationWarnings = $state([]);
  let fileResults = $state([]);
  let failedFileNames = $state([]);

  // Group + upgrade
  let selectedGroup = $state(null);
  let isSavingToGroup = $state(false);
  let groupSaveSuccess = $state("");
  let showUpgradeModal = $state(false);

  onMount(async () => {
    if (!localSubscription) {
      try {
        localSubscription = await ApiService.getSubscriptionStatus();
      } catch (e) {
        console.error("Failed to fetch subscription:", e);
      }
    }
  });

  // ---- file intake (drag / click / paste) ----
  function addFiles(list) {
    const valid = Array.from(list).filter(
      (f) => f.type.startsWith("image/") || f.type === "application/pdf",
    );
    if (valid.length) files = [...files, ...valid];
  }
  function onDrop(e) {
    e.preventDefault();
    dragOver = false;
    if (e.dataTransfer?.files) addFiles(e.dataTransfer.files);
  }
  function onPick(e) {
    if (e.target.files) addFiles(e.target.files);
    e.target.value = "";
  }
  function openPicker() {
    if (!isProcessing) document.getElementById("scannerFileInput")?.click();
  }
  function removeFile(i) {
    files = files.filter((_, idx) => idx !== i);
  }
  function clearFiles() {
    files = [];
  }
  async function handlePaste() {
    try {
      const items = await navigator.clipboard.read();
      for (const item of items) {
        const t = item.types.find((x) => x.startsWith("image/"));
        if (t) {
          const blob = await item.getType(t);
          files = [
            ...files,
            new File([blob], `tempel-${Date.now()}.png`, { type: t }),
          ];
        }
      }
    } catch (err) {
      console.error("Clipboard read failed", err);
    }
  }
  onMount(() => {
    const onWindowPaste = (e) => {
      if (e.clipboardData?.files?.length) addFiles(e.clipboardData.files);
    };
    window.addEventListener("paste", onWindowPaste);
    return () => window.removeEventListener("paste", onWindowPaste);
  });

  // A policy PDF is the one file that enriches insurance columns — flag it so the
  // operator can see it was recognised as a POLIS (display hint only; the backend
  // does the authoritative detection).
  function looksLikePolicy(f) {
    return f.type === "application/pdf" && /pol(is|icy)|asuransi/i.test(f.name);
  }
  function fileIcon(f) {
    if (looksLikePolicy(f)) return ShieldCheck;
    return f.type === "application/pdf" ? FileText : ImageIcon;
  }
  function fmtSize(b) {
    if (b < 1024) return b + " B";
    if (b < 1048576) return (b / 1024).toFixed(0) + " KB";
    return (b / 1048576).toFixed(1) + " MB";
  }
  function sessionId() {
    return Math.random().toString(36).slice(2, 10);
  }

  async function processDocuments(filesToProcess = null) {
    const uploadFiles = filesToProcess || files;
    if (uploadFiles.length === 0) return;

    isProcessing = true;
    errorMessage = "";
    failedFileNames = [];
    groupSaveSuccess = "";

    try {
      const result = await ApiService.uploadDocuments(uploadFiles, sessionId(), {
        cacheMode: "default",
      });
      previewData = result.data || [];
      validationWarnings = result.validation_warnings || [];
      fileResults = result.file_results || [];
      failedFileNames = fileResults
        .filter((fr) => fr.status === "failed")
        .map((fr) => fr.filename);
      showModal = true;
      try {
        localSubscription = await ApiService.getSubscriptionStatus();
      } catch {}
    } catch (err) {
      errorMessage = err.message;
    } finally {
      isProcessing = false;
    }
  }

  function retryFailed() {
    const retry = files.filter((f) => failedFileNames.includes(f.name));
    if (retry.length) processDocuments(retry);
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
      resetAfterExport();
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
      const added = result?.added_count ?? previewData.length;
      const updated = result?.updated_count ?? 0;
      groupSaveSuccess = `${added} data baru dan ${updated} update tersimpan ke grup "${selectedGroup.name}".`;
      selectedGroup = {
        ...selectedGroup,
        member_count: (selectedGroup.member_count || 0) + added,
      };
      resetAfterExport();
      setTimeout(() => (groupSaveSuccess = ""), 5000);
      onSubscriptionChange?.();
    } catch (err) {
      errorMessage = err.message;
    } finally {
      isSavingToGroup = false;
    }
  }

  function resetAfterExport() {
    showModal = false;
    files = [];
    previewData = [];
    validationWarnings = [];
    fileResults = [];
    failedFileNames = [];
  }

  async function viewGroupData(group) {
    errorMessage = "";
    try {
      const full = await ApiService.getGroup(group.id);
      previewData = full.members || [];
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

  // Gating: Pro keeps the editable table; only an explicit backend block locks.
  let isPro = $derived(
    isProOrHigher(localSubscription?.plan) &&
      localSubscription?.status !== "expired",
  );
  let isBlocked = $derived(!isPro && localSubscription?.allowed === false);

  let hasResults = $derived(previewData.length > 0);
  let polisCount = $derived(
    fileResults.filter((fr) => fr.doc_type === "polis" && fr.status === "completed")
      .length,
  );
  let scanLabel = $derived(
    isProcessing
      ? "Memindai…"
      : `Pindai${files.length ? " " + files.length : ""} dokumen`,
  );
</script>

<div class="scanner-page page-enter">
  <PageHeader
    kicker="Fitur Unggulan"
    title="AI Scanner Dokumen"
    subtitle="Unggah dokumen jamaah dan polis asuransi sekaligus — AI mengisi data Siskopatuh otomatis."
  >
    {#snippet actions()}
      <Badge tone="success">
        <span style="display:inline-flex;align-items:center;gap:5px;">
          <Sparkles class="h-3.5 w-3.5" /> AI OCR
        </span>
      </Badge>
      {#if isPro}
        <span class="pro-pill">
          <Crown class="h-4 w-4" style="color:var(--c-accent)" /> Pro Active
        </span>
      {/if}
    {/snippet}
  </PageHeader>

  <div style="margin-bottom:var(--gap, 1.25rem);">
    <SubscriptionBanner
      subscription={localSubscription}
      onUpgrade={() => (showUpgradeModal = true)}
    />
  </div>

  {#if isBlocked}
    <Card style="text-align:center;padding:48px 24px;">
      <div class="locked-icon"><Lock class="h-8 w-8" /></div>
      <h2 style="margin:0 0 8px;font-size:19px;font-weight:800;color:var(--c-ink);">
        Akses Terbatas
      </h2>
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
    <div class="flow">
      {#if groupSaveSuccess}
        <div class="success-banner">
          <CheckCircle class="h-5 w-5" style="color:var(--c-success);flex-shrink:0;" />
          <span>{groupSaveSuccess}</span>
        </div>
      {/if}

      <!-- Step 1 — unggah & pindai -->
      <Card>
        <div
          class="dropzone"
          class:drag={dragOver}
          class:scanning={isProcessing}
          role="button"
          tabindex="0"
          onclick={openPicker}
          onkeydown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              openPicker();
            }
          }}
          ondragover={(e) => {
            e.preventDefault();
            dragOver = true;
          }}
          ondragleave={() => (dragOver = false)}
          ondrop={onDrop}
        >
          <input
            id="scannerFileInput"
            type="file"
            multiple
            accept="image/*,.pdf"
            class="hidden-input"
            onchange={onPick}
          />
          <div class="dz-icon">
            {#if isProcessing}
              <ScanLine class="h-6 w-6" />
            {:else}
              <UploadCloud class="h-6 w-6" />
            {/if}
          </div>
          <div class="dz-title">
            {#if isProcessing}
              AI sedang membaca {files.length} dokumen…
            {:else}
              Tarik &amp; lepas dokumen di sini
            {/if}
          </div>
          <div class="dz-sub">
            {#if isProcessing}
              Mohon tunggu sebentar.
            {:else}
              atau klik untuk pilih · KTP, Paspor, Kartu Keluarga, dan PDF polis
              asuransi
            {/if}
          </div>

          {#if isProcessing}<span class="scanline"></span>{/if}

          {#if !isProcessing}
            <button
              type="button"
              class="paste-btn"
              onclick={(e) => {
                e.stopPropagation();
                handlePaste();
              }}
            >
              <Clipboard class="h-3.5 w-3.5" />
              <span>Tempel</span>
            </button>
          {/if}
        </div>

        <!-- Policy nudge -->
        <div class="polis-hint">
          <ShieldCheck class="h-4 w-4" style="color:var(--c-accent);flex-shrink:0;" />
          <span>
            Sertakan <strong>PDF polis asuransi</strong> — kolom Asuransi &amp; No
            Polis terisi otomatis per jamaah (dicocokkan lewat nomor paspor).
          </span>
        </div>

        {#if errorMessage}
          <div class="error-banner">
            <AlertCircle class="h-5 w-5" style="flex-shrink:0;" />
            <span>{errorMessage}</span>
          </div>
        {/if}

        <!-- Selected files -->
        {#if files.length > 0}
          <div class="files-head">
            <span>{files.length} file siap dipindai</span>
            {#if !isProcessing}
              <button type="button" class="link-danger" onclick={clearFiles}>
                Hapus semua
              </button>
            {/if}
          </div>
          <div class="chips">
            {#each files as file, i}
              {@const Ico = fileIcon(file)}
              <span class="chip" class:polis={looksLikePolicy(file)}>
                <Ico class="h-4 w-4" />
                <span class="chip-name" title={file.name}>{file.name}</span>
                <span class="chip-size">{fmtSize(file.size)}</span>
                {#if looksLikePolicy(file)}<span class="chip-badge">polis</span>{/if}
                {#if !isProcessing}
                  <button
                    type="button"
                    class="chip-x"
                    aria-label="Hapus {file.name}"
                    onclick={() => removeFile(i)}
                  >
                    <X class="h-3.5 w-3.5" />
                  </button>
                {/if}
              </span>
            {/each}
          </div>
        {/if}

        <!-- Group (optional) + scan action -->
        <div class="action-row">
          <div class="group-inline">
            <span class="group-label"><Users class="h-4 w-4" /> Simpan ke grup</span>
            <GroupSelector
              bind:selectedGroup
              onGroupSelect={(g) => (selectedGroup = g)}
              onViewGroup={viewGroupData}
              isPro={isPro && localSubscription?.status === "active"}
            />
            <span class="group-opt">opsional</span>
          </div>
          <Button
            variant="primary"
            icon={ScanLine}
            disabled={files.length === 0 || isProcessing}
            onclick={() => processDocuments()}
          >
            {scanLabel}
          </Button>
        </div>
      </Card>

      <!-- Step 2 — hasil -->
      {#if hasResults && !showModal}
        <Card>
          <div class="result-band">
            <div class="rb-left">
              <div class="rb-icon"><CheckCircle class="h-6 w-6" /></div>
              <div>
                <div class="rb-title">
                  {previewData.length} jamaah diekstrak
                  {#if polisCount > 0}
                    <span class="rb-dot">·</span>
                    <span class="rb-polis">{polisCount} polis terbaca</span>
                  {/if}
                </div>
                <div class="rb-sub">
                  {#if validationWarnings.length > 0}
                    {validationWarnings.length} data perlu diperiksa — tinjau sebelum
                    ekspor.
                  {:else}
                    Tinjau datanya sebelum diekspor.
                  {/if}
                </div>
              </div>
            </div>
            <div class="rb-actions">
              <Button variant="primary" icon={Table} onclick={() => (showModal = true)}>
                Tinjau &amp; edit
              </Button>
              <Button variant="ghost" icon={FileSpreadsheet} onclick={generateExcel}>
                {isGenerating ? "Mengekspor…" : "Export Excel"}
              </Button>
              {#if selectedGroup}
                <Button
                  variant="ghost"
                  icon={Users}
                  onclick={saveToGroup}
                  disabled={isSavingToGroup}
                >
                  {isSavingToGroup ? "Menyimpan…" : "Simpan ke grup"}
                </Button>
              {/if}
            </div>
          </div>
        </Card>
      {/if}

      <!-- Retry failed -->
      {#if failedFileNames.length > 0 && !isProcessing && !showModal}
        <div class="retry-banner">
          <div class="retry-text">
            <AlertCircle class="h-5 w-5" style="color:var(--c-danger);flex-shrink:0;" />
            <span>
              <strong>{failedFileNames.length}</strong> file gagal dipindai:
              {failedFileNames.join(", ")}
            </span>
          </div>
          <Button variant="danger" size="sm" onclick={retryFailed}>Coba Lagi</Button>
        </div>
      {/if}
    </div>
  {/if}
</div>

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

  /* Single, focused column — the workflow is a sequence, not a split. */
  .flow {
    display: flex;
    flex-direction: column;
    gap: var(--gap, 1.25rem);
    max-width: 760px;
    margin: 0 auto;
  }

  .pro-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px;
    border-radius: 999px;
    background: var(--c-accent-soft);
    color: #8a6a1d;
    font-size: 13px;
    font-weight: 700;
    border: 1px solid color-mix(in srgb, var(--c-accent) 30%, transparent);
  }

  .hidden-input {
    display: none;
  }

  /* The hero: one combined dropzone. */
  .dropzone {
    position: relative;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 4px;
    padding: 30px 20px;
    border: 2px dashed var(--c-line);
    border-radius: var(--radius-lg);
    background: var(--c-primary-tint);
    cursor: pointer;
    transition: border-color 0.15s, background 0.15s;
  }
  .dropzone:hover {
    border-color: color-mix(in srgb, var(--c-primary) 45%, var(--c-line));
  }
  .dropzone.drag {
    border-color: var(--c-primary);
    background: var(--c-primary-soft);
  }
  .dropzone.scanning {
    cursor: default;
    border-color: var(--c-primary);
  }
  .dropzone:focus-visible {
    outline: none;
    border-color: var(--c-primary);
    box-shadow: 0 0 0 3px var(--c-primary-soft);
  }
  .dz-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 52px;
    height: 52px;
    margin-bottom: 6px;
    border-radius: 50%;
    background: var(--c-primary-soft);
    color: var(--c-primary);
  }
  .dz-title {
    font-size: 15.5px;
    font-weight: 700;
    color: var(--c-ink);
  }
  .dz-sub {
    font-size: 13px;
    color: var(--c-muted);
    max-width: 420px;
  }

  /* Signature: a scan line sweeping the dropzone while the AI reads. */
  .scanline {
    position: absolute;
    left: 8%;
    right: 8%;
    top: 0;
    height: 2px;
    border-radius: 2px;
    background: linear-gradient(
      90deg,
      transparent,
      var(--c-primary),
      transparent
    );
    animation: sweep 1.5s ease-in-out infinite;
  }
  @keyframes sweep {
    0% {
      top: 6%;
      opacity: 0;
    }
    20% {
      opacity: 1;
    }
    80% {
      opacity: 1;
    }
    100% {
      top: 94%;
      opacity: 0;
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .scanline {
      animation: none;
      top: 50%;
      opacity: 0.6;
    }
  }

  .paste-btn {
    position: absolute;
    top: 10px;
    right: 10px;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 5px 10px;
    font-size: 12.5px;
    font-weight: 600;
    color: var(--c-ink-soft);
    background: var(--c-surface);
    border: 1px solid var(--c-line);
    border-radius: 999px;
    cursor: pointer;
  }
  .paste-btn:hover {
    background: var(--c-bg);
  }

  .polis-hint {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    margin-top: 12px;
    padding: 10px 12px;
    font-size: 13px;
    line-height: 1.45;
    color: #7a5e16;
    background: var(--c-accent-soft);
    border-radius: var(--radius);
  }
  .polis-hint strong {
    font-weight: 700;
  }

  .error-banner {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-top: 12px;
    padding: 0.8rem 1rem;
    font-size: 13.5px;
    color: var(--c-danger);
    background: var(--c-danger-soft);
    border-radius: var(--radius);
  }

  .success-banner {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 0.85rem 1rem;
    font-size: 14px;
    color: var(--c-primary-deep);
    border: 1px solid var(--c-primary-soft);
    background: var(--c-primary-soft);
    border-radius: var(--radius);
  }

  .files-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 18px 0 10px;
    font-size: 13px;
    font-weight: 600;
    color: var(--c-ink-soft);
  }
  .link-danger {
    background: none;
    border: none;
    padding: 0;
    font-size: 13px;
    color: var(--c-danger);
    cursor: pointer;
  }
  .link-danger:hover {
    text-decoration: underline;
  }

  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    max-width: 100%;
    padding: 6px 10px;
    font-size: 12.5px;
    color: var(--c-ink-soft);
    background: var(--c-surface);
    border: 1px solid var(--c-line);
    border-radius: var(--radius);
  }
  .chip.polis {
    background: var(--c-primary-soft);
    border-color: color-mix(in srgb, var(--c-primary) 28%, transparent);
    color: var(--c-primary-deep);
  }
  .chip-name {
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .chip-size {
    font-size: 11px;
    color: var(--c-faint);
  }
  .chip-badge {
    padding: 1px 7px;
    font-size: 10.5px;
    font-weight: 700;
    color: #fff;
    background: var(--c-primary);
    border-radius: 999px;
  }
  .chip-x {
    display: inline-flex;
    padding: 0;
    background: none;
    border: none;
    color: var(--c-faint);
    cursor: pointer;
  }
  .chip-x:hover {
    color: var(--c-danger);
  }

  .action-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 14px;
    margin-top: 18px;
    padding-top: 18px;
    border-top: 1px solid var(--c-line-soft);
  }
  .group-inline {
    display: flex;
    align-items: center;
    gap: 10px;
    flex-wrap: wrap;
    font-size: 13px;
    color: var(--c-muted);
  }
  .group-label {
    display: inline-flex;
    align-items: center;
    gap: 6px;
  }
  .group-opt {
    font-size: 12px;
    color: var(--c-faint);
  }

  /* Result band */
  .result-band {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 14px;
  }
  .rb-left {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
  }
  .rb-icon {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 42px;
    height: 42px;
    border-radius: var(--radius);
    background: var(--c-primary-soft);
    color: var(--c-primary);
  }
  .rb-title {
    font-size: 15px;
    font-weight: 800;
    color: var(--c-ink);
  }
  .rb-dot {
    color: var(--c-faint);
    font-weight: 400;
  }
  .rb-polis {
    color: var(--c-primary);
  }
  .rb-sub {
    margin-top: 1px;
    font-size: 13px;
    color: var(--c-muted);
  }
  .rb-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .retry-banner {
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
  .retry-text {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    font-size: 13.5px;
    color: var(--c-danger);
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
</style>
