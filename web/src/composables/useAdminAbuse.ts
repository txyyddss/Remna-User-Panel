import { onMounted, shallowRef } from 'vue'

import { abuseApi, type AbuseNode, type AbusePolicy, type AbusePunishment, type AbuseRecord, type AbuseRule } from '@/api/abuse'
import { localizedError } from '@/i18n'

export function useAdminAbuse() {
  const policy = shallowRef<AbusePolicy | null>(null); const nodes = shallowRef<AbuseNode[]>([]); const rules = shallowRef<AbuseRule[]>([]); const punishments = shallowRef<AbusePunishment[]>([]); const records = shallowRef<AbuseRecord[]>([]); const statistics = shallowRef<Record<string, number>>({}); const whitelist = shallowRef<string[]>([]); const loading = shallowRef(true); const busy = shallowRef(false); const error = shallowRef<string | null>(null)
  async function load(): Promise<void> { loading.value = true; error.value = null; try { const [nextPolicy,nextNodes,nextRules,nextPunishments,page,nextStatistics,nextWhitelist] = await Promise.all([abuseApi.policy(), abuseApi.nodes(), abuseApi.rules(), abuseApi.punishments(), abuseApi.adminRecords(), abuseApi.statistics(), abuseApi.whitelist()]); policy.value=nextPolicy;nodes.value=nextNodes;rules.value=nextRules;punishments.value=nextPunishments;records.value=page.items;statistics.value=nextStatistics;whitelist.value=nextWhitelist } catch (caught) { error.value = localizedError(caught, 'adminAbuse.loadFailed') } finally { loading.value = false } }
  async function save(work: () => Promise<void>): Promise<void> { busy.value = true; error.value = null; try { await work(); await load() } catch (caught) { error.value = localizedError(caught, 'adminAbuse.saveFailed') } finally { busy.value = false } }
  onMounted(load); return { policy,nodes,rules,punishments,records,statistics,whitelist,loading,busy,error,load,save }
}
