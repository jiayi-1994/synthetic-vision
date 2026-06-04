<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const mode = ref<'login' | 'register'>('login')
const username = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

function setMode(m: 'login' | 'register') {
  mode.value = m
  error.value = ''
}

async function submit() {
  if (loading.value) return
  error.value = ''
  loading.value = true
  try {
    if (mode.value === 'login') {
      await auth.login(email.value, password.value)
    } else {
      await auth.register(username.value, email.value, password.value)
    }
    router.push('/')
  } catch (e: any) {
    error.value =
      e?.response?.data?.error ||
      e?.message ||
      (mode.value === 'login' ? '认证失败。' : '配额申请失败。')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div
    class="bg-background text-on-background min-h-screen flex items-center justify-center relative overflow-hidden font-body-md antialiased"
  >
    <!-- Atmospheric Background Layers -->
    <div class="absolute inset-0 z-0 bg-surface-container-lowest">
      <!-- Radial neural-net glow base -->
      <div
        class="absolute inset-0 opacity-20 mix-blend-screen"
        style="
          background-image: radial-gradient(circle at 30% 30%, rgba(124, 58, 237, 0.35), transparent 45%),
            radial-gradient(circle at 70% 70%, rgba(0, 162, 230, 0.25), transparent 50%);
        "
      ></div>
      <!-- Subtle Glows -->
      <div
        class="absolute top-1/4 left-1/4 w-[500px] h-[500px] bg-primary/10 rounded-full blur-[120px] mix-blend-screen pointer-events-none"
      ></div>
      <div
        class="absolute bottom-1/4 right-1/4 w-[400px] h-[400px] bg-secondary-container/10 rounded-full blur-[100px] mix-blend-screen pointer-events-none"
      ></div>
    </div>

    <!-- Main Auth Canvas -->
    <main class="relative z-10 w-full max-w-md px-margin-sm md:px-0">
      <!-- Glassmorphism Card -->
      <div
        class="bg-surface-container-low/60 backdrop-blur-xl border border-outline-variant/30 rounded-xl shadow-[0_8px_32px_rgba(124,58,237,0.1)] p-8 md:p-10 relative overflow-hidden"
      >
        <!-- Top Inner Glow / Rim Light effect -->
        <div
          class="absolute top-0 left-0 w-full h-[1px] bg-gradient-to-r from-transparent via-primary/30 to-transparent"
        ></div>

        <!-- Header Section -->
        <div class="flex flex-col items-center mb-8 text-center">
          <div
            class="h-12 w-12 rounded-lg bg-surface-variant/50 border border-outline-variant/20 flex items-center justify-center mb-4 shadow-inner relative"
          >
            <span class="material-symbols-outlined text-primary text-3xl">auto_awesome</span>
            <span
              class="absolute top-1 right-1 w-2 h-2 rounded-full bg-secondary-container animate-pulse"
            ></span>
          </div>
          <h1
            class="font-headline-lg-mobile md:font-headline-lg text-headline-lg-mobile md:text-headline-lg text-primary tracking-tight mb-2"
          >
            Synthetic Vision
          </h1>
          <p class="font-body-md text-body-md text-on-surface-variant">
            {{
              mode === 'login'
                ? '认证以访问 V3.5 生成引擎。'
                : '在 V3.5 引擎上申请新的操作员配额。'
            }}
          </p>
        </div>

        <!-- Auth Form -->
        <form class="space-y-6" @submit.prevent="submit">
          <!-- Username Field (Register only) -->
          <div v-if="mode === 'register'" class="space-y-2">
            <label
              class="font-label-sm text-label-sm text-on-surface-variant block uppercase tracking-wider"
              for="username"
            >
              操作员代号
            </label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <span class="material-symbols-outlined text-outline-variant text-sm">badge</span>
              </div>
              <input
                id="username"
                v-model="username"
                class="block w-full pl-10 pr-3 py-3 border border-outline-variant/50 rounded-lg leading-5 bg-surface-variant/50 text-on-surface placeholder:text-outline-variant focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary transition-all duration-200 font-label-sm sm:text-sm"
                name="username"
                placeholder="creator_alpha"
                required
                type="text"
              />
            </div>
          </div>

          <!-- Email Field -->
          <div class="space-y-2">
            <label
              class="font-label-sm text-label-sm text-on-surface-variant block uppercase tracking-wider"
              for="email"
            >
              邮箱地址
            </label>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <span class="material-symbols-outlined text-outline-variant text-sm">mail</span>
              </div>
              <input
                id="email"
                v-model="email"
                class="block w-full pl-10 pr-3 py-3 border border-outline-variant/50 rounded-lg leading-5 bg-surface-variant/50 text-on-surface placeholder:text-outline-variant focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary transition-all duration-200 font-label-sm sm:text-sm"
                name="email"
                placeholder="operator@synthetic.ai"
                required
                type="email"
              />
            </div>
          </div>

          <!-- Password Field -->
          <div class="space-y-2">
            <div class="flex justify-between items-center">
              <label
                class="font-label-sm text-label-sm text-on-surface-variant block uppercase tracking-wider"
                for="password"
              >
                访问密钥
              </label>
              <a
                class="font-label-sm text-label-sm text-primary hover:text-primary-fixed transition-colors"
                href="#"
                @click.prevent
              >
                忘记密钥？
              </a>
            </div>
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <span class="material-symbols-outlined text-outline-variant text-sm">lock</span>
              </div>
              <input
                id="password"
                v-model="password"
                class="block w-full pl-10 pr-3 py-3 border border-outline-variant/50 rounded-lg leading-5 bg-surface-variant/50 text-on-surface placeholder:text-outline-variant focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary transition-all duration-200 font-label-sm sm:text-sm"
                name="password"
                placeholder="••••••••••••"
                required
                type="password"
              />
            </div>
          </div>

          <!-- Error Display -->
          <div
            v-if="error"
            class="flex items-center gap-2 px-3 py-2.5 rounded-lg bg-error-container/20 border border-error/30 text-error font-label-sm text-label-sm"
          >
            <span class="material-symbols-outlined text-sm">error</span>
            <span>{{ error }}</span>
          </div>

          <!-- Actions -->
          <div class="pt-2 flex flex-col gap-4">
            <button
              class="w-full flex justify-center items-center gap-2 py-3 px-4 border border-transparent rounded-lg shadow-sm text-sm font-medium text-on-primary-container bg-primary-container hover:bg-inverse-primary focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary focus:ring-offset-background transition-all duration-200 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed"
              type="submit"
              :disabled="loading"
            >
              <span
                v-if="loading"
                class="material-symbols-outlined text-[18px] animate-spin"
                >progress_activity</span
              >
              {{ mode === 'login' ? '初始化会话' : '申请新配额' }}
            </button>
            <button
              class="w-full flex justify-center py-3 px-4 border border-outline-variant rounded-lg shadow-sm text-sm font-medium text-on-surface bg-surface-container/30 hover:bg-surface-variant/50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-outline focus:ring-offset-background transition-all duration-200 active:scale-[0.98]"
              type="button"
              @click="setMode(mode === 'login' ? 'register' : 'login')"
            >
              {{ mode === 'login' ? '申请新配额' : '初始化会话' }}
            </button>
          </div>
        </form>

        <!-- Decorative Bottom Element -->
        <div
          class="mt-8 pt-6 border-t border-outline-variant/20 flex justify-center items-center gap-2 text-outline-variant text-xs font-label-sm"
        >
          <span class="material-symbols-outlined text-sm">shield_locked</span>
          <span>端到端加密链路</span>
        </div>
      </div>
    </main>
  </div>
</template>
