import { onMounted, shallowRef } from 'vue'

import { abuseApi, type AbuseRecord } from '@/api/abuse'
import { localizedError } from '@/i18n'

export function useAbuseRecords() {
  const records = shallowRef<AbuseRecord[]>([])
  const loading = shallowRef(true)
  const error = shallowRef<string | null>(null)
  async function load(): Promise<void> { loading.value = true; error.value = null; try { records.value = (await abuseApi.records()).items } catch (caught) { error.value = localizedError(caught, 'abuse.loadFailed') } finally { loading.value = false } }
  onMounted(load)
  return { records, loading, error, load }
}
