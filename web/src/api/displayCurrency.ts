import type { DisplayCurrencyPreference } from './types'
import { request } from './http'

export const displayCurrencyApi = {
  get: () => request<DisplayCurrencyPreference>('/api/v1/me/display-currency'),
  update: (currency: DisplayCurrencyPreference['currency']) => request<DisplayCurrencyPreference>('/api/v1/me/display-currency', {
    method: 'PUT', body: { currency },
  }),
}
