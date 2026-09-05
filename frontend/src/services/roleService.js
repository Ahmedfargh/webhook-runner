import api from './api'

export const roleService = {
  async listRoles(params = {}) {
    const response = await api.get('/roles', { params })
    return response.data
  },

  async getRole(id) {
    const response = await api.get(`/roles/${id}`)
    return response.data
  },

  async createRole(payload) {
    const response = await api.post('/roles', payload)
    return response.data
  },

  async updateRole(id, payload) {
    const response = await api.put(`/roles/${id}`, payload)
    return response.data
  },

  async deleteRole(id) {
    const response = await api.delete(`/roles/${id}`)
    return response.data
  },

  async assignPermissions(roleId, permissionIds) {
    const response = await api.post(`/roles/${roleId}/permissions`, { permission_ids: permissionIds })
    return response.data
  },
}
