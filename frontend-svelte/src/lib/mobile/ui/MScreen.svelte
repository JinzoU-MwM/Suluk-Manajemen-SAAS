<script>
  import { ChevronLeft } from "lucide-svelte";
  let { title = "", onBack, right = null, children } = $props();
</script>

<div class="m-screen m-slide">
  <div class="m-nav">
    <button type="button" class="m-nav-btn" onclick={onBack}><ChevronLeft size={22} />Kembali</button>
    <div class="m-nav-title">{title}</div>
    {#if right}{@render right()}{:else}<div style="width:38px"></div>{/if}
  </div>
  <div class="m-scroll">
    {@render children?.()}
    <div style="height:28px"></div>
  </div>
</div>
