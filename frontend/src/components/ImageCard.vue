<script setup lang="ts">
import type { Generation } from '@/types'
import { relativeTime } from '@/lib/format'

const props = defineProps<{ gen: Generation }>()
const emit = defineEmits<{ (e: 'delete', id: string): void }>()

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
    class="group relative aspect-[4/5] rounded-xl overflow-hidden glass-panel glow-hover transition-all duration-300"
  >
    <img
      :alt="gen.prompt"
      :src="gen.image_url"
      class="w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
    />
    <!-- Hover Overlay -->
    <div
      class="absolute inset-0 bg-gradient-to-t from-surface-container-lowest/90 via-surface/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300 flex flex-col justify-end p-4"
    >
      <div class="translate-y-4 group-hover:translate-y-0 transition-transform duration-300">
        <p class="font-label-sm text-label-sm text-primary mb-1">
          {{ gen.width }}x{{ gen.height }} • {{ relativeTime(gen.created_at) }}
        </p>
        <p class="text-sm text-on-surface line-clamp-2 mb-4 leading-snug">
          "{{ gen.prompt }}"
        </p>
        <div class="flex items-center gap-2">
          <button
            class="flex-1 bg-surface-variant/80 hover:bg-primary-container text-on-surface hover:text-on-primary-container py-2 rounded-lg backdrop-blur-sm border border-outline-variant/30 flex justify-center items-center gap-2 transition-colors text-sm font-medium"
            type="button"
            @click="download"
          >
            <span class="material-symbols-outlined text-[18px]">download</span> 下载
          </button>
          <button
            class="w-10 h-10 bg-surface-variant/80 hover:bg-error-container text-on-surface hover:text-on-error-container rounded-lg backdrop-blur-sm border border-outline-variant/30 flex justify-center items-center transition-colors"
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
