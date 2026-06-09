<script>
    import { onMount } from "svelte";
    import {
        User,
        UserCheck,
        Crown,
        Shield,
        Lock,
        Eye,
        Check,
        CheckCircle,
        AlertCircle,
        Clock,
        Zap,
        Trash2,
        Activity,
        Bell,
        Moon,
        Sun,
        FolderOpen,
        BarChart3,
        ChevronRight,
        Sparkles,
        Loader2,
    } from "lucide-svelte";
    import { ApiService } from "../services/api";
    import { planMeta, limitLabel, isProOrHigher } from "../config/pricing.js";
    import PageHeader from "../components/PageHeader.svelte";
    import EmptyState from "../components/EmptyState.svelte";
    import Card from "../components/ui/Card.svelte";
    import Badge from "../components/ui/Badge.svelte";
    import Button from "../components/ui/Button.svelte";
    import ProgressBar from "../components/ui/ProgressBar.svelte";

    let {
        onLogout,
        onUpgradeRequest,
        user = null,
        subscription: initialSubscription = null,
        trialAvailable: initialTrialAvailable = false,
        groups: initialGroups = [],
    } = $props();

    $effect(() => {
        document.title = "Profil - Jamaah.in";
    });

    // Profile state
    let profile = $state(null);
    let subscription = $state(null);
    let trialStatus = $state(null);
    let loading = $state(true);
    let groupCount = $state(0);

    // Sync initial props to state (silences state_referenced_locally warning)
    $effect.pre(() => {
        subscription ??= initialSubscription;
        if (initialTrialAvailable && !trialStatus) {
            trialStatus = { can_activate: true };
        }
        if (initialGroups?.length && !groupCount) {
            groupCount = initialGroups.length;
        }
    });

    // Active settings tab
    let tab = $state("profil");
    const TABS = [
        { id: "profil", label: "Profil Saya", icon: UserCheck },
        { id: "keamanan", label: "Keamanan", icon: Eye },
        { id: "langganan", label: "Langganan", icon: Crown },
        { id: "notifikasi", label: "Notifikasi", icon: Bell },
    ];

    // Edit profile
    let editName = $state("");
    let savingProfile = $state(false);
    let profileMsg = $state({ type: "", text: "" });

    // Avatar color
    let selectedColor = $state("blue");
    const avatarColors = [
        { name: "emerald", hex: "#1B7F5A" },
        { name: "blue", hex: "#2563c9" },
        { name: "purple", hex: "#7a5ae0" },
        { name: "rose", hex: "#c0392b" },
        { name: "amber", hex: "#C99A2E" },
        { name: "cyan", hex: "#0f7a5a" },
        { name: "indigo", hex: "#15564a" },
        { name: "slate", hex: "#6b7d77" },
    ];

    // Change password
    let currentPassword = $state("");
    let newPassword = $state("");
    let confirmPassword = $state("");
    let savingPassword = $state(false);
    let passwordMsg = $state({ type: "", text: "" });

    // Notification prefs
    let notifyUsageLimit = $state(true);
    let notifyExpiry = $state(true);
    let savingNotifs = $state(false);
    let notifMsg = $state({ type: "", text: "" });

    // Dark mode
    let darkMode = $state(false);

    // Activity log
    let activities = $state([]);
    let loadingActivity = $state(true);

    // Delete account
    let showDeleteModal = $state(false);
    let deletePassword = $state("");
    let deleting = $state(false);
    let deleteError = $state("");

    onMount(async () => {
        darkMode = localStorage.getItem("darkMode") === "true";
        applyDarkMode();

        try {
            const [me, actData] = await Promise.all([
                ApiService.getMe(),
                ApiService.getActivity().catch(() => ({ activities: [] })),
            ]);
            profile = me;
            editName = me.name;
            selectedColor = me.avatar_color || "blue";
            notifyUsageLimit =
                me.notify_usage_limit !== undefined
                    ? me.notify_usage_limit
                    : true;
            notifyExpiry =
                me.notify_expiry !== undefined ? me.notify_expiry : true;
            activities = actData.activities || [];

            if (!subscription) {
                subscription = await ApiService.getSubscriptionStatus();
            }
            if (!trialStatus) {
                try {
                    trialStatus = await ApiService.getTrialStatus();
                } catch {}
            }
        } catch (e) {
            console.error("Failed to load profile:", e);
        } finally {
            loading = false;
            loadingActivity = false;
        }

        if (!groupCount) {
            try {
                const groupData = await ApiService.listGroups();
                groupCount = groupData.groups?.length || 0;
            } catch {}
        }
    });

    function applyDarkMode() {
        if (darkMode) {
            document.documentElement.classList.add("dark");
        } else {
            document.documentElement.classList.remove("dark");
        }
    }

    function toggleDarkMode() {
        darkMode = !darkMode;
        localStorage.setItem("darkMode", darkMode.toString());
        applyDarkMode();
    }

    async function saveProfile() {
        if (!editName.trim()) return;
        savingProfile = true;
        profileMsg = { type: "", text: "" };
        try {
            await ApiService.updateProfile({ name: editName.trim() });
            profile.name = editName.trim();
            const stored = JSON.parse(localStorage.getItem("user") || "{}");
            stored.name = editName.trim();
            localStorage.setItem("user", JSON.stringify(stored));
            profileMsg = {
                type: "success",
                text: "Profil berhasil diperbarui!",
            };
        } catch (e) {
            profileMsg = { type: "error", text: e.message };
        } finally {
            savingProfile = false;
        }
    }

    async function selectAvatarColor(color) {
        selectedColor = color;
        try {
            await ApiService.updateProfile({ avatar_color: color });
            if (profile) profile.avatar_color = color;
        } catch (e) {
            console.error("Failed to save avatar color:", e);
        }
    }

    async function saveNotificationPrefs() {
        savingNotifs = true;
        notifMsg = { type: "", text: "" };
        try {
            await ApiService.updateProfile({
                notify_usage_limit: notifyUsageLimit,
                notify_expiry: notifyExpiry,
            });
            notifMsg = {
                type: "success",
                text: "Preferensi notifikasi disimpan!",
            };
        } catch (e) {
            notifMsg = { type: "error", text: e.message };
        } finally {
            savingNotifs = false;
        }
    }

    async function savePassword() {
        passwordMsg = { type: "", text: "" };
        if (newPassword.length < 6) {
            passwordMsg = {
                type: "error",
                text: "Password baru minimal 6 karakter",
            };
            return;
        }
        if (newPassword !== confirmPassword) {
            passwordMsg = {
                type: "error",
                text: "Konfirmasi password tidak cocok",
            };
            return;
        }
        savingPassword = true;
        try {
            await ApiService.changePassword(currentPassword, newPassword);
            passwordMsg = {
                type: "success",
                text: "Password berhasil diubah!",
            };
            currentPassword = "";
            newPassword = "";
            confirmPassword = "";
        } catch (e) {
            passwordMsg = { type: "error", text: e.message };
        } finally {
            savingPassword = false;
        }
    }

    async function confirmDeleteAccount() {
        deleting = true;
        deleteError = "";
        try {
            await ApiService.deleteAccount(deletePassword);
            localStorage.clear();
            onLogout();
        } catch (e) {
            deleteError = e.message;
        } finally {
            deleting = false;
        }
    }

    function formatDate(iso) {
        if (!iso) return "-";
        const d = new Date(iso);
        return d.toLocaleDateString("id-ID", {
            day: "numeric",
            month: "short",
            year: "numeric",
        });
    }

    function formatDateTime(iso) {
        if (!iso) return "-";
        const d = new Date(iso);
        return d.toLocaleDateString("id-ID", {
            day: "numeric",
            month: "short",
            year: "numeric",
            hour: "2-digit",
            minute: "2-digit",
        });
    }

    function actionLabel(action) {
        if (action === "document_scan") return "Scan Dokumen";
        return action;
    }

    let planLabel = $derived(planMeta(subscription?.plan).name + " Plan");
    let usagePercent = $derived(
        subscription?.usage_limit
            ? Math.min(
                  100,
                  Math.round(
                      (subscription.usage_count / subscription.usage_limit) *
                          100,
                  ),
              )
            : 0,
    );
    let usageColor = $derived(
        usagePercent > 80
            ? "var(--c-danger)"
            : usagePercent > 50
              ? "var(--c-warning)"
              : "var(--c-primary)",
    );
    let avatarHex = $derived(
        avatarColors.find((c) => c.name === selectedColor)?.hex || "#2563c9",
    );
    let groupLimit = $derived(
        limitLabel(
            subscription?.max_groups ?? planMeta(subscription?.plan).maxGroups,
        ),
    );
</script>

<div class="profile-page">
    {#if loading}
        <div class="loading-container">
            <Loader2 class="h-8 w-8 animate-spin" style="color:var(--c-primary)" />
        </div>
    {:else if profile}
        <PageHeader
            kicker="Pengaturan Akun"
            title="Profil Saya"
            subtitle="Kelola informasi pribadi, langganan, keamanan, dan preferensi akun Anda."
        >
            {#snippet actions()}
                <Button variant="ghost" icon={darkMode ? Sun : Moon} onclick={toggleDarkMode}>
                    {darkMode ? "Mode Terang" : "Mode Gelap"}
                </Button>
            {/snippet}
        </PageHeader>

        <!-- Mobile tab pills -->
        <div class="mobile-tabs">
            <div class="mobile-tabs-inner">
                {#each TABS as t}
                    <button
                        type="button"
                        class="mobile-tab {tab === t.id ? 'mobile-tab-active' : ''}"
                        onclick={() => (tab = t.id)}
                    >
                        {t.label}
                    </button>
                {/each}
            </div>
        </div>

        <div class="profile-grid">
            <!-- Left: Summary + Nav -->
            <div class="profile-col-left">
                <Card pad={false} style="overflow:hidden">
                    <div class="summary-banner">
                        <Sparkles
                            size={90}
                            class="banner-spark"
                        />
                    </div>
                    <div class="summary-body">
                        <div class="summary-avatar-wrap">
                            <div class="summary-avatar-ring" style="background:{avatarHex}1f">
                                <span class="summary-avatar-initial" style="color:{avatarHex}">
                                    {profile.name?.charAt(0)?.toUpperCase() || "U"}
                                </span>
                            </div>
                        </div>
                        <div class="summary-name">{profile.name}</div>
                        <div class="summary-email">{profile.email}</div>
                        <div class="summary-badges">
                            {#if profile.is_admin}
                                <Badge tone="success" dot>Admin</Badge>
                            {/if}
                            <Badge tone={isProOrHigher(subscription?.plan) ? "warning" : "info"}>
                                {planLabel}
                            </Badge>
                        </div>

                        <div class="summary-stats">
                            <div class="summary-stat">
                                <div class="summary-stat-value tabular">
                                    {profile.usage?.total || 0}{#if profile.usage?.limit}<span
                                            class="summary-stat-limit">/{profile.usage.limit}</span
                                        >{/if}
                                </div>
                                <div class="summary-stat-label">Scan</div>
                            </div>
                            <div class="summary-stat summary-stat-bordered">
                                <div class="summary-stat-value tabular">
                                    {groupCount}<span class="summary-stat-limit">/{groupLimit}</span>
                                </div>
                                <div class="summary-stat-label">Grup</div>
                            </div>
                            <div class="summary-stat summary-stat-bordered">
                                <div class="summary-stat-value tabular">
                                    {formatDate(profile.created_at)}
                                </div>
                                <div class="summary-stat-label">Bergabung</div>
                            </div>
                        </div>
                    </div>
                </Card>

                <!-- Desktop nav -->
                <Card pad={false} class="nav-card">
                    <div class="nav-list">
                        {#each TABS as t}
                            {@const active = tab === t.id}
                            <button
                                type="button"
                                class="nav-item {active ? 'nav-item-active' : ''}"
                                onclick={() => (tab = t.id)}
                            >
                                <t.icon size={18} />
                                <span class="nav-item-label">{t.label}</span>
                                {#if active}<ChevronRight size={16} />{/if}
                            </button>
                        {/each}
                    </div>
                </Card>

                {#if user?.is_super_admin}
                    <a href="/#/super-admin" class="superadmin-link">
                        <Shield size={16} />
                        <span>Super Admin Dashboard</span>
                    </a>
                {/if}
            </div>

            <!-- Right: Active panel -->
            <div class="profile-col-right">
                {#if tab === "profil"}
                    <Card>
                        <div class="section-head">
                            <div>
                                <div class="section-title">Informasi Pribadi</div>
                                <div class="section-desc">
                                    Data ini ditampilkan pada akun internal dan dokumen travel.
                                </div>
                            </div>
                        </div>

                        <div class="fields-grid">
                            <div class="field">
                                <label class="field-label" for="profile-email">Email</label>
                                <input
                                    id="profile-email"
                                    type="email"
                                    value={profile.email}
                                    disabled
                                    class="field-input field-disabled"
                                />
                                <p class="field-hint">Email tidak dapat diubah</p>
                            </div>
                            <div class="field">
                                <label class="field-label" for="profile-name">Nama Lengkap</label>
                                <input
                                    id="profile-name"
                                    type="text"
                                    bind:value={editName}
                                    class="field-input"
                                    placeholder="Nama lengkap"
                                />
                            </div>
                        </div>

                        <div class="avatar-picker">
                            <span class="field-label" style="margin:0">Warna Avatar</span>
                            <div class="color-options">
                                {#each avatarColors as color}
                                    <button
                                        type="button"
                                        onclick={() => selectAvatarColor(color.name)}
                                        class="color-dot {selectedColor === color.name ? 'color-dot-active' : ''}"
                                        style="background:{color.hex}"
                                        title={color.name}
                                        aria-label="Pilih warna {color.name}"
                                    ></button>
                                {/each}
                            </div>
                        </div>

                        {#if profileMsg.text}
                            <div class="msg {profileMsg.type === 'success' ? 'msg-success' : 'msg-error'}">
                                {#if profileMsg.type === "success"}
                                    <CheckCircle class="h-4 w-4" />
                                {:else}
                                    <AlertCircle class="h-4 w-4" />
                                {/if}
                                {profileMsg.text}
                            </div>
                        {/if}

                        <div class="section-action">
                            <Button
                                variant="primary"
                                icon={Check}
                                onclick={saveProfile}
                                disabled={savingProfile || editName.trim() === profile.name}
                            >
                                {savingProfile ? "Menyimpan..." : "Simpan Perubahan"}
                            </Button>
                        </div>
                    </Card>

                    <Card>
                        <div class="section-head">
                            <div>
                                <div class="section-title">Aktivitas Terakhir</div>
                                <div class="section-desc">Riwayat scan dokumen terbaru.</div>
                            </div>
                            <div class="section-icon">
                                <Activity size={19} />
                            </div>
                        </div>

                        {#if loadingActivity}
                            <div class="activity-loading">
                                <Loader2 class="h-5 w-5 animate-spin" style="color:var(--c-faint)" />
                            </div>
                        {:else if activities.length === 0}
                            <EmptyState
                                icon={Activity}
                                title="Belum ada aktivitas"
                                text="Aktivitas scan dokumen Anda akan muncul di sini."
                            />
                        {:else}
                            <div class="activity-list">
                                {#each activities as act}
                                    <div class="activity-row">
                                        <div class="activity-icon">
                                            <BarChart3 size={18} />
                                        </div>
                                        <div class="activity-info">
                                            <div class="activity-action">{actionLabel(act.action)}</div>
                                            <div class="activity-count">{act.count} dokumen</div>
                                        </div>
                                        <span class="activity-date">{formatDateTime(act.created_at)}</span>
                                    </div>
                                {/each}
                            </div>
                        {/if}
                    </Card>
                {:else if tab === "keamanan"}
                    <Card>
                        <div class="section-head">
                            <div>
                                <div class="section-title">Kata Sandi</div>
                                <div class="section-desc">
                                    Perbarui sandi secara berkala untuk menjaga keamanan akun.
                                </div>
                            </div>
                            <div class="section-icon">
                                <Lock size={19} />
                            </div>
                        </div>

                        <div class="fields-grid">
                            <div class="field field-full">
                                <label class="field-label" for="current-password">Password Saat Ini</label>
                                <input
                                    id="current-password"
                                    type="password"
                                    bind:value={currentPassword}
                                    class="field-input"
                                    placeholder="Masukkan password saat ini"
                                />
                            </div>
                            <div class="field">
                                <label class="field-label" for="new-password">Password Baru</label>
                                <input
                                    id="new-password"
                                    type="password"
                                    bind:value={newPassword}
                                    class="field-input"
                                    placeholder="Minimal 6 karakter"
                                />
                            </div>
                            <div class="field">
                                <label class="field-label" for="confirm-password">Konfirmasi Password Baru</label>
                                <input
                                    id="confirm-password"
                                    type="password"
                                    bind:value={confirmPassword}
                                    class="field-input"
                                    placeholder="Ulangi password baru"
                                />
                            </div>
                        </div>

                        {#if passwordMsg.text}
                            <div class="msg {passwordMsg.type === 'success' ? 'msg-success' : 'msg-error'}">
                                {#if passwordMsg.type === "success"}
                                    <CheckCircle class="h-4 w-4" />
                                {:else}
                                    <AlertCircle class="h-4 w-4" />
                                {/if}
                                {passwordMsg.text}
                            </div>
                        {/if}

                        <div class="section-action">
                            <Button
                                variant="primary"
                                icon={Check}
                                onclick={savePassword}
                                disabled={savingPassword || !currentPassword || !newPassword || !confirmPassword}
                            >
                                {savingPassword ? "Mengubah..." : "Perbarui Sandi"}
                            </Button>
                        </div>
                    </Card>

                    <Card>
                        <div class="section-head">
                            <div>
                                <div class="section-title">Preferensi Tampilan</div>
                                <div class="section-desc">Sesuaikan tampilan aplikasi sesuai selera.</div>
                            </div>
                        </div>
                        <div class="setting-row" style="border-top:none">
                            <div class="setting-icon">
                                {#if darkMode}<Sun size={19} />{:else}<Moon size={19} />{/if}
                            </div>
                            <div class="setting-text">
                                <div class="setting-title">Mode Gelap</div>
                                <div class="setting-desc">Tampilan dengan latar gelap yang nyaman di mata.</div>
                            </div>
                            <button
                                type="button"
                                onclick={toggleDarkMode}
                                class="toggle {darkMode ? 'toggle-on' : ''}"
                                role="switch"
                                aria-checked={darkMode}
                                aria-label="Toggle Mode Gelap"
                            >
                                <span class="toggle-knob"></span>
                            </button>
                        </div>
                    </Card>

                    <Card class="danger-card">
                        <div class="section-head">
                            <div>
                                <div class="section-title" style="color:var(--c-danger)">Zona Berbahaya</div>
                                <div class="section-desc">
                                    Hapus akun secara permanen. Semua data grup dan riwayat akan dihapus.
                                    Tindakan ini tidak dapat dibatalkan.
                                </div>
                            </div>
                            <div class="section-icon section-icon-danger">
                                <Trash2 size={19} />
                            </div>
                        </div>
                        <div class="section-action">
                            <Button variant="danger" icon={Trash2} onclick={() => (showDeleteModal = true)}>
                                Hapus Akun Saya
                            </Button>
                        </div>
                    </Card>
                {:else if tab === "langganan"}
                    <Card>
                        <div class="section-head">
                            <div>
                                <div class="section-title">Paket Langganan</div>
                                <div class="section-desc">Detail paket aktif dan batas penggunaan akun Anda.</div>
                            </div>
                            <div class="section-icon section-icon-accent">
                                {#if isProOrHigher(subscription?.plan)}
                                    <Crown size={19} />
                                {:else}
                                    <Zap size={19} />
                                {/if}
                            </div>
                        </div>

                        <div class="plan-hero">
                            <div class="plan-hero-icon">
                                {#if isProOrHigher(subscription?.plan)}
                                    <Crown size={26} />
                                {:else}
                                    <Zap size={26} />
                                {/if}
                            </div>
                            <div class="plan-hero-info">
                                <div class="plan-hero-name">{planLabel}</div>
                                <div class="plan-hero-sub">
                                    {planMeta(subscription?.plan).desc}
                                </div>
                            </div>
                            <Badge tone={isProOrHigher(subscription?.plan) ? "warning" : "info"}>
                                {subscription?.status === "trial" ? "Trial" : "Aktif"}
                            </Badge>
                        </div>

                        <div class="limits-grid">
                            <div class="limit-box">
                                <div class="limit-icon"><BarChart3 size={18} /></div>
                                <div>
                                    <div class="limit-value tabular">
                                        {profile.usage?.total || 0}{#if profile.usage?.limit}<span
                                                class="limit-cap">/{profile.usage.limit}</span
                                            >{/if}
                                    </div>
                                    <div class="limit-label">Scan dokumen</div>
                                </div>
                            </div>
                            <div class="limit-box">
                                <div class="limit-icon"><FolderOpen size={18} /></div>
                                <div>
                                    <div class="limit-value tabular">
                                        {groupCount}<span class="limit-cap">/{groupLimit}</span>
                                    </div>
                                    <div class="limit-label">Grup</div>
                                </div>
                            </div>
                            <div class="limit-box">
                                <div class="limit-icon"><User size={18} /></div>
                                <div>
                                    <div class="limit-value tabular">
                                        {limitLabel(
                                            subscription?.max_users ??
                                                planMeta(subscription?.plan).maxUsers,
                                        )}
                                    </div>
                                    <div class="limit-label">Pengguna</div>
                                </div>
                            </div>
                        </div>

                        {#if !isProOrHigher(subscription?.plan) && subscription?.usage_limit}
                            <div class="usage-block">
                                <div class="usage-block-head">
                                    <span class="usage-block-label">Kuota scan terpakai</span>
                                    <span class="usage-block-pct tabular">{usagePercent}%</span>
                                </div>
                                <ProgressBar value={usagePercent} max={100} color={usageColor} />
                            </div>
                        {/if}

                        <div class="sub-details">
                            {#if isProOrHigher(subscription?.plan)}
                                <div class="sub-detail-row">
                                    <span class="sub-detail-label">Berlangganan sejak</span>
                                    <span class="sub-detail-value">{formatDate(subscription.subscribed_at)}</span>
                                </div>
                                <div class="sub-detail-row">
                                    <span class="sub-detail-label">Berlaku hingga</span>
                                    <span class="sub-detail-value">{formatDate(subscription.subscription_ends)}</span>
                                </div>
                            {:else}
                                {#if subscription?.trial_ends && subscription?.status === "trial"}
                                    <div class="sub-detail-row">
                                        <span class="sub-detail-label">
                                            <Clock size={14} style="color:var(--c-warning)" />
                                            Trial berakhir
                                        </span>
                                        <span class="sub-detail-value">{formatDate(subscription.trial_ends)}</span>
                                    </div>
                                {/if}
                            {/if}
                        </div>

                        {#if !isProOrHigher(subscription?.plan)}
                            <div class="section-action upgrade-actions">
                                <Button variant="accent" icon={Crown} full onclick={onUpgradeRequest}>
                                    Upgrade ke Pro - Rp299.000/bulan
                                </Button>
                                {#if trialStatus?.can_activate}
                                    <Button variant="soft" full onclick={onUpgradeRequest}>
                                        Coba Pro 7 Hari Gratis
                                    </Button>
                                {/if}
                            </div>
                        {/if}
                    </Card>
                {:else if tab === "notifikasi"}
                    <Card>
                        <div class="section-head">
                            <div>
                                <div class="section-title">Notifikasi</div>
                                <div class="section-desc">
                                    Pilih kejadian yang ingin Anda terima pemberitahuannya.
                                </div>
                            </div>
                            <div class="section-icon">
                                <Bell size={19} />
                            </div>
                        </div>

                        <div class="setting-list">
                            <div class="setting-row">
                                <div class="setting-icon"><BarChart3 size={19} /></div>
                                <div class="setting-text">
                                    <div class="setting-title">Peringatan batas kuota</div>
                                    <div class="setting-desc">Notifikasi saat mendekati batas scan.</div>
                                </div>
                                <button
                                    type="button"
                                    onclick={() => (notifyUsageLimit = !notifyUsageLimit)}
                                    class="toggle {notifyUsageLimit ? 'toggle-on' : ''}"
                                    role="switch"
                                    aria-checked={notifyUsageLimit}
                                    aria-label="Toggle Peringatan batas kuota"
                                >
                                    <span class="toggle-knob"></span>
                                </button>
                            </div>
                            <div class="setting-row">
                                <div class="setting-icon"><Clock size={19} /></div>
                                <div class="setting-text">
                                    <div class="setting-title">Peringatan masa berlaku</div>
                                    <div class="setting-desc">
                                        Notifikasi sebelum trial/langganan berakhir.
                                    </div>
                                </div>
                                <button
                                    type="button"
                                    onclick={() => (notifyExpiry = !notifyExpiry)}
                                    class="toggle {notifyExpiry ? 'toggle-on' : ''}"
                                    role="switch"
                                    aria-checked={notifyExpiry}
                                    aria-label="Toggle Peringatan masa berlaku"
                                >
                                    <span class="toggle-knob"></span>
                                </button>
                            </div>
                        </div>

                        {#if notifMsg.text}
                            <div class="msg {notifMsg.type === 'success' ? 'msg-success' : 'msg-error'}">
                                {#if notifMsg.type === "success"}
                                    <CheckCircle class="h-4 w-4" />
                                {:else}
                                    <AlertCircle class="h-4 w-4" />
                                {/if}
                                {notifMsg.text}
                            </div>
                        {/if}

                        <div class="section-action">
                            <Button
                                variant="primary"
                                icon={Check}
                                onclick={saveNotificationPrefs}
                                disabled={savingNotifs}
                            >
                                {savingNotifs ? "Menyimpan..." : "Simpan Preferensi"}
                            </Button>
                        </div>
                    </Card>
                {/if}
            </div>
        </div>
    {/if}
</div>

<!-- Delete Account Modal -->
{#if showDeleteModal}
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <div
        class="modal-overlay"
        onclick={() => (showDeleteModal = false)}
        onkeydown={(e) => {
            if (e.key === "Escape") showDeleteModal = false;
        }}
        role="dialog"
        aria-modal="true"
        tabindex="-1"
    >
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div class="modal-content" onclick={(e) => e.stopPropagation()} onkeydown={() => {}}>
            <div class="modal-header">
                <h3 class="modal-title" style="color:var(--c-danger)">
                    <Trash2 class="h-5 w-5" />
                    Hapus Akun
                </h3>
                <button type="button" onclick={() => (showDeleteModal = false)} class="modal-close">×</button>
            </div>
            <div class="modal-body">
                <p class="delete-warning">
                    Tindakan ini <strong>tidak dapat dibatalkan</strong>. Masukkan password untuk
                    konfirmasi.
                </p>
                <input
                    type="password"
                    bind:value={deletePassword}
                    placeholder="Password akun"
                    class="field-input"
                />
                {#if deleteError}
                    <div class="msg msg-error">
                        <AlertCircle class="h-4 w-4" />
                        {deleteError}
                    </div>
                {/if}
                <div class="delete-actions">
                    <Button variant="ghost" onclick={() => (showDeleteModal = false)}>Batal</Button>
                    <Button
                        variant="primary"
                        icon={Trash2}
                        onclick={confirmDeleteAccount}
                        disabled={deleting || !deletePassword}
                        style="background:var(--c-danger);border-color:var(--c-danger)"
                    >
                        {deleting ? "Menghapus..." : "Hapus Permanen"}
                    </Button>
                </div>
            </div>
        </div>
    </div>
{/if}

<style>
    .profile-page {
        padding: 24px;
        background: var(--c-bg);
        min-height: 100vh;
    }
    @media (max-width: 640px) {
        .profile-page {
            padding: 16px;
        }
    }

    .loading-container {
        display: flex;
        justify-content: center;
        align-items: center;
        padding: 6rem 0;
    }

    /* ---- Grid ---- */
    .profile-grid {
        display: grid;
        grid-template-columns: 300px 1fr;
        gap: var(--gap);
        align-items: start;
    }
    @media (max-width: 1024px) {
        .profile-grid {
            grid-template-columns: 1fr;
        }
    }
    .profile-col-left {
        display: flex;
        flex-direction: column;
        gap: var(--gap);
        position: sticky;
        top: 0;
    }
    @media (max-width: 1024px) {
        .profile-col-left {
            position: static;
        }
    }
    .profile-col-right {
        display: flex;
        flex-direction: column;
        gap: var(--gap);
        min-width: 0;
    }

    /* ---- Summary card ---- */
    .summary-banner {
        height: 90px;
        background: linear-gradient(120deg, var(--c-primary-deep), var(--c-primary));
        position: relative;
        overflow: hidden;
    }
    :global(.banner-spark) {
        position: absolute;
        right: -16px;
        top: -16px;
        color: rgba(255, 255, 255, 0.12);
    }
    .summary-body {
        padding: 0 22px 22px;
        margin-top: -42px;
        text-align: center;
    }
    .summary-avatar-wrap {
        width: 88px;
        height: 88px;
        margin: 0 auto 14px;
    }
    .summary-avatar-ring {
        width: 88px;
        height: 88px;
        border-radius: 50%;
        border: 4px solid var(--c-surface);
        display: flex;
        align-items: center;
        justify-content: center;
        box-sizing: border-box;
    }
    .summary-avatar-initial {
        font-size: 32px;
        font-weight: 800;
    }
    .summary-name {
        font-size: 18.5px;
        font-weight: 800;
        color: var(--c-ink);
    }
    .summary-email {
        font-size: 13px;
        color: var(--c-muted);
        margin-top: 3px;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    .summary-badges {
        margin-top: 12px;
        display: flex;
        gap: 6px;
        justify-content: center;
        flex-wrap: wrap;
    }
    .summary-stats {
        display: flex;
        margin-top: 20px;
        padding-top: 18px;
        border-top: 1px solid var(--c-line-soft);
    }
    .summary-stat {
        flex: 1;
        min-width: 0;
        padding: 0 4px;
    }
    .summary-stat-bordered {
        border-left: 1px solid var(--c-line-soft);
    }
    .summary-stat-value {
        font-size: 15px;
        font-weight: 800;
        color: var(--c-ink);
    }
    .summary-stat-limit {
        font-size: 12px;
        font-weight: 500;
        color: var(--c-faint);
    }
    .summary-stat-label {
        font-size: 11px;
        color: var(--c-faint);
        margin-top: 2px;
    }

    /* ---- Nav card ---- */
    :global(.nav-card) {
        display: block;
    }
    @media (max-width: 1024px) {
        :global(.nav-card) {
            display: none;
        }
    }
    .nav-list {
        padding: 10px;
    }
    .nav-item {
        width: 100%;
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 11px 13px;
        border-radius: var(--radius);
        margin-bottom: 2px;
        text-align: left;
        font-size: 14px;
        font-weight: 500;
        color: var(--c-ink-soft);
        background: transparent;
        transition: background 0.14s;
    }
    .nav-item:hover {
        background: var(--c-bg);
    }
    .nav-item-active {
        font-weight: 700;
        color: var(--c-primary-deep);
        background: var(--c-primary-soft);
    }
    .nav-item-label {
        flex: 1;
    }

    /* ---- Super admin link ---- */
    .superadmin-link {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        padding: 11px 16px;
        background: var(--c-accent-soft);
        color: var(--c-accent);
        border-radius: var(--radius);
        text-decoration: none;
        font-size: 14px;
        font-weight: 700;
        transition: filter 0.15s;
    }
    .superadmin-link:hover {
        filter: brightness(0.97);
    }

    /* ---- Mobile tabs ---- */
    .mobile-tabs {
        display: none;
        margin-bottom: var(--gap);
        overflow-x: auto;
    }
    @media (max-width: 1024px) {
        .mobile-tabs {
            display: block;
        }
    }
    .mobile-tabs-inner {
        display: inline-flex;
        gap: 4px;
        background: var(--c-bg-2);
        padding: 4px;
        border-radius: var(--radius);
    }
    .mobile-tab {
        padding: 8px 14px;
        font-size: 13px;
        font-weight: 600;
        border-radius: var(--radius-sm);
        white-space: nowrap;
        color: var(--c-muted);
        background: transparent;
    }
    .mobile-tab-active {
        color: var(--c-ink);
        background: var(--c-surface);
        box-shadow: var(--shadow-sm);
    }

    /* ---- Section head ---- */
    .section-head {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        gap: 16px;
        margin-bottom: 18px;
    }
    .section-title {
        font-size: 16.5px;
        font-weight: 800;
        color: var(--c-ink);
    }
    .section-desc {
        font-size: 13px;
        color: var(--c-muted);
        margin-top: 4px;
        line-height: 1.45;
    }
    .section-icon {
        width: 38px;
        height: 38px;
        border-radius: var(--radius);
        background: var(--c-primary-soft);
        color: var(--c-primary-deep);
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    .section-icon-accent {
        background: var(--c-accent-soft);
        color: var(--c-accent);
    }
    .section-icon-danger {
        background: var(--c-danger-soft);
        color: var(--c-danger);
    }
    .section-action {
        margin-top: 18px;
    }

    /* ---- Fields ---- */
    .fields-grid {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 18px 22px;
    }
    @media (max-width: 560px) {
        .fields-grid {
            grid-template-columns: 1fr;
        }
    }
    .field-full {
        grid-column: 1 / -1;
    }
    .field-label {
        display: block;
        font-size: 11.5px;
        font-weight: 700;
        letter-spacing: 0.04em;
        text-transform: uppercase;
        color: var(--c-faint);
        margin-bottom: 7px;
    }
    .field-input {
        width: 100%;
        padding: 11px 13px;
        font-size: 14px;
        font-weight: 500;
        color: var(--c-ink);
        background: var(--c-surface);
        border: 1px solid var(--c-line);
        border-radius: var(--radius);
        outline: none;
        box-sizing: border-box;
        transition: border-color 0.15s, box-shadow 0.15s;
    }
    .field-input:focus {
        border-color: var(--c-primary);
        box-shadow: 0 0 0 3px var(--c-primary-soft);
    }
    .field-disabled {
        background: var(--c-bg-2);
        color: var(--c-muted);
        cursor: not-allowed;
    }
    .field-hint {
        font-size: 11.5px;
        color: var(--c-faint);
        margin-top: 6px;
    }

    /* ---- Avatar picker ---- */
    .avatar-picker {
        margin-top: 18px;
        padding-top: 18px;
        border-top: 1px solid var(--c-line-soft);
        display: flex;
        align-items: center;
        gap: 16px;
        flex-wrap: wrap;
    }
    .color-options {
        display: flex;
        gap: 8px;
        flex-wrap: wrap;
    }
    .color-dot {
        width: 26px;
        height: 26px;
        border-radius: 50%;
        border: 2px solid transparent;
        cursor: pointer;
        transition: transform 0.15s;
    }
    .color-dot:hover {
        transform: scale(1.15);
    }
    .color-dot-active {
        border-color: var(--c-surface);
        box-shadow: 0 0 0 2px var(--c-ink);
    }

    /* ---- Setting rows ---- */
    .setting-list {
        margin-top: -4px;
    }
    .setting-row {
        display: flex;
        align-items: center;
        gap: 14px;
        padding: 16px 0;
        border-top: 1px solid var(--c-line-soft);
    }
    .setting-icon {
        width: 38px;
        height: 38px;
        border-radius: var(--radius);
        background: var(--c-primary-soft);
        color: var(--c-primary-deep);
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    .setting-text {
        flex: 1;
        min-width: 0;
    }
    .setting-title {
        font-size: 14px;
        font-weight: 700;
        color: var(--c-ink);
    }
    .setting-desc {
        font-size: 12.5px;
        color: var(--c-muted);
        margin-top: 2px;
        line-height: 1.45;
    }

    /* ---- Toggle ---- */
    .toggle {
        width: 46px;
        height: 27px;
        border-radius: 999px;
        padding: 3px;
        background: var(--c-bg-2);
        transition: background 0.2s;
        flex-shrink: 0;
        display: flex;
        justify-content: flex-start;
        border: none;
        cursor: pointer;
    }
    .toggle-on {
        background: var(--c-primary);
        justify-content: flex-end;
    }
    .toggle-knob {
        width: 21px;
        height: 21px;
        border-radius: 999px;
        background: #fff;
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
        transition: all 0.2s;
    }

    /* ---- Plan hero ---- */
    .plan-hero {
        display: flex;
        align-items: center;
        gap: 16px;
        padding: 16px;
        background: var(--c-primary-tint);
        border: 1px solid var(--c-line-soft);
        border-radius: var(--radius-lg);
        margin-bottom: 18px;
    }
    .plan-hero-icon {
        width: 52px;
        height: 52px;
        border-radius: var(--radius-lg);
        background: var(--c-accent-soft);
        color: var(--c-accent);
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    .plan-hero-info {
        flex: 1;
        min-width: 0;
    }
    .plan-hero-name {
        font-size: 16px;
        font-weight: 800;
        color: var(--c-ink);
    }
    .plan-hero-sub {
        font-size: 12.5px;
        color: var(--c-muted);
        margin-top: 3px;
    }

    /* ---- Limits grid ---- */
    .limits-grid {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 12px;
        margin-bottom: 4px;
    }
    @media (max-width: 560px) {
        .limits-grid {
            grid-template-columns: 1fr;
        }
    }
    .limit-box {
        display: flex;
        align-items: center;
        gap: 10px;
        padding: 14px;
        background: var(--c-bg);
        border: 1px solid var(--c-line-soft);
        border-radius: var(--radius);
    }
    .limit-icon {
        width: 36px;
        height: 36px;
        border-radius: var(--radius);
        background: var(--c-primary-soft);
        color: var(--c-primary-deep);
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    .limit-value {
        font-size: 15px;
        font-weight: 800;
        color: var(--c-ink);
        line-height: 1.2;
    }
    .limit-cap {
        font-size: 12px;
        font-weight: 500;
        color: var(--c-faint);
    }
    .limit-label {
        font-size: 11.5px;
        color: var(--c-faint);
        margin-top: 2px;
    }

    /* ---- Usage block ---- */
    .usage-block {
        margin-top: 16px;
    }
    .usage-block-head {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;
    }
    .usage-block-label {
        font-size: 13px;
        color: var(--c-muted);
    }
    .usage-block-pct {
        font-size: 13px;
        font-weight: 700;
        color: var(--c-ink);
    }

    /* ---- Subscription details ---- */
    .sub-details {
        margin-top: 16px;
        display: flex;
        flex-direction: column;
        gap: 10px;
    }
    .sub-detail-row {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 8px;
        font-size: 13.5px;
    }
    .sub-detail-label {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        color: var(--c-muted);
    }
    .sub-detail-value {
        color: var(--c-ink);
        font-weight: 600;
    }
    .upgrade-actions {
        display: flex;
        flex-direction: column;
        gap: 10px;
    }

    /* ---- Messages ---- */
    .msg {
        display: flex;
        align-items: center;
        gap: 8px;
        padding: 9px 13px;
        border-radius: var(--radius);
        font-size: 13px;
        font-weight: 500;
        margin-top: 16px;
    }
    .msg-success {
        background: var(--c-success-soft);
        color: var(--c-success);
    }
    .msg-error {
        background: var(--c-danger-soft);
        color: var(--c-danger);
    }

    /* ---- Activity ---- */
    .activity-loading {
        display: flex;
        justify-content: center;
        padding: 2rem;
    }
    .activity-list {
        display: flex;
        flex-direction: column;
    }
    .activity-row {
        display: flex;
        align-items: center;
        gap: 14px;
        padding: 14px 0;
        border-top: 1px solid var(--c-line-soft);
    }
    .activity-row:first-child {
        border-top: none;
    }
    .activity-icon {
        width: 38px;
        height: 38px;
        border-radius: var(--radius);
        background: var(--c-primary-soft);
        color: var(--c-primary-deep);
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    .activity-info {
        flex: 1;
        min-width: 0;
    }
    .activity-action {
        font-size: 14px;
        font-weight: 700;
        color: var(--c-ink);
    }
    .activity-count {
        font-size: 12.5px;
        color: var(--c-muted);
        margin-top: 2px;
    }
    .activity-date {
        font-size: 12px;
        color: var(--c-faint);
        white-space: nowrap;
    }

    /* ---- Danger card ---- */
    :global(.danger-card) {
        border-color: var(--c-danger-soft) !important;
    }

    /* ---- Modal ---- */
    .modal-overlay {
        position: fixed;
        inset: 0;
        background: rgba(15, 33, 28, 0.5);
        backdrop-filter: blur(4px);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 50;
        padding: 1rem;
    }
    .modal-content {
        background: var(--c-surface);
        border-radius: var(--radius-lg);
        width: 100%;
        max-width: 24rem;
        overflow: hidden;
        box-shadow: var(--shadow-lg);
    }
    .modal-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 16px 20px;
        border-bottom: 1px solid var(--c-line);
    }
    .modal-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 16px;
        font-weight: 800;
    }
    .modal-close {
        background: none;
        border: none;
        color: var(--c-faint);
        font-size: 22px;
        line-height: 1;
        cursor: pointer;
        padding: 4px;
    }
    .modal-body {
        padding: 20px;
        display: flex;
        flex-direction: column;
        gap: 14px;
    }
    .delete-warning {
        font-size: 13.5px;
        color: var(--c-muted);
        line-height: 1.5;
    }
    .delete-actions {
        display: flex;
        gap: 10px;
        justify-content: flex-end;
    }
</style>
