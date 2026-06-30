<script>
  // Public pricing page (/harga). Monthly/annual toggle (annual = pay 10 months,
  // get 12 -> ~17% off / "2 bulan gratis"), plus a voucher field. Uses the shared
  // .gp-* marketing chrome; pricing-specific UI is scoped below. Brand: deep green
  // + gold (one accent each), one radius scale, zero em-dashes.
  import { goto } from "$app/navigation";
  import { PLANS, formatIDR, UNLIMITED } from "$lib/config/pricing.js";

  let { onGoToApp = () => {} } = $props();

  // "monthly" | "annual"
  let billing = $state("annual");

  // Frontend voucher catalog. (Server-side validation is applied at checkout; this
  // is the marketing-page preview of the discount.)
  const VOUCHERS = {
    SULUK20: { type: "percent", value: 20, label: "Diskon 20%" },
    HEMAT10: { type: "percent", value: 10, label: "Diskon 10%" },
    UMRAH50: { type: "amount", value: 50000, label: "Potongan Rp 50.000" },
  };
  let voucherInput = $state("");
  let appliedCode = $state("");
  let applied = $state(null);
  let voucherError = $state("");

  function applyVoucher(e) {
    e?.preventDefault();
    const code = voucherInput.trim().toUpperCase();
    if (!code) return;
    const v = VOUCHERS[code];
    if (!v) {
      voucherError = "Kode voucher tidak valid atau sudah berakhir.";
      return;
    }
    applied = v;
    appliedCode = code;
    voucherError = "";
    voucherInput = "";
  }
  function removeVoucher() {
    applied = null;
    appliedCode = "";
    voucherError = "";
  }

  function periodBase(p) {
    return billing === "annual" ? p.annualPrice : p.monthlyPrice;
  }
  function discounted(base) {
    if (!applied) return base;
    if (applied.type === "percent") return Math.round(base * (1 - applied.value / 100));
    if (applied.type === "amount") return Math.max(0, base - applied.value);
    return base;
  }

  const mainPlans = PLANS.filter((p) => p.key === "starter" || p.key === "pro" || p.key === "bisnis");
  const gratisPlan = PLANS.find((p) => p.key === "gratis");
  const enterprisePlan = PLANS.find((p) => p.key === "enterprise");

  const SALES_WA =
    "https://wa.me/6285159980404?text=" +
    encodeURIComponent("Halo, saya tertarik dengan paket Enterprise Suluk.");

  function handleCta(p) {
    if (p.key === "enterprise") {
      window.open(SALES_WA, "_blank");
      return;
    }
    const params = new URLSearchParams({ plan: p.key, billing });
    if (applied) params.set("voucher", appliedCode);
    goto(`/daftar?${params.toString()}`);
  }

  function userLimit(p) {
    return p.maxUsers === UNLIMITED ? "tak terbatas" : `${p.maxUsers} pengguna`;
  }

  const faqs = [
    {
      q: "Apa bedanya tagihan bulanan dan tahunan?",
      a: "Tahunan dibayar di muka untuk 12 bulan dengan harga 10 bulan, jadi Anda hemat 2 bulan (sekitar 17%). Bulanan lebih fleksibel dan bisa dibatalkan kapan saja.",
    },
    {
      q: "Bisa ganti paket nanti?",
      a: "Bisa. Naik atau turun paket kapan saja dari dalam aplikasi. Saat naik paket, selisihnya dihitung prorata.",
    },
    {
      q: "Bagaimana cara pakai voucher?",
      a: "Masukkan kode voucher di atas, lalu harga paket berbayar akan otomatis menyesuaikan. Kode juga terbawa ke halaman pendaftaran.",
    },
    {
      q: "Metode pembayaran apa saja yang didukung?",
      a: "Transfer bank, virtual account, dan e-wallet melalui gateway pembayaran. Invoice resmi otomatis terbit setelah pembayaran.",
    },
    {
      q: "Apakah ada masa coba gratis?",
      a: "Ya. Setiap akun baru mendapat trial Pro 14 hari, tanpa kartu kredit. Setelahnya turun ke Gratis bila tidak berlangganan.",
    },
  ];
</script>

<div class="hp">
  <!-- HERO + CONTROLS -->
  <header class="gp-hero">
    <div class="gp-wide hp-hero">
      <span class="gp-badge">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 12V8H6a2 2 0 0 1-2-2c0-1.1.9-2 2-2h12v4"/><path d="M4 6v12c0 1.1.9 2 2 2h14v-4"/><path d="M18 12a2 2 0 0 0 0 4h4v-4z"/></svg>
        Harga transparan, tanpa biaya tersembunyi
      </span>
      <h1 class="gp-h1">Harga yang tumbuh bersama travel Anda</h1>
      <p class="hp-lead">Mulai gratis, naik kelas saat bisnis berkembang. Bayar tahunan dan hemat 2 bulan. Batalkan kapan saja.</p>

      <div class="hp-toggle" role="group" aria-label="Periode tagihan">
        <button type="button" class:active={billing === "monthly"} onclick={() => (billing = "monthly")}>Bulanan</button>
        <button type="button" class:active={billing === "annual"} onclick={() => (billing = "annual")}>
          Tahunan <span class="hp-save">Hemat 2 bulan</span>
        </button>
      </div>

      <div class="hp-voucher">
        {#if applied}
          <div class="hp-chip">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg>
            Kode <strong>{appliedCode}</strong> diterapkan, {applied.label}
            <button type="button" class="hp-chip-x" onclick={removeVoucher} aria-label="Hapus voucher">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round"><path d="M18 6 6 18M6 6l12 12"/></svg>
            </button>
          </div>
        {:else}
          <form class="hp-voucher-form" onsubmit={applyVoucher}>
            <label class="hp-voucher-label" for="voucher">Punya kode voucher?</label>
            <div class="hp-voucher-row">
              <input id="voucher" type="text" bind:value={voucherInput} placeholder="mis. SULUK20" autocomplete="off" spellcheck="false" />
              <button type="submit" class="hp-voucher-btn">Terapkan</button>
            </div>
          </form>
          {#if voucherError}<div class="hp-voucher-err" role="alert">{voucherError}</div>{/if}
        {/if}
      </div>
    </div>
  </header>

  <!-- PLAN CARDS -->
  <section class="hp-plans-sec">
    <div class="gp-wide">
      <div class="hp-grid">
        {#each mainPlans as p}
          {@const base = periodBase(p)}
          {@const final = discounted(base)}
          <div class="hp-card" class:pop={p.popular}>
            {#if p.popular}<span class="hp-badge-pop">Paling Populer</span>{/if}
            <h2 class="hp-name">{p.name}</h2>
            <p class="hp-desc">{p.desc}</p>

            <div class="hp-price">
              {#if applied && final < base}<span class="hp-orig">{formatIDR(base)}</span>{/if}
              <div class="hp-price-main">
                <span class="hp-amt">{formatIDR(final)}</span>
                <span class="hp-per">/{billing === "annual" ? "tahun" : "bulan"}</span>
              </div>
              {#if billing === "annual"}
                <div class="hp-eff">setara {formatIDR(Math.round(final / 12))}/bulan, hemat 2 bulan</div>
              {:else}
                <div class="hp-eff">ditagih bulanan, batal kapan saja</div>
              {/if}
            </div>

            <button class="hp-cta {p.popular ? 'primary' : 'ghost'}" onclick={() => handleCta(p)}>{p.cta}</button>

            <ul class="hp-feats">
              <li class="hp-feats-head">{p.rank > 1 ? `Semua di ${PLANS[p.rank - 1].name}, plus:` : "Termasuk:"}</li>
              {#each p.features as f}
                <li><svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg> {f}</li>
              {/each}
            </ul>
          </div>
        {/each}
      </div>

      <!-- Gratis + Enterprise (secondary) -->
      <div class="hp-secondary">
        <div class="hp-sec-card">
          <div class="hp-sec-top">
            <h3>{gratisPlan.name}</h3>
            <span class="hp-sec-price">Rp 0</span>
          </div>
          <p>{gratisPlan.desc} Kelola hingga {gratisPlan.maxJamaah} jamaah, {userLimit(gratisPlan)}, tanpa kartu kredit.</p>
          <button class="hp-cta ghost sm" onclick={() => handleCta(gratisPlan)}>{gratisPlan.cta}</button>
        </div>
        <div class="hp-sec-card">
          <div class="hp-sec-top">
            <h3>{enterprisePlan.name}</h3>
            <span class="hp-sec-price">Custom</span>
          </div>
          <p>{enterprisePlan.desc} Multi-PT, akses API, pengguna tak terbatas, dan dukungan prioritas 24/7.</p>
          <button class="hp-cta ghost sm" onclick={() => handleCta(enterprisePlan)}>{enterprisePlan.cta}</button>
        </div>
      </div>
    </div>
  </section>

  <!-- FAQ -->
  <section class="gp-sec hp-faq-sec">
    <div class="gp-wrap">
      <h2 class="gp-h2" style="text-align:center">Pertanyaan seputar langganan</h2>
      <div class="gp-faq" style="margin-top:28px">
        {#each faqs as f}
          <details>
            <summary>{f.q}</summary>
            <div class="ans">{f.a}</div>
          </details>
        {/each}
      </div>
    </div>
  </section>

  <!-- CTA -->
  <section class="gp-sec" style="padding-top:0">
    <div class="gp-wide">
      <div class="gp-cta-box">
        <h2>Mulai gratis hari ini</h2>
        <p>Coba semua fitur Pro gratis 14 hari. Tanpa kartu kredit, tanpa komitmen.</p>
        <button class="gp-btn" style="background:#fff;color:var(--c-primary-deep)" onclick={() => goto("/daftar")}>Coba Gratis 14 Hari</button>
      </div>
    </div>
  </section>
</div>

<style>
  .hp { font-family: var(--font-ui); color: var(--c-ink); }
  .hp :global(*) { box-sizing: border-box; }

  /* hero */
  .hp-hero { text-align: center; max-width: 760px; margin: 0 auto; }
  .hp-hero .gp-h1 { margin-bottom: 14px; }
  .hp-lead { font-size: 18px; line-height: 1.6; color: var(--c-muted); margin: 0 auto 26px; max-width: 560px; }

  /* billing toggle */
  .hp-toggle { display: inline-flex; gap: 4px; padding: 4px; background: var(--c-bg); border: 1px solid var(--c-line); border-radius: 999px; }
  .hp-toggle button {
    display: inline-flex; align-items: center; gap: 8px; border: none; cursor: pointer;
    background: transparent; color: var(--c-ink-soft); font-family: inherit; font-size: 14.5px; font-weight: 700;
    padding: 9px 20px; border-radius: 999px; transition: background .18s, color .18s, box-shadow .18s;
  }
  .hp-toggle button.active { background: var(--c-surface); color: var(--c-primary-deep); box-shadow: 0 2px 8px -2px rgba(16,33,28,.18); }
  .hp-save { font-size: 11px; font-weight: 800; letter-spacing: .02em; color: #fff; background: var(--c-accent); padding: 2px 8px; border-radius: 999px; }
  .hp-toggle button.active .hp-save { background: var(--c-primary); }

  /* voucher */
  .hp-voucher { margin-top: 18px; min-height: 44px; }
  .hp-voucher-label { display: block; font-size: 13px; font-weight: 600; color: var(--c-faint); margin-bottom: 7px; }
  .hp-voucher-row { display: inline-flex; gap: 8px; }
  .hp-voucher-row input {
    width: 220px; max-width: 60vw; font-family: inherit; font-size: 14px; font-weight: 600; letter-spacing: .04em; text-transform: uppercase;
    padding: 10px 14px; border: 1px solid var(--c-line); border-radius: 10px; background: var(--c-surface); color: var(--c-ink); outline: none;
  }
  .hp-voucher-row input::placeholder { letter-spacing: normal; text-transform: none; color: var(--c-faint); font-weight: 500; }
  .hp-voucher-row input:focus { border-color: var(--c-primary); box-shadow: 0 0 0 3px var(--c-primary-soft); }
  .hp-voucher-btn { font-family: inherit; font-size: 14px; font-weight: 700; color: #fff; background: var(--c-primary); border: none; border-radius: 10px; padding: 10px 18px; cursor: pointer; transition: filter .15s, transform .1s; }
  .hp-voucher-btn:hover { filter: brightness(1.05); }
  .hp-voucher-btn:active { transform: translateY(1px); }
  .hp-voucher-err { margin-top: 8px; font-size: 13px; color: var(--c-danger); }
  .hp-chip {
    display: inline-flex; align-items: center; gap: 8px; font-size: 14px; font-weight: 600; color: var(--c-primary-deep);
    background: var(--c-primary-soft); border: 1px solid color-mix(in srgb, var(--c-primary) 22%, transparent); padding: 9px 10px 9px 14px; border-radius: 999px;
  }
  .hp-chip strong { font-weight: 800; letter-spacing: .03em; }
  .hp-chip-x { display: inline-flex; align-items: center; justify-content: center; width: 22px; height: 22px; border: none; border-radius: 50%; background: rgba(15,61,46,.1); color: var(--c-primary-deep); cursor: pointer; }
  .hp-chip-x:hover { background: rgba(15,61,46,.2); }

  /* plan cards */
  .hp-plans-sec { padding: 8px 0 8px; }
  .hp-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; max-width: 1000px; margin: 0 auto; align-items: start; }
  .hp-card {
    position: relative; display: flex; flex-direction: column; background: var(--c-surface);
    border: 1px solid var(--c-line); border-radius: 20px; padding: 28px 24px;
    box-shadow: 0 1px 2px rgba(16,33,28,.04); transition: transform .18s, box-shadow .18s, border-color .18s;
  }
  .hp-card:hover { transform: translateY(-4px); box-shadow: 0 22px 48px -24px rgba(16,33,28,.26); border-color: var(--c-primary-soft); }
  .hp-card.pop { border: 2px solid var(--c-primary); box-shadow: 0 28px 60px -28px var(--c-primary); }
  .hp-card.pop:hover { box-shadow: 0 34px 70px -28px var(--c-primary); border-color: var(--c-primary); }
  .hp-badge-pop { position: absolute; top: -13px; left: 50%; transform: translateX(-50%); background: var(--c-primary); color: #fff; font-size: 12px; font-weight: 700; padding: 5px 14px; border-radius: 999px; white-space: nowrap; }
  .hp-name { font-family: var(--font-display); font-size: 22px; font-weight: 800; margin: 0 0 6px; }
  .hp-desc { font-size: 13.5px; color: var(--c-muted); line-height: 1.5; margin: 0 0 18px; min-height: 40px; }
  .hp-price { margin-bottom: 20px; }
  .hp-orig { font-size: 15px; color: var(--c-faint); text-decoration: line-through; }
  .hp-price-main { display: flex; align-items: baseline; gap: 5px; flex-wrap: nowrap; }
  /* nowrap + a size that keeps the 7-digit annual prices on ONE line across all
     cards, so the price block height (and everything below it) stays aligned. */
  .hp-amt { font-size: 28px; font-weight: 800; letter-spacing: -.03em; line-height: 1.05; color: var(--c-ink); white-space: nowrap; }
  .hp-per { font-size: 14px; font-weight: 600; color: var(--c-muted); }
  .hp-eff { font-size: 12.5px; color: var(--c-primary-deep); font-weight: 600; margin-top: 8px; }
  .hp-feats { list-style: none; padding: 0; margin: 22px 0 0; display: flex; flex-direction: column; gap: 12px; }
  .hp-feats-head { font-size: 12px; font-weight: 700; letter-spacing: .04em; text-transform: uppercase; color: var(--c-faint); }
  .hp-feats li:not(.hp-feats-head) { display: flex; gap: 10px; align-items: flex-start; font-size: 14px; color: var(--c-ink-soft); line-height: 1.45; }
  .hp-feats li svg { color: var(--c-primary); flex-shrink: 0; margin-top: 1px; }

  /* CTAs (one radius scale: 12px) */
  .hp-cta { width: 100%; font-family: inherit; font-size: 15px; font-weight: 700; border-radius: 12px; padding: 13px 18px; cursor: pointer; transition: transform .12s, box-shadow .15s, filter .15s, background .15s; }
  .hp-cta.primary { background: var(--c-primary); color: #fff; border: none; box-shadow: 0 10px 26px -10px var(--c-primary); }
  .hp-cta.primary:hover { transform: translateY(-2px); filter: brightness(1.05); }
  .hp-cta.ghost { background: var(--c-surface); color: var(--c-ink); border: 1px solid var(--c-line); }
  .hp-cta.ghost:hover { border-color: var(--c-primary); color: var(--c-primary-deep); background: var(--c-primary-soft); }
  .hp-cta.sm { width: auto; padding: 11px 20px; font-size: 14px; }
  .hp-cta:focus-visible { outline: 2px solid var(--c-primary); outline-offset: 3px; }

  /* secondary band */
  .hp-secondary { display: grid; grid-template-columns: 1fr 1fr; gap: 18px; max-width: 1000px; margin: 28px auto 0; }
  .hp-sec-card { display: flex; flex-direction: column; align-items: flex-start; gap: 12px; border: 1px solid var(--c-line); border-radius: 18px; padding: 24px; background: var(--c-bg); }
  .hp-sec-top { display: flex; align-items: baseline; justify-content: space-between; width: 100%; }
  .hp-sec-top h3 { font-family: var(--font-display); font-size: 19px; font-weight: 800; margin: 0; }
  .hp-sec-price { font-size: 18px; font-weight: 800; color: var(--c-primary-deep); }
  .hp-sec-card p { font-size: 14px; color: var(--c-muted); line-height: 1.55; margin: 0; }

  .hp-faq-sec :global(.gp-faq) { max-width: 760px; margin-left: auto; margin-right: auto; }

  @media (max-width: 900px) {
    .hp-grid { grid-template-columns: 1fr; max-width: 460px; }
    .hp-card.pop { order: -1; }
    .hp-secondary { grid-template-columns: 1fr; max-width: 460px; }
  }
  @media (max-width: 560px) {
    .hp-toggle button { padding: 9px 16px; }
    .hp-lead { font-size: 16px; }
  }
</style>
