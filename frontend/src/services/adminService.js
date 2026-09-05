import api from './api'

export const adminService = {
  async listAdmins(params = {}) {
    const response = await api.get('/admins', { params })
    return response.data
  },

  async getAdmin(id) {
    const response = await api.get(`/admins/${id}`)
    return response.data
  },

  async createAdmin(payload) {
    const response = await api.post('/admins', payload)
    return response.data
  },

  async updateAdmin(id, payload) {
    const response = await api.put(`/admins/${id}`, payload)
    return response.data
  },

  async deleteAdmin(id) {
    const response = await api.delete(`/admins/${id}`)
    return response.data
  },

  async assignRoles(adminId, roleIds) {
    const response = await api.post(`/admins/${adminId}/roles`, { role_ids: roleIds })
    return response.data
  },

  async assignPermissions(adminId, permissionIds) {
    const response = await api.post(`/admins/${adminId}/permissions`, { permission_ids: permissionIds })
    return response.data
  },
}
