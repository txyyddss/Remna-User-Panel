import { localizedError, t } from '@/i18n'
import type { ConnectivityAttempt } from '@/api/connectivity'

export function connectivityError(caught: unknown, fallback = 'hostConnectivity.loadFailed'): string {
  const code = typeof caught === 'string' ? caught : caught && typeof caught === 'object' && 'code' in caught ? String(caught.code) : ''
  const key = `hostConnectivity.codes.${code}`
  const message = t(key)
  return message !== key ? message : localizedError(caught, fallback)
}

export function connectivityTone(status?: ConnectivityAttempt['status']): 'neutral' | 'success' | 'warning' | 'danger' {
  if (status === 'connected') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'error' || status === 'unsupported' || status === 'interrupted') return 'warning'
  return 'neutral'
}
