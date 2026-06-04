<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { avatarUrl } from '../lib/format'

const auth = useAuthStore()
const router = useRouter()

const credits = computed(() => auth.user?.credits ?? 0)
const avatar = computed(() => avatarUrl(auth.user?.avatar_seed ?? ''))

function generate() {
  router.push('/')
}
</script>

<template>
  <!-- TopNavBar -->
  <header
    class="bg-surface-container/60 backdrop-blur-xl border-b border-outline-variant/20 shadow-sm shadow-primary/5 fixed top-0 w-full md:w-[calc(100%-18rem)] z-50 flex justify-between items-center px-margin-lg h-16 transition-all"
  >
    <!-- Mobile Brand (Hidden on Desktop) -->
    <div class="flex items-center gap-3 md:hidden">
      <h1 class="font-headline-lg-mobile text-headline-lg-mobile text-primary tracking-tight">
        Synthetic Vision
      </h1>
    </div>
    <!-- Empty div for flex spacing on desktop where sidebar handles branding -->
    <div class="hidden md:block"></div>
    <div class="flex items-center gap-4">
      <!-- Credits Pill -->
      <div
        class="glass-panel px-4 py-1.5 rounded-full flex items-center gap-2 border border-outline-variant/40"
      >
        <div class="w-2 h-2 rounded-full bg-secondary animate-pulse"></div>
        <span class="font-label-sm text-label-sm text-on-surface"
          >{{ credits.toLocaleString('en-US') }} 积分</span
        >
      </div>
      <!-- Notifications -->
      <button
        class="w-8 h-8 rounded-full flex items-center justify-center text-on-surface-variant hover:text-primary hover:bg-surface-variant/50 transition-colors active:scale-95"
      >
        <span class="material-symbols-outlined">notifications</span>
      </button>
      <!-- Avatar -->
      <div
        class="w-9 h-9 rounded-full bg-surface-variant overflow-hidden border-2 border-outline-variant/50 hover:border-primary transition-colors cursor-pointer"
      >
        <img alt="用户头像" class="w-full h-full object-cover" :src="avatar" />
      </div>
      <!-- Generate -->
      <button
        class="bg-primary text-on-primary font-bold px-5 py-2 rounded-lg hover:bg-primary/90 transition-colors glow-shadow active:scale-95 hidden sm:block"
        @click="generate"
      >
        生成
      </button>
    </div>
  </header>
</template>
