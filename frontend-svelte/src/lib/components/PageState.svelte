<!--
  PageState.svelte — standardizes the loading / error / empty / content fork so every
  page handles the three states consistently.
  Props:
    loading      - show skeleton
    error        - error message string (run through mapError); shows a banner
    isEmpty      - show the `empty` snippet
    skeletonCount/skeletonType - forwarded to SkeletonLoader
  Snippets: `empty` (no-data view), default children (the content).
-->
<script>
  import { AlertCircle } from "lucide-svelte";
  import SkeletonLoader from "./SkeletonLoader.svelte";
  import { mapError } from "../services/toast.svelte.js";

  let {
    loading = false,
    error = "",
    isEmpty = false,
    skeletonCount = 5,
    skeletonType = "row",
    empty = null,
    children,
  } = $props();
</script>

{#if loading}
  <SkeletonLoader count={skeletonCount} type={skeletonType} />
{:else if error}
  <div class="flex items-start gap-3 rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
    <AlertCircle class="mt-0.5 h-5 w-5 flex-shrink-0" />
    <span>{mapError(error)}</span>
  </div>
{:else if isEmpty}
  {#if empty}{@render empty()}{/if}
{:else}
  {@render children?.()}
{/if}
