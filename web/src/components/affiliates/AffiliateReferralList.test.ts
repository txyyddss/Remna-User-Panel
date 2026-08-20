import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import AffiliateReferralList from './AffiliateReferralList.vue'

describe('AffiliateReferralList', () => {
  it('renders a Telegram full name with one pending state and emits a server page selection', async () => {
    const wrapper = mount(AffiliateReferralList, { props: { loading: false, page: { page: 1, pageSize: 5, total: 6, totalPages: 2, items: [{ firstName: 'Ada', lastName: 'Lovelace', registeredAt: '2026-08-20T00:00:00Z', status: 'pending' }] } } })
    expect(wrapper.text()).toContain('Ada Lovelace')
    expect(wrapper.text().match(/Pending/g)).toHaveLength(1)
    wrapper.getComponent({ name: 'Pagination' }).vm.$emit('update:page', 2)
    expect(wrapper.emitted('update:page')).toEqual([[2]])
  })
})
