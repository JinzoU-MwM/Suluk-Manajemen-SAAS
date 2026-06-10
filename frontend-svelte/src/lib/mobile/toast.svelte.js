// Suluk Mobile — global toast store (Svelte 5 runes module)
export const toastState = $state({ items: [] });
let seq = 0;

/** Show a transient toast. icon is an optional lucide-svelte component. */
export function toast(msg, icon = null) {
  const id = ++seq;
  toastState.items = [...toastState.items, { id, msg, icon }];
  setTimeout(() => {
    toastState.items = toastState.items.filter((t) => t.id !== id);
  }, 2400);
}
