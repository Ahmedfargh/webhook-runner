import api from './api'

export const manualPaymentService = {
  async submitPayment(data) {
    const res = await api.post('/manual-payments', data)
    return res.data
  },

  async listAllPayments(params = {}) {
    const res = await api.get('/manual-payments/admin/all', { params })
    return res.data
  },

  async reviewPayment(id, data) {
    const res = await api.post(`/manual-payments/admin/${id}/review`, data)
    return res.data
  },
}
