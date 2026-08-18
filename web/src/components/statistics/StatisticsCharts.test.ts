import { mount } from '@vue/test-utils'
import { h } from 'vue'
import { describe, expect, it } from 'vitest'
import UApp from '@nuxt/ui/runtime/components/App.vue'

import type { StatisticsSnapshot } from '@/api/types'
import StatisticsDistribution from './StatisticsDistribution.vue'
import StatisticsDonut from './StatisticsDonut.vue'
import StatisticsPie from './StatisticsPie.vue'

const shares = [
  { id: 'active', label: 'Active', value: 3 },
  { id: 'queued', label: 'Queued', value: 1 },
]

describe('statistics SVG charts', () => {
  it('renders a donut with real SVG ring segments and center text', () => {
    const wrapper = mount(StatisticsDonut, {
      props: { items: shares, centerLabel: 'Members', centerValue: '4', chartLabel: 'Member split' },
    })

    expect(wrapper.find('svg[role="img"]').attributes('aria-label')).toBe('Member split')
    expect(wrapper.findAll('.statistics-ring-segment')).toHaveLength(2)
    expect(wrapper.find('.statistics-donut__center').text()).toContain('4')
    expect(wrapper.html()).not.toContain('conic-gradient')
  })

  it('renders outside labels for the DB4 pie and a complete label list', () => {
    const wrapper = mount(StatisticsPie, {
      props: { items: shares, chartLabel: 'Subscription states', sliceLabels: true },
    })

    expect(wrapper.findAll('.statistics-pie > g > path:not(.statistics-pie__line)')).toHaveLength(2)
    expect(wrapper.findAll('.statistics-pie__label')).toHaveLength(2)
    expect(wrapper.findAll('.statistics-legend li')).toHaveLength(2)
  })

  it('normalizes each DB12 stacked bar to one hundred percent', () => {
    const zero = { currency: 'TXB', minor: '0', display: '0.00 TXB' } as const
    const database = {
      newUserConversionPercent: 0, averageSpend: zero, spendMinimum: zero, spendMaximum: zero,
      subscriptionStates: [], averageRollover: zero, averageCheckInReward: zero, comboShares: [],
      groupMessagesTotal: 0, averageOptionalSquads: 0, paymentStatuses: [], databaseBytes: '0',
      squadByCombo: [],
      comboBySquad: [{
        id: 'squad-a', label: 'Squad A',
        segments: [{ id: 'combo-a', label: 'Combo A', value: 2 }, { id: 'combo-b', label: 'Combo B', value: 1 }],
      }],
    } satisfies StatisticsSnapshot['database']
    const wrapper = mount(UApp, {
      global: { stubs: { UTabs: true, UIcon: true } },
      slots: { default: () => h(StatisticsDistribution, { database }) },
    })
    const segments = wrapper.findAll('.statistics-distribution__segment')

    expect(segments).toHaveLength(2)
    expect(Number(segments[0]?.attributes('width'))).toBeCloseTo(66.667, 2)
    expect(Number(segments[1]?.attributes('x'))).toBeCloseTo(66.667, 2)
    expect(segments.reduce((total, segment) => total + Number(segment.attributes('width')), 0)).toBeCloseTo(100)
  })
})
