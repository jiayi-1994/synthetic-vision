/** @type {import('tailwindcss').Config} */
// "Aurora Grid" — Cyber Neon. Semantic token NAMES kept stable so every view
// (incl. untouched Admin/Marketplace/Analytics) inherits the neon skin; only
// the VALUES changed. Cyan = primary/focus, Magenta = secondary/energy,
// Violet = tertiary (gradient mid). See DESIGN.md.
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{vue,ts}'],
  theme: {
    extend: {
      colors: {
        // —— primary: Cyan ——
        primary: 'rgb(var(--cyan-rgb) / <alpha-value>)',
        'on-primary': 'rgb(var(--on-cyan-rgb) / <alpha-value>)',
        'primary-container': 'rgb(var(--cyan-deep-rgb) / <alpha-value>)',
        'on-primary-container': 'rgb(var(--on-cyan-rgb) / <alpha-value>)',
        'primary-fixed': 'rgb(var(--cyan-rgb) / <alpha-value>)',
        'primary-fixed-dim': 'rgb(var(--cyan-rgb) / <alpha-value>)',
        'on-primary-fixed': 'rgb(var(--on-cyan-rgb) / <alpha-value>)',
        'on-primary-fixed-variant': 'rgb(var(--cyan-deep-rgb) / <alpha-value>)',
        'inverse-primary': 'rgb(var(--cyan-deep-rgb) / <alpha-value>)',
        'surface-tint': 'rgb(var(--cyan-rgb) / <alpha-value>)',
        // —— secondary: Magenta ——
        secondary: 'rgb(var(--magenta-rgb) / <alpha-value>)',
        'on-secondary': 'rgb(var(--on-magenta-rgb) / <alpha-value>)',
        'secondary-container': 'rgb(var(--magenta-deep-rgb) / <alpha-value>)',
        'on-secondary-container': 'rgb(var(--on-magenta-rgb) / <alpha-value>)',
        'secondary-fixed': 'rgb(var(--magenta-rgb) / <alpha-value>)',
        'secondary-fixed-dim': 'rgb(var(--magenta-rgb) / <alpha-value>)',
        'on-secondary-fixed': 'rgb(var(--on-magenta-rgb) / <alpha-value>)',
        'on-secondary-fixed-variant': 'rgb(var(--magenta-deep-rgb) / <alpha-value>)',
        // —— tertiary: Violet ——
        tertiary: 'rgb(var(--violet-rgb) / <alpha-value>)',
        'on-tertiary': 'rgb(var(--on-violet-rgb) / <alpha-value>)',
        'tertiary-container': 'rgb(var(--violet-deep-rgb) / <alpha-value>)',
        'on-tertiary-container': 'rgb(var(--on-violet-rgb) / <alpha-value>)',
        'tertiary-fixed': 'rgb(var(--violet-rgb) / <alpha-value>)',
        'tertiary-fixed-dim': 'rgb(var(--violet-rgb) / <alpha-value>)',
        'on-tertiary-fixed': 'rgb(var(--on-violet-rgb) / <alpha-value>)',
        'on-tertiary-fixed-variant': 'rgb(var(--violet-deep-rgb) / <alpha-value>)',
        // —— error ——
        error: 'rgb(var(--error-rgb) / <alpha-value>)',
        'on-error': 'rgb(var(--void-rgb) / <alpha-value>)',
        'error-container': 'rgb(var(--error-rgb) / <alpha-value>)',
        'on-error-container': 'rgb(var(--text-hi-rgb) / <alpha-value>)',
        // —— success / warning (extra) ——
        success: 'rgb(var(--success-rgb) / <alpha-value>)',
        warning: 'rgb(var(--warning-rgb) / <alpha-value>)',
        // —— surfaces / void ——
        background: 'rgb(var(--bg-rgb) / <alpha-value>)',
        'on-background': 'rgb(var(--text-rgb) / <alpha-value>)',
        surface: 'rgb(var(--bg-rgb) / <alpha-value>)',
        'surface-dim': 'rgb(var(--void-rgb) / <alpha-value>)',
        'surface-bright': 'rgb(var(--surface-3-rgb) / <alpha-value>)',
        'on-surface': 'rgb(var(--text-hi-rgb) / <alpha-value>)',
        'on-surface-variant': 'rgb(var(--text-dim-rgb) / <alpha-value>)',
        'surface-container-lowest': 'rgb(var(--void-rgb) / <alpha-value>)',
        'surface-container-low': 'rgb(var(--surface-1-rgb) / <alpha-value>)',
        'surface-container': 'rgb(var(--surface-2-rgb) / <alpha-value>)',
        'surface-container-high': 'rgb(var(--surface-3-rgb) / <alpha-value>)',
        'surface-container-highest': 'rgb(var(--surface-3-rgb) / <alpha-value>)',
        'surface-variant': 'rgb(var(--surface-3-rgb) / <alpha-value>)',
        outline: 'rgb(var(--outline-rgb) / <alpha-value>)',
        'outline-variant': 'rgb(var(--outline-variant-rgb) / <alpha-value>)',
        'inverse-surface': 'rgb(var(--inverse-surface-rgb) / <alpha-value>)',
        'inverse-on-surface': 'rgb(var(--inverse-on-surface-rgb) / <alpha-value>)',
        // —— neon raw (for arbitrary gradients in markup) ——
        void: 'rgb(var(--void-rgb) / <alpha-value>)',
        cyan: 'rgb(var(--cyan-rgb) / <alpha-value>)',
        magenta: 'rgb(var(--magenta-rgb) / <alpha-value>)',
        violet: 'rgb(var(--violet-rgb) / <alpha-value>)',
      },
      borderRadius: {
        DEFAULT: '0.25rem',
        lg: '0.5rem',
        xl: '0.75rem',
        '2xl': '1rem',
        full: '9999px',
      },
      spacing: {
        'margin-lg': '40px',
        'margin-sm': '16px',
        gutter: '24px',
        unit: '4px',
        'container-max': '1440px',
      },
      fontFamily: {
        // Display = Space Grotesk (geometric, techy). Body = Inter. Label = JetBrains Mono.
        // CJK falls through to OS-native faces after the Latin face.
        display: ['Space Grotesk', 'PingFang SC', 'Microsoft YaHei', 'Noto Sans SC', 'sans-serif'],
        'display-lg': ['Space Grotesk', 'PingFang SC', 'Microsoft YaHei', 'Noto Sans SC', 'sans-serif'],
        'headline-lg': ['Space Grotesk', 'PingFang SC', 'Microsoft YaHei', 'Noto Sans SC', 'sans-serif'],
        'headline-lg-mobile': ['Space Grotesk', 'PingFang SC', 'Microsoft YaHei', 'Noto Sans SC', 'sans-serif'],
        'body-md': ['Inter', 'PingFang SC', 'Microsoft YaHei', 'Noto Sans SC', 'sans-serif'],
        'label-sm': ['JetBrains Mono', 'PingFang SC', 'Microsoft YaHei', 'monospace'],
        mono: ['JetBrains Mono', 'PingFang SC', 'Microsoft YaHei', 'monospace'],
      },
      fontSize: {
        'display-xl': ['64px', { lineHeight: '1.04', letterSpacing: '-0.03em', fontWeight: '700' }],
        'display-lg': ['44px', { lineHeight: '1.08', letterSpacing: '-0.02em', fontWeight: '700' }],
        'headline-lg': ['30px', { lineHeight: '1.15', letterSpacing: '-0.01em', fontWeight: '600' }],
        'headline-lg-mobile': ['22px', { lineHeight: '1.2', fontWeight: '600' }],
        title: ['18px', { lineHeight: '1.3', fontWeight: '600' }],
        'body-md': ['15px', { lineHeight: '1.65', fontWeight: '400' }],
        'label-sm': ['12px', { lineHeight: '1.0', letterSpacing: '0.14em', fontWeight: '500' }],
        micro: ['10px', { lineHeight: '1.0', letterSpacing: '0.18em', fontWeight: '500' }],
      },
      keyframes: {
        'gradient-pan': {
          '0%,100%': { backgroundPosition: '0% 50%' },
          '50%': { backgroundPosition: '100% 50%' },
        },
        'rise-in': {
          '0%': { opacity: '0', transform: 'translateY(16px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        'spin-slow': { to: { transform: 'rotate(360deg)' } },
        'pulse-dot': {
          '0%,100%': { opacity: '1', boxShadow: '0 0 0 0 rgba(255,61,240,.5)' },
          '50%': { opacity: '.5', boxShadow: '0 0 0 6px rgba(255,61,240,0)' },
        },
        scanline: { '0%': { transform: 'translateY(-100%)' }, '100%': { transform: 'translateY(100%)' } },
      },
      animation: {
        'gradient-pan': 'gradient-pan 6s ease infinite',
        'rise-in': 'rise-in .6s cubic-bezier(.22,1,.36,1) both',
        'spin-slow': 'spin-slow 3s linear infinite',
        'pulse-dot': 'pulse-dot 2s ease-in-out infinite',
        scanline: 'scanline 2.4s linear infinite',
      },
    },
  },
  plugins: [],
}
