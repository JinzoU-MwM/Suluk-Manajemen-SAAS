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
    Files,
    Image as ImageIcon,
    ShieldCheck,
    Clipboard,
    X,
  } from "lucide-svelte";
  import TableResult from "../components/TableResult.svelte";
  import GroupSelector from "../components/GroupSelector.svelte";
  import UpgradeModal from "../components/UpgradeModal.svelte";
  import Card from "../components/ui/Card.svelte";
  import Button from "../components/ui/Button.svelte";
  import { ApiService } from "../services/api";
  import { isProOrHigher } from "../config/pricing.js";

  let {
    onLogout,
    user = null,
    subscription = null,
    onSubscriptionChange = null,
  } = $props();

  let files = $state([]);
  let dragOver = $state(false);
  let isProcessing = $state(false);
  let errorMessage = $state("");
  let localSubscription = $state(null);
  $effect(() => {
    if (subscription) localSubscription = subscription;
  });

  let showModal = $state(false);
  let previewData = $state([]);
  let isGenerating = $state(false);
  let validationWarnings = $state([]);
  let fileResults = $state([]);
  let failedFileNames = $state([]);

  let selectedGroup = $state(null);
  let isSavingToGroup = $state(false);
  let groupSaveSuccess = $state("");
  let showUpgradeModal = $state(false);

  // On-device recent-scan history (the instant-scan flow keeps no server job).
  const RECENT_KEY = "suluk_scanner_recent";
  let recent = $state([]);
  function loadRecent() {
    try {
      return JSON.parse(localStorage.getItem(RECENT_KEY) || "[]");
    } catch {
      return [];
    }
  }
  function saveRecent(list) {
    recent = list.slice(0, 4);
    try {
      localStorage.setItem(RECENT_KEY, JSON.stringify(recent));
    } catch {}
  }
  function pushRecent(entry) {
    saveRecent([entry, ...recent]);
  }
  function markLatestDone() {
    if (recent.length) saveRecent([{ ...recent[0], status: "Selesai" }, ...recent.slice(1)]);
  }
  function relTime(ts) {
    const s = Math.floor((Date.now() - ts) / 1000);
    if (s < 60) return "baru saja";
    const m = Math.floor(s / 60);
    if (m < 60) return `${m} menit lalu`;
    const h = Math.floor(m / 60);
    if (h < 24) return `${h} jam lalu`;
    return `${Math.floor(h / 24)} hari lalu`;
  }

  onMount(async () => {
    recent = loadRecent();
    if (!localSubscription) {
      try {
        localSubscription = await ApiService.getSubscriptionStatus();
      } catch (e) {
        console.error("Failed to fetch subscription:", e);
      }
    }
  });
  onMount(() => {
    const onWindowPaste = (e) => {
      if (e.clipboardData?.files?.length) addFiles(e.clipboardData.files);
    };
    window.addEventListener("paste", onWindowPaste);
    return () => window.removeEventListener("paste", onWindowPaste);
  });

  // ---- file intake ----
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
      const polis = fileResults.filter(
        (fr) => fr.doc_type === "polis" && fr.status === "completed",
      ).length;
      if (previewData.length > 0) {
        pushRecent({
          label: selectedGroup?.name || `${previewData.length} jamaah`,
          count: previewData.length,
          polis,
          status: "Review",
          ts: Date.now(),
        });
      }
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
      markLatestDone();
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
      markLatestDone();
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

  // Gating + plan
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
  let usageCount = $derived(localSubscription?.usage_count ?? null);
  let usageLimit = $derived(localSubscription?.usage_limit ?? null);
  let remaining = $derived(
    usageLimit ? Math.max(0, usageLimit - (usageCount || 0)) : null,
  );
  let usagePct = $derived(
    usageLimit ? Math.min(100, Math.round(((usageCount || 0) / usageLimit) * 100)) : 0,
  );
  let scanLabel = $derived(
    isProcessing
      ? "Memindai…"
      : `Pindai${files.length ? " " + files.length : ""} dokumen`,
  );
</script>

<div class="scanner-page page-enter">
  <!-- Header -->
  <div class="sc-head">
    <div>
      <div class="eyebrow">Fitur unggulan</div>
      <h1 class="sc-title">AI Scanner Dokumen</h1>
      <p class="sc-sub">
        Unggah dokumen jamaah &amp; polis asuransi — AI mengisi data Siskopatuh
        otomatis.
      </p>
    </div>
    <div class="head-badges">
      <span class="badge-ai"><Sparkles class="h-3.5 w-3.5" /> AI OCR</span>
      {#if isPro}
        <span class="badge-pro"><Crown class="h-3.5 w-3.5" /> Pro</span>
      {/if}
    </div>
  </div>

  {#if groupSaveSuccess}
    <div class="success-banner">
      <CheckCircle class="h-5 w-5" style="color:var(--c-success);flex-shrink:0;" />
      <span>{groupSaveSuccess}</span>
    </div>
  {/if}

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
          Upgrade ke Pro
        </Button>
      </div>
    </Card>
  {:else}
    <div class="scan-grid">
      <!-- LEFT: action -->
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
            {#if isProcessing}<ScanLine class="h-6 w-6" />{:else}<UploadCloud
                class="h-6 w-6"
              />{/if}
          </div>
          <div class="dz-title">
            {#if isProcessing}AI sedang membaca {files.length} dokumen…{:else}Tarik
              &amp; lepas dokumen{/if}
          </div>
          <div class="dz-sub">
            {#if isProcessing}Mohon tunggu sebentar.{:else}atau
              <span class="dz-link">klik untuk pilih</span> · KTP, Paspor, KK, PDF
              polis{/if}
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
              <Clipboard class="h-3.5 w-3.5" /><span>Tempel</span>
            </button>
          {/if}
        </div>

        <div class="polis-hint">
          <ShieldCheck class="h-4 w-4" style="color:var(--c-accent);flex-shrink:0;" />
          <span>
            Sertakan <strong>PDF polis asuransi</strong> — kolom Asuransi &amp; No
            Polis terisi otomatis, dicocokkan lewat nomor paspor.
          </span>
        </div>

        {#if errorMessage}
          <div class="error-banner">
            <AlertCircle class="h-5 w-5" style="flex-shrink:0;" /><span
              >{errorMessage}</span
            >
          </div>
        {/if}

        <div class="divider"></div>

        <!-- Group (optional) -->
        <div class="group-label-row">
          <span class="group-label"><Users class="h-4 w-4" /> Simpan ke grup</span>
          <span class="opt-pill">opsional</span>
        </div>
        <GroupSelector
          bind:selectedGroup
          onGroupSelect={(g) => (selectedGroup = g)}
          onViewGroup={viewGroupData}
          isPro={isPro && localSubscription?.status === "active"}
        />

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

        <!-- Action footer -->
        <div class="action-foot">
          <span class="file-count">
            <Files class="h-4 w-4" />
            {files.length > 0 ? `${files.length} dokumen` : "Belum ada dokumen"}
          </span>
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

      <!-- RIGHT: context -->
      <div class="context-col">
        <!-- Plan / usage -->
        <Card>
          <div class="ctx-plan-head">
            <Crown class="h-4 w-4" style="color:var(--c-accent)" />
            <span>Paket {isPro ? "Pro" : "Gratis"}</span>
          </div>
          {#if isPro}
            <div class="ctx-num">{(usageCount ?? 0).toLocaleString("id-ID")}</div>
            <div class="ctx-num-sub">dokumen di-scan · tanpa batas</div>
          {:else if usageLimit}
            <div class="ctx-num">{remaining}</div>
            <div class="ctx-num-sub">scan tersisa dari {usageLimit}</div>
            <div class="quota-bar">
              <div class="quota-fill" style="width:{usagePct}%"></div>
            </div>
            <Button
              variant="soft"
              icon={Crown}
              full
              size="sm"
              onclick={() => (showUpgradeModal = true)}
            >
              Upgrade ke Pro
            </Button>
          {:else}
            <div class="ctx-num-sub" style="margin-top:4px;">
              Pindai dokumen untuk mulai.
            </div>
          {/if}
        </Card>

        <!-- How it works -->
        <Card>
          <div class="ctx-title">Cara kerja</div>
          {#each ["Unggah dokumen jamaah & PDF polis", "AI membaca & mencocokkan via nomor paspor", "Data Siskopatuh terisi, siap di-review"] as step, i}
            <div class="step-row">
              <div class="step-num">{i + 1}</div>
              <div class="step-text">{step}</div>
            </div>
          {/each}
        </Card>

        <!-- Recent scans -->
        <Card>
          <div class="ctx-title">Scan terakhir</div>
          {#if recent.length === 0}
            <div class="recent-empty">
              Belum ada — hasil pindaian terakhir muncul di sini.
            </div>
          {:else}
            {#each recent as r}
              <div class="recent-row">
                <Files class="h-4 w-4" style="color:var(--c-faint);flex-shrink:0;" />
                <div class="recent-meta">
                  <div class="recent-label" title={r.label}>{r.label}</div>
                  <div class="recent-time">
                    {relTime(r.ts)}{#if r.polis > 0} · {r.polis} polis{/if}
                  </div>
                </div>
                <span class="recent-badge" class:done={r.status === "Selesai"}>
                  {r.status}
                </span>
              </div>
            {/each}
          {/if}
        </Card>
      </div>
    </div>

    <!-- Result band (full width) -->
    {#if hasResults && !showModal}
      <Card style="margin-top:var(--gap,1.25rem);">
        <div class="result-band">
          <div class="rb-left">
            <div class="rb-icon"><CheckCircle class="h-6 w-6" /></div>
            <div>
              <div class="rb-title">
                {previewData.length} jamaah diekstrak
                {#if polisCount > 0}<span class="rb-dot">·</span><span class="rb-polis"
                    >{polisCount} polis terbaca</span
                  >{/if}
              </div>
              <div class="rb-sub">
                {#if validationWarnings.length > 0}{validationWarnings.length} data perlu
                  diperiksa — tinjau sebelum ekspor.{:else}Tinjau datanya sebelum
                  diekspor.{/if}
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

    {#if failedFileNames.length > 0 && !isProcessing && !showModal}
      <div class="retry-banner">
        <div class="retry-text">
          <AlertCircle class="h-5 w-5" style="color:var(--c-danger);flex-shrink:0;" />
          <span
            ><strong>{failedFileNames.length}</strong> file gagal dipindai:
            {failedFileNames.join(", ")}</span
          >
        </div>
        <Button variant="danger" size="sm" onclick={retryFailed}>Coba Lagi</Button>
      </div>
    {/if}
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

  .sc-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 1.25rem;
  }
  .eyebrow {
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--c-primary);
    margin-bottom: 5px;
  }
  .sc-title {
    font-family: var(--font-display, Georgia, serif);
    font-size: 26px;
    font-weight: 700;
    line-height: 1.15;
    color: var(--c-ink);
    margin: 0 0 5px;
  }
  .sc-sub {
    margin: 0;
    font-size: 13.5px;
    color: var(--c-muted);
    max-width: 480px;
  }
  .head-badges {
    display: flex;
    gap: 8px;
    flex-shrink: 0;
  }
  .badge-ai,
  .badge-pro {
    display: inline-flex;
    align-items: center;
    gap: 5px;
    font-size: 12px;
    font-weight: 700;
    padding: 5px 10px;
    border-radius: 999px;
    white-space: nowrap;
  }
  .badge-ai {
    background: var(--c-primary-soft);
    color: var(--c-primary-deep);
  }
  .badge-pro {
    background: var(--c-accent-soft);
    color: #8a6a1d;
    border: 1px solid color-mix(in srgb, var(--c-accent) 30%, transparent);
  }

  .scan-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--gap, 1.25rem);
    align-items: start;
  }
  @media (min-width: 960px) {
    .scan-grid {
      grid-template-columns: 1.55fr 1fr;
    }
  }
  .context-col {
    display: flex;
    flex-direction: column;
    gap: var(--gap, 1.25rem);
  }

  .hidden-input {
    display: none;
  }

  /* Dropzone */
  .dropzone {
    position: relative;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 3px;
    padding: 28px 18px;
    border: 1.5px dashed var(--c-line);
    border-radius: var(--radius);
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
    width: 46px;
    height: 46px;
    margin-bottom: 6px;
    border-radius: 50%;
    background: var(--c-primary-soft);
    color: var(--c-primary);
  }
  .dz-title {
    font-size: 15px;
    font-weight: 700;
    color: var(--c-ink);
  }
  .dz-sub {
    font-size: 12.5px;
    color: var(--c-muted);
  }
  .dz-link {
    color: var(--c-primary);
    font-weight: 600;
  }
  .scanline {
    position: absolute;
    left: 8%;
    right: 8%;
    top: 0;
    height: 2px;
    border-radius: 2px;
    background: linear-gradient(90deg, transparent, var(--c-primary), transparent);
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
    gap: 5px;
    padding: 4px 9px;
    font-size: 12px;
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
    gap: 9px;
    margin-top: 12px;
    padding: 9px 12px;
    font-size: 12.5px;
    line-height: 1.45;
    color: #7a5e16;
    background: var(--c-accent-soft);
    border-radius: var(--radius);
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
    margin-bottom: var(--gap, 1.25rem);
    padding: 0.85rem 1rem;
    font-size: 14px;
    color: var(--c-primary-deep);
    border: 1px solid var(--c-primary-soft);
    background: var(--c-primary-soft);
    border-radius: var(--radius);
  }

  .divider {
    height: 1px;
    background: var(--c-line-soft);
    margin: 1rem 0;
  }
  .group-label-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 9px;
  }
  .group-label {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 13px;
    font-weight: 600;
    color: var(--c-ink-soft);
  }
  .opt-pill {
    font-size: 10.5px;
    color: var(--c-faint);
    border: 1px solid var(--c-line);
    padding: 1px 7px;
    border-radius: 999px;
  }

  .files-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin: 16px 0 9px;
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
    max-width: 180px;
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

  .action-foot {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-top: 1rem;
  }
  .file-count {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    font-size: 12px;
    color: var(--c-faint);
  }

  /* Context cards */
  .ctx-plan-head {
    display: flex;
    align-items: center;
    gap: 7px;
    font-size: 12.5px;
    font-weight: 600;
    color: var(--c-muted);
    margin-bottom: 8px;
  }
  .ctx-num {
    font-size: 30px;
    font-weight: 800;
    line-height: 1;
    color: var(--c-ink);
  }
  .ctx-num-sub {
    font-size: 11.5px;
    color: var(--c-faint);
    margin-top: 4px;
  }
  .quota-bar {
    height: 6px;
    background: var(--c-bg-2);
    border-radius: 999px;
    overflow: hidden;
    margin: 12px 0;
  }
  .quota-fill {
    height: 100%;
    background: var(--c-primary);
    border-radius: 999px;
    transition: width 0.4s;
  }
  .ctx-title {
    font-size: 13px;
    font-weight: 700;
    color: var(--c-ink);
    margin-bottom: 11px;
  }
  .step-row {
    display: flex;
    gap: 10px;
    margin-bottom: 10px;
  }
  .step-row:last-child {
    margin-bottom: 0;
  }
  .step-num {
    flex-shrink: 0;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: var(--c-primary-soft);
    color: var(--c-primary-deep);
    font-size: 11px;
    font-weight: 700;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .step-text {
    font-size: 12.5px;
    color: var(--c-muted);
    line-height: 1.45;
  }
  .recent-empty {
    font-size: 12px;
    color: var(--c-faint);
    line-height: 1.5;
  }
  .recent-row {
    display: flex;
    align-items: center;
    gap: 9px;
    padding: 7px 0;
    border-bottom: 1px solid var(--c-line-soft);
  }
  .recent-row:last-child {
    border-bottom: none;
  }
  .recent-meta {
    flex: 1;
    min-width: 0;
  }
  .recent-label {
    font-size: 12.5px;
    font-weight: 600;
    color: var(--c-ink-soft);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .recent-time {
    font-size: 11px;
    color: var(--c-faint);
  }
  .recent-badge {
    flex-shrink: 0;
    font-size: 10.5px;
    font-weight: 700;
    padding: 2px 8px;
    border-radius: 999px;
    background: var(--c-warning-soft);
    color: var(--c-warning);
  }
  .recent-badge.done {
    background: var(--c-primary-soft);
    color: var(--c-primary-deep);
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
    margin: 0 4px;
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
    margin-top: var(--gap, 1.25rem);
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
