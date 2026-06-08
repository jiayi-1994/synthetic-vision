<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useGenerationsStore } from '../stores/generations'
import { avatarUrl } from '../lib/format'
import { useMagnet, clickSpark } from '../composables/useInteractions'

const auth = useAuthStore()
const gen = useGenerationsStore()
const router = useRouter()

const credits = computed(() => auth.user?.credits ?? 0)
const avatar = computed(() => avatarUrl(auth.user?.avatar_seed ?? ''))
const showNotifications = ref(false)

const genBtn = useMagnet(5)

const notifications = computed(() => {
  const recent = [...gen.items]
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 5)

  if (recent.length === 0) {
    return [
      {
        id: 'empty',
        icon: 'notifications',
        title: '暂无任务通知',
        body: '开始生成后，完成、失败和批量进度会显示在这里。',
        tone: 'text-on-surface-variant',
      },
    ]
  }

  return recent.map((item) => {
    if (item.status === 'completed') {
      return {
        id: item.id,
        icon: 'check_circle',
        title: '图像已完成',
        body: `${item.resolution} / ${item.aspect_ratio} · ${item.cost} CR`,
        tone: 'text-success',
      }
    }
    if (item.status === 'failed') {
      return {
        id: item.id,
        icon: 'error',
        title: '生成失败，积分已返还',
        body: item.error || 'Provider 未返回可用图像。',
        tone: 'text-error',
      }
    }
    return {
      id: item.id,
      icon: 'progress_activity',
      title: item.status === 'processing' ? '正在合成' : '任务排队中',
      body: `${item.resolution} / ${item.aspect_ratio}`,
      tone: 'text-primary',
    }
  })
})

const unreadCount = computed(() =>
  gen.items.filter((item) => item.status === 'completed' || item.status === 'failed').slice(0, 9).length
)

function generate(e: MouseEvent) {
  clickSpark(e)
  router.push('/')
}

function toggleNotifications() {
  showNotifications.value = !showNotifications.value
  if (showNotifications.value && gen.items.length === 0) {
    gen.fetchAll({ limit: 5 }).catch(() => {})
  }
}

function openGallery() {
  showNotifications.value = false
  router.push('/gallery')
}

onMounted(() => {
  if (auth.isAuthenticated && gen.items.length === 0) {
    gen.fetchAll({ limit: 5 }).catch(() => {})
  }
})
</script>

<template>
  <!-- TopNavBar -->
  <header
    class="bg-surface-container/55 backdrop-blur-xl border-b border-outline-variant/20 fixed top-0 w-full md:w-[calc(100%-18rem)] z-50 flex justify-between items-center px-margin-sm md:px-margin-lg h-16 transition-all"
  >
    <!-- top neon hairline -->
    <div
      class="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-primary/30 to-transparent"
    ></div>

    <!-- Mobile Brand (Hidden on Desktop) -->
    <div class="flex items-center gap-3 md:hidden">
      <h1 class="font-display text-headline-lg-mobile font-bold text-on-surface tracking-tight">
        Synthetic <span class="text-neon">Vision</span>
      </h1>
    </div>
    <!-- breadcrumb / status on desktop -->
    <div class="hidden md:flex items-center gap-2 font-mono text-micro text-on-surface-variant uppercase">
      <span class="w-1.5 h-1.5 rounded-full bg-success shadow-[0_0_8px_rgba(61,255,176,0.8)]"></span>
      ENGINE ONLINE
    </div>

    <div class="flex items-center gap-3 md:gap-4">
      <!-- Credits Pill -->
      <div
        class="glass-panel px-4 py-1.5 rounded-full flex items-center gap-2 border border-outline-variant/40"
      >
        <div class="w-2 h-2 rounded-full bg-secondary animate-pulse-dot"></div>
        <span class="font-mono text-[12px] tracking-wider text-on-surface"
          >{{ credits.toLocaleString('en-US') }} <span class="text-on-surface-variant">CR</span></span
        >
      </div>
      <!-- Notifications -->
      <div class="relative">
        <button
          type="button"
          class="relative w-11 h-11 rounded-full flex items-center justify-center text-on-surface-variant hover:text-primary hover:bg-primary/8 transition-colors active:scale-95"
          :aria-expanded="showNotifications"
          aria-label="打开通知"
          @click="toggleNotifications"
        >
          <span class="material-symbols-outlined text-[22px]">notifications</span>
          <span
            v-if="unreadCount > 0"
            class="absolute -top-0.5 -right-0.5 min-w-5 h-5 px-1 rounded-full bg-secondary text-on-secondary border border-surface font-mono text-[10px] flex items-center justify-center"
          >
            {{ unreadCount }}
          </span>
        </button>
        <transition
          enter-active-class="transition duration-150 ease-out"
          enter-from-class="opacity-0 translate-y-2 scale-95"
          leave-active-class="transition duration-100 ease-in"
          leave-to-class="opacity-0 translate-y-2 scale-95"
        >
          <div
            v-if="showNotifications"
            class="absolute right-0 top-12 w-[min(88vw,360px)] glass-panel grad-border rounded-2xl p-4 shadow-[0_18px_80px_rgba(0,0,0,0.45)] z-50"
          >
            <div class="flex items-center justify-between gap-3 mb-3">
              <div>
                <p class="font-display font-bold text-on-surface">任务通知</p>
                <p class="font-mono text-[10px] text-on-surface-variant uppercase">RECENT JOB EVENTS</p>
              </div>
              <button
                type="button"
                class="font-mono text-[11px] text-primary hover:text-secondary transition-colors"
                @click="openGallery"
              >
                查看作品库
              </button>
            </div>
            <div class="space-y-2">
              <article
                v-for="item in notifications"
                :key="item.id"
                class="rounded-xl border border-outline-variant/25 bg-surface-container/70 p-3 flex gap-3"
              >
                <span class="material-symbols-outlined text-[20px] mt-0.5" :class="item.tone">{{
                  item.icon
                }}</span>
                <div class="min-w-0">
                  <p class="font-display text-sm font-semibold text-on-surface">{{ item.title }}</p>
                  <p class="font-mono text-[11px] text-on-surface-variant truncate">{{ item.body }}</p>
                </div>
              </article>
            </div>
          </div>
        </transition>
      </div>
      <!-- Avatar -->
      <div
        class="w-9 h-9 rounded-full bg-surface-variant overflow-hidden border border-outline-variant/50 hover:border-primary hover:shadow-[0_0_12px_rgba(56,232,255,0.4)] transition-all cursor-pointer"
      >
        <img alt="用户头像" class="w-full h-full object-cover" :src="avatar" />
      </div>
      <!-- Generate -->
      <button
        ref="genBtn"
        class="hidden sm:flex items-center gap-2 bg-primary-container text-on-primary-container font-display font-bold px-5 py-2 rounded-xl border border-primary/50 glow-shadow hover:brightness-110 transition-all active:scale-95"
        @click="generate"
      >
        <span class="material-symbols-outlined text-[18px]">bolt</span>
        生成
      </button>
    </div>
  </header>
</template>
