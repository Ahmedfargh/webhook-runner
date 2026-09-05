import api from './api'

export const appService = {
  async listApps(params = {}) {
    const res = await api.get('/apps', { params })
    return res.data
  },

  async getApp(id) {
    const res = await api.get(`/apps/${id}`)
    return res.data
  },

  async createApp(data) {
    const res = await api.post('/apps', data)
    return res.data
  },

  async updateApp(id, data) {
    const res = await api.put(`/apps/${id}`, data)
    return res.data
  },

  async deleteApp(id) {
    const res = await api.delete(`/apps/${id}`)
    return res.data
  },

  async rotateSecrets(id, data = {}) {
    const res = await api.post(`/apps/${id}/rotate-secrets`, data)
    return res.data
  },
}
