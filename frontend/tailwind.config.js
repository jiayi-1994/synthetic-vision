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
        primary: '#38e8ff',
        'on-primary': '#00131a',
        'primary-container': '#0bb5d6',
        'on-primary-container': '#00131a',
        'primary-fixed': '#9bf6ff',
        'primary-fixed-dim': '#38e8ff',
        'on-primary-fixed': '#001016',
        'on-primary-fixed-variant': '#007e96',
        'inverse-primary': '#0bb5d6',
        'surface-tint': '#38e8ff',
        // —— secondary: Magenta ——
        secondary: '#ff3df0',
        'on-secondary': '#1a0019',
        'secondary-container': '#c41fb8',
        'on-secondary-container': '#ffd6f8',
        'secondary-fixed': '#ffaef3',
        'secondary-fixed-dim': '#ff3df0',
        'on-secondary-fixed': '#330030',
        'on-secondary-fixed-variant': '#8f0085',
        // —— tertiary: Violet ——
        tertiary: '#8b5cff',
        'on-tertiary': '#1a0040',
        'tertiary-container': '#5a2ee4',
        'on-tertiary-container': '#e7dcff',
        'tertiary-fixed': '#c7b3ff',
        'tertiary-fixed-dim': '#8b5cff',
        'on-tertiary-fixed': '#13003a',
        'on-tertiary-fixed-variant': '#4500b3',
        // —— error ——
        error: '#ff5d6c',
        'on-error': '#1a0004',
        'error-container': '#5a0a14',
        'on-error-container': '#ffd9dd',
        // —— success / warning (extra) ——
        success: '#3dffb0',
        warning: '#ffce4d',
        // —— surfaces / void ——
        background: '#070912',
        'on-background': '#c4cce0',
        surface: '#070912',
        'surface-dim': '#05060f',
        'surface-bright': '#222a4d',
        'on-surface': '#eaf2ff',
        'on-surface-variant': '#7e88a6',
        'surface-container-lowest': '#05060f',
        'surface-container-low': '#0b0e1c',
        'surface-container': '#11152a',
        'surface-container-high': '#171c38',
        'surface-container-highest': '#1d2342',
        'surface-variant': '#1d2342',
        outline: '#7e88a6',
        'outline-variant': '#39406a',
        'inverse-surface': '#eaf2ff',
        'inverse-on-surface': '#11152a',
        // —— neon raw (for arbitrary gradients in markup) ——
        void: '#05060f',
        cyan: '#38e8ff',
        magenta: '#ff3df0',
        violet: '#8b5cff',
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
