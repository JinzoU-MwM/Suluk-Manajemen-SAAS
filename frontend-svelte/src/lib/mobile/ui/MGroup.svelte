<script>
  let { class: cls = "", style = "", children } = $props();
</script>

<div class="m-card {cls}" style="overflow:hidden;{style}">{@render children?.()}</div>
