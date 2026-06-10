<script>
  import Seo from "$lib/components/Seo.svelte";
  import GuideCrm from "$lib/pages/marketing/GuideCrm.svelte";
  import { gotoMarketing, gotoApp } from "$lib/stores/session.svelte.js";
  import { breadcrumb, faqPage, serviceSchema } from "$lib/seo/schema.js";
  import { FAQ } from "$lib/seo/faq.js";

  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Software Travel Umrah", path: "/software-travel-umrah" },
      { name: "CRM Jamaah", path: "/fitur/crm-jamaah" },
    ]),
    serviceSchema({
      name: "CRM Jamaah",
      description:
        "Pipeline CRM khusus travel umrah: lacak lead dari iklan hingga closing dan konversi calon jamaah jadi keberangkatan.",
      path: "/fitur/crm-jamaah",
    }),
    faqPage(FAQ.crm),
  ];
</script>

<Seo
  title="CRM Jamaah — Kelola Prospek hingga Berangkat | Suluk"
  description="Pipeline CRM khusus travel umrah: lacak lead dari iklan hingga closing, follow-up otomatis, dan konversi calon jamaah jadi keberangkatan."
  path="/fitur/crm-jamaah"
  {schema}
/>

<GuideCrm onNav={gotoMarketing} onGoToApp={gotoApp} />
