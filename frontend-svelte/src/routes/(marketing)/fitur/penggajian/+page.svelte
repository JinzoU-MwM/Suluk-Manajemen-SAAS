<script>
  import Seo from "$lib/components/Seo.svelte";
  import GuidePayroll from "$lib/pages/marketing/GuidePayroll.svelte";
  import { gotoMarketing, gotoApp } from "$lib/stores/session.svelte.js";
  import { breadcrumb, faqPage, serviceSchema } from "$lib/seo/schema.js";
  import { FAQ } from "$lib/seo/faq.js";

  const schema = [
    breadcrumb([
      { name: "Beranda", path: "/" },
      { name: "Software Travel Umrah", path: "/software-travel-umrah" },
      { name: "Penggajian", path: "/fitur/penggajian" },
    ]),
    serviceSchema({
      name: "Penggajian & Honor Muthawwif",
      description:
        "Gaji karyawan dan honor muthawwif per trip, hitung komponen termasuk PPh 21 & BPJS, dan rekap transfer untuk pembayaran massal.",
      path: "/fitur/penggajian",
    }),
    faqPage(FAQ.payroll),
  ];
</script>

<Seo
  title="Penggajian Karyawan & Honor Muthawwif Travel Umrah | Suluk"
  description="Kelola gaji karyawan dan honor muthawwif per trip dengan rapi. Hitung komponen, catat pembayaran, dan integrasikan ke laporan keuangan."
  path="/fitur/penggajian"
  {schema}
/>

<GuidePayroll onNav={gotoMarketing} onGoToApp={gotoApp} />
