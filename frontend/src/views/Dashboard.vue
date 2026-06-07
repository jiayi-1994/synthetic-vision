<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useRoute, useRouter } from 'vue-router'
import { useGenerationsStore } from '@/stores/generations'
import { COSTS, RESOLUTIONS, ASPECTS } from '@/lib/format'
import { getPresetById } from '@/lib/presets'
import { useMagnet, clickSpark } from '@/composables/useInteractions'
import type { Resolution, AspectRatio, GenerationMode, Preset } from '@/types'

const gen = useGenerationsStore()
const { active, activeBatchProgress } = storeToRefs(gen)

const prompt = ref('')
const negativePrompt = ref('')
const showNegative = ref(false)
const resolution = ref<Resolution>('2K')
const aspect = ref<AspectRatio>('1:1')
const style = ref('Cinematic')
const errorMsg = ref('')
const presetHint = ref('')
const mode = ref<GenerationMode>('text')
const sourceInput = ref<HTMLInputElement | null>(null)
const sourceFile = ref<File | null>(null)
const sourcePreview = ref('')
const sourceDims = ref({ width: 0, height: 0 })
const maskCanvas = ref<HTMLCanvasElement | null>(null)
const brushSize = ref(44)
const isErasing = ref(false)
const hasMask = ref(false)
const isPainting = ref(false)

const route = useRoute()
const router = useRouter()

const genBtn = useMagnet(5)
const MAX_BATCH_COUNT = 8

const modes: { id: GenerationMode; label: string; icon: string; helper: string }[] = [
  { id: 'text', label: '文生图', icon: 'auto_awesome', helper: '从提示词直接生成新画面' },
  { id: 'image', label: '图生图', icon: 'add_photo_alternate', helper: '上传参考图后重绘风格与细节' },
  { id: 'edit', label: '局部修图', icon: 'brush', helper: '涂抹区域后进行局部替换或移除' },
]
const editActions = [
  { id: 'replace', label: '替换选区', icon: 'ink_eraser', helper: '默认：只重绘你涂抹的区域，未选区保持原图。' },
  { id: 'remove', label: '移除物体', icon: 'backspace', helper: '适合擦掉选区内的物体，并用周围内容自然补全。' },
  { id: 'outpaint', label: '扩展背景', icon: 'open_in_full', helper: '适合延展边缘、背景和构图空间。' },
  { id: 'material', label: '重绘材质', icon: 'texture', helper: '保留选区结构，替换表面质感、颜色或材质。' },
] as const
type EditActionId = (typeof editActions)[number]['id']
const editAction = ref<EditActionId>('replace')
const batchCount = ref(1)

const cost = computed(() => COSTS[resolution.value])
const normalizedBatchCount = computed(() => normalizeBatchCount(batchCount.value))
const estimatedCost = computed(() => cost.value * normalizedBatchCount.value)
const needsSource = computed(() => mode.value === 'image' || mode.value === 'edit')
const selectedEditAction = computed(
  () => editActions.find((action) => action.id === editAction.value) ?? editActions[0]
)
const activeOutputSourceUrl = computed(() => {
  if (!needsSource.value || sourceFile.value) return ''
  if (active.value?.status !== 'completed' || !active.value.image_url) return ''
  return active.value.image_url
})
const effectiveSourcePreview = computed(() =>
  needsSource.value ? sourcePreview.value || activeOutputSourceUrl.value : ''
)
const hasSourceForRequest = computed(() => !!sourceFile.value || !!activeOutputSourceUrl.value)
const sourceName = computed(() => {
  if (sourceFile.value) return sourceFile.value.name
  if (activeOutputSourceUrl.value && active.value) return `当前结果 · ${active.value.id.slice(0, 8)}`
  return '未选择参考图'
})
const modeLabel = computed(() => {
  if (mode.value === 'image') return '图生图'
  if (mode.value === 'edit') return '局部修图'
  return '文生图'
})
const promptPlaceholder = computed(() => {
  if (mode.value === 'image') return '描述你想如何重绘参考图… 例如：“保留主体构图，改成赛博霓虹摄影棚海报…”'
  if (mode.value === 'edit') return '描述涂抹区域要如何变化… 例如：“把选区里的旧路灯替换成发光晶体装置…”'
  return '描述你的画面… 例如：“午夜里一片生物荧光森林，悬浮着晶莹剔透的结晶体…”'
})
const submitLabel = computed(() => {
  if (isBusy.value) return '合成中'
  if (mode.value === 'image') return '重绘'
  if (mode.value === 'edit') return '修图'
  return normalizedBatchCount.value > 1 ? `生成${normalizedBatchCount.value}张` : '生成'
})

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

function normalizeBatchCount(v: number): number {
  if (!Number.isFinite(v) || Number.isNaN(v)) return 1
  const c = Math.floor(v)
  if (c < 1) return 1
  if (c > MAX_BATCH_COUNT) return MAX_BATCH_COUNT
  return c
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
const batchProgressPct = computed(
  () =>
    activeBatchProgress.value.total === 0
      ? 0
      : Math.round((activeBatchProgress.value.done / activeBatchProgress.value.total) * 100),
)

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
  revokeSourcePreview()
})

function pickResolution(r: Resolution) {
  resolution.value = r
}
function pickAspect(a: AspectRatio) {
  aspect.value = a
}

function pickMode(next: GenerationMode) {
  mode.value = next
  errorMsg.value = ''
  if (next === 'edit') {
    hasMask.value = false
    nextTick(() => resetMaskCanvas())
  }
}

// Approximate ratio box dimensions for the aspect-grid icons
function ratioBox(a: { w: number; h: number }) {
  const max = 36
  const longest = Math.max(a.w, a.h)
  const width = Math.round((a.w / longest) * max)
  const height = Math.round((a.h / longest) * max)
  return { width: `${Math.max(width, 14)}px`, height: `${Math.max(height, 14)}px` }
}

async function generate(e?: MouseEvent) {
  if (isBusy.value) return
  const promptText = prompt.value.trim()
  if (!promptText) {
    errorMsg.value = '请先描述你的画面再开始合成。'
    return
  }
  if (needsSource.value && !hasSourceForRequest.value) {
    errorMsg.value = '请先上传一张参考图片，或先生成一张图片后再二次修改。'
    return
  }
  let maskImage: Blob | undefined
  if (mode.value === 'edit') {
    if (!hasMask.value) {
      errorMsg.value = '请先在参考图上涂抹要修图的区域。'
      return
    }
    maskImage = await exportMaskBlob()
    if (!maskImage) {
      errorMsg.value = 'Mask 导出失败，请清除后重新涂抹。'
      return
    }
  }
  const sourceImage = await sourceImageForRequest()
  if (needsSource.value && !sourceImage) {
    errorMsg.value = '当前图片读取失败，请上传一张参考图后重试。'
    return
  }
  const requestCount = normalizedBatchCount.value
  batchCount.value = requestCount
  if (e) clickSpark(e)
  errorMsg.value = ''
  const effectivePrompt =
    mode.value === 'edit' ? `${selectedEditAction.value.label}：${promptText}` : promptText
  try {
    await gen.create({
      mode: mode.value,
      prompt: effectivePrompt,
      negative_prompt: negativePrompt.value.trim() || undefined,
      style: style.value,
      resolution: resolution.value,
      aspect_ratio: aspect.value,
      count: requestCount,
      source_image: sourceImage,
      mask_image: maskImage,
    })
  } catch (err: any) {
    if (err?.response?.status === 402) {
      errorMsg.value = '积分不足。'
    } else {
      errorMsg.value = err?.response?.data?.error || err?.message || '合成失败。'
    }
  }
}

async function sourceImageForRequest(): Promise<File | undefined> {
  if (sourceFile.value) return sourceFile.value
  const url = activeOutputSourceUrl.value
  if (!url || !active.value) return undefined
  const response = await fetch(url)
  if (!response.ok) return undefined
  const blob = await response.blob()
  return new File([blob], `${active.value.id}.png`, { type: blob.type || 'image/png' })
}

function openSourcePicker() {
  sourceInput.value?.click()
}

function handleSourceSelect(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  setSourceFile(file)
}

function setSourceFile(file: File) {
  errorMsg.value = ''
  const validTypes = ['image/png', 'image/jpeg', 'image/webp']
  if (!validTypes.includes(file.type)) {
    errorMsg.value = '参考图仅支持 PNG、JPEG 或 WebP。'
    return
  }
  if (file.size > 10 * 1024 * 1024) {
    errorMsg.value = '参考图不能超过 10MB。'
    return
  }
  revokeSourcePreview()
  sourceFile.value = file
  sourcePreview.value = URL.createObjectURL(file)
  if (mode.value === 'text') {
    mode.value = 'image'
  }
  hasMask.value = false
  nextTick(() => resetMaskCanvas())
}

function removeSource() {
  revokeSourcePreview()
  sourceFile.value = null
  sourceDims.value = { width: 0, height: 0 }
  clearMask()
}

function revokeSourcePreview() {
  if (sourcePreview.value) {
    URL.revokeObjectURL(sourcePreview.value)
    sourcePreview.value = ''
  }
}

function onSourceImageLoad(event: Event) {
  const img = event.target as HTMLImageElement
  sourceDims.value = {
    width: img.naturalWidth || 0,
    height: img.naturalHeight || 0,
  }
  resetMaskCanvas()
}

function resetMaskCanvas() {
  const canvas = maskCanvas.value
  if (!canvas || sourceDims.value.width <= 0 || sourceDims.value.height <= 0) return
  canvas.width = sourceDims.value.width
  canvas.height = sourceDims.value.height
  const ctx = canvas.getContext('2d')
  ctx?.clearRect(0, 0, canvas.width, canvas.height)
  hasMask.value = false
}

function clearMask() {
  const canvas = maskCanvas.value
  if (!canvas) {
    hasMask.value = false
    return
  }
  canvas.getContext('2d')?.clearRect(0, 0, canvas.width, canvas.height)
  hasMask.value = false
}

function paintMask(event: PointerEvent) {
  const canvas = maskCanvas.value
  if (!canvas) return
  const rect = canvas.getBoundingClientRect()
  if (rect.width <= 0 || rect.height <= 0) return
  const x = ((event.clientX - rect.left) / rect.width) * canvas.width
  const y = ((event.clientY - rect.top) / rect.height) * canvas.height
  const ctx = canvas.getContext('2d')
  if (!ctx) return
  ctx.save()
  ctx.globalCompositeOperation = isErasing.value ? 'destination-out' : 'source-over'
  ctx.fillStyle = 'rgba(56, 232, 255, 0.42)'
  ctx.shadowColor = 'rgba(255, 61, 240, 0.45)'
  ctx.shadowBlur = Math.max(4, brushSize.value / 4)
  ctx.beginPath()
  ctx.arc(x, y, brushSize.value, 0, Math.PI * 2)
  ctx.fill()
  ctx.restore()
  if (!isErasing.value) {
    hasMask.value = true
    if (errorMsg.value.includes('涂抹')) {
      errorMsg.value = ''
    }
  }
}

function startPainting(event: PointerEvent) {
  if (mode.value !== 'edit') return
  isPainting.value = true
  ;(event.currentTarget as HTMLCanvasElement).setPointerCapture(event.pointerId)
  paintMask(event)
}

function continuePainting(event: PointerEvent) {
  if (!isPainting.value) return
  paintMask(event)
}

function stopPainting(event: PointerEvent) {
  isPainting.value = false
  try {
    ;(event.currentTarget as HTMLCanvasElement).releasePointerCapture(event.pointerId)
  } catch {
    // Pointer capture may already be released by the browser.
  }
}

function exportMaskBlob(): Promise<Blob | null> {
  const canvas = maskCanvas.value
  if (!canvas || !hasMask.value) return Promise.resolve(null)
  const mask = document.createElement('canvas')
  mask.width = canvas.width
  mask.height = canvas.height
  const ctx = mask.getContext('2d')
  if (!ctx) return Promise.resolve(null)
  const overlay = canvas.getContext('2d')?.getImageData(0, 0, canvas.width, canvas.height)
  if (!overlay) return Promise.resolve(null)
  const out = ctx.createImageData(mask.width, mask.height)
  for (let i = 0; i < out.data.length; i += 4) {
    const selected = overlay.data[i + 3] > 0
    out.data[i] = 0
    out.data[i + 1] = 0
    out.data[i + 2] = 0
    out.data[i + 3] = selected ? 0 : 255
  }
  ctx.putImageData(out, 0, 0)
  return new Promise((resolve) => mask.toBlob((blob) => resolve(blob), 'image/png'))
}

function applyEditPreset(action: (typeof editActions)[number]) {
  editAction.value = action.id
  if (!prompt.value.trim()) {
    prompt.value = action.helper
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
  <div
    class="flex-1 flex flex-col lg:flex-row p-margin-sm md:p-margin-lg gap-gutter h-full overflow-y-auto lg:overflow-hidden"
  >
    <!-- Left Panel: Settings Toolbar -->
    <aside class="w-full lg:w-[320px] flex-shrink-0 flex flex-col gap-6 animate-rise-in">
      <div class="glass-panel rounded-2xl p-6 flex flex-col gap-6 h-full overflow-y-auto">
        <h2
          class="font-display text-title font-bold text-on-surface border-b border-outline-variant/30 pb-3 flex items-center gap-2"
        >
          <span class="material-symbols-outlined text-primary">tune</span> 生成参数
          <span class="ml-auto font-mono text-micro text-on-surface-variant uppercase">PARAMS</span>
        </h2>

        <div v-if="presetHint" class="font-mono text-[11px] text-tertiary leading-relaxed">
          {{ presetHint }}
        </div>

        <!-- Generation Mode -->
        <div class="flex flex-col gap-3">
          <label class="font-mono text-micro text-on-surface-variant uppercase">模式 · MODE</label>
          <div
            class="grid grid-cols-3 gap-1 bg-surface-container-lowest p-1 rounded-xl border border-outline-variant/30"
          >
            <button
              v-for="m in modes"
              :key="m.id"
              type="button"
              class="min-w-0 px-2 py-2.5 rounded-lg border text-center transition-all"
              :class="
                mode === m.id
                  ? 'border-primary/50 bg-primary/10 text-primary shadow-[0_0_16px_rgba(56,232,255,0.12)]'
                  : 'border-transparent text-on-surface-variant hover:bg-primary/5 hover:text-primary'
              "
              :title="m.helper"
              @click="pickMode(m.id)"
            >
              <span class="material-symbols-outlined text-[18px] block mx-auto mb-1">{{ m.icon }}</span>
              <span class="font-mono text-[10px] leading-none">{{ m.label }}</span>
            </button>
          </div>
          <p class="font-mono text-[11px] leading-relaxed text-on-surface-variant">
            {{ modes.find((m) => m.id === mode)?.helper }}
          </p>
        </div>

        <!-- Resolution -->
        <div class="flex flex-col gap-3">
          <label class="font-mono text-micro text-on-surface-variant uppercase">分辨率 · RESOLUTION</label>
          <div
            class="grid grid-cols-2 gap-2 bg-surface-container-lowest p-1 rounded-xl border border-outline-variant/30"
          >
            <button
              v-for="r in RESOLUTIONS"
              :key="r"
              type="button"
              class="py-2.5 text-center rounded-lg font-mono text-[13px] tracking-wider transition-all"
              :class="
                resolution === r
                  ? 'bg-primary/15 text-primary border border-primary/40 shadow-[0_0_16px_rgba(56,232,255,0.15)]'
                  : 'text-on-surface-variant border border-transparent hover:bg-primary/5 hover:text-primary'
              "
              @click="pickResolution(r)"
            >
              {{ r }}
            </button>
          </div>
        </div>

        <!-- Aspect Ratio -->
        <div class="flex flex-col gap-3">
          <label class="font-mono text-micro text-on-surface-variant uppercase">宽高比 · ASPECT</label>
          <div class="grid grid-cols-2 gap-3">
            <button
              v-for="a in ASPECTS"
              :key="a.id"
              type="button"
              class="flex flex-col items-center justify-center gap-2 p-3 rounded-xl border transition-all"
              :class="
                aspect === a.id
                  ? 'border-primary/50 bg-primary/10 text-primary shadow-[0_0_16px_rgba(56,232,255,0.12)]'
                  : 'border-outline-variant/40 hover:bg-surface-variant/30 text-on-surface-variant hover:border-primary/30'
              "
              @click="pickAspect(a.id)"
            >
              <div class="border-2 border-current rounded-sm" :style="ratioBox(a)"></div>
              <span class="font-mono text-[11px]" :class="aspect === a.id ? 'font-bold' : ''">{{
                a.label
              }}</span>
            </button>
          </div>
        </div>

        <!-- Energy Cost Estimation -->
        <div
          class="mt-auto grad-border bg-surface-container p-4 rounded-xl flex flex-col gap-3"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <span class="material-symbols-outlined text-secondary">flash_on</span>
              <span class="font-mono text-micro text-on-surface-variant uppercase">能量消耗</span>
            </div>
            <span class="font-mono text-[11px] text-on-surface-variant uppercase">每次生成</span>
          </div>
          <div class="flex items-center justify-between gap-3">
            <label class="font-mono text-micro text-on-surface-variant uppercase flex items-center gap-2">
              <span class="material-symbols-outlined text-[16px]">filter_1</span>
              生成数量
            </label>
            <input
              v-model.number="batchCount"
              type="number"
              min="1"
              :max="MAX_BATCH_COUNT"
              class="w-20 px-2.5 py-1 rounded-md bg-surface-variant border border-outline-variant/40 text-on-surface text-right"
              @change="batchCount = normalizeBatchCount(batchCount)"
            />
          </div>
          <p class="font-mono text-[10px] text-on-surface-variant">
            取值 1~{{ MAX_BATCH_COUNT }}，空值/非法会回退为 1；每次提交按“每张{{ cost }} CR”计费。
          </p>
          <div class="flex justify-between items-center">
            <span class="font-mono text-micro text-on-surface-variant uppercase">本次预估</span>
            <span class="font-display text-lg font-bold text-on-surface"
              >−{{ estimatedCost }}<span class="text-on-surface-variant text-sm font-mono ml-1">CR</span></span
            >
          </div>
        </div>
      </div>
    </aside>

    <!-- Central Area: Canvas & Prompt -->
    <section class="flex-1 flex flex-col gap-6 lg:overflow-hidden animate-rise-in" style="animation-delay:.08s">
      <!-- Prompt Input Area -->
      <div
        class="glass-panel rounded-2xl p-1 relative flex-shrink-0 group focus-within:border-primary/50 focus-within:glow-shadow transition-all duration-300"
      >
        <div class="bg-surface-container-lowest rounded-xl overflow-hidden flex flex-col relative">
          <textarea
            v-model="prompt"
            class="w-full h-32 bg-transparent border-none text-on-surface font-body-md p-4 resize-none focus:ring-0 focus:outline-none placeholder:text-on-surface-variant/40 leading-relaxed"
            :placeholder="promptPlaceholder"
          ></textarea>
          <div
            class="bg-surface-container/80 backdrop-blur-md px-4 py-3 flex justify-between items-center border-t border-outline-variant/20"
          >
            <div class="flex gap-2 items-center">
              <button
                type="button"
                class="w-10 h-10 rounded-lg bg-surface-variant text-on-surface-variant flex items-center justify-center hover:text-primary hover:bg-primary/10 transition-colors"
                title="添加参考图片"
                @click="openSourcePicker"
              >
                <span class="material-symbols-outlined text-[18px]">add_photo_alternate</span>
              </button>
              <input
                ref="sourceInput"
                type="file"
                accept="image/png,image/jpeg,image/webp"
                class="hidden"
                @change="handleSourceSelect"
              />
              <button
                type="button"
                class="w-10 h-10 rounded-lg flex items-center justify-center transition-colors"
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
                class="flex items-center px-2.5 py-1 rounded-lg bg-surface text-tertiary border border-tertiary/30 font-mono text-[10px] uppercase tracking-wider"
              >
                <span class="w-1.5 h-1.5 rounded-full bg-tertiary mr-2"></span> {{ style }}
              </div>
            </div>
            <button
              ref="genBtn"
              type="button"
              class="btn-primary flex items-center gap-2 disabled:opacity-60"
              :disabled="isBusy"
              @click="generate($event)"
            >
              <span>{{ submitLabel }}</span>
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
          class="field font-mono text-[12px]"
          placeholder="负面提示词 — 你想避免出现的内容…"
        />
      </div>

      <!-- Source image controls -->
      <div
        v-if="needsSource || sourceFile"
        class="flex-shrink-0 glass-panel rounded-2xl p-4 flex flex-col gap-3"
      >
        <div class="flex flex-col md:flex-row md:items-center gap-3 justify-between">
          <div>
            <div class="flex items-center gap-2 text-on-surface">
              <span class="material-symbols-outlined text-primary text-[18px]">image_search</span>
              <span class="font-display text-sm font-bold">{{ modeLabel }}参考图</span>
            </div>
            <p class="font-mono text-[11px] text-on-surface-variant mt-1">
              {{
                hasSourceForRequest
                  ? sourceName
                  : active && active.status === 'completed'
                    ? '当前生成结果可直接圈选；也可以上传 PNG / JPEG / WebP。'
                    : '上传 PNG / JPEG / WebP，最大 10MB；或先生成一张图后再二次修改。'
              }}
              <span v-if="sourceDims.width"> · {{ sourceDims.width }}×{{ sourceDims.height }}</span>
            </p>
          </div>
          <div class="flex items-center gap-2">
            <button type="button" class="btn-ghost !min-h-9 !py-2 !px-3 text-xs" @click="openSourcePicker">
              {{ sourceFile || activeOutputSourceUrl ? '上传新图' : '上传图片' }}
            </button>
            <button
              v-if="sourceFile"
              type="button"
              class="btn-ghost !min-h-9 !py-2 !px-3 text-xs"
              @click="removeSource"
            >
              移除
            </button>
          </div>
        </div>

        <div v-if="mode === 'edit'" class="border-t border-outline-variant/20 pt-3 flex flex-col gap-3">
          <div
            class="rounded-xl border border-primary/25 bg-surface-container-lowest/80 p-3 shadow-[0_0_24px_rgba(56,232,255,0.08)]"
          >
            <div class="mb-2 flex flex-col gap-1 md:flex-row md:items-center md:justify-between">
              <span class="font-mono text-[11px] uppercase tracking-wider text-primary">
                处理方式 · ACTION
              </span>
              <span class="font-mono text-[10px] text-on-surface-variant">
                默认已选中：{{ selectedEditAction.label }}
              </span>
            </div>
            <div class="grid grid-cols-2 gap-2 md:grid-cols-4">
              <button
                v-for="action in editActions"
                :key="action.id"
                type="button"
                class="group flex min-h-14 items-center gap-2 rounded-lg border px-2.5 py-2 text-left transition-all focus:outline-none focus:ring-2 focus:ring-primary/50"
                :class="
                  editAction === action.id
                    ? 'border-primary/60 bg-primary/15 text-primary shadow-[0_0_18px_rgba(56,232,255,0.16)]'
                    : 'border-outline-variant/35 bg-surface-container/60 text-on-surface-variant hover:border-primary/40 hover:bg-primary/10 hover:text-primary'
                "
                @click="applyEditPreset(action)"
              >
                <span
                  class="material-symbols-outlined text-[18px]"
                  :class="
                    editAction === action.id
                      ? 'text-secondary'
                      : 'text-on-surface-variant group-hover:text-primary'
                  "
                  >{{ action.icon }}</span
                >
                <span class="min-w-0">
                  <span class="block font-mono text-[11px] leading-tight">{{ action.label }}</span>
                  <span
                    v-if="editAction === action.id"
                    class="mt-0.5 block font-mono text-[9px] uppercase tracking-wider text-secondary"
                    >SELECTED</span
                  >
                </span>
              </button>
            </div>
            <p class="mt-2 font-mono text-[11px] leading-relaxed text-on-surface-variant">
              {{ selectedEditAction.helper }}
            </p>
          </div>
          <div class="grid gap-3 md:grid-cols-[1fr_auto_auto] md:items-center">
            <label class="flex items-center gap-3 font-mono text-[11px] text-on-surface-variant">
              BRUSH
              <input v-model.number="brushSize" type="range" min="12" max="120" class="w-full accent-cyan-300" />
              <span class="w-10 text-right text-primary">{{ brushSize }}</span>
            </label>
            <button type="button" class="btn-ghost !min-h-9 !py-2 !px-3 text-xs" @click="isErasing = !isErasing">
              {{ isErasing ? '擦除中' : '画笔' }}
            </button>
            <button type="button" class="btn-ghost !min-h-9 !py-2 !px-3 text-xs" @click="clearMask">
              清除 Mask
            </button>
          </div>
        </div>
      </div>

      <!-- Inline error -->
      <div
        v-if="errorMsg"
        class="flex-shrink-0 flex items-center gap-2 px-4 py-2.5 rounded-xl bg-error/10 border border-error/30 text-error font-mono text-[12px]"
      >
        <span class="material-symbols-outlined text-sm">error</span>
        <span>{{ errorMsg }}</span>
      </div>

      <!-- Canvas / Result Area -->
      <div
        class="flex-1 glass-panel rounded-2xl overflow-hidden relative flex flex-col items-center justify-center min-h-[400px]"
      >
        <!-- Batch progress -->
        <div
          v-if="activeBatchProgress.total > 1"
          class="mb-4 w-full flex-shrink-0 grad-border bg-surface-container p-3 rounded-xl flex flex-col gap-2"
        >
          <div class="flex items-center justify-between">
            <span class="font-mono text-micro text-on-surface-variant uppercase">本次批量进度</span>
            <span class="font-mono text-micro text-primary">{{ batchProgressPct }}%</span>
          </div>
          <div class="w-full h-1.5 bg-surface-variant rounded-full overflow-hidden">
            <div
              class="h-full rounded-full bg-gradient-to-r from-secondary via-tertiary to-primary transition-all duration-500"
              :style="{ width: `${batchProgressPct}%` }"
            ></div>
          </div>
          <p class="font-mono text-[11px] text-on-surface-variant text-right">
            已完成 {{ activeBatchProgress.done }} / {{ activeBatchProgress.total }} 张
          </p>
        </div>

        <!-- Background grid for Canvas Area -->
        <div
          class="absolute inset-0 opacity-[0.5]"
          style="
            background-image: linear-gradient(rgba(56, 232, 255, 0.04) 1px, transparent 1px),
              linear-gradient(90deg, rgba(56, 232, 255, 0.04) 1px, transparent 1px);
            background-size: 28px 28px;
          "
        ></div>

        <!-- Source preview / mask editor. In image/edit modes this can be either
             an uploaded reference or the currently completed generation. -->
        <div
          v-if="effectiveSourcePreview"
          class="relative z-10 w-full h-full flex flex-col items-center justify-center p-6 gap-4"
        >
          <div
            class="relative max-w-full max-h-[min(58vh,620px)] rounded-2xl overflow-hidden border border-primary/20 shadow-[0_0_60px_rgba(56,232,255,0.16)]"
          >
            <img
              :src="effectiveSourcePreview"
              alt="参考图预览"
              class="block max-w-full max-h-[min(58vh,620px)] object-contain"
              @load="onSourceImageLoad"
            />
            <canvas
              v-if="mode === 'edit'"
              ref="maskCanvas"
              class="absolute inset-0 w-full h-full touch-none cursor-crosshair"
              @pointerdown="startPainting"
              @pointermove="continuePainting"
              @pointerup="stopPainting"
              @pointerleave="stopPainting"
            ></canvas>
          </div>
          <div class="text-center">
            <p class="font-display text-lg font-bold text-on-surface">
              {{ mode === 'edit' ? '涂抹要修图的区域' : '参考图已锁定' }}
            </p>
            <p class="font-mono text-[11px] text-on-surface-variant mt-1 uppercase tracking-wider">
              {{
                sourceFile
                  ? mode === 'edit'
                    ? 'UPLOADED MASK PAINT · ALPHA EXPORT'
                    : 'UPLOADED REFERENCE · IMAGE-TO-IMAGE'
                  : mode === 'edit'
                    ? 'CURRENT RESULT · MASK PAINT'
                    : 'CURRENT RESULT · REMIX'
              }}
            </p>
          </div>
        </div>

        <!-- Synthesizing state (signature energy moment) -->
        <div
          v-else-if="isBusy"
          class="absolute inset-0 flex flex-col items-center justify-center z-10 bg-surface-container-lowest/85 backdrop-blur-sm overflow-hidden"
        >
          <!-- scanline -->
          <div
            class="pointer-events-none absolute inset-x-0 h-24 bg-gradient-to-b from-transparent via-primary/10 to-transparent animate-scanline"
          ></div>
          <div class="relative w-28 h-28 mb-6">
            <div class="absolute inset-0 border-2 border-surface-variant rounded-full"></div>
            <div
              class="absolute inset-0 border-2 rounded-full border-t-primary border-r-tertiary border-b-secondary border-l-transparent animate-spin-slow shadow-[0_0_24px_rgba(56,232,255,0.3)]"
            ></div>
            <div class="absolute inset-2 border border-outline-variant/30 rounded-full"></div>
            <div class="absolute inset-0 flex items-center justify-center">
              <span class="material-symbols-outlined text-primary text-3xl animate-pulse">blur_on</span>
            </div>
          </div>
          <h3 class="font-display text-xl font-bold text-on-surface mb-2">正在合成画面</h3>
          <p class="font-mono text-[12px] text-on-surface-variant max-w-md text-center uppercase tracking-wider">
            RENDERING VOLUMETRIC LIGHT · CRYSTAL STRUCTURE
          </p>
          <div class="w-72 h-1.5 bg-surface-variant rounded-full mt-6 overflow-hidden">
            <div
              class="h-full rounded-full bg-gradient-to-r from-secondary via-tertiary to-primary bg-[length:200%_auto] animate-gradient-pan transition-all duration-700"
              :style="{ width: progressPct + '%' }"
            ></div>
          </div>
          <p class="font-mono text-micro text-primary mt-2">{{ progressPct }}%</p>
        </div>

        <!-- Completed image -->
        <img
          v-else-if="active && active.status === 'completed' && active.image_url"
          :src="active.image_url"
          :alt="active.prompt"
          class="relative z-10 max-w-full max-h-full object-contain rounded-lg shadow-[0_0_60px_rgba(56,232,255,0.15)]"
        />

        <!-- Failed state -->
        <div
          v-else-if="active && active.status === 'failed'"
          class="relative z-10 flex flex-col items-center justify-center text-center px-6"
        >
          <span class="material-symbols-outlined text-error text-5xl mb-4">error</span>
          <h3 class="font-display text-xl font-bold text-on-surface mb-2">合成失败</h3>
          <p class="font-body-md text-body-md text-on-surface-variant max-w-md">
            {{ active.error || '渲染未能完成，积分已退还。' }}
          </p>
        </div>

        <!-- Idle placeholder -->
        <div v-else class="relative z-10 flex flex-col items-center justify-center text-center px-6">
          <div
            class="relative w-20 h-20 rounded-2xl grad-border bg-surface-container flex items-center justify-center mb-6 glow-shadow"
          >
            <span class="material-symbols-outlined text-primary text-4xl">auto_awesome</span>
          </div>
          <h3 class="font-display text-xl font-bold text-on-surface mb-2">画布就绪</h3>
          <p class="font-body-md text-body-md text-on-surface-variant max-w-md">
            配置参数、描述你的画面，然后开始合成。
          </p>
        </div>
      </div>
    </section>
  </div>
</template>

