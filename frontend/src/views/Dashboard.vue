<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'
import { useGenerationsStore } from '@/stores/generations'
import { COSTS, RESOLUTIONS, ASPECTS } from '@/lib/format'
import { getPresetById } from '@/lib/presets'
import type { Resolution, AspectRatio, Preset } from '@/types'

const gen = useGenerationsStore()
const { active } = storeToRefs(gen)

const prompt = ref('')
const negativePrompt = ref('')
const showNegative = ref(false)
const resolution = ref<Resolution>('2K')
const aspect = ref<AspectRatio>('1:1')
const style = ref('Cinematic')
const errorMsg = ref('')
const presetHint = ref('')

const route = useRoute()
const router = useRouter()

const cost = computed(() => COSTS[resolution.value])

const isBusy = computed(
  () => active.value?.status === 'pending' || active.value?.status === 'processing'
)

// Time-based fake progress: an ease-out climb over ~50s that caps at 95%,
// snapping to 100 only when the job actually completes (or 0 on failure).
const simProgress = ref(0)
let progTimer: ReturnType<typeof setInterval> | undefined

function startFakeProgress() {
  stopFakeProgress()
  simProgress.value = 3
  const start = performance.now()
  progTimer = setInterval(() => {
    const t = Math.min(1, (performance.now() - start) / 50000)
    simProgress.value = Math.min(95, Math.round(95 * (1 - (1 - t) * (1 - t))))
  }, 400)
}

function stopFakeProgress() {
  if (progTimer) {
    clearInterval(progTimer)
    progTimer = undefined
  }
}

watch(isBusy, (busy) => {
  if (busy) startFakeProgress()
  else stopFakeProgress()
})

watch(
  () => active.value?.status,
  (status) => {
    if (status === 'completed') {
      stopFakeProgress()
      simProgress.value = 100
    } else if (status === 'failed') {
      stopFakeProgress()
      simProgress.value = 0
    }
  }
)

watch(
  () => route.query.preset,
  (presetId) => {
    applyPresetFromQuery(presetId)
  },
  { immediate: true }
)

const progressPct = computed(() => simProgress.value)

onMounted(() => {
  // A return visit should start idle: don't re-show a previously finished image
  // (the store's `active` persists across route changes). An in-flight job keeps
  // polling — its store-owned timer survives navigation, so re-arm defensively.
  if (active.value && (active.value.status === 'completed' || active.value.status === 'failed')) {
    gen.active = null
  } else if (active.value) {
    gen.pollActive()
  }
})

onUnmounted(() => {
  stopFakeProgress()
})

function pickResolution(r: Resolution) {
  resolution.value = r
}
function pickAspect(a: AspectRatio) {
  aspect.value = a
}

// Approximate ratio box dimensions for the aspect-grid icons
function ratioBox(a: { w: number; h: number }) {
  const max = 40
  const longest = Math.max(a.w, a.h)
  const width = Math.round((a.w / longest) * max)
  const height = Math.round((a.h / longest) * max)
  return { width: `${Math.max(width, 16)}px`, height: `${Math.max(height, 16)}px` }
}

async function generate() {
  if (isBusy.value) return
  if (!prompt.value.trim()) {
    errorMsg.value = '请先描述你的画面再开始合成。'
    return
  }
  errorMsg.value = ''
  try {
    await gen.create({
      prompt: prompt.value.trim(),
      negative_prompt: negativePrompt.value.trim() || undefined,
      style: style.value,
      resolution: resolution.value,
      aspect_ratio: aspect.value,
    })
  } catch (e: any) {
    if (e?.response?.status === 402) {
      errorMsg.value = '积分不足。'
    } else {
      errorMsg.value = e?.response?.data?.error || e?.message || '合成失败。'
    }
  }
}

function applyPreset(preset: Preset) {
  prompt.value = preset.prompt_seed
  style.value = preset.style
  resolution.value = preset.suggested_resolution
  aspect.value = preset.suggested_aspect_ratio
}

function applyPresetFromQuery(rawPresetId: unknown) {
  if (typeof rawPresetId !== 'string' || rawPresetId.trim() === '') {
    presetHint.value = ''
    return
  }

  const preset = getPresetById(rawPresetId)
  if (!preset) {
    presetHint.value = '未识别的预设参数，已忽略。'
    return
  }

  applyPreset(preset)
  presetHint.value = `已载入「${preset.title}」预设，参数可继续修改。`
  if (route.path === '/') {
    router.replace({ path: '/', query: {} })
  }
}
</script>

<template>
  <!-- Workspace Area -->
  <div class="flex-1 flex flex-col lg:flex-row p-margin-sm md:p-margin-lg gap-gutter h-full overflow-y-auto lg:overflow-hidden">
    <!-- Left Panel: Settings Toolbar -->
    <aside class="w-full lg:w-[320px] flex-shrink-0 flex flex-col gap-6">
      <div class="glass-panel rounded-xl p-6 flex flex-col gap-6 h-full overflow-y-auto">
        <h2
          class="font-body-md text-body-md font-bold text-on-surface border-b border-outline-variant/30 pb-3 flex items-center gap-2"
        >
          <span class="material-symbols-outlined text-primary">tune</span> 生成参数
        </h2>

        <div v-if="presetHint" class="text-label-sm text-on-surface-variant">{{ presetHint }}</div>

        <!-- Resolution -->
        <div class="flex flex-col gap-3">
          <label class="font-label-sm text-label-sm text-on-surface-variant uppercase tracking-wider"
            >分辨率</label
          >
          <div
            class="grid grid-cols-2 gap-2 bg-surface-container-lowest p-1 rounded-lg border border-outline-variant/30"
          >
            <button
              v-for="r in RESOLUTIONS"
              :key="r"
              type="button"
              class="py-2 text-center rounded-md font-label-sm text-label-sm transition-colors"
              :class="
                resolution === r
                  ? 'bg-primary-container text-on-primary-container font-bold shadow-sm glow-shadow'
                  : 'text-on-surface-variant hover:bg-surface-variant/50'
              "
              @click="pickResolution(r)"
            >
              {{ r }}
            </button>
          </div>
        </div>

        <!-- Aspect Ratio -->
        <div class="flex flex-col gap-3">
          <label class="font-label-sm text-label-sm text-on-surface-variant uppercase tracking-wider"
            >宽高比</label
          >
          <div class="grid grid-cols-2 gap-3">
            <button
              v-for="a in ASPECTS"
              :key="a.id"
              type="button"
              class="flex flex-col items-center justify-center gap-2 p-3 rounded-lg border transition-all"
              :class="
                aspect === a.id
                  ? 'border-primary/50 bg-primary/10 text-primary glow-shadow'
                  : 'border-outline-variant/50 hover:bg-surface-variant/30 text-on-surface-variant'
              "
              @click="pickAspect(a.id)"
            >
              <div
                class="border-2 border-current rounded-sm"
                :style="ratioBox(a)"
              ></div>
              <span
                class="font-label-sm text-label-sm"
                :class="aspect === a.id ? 'font-bold' : ''"
                >{{ a.label }}</span
              >
            </button>
          </div>
        </div>

        <!-- Energy Cost Estimation -->
        <div
          class="mt-auto bg-surface-container p-4 rounded-lg border border-outline-variant/20 flex justify-between items-center"
        >
          <div class="flex items-center gap-2">
            <span class="material-symbols-outlined text-secondary">flash_on</span>
            <span class="font-label-sm text-label-sm text-on-surface-variant">能量消耗</span>
          </div>
          <span class="font-body-md text-body-md font-bold text-on-surface">-{{ cost }} 积分</span>
        </div>
      </div>
    </aside>

    <!-- Central Area: Canvas & Prompt -->
    <section class="flex-1 flex flex-col gap-6 lg:overflow-hidden">
      <!-- Prompt Input Area -->
      <div
        class="glass-panel rounded-xl p-1 relative flex-shrink-0 group focus-within:border-primary/50 focus-within:glow-shadow transition-all duration-300"
      >
        <div class="bg-surface-container-lowest rounded-lg overflow-hidden flex flex-col relative">
          <textarea
            v-model="prompt"
            class="w-full h-32 bg-transparent border-none text-on-surface font-body-md p-4 resize-none focus:ring-0 focus:outline-none placeholder:text-on-surface-variant/50 leading-relaxed"
            placeholder="描述你的画面… 例如：“午夜里一片生物荧光森林，悬浮着晶莹剔透的结晶体…”"
          ></textarea>
          <div
            class="bg-surface-container/80 backdrop-blur-md px-4 py-3 flex justify-between items-center border-t border-outline-variant/20"
          >
            <div class="flex gap-2 items-center">
              <button
                type="button"
                class="w-8 h-8 rounded bg-surface-variant text-on-surface-variant flex items-center justify-center hover:text-primary hover:bg-primary/10 transition-colors"
                title="添加参考图片"
              >
                <span class="material-symbols-outlined text-[18px]">add_photo_alternate</span>
              </button>
              <button
                type="button"
                class="w-8 h-8 rounded flex items-center justify-center transition-colors"
                :class="
                  showNegative
                    ? 'bg-primary/10 text-primary'
                    : 'bg-surface-variant text-on-surface-variant hover:text-primary hover:bg-primary/10'
                "
                title="负面提示词"
                @click="showNegative = !showNegative"
              >
                <span class="material-symbols-outlined text-[18px]">block</span>
              </button>
              <div class="h-8 w-px bg-outline-variant/40 mx-1"></div>
              <div
                class="flex items-center px-2 py-1 rounded bg-surface text-on-surface-variant border border-outline-variant/30 font-label-sm text-[10px]"
              >
                <span class="w-2 h-2 rounded-full bg-primary/50 mr-2"></span> 风格：{{ style }}
              </div>
            </div>
            <button
              type="button"
              class="bg-primary text-on-primary font-bold px-6 py-2 rounded-lg hover:bg-primary/90 transition-all glow-shadow active:scale-95 flex items-center gap-2 disabled:opacity-60 disabled:cursor-not-allowed"
              :disabled="isBusy"
              @click="generate"
            >
              <span>{{ isBusy ? '合成中' : '生成' }}</span>
              <span
                class="material-symbols-outlined text-[18px]"
                :class="isBusy ? 'animate-spin' : ''"
                >{{ isBusy ? 'progress_activity' : 'bolt' }}</span
              >
            </button>
          </div>
        </div>
      </div>

      <!-- Negative prompt input (revealed) -->
      <div v-if="showNegative" class="flex-shrink-0">
        <input
          v-model="negativePrompt"
          type="text"
          class="w-full px-4 py-2.5 bg-surface-container-lowest border border-outline-variant/30 rounded-lg text-on-surface font-label-sm text-label-sm placeholder:text-on-surface-variant/50 focus:outline-none focus:border-primary/50 transition-colors"
          placeholder="负面提示词 — 你想避免出现的内容…"
        />
      </div>

      <!-- Inline error -->
      <div
        v-if="errorMsg"
        class="flex-shrink-0 flex items-center gap-2 px-4 py-2.5 rounded-lg bg-error-container/20 border border-error/30 text-error font-label-sm text-label-sm"
      >
        <span class="material-symbols-outlined text-sm">error</span>
        <span>{{ errorMsg }}</span>
      </div>

      <!-- Canvas / Result Area -->
      <div
        class="flex-1 glass-panel rounded-xl overflow-hidden relative flex flex-col items-center justify-center min-h-[400px]"
      >
        <!-- Background Pattern for Canvas Area -->
        <div
          class="absolute inset-0 opacity-[0.03]"
          style="background-image: radial-gradient(#d2bbff 1px, transparent 1px); background-size: 20px 20px;"
        ></div>

        <!-- Completed image -->
        <img
          v-if="active && active.status === 'completed' && active.image_url"
          :src="active.image_url"
          :alt="active.prompt"
          class="relative z-10 max-w-full max-h-full object-contain"
        />

        <!-- Synthesizing state -->
        <div
          v-else-if="isBusy"
          class="absolute inset-0 flex flex-col items-center justify-center z-10 bg-surface-container-lowest/80 backdrop-blur-sm"
        >
          <div class="relative w-24 h-24 mb-6">
            <div class="absolute inset-0 border-4 border-surface-variant rounded-full"></div>
            <div
              class="absolute inset-0 border-4 border-primary rounded-full border-t-transparent animate-spin"
            ></div>
            <div class="absolute inset-0 flex items-center justify-center">
              <span class="material-symbols-outlined text-primary text-3xl animate-pulse"
                >model_training</span
              >
            </div>
          </div>
          <h3 class="font-headline-lg-mobile text-headline-lg-mobile text-on-surface mb-2">
            正在合成画面
          </h3>
          <p class="font-body-md text-body-md text-on-surface-variant max-w-md text-center">
            正根据你的提示词渲染体积光与晶体结构…
          </p>
          <div class="w-64 h-1.5 bg-surface-variant rounded-full mt-6 overflow-hidden">
            <div
              class="h-full bg-primary rounded-full shadow-[0_0_10px_rgba(124,58,237,0.5)] transition-all duration-700"
              :style="{ width: progressPct + '%' }"
            ></div>
          </div>
        </div>

        <!-- Failed state -->
        <div
          v-else-if="active && active.status === 'failed'"
          class="relative z-10 flex flex-col items-center justify-center text-center px-6"
        >
          <span class="material-symbols-outlined text-error text-5xl mb-4">error</span>
          <h3 class="font-headline-lg-mobile text-headline-lg-mobile text-on-surface mb-2">
            合成失败
          </h3>
          <p class="font-body-md text-body-md text-on-surface-variant max-w-md">
            {{ active.error || '渲染未能完成，积分已退还。' }}
          </p>
        </div>

        <!-- Idle placeholder -->
        <div
          v-else
          class="relative z-10 flex flex-col items-center justify-center text-center px-6"
        >
          <div
            class="w-20 h-20 rounded-2xl bg-surface-variant/40 border border-outline-variant/30 flex items-center justify-center mb-6"
          >
            <span class="material-symbols-outlined text-primary text-4xl">auto_awesome</span>
          </div>
          <h3 class="font-headline-lg-mobile text-headline-lg-mobile text-on-surface mb-2">
            画布就绪
          </h3>
          <p class="font-body-md text-body-md text-on-surface-variant max-w-md">
            配置参数、描述你的画面，然后开始合成。
          </p>
        </div>
      </div>
    </section>
  </div>
</template>
