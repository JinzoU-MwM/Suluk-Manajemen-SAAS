<script>
  import { onMount } from "svelte";
  import { MessageCircle, Phone, CircleCheck, Clock, Wallet, Pencil } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRp } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MAvatar from "../ui/MAvatar.svelte";
  import MBadge from "../ui/MBadge.svelte";
  import MSection from "../ui/MSection.svelte";
  import MGroup from "../ui/MGroup.svelte";
  import MKV from "../ui/MKV.svelte";
  import MProgress from "../ui/MProgress.svelte";

  let { nav, params } = $props();
  let j = $state(params?.jamaah || {});
  let docs = $state([]);

  onMount(async () => {
    const [detail, dl] = await Promise.all([
      ApiService.getJamaah(params.id).catch(() => null),
      ApiService.listDocuments(params.id).catch(() => []),
    ]);
    if (detail) j = { ...j, ...detail };
    docs = Array.isArray(dl) ? dl : dl?.data || [];
  });

  const nm = (x) => x.nama || x.name || "Jamaah";
  let total = $derived(Number(j.total ?? j.total_amount ?? 0));
  let paid = $derived(Number(j.dp ?? j.amount_paid ?? 0));
  const DONE = ["diterima", "selesai", "lengkap", "verified", "approved"];
  let kv = $derived(
    [
      { k: "Paket", v: j.paket || j.package_name || "—", full: true },
      { k: "Grup", v: j.grup || j.group_name || "—", full: true },
      { k: "Keberangkatan", v: j.berangkat || j.departure_date || "—" },
      { k: "No. Paspor", v: j.no_paspor || j.noPaspor || "—" },
      { k: "No. HP", v: j.no_hp || j.hp || "—" },
      { k: "Kota Asal", v: j.kota || "—" },
    ],
  );
</script>

<MScreen title={j.id || "Detail Jamaah"} onBack={nav.back}>
  <div style="text-align:center;padding:22px 18px 18px">
    <div style="display:inline-block"><MAvatar name={nm(j)} size={78} /></div>
    <div style="font-size:20px;font-weight:800;margin-top:12px">{nm(j)}</div>
    <div style="font-size:13.5px;color:var(--c-muted);margin-top:3px">
      {(j.gender === "P" ? "Perempuan" : j.gender === "L" ? "Laki-laki" : "") + (j.umur ? " · " + j.umur + " th" : "") + (j.kota ? " · " + j.kota : "")}
    </div>
    {#if j.status}<div style="margin-top:10px"><MBadge status={j.status} dot /></div>{/if}
  </div>

  <div style="display:flex;gap:10px;padding:0 18px 18px">
    <button type="button" class="m-btn m-btn-soft" onclick={() => nav.toast("Membuka WhatsApp…", MessageCircle)}><MessageCircle size={18} />WhatsApp</button>
    <button type="button" class="m-btn m-btn-ghost" onclick={() => nav.toast("Menelepon…", Phone)}><Phone size={18} />Telepon</button>
  </div>

  {#if total > 0}
    <div style="padding:0 18px">
      <div class="m-card m-card-pad" style="background:var(--c-primary-tint);border:none">
        <div style="display:flex;justify-content:space-between;align-items:baseline">
          <div style="font-size:12px;font-weight:700;color:var(--c-primary-deep);letter-spacing:.03em">PEMBAYARAN</div>
          <div class="tnum" style="font-size:12.5px;color:var(--c-muted)">{Math.round((paid / total) * 100)}% lunas</div>
        </div>
        <div class="tnum" style="font-size:24px;font-weight:800;margin-top:6px">{fmtRp(paid)}</div>
        <div class="tnum" style="font-size:13px;color:var(--c-muted);margin-bottom:12px">dari {fmtRp(total)}</div>
        <MProgress value={paid} max={total} color={paid >= total ? "var(--c-success)" : "var(--c-accent)"} />
        {#if paid < total}
          <button type="button" class="m-btn m-btn-primary" style="margin-top:14px" onclick={() => nav.go("bayar", { jamaah: j })}>
            <Wallet size={18} />Catat Pembayaran
          </button>
        {/if}
      </div>
    </div>
  {/if}

  <MSection label="Detail Jamaah" style="padding-top:20px">
    <div class="m-card m-card-pad"><MKV items={kv} /></div>
  </MSection>

  <MSection label="Dokumen" style="padding-top:20px">
    <MGroup>
      {#if docs.length}
        {#each docs as d, i (i)}
          {@const done = DONE.includes(String(d.status || "").toLowerCase())}
          <div class="m-row">
            <div class="m-row-ic" style="background:{done ? 'var(--c-success-soft)' : 'var(--c-bg-2)'};color:{done ? 'var(--c-success)' : 'var(--c-faint)'}">
              {#if done}<CircleCheck size={18} />{:else}<Clock size={18} />{/if}
            </div>
            <div class="m-row-main"><div class="m-row-title">{d.doc_type || d.type || "Dokumen"}</div></div>
            <span style="font-size:12.5px;font-weight:600;color:{done ? 'var(--c-success)' : 'var(--c-faint)'}">{done ? "Lengkap" : "Menunggu"}</span>
          </div>
        {/each}
      {:else}
        <div class="m-empty" style="padding:24px 20px">Belum ada dokumen</div>
      {/if}
    </MGroup>
  </MSection>

  <div style="padding:20px 18px 0">
    <button type="button" class="m-btn m-btn-ghost" onclick={() => nav.toast("Mode edit jamaah")}><Pencil size={18} />Edit Data Jamaah</button>
  </div>
</MScreen>
