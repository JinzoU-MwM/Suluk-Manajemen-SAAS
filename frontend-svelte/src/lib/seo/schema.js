// JSON-LD structured-data builders for the prerendered marketing pages.
// Rendered into <svelte:head> by Seo.svelte. Keep FAQ text in sync with the
// visible FAQ on each page (Google requires structured data to match content).
const BASE = "https://suluk.site";

export function breadcrumb(items) {
  return {
    "@context": "https://schema.org",
    "@type": "BreadcrumbList",
    itemListElement: items.map((it, i) => ({
      "@type": "ListItem",
      position: i + 1,
      name: it.name,
      item: `${BASE}${it.path}`,
    })),
  };
}

export function faqPage(faqs) {
  return {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: faqs.map((f) => ({
      "@type": "Question",
      name: f.q,
      acceptedAnswer: { "@type": "Answer", text: f.a },
    })),
  };
}

export function serviceSchema({ name, description, path }) {
  return {
    "@context": "https://schema.org",
    "@type": "Service",
    serviceType: name,
    name,
    description,
    provider: { "@type": "Organization", name: "Suluk", url: BASE },
    areaServed: { "@type": "Country", name: "Indonesia" },
    url: `${BASE}${path}`,
  };
}

export function article({ title, description, slug, datePublished }) {
  return {
    "@context": "https://schema.org",
    "@type": "Article",
    headline: title,
    description,
    inLanguage: "id-ID",
    datePublished,
    dateModified: datePublished,
    mainEntityOfPage: { "@type": "WebPage", "@id": `${BASE}/panduan/${slug}` },
    author: { "@type": "Organization", name: "Suluk", url: BASE },
    publisher: {
      "@type": "Organization",
      name: "Suluk",
      url: BASE,
      logo: { "@type": "ImageObject", url: `${BASE}/icon-512.png` },
    },
    image: `${BASE}/og-suluk.png`,
  };
}

// Build the <script type="application/ld+json"> tag string for {@html}. The
// closing tag is split so it doesn't terminate the component's own <script>.
export function jsonLdTags(schema) {
  if (!schema) return "";
  const arr = Array.isArray(schema) ? schema : [schema];
  return arr
    .map((o) => `<script type="application/ld+json">${JSON.stringify(o)}<\/script>`)
    .join("\n");
}
