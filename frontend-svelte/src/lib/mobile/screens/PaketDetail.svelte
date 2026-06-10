<script>
  import { onMount } from "svelte";
  import { Globe } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRp, fmtRpShort } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MSection from "../ui/MSection.svelte";
  import MGroup from "../ui/MGroup.svelte";
  import MKV from "../ui/MKV.svelte";

  let { nav, params } = $props();
  let p = $state(params?.pkg || {});

  onMount(async () => {
    try {
      const detail = await ApiService.getPackage(params.id);
      if (detail) p = { ...p, ...detail };
    } catch {}
  });

  const ptype = (x) => x.package_type || x.tipe || "Umrah";
  let tiers = $derived(
    (p.pricing_tiers || []).length
      ? p.pricing_tiers.map((t) => [t.label || t.room_type || "Tier", Number(t.price)])
      : [],
  );
  let minPrice = $derived(tiers.length ? Math.min(...tiers.map((t) => t[1])) : Number(p.price ?? 0));
  let reserved = $derived(p.reserved_seats ?? p.terisi ?? 0);
  let totalSeats = $derived(p.total_seats ?? p.kuota ?? 0);
  let info = $derived(
    [
      { k: "Durasi", v: (p.duration_days || p.durasi || "—") + " hari" },
      { k: "Keberangkatan", v: p.departure_date || p.tgl || "—" },
      { k: "Maskapai", v: p.airline || p.maskapai || "—", full: true },
      { k: "Hotel Mekkah", v: p.hotel_makkah_name || p.hotelMekkah || "—", full: true },
      { k: "Hotel Madinah", v: p.hotel_madinah_name || p.hotelMadinah || "—", full: true },
    ],
  );
</script>

<MScreen title={p.id || "Detail Paket"} onBack={nav.back}>
  <div style="padding:16px 18px 0">
    <div class="m-card m-card-pad" style="background:linear-gradient(120deg,#1B7F5A,#0F3D2E);border:none;color:#fff">
      <div style="font-size:11.5px;opacity:.9;font-weight:700;letter-spacing:.04em;text-transform:uppercase">{ptype(p)}</div>
      <div style="font-size:19px;font-weight:800;margin-top:6px;line-height:1.2">{p.name || p.nama || "Paket"}</div>
      <div class="tnum" style="font-size:28px;font-weight:800;margin-top:14px">{fmtRp(minPrice)}</div>
      <div style="font-size:12.5px;opacity:.9;margin-top:2px">{reserved}/{totalSeats} kursi · {fmtRpShort(minPrice * reserved)} omzet</div>
    </div>
  </div>

  <MSection label="Informasi Paket" style="padding-top:20px">
    <div class="m-card m-card-pad"><MKV items={info} /></div>
  </MSection>

  {#if tiers.length}
    <MSection label="Tier Harga" style="padding-top:20px">
      <MGroup>
        {#each tiers as t (t[0])}
          <div class="m-row">
            <div class="m-row-main"><div class="m-row-title">{t[0]}</div></div>
            <span class="tnum" style="font-weight:800;font-size:14.5px">{fmtRp(t[1])}</span>
          </div>
        {/each}
      </MGroup>
    </MSection>
  {/if}

  <div style="padding:20px 18px 0">
    <button type="button" class="m-btn m-btn-primary" onclick={() => { if (p.slug) { navigator.clipboard?.writeText(location.origin + "/#/paket/" + p.slug); nav.toast("Link publik disalin"); } else nav.toast("Slug paket belum tersedia"); }}>
      <Globe size={18} />Bagikan Link Publik
    </button>
  </div>
</MScreen>
