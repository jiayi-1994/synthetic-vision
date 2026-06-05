<script setup lang="ts">
import type { Generation } from '@/types'
import { relativeTime } from '@/lib/format'
import { useSpotlight } from '@/composables/useInteractions'

const props = defineProps<{ gen: Generation }>()
const emit = defineEmits<{ (e: 'delete', id: string): void }>()

const card = useSpotlight()

function download() {
  const a = document.createElement('a')
  a.href = props.gen.image_url
  a.download = `synthetic-vision-${props.gen.id}.png`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}
</script>

<template>
  <div
    ref="card"
    class="spotlight group relative aspect-[4/5] rounded-2xl overflow-hidden glass-panel glow-hover"
  >
    <img
      :alt="gen.prompt"
      :src="gen.image_url"
      class="w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
    />
    <!-- neon corner ticks -->
    <span class="pointer-events-none absolute top-3 left-3 w-4 h-4 border-l border-t border-primary/40 rounded-tl"></span>
    <span class="pointer-events-none absolute bottom-3 right-3 w-4 h-4 border-r border-b border-secondary/40 rounded-br opacity-0 group-hover:opacity-100 transition-opacity"></span>

    <!-- Hover Overlay -->
    <div
      class="absolute inset-0 bg-gradient-to-t from-void/95 via-surface/30 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex flex-col justify-end p-4 z-10"
    >
      <div class="translate-y-4 group-hover:translate-y-0 transition-transform duration-300">
        <p class="font-mono text-[11px] tracking-wider text-primary mb-1.5 uppercase">
          {{ gen.width }}×{{ gen.height }} · {{ relativeTime(gen.created_at) }}
        </p>
        <p class="text-sm text-on-surface line-clamp-2 mb-4 leading-snug">
          "{{ gen.prompt }}"
        </p>
        <div class="flex items-center gap-2">
          <button
            class="flex-1 bg-primary/10 hover:bg-primary-container text-primary hover:text-on-primary-container py-2 rounded-xl backdrop-blur-sm border border-primary/30 flex justify-center items-center gap-2 transition-colors text-sm font-display font-semibold"
            type="button"
            @click="download"
          >
            <span class="material-symbols-outlined text-[18px]">download</span> 下载
          </button>
          <button
            class="w-10 h-10 bg-surface-variant/80 hover:bg-error/20 text-on-surface hover:text-error rounded-xl backdrop-blur-sm border border-outline-variant/30 hover:border-error/40 flex justify-center items-center transition-colors"
            type="button"
            title="删除"
            @click="emit('delete', gen.id)"
          >
            <span class="material-symbols-outlined text-[18px]">delete</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
