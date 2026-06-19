<script>
  import Seo from "$lib/components/Seo.svelte";
  import BantuanArticlePage from "$lib/pages/BantuanArticlePage.svelte";

  let { data } = $props();
</script>

<Seo
  title={`${data.guide.title} - Bantuan Suluk`}
  description={data.guide.summary}
  path={`/app/bantuan/${data.guide.slug}`}
  robots="noindex,nofollow"
/>

<BantuanArticlePage area="app" basePath="/app/bantuan" guide={data.guide} />
