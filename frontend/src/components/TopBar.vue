<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { avatarUrl } from '../lib/format'
import { useMagnet, clickSpark } from '../composables/useInteractions'

const auth = useAuthStore()
const router = useRouter()

const credits = computed(() => auth.user?.credits ?? 0)
const avatar = computed(() => avatarUrl(auth.user?.avatar_seed ?? ''))

const genBtn = useMagnet(5)

function generate(e: MouseEvent) {
  clickSpark(e)
  router.push('/')
}
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
      <button
        class="w-11 h-11 rounded-full flex items-center justify-center text-on-surface-variant hover:text-primary hover:bg-primary/8 transition-colors active:scale-95"
      >
        <span class="material-symbols-outlined text-[22px]">notifications</span>
      </button>
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
