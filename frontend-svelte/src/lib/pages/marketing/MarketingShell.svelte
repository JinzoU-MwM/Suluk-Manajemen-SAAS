<!--
  MarketingShell — shared public chrome (sticky nav + footer) for the SEO/legal/
  content pages, ported from the design's guides.css (.gp-* classes).
  Children render the page body (gp-hero + gp-sec... or gp-legal).
-->
<script>
  import "../../styles/marketing.css";
  import LandingNav from "$lib/components/LandingNav.svelte";
  let { onNav = () => {}, onGoToApp = () => {}, onHome = () => {}, children } = $props();

  const FOOT_FITUR = [
    { slug: "fitur-invoice", label: "Invoice" },
    { slug: "fitur-crm", label: "CRM" },
    { slug: "fitur-keuangan", label: "Keuangan" },
    { slug: "fitur-kontrak", label: "E-Kontrak" },
    { slug: "fitur-payroll", label: "Payroll" },
  ];
  const FOOT_COMPANY = [
    { slug: "about", label: "Tentang" },
    { slug: "contact", label: "Hubungi Kami" },
    { slug: "privacy", label: "Kebijakan Privasi" },
    { slug: "terms", label: "Syarat & Ketentuan" },
  ];
</script>

<div class="gp">
  <LandingNav />

  {@render children?.()}

  <footer class="gp-footer">
    <div class="gp-wide">
      <div class="gp-footer-grid">
        <div class="gp-footer-about">
          <button class="gp-brand" style="background:none;border:none;cursor:pointer;padding:0" onclick={() => onHome()}>
            <span class="gp-brand-mark"><img src="/brand/suluk-mark-white.png" alt="Suluk" style="height:34px;width:auto;display:block" /></span>
            <span class="gp-brand-name" style="color:#fff">Suluk</span>
          </button>
          <p>Sistem manajemen all-in-one untuk bisnis travel umrah &amp; haji di Indonesia.</p>
        </div>
        <div>
          <h4>Produk</h4>
          <ul>
            <li><button type="button" class="foot-link" onclick={() => onHome()}>Beranda</button></li>
            <li><button type="button" class="foot-link" onclick={() => onNav("software")}>Software</button></li>
            <li><button type="button" class="foot-link" onclick={() => onNav("unduh")}>Aplikasi Mobile</button></li>
            <li><button type="button" class="foot-link" onclick={() => onGoToApp()}>Demo</button></li>
          </ul>
        </div>
        <div>
          <h4>Fitur</h4>
          <ul>
            {#each FOOT_FITUR as f}<li><button type="button" class="foot-link" onclick={() => onNav(f.slug)}>{f.label}</button></li>{/each}
          </ul>
        </div>
        <div>
          <h4>Perusahaan</h4>
          <ul>
            {#each FOOT_COMPANY as f}<li><button type="button" class="foot-link" onclick={() => onNav(f.slug)}>{f.label}</button></li>{/each}
          </ul>
        </div>
      </div>
      <div class="gp-footer-bottom">
        <div>© 2026 Jamnas — PT Jaminan Nasional Indonesia. Hak cipta dilindungi.</div>
        <div>Dibuat untuk travel umrah &amp; haji Indonesia</div>
      </div>
    </div>
  </footer>
</div>

<style>
  .foot-link {
    background: none;
    border: none;
    padding: 0;
    cursor: pointer;
    font-size: 14px;
    color: #a9c2b8;
    text-align: left;
  }
  .foot-link:hover { color: #fff; }
</style>
