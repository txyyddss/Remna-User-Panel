import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import type { StatisticsNode, StatisticsNodesSnapshot, StatisticsSnapshot } from '@/api/types'
import StatisticsPage from './StatisticsPage.vue'

const apiMocks = vi.hoisted(() => ({
  getNodeGeocheck: vi.fn(),
  getStatistics: vi.fn(),
  getStatisticsNodes: vi.fn(),
}))

vi.mock('@/api/client', () => ({ api: apiMocks }))
vi.mock('@/composables/useTelegramBackButton', () => ({ useTelegramBackButton: () => undefined }))

const node: StatisticsNode = {
  uuid: '373f14bc-089a-4c3a-91c3-3421e7c83367', name: 'Tokyo', countryCode: 'JP', online: true,
  usersOnline: 1, rxBytesPerSec: '2', txBytesPerSec: '3', xrayVersion: '1.0', multiplier: 1,
}
const nodes: StatisticsNodesSnapshot = { generatedAt: '2026-08-19T12:00:00Z', stale: false, nodes: [node] }
const money = { currency: 'TXB', minor: '0', display: '0.00 TXB' } as const
const statistics: StatisticsSnapshot = {
  generatedAt: '2026-08-19T12:00:00Z', remoteGeneratedAt: '2026-08-19T12:00:00Z', databaseGeneratedAt: '2026-08-19T12:00:00Z', stalePartitions: [],
  remote: { weeklyUserIncrease: 0, monthlyAverageUsagePercent: 0, trafficDates: [], trafficSeries: [] },
  database: { newUserConversionPercent: 0, averageSpend: money, spendMinimum: money, spendMaximum: money, subscriptionStates: [], averageRollover: money, averageCheckInReward: money, comboShares: [], groupMessagesTotal: 0, averageOptionalSquads: 0, paymentStatuses: [], databaseBytes: '0', squadByCombo: [], comboBySquad: [] },
}

describe('StatisticsPage', () => {
  it('passes a selected card image into the Geocheck modal', async () => {
    apiMocks.getStatistics.mockResolvedValue(statistics)
    apiMocks.getStatisticsNodes.mockResolvedValue(nodes)
    apiMocks.getNodeGeocheck.mockResolvedValue({ nodeUuid: node.uuid, checkedAt: nodes.generatedAt, image: { format: 'svg', mediaType: 'image/svg+xml', encoding: 'base64', data: 'PHN2Zy8+' } })
    const wrapper = mount(StatisticsPage, {
      global: {
        stubs: {
          InlineNotice: true, SkeletonBlock: true, StatisticsFreshness: true, StatisticsOverview: true,
          StatisticsTrafficChart: true, StatisticsShareCharts: true, StatisticsDistribution: true, UTooltip: true, UButton: true,
          StatisticsNodes: { props: ['snapshot'], emits: ['openGeocheck'], template: '<div data-testid="geocheck-card" @click="$emit(\'openGeocheck\', snapshot.nodes[0])"></div>' },
          StatisticsGeocheckModal: { props: ['open', 'node', 'result'], template: '<output data-testid="geocheck-modal" :data-open="String(open)" :data-node="node && node.uuid" :data-image="result && result.image.data" />' },
        },
        mocks: { $t: (key: string) => key },
      },
    })
    await flushPromises()
    expect(apiMocks.getNodeGeocheck).not.toHaveBeenCalled()

    await wrapper.get('[data-testid="geocheck-card"]').trigger('click')
    await flushPromises()

    const modal = wrapper.get('[data-testid="geocheck-modal"]')
    expect(apiMocks.getNodeGeocheck).toHaveBeenCalledWith(node.uuid)
    expect(modal.attributes('data-open')).toBe('true')
    expect(modal.attributes('data-node')).toBe(node.uuid)
    expect(modal.attributes('data-image')).toBe('PHN2Zy8+')
    wrapper.unmount()
  })
})
