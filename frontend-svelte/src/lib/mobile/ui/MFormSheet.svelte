<script>
  import MSheet from "./MSheet.svelte";

  // fields: [{ key, label, type?, required?, options?, placeholder?, full?, inputmode? }]
  //   type: text(default) | tel | email | number | select | textarea | date
  //   options (select): [{ value, label }] or string[]
  let {
    open = false,
    title = "",
    fields = [],
    initial = {},
    submitLabel = "Simpan",
    onClose,
    onSubmit,
  } = $props();

  let values = $state({});
  let saving = $state(false);
  let error = $state("");
  let seeded = false;

  // Seed the form once each time it opens (not on every keystroke).
  $effect(() => {
    if (open && !seeded) {
      const v = {};
      for (const f of fields) v[f.key] = initial[f.key] ?? "";
      values = v;
      error = "";
      seeded = true;
    } else if (!open && seeded) {
      seeded = false;
    }
  });

  const optVal = (o) => (typeof o === "string" ? o : o.value);
  const optLabel = (o) => (typeof o === "string" ? o : o.label);

  async function submit() {
    for (const f of fields) {
      if (f.required && !String(values[f.key] ?? "").trim()) {
        error = f.label + " wajib diisi";
        return;
      }
    }
    saving = true;
    error = "";
    try {
      await onSubmit({ ...values });
      onClose?.();
    } catch (e) {
      error = e?.message || "Gagal menyimpan";
    } finally {
      saving = false;
    }
  }
</script>

<MSheet {open} {onClose} {title}>
  <div style="display:flex;flex-direction:column;gap:14px;padding-bottom:6px">
    {#each fields as f (f.key)}
      <div class="m-field">
        <label for="ff-{f.key}">{f.label}{f.required ? " *" : ""}</label>
        {#if f.type === "select"}
          <select id="ff-{f.key}" class="m-input" bind:value={values[f.key]}>
            <option value="" disabled>Pilih…</option>
            {#each f.options || [] as o}
              <option value={optVal(o)}>{optLabel(o)}</option>
            {/each}
          </select>
        {:else if f.type === "textarea"}
          <textarea id="ff-{f.key}" class="m-input" rows="3" placeholder={f.placeholder || ""} bind:value={values[f.key]}></textarea>
        {:else}
          <input
            id="ff-{f.key}"
            class="m-input"
            type={f.type === "number" ? "text" : f.type === "date" ? "date" : "text"}
            inputmode={f.inputmode || (f.type === "number" ? "numeric" : f.type === "tel" ? "tel" : "text")}
            placeholder={f.placeholder || ""}
            value={values[f.key]}
            oninput={(e) => (values[f.key] = f.type === "number" ? e.currentTarget.value.replace(/[^0-9]/g, "") : e.currentTarget.value)}
          />
        {/if}
      </div>
    {/each}

    {#if error}
      <div style="background:var(--c-danger-soft);color:var(--c-danger);font-size:13px;font-weight:600;padding:10px 12px;border-radius:10px">{error}</div>
    {/if}

    <button type="button" class="m-btn m-btn-primary" disabled={saving} onclick={submit}>
      {saving ? "Menyimpan…" : submitLabel}
    </button>
  </div>
</MSheet>
