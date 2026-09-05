import api from './api'

export const countryService = {
  async listCountries() {
    const response = await api.get('/countries')
    return response.data
  },
}
