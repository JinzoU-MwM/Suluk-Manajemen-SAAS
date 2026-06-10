<script>
  import { CreditCard, Globe, ScanLine, Loader, RefreshCw, UserPlus, Sparkles } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import MAnimCheck from "../ui/MAnimCheck.svelte";
  import MChips from "../ui/MChips.svelte";
  import MProgress from "../ui/MProgress.svelte";

  let { nav } = $props();

  let doc = $state("KTP"); // KTP | Paspor
  let stage = $state("idle"); // idle | scan | done
  let prog = $state(0);
  let fileInput;
  let fields = $state([]);
  let saving = $state(false);

  const TEMPLATES = {
    KTP: [["NIK", ""], ["Nama", ""], ["Tgl Lahir", ""], ["Jenis Kelamin", ""], ["Kota", ""]],
    Paspor: [["No. Paspor", ""], ["Nama", ""], ["Kebangsaan", "Indonesia"], ["Berlaku", ""], ["Penerbit", ""]],
  };

  function pickFile() {
    if (stage === "scan") return;
    fileInput?.click();
  }

  // Flatten an OCR result into a lookup of lowercased-key -> value (best-effort).
  function flatten(obj, out = {}) {
    if (!obj || typeof obj !== "object") return out;
    for (const [k, v] of Object.entries(obj)) {
      if (v && typeof v === "object") flatten(v, out);
      else if (v != null && v !== "") out[k.toLowerCase().replace(/[^a-z]/g, "")] = String(v);
    }
    return out;
  }
  function prefill(lookup) {
    const aliases = {
      NIK: ["nik", "noidentitas", "noktp"],
      Nama: ["nama", "name", "namalengkap"],
      "Tgl Lahir": ["tgllahir", "tanggallahir", "birthdate", "dob"],
      "Jenis Kelamin": ["jeniskelamin", "gender"],
      Kota: ["kota", "kotalahir", "tempatlahir", "city"],
      "No. Paspor": ["nopaspor", "passportno", "nopassport"],
      Kebangsaan: ["kebangsaan", "nationality"],
      Berlaku: ["berlaku", "expiry", "validuntil", "tanggalpaspor"],
      Penerbit: ["penerbit", "issuer", "kantorimigrasi"],
    };
    return TEMPLATES[doc].map(([k, def]) => {
      const keys = aliases[k] || [];
      const hit = keys.map((a) => lookup[a]).find(Boolean);
      return [k, hit || def];
    });
  }

  async function onFile(e) {
    const file = e.target.files?.[0];
    if (!file) return;
    stage = "scan";
    prog = 0;
    const t0 = Date.now();
    const tick = setInterval(() => {
      prog = Math.min(95, ((Date.now() - t0) / 2200) * 100);
    }, 50);
    try {
      const res = await ApiService.uploadDocuments([file], null, { cacheMode: "default" });
      fields = prefill(flatten(res));
    } catch {
      fields = TEMPLATES[doc].slice();
      nav.toast("OCR gagal — isi data manual");
    } finally {
      clearInterval(tick);
      prog = 100;
      stage = "done";
    }
  }

  function reset() {
    stage = "idle";
    prog = 0;
    fields = [];
    if (fileInput) fileInput.value = "";
  }

  async function daftarkan() {
    const get = (k) => fields.find((f) => f[0] === k)?.[1] || "";
    const nama = get("Nama");
    if (!nama) {
      nav.toast("Nama wajib diisi");
      return;
    }
    saving = true;
    try {
      await ApiService.createProfile({ nama, no_identitas: get("NIK") || undefined, no_paspor: get("No. Paspor") || undefined });
      nav.toast("Jamaah berhasil didaftarkan!", UserPlus);
      reset();
      nav.switchTab("jamaah");
    } catch (err) {
      nav.toast(err?.message || "Gagal mendaftarkan");
    } finally {
      saving = false;
    }
  }
</script>

<div class="m-screen-root">
  <div class="m-head" style="padding-bottom:8px">
    <div class="m-head-title">AI Scanner</div>
    <div class="m-head-sub">Pindai dokumen, data terisi otomatis</div>
  </div>
  <div style="padding-bottom:12px">
    <MChips tabs={["KTP", "Kartu Keluarga", "Paspor"]} value={doc === "Paspor" ? "Paspor" : "KTP"} onChange={(v) => { doc = v === "Paspor" ? "Paspor" : "KTP"; reset(); }} />
  </div>

  <input bind:this={fileInput} onchange={onFile} type="file" accept="image/*" capture="environment" style="display:none" />

  <div class="m-scroll">
    <div style="padding:4px 18px 0">
      <!-- camera viewport -->
      <button type="button" onclick={pickFile} style="width:100%;position:relative;border-radius:20px;overflow:hidden;background:linear-gradient(150deg,#1a3d33,#0c2b22);aspect-ratio:1.4;display:flex;align-items:center;justify-content:center">
        <div style="text-align:center;color:rgba(255,255,255,.5)">
          {#if doc === "Paspor"}<Globe size={46} />{:else}<CreditCard size={46} />{/if}
          <div style="font-size:11.5px;margin-top:8px;font-weight:600;letter-spacing:.05em">{doc === "Paspor" ? "PASPOR RI" : "KARTU TANDA PENDUDUK"}</div>
        </div>
        {#each [["top:14px;left:14px", "3px 0 0 3px"], ["top:14px;right:14px", "3px 3px 0 0"], ["bottom:14px;left:14px", "0 0 3px 3px"], ["bottom:14px;right:14px", "0 3px 3px 0"]] as c}
          <span style="position:absolute;{c[0]};width:26px;height:26px;border-color:rgba(255,255,255,.7);border-style:solid;border-radius:3px;border-width:{c[1]}"></span>
        {/each}
        {#if stage === "scan"}
          <div class="m-scanline" style="position:absolute;left:14px;right:14px;height:3px;background:var(--c-accent);border-radius:2px;top:{prog}%"></div>
        {/if}
        {#if stage === "done"}
          <div style="position:absolute;inset:0;background:rgba(15,61,46,.55);display:flex;align-items:center;justify-content:center">
            <MAnimCheck size={64} />
          </div>
        {/if}
      </button>

      {#if stage === "idle"}
        <button type="button" class="m-btn m-btn-primary" style="margin-top:16px" onclick={pickFile}><ScanLine size={19} />Pindai {doc}</button>
      {:else if stage === "scan"}
        <div style="margin-top:16px">
          <div style="display:flex;justify-content:space-between;font-size:12.5px;font-weight:600;margin-bottom:8px">
            <span style="display:flex;align-items:center;gap:7px;color:var(--c-muted)"><span class="m-spin"><Loader size={15} /></span>Mengekstrak data…</span>
            <span class="tnum" style="color:var(--c-primary)">{Math.round(prog)}%</span>
          </div>
          <MProgress value={prog} />
        </div>
      {:else}
        <div class="m-enter" style="margin-top:18px">
          <div style="display:flex;align-items:center;gap:8px;margin-bottom:12px">
            <div class="m-label" style="padding:0">Hasil Ekstraksi</div>
            <span class="m-chip" style="background:var(--c-success-soft);color:var(--c-success)"><Sparkles size={12} />Periksa & lengkapi</span>
          </div>
          <div class="m-card m-card-pad" style="display:flex;flex-direction:column;gap:13px">
            {#each fields as f, i (f[0])}
              <div>
                <label for="sf-{i}" style="font-size:11px;font-weight:700;letter-spacing:.03em;text-transform:uppercase;color:var(--c-faint);display:block;margin-bottom:5px">{f[0]}</label>
                <input id="sf-{i}" class="m-input" bind:value={fields[i][1]} style="padding:10px 12px;font-size:14px" />
              </div>
            {/each}
          </div>
          <div style="display:flex;gap:10px;margin-top:14px">
            <button type="button" class="m-btn m-btn-ghost" onclick={reset}><RefreshCw size={17} />Ulang</button>
            <button type="button" class="m-btn m-btn-primary" disabled={saving} onclick={daftarkan}><UserPlus size={18} />{saving ? "Menyimpan…" : "Daftarkan"}</button>
          </div>
        </div>
      {/if}
    </div>
    <div style="height:24px"></div>
  </div>
</div>
