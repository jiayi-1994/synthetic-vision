<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { PRESET_CATALOG, byCategory } from '@/lib/presets'
import { clickSpark } from '@/composables/useInteractions'
import type { Preset, PresetCategory } from '@/types'

const router = useRouter()
const selectedCategory = ref<PresetCategory | 'all'>('all')

const categories = [
  { key: 'all', label: '全部' },
  { key: 'photoreal', label: '写实' },
  { key: 'illustration', label: '插画' },
  { key: 'abstract', label: '抽象' },
  { key: 'product', label: '商品' },
  { key: 'retro', label: '复古' },
  { key: 'portrait', label: '人像' },
] as const

const visiblePresets = computed(() => byCategory(PRESET_CATALOG, selectedCategory.value))

// Per-card spotlight: set --mx/--my on the hovered <article> (rAF-throttled).
let raf = 0
function onCardMove(e: PointerEvent) {
  const el = e.currentTarget as HTMLElement
  if (raf) return
  raf = requestAnimationFrame(() => {
    const r = el.getBoundingClientRect()
    el.style.setProperty('--mx', `${((e.clientX - r.left) / r.width) * 100}%`)
    el.style.setProperty('--my', `${((e.clientY - r.top) / r.height) * 100}%`)
    raf = 0
  })
}

function applyPreset(preset: Preset, e: MouseEvent) {
  clickSpark(e)
  router.push({ path: '/', query: { preset: preset.id } })
}
</script>

<template>
  <main class="flex-1 p-margin-sm md:p-margin-lg overflow-y-auto relative">
    <div class="max-w-container-max mx-auto space-y-6">
      <section v-reveal class="glass-panel rounded-2xl p-6 md:p-8 grad-border relative overflow-hidden">
        <p class="font-mono text-micro text-on-surface-variant mb-3 uppercase">创作加速器 · ACCELERATOR</p>
        <h1 class="font-display text-headline-lg-mobile md:text-headline-lg font-bold text-on-surface mb-2">
          市场<span class="text-neon">预设</span>
        </h1>
        <p class="font-body-md text-body-md text-on-surface-variant max-w-2xl">
          使用内置模板快速启动创作方向。选择一个场景预设后会跳转到工作台进行参数确认，可自由调整后再生成。
        </p>
      </section>

      <!-- Category chips -->
      <div class="flex flex-wrap gap-2" v-reveal>
        <button
          v-for="c in categories"
          :key="c.key"
          type="button"
          class="min-h-[44px] px-4 rounded-full border transition-all font-mono text-[12px] tracking-wider"
          :class="
            selectedCategory === c.key
              ? 'bg-primary/15 text-primary border-primary/50 shadow-[0_0_16px_rgba(56,232,255,0.15)]'
              : 'bg-surface-container text-on-surface-variant border-outline-variant/40 hover:border-primary/30 hover:text-primary'
          "
          @click="selectedCategory = c.key as PresetCategory | 'all'"
        >
          {{ c.label }}
        </button>
      </div>

      <!-- Preset cards -->
      <section class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6">
        <article
          v-for="(preset, i) in visiblePresets"
          :key="preset.id"
          v-reveal="(i % 2) * 70"
          class="spotlight glass-panel glow-hover rounded-2xl p-6"
          @pointermove="onCardMove"
        >
          <div class="flex items-start justify-between gap-4 relative z-10">
            <div>
              <p class="font-display text-lg font-bold text-on-surface mb-2">{{ preset.title }}</p>
              <p class="font-mono text-[11px] text-primary mb-2 uppercase tracking-wider">{{ preset.preview }}</p>
              <p class="font-body-md text-body-md text-on-surface-variant">{{ preset.description }}</p>
            </div>
            <span
              class="flex-shrink-0 inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-secondary/10 text-secondary border border-secondary/30 font-mono text-[11px]"
            >
              <span class="material-symbols-outlined text-[14px]">toll</span>{{ preset.estimated_cost }}
            </span>
          </div>

          <div
            class="mt-5 pt-4 border-t border-outline-variant/30 flex flex-wrap gap-2 items-center justify-between relative z-10"
          >
            <div class="flex items-center gap-2 font-mono text-[11px] text-on-surface-variant">
              <span class="inline-flex px-2 py-1 rounded-lg bg-surface-container border border-outline-variant/30">{{ preset.suggested_resolution }}</span>
              <span class="inline-flex px-2 py-1 rounded-lg bg-surface-container border border-outline-variant/30">{{ preset.suggested_aspect_ratio }}</span>
              <span class="inline-flex px-2 py-1 rounded-lg bg-tertiary/10 text-tertiary border border-tertiary/30">{{ preset.style }}</span>
            </div>
            <button
              type="button"
              class="btn-primary text-sm py-2 px-4 flex items-center gap-1.5"
              @click="applyPreset(preset, $event)"
            >
              应用 <span class="material-symbols-outlined text-[16px]">arrow_forward</span>
            </button>
          </div>

          <div class="mt-4 flex flex-wrap gap-2 relative z-10">
            <span
              v-for="tag in preset.tags"
              :key="`${preset.id}-${tag}`"
              class="px-2 py-1 rounded-full font-mono text-[10px] bg-surface-container-lowest border border-outline-variant/30 text-on-surface-variant"
            >
              #{{ tag }}
            </span>
          </div>
        </article>
      </section>

      <section v-if="visiblePresets.length === 0" class="glass-panel rounded-2xl p-6 text-center">
        <p class="text-on-surface-variant">当前分类暂无可用预设，请选择其它标签。</p>
      </section>
    </div>
  </main>
</template>
