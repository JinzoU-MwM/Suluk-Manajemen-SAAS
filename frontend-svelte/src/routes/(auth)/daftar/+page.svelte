<script>
  import Seo from "$lib/components/Seo.svelte";
  import Login from "$lib/pages/Login.svelte";
  import { goto } from "$app/navigation";
  import { session, homePath } from "$lib/stores/session.svelte.js";

  function handleLoginSuccess(userData) {
    session.setUserAndLoad(userData);
    goto(homePath());
  }
</script>

<Seo
  title="Daftar - Suluk"
  description="Daftar akun Suluk untuk mulai otomatisasi data jamaah."
  path="/daftar"
  robots="noindex,nofollow"
/>

<Login initialMode="register" onLoginSuccess={handleLoginSuccess} onBack={() => goto("/")} />
