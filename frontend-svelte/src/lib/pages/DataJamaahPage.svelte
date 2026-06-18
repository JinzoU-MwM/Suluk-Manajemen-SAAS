<script>
  import { onMount } from "svelte";
  import {
    AlertCircle,
    BadgeCheck,
    BookUser,
    CheckCircle2,
    Clock,
    FileText,
    Loader2,
    Plane,
    Search,
    Sparkles,
    UserPlus,
    UsersRound,
    X,
  } from "lucide-svelte";
  import { ApiService } from "../services/api.js";
  import { mapError, showToast } from "../services/toast.svelte.js";
  import SlideDrawer from "../components/SlideDrawer.svelte";
  import EmptyState from "../components/EmptyState.svelte";
  import Pager from "../components/Pager.svelte";
  import PageHeader from "../components/PageHeader.svelte";
  import StatCard from "../components/StatCard.svelte";
  import Avatar from "../components/Avatar.svelte";
  import Card from "../components/ui/Card.svelte";
  import Badge from "../components/ui/Badge.svelte";
  import Button from "../components/ui/Button.svelte";
  import FilterTabs from "../components/ui/FilterTabs.svelte";
  import ProgressBar from "../components/ui/ProgressBar.svelte";

  let { onNavigate = () => {} } = $props();

  let groups = $state([]);
  let selectedGroupId = $state("");
  let members = $state([]);
  let isLoadingGroups = $state(true);
  let isLoadingMembers = $state(false);
  let error = $state("");
  let search = $state("");
  let tab = $state("Semua");
  let selected = $state(null);

  let selectedGroup = $derived(
    groups.find((group) => String(group.id) === String(selectedGroupId)) || null,
  );

  // --- Manual input (alternative to the AI Scanner) ---
  const emptyManual = {
    nama: "",
    gender: "",
    no_hp: "",
    no_identitas: "",
    no_paspor: "",
    tanggal_lahir: "",
    tempat_lahir: "",
    alamat: "",
    email: "",
  };
  let showManual = $state(false);
  let savingManual = $state(false);
  let manual = $state({ ...emptyManual });

  function openManual() {
    manual = { ...emptyManual };
    showManual = true;
  }

  async function saveManual() {
    if (!manual.nama.trim()) {
      showToast("Nama wajib diisi", "warning");
      return;
    }
    if (!selectedGroupId) {
      showToast("Pilih grup keberangkatan terlebih dahulu", "warning");
      return;
    }
    savingManual = true;
    try {
      // 1) Create the jamaah profile (stores the full data).
      const profile = await ApiService.createProfile({ ...manual, lead_source: "walk_in" });
      // 2) Link the new profile to the selected group so it shows in this list.
      if (profile?.id) {
        await ApiService.addGroupMembers(selectedGroupId, [
          { member_id: profile.id, name: manual.nama.trim(), phone: manual.no_hp.trim() },
        ]);
      }
      showToast("Jamaah berhasil ditambahkan", "success");
      showManual = false;
      await loadMembers();
    } catch (e) {
      showToast(mapError(e.message), "error");
    } finally {
      savingManual = false;
    }
  }

  // Documents a member can hold (drives checklist + completeness filter).
  function docList(member) {
    return [
      { label: "Identitas", ok: !!member.no_identitas },
      { label: "Paspor", ok: !!member.no_paspor },
      { label: "Visa", ok: !!member.no_visa },
      { label: "Asuransi", ok: !!(member.asuransi || member.no_polis) },
    ];
  }
  function docsDone(member) {
    return docList(member).filter((d) => d.ok).length;
  }
  function docCount() {
    return 4;
  }
  function isComplete(member) {
    return docsDone(member) === docCount();
  }

  let searchedMembers = $derived(
    members.filter((member) => {
      const needle = search.trim().toLowerCase();
      if (!needle) return true;
      return [
        member.nama,
        member.nama_paspor,
        member.no_paspor,
        member.no_identitas,
        member.no_visa,
      ]
        .filter(Boolean)
        .some((value) => String(value).toLowerCase().includes(needle));
    }),
  );

  let filteredMembers = $derived(
    searchedMembers.filter((member) => {
      if (tab === "Lengkap") return isComplete(member);
      if (tab === "Belum Lengkap") return !isComplete(member);
      return true;
    }),
  );

  // Filter tabs with live counts (Suluk design).
  let tabs = $derived([
    { value: "Semua", label: "Semua", count: searchedMembers.length },
    { value: "Lengkap", label: "Lengkap", count: searchedMembers.filter(isComplete).length },
    { value: "Belum Lengkap", label: "Belum Lengkap", count: searchedMembers.filter((m) => !isComplete(m)).length },
  ]);

  // Summary tiles (Suluk design). Counts are over the loaded group's members.
  let statTiles = $derived([
    { label: "Total di Grup", value: String(members.length), icon: UsersRound, accent: "var(--c-primary)" },
    { label: "Punya Paspor", value: String(members.filter((m) => m.no_paspor).length), icon: BookUser, accent: "var(--c-info)" },
    { label: "Punya Identitas", value: String(members.filter((m) => m.no_identitas).length), icon: BadgeCheck, accent: "var(--c-accent)" },
    { label: "Punya Visa", value: String(members.filter((m) => m.no_visa).length), icon: Plane, accent: "var(--c-success)" },
  ]);

  // Pagination (client-side over the filtered members)
  const PAGE_SIZE = 25;
  let page = $state(1);
  let pagedMembers = $derived(filteredMembers.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE));
  $effect(() => {
    search; selectedGroupId; tab;
    page = 1;
  });

  onMount(loadGroups);

  async function loadGroups() {
    isLoadingGroups = true;
    error = "";
    try {
      const result = await ApiService.listGroups();
      groups = result.groups || [];
      selectedGroupId = groups[0]?.id ? String(groups[0].id) : "";
      await loadMembers();
    } catch (e) {
      error = e.message;
    } finally {
      isLoadingGroups = false;
    }
  }

  async function loadMembers() {
    members = [];
    if (!selectedGroupId) return;
    isLoadingMembers = true;
    error = "";
    try {
      const group = await ApiService.getGroup(selectedGroupId);
      members = group.members || [];
    } catch (e) {
      error = e.message;
    } finally {
      isLoadingMembers = false;
    }
  }

  function displayName(member) {
    return member.nama || member.nama_paspor || "Tanpa nama";
  }

  function genderLabel(member) {
    const g = String(member.jenis_kelamin || "").toUpperCase();
    if (g === "L" || g === "LAKI-LAKI") return "Laki-laki";
    if (g === "P" || g === "PEREMPUAN") return "Perempuan";
    return "—";
  }

  // KeyValue items for the drawer (binds to real scanned fields).
  function keyValues(member) {
    return [
      { k: "Title", v: member.title || "—" },
      { k: "Jenis Kelamin", v: genderLabel(member) },
      { k: "Tempat Lahir", v: member.tempat_lahir || "—" },
      { k: "Tanggal Lahir", v: member.tanggal_lahir || "—" },
      { k: "No. Identitas", v: member.no_identitas || "—" },
      { k: "No. Paspor", v: member.no_paspor || "—" },
      { k: "No. Visa", v: member.no_visa || "—" },
      { k: "No. HP", v: member.no_hp || member.no_telepon || "—" },
      { k: "Alamat", v: member.alamat || "—", full: true },
    ];
  }
</script>

<div style="background:var(--c-bg)" class="min-h-screen p-4 lg:p-8">
  <PageHeader
    kicker="CRM & Jamaah"
    title="Data Jamaah"
    subtitle="Kelola seluruh data calon jamaah dari setiap grup keberangkatan."
  >
    {#snippet actions()}
      <div style="display:flex;gap:10px;flex-wrap:wrap">
        <Button variant="ghost" icon={Sparkles} onclick={() => onNavigate("scanner")}>
          Scan AI
        </Button>
        <Button variant="primary" icon={UserPlus} onclick={openManual}>
          Input Manual
        </Button>
      </div>
    {/snippet}
  </PageHeader>

  {#if error}
    <div class="mb-5 flex items-start gap-3 rounded-2xl border border-red-100 bg-red-50 p-4 text-sm text-red-700">
      <AlertCircle class="mt-0.5 h-5 w-5 flex-shrink-0" />
      <span>{mapError(error)}</span>
    </div>
  {/if}

  <!-- Summary cards (Suluk design) -->
  <div class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
    {#each statTiles as t}
      <StatCard icon={t.icon} label={t.label} value={t.value} accent={t.accent} />
    {/each}
  </div>

  <!-- Group selector -->
  <Card style="margin-bottom:var(--gap)">
    <label for="jamaah-group-select" style="display:block;margin-bottom:8px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint)">Grup Keberangkatan</label>
    <select
      id="jamaah-group-select"
      bind:value={selectedGroupId}
      onchange={loadMembers}
      style="width:100%;padding:12px 14px;font-size:13.5px;font-weight:500;color:var(--c-ink);background:var(--c-bg);border:1px solid var(--c-line);border-radius:var(--radius);outline:none"
    >
      {#if isLoadingGroups}
        <option>Memuat grup...</option>
      {:else if groups.length === 0}
        <option value="">Belum ada grup</option>
      {:else}
        {#each groups as group}
          <option value={String(group.id)}>{group.name} - {group.member_count || 0} jamaah</option>
        {/each}
      {/if}
    </select>
  </Card>

  <!-- Jamaah list -->
  <Card pad={false}>
    <div style="display:flex;justify-content:space-between;align-items:center;padding:16px 20px;gap:16px;flex-wrap:wrap">
      <div class="min-w-0">
        <div style="font-size:15px;font-weight:800;color:var(--c-ink)">{selectedGroup?.name || "Daftar Jamaah"}</div>
        <div style="font-size:12.5px;color:var(--c-muted);margin-top:2px">Data operasional yang tersimpan dari hasil scan dan input grup.</div>
      </div>
      <div style="display:flex;align-items:center;gap:12px;flex-wrap:wrap">
        <FilterTabs {tabs} value={tab} onChange={(v) => (tab = v)} />
        <div style="position:relative;width:260px;max-width:100%">
          <Search style="position:absolute;left:13px;top:50%;transform:translateY(-50%);color:var(--c-faint);pointer-events:none" size={17} />
          <input
            bind:value={search}
            placeholder="Cari nama, paspor, NIK…"
            style="width:100%;padding:10px 14px 10px 38px;font-size:13.5px;color:var(--c-ink);background:var(--c-surface);border:1px solid var(--c-line);border-radius:var(--radius);outline:none"
          />
        </div>
      </div>
    </div>

    {#if isLoadingMembers}
      <div style="color:var(--c-muted)" class="flex items-center justify-center gap-3 p-12 text-sm">
        <Loader2 class="h-5 w-5 animate-spin" style="color:var(--c-primary)" />
        Memuat data jamaah...
      </div>
    {:else if filteredMembers.length === 0}
      <EmptyState
        icon={UsersRound}
        title={search || tab !== "Semua" ? "Tidak ada jamaah yang cocok" : "Belum ada data jamaah"}
        text={search || tab !== "Semua" ? "Coba ubah filter atau kata kunci pencarian." : "Pilih grup lain atau tambah data dari AI Scanner."}
      />
    {:else}
      <div style="padding:0 4px 8px;overflow-x:auto">
        <table style="width:100%;border-collapse:collapse;font-size:13.5px">
          <thead>
            <tr>
              <th style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Jamaah</th>
              <th class="hidden md:table-cell" style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Paspor</th>
              <th class="hidden lg:table-cell" style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Identitas</th>
              <th class="hidden lg:table-cell" style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Visa</th>
              <th style="text-align:left;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Dokumen</th>
              <th style="text-align:center;padding:0 16px 12px;font-size:11.5px;font-weight:700;letter-spacing:.05em;text-transform:uppercase;color:var(--c-faint);white-space:nowrap;border-bottom:1px solid var(--c-line)">Status</th>
            </tr>
          </thead>
          <tbody>
            {#each pagedMembers as member}
              <tr
                class="suluk-row"
                style="cursor:pointer;transition:background .12s"
                onclick={() => (selected = member)}
              >
                <td style="padding:13px 16px;border-bottom:1px solid var(--c-line-soft);color:var(--c-ink-soft);vertical-align:middle">
                  <div style="display:flex;gap:12px;align-items:center">
                    <Avatar name={displayName(member)} size={38} />
                    <div class="min-w-0">
                      <div style="font-weight:700;color:var(--c-ink)" class="truncate">{displayName(member)}</div>
                      <div style="font-size:12px;color:var(--c-faint);margin-top:2px" class="truncate">
                        {member.title || ""}{member.title && member.tanggal_lahir ? " · " : ""}{member.tanggal_lahir || (member.title ? "" : "—")}
                      </div>
                    </div>
                  </div>
                </td>
                <td class="hidden md:table-cell" style="padding:13px 16px;border-bottom:1px solid var(--c-line-soft);color:var(--c-ink-soft);vertical-align:middle;white-space:nowrap">{member.no_paspor || "—"}</td>
                <td class="hidden lg:table-cell" style="padding:13px 16px;border-bottom:1px solid var(--c-line-soft);color:var(--c-ink-soft);vertical-align:middle;white-space:nowrap">{member.no_identitas || "—"}</td>
                <td class="hidden lg:table-cell" style="padding:13px 16px;border-bottom:1px solid var(--c-line-soft);color:var(--c-ink-soft);vertical-align:middle;white-space:nowrap">{member.no_visa || "—"}</td>
                <td style="padding:13px 16px;border-bottom:1px solid var(--c-line-soft);color:var(--c-ink-soft);vertical-align:middle;min-width:150px">
                  <div style="display:flex;justify-content:space-between;font-size:12px;margin-bottom:5px">
                    <span style="font-weight:700;color:var(--c-ink)" class="tabular">{docsDone(member)}/{docCount()}</span>
                    <span style="color:var(--c-faint)">dokumen</span>
                  </div>
                  <ProgressBar
                    value={docsDone(member)}
                    max={docCount()}
                    color={isComplete(member) ? "var(--c-success)" : "var(--c-accent)"}
                  />
                </td>
                <td style="padding:13px 16px;border-bottom:1px solid var(--c-line-soft);text-align:center;vertical-align:middle">
                  <Badge status={isComplete(member) ? "Lengkap" : "Sebagian"} tone={isComplete(member) ? "success" : "info"} dot />
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
        <div class="px-5">
          <Pager {page} pageSize={PAGE_SIZE} total={filteredMembers.length} onchange={(p) => (page = p)} />
        </div>
      </div>
    {/if}
  </Card>
</div>

<!-- Manual input modal -->
<SlideDrawer open={showManual} title="Input Jamaah Manual" width="600px" onClose={() => (showManual = false)}>
  <form onsubmit={(e) => { e.preventDefault(); saveManual(); }}>
    <div style="padding:22px 24px;display:flex;flex-direction:column;gap:15px">
      <div style="font-size:12.5px;color:var(--c-muted)">
        Ditambahkan ke grup: <strong style="color:var(--c-ink)">{selectedGroup?.name || "—"}</strong>
      </div>

      <div style="display:flex;flex-direction:column;gap:6px">
        <label for="m-nama" style="font-size:11.5px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:var(--c-faint)">Nama Lengkap <span style="color:#dc2626">*</span></label>
        <input id="m-nama" type="text" bind:value={manual.nama} placeholder="Nama sesuai paspor/identitas"
          style="width:100%;padding:11px 14px;font-size:13.5px;color:var(--c-ink);background:var(--c-bg);border:1px solid var(--c-line);border-radius:var(--radius);outline:none" />
      </div>

      <div style="display:grid;grid-template-columns:1fr 1fr;gap:15px">
        <div style="display:flex;flex-direction:column;gap:6px">
          <label for="m-gender" style="font-size:11.5px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:var(--c-faint)">Jenis Kelamin</label>
          <select id="m-gender" bind:value={manual.gender}
            style="width:100%;padding:11px 14px;font-size:13.5px;color:var(--c-ink);background:var(--c-bg);border:1px solid var(--c-line);border-radius:var(--radius);outline:none">
            <option value="">—</option>
            <option value="L">Laki-laki</option>
            <option value="P">Perempuan</option>
          </select>
        </div>
        <div style="display:flex;flex-direction:column;gap:6px">
          <label for="m-hp" style="font-size:11.5px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:var(--c-faint)">No. HP</label>
          <input id="m-hp" type="tel" bind:value={manual.no_hp} placeholder="08…"
            style="width:100%;padding:11px 14px;font-size:13.5px;color:var(--c-ink);background:var(--c-bg);border:1px solid var(--c-line);border-radius:var(--radius);outline:none" />
        </div>
      </div>

      <div style="display:grid;grid-template-columns:1fr 1fr;gap:15px">
        <div style="display:flex;flex-direction:column;gap:6px">
          <label for="m-nik" style="font-size:11.5px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:var(--c-faint)">No. Identitas (NIK)</label>
          <input id="m-nik" type="text" bind:value={manual.no_identitas}
            style="width:100%;padding:11px 14px;font-size:13.5px;color:var(--c-ink);background:var(--c-bg);border:1px solid var(--c-line);border-radius:var(--radius);outline:none" />
        </div>
        <div style="display:flex;flex-direction:column;gap:6px">
          <label for="m-paspor" style="font-size:11.5px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:var(--c-faint)">No. Paspor</label>
          <input id="m-paspor" type="text" bind:value={manual.no_paspor}
            style="width:100%;padding:11px 14px;font-size:13.5px;color:var(--c-ink);background:var(--c-bg);border:1px solid var(--c-line);border-radius:var(--radius);outline:none" />
        </div>
      </div>

      <div style="display:grid;grid-template-columns:1fr 1fr;gap:15px">
        <div style="display:flex;flex-direction:column;gap:6px">
          <label for="m-tlahir" style="font-size:11.5px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:var(--c-faint)">Tempat Lahir</label>
          <input id="m-tlahir" type="text" bind:value={manual.tempat_lahir}
            style="width:100%;padding:11px 14px;font-size:13.5px;color:var(--c-ink);background:var(--c-bg);border:1px solid var(--c-line);border-radius:var(--radius);outline:none" />
        </div>
        <div style="display:flex;flex-direction:column;gap:6px">
          <label for="m-dlahir" style="font-size:11.5px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:var(--c-faint)">Tanggal Lahir</label>
          <input id="m-dlahir" type="date" bind:value={manual.tanggal_lahir}
            style="width:100%;padding:11px 14px;font-size:13.5px;color:var(--c-ink);background:var(--c-bg);border:1px solid var(--c-line);border-radius:var(--radius);outline:none" />
        </div>
      </div>

      <div style="display:flex;flex-direction:column;gap:6px">
        <label for="m-email" style="font-size:11.5px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:var(--c-faint)">Email</label>
        <input id="m-email" type="email" bind:value={manual.email}
          style="width:100%;padding:11px 14px;font-size:13.5px;color:var(--c-ink);background:var(--c-bg);border:1px solid var(--c-line);border-radius:var(--radius);outline:none" />
      </div>

      <div style="display:flex;flex-direction:column;gap:6px">
        <label for="m-alamat" style="font-size:11.5px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:var(--c-faint)">Alamat</label>
        <textarea id="m-alamat" bind:value={manual.alamat} rows="2"
          style="width:100%;padding:11px 14px;font-size:13.5px;color:var(--c-ink);background:var(--c-bg);border:1px solid var(--c-line);border-radius:var(--radius);outline:none;resize:vertical"></textarea>
      </div>
    </div>

    <div style="padding:16px 24px;border-top:1px solid var(--c-line);display:flex;gap:10px;justify-content:flex-end;background:var(--c-bg)">
      <button type="button" onclick={() => (showManual = false)}
        style="padding:10px 18px;font-size:13.5px;font-weight:700;color:var(--c-muted);background:transparent;border:1px solid var(--c-line);border-radius:var(--radius);cursor:pointer">Batal</button>
      <button type="submit" disabled={savingManual}
        style="padding:10px 20px;font-size:13.5px;font-weight:700;color:#fff;background:var(--c-primary);border:none;border-radius:var(--radius);cursor:pointer;opacity:{savingManual ? 0.7 : 1}">
        {savingManual ? "Menyimpan…" : "Simpan Jamaah"}
      </button>
    </div>
  </form>
</SlideDrawer>

<!-- Detail Drawer (Suluk design) -->
{#if selected}
  <button
    type="button"
    onclick={() => (selected = null)}
    aria-label="Tutup"
    style="position:fixed;inset:0;z-index:90;background:rgba(16,33,28,0.34);backdrop-filter:blur(2px);border:none;cursor:default"
  ></button>
  <div
    role="dialog"
    aria-modal="true"
    aria-label={displayName(selected)}
    style="position:fixed;top:0;right:0;bottom:0;z-index:91;width:460px;max-width:94vw;background:var(--c-surface);box-shadow:var(--shadow-lg);display:flex;flex-direction:column"
  >
    <!-- Header -->
    <div style="padding:20px 24px;border-bottom:1px solid var(--c-line);display:flex;justify-content:space-between;align-items:flex-start;gap:12px">
      <div class="min-w-0">
        <div style="font-size:17px;font-weight:800;color:var(--c-ink)" class="truncate">{displayName(selected)}</div>
        <div style="font-size:13px;color:var(--c-muted);margin-top:3px" class="truncate">{selectedGroup?.name || "Jamaah"}</div>
      </div>
      <button
        type="button"
        onclick={() => (selected = null)}
        aria-label="Tutup"
        style="width:36px;height:36px;border-radius:var(--radius);display:flex;align-items:center;justify-content:center;color:var(--c-muted);background:transparent;flex-shrink:0"
      >
        <X size={18} />
      </button>
    </div>

    <!-- Body -->
    <div style="flex:1;overflow-y:auto;padding:24px">
      <div style="display:flex;flex-direction:column;gap:22px">
        <!-- Identity -->
        <div style="display:flex;gap:16px;align-items:center">
          <Avatar name={displayName(selected)} size={64} />
          <div class="min-w-0">
            <div style="font-size:18px;font-weight:800;color:var(--c-ink)" class="truncate">{displayName(selected)}</div>
            <div style="font-size:13px;color:var(--c-muted);margin-top:4px">
              {genderLabel(selected)}{selected.tanggal_lahir ? " · " + selected.tanggal_lahir : ""}
            </div>
            <div style="margin-top:8px">
              <Badge status={isComplete(selected) ? "Lengkap" : "Sebagian"} tone={isComplete(selected) ? "success" : "info"} dot />
            </div>
          </div>
        </div>

        <!-- Document completeness summary -->
        <div style="background:var(--c-primary-tint);border-radius:var(--radius);padding:18px">
          <div style="display:flex;justify-content:space-between;align-items:baseline;margin-bottom:10px">
            <div style="font-size:12.5px;font-weight:700;color:var(--c-primary-deep)">KELENGKAPAN DOKUMEN</div>
            <div style="font-size:13px;color:var(--c-muted)" class="tabular">{Math.round((docsDone(selected) / docCount()) * 100)}% lengkap</div>
          </div>
          <div style="font-size:22px;font-weight:800;margin-bottom:4px;color:var(--c-ink)" class="tabular">{docsDone(selected)} / {docCount()}</div>
          <div style="font-size:13px;color:var(--c-muted);margin-bottom:12px">dokumen tersimpan</div>
          <ProgressBar
            value={docsDone(selected)}
            max={docCount()}
            color={isComplete(selected) ? "var(--c-success)" : "var(--c-accent)"}
            height={9}
          />
        </div>

        <!-- KeyValue grid -->
        <div style="display:grid;grid-template-columns:1fr 1fr;gap:16px 20px">
          {#each keyValues(selected) as it}
            <div style={it.full ? "grid-column:1 / -1" : ""}>
              <div style="font-size:11.5px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:var(--c-faint);margin-bottom:5px">{it.k}</div>
              <div style="font-size:14.5px;font-weight:600;color:var(--c-ink)">{it.v}</div>
            </div>
          {/each}
        </div>

        <!-- Document checklist -->
        <div>
          <div style="font-size:12.5px;font-weight:700;color:var(--c-faint);text-transform:uppercase;letter-spacing:.04em;margin-bottom:10px">Dokumen</div>
          <div style="display:flex;flex-direction:column;gap:8px">
            {#each docList(selected) as d}
              <div style="display:flex;align-items:center;gap:10px;padding:10px 14px;background:var(--c-bg);border-radius:var(--radius-sm)">
                <div style="width:28px;height:28px;border-radius:7px;display:flex;align-items:center;justify-content:center;background:{d.ok ? 'var(--c-success-soft)' : 'var(--c-bg-2)'};color:{d.ok ? 'var(--c-success)' : 'var(--c-faint)'}">
                  {#if d.ok}<CheckCircle2 size={15} />{:else}<Clock size={15} />{/if}
                </div>
                <span style="flex:1;font-size:13.5px;font-weight:600;color:var(--c-ink)">{d.label}</span>
                <span style="font-size:12px;font-weight:600;color:{d.ok ? 'var(--c-success)' : 'var(--c-faint)'}">{d.ok ? "Lengkap" : "Menunggu"}</span>
              </div>
            {/each}
          </div>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div style="padding:16px 24px;border-top:1px solid var(--c-line);display:flex;gap:10px;background:var(--c-bg)">
      <Button variant="ghost" icon={FileText} full onclick={() => onNavigate("documents")}>Dokumen</Button>
      <Button variant="primary" icon={UserPlus} full onclick={() => onNavigate("scanner")}>Tambah Data</Button>
    </div>
  </div>
{/if}

<style>
  .tabular {
    font-variant-numeric: tabular-nums;
  }
  .suluk-row:hover {
    background: var(--c-primary-tint);
  }
</style>
