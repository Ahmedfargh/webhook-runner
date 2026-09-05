import api from './api'

export const permissionService = {
  async listPermissions(params = {}) {
    const response = await api.get('/permissions', { params })
    return response.data
  },

  async getPermission(id) {
    const response = await api.get(`/permissions/${id}`)
    return response.data
  },

  async createPermission(payload) {
    const response = await api.post('/permissions', payload)
    return response.data
  },

  async updatePermission(id, payload) {
    const response = await api.put(`/permissions/${id}`, payload)
    return response.data
  },

  async deletePermission(id) {
    const response = await api.delete(`/permissions/${id}`)
    return response.data
  },
}
