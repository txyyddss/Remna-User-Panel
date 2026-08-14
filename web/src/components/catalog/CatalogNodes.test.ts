import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { PurchaseQuote } from '@/api/types'
import CatalogNodes from './CatalogNodes.vue'

const quote: PurchaseQuote = {
  comboId: 'combo-1',
  comboName: 'North',
  grossPrice: { currency: 'TXB', minor: '1000', display: '10.00 TXB' },
  discount: { currency: 'TXB', minor: '0', display: '0.00 TXB' },
  netPrice: { currency: 'TXB', minor: '1000', display: '10.00 TXB' },
  effectiveAt: '2026-08-14T00:00:00Z',
  expiresAt: '2026-09-13T00:00:00Z',
  queued: false,
  addonSquadUuids: [],
  accessibleNodes: [{
    uuid: '00000000-0000-4000-8000-000000000001',
    name: 'Tokyo relay',
    countryCode: 'JP',
    consumptionMultiplier: 1,
    activeInboundUuids: [],
    accessible: true,
    providerName: 'Transit provider',
    providerFaviconUrl: 'https://example.test/favicon.png',
  }],
}

describe('CatalogNodes', () => {
  it('renders the projected provider name and favicon', () => {
    const wrapper = mount(CatalogNodes, { props: { quote, loading: false } })

    expect(wrapper.text()).toContain('Transit provider')
    expect(wrapper.get('.catalog-node__provider-icon').attributes('src')).toBe('https://example.test/favicon.png')
  })
})
