<script>
  let { label = "", children } = $props();
</script>

<div class="m-field"><label>{label}</label>{@render children?.()}</div>
