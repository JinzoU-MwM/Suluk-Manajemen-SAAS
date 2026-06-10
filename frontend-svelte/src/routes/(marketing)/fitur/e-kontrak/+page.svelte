<script>
  import Seo from "$lib/components/Seo.svelte";
  import GuideKontrak from "$lib/pages/marketing/GuideKontrak.svelte";
  import { gotoMarketing, gotoApp } from "$lib/stores/session.svelte.js";
  import { breadcrumb, faqPage, serviceSchema } from "$lib/seo/schema.js";
  import { FAQ } from "$lib/seo/faq.js";

  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Software Travel Umrah", path: "/software-travel-umrah" },
      { name: "E-Kontrak", path: "/fitur/e-kontrak" },
    ]),
    serviceSchema({
      name: "E-Kontrak Digital Jamaah",
      description:
        "Akad perjalanan ditandatangani jamaah dari ponsel, template kontrak dengan variabel otomatis, dan arsip PDF immutable.",
      path: "/fitur/e-kontrak",
    }),
    faqPage(FAQ.kontrak),
  ];
</script>

<Seo
  title="E-Kontrak Digital Jamaah — Tanda Tangan Online | Suluk"
  description="Kirim akad perjalanan untuk ditandatangani jamaah langsung dari ponsel. Template kontrak dengan variabel otomatis dan arsip digital yang rapi."
  path="/fitur/e-kontrak"
  {schema}
/>

<GuideKontrak onNav={gotoMarketing} onGoToApp={gotoApp} />
