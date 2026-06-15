<script>
  import { onMount } from 'svelte';
  import { formatRupiah as formatIDR, formatDate } from '../utils/formatting.js';
  import {
    CalendarDays,
    ChevronRight,
    Copy,
    Edit,
    ExternalLink,
    Globe,
    Hotel,
    Layers,
    Link2,
    Loader2,
    Lock,
    Package,
    Plane,
    Plus,
    Trash2,
    Users,
  } from 'lucide-svelte';
  import StatusBadge from '../components/StatusBadge.svelte';
  import SlideDrawer from '../components/SlideDrawer.svelte';
  import IDRInput from '../components/IDRInput.svelte';
  import PageHeader from '../components/PageHeader.svelte';
  import StatCard from '../components/StatCard.svelte';
  import EmptyState from '../components/EmptyState.svelte';
  import Card from '../components/ui/Card.svelte';
  import Button from '../components/ui/Button.svelte';
  import FilterTabs from '../components/ui/FilterTabs.svelte';
  import ProgressBar from '../components/ui/ProgressBar.svelte';
  import { ApiService } from '../services/api';
  import { showToast } from '../services/toast.svelte.js';

  let { onNavigate, user = null } = $props();

  const STATUS_TABS = [
    { id: 'all', label: 'Semua' },
    { id: 'open', label: 'Open' },
    { id: 'draft', label: 'Draft' },
    { id: 'full', label: 'Penuh' },
    { id: 'closed', label: 'Ditutup' },
    { id: 'done', label: 'Selesai' },
  ];

  const PACKAGE_TYPES = [
    { value: 'umroh_reguler', label: 'Umroh Reguler' },
    { value: 'umroh_plus', label: 'Umroh Plus' },
    { value: 'haji_khusus', label: 'Haji Khusus' },
    { value: 'haji_onh_plus', label: 'Haji ONH Plus' },
  ];

  const STATUS_OPTIONS = [
    { value: 'draft', label: 'Draft' },
    { value: 'open', label: 'Open' },
    { value: 'full', label: 'Penuh' },
    { value: 'closed', label: 'Ditutup' },
    { value: 'done', label: 'Selesai' },
  ];

  const ROOM_FIELDS = [
    { room_type: 'quad', label: 'Quad (4 orang)' },
    { room_type: 'triple', label: 'Triple (3 orang)' },
    { room_type: 'double', label: 'Double (2 orang)' },
    { room_type: 'single', label: 'Single (1 orang)' },
  ];

  let packages = $state([]);
  let isLoading = $state(true);
  let filterStatus = $state('all');
  let drawerOpen = $state(false);
  let detailLoading = $state(false);
  let selectedPackage = $state(null);
  let formDrawerOpen = $state(false);
  let formMode = $state('create');
  let savingForm = $state(false);
  let changingStatus = $state(false);
  let publishingId = $state('');
  let deletingId = $state('');
  let formError = $state('');
  let formState = $state(createEmptyForm());

  let currentRole = $derived(user?.is_super_admin ? 'owner' : (user?.role ?? 'viewer'));
  let canEditPackages = $derived(currentRole === 'owner' || currentRole === 'admin');
  let canPublishPackages = $derived(currentRole === 'owner');
  let canDeletePackages = $derived(currentRole === 'owner');

  let filtered = $derived(
    filterStatus === 'all'
      ? packages
      : packages.filter((pkg) => pkg.status === filterStatus)
  );

  // Summary tiles (Suluk design). All values derived from the loaded packages.
  let summaryStats = $derived({
    totalPackages: packages.length,
    openPackages: packages.filter((pkg) => pkg.status === 'open').length,
    publishedPackages: packages.filter((pkg) => pkg.is_published).length,
    totalSeats: packages.reduce((sum, pkg) => sum + (pkg.total_seats || 0), 0),
    reservedSeats: packages.reduce((sum, pkg) => sum + (pkg.reserved_seats || 0), 0),
  });

  onMount(loadPackages);

  function createEmptyForm() {
    return {
      id: '',
      name: '',
      packageType: 'umroh_reguler',
      departureDate: '',
      returnDate: '',
      totalSeats: 40,
      status: 'draft',
      description: '',
      airline: '',
      flightNumberGo: '',
      flightNumberReturn: '',
      hotelMakkahName: '',
      hotelMadinahName: '',
      prices: {
        quad: 0,
        triple: 0,
        double: 0,
        single: 0,
      },
      quotas: {
        quad: 0,
        triple: 0,
        double: 0,
        single: 0,
      },
      existingPricingTiers: [],
    };
  }

  function normalizePackage(pkg) {
    const pricingTiers = [...(pkg?.pricing_tiers || [])].sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0));
    const reservedSeats = Number(pkg?.reserved_seats || 0);
    const totalSeats = Number(pkg?.total_seats || 0);
    return {
      ...pkg,
      pricing_tiers: pricingTiers,
      reserved_seats: reservedSeats,
      total_seats: totalSeats,
      available_seats: Math.max(0, totalSeats - reservedSeats),
    };
  }

  function getPackagesFromResponse(response) {
    if (Array.isArray(response?.data)) return response.data;
    if (Array.isArray(response)) return response;
    return [];
  }

  async function loadPackages() {
    isLoading = true;
    try {
      const response = await ApiService.listPackages({ pageSize: 100 });
      packages = getPackagesFromResponse(response).map(normalizePackage);
    } catch (error) {
      packages = [];
      showToast(error.message || 'Gagal memuat data paket', 'error');
    } finally {
      isLoading = false;
    }
  }

  async function openDetail(pkg) {
    drawerOpen = true;
    detailLoading = true;
    selectedPackage = normalizePackage(pkg);
    try {
      const response = await ApiService.getPackage(pkg.id);
      selectedPackage = normalizePackage(response.data ?? response);
      mergePackage(selectedPackage);
    } catch (error) {
      showToast(error.message || 'Gagal memuat detail paket', 'error');
    } finally {
      detailLoading = false;
    }
  }

  function mergePackage(pkg) {
    packages = packages.map((item) => (item.id === pkg.id ? normalizePackage(pkg) : item));
  }

  function removePackage(packageId) {
    packages = packages.filter((pkg) => pkg.id !== packageId);
    if (selectedPackage?.id === packageId) {
      selectedPackage = null;
      drawerOpen = false;
    }
  }

  function toDateInput(dateStr) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    if (Number.isNaN(date.getTime())) return '';
    return date.toISOString().slice(0, 10);
  }

  function getLowestPrice(pkg) {
    const prices = (pkg.pricing_tiers || [])
      .map((tier) => Number(tier.price || 0))
      .filter((price) => price > 0);
    return prices.length > 0 ? Math.min(...prices) : null;
  }

  function getPublicLink(slug) {
    if (typeof window === 'undefined') {
      return `/#/paket/${slug}`;
    }
    return `${window.location.origin}/#/paket/${slug}`;
  }

  async function copyPublicLink(pkg) {
    const link = getPublicLink(pkg.slug);
    try {
      await navigator.clipboard.writeText(link);
      showToast('Link paket berhasil disalin', 'success');
    } catch {
      const textArea = document.createElement('textarea');
      textArea.value = link;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      showToast('Link paket berhasil disalin', 'success');
    }
  }

  function openPublicLink(pkg) {
    if (!pkg.is_published) {
      showToast('Paket masih private. Publikasikan dulu untuk membuka halaman publik.', 'error');
      return;
    }
    window.open(getPublicLink(pkg.slug), '_blank', 'noopener,noreferrer');
  }

  async function togglePublish(pkg) {
    if (!canPublishPackages) return;
    publishingId = pkg.id;
    try {
      const response = await ApiService.updatePackage(pkg.id, {
        is_published: !pkg.is_published,
      });
      const updated = normalizePackage(response.data ?? response);
      mergePackage(updated);
      if (selectedPackage?.id === updated.id) {
        selectedPackage = updated;
      }
      showToast(
        updated.is_published ? 'Paket berhasil dipublikasikan' : 'Paket dikembalikan ke mode internal',
        'success'
      );
    } catch (error) {
      showToast(error.message || 'Gagal mengubah publikasi paket', 'error');
    } finally {
      publishingId = '';
    }
  }

  async function updatePackageStatus(pkg, status) {
    if (!canEditPackages || status === pkg.status) return;
    changingStatus = true;
    try {
      const response = await ApiService.updatePackageStatus(pkg.id, status);
      const updated = normalizePackage(response.data ?? response);
      mergePackage(updated);
      if (selectedPackage?.id === updated.id) {
        selectedPackage = updated;
      }
      showToast('Status paket diperbarui', 'success');
    } catch (error) {
      showToast(error.message || 'Gagal memperbarui status paket', 'error');
    } finally {
      changingStatus = false;
    }
  }

  async function deletePackage(pkg) {
    if (!canDeletePackages) return;
    if (!confirm(`Hapus paket "${pkg.name}"? Tindakan ini tidak dapat dibatalkan.`)) {
      return;
    }
    deletingId = pkg.id;
    try {
      await ApiService.deletePackage(pkg.id);
      removePackage(pkg.id);
      showToast('Paket berhasil dihapus', 'success');
    } catch (error) {
      showToast(error.message || 'Gagal menghapus paket', 'error');
    } finally {
      deletingId = '';
    }
  }

  function openCreateForm() {
    if (!canEditPackages) return;
    formMode = 'create';
    formError = '';
    formState = createEmptyForm();
    formDrawerOpen = true;
  }

  function packageToForm(pkg) {
    const prices = { quad: 0, triple: 0, double: 0, single: 0 };
    const quotas = { quad: 0, triple: 0, double: 0, single: 0 };
    for (const tier of pkg.pricing_tiers || []) {
      prices[tier.room_type] = Number(tier.price || 0);
      quotas[tier.room_type] = Number(tier.quota_seats || 0);
    }

    return {
      id: pkg.id,
      name: pkg.name || '',
      packageType: pkg.package_type || 'umroh_reguler',
      departureDate: toDateInput(pkg.departure_date),
      returnDate: toDateInput(pkg.return_date),
      totalSeats: Number(pkg.total_seats || 0),
      status: pkg.status || 'draft',
      description: pkg.description || '',
      airline: pkg.airline || '',
      flightNumberGo: pkg.flight_number_go || '',
      flightNumberReturn: pkg.flight_number_return || '',
      hotelMakkahName: pkg.hotel_makkah_name || '',
      hotelMadinahName: pkg.hotel_madinah_name || '',
      prices,
      quotas,
      existingPricingTiers: pkg.pricing_tiers || [],
    };
  }

  function openEditForm() {
    if (!canEditPackages || !selectedPackage) return;
    formMode = 'edit';
    formError = '';
    formState = packageToForm(selectedPackage);
    formDrawerOpen = true;
  }

  function buildTierPayloads() {
    let sortOrder = 0;
    return ROOM_FIELDS
      .map((room) => {
        const price = Number(formState.prices[room.room_type] || 0);
        if (price <= 0) return null;
        const existing = (formState.existingPricingTiers || []).find((tier) => tier.room_type === room.room_type);
        return {
          id: existing?.id,
          room_type: room.room_type,
          price,
          label: room.label,
          is_early_bird: false,
          sort_order: sortOrder++,
          quota_seats: Number(formState.quotas?.[room.room_type] || 0),
        };
      })
      .filter(Boolean);
  }

  function buildPackagePayload() {
    return {
      name: formState.name.trim(),
      package_type: formState.packageType,
      departure_date: formState.departureDate || undefined,
      return_date: formState.returnDate || undefined,
      total_seats: Number(formState.totalSeats || 0),
      description: formState.description.trim() || undefined,
      airline: formState.airline.trim() || undefined,
      flight_number_go: formState.flightNumberGo.trim() || undefined,
      flight_number_return: formState.flightNumberReturn.trim() || undefined,
      hotel_makkah_name: formState.hotelMakkahName.trim() || undefined,
      hotel_madinah_name: formState.hotelMadinahName.trim() || undefined,
    };
  }

  function validateForm() {
    if (!formState.name.trim()) return 'Nama paket wajib diisi.';
    if (!formState.departureDate) return 'Tanggal berangkat wajib diisi.';
    if (!formState.returnDate) return 'Tanggal pulang wajib diisi.';
    if (new Date(formState.returnDate) < new Date(formState.departureDate)) {
      return 'Tanggal pulang tidak boleh lebih awal dari tanggal berangkat.';
    }
    if (Number(formState.totalSeats || 0) < 1) {
      return 'Total kursi minimal 1.';
    }
    if (buildTierPayloads().length === 0) {
      return 'Minimal satu harga kamar harus diisi.';
    }
    return '';
  }

  async function syncPricingTiers(packageId, existingPricingTiers, desiredPricingTiers) {
    const existingByRoom = new Map((existingPricingTiers || []).map((tier) => [tier.room_type, tier]));
    const desiredRooms = new Set(desiredPricingTiers.map((tier) => tier.room_type));

    for (const tier of desiredPricingTiers) {
      const existing = existingByRoom.get(tier.room_type);
      if (existing) {
        await ApiService.updatePricingTier(packageId, existing.id, {
          room_type: tier.room_type,
          price: tier.price,
          label: tier.label,
          is_early_bird: false,
          sort_order: tier.sort_order,
          quota_seats: tier.quota_seats,
        });
      } else {
        await ApiService.createPricingTier(packageId, {
          room_type: tier.room_type,
          price: tier.price,
          label: tier.label,
          is_early_bird: false,
          sort_order: tier.sort_order,
          quota_seats: tier.quota_seats,
        });
      }
    }

    for (const existing of existingPricingTiers || []) {
      if (!desiredRooms.has(existing.room_type)) {
        await ApiService.deletePricingTier(packageId, existing.id);
      }
    }
  }

  async function saveForm() {
    if (!canEditPackages) return;

    formError = validateForm();
    if (formError) {
      return;
    }

    savingForm = true;
    try {
      const tierPayloads = buildTierPayloads();
      let savedPackage;

      if (formMode === 'create') {
        const response = await ApiService.createPackage({
          ...buildPackagePayload(),
          pricing_tiers: tierPayloads,
        });
        savedPackage = normalizePackage(response.data ?? response);
      } else {
        await ApiService.updatePackage(formState.id, buildPackagePayload());
        await syncPricingTiers(formState.id, formState.existingPricingTiers, tierPayloads);
        await ApiService.updatePackageStatus(formState.id, formState.status);
        const response = await ApiService.getPackage(formState.id);
        savedPackage = normalizePackage(response.data ?? response);
      }

      if (formMode === 'create') {
        packages = [savedPackage, ...packages];
      } else {
        mergePackage(savedPackage);
      }

      selectedPackage = savedPackage;
      drawerOpen = true;
      formDrawerOpen = false;
      showToast(
        formMode === 'create' ? 'Paket berhasil dibuat' : 'Paket berhasil diperbarui',
        'success'
      );
    } catch (error) {
      formError = error.message || 'Gagal menyimpan paket';
    } finally {
      savingForm = false;
    }
  }

  function typeLabel(packageType) {
    return PACKAGE_TYPES.find((type) => type.value === packageType)?.label || packageType;
  }

  // Deterministic brand color per package type for the card/drawer gradient banner.
  const TYPE_COLORS = {
    umroh_reguler: '#1B7F5A',
    umroh_plus: '#0F3D2E',
    haji_khusus: '#C99A2E',
    haji_onh_plus: '#2563a8',
  };
  function brandColor(pkg) {
    return TYPE_COLORS[pkg?.package_type] || '#1B7F5A';
  }

  // Trip duration in days (inclusive) from departure → return, when both exist.
  function durationDays(pkg) {
    if (!pkg?.departure_date || !pkg?.return_date) return null;
    const start = new Date(pkg.departure_date);
    const end = new Date(pkg.return_date);
    if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return null;
    const diff = Math.round((end.getTime() - start.getTime()) / 86400000) + 1;
    return diff > 0 ? diff : null;
  }
</script>

<div class="flex h-screen flex-col bg-slate-50">
  <div class="flex-1 overflow-y-auto px-6 py-6">
    <PageHeader
      kicker="Penjualan"
      title="Paket & Harga"
      subtitle="Kelola paket umroh & haji, harga, kuota, dan publikasi ke halaman publik."
    >
      {#snippet actions()}
        {#if canEditPackages}
          <Button variant="primary" icon={Plus} onclick={openCreateForm}>Buat Paket</Button>
        {/if}
      {/snippet}
    </PageHeader>

    <!-- Summary cards (Suluk design) -->
    <div class="grid grid-cols-2 gap-3 lg:grid-cols-4">
      <StatCard icon={Layers} label="Total Paket" value={`${summaryStats.totalPackages}`} accent="#1B7F5A" sub="paket tersedia" />
      <StatCard icon={Globe} label="Dipublikasikan" value={`${summaryStats.publishedPackages}`} accent="#2563a8" sub="tampil ke publik" />
      <StatCard icon={Package} label="Paket Open" value={`${summaryStats.openPackages}`} accent="#C99A2E" sub="menerima pendaftaran" />
      <StatCard icon={Users} label="Kuota Terisi" value={`${summaryStats.reservedSeats}/${summaryStats.totalSeats}`} accent="#c0392b" sub="kursi terbooking" />
    </div>

    <!-- Filter tabs -->
    <div class="mt-5 overflow-x-auto pb-0.5">
      <FilterTabs
        tabs={STATUS_TABS.map((tab) => ({ value: tab.id, label: tab.label }))}
        value={filterStatus}
        onChange={(value) => (filterStatus = value)}
      />
    </div>

    <!-- Package grid -->
    <div class="mt-5">
      {#if isLoading}
        <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {#each [1, 2, 3] as item}
            <div class="h-64 animate-pulse rounded-2xl bg-white" aria-hidden="true">{item}</div>
          {/each}
        </div>
      {:else if filtered.length === 0}
        <EmptyState
          icon={Package}
          title="Belum ada paket"
          text={canEditPackages
            ? 'Buat paket pertama untuk mulai membagikan penawaran ke jamaah.'
            : 'Belum ada paket yang bisa ditampilkan.'}
        />
      {:else}
        <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {#each filtered as pkg}
            {@const accent = brandColor(pkg)}
            {@const duration = durationDays(pkg)}
            <Card hover pad={false} onclick={() => openDetail(pkg)} style="overflow:hidden;display:flex;flex-direction:column">
              <!-- Gradient brand banner -->
              <div
                class="relative flex min-h-[104px] flex-col justify-end p-4"
                style="background:linear-gradient(120deg, {accent}, {accent}cc)"
              >
                <div class="absolute right-3.5 top-3.5">
                  <StatusBadge status={pkg.status} size="xs" />
                </div>
                <div class="absolute left-4 top-3.5 flex items-center gap-1.5 text-[11px] font-bold uppercase tracking-[0.05em] text-white/90">
                  {#if pkg.is_published}
                    <Globe class="h-3.5 w-3.5" /> Publik
                  {:else}
                    <Lock class="h-3.5 w-3.5" /> Internal
                  {/if}
                </div>
                <h3 class="font-serif text-[20px] font-bold leading-tight text-white" style="text-shadow:0 1px 8px rgba(0,0,0,.2)">
                  {pkg.name}
                </h3>
                <p class="mt-0.5 text-[12px] font-medium text-white/85">{typeLabel(pkg.package_type)}</p>
              </div>

              <!-- Body -->
              <div class="flex flex-1 flex-col gap-3.5 p-5">
                <div class="flex flex-wrap gap-x-4 gap-y-1.5 text-[12.5px]" style="color:var(--c-muted)">
                  <span class="flex items-center gap-1.5">
                    <CalendarDays class="h-3.5 w-3.5 flex-shrink-0" />
                    {formatDate(pkg.departure_date)} - {formatDate(pkg.return_date)}
                  </span>
                  {#if duration}
                    <span class="flex items-center gap-1.5">
                      <Layers class="h-3.5 w-3.5 flex-shrink-0" />
                      {duration} hari
                    </span>
                  {/if}
                  <span class="flex items-center gap-1.5">
                    <Plane class="h-3.5 w-3.5 flex-shrink-0" />
                    {pkg.airline || '-'} {#if pkg.flight_number_go}&middot; {pkg.flight_number_go}{/if}
                  </span>
                  <span class="flex items-center gap-1.5">
                    <Hotel class="h-3.5 w-3.5 flex-shrink-0" />
                    {pkg.hotel_makkah_name || '-'}
                  </span>
                </div>

                <div>
                  <div class="mb-1.5 flex items-center justify-between text-[12px]">
                    <span class="font-semibold" style="color:var(--c-muted)">Kuota terisi</span>
                    <span class="font-bold" style="font-variant-numeric:tabular-nums;color:var(--c-ink)">{pkg.reserved_seats}/{pkg.total_seats}</span>
                  </div>
                  <ProgressBar
                    value={pkg.reserved_seats}
                    max={pkg.total_seats}
                    color={pkg.reserved_seats >= pkg.total_seats ? 'var(--c-danger)' : accent}
                  />
                </div>

                <div class="mt-auto flex items-end justify-between border-t pt-3.5" style="border-color:var(--c-line-soft)">
                  <div>
                    <p class="text-[11.5px] font-semibold" style="color:var(--c-faint)">Mulai dari</p>
                    {#if getLowestPrice(pkg)}
                      <p class="text-[19px] font-extrabold" style="font-variant-numeric:tabular-nums;color:var(--c-ink)">{formatIDR(getLowestPrice(pkg))}</p>
                    {:else}
                      <p class="text-sm font-semibold" style="color:var(--c-faint)">Harga belum diatur</p>
                    {/if}
                  </div>
                  <span class="flex items-center gap-1 text-[12.5px] font-bold" style="color:var(--c-primary)">
                    Detail <ChevronRight class="h-4 w-4" />
                  </span>
                </div>
              </div>
            </Card>
          {/each}
        </div>
      {/if}
    </div>
  </div>
</div>

<SlideDrawer
  open={drawerOpen}
  title={selectedPackage?.name || 'Detail Paket'}
  width="640px"
  onClose={() => (drawerOpen = false)}
>
  {#if detailLoading}
    <div class="flex h-full min-h-[320px] items-center justify-center gap-3 text-slate-500">
      <Loader2 class="h-5 w-5 animate-spin" />
      <span>Memuat detail paket...</span>
    </div>
  {:else if selectedPackage}
    <div class="space-y-6 p-6">
      <div class="flex flex-wrap items-center gap-3">
        <StatusBadge status={selectedPackage.status} />
        <span class="inline-flex items-center gap-1 rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-600">
          {#if selectedPackage.is_published}
            <Globe class="h-3.5 w-3.5" /> Publik
          {:else}
            <Lock class="h-3.5 w-3.5" /> Internal
          {/if}
        </span>
        <span class="inline-flex rounded-full bg-primary-50 px-3 py-1 text-xs font-semibold text-primary-700">
          {typeLabel(selectedPackage.package_type)}
        </span>
      </div>

      <!-- Brand price hero (Suluk design) -->
      {#if getLowestPrice(selectedPackage)}
        <div
          class="rounded-2xl p-5 text-white"
          style="background:linear-gradient(120deg, {brandColor(selectedPackage)}, {brandColor(selectedPackage)}cc)"
        >
          <p class="text-xs font-bold uppercase tracking-[0.04em] text-white/85">Mulai dari / Jamaah</p>
          <p class="mt-1 text-[30px] font-extrabold" style="font-variant-numeric:tabular-nums">{formatIDR(getLowestPrice(selectedPackage))}</p>
          <p class="mt-1 text-[13px] text-white/90">
            {selectedPackage.reserved_seats} dari {selectedPackage.total_seats} kursi terisi
            {#if getLowestPrice(selectedPackage) && selectedPackage.reserved_seats > 0}
              &middot; {formatIDR(getLowestPrice(selectedPackage) * selectedPackage.reserved_seats)} potensi omzet
            {/if}
          </p>
        </div>
      {/if}

      <div class="rounded-2xl border border-slate-100 p-4">
        <div class="mb-3 flex items-center gap-2">
          <Link2 class="h-4 w-4 text-slate-500" />
          <h3 class="text-sm font-bold text-slate-800">Link Publik Paket</h3>
        </div>
        <div class="rounded-xl bg-slate-50 px-3 py-2 text-xs text-slate-600">
          {getPublicLink(selectedPackage.slug)}
        </div>
        <div class="mt-3 flex flex-wrap gap-2">
          <button
            type="button"
            onclick={() => copyPublicLink(selectedPackage)}
            class="inline-flex items-center gap-2 rounded-xl border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-600 transition-colors hover:bg-slate-50"
          >
            <Copy class="h-4 w-4" />
            Salin Link
          </button>
          <button
            type="button"
            onclick={() => openPublicLink(selectedPackage)}
            disabled={!selectedPackage.is_published}
            class="inline-flex items-center gap-2 rounded-xl border border-slate-200 px-3 py-2 text-sm font-semibold text-slate-600 transition-colors hover:bg-slate-50 disabled:cursor-not-allowed disabled:opacity-50"
          >
            <ExternalLink class="h-4 w-4" />
            Buka Halaman
          </button>
          {#if canPublishPackages}
            <button
              type="button"
              onclick={() => togglePublish(selectedPackage)}
              disabled={publishingId === selectedPackage.id}
              class="inline-flex items-center gap-2 rounded-xl bg-slate-900 px-3 py-2 text-sm font-semibold text-white transition-colors hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {#if publishingId === selectedPackage.id}
                <Loader2 class="h-4 w-4 animate-spin" />
              {:else if selectedPackage.is_published}
                <Lock class="h-4 w-4" />
              {:else}
                <Globe class="h-4 w-4" />
              {/if}
              {selectedPackage.is_published ? 'Jadikan Internal' : 'Publikasikan'}
            </button>
          {/if}
        </div>
        <p class="mt-3 text-xs text-slate-400">
          {#if selectedPackage.is_published}
            Link aktif dan bisa dibagikan ke calon jamaah.
          {:else}
            Link sudah tersedia tetapi belum bisa diakses publik sampai owner mempublikasikan paket.
          {/if}
        </p>
      </div>

      <div class="rounded-2xl border border-slate-100 p-4">
        <div class="mb-3 flex items-center justify-between gap-3">
          <h3 class="text-sm font-bold text-slate-800">Status Paket</h3>
          {#if canEditPackages}
            <select
              value={selectedPackage.status}
              disabled={changingStatus}
              onchange={(event) => updatePackageStatus(selectedPackage, event.currentTarget.value)}
              class="rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-medium text-slate-700 focus:border-primary-400 focus:outline-none focus:ring-2 focus:ring-primary-100 disabled:opacity-60"
            >
              {#each STATUS_OPTIONS as option}
                <option value={option.value}>{option.label}</option>
              {/each}
            </select>
          {/if}
        </div>
        <p class="text-sm text-slate-500">
          Status operasional paket dipakai untuk filter internal dan kontrol kuota.
        </p>
      </div>

      <div class="rounded-2xl border border-slate-100 p-4 space-y-3">
        <h3 class="text-sm font-bold text-slate-800">Info Keberangkatan</h3>
        {@render InfoRow('Tanggal Berangkat', formatDate(selectedPackage.departure_date))}
        {@render InfoRow('Tanggal Pulang', formatDate(selectedPackage.return_date))}
        {@render InfoRow('Maskapai', selectedPackage.airline || '-')}
        {@render InfoRow('Flight Berangkat', selectedPackage.flight_number_go || '-')}
        {@render InfoRow('Flight Pulang', selectedPackage.flight_number_return || '-')}
        {@render InfoRow('Hotel Makkah', selectedPackage.hotel_makkah_name || '-')}
        {@render InfoRow('Hotel Madinah', selectedPackage.hotel_madinah_name || '-')}
        {@render InfoRow('Deskripsi', selectedPackage.description || '-')}
      </div>

      <div class="rounded-2xl border border-slate-100 p-4">
        <h3 class="mb-3 text-sm font-bold text-slate-800">Harga per Tipe Kamar</h3>
        {#if selectedPackage.pricing_tiers?.length > 0}
          <table class="w-full text-sm">
            <thead>
              <tr class="text-left text-xs text-slate-400">
                <th class="pb-2 font-semibold">Tipe</th>
                <th class="pb-2 text-right font-semibold">Harga</th>
                <th class="pb-2 text-right font-semibold">Kuota</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-50">
              {#each ROOM_FIELDS as room}
                {@const tier = selectedPackage.pricing_tiers.find((item) => item.room_type === room.room_type)}
                {#if tier}
                  <tr>
                    <td class="py-2 text-slate-700">{room.label}</td>
                    <td class="py-2 text-right font-semibold text-slate-800">{formatIDR(tier.price)}</td>
                    <td class="py-2 text-right text-slate-600">
                      {#if Number(tier.quota_seats) > 0}
                        <span class:text-red-600={Number(tier.reserved_seats) >= Number(tier.quota_seats)}>{tier.reserved_seats || 0}/{tier.quota_seats}</span>
                      {:else}
                        <span class="text-slate-300">—</span>
                      {/if}
                    </td>
                  </tr>
                {/if}
              {/each}
            </tbody>
          </table>
        {:else}
          <p class="text-sm text-slate-400">Harga kamar belum diatur.</p>
        {/if}
      </div>

      <div class="rounded-2xl border border-slate-100 p-4">
        <h3 class="mb-3 text-sm font-bold text-slate-800">Kuota</h3>
        <div class="mb-2 flex items-end justify-between">
          <span class="text-3xl font-bold text-slate-800">{selectedPackage.reserved_seats}</span>
          <span class="text-slate-400">/ {selectedPackage.total_seats} kursi</span>
        </div>
        <div class="h-2 overflow-hidden rounded-full bg-slate-100">
          <div
            class="h-full rounded-full {selectedPackage.reserved_seats >= selectedPackage.total_seats ? 'bg-red-400' : 'bg-emerald-400'}"
            style="width: {selectedPackage.total_seats > 0 ? Math.min(100, Math.round((selectedPackage.reserved_seats / selectedPackage.total_seats) * 100)) : 0}%"
          ></div>
        </div>
        <p class="mt-1.5 text-xs text-slate-400">{selectedPackage.available_seats} kursi tersisa</p>
      </div>

      <div class="flex flex-wrap gap-3">
        {#if canEditPackages}
          <button
            type="button"
            onclick={openEditForm}
            class="flex items-center justify-center gap-2 rounded-xl border border-slate-200 px-4 py-2.5 text-sm font-semibold text-slate-600 transition-colors hover:bg-slate-50"
          >
            <Edit class="h-4 w-4" />
            Edit Paket
          </button>
        {/if}
        {#if canDeletePackages}
          <button
            type="button"
            onclick={() => deletePackage(selectedPackage)}
            disabled={deletingId === selectedPackage.id}
            class="flex items-center justify-center gap-2 rounded-xl border border-red-200 px-4 py-2.5 text-sm font-semibold text-red-600 transition-colors hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {#if deletingId === selectedPackage.id}
              <Loader2 class="h-4 w-4 animate-spin" />
            {:else}
              <Trash2 class="h-4 w-4" />
            {/if}
            Hapus Paket
          </button>
        {/if}
        <button
          type="button"
          onclick={() => {
            onNavigate?.('crm');
            drawerOpen = false;
          }}
          class="flex items-center justify-center gap-2 rounded-xl bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-primary-700"
        >
          <Users class="h-4 w-4" />
          Lihat CRM
        </button>
      </div>
    </div>
  {/if}
</SlideDrawer>

<SlideDrawer
  open={formDrawerOpen}
  title={formMode === 'create' ? 'Buat Paket Baru' : 'Edit Paket'}
  width="700px"
  onClose={() => (formDrawerOpen = false)}
>
  <div class="space-y-6 p-6">
    {#if formError}
      <div class="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">
        {formError}
      </div>
    {/if}

    <div class="grid gap-4 sm:grid-cols-2">
      <div class="sm:col-span-2">
        <label for="p-name" class="mb-1 block text-sm font-medium text-slate-700">Nama Paket</label>
        <input
          id="p-name"
          type="text"
          bind:value={formState.name}
          placeholder="Contoh: Umroh Reguler Ramadan 2027"
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>

      <div>
        <label for="p-type" class="mb-1 block text-sm font-medium text-slate-700">Jenis Paket</label>
        <select
          id="p-type"
          bind:value={formState.packageType}
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        >
          {#each PACKAGE_TYPES as type}
            <option value={type.value}>{type.label}</option>
          {/each}
        </select>
      </div>

      <div>
        <label for="p-status" class="mb-1 block text-sm font-medium text-slate-700">Status</label>
        <select
          id="p-status"
          bind:value={formState.status}
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        >
          {#each STATUS_OPTIONS as option}
            <option value={option.value}>{option.label}</option>
          {/each}
        </select>
      </div>

      <div>
        <label for="p-dep-date" class="mb-1 block text-sm font-medium text-slate-700">Tanggal Berangkat</label>
        <input
          id="p-dep-date"
          type="date"
          bind:value={formState.departureDate}
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>

      <div>
        <label for="p-ret-date" class="mb-1 block text-sm font-medium text-slate-700">Tanggal Pulang</label>
        <input
          id="p-ret-date"
          type="date"
          bind:value={formState.returnDate}
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>

      <div>
        <label for="p-seats" class="mb-1 block text-sm font-medium text-slate-700">Total Kursi</label>
        <input
          id="p-seats"
          type="number"
          min="1"
          bind:value={formState.totalSeats}
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>

      <div>
        <label for="p-airline" class="mb-1 block text-sm font-medium text-slate-700">Maskapai</label>
        <input
          id="p-airline"
          type="text"
          bind:value={formState.airline}
          placeholder="Contoh: Saudi Airlines"
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>

      <div>
        <label for="p-flight-go" class="mb-1 block text-sm font-medium text-slate-700">Flight Berangkat</label>
        <input
          id="p-flight-go"
          type="text"
          bind:value={formState.flightNumberGo}
          placeholder="Contoh: SV812"
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>

      <div>
        <label for="p-flight-ret" class="mb-1 block text-sm font-medium text-slate-700">Flight Pulang</label>
        <input
          id="p-flight-ret"
          type="text"
          bind:value={formState.flightNumberReturn}
          placeholder="Contoh: SV813"
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>

      <div>
        <label for="p-hotel-mak" class="mb-1 block text-sm font-medium text-slate-700">Hotel Makkah</label>
        <input
          id="p-hotel-mak"
          type="text"
          bind:value={formState.hotelMakkahName}
          placeholder="Contoh: Hilton Makkah"
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>

      <div>
        <label for="p-hotel-mad" class="mb-1 block text-sm font-medium text-slate-700">Hotel Madinah</label>
        <input
          id="p-hotel-mad"
          type="text"
          bind:value={formState.hotelMadinahName}
          placeholder="Contoh: Anwar Al Madinah"
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        />
      </div>

      <div class="sm:col-span-2">
        <label for="p-desc" class="mb-1 block text-sm font-medium text-slate-700">Deskripsi Paket</label>
        <textarea
          id="p-desc"
          rows="4"
          bind:value={formState.description}
          placeholder="Ringkasan itinerary, fasilitas, atau catatan paket"
          class="w-full rounded-xl border border-slate-200 bg-white px-3 py-2.5 text-sm text-slate-800 outline-none transition-colors focus:border-primary-400 focus:ring-2 focus:ring-primary-100"
        ></textarea>
      </div>
    </div>

    <div class="rounded-2xl border border-slate-100 p-4">
      <h3 class="mb-1 text-sm font-bold text-slate-800">Harga &amp; Kuota per Tipe Kamar</h3>
      <p class="mb-4 text-xs text-slate-400">Kuota 0 = tanpa batas tipe kamar (hanya total kursi yang berlaku).</p>
      <div class="grid gap-4 sm:grid-cols-2">
        {#each ROOM_FIELDS as room}
          <div class="rounded-xl border border-slate-100 p-3">
            <p class="mb-2 text-xs font-semibold text-slate-600">{room.label}</p>
            <IDRInput bind:value={formState.prices[room.room_type]} label="Harga" />
            <div class="mt-2 flex flex-col gap-1">
              <label for="q-{room.room_type}" class="text-xs font-medium text-slate-500">Kuota kursi</label>
              <input id="q-{room.room_type}" type="number" min="0" bind:value={formState.quotas[room.room_type]} class="rounded-xl border border-slate-200 px-3 py-2 text-sm outline-none focus:border-primary-400" />
            </div>
          </div>
        {/each}
      </div>
    </div>

    <div class="flex flex-wrap justify-end gap-3 border-t border-slate-100 pt-4">
      <button
        type="button"
        onclick={() => (formDrawerOpen = false)}
        class="rounded-xl border border-slate-200 px-4 py-2.5 text-sm font-semibold text-slate-600 transition-colors hover:bg-slate-50"
      >
        Batal
      </button>
      <button
        type="button"
        onclick={saveForm}
        disabled={savingForm}
        class="inline-flex items-center gap-2 rounded-xl bg-primary-600 px-4 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-primary-700 disabled:cursor-not-allowed disabled:opacity-60"
      >
        {#if savingForm}
          <Loader2 class="h-4 w-4 animate-spin" />
        {/if}
        {formMode === 'create' ? 'Simpan Paket' : 'Perbarui Paket'}
      </button>
    </div>
  </div>
</SlideDrawer>

{#snippet InfoRow(label, value)}
  <div class="flex items-start justify-between gap-4 text-sm">
    <span class="flex-shrink-0 text-slate-400">{label}</span>
    <span class="text-right font-medium text-slate-700">{value}</span>
  </div>
{/snippet}
