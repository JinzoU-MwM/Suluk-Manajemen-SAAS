<script>
  import { onMount } from "svelte";
  import { FileSignature } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import MScreen from "../ui/MScreen.svelte";
  import MGroup from "../ui/MGroup.svelte";
  import MBadge from "../ui/MBadge.svelte";

  let { nav } = $props();
  let all = $state([]);
  let loading = $state(true);

  const LABEL = { signed: "Ditandatangani", sent: "Terkirim", pending: "Menunggu TTD", draft: "Draft" };
  const label = (s) => LABEL[s] || s || "—";
  const fmtDate = (s) => {
    if (!s) return "";
    const d = new Date(s);
    return isNaN(d.getTime()) ? String(s) : d.toLocaleDateString("id-ID", { day: "numeric", month: "short", year: "numeric" });
  };

  onMount(async () => {
    try {
      const res = await ApiService.listContracts();
      all = res?.contracts || res?.data || (Array.isArray(res) ? res : []) || [];
    } catch {
      all = [];
    } finally {
      loading = false;
    }
  });

  let signed = $derived(all.filter((c) => label(c.status) === "Ditandatangani").length);
  let waiting = $derived(all.filter((c) => ["Terkirim", "Menunggu TTD"].includes(label(c.status))).length);
</script>

<MScreen title="Kontrak & Akad" onBack={nav.back}>
  <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px;padding:16px 18px 0">
    <div class="m-card m-card-pad"><div class="tnum" style="font-size:20px;font-weight:800">{signed}</div><div style="font-size:12px;color:var(--c-muted);margin-top:2px">Ditandatangani</div></div>
    <div class="m-card m-card-pad"><div class="tnum" style="font-size:20px;font-weight:800">{waiting}</div><div style="font-size:12px;color:var(--c-muted);margin-top:2px">Menunggu</div></div>
  </div>
  <div style="padding:16px 18px 0">
    {#if loading}
      <div class="m-loading" style="padding:50px 0">Memuat…</div>
    {:else if all.length}
      <MGroup>
        {#each all as c (c.id)}
          <div class="m-row" role="button" tabindex="0" onclick={() => nav.toast("Buka kontrak")} onkeydown={() => {}}>
            <div class="m-row-ic" style="background:var(--c-accent-soft);color:var(--c-accent)"><FileSignature size={18} /></div>
            <div class="m-row-main">
              <div class="m-row-title">{c.recipient_name || c.jamaah || "Penerima"}</div>
              <div class="m-row-sub">{(c.template_name || c.template || "Kontrak") + (c.signed_at || c.created_at ? " · " + fmtDate(c.signed_at || c.created_at) : "")}</div>
            </div>
            <MBadge status={label(c.status)} />
          </div>
        {/each}
      </MGroup>
    {:else}
      <div class="m-empty">Belum ada kontrak</div>
    {/if}
    <div style="height:24px"></div>
  </div>
</MScreen>
