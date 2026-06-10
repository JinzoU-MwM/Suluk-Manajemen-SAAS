<script>
  import { onMount } from "svelte";
  import { CircleCheck, X, Check, XCircle } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRp } from "../format.js";
  import MScreen from "../ui/MScreen.svelte";
  import MBadge from "../ui/MBadge.svelte";

  let { nav } = $props();
  let items = $state([]);
  let loading = $state(true);
  let busy = $state(null);

  onMount(load);
  async function load() {
    try {
      const res = await ApiService.listRefunds({ status: "pending", limit: 50 });
      const list = res?.data || res?.refunds || (Array.isArray(res) ? res : []) || [];
      items = list.map((r) => ({
        id: r.id,
        jamaah: r.jamaah_name || r.jamaah || "Jamaah",
        reason: r.reason || r.alasan || "—",
        amount: Number(r.amount ?? r.refund_amount ?? 0),
        detail: r.invoice_number || r.notes || r.package_name || "Refund",
      }));
    } catch {
      items = [];
    } finally {
      loading = false;
    }
  }

  async function act(id, ok) {
    busy = id;
    try {
      if (ok) await ApiService.approveRefund(id);
      else await ApiService.rejectRefund(id);
      items = items.filter((x) => x.id !== id);
      nav.toast(ok ? "Pengajuan disetujui" : "Pengajuan ditolak", ok ? Check : X);
    } catch (err) {
      nav.toast(err?.message || "Gagal memproses");
    } finally {
      busy = null;
    }
  }
</script>

<MScreen title="Persetujuan" onBack={nav.back}>
  <div style="padding:16px 18px 0">
    <div class="m-card m-card-pad" style="display:flex;align-items:center;gap:12px;margin-bottom:16px;background:var(--c-primary-tint);border:none">
      <div style="width:40px;height:40px;border-radius:11px;background:var(--c-primary);color:#fff;display:flex;align-items:center;justify-content:center"><CircleCheck size={21} /></div>
      <div>
        <div style="font-size:15px;font-weight:800">{items.length} menunggu persetujuan</div>
        <div style="font-size:12.5px;color:var(--c-muted)">Refund & diskon perlu otorisasi Anda</div>
      </div>
    </div>

    {#if loading}
      <div class="m-loading" style="padding:40px 0">Memuat…</div>
    {:else if items.length}
      {#each items as it (it.id)}
        <div class="m-card m-card-pad m-enter" style="margin-bottom:12px">
          <div style="display:flex;align-items:center;gap:11px;margin-bottom:12px">
            <div style="width:40px;height:40px;border-radius:11px;background:#c0392b1c;color:#c0392b;display:flex;align-items:center;justify-content:center;flex-shrink:0"><XCircle size={20} /></div>
            <div style="flex:1;min-width:0">
              <div style="display:flex;align-items:center;gap:7px"><span style="font-size:15px;font-weight:700">{it.jamaah}</span><MBadge status="Belum Bayar">Refund</MBadge></div>
              <div style="font-size:12.5px;color:var(--c-muted);margin-top:1px">{it.detail}</div>
            </div>
          </div>
          <div style="background:var(--c-bg);border-radius:12px;padding:11px 13px;margin-bottom:12px">
            <div style="display:flex;justify-content:space-between;font-size:13px"><span style="color:var(--c-muted)">Alasan</span><span style="font-weight:600">{it.reason}</span></div>
            <div style="display:flex;justify-content:space-between;font-size:13px;margin-top:7px"><span style="color:var(--c-muted)">Nilai</span><span class="tnum" style="font-weight:800;color:#c0392b">{fmtRp(it.amount)}</span></div>
          </div>
          <div style="display:flex;gap:10px">
            <button type="button" class="m-btn m-btn-danger" disabled={busy === it.id} onclick={() => act(it.id, false)}><X size={17} />Tolak</button>
            <button type="button" class="m-btn m-btn-primary" disabled={busy === it.id} onclick={() => act(it.id, true)}><Check size={17} />Setujui</button>
          </div>
        </div>
      {/each}
    {:else}
      <div class="m-empty" style="padding:50px 20px">
        <CircleCheck size={36} style="color:var(--c-success)" />
        <div style="margin-top:10px;font-size:14.5px;font-weight:600;color:var(--c-ink)">Semua beres!</div>
        <div style="font-size:13px;margin-top:3px">Tidak ada pengajuan tertunda.</div>
      </div>
    {/if}
  </div>
</MScreen>
