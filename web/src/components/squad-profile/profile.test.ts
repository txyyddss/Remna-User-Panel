import { describe, expect, it } from 'vitest'

import { countryOptions } from './profile'

describe('squad profile display helpers', () => {
  it('provides localized ISO country options without hardcoded labels', () => {
    const options = countryOptions('en')
    expect(options.find((item) => item.value === 'TW')?.label).toBe('Taiwan')
    expect(options.find((item) => item.value === 'US')?.label).toBe('United States')
  })
})
