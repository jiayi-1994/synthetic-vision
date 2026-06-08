import { computed, ref } from 'vue'
import {
  defaultWorkspacePreferences,
  loadWorkspacePreferences,
  saveWorkspacePreferences,
  type ThemeId,
} from './preferences'

export interface ThemeOption {
  id: ThemeId
  label: string
  shortLabel: string
  icon: string
  helper: string
  preview: string
}

export const themeOptions: ThemeOption[] = [
  {
    id: 'aurora',
    label: '暗夜霓虹',
    shortLabel: '霓虹',
    icon: 'dark_mode',
    helper: '默认的深色 Aurora Grid，适合长时间创作。',
    preview: 'from-[#05060f] via-[#11152a] to-[#38e8ff]',
  },
  {
    id: 'daybreak',
    label: '晨雾浅色',
    shortLabel: '浅色',
    icon: 'light_mode',
    helper: '浅色玻璃面板，更适合白天展示和截图。',
    preview: 'from-[#f7faff] via-[#eaf1ff] to-[#00a3bf]',
  },
  {
    id: 'contrast',
    label: '高对比',
    shortLabel: '高对比',
    icon: 'contrast',
    helper: '提升文字和边界对比，适合强光环境。',
    preview: 'from-[#000000] via-[#141414] to-[#00f5ff]',
  },
]

const currentTheme = ref<ThemeId>(defaultWorkspacePreferences.theme)

function normalizeTheme(theme: ThemeId | undefined): ThemeId {
  return themeOptions.some((item) => item.id === theme) ? theme : defaultWorkspacePreferences.theme
}

export function applyTheme(theme: ThemeId) {
  if (typeof document === 'undefined') return
  const next = normalizeTheme(theme)
  document.documentElement.dataset.theme = next
  document.documentElement.classList.toggle('dark', next !== 'daybreak')
}

export function previewTheme(theme: ThemeId) {
  const next = normalizeTheme(theme)
  currentTheme.value = next
  applyTheme(next)
}

export function setTheme(theme: ThemeId) {
  const next = normalizeTheme(theme)
  previewTheme(next)
  saveWorkspacePreferences({
    ...loadWorkspacePreferences(),
    theme: next,
  })
}

export function cycleTheme() {
  const index = themeOptions.findIndex((item) => item.id === currentTheme.value)
  const next = themeOptions[(index + 1) % themeOptions.length]
  setTheme(next.id)
}

export function initTheme() {
  const storedTheme = normalizeTheme(loadWorkspacePreferences().theme)
  previewTheme(storedTheme)
}

export function useTheme() {
  const activeTheme = computed(
    () => themeOptions.find((item) => item.id === currentTheme.value) ?? themeOptions[0]
  )

  return {
    activeTheme,
    currentTheme,
    themeOptions,
    previewTheme,
    setTheme,
    cycleTheme,
  }
}
