<script>
  // Wraps a Pro-only page: renders its children when the user is Pro, otherwise
  // the upsell/trial screen — mirrors the inline pro-gating that lived in App.svelte.
  import ProGateScreen from "$lib/components/ProGateScreen.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";

  let { name, desc, highlights = [], children } = $props();
</script>

{#if session.isPro}
  {@render children()}
{:else}
  <ProGateScreen
    featureName={name}
    featureDescription={desc}
    {highlights}
    trialAvailable={session.trialAvailable}
    onUpgrade={() => navigate("profile:upgrade")}
    onTrial={() => navigate("profile:upgrade")}
  />
{/if}
