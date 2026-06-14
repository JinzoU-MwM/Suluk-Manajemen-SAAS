<script>
  import { onMount } from "svelte";
  import { LockOpen, Lock } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRp } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MSection from "../ui/MSection.svelte";
  import MFormSheet from "../ui/MFormSheet.svelte";

  let { nav } = $props();

  const STATUS_LABEL = { open: "Terbuka", closed: "Tertutup" };

  let loading = $state(true);
  let activeSession = $state(null);
  let history = $state([]);
  let showOpen = $state(false);
  let showClose = $state(false);

  async function load() {
    loading = true;
    try {
      const [act, hist] = await Promise.all([
        ApiService.getActiveCashSession().catch(() => null),
        ApiService.listCashSessions({ limit: 30 }).catch(() => []),
      ]);
      activeSession = act && act.id ? act : null;
      history = hist || [];
    } catch (e) {
      nav.toast?.(e?.message || "Gagal memuat sesi kas", "error");
    } finally {
      loading = false;
    }
  }

  async function submitOpen(v) {
    await ApiService.openCashSession({
      opening_float: parseInt(v.opening_float || "0", 10) || 0,
      notes: v.notes || "",
    });
    nav.toast?.("Kas dibuka", "success");
    await load();
  }

  async function submitClose(v) {
    await ApiService.closeCashSession(activeSession.id, {
      counted_cash: parseInt(v.counted_cash || "0", 10) || 0,
    });
    nav.toast?.("Kas ditutup", "success");
    await load();
  }

  function diffText(d) {
    if (d === 0 || d == null) return "Pas";
    return d > 0 ? "Lebih " + fmtRp(d) : "Kurang " + fmtRp(-d);
  }
  const diffColor = (d) => (d < 0 ? "var(--c-danger)" : "var(--c-primary-deep)");

  onMount(load);
</script>

<MScreen title="Kasir (POS)" onBack={nav.back}>
  {#if loading}
    <div style="padding:40px;text-align:center;color:var(--c-ink-soft)">Memuat…</div>
  {:else}
    <div style="padding:16px 18px 0">
      {#if activeSession}
        <div class="m-card m-card-pad" style="background:linear-gradient(150deg,var(--c-primary-deep),var(--c-primary));border:none;color:#fff">
          <div style="font-size:12.5px;opacity:.85;font-weight:600">Perkiraan Kas di Laci</div>
          <div class="tnum" style="font-size:28px;font-weight:800;margin-top:4px">{fmtRp(activeSession.expected_cash ?? activeSession.opening_float)}</div>
          <div style="font-size:12px;opacity:.9;margin-top:8px">Modal awal {fmtRp(activeSession.opening_float)} · sesi terbuka</div>
        </div>
        <button type="button" class="m-btn m-btn-primary" style="margin-top:14px" onclick={() => (showClose = true)}>
          <Lock size={18} /> Tutup Kas
        </button>
      {:else}
        <div class="m-card m-card-pad" style="text-align:center;color:var(--c-ink-soft)">
          Tidak ada sesi kas terbuka.
        </div>
        <button type="button" class="m-btn m-btn-primary" style="margin-top:14px" onclick={() => (showOpen = true)}>
          <LockOpen size={18} /> Buka Kas
        </button>
      {/if}
    </div>

    <MSection label="Riwayat Sesi" style="padding-top:20px">
      <div style="display:flex;flex-direction:column;gap:10px">
        {#each history as s}
          <div class="m-card m-card-pad">
            <div style="display:flex;justify-content:space-between;font-size:12px;color:var(--c-ink-soft)">
              <span>{STATUS_LABEL[s.status] || s.status}</span>
              <span class="tnum">{(s.opened_at || "").slice(0, 10)}</span>
            </div>
            <div style="display:flex;justify-content:space-between;margin-top:5px;font-size:13px">
              <span>Modal {fmtRp(s.opening_float)}</span>
              {#if s.status === "closed"}
                <span style="font-weight:700;color:{diffColor(s.difference)}">{diffText(s.difference)}</span>
              {/if}
            </div>
          </div>
        {:else}
          <div style="font-size:13px;color:var(--c-ink-soft)">Belum ada sesi kas.</div>
        {/each}
      </div>
    </MSection>
  {/if}
</MScreen>

<MFormSheet
  open={showOpen}
  title="Buka Kas"
  submitLabel="Buka"
  fields={[
    { key: "opening_float", label: "Modal Awal", type: "number", required: true, placeholder: "0" },
    { key: "notes", label: "Catatan (opsional)", type: "text", placeholder: "Misal: shift pagi…" },
  ]}
  onClose={() => (showOpen = false)}
  onSubmit={submitOpen}
/>

<MFormSheet
  open={showClose}
  title="Tutup Kas"
  submitLabel="Tutup Kas"
  fields={[
    { key: "counted_cash", label: "Kas Dihitung (fisik)", type: "number", required: true, placeholder: "0" },
  ]}
  onClose={() => (showClose = false)}
  onSubmit={submitClose}
/>
