<!-- Badge — Suluk design status pill (mirrors components.jsx Badge STATUS_MAP). -->
<script>
  // status text → [fg token, bg token]
  const MAP = {
    Lunas: ["--c-success", "--c-success-soft"],
    Dibayar: ["--c-success", "--c-success-soft"],
    Ditandatangani: ["--c-success", "--c-success-soft"],
    Published: ["--c-success", "--c-success-soft"],
    Aktif: ["--c-success", "--c-success-soft"],
    Aman: ["--c-success", "--c-success-soft"],
    Closing: ["--c-success", "--c-success-soft"],
    Selesai: ["--c-success", "--c-success-soft"],
    Cicilan: ["--c-info", "--c-info-soft"],
    Sebagian: ["--c-info", "--c-info-soft"],
    Terkirim: ["--c-info", "--c-info-soft"],
    Diproses: ["--c-info", "--c-info-soft"],
    DP: ["--c-warning", "--c-warning-soft"],
    Verifikasi: ["--c-warning", "--c-warning-soft"],
    "Menunggu TTD": ["--c-warning", "--c-warning-soft"],
    Pending: ["--c-warning", "--c-warning-soft"],
    Menipis: ["--c-warning", "--c-warning-soft"],
    "Belum Bayar": ["--c-danger", "--c-danger-soft"],
    "Jatuh Tempo": ["--c-danger", "--c-danger-soft"],
    Kritis: ["--c-danger", "--c-danger-soft"],
    Draft: ["--c-muted", "--c-bg-2"],
    Nonaktif: ["--c-muted", "--c-bg-2"],
  };
  let { status = "", label = "", tone = "", dot = false, children = null } = $props();
  // tone overrides the lookup: success|info|warning|danger|muted
  const TONES = {
    success: ["--c-success", "--c-success-soft"],
    info: ["--c-info", "--c-info-soft"],
    warning: ["--c-warning", "--c-warning-soft"],
    danger: ["--c-danger", "--c-danger-soft"],
    muted: ["--c-muted", "--c-bg-2"],
  };
  let pair = $derived(TONES[tone] || MAP[status] || ["--c-muted", "--c-bg-2"]);
  let text = $derived(label || status);
</script>

<span
  style="display:inline-flex;align-items:center;gap:6px;padding:4px 10px;border-radius:999px;background:var({pair[1]});color:var({pair[0]});font-size:12px;font-weight:700;white-space:nowrap;line-height:1.3"
>
  {#if dot}<span style="width:6px;height:6px;border-radius:999px;background:var({pair[0]})"></span>{/if}
  {#if children}{@render children()}{:else}{text}{/if}
</span>
