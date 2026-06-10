<script>
  import Seo from "$lib/components/Seo.svelte";
  import GuideKeuangan from "$lib/pages/marketing/GuideKeuangan.svelte";
  import { gotoMarketing, gotoApp } from "$lib/stores/session.svelte.js";
  import { breadcrumb, faqPage, serviceSchema } from "$lib/seo/schema.js";
  import { FAQ } from "$lib/seo/faq.js";

  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Software Travel Umrah", path: "/software-travel-umrah" },
      { name: "Laporan Keuangan", path: "/fitur/laporan-keuangan" },
    ]),
    serviceSchema({
      name: "Laporan Keuangan Travel Umrah",
      description:
        "Laba rugi per trip, arus kas, piutang, dan biaya operasional travel umrah secara real-time tanpa rekap manual.",
      path: "/fitur/laporan-keuangan",
    }),
    faqPage(FAQ.keuangan),
  ];
</script>

<Seo
  title="Laporan Keuangan Travel Umrah Real-Time | Suluk"
  description="Laba rugi per trip, arus kas, piutang, dan biaya operasional travel umrah secara real-time — tanpa rekap manual di spreadsheet."
  path="/fitur/laporan-keuangan"
  {schema}
/>

<GuideKeuangan onNav={gotoMarketing} onGoToApp={gotoApp} />
