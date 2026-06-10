<script>
  import Seo from "$lib/components/Seo.svelte";
  import PayrollPage from "$lib/pages/PayrollPage.svelte";
  import { session, navigate } from "$lib/stores/session.svelte.js";
</script>

<Seo title="Penggajian - Suluk" description="Penggajian karyawan & honor muthawwif." path="/app/payroll" robots="noindex,nofollow" />

<PayrollPage onNavigate={navigate} user={session.user} />
