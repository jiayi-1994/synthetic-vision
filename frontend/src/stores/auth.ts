import { defineStore } from 'pinia'
import { AuthAPI } from '../api/client'
import { useGenerationsStore } from './generations'
import type { User } from '../types'

const TOKEN_KEY = 'sv_token'

interface AuthState {
  token: string | null
  user: User | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    token: localStorage.getItem(TOKEN_KEY),
    user: null,
  }),

  getters: {
    isAuthenticated: (state): boolean => !!state.token,
    isAdmin: (state): boolean => state.user?.role === 'admin',
  },

  actions: {
    persist(token: string, user: User) {
      this.token = token
      this.user = user
      localStorage.setItem(TOKEN_KEY, token)
    },

    async login(email: string, password: string) {
      // Drop any prior account's cached generations before switching identity.
      useGenerationsStore().reset()
      const { token, user } = await AuthAPI.login({ email, password })
      this.persist(token, user)
      return user
    },

    async register(username: string, email: string, password: string) {
      useGenerationsStore().reset()
      const { token, user } = await AuthAPI.register({ username, email, password })
      this.persist(token, user)
      return user
    },

    async fetchMe() {
      const { user } = await AuthAPI.me()
      this.user = user
      return user
    },

    async refresh() {
      return this.fetchMe()
    },

    setCredits(n: number) {
      if (this.user) {
        this.user.credits = n
      }
    },

    logout() {
      this.token = null
      this.user = null
      localStorage.removeItem(TOKEN_KEY)
      // Clear the other account-scoped store so a re-login starts clean.
      useGenerationsStore().reset()
    },
  },
})
