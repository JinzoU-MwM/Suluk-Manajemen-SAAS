<script>
  import { statusTone } from "../format.js";
  let { status = "", tone = null, dot = false, children = null } = $props();
  let pair = $derived(tone ? [tone, tone + "1c"] : statusTone(status));
</script>

<span class="m-chip" style="background:{pair[1]};color:{pair[0]}">
  {#if dot}<span style="width:6px;height:6px;border-radius:999px;background:{pair[0]}"></span>{/if}
  {#if children}{@render children()}{:else}{status}{/if}
</span>
