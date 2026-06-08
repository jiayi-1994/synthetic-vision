<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ASPECTS, RESOLUTIONS } from '@/lib/format'
import {
  defaultWorkspacePreferences,
  loadWorkspacePreferences,
  saveWorkspacePreferences,
  type WorkspacePreferences,
} from '@/lib/preferences'
import type { GenerationMode } from '@/types'

const router = useRouter()
const route = useRoute()

const modes: { id: GenerationMode; label: string; helper: string; icon: string }[] = [
  { id: 'text', label: '文生图', helper: '默认从提示词创建新图。', icon: 'auto_awesome' },
  { id: 'image', label: '图生图', helper: '默认提示你上传参考图。', icon: 'add_photo_alternate' },
  { id: 'edit', label: '局部修图', helper: '默认进入涂抹修图工作流。', icon: 'brush' },
]

const styleOptions = ['Cinematic', 'Editorial', 'Product', 'Anime', 'Photoreal', 'Cyber Neon']
const focusBilling = computed(() => route.query.section === 'billing')

const planCards = [
  {
    name: 'Free',
    price: '0 CR / 月',
    tone: 'text-on-surface',
    features: ['注册赠送 1,250 CR', '文生图 / 图生图 / 局部修图', '本地 Mock Provider 可完整演示'],
  },
  {
    name: 'Pro',
    price: '管理员充值',
    tone: 'text-primary',
    features: ['适合真实 OpenAI-compatible Provider', '批量生成最高 8 张', '通过 Admin public_id 注入积分'],
  },
]

const prefs = reactive<WorkspacePreferences>({ ...defaultWorkspacePreferences })
const saved = ref(false)
let savedTimer: ReturnType<typeof setTimeout> | undefined

function apply(next: WorkspacePreferences) {
  prefs.defaultMode = next.defaultMode
  prefs.defaultResolution = next.defaultResolution
  prefs.defaultAspectRatio = next.defaultAspectRatio
  prefs.defaultStyle = next.defaultStyle
  prefs.compactGallery = next.compactGallery
  prefs.emailDigest = next.emailDigest
}

function save() {
  saveWorkspacePreferences({ ...prefs })
  saved.value = true
  if (savedTimer) clearTimeout(savedTimer)
  savedTimer = setTimeout(() => (saved.value = false), 2600)
}

function reset() {
  apply({ ...defaultWorkspacePreferences })
  save()
}

function goDashboard() {
  router.push('/')
}

onMounted(() => {
  apply(loadWorkspacePreferences())
})
</script>

<template>
  <main class="flex-1 p-margin-sm md:p-margin-lg overflow-y-auto relative">
    <div class="max-w-container-max mx-auto space-y-gutter">
      <section v-reveal class="glass-panel rounded-2xl p-6 md:p-8 flex flex-col gap-5 relative overflow-hidden">
        <div
          class="pointer-events-none absolute -right-20 -top-24 h-64 w-64 rounded-full bg-primary/10 blur-3xl"
        ></div>
        <div class="relative z-10 flex flex-col lg:flex-row lg:items-center justify-between gap-5">
          <div>
            <p class="font-mono text-micro text-on-surface-variant mb-3 uppercase">偏好设置 · SETTINGS</p>
            <h1 class="font-display text-headline-lg-mobile md:text-headline-lg font-bold text-on-surface mb-2">
              工作台<span class="text-neon">默认值</span>
            </h1>
            <p class="font-body-md text-body-md text-on-surface-variant max-w-2xl">
              这里的配置保存在当前浏览器，用于工作台首次打开时的模式、分辨率和宽高比。不会影响服务器数据。
            </p>
          </div>
          <button type="button" class="btn-primary flex items-center gap-2 self-start lg:self-auto" @click="goDashboard">
            <span class="material-symbols-outlined text-[18px]">bolt</span>
            打开工作台
          </button>
        </div>
      </section>

      <section class="grid grid-cols-1 xl:grid-cols-12 gap-gutter items-start">
        <article v-reveal class="xl:col-span-8 glass-panel rounded-2xl p-6 space-y-6">
          <div>
            <h2 class="font-display text-title font-bold text-on-surface flex items-center gap-2">
              <span class="material-symbols-outlined text-primary text-[20px]">tune</span>
              默认生成参数
            </h2>
            <p class="font-mono text-[11px] text-on-surface-variant mt-1 uppercase">
              APPLIED ON DASHBOARD START
            </p>
          </div>

          <div class="space-y-3">
            <label class="font-mono text-micro text-on-surface-variant uppercase">默认模式</label>
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <button
                v-for="m in modes"
                :key="m.id"
                type="button"
                class="text-left rounded-xl border p-4 transition-all"
                :class="
                  prefs.defaultMode === m.id
                    ? 'border-primary/60 bg-primary/10 text-primary shadow-[0_0_22px_rgba(56,232,255,0.14)]'
                    : 'border-outline-variant/35 bg-surface-container/50 text-on-surface-variant hover:border-primary/40 hover:text-primary'
                "
                @click="prefs.defaultMode = m.id"
              >
                <span class="material-symbols-outlined text-[20px]">{{ m.icon }}</span>
                <span class="block font-display font-bold mt-2">{{ m.label }}</span>
                <span class="block font-mono text-[10px] leading-relaxed mt-1">{{ m.helper }}</span>
              </button>
            </div>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
            <div class="space-y-3">
              <label class="font-mono text-micro text-on-surface-variant uppercase">默认分辨率</label>
              <div class="grid grid-cols-3 gap-2">
                <button
                  v-for="r in RESOLUTIONS"
                  :key="r"
                  type="button"
                  class="rounded-xl border px-3 py-3 font-mono text-[12px] transition-all"
                  :class="
                    prefs.defaultResolution === r
                      ? 'border-primary/60 bg-primary/15 text-primary'
                      : 'border-outline-variant/35 bg-surface-container/50 text-on-surface-variant hover:border-primary/40'
                  "
                  @click="prefs.defaultResolution = r"
                >
                  {{ r }}
                </button>
              </div>
            </div>

            <div class="space-y-3">
              <label class="font-mono text-micro text-on-surface-variant uppercase">默认风格</label>
              <select v-model="prefs.defaultStyle" class="field font-mono text-[12px]">
                <option v-for="style in styleOptions" :key="style" :value="style">{{ style }}</option>
              </select>
            </div>
          </div>

          <div class="space-y-3">
            <label class="font-mono text-micro text-on-surface-variant uppercase">默认宽高比</label>
            <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
              <button
                v-for="a in ASPECTS"
                :key="a.id"
                type="button"
                class="rounded-xl border px-3 py-3 font-mono text-[11px] transition-all"
                :class="
                  prefs.defaultAspectRatio === a.id
                    ? 'border-secondary/60 bg-secondary/10 text-secondary'
                    : 'border-outline-variant/35 bg-surface-container/50 text-on-surface-variant hover:border-secondary/40'
                "
                @click="prefs.defaultAspectRatio = a.id"
              >
                {{ a.label }}
              </button>
            </div>
          </div>
        </article>

        <aside class="xl:col-span-4 space-y-6">
          <article
            v-reveal="60"
            class="grad-border glass-panel rounded-2xl p-6 space-y-4"
            :class="focusBilling ? 'shadow-[0_0_48px_rgba(56,232,255,0.18)]' : ''"
          >
            <div>
              <p class="font-mono text-micro text-on-surface-variant uppercase">套餐与积分 · BILLING</p>
              <h2 class="font-display text-title font-bold text-on-surface mt-2">升级套餐</h2>
              <p class="text-sm text-on-surface-variant mt-2">
                当前版本不接第三方支付；升级入口会说明积分规则，并引导管理员通过 Admin 页面充值。
              </p>
            </div>

            <div class="space-y-3">
              <div
                v-for="plan in planCards"
                :key="plan.name"
                class="rounded-xl border border-outline-variant/30 bg-surface-container/45 p-4"
              >
                <div class="flex items-center justify-between gap-3">
                  <span class="font-display font-bold" :class="plan.tone">{{ plan.name }}</span>
                  <span class="font-mono text-[11px] text-secondary">{{ plan.price }}</span>
                </div>
                <ul class="mt-3 space-y-2">
                  <li
                    v-for="feature in plan.features"
                    :key="feature"
                    class="flex gap-2 text-xs text-on-surface-variant"
                  >
                    <span class="material-symbols-outlined text-success text-[15px]">check_circle</span>
                    <span>{{ feature }}</span>
                  </li>
                </ul>
              </div>
            </div>
          </article>

          <article v-reveal="80" class="glass-panel rounded-2xl p-6 space-y-5">
            <h2 class="font-display text-title font-bold text-on-surface flex items-center gap-2">
              <span class="material-symbols-outlined text-tertiary text-[20px]">view_comfy</span>
              界面偏好
            </h2>

            <label class="flex items-start gap-3 rounded-xl border border-outline-variant/30 bg-surface-container/40 p-4">
              <input v-model="prefs.compactGallery" type="checkbox" class="mt-1 accent-cyan-300" />
              <span>
                <span class="block font-display text-on-surface font-semibold">紧凑作品库</span>
                <span class="block font-mono text-[11px] text-on-surface-variant mt-1">
                  Gallery 会使用更密集的网格展示更多作品。
                </span>
              </span>
            </label>

            <label class="flex items-start gap-3 rounded-xl border border-outline-variant/30 bg-surface-container/40 p-4">
              <input v-model="prefs.emailDigest" type="checkbox" class="mt-1 accent-cyan-300" />
              <span>
                <span class="block font-display text-on-surface font-semibold">邮件摘要提示</span>
                <span class="block font-mono text-[11px] text-on-surface-variant mt-1">
                  预留给后续邮件摘要；当前仅保存偏好，不会发送邮件。
                </span>
              </span>
            </label>
          </article>

          <article v-reveal="140" class="grad-border glass-panel rounded-2xl p-6 space-y-4">
            <div>
              <p class="font-mono text-micro text-on-surface-variant uppercase">本地配置 · LOCAL ONLY</p>
              <p class="text-on-surface-variant text-sm mt-2">
                设置写入浏览器 localStorage。换浏览器或清理缓存后会回到默认值。
              </p>
            </div>
            <div class="flex gap-3">
              <button type="button" class="btn-primary flex-1" @click="save">保存设置</button>
              <button type="button" class="btn-ghost flex-1" @click="reset">恢复默认</button>
            </div>
            <p v-if="saved" class="font-mono text-[12px] text-success flex items-center gap-2">
              <span class="material-symbols-outlined text-[16px]">check_circle</span>
              已保存，下次进入工作台生效。
            </p>
          </article>
        </aside>
      </section>
    </div>
  </main>
</template>
