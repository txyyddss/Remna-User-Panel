import { t } from '@/i18n'

export const factKeys = ['country', 'region', 'city', 'isp', 'asn', 'networkType'] as const
export const riskKeys = ['abuse', 'datacenter', 'vpn', 'proxy', 'tor'] as const
const knownErrors = new Set(['IP_LOOKUP_INVALID_IP', 'IP_LOOKUP_DISABLED', 'IP_LOOKUP_QUOTE_CHANGED', 'IP_LOOKUP_REFRESH_UNAVAILABLE', 'IP_LOOKUP_CONFIGURATION_REQUIRED', 'IP_LOOKUP_CREDENTIAL_REQUIRED', 'IP_LOOKUP_INVALID_CONFIG', 'IP_LOOKUP_ALL_PROVIDERS_FAILED', 'IP_LOOKUP_BUSY', 'IP_LOOKUP_CONFLICT', 'INSUFFICIENT_BALANCE'])

export function ipLookupErrorText(code: string): string {
  return t(`ipLookup.errors.${knownErrors.has(code) ? code : 'IP_LOOKUP_FAILED'}`)
}

export function providerName(id: string): string { return t(`ipLookup.providers.${id}`) }
export function reasonText(reason: string): string {
  const [source, value] = reason.includes(':') ? reason.split(':') : ['', reason]
  const text = t(`ipLookup.reasons.${value}`)
  return source ? `${providerName(source)}: ${text}` : text
}
