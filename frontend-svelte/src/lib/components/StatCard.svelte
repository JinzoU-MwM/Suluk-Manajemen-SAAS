<!-- StatCard.svelte — Suluk design stat card: accent icon tile + value + label + optional delta/sub. -->
<script>
  import { ArrowUp, ArrowDown } from "lucide-svelte";
  let { icon = null, label = "", value = "", accent = "#1B7F5A", delta = null, deltaUp = true, sub = "" } = $props();
  let Icon = $derived(icon);
</script>

<div class="rounded-2xl border border-slate-200/70 bg-white p-5 shadow-sm">
  <div class="mb-3.5 flex items-start justify-between">
    <div class="flex h-[42px] w-[42px] items-center justify-center rounded-xl" style="background:{accent}18;color:{accent}">
      {#if Icon}<Icon class="h-5 w-5" />{/if}
    </div>
    {#if delta != null}
      <span class="inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-bold" style={deltaUp ? "color:#1B7F5A;background:#E8F4EF" : "color:#c0392b;background:#fbe9e7"}>
        {#if deltaUp}<ArrowUp class="h-3 w-3" />{:else}<ArrowDown class="h-3 w-3" />{/if}{delta}
      </span>
    {/if}
  </div>
  <p class="text-[27px] font-extrabold leading-none tracking-tight text-[#10211c]" style="font-variant-numeric:tabular-nums">{value}</p>
  <p class="mt-1.5 text-[13.5px] font-medium text-slate-500">{label}</p>
  {#if sub}<p class="mt-1 text-xs text-slate-400">{sub}</p>{/if}
</div>
