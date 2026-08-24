import { flushPromises, shallowMount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { CatalogNode, SquadProduct } from '@/api/types'
import CatalogSquadStep from './CatalogSquadStep.vue'
import SquadSelector from './SquadSelector.vue'
import StatisticsGeocheckModal from '../statistics/StatisticsGeocheckModal.vue'

const apiMocks = vi.hoisted(() => ({
  getNodeGeocheck: vi.fn(),
}))

vi.mock('@/api/client', () => ({ api: apiMocks }))

const node: CatalogNode = {
  uuid: '00000000-0000-4000-8000-000000000001',
  name: 'Tokyo relay',
  countryCode: 'JP',
  consumptionMultiplier: 1,
  providerName: 'Transit provider',
}
const squad: SquadProduct = {
  id: 'squad-1', remnaSquadUuid: '00000000-0000-4000-8000-000000000002', name: 'North transit', description: '', profile: null,
  price: { currency: 'TXB', minor: '100', display: '1.00 TXB' }, visible: true, upstreamPresent: true, stockHeldByCurrentUser: false,
  activationRequired: false, accessibleNodes: [node], stockRemaining: 4,
  createdAt: '2026-08-21T00:00:00Z', updatedAt: '2026-08-21T00:00:00Z',
}
function mountStep(featuredIds: string[] = []) {
  return shallowMount(CatalogSquadStep, {
    props: { squads: [squad], selectedIds: [], includedIds: [], featuredIds, orderedIds: [squad.id] },
  })
}

describe('CatalogSquadStep', () => {
  beforeEach(() => vi.clearAllMocks())

  it('hands preloaded composition presentation to the selector', () => {
    const wrapper = mountStep([squad.id])

    expect(wrapper.getComponent(SquadSelector).props('featuredIds')).toEqual([squad.id])
    expect(wrapper.getComponent(SquadSelector).props('orderedIds')).toEqual([squad.id])
  })

  it('opens the shared modal for the exact catalog node', async () => {
    const result = { nodeUuid: node.uuid, checkedAt: '2026-08-21T00:00:00Z', image: { format: 'svg', mediaType: 'image/svg+xml', encoding: 'base64', data: 'PHN2Zy8+' } }
    apiMocks.getNodeGeocheck.mockResolvedValue(result)
    const wrapper = mountStep()

    wrapper.getComponent(SquadSelector).vm.$emit('openGeocheck', node)
    await flushPromises()

    expect(apiMocks.getNodeGeocheck).toHaveBeenCalledWith(node.uuid)
    expect(wrapper.getComponent(StatisticsGeocheckModal).props('node')).toStrictEqual(node)
    expect(wrapper.getComponent(StatisticsGeocheckModal).props('result')).toEqual(result)
  })
})
