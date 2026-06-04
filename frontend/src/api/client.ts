import axios, { type AxiosInstance } from 'axios'
import type {
  User,
  Stats,
  Generation,
  GenStatus,
  Resolution,
  CreateGenInput,
  AdminUser,
  AnalyticsResponse,
} from '../types'

const TOKEN_KEY = 'sv_token'

export const api: AxiosInstance = axios.create({
  baseURL: '/api',
})

// Attach bearer token from localStorage on every request.
api.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY)
  if (token && config.headers) {
    config.headers.set('Authorization', `Bearer ${token}`)
  }
  return config
})

// On 401, clear the token and bounce to /login.
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error?.response?.status === 401) {
      localStorage.removeItem(TOKEN_KEY)
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    return Promise.reject(error)
  },
)

export const AuthAPI = {
  async register(b: { username: string; email: string; password: string }): Promise<{
    token: string
    user: User
  }> {
    const { data } = await api.post('/auth/register', b)
    return data
  },
  async login(b: { email: string; password: string }): Promise<{ token: string; user: User }> {
    const { data } = await api.post('/auth/login', b)
    return data
  },
  async me(): Promise<{ user: User }> {
    const { data } = await api.get('/auth/me')
    return data
  },
  async stats(): Promise<Stats> {
    const { data } = await api.get('/me/stats')
    return data
  },
}

export const AnalyticsAPI = {
  async meAnalytics(): Promise<AnalyticsResponse> {
    const { data } = await api.get('/me/analytics')
    return data
  },
}

export const GenAPI = {
  async cost(resolution: Resolution): Promise<number> {
    const { data } = await api.get('/generations/cost', { params: { resolution } })
    return data.cost
  },
  async create(input: CreateGenInput): Promise<Generation> {
    const { data } = await api.post('/generations', input)
    return data
  },
  async list(params?: { status?: GenStatus; limit?: number }): Promise<Generation[]> {
    const { data } = await api.get('/generations', { params })
    return data.generations
  },
  async get(id: string): Promise<Generation> {
    const { data } = await api.get(`/generations/${id}`)
    return data
  },
  async remove(id: string): Promise<void> {
    await api.delete(`/generations/${id}`)
  },
}

export const AdminAPI = {
  async users(p: {
    search?: string
    page?: number
    page_size?: number
  }): Promise<{ users: AdminUser[]; total: number; page: number; page_size: number }> {
    const { data } = await api.get('/admin/users', { params: p })
    return data
  },
  async inject(b: { target_public_id: string; amount: number }): Promise<{ user: User }> {
    const { data } = await api.post('/admin/credits', b)
    return data
  },
  async cluster(): Promise<{ status: string; load_percent: number }> {
    const { data } = await api.get('/admin/cluster')
    return data
  },
}

export default api
