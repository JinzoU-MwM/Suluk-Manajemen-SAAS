<script>
  import { avatarColor, initials } from "../format.js";
  let { name = "", size = 42 } = $props();
  let c = $derived(avatarColor(name));
</script>

<div
  style="width:{size}px;height:{size}px;border-radius:50%;flex-shrink:0;background:{c}22;color:{c};display:flex;align-items:center;justify-content:center;font-weight:700;font-size:{size *
    0.36}px"
>{initials(name)}</div>
