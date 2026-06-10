<script>
  import { Check } from "lucide-svelte";
  import { toastState } from "../toast.svelte.js";
</script>

<div class="m-toast-wrap">
  {#each toastState.items as t (t.id)}
    {@const Icon = t.icon || Check}
    <div class="m-toast">
      <span
        style="width:22px;height:22px;border-radius:999px;background:var(--c-primary);display:flex;align-items:center;justify-content:center;flex-shrink:0;color:#fff"
      ><Icon size={13} /></span>
      {t.msg}
    </div>
  {/each}
</div>
