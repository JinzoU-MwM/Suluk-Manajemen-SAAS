<script>
  import { onMount } from "svelte";
  import { goto } from "$app/navigation";
  import LandingPage from "$lib/pages/LandingPage.svelte";
  import { gotoMarketing, homePath } from "$lib/stores/session.svelte.js";

  const BASE_URL = "https://suluk.site";
  const title = "Software Travel Umrah | Scan KTP/KK ke Siskopatuh - Suluk";
  const description =
    "Otomatisasi input data jamaah 10x lebih cepat. Scan KTP/KK langsung jadi Excel Siskopatuh 32 kolom, auto rooming hotel, manifest digital mutawwif. Coba gratis 7 hari.";

  // Structured data (homepage). Kept as raw <script> tags injected via {@html}.
  const jsonLd = `
<script type="application/ld+json">
{"@context":"https://schema.org","@type":"SoftwareApplication","name":"Suluk","description":"Software travel umrah untuk otomatisasi input data jamaah, scan KTP/KK ke Excel Siskopatuh, auto rooming hotel, dan manifest digital mutawwif.","url":"https://suluk.site","applicationCategory":"BusinessApplication","operatingSystem":"Web Browser","offers":{"@type":"Offer","price":"80000","priceCurrency":"IDR","priceValidUntil":"2026-12-31"},"publisher":{"@type":"Organization","name":"Suluk","url":"https://suluk.site"}}
<\/script>
<script type="application/ld+json">
{"@context":"https://schema.org","@type":"Organization","name":"Suluk","url":"https://suluk.site","logo":"https://suluk.site/icon-512.png"}
<\/script>
<script type="application/ld+json">
{"@context":"https://schema.org","@type":"WebSite","name":"Suluk","url":"https://suluk.site","inLanguage":"id-ID"}
<\/script>
<script type="application/ld+json">
{"@context":"https://schema.org","@type":"FAQPage","mainEntity":[{"@type":"Question","name":"Apa itu Suluk?","acceptedAnswer":{"@type":"Answer","text":"Suluk adalah software travel umrah untuk otomatisasi input data jamaah, auto rooming, dan manifest digital."}},{"@type":"Question","name":"Dokumen apa saja yang didukung?","acceptedAnswer":{"@type":"Answer","text":"Suluk mendukung pemrosesan dokumen KTP/KK, paspor, dan visa."}}]}
<\/script>`;

  onMount(() => {
    // Logged-in users skip the marketing homepage and go straight to the app.
    try {
      if (localStorage.getItem("user")) goto(homePath(), { replaceState: true });
    } catch {}
  });
</script>

<svelte:head>
  <title>{title}</title>
  <meta name="title" content={title} />
  <meta name="description" content={description} />
  <meta
    name="keywords"
    content="software travel umrah, aplikasi travel haji, sistem manajemen jamaah, input data jamaah siskopatuh, scan KTP/KK ke excel, auto rooming hotel umrah, manifest digital mutawwif"
  />
  <meta name="robots" content="index, follow" />
  <link rel="canonical" href="{BASE_URL}/" />
  <link rel="alternate" hreflang="id-ID" href="{BASE_URL}/" />
  <meta property="og:type" content="website" />
  <meta property="og:url" content="{BASE_URL}/" />
  <meta property="og:title" content={title} />
  <meta property="og:description" content={description} />
  <meta property="og:image" content="{BASE_URL}/og-suluk.png" />
  <meta property="og:image:width" content="1200" />
  <meta property="og:image:height" content="630" />
  <meta property="og:image:alt" content="Suluk - Software Travel Umrah" />
  <meta property="og:locale" content="id_ID" />
  <meta property="og:site_name" content="Suluk" />
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content={title} />
  <meta name="twitter:description" content={description} />
  <meta name="twitter:image" content="{BASE_URL}/og-suluk.png" />
  {@html jsonLd}
</svelte:head>

<LandingPage
  onGoToLogin={() => goto("/login")}
  onGoToRegister={() => goto("/daftar")}
  onNavigate={gotoMarketing}
/>
