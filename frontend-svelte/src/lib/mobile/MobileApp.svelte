<script>
  import "./mobile.css";
  import { LayoutDashboard, Users, ScanLine, Briefcase, Grid3x3 } from "lucide-svelte";
  import MToast from "./ui/MToast.svelte";
  import { toast } from "./toast.svelte.js";

  import Beranda from "./screens/Beranda.svelte";
  import Jamaah from "./screens/Jamaah.svelte";
  import JamaahDetail from "./screens/JamaahDetail.svelte";
  import Scanner from "./screens/Scanner.svelte";
  import CRM from "./screens/CRM.svelte";
  import LeadDetail from "./screens/LeadDetail.svelte";
  import Bayar from "./screens/Bayar.svelte";
  import Approval from "./screens/Approval.svelte";
  import Notifikasi from "./screens/Notifikasi.svelte";
  import Lainnya from "./screens/Lainnya.svelte";

  let { user = null, onExit = () => {}, onLogout = () => {} } = $props();

  const TAB_ROOT = { beranda: Beranda, jamaah: Jamaah, scan: Scanner, crm: CRM, lainnya: Lainnya };
  const SCREENS = {
    "jamaah-detail": JamaahDetail,
    "lead-detail": LeadDetail,
    bayar: Bayar,
    approval: Approval,
    notifikasi: Notifikasi,
  };

  const TABS = [
    { id: "beranda", label: "Beranda", icon: LayoutDashboard },
    { id: "jamaah", label: "Jamaah", icon: Users },
    { id: "scan", label: "Scan", scan: true, icon: ScanLine },
    { id: "crm", label: "CRM", icon: Briefcase },
    { id: "lainnya", label: "Lainnya", icon: Grid3x3 },
  ];

  let tab = $state("beranda");
  let stack = $state([]); // [{ screen, params }]

  const nav = {
    go: (screen, params = {}) => (stack = [...stack, { screen, params }]),
    back: () => (stack = stack.slice(0, -1)),
    switchTab: (t) => {
      stack = [];
      tab = t;
    },
    toast,
    get user() {
      return user;
    },
    onExit,
    onLogout,
  };

  let RootComp = $derived(TAB_ROOT[tab]);
</script>

<div class="m-app">
  <div style="flex:1;position:relative;overflow:hidden;display:flex;flex-direction:column">
    {#key tab}
      <div class="m-tabview" style="flex:1;display:flex;flex-direction:column;min-height:0">
        <RootComp {nav} />
      </div>
    {/key}

    {#each stack as s, i (i + s.screen)}
      {@const Comp = SCREENS[s.screen]}
      <div style="position:absolute;inset:0;z-index:{30 + i}">
        {#if Comp}<Comp {nav} params={s.params} />{/if}
      </div>
    {/each}
  </div>

  <div class="m-tabbar">
    {#each TABS as t}
      {@const Icon = t.icon}
      {#if t.scan}
        <button type="button" class="m-tab m-tab-scan" onclick={() => nav.switchTab("scan")}>
          <div class="ic"><Icon size={25} /></div>
        </button>
      {:else}
        {@const active = tab === t.id && stack.length === 0}
        <button type="button" class="m-tab {active ? 'on' : ''}" onclick={() => nav.switchTab(t.id)}>
          <Icon size={23} />
          <span>{t.label}</span>
          <span class="m-tab-dot"></span>
        </button>
      {/if}
    {/each}
  </div>

  <MToast />
</div>
