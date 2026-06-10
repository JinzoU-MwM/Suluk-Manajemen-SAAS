<script>
  import { onMount } from "svelte";
  import { Bell, Wallet, TriangleAlert, CircleAlert, CircleCheck } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import MScreen from "../ui/MScreen.svelte";
  import MGroup from "../ui/MGroup.svelte";

  let { nav } = $props();
  let items = $state([]);
  let loading = $state(true);

  const TONE = {
    success: ["#1B7F5A", Wallet],
    info: ["#2563a8", Bell],
    warning: ["#b8860b", TriangleAlert],
    danger: ["#c0392b", CircleAlert],
    error: ["#c0392b", CircleAlert],
  };

  function ago(s) {
    if (!s) return "";
    const d = new Date(s);
    if (isNaN(d.getTime())) return String(s);
    const diff = (Date.now() - d.getTime()) / 1000;
    if (diff < 60) return "baru";
    if (diff < 3600) return Math.floor(diff / 60) + " mnt";
    if (diff < 86400) return Math.floor(diff / 3600) + " jam";
    return Math.floor(diff / 86400) + " hr";
  }

  onMount(async () => {
    try {
      const data = await ApiService.getNotifications();
      const list = data?.items || data?.notifications || (Array.isArray(data) ? data : []) || [];
      items = list.map((n) => ({
        id: n.id,
        title: n.title || n.t || "Notifikasi",
        sub: n.message || n.body || n.d || "",
        time: ago(n.created_at || n.time),
        tone: TONE[String(n.severity || "info").toLowerCase()] || TONE.info,
      }));
    } catch {
      items = [];
    } finally {
      loading = false;
    }
  });

  async function readAll() {
    try {
      await ApiService.markAllNotificationsRead();
    } catch {}
    nav.toast("Semua ditandai dibaca");
  }
</script>

<MScreen title="Notifikasi" onBack={nav.back}>
  {#snippet right()}
    <button type="button" class="m-nav-btn" style="font-size:13px" onclick={readAll}>Baca</button>
  {/snippet}

  <div style="padding:14px 18px 0">
    {#if loading}
      <div class="m-loading" style="padding:40px 0">Memuat…</div>
    {:else if items.length}
      <MGroup>
        {#each items as n (n.id)}
          {@const Icon = n.tone[1]}
          <div class="m-row">
            <div class="m-row-ic" style="background:{n.tone[0]}1c;color:{n.tone[0]}"><Icon size={18} /></div>
            <div class="m-row-main">
              <div class="m-row-title">{n.title}</div>
              {#if n.sub}<div class="m-row-sub">{n.sub}</div>{/if}
            </div>
            <span style="font-size:11.5px;color:var(--c-faint);flex-shrink:0">{n.time}</span>
          </div>
        {/each}
      </MGroup>
    {:else}
      <div class="m-empty" style="padding:50px 20px">
        <CircleCheck size={34} style="color:var(--c-success)" />
        <div style="margin-top:10px;font-size:14px">Tidak ada notifikasi</div>
      </div>
    {/if}
  </div>
</MScreen>
