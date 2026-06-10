<script>
  import { ChevronRight } from "lucide-svelte";
  import MAvatar from "./MAvatar.svelte";

  let {
    icon = null,
    iconColor = "var(--c-primary)",
    avatar = null,
    title = "",
    sub = "",
    right = null,
    onclick = null,
    chevron = true,
    children,
  } = $props();
  const Icon = icon;
</script>

<svelte:element this={onclick ? "button" : "div"} type="button" class="m-row" onclick={onclick}>
  {#if avatar}
    <MAvatar name={avatar} size={40} />
  {:else if Icon}
    <div class="m-row-ic" style="background:{iconColor}1c;color:{iconColor}"><Icon size={19} /></div>
  {/if}
  <div class="m-row-main">
    <div class="m-row-title">{title}</div>
    {#if sub}<div class="m-row-sub">{sub}</div>{/if}
  </div>
  {#if right}{@render right()}{/if}
  {#if children}{@render children()}{/if}
  {#if onclick && chevron}<ChevronRight size={18} class="m-chev" />{/if}
</svelte:element>
