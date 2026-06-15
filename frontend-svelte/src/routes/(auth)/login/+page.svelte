<script>
  import Seo from "$lib/components/Seo.svelte";
  import Login from "$lib/pages/Login.svelte";
  import { goto } from "$app/navigation";
  import { session, homePathForUser } from "$lib/stores/session.svelte.js";

  function handleLoginSuccess(userData) {
    session.setUserAndLoad(userData);
    goto(homePathForUser(userData));
  }
</script>

<Seo
  title="Login - Suluk"
  description="Login ke dashboard Suluk untuk kelola data jamaah dan operasional umrah."
  path="/login"
  robots="noindex,nofollow"
/>

<Login initialMode="login" onLoginSuccess={handleLoginSuccess} onBack={() => goto("/")} />
