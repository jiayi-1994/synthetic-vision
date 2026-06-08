<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { AuthAPI } from '@/api/client'
import { useGenerationsStore } from '@/stores/generations'
import { useAuthStore } from '@/stores/auth'
import { avatarUrl } from '@/lib/format'
import { loadWorkspacePreferences } from '@/lib/preferences'
import { useSpotlight } from '@/composables/useInteractions'
import type { AspectRatio, GenerationMode, Resolution, Stats } from '@/types'
import ImageCard from '@/components/ImageCard.vue'

const gen = useGenerationsStore()
const auth = useAuthStore()
const { completed } = storeToRefs(gen)

const stats = ref<Stats | null>(null)
const profileCard = useSpotlight()
const modeFilter = ref<GenerationMode | 'all'>('all')
const resolutionFilter = ref<Resolution | 'all'>('all')
const aspectFilter = ref<AspectRatio | 'all'>('all')
const search = ref('')
const compactGallery = ref(false)

const username = computed(() => stats.value?.user.username || auth.user?.username || 'Operator')
const avatarSeed = computed(
  () => stats.value?.user.avatar_seed || auth.user?.avatar_seed || username.value
)
const plan = computed(() => stats.value?.user.plan || auth.user?.plan || 'free')
const planLabel = computed(() =>
  plan.value.toLowerCase() === 'free' ? '免费会员' : '专业版会员'
)
const totalGenerations = computed(() => stats.value?.total_generations ?? 0)
const creditBalance = computed(() => stats.value?.credit_balance ?? auth.user?.credits ?? 0)
const filteredCompleted = computed(() => {
  const q = search.value.trim().toLowerCase()
  return completed.value.filter((item) => {
    if (modeFilter.value !== 'all' && item.mode !== modeFilter.value) return false
    if (resolutionFilter.value !== 'all' && item.resolution !== resolutionFilter.value) return false
    if (aspectFilter.value !== 'all' && item.aspect_ratio !== aspectFilter.value) return false
    if (q && !item.prompt.toLowerCase().includes(q)) return false
    return true
  })
})
const hasActiveFilters = computed(
  () =>
    modeFilter.value !== 'all' ||
    resolutionFilter.value !== 'all' ||
    aspectFilter.value !== 'all' ||
    search.value.trim() !== ''
)

// Bento: feature the newest image as a 2×2 hero tile when the set is large enough.
function cellClass(index: number): string {
  if (compactGallery.value) {
    return '!aspect-auto h-full'
  }
  if (filteredCompleted.value.length >= 5 && index === 0) {
    return 'sm:col-span-2 sm:row-span-2 !aspect-auto h-full'
  }
  return '!aspect-auto h-full'
}

function fmt(n: number): string {
  return n.toLocaleString('en-US')
}

async function loadStats() {
  try {
    stats.value = await AuthAPI.stats()
  } catch {
    /* fall back to auth store values */
  }
}

async function handleDelete(id: string) {
  await gen.remove(id)
  await loadStats()
}

function modeLabel(mode: GenerationMode): string {
  switch (mode) {
    case 'image':
      return '图生图'
    case 'edit':
      return '局部修图'
    default:
      return '文生图'
  }
}

function clearFilters() {
  modeFilter.value = 'all'
  resolutionFilter.value = 'all'
  aspectFilter.value = 'all'
  search.value = ''
}

onMounted(async () => {
  compactGallery.value = loadWorkspacePreferences().compactGallery
  await Promise.all([loadStats(), gen.fetchAll()])
})
</script>

<template>
  <main class="flex-1 p-margin-sm md:p-margin-lg min-h-full relative overflow-y-auto">
    <div class="max-w-container-max mx-auto space-y-gutter">
      <!-- Profile Header -->
      <section
        ref="profileCard"
        v-reveal
        class="spotlight glass-panel rounded-2xl p-6 md:p-8 flex flex-col md:flex-row items-start md:items-center justify-between gap-6 relative overflow-hidden"
      >
        <div class="flex items-center gap-6 relative z-10">
          <div class="relative">
            <div class="absolute -inset-1 rounded-full bg-gradient-to-tr from-primary via-tertiary to-secondary opacity-60 blur-[6px]"></div>
            <img
              :src="avatarUrl(avatarSeed)"
              :alt="username"
              class="relative w-20 h-20 md:w-24 md:h-24 rounded-full border border-primary/40 object-cover bg-surface-variant"
            />
            <div
              class="absolute bottom-0 right-0 w-6 h-6 bg-surface-container rounded-full flex items-center justify-center border border-outline-variant"
            >
              <span class="w-3 h-3 bg-success rounded-full shadow-[0_0_8px_rgba(61,255,176,0.8)]"></span>
            </div>
          </div>
          <div>
            <h1 class="font-display text-2xl md:text-headline-lg font-bold text-on-surface mb-1.5">
              {{ username }}
            </h1>
            <p class="font-mono text-micro text-on-surface-variant flex items-center gap-2 uppercase">
              <span class="material-symbols-outlined text-[14px] text-primary">verified</span>
              {{ planLabel }}
            </p>
          </div>
        </div>
        <div class="flex gap-4 md:gap-6 w-full md:w-auto overflow-x-auto pb-2 md:pb-0 relative z-10">
          <div class="grad-border bg-surface-container px-5 py-3 rounded-xl flex-shrink-0 min-w-[130px]">
            <p class="font-mono text-micro text-on-surface-variant mb-2 uppercase">总生成数 · TOTAL</p>
            <p class="font-display text-2xl md:text-[32px] font-bold text-neon leading-none">
              {{ fmt(totalGenerations) }}
            </p>
          </div>
          <div class="grad-border bg-surface-container px-5 py-3 rounded-xl flex-shrink-0 min-w-[130px]">
            <p class="font-mono text-micro text-on-surface-variant mb-2 uppercase">积分余额 · CREDITS</p>
            <div class="flex items-center gap-2">
              <span class="material-symbols-outlined text-secondary text-[20px]">toll</span>
              <p class="font-display text-2xl md:text-[32px] font-bold text-on-surface leading-none">
                {{ fmt(creditBalance) }}
              </p>
            </div>
          </div>
        </div>
      </section>

      <!-- Gallery Grid -->
      <section>
        <div class="flex flex-col lg:flex-row lg:items-end justify-between gap-4 mb-6" v-reveal>
          <div>
            <h2 class="font-display text-headline-lg-mobile md:text-headline-lg font-bold text-on-surface">
              最近作品
            </h2>
            <p class="font-mono text-micro text-on-surface-variant uppercase mt-1">RECENT OUTPUT</p>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <div class="relative">
              <span
                class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-outline text-[18px]"
                >search</span
              >
              <input
                v-model="search"
                class="field pl-10 !py-2.5 font-mono text-[12px] w-56"
                placeholder="搜索 prompt"
                type="search"
              />
            </div>
            <select v-model="modeFilter" class="field !py-2.5 font-mono text-[12px] w-32">
              <option value="all">全部模式</option>
              <option value="text">文生图</option>
              <option value="image">图生图</option>
              <option value="edit">局部修图</option>
            </select>
            <select v-model="resolutionFilter" class="field !py-2.5 font-mono text-[12px] w-28">
              <option value="all">全部分辨率</option>
              <option value="1K">1K</option>
              <option value="2K">2K</option>
              <option value="4K">4K</option>
            </select>
            <select v-model="aspectFilter" class="field !py-2.5 font-mono text-[12px] w-28">
              <option value="all">全部比例</option>
              <option value="1:1">1:1</option>
              <option value="4:3">4:3</option>
              <option value="16:9">16:9</option>
              <option value="9:16">9:16</option>
            </select>
            <button
              v-if="hasActiveFilters"
              type="button"
              class="btn-ghost flex items-center gap-2 text-sm !min-h-[42px] !py-2 !px-3"
              @click="clearFilters"
            >
              <span class="material-symbols-outlined text-[16px]">filter_alt_off</span> 清除
            </button>
          </div>
        </div>

        <div
          v-if="hasActiveFilters"
          v-reveal
          class="mb-4 flex flex-wrap items-center gap-2 font-mono text-[11px] text-on-surface-variant"
        >
          <span class="uppercase">FILTERED</span>
          <span class="px-2 py-1 rounded-full bg-primary/10 text-primary border border-primary/25">
            {{ filteredCompleted.length }} / {{ completed.length }} 张
          </span>
          <span v-if="modeFilter !== 'all'" class="px-2 py-1 rounded-full bg-surface-container border border-outline-variant/30">
            {{ modeLabel(modeFilter as GenerationMode) }}
          </span>
          <span v-if="resolutionFilter !== 'all'" class="px-2 py-1 rounded-full bg-surface-container border border-outline-variant/30">
            {{ resolutionFilter }}
          </span>
          <span v-if="aspectFilter !== 'all'" class="px-2 py-1 rounded-full bg-surface-container border border-outline-variant/30">
            {{ aspectFilter }}
          </span>
        </div>

        <!-- Empty State -->
        <div
          v-if="completed.length === 0"
          v-reveal
          class="glass-panel rounded-2xl p-12 flex flex-col items-center justify-center text-center"
        >
          <div
            class="w-20 h-20 rounded-2xl grad-border bg-surface-container flex items-center justify-center mb-6 glow-shadow"
          >
            <span class="material-symbols-outlined text-primary text-4xl">grid_view</span>
          </div>
          <h3 class="font-display text-xl font-bold text-on-surface mb-2">暂无作品</h3>
          <p class="font-body-md text-body-md text-on-surface-variant max-w-md">
            你完成的合成作品会显示在这里。前往工作台生成你的第一张画面。
          </p>
        </div>

        <div
          v-else-if="filteredCompleted.length === 0"
          v-reveal
          class="glass-panel rounded-2xl p-12 flex flex-col items-center justify-center text-center"
        >
          <div
            class="w-20 h-20 rounded-2xl grad-border bg-surface-container flex items-center justify-center mb-6 glow-shadow"
          >
            <span class="material-symbols-outlined text-primary text-4xl">filter_alt_off</span>
          </div>
          <h3 class="font-display text-xl font-bold text-on-surface mb-2">没有匹配作品</h3>
          <p class="font-body-md text-body-md text-on-surface-variant max-w-md mb-6">
            当前筛选条件下没有已完成作品。清除筛选或更换关键词后再试。
          </p>
          <button type="button" class="btn-primary" @click="clearFilters">清除筛选</button>
        </div>

        <!-- Bento Grid -->
        <div
          v-else
          v-reveal
          class="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3 md:gap-5"
          :class="compactGallery ? 'auto-rows-[150px] sm:auto-rows-[180px]' : 'auto-rows-[180px] sm:auto-rows-[220px]'"
        >
          <ImageCard
            v-for="(g, i) in filteredCompleted"
            :key="g.id"
            :gen="g"
            :class="cellClass(i)"
            @delete="handleDelete"
          />
        </div>
      </section>
    </div>
  </main>
</template>
