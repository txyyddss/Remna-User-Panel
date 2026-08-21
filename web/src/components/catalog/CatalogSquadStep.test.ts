import { flushPromises, shallowMount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { CatalogNode, SquadProduct, StatisticsSnapshot } from '@/api/types'
import CatalogSquadStep from './CatalogSquadStep.vue'
import SquadSelector from './SquadSelector.vue'
import StatisticsGeocheckModal from '../statistics/StatisticsGeocheckModal.vue'

const apiMocks = vi.hoisted(() => ({
  getNodeGeocheck: vi.fn(),
  getStatistics: vi.fn(),
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
  price: { currency: 'TXB', minor: '100', display: '1.00 TXB' }, visible: true, upstreamPresent: true,
  activationRequired: false, accessibleNodes: [node], stockRemaining: 4,
  createdAt: '2026-08-21T00:00:00Z', updatedAt: '2026-08-21T00:00:00Z',
}
const money = { currency: 'TXB', minor: '0', display: '0.00 TXB' } as const
const statistics: StatisticsSnapshot = {
  generatedAt: '2026-08-21T00:00:00Z', remoteGeneratedAt: '2026-08-21T00:00:00Z', databaseGeneratedAt: '2026-08-21T00:00:00Z', stalePartitions: [],
  remote: { weeklyUserIncrease: 0, monthlyAverageUsagePercent: 0, predictedAverageRollover: money, trafficDates: [], trafficSeries: [] },
  database: {
    newUserConversionPercent: 0, averageSpend: money, spendMinimum: money, spendMaximum: money, subscriptionStates: [], averageCheckInReward: money,
    comboShares: [], groupMessagesTotal: 0, averageOptionalSquads: 0, paymentStatuses: [], databaseBytes: '0', comboBySquad: [],
    squadByCombo: [{ id: 'combo-1', label: 'Core', segments: [{ id: squad.remnaSquadUuid, label: squad.name, value: 12 }] }],
  },
}

function mountStep() {
  return shallowMount(CatalogSquadStep, {
    props: { comboId: 'combo-1', squads: [squad], selectedIds: [], includedIds: [] },
  })
}

describe('CatalogSquadStep', () => {
  beforeEach(() => vi.clearAllMocks())

  it('hands the selected combo leaders to the selector', async () => {
    apiMocks.getStatistics.mockResolvedValue(statistics)
    const wrapper = mountStep()
    await flushPromises()

    expect(wrapper.getComponent(SquadSelector).props('featuredIds')).toEqual([squad.id])
  })

  it('keeps squad selection available when statistics fail', async () => {
    apiMocks.getStatistics.mockRejectedValue(new Error('unavailable'))
    const wrapper = mountStep()
    expect(wrapper.getComponent(SquadSelector).props('featuredIds')).toEqual([])

    await flushPromises()
    expect(wrapper.getComponent(SquadSelector).props('featuredIds')).toEqual([])
  })

  it('opens the shared modal for the exact catalog node', async () => {
    const result = { nodeUuid: node.uuid, checkedAt: '2026-08-21T00:00:00Z', image: { format: 'svg', mediaType: 'image/svg+xml', encoding: 'base64', data: 'PHN2Zy8+' } }
    apiMocks.getStatistics.mockResolvedValue(statistics)
    apiMocks.getNodeGeocheck.mockResolvedValue(result)
    const wrapper = mountStep()

    wrapper.getComponent(SquadSelector).vm.$emit('openGeocheck', node)
    await flushPromises()

    expect(apiMocks.getNodeGeocheck).toHaveBeenCalledWith(node.uuid)
    expect(wrapper.getComponent(StatisticsGeocheckModal).props('node')).toBe(node)
    expect(wrapper.getComponent(StatisticsGeocheckModal).props('result')).toEqual(result)
  })
})
