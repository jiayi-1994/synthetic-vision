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

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <!-- SideNavBar (Desktop Only) -->
  <nav
    class="hidden md:flex bg-surface-container-low/60 backdrop-blur-xl border-r border-outline-variant/20 fixed left-0 top-0 h-full w-72 flex-col p-margin-sm gap-4 z-40"
  >
    <div class="flex items-center gap-3 mb-8 px-4 mt-4">
      <div
        class="w-10 h-10 rounded-full bg-primary-container flex items-center justify-center glow-shadow"
      >
        <span class="material-symbols-outlined text-on-primary-container">auto_awesome</span>
      </div>
      <div>
        <h1 class="text-headline-lg font-headline-lg text-primary tracking-tight text-[20px]">
          Synthetic Vision
        </h1>
        <p class="font-label-sm text-label-sm text-on-surface-variant uppercase mt-1">V3.5 引擎</p>
      </div>
    </div>
    <div class="flex flex-col gap-2 flex-grow">
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
            class="flex items-center gap-3 px-4 py-3 rounded-xl group transition-all duration-200 active:translate-x-1"
            :class="
              (item.to === '/' ? isExactActive : isActive)
                ? 'bg-primary-container text-on-primary-container font-bold hover:bg-surface-variant/50'
                : 'text-on-surface-variant font-medium hover:bg-surface-variant/30 hover:text-primary'
            "
            @click="navigate"
          >
            <span class="material-symbols-outlined">{{ item.icon }}</span>
            <span>{{ item.label }}</span>
          </a>
        </RouterLink>
      </template>
    </div>
    <div class="mt-auto flex flex-col gap-2 border-t border-outline-variant/20 pt-4">
      <button
        class="w-full py-3 px-4 rounded-xl border border-outline-variant text-primary font-bold hover:bg-surface-variant/50 transition-colors mb-2"
      >
        升级套餐
      </button>
      <a
        class="flex items-center gap-3 px-4 py-3 text-on-surface-variant font-medium rounded-xl group hover:bg-surface-variant/30 hover:text-primary transition-all duration-200"
        href="#"
      >
        <span class="material-symbols-outlined">help_outline</span>
        <span>帮助</span>
      </a>
      <a
        class="flex items-center gap-3 px-4 py-3 text-on-surface-variant font-medium rounded-xl group hover:bg-surface-variant/30 hover:text-primary transition-all duration-200"
        href="#"
      >
        <span class="material-symbols-outlined">settings</span>
        <span>设置</span>
      </a>
      <button
        class="flex items-center gap-3 px-4 py-3 text-on-surface-variant font-medium rounded-xl group hover:bg-error/10 hover:text-error transition-all duration-200"
        @click="logout"
      >
        <span class="material-symbols-outlined">logout</span>
        <span>退出登录</span>
      </button>
    </div>
  </nav>
</template>
