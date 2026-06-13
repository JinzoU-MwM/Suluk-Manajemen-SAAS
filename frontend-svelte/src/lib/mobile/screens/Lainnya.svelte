<script>
  import {
    Package, Users2, FileText, Wallet, BedDouble, FileSignature, Map, Truck,
    Handshake, Banknote, Boxes, Ban, User, Monitor, LogOut, BookOpen,
    PiggyBank, Calculator,
  } from "lucide-svelte";
  import MSection from "../ui/MSection.svelte";
  import MGroup from "../ui/MGroup.svelte";
  import MAvatar from "../ui/MAvatar.svelte";

  let { nav } = $props();

  const modules = [
    { ic: Package, label: "Paket", c: "#1B7F5A", id: "paket" },
    { ic: Users2, label: "Grup", c: "#2563a8", id: "grup" },
    { ic: FileText, label: "Invoice", c: "#C99A2E", id: "invoice" },
    { ic: Wallet, label: "Keuangan", c: "#15564a", id: "keuangan" },
    { ic: Calculator, label: "Kasir", c: "#1B7F5A", id: "kasir" },
    { ic: PiggyBank, label: "Tabungan", c: "#C99A2E", id: "tabungan" },
    { ic: BookOpen, label: "Akuntansi", c: "#0d7334", id: "akuntansi" },
    { ic: BedDouble, label: "Rooming", c: "#7a5ae0", id: "rooming" },
    { ic: FileSignature, label: "Kontrak", c: "#b8860b", id: "kontrak" },
    { ic: Map, label: "Itinerary", c: "#2563a8", id: "itinerary" },
    { ic: Truck, label: "Vendor", c: "#0f7a5a", id: "vendor" },
    { ic: Handshake, label: "Agen", c: "#a9842f", id: "agen" },
    { ic: Banknote, label: "Payroll", c: "#1B7F5A", id: "payroll" },
    { ic: Boxes, label: "Inventaris", c: "#b8860b", id: "inventaris" },
    { ic: Ban, label: "Pembatalan", c: "#c0392b", id: "pembatalan" },
  ];
  const name = nav.user?.name || "Admin";
</script>

<div class="m-screen-root">
  <div class="m-head"><div class="m-head-title">Lainnya</div><div class="m-head-sub">Semua modul & akun</div></div>

  <div class="m-scroll">
    <!-- profile card -->
    <div style="padding:4px 18px 0">
      <div class="m-card m-card-pad" style="display:flex;align-items:center;gap:13px">
        <MAvatar {name} size={48} />
        <div style="flex:1;min-width:0">
          <div style="font-size:16px;font-weight:800">{name}</div>
          <div style="font-size:12.5px;color:var(--c-muted)">{nav.user?.email || "Staff · Suluk"}</div>
        </div>
      </div>
    </div>

    <MSection label="Modul" style="padding-top:20px">
      <div style="display:grid;grid-template-columns:repeat(4,1fr);gap:12px">
        {#each modules as m}
          {@const Icon = m.ic}
          <button type="button" class="m-qa" onclick={() => nav.go(m.id)}
            style="background:var(--c-surface);border:1px solid var(--c-line);border-radius:16px;padding:14px 4px;display:flex;flex-direction:column;align-items:center;gap:8px">
            <div style="width:42px;height:42px;border-radius:12px;background:{m.c}1c;color:{m.c};display:flex;align-items:center;justify-content:center"><Icon size={20} /></div>
            <span style="font-size:11px;font-weight:600;color:var(--c-ink);text-align:center">{m.label}</span>
          </button>
        {/each}
      </div>
    </MSection>

    <MSection label="Akun" style="padding-top:24px">
      <MGroup>
        <button type="button" class="m-row" onclick={() => nav.go("profil")}>
          <div class="m-row-ic" style="background:var(--c-primary-soft);color:var(--c-primary)"><User size={19} /></div>
          <div class="m-row-main"><div class="m-row-title">Profil Saya</div></div>
        </button>
        <button type="button" class="m-row" onclick={() => nav.onExit?.()}>
          <div class="m-row-ic" style="background:var(--c-info-soft);color:var(--c-info)"><Monitor size={19} /></div>
          <div class="m-row-main"><div class="m-row-title">Buka Versi Desktop</div></div>
        </button>
        <button type="button" class="m-row" onclick={() => nav.onLogout?.()}>
          <div class="m-row-ic" style="background:var(--c-danger-soft);color:var(--c-danger)"><LogOut size={19} /></div>
          <div class="m-row-main"><div class="m-row-title" style="color:var(--c-danger)">Keluar</div></div>
        </button>
      </MGroup>
    </MSection>
    <div style="height:24px"></div>
  </div>
</div>
