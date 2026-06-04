<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { AuthAPI } from '@/api/client'
import { useGenerationsStore } from '@/stores/generations'
import { useAuthStore } from '@/stores/auth'
import { avatarUrl } from '@/lib/format'
import type { Stats } from '@/types'
import ImageCard from '@/components/ImageCard.vue'

const gen = useGenerationsStore()
const auth = useAuthStore()
const { completed } = storeToRefs(gen)

const stats = ref<Stats | null>(null)

const username = computed(() => stats.value?.user.username || auth.user?.username || 'Operator')
const avatarSeed = computed(
  () => stats.value?.user.avatar_seed || auth.user?.avatar_seed || username.value
)
const plan = computed(() => stats.value?.user.plan || auth.user?.plan || 'free')
const planLabel = computed(() =>
  plan.value.toLowerCase() === 'free' ? '免费会员' : '专业版会员'
)
const totalGenerations = computed(() => stats.value?.total_generations ?? 0)
const creditBalance = computed(
  () => stats.value?.credit_balance ?? auth.user?.credits ?? 0
)

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
    <!-- Ambient Background Glow -->
    <div
      class="absolute top-0 left-1/4 w-1/2 h-64 bg-primary/5 blur-[120px] rounded-full pointer-events-none -z-10"
    ></div>
    <div class="max-w-container-max mx-auto space-y-gutter">
      <!-- Profile Header -->
      <section
        class="glass-panel rounded-xl p-6 md:p-8 flex flex-col md:flex-row items-start md:items-center justify-between gap-6"
      >
        <div class="flex items-center gap-6">
          <div class="relative">
            <img
              :src="avatarUrl(avatarSeed)"
              :alt="username"
              class="w-20 h-20 md:w-24 md:h-24 rounded-full border-2 border-primary/30 object-cover bg-surface-variant shadow-[0_0_20px_rgba(124,58,237,0.2)]"
            />
            <div
              class="absolute bottom-0 right-0 w-6 h-6 bg-surface-container rounded-full flex items-center justify-center border border-outline-variant"
            >
              <span class="w-3 h-3 bg-secondary rounded-full"></span>
            </div>
          </div>
          <div>
            <h1
              class="font-headline-lg-mobile md:font-headline-lg md:text-headline-lg text-on-surface mb-1"
            >
              {{ username }}
            </h1>
            <p class="font-label-sm text-label-sm text-on-surface-variant flex items-center gap-2">
              <span class="material-symbols-outlined text-[14px]">verified</span> {{ planLabel }}
            </p>
          </div>
        </div>
        <div
          class="flex gap-4 md:gap-8 w-full md:w-auto overflow-x-auto pb-2 md:pb-0"
        >
          <div
            class="bg-surface-container-highest/50 px-5 py-3 rounded-lg border border-outline-variant/20 flex-shrink-0"
          >
            <p
              class="font-label-sm text-label-sm text-on-surface-variant mb-1 uppercase tracking-wider"
            >
              总生成数
            </p>
            <p
              class="font-display-lg text-display-lg text-primary leading-none text-2xl md:text-[32px]"
            >
              {{ fmt(totalGenerations) }}
            </p>
          </div>
          <div
            class="bg-surface-container-highest/50 px-5 py-3 rounded-lg border border-outline-variant/20 flex-shrink-0"
          >
            <p
              class="font-label-sm text-label-sm text-on-surface-variant mb-1 uppercase tracking-wider"
            >
              积分余额
            </p>
            <div class="flex items-center gap-2">
              <span class="material-symbols-outlined text-secondary text-[20px]">toll</span>
              <p
                class="font-display-lg text-display-lg text-on-surface leading-none text-2xl md:text-[32px]"
              >
                {{ fmt(creditBalance) }}
              </p>
            </div>
          </div>
        </div>
      </section>

      <!-- Gallery Grid -->
      <section>
        <div class="flex items-center justify-between mb-6">
          <h2 class="font-bold text-xl text-on-surface">最近作品</h2>
          <div class="flex gap-2">
            <button
              type="button"
              class="bg-surface-container hover:bg-surface-variant text-on-surface px-3 py-1.5 rounded-lg border border-outline-variant/30 font-label-sm text-label-sm flex items-center gap-2 transition-colors"
            >
              <span class="material-symbols-outlined text-[16px]">filter_list</span> 筛选
            </button>
          </div>
        </div>

        <!-- Empty State -->
        <div
          v-if="completed.length === 0"
          class="glass-panel rounded-xl p-12 flex flex-col items-center justify-center text-center"
        >
          <div
            class="w-20 h-20 rounded-2xl bg-surface-variant/40 border border-outline-variant/30 flex items-center justify-center mb-6"
          >
            <span class="material-symbols-outlined text-primary text-4xl">grid_view</span>
          </div>
          <h3 class="font-headline-lg-mobile text-headline-lg-mobile text-on-surface mb-2">
            暂无作品
          </h3>
          <p class="font-body-md text-body-md text-on-surface-variant max-w-md">
            你完成的合成作品会显示在这里。前往工作台生成你的第一张画面。
          </p>
        </div>

        <!-- Grid -->
        <div
          v-else
          class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 md:gap-6"
        >
          <ImageCard
            v-for="g in completed"
            :key="g.id"
            :gen="g"
            @delete="handleDelete"
          />
        </div>
      </section>
    </div>
  </main>
</template>
