<script>
  import { onMount } from "svelte";
  import { Pencil, LogOut, Wallet, UserPlus, FileSignature, MessageCircle, Mail } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import MScreen from "../ui/MScreen.svelte";
  import MAvatar from "../ui/MAvatar.svelte";
  import MSegmented from "../ui/MSegmented.svelte";
  import MKV from "../ui/MKV.svelte";
  import MGroup from "../ui/MGroup.svelte";

  let { nav } = $props();
  let me = $state(nav.user || {});
  let org = $state(null);
  let tab = $state("profil");
  const segs = [{ value: "profil", label: "Profil" }, { value: "perusahaan", label: "Perusahaan" }, { value: "notif", label: "Notifikasi" }];

  let prefs = $state({ bayar: true, jamaah: true, kontrak: true, wa: true, email: false });
  const prefRows = [
    { k: "bayar", label: "Pembayaran masuk", icon: Wallet },
    { k: "jamaah", label: "Jamaah baru", icon: UserPlus },
    { k: "kontrak", label: "Kontrak ditandatangani", icon: FileSignature },
    { k: "wa", label: "Notifikasi WhatsApp", icon: MessageCircle },
    { k: "email", label: "Ringkasan email", icon: Mail },
  ];

  onMount(async () => {
    const [u, o] = await Promise.all([ApiService.getMe().catch(() => null), ApiService.getOrganization().catch(() => null)]);
    if (u) me = { ...me, ...u };
    org = o;
  });

  const fmtDate = (s) => {
    if (!s) return "—";
    const d = new Date(s);
    return isNaN(d.getTime()) ? String(s) : d.toLocaleDateString("id-ID", { month: "short", year: "numeric" });
  };
  let profilKV = $derived([
    { k: "Email", v: me.email || "—", full: true },
    { k: "No. HP", v: me.phone || me.no_hp || "—" },
    { k: "Peran", v: me.role || "Staff" },
    { k: "Bergabung", v: fmtDate(me.created_at), full: true },
  ]);
  let orgKV = $derived([
    { k: "Nama Legal", v: org?.name || org?.legal_name || "—", full: true },
    { k: "Izin PPIU", v: org?.ppiu_number || "—" },
    { k: "NPWP", v: org?.npwp || "—" },
    { k: "SK Kemenag", v: org?.sk_number || "—", full: true },
  ]);
</script>

<MScreen title="Profil Saya" onBack={nav.back}>
  <div style="text-align:center;padding:20px 18px 16px">
    <div style="display:inline-block"><MAvatar name={me.name || "Admin"} size={84} /></div>
    <div style="font-size:19px;font-weight:800;margin-top:12px">{me.name || "Admin"}</div>
    <div style="font-size:13px;color:var(--c-muted);margin-top:2px">{me.role || "Staff"}{org?.name ? " · " + org.name : ""}</div>
  </div>

  <div style="padding:0 18px 16px"><MSegmented tabs={segs} value={tab} onChange={(v) => (tab = v)} /></div>

  {#if tab === "profil"}
    <div style="padding:0 18px">
      <div class="m-card m-card-pad"><MKV items={profilKV} /></div>
      <div style="margin-top:16px"><button type="button" class="m-btn m-btn-ghost" onclick={() => nav.toast("Mode edit profil")}><Pencil size={18} />Edit Profil</button></div>
    </div>
  {:else if tab === "perusahaan"}
    <div style="padding:0 18px"><div class="m-card m-card-pad"><MKV items={orgKV} /></div></div>
  {:else}
    <div style="padding:0 18px">
      <MGroup>
        {#each prefRows as row (row.k)}
          {@const Icon = row.icon}
          <div class="m-row">
            <div class="m-row-ic" style="background:var(--c-primary-soft);color:var(--c-primary-deep)"><Icon size={18} /></div>
            <div class="m-row-main"><div class="m-row-title">{row.label}</div></div>
            <button type="button" onclick={() => (prefs[row.k] = !prefs[row.k])} aria-label={row.label}
              style="width:46px;height:27px;border-radius:999px;padding:3px;background:{prefs[row.k] ? 'var(--c-primary)' : 'var(--c-bg-2)'};display:flex;justify-content:{prefs[row.k] ? 'flex-end' : 'flex-start'};transition:background .2s;flex-shrink:0">
              <span style="width:21px;height:21px;border-radius:999px;background:#fff;box-shadow:0 1px 3px rgba(0,0,0,.2)"></span>
            </button>
          </div>
        {/each}
      </MGroup>
    </div>
  {/if}

  <div style="padding:20px 18px 0">
    <button type="button" class="m-btn m-btn-danger" onclick={() => nav.onLogout?.()}><LogOut size={18} />Keluar</button>
  </div>
</MScreen>
