<script>
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
</script>

<svelte:head>
  <title>Halaman tidak ditemukan - Suluk</title>
  <meta name="robots" content="noindex" />
</svelte:head>

<div class="min-h-screen flex items-center justify-center bg-slate-50 px-6">
  <div class="text-center">
    <div class="text-6xl font-bold text-emerald-700 mb-2">{$page.status}</div>
    <h1 class="text-xl font-semibold text-slate-800 mb-2">
      {$page.status === 404 ? "Halaman tidak ditemukan" : "Terjadi kesalahan"}
    </h1>
    <p class="text-slate-500 mb-6">
      {$page.error?.message || "Maaf, halaman yang Anda cari tidak tersedia."}
    </p>
    <button
      class="px-5 py-2.5 rounded-lg bg-emerald-700 text-white font-semibold hover:bg-emerald-800"
      onclick={() => goto("/")}
    >
      Kembali ke beranda
    </button>
  </div>
</div>
