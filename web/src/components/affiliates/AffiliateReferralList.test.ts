import { mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'

import AffiliateReferralList from './AffiliateReferralList.vue'

describe('AffiliateReferralList', () => {
  afterEach(() => vi.clearAllMocks())

  it('renders a Telegram full name with one pending state and emits a server page selection', async () => {
    const selectionChanged = vi.fn()
    const feedback = window.Telegram?.WebApp?.HapticFeedback
    if (!feedback) throw new Error('Telegram haptic fixture is unavailable')
    feedback.selectionChanged = selectionChanged
    const wrapper = mount(AffiliateReferralList, { props: { loading: false, page: { page: 1, pageSize: 5, total: 6, totalPages: 2, items: [{ firstName: 'Ada', lastName: 'Lovelace', registeredAt: '2026-08-20T00:00:00Z', status: 'pending' }] } } })
    expect(wrapper.text()).toContain('Ada Lovelace')
    expect(wrapper.text().match(/Pending/g)).toHaveLength(1)
    const pagination = wrapper.getComponent({ name: 'Pagination' })
    expect(pagination.attributes('data-haptic')).toBeUndefined()
    pagination.vm.$emit('update:page', 1)
    expect(selectionChanged).not.toHaveBeenCalled()
    pagination.vm.$emit('update:page', 2)
    expect(wrapper.emitted('update:page')).toEqual([[2]])
    expect(selectionChanged).toHaveBeenCalledOnce()
  })
})
