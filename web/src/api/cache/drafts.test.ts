import { describe, expect, it } from 'vitest'
import { mergeRefreshedDraft } from './drafts'

describe('cached form refresh', () => {
  it('updates unchanged fields while preserving edits made during the request', () => {
    const previous = { name: 'Carpool Pay', endpoint: 'https://old.example', channels: ['alipay'] }
    const draft = { ...previous, channels: ['wxpay'] }
    mergeRefreshedDraft(draft, { name: 'TX Payments', endpoint: 'https://new.example', channels: ['alipay', 'wxpay'] }, previous)
    expect(draft).toEqual({ name: 'TX Payments', endpoint: 'https://new.example', channels: ['wxpay'] })
  })
})
