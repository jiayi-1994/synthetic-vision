import { defineStore } from 'pinia'
import { GenAPI } from '../api/client'
import { useAuthStore } from './auth'
import type { Generation, GenStatus, CreateGenInput } from '../types'

const POLL_INTERVAL_MS = 1500

interface GenerationsState {
  items: Generation[]
  active: Generation | null
  loading: boolean
  pollingIds: string[]
  activeBatchIds: string[]
  // Handle of the in-flight poll timer, so it can be cancelled on logout/reset.
  pollTimer: number | null
}

function isTerminal(status: GenStatus): boolean {
  return status === 'completed' || status === 'failed'
}

function dedupeIds(ids: string[]): string[] {
  return [...new Set(ids)]
}

export const useGenerationsStore = defineStore('generations', {
  state: (): GenerationsState => ({
    items: [],
    active: null,
    loading: false,
    pollingIds: [],
    activeBatchIds: [],
    pollTimer: null,
  }),

  getters: {
    completed: (state): Generation[] => state.items.filter((g) => g.status === 'completed'),

    activeBatch(state): string[] {
      return dedupeIds(state.activeBatchIds)
    },

    activeBatchProgress(state): { done: number; total: number } {
      const ids = dedupeIds(state.activeBatchIds)
      if (ids.length === 0) {
        return { done: 0, total: 0 }
      }

      const batchItems = ids
        .map((id) => state.items.find((g) => g.id === id))
        .filter((g): g is Generation => g !== undefined)

      const done = batchItems.filter((g) => isTerminal(g.status)).length
      return { done, total: ids.length }
    },
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

    async create(input: CreateGenInput): Promise<Generation[]> {
      const auth = useAuthStore()
      const gens = await GenAPI.create(input)
      if (gens.length > 0) {
        this.active = gens[0]
        gens.forEach((g) => this.upsert(g))
        this.activeBatchIds = dedupeIds(gens.map((g) => g.id))
        const inFlight = gens
          .filter((g) => !isTerminal(g.status))
          .map((g) => g.id)
        if (inFlight.length > 0) {
          this.pollingIds = dedupeIds([...this.pollingIds, ...inFlight])
        }
      } else {
        this.active = null
      }
      // Credits were deducted on the backend at create time; sync the pill.
      auth.refresh().catch(() => {})
      // Begin polling until all submitted jobs finish (or fail).
      if (this.pollingIds.length > 0) {
        this.pollActive()
      }
      return gens
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
      this.pollingIds = []
      this.activeBatchIds = []
      this.loading = false
    },

    async pollActive() {
      const auth = useAuthStore()
      this.stopPolling()

      const tick = async () => {
        // Bail if logged out.
        if (!auth.isAuthenticated) {
          this.stopPolling()
          return
        }

        const ids = dedupeIds(this.pollingIds)
        if (ids.length === 0) {
          this.pollingIds = []
          this.stopPolling()
          return
        }

        const results = await Promise.all(
          ids.map(async (id) => {
            try {
              const gen = await GenAPI.get(id)
              return { id, gen, ok: true as const }
            } catch {
              return { id, gen: null as Generation | null, ok: false as const }
            }
          }),
        )

        const nextPollSet = new Set(ids)
        let shouldRefreshCredits = false

        for (const result of results) {
          if (!result.ok || result.gen === null) {
            // Keep the ID when fetch fails transiently; retry on next tick.
            continue
          }

          this.upsert(result.gen)
          if (this.active?.id === result.id) {
            this.active = result.gen
          }

          if (isTerminal(result.gen.status)) {
            shouldRefreshCredits = true
            nextPollSet.delete(result.id)
          } else {
            nextPollSet.add(result.id)
          }
        }

        this.pollingIds = dedupeIds(Array.from(nextPollSet))
        if (shouldRefreshCredits) {
          // On completion or failed, credit balance may have changed.
          auth.refresh().catch(() => {})
        }

        if (this.pollingIds.length > 0) {
          this.pollTimer = window.setTimeout(tick, POLL_INTERVAL_MS)
        } else {
          this.stopPolling()
        }
      }

      // Start from known non-terminal IDs.
      this.pollingIds = dedupeIds(this.pollingIds)
      if (this.pollingIds.length === 0) {
        return
      }
      this.pollTimer = window.setTimeout(tick, POLL_INTERVAL_MS)
    },

    async remove(id: string) {
      await GenAPI.remove(id)
      this.items = this.items.filter((g) => g.id !== id)
      this.pollingIds = this.pollingIds.filter((pid) => pid !== id)
      this.activeBatchIds = this.activeBatchIds.filter((bid) => bid !== id)
      if (this.active?.id === id) {
        const nextActive = this.items.find((g) => g.id !== id && !isTerminal(g.status))
        this.active = nextActive ?? null
      }
      if (this.pollingIds.length === 0) {
        this.stopPolling()
      }
    },
  },
})
