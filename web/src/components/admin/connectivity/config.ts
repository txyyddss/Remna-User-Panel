import { z } from 'zod'
import type { ConnectivityConfig } from '@/api/connectivity'

export const defaultConnectivityConfig: ConnectivityConfig = {
  scheduledEnabled: false, remnawaveUserId: 0, intervalSeconds: 300,
  timeoutSeconds: 15, probeUrl: 'https://cp.cloudflare.com/generate_204',
}

const configSchema = z.object({
  scheduledEnabled: z.boolean(),
  remnawaveUserId: z.number().int().nonnegative().max(Number.MAX_SAFE_INTEGER),
  intervalSeconds: z.number().int().min(60).max(86400),
  timeoutSeconds: z.number().int().min(1).max(60),
  probeUrl: z.string().max(2048).refine(value => {
    try {
      const url = new URL(value)
      return url.protocol === 'https:' && Boolean(url.hostname) && !url.username && !url.password
    } catch { return false }
  }),
}).refine(value => !value.scheduledEnabled || value.remnawaveUserId > 0)

export function validConnectivityConfig(config: ConnectivityConfig): boolean {
  return configSchema.safeParse(config).success
}

export function sameConnectivityConfig(a: ConnectivityConfig, b: ConnectivityConfig): boolean {
  return a.scheduledEnabled === b.scheduledEnabled && a.remnawaveUserId === b.remnawaveUserId
    && a.intervalSeconds === b.intervalSeconds && a.timeoutSeconds === b.timeoutSeconds && a.probeUrl === b.probeUrl
}
