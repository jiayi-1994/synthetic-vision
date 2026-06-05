<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { AuthAPI } from '@/api/client'
import { useGenerationsStore } from '@/stores/generations'
import { useAuthStore } from '@/stores/auth'
import { avatarUrl } from '@/lib/format'
import { useSpotlight } from '@/composables/useInteractions'
import type { Stats } from '@/types'
import ImageCard from '@/components/ImageCard.vue'

const gen = useGenerationsStore()
const auth = useAuthStore()
const { completed } = storeToRefs(gen)

const stats = ref<Stats | null>(null)
const profileCard = useSpotlight()

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

// Bento: feature the newest image as a 2×2 hero tile when the set is large enough.
function cellClass(index: number): string {
  if (completed.value.length >= 5 && index === 0) {
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

onMounted(async () => {
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
        <div class="flex items-center justify-between mb-6" v-reveal>
          <div>
            <h2 class="font-display text-headline-lg-mobile md:text-headline-lg font-bold text-on-surface">
              最近作品
            </h2>
            <p class="font-mono text-micro text-on-surface-variant uppercase mt-1">RECENT OUTPUT</p>
          </div>
          <button
            type="button"
            class="btn-ghost flex items-center gap-2 text-sm min-h-[44px] px-4"
          >
            <span class="material-symbols-outlined text-[16px]">filter_list</span> 筛选
          </button>
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

        <!-- Bento Grid -->
        <div
          v-else
          v-reveal
          class="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 auto-rows-[180px] sm:auto-rows-[220px] gap-3 md:gap-5"
        >
          <ImageCard
            v-for="(g, i) in completed"
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
