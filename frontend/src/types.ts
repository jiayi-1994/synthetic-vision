export interface User {
  id: string
  public_id: string
  username: string
  email: string
  role: 'user' | 'admin'
  plan: string
  credits: number
  avatar_seed: string
  created_at: string
}

export type GenStatus = 'pending' | 'processing' | 'completed' | 'failed'
export type GenerationMode = 'text' | 'image' | 'edit'
export type Resolution = '1K' | '2K' | '4K'
export type AspectRatio = '1:1' | '4:3' | '16:9' | '9:16'

export interface Generation {
  id: string
  mode: GenerationMode
  prompt: string
  negative_prompt: string
  resolution: Resolution
  aspect_ratio: AspectRatio
  style: string
  width: number
  height: number
  seed: number
  status: GenStatus
  cost: number
  image_url: string
  has_source_image: boolean
  has_mask_image: boolean
  error: string
  created_at: string
  completed_at: string | null
}

export interface AdminUser {
  public_id: string
  username: string
  email: string
  credits: number
  role: string
  plan: string
  last_activity_at: string
  initials: string
  created_at: string
}

export interface Stats {
  user: User
  total_generations: number
  completed_generations: number
  credit_balance: number
}

export interface CreateGenInput {
  mode?: GenerationMode
  prompt: string
  negative_prompt?: string
  style?: string
  resolution: Resolution
  aspect_ratio: AspectRatio
  source_image?: File
  mask_image?: Blob | File
}

export interface AnalyticsDistributionItem {
  label: string
  count: number
  percentage: number
}

export interface AnalyticsSummary {
  total_generations: number
  completed_generations: number
  failed_generations: number
  pending_generations: number
  processing_generations: number
  success_rate: number
  credits_spent: number
  credits_refunded: number
  credit_balance: number
}

export interface AnalyticsCreditBreakdown {
  generation_debits: number
  refunds: number
  admin_topups: number
  signup_bonus: number
}

export interface AnalyticsRecentGeneration {
  id: string
  prompt: string
  status: GenStatus
  resolution: Resolution
  aspect_ratio: AspectRatio
  cost: number
  error: string
  created_at: string
  completed_at: string | null
}

export interface AnalyticsResponse {
  user: User
  summary: AnalyticsSummary
  status_distribution: AnalyticsDistributionItem[]
  resolution_distribution: AnalyticsDistributionItem[]
  aspect_ratio_distribution: AnalyticsDistributionItem[]
  credit_breakdown: AnalyticsCreditBreakdown
  recent_generations: AnalyticsRecentGeneration[]
}

export type PresetCategory = 'photoreal' | 'illustration' | 'abstract' | 'product' | 'retro' | 'portrait'

export interface Preset {
  id: string
  title: string
  description: string
  prompt_seed: string
  style: string
  suggested_resolution: Resolution
  suggested_aspect_ratio: AspectRatio
  tags: PresetCategory[]
  estimated_cost: number
  preview: string
}
