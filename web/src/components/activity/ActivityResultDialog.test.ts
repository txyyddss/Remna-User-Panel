import { defineComponent } from 'vue'
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import type { ActivityResult } from '@/api/features'

import ActivityResultDialog from './ActivityResultDialog.vue'
import BetSuccessFireworks from './BetSuccessFireworks.vue'

const ModalStub = defineComponent({
  props: { open: Boolean },
  emits: ['update:open'],
  template: '<div data-modal><slot name="body" /><slot name="footer" /></div>',
})

function result(kind: ActivityResult['kind'], outcome: ActivityResult['outcome']): ActivityResult {
  return {
    id: `${kind}-${outcome}`,
    kind,
    outcome,
    message: '',
    reward: { kind: 'none' },
    balanceAfter: { currency: 'TXB', minor: '100', display: '1.00 TXB' },
    createdAt: '2026-08-11T00:00:00Z',
  }
}

function mountDialog(activityResult: ActivityResult) {
  return mount(ActivityResultDialog, {
    props: { result: activityResult },
    global: {
      stubs: { UIcon: true, UButton: true, UModal: ModalStub },
    },
  })
}

describe('ActivityResultDialog', () => {
  it('shows fireworks only for a winning bet', () => {
    expect(mountDialog(result('bet', 'win')).findComponent(BetSuccessFireworks).exists()).toBe(true)
    expect(mountDialog(result('bet', 'loss')).findComponent(BetSuccessFireworks).exists()).toBe(false)
    expect(mountDialog(result('check_in', 'complete')).findComponent(BetSuccessFireworks).exists()).toBe(false)
    expect(mountDialog(result('draw', 'complete')).findComponent(BetSuccessFireworks).exists()).toBe(false)
  })
})
