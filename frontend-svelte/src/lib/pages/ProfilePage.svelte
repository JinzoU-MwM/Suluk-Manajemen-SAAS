<script>
    import { onMount } from "svelte";
    import {
        UserCheck,
        Building,
        Shield,
        Lock,
        Eye,
        Mail,
        Phone,
        MapPin,
        Calendar,
        Check,
        CheckCircle,
        AlertCircle,
        Clock,
        Zap,
        Crown,
        Trash2,
        Activity,
        Bell,
        Moon,
        Sun,
        BarChart3,
        LogOut,
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
        document.title = "Profil - Suluk";
    });

    // Profile state
    let profile = $state(null);
    let subscription = $state(null);
    let trialStatus = $state(null);
    let loading = $state(true);
    let groupCount = $state(0);

    // Organization / company profile (Perusahaan tab)
    let org = $state(null);
    let editingOrg = $state(false);
    let savingOrg = $state(false);
    let orgMsg = $state({ type: "", text: "" });
    let orgForm = $state({
        name: "", ppiu_number: "", npwp: "", sk_number: "",
        phone: "", email: "", address: "",
    });

    function startEditOrg() {
        orgForm = {
            name: org?.name || "",
            ppiu_number: org?.ppiu_number || "",
            npwp: org?.npwp || "",
            sk_number: org?.sk_number || "",
            phone: org?.phone || "",
            email: org?.email || "",
            address: org?.address || "",
        };
        orgMsg = { type: "", text: "" };
        editingOrg = true;
    }

    async function saveOrg() {
        savingOrg = true;
        orgMsg = { type: "", text: "" };
        try {
            org = await ApiService.updateOrganization(orgForm);
            editingOrg = false;
            orgMsg = { type: "success", text: "Profil perusahaan diperbarui!" };
        } catch (e) {
            orgMsg = { type: "error", text: e.message };
        } finally {
            savingOrg = false;
        }
    }

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
        { id: "perusahaan", label: "Perusahaan", icon: Building },
        { id: "notifikasi", label: "Notifikasi", icon: Bell },
    ];

    // Edit profile (Profil tab)
    let editing = $state(false);
    let editForm = $state({ name: "", phone: "", city: "", bio: "" });
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
            editForm = { name: me.name || "", phone: me.phone || "", city: me.city || "", bio: me.bio || "" };
            ApiService.getOrganization().then((o) => { org = o; }).catch(() => {});
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
        if (!editForm.name.trim()) return;
        savingProfile = true;
        profileMsg = { type: "", text: "" };
        try {
            const updated = await ApiService.updateProfile({
                name: editForm.name.trim(),
                phone: editForm.phone.trim(),
                city: editForm.city.trim(),
                bio: editForm.bio.trim(),
            });
            profile = updated;
            const stored = JSON.parse(localStorage.getItem("user") || "{}");
            stored.name = updated.name;
            localStorage.setItem("user", JSON.stringify(stored));
            profileMsg = { type: "success", text: "Profil berhasil diperbarui!" };
            editing = false;
        } catch (e) {
            profileMsg = { type: "error", text: e.message };
        } finally {
            savingProfile = false;
        }
    }

    function cancelEdit() {
        editForm = { name: profile?.name || "", phone: profile?.phone || "", city: profile?.city || "", bio: profile?.bio || "" };
        profileMsg = { type: "", text: "" };
        editing = false;
    }

    async function confirmCancelSubscription() {
        savingSub = true;
        try {
            subscription = await ApiService.cancelSubscription();
            showCancelConfirm = false;
        } catch (e) {
            alert(e.message || "Gagal membatalkan langganan.");
        } finally {
            savingSub = false;
        }
    }

    async function resumeSubscription() {
        savingSub = true;
        try {
            subscription = await ApiService.resumeSubscription();
        } catch (e) {
            alert(e.message || "Gagal melanjutkan langganan.");
        } finally {
            savingSub = false;
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
        if (!iso) return "—";
        const d = new Date(iso);
        return d.toLocaleDateString("id-ID", {
            day: "numeric",
            month: "short",
            year: "numeric",
        });
    }

    function formatJoin(iso) {
        if (!iso) return "—";
        const d = new Date(iso);
        return d.toLocaleDateString("id-ID", {
            month: "long",
            year: "numeric",
        });
    }

    function formatDateTime(iso) {
        if (!iso) return "—";
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

    function roleLabel(role) {
        if (!role) return "Anggota";
        const map = {
            owner: "Owner / Direktur Utama",
            admin: "Administrator",
            staff: "Staf",
            member: "Anggota",
        };
        return map[role] || role;
    }

    let savingSub = $state(false);
    let showCancelConfirm = $state(false);
    let cancelAtPeriodEnd = $derived(!!subscription?.cancel_at_period_end);

    let planName = $derived(planMeta(subscription?.plan).name);
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
    let pro = $derived(isProOrHigher(subscription?.plan));
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
            subtitle="Kelola informasi pribadi, keamanan, dan preferensi akun Anda."
        >
            {#snippet actions()}
                <Button variant="ghost" icon={LogOut} onclick={onLogout}>Keluar</Button>
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
            <!-- LEFT: Summary + Nav -->
            <div class="profile-col-left">
                <Card pad={false} style="overflow:hidden">
                    <div class="summary-banner">
                        <Sparkles size={90} class="banner-spark" />
                    </div>
                    <div class="summary-body">
                        <div class="summary-avatar-wrap">
                            <div class="summary-avatar-ring">
                                <div
                                    class="summary-avatar"
                                    style="background:{avatarHex}1f;color:{avatarHex}"
                                >
                                    {profile.name?.charAt(0)?.toUpperCase() || "U"}
                                </div>
                            </div>
                        </div>
                        <div class="summary-name">{profile.name}</div>
                        <div class="summary-role">{roleLabel(profile.role)}</div>
                        <div class="summary-badges">
                            <Badge tone={pro ? "warning" : "info"} dot>
                                {planName} Plan
                            </Badge>
                            {#if profile.is_admin}
                                <Badge tone="success" dot>Admin</Badge>
                            {/if}
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
                                    {formatJoin(profile.created_at)}
                                </div>
                                <div class="summary-stat-label">Bergabung</div>
                            </div>
                        </div>

                        <!-- Subscription block (folded into summary) -->
                        <div class="plan-block">
                            <div class="plan-block-row">
                                <div class="plan-block-icon">
                                    {#if pro}<Crown size={16} />{:else}<Zap size={16} />{/if}
                                </div>
                                <div class="plan-block-info">
                                    <div class="plan-block-name">{planName} Plan</div>
                                    <div class="plan-block-desc">
                                        {planMeta(subscription?.plan).desc}
                                    </div>
                                </div>
                            </div>

                            {#if subscription?.status === "expired"}
                                <div class="plan-expiry plan-expiry-danger">Langganan kedaluwarsa</div>
                            {:else if cancelAtPeriodEnd && subscription?.expires_at}
                                <div class="plan-expiry plan-expiry-danger">
                                    Berakhir {formatDate(subscription.expires_at)} · tidak diperpanjang
                                </div>
                            {:else if subscription?.expires_at}
                                <div class="plan-expiry">Berlaku hingga {formatDate(subscription.expires_at)}</div>
                            {/if}

                            {#if pro && subscription?.status !== "expired"}
                                {#if cancelAtPeriodEnd}
                                    <Button
                                        variant="soft"
                                        size="sm"
                                        full
                                        disabled={savingSub}
                                        onclick={resumeSubscription}
                                    >
                                        {savingSub ? "Memproses…" : "Lanjutkan langganan"}
                                    </Button>
                                {:else if showCancelConfirm}
                                    <div class="plan-confirm">
                                        <div class="plan-confirm-text">
                                            Anda tetap bisa memakai {planName} sampai
                                            {formatDate(subscription?.expires_at)}, lalu paket turun ke Gratis.
                                        </div>
                                        <div class="plan-confirm-actions">
                                            <Button variant="ghost" size="sm" onclick={() => (showCancelConfirm = false)}>
                                                Batal
                                            </Button>
                                            <Button variant="danger" size="sm" disabled={savingSub} onclick={confirmCancelSubscription}>
                                                {savingSub ? "Memproses…" : "Ya, batalkan"}
                                            </Button>
                                        </div>
                                    </div>
                                {:else}
                                    <button type="button" class="plan-cancel-link" onclick={() => (showCancelConfirm = true)}>
                                        Batalkan perpanjangan
                                    </button>
                                {/if}
                            {/if}

                            {#if !pro && subscription?.usage_limit}
                                <div class="plan-usage">
                                    <div class="plan-usage-head">
                                        <span>Kuota scan</span>
                                        <span class="tabular">{usagePercent}%</span>
                                    </div>
                                    <ProgressBar value={usagePercent} max={100} color={usageColor} />
                                </div>
                            {/if}

                            {#if !pro}
                                <Button
                                    variant="accent"
                                    size="sm"
                                    icon={Crown}
                                    full
                                    onclick={onUpgradeRequest}
                                >
                                    Upgrade paket
                                </Button>
                                {#if trialStatus?.can_activate}
                                    <Button variant="soft" size="sm" full onclick={onUpgradeRequest}>
                                        Coba 7 Hari Gratis
                                    </Button>
                                {/if}
                            {/if}
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
                    <a href="/super-admin" class="superadmin-link">
                        <Shield size={16} />
                        <span>Super Admin Dashboard</span>
                    </a>
                {/if}
            </div>

            <!-- RIGHT: Active panel -->
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
                            {#if editing}
                                <div class="section-head-actions">
                                    <Button variant="ghost" size="sm" onclick={cancelEdit}>Batal</Button>
                                    <Button
                                        variant="primary"
                                        size="sm"
                                        icon={Check}
                                        onclick={saveProfile}
                                        disabled={savingProfile || !editForm.name.trim()}
                                    >
                                        {savingProfile ? "Menyimpan…" : "Simpan"}
                                    </Button>
                                </div>
                            {:else}
                                <Button
                                    variant="soft"
                                    size="sm"
                                    icon={UserCheck}
                                    onclick={() => (editing = true)}
                                >
                                    Edit
                                </Button>
                            {/if}
                        </div>

                        <div class="fields-grid">
                            <!-- Nama -->
                            <div class="field">
                                <label class="field-label" for="f-name">Nama Lengkap</label>
                                {#if editing}
                                    <input
                                        id="f-name"
                                        type="text"
                                        bind:value={editForm.name}
                                        class="field-input"
                                        placeholder="Nama lengkap"
                                    />
                                {:else}
                                    <div class="field-view">{profile.name || "—"}</div>
                                {/if}
                            </div>
                            <!-- Jabatan -->
                            <div class="field">
                                <span class="field-label">Jabatan</span>
                                <div class="field-view">{roleLabel(profile.role)}</div>
                            </div>
                            <!-- Email -->
                            <div class="field">
                                <span class="field-label">Email</span>
                                <div class="field-view">
                                    <Mail size={16} class="field-ic" />
                                    {profile.email || "—"}
                                </div>
                            </div>
                            <!-- Telepon -->
                            <div class="field">
                                <span class="field-label">No. Telepon</span>
                                {#if editing}
                                    <input class="field-input" type="tel" bind:value={editForm.phone} placeholder="cth. 0812-3456-7890" />
                                {:else}
                                    <div class="field-view"><Phone size={16} class="field-ic" />{profile.phone || "—"}</div>
                                {/if}
                            </div>
                            <!-- Kota -->
                            <div class="field">
                                <span class="field-label">Kota</span>
                                {#if editing}
                                    <input class="field-input" bind:value={editForm.city} placeholder="cth. Bandung" />
                                {:else}
                                    <div class="field-view"><MapPin size={16} class="field-ic" />{profile.city || "—"}</div>
                                {/if}
                            </div>
                            <!-- Bergabung -->
                            <div class="field">
                                <span class="field-label">Bergabung Sejak</span>
                                <div class="field-view">
                                    <Calendar size={16} class="field-ic" />
                                    {formatJoin(profile.created_at)}
                                </div>
                            </div>
                            <!-- Bio (full) -->
                            <div class="field field-full">
                                <span class="field-label">Bio</span>
                                {#if editing}
                                    <input class="field-input" bind:value={editForm.bio} placeholder="Ceritakan sedikit tentang Anda" />
                                {:else}
                                    <div class="field-bio">{profile.bio || "—"}</div>
                                {/if}
                            </div>
                        </div>

                        <!-- Avatar color picker -->
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

                    <Card>
                        <div class="section-head">
                            <div>
                                <div class="section-title">Aktivitas Terakhir</div>
                                <div class="section-desc">Riwayat scan dokumen terbaru.</div>
                            </div>
                            <div class="section-icon"><Activity size={19} /></div>
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
                                        <div class="activity-icon"><BarChart3 size={18} /></div>
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
                            <div class="section-icon"><Lock size={19} /></div>
                        </div>

                        <div class="fields-grid">
                            <div class="field field-full">
                                <label class="field-label" for="current-password">Sandi Saat Ini</label>
                                <div class="field-edit">
                                    <Eye size={16} class="field-edit-ic" />
                                    <input
                                        id="current-password"
                                        type="password"
                                        bind:value={currentPassword}
                                        class="field-input field-input-ic"
                                        placeholder="Masukkan sandi saat ini"
                                    />
                                </div>
                            </div>
                            <div class="field">
                                <label class="field-label" for="new-password">Sandi Baru</label>
                                <div class="field-edit">
                                    <Eye size={16} class="field-edit-ic" />
                                    <input
                                        id="new-password"
                                        type="password"
                                        bind:value={newPassword}
                                        class="field-input field-input-ic"
                                        placeholder="Minimal 6 karakter"
                                    />
                                </div>
                            </div>
                            <div class="field">
                                <label class="field-label" for="confirm-password">Konfirmasi Sandi Baru</label>
                                <div class="field-edit">
                                    <Eye size={16} class="field-edit-ic" />
                                    <input
                                        id="confirm-password"
                                        type="password"
                                        bind:value={confirmPassword}
                                        class="field-input field-input-ic"
                                        placeholder="Ulangi sandi baru"
                                    />
                                </div>
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
                                size="sm"
                                icon={Check}
                                onclick={savePassword}
                                disabled={savingPassword || !currentPassword || !newPassword || !confirmPassword}
                            >
                                {savingPassword ? "Mengubah…" : "Perbarui Sandi"}
                            </Button>
                        </div>
                    </Card>

                    <Card>
                        <div class="section-head">
                            <div>
                                <div class="section-title">Keamanan Akun</div>
                                <div class="section-desc">Lapisan perlindungan tambahan untuk akun Anda.</div>
                            </div>
                        </div>
                        <div class="setting-list">
                            <div class="setting-row">
                                <div class="setting-icon"><CheckCircle size={19} /></div>
                                <div class="setting-text">
                                    <div class="setting-title">Verifikasi Dua Langkah (2FA)</div>
                                    <div class="setting-desc">
                                        Lapisan keamanan tambahan saat masuk dari perangkat baru.
                                    </div>
                                </div>
                                <button
                                    type="button"
                                    class="toggle"
                                    role="switch"
                                    aria-checked="false"
                                    aria-label="Toggle 2FA"
                                    disabled
                                    title="Segera hadir"
                                >
                                    <span class="toggle-knob"></span>
                                </button>
                            </div>
                        </div>
                    </Card>

                    <Card>
                        <div class="section-head">
                            <div>
                                <div class="section-title">Sesi Aktif</div>
                                <div class="section-desc">Perangkat yang saat ini masuk ke akun Anda.</div>
                            </div>
                        </div>
                        <div class="setting-list">
                            <div class="setting-row">
                                <div class="setting-icon setting-icon-muted">
                                    <Activity size={19} />
                                </div>
                                <div class="setting-text">
                                    <div class="setting-title">Perangkat ini</div>
                                    <div class="setting-desc">Sesi yang sedang Anda gunakan.</div>
                                </div>
                                <Badge status="Aktif" dot>Sekarang</Badge>
                            </div>
                        </div>
                    </Card>

                    <Card class="danger-card">
                        <div class="setting-row setting-row-danger" style="border-top:none">
                            <div class="setting-icon section-icon-danger"><Trash2 size={19} /></div>
                            <div class="setting-text">
                                <div class="setting-title" style="color:var(--c-danger)">Hapus Akun</div>
                                <div class="setting-desc">
                                    Hapus akun secara permanen. Semua data grup dan riwayat akan dihapus.
                                    Tindakan ini tidak dapat dibatalkan.
                                </div>
                            </div>
                            <Button
                                variant="danger"
                                size="sm"
                                icon={Trash2}
                                onclick={() => (showDeleteModal = true)}
                            >
                                Hapus Akun
                            </Button>
                        </div>
                    </Card>
                {:else if tab === "perusahaan"}
                    <Card>
                        <div class="section-head">
                            <div>
                                <div class="section-title">Profil Perusahaan</div>
                                <div class="section-desc">
                                    Informasi legal travel yang tampil pada invoice & kontrak.
                                </div>
                            </div>
                            {#if editingOrg}
                                <div style="display:flex;gap:8px">
                                    <Button variant="ghost" size="sm" onclick={() => (editingOrg = false)}>Batal</Button>
                                    <Button variant="primary" size="sm" icon={Check} onclick={saveOrg} disabled={savingOrg}>
                                        {savingOrg ? "Menyimpan…" : "Simpan"}
                                    </Button>
                                </div>
                            {:else}
                                <Button variant="soft" size="sm" icon={UserCheck} onclick={startEditOrg}>Edit</Button>
                            {/if}
                        </div>

                        {#if orgMsg.text}
                            <div class="msg {orgMsg.type === 'success' ? 'msg-success' : 'msg-error'}">{orgMsg.text}</div>
                        {/if}

                        <div class="company-head">
                            <div class="company-tile"><Building size={28} /></div>
                            <div>
                                <div class="company-name">{org?.name || "—"}</div>
                                <div class="company-sub">
                                    <Badge status="Lunas">PPIU Berizin</Badge>
                                    Travel Umrah & Haji
                                </div>
                            </div>
                        </div>

                        <div class="fields-grid">
                            <div class="field">
                                <span class="field-label">Nama Legal</span>
                                {#if editingOrg}
                                    <input class="field-input" bind:value={orgForm.name} placeholder="Nama legal perusahaan" />
                                {:else}
                                    <div class="field-view">{org?.name || "—"}</div>
                                {/if}
                            </div>
                            <div class="field">
                                <span class="field-label">No. Izin PPIU</span>
                                {#if editingOrg}
                                    <input class="field-input" bind:value={orgForm.ppiu_number} placeholder="cth. U.123 Tahun 2024" />
                                {:else}
                                    <div class="field-view">{org?.ppiu_number || "—"}</div>
                                {/if}
                            </div>
                            <div class="field">
                                <span class="field-label">NPWP</span>
                                {#if editingOrg}
                                    <input class="field-input" bind:value={orgForm.npwp} placeholder="cth. 01.234.567.8-901.000" />
                                {:else}
                                    <div class="field-view">{org?.npwp || "—"}</div>
                                {/if}
                            </div>
                            <div class="field">
                                <span class="field-label">No. SK Kemenag</span>
                                {#if editingOrg}
                                    <input class="field-input" bind:value={orgForm.sk_number} placeholder="cth. KMA 1234/2024" />
                                {:else}
                                    <div class="field-view">{org?.sk_number || "—"}</div>
                                {/if}
                            </div>
                            <div class="field">
                                <span class="field-label">Telepon Kantor</span>
                                {#if editingOrg}
                                    <input class="field-input" bind:value={orgForm.phone} placeholder="cth. 022-2358-9000" />
                                {:else}
                                    <div class="field-view"><Phone size={16} class="field-ic" />{org?.phone || "—"}</div>
                                {/if}
                            </div>
                            <div class="field">
                                <span class="field-label">Email Resmi</span>
                                {#if editingOrg}
                                    <input class="field-input" type="email" bind:value={orgForm.email} placeholder="cth. cs@travel.id" />
                                {:else}
                                    <div class="field-view"><Mail size={16} class="field-ic" />{org?.email || "—"}</div>
                                {/if}
                            </div>
                            <div class="field field-full">
                                <span class="field-label">Alamat Kantor</span>
                                {#if editingOrg}
                                    <input class="field-input" bind:value={orgForm.address} placeholder="Alamat lengkap kantor" />
                                {:else}
                                    <div class="field-view"><MapPin size={16} class="field-ic" />{org?.address || "—"}</div>
                                {/if}
                            </div>
                        </div>
                    </Card>
                {:else if tab === "notifikasi"}
                    <Card>
                        <div class="section-head">
                            <div>
                                <div class="section-title">Notifikasi Aktivitas</div>
                                <div class="section-desc">
                                    Pilih kejadian yang ingin Anda terima pemberitahuannya.
                                </div>
                            </div>
                            <div class="section-icon"><Bell size={19} /></div>
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
                                size="sm"
                                icon={Check}
                                onclick={saveNotificationPrefs}
                                disabled={savingNotifs}
                            >
                                {savingNotifs ? "Menyimpan…" : "Simpan Preferensi"}
                            </Button>
                        </div>
                    </Card>

                    <Card>
                        <div class="section-head">
                            <div>
                                <div class="section-title">Tampilan</div>
                                <div class="section-desc">Atur tampilan aplikasi.</div>
                            </div>
                        </div>
                        <div class="setting-list">
                            <div class="setting-row">
                                <div class="setting-icon">
                                    {#if darkMode}<Sun size={19} />{:else}<Moon size={19} />{/if}
                                </div>
                                <div class="setting-text">
                                    <div class="setting-title">Mode Gelap</div>
                                    <div class="setting-desc">Tampilan dengan latar gelap.</div>
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
                        {deleting ? "Menghapus…" : "Hapus Permanen"}
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
        position: relative;
        width: 88px;
        height: 88px;
        margin: 0 auto 14px;
    }
    .summary-avatar-ring {
        width: 88px;
        height: 88px;
        border-radius: 50%;
        border: 4px solid var(--c-surface);
        background: var(--c-surface);
        box-sizing: border-box;
    }
    .summary-avatar {
        width: 100%;
        height: 100%;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 32px;
        font-weight: 800;
    }
    .summary-name {
        font-size: 18.5px;
        font-weight: 800;
        color: var(--c-ink);
    }
    .summary-role {
        font-size: 13px;
        color: var(--c-muted);
        margin-top: 3px;
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

    /* ---- Plan block (folded subscription) ---- */
    .plan-block {
        margin-top: 18px;
        padding-top: 18px;
        border-top: 1px solid var(--c-line-soft);
        display: flex;
        flex-direction: column;
        gap: 12px;
        text-align: left;
    }
    .plan-block-row {
        display: flex;
        align-items: center;
        gap: 11px;
    }
    .plan-block-icon {
        width: 40px;
        height: 40px;
        border-radius: var(--radius);
        background: var(--c-accent-soft);
        color: var(--c-accent);
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    .plan-block-info {
        min-width: 0;
    }
    .plan-block-name {
        font-size: 14px;
        font-weight: 800;
        color: var(--c-ink);
    }
    .plan-block-desc {
        font-size: 11.5px;
        color: var(--c-muted);
        margin-top: 2px;
        line-height: 1.4;
    }
    .plan-usage {
        display: flex;
        flex-direction: column;
        gap: 6px;
    }
    .plan-usage-head {
        display: flex;
        justify-content: space-between;
        font-size: 12px;
        color: var(--c-muted);
        font-weight: 600;
    }
    .plan-expiry { font-size: 12px; color: var(--c-muted); font-weight: 600; }
    .plan-expiry-danger { color: var(--c-danger); }
    .plan-cancel-link {
        background: none;
        border: none;
        padding: 0;
        font-size: 0.78rem;
        color: #94a3b8;
        text-decoration: underline;
        cursor: pointer;
        align-self: flex-start;
    }
    .plan-confirm {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        padding: 0.625rem;
        border-radius: 0.625rem;
        background: rgba(244, 63, 94, 0.06);
    }
    .plan-confirm-text {
        font-size: 0.78rem;
        color: #475569;
        line-height: 1.4;
    }
    .plan-confirm-actions {
        display: flex;
        gap: 0.5rem;
        justify-content: flex-end;
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
    .section-head-actions {
        display: flex;
        gap: 8px;
        flex-shrink: 0;
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
        gap: 20px 22px;
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
    .field-view {
        display: flex;
        align-items: center;
        gap: 9px;
        font-size: 14.5px;
        font-weight: 600;
        color: var(--c-ink);
        padding: 10px 0;
    }
    :global(.field-ic) {
        color: var(--c-primary);
        flex-shrink: 0;
    }
    .field-bio {
        font-size: 14px;
        color: var(--c-ink-soft);
        line-height: 1.6;
    }
    .field-edit {
        position: relative;
    }
    :global(.field-edit-ic) {
        position: absolute;
        left: 13px;
        top: 50%;
        transform: translateY(-50%);
        color: var(--c-faint);
        pointer-events: none;
        z-index: 1;
    }
    .field-input-ic {
        padding-left: 38px !important;
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

    /* ---- Company head ---- */
    .company-head {
        display: flex;
        align-items: center;
        gap: 16px;
        padding: 4px 0 20px;
        margin-bottom: 18px;
        border-bottom: 1px solid var(--c-line-soft);
    }
    .company-tile {
        width: 60px;
        height: 60px;
        border-radius: var(--radius-lg);
        background: var(--c-primary-soft);
        color: var(--c-primary-deep);
        display: flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    .company-name {
        font-size: 17px;
        font-weight: 800;
        color: var(--c-ink);
    }
    .company-sub {
        font-size: 13px;
        color: var(--c-muted);
        margin-top: 6px;
        display: flex;
        align-items: center;
        gap: 7px;
        flex-wrap: wrap;
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
    .setting-row-danger {
        padding: 0;
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
    .setting-icon-muted {
        background: var(--c-bg-2);
        color: var(--c-ink-soft);
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
    .toggle:disabled {
        opacity: 0.55;
        cursor: not-allowed;
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
