import api from './api'

export const subscriptionService = {
  async getCurrentSubscription() {
    const res = await api.get('/subscriptions/current')
    return res.data
  },

  async subscribe(data) {
    const res = await api.post('/subscriptions/subscribe', data)
    return res.data
  },

  async cancelSubscription(data = {}) {
    const res = await api.post('/subscriptions/cancel', data)
    return res.data
  },

  async listAllSubscriptions(params = {}) {
    const res = await api.get('/subscriptions/admin/all', { params })
    return res.data
  },

  async adminOverrideSubscription(data) {
    const res = await api.post('/subscriptions/admin/override', data)
    return res.data
  },
}
