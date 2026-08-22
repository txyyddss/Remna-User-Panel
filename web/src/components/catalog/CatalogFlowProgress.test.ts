import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import CatalogFlowProgress from './CatalogFlowProgress.vue'
import { indicatorIcon } from './catalogFlowProgress'

function mountProgress(step: number) {
  return mount(CatalogFlowProgress, { props: { modelValue: step } })
}

describe('CatalogFlowProgress', () => {
  it('maps the one-based catalog step to completed and active states', () => {
    const wrapper = mountProgress(3)

    expect(wrapper.findAll('[data-slot="item"]').map((item) => item.attributes('data-state'))).toEqual([
      'completed',
      'completed',
      'active',
      'inactive',
    ])
  })

  it('shows checks only for completed steps', () => {
    expect(indicatorIcon(2, 3, 'i-ph-users-three')).toBe('i-ph-check-bold')
    expect(indicatorIcon(3, 3, 'i-ph-network')).toBe('i-ph-network')
    expect(indicatorIcon(4, 3, 'i-ph-ticket')).toBe('i-ph-ticket')
  })

  it('marks only completed step icons for the success color', () => {
    const wrapper = mountProgress(3)
    const icons = wrapper.findAll('[data-slot="icon"]')

    expect(icons.map((icon) => icon.classes().includes('catalog-progress__icon--completed'))).toEqual([
      true,
      true,
      false,
      false,
    ])
  })
})
