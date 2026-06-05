import type { Directive } from 'vue'

// v-reveal: fade/blur-up when the element scrolls into view. Adds `.reveal`
// immediately and `.is-in` on intersection. Honors prefers-reduced-motion by
// revealing instantly (CSS handles the static fallback).
const reduce = () =>
  typeof window !== 'undefined' &&
  window.matchMedia &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches

let io: IntersectionObserver | null = null
const ensureObserver = () => {
  if (io || typeof IntersectionObserver === 'undefined') return io
  io = new IntersectionObserver(
    (entries) => {
      for (const e of entries) {
        if (e.isIntersecting) {
          e.target.classList.add('is-in')
          io!.unobserve(e.target)
        }
      }
    },
    { threshold: 0.12, rootMargin: '0px 0px -8% 0px' }
  )
  return io
}

export const reveal: Directive<HTMLElement, number | undefined> = {
  mounted(el, binding) {
    el.classList.add('reveal')
    if (binding.value) el.style.transitionDelay = `${binding.value}ms`
    if (reduce() || typeof IntersectionObserver === 'undefined') {
      el.classList.add('is-in')
      return
    }
    ensureObserver()?.observe(el)
  },
  unmounted(el) {
    io?.unobserve(el)
  },
}
