import { onMounted, onUnmounted, ref, type Ref } from 'vue'

const hasHover = () =>
  typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(hover: hover)').matches

const reduce = () =>
  typeof window !== 'undefined' &&
  window.matchMedia &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches

/**
 * Global cursor light: writes --mx/--my (px) onto :root so the fixed .cursor-glow
 * layer and any element reading the root vars can follow the pointer. rAF-throttled.
 * No-op on touch / reduced-motion devices.
 */
export function useCursorGlow() {
  let raf = 0
  let px = 0
  let py = 0
  const onMove = (e: PointerEvent) => {
    px = e.clientX
    py = e.clientY
    if (raf) return
    raf = requestAnimationFrame(() => {
      document.documentElement.style.setProperty('--mx', `${px}px`)
      document.documentElement.style.setProperty('--my', `${py}px`)
      raf = 0
    })
  }
  onMounted(() => {
    if (!hasHover() || reduce()) return
    window.addEventListener('pointermove', onMove, { passive: true })
  })
  onUnmounted(() => {
    window.removeEventListener('pointermove', onMove)
    if (raf) cancelAnimationFrame(raf)
  })
}

/**
 * Per-element spotlight: sets --mx/--my (%) on the element under the pointer so
 * `.spotlight::before` blooms locally. rAF-throttled, hover-only.
 */
export function useSpotlight() {
  const el = ref<HTMLElement | null>(null)
  let raf = 0
  let mx = 50
  let my = 50
  const onMove = (e: PointerEvent) => {
    const node = el.value
    if (!node) return
    const r = node.getBoundingClientRect()
    mx = ((e.clientX - r.left) / r.width) * 100
    my = ((e.clientY - r.top) / r.height) * 100
    if (raf) return
    raf = requestAnimationFrame(() => {
      node.style.setProperty('--mx', `${mx}%`)
      node.style.setProperty('--my', `${my}%`)
      raf = 0
    })
  }
  onMounted(() => {
    if (!hasHover() || reduce() || !el.value) return
    el.value.addEventListener('pointermove', onMove, { passive: true })
  })
  onUnmounted(() => {
    el.value?.removeEventListener('pointermove', onMove)
    if (raf) cancelAnimationFrame(raf)
  })
  return el
}

/**
 * Magnetic element: nudges toward the pointer (<=strength px) while hovered,
 * springs back on leave. Returns a template ref. Hover-only, rAF-throttled.
 */
export function useMagnet(strength = 6): Ref<HTMLElement | null> {
  const el = ref<HTMLElement | null>(null)
  let raf = 0
  const apply = (x: number, y: number) => {
    if (raf) cancelAnimationFrame(raf)
    raf = requestAnimationFrame(() => {
      if (el.value) el.value.style.transform = `translate(${x}px, ${y}px)`
    })
  }
  const onMove = (e: PointerEvent) => {
    const node = el.value
    if (!node) return
    const r = node.getBoundingClientRect()
    const dx = (e.clientX - (r.left + r.width / 2)) / (r.width / 2)
    const dy = (e.clientY - (r.top + r.height / 2)) / (r.height / 2)
    apply(dx * strength, dy * strength)
  }
  const onLeave = () => apply(0, 0)
  onMounted(() => {
    if (!hasHover() || reduce() || !el.value) return
    el.value.addEventListener('pointermove', onMove, { passive: true })
    el.value.addEventListener('pointerleave', onLeave, { passive: true })
  })
  onUnmounted(() => {
    el.value?.removeEventListener('pointermove', onMove)
    el.value?.removeEventListener('pointerleave', onLeave)
    if (raf) cancelAnimationFrame(raf)
  })
  return el
}

/**
 * Count-up: eases a reactive number from 0 to `target` once mounted. Returns a
 * ref holding the (rounded) display value. Reduced-motion shows the final value
 * immediately. Re-runs when the target changes.
 */
export function useCountUp(target: () => number, durationMs = 1100) {
  const display = ref(0)
  let raf = 0
  const run = () => {
    const end = target()
    if (reduce()) {
      display.value = end
      return
    }
    const begin = display.value
    const t0 = performance.now()
    const step = (now: number) => {
      const p = Math.min(1, (now - t0) / durationMs)
      const eased = 1 - Math.pow(1 - p, 3)
      display.value = Math.round(begin + (end - begin) * eased)
      if (p < 1) raf = requestAnimationFrame(step)
    }
    if (raf) cancelAnimationFrame(raf)
    raf = requestAnimationFrame(step)
  }
  onMounted(run)
  onUnmounted(() => raf && cancelAnimationFrame(raf))
  return { display, run }
}

/**
 * click-spark: bursts N cyan particles from the click point. One-shot, self-cleaning.
 */
export function clickSpark(e: MouseEvent, count = 6) {
  if (reduce()) return
  const x = e.clientX
  const y = e.clientY
  for (let i = 0; i < count; i++) {
    const angle = (Math.PI * 2 * i) / count
    const dist = 26 + (i % 3) * 8
    const p = document.createElement('span')
    p.className = 'spark'
    p.style.left = `${x}px`
    p.style.top = `${y}px`
    p.style.setProperty('--dx', `${Math.cos(angle) * dist}px`)
    p.style.setProperty('--dy', `${Math.sin(angle) * dist}px`)
    document.body.appendChild(p)
    window.setTimeout(() => p.remove(), 520)
  }
}
