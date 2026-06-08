<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { AdminAPI } from '@/api/client'
import { relativeTime } from '@/lib/format'
import { useMagnet, clickSpark } from '@/composables/useInteractions'
import type { AdminUser } from '@/types'

const injectBtn = useMagnet(4)

const PAGE_SIZE = 10

const users = ref<AdminUser[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(PAGE_SIZE)
const search = ref('')

const cluster = ref<{ status: string; load_percent: number } | null>(null)

const form = reactive({ target_public_id: '', amount: 0 })
const injecting = ref(false)
const formError = ref('')

const toast = ref<{ msg: string; kind: 'success' | 'error' } | null>(null)
let toastTimer: ReturnType<typeof setTimeout> | undefined

function showToast(msg: string, kind: 'success' | 'error') {
  toast.value = { msg, kind }
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => (toast.value = null), 3200)
}

// Pagination display: "Showing X-Y of N"
const rangeStart = computed(() => (total.value === 0 ? 0 : (page.value - 1) * pageSize.value + 1))
const rangeEnd = computed(() => Math.min(page.value * pageSize.value, total.value))
const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize.value)))
const hasSearch = computed(() => search.value.trim() !== '')

function fmt(n: number): string {
  return n.toLocaleString('en-US')
}

// Credit pill color by amount: 0 -> error, low -> secondary, high -> primary
function pillClass(credits: number): string {
  if (credits <= 0)
    return 'bg-error-container/20 border border-error/30 text-error'
  if (credits < 1000)
    return 'bg-surface-container border border-outline-variant/30 text-secondary'
  return 'bg-surface-container border border-outline-variant/30 text-primary'
}

// Initials chip color cycling for visual variety
function chipClass(u: AdminUser): string {
  if (u.credits <= 0)
    return 'bg-surface-bright text-on-surface border border-outline-variant/30'
  if (u.credits < 1000)
    return 'bg-tertiary-container/20 text-tertiary border border-tertiary/20'
  return 'bg-secondary-container/20 text-secondary border border-secondary/20'
}

async function loadUsers() {
  try {
    const res = await AdminAPI.users({
      search: search.value || undefined,
      page: page.value,
      page_size: pageSize.value,
    })
    users.value = res.users
    total.value = res.total
    page.value = res.page
    pageSize.value = res.page_size
  } catch (e: any) {
    showToast(e?.response?.data?.error || e?.message || '加载用户失败。', 'error')
  }
}

async function loadCluster() {
  try {
    cluster.value = await AdminAPI.cluster()
  } catch {
    /* non-fatal */
  }
}

function prefillRecharge(u: AdminUser) {
  form.target_public_id = u.public_id
  formError.value = ''
}

function applySearch() {
  page.value = 1
  loadUsers()
}

function clearSearch() {
  search.value = ''
  applySearch()
}

function goPrev() {
  if (page.value > 1) {
    page.value -= 1
    loadUsers()
  }
}
function goNext() {
  if (page.value < totalPages.value) {
    page.value += 1
    loadUsers()
  }
}

async function inject(e?: MouseEvent) {
  if (injecting.value) return
  if (e) clickSpark(e)
  formError.value = ''
  if (!form.target_public_id.trim()) {
    formError.value = '请输入目标用户 ID。'
    return
  }
  if (!Number.isFinite(form.amount) || form.amount <= 0) {
    formError.value = '请输入正数的积分数量。'
    return
  }
  injecting.value = true
  try {
    const res = await AdminAPI.inject({
      target_public_id: form.target_public_id.trim(),
      amount: Number(form.amount),
    })
    showToast(
      `已向 ${res.user.public_id} 注入 ${fmt(Number(form.amount))} 积分。`,
      'success'
    )
    form.amount = 0
    await loadUsers()
  } catch (e: any) {
    const msg = e?.response?.data?.error || e?.message || '注入失败。'
    formError.value = msg
    showToast(msg, 'error')
  } finally {
    injecting.value = false
  }
}

onMounted(() => {
  loadUsers()
  loadCluster()
})
</script>

<template>
  <main class="flex-1 p-margin-sm md:p-margin-lg overflow-y-auto">
    <div class="max-w-container-max mx-auto space-y-gutter">
      <!-- Page Header -->
      <div class="mb-8" v-reveal>
        <p class="font-mono text-micro text-on-surface-variant mb-3 uppercase">控制台 · ADMIN CONSOLE</p>
        <h2 class="font-display text-[32px] md:text-display-lg font-bold text-on-surface mb-2">
          用户<span class="text-neon">目录</span>
        </h2>
        <p class="text-on-surface-variant max-w-2xl">
          管理平台访问权限，监控生成活动，并向目标账户注入管理积分。
        </p>
      </div>

      <!-- Bento Grid Layout -->
      <div class="grid grid-cols-1 xl:grid-cols-12 gap-gutter items-start">
        <!-- Table Section (Spans 8 cols on XL) -->
        <div
          v-reveal
          class="xl:col-span-8 glass-panel rounded-2xl overflow-hidden flex flex-col"
        >
          <div
            class="p-6 border-b border-outline-variant/20 flex flex-col lg:flex-row lg:items-center justify-between gap-4 bg-surface-container/50"
          >
            <h3 class="font-display font-bold text-title flex items-center gap-2 text-on-surface">
              <span class="material-symbols-outlined text-primary text-[20px]">group</span>
              活跃用户
              <span class="font-mono text-micro text-on-surface-variant uppercase ml-1">DIRECTORY</span>
            </h3>
            <form class="flex flex-col sm:flex-row gap-2 sm:items-center" @submit.prevent="applySearch">
              <div class="relative">
                <span class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-outline text-[18px]">search</span>
                <input
                  v-model="search"
                  class="field pl-10 !py-2.5 font-mono text-[12px] w-full sm:w-72"
                  placeholder="搜索用户名 / 邮箱 / 用户 ID"
                  type="search"
                />
              </div>
              <button
                type="submit"
                class="min-h-[40px] px-3 rounded-lg border border-outline-variant/50 text-on-surface-variant font-mono text-[11px] uppercase hover:border-primary/40 hover:text-primary transition-colors flex items-center justify-center gap-1.5"
              >
                <span class="material-symbols-outlined text-[16px]">filter_list</span> 搜索
              </button>
              <button
                v-if="hasSearch"
                type="button"
                class="min-h-[40px] px-3 rounded-lg border border-outline-variant/50 text-on-surface-variant font-mono text-[11px] uppercase hover:border-secondary/40 hover:text-secondary transition-colors"
                @click="clearSearch"
              >
                清空
              </button>
            </form>
          </div>

          <div
            v-if="hasSearch"
            class="px-6 py-3 border-b border-outline-variant/20 bg-primary/5 font-mono text-[11px] text-on-surface-variant flex flex-wrap items-center gap-2"
          >
            <span class="uppercase">SEARCH</span>
            <span class="px-2 py-1 rounded-full bg-primary/10 text-primary border border-primary/25">
              “{{ search.trim() }}”
            </span>
            <span>匹配 {{ fmt(total) }} 个用户</span>
          </div>

          <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="bg-surface-container-highest/20 border-b border-outline-variant/20">
                  <th class="py-3 px-6 font-mono text-micro text-on-surface-variant uppercase">用户名 / ID</th>
                  <th class="py-3 px-6 font-mono text-micro text-on-surface-variant uppercase">邮箱地址</th>
                  <th class="py-3 px-6 font-mono text-micro text-on-surface-variant uppercase">当前积分</th>
                  <th class="py-3 px-6 font-mono text-micro text-on-surface-variant uppercase">最近活动</th>
                  <th class="py-3 px-6 font-mono text-micro text-on-surface-variant uppercase text-right">操作</th>
                </tr>
              </thead>
              <tbody class="font-body-md text-sm divide-y divide-outline-variant/10">
                <tr
                  v-for="u in users"
                  :key="u.public_id"
                  class="hover:bg-primary/5 transition-colors group"
                >
                  <td class="py-4 px-6">
                    <div class="flex items-center gap-3">
                      <div
                        class="w-9 h-9 rounded-xl flex items-center justify-center font-display font-bold text-xs"
                        :class="chipClass(u)"
                      >
                        {{ u.initials }}
                      </div>
                      <div>
                        <div class="font-semibold text-on-surface">{{ u.username }}</div>
                        <div class="font-mono text-[10px] text-on-surface-variant opacity-70">
                          {{ u.public_id }}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td class="py-4 px-6 text-on-surface-variant font-mono text-[12px]">{{ u.email }}</td>
                  <td class="py-4 px-6">
                    <span
                      class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full font-mono text-[12px]"
                      :class="pillClass(u.credits)"
                    >
                      <span
                        v-if="u.credits >= 1000"
                        class="w-1.5 h-1.5 rounded-full bg-primary animate-pulse-dot"
                      ></span>
                      {{ fmt(u.credits) }}
                    </span>
                  </td>
                  <td class="py-4 px-6 text-on-surface-variant font-mono text-[11px]">
                    {{ relativeTime(u.last_activity_at) }}
                  </td>
                  <td class="py-4 px-6 text-right">
                    <button
                      type="button"
                      class="min-h-[40px] px-3 rounded-lg border border-outline-variant/50 hover:border-primary/50 text-primary font-mono text-[11px] uppercase transition-all bg-surface-container/30 hover:bg-primary/8"
                      @click="prefillRecharge(u)"
                    >
                      充值
                    </button>
                  </td>
                </tr>

                <tr v-if="users.length === 0">
                  <td colspan="5" class="py-10 px-6 text-center text-on-surface-variant">
                    未找到用户。
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div
            class="p-4 border-t border-outline-variant/20 bg-surface-container/30 flex justify-between items-center font-mono text-[11px] text-on-surface-variant"
          >
            <span>显示第 {{ rangeStart }}-{{ rangeEnd }} 项，共 {{ fmt(total) }} 个用户</span>
            <div class="flex gap-1">
              <button
                type="button"
                class="w-11 h-11 rounded-lg flex items-center justify-center hover:bg-primary/10 hover:text-primary transition-colors disabled:opacity-40 disabled:hover:bg-transparent"
                :disabled="page <= 1"
                @click="goPrev"
              >
                <span class="material-symbols-outlined text-[18px]">chevron_left</span>
              </button>
              <button
                type="button"
                class="w-11 h-11 rounded-lg flex items-center justify-center hover:bg-primary/10 hover:text-primary transition-colors disabled:opacity-40 disabled:hover:bg-transparent"
                :disabled="page >= totalPages"
                @click="goNext"
              >
                <span class="material-symbols-outlined text-[18px]">chevron_right</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Quick Action Section (Spans 4 cols on XL) -->
        <div class="xl:col-span-4 space-y-6">
          <!-- Manual Injection Form -->
          <div
            v-reveal="80"
            v-spotlight
            class="spotlight glass-panel grad-border rounded-2xl p-6 relative overflow-hidden"
          >
            <h3 class="font-display font-bold text-title flex items-center gap-2 mb-6 text-primary relative z-10">
              <span class="material-symbols-outlined text-[20px]" style="font-variation-settings: 'FILL' 1">bolt</span>
              手动注入积分
            </h3>
            <form class="space-y-5 relative z-10" @submit.prevent="inject()">
              <div class="space-y-2">
                <label class="block font-mono text-micro text-on-surface-variant uppercase">目标用户 ID</label>
                <div class="relative">
                  <span class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-outline text-[18px]">person_search</span>
                  <input
                    v-model="form.target_public_id"
                    class="field pl-10 font-mono text-[13px]"
                    placeholder="例如 USR-9102B"
                    type="text"
                  />
                </div>
              </div>
              <div class="space-y-2">
                <label class="block font-mono text-micro text-on-surface-variant uppercase">积分数量</label>
                <div class="relative">
                  <span class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-outline text-[18px]">database</span>
                  <input
                    v-model.number="form.amount"
                    class="field pl-10 font-mono text-[13px]"
                    placeholder="0"
                    type="number"
                  />
                </div>
              </div>

              <p v-if="formError" class="flex items-center gap-2 text-error font-mono text-[12px]">
                <span class="material-symbols-outlined text-sm">error</span>{{ formError }}
              </p>

              <div class="pt-1">
                <button
                  ref="injectBtn"
                  class="btn-primary w-full flex justify-center items-center gap-2"
                  type="button"
                  :disabled="injecting"
                  @click="inject($event)"
                >
                  {{ injecting ? '注入中…' : '确认充值' }}
                  <span
                    class="material-symbols-outlined text-[18px]"
                    :class="injecting ? 'animate-spin' : ''"
                    >{{ injecting ? 'progress_activity' : 'arrow_forward' }}</span
                  >
                </button>
              </div>
            </form>
          </div>

          <!-- Compute Cluster Status -->
          <div
            v-reveal="140"
            class="glass-panel rounded-2xl p-5 flex items-center justify-between"
          >
            <div>
              <div class="font-mono text-micro text-on-surface-variant mb-2 uppercase">计算集群 · CLUSTER</div>
              <div class="flex items-center gap-2">
                <span class="w-2 h-2 rounded-full bg-success shadow-[0_0_8px_rgba(61,255,176,0.8)]"></span>
                <span class="font-mono text-[12px] text-on-surface">
                  {{ cluster ? `${cluster.status} · LOAD ${cluster.load_percent}%` : '连接中…' }}
                </span>
              </div>
              <div v-if="cluster" class="w-full h-1.5 bg-surface-variant rounded-full overflow-hidden mt-3">
                <div
                  class="h-full rounded-full bg-gradient-to-r from-success via-primary to-secondary transition-all duration-1000"
                  :style="{ width: `${cluster.load_percent}%` }"
                ></div>
              </div>
            </div>
            <span class="material-symbols-outlined text-primary ml-4">memory</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Toast -->
    <transition
      enter-active-class="transition duration-200 ease-out"
      enter-from-class="opacity-0 translate-y-2"
      leave-active-class="transition duration-200 ease-in"
      leave-to-class="opacity-0 translate-y-2"
    >
      <div
        v-if="toast"
        class="fixed bottom-6 right-6 z-50 flex items-center gap-2 px-4 py-3 rounded-xl glass-panel grad-border max-w-sm"
        :class="toast.kind === 'success' ? 'text-success' : 'text-error'"
      >
        <span class="material-symbols-outlined text-[18px]">{{
          toast.kind === 'success' ? 'check_circle' : 'error'
        }}</span>
        <span class="text-sm text-on-surface">{{ toast.msg }}</span>
      </div>
    </transition>
  </main>
</template>
