import type { ConnectionIP } from '@/api/types'

export interface ConnectionTarget {
  nodeName: string
  countryCode: string
  connection: ConnectionIP
}
