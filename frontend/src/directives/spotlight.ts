import type { Directive } from 'vue'

// v-spotlight: writes local percentage-based --mx/--my variables so
// `.spotlight::before` blooms under the pointer instead of inheriting the
// global cursor coordinates. Hover-only and rAF-throttled for low overhead.
const hasHover = () =>
  typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(hover: hover)').matches

const reduce = () =>
  typeof window !== 'undefined' &&
  window.matchMedia &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches

interface SpotlightState {
  raf: number
  handler: (e: PointerEvent) => void
}

const states = new WeakMap<HTMLElement, SpotlightState>()

export const spotlight: Directive<HTMLElement> = {
  mounted(el) {
    if (!hasHover() || reduce()) return

    const state: SpotlightState = {
      raf: 0,
      handler: (e: PointerEvent) => {
        if (state.raf) return
        state.raf = requestAnimationFrame(() => {
          const r = el.getBoundingClientRect()
          el.style.setProperty('--mx', `${((e.clientX - r.left) / r.width) * 100}%`)
          el.style.setProperty('--my', `${((e.clientY - r.top) / r.height) * 100}%`)
          state.raf = 0
        })
      },
    }

    states.set(el, state)
    el.addEventListener('pointermove', state.handler, { passive: true })
  },
  unmounted(el) {
    const state = states.get(el)
    if (!state) return
    el.removeEventListener('pointermove', state.handler)
    if (state.raf) cancelAnimationFrame(state.raf)
    states.delete(el)
  },
}
