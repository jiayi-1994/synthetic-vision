import type { Resolution, AspectRatio } from '../types'

/**
 * Human-friendly relative time: "just now", "2 mins ago", "5 hours ago", "1 day ago".
 */
export function relativeTime(iso: string): string {
  if (!iso) return ''
  const then = new Date(iso).getTime()
  if (Number.isNaN(then)) return ''
  const diff = Math.max(0, Date.now() - then)
  const sec = Math.floor(diff / 1000)
  if (sec < 45) return '刚刚'
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min} 分钟前`
  const hr = Math.floor(min / 60)
  if (hr < 24) return `${hr} 小时前`
  const day = Math.floor(hr / 24)
  if (day < 30) return `${day} 天前`
  const mon = Math.floor(day / 30)
  if (mon < 12) return `${mon} 个月前`
  const yr = Math.floor(mon / 12)
  return `${yr} 年前`
}

/**
 * Deterministic avatar via DiceBear bottts-neutral.
 */
export function avatarUrl(seed: string): string {
  const s = encodeURIComponent(seed || 'synthetic')
  return `https://api.dicebear.com/9.x/bottts-neutral/svg?seed=${s}`
}

export const COSTS: Record<Resolution, number> = {
  '1K': 5,
  '2K': 15,
  '4K': 40,
}

export const RESOLUTIONS: Resolution[] = ['1K', '2K']

export const ASPECTS: { id: AspectRatio; label: string; w: number; h: number }[] = [
  { id: '1:1', label: '1:1 正方形', w: 8, h: 8 },
  { id: '4:3', label: '4:3 标准', w: 8, h: 6 },
  { id: '16:9', label: '16:9 宽屏', w: 10, h: 6 },
  { id: '9:16', label: '9:16 竖屏', w: 6, h: 10 },
]
