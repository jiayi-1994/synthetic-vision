<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const userId = computed(() => auth.user?.public_id || 'USR-UNKNOWN')

const runbook = [
  {
    title: '生成排队或长时间处理中',
    icon: 'hourglass_top',
    body: '后台 worker 会自动重扫 pending 任务。刷新 Gallery 或 Analytics 可以查看最新状态；失败任务会自动退回积分。',
  },
  {
    title: '图生图 / 局部修图不可提交',
    icon: 'image_search',
    body: '图生图需要 PNG/JPEG/WebP 参考图；局部修图还需要先在画布上涂抹 mask 区域。单个上传文件上限 10MB。',
  },
  {
    title: '积分不足',
    icon: 'toll',
    body: '1K/2K/4K 分别消耗 5/15/40 CR。管理员可在 Admin 页面通过 public_id 为用户充值。',
  },
]

const contactCards = computed(() => [
  { label: '当前用户 ID', value: userId.value, icon: 'badge' },
  { label: '服务入口', value: 'Go API + Vue SPA', icon: 'dns' },
  { label: '默认 Provider', value: 'Mock / OpenAI-compatible', icon: 'hub' },
])

function goDashboard() {
  router.push('/')
}

function goAnalytics() {
  router.push('/analytics')
}
</script>

<template>
  <main class="flex-1 p-margin-sm md:p-margin-lg overflow-y-auto relative">
    <div class="max-w-container-max mx-auto space-y-gutter">
      <section
        v-reveal
        class="glass-panel rounded-2xl p-6 md:p-8 flex flex-col xl:flex-row xl:items-end justify-between gap-6 relative overflow-hidden"
      >
        <div class="pointer-events-none absolute -left-20 -bottom-24 h-64 w-64 rounded-full bg-secondary/10 blur-3xl"></div>
        <div class="relative z-10">
          <p class="font-mono text-micro text-on-surface-variant mb-3 uppercase">帮助中心 · SUPPORT</p>
          <h1 class="font-display text-headline-lg-mobile md:text-headline-lg font-bold text-on-surface mb-2">
            操作<span class="text-neon">排障</span>
          </h1>
          <p class="font-body-md text-body-md text-on-surface-variant max-w-2xl">
            这里把最常见的生成、修图、积分问题做成可执行提示，避免侧边栏“帮助”只是空链接。
          </p>
        </div>
        <div class="relative z-10 flex flex-wrap gap-3">
          <button type="button" class="btn-primary flex items-center gap-2" @click="goDashboard">
            <span class="material-symbols-outlined text-[18px]">bolt</span>
            去生成
          </button>
          <button type="button" class="btn-ghost flex items-center gap-2" @click="goAnalytics">
            <span class="material-symbols-outlined text-[18px]">insights</span>
            看状态
          </button>
        </div>
      </section>

      <section class="grid grid-cols-1 xl:grid-cols-12 gap-gutter items-start">
        <div class="xl:col-span-8 space-y-6">
          <article
            v-for="(item, i) in runbook"
            :key="item.title"
            v-reveal="i * 70"
            class="glass-panel glow-hover rounded-2xl p-6 flex gap-4"
          >
            <div
              class="w-12 h-12 rounded-xl bg-primary/10 border border-primary/30 flex items-center justify-center flex-shrink-0"
            >
              <span class="material-symbols-outlined text-primary">{{ item.icon }}</span>
            </div>
            <div>
              <h2 class="font-display text-title font-bold text-on-surface mb-2">{{ item.title }}</h2>
              <p class="text-on-surface-variant leading-relaxed">{{ item.body }}</p>
            </div>
          </article>
        </div>

        <aside class="xl:col-span-4 space-y-6">
          <article v-reveal class="grad-border glass-panel rounded-2xl p-6 space-y-4">
            <h2 class="font-display text-title font-bold text-on-surface flex items-center gap-2">
              <span class="material-symbols-outlined text-secondary text-[20px]">support_agent</span>
              支持信息
            </h2>
            <div class="space-y-3">
              <div
                v-for="card in contactCards"
                :key="card.label"
                class="rounded-xl border border-outline-variant/30 bg-surface-container/40 p-4"
              >
                <div class="flex items-center gap-2 font-mono text-[10px] text-on-surface-variant uppercase">
                  <span class="material-symbols-outlined text-[16px] text-primary">{{ card.icon }}</span>
                  {{ card.label }}
                </div>
                <p class="mt-2 font-display text-on-surface font-bold break-all">{{ card.value }}</p>
              </div>
            </div>
          </article>

          <article v-reveal="100" class="glass-panel rounded-2xl p-6">
            <h2 class="font-display text-title font-bold text-on-surface mb-3">排障顺序</h2>
            <ol class="space-y-3 text-sm text-on-surface-variant">
              <li class="flex gap-3">
                <span class="font-mono text-primary">01</span>
                先在 Analytics 看最近任务是否 failed。
              </li>
              <li class="flex gap-3">
                <span class="font-mono text-primary">02</span>
                Gallery 确认完成图是否已经落盘。
              </li>
              <li class="flex gap-3">
                <span class="font-mono text-primary">03</span>
                若是 Provider 错误，失败行会保留错误文本并返还积分。
              </li>
            </ol>
          </article>
        </aside>
      </section>
    </div>
  </main>
</template>
