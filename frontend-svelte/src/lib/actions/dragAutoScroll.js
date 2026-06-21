// Symmetric drag-auto-scroll for horizontally scrollable containers (e.g. the
// CRM Kanban). Native HTML5 drag edge-scroll has a tiny, layout-dependent
// trigger zone — and on the left it gets squeezed by the app chrome — so we
// drive the scroll ourselves with equal, generous hotspots on both edges.

/**
 * Horizontal auto-scroll velocity (px per animation frame) for a pointer at
 * `pointerX` over a container spanning [`rectLeft`, `rectRight`].
 * Negative scrolls left, positive scrolls right, 0 means no scroll. Speed ramps
 * linearly from 0 at the inner edge of the hotspot to `maxSpeed` at (or past)
 * the container edge.
 *
 * @param {number} pointerX  viewport X of the pointer (e.clientX)
 * @param {number} rectLeft  container left edge (getBoundingClientRect().left)
 * @param {number} rectRight container right edge
 * @param {number} edge      hotspot width in px at each side
 * @param {number} maxSpeed  max px/frame
 * @returns {number}
 */
export function edgeScrollVelocity(pointerX, rectLeft, rectRight, edge, maxSpeed) {
  const leftDepth = rectLeft + edge - pointerX;
  if (leftDepth > 0) {
    const ramp = Math.min(leftDepth, edge) / edge;
    return -Math.ceil(ramp * maxSpeed);
  }
  const rightDepth = pointerX - (rectRight - edge);
  if (rightDepth > 0) {
    const ramp = Math.min(rightDepth, edge) / edge;
    return Math.ceil(ramp * maxSpeed);
  }
  return 0;
}

/**
 * Svelte action: while a drag is in progress over `node`, auto-scroll it
 * horizontally when the pointer is near either edge.
 *
 * @param {HTMLElement} node
 * @param {{ edge?: number, maxSpeed?: number }} [options]
 */
export function dragAutoScroll(node, options = {}) {
  let edge = options.edge ?? 80;
  let maxSpeed = options.maxSpeed ?? 18;
  let pointerX = null;
  let raf = 0;

  function step() {
    if (pointerX == null) {
      raf = 0;
      return;
    }
    const rect = node.getBoundingClientRect();
    const v = edgeScrollVelocity(pointerX, rect.left, rect.right, edge, maxSpeed);
    if (v !== 0) node.scrollLeft += v;
    raf = requestAnimationFrame(step);
  }

  function onDragOver(e) {
    pointerX = e.clientX;
    if (!raf) raf = requestAnimationFrame(step);
  }

  function stop() {
    pointerX = null;
    if (raf) {
      cancelAnimationFrame(raf);
      raf = 0;
    }
  }

  function onDragLeave(e) {
    // Ignore leaves into inner columns/cards; stop only when the pointer truly
    // exits the container (or there is no related target, e.g. left the window).
    if (!e.relatedTarget || !node.contains(e.relatedTarget)) stop();
  }

  node.addEventListener('dragover', onDragOver);
  node.addEventListener('drop', stop);
  node.addEventListener('dragend', stop);
  node.addEventListener('dragleave', onDragLeave);

  return {
    update(opts = {}) {
      edge = opts.edge ?? edge;
      maxSpeed = opts.maxSpeed ?? maxSpeed;
    },
    destroy() {
      stop();
      node.removeEventListener('dragover', onDragOver);
      node.removeEventListener('drop', stop);
      node.removeEventListener('dragend', stop);
      node.removeEventListener('dragleave', onDragLeave);
    },
  };
}