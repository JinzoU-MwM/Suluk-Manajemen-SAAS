<script>
  import { ArrowLeft, Loader2 } from "lucide-svelte";
  import { ApiService } from "../../services/api.js";
  import "../mobile.css";

  let { onLoginSuccess } = $props();

  // mode: "login" | "register" | "verify" | "forgot" | "forgot-reset"
  let mode = $state("login");
  let email = $state("");
  let password = $state("");
  let name = $state("");
  let error = $state("");
  let success = $state("");
  let isLoading = $state(false);

  let otpCode = $state("");
  let verifyAddr = $state("");
  let resetCode = $state("");
  let newPassword = $state("");
  let confirmPassword = $state("");
  let resendCooldown = $state(0);
  let cooldownInterval = null;

  function startCooldown(s = 60) {
    resendCooldown = s;
    if (cooldownInterval) clearInterval(cooldownInterval);
    cooldownInterval = setInterval(() => {
      resendCooldown--;
      if (resendCooldown <= 0) clearInterval(cooldownInterval);
    }, 1000);
  }
  function go(m) {
    mode = m;
    error = "";
    success = "";
  }

  async function handleLogin(e) {
    e.preventDefault();
    error = ""; success = ""; isLoading = true;
    try {
      const r = await ApiService.login(email, password);
      if (r.access_token) localStorage.setItem("access_token", r.access_token);
      if (r.refresh_token) localStorage.setItem("refresh_token", r.refresh_token);
      localStorage.setItem("user", JSON.stringify(r.user));
      onLoginSuccess(r.user);
    } catch (err) {
      const msg = err.message || "Gagal masuk";
      if (msg.includes("belum diverifikasi")) {
        verifyAddr = email; mode = "verify"; error = "";
        success = "Masukkan kode OTP yang dikirim ke email Anda.";
      } else error = msg;
    } finally {
      isLoading = false;
    }
  }

  async function handleRegister(e) {
    e.preventDefault();
    error = ""; success = "";
    if (password.length < 6) { error = "Kata sandi minimal 6 karakter"; return; }
    isLoading = true;
    try {
      const r = await ApiService.register(email, password, name);
      verifyAddr = email; mode = "verify";
      success = r.message || "Kode verifikasi dikirim ke email Anda";
      startCooldown(60);
    } catch (err) {
      error = err.message;
    } finally {
      isLoading = false;
    }
  }

  async function handleVerify(e) {
    e.preventDefault();
    error = ""; success = "";
    if (otpCode.length !== 6) { error = "Masukkan 6 digit kode OTP"; return; }
    isLoading = true;
    try {
      const r = await ApiService.verifyEmail(verifyAddr, otpCode);
      if (r.access_token) {
        localStorage.setItem("access_token", r.access_token);
        if (r.refresh_token) localStorage.setItem("refresh_token", r.refresh_token);
        localStorage.setItem("user", JSON.stringify(r.user));
        onLoginSuccess(r.user);
      } else { success = r.message; mode = "login"; }
    } catch (err) {
      error = err.message;
    } finally {
      isLoading = false;
    }
  }

  async function handleResend() {
    error = ""; success = ""; isLoading = true;
    try {
      const r = await ApiService.resendOtp(verifyAddr);
      success = r.message || "Kode verifikasi baru dikirim";
      startCooldown(60);
    } catch (err) {
      error = err.message;
    } finally {
      isLoading = false;
    }
  }

  async function handleForgotSend(e) {
    e.preventDefault();
    error = ""; success = ""; isLoading = true;
    try {
      const r = await ApiService.forgotPassword(email);
      success = r.message || "Kode reset dikirim ke email Anda";
      verifyAddr = email; mode = "forgot-reset"; startCooldown(60);
    } catch (err) {
      error = err.message;
    } finally {
      isLoading = false;
    }
  }

  async function handleReset(e) {
    e.preventDefault();
    error = ""; success = "";
    if (newPassword.length < 6) { error = "Kata sandi baru minimal 6 karakter"; return; }
    if (newPassword !== confirmPassword) { error = "Konfirmasi kata sandi tidak cocok"; return; }
    isLoading = true;
    try {
      const r = await ApiService.resetPassword(verifyAddr, resetCode, newPassword);
      success = r.message || "Kata sandi berhasil direset. Silakan masuk.";
      mode = "login"; resetCode = ""; newPassword = ""; confirmPassword = "";
    } catch (err) {
      error = err.message;
    } finally {
      isLoading = false;
    }
  }

  let heading = $derived(
    mode === "register" ? "Buat Akun" :
    mode === "verify" ? "Verifikasi Email" :
    mode === "forgot" ? "Lupa Kata Sandi" :
    mode === "forgot-reset" ? "Atur Ulang Sandi" : "Selamat Datang",
  );
  let subheading = $derived(
    mode === "register" ? "Daftarkan travel Anda di Suluk" :
    mode === "verify" ? "Kode OTP dikirim ke " + verifyAddr :
    mode === "forgot" ? "Kirim kode reset ke email Anda" :
    mode === "forgot-reset" ? "Masukkan kode + sandi baru" : "Masuk untuk kelola travel Anda",
  );
</script>

<div class="ml-wrap">
  <div class="ml-hero">
    <img src="/brand/suluk-mark-white.png?v=1" alt="Suluk" class="ml-mark" />
    <div class="ml-brand">Suluk</div>
    <div class="ml-tag">ERP FOR TRAVEL</div>
  </div>

  <div class="ml-sheet m-slide">
    {#if mode !== "login"}
      <button type="button" class="ml-back" onclick={() => go(mode === "forgot-reset" ? "forgot" : "login")}>
        <ArrowLeft size={16} /> Kembali
      </button>
    {/if}

    <div class="ml-title">{heading}</div>
    <div class="ml-sub">{subheading}</div>

    {#if mode === "login" || mode === "register"}
      <div class="m-seg" style="margin:16px 0">
        <button type="button" class={mode === "login" ? "on" : ""} onclick={() => go("login")}>Masuk</button>
        <button type="button" class={mode === "register" ? "on" : ""} onclick={() => go("register")}>Daftar</button>
      </div>
    {/if}

    {#if error}<div class="ml-msg ml-err">{error}</div>{/if}
    {#if success}<div class="ml-msg ml-ok">{success}</div>{/if}

    {#if mode === "login"}
      <form onsubmit={handleLogin} class="ml-form">
        <div class="m-field"><label for="ml-email">Email</label><input id="ml-email" class="m-input" type="email" bind:value={email} placeholder="nama@travel.com" autocomplete="email" required /></div>
        <div class="m-field"><label for="ml-pass">Kata Sandi</label><input id="ml-pass" class="m-input" type="password" bind:value={password} placeholder="••••••••" autocomplete="current-password" required /></div>
        <button type="button" class="ml-link" onclick={() => go("forgot")}>Lupa kata sandi?</button>
        <button type="submit" class="m-btn m-btn-primary" disabled={isLoading}>{#if isLoading}<Loader2 size={18} class="m-spin" />{/if}Masuk</button>
      </form>
    {:else if mode === "register"}
      <form onsubmit={handleRegister} class="ml-form">
        <div class="m-field"><label for="ml-name">Nama</label><input id="ml-name" class="m-input" bind:value={name} placeholder="Nama Anda / travel" required /></div>
        <div class="m-field"><label for="ml-remail">Email</label><input id="ml-remail" class="m-input" type="email" bind:value={email} placeholder="nama@travel.com" autocomplete="email" required /></div>
        <div class="m-field"><label for="ml-rpass">Kata Sandi</label><input id="ml-rpass" class="m-input" type="password" bind:value={password} placeholder="Minimal 6 karakter" autocomplete="new-password" required /></div>
        <button type="submit" class="m-btn m-btn-primary" disabled={isLoading}>{#if isLoading}<Loader2 size={18} class="m-spin" />{/if}Daftar</button>
      </form>
    {:else if mode === "verify"}
      <form onsubmit={handleVerify} class="ml-form">
        <div class="m-field"><label for="ml-otp">Kode OTP (6 digit)</label><input id="ml-otp" class="m-input" inputmode="numeric" maxlength="6" bind:value={otpCode} placeholder="••••••" style="text-align:center;letter-spacing:6px;font-size:20px" /></div>
        <button type="submit" class="m-btn m-btn-primary" disabled={isLoading}>{#if isLoading}<Loader2 size={18} class="m-spin" />{/if}Verifikasi</button>
        <button type="button" class="ml-link" style="text-align:center;width:100%" disabled={resendCooldown > 0} onclick={handleResend}>
          {resendCooldown > 0 ? `Kirim ulang dalam ${resendCooldown}s` : "Kirim ulang kode"}
        </button>
      </form>
    {:else if mode === "forgot"}
      <form onsubmit={handleForgotSend} class="ml-form">
        <div class="m-field"><label for="ml-femail">Email</label><input id="ml-femail" class="m-input" type="email" bind:value={email} placeholder="nama@travel.com" autocomplete="email" required /></div>
        <button type="submit" class="m-btn m-btn-primary" disabled={isLoading}>{#if isLoading}<Loader2 size={18} class="m-spin" />{/if}Kirim Kode Reset</button>
      </form>
    {:else if mode === "forgot-reset"}
      <form onsubmit={handleReset} class="ml-form">
        <div class="m-field"><label for="ml-code">Kode Reset</label><input id="ml-code" class="m-input" inputmode="numeric" bind:value={resetCode} placeholder="Kode dari email" required /></div>
        <div class="m-field"><label for="ml-np">Kata Sandi Baru</label><input id="ml-np" class="m-input" type="password" bind:value={newPassword} placeholder="Minimal 6 karakter" autocomplete="new-password" required /></div>
        <div class="m-field"><label for="ml-cp">Konfirmasi Sandi</label><input id="ml-cp" class="m-input" type="password" bind:value={confirmPassword} placeholder="Ulangi kata sandi" autocomplete="new-password" required /></div>
        <button type="submit" class="m-btn m-btn-primary" disabled={isLoading}>{#if isLoading}<Loader2 size={18} class="m-spin" />{/if}Atur Ulang Sandi</button>
      </form>
    {/if}
  </div>
</div>

<style>
  .ml-wrap {
    position: fixed; inset: 0; height: 100dvh; z-index: 50;
    display: flex; flex-direction: column; overflow-y: auto;
    background: linear-gradient(165deg, var(--c-primary-deep), var(--c-primary));
    font-family: var(--font-ui); color: var(--c-ink);
  }
  .ml-hero {
    flex-shrink: 0; text-align: center; color: #fff;
    padding: calc(env(safe-area-inset-top) + 56px) 24px 32px;
  }
  .ml-mark { height: 72px; width: auto; display: inline-block; }
  .ml-brand { font-family: var(--font-display, Georgia, serif); font-size: 30px; font-weight: 800; margin-top: 12px; letter-spacing: -0.01em; }
  .ml-tag { font-size: 11px; font-weight: 700; letter-spacing: 0.22em; color: var(--c-accent); margin-top: 4px; }
  .ml-sheet {
    flex: 1; background: var(--c-surface);
    border-radius: 26px 26px 0 0; padding: 26px 22px calc(env(safe-area-inset-bottom) + 32px);
    box-shadow: 0 -8px 30px rgba(0,0,0,.12); min-height: 52%;
  }
  .ml-back {
    display: inline-flex; align-items: center; gap: 4px; font-size: 14px; font-weight: 600;
    color: var(--c-muted); margin-bottom: 14px;
  }
  .ml-back:active { opacity: .5; }
  .ml-title { font-family: var(--font-display, Georgia, serif); font-size: 26px; font-weight: 800; letter-spacing: -0.02em; }
  .ml-sub { font-size: 13.5px; color: var(--c-muted); margin-top: 4px; overflow-wrap: anywhere; }
  .ml-form { display: flex; flex-direction: column; gap: 14px; margin-top: 4px; }
  .ml-link { align-self: flex-end; font-size: 13px; font-weight: 600; color: var(--c-primary); }
  .ml-link:disabled { color: var(--c-faint); }
  .ml-msg { font-size: 13px; font-weight: 600; padding: 10px 13px; border-radius: 11px; margin-top: 14px; }
  .ml-err { background: var(--c-danger-soft); color: var(--c-danger); }
  .ml-ok { background: var(--c-success-soft); color: var(--c-success); }
</style>
