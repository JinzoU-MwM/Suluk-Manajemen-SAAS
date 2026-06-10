<script>
  import Seo from "$lib/components/Seo.svelte";
  import { breadcrumb } from "$lib/seo/schema.js";
  import { PANDUAN } from "$lib/content/panduan.js";

  const schema = breadcrumb([
    { name: "Beranda", path: "/" },
    { name: "Panduan", path: "/panduan" },
  ]);
</script>

<Seo
  title="Panduan Operasional Travel Umrah & Haji | Suluk"
  description="Tips dan langkah praktis mengelola jamaah, dokumen Siskopatuh, manifest, dan keuangan travel umrah & haji. Panduan lengkap dari tim Suluk."
  path="/panduan"
  {schema}
/>

<section class="gp-hero">
  <div class="gp-wrap">
    <p class="gp-badge">Panduan</p>
    <h1 class="gp-h1">Panduan Operasional <em>Travel Umrah</em></h1>
    <p class="gp-lead">
      Tips praktis dan langkah demi langkah untuk mengelola jamaah, dokumen, dan
      keuangan travel umrah &amp; haji Anda — agar operasional lebih rapi dan efisien.
    </p>
  </div>
</section>

<section class="gp-sec">
  <div class="gp-wrap">
    <div class="pd-list">
      {#each PANDUAN as p}
        <a class="pd-card" href={`/panduan/${p.slug}`}>
          <h2>{p.title}</h2>
          <p>{p.description}</p>
          <span class="pd-meta">{p.readMinutes} menit baca · {p.date}</span>
        </a>
      {/each}
    </div>
  </div>
</section>

<style>
  .pd-list { display: grid; gap: 18px; }
  .pd-card {
    display: block;
    padding: 22px 24px;
    border: 1px solid var(--c-line, #e6e9e7);
    border-radius: 16px;
    text-decoration: none;
    background: var(--c-surface, #fff);
    transition: box-shadow 0.15s, transform 0.1s;
  }
  .pd-card:hover {
    box-shadow: var(--shadow-lg, 0 18px 40px rgba(15, 61, 46, 0.12));
    transform: translateY(-2px);
  }
  .pd-card h2 { font-size: 20px; font-weight: 800; color: var(--c-ink, #14201b); margin: 0 0 6px; }
  .pd-card p { font-size: 15px; color: var(--c-ink-soft, #4a5a52); margin: 0 0 10px; line-height: 1.6; }
  .pd-meta { font-size: 13px; color: var(--c-accent, #c79a3e); font-weight: 700; }
</style>
