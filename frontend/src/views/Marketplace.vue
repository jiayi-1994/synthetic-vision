<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { PRESET_CATALOG, byCategory } from '@/lib/presets'
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

function applyPreset(preset: Preset) {
  router.push({ path: '/', query: { preset: preset.id } })
}
</script>

<template>
  <main class="flex-1 p-margin-sm md:p-margin-lg overflow-y-auto relative">
    <!-- Ambient Background Glow -->
    <div class="absolute top-0 left-1/4 w-1/2 h-64 bg-primary/5 blur-[120px] rounded-full pointer-events-none -z-10"></div>

    <div class="max-w-container-max mx-auto space-y-6">
      <section class="glass-panel rounded-xl p-6 md:p-8">
        <p class="font-label-sm text-label-sm text-on-surface-variant mb-2 uppercase tracking-wider">创作加速器</p>
        <h1 class="font-headline-lg md:font-headline-lg text-on-surface mb-2">市场预设</h1>
        <p class="font-body-md text-body-md text-on-surface-variant">
          使用内置模板快速启动创作方向。选择一个场景预设后会跳转到工作台进行参数确认，可自由调整后再生成。
        </p>
      </section>

      <!-- Category chips -->
      <div class="flex flex-wrap gap-2">
        <button
          v-for="c in categories"
          :key="c.key"
          type="button"
          class="px-4 py-2 rounded-full border transition-colors text-label-sm font-bold"
          :class="
            selectedCategory === c.key
              ? 'bg-primary text-on-primary border-primary/70'
              : 'bg-surface-container text-on-surface border-outline-variant/40 hover:bg-surface-variant/40'
          "
          @click="selectedCategory = c.key as PresetCategory | 'all'"
        >
          {{ c.label }}
        </button>
      </div>

      <!-- Preset cards -->
      <section class="grid grid-cols-1 lg:grid-cols-2 gap-4 md:gap-6">
        <article
          v-for="preset in visiblePresets"
          :key="preset.id"
          class="glass-panel rounded-xl p-6 border border-outline-variant/25"
        >
          <div class="flex items-start justify-between gap-4">
            <div>
              <div class="flex items-center gap-2 mb-3">
                <p class="font-bold text-lg text-on-surface">{{ preset.title }}</p>
              </div>
              <p class="font-label-sm text-label-sm text-primary mb-2">{{ preset.preview }}</p>
              <p class="font-body-md text-body-md text-on-surface-variant">{{ preset.description }}</p>
            </div>
            <span class="inline-flex items-center px-2 py-1 rounded-md bg-surface-container text-label-sm text-on-surface-variant border border-outline-variant/30">
              {{ preset.estimated_cost }} 积分
            </span>
          </div>

          <div class="mt-5 pt-4 border-t border-outline-variant/30 flex flex-wrap gap-2 items-center justify-between">
            <div class="flex items-center gap-2 text-label-sm text-on-surface-variant">
              <span class="inline-flex px-2 py-1 rounded-md bg-surface-container border border-outline-variant/30">{{ preset.suggested_resolution }}</span>
              <span class="inline-flex px-2 py-1 rounded-md bg-surface-container border border-outline-variant/30">{{ preset.suggested_aspect_ratio }}</span>
              <span class="inline-flex px-2 py-1 rounded-md bg-surface-container border border-outline-variant/30">{{ preset.style }}</span>
            </div>
            <button
              type="button"
              class="bg-primary text-on-primary font-bold px-4 py-2 rounded-lg hover:bg-primary/90 transition-colors"
              @click="applyPreset(preset)"
            >
              应用到工作台
            </button>
          </div>

          <div class="mt-4 flex flex-wrap gap-2">
            <span
              v-for="tag in preset.tags"
              :key="`${preset.id}-${tag}`"
              class="px-2 py-1 rounded-full text-[11px] bg-surface-container-lowest border border-outline-variant/30 text-on-surface-variant"
            >
              #{{ tag }}
            </span>
          </div>
        </article>
      </section>      <section class="glass-panel rounded-xl p-6 text-center" v-if="visiblePresets.length === 0">
        <p class="text-on-surface-variant">当前分类暂无可用预设，请选择其它标签。</p>
      </section>
    </div>
  </main>
</template>

