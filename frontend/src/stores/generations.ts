import { defineStore } from 'pinia'
import { GenAPI } from '../api/client'
import { useAuthStore } from './auth'
import type { Generation, GenStatus, CreateGenInput } from '../types'

const POLL_INTERVAL_MS = 1500

interface GenerationsState {
  items: Generation[]
  active: Generation | null
  loading: boolean
  // Handle of the in-flight poll timer, so it can be cancelled on logout/reset.
  pollTimer: number | null
}

function isTerminal(status: GenStatus): boolean {
  return status === 'completed' || status === 'failed'
}

export const useGenerationsStore = defineStore('generations', {
  state: (): GenerationsState => ({
    items: [],
    active: null,
    loading: false,
    pollTimer: null,
  }),

  getters: {
    completed: (state): Generation[] => state.items.filter((g) => g.status === 'completed'),
  },

  actions: {
    upsert(gen: Generation) {
      const idx = this.items.findIndex((g) => g.id === gen.id)
      if (idx === -1) {
        this.items.unshift(gen)
      } else {
        this.items[idx] = gen
      }
    },

    async fetchAll(params?: { status?: GenStatus; limit?: number }) {
      this.loading = true
      try {
        this.items = await GenAPI.list(params)
      } finally {
        this.loading = false
      }
    },

    async create(input: CreateGenInput): Promise<Generation> {
      const auth = useAuthStore()
      const gen = await GenAPI.create(input)
      this.active = gen
      this.upsert(gen)
      // Credits were deducted on the backend at create time; sync the pill.
      auth.refresh().catch(() => {})
      // Begin polling until the worker finishes (or fails).
      this.pollActive()
      return gen
    },

    stopPolling() {
      if (this.pollTimer !== null) {
        window.clearTimeout(this.pollTimer)
        this.pollTimer = null
      }
    },

    // reset clears all per-account state and cancels polling. Called on logout
    // and account switch so one user never sees another's generations/active image.
    reset() {
      this.stopPolling()
      this.items = []
      this.active = null
      this.loading = false
    },

    async pollActive() {
      const auth = useAuthStore()
      const id = this.active?.id
      if (!id) return
      this.stopPolling()

      const tick = async () => {
        // Bail if logged out or the active generation has been swapped out.
        if (!auth.isAuthenticated) {
          this.stopPolling()
          return
        }
        if (!this.active || this.active.id !== id) return
        try {
          const gen = await GenAPI.get(id)
          this.active = gen
          this.upsert(gen)
          if (isTerminal(gen.status)) {
            // On completion or refund (failed), credit balance may have changed.
            auth.refresh().catch(() => {})
            return
          }
        } catch (e) {
          // Stop polling on error to avoid a tight failure loop.
          return
        }
        if (this.active && this.active.id === id && !isTerminal(this.active.status)) {
          this.pollTimer = window.setTimeout(tick, POLL_INTERVAL_MS)
        }
      }

      this.pollTimer = window.setTimeout(tick, POLL_INTERVAL_MS)
    },

    async remove(id: string) {
      await GenAPI.remove(id)
      this.items = this.items.filter((g) => g.id !== id)
      if (this.active?.id === id) {
        this.stopPolling()
        this.active = null
      }
    },
  },
})
