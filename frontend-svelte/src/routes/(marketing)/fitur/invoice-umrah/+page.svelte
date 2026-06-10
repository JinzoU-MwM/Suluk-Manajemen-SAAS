<script>
  import Seo from "$lib/components/Seo.svelte";
  import GuideInvoice from "$lib/pages/marketing/GuideInvoice.svelte";
  import { gotoMarketing, gotoApp } from "$lib/stores/session.svelte.js";
  import { breadcrumb, faqPage, serviceSchema } from "$lib/seo/schema.js";
  import { FAQ } from "$lib/seo/faq.js";

  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Software Travel Umrah", path: "/software-travel-umrah" },
      { name: "Invoice Umrah", path: "/fitur/invoice-umrah" },
    ]),
    serviceSchema({
      name: "Invoice & Tagihan Travel Umrah",
      description:
        "Invoice jamaah otomatis dari paket, cicilan, pemantauan tunggakan, dan kwitansi PDF untuk travel umrah & haji.",
      path: "/fitur/invoice-umrah",
    }),
    faqPage(FAQ.invoice),
  ];
</script>

<Seo
  title="Invoice & Tagihan Cicilan Otomatis untuk Travel Umrah | Suluk"
  description="Buat invoice jamaah otomatis dari paket, catat cicilan, pantau tunggakan, dan cetak kwitansi PDF — khusus alur kerja travel umrah & haji."
  path="/fitur/invoice-umrah"
  {schema}
/>

<GuideInvoice onNav={gotoMarketing} onGoToApp={gotoApp} />
