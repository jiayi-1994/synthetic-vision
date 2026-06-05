# DESIGN.md — Synthetic Vision · "Aurora Grid" (Cyber Neon)

> 设计方向：**赛博霓虹 / Cyber Neon**。深黑底 + 青/品红霓虹流体，发光边、网格透视、生成式 AI 未来感。
> 交互档位：**L3 沉浸体验**。WebGL 极光、光标光斑跟随、巨型标题 mask reveal、磁吸 CTA、SpotlightCard 阵列、click-spark。
> 适用页面：Login / Dashboard(工作台) / Gallery(作品库)。其余视图（Admin/Marketplace/Analytics）通过同一套 token 自动继承霓虹皮肤。
> 落地约束：Vue 3 + TS + Tailwind 3。所有颜色经 CSS 变量 / tailwind token 引用，零硬编码 hex。

---

## 1. Visual Theme & Atmosphere

**设计哲学**：一台正在思考的机器。深空黑底是"未渲染的虚空"，霓虹流体是"涌现的视觉"。每个发光元素都像通了电——边缘漏光、网格在脚下透视延伸、光标所到之处空间被点亮。克制留白 + 高对比 + 精确网格，避免廉价"赛博朋克贴纸感"。

**氛围关键词**：electric · volumetric · precise · alive · glass-on-void

**一句话定调**：*在黑色虚空里，用青与品红的光绘制尚未存在的图像。*

- 底色趋近纯黑（`#05060f`），不是灰。让霓虹有真正的对比落差。
- 主体是 **glass-on-void**：半透明深色玻璃面板悬浮在发光网格之上。
- 强调=**光**，不是填充色块。用 glow（box-shadow 扩散）和 gradient-border 传达"通电"。
- 动效是氛围的一部分：背景极光缓慢流动、光标拖出光斑、卡片在光标下被点亮。静止画面也要"有电"。

---

## 2. Color Palette & Roles

霓虹双主色：**Cyan（青）** 为主操作 / 焦点，**Magenta（品红）** 为次强调 / 能量。两者构成签名线性渐变。

```css
:root {
  /* —— Void / surfaces —— */
  --void:               #05060f;   /* 5 6 15      页面最底 */
  --void-rgb:           5 6 15;
  --bg:                 #070912;   /* 7 9 18      app 背景 */
  --surface-1:          #0b0e1c;   /* 11 14 28    低层面板 */
  --surface-2:          #11152a;   /* 17 21 42    面板 */
  --surface-3:          #171c38;   /* 23 28 56    高层/hover */
  --surface-glass:      17 22 46;  /* rgba 用，glass 面板 */

  /* —— Neon primary: Cyan —— */
  --cyan:               #38e8ff;   /* 56 232 255  主操作/焦点 */
  --cyan-rgb:           56 232 255;
  --cyan-deep:          #0bb5d6;   /* 实心按钮底 */
  --on-cyan:            #00131a;   /* 青底上的字（近黑） */

  /* —— Neon secondary: Magenta —— */
  --magenta:            #ff3df0;   /* 255 61 240  次强调/能量 */
  --magenta-rgb:        255 61 240;
  --magenta-deep:       #c41fb8;
  --on-magenta:         #1a0019;

  /* —— Neon tertiary: Violet (渐变中段) —— */
  --violet:             #8b5cff;
  --violet-rgb:         139 92 255;

  /* —— Text —— */
  --text-hi:            #eaf2ff;   /* 标题/重点 */
  --text:               #c4cce0;   /* 正文 */
  --text-dim:           #7e88a6;   /* 次要/label */
  --text-faint:         #4a5270;   /* 占位/禁用 */

  /* —— Lines —— */
  --line:               rgba(120 140 200 / 0.14);  /* 默认细边 */
  --line-strong:        rgba(120 140 200 / 0.28);
  --grid:               rgba(56 232 255 / 0.06);    /* 透视网格线 */

  /* —— Status —— */
  --success:            #3dffb0;
  --warning:            #ffce4d;
  --error:              #ff5d6c;
  --error-rgb:          255 93 108;
  --on-error:           #1a0004;

  /* —— Signature gradients —— */
  --grad-neon:   linear-gradient(110deg, var(--cyan) 0%, var(--violet) 50%, var(--magenta) 100%);
  --grad-text:   linear-gradient(110deg, #9bf6ff 0%, #c7b3ff 45%, #ff8df5 100%);
  --glow-cyan:   0 0 0 1px rgba(56 232 255 / .35), 0 8px 40px rgba(56 232 255 / .18);
  --glow-magenta:0 0 0 1px rgba(255 61 240 / .30), 0 8px 40px rgba(255 61 240 / .16);
}
```

**Tailwind token 映射**（保留现有语义名，只换值 → 全站统一切换、不破坏未改造视图）：

| token 名 | 新值 | 角色 |
|---|---|---|
| `background` / `surface` / `surface-dim` | `#070912` | 页面背景 |
| `surface-container-lowest` | `#05060f` | 输入/最底 |
| `surface-container-low` | `#0b0e1c` | 侧栏/卡底 |
| `surface-container` | `#11152a` | 面板 |
| `surface-container-high` | `#171c38` | hover/高层 |
| `surface-container-highest` / `surface-variant` | `#1d2342` | chip/stat |
| `surface-bright` | `#222a4d` | — |
| `primary` | `#38e8ff` (cyan) | 主操作/焦点 |
| `primary-container` | `#0bb5d6` | 实心主按钮底 |
| `on-primary` / `on-primary-container` | `#00131a` | 青底上字 |
| `secondary` | `#ff3df0` (magenta) | 次强调/能量/脉冲 |
| `secondary-container` | `#c41fb8` | — |
| `on-secondary` / `on-secondary-container` | `#1a0019` | — |
| `tertiary` | `#8b5cff` (violet) | 渐变中段/标签 |
| `on-surface` | `#eaf2ff` | 标题 |
| `on-background` | `#c4cce0` | 正文 |
| `on-surface-variant` | `#7e88a6` | 次要/label |
| `outline` | `#7e88a6` | — |
| `outline-variant` | `#39406a` | 细边基色 |
| `error` | `#ff5d6c` / `error-container` `#5a0a14` / `on-error-container` `#ffd9dd` | 错误 |

> 注：上表"现有语义名换值"使未触及的 Admin/Marketplace/Analytics 三视图自动获得霓虹皮肤，保持全站一致。

---

## 3. Typography Rules

**字体族**（全部 self-host，`@fontsource`，无外部 Google Fonts；中英混排 CJK fallback）：

- **Display / 标题**：`Space Grotesk`（600/700）— 几何无衬线，科技克制，霓虹标配。
- **Body / 正文**：`Inter`（400/600）。
- **Mono / 标签·数字·代码·参数**：`JetBrains Mono`（500）— 霓虹界面的"机器感"靠 mono label 撑。

```
import '@fontsource/space-grotesk/500.css'
import '@fontsource/space-grotesk/600.css'
import '@fontsource/space-grotesk/700.css'
import '@fontsource/inter/400.css'
import '@fontsource/inter/600.css'
import '@fontsource/jetbrains-mono/500.css'
```

```css
--font-display: 'Space Grotesk','PingFang SC','Microsoft YaHei','Noto Sans SC',sans-serif;
--font-body:    'Inter','PingFang SC','Microsoft YaHei','Noto Sans SC',sans-serif;
--font-mono:    'JetBrains Mono','PingFang SC','Microsoft YaHei',monospace;
```

**字号层级**：

| 级别 | size / line / weight / tracking | 用途 |
|---|---|---|
| display-xl | 64px / 1.04 / 700 / -0.03em | Login 巨型 hero（桌面） |
| display-lg | 44px / 1.08 / 700 / -0.02em | 页面主标题 |
| headline | 30px / 1.15 / 600 / -0.01em | section H2 |
| headline-mobile | 22px / 1.2 / 600 | 移动标题 |
| title | 18px / 1.3 / 600 | 卡片/面板标题 |
| body | 15px / 1.65 / 400 | 正文（中文 ≥15px） |
| label | 12px / 1.0 / 500 / 0.14em / **uppercase** / mono | eyebrow / 参数标签 / 数据 |
| micro | 10px / 1.0 / 500 / 0.18em / uppercase / mono | 角标/状态码 |

**中文规则**：标题用 Space Grotesk + CJK fallback；正文行高 ≥1.7、`letter-spacing:.02em`；mono label 的 uppercase/tracking 只对拉丁字符有意义，中文 label 不强制 uppercase。

**禁止字体**：系统默认 serif、Arial/Times、任何 cursive、Comic Sans、Orbitron / 像素体（廉价赛博感）。

---

## 4. Component Stylings（含全部状态）

### 4.1 Glass panel（基底容器）
```css
.glass {
  background: rgba(var(--surface-glass) / 0.55);
  backdrop-filter: blur(14px);            /* ≤14，性能红线 */
  -webkit-backdrop-filter: blur(14px);
  border: 1px solid var(--line);
  border-radius: 16px;
  position: relative;
}
/* 顶部漏光 rim */
.glass::before {
  content:''; position:absolute; inset:0 0 auto 0; height:1px; border-radius:16px 16px 0 0;
  background: linear-gradient(90deg, transparent, rgba(var(--cyan-rgb)/.5), transparent);
  opacity:.6; pointer-events:none;
}
.glass:hover { border-color: var(--line-strong); }
```

### 4.2 Primary button（青·磁吸）
```css
.btn-primary{
  font-family:var(--font-display); font-weight:600; letter-spacing:.01em;
  color:var(--on-cyan); background:var(--cyan-deep);
  border:1px solid rgba(var(--cyan-rgb)/.6); border-radius:12px;
  padding:.7rem 1.4rem; position:relative; isolation:isolate;
  box-shadow:var(--glow-cyan); transition:transform .18s, box-shadow .25s, filter .2s;
}
.btn-primary:hover{ filter:brightness(1.12); box-shadow:0 0 0 1px rgba(var(--cyan-rgb)/.6),0 10px 48px rgba(var(--cyan-rgb)/.32); }
.btn-primary:focus-visible{ outline:none; box-shadow:0 0 0 2px var(--void),0 0 0 4px var(--cyan); }
.btn-primary:active{ transform:translateY(1px) scale(.985); }
.btn-primary:disabled{ filter:saturate(.4) brightness(.7); box-shadow:none; cursor:not-allowed; opacity:.6; }
/* 磁吸：JS 给元素加 --mx/--my 位移（≤6px），rAF 节流 */
```

### 4.3 Secondary / ghost button
```css
.btn-ghost{
  font-family:var(--font-display); font-weight:500; color:var(--text-hi);
  background:rgba(var(--surface-glass)/.4); border:1px solid var(--line-strong);
  border-radius:12px; padding:.7rem 1.4rem; transition:.2s;
}
.btn-ghost:hover{ border-color:rgba(var(--cyan-rgb)/.5); color:var(--cyan); background:rgba(var(--cyan-rgb)/.06); }
.btn-ghost:focus-visible{ outline:none; box-shadow:0 0 0 2px var(--void),0 0 0 4px rgba(var(--cyan-rgb)/.6); }
.btn-ghost:active{ transform:scale(.985); }
.btn-ghost:disabled{ opacity:.45; cursor:not-allowed; }
```

### 4.4 SpotlightCard（Gallery / Dashboard 阵列签名件）
```css
.spot{ position:relative; border-radius:16px; border:1px solid var(--line);
       background:rgba(var(--surface-glass)/.5); overflow:hidden; transition:border-color .25s, transform .25s; }
.spot::before{                                   /* 光标跟随光斑 */
  content:''; position:absolute; inset:0; pointer-events:none; opacity:0; transition:opacity .25s;
  background:radial-gradient(220px circle at var(--mx,50%) var(--my,50%), rgba(var(--cyan-rgb)/.16), transparent 60%);
}
.spot:hover{ border-color:rgba(var(--cyan-rgb)/.35); transform:translateY(-3px); }
.spot:hover::before{ opacity:1; }
```

### 4.5 Input / textarea
```css
.field{
  width:100%; font-family:var(--font-body); color:var(--text-hi);
  background:var(--surface-1); border:1px solid var(--line-strong); border-radius:12px;
  padding:.75rem 1rem; transition:border-color .2s, box-shadow .2s;
}
.field::placeholder{ color:var(--text-faint); }
.field:hover{ border-color:rgba(var(--cyan-rgb)/.3); }
.field:focus{ outline:none; border-color:rgba(var(--cyan-rgb)/.6); box-shadow:0 0 0 3px rgba(var(--cyan-rgb)/.14); }
.field:disabled{ opacity:.5; cursor:not-allowed; }
.field[aria-invalid=true]{ border-color:rgba(var(--error-rgb)/.6); box-shadow:0 0 0 3px rgba(var(--error-rgb)/.14); }
```

### 4.6 Nav link（侧栏）
```css
.nav{ display:flex; gap:.75rem; align-items:center; padding:.7rem 1rem; border-radius:12px;
      font-family:var(--font-display); color:var(--text-dim); transition:.2s; position:relative; }
.nav:hover{ color:var(--cyan); background:rgba(var(--cyan-rgb)/.06); }
.nav:focus-visible{ outline:none; box-shadow:0 0 0 2px var(--void),0 0 0 4px rgba(var(--cyan-rgb)/.5); }
.nav--active{ color:var(--text-hi); background:rgba(var(--cyan-rgb)/.10); }
.nav--active::before{ content:''; position:absolute; left:0; top:18%; bottom:18%; width:3px; border-radius:3px;
  background:var(--grad-neon); box-shadow:0 0 12px rgba(var(--cyan-rgb)/.7); }
```

### 4.7 Chip / Tag / Credits pill
```css
.chip{ font-family:var(--font-mono); font-size:11px; letter-spacing:.1em; text-transform:uppercase;
       color:var(--text-dim); background:var(--surface-2); border:1px solid var(--line); border-radius:999px; padding:.3rem .7rem; }
.chip--on{ color:var(--cyan); border-color:rgba(var(--cyan-rgb)/.4); background:rgba(var(--cyan-rgb)/.08); }
.pill-credits{ /* secondary 脉冲点 + mono 数字 */ }
```

### 4.8 Gradient-border（强调容器 / hero 卡）
```css
.grad-border{ position:relative; border-radius:18px; background:var(--surface-1); }
.grad-border::after{ content:''; position:absolute; inset:0; border-radius:18px; padding:1px;
  background:var(--grad-neon); -webkit-mask:linear-gradient(#000 0 0) content-box,linear-gradient(#000 0 0);
  -webkit-mask-composite:xor; mask-composite:exclude; opacity:.6; pointer-events:none; }
```

---

## 5. Layout Principles

- **网格**：12 列，`gutter 24px`，容器 `max-width 1440px`。Gallery 用 **Bento 不等大**：`grid-auto-rows` + 个别卡 `row-span-2 / col-span-2`（每 5-6 张出现一张大卡）。
- **间距梯度**（4px 基）：4 / 8 / 12 / 16 / 24 / 40 / 64。
- **AppShell**：左固定侧栏 `w-72`，顶栏 `h-16`，内容区滚动。移动端侧栏隐藏。
- **透视网格地面**：背景层用 `linear-gradient` 画 1px 网格 + 远端淡出，制造"站在数据平面上"的纵深（性能=纯渐变，零成本）。
- **节奏**：Login 单屏居中聚焦；Dashboard 双栏（左 320 参数 / 右 画布）；Gallery 头部 profile + Bento 阵列。

---

## 6. Depth & Elevation

光即层级——越高层，glow 越亮、blur 越实。

```css
--e0: none;                                             /* 贴地网格 */
--e1: 0 2px 12px rgba(0 0 0 / .4);                      /* 面板基底 */
--e2: 0 8px 32px rgba(0 0 0 / .5);                      /* 浮起面板 */
--e3: 0 16px 60px rgba(0 0 0 / .6);                     /* 模态/活跃画布 */
--glow-focus: 0 0 0 1px rgba(56 232 255/.5), 0 8px 40px rgba(56 232 255/.22);  /* 焦点元素 */
--glow-energy:0 0 0 1px rgba(255 61 240/.45),0 8px 40px rgba(255 61 240/.20);  /* 能量/合成中 */
```
阴影一律冷调（黑 + 霓虹漏光），禁止暖色/纯黑硬投影。

---

## 7. Animation & Interaction（L3）

**依赖**：零第三方动画库。WebGL 用原生 GLSL（无 three.js，离线友好）；其余纯 CSS + IntersectionObserver + rAF。`prefers-reduced-motion` 全量降级。

### 7.1 入场（load）
- Hero H1 **mask reveal**：`clip-path` inset 由 100%→0 + 字体渐变流动，一次性，0.9s ease-out。
- 面板/卡 **stagger fade-up**：`translateY(16px)+opacity` 依次 60ms 错峰，CSS 一次性。

### 7.2 滚动（Gallery / 长内容）
- `v-reveal` 指令：IntersectionObserver 进入视口 → `is-in` → fade-up + blur(6px→0)。
- Section H2 滚动触发：字符级 blur reveal（CSS 变量延迟）。

### 7.3 悬停
- SpotlightCard：`--mx/--my` 跟随光标（rAF 节流），`::before` radial 点亮 + `translateY(-3px)`。
- 图片卡：`scale(1.05)` 700ms + 底部信息上滑浮现。
- Magnet CTA：指针在按钮附近时，按钮向指针偏移 ≤6px（rAF 节流），离开归位。

### 7.4 签名特效（signature moments，L3 ≥6）
1. **WebGL 极光背景**（Login 全屏 + 全局极淡）：GLSL fragment shader 流动青/品红/紫极光。`IntersectionObserver` 不可见暂停；移动端降 DPR；reduced-motion → 静态渐变快照。**全页仅 1 个 WebGL scene**。
2. **巨型标题 mask reveal + 渐变流动**（Login "SYNTHETIC VISION"）。
3. **全局光标光斑**：`pointermove`(rAF) → 根 `--mx/--my` → 一层固定 radial-gradient 跟随（`hover:hover` 才启用）。
4. **磁吸 CTA**（生成 / 初始化会话）。
5. **SpotlightCard 阵列**（Gallery 作品 / Dashboard 参数块）。
6. **click-spark**：点击主按钮迸发 6 条青色火花（一次性 CSS，结束即移除）。
7. **合成中能量态**：画布旋转霓虹光环 + 扫描线 + 渐变进度条（magenta→cyan 流动），脉冲。

> **关于 pin-scrub / 左pin右swap**：本产品是 **app 工作台**（固定高度的功能屏，非营销长页），强行 ScrollTrigger pin 会破坏可用性。L3 在此**以 WebGL 签名 + 光标驱动空间 + mask reveal + 磁吸**替代滚动叙事——这是针对 app shell 的有意取舍，已在此记录。Gallery 作为唯一长滚动页承载 `v-reveal` + Bento 节奏。

### 7.5 reduced-motion 降级
```css
@media (prefers-reduced-motion: reduce){
  *,*::before,*::after{ animation-duration:.001ms!important; animation-iteration-count:1!important;
    transition-duration:.001ms!important; scroll-behavior:auto!important; }
  /* WebGL canvas 停渲 → 显示静态渐变快照；光标光斑/磁吸禁用 */
}
```

---

## 8. Do's and Don'ts

**Do**
1. 强调一律用**光**（glow / gradient-border / 漏光 rim），而非大面积实心霓虹填充。
2. 霓虹只点缀焦点、能量、数据；大面积保持深空黑 + 玻璃。
3. mono 字体承载所有 label / 数字 / 参数 / 状态码——这是霓虹界面"机器感"来源。
4. 每 1-2 屏（Gallery）至少一个 signature moment；静屏也要"通电"（rim 光 / 网格）。
5. 光标交互仅在 `matchMedia('(hover:hover)')` 启用。
6. 所有动效给 `prefers-reduced-motion` 完整降级。
7. 颜色全部走 CSS 变量 / tailwind token。
8. 渐变文字只用在 ≤2 处巨型 hero（见 text-decoration 规则），正文永不渐变。

**Don'ts**
1. ❌ 大面积纯霓虹背景填充（刺眼、廉价、毁对比）。
2. ❌ `filter:blur()` 加在移动/动画元素上（性能红线）；景深用 opacity+scale。
3. ❌ `backdrop-filter` blur > 14px，或覆盖大滚动区。
4. ❌ 多个 WebGL / Three.js 常驻渲染；离开视口不暂停。
5. ❌ Orbitron/像素体/霓虹描边字 + 紫色霓虹故障字堆砌（AI slop 赛博感）。
6. ❌ Emoji 当图标（用 Material Symbols / 内联 SVG）。
7. ❌ 纯色块占位图（用 mock 图 / 渐变占位）。
8. ❌ 在 app 工作屏上强加 scroll-jacking / Lenis / 多 pin。
9. ❌ 暖色/硬黑投影（阴影只冷调霓虹漏光）。

---

## 9. Responsive Behavior

- **断点**：`sm 640 / md 768 / lg 1024 / xl 1280`。
- **侧栏**：`< md` 隐藏（`hidden md:flex`），顶栏显示移动品牌。
- **Dashboard**：`< lg` 双栏改纵向堆叠（参数面板在上，画布在下），整页改 `overflow-y-auto`。
- **Gallery Bento**：`1 → 2 → 3 → 4` 列（sm/lg/xl）；移动端取消 col/row-span，回退等大。
- **Login hero**：display-xl 64px → 移动 40px；WebGL 降 DPR(1)。
- **触摸目标** ≥ 44×44px；移动端 ≤600px **无横向溢出**。
- **性能降级**：移动端 WebGL → 静态渐变；光标光斑/磁吸在无 hover 设备禁用。

---

*Motion effects derived from [vue-bits](https://github.com/DavidHDev/vue-bits) by DavidHDev (MIT): SpotlightCard, Magnet, ClickSpark, ScrollReveal, GradientText 概念。WebGL aurora 为本项目原生 GLSL 实现。*
