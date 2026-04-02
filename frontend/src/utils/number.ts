export function sanitizeDecimalInput(value: string): string {
  const compact = value.replace(/\s+/g, '')
  const cleaned = compact.replace(/[^\d.]/g, '')
  const firstDot = cleaned.indexOf('.')
  if (firstDot === -1) {
    return cleaned
  }
  return `${cleaned.slice(0, firstDot + 1)}${cleaned.slice(firstDot + 1).replace(/\./g, '')}`
}

export function parseSanitizedDecimal(value: string): number {
  const parsed = Number.parseFloat(sanitizeDecimalInput(value))
  return Number.isFinite(parsed) ? parsed : 0
}
