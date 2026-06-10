<script>
  import { MessageCircle, Phone, UserPlus } from "lucide-svelte";
  import { fmtRp } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MAvatar from "../ui/MAvatar.svelte";
  import MSection from "../ui/MSection.svelte";
  import MSegmented from "../ui/MSegmented.svelte";
  import MKV from "../ui/MKV.svelte";

  let { nav, params } = $props();
  const c = params?.lead || {};
  const STAGES = ["Prospek", "Follow Up", "Negosiasi", "Closing"];
  let stage = $state(STAGES.includes(c.stage) ? c.stage : "Prospek");
</script>

<MScreen title={c.id || "Lead"} onBack={nav.back}>
  <div style="text-align:center;padding:22px 18px 16px">
    <div style="display:inline-block"><MAvatar name={c.nama} size={74} /></div>
    <div style="font-size:19px;font-weight:800;margin-top:12px">{c.nama}</div>
    <div style="font-size:13.5px;color:var(--c-muted);margin-top:3px">{c.minat}</div>
    <div class="tnum" style="font-size:24px;font-weight:800;color:var(--c-primary);margin-top:10px">{fmtRp(c.nilai || 0)}</div>
  </div>

  <div style="display:flex;gap:10px;padding:0 18px 18px">
    <button type="button" class="m-btn m-btn-soft" onclick={() => nav.toast("Membuka WhatsApp…", MessageCircle)}><MessageCircle size={18} />Chat</button>
    <button type="button" class="m-btn m-btn-ghost" onclick={() => nav.toast("Menelepon…", Phone)}><Phone size={18} />Telepon</button>
  </div>

  <MSection label="Pindahkan Tahap">
    <div class="m-card m-card-pad">
      <MSegmented tabs={STAGES} value={stage} onChange={(v) => { stage = v; nav.toast("Dipindahkan ke " + v); }} />
    </div>
  </MSection>

  <MSection label="Detail Lead" style="padding-top:20px">
    <div class="m-card m-card-pad">
      <MKV items={[{ k: "Sumber", v: c.sumber || "—" }, { k: "No. HP", v: c.hp || "—" }, { k: "Minat Paket", v: c.minat || "—", full: true }]} />
    </div>
  </MSection>

  <div style="padding:20px 18px 0">
    <button type="button" class="m-btn m-btn-primary" onclick={() => { nav.toast("Lead dikonversi jadi jamaah!"); nav.back(); }}><UserPlus size={18} />Konversi ke Jamaah</button>
  </div>
</MScreen>
