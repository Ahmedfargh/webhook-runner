import api from './api'

export const webhookService = {
  async sendWebhook(data) {
    const res = await api.post('/webhooks/send', data)
    return res.data
  },

  async listWebhookCalls(params = {}) {
    const res = await api.get('/webhooks/calls', { params })
    return res.data
  },

  async getWebhookCall(id) {
    const res = await api.get(`/webhooks/calls/${id}`)
    return res.data
  },

  async retryWebhookCall(id) {
    const res = await api.post(`/webhooks/calls/${id}/retry`)
    return res.data
  },
}
