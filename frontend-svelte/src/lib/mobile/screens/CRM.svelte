<script>
  import { onMount } from "svelte";
  import { Plus, ChevronRight } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRpShort } from "../format.js";
  import MChips from "../ui/MChips.svelte";
  import MAvatar from "../ui/MAvatar.svelte";
  import MFormSheet from "../ui/MFormSheet.svelte";

  let { nav } = $props();

  let items = $state([]);
  let loading = $state(true);
  let stage = $state("Semua");
  let formOpen = $state(false);

  const FIELDS = [
    { key: "nama", label: "Nama Lead", required: true },
    { key: "no_hp", label: "No. HP", type: "tel" },
    { key: "lead_source", label: "Sumber", type: "select", options: [{ value: "instagram", label: "Instagram Ads" }, { value: "referral", label: "Referral" }, { value: "whatsapp", label: "WhatsApp" }, { value: "website", label: "Website" }, { value: "event", label: "Event/Pameran" }, { value: "lainnya", label: "Lainnya" }] },
  ];

  async function submitLead(data) {
    const payload = Object.fromEntries(Object.entries(data).filter(([, v]) => v !== "" && v != null));
    await ApiService.createProfile(payload);
    nav.toast("Lead ditambahkan");
    await load();
  }

  const STAGE_COLOR = { Prospek: "#2563a8", "Follow Up": "#C99A2E", Negosiasi: "#7a5ae0", Closing: "#1B7F5A", Semua: "#1B7F5A" };
  const tempColor = (t) => (t === "hot" ? "var(--c-danger)" : t === "warm" ? "var(--c-warning)" : "var(--c-muted)");
  const tempBg = (t) => (t === "hot" ? "var(--c-danger-soft)" : t === "warm" ? "var(--c-warning-soft, #fef3c7)" : "var(--c-bg-2)");

  async function load() {
    try {
      const res = await ApiService.listCRM({ pageSize: 50 });
      items = (res?.data || []).map((x) => ({
        id: x.id,
        nama: x.nama || x.name || "Lead",
        minat: x.minat || x.package_name || x.paket || "—",
        nilai: Number(x.balance ?? x.outstanding ?? x.nilai ?? x.total_amount ?? 0),
        sumber: x.sumber || x.lead_source || "—",
        hp: x.no_hp || x.hp || "—",
        stage: x.pipeline_status || x.stage || "Semua",
        lead_score: x.lead_score ?? null,
        lead_temp: x.lead_temp || "cold",
        raw: x,
      }));
    } catch {
      items = [];
    } finally {
      loading = false;
    }
  }
  onMount(load);

  let stages = $derived(["Semua", ...Array.from(new Set(items.map((i) => i.stage))).filter((s) => s && s !== "Semua")]);
  let chips = $derived(stages.map((s) => ({ value: s, label: s === "Semua" ? `Semua (${items.length})` : `${s} (${items.filter((i) => i.stage === s).length})` })));
  let cards = $derived(stage === "Semua" ? items : items.filter((i) => i.stage === stage));
  let total = $derived(items.reduce((s, c) => s + c.nilai, 0));
</script>

<div class="m-screen-root">
  <div class="m-head" style="padding-bottom:10px">
    <div class="m-head-row">
      <div style="flex:1">
        <div class="m-head-title">CRM</div>
        <div class="m-head-sub">Potensi pipeline {fmtRpShort(total)}</div>
      </div>
      <button type="button" onclick={() => (formOpen = true)} aria-label="Lead baru" style="width:42px;height:42px;border-radius:13px;background:var(--c-primary);color:#fff;display:flex;align-items:center;justify-content:center">
        <Plus size={22} />
      </button>
    </div>
  </div>
  <div style="padding-bottom:10px"><MChips tabs={chips} value={stage} onChange={(v) => (stage = v)} /></div>

  <div class="m-scroll">
    <div style="padding:4px 18px 0;display:flex;flex-direction:column;gap:12px">
      {#if loading}
        <div class="m-loading" style="padding:50px 0">Memuat…</div>
      {:else if cards.length}
        {#each cards as c, idx (c.id)}
          {@const col = STAGE_COLOR[c.stage] || "#1B7F5A"}
          <div class="m-card m-card-pad m-enter" role="button" tabindex="0" style="border-left:3px solid {col};animation-delay:{idx * 0.05}s"
            onclick={() => nav.go("lead-detail", { lead: c })} onkeydown={(e) => e.key === "Enter" && nav.go("lead-detail", { lead: c })}>
            <div style="display:flex;align-items:center;gap:11px">
              <MAvatar name={c.nama} size={40} />
              <div style="flex:1;min-width:0">
                <div style="font-size:15px;font-weight:700">{c.nama}</div>
                <div style="font-size:12.5px;color:var(--c-muted);margin-top:1px">{c.minat}</div>
              </div>
              {#if c.lead_score != null}
                <span class="tnum" style="font-size:11px;font-weight:800;padding:2px 8px;border-radius:999px;background:{tempBg(c.lead_temp)};color:{tempColor(c.lead_temp)}">{c.lead_score}</span>
              {/if}
              <ChevronRight size={18} class="m-chev" />
            </div>
            <div style="display:flex;justify-content:space-between;align-items:center;margin-top:12px;padding-top:12px;border-top:1px solid var(--c-line-soft)">
              <span class="m-chip" style="background:var(--c-bg-2);color:var(--c-muted)">{c.sumber}</span>
              <span class="tnum" style="font-size:15px;font-weight:800;color:var(--c-primary)">{fmtRpShort(c.nilai)}</span>
            </div>
          </div>
        {/each}
      {:else}
        <div class="m-empty">Tidak ada lead</div>
      {/if}
    </div>
    <div style="height:24px"></div>
  </div>
</div>

<MFormSheet open={formOpen} title="Lead Baru" fields={FIELDS} submitLabel="Tambah Lead" onClose={() => (formOpen = false)} onSubmit={submitLead} />
