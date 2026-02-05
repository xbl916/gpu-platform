import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('accessToken')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('accessToken')
      localStorage.removeItem('refreshToken')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const authService = {
  async register(data: { email: string; password: string; name: string }) {
    const response = await api.post('/auth/register', data)
    return response.data
  },

  async login(data: { email: string; password: string }) {
    const response = await api.post('/auth/login', data)
    return response.data
  },

  logout() {
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
  },
}

export const containerService = {
  async list(params?: { limit?: number; offset?: number }) {
    const response = await api.get('/containers', { params })
    return response.data
  },

  async get(id: string) {
    const response = await api.get(`/containers/${id}`)
    return response.data
  },

  async create(data: { name: string; resources: any; templateId?: string }) {
    const response = await api.post('/containers', data)
    return response.data
  },

  async delete(id: string) {
    const response = await api.delete(`/containers/${id}`)
    return response.data
  },

  async start(id: string) {
    const response = await api.post(`/containers/${id}/start`)
    return response.data
  },

  async stop(id: string) {
    const response = await api.post(`/containers/${id}/stop`)
    return response.data
  },

  async restart(id: string) {
    const response = await api.post(`/containers/${id}/restart`)
    return response.data
  },

  async getLogs(id: string) {
    const response = await api.get(`/containers/${id}/logs`)
    return response.data
  },

  async getMetrics(id: string) {
    const response = await api.get(`/containers/${id}/metrics`)
    return response.data
  },
}

export const templateService = {
  async list(params?: { limit?: number; offset?: number }) {
    const response = await api.get('/templates', { params })
    return response.data
  },

  async get(id: string) {
    const response = await api.get(`/templates/${id}`)
    return response.data
  },

  async create(data: { name: string; description?: string; config: any }) {
    const response = await api.post('/templates', data)
    return response.data
  },

  async update(id: string, data: { name: string; description?: string; config: any }) {
    const response = await api.put(`/templates/${id}`, data)
    return response.data
  },

  async delete(id: string) {
    const response = await api.delete(`/templates/${id}`)
    return response.data
  },

  async build(id: string, data: { baseImage?: string; pipPackages?: string[]; envVars?: Record<string, string> }) {
    const response = await api.post(`/templates/${id}/build`, data)
    return response.data
  },

  async publish(id: string) {
    const response = await api.post(`/templates/${id}/publish`)
    return response.data
  },
}

export const resourceService = {
  async getAvailability() {
    const response = await api.get('/resources/availability')
    return response.data
  },

  async listGPUPools() {
    const response = await api.get('/resources/gpu-pools')
    return response.data
  },

  async listGPUServers() {
    const response = await api.get('/resources/gpu-servers')
    return response.data
  },

  async getQuota() {
    const response = await api.get('/users/me/quota')
    return response.data
  },
}

export const monitorService = {
  async getDashboard() {
    const response = await api.get('/monitor/dashboard')
    return response.data
  },

  async getAlerts() {
    const response = await api.get('/monitor/alerts')
    return response.data
  },
}

export default api
