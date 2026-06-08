<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

interface NavItem {
  to: string
  label: string
  icon: string
  admin?: boolean
}

const navItems: NavItem[] = [
  { to: '/', label: '工作台', icon: 'auto_awesome' },
  { to: '/gallery', label: '作品库', icon: 'grid_view' },
  { to: '/marketplace', label: '市场', icon: 'storefront' },
  { to: '/admin', label: '管理后台', icon: 'admin_panel_settings', admin: true },
  { to: '/analytics', label: '数据分析', icon: 'insights' },
]

function goUpgrade() {
  router.push('/settings?section=billing')
}

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <!-- SideNavBar (Desktop Only) -->
  <nav
    class="hidden md:flex bg-surface-container-low/70 backdrop-blur-xl border-r border-outline-variant/20 fixed left-0 top-0 h-full w-72 flex-col p-margin-sm gap-4 z-40"
  >
    <!-- right-edge neon hairline -->
    <div
      class="pointer-events-none absolute inset-y-0 right-0 w-px bg-gradient-to-b from-transparent via-primary/40 to-transparent"
    ></div>

    <div class="flex items-center gap-3 mb-8 px-3 mt-4">
      <div
        class="relative w-10 h-10 rounded-xl grad-border bg-surface-container flex items-center justify-center glow-shadow"
      >
        <span class="material-symbols-outlined text-primary">auto_awesome</span>
      </div>
      <div>
        <h1 class="font-display text-[19px] font-bold text-on-surface tracking-tight leading-none">
          Synthetic <span class="text-neon">Vision</span>
        </h1>
        <p class="font-mono text-micro text-on-surface-variant uppercase mt-1.5">V3.5 · NEURAL ENGINE</p>
      </div>
    </div>

    <div class="flex flex-col gap-1.5 flex-grow">
      <template v-for="item in navItems" :key="item.to">
        <RouterLink
          v-if="!item.admin || auth.isAdmin"
          :to="item.to"
          custom
          v-slot="{ href, navigate, isActive, isExactActive }"
        >
          <!-- Pick exactly one class set: mixing active + inactive text/weight
               classes lets CSS source-order (not intent) decide the winner.
               '/' uses exact-active so it isn't lit on every child route. -->
          <a
            :href="href"
            class="relative flex items-center gap-3 px-4 py-3 rounded-xl font-display group transition-all duration-200"
            :class="
              (item.to === '/' ? isExactActive : isActive)
                ? 'text-on-surface bg-primary/10'
                : 'text-on-surface-variant hover:bg-primary/5 hover:text-primary'
            "
            @click="navigate"
          >
            <span
              v-if="item.to === '/' ? isExactActive : isActive"
              class="absolute left-0 top-[18%] bottom-[18%] w-[3px] rounded-full bg-gradient-to-b from-primary via-tertiary to-secondary shadow-[0_0_12px_rgba(56,232,255,0.7)]"
            ></span>
            <span class="material-symbols-outlined text-[22px]">{{ item.icon }}</span>
            <span class="text-[15px]">{{ item.label }}</span>
          </a>
        </RouterLink>
      </template>
    </div>

    <div class="mt-auto flex flex-col gap-1.5 border-t border-outline-variant/20 pt-4">
      <button
        type="button"
        class="relative w-full py-3 px-4 rounded-xl grad-border bg-surface-container font-display font-bold text-on-surface hover:text-primary hover:glow-shadow transition-all mb-1.5 flex items-center justify-center gap-2"
        @click="goUpgrade"
      >
        <span class="material-symbols-outlined text-[18px] text-secondary">bolt</span>
        升级套餐
      </button>
      <RouterLink
        to="/support"
        class="flex items-center gap-3 px-4 py-2.5 text-on-surface-variant font-display rounded-xl hover:bg-primary/5 hover:text-primary transition-all duration-200"
      >
        <span class="material-symbols-outlined text-[22px]">help_outline</span>
        <span class="text-[15px]">帮助</span>
      </RouterLink>
      <RouterLink
        to="/settings"
        class="flex items-center gap-3 px-4 py-2.5 text-on-surface-variant font-display rounded-xl hover:bg-primary/5 hover:text-primary transition-all duration-200"
      >
        <span class="material-symbols-outlined text-[22px]">settings</span>
        <span class="text-[15px]">设置</span>
      </RouterLink>
      <button
        class="flex items-center gap-3 px-4 py-2.5 text-on-surface-variant font-display rounded-xl hover:bg-error/10 hover:text-error transition-all duration-200"
        @click="logout"
      >
        <span class="material-symbols-outlined text-[22px]">logout</span>
        <span class="text-[15px]">退出登录</span>
      </button>
    </div>
  </nav>
</template>
