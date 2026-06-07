<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AuroraCanvas from '@/components/AuroraCanvas.vue'
import { useMagnet, clickSpark } from '@/composables/useInteractions'

const router = useRouter()
const auth = useAuthStore()

const mode = ref<'login' | 'register'>('login')
const username = ref('')
const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

const submitBtn = useMagnet(5)

function setMode(m: 'login' | 'register') {
  mode.value = m
  error.value = ''
}

async function submit(e?: MouseEvent) {
  if (loading.value) return
  if (e) clickSpark(e)
  error.value = ''
  loading.value = true
  try {
    if (mode.value === 'login') {
      await auth.login(email.value, password.value)
    } else {
      await auth.register(username.value, email.value, password.value)
    }
    router.push('/')
  } catch (err: any) {
    error.value =
      err?.response?.data?.error ||
      err?.message ||
      (mode.value === 'login' ? '认证失败。' : '注册失败。')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div
    class="text-on-background min-h-screen grid lg:grid-cols-2 relative overflow-hidden font-body-md"
  >
    <!-- ============ Left: signature aurora hero (signature moment #1 + #2) ============ -->
    <section class="relative hidden lg:flex flex-col justify-between p-12 xl:p-16 overflow-hidden">
      <AuroraCanvas :intensity="1.1" />
      <!-- darken for legibility -->
      <div class="absolute inset-0 bg-gradient-to-br from-void/30 via-void/50 to-void/80"></div>
      <div
        class="absolute inset-0 opacity-[0.5]"
        style="
          background-image: linear-gradient(rgba(56, 232, 255, 0.05) 1px, transparent 1px),
            linear-gradient(90deg, rgba(56, 232, 255, 0.05) 1px, transparent 1px);
          background-size: 40px 40px;
          -webkit-mask-image: radial-gradient(ellipse 80% 80% at 30% 40%, #000, transparent 75%);
          mask-image: radial-gradient(ellipse 80% 80% at 30% 40%, #000, transparent 75%);
        "
      ></div>

      <div class="relative z-10 flex items-center gap-2 font-mono text-micro text-cyan/80 uppercase">
        <span class="w-1.5 h-1.5 rounded-full bg-cyan animate-pulse-dot"></span>
        V3.5 · NEURAL IMAGE ENGINE
      </div>

      <div class="relative z-10">
        <!-- Hero mask reveal + flowing gradient -->
        <h1
          class="reveal-mask font-display text-[64px] xl:text-[80px] font-bold leading-[0.98] tracking-[-0.03em] mb-6"
        >
          <span class="block text-on-surface">SYNTHETIC</span>
          <span class="block text-neon">VISION</span>
        </h1>
        <p class="max-w-md text-on-surface-variant text-[17px] leading-relaxed">
          在黑色虚空里，用青与品红的光，绘制尚未存在的图像。文本到图像的神经合成引擎。
        </p>
      </div>

      <div class="relative z-10 flex gap-8 font-mono text-micro uppercase text-on-surface-variant">
        <div>
          <div class="text-cyan text-[22px] font-display font-bold tracking-normal">4K</div>
          ULTRA RES
        </div>
        <div>
          <div class="text-tertiary text-[22px] font-display font-bold tracking-normal">∞</div>
          STYLES
        </div>
        <div>
          <div class="text-secondary text-[22px] font-display font-bold tracking-normal">1250</div>
          START CR
        </div>
      </div>
    </section>

    <!-- ============ Right: auth panel ============ -->
    <section class="relative flex items-center justify-center p-6 sm:p-10">
      <!-- mobile aurora peek -->
      <div class="lg:hidden absolute inset-0 opacity-60"><AuroraCanvas :intensity="0.7" /></div>
      <div class="lg:hidden absolute inset-0 bg-void/70"></div>

      <main class="relative z-10 w-full max-w-md">
        <div class="grad-border rounded-[20px]">
          <div
            class="glass-panel rounded-[20px] p-8 md:p-10 relative overflow-hidden"
          >
            <!-- Header -->
            <div class="flex flex-col items-center mb-8 text-center">
              <div
                class="relative h-14 w-14 rounded-2xl grad-border bg-surface-container flex items-center justify-center mb-5 glow-shadow"
              >
                <span class="material-symbols-outlined text-primary text-3xl">auto_awesome</span>
                <span
                  class="absolute -top-1 -right-1 w-3 h-3 rounded-full bg-secondary animate-pulse-dot"
                ></span>
              </div>
              <h2 class="font-display text-2xl font-bold text-on-surface tracking-tight mb-2">
                {{ mode === 'login' ? '登录' : '注册' }}
              </h2>
              <p class="text-on-surface-variant text-sm leading-relaxed">
                {{
                  mode === 'login'
                    ? '登录以访问 V3.5 生成引擎。'
                    : '注册后即可开始使用 V3.5 生成引擎。'
                }}
              </p>
            </div>

            <!-- Form -->
            <form class="space-y-5" @submit.prevent="submit()">
              <div v-if="mode === 'register'" class="space-y-2">
                <label
                  class="font-mono text-micro text-on-surface-variant block uppercase"
                  for="username"
                >
                  操作员代号 · CODENAME
                </label>
                <div class="relative">
                  <span
                    class="material-symbols-outlined text-outline-variant text-[18px] absolute inset-y-0 left-3 flex items-center"
                    >badge</span
                  >
                  <input
                    id="username"
                    v-model="username"
                    class="field pl-10"
                    name="username"
                    placeholder="creator_alpha"
                    required
                    type="text"
                  />
                </div>
              </div>

              <div class="space-y-2">
                <label class="font-mono text-micro text-on-surface-variant block uppercase" for="email">
                  邮箱地址 · EMAIL
                </label>
                <div class="relative">
                  <span
                    class="material-symbols-outlined text-outline-variant text-[18px] absolute inset-y-0 left-3 flex items-center"
                    >mail</span
                  >
                  <input
                    id="email"
                    v-model="email"
                    class="field pl-10"
                    name="email"
                    placeholder="operator@synthetic.ai"
                    required
                    type="email"
                  />
                </div>
              </div>

              <div class="space-y-2">
                <div class="flex justify-between items-center">
                  <label
                    class="font-mono text-micro text-on-surface-variant block uppercase"
                    for="password"
                  >
                    访问密钥 · ACCESS KEY
                  </label>
                  <a
                    class="font-mono text-micro text-primary hover:text-primary-fixed transition-colors uppercase"
                    href="#"
                    @click.prevent
                  >
                    忘记密钥？
                  </a>
                </div>
                <div class="relative">
                  <span
                    class="material-symbols-outlined text-outline-variant text-[18px] absolute inset-y-0 left-3 flex items-center"
                    >lock</span
                  >
                  <input
                    id="password"
                    v-model="password"
                    class="field pl-10"
                    name="password"
                    placeholder="••••••••••••"
                    required
                    type="password"
                  />
                </div>
              </div>

              <!-- Error -->
              <div
                v-if="error"
                class="flex items-center gap-2 px-3 py-2.5 rounded-xl bg-error/10 border border-error/30 text-error font-mono text-[12px]"
              >
                <span class="material-symbols-outlined text-sm">error</span>
                <span>{{ error }}</span>
              </div>

              <!-- Actions -->
              <div class="pt-1 flex flex-col gap-3">
                <button
                  ref="submitBtn"
                  class="btn-primary w-full flex justify-center items-center gap-2"
                  type="button"
                  :disabled="loading"
                  @click="submit($event)"
                >
                  <span v-if="loading" class="material-symbols-outlined text-[18px] animate-spin"
                    >progress_activity</span
                  >
                  <span v-else class="material-symbols-outlined text-[18px]">bolt</span>
                  {{ mode === 'login' ? '登录' : '注册' }}
                </button>
                <button
                  class="btn-ghost w-full"
                  type="button"
                  @click="setMode(mode === 'login' ? 'register' : 'login')"
                >
                  {{ mode === 'login' ? '注册' : '登录' }}
                </button>
              </div>
            </form>

            <div
              class="mt-7 pt-5 border-t border-outline-variant/20 flex justify-center items-center gap-2 text-outline-variant font-mono text-micro uppercase"
            >
              <span class="material-symbols-outlined text-sm">shield_locked</span>
              <span>END-TO-END ENCRYPTED LINK</span>
            </div>
          </div>
        </div>
      </main>
    </section>
  </div>
</template>
