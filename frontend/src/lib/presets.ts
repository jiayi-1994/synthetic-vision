import type { Preset, PresetCategory } from '@/types'

export const PRESET_CATALOG: Preset[] = [
  {
    id: 'dreamscape-01',
    title: '梦境森林',
    description:
      '营造高雾、高光的超现实森林场景，强调微光颗粒、体积光与深度景深的层次。',
    prompt_seed: 'dreamy bioluminescent forest at midnight with volumetric light shafts and cinematic atmosphere',
    style: 'Cinematic',
    suggested_resolution: '2K',
    suggested_aspect_ratio: '16:9',
    tags: ['photoreal', 'illustration'],
    estimated_cost: 15,
    preview: '超现实森林',
  },
  {
    id: 'urban-neon-02',
    title: '赛博城市',
    description: '以夜色霓虹和湿润反射打造赛博都市远景，适合氛围海报与封面。',
    prompt_seed: 'cyberpunk city at night, neon reflections on wet streets, dramatic rain, high detail',
    style: 'Cyber',
    suggested_resolution: '2K',
    suggested_aspect_ratio: '16:9',
    tags: ['illustration', 'photoreal'],
    estimated_cost: 15,
    preview: '霓虹质感',
  },
  {
    id: 'portrait-editorial-03',
    title: '人像影棚',
    description: '偏纪实的人像照明，浅景深 + 柔光勾勒，适合展示人物肖像系列。',
    prompt_seed: 'editorial portrait of a graceful figure in soft rim light, shallow depth of field, cinematic realism',
    style: 'Portrait',
    suggested_resolution: '1K',
    suggested_aspect_ratio: '4:3',
    tags: ['portrait', 'photoreal'],
    estimated_cost: 5,
    preview: '人物棚拍',
  },
  {
    id: 'abstract-fluid-04',
    title: '流体抽象',
    description: '液态几何与发光纹理交错的抽象画风，强调动感笔触和色彩对撞。',
    prompt_seed: 'abstract fluid geometry composition with glowing gradients and energetic brush strokes',
    style: 'Abstract',
    suggested_resolution: '1K',
    suggested_aspect_ratio: '1:1',
    tags: ['abstract', 'illustration'],
    estimated_cost: 5,
    preview: '流动纹理',
  },
  {
    id: 'product-showcase-05',
    title: '产品展示',
    description: '用于商品海报/详情图的纯净中性场景，减少杂物并强化主件质感。',
    prompt_seed:
      'product shot of a premium gadget on a clean neutral background, soft studio lighting, reflective surfaces',
    style: 'Photographic',
    suggested_resolution: '2K',
    suggested_aspect_ratio: '1:1',
    tags: ['product', 'photoreal'],
    estimated_cost: 15,
    preview: '商品主图',
  },
  {
    id: 'retro-console-06',
    title: '复古游戏',
    description: '复古像素与电影级调色的混搭，用于游戏界面、插画素材或角色切换页。',
    prompt_seed: 'retro arcade game scene, pixel-art inspired atmosphere with cinematic glow and clean edges',
    style: 'Retro',
    suggested_resolution: '1K',
    suggested_aspect_ratio: '16:9',
    tags: ['retro', 'illustration'],
    estimated_cost: 5,
    preview: '复古像素',
  },
] 

export function getPresetById(id: string): Preset | undefined {
  return PRESET_CATALOG.find((preset) => preset.id === id)
}

export function byCategory(presets: Preset[], category: PresetCategory | 'all'): Preset[] {
  if (category === 'all') return presets
  return presets.filter((preset) => preset.tags.includes(category))
}
