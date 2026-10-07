import '@/styles/main.css'
import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import ui from '@nuxt/ui/vue-plugin'
import { abuseApi, type AbusePolicy, type AbuseRule } from '@/api/abuse'
import { t } from '@/i18n'
import AbuseAudit from './AbuseAudit.vue'

const parameters = new URLSearchParams(location.search)
if (parameters.has('reduced')) {
  const nativeMatchMedia = window.matchMedia.bind(window)
  window.matchMedia = query => {
    const media = nativeMatchMedia(query)
    if (query === '(prefers-reduced-motion: reduce)') Object.defineProperty(media, 'matches', { value: true })
    return media
  }
}
const now = new Date().toISOString()
const state = {
  saves: 0, deletions: 0, whitelistUpdates: 0,
  policy: { outboundTags: ['direct'], globalEnabled: true, globalLimit: 20, streakSeconds: 10,
    warningValidityDays: 7, warningCooldownMinutes: 60, revision: 1 } as AbusePolicy,
  rules: [{ id: 'audit-rule', name: 'Blocked destination', expression: 'example.net', qpsLimit: 10,
    enabled: true, revision: 1 }] as AbuseRule[],
}
async function pause(): Promise<void> {
  if (parameters.has('slow')) await new Promise(resolve => setTimeout(resolve, Number(parameters.get('slow')) || 2500))
  if (parameters.has('error')) throw new Error('Constructed abuse loading failure')
}
abuseApi.policy = async () => { await pause(); return state.policy }
abuseApi.savePolicy = async value => { state.saves++; state.policy = { ...value, revision: value.revision + 1 }; return state.policy }
abuseApi.nodes = async () => {
  await pause()
  return parameters.has('empty') ? [] : [{ uuid: 'audit-node', name: 'Tokyo edge', rotatedAt: now, lastReportAt: now }]
}
abuseApi.rules = async () => { await pause(); return parameters.has('empty') ? [] : state.rules }
abuseApi.punishments = async () => { await pause(); return [] }
abuseApi.adminRecords = async () => { await pause(); return { items: [], nextCursor: '' } }
abuseApi.statistics = async () => { await pause(); return { average: 12, minimum: 2, maximum: 20 } }
abuseApi.whitelist = async () => { await pause(); return [] }
abuseApi.deleteRule = async id => { state.deletions++; state.rules = state.rules.filter(rule => rule.id !== id) }
abuseApi.setWhitelist = async () => { state.whitelistUpdates++ }
abuseApi.copyNodeKey = async () => ({ key: 'constructed-node-key' })
abuseApi.rotateNodeKey = async () => ({ key: 'constructed-rotated-key' })
Object.assign(window, { __abuseAudit: state })
const app = createApp(AbuseAudit)
const router = createRouter({ history: createWebHistory(), routes: [{ path: '/fixtures/abuse-audit.html', component: AbuseAudit }] })
app.config.globalProperties.$t = t
app.use(router)
app.use(ui)
void router.isReady().then(() => app.mount('#app'))
