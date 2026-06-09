<script>
  import { onMount } from 'svelte';
  import {
    CheckCircle,
    ChevronRight,
    Clock,
    Copy,
    Eye,
    FileSignature,
    FileText,
    Link2,
    Loader2,
    Pencil,
    Plus,
    Save,
    Send,
    ShieldCheck,
    Sparkles,
    Trash2,
  } from 'lucide-svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import PageHeader from '../components/PageHeader.svelte';
  import StatCard from '../components/StatCard.svelte';
  import EmptyState from '../components/EmptyState.svelte';
  import Card from '../components/ui/Card.svelte';
  import Badge from '../components/ui/Badge.svelte';
  import Button from '../components/ui/Button.svelte';
  import FilterTabs from '../components/ui/FilterTabs.svelte';
  import { ApiService } from '../services/api';
  import { showToast } from '../services/toast.svelte.js';

  let { user = null } = $props();

  const PACKAGE_TYPES = [
    { value: '', label: 'Semua Paket' },
    { value: 'umroh_reguler', label: 'Umroh Reguler' },
    { value: 'umroh_plus', label: 'Umroh Plus' },
    { value: 'haji_khusus', label: 'Haji Khusus' },
    { value: 'haji_onh_plus', label: 'Haji ONH Plus' },
  ];

  const CONTRACT_STATUSES = [
    { value: '', label: 'Semua Status' },
    { value: 'terkirim', label: 'Terkirim' },
    { value: 'ditandatangani', label: 'Ditandatangani' },
    { value: 'expired', label: 'Expired' },
  ];

  const VARIABLES = [
    'nama_jamaah',
    'no_paspor',
    'nama_paket',
    'tanggal_berangkat',
    'tanggal_pulang',
    'tipe_kamar',
    'harga_paket',
    'skema_bayar',
    'nama_travel',
    'tanggal_kontrak',
    'ketentuan_refund',
  ];

  const SAMPLE_VALUES = {
    nama_jamaah: 'Ahmad Fauzi',
    no_paspor: 'X1234567',
    nama_paket: 'Umroh Reguler Ramadan 2027',
    tanggal_berangkat: '12 Maret 2027',
    tanggal_pulang: '24 Maret 2027',
    tipe_kamar: 'Quad',
    harga_paket: 'Rp 29.500.000',
    skema_bayar: 'DP Rp 8.000.000, pelunasan H-45',
    nama_travel: 'Jamaah.in Travel',
    tanggal_kontrak: '22 Mei 2026',
    ketentuan_refund: 'Pembatalan >45 hari: potongan 10%. Pembatalan <30 hari: mengikuti biaya vendor aktual.',
  };

  let templates = $state([]);
  let contracts = $state([]);
  let isLoadingTemplates = $state(true);
  let isLoadingContracts = $state(true);
  let selectedType = $state('');
  let selectedStatus = $state('');
  let drawerOpen = $state(false);
  let previewOpen = $state(false);
  let generateOpen = $state(false);
  let saving = $state(false);
  let generating = $state(false);
  let deletingId = $state('');
  let previewLoading = $state(false);
  let previewHtml = $state('');
  let previewTitle = $state('');
  let formError = $state('');
  let generateError = $state('');
  let editor = $state(createEmptyTemplate());
  let generator = $state(createEmptyGenerator());

  let currentRole = $derived(user?.is_super_admin ? 'owner' : (user?.role ?? 'viewer'));
  let canEditContracts = $derived(currentRole === 'owner' || currentRole === 'admin');

  let filteredTemplates = $derived(
    selectedType
      ? templates.filter((tpl) => (tpl.package_type || '') === selectedType)
      : templates
  );

  let summaryStats = $derived({
    totalKontrak: contracts.length,
    ditandatangani: contracts.filter((c) => c.status === 'ditandatangani').length,
    menungguTtd: contracts.filter((c) => c.status === 'terkirim').length,
    expired: contracts.filter((c) => c.status === 'expired').length,
  });

  onMount(async () => {
    await Promise.all([loadTemplates(), loadContracts()]);
  });

  function createEmptyTemplate() {
    return {
      id: '',
      name: '',
      package_type: '',
      content: `PERJANJIAN PERJALANAN UMRAH\n\nSaya yang bertanda tangan di bawah ini {{nama_jamaah}} dengan nomor paspor {{no_paspor}} setuju mengikuti paket {{nama_paket}} pada tanggal {{tanggal_berangkat}} sampai {{tanggal_pulang}}.\n\nTipe kamar: {{tipe_kamar}}\nHarga paket: {{harga_paket}}\nSkema bayar: {{skema_bayar}}\n\nDengan ini saya menyetujui seluruh ketentuan perjalanan dari {{nama_travel}}, termasuk kebijakan refund berikut:\n{{ketentuan_refund}}\n\nKontrak ini dibuat pada {{tanggal_kontrak}}.`,
      is_active: true,
    };
  }

  function createEmptyGenerator() {
    return {
      template_id: '',
      recipient_name: '',
      recipient_phone: '',
      recipient_email: '',
      package_type: '',
      expires_in_days: 7,
      variables: { ...SAMPLE_VALUES },
    };
  }

  function typeLabel(value) {
    return PACKAGE_TYPES.find((item) => item.value === value)?.label || 'Semua Paket';
  }

  // Maps a contract status to the design Badge status text.
  function badgeStatus(status) {
    switch (status) {
      case 'ditandatangani':
        return 'Ditandatangani';
      case 'terkirim':
        return 'Terkirim';
      case 'expired':
        return 'Expired';
      default:
        return statusLabel(status);
    }
  }

  function statusLabel(status) {
    return CONTRACT_STATUSES.find((item) => item.value === status)?.label || status;
  }

  function renderContent(content) {
    return String(content || '')
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/\n/g, '<br>');
  }

  function publicLink(token) {
    if (typeof window === 'undefined') return `/#/kontrak/${token}`;
    return `${window.location.origin}/#/kontrak/${token}`;
  }

  function formatDateTime(value) {
    if (!value) return '-';
    return new Date(value).toLocaleString('id-ID', {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  async function loadTemplates() {
    isLoadingTemplates = true;
    try {
      const response = await ApiService.listContractTemplates(true);
      templates = response.data ?? response ?? [];
    } catch (error) {
      templates = [];
      showToast(error.message || 'Gagal memuat template kontrak', 'error');
    } finally {
      isLoadingTemplates = false;
    }
  }

  async function loadContracts() {
    isLoadingContracts = true;
    try {
      const response = await ApiService.listContracts(selectedStatus);
      contracts = response.data ?? response ?? [];
    } catch (error) {
      contracts = [];
      showToast(error.message || 'Gagal memuat kontrak yang dikirim', 'error');
    } finally {
      isLoadingContracts = false;
    }
  }

  function changeStatus(value) {
    selectedStatus = value;
    loadContracts();
  }

  function openCreateDrawer() {
    if (!canEditContracts) return;
    editor = createEmptyTemplate();
    formError = '';
    drawerOpen = true;
  }

  function openEditDrawer(template) {
    if (!canEditContracts) return;
    editor = {
      id: template.id,
      name: template.name,
      package_type: template.package_type || '',
      content: template.content,
      is_active: template.is_active,
    };
    formError = '';
    drawerOpen = true;
  }

  function openGenerateDrawer(template = null) {
    if (!canEditContracts) return;
    generator = createEmptyGenerator();
    if (template) {
      generator.template_id = template.id;
      generator.package_type = template.package_type || '';
    }
    generateError = '';
    generateOpen = true;
  }

  function insertVariable(variableName) {
    editor.content = `${editor.content} {{${variableName}}}`.trim();
  }

  async function openPreview(template = editor) {
    previewOpen = true;
    previewLoading = true;
    previewTitle = template.name || 'Preview Template';
    try {
      const response = await ApiService.previewContractTemplate({
        content: template.content,
        data: SAMPLE_VALUES,
      });
      previewHtml = renderContent((response.data ?? response)?.rendered || '');
    } catch (error) {
      previewHtml = '';
      showToast(error.message || 'Gagal menampilkan preview kontrak', 'error');
    } finally {
      previewLoading = false;
    }
  }

  async function saveTemplate() {
    if (!canEditContracts) return;
    formError = '';
    if (!editor.name.trim()) {
      formError = 'Nama template wajib diisi.';
      return;
    }
    if (!editor.content.trim()) {
      formError = 'Isi template wajib diisi.';
      return;
    }

    saving = true;
    try {
      let response;
      const payload = {
        name: editor.name.trim(),
        package_type: editor.package_type || undefined,
        content: editor.content.trim(),
        is_active: editor.is_active,
      };
      if (editor.id) {
        response = await ApiService.updateContractTemplate(editor.id, payload);
      } else {
        response = await ApiService.createContractTemplate(payload);
      }
      const saved = response.data ?? response;
      if (editor.id) {
        templates = templates.map((tpl) => (tpl.id === saved.id ? saved : tpl));
      } else {
        templates = [saved, ...templates];
      }
      drawerOpen = false;
      showToast(editor.id ? 'Template kontrak diperbarui' : 'Template kontrak dibuat', 'success');
    } catch (error) {
      formError = error.message || 'Gagal menyimpan template';
    } finally {
      saving = false;
    }
  }

  async function saveContractGeneration() {
    if (!canEditContracts) return;
    generateError = '';
    if (!generator.template_id) {
      generateError = 'Pilih template kontrak terlebih dahulu.';
      return;
    }
    if (!generator.recipient_name.trim()) {
      generateError = 'Nama jamaah wajib diisi.';
      return;
    }

    generating = true;
    try {
      const payload = {
        template_id: generator.template_id,
        recipient_name: generator.recipient_name.trim(),
        recipient_phone: generator.recipient_phone.trim() || undefined,
        recipient_email: generator.recipient_email.trim() || undefined,
        package_type: generator.package_type || undefined,
        expires_in_days: Number(generator.expires_in_days || 7),
        variables: Object.fromEntries(
          Object.entries(generator.variables)
            .map(([key, value]) => [key, String(value || '').trim()])
            .filter(([, value]) => value)
        ),
      };
      const response = await ApiService.createContract(payload);
      const saved = response.data ?? response;
      contracts = [saved, ...contracts];
      generateOpen = false;
      showToast('Kontrak berhasil digenerate', 'success');
    } catch (error) {
      generateError = error.message || 'Gagal generate kontrak';
    } finally {
      generating = false;
    }
  }

  async function deleteTemplate(template) {
    if (!canEditContracts) return;
    if (!confirm(`Hapus template "${template.name}"?`)) return;
    deletingId = template.id;
    try {
      await ApiService.deleteContractTemplate(template.id);
      templates = templates.filter((tpl) => tpl.id !== template.id);
      showToast('Template kontrak dihapus', 'success');
    } catch (error) {
      showToast(error.message || 'Gagal menghapus template kontrak', 'error');
    } finally {
      deletingId = '';
    }
  }

  async function copySigningLink(contract) {
    try {
      await navigator.clipboard.writeText(publicLink(contract.public_token));
      showToast('Link kontrak disalin', 'success');
    } catch {
      showToast('Gagal menyalin link kontrak', 'error');
    }
  }

  function fillGeneratorVariable(key, value) {
    generator.variables = {
      ...generator.variables,
      [key]: value,
    };
  }
</script>

<div class="contracts-page" style="min-height:100%;background:var(--c-bg)">
  <div class="px-4 py-6 lg:px-8">
    <PageHeader
      kicker="E-Kontrak"
      title="Kontrak & Akad"
      subtitle="Kelola template akad, generate kontrak per jamaah, dan pantau status penandatanganan digital."
    >
      {#snippet actions()}
        {#if canEditContracts}
          <Button variant="ghost" icon={Send} onclick={() => openGenerateDrawer()}>Generate Kontrak</Button>
          <Button variant="primary" icon={Plus} onclick={openCreateDrawer}>Template Baru</Button>
        {/if}
      {/snippet}
    </PageHeader>

    <!-- Summary cards (Suluk design) -->
    <div class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard icon={FileSignature} label="Total Kontrak" value={String(summaryStats.totalKontrak)} accent="var(--c-primary)" />
      <StatCard icon={CheckCircle} label="Ditandatangani" value={String(summaryStats.ditandatangani)} accent="var(--c-success)" />
      <StatCard icon={Clock} label="Menunggu TTD" value={String(summaryStats.menungguTtd)} accent="var(--c-warning)" />
      <StatCard icon={ShieldCheck} label="Expired" value={String(summaryStats.expired)} accent="var(--c-danger)" />
    </div>

    <!-- Sent contracts table (design Kontrak) -->
    <Card pad={false} class="mb-6" style="overflow:hidden">
      <div
        class="flex flex-wrap items-center justify-between gap-3 px-5 py-4"
        style="border-bottom:1px solid var(--c-line)"
      >
        <div class="flex items-center gap-2">
          <Link2 class="h-5 w-5" style="color:var(--c-primary)" />
          <h2 class="text-[15px] font-extrabold" style="color:var(--c-ink)">Kontrak Terkirim</h2>
        </div>
        <FilterTabs tabs={CONTRACT_STATUSES} value={selectedStatus} onChange={changeStatus} />
      </div>

      {#if isLoadingContracts}
        <div class="flex items-center justify-center gap-3 py-20" style="color:var(--c-muted)">
          <Loader2 class="h-5 w-5 animate-spin" />
          <span>Memuat kontrak...</span>
        </div>
      {:else if contracts.length === 0}
        <EmptyState
          icon={FileSignature}
          title="Belum ada kontrak yang digenerate"
          text="Generate kontrak dari template untuk mulai kirim link tanda tangan ke jamaah."
        />
      {:else}
        <div class="overflow-x-auto">
          <table class="w-full" style="border-collapse:collapse;font-size:13.5px">
            <thead>
              <tr>
                <th class="contracts-th" style="text-align:left">No. Kontrak</th>
                <th class="contracts-th" style="text-align:left">Jamaah</th>
                <th class="contracts-th" style="text-align:left">Template</th>
                <th class="contracts-th" style="text-align:center">Status</th>
                <th class="contracts-th" style="text-align:left">Expired</th>
                <th class="contracts-th" style="text-align:right">Link</th>
              </tr>
            </thead>
            <tbody>
              {#each contracts as contract}
                <tr class="contracts-row">
                  <td class="contracts-td">
                    <div class="font-bold" style="color:var(--c-ink)">{contract.recipient_name}</div>
                    {#if contract.recipient_phone}
                      <div class="mt-0.5 text-xs" style="color:var(--c-faint)">{contract.recipient_phone}</div>
                    {/if}
                  </td>
                  <td class="contracts-td">
                    <span style="color:var(--c-ink-soft)">{contract.recipient_email || '-'}</span>
                  </td>
                  <td class="contracts-td">
                    <span style="color:var(--c-muted)">{contract.template_name}</span>
                  </td>
                  <td class="contracts-td" style="text-align:center">
                    <Badge status={badgeStatus(contract.status)} dot />
                  </td>
                  <td class="contracts-td">
                    <div style="color:var(--c-ink-soft)">{formatDateTime(contract.expires_at)}</div>
                    {#if contract.signed_at}
                      <div class="mt-0.5 text-xs" style="color:var(--c-success)">Signed: {formatDateTime(contract.signed_at)}</div>
                    {/if}
                  </td>
                  <td class="contracts-td" style="text-align:right">
                    <button
                      type="button"
                      onclick={() => copySigningLink(contract)}
                      class="inline-flex items-center gap-1.5 text-xs font-bold"
                      style="color:var(--c-primary)"
                    >
                      <Copy class="h-3.5 w-3.5" />
                      Copy Link
                      <ChevronRight class="h-3.5 w-3.5" />
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </Card>

    <div class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
      <!-- Template management -->
      <Card pad={false} style="overflow:hidden">
        <div
          class="flex flex-wrap items-center justify-between gap-3 px-5 py-4"
          style="border-bottom:1px solid var(--c-line)"
        >
          <div class="flex items-center gap-2">
            <FileText class="h-5 w-5" style="color:var(--c-primary)" />
            <h2 class="text-[15px] font-extrabold" style="color:var(--c-ink)">Template Kontrak</h2>
          </div>
          <select
            bind:value={selectedType}
            class="rounded-xl bg-white px-3 py-2 text-sm outline-none"
            style="border:1px solid var(--c-line);color:var(--c-ink-soft)"
          >
            {#each PACKAGE_TYPES as item}
              <option value={item.value}>{item.label}</option>
            {/each}
          </select>
        </div>

        <div class="p-5">
          {#if isLoadingTemplates}
            <div class="flex items-center justify-center gap-3 py-16" style="color:var(--c-muted)">
              <Loader2 class="h-5 w-5 animate-spin" />
              <span>Memuat template...</span>
            </div>
          {:else if filteredTemplates.length === 0}
            <EmptyState
              icon={FileText}
              title="Belum ada template kontrak"
              text="Buat template untuk paket reguler, plus, atau haji khusus."
            />
          {:else}
            <div class="space-y-3">
              {#each filteredTemplates as template}
                <article
                  class="rounded-2xl p-4"
                  style="border:1px solid var(--c-line)"
                >
                  <div class="flex flex-wrap items-start justify-between gap-3">
                    <div class="min-w-0">
                      <div class="flex flex-wrap items-center gap-2">
                        <h3 class="text-sm font-bold" style="color:var(--c-ink)">{template.name}</h3>
                        <span
                          class="rounded-full px-2.5 py-1 text-[11px] font-semibold"
                          style="background:var(--c-bg-2);color:var(--c-muted)"
                        >
                          {typeLabel(template.package_type || '')}
                        </span>
                        <Badge status={template.is_active ? 'Aktif' : 'Nonaktif'} />
                      </div>
                      <p
                        class="mt-2 line-clamp-3 text-sm leading-relaxed"
                        style="color:var(--c-muted)"
                      >
                        {template.content}
                      </p>
                    </div>

                    <div class="flex flex-wrap items-center gap-2">
                      <Button variant="ghost" size="sm" icon={Eye} onclick={() => openPreview(template)}>Preview</Button>
                      {#if canEditContracts}
                        <Button variant="soft" size="sm" icon={FileSignature} onclick={() => openGenerateDrawer(template)}>Generate</Button>
                        <Button variant="ghost" size="sm" icon={Pencil} onclick={() => openEditDrawer(template)}>Edit</Button>
                        <Button
                          variant="danger"
                          size="sm"
                          icon={deletingId === template.id ? undefined : Trash2}
                          disabled={deletingId === template.id}
                          onclick={() => deleteTemplate(template)}
                        >
                          {#if deletingId === template.id}<Loader2 class="h-3.5 w-3.5 animate-spin" />{/if}
                          Hapus
                        </Button>
                      {/if}
                    </div>
                  </div>
                </article>
              {/each}
            </div>
          {/if}
        </div>
      </Card>

      <!-- Auto variables helper -->
      <Card style="border-color:var(--c-accent-soft);background:linear-gradient(180deg,#fff8eb 0%,#ffffff 100%)">
        <div class="mb-4 flex items-center gap-2">
          <Sparkles class="h-5 w-5" style="color:var(--c-accent)" />
          <h2 class="text-[15px] font-extrabold" style="color:var(--c-ink)">Variabel Otomatis</h2>
        </div>
        <p class="mb-4 text-sm leading-relaxed" style="color:var(--c-muted)">
          Klik variabel saat drawer template terbuka untuk menyisipkan placeholder sesuai PRD.
        </p>
        <div class="flex flex-wrap gap-2">
          {#each VARIABLES as variableName}
            <button
              type="button"
              onclick={() => {
                if (canEditContracts && drawerOpen) insertVariable(variableName);
              }}
              class="rounded-full bg-white px-3 py-1.5 text-xs font-semibold"
              style="border:1px solid var(--c-accent-soft);color:var(--c-accent)"
            >
              {'{{'}{variableName}{'}}'}
            </button>
          {/each}
        </div>
      </Card>
    </div>
  </div>
</div>

<SlideDrawer
  open={drawerOpen}
  title={editor.id ? 'Edit Template Kontrak' : 'Template Kontrak Baru'}
  width="760px"
  onClose={() => (drawerOpen = false)}
>
  <div class="space-y-6 p-6">
    {#if formError}
      <div class="rounded-2xl px-4 py-3 text-sm" style="border:1px solid var(--c-danger-soft);background:var(--c-danger-soft);color:var(--c-danger)">
        {formError}
      </div>
    {/if}

    <div class="grid gap-4 sm:grid-cols-2">
      <div>
        <label for="contract-template-name" class="mb-1 block text-sm font-medium" style="color:var(--c-ink-soft)">Nama Template</label>
        <input
          id="contract-template-name"
          type="text"
          bind:value={editor.name}
          class="w-full rounded-xl bg-white px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);color:var(--c-ink)"
        />
      </div>
      <div>
        <label for="contract-template-type" class="mb-1 block text-sm font-medium" style="color:var(--c-ink-soft)">Jenis Paket</label>
        <select
          id="contract-template-type"
          bind:value={editor.package_type}
          class="w-full rounded-xl bg-white px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);color:var(--c-ink)"
        >
          {#each PACKAGE_TYPES as item}
            <option value={item.value}>{item.label}</option>
          {/each}
        </select>
      </div>
      <label class="inline-flex items-center gap-2 text-sm font-medium" style="color:var(--c-ink-soft)">
        <input type="checkbox" bind:checked={editor.is_active} class="rounded" style="border:1px solid var(--c-line)" />
        Template aktif
      </label>
    </div>

    <div>
      <label for="contract-template-content" class="mb-1 block text-sm font-medium" style="color:var(--c-ink-soft)">Isi Template</label>
      <textarea
        id="contract-template-content"
        rows="16"
        bind:value={editor.content}
        class="w-full rounded-2xl bg-white px-4 py-3 text-sm leading-relaxed outline-none"
        style="border:1px solid var(--c-line);color:var(--c-ink)"
      ></textarea>
    </div>

    <div class="flex flex-wrap justify-between gap-3 pt-4" style="border-top:1px solid var(--c-line-soft)">
      <Button variant="ghost" icon={Eye} onclick={() => openPreview(editor)}>Preview</Button>
      <div class="flex flex-wrap gap-3">
        <Button variant="ghost" onclick={() => (drawerOpen = false)}>Batal</Button>
        <Button
          variant="primary"
          icon={saving ? undefined : Save}
          disabled={saving}
          onclick={saveTemplate}
        >
          {#if saving}<Loader2 class="h-4 w-4 animate-spin" />{/if}
          Simpan Template
        </Button>
      </div>
    </div>
  </div>
</SlideDrawer>

<SlideDrawer
  open={generateOpen}
  title="Generate Kontrak Jamaah"
  width="760px"
  onClose={() => (generateOpen = false)}
>
  <div class="space-y-6 p-6">
    {#if generateError}
      <div class="rounded-2xl px-4 py-3 text-sm" style="border:1px solid var(--c-danger-soft);background:var(--c-danger-soft);color:var(--c-danger)">
        {generateError}
      </div>
    {/if}

    <div class="grid gap-4 sm:grid-cols-2">
      <div>
        <label for="contract-generator-template" class="mb-1 block text-sm font-medium" style="color:var(--c-ink-soft)">Template</label>
        <select
          id="contract-generator-template"
          bind:value={generator.template_id}
          class="w-full rounded-xl bg-white px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);color:var(--c-ink)"
        >
          <option value="">Pilih template</option>
          {#each templates.filter((tpl) => tpl.is_active) as template}
            <option value={template.id}>{template.name}</option>
          {/each}
        </select>
      </div>
      <div>
        <label for="contract-generator-type" class="mb-1 block text-sm font-medium" style="color:var(--c-ink-soft)">Jenis Paket</label>
        <select
          id="contract-generator-type"
          bind:value={generator.package_type}
          class="w-full rounded-xl bg-white px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);color:var(--c-ink)"
        >
          {#each PACKAGE_TYPES as item}
            <option value={item.value}>{item.label}</option>
          {/each}
        </select>
      </div>
      <div>
        <label for="contract-generator-name" class="mb-1 block text-sm font-medium" style="color:var(--c-ink-soft)">Nama Jamaah</label>
        <input
          id="contract-generator-name"
          type="text"
          bind:value={generator.recipient_name}
          class="w-full rounded-xl bg-white px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);color:var(--c-ink)"
        />
      </div>
      <div>
        <label for="contract-generator-phone" class="mb-1 block text-sm font-medium" style="color:var(--c-ink-soft)">WhatsApp</label>
        <input
          id="contract-generator-phone"
          type="text"
          bind:value={generator.recipient_phone}
          class="w-full rounded-xl bg-white px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);color:var(--c-ink)"
        />
      </div>
      <div>
        <label for="contract-generator-email" class="mb-1 block text-sm font-medium" style="color:var(--c-ink-soft)">Email</label>
        <input
          id="contract-generator-email"
          type="email"
          bind:value={generator.recipient_email}
          class="w-full rounded-xl bg-white px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);color:var(--c-ink)"
        />
      </div>
      <div>
        <label for="contract-generator-expiry" class="mb-1 block text-sm font-medium" style="color:var(--c-ink-soft)">Expiry (hari)</label>
        <input
          id="contract-generator-expiry"
          type="number"
          min="1"
          max="30"
          bind:value={generator.expires_in_days}
          class="w-full rounded-xl bg-white px-3 py-2.5 text-sm outline-none"
          style="border:1px solid var(--c-line);color:var(--c-ink)"
        />
      </div>
    </div>

    <div>
      <div class="mb-2 flex items-center justify-between">
        <p class="block text-sm font-medium" style="color:var(--c-ink-soft)">Variabel Kontrak</p>
        <button
          type="button"
          onclick={() => (generator.variables = { ...SAMPLE_VALUES, nama_jamaah: generator.recipient_name || SAMPLE_VALUES.nama_jamaah })}
          class="text-xs font-semibold"
          style="color:var(--c-primary)"
        >
          Reset Sample
        </button>
      </div>
      <div class="grid gap-3 sm:grid-cols-2">
        {#each VARIABLES as variableName}
          <div>
            <label for={`generator-${variableName}`} class="mb-1 block text-xs font-bold uppercase tracking-wide" style="color:var(--c-faint)">{'{{'}{variableName}{'}}'}</label>
            <input
              id={`generator-${variableName}`}
              type="text"
              value={generator.variables?.[variableName] || ''}
              oninput={(event) => fillGeneratorVariable(variableName, event.currentTarget.value)}
              class="w-full rounded-xl bg-white px-3 py-2.5 text-sm outline-none"
              style="border:1px solid var(--c-line);color:var(--c-ink)"
            />
          </div>
        {/each}
      </div>
    </div>

    <div class="flex flex-wrap justify-end gap-3 pt-4" style="border-top:1px solid var(--c-line-soft)">
      <Button variant="ghost" onclick={() => (generateOpen = false)}>Batal</Button>
      <Button
        variant="primary"
        icon={generating ? undefined : Send}
        disabled={generating}
        onclick={saveContractGeneration}
      >
        {#if generating}<Loader2 class="h-4 w-4 animate-spin" />{/if}
        Generate Link
      </Button>
    </div>
  </div>
</SlideDrawer>

<SlideDrawer
  open={previewOpen}
  title={previewTitle}
  width="720px"
  onClose={() => (previewOpen = false)}
>
  {#if previewLoading}
    <div class="flex min-h-[320px] items-center justify-center gap-3" style="color:var(--c-muted)">
      <Loader2 class="h-5 w-5 animate-spin" />
      <span>Memuat preview kontrak...</span>
    </div>
  {:else}
    <div class="p-6">
      <div class="rounded-3xl bg-white p-6" style="border:1px solid var(--c-line);box-shadow:var(--shadow-sm)">
        <div class="text-sm leading-relaxed" style="color:var(--c-ink-soft)">
          {@html previewHtml}
        </div>
      </div>
    </div>
  {/if}
</SlideDrawer>

<style>
  .contracts-th {
    padding: 0 16px 12px;
    font-size: 11.5px;
    font-weight: 700;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    color: var(--c-faint);
    white-space: nowrap;
    border-bottom: 1px solid var(--c-line);
  }
  .contracts-th:first-child {
    padding-left: 20px;
  }
  .contracts-th:last-child {
    padding-right: 20px;
  }
  .contracts-td {
    padding: 14px 16px;
    border-bottom: 1px solid var(--c-line-soft);
    color: var(--c-ink-soft);
    white-space: nowrap;
    vertical-align: middle;
  }
  .contracts-td:first-child {
    padding-left: 20px;
  }
  .contracts-td:last-child {
    padding-right: 20px;
  }
  .contracts-row {
    transition: background 0.12s;
  }
  .contracts-row:hover {
    background: var(--c-primary-tint);
  }
  .contracts-row:last-child .contracts-td {
    border-bottom: none;
  }
</style>
