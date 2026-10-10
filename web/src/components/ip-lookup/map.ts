const mercatorLimit = 85.05112878

export function ipLocationMap(latitude: number | null | undefined, longitude: number | null | undefined) {
  if (latitude == null || longitude == null || !Number.isFinite(latitude) || !Number.isFinite(longitude)
    || Math.abs(latitude) > 90 || Math.abs(longitude) > 180) return null
  const center = Math.max(-mercatorLimit + 0.3, Math.min(mercatorLimit - 0.3, latitude))
  const west = Math.max(-180, Math.min(179.4, longitude - 0.3))
  const bbox = [west, center - 0.3, west + 0.6, center + 0.3]
  const embed = new URL('https://www.openstreetmap.org/export/embed.html')
  embed.search = new URLSearchParams({ bbox: bbox.join(','), layer: 'mapnik', marker: `${latitude},${longitude}`, theme: 'dark' }).toString()
  const link = new URL('https://www.openstreetmap.org/')
  link.search = new URLSearchParams({ mlat: String(latitude), mlon: String(longitude) }).toString()
  link.hash = `map=10/${latitude}/${longitude}`
  return { embed: embed.toString(), link: link.toString() }
}
