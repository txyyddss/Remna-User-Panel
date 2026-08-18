import type { AdminEntitlement, AdminUserDetail } from '@/api/adminOperations'

export function adminUserName(detail: AdminUserDetail): string {
  const name = [detail.user.firstName, detail.user.lastName].filter(Boolean).join(' ').trim()
  return name || detail.user.username || detail.user.telegramUsername || detail.user.telegramId
}

export function toLocalDateTime(value: string): string {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return ''
  const offset = parsed.getTimezoneOffset() * 60_000
  return new Date(parsed.getTime() - offset).toISOString().slice(0, 16)
}

export function fromLocalDateTime(value: string): string {
  const parsed = new Date(value)
  return Number.isNaN(parsed.getTime()) ? '' : parsed.toISOString()
}

export function entitlementTone(status: AdminEntitlement['status']): 'neutral' | 'success' | 'warning' | 'danger' {
  if (status === 'active') return 'success'
  if (status === 'activating' || status === 'queued') return 'warning'
  if (status === 'failed') return 'danger'
  return 'neutral'
}
