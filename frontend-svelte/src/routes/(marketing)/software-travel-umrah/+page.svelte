<script>
  import Seo from "$lib/components/Seo.svelte";
  import Pillar from "$lib/pages/marketing/Pillar.svelte";
  import { gotoMarketing, gotoApp } from "$lib/stores/session.svelte.js";
  import { breadcrumb, faqPage, serviceSchema } from "$lib/seo/schema.js";
  import { FAQ } from "$lib/seo/faq.js";

  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Software Travel Umrah", path: "/software-travel-umrah" },
    ]),
    serviceSchema({
      name: "Software Travel Umrah",
      description:
        "Platform operasional all-in-one untuk travel umrah & haji: jamaah, paket, invoice, keuangan, CRM, e-kontrak, rooming, vendor, dan payroll.",
      path: "/software-travel-umrah",
    }),
    faqPage(FAQ.pillar),
  ];
</script>

<Seo
  title="Software Travel Umrah Lengkap — Satu Platform untuk Seluruh Operasional | Suluk"
  description="Satu software untuk seluruh operasional travel umrah & haji: pendaftaran jamaah, paket, invoice, keuangan, CRM, e-kontrak, rooming, vendor, dan payroll."
  path="/software-travel-umrah"
  {schema}
/>

<Pillar onNav={gotoMarketing} onGoToApp={gotoApp} />
