<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { AdminAPI } from '@/api/client'
import { relativeTime } from '@/lib/format'
import type { AdminUser } from '@/types'

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

async function inject() {
  if (injecting.value) return
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
      <div class="mb-8">
        <h2 class="text-display-lg font-display-lg text-on-surface mb-2 text-[32px] md:text-[48px]">
          用户目录
        </h2>
        <p class="text-on-surface-variant max-w-2xl">
          管理平台访问权限，监控生成活动，并向目标账户注入管理积分。
        </p>
      </div>

      <!-- Bento Grid Layout -->
      <div class="grid grid-cols-1 xl:grid-cols-12 gap-gutter items-start">
        <!-- Table Section (Spans 8 cols on XL) -->
        <div
          class="xl:col-span-8 bg-surface-container-low/60 backdrop-blur-xl border border-outline-variant/30 rounded-xl overflow-hidden shadow-sm shadow-primary/5 flex flex-col"
        >
          <div
            class="p-6 border-b border-outline-variant/20 flex justify-between items-center bg-surface-container/50"
          >
            <h3 class="font-bold text-lg flex items-center gap-2">
              <span class="material-symbols-outlined text-primary text-[20px]">group</span>
              活跃用户
            </h3>
            <div class="flex gap-2">
              <button
                type="button"
                class="px-3 py-1.5 rounded border border-outline-variant/50 text-on-surface-variant font-label-sm text-label-sm hover:bg-surface-variant/50 transition-colors flex items-center gap-1"
              >
                <span class="material-symbols-outlined text-[16px]">filter_list</span> 筛选
              </button>
            </div>
          </div>

          <div class="overflow-x-auto">
            <table class="w-full text-left border-collapse">
              <thead>
                <tr class="bg-surface-container-highest/20 border-b border-outline-variant/20">
                  <th
                    class="py-3 px-6 font-label-sm text-label-sm text-on-surface-variant uppercase tracking-wider font-medium"
                  >
                    用户名 / ID
                  </th>
                  <th
                    class="py-3 px-6 font-label-sm text-label-sm text-on-surface-variant uppercase tracking-wider font-medium"
                  >
                    邮箱地址
                  </th>
                  <th
                    class="py-3 px-6 font-label-sm text-label-sm text-on-surface-variant uppercase tracking-wider font-medium"
                  >
                    当前积分
                  </th>
                  <th
                    class="py-3 px-6 font-label-sm text-label-sm text-on-surface-variant uppercase tracking-wider font-medium"
                  >
                    最近活动
                  </th>
                  <th
                    class="py-3 px-6 font-label-sm text-label-sm text-on-surface-variant uppercase tracking-wider font-medium text-right"
                  >
                    操作
                  </th>
                </tr>
              </thead>
              <tbody class="font-body-md text-sm divide-y divide-outline-variant/10">
                <tr
                  v-for="u in users"
                  :key="u.public_id"
                  class="hover:bg-surface-variant/20 transition-colors group"
                >
                  <td class="py-4 px-6">
                    <div class="flex items-center gap-3">
                      <div
                        class="w-8 h-8 rounded-full flex items-center justify-center font-bold text-xs"
                        :class="chipClass(u)"
                      >
                        {{ u.initials }}
                      </div>
                      <div>
                        <div class="font-semibold text-on-surface">{{ u.username }}</div>
                        <div
                          class="font-label-sm text-label-sm text-on-surface-variant opacity-60"
                        >
                          {{ u.public_id }}
                        </div>
                      </div>
                    </div>
                  </td>
                  <td class="py-4 px-6 text-on-surface-variant">{{ u.email }}</td>
                  <td class="py-4 px-6">
                    <span
                      class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full font-label-sm text-label-sm"
                      :class="pillClass(u.credits)"
                    >
                      <span
                        v-if="u.credits >= 1000"
                        class="w-1.5 h-1.5 rounded-full bg-primary animate-pulse"
                      ></span>
                      {{ fmt(u.credits) }}
                    </span>
                  </td>
                  <td class="py-4 px-6 text-on-surface-variant">
                    {{ relativeTime(u.last_activity_at) }}
                  </td>
                  <td class="py-4 px-6 text-right">
                    <button
                      type="button"
                      class="px-3 py-1.5 rounded border border-outline-variant hover:border-primary/50 text-primary font-label-sm text-label-sm transition-all duration-200 bg-surface-container/30 hover:bg-surface-variant/50"
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
            class="p-4 border-t border-outline-variant/20 bg-surface-container/30 flex justify-between items-center font-label-sm text-label-sm text-on-surface-variant"
          >
            <span>显示第 {{ rangeStart }}-{{ rangeEnd }} 项，共 {{ fmt(total) }} 个用户</span>
            <div class="flex gap-2">
              <button
                type="button"
                class="p-1 rounded hover:bg-surface-variant/50 transition-colors disabled:opacity-50"
                :disabled="page <= 1"
                @click="goPrev"
              >
                <span class="material-symbols-outlined text-[18px]">chevron_left</span>
              </button>
              <button
                type="button"
                class="p-1 rounded hover:bg-surface-variant/50 transition-colors disabled:opacity-50"
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
            class="bg-surface-container-low/60 backdrop-blur-xl border border-outline-variant/30 rounded-xl p-6 shadow-[0_8px_32px_rgba(124,58,237,0.05)] relative overflow-hidden group"
          >
            <div
              class="absolute top-0 right-0 w-32 h-32 bg-primary/10 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2 group-hover:bg-primary/20 transition-all duration-500"
            ></div>
            <h3
              class="font-bold text-lg flex items-center gap-2 mb-6 text-primary relative z-10"
            >
              <span
                class="material-symbols-outlined text-[20px]"
                style="font-variation-settings: 'FILL' 1"
                >bolt</span
              >
              手动注入积分
            </h3>
            <form class="space-y-5 relative z-10" @submit.prevent="inject">
              <div class="space-y-2">
                <label
                  class="block font-label-sm text-label-sm text-on-surface-variant uppercase tracking-wider"
                  >目标用户 ID</label
                >
                <div class="relative">
                  <span
                    class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-outline text-sm"
                    >person_search</span
                  >
                  <input
                    v-model="form.target_public_id"
                    class="w-full pl-9 pr-4 py-2.5 bg-surface-container border border-outline-variant/50 rounded-lg font-label-sm text-label-sm text-on-surface focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-all placeholder:text-outline/50"
                    placeholder="例如 USR-9102B"
                    type="text"
                  />
                </div>
              </div>
              <div class="space-y-2">
                <label
                  class="block font-label-sm text-label-sm text-on-surface-variant uppercase tracking-wider"
                  >积分数量</label
                >
                <div class="relative">
                  <span
                    class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-outline text-sm"
                    >database</span
                  >
                  <input
                    v-model.number="form.amount"
                    class="w-full pl-9 pr-4 py-2.5 bg-surface-container border border-outline-variant/50 rounded-lg font-label-sm text-label-sm text-on-surface focus:outline-none focus:border-primary focus:ring-1 focus:ring-primary transition-all placeholder:text-outline/50"
                    placeholder="0"
                    type="number"
                  />
                </div>
              </div>

              <p
                v-if="formError"
                class="flex items-center gap-2 text-error font-label-sm text-label-sm"
              >
                <span class="material-symbols-outlined text-sm">error</span>{{ formError }}
              </p>

              <div class="pt-2">
                <button
                  class="w-full bg-primary text-on-primary-fixed py-2.5 rounded-lg font-bold hover:brightness-110 active:scale-[0.98] transition-all flex justify-center items-center gap-2 shadow-lg shadow-primary/20 disabled:opacity-60 disabled:cursor-not-allowed"
                  type="submit"
                  :disabled="injecting"
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
            class="bg-surface-container border border-outline-variant/20 rounded-xl p-4 flex items-center justify-between"
          >
            <div>
              <div class="font-label-sm text-label-sm text-on-surface-variant mb-1">
                计算集群
              </div>
              <div class="flex items-center gap-2">
                <span
                  class="w-2 h-2 rounded-full bg-secondary shadow-[0_0_8px_rgba(137,206,255,0.8)]"
                ></span>
                <span class="font-medium text-sm text-on-surface">
                  {{ cluster ? `${cluster.status}（负载 ${cluster.load_percent}%）` : '连接中…' }}
                </span>
              </div>
            </div>
            <span class="material-symbols-outlined text-outline-variant">memory</span>
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
        class="fixed bottom-6 right-6 z-50 flex items-center gap-2 px-4 py-3 rounded-lg glass-panel shadow-lg max-w-sm"
        :class="toast.kind === 'success' ? 'text-secondary' : 'text-error'"
      >
        <span class="material-symbols-outlined text-[18px]">{{
          toast.kind === 'success' ? 'check_circle' : 'error'
        }}</span>
        <span class="text-sm text-on-surface">{{ toast.msg }}</span>
      </div>
    </transition>
  </main>
</template>
