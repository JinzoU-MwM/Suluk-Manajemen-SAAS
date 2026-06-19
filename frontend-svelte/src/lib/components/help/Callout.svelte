<script>
  // Kotak sorotan untuk panduan: info | tip | warning. Dipakai oleh GuideBody.
  // Gaya mengikuti chip lembut aplikasi (Tailwind soft colors).
  import { Info, Lightbulb, TriangleAlert } from "lucide-svelte";

  let { variant = "info", text = "" } = $props();

  const META = {
    info: { Icon: Info, label: "Info", box: "bg-blue-50 text-blue-900", ico: "text-blue-500" },
    tip: { Icon: Lightbulb, label: "Tips", box: "bg-primary-50 text-primary-800", ico: "text-primary-600" },
    warning: { Icon: TriangleAlert, label: "Perhatian", box: "bg-amber-50 text-amber-900", ico: "text-amber-600" },
  };

  let meta = $derived(META[variant] ?? META.info);
  let Icon = $derived(meta.Icon);
</script>

<div class="my-3.5 flex items-start gap-2.5 rounded-xl p-3.5 text-[13.5px] leading-relaxed {meta.box}" role="note" aria-label={meta.label}>
  <Icon class="mt-0.5 h-[18px] w-[18px] shrink-0 {meta.ico}" aria-hidden="true" />
  <p class="m-0">{text}</p>
</div>
