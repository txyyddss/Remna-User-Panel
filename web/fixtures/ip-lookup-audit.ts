import '@/styles/main.css'
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import ui from '@nuxt/ui/vue-plugin'
import { api } from '@/api/client'
import { ApiError } from '@/api/http'
import { ipLookupApi, type IPLookupAdminSettings, type IPLookupCheck, type IPLookupQuote, type IPLookupReport, type IPLookupState } from '@/api/ipLookup'
import { useSessionStore } from '@/stores/session'
import { setLocale, t } from '@/i18n'
import IPLookupPage from '@/components/ip-lookup/IPLookupPage.vue'
import AdminIPLookupPage from '@/components/admin/ip-lookup/AdminIPLookupPage.vue'
import IPLookupAudit from './IPLookupAudit.vue'

const params = new URLSearchParams(location.search)
const mode = params.get('mode') ?? 'empty'
const ids = ['abuseipdb', 'scamalytics', 'ipapi', 'maxmind', 'ipqs', 'ip2location'] as const
const money = (minor: number) => ({ currency: 'TXB' as const, minor: String(minor), display: `${(minor / 100).toFixed(2)} TXB` })
const now = () => new Date().toISOString()
const state: IPLookupState = { enabled: mode !== 'disabled', lookupFee: money(250), refreshFee: money(350), allowance: params.has('no-combo') ? null : { purchaseId: 'term', total: 2, remaining: 2, validUntil: '2026-12-01T00:00:00Z' } }
const settings: IPLookupAdminSettings = { enabled: false, lookupFeeTxb: '', refreshFeeTxb: '', providers: ids.map(id => ({ id, enabled: false, accountId: id === 'maxmind' ? '123456' : id === 'scamalytics' ? 'fixture-account' : '' })), credentials: {}, configured: Object.fromEntries(ids.map(id => [id, true])), comboQuotas: { 'standard': null, 'weekend': 0 } }
const facts = { country: 'JP', region: 'Tokyo', city: 'Tokyo', isp: 'Sony Network Communications Inc.', asn: '2527', networkType: 'residential' }
let cached = params.has('cached')
let polls = 0
let submitCount = 0
let savedBody: IPLookupAdminSettings | null = null
let lastQuote: IPLookupQuote | null = null
const receipt = () => ({ id: 'check-audit', kind: 'ip_reputation_lookup', status: 'queued' as const, errorCode: null, createdAt: now(), updatedAt: now(), completedAt: null })

async function pause() {
  if (mode === 'loading') await new Promise(resolve => setTimeout(resolve, 5000))
  if (mode === 'error') throw new ApiError(502, { code: 'IP_LOOKUP_FAILED', message: 'Constructed provider outage' })
}
ipLookupApi.state = async () => { await pause(); return structuredClone(state) }
ipLookupApi.settings = async () => { await pause(); return structuredClone(settings) }
ipLookupApi.saveSettings = async value => {
  savedBody = JSON.parse(JSON.stringify(value)) as IPLookupAdminSettings
  if (mode === 'save-error' || (value.enabled && (value.lookupFeeTxb === '' || value.refreshFeeTxb === '' || Object.values(value.comboQuotas).some(q => q === null)))) throw new ApiError(422, { code: 'IP_LOOKUP_CONFIGURATION_REQUIRED', message: 'Constructed missing configuration' })
  Object.assign(settings, JSON.parse(JSON.stringify(value)), { credentials: {} })
  state.enabled = value.enabled
  return structuredClone(settings)
}
ipLookupApi.quote = async (ip, refresh) => {
  const useQuota = !refresh && !cached && !!state.allowance?.remaining
  return { ip, refresh, charge: money(useQuota || (!refresh && cached) ? 0 : refresh ? 350 : 250), useQuota, purchaseId: state.allowance?.purchaseId ?? '', remaining: state.allowance?.remaining ?? 0, cacheReportId: cached ? 'report-audit' : '', configHash: 'fixture-config', expiresAt: Math.floor(Date.now() / 1000) + 120, token: 'fixture-token' }
}
ipLookupApi.submit = async quote => {
  submitCount++
  lastQuote = structuredClone(quote)
  if (mode === 'balance-error') throw new ApiError(409, { code: 'INSUFFICIENT_BALANCE', message: 'Constructed low balance' })
  if (quote.useQuota && state.allowance) state.allowance.remaining--
  polls = 0
  return receipt()
}
ipLookupApi.check = async () => {
  polls++
  const operation = receipt()
  const report: IPLookupReport = { id: 'report-audit', ip: lastQuote?.ip ?? '150.249.241.62', status: 'succeeded', refundRequired: mode === 'partial' || mode === 'all-failed', verdict: mode === 'risk' ? 'unsuitable' : mode === 'partial' ? 'inconclusive' : 'suitable', reasons: mode === 'risk' ? ['abuseipdb:abuse_reports'] : mode === 'partial' ? ['incomplete_coverage'] : [], facts, sources: Object.fromEntries(Object.keys(facts).map(k => [k, 'ipapi'])), providers: ids.map(id => ({ id, riskLevel: id === 'scamalytics' ? 'low' : '', status: 'success', complete: true, errorCode: '', checkedAt: now(), facts: id === 'ipqs' ? { ...facts, city: 'Chiyoda City' } : facts, signals: { abuse: false, datacenter: false, vpn: false, proxy: false, tor: false }, scores: id === 'scamalytics' ? { fraud_score: 5, isp_risk_score: 5 } : id === 'ipapi' ? { company_abuse_ratio: 0.0002, asn_abuse_ratio: 0.0003 } : { fraud_score: 0 }, reports: id === 'abuseipdb' ? 0 : null })), checkedAt: now(), policyVersion: 'residential-v1', parserVersion: 'residential-v1' }
  if (mode === 'risk') { report.providers[0]!.reports = 1; report.providers.slice(1).forEach(p => { p.status = 'skipped' }) }
  if (mode === 'scamalytics-risk') { report.verdict = 'unsuitable'; report.reasons = ['scamalytics:fraud_score']; report.providers[1]!.riskLevel = 'high'; report.providers[1]!.scores = { fraud_score: 85, isp_risk_score: 5 }; report.providers.slice(2).forEach(p => { p.status = 'skipped' }) }
  if (mode === 'partial') { report.status = 'partial'; report.providers[3]!.status = 'error'; report.providers[3]!.complete = false; report.providers[3]!.signals = { abuse: null, datacenter: null, vpn: null, proxy: null, tor: null } }
  const pending = polls < 3
  const failed = mode === 'all-failed'
  const result: IPLookupCheck = { operation: { ...operation, status: pending ? 'processing' : failed ? 'failed' : report.status, errorCode: !pending && failed ? 'IP_LOOKUP_ALL_PROVIDERS_FAILED' : null, completedAt: pending ? null : now() }, report: pending || failed ? null : report, cached: cached && !lastQuote?.refresh, charge: failed || mode === 'partial' ? money(0) : lastQuote?.charge ?? money(0), usedQuota: lastQuote?.useQuota ?? false, refunded: !pending && (failed || mode === 'partial') }
  if (!pending && !failed) cached = true
  if (!pending && (failed || mode === 'partial') && lastQuote?.useQuota && state.allowance) state.allowance.remaining++
  return result
}
api.getCommunityAccess = async () => ({ activeCombo: false })
api.getAdminResource = async () => ({ items: ['standard', 'weekend'].map((id, index) => ({ id, name: index ? 'Weekend' : 'Standard', ipLookupQuota: settings.comboQuotas[id] })) }) as never
const app = createApp(IPLookupAudit)
app.config.globalProperties.$t = t
const pinia = createPinia()
app.use(pinia)
const session = useSessionStore(pinia)
session.session = { authenticated: true, user: { id: 'ip-audit-member', telegramId: '96001', firstName: 'Morgan', username: 'morgan', telegramUsername: 'morgan', role: params.has('admin') ? 'admin' : 'member', onboardingState: 'complete' } } as never
session.status = 'ready'
const router = createRouter({ history: createWebHistory(), routes: [{ path: '/:pathMatch(.*)*', component: params.has('admin') ? AdminIPLookupPage : IPLookupPage }] })
app.use(router)
app.use(ui)
setLocale(params.has('zh') ? 'zh-CN' : 'en')
document.title = t('ipLookup.title')
;(window as unknown as { __ipAudit: object }).__ipAudit = { state, settings, counters: () => ({ submitCount, polls, savedBody }), setCached: (value: boolean) => { cached = value } }
void router.isReady().then(() => app.mount('#app'))
