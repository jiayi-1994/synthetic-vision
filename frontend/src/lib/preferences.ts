import type { AspectRatio, GenerationMode, Resolution } from '@/types'

const STORAGE_KEY = 'sv_workspace_preferences'

const modes: GenerationMode[] = ['text', 'image', 'edit']
const resolutions: Resolution[] = ['1K', '2K', '4K']
const aspects: AspectRatio[] = ['1:1', '4:3', '16:9', '9:16']

export interface WorkspacePreferences {
  defaultMode: GenerationMode
  defaultResolution: Resolution
  defaultAspectRatio: AspectRatio
  defaultStyle: string
  compactGallery: boolean
  emailDigest: boolean
}

export const defaultWorkspacePreferences: WorkspacePreferences = {
  defaultMode: 'text',
  defaultResolution: '2K',
  defaultAspectRatio: '1:1',
  defaultStyle: 'Cinematic',
  compactGallery: false,
  emailDigest: false,
}

function isGenerationMode(value: unknown): value is GenerationMode {
  return typeof value === 'string' && modes.includes(value as GenerationMode)
}

function isResolution(value: unknown): value is Resolution {
  return typeof value === 'string' && resolutions.includes(value as Resolution)
}

function isAspectRatio(value: unknown): value is AspectRatio {
  return typeof value === 'string' && aspects.includes(value as AspectRatio)
}

export function loadWorkspacePreferences(): WorkspacePreferences {
  if (typeof localStorage === 'undefined') {
    return { ...defaultWorkspacePreferences }
  }

  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return { ...defaultWorkspacePreferences }
    }
    const parsed = JSON.parse(raw) as Partial<WorkspacePreferences>
    return {
      defaultMode: isGenerationMode(parsed.defaultMode)
        ? parsed.defaultMode
        : defaultWorkspacePreferences.defaultMode,
      defaultResolution: isResolution(parsed.defaultResolution)
        ? parsed.defaultResolution
        : defaultWorkspacePreferences.defaultResolution,
      defaultAspectRatio: isAspectRatio(parsed.defaultAspectRatio)
        ? parsed.defaultAspectRatio
        : defaultWorkspacePreferences.defaultAspectRatio,
      defaultStyle:
        typeof parsed.defaultStyle === 'string' && parsed.defaultStyle.trim()
          ? parsed.defaultStyle.trim()
          : defaultWorkspacePreferences.defaultStyle,
      compactGallery: parsed.compactGallery === true,
      emailDigest: parsed.emailDigest === true,
    }
  } catch {
    return { ...defaultWorkspacePreferences }
  }
}

export function saveWorkspacePreferences(preferences: WorkspacePreferences) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(preferences))
}
