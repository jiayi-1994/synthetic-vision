<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { AnalyticsAPI } from '@/api/client'
import { relativeTime } from '@/lib/format'
import type { AnalyticsResponse, AnalyticsDistributionItem, GenStatus } from '@/types'

interface MetricCard {
  label: string
  value: string
  helper: string
  icon: string
}

const router = useRouter()
const loading = ref(true)
const error = ref('')
const data = ref<AnalyticsResponse | null>(null)

const summaryCards = computed<MetricCard[]>(() => {
  if (!data.value) return []
  const s = data.value.summary
  return [
    { label: '总生成数', value: String(s.total_generations), helper: '按创建时间汇总', icon: 'auto_awesome' },
    { label: '成功率', value: `${s.success_rate.toFixed(1)}%`, helper: '完成 / 总任务', icon: 'verified' },
    { label: '消耗积分', value: `${s.credits_spent} 积分`, helper: `退款 ${s.credits_refunded} 积分`, icon: 'flash_on' },
    { label: '当前积分', value: `${s.credit_balance}`, helper: '账户余额', icon: 'toll' },
  ]
})

const statusItems = computed(() => data.value?.status_distribution ?? [])
const resolutionItems = computed(() => data.value?.resolution_distribution ?? [])
const aspectItems = computed(() => data.value?.aspect_ratio_distribution ?? [])
const recent = computed(() => data.value?.recent_generations ?? [])

const hasRecent = computed(() => recent.value.length > 0)
const hasData = computed(() => (data.value?.summary.total_generations ?? 0) > 0)

function formatStatus(status: GenStatus) {
  switch (status) {
    case 'pending':
      return '排队中'
    case 'processing':
      return '生成中'
    case 'completed':
      return '完成'
    case 'failed':
      return '失败'
    default:
      return status
  }
}

function statusBadgeClass(status: GenStatus) {
  switch (status) {
    case 'completed':
      return 'bg-primary/15 text-primary'
    case 'processing':
      return 'bg-secondary/15 text-secondary'
    case 'pending':
      return 'bg-outline text-on-surface-variant border border-outline-variant/40'
    case 'failed':
      return 'bg-error/15 text-error'
    default:
      return 'bg-surface-variant text-on-surface-variant'
  }
}

function percentBarWidth(item: AnalyticsDistributionItem) {
  const normalized = Number.isFinite(item.percentage) ? item.percentage : 0
  return `${Math.max(0, Math.min(100, normalized))}%`
}

async function loadAnalytics() {
  loading.value = true
  error.value = ''
  try {
    data.value = await AnalyticsAPI.meAnalytics()
  } catch (e: any) {
    error.value = e?.response?.data?.error || '无法读取数据，请稍后重试。'
  } finally {
    loading.value = false
  }
}

onMounted(loadAnalytics)
</script>

<template>
  <main class="flex-1 p-margin-sm md:p-margin-lg overflow-y-auto relative">
    <!-- Ambient Background Glow -->
    <div class="absolute top-0 left-1/4 w-1/2 h-64 bg-primary/5 blur-[120px] rounded-full pointer-events-none -z-10"></div>
    <div class="max-w-container-max mx-auto space-y-gutter">
      <!-- Header -->
      <section class="glass-panel rounded-xl p-6 md:p-8 flex flex-wrap gap-4 items-center justify-between">
        <div>
          <p class="font-label-sm text-label-sm text-on-surface-variant mb-1 uppercase tracking-wider">个人洞察</p>
          <h1 class="font-headline-lg md:font-headline-lg text-on-surface mb-1">分析看板</h1>
          <p class="font-body-md text-body-md text-on-surface-variant">
            基于你自己的生成历史，展示积分使用、状态分布和近期活动。
          </p>
        </div>
        <button
          type="button"
          class="bg-primary text-on-primary font-bold px-5 py-2 rounded-lg hover:bg-primary/90 transition-colors glow-shadow active:scale-95"
          @click="router.push('/')"
        >
          去生成
        </button>
      </section>

      <div v-if="loading" class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4 md:gap-6">
        <div
          v-for="i in 4"
          :key="i"
          class="glass-panel rounded-xl p-6 border border-outline-variant/30 animate-pulse bg-surface-container/30 h-28"
        ></div>
      </div>

      <div v-else-if="error" class="glass-panel rounded-xl p-8 border border-error/30 text-error">
        <div class="flex items-start gap-3">
          <span class="material-symbols-outlined text-2xl text-error">error</span>
          <div>
            <h3 class="font-semibold text-on-surface mb-1">数据加载失败</h3>
            <p class="font-body-md text-body-md text-on-surface-variant">{{ error }}</p>
          </div>
        </div>
      </div>

      <template v-else>
        <!-- Empty state -->
        <section
          v-if="!hasData"
          class="glass-panel rounded-xl p-12 flex flex-col items-center justify-center text-center"
        >
          <div class="w-20 h-20 rounded-2xl bg-surface-variant/40 border border-outline-variant/30 flex items-center justify-center mb-6">
            <span class="material-symbols-outlined text-primary text-4xl">insights</span>
          </div>
          <h3 class="font-headline-lg-mobile text-headline-lg-mobile text-on-surface mb-2">暂无历史数据</h3>
          <p class="font-body-md text-body-md text-on-surface-variant max-w-md mb-6">
            你还没有生成记录，先去试一次合成，解析面板会自动出现真实统计。
          </p>
          <button
            class="bg-primary text-on-primary px-5 py-2.5 rounded-lg font-bold hover:bg-primary/90 transition-colors"
            @click="router.push('/')"
          >
            立即创建首张作品
          </button>
        </section>

        <template v-else>
          <!-- Summary cards -->
          <section class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4 md:gap-6">
            <article
              v-for="card in summaryCards"
              :key="card.label"
              class="glass-panel rounded-xl p-6 border border-outline-variant/20"
            >
              <div class="flex items-start justify-between gap-4">
                <div>
                  <p class="font-label-sm text-label-sm text-on-surface-variant">{{ card.label }}</p>
                  <p class="font-display-lg md:text-[32px] text-on-surface mt-2">{{ card.value }}</p>
                </div>
                <span
                  class="w-10 h-10 rounded-lg bg-surface-container flex items-center justify-center border border-outline-variant/30"
                >
                  <span class="material-symbols-outlined text-on-surface-variant">{{ card.icon }}</span>
                </span>
              </div>
              <p class="font-label-sm text-label-sm text-on-surface-variant mt-4">{{ card.helper }}</p>
            </article>
          </section>

          <!-- Visual breakdowns -->
          <section class="grid grid-cols-1 xl:grid-cols-2 gap-4 md:gap-6">
            <article class="glass-panel rounded-xl p-6">
              <h2 class="font-bold text-on-surface text-lg mb-4">状态与耗时</h2>
              <div class="space-y-4">
                <div
                  v-for="item in statusItems"
                  :key="`status-${item.label}`"
                  class="space-y-2"
                >
                  <div class="flex items-center justify-between text-sm">
                    <span class="font-label-sm text-label-sm text-on-surface-variant uppercase">{{ item.label }}</span>
                    <span class="font-label-sm text-label-sm text-on-surface">{{ item.count }}</span>
                  </div>
                  <div class="w-full h-2 bg-surface-variant rounded-full overflow-hidden">
                    <div
                      class="h-full bg-primary rounded-full transition-all"
                      :style="{ width: percentBarWidth(item) }"
                    ></div>
                  </div>
                </div>
              </div>
            </article>

            <article class="glass-panel rounded-xl p-6 space-y-6">
              <div>
                <h2 class="font-bold text-on-surface text-lg mb-4">分辨率分布</h2>
                <div class="space-y-3">
                  <div
                    v-for="item in resolutionItems"
                    :key="`res-${item.label}`"
                    class="flex items-center justify-between gap-3"
                  >
                    <span class="font-label-sm text-label-sm text-on-surface-variant">{{ item.label }}</span>
                    <span class="font-label-sm text-label-sm text-on-surface">{{ item.count }} 张</span>
                    <span class="font-label-sm text-label-sm text-primary">{{ item.percentage.toFixed(1) }}%</span>
                  </div>
                </div>
              </div>
              <div>
                <h2 class="font-bold text-on-surface text-lg mb-4">宽高比分布</h2>
                <div class="space-y-3">
                  <div
                    v-for="item in aspectItems"
                    :key="`aspect-${item.label}`"
                    class="flex items-center justify-between gap-3"
                  >
                    <span class="font-label-sm text-label-sm text-on-surface-variant">{{ item.label }}</span>
                    <span class="font-label-sm text-label-sm text-on-surface">{{ item.count }} 张</span>
                    <span class="font-label-sm text-label-sm text-primary">{{ item.percentage.toFixed(1) }}%</span>
                  </div>
                </div>
              </div>
            </article>
          </section>

          <!-- Recent activity -->
          <section class="glass-panel rounded-xl p-6 space-y-4">
            <h2 class="font-bold text-on-surface text-lg">近期活动</h2>
            <div v-if="!hasRecent" class="text-body-md text-on-surface-variant">
              近期暂无活动。前往“工作台”生成后会展示最新记录。
            </div>
            <div v-else class="space-y-3">
              <article
                v-for="gen in recent"
                :key="gen.id"
                class="p-4 rounded-lg border border-outline-variant/30 bg-surface-container/40 flex flex-col md:flex-row md:items-start md:justify-between gap-3"
              >
                <div class="space-y-1 flex-1">
                  <div class="flex items-center gap-2 flex-wrap text-label-sm text-on-surface-variant">
                    <span class="font-label-sm">{{ gen.created_at ? relativeTime(gen.created_at) : '时间未知' }}</span>
                    <span class="w-1.5 h-1.5 bg-outline rounded-full"></span>
                    <span class="font-label-sm">{{ gen.resolution }} / {{ gen.aspect_ratio }}</span>
                    <span class="w-1.5 h-1.5 bg-outline rounded-full"></span>
                    <span class="font-label-sm">{{ gen.cost }} 积分</span>
                  </div>
                  <p class="text-sm text-on-surface line-clamp-2">{{ gen.prompt }}</p>
                  <p v-if="gen.status === 'failed'" class="font-label-sm text-error text-sm">
                    {{ gen.error || '生成失败，已返还积分。' }}
                  </p>
                </div>
                <div class="flex items-center gap-2">
                  <span
                    class="px-3 py-1 rounded-full font-label-sm text-label-sm border"
                    :class="statusBadgeClass(gen.status as GenStatus)"
                  >
                    {{ formatStatus(gen.status as GenStatus) }}
                  </span>
                </div>
              </article>
            </div>
          </section>
        </template>
      </template>
    </div>
  </main>
</template>

