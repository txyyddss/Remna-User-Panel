import { describe, expect, it } from 'vitest'

import { draftFromProfile, profileFromDraft } from './profileForm'

describe('squad profile form mapping', () => {
  it('maps unlimited international ports and comma-separated carriers', () => {
    const draft = draftFromProfile(null)
    draft.countryCode = 'TW'
    draft.upstreamCarriers = 'Carrier A, Carrier B, Carrier A'
    const profile = profileFromDraft(draft)
    expect(profile).toEqual({
      type: 'international_network',
      portMbps: null,
      countryCode: 'TW',
      upstreamCarriers: ['Carrier A', 'Carrier B'],
    })
  })

  it('does not create an incomplete broadband profile', () => {
    const draft = draftFromProfile(null)
    draft.type = 'broadband'
    expect(profileFromDraft(draft)).toBeNull()
  })
})
