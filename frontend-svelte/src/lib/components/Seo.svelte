<script>
  // Reusable per-page SEO head. Emits title, description, canonical, robots,
  // OG/Twitter cards, and optional JSON-LD structured data. Used by prerendered
  // marketing pages (index,follow) and the SPA app pages (noindex,nofollow).
  import { jsonLdTags } from "$lib/seo/schema.js";

  const BASE_URL = "https://suluk.site";
  let {
    title,
    description = "",
    path = "/",
    robots = "index,follow",
    image = `${BASE_URL}/og-suluk.png`,
    schema = null, // a JSON-LD object or array of objects
  } = $props();

  let canonical = $derived(`${BASE_URL}${path}`);
  let schemaHtml = $derived(jsonLdTags(schema));
</script>

<svelte:head>
  <title>{title}</title>
  <meta name="title" content={title} />
  <meta name="description" content={description} />
  <meta name="robots" content={robots} />
  <link rel="canonical" href={canonical} />
  <meta property="og:type" content="website" />
  <meta property="og:url" content={canonical} />
  <meta property="og:title" content={title} />
  <meta property="og:description" content={description} />
  <meta property="og:image" content={image} />
  <meta property="og:image:width" content="1200" />
  <meta property="og:image:height" content="630" />
  <meta property="og:image:alt" content="Suluk - Software Travel Umrah" />
  <meta property="og:locale" content="id_ID" />
  <meta property="og:site_name" content="Suluk" />
  <meta name="twitter:card" content="summary_large_image" />
  <meta name="twitter:title" content={title} />
  <meta name="twitter:description" content={description} />
  <meta name="twitter:image" content={image} />
  {@html schemaHtml}
</svelte:head>
