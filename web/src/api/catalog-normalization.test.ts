import { describe, expect, it } from 'vitest'

import type { Catalog, PurchaseQuote } from './types'
import { normalizeCatalog, normalizePurchaseQuote } from './catalog-normalization'

describe('catalog response normalization', () => {
  it('normalizes missing and null catalog collections recursively', () => {
    const catalog = normalizeCatalog({
      combos: [
        { id: 'combo-1', includedSquads: null },
        {
          id: 'combo-2',
          includedSquads: [
            { id: 'included-1', accessibleNodes: null },
          ],
        },
      ],
      addons: [
        { id: 'addon-1', accessibleNodes: null },
      ],
      nodes: null,
    } as unknown as Catalog)

    expect(catalog.combos).toHaveLength(2)
    expect(catalog.combos[0]?.includedSquads).toEqual([])
    expect(catalog.combos[1]?.includedSquads[0]?.accessibleNodes).toEqual([])
    expect(catalog.addons[0]?.accessibleNodes).toEqual([])
    expect(catalog.nodes).toEqual([])
  })

  it('normalizes missing top-level combos and addons', () => {
    const catalog = normalizeCatalog({
      combos: undefined,
      addons: null,
      nodes: [],
    } as unknown as Catalog)

    expect(catalog.combos).toEqual([])
    expect(catalog.addons).toEqual([])
  })

  it('normalizes quote accessibleNodes', () => {
    const quote = normalizePurchaseQuote({
      comboId: 'combo-1',
      accessibleNodes: null,
    } as unknown as PurchaseQuote)

    expect(quote.accessibleNodes).toEqual([])
  })
})
