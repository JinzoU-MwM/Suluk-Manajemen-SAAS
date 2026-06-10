<script>
  // Routes a marketing slug to its content page inside the shared MarketingShell.
  import MarketingShell from "./MarketingShell.svelte";
  import About from "./About.svelte";
  import Contact from "./Contact.svelte";
  import Privacy from "./Privacy.svelte";
  import Terms from "./Terms.svelte";
  import NotFound from "./NotFound.svelte";
  import Pillar from "./Pillar.svelte";
  import GuideInvoice from "./GuideInvoice.svelte";
  import GuideCrm from "./GuideCrm.svelte";
  import GuideKeuangan from "./GuideKeuangan.svelte";
  import GuideKontrak from "./GuideKontrak.svelte";
  import GuidePayroll from "./GuidePayroll.svelte";
  import Unduh from "./Unduh.svelte";

  let { slug = "", onNav = () => {}, onGoToApp = () => {}, onHome = () => {} } = $props();

  const MAP = {
    about: About,
    contact: Contact,
    privacy: Privacy,
    terms: Terms,
    software: Pillar,
    "fitur-invoice": GuideInvoice,
    "fitur-crm": GuideCrm,
    "fitur-keuangan": GuideKeuangan,
    "fitur-kontrak": GuideKontrak,
    "fitur-payroll": GuidePayroll,
    unduh: Unduh,
  };
  let Comp = $derived(MAP[slug] || NotFound);
</script>

<MarketingShell {onNav} {onGoToApp} {onHome}>
  <Comp {onNav} {onGoToApp} />
</MarketingShell>
