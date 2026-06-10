<script>
  import { onMount } from "svelte";
  import { Check, CreditCard, CircleCheck } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import { fmtRp, fmtRpShort } from "../format.js";
  import MField from "../ui/MField.svelte";

  let { nav, params } = $props();

  let invoices = $state([]);
  let selected = $state(null);
  let amount = $state("");
  let method = $state("transfer_bank");
  let saving = $state(false);
  const methods = [
    { value: "transfer_bank", label: "Transfer Bank" },
    { value: "qris", label: "QRIS" },
    { value: "tunai", label: "Tunai" },
    { value: "e_wallet", label: "E-Wallet" },
    { value: "kartu_kredit", label: "Kartu Kredit" },
  ];

  onMount(async () => {
    try {
      const res = await ApiService.listInvoices({ pageSize: 50 });
      const list = res?.invoices || res?.data || (Array.isArray(res) ? res : []) || [];
      invoices = list.filter((i) => remaining(i) > 0);
      // Prefill if opened from a jamaah
      if (params?.jamaah) {
        const nm = params.jamaah.nama || params.jamaah.name;
        selected = invoices.find((i) => (i.jamaah_name || "") === nm) || null;
      }
    } catch {
      invoices = [];
    }
  });

  const remaining = (i) => Number(i.amount_remaining ?? i.total_amount - i.amount_paid) || 0;

  async function save() {
    if (!selected) {
      nav.toast("Pilih invoice dulu");
      return;
    }
    const amt = Number(amount);
    if (!amt || amt <= 0) {
      nav.toast("Masukkan jumlah pembayaran");
      return;
    }
    saving = true;
    try {
      await ApiService.recordPayment(selected.id, { amount: amt, payment_method: method, paid_at: new Date().toISOString().slice(0, 10) });
      nav.toast("Pembayaran berhasil dicatat!", CircleCheck);
      nav.back();
    } catch (err) {
      nav.toast(err?.message || "Gagal mencatat pembayaran");
    } finally {
      saving = false;
    }
  }
</script>

<div class="m-screen m-slide">
  <div class="m-nav">
    <button type="button" class="m-nav-btn" onclick={nav.back}>Kembali</button>
    <div class="m-nav-title">Catat Pembayaran</div>
    <div style="width:38px"></div>
  </div>
  <div class="m-scroll">
    <div style="padding:18px 18px 0">
      <MField label="Pilih Invoice">
        <div style="display:flex;flex-direction:column;gap:8px">
          {#if invoices.length}
            {#each invoices as inv (inv.id)}
              <button type="button" class="m-card" onclick={() => { selected = inv; if (!amount) amount = String(remaining(inv)); }}
                style="display:flex;align-items:center;gap:11px;padding:13px 14px;text-align:left;border-color:{selected?.id === inv.id ? 'var(--c-primary)' : 'var(--c-line)'};background:{selected?.id === inv.id ? 'var(--c-primary-tint)' : 'var(--c-surface)'}">
                <div style="flex:1;min-width:0">
                  <div style="font-size:14.5px;font-weight:700;white-space:nowrap;overflow:hidden;text-overflow:ellipsis">{inv.jamaah_name || inv.invoice_number || inv.id}</div>
                  <div class="tnum" style="font-size:12.5px;color:var(--c-muted)">Sisa {fmtRp(remaining(inv))}</div>
                </div>
                {#if selected?.id === inv.id}<CircleCheck size={20} style="color:var(--c-primary)" />{/if}
              </button>
            {/each}
          {:else}
            <div class="m-empty" style="padding:24px 0">Tidak ada tagihan tertunggak</div>
          {/if}
        </div>
      </MField>

      <div style="text-align:center;padding:18px 0 22px">
        <div style="font-size:12px;font-weight:700;letter-spacing:.04em;text-transform:uppercase;color:var(--c-faint);margin-bottom:8px">Jumlah Pembayaran</div>
        <div style="display:flex;align-items:center;justify-content:center;gap:6px">
          <span style="font-size:24px;font-weight:800;color:var(--c-muted)">Rp</span>
          <input value={amount} oninput={(e) => (amount = e.currentTarget.value.replace(/[^0-9]/g, ""))} inputmode="numeric" placeholder="0"
            style="width:200px;border:none;outline:none;background:transparent;font-size:38px;font-weight:800;text-align:center;font-family:inherit;color:var(--c-ink);font-variant-numeric:tabular-nums" />
        </div>
        {#if amount}<div class="tnum" style="font-size:14px;color:var(--c-muted);margin-top:4px">Rp {Number(amount).toLocaleString("id-ID")}</div>{/if}
      </div>

      <div style="display:flex;gap:8px;margin-bottom:20px">
        {#each [5000000, 10000000, 25000000] as v}
          <button type="button" onclick={() => (amount = String(v))} class="m-hchip" style="flex:1">{fmtRpShort(v)}</button>
        {/each}
      </div>

      <MField label="Metode Pembayaran">
        <div style="display:flex;flex-direction:column;gap:8px">
          {#each methods as m}
            <button type="button" onclick={() => (method = m.value)} class="m-card"
              style="display:flex;align-items:center;gap:11px;padding:13px 14px;text-align:left;border-color:{method === m.value ? 'var(--c-primary)' : 'var(--c-line)'};background:{method === m.value ? 'var(--c-primary-tint)' : 'var(--c-surface)'}">
              <div style="width:34px;height:34px;border-radius:9px;background:{method === m.value ? 'var(--c-primary)' : 'var(--c-bg-2)'};color:{method === m.value ? '#fff' : 'var(--c-muted)'};display:flex;align-items:center;justify-content:center"><CreditCard size={17} /></div>
              <span style="flex:1;font-size:14.5px;font-weight:600">{m.label}</span>
              {#if method === m.value}<CircleCheck size={20} style="color:var(--c-primary)" />{/if}
            </button>
          {/each}
        </div>
      </MField>
    </div>
    <div style="padding:22px 18px 0">
      <button type="button" class="m-btn m-btn-primary" disabled={saving} onclick={save}><Check size={18} />{saving ? "Menyimpan…" : "Simpan Pembayaran"}</button>
    </div>
    <div style="height:28px"></div>
  </div>
</div>
