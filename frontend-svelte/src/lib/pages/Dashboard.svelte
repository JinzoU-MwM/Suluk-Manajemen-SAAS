<script>
  import { onMount } from "svelte";
  import {
    AlertTriangle,
    ArrowRight,
    Banknote,
    Bell,
    Building2,
    Calendar,
    CheckCircle,
    ChevronRight,
    ClipboardCheck,
    CreditCard,
    DollarSign,
    FileText,
    Hotel,
    Loader2,
    Package,
    Receipt,
    ScanLine,
    Search,
    TrendingUp,
    UserPlus,
    Users,
  } from "lucide-svelte";
  import { ApiService } from "../services/api";
  import { formatRupiah, formatDate, formatPct } from "../utils/formatting.js";

  let { user = null, subscription = null, onNavigate = null } = $props();

  let stats = $state(null);
  let ownerDash = $state(null);
  let isLoading = $state(true);
  let ownerLoading = $state(false);
  let error = $state("");
  let ownerError = $state("");
  let searchQuery = $state("");
  let unreadCount = $state(0);

  let isOwner = $derived(user?.role === "owner" || user?.role === "admin");

  onMount(async () => {
    try {
      stats = await ApiService.getDashboardStats();
    } catch (e) {
      error = e.message;
    } finally {
      isLoading = false;
    }
    try {
      unreadCount = await ApiService.getUnreadNotificationCount();
    } catch (_) {
      // notifications are non-critical
    }
  });

  // Reactive fetch: triggers whenever user (and thus isOwner) changes after mount.
  // Guards against calling multiple times by checking ownerDash / ownerLoading.
  $effect(() => {
    if (isOwner && ownerDash === null && !ownerLoading) {
      ownerLoading = true;
      ApiService.getOwnerDashboard()
        .then((d) => { ownerDash = d; })
        .catch((e) => { ownerError = e.message; })
        .finally(() => { ownerLoading = false; });
    }
  });

  let isPro = $derived(subscription?.plan === "pro" && subscription?.status !== "expired");
  let displayName = $derived(user?.name || "Admin Travel");
  let displayEmail = $derived(user?.email || "admin@travel.co.id");
  let initials = $derived(
    displayName
      .split(" ")
      .map((part) => part.charAt(0))
      .join("")
      .slice(0, 2)
      .toUpperCase(),
  );

  let gender = $derived(stats?.gender_breakdown || { male: 0, female: 0, unknown: 0 });
  let genderTotal = $derived(gender.male + gender.female + gender.unknown);
  let malePct = $derived(genderTotal > 0 ? Math.round((gender.male / genderTotal) * 100) : 0);
  let femalePct = $derived(genderTotal > 0 ? Math.round((gender.female / genderTotal) * 100) : 0);
  let unknownPct = $derived(Math.max(0, 100 - malePct - femalePct));
  let genderDonut = $derived(
    `conic-gradient(#2563eb 0 ${malePct > 0 ? malePct + 0.5 : 0}%, #ec4899 ${malePct > 0 ? malePct + 0.5 : 0}% ${malePct + femalePct > 0 ? malePct + femalePct + 0.5 : 0}%, #cbd5e1 ${malePct + femalePct > 0 ? malePct + femalePct + 0.5 : 0}% 100%)`,
  );
  let trend = $derived(stats?.monthly_trend || []);
  let trendMax = $derived(Math.max(...trend.map((item) => item.count), 1));
  let revTrend = $derived(ownerDash?.revenue_chart || []);
  let revMax = $derived(Math.max(...revTrend.map((item) => item.total ?? 0), 1));

  let statCards = $derived([
    {
      icon: Users,
      label: "Total Jamaah",
      value: stats?.total_jamaah ?? 0,
      badge: stats?.jamaah_this_month > 0 ? `+${stats.jamaah_this_month} bulan ini` : "Live",
      color: "emerald",
      line: "from-emerald-500 to-emerald-400",
    },
    {
      icon: Building2,
      label: "Total Grup Keberangkatan",
      value: stats?.total_groups ?? 0,
      badge: stats?.groups_this_month > 0 ? `+${stats.groups_this_month} bulan ini` : `${stats?.total_groups ?? 0} grup`,
      color: "primary",
      line: "from-primary-500 to-primary-400",
    },
    {
      icon: CheckCircle,
      label: "Perlengkapan Terpenuhi",
      value: `${stats?.equipment_rate ?? 0}%`,
      badge: "Inventory",
      color: "amber",
      line: "from-amber-500 to-amber-400",
    },
    {
      icon: AlertTriangle,
      label: "Paspor Segera Habis",
      value: stats?.passport_expiring_soon ?? 0,
      badge: "90 hari",
      color: (stats?.passport_expiring_soon ?? 0) > 0 ? "red" : "slate",
      line: (stats?.passport_expiring_soon ?? 0) > 0 ? "from-red-500 to-red-400" : "from-slate-400 to-slate-300",
    },
  ]);

  let summaryCards = $derived(isOwner && ownerDash ? [
    {
      icon: DollarSign,
      label: "Total Revenue",
      value: formatRupiah(ownerDash.summary?.total_revenue ?? 0),
      badge: "Terkumpul",
      color: "emerald",
      line: "from-emerald-500 to-emerald-400",
    },
    {
      icon: Receipt,
      label: "Piutang",
      value: formatRupiah(ownerDash.summary?.total_piutang ?? 0),
      badge: "Belum dibayar",
      color: "amber",
      line: "from-amber-500 to-amber-400",
    },
    {
      icon: CreditCard,
      label: "Hutang Vendor",
      value: formatRupiah(ownerDash.summary?.total_debt ?? 0),
      badge: "Tagihan",
      color: "red",
      line: "from-red-500 to-red-400",
    },
    {
      icon: Package,
      label: "Paket Aktif",
      value: ownerDash.summary?.total_packages ?? 0,
      badge: ownerDash.summary?.total_jamaah ? `${ownerDash.summary.total_jamaah} jamaah` : "Live",
      color: "primary",
      line: "from-primary-500 to-primary-400",
    },
  ] : []);

  const quickActions = [
    {
      label: "Scan Dokumen",
      desc: "KTP/Paspor -> Excel",
      icon: ScanLine,
      page: "scanner",
      color: "from-primary-500 to-primary-600",
      shadow: "shadow-primary-500/20",
    },
    {
      label: "Auto-Rooming",
      desc: "Alokasi kamar",
      icon: Hotel,
      page: "rooming",
      color: "from-emerald-500 to-emerald-600",
      shadow: "shadow-emerald-500/20",
    },
    {
      label: "Manifest",
      desc: "Akses tim lapangan",
      icon: ClipboardCheck,
      page: "team",
      color: "from-amber-500 to-orange-500",
      shadow: "shadow-amber-500/20",
    },
    {
      label: "Tambah Jamaah",
      desc: "Input atau import",
      icon: UserPlus,
      page: "scanner",
      color: "from-violet-500 to-purple-600",
      shadow: "shadow-violet-500/20",
    },
  ];

  let operationCards = $derived([
    {
      label: "Dokumen selesai",
      value: formatPct(stats?.document_rate ?? 0),
      icon: FileText,
      color: "primary",
    },
    {
      label: "Perlengkapan terpenuhi",
      value: formatPct(stats?.equipment_rate ?? 0),
      icon: Package,
      color: "amber",
    },
  ]);

  function alertTone(type) {
    const tones = {
      passport: "border-amber-200 bg-amber-50 text-amber-900",
      documents: "border-orange-200 bg-orange-50 text-orange-900",
      payments: "border-red-200 bg-red-50 text-red-900",
    };
    return tones[type] || tones.payments;
  }

  function alertIconTone(type) {
    const tones = {
      passport: "text-amber-600",
      documents: "text-orange-600",
      payments: "text-red-600",
    };
    return tones[type] || tones.payments;
  }

  function iconTone(color) {
    const tones = {
      emerald: "bg-emerald-50 text-emerald-600",
      primary: "bg-primary-50 text-primary-600",
      amber: "bg-amber-50 text-amber-600",
      violet: "bg-violet-50 text-violet-600",
      red: "bg-red-50 text-red-600",
      slate: "bg-slate-50 text-slate-500",
    };
    return tones[color] || tones.slate;
  }

  function badgeTone(color) {
    const tones = {
      emerald: "bg-emerald-50 text-emerald-600",
      primary: "bg-primary-50 text-primary-600",
      amber: "bg-amber-50 text-amber-600",
      violet: "bg-violet-50 text-violet-600",
      red: "bg-red-50 text-red-600",
      slate: "bg-slate-100 text-slate-500",
    };
    return tones[color] || tones.slate;
  }
</script>

<div class="min-h-screen bg-slate-50/70">
  <header class="sticky top-16 z-20 flex min-h-[72px] items-center justify-between gap-4 border-b border-slate-200/80 bg-white/80 px-4 backdrop-blur-xl lg:top-0 lg:px-8">
    <div class="min-w-0">
      <h1 class="text-lg font-bold text-slate-900">Dashboard</h1>
      <p class="hidden text-xs text-slate-400 sm:block">Selamat datang kembali, {displayName}</p>
    </div>

    <div class="flex items-center gap-2 sm:gap-3">
      <label class="relative hidden md:flex">
        <Search class="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
        <input type="search" bind:value={searchQuery} onkeydown={(e) => e.key === "Enter" && onNavigate?.("jamaah")} placeholder="Cari jamaah, grup..." class="w-56 rounded-xl border border-slate-200 bg-slate-50 py-2.5 pl-10 pr-4 text-sm outline-none transition-all focus:border-primary-400 focus-visible:outline-2 focus-visible:outline-primary-500 lg:w-72" />
      </label>
      <button type="button" class="flex h-10 w-10 items-center justify-center rounded-xl text-slate-500 transition-colors hover:bg-slate-100 focus-visible:outline-2 focus-visible:outline-primary-500 md:hidden" aria-label="Cari">
        <Search class="h-5 w-5" />
      </button>
      <button type="button" class="relative flex h-10 w-10 items-center justify-center rounded-xl text-slate-500 transition-colors hover:bg-slate-100 focus-visible:outline-2 focus-visible:outline-primary-500" aria-label="Notifikasi">
        <Bell class="h-5 w-5" />
        {#if unreadCount > 0}
          <span class="badge-pulse absolute right-1.5 top-1.5 h-2.5 w-2.5 rounded-full border-2 border-white bg-red-500"></span>
        {/if}
      </button>
      <div class="hidden h-8 w-px bg-slate-200 sm:block"></div>
      <div class="flex cursor-pointer items-center gap-3 rounded-xl px-2 py-1.5 transition-colors hover:bg-slate-50 focus-visible:outline-2 focus-visible:outline-primary-500" tabindex="0" role="button" aria-label="Profil">
        <div class="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-primary-500 to-primary-600 text-sm font-bold text-white shadow-lg shadow-primary-500/20">{initials}</div>
        <div class="hidden min-w-0 text-left sm:block">
          <p class="truncate text-sm font-semibold leading-none text-slate-800">{displayName}</p>
          <p class="mt-1 truncate text-[11px] text-slate-400">{displayEmail}</p>
        </div>
      </div>
    </div>
  </header>

  <div class="space-y-6 p-4 lg:p-8">
    <div class="rounded-3xl border border-slate-100 bg-white p-5 shadow-sm">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <p class="text-xs font-bold uppercase tracking-wider text-primary-600">Ringkasan Operasional</p>
          <h2 class="mt-1 text-xl font-extrabold text-slate-900">Pantau semua progres jamaah hari ini.</h2>
          <p class="mt-1 text-sm text-slate-500">Data jamaah, grup, dokumen, dan kesiapan lapangan dalam satu dashboard.</p>
        </div>
        <span class="w-fit rounded-full bg-emerald-50 px-3 py-1 text-xs font-bold uppercase text-emerald-700">{isPro ? "Pro Active" : "Starter"}</span>
      </div>
    </div>
    <!-- ── Owner / Admin section (independent of operational stats) ── -->
    {#if isOwner}
      {#if ownerLoading}
        <div class="flex items-center justify-center rounded-2xl border border-slate-100 bg-white py-8 text-slate-400 shadow-sm">
          <Loader2 class="mr-2 h-5 w-5 animate-spin" /> Memuat data keuangan...
        </div>
      {:else if ownerError}
        <div class="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">{ownerError}</div>
      {:else if ownerDash}
        <!-- Partial data warning: some upstream services failed -->
        {#if ownerDash.partial}
          <div class="flex items-start gap-3 rounded-2xl border border-amber-200 bg-amber-50 p-4 text-amber-900">
            <AlertTriangle class="mt-0.5 h-5 w-5 flex-shrink-0 text-amber-600" />
            <div>
              <p class="text-sm font-bold">Sebagian data keuangan belum dapat dimuat.</p>
              <p class="text-xs text-amber-700">Beberapa layanan sedang tidak tersedia, jadi angka di bawah mungkin belum lengkap. Coba muat ulang beberapa saat lagi.</p>
            </div>
          </div>
        {/if}

        <!-- Financial summary cards -->
        {#if summaryCards.length > 0}
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4 lg:gap-5">
            {#each summaryCards as card}
              {@const CardIcon = card.icon}
              <div class={`flex h-full flex-col relative overflow-hidden rounded-2xl border border-slate-100 bg-white p-5 shadow-sm before:absolute before:left-0 before:right-0 before:top-0 before:h-[3px] before:bg-gradient-to-r before:${card.line}`}>
                <div class="mb-3 flex items-center justify-between">
                  <div class={`flex h-11 w-11 items-center justify-center rounded-xl ${iconTone(card.color)}`}>
                    <CardIcon class="h-6 w-6" />
                  </div>
                  <span class={`rounded-full px-2 py-0.5 text-xs font-semibold ${badgeTone(card.color)}`}>{card.badge}</span>
                </div>
                <p class="text-2xl font-extrabold text-slate-900">{card.value}</p>
                <p class="mt-0.5 text-xs text-slate-400">{card.label}</p>
              </div>
            {/each}
          </div>
        {/if}

        <!-- Alert banners -->
        {#if ownerDash.alerts?.passport_expiring_soon > 0}
          <div class="flex flex-col gap-3 rounded-2xl border border-amber-200 bg-amber-50 p-4 text-amber-900 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex items-start gap-3">
              <AlertTriangle class="mt-0.5 h-5 w-5 flex-shrink-0 text-amber-600" />
              <div>
                <p class="text-sm font-bold">{ownerDash.alerts.passport_expiring_soon} paspor perlu dicek sebelum 90 hari.</p>
                <p class="text-xs text-amber-700">Prioritaskan verifikasi dokumen sebelum export final atau keberangkatan.</p>
              </div>
            </div>
            <button type="button" onclick={() => onNavigate?.("scanner")} class="rounded-xl bg-amber-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-amber-700 focus-visible:outline-2 focus-visible:outline-primary-500">Cek Dokumen</button>
          </div>
        {/if}
        {#if ownerDash.alerts?.incomplete_documents > 0}
          <div class="flex flex-col gap-3 rounded-2xl border border-orange-200 bg-orange-50 p-4 text-orange-900 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex items-start gap-3">
              <FileText class="mt-0.5 h-5 w-5 flex-shrink-0 text-orange-600" />
              <div>
                <p class="text-sm font-bold">{ownerDash.alerts.incomplete_documents} jamaah belum melengkapi dokumen.</p>
                <p class="text-xs text-orange-700">Hubungi jamaah untuk melengkapi persyaratan sebelum keberangkatan.</p>
              </div>
            </div>
            <button type="button" onclick={() => onNavigate?.("jamaah")} class="rounded-xl bg-orange-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-orange-700 focus-visible:outline-2 focus-visible:outline-primary-500">Lihat Jamaah</button>
          </div>
        {/if}
        {#if ownerDash.alerts?.overdue_payments > 0}
          <div class="flex flex-col gap-3 rounded-2xl border border-red-200 bg-red-50 p-4 text-red-900 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex items-start gap-3">
              <Banknote class="mt-0.5 h-5 w-5 flex-shrink-0 text-red-600" />
              <div>
                <p class="text-sm font-bold">{ownerDash.alerts.overdue_payments} pembayaran melewati jatuh tempo.</p>
                <p class="text-xs text-red-700">Segera tagih jamaah untuk menghindari pembatalan.</p>
              </div>
            </div>
            <button type="button" onclick={() => onNavigate?.("invoices")} class="rounded-xl bg-red-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-red-700 focus-visible:outline-2 focus-visible:outline-primary-500">Cek Invoice</button>
          </div>
        {/if}
        {#if !ownerDash.alerts?.passport_expiring_soon && !ownerDash.alerts?.incomplete_documents && !ownerDash.alerts?.overdue_payments}
          <div class="flex items-center gap-3 rounded-2xl border border-emerald-200 bg-emerald-50 p-4 text-emerald-900">
            <CheckCircle class="h-5 w-5 flex-shrink-0 text-emerald-600" />
            <p class="text-sm font-bold">Semua beres! Tidak ada peringatan yang perlu ditindaklanjuti saat ini.</p>
          </div>
        {/if}

        <!-- Active packages table -->
        {#if ownerDash.active_packages?.length > 0}
          <section class="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm">
            <div class="mb-5 flex items-center justify-between">
              <div>
                <h2 class="text-base font-bold text-slate-900">Paket Aktif</h2>
                <p class="mt-0.5 text-xs text-slate-400">Progress pembayaran per paket</p>
              </div>
              <Package class="h-5 w-5 text-slate-300" />
            </div>
            <div class="overflow-x-auto">
              <table class="w-full text-left text-sm">
                <thead>
                  <tr class="border-b border-slate-100">
                    <th class="pb-3 pr-4 font-semibold text-slate-500">Paket</th>
                    <th class="pb-3 pr-4 font-semibold text-slate-500">Kuota</th>
                    <th class="pb-3 pr-4 font-semibold text-slate-500">Terdaftar</th>
                    <th class="pb-3 pr-4 font-semibold text-slate-500">Progress Bayar</th>
                    <th class="pb-3 pr-4 font-semibold text-slate-500">Terkumpul</th>
                    <th class="pb-3"></th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-50">
                  {#each ownerDash.active_packages as pkg}
                    <tr class="transition-colors hover:bg-slate-50/50">
                      <td class="py-3 pr-4"><p class="font-semibold text-slate-800">{pkg.name}</p></td>
                      <td class="py-3 pr-4 text-slate-600">{pkg.total_seats ?? 0}</td>
                      <td class="py-3 pr-4 text-slate-600">{pkg.reserved_seats ?? 0}</td>
                      <td class="py-3 pr-4">
                        <div class="flex items-center gap-2">
                          <div class="h-2 w-24 overflow-hidden rounded-full bg-slate-100">
                            <div class="h-full rounded-full bg-gradient-to-r from-primary-500 to-primary-400" style:width={`${formatPct(pkg.payment_pct)}%`}></div>
                          </div>
                          <span class="text-xs font-semibold text-slate-600">{formatPct(pkg.payment_pct)}%</span>
                        </div>
                      </td>
                      <td class="py-3 pr-4 text-sm font-semibold text-slate-800">{formatRupiah(pkg.paid ?? 0)}</td>
                      <td class="py-3">
                        <button type="button" onclick={() => onNavigate?.("invoices")} class="rounded-lg p-1.5 text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600">
                          <ChevronRight class="h-4 w-4" />
                        </button>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          </section>
        {:else if ownerDash.summary?.total_packages === 0}
          <section class="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm">
            <div class="flex flex-col items-center py-6 text-center">
              <Package class="mb-3 h-10 w-10 text-slate-300" />
              <p class="text-sm font-semibold text-slate-600">Belum ada paket aktif</p>
              <p class="mt-1 text-xs text-slate-400">Buat paket umroh atau haji pertama Anda untuk mulai mengelola pendaftaran.</p>
            </div>
          </section>
        {/if}

        <!-- Monthly revenue chart (revenue_chart field from backend) -->
        {#if revTrend.length > 1}
          <section class="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm">
            <div class="mb-5 flex items-center justify-between gap-3">
              <div>
                <h2 class="text-base font-bold text-slate-900">Tren Revenue</h2>
                <p class="mt-0.5 text-xs text-slate-400">Pendapatan 6 bulan terakhir</p>
              </div>
              <span class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-600"><TrendingUp class="h-3.5 w-3.5" /> Keuangan</span>
            </div>
            <div class="flex h-64 items-end gap-3 rounded-2xl bg-slate-50 p-4">
              {#each revTrend as item}
                {@const revVal = item.total ?? 0}
                <div class="flex h-full min-w-0 flex-1 flex-col justify-end gap-2">
                  <div class="relative flex flex-1 items-end">
                    <div class="w-full rounded-t-xl bg-gradient-to-t from-emerald-600 to-emerald-400 shadow-lg shadow-emerald-500/10" style:height={`${Math.max(8, Math.round((revVal / revMax) * 100))}%`} title={`${item.month ?? ""}: ${formatRupiah(revVal)}`}></div>
                  </div>
                  <div class="text-center">
                    <p class="truncate text-[10px] font-medium text-slate-400">{item.month ?? ""}</p>
                    <p class="text-[10px] font-bold text-slate-700">{formatRupiah(revVal)}</p>
                  </div>
                </div>
              {/each}
            </div>
          </section>
        {/if}
      {/if}
    {/if}

    <!-- ── Operational stats section ── -->
    {#if isLoading}
      <div class="flex min-h-[360px] items-center justify-center rounded-3xl border border-slate-200 bg-white text-slate-400 shadow-sm">
        <Loader2 class="mr-2 h-6 w-6 animate-spin" /> Memuat statistik...
      </div>
    {:else if error}
      <div class="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-600">{error}</div>
    {:else if stats}
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4 lg:gap-4">
        {#each quickActions as action}
          {@const ActionIcon = action.icon}
          <button type="button" onclick={() => onNavigate?.(action.page)} class="quick-action flex min-h-[92px] items-center gap-3 rounded-2xl border border-slate-100 bg-white p-4 text-left shadow-sm focus-visible:outline-2 focus-visible:outline-primary-500">
            <div class={`flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br ${action.color} shadow-lg ${action.shadow}`}>
              <ActionIcon class="h-5 w-5 text-white" />
            </div>
            <div class="min-w-0">
              <p class="truncate text-sm font-bold text-slate-800">{action.label}</p>
              <p class="truncate text-[11px] text-slate-400">{action.desc}</p>
            </div>
          </button>
        {/each}
      </div>

      <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4 lg:gap-5">
        {#each statCards as card}
          {@const CardIcon = card.icon}
          <div class={`flex h-full flex-col relative overflow-hidden rounded-2xl border border-slate-100 bg-white p-5 shadow-sm before:absolute before:left-0 before:right-0 before:top-0 before:h-[3px] before:bg-gradient-to-r before:${card.line}`}>
            <div class="mb-3 flex items-center justify-between">
              <div class={`flex h-11 w-11 items-center justify-center rounded-xl ${iconTone(card.color)}`}>
                <CardIcon class="h-6 w-6" />
              </div>
              <span class={`rounded-full px-2 py-0.5 text-xs font-semibold ${badgeTone(card.color)}`}>{card.badge}</span>
            </div>
            <p class="text-2xl font-extrabold text-slate-900">{card.value}</p>
            <p class="mt-0.5 text-xs text-slate-400">{card.label}</p>
          </div>
        {/each}
      </div>

      {#if stats.passport_expiring_soon > 0 && !isOwner}
        <div class="flex flex-col gap-3 rounded-2xl border border-amber-200 bg-amber-50 p-4 text-amber-900 sm:flex-row sm:items-center sm:justify-between">
          <div class="flex items-start gap-3">
            <AlertTriangle class="mt-0.5 h-5 w-5 flex-shrink-0 text-amber-600" />
            <div>
              <p class="text-sm font-bold">{stats.passport_expiring_soon} paspor perlu dicek sebelum 90 hari.</p>
              <p class="text-xs text-amber-700">Prioritaskan verifikasi dokumen sebelum export final atau keberangkatan.</p>
            </div>
          </div>
          <button type="button" onclick={() => onNavigate?.("scanner")} class="rounded-xl bg-amber-600 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-amber-700 focus-visible:outline-2 focus-visible:outline-primary-500">Cek Dokumen</button>
        </div>
      {/if}

      <div class="grid grid-cols-1 gap-5 xl:grid-cols-3">
        <section class="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm xl:col-span-2">
          <div class="mb-5 flex items-center justify-between gap-3">
            <div>
              <h2 class="text-base font-bold text-slate-900">Tren Pendaftaran</h2>
              <p class="mt-0.5 text-xs text-slate-400">6 bulan terakhir</p>
            </div>
            <span class="inline-flex items-center gap-1 rounded-full bg-primary-50 px-3 py-1 text-xs font-semibold text-primary-600"><TrendingUp class="h-3.5 w-3.5" /> Live</span>
          </div>
          {#if trend.length >= 2}
            <div class="flex h-64 items-end gap-3 rounded-2xl bg-slate-50 p-4">
              {#each trend as item}
                <div class="flex h-full min-w-0 flex-1 flex-col justify-end gap-2">
                  <div class="relative flex flex-1 items-end">
                    <div class="w-full rounded-t-xl bg-gradient-to-t from-primary-600 to-primary-400 shadow-lg shadow-primary-500/10" style:height={`${Math.max(8, Math.round((item.count / trendMax) * 100))}%`} title={`${item.label}: ${item.count} jamaah`}></div>
                  </div>
                  <div class="text-center">
                    <p class="truncate text-[10px] font-medium text-slate-400">{item.label}</p>
                    <p class="text-xs font-bold text-slate-700">{item.count}</p>
                  </div>
                </div>
              {/each}
            </div>
          {:else}
            <div class="flex h-64 items-center justify-center rounded-2xl bg-slate-50 text-sm text-slate-400">Belum cukup data untuk menampilkan tren.</div>
          {/if}
        </section>

        <section class="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm">
          <div class="mb-5">
            <h2 class="text-base font-bold text-slate-900">Komposisi Gender</h2>
            <p class="mt-0.5 text-xs text-slate-400">Distribusi jamaah</p>
          </div>
          <div class="flex flex-col items-center">
            {#if genderTotal > 0}
              <div class="relative flex h-44 w-44 items-center justify-center rounded-full" style:background={genderDonut}>
                <div class="flex h-28 w-28 flex-col items-center justify-center rounded-full bg-white shadow-inner">
                  <p class="text-2xl font-extrabold text-slate-900">{genderTotal}</p>
                  <p class="text-[10px] text-slate-400">Total</p>
                </div>
              </div>
              <div class="mt-5 w-full space-y-3">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2"><span class="h-3 w-3 rounded-full bg-primary-500"></span><span class="text-sm text-slate-600">Laki-laki</span></div>
                  <span class="text-sm font-bold text-slate-800">{gender.male} <span class="font-medium text-slate-400">({malePct}%)</span></span>
                </div>
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2"><span class="h-3 w-3 rounded-full bg-pink-500"></span><span class="text-sm text-slate-600">Perempuan</span></div>
                  <span class="text-sm font-bold text-slate-800">{gender.female} <span class="font-medium text-slate-400">({femalePct}%)</span></span>
                </div>
                {#if gender.unknown > 0}
                  <div class="flex items-center justify-between">
                    <div class="flex items-center gap-2"><span class="h-3 w-3 rounded-full bg-slate-300"></span><span class="text-sm text-slate-600">Belum diisi</span></div>
                    <span class="text-sm font-bold text-slate-800">{gender.unknown} <span class="font-medium text-slate-400">({unknownPct}%)</span></span>
                  </div>
                {/if}
              </div>
            {:else}
              <div class="flex h-44 w-44 items-center justify-center rounded-full border-2 border-dashed border-slate-200 bg-slate-50">
                <Users class="h-8 w-8 text-slate-300" />
              </div>
              <p class="mt-3 text-center text-xs text-slate-400">Belum ada data jamaah untuk ditampilkan.</p>
            {/if}
          </div>
        </section>
      </div>

      <div class="grid grid-cols-1 gap-5 xl:grid-cols-3">
        <section class="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm xl:col-span-2">
          <div class="mb-5 flex items-center justify-between">
            <div>
              <h2 class="text-base font-bold text-slate-900">Kesiapan Operasional</h2>
              <p class="mt-0.5 text-xs text-slate-400">Ringkasan modul utama dari referensi dashboard.</p>
            </div>
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            {#each operationCards as item}
              {@const OperationIcon = item.icon}
              <div class="rounded-2xl border border-slate-100 bg-slate-50/70 p-4">
                <div class="mb-3 flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <div class={`flex h-9 w-9 items-center justify-center rounded-xl ${iconTone(item.color)}`}><OperationIcon class="h-[18px] w-[18px]" /></div>
                    <p class="text-sm font-bold text-slate-800">{item.label}</p>
                  </div>
                  <span class={`rounded-full px-2 py-0.5 text-xs font-semibold ${badgeTone(item.color)}`}>{item.value}%</span>
                </div>
                <div class="h-2 overflow-hidden rounded-full bg-white">
                  <div class={`h-full rounded-full bg-gradient-to-r ${item.color === "primary" ? "from-primary-500 to-primary-400" : item.color === "emerald" ? "from-emerald-500 to-emerald-400" : item.color === "amber" ? "from-amber-500 to-amber-400" : "from-violet-500 to-violet-400"}`} style:width={`${item.value}%`}></div>
                </div>
              </div>
            {/each}
          </div>
        </section>

        <section class="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm">
          <div class="mb-4 flex items-center justify-between">
            <div>
              <h2 class="text-base font-bold text-slate-900">Grup Terbaru</h2>
              <p class="mt-0.5 text-xs text-slate-400">Aktivitas terakhir</p>
            </div>
            <Building2 class="h-5 w-5 text-slate-300" />
          </div>
          {#if stats.recent_groups?.length > 0}
            <div class="divide-y divide-slate-100">
              {#each stats.recent_groups as group}
                <div class="flex items-center justify-between gap-3 py-3 first:pt-0 last:pb-0">
                  <div class="min-w-0">
                    <p class="truncate text-sm font-semibold text-slate-700">{group.name}</p>
                    <p class="text-xs text-slate-400">{formatDate(group.created_at)}</p>
                  </div>
                  <span class="rounded-lg bg-primary-50 px-2 py-1 text-xs font-semibold text-primary-600">{group.member_count} jamaah</span>
                </div>
              {/each}
            </div>
          {:else}
            <div class="rounded-2xl bg-slate-50 p-6 text-center text-sm text-slate-400">Belum ada grup keberangkatan.</div>
          {/if}
        </section>
      </div>
    {/if}
  </div>
</div>

<style>
  .quick-action {
    transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  }

  .quick-action:hover {
    transform: translateY(-4px);
    box-shadow: 0 12px 24px -8px rgba(15, 23, 42, 0.16);
  }

  .badge-pulse {
    animation: pulse-ring 2s infinite;
  }

  @keyframes pulse-ring {
    0% {
      box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.4);
    }
    70% {
      box-shadow: 0 0 0 6px rgba(239, 68, 68, 0);
    }
    100% {
      box-shadow: 0 0 0 0 rgba(239, 68, 68, 0);
    }
  }
</style>
