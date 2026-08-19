import { describe, expect, it, vi } from 'vitest'

import { api } from '@/api/client'
import type { StatisticsNode } from '@/api/types'
import { useNodeGeocheck } from './useNodeGeocheck'

vi.mock('@/api/client', () => ({ api: { getNodeGeocheck: vi.fn() } }))
vi.mock('@/i18n', () => ({ localizedError: () => 'unavailable' }))

const node: StatisticsNode = { uuid: '373f14bc-089a-4c3a-91c3-3421e7c83367', name: 'Tokyo', countryCode: 'JP', online: true, usersOnline: 1, rxBytesPerSec: '2', txBytesPerSec: '3', xrayVersion: '1.0', multiplier: 1 }

describe('useNodeGeocheck', () => {
  it('loads an image only after explicit selection and clears it when closed', async () => {
    vi.mocked(api.getNodeGeocheck).mockResolvedValue({ nodeUuid: node.uuid, checkedAt: '2026-08-19T12:00:00Z', image: { format: 'svg', mediaType: 'image/svg+xml', encoding: 'base64', data: 'PHN2Zy8+' } })
    const geocheck = useNodeGeocheck()
    expect(api.getNodeGeocheck).not.toHaveBeenCalled()
    await geocheck.show(node)
    expect(api.getNodeGeocheck).toHaveBeenCalledWith(node.uuid)
    expect(geocheck.result.value?.image.data).toBe('PHN2Zy8+')
    geocheck.close()
    expect(geocheck.result.value).toBeNull()
    expect(geocheck.isOpen.value).toBe(false)
  })

  it('keeps the unavailable state localized after a failed request', async () => {
    vi.mocked(api.getNodeGeocheck).mockRejectedValue(new Error('missing'))
    const geocheck = useNodeGeocheck()
    await geocheck.show(node)
    expect(geocheck.error.value).toBeTruthy()
    expect(geocheck.loading.value).toBe(false)
  })
})
