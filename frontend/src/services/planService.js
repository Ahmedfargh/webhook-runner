import api from './api'

export const planService = {
  async listPlans(includeInactive = false) {
    const res = await api.get('/plans', { params: { include_inactive: includeInactive } })
    return res.data
  },

  async getPlan(id) {
    const res = await api.get(`/plans/${id}`)
    return res.data
  },

  async createPlan(data) {
    const res = await api.post('/admin/plans', data)
    return res.data
  },

  async updatePlan(id, data) {
    const res = await api.put(`/admin/plans/${id}`, data)
    return res.data
  },

  async deletePlan(id) {
    const res = await api.delete(`/admin/plans/${id}`)
    return res.data
  },
}
