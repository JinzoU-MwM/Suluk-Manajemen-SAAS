<script>
  // Shared public chrome (sticky nav + footer) for the SEO/legal/content pages.
  import MarketingShell from "$lib/pages/marketing/MarketingShell.svelte";
  import { goto } from "$app/navigation";
  import { gotoMarketing, gotoApp } from "$lib/stores/session.svelte.js";

  let { children } = $props();
</script>

<MarketingShell onNav={gotoMarketing} onGoToApp={gotoApp} onHome={() => goto("/")}>
  {@render children()}
</MarketingShell>
