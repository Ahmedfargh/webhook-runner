import api from './api'

export const invoiceService = {
  async getMyInvoices(params = {}) {
    const res = await api.get('/invoices', { params })
    return res.data
  },

  async getInvoice(id) {
    const res = await api.get(`/invoices/${id}`)
    return res.data
  },

  async listAllInvoices(params = {}) {
    const res = await api.get('/invoices/admin/all', { params })
    return res.data
  },

  async createManualInvoice(data) {
    const res = await api.post('/invoices/admin/create', data)
    return res.data
  },

  async markInvoicePaid(id, data = {}) {
    const res = await api.post(`/invoices/admin/${id}/mark-paid`, data)
    return res.data
  },

  async voidInvoice(id, reason = '') {
    const res = await api.post(`/invoices/admin/${id}/void`, { reason })
    return res.data
  },
}
