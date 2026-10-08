const nativeFetch = window.fetch.bind(window)
const params = new URLSearchParams(location.search)
const accountNames = { 7: 'monitoring_account', 9: 'monitoring_other' }
const config = { scheduledEnabled: false, remnawaveUserId: 7, intervalSeconds: 300, timeoutSeconds: 15, probeUrl: 'https://cp.cloudflare.com/generate_204' }
const hostIds = ['11111111-1111-4111-8111-111111111111', '22222222-2222-4222-8222-222222222222', '33333333-3333-4333-8333-333333333333']
const now = () => new Date().toISOString()
const attempt = (index, status = 'connected', offset = 0) => ({ id: `aaaaaaaa-aaaa-4aaa-8aaa-${String(index + offset).padStart(12,'0')}`, runId: '99999999-9999-4999-8999-999999999999', hostUuid: hostIds[index % 3], configHash: 'a'.repeat(64), remnawaveUserId: 7, trigger: 'manual', startedAt: now(), finishedAt: now(), status, latencyMs: status === 'connected' ? 124.6 : null, httpStatus: status === 'connected' ? 204 : null, errorCode: status === 'failed' ? 'CONNECTIVITY_PROBE_TIMEOUT' : status === 'unsupported' ? 'CONNECTIVITY_UNSUPPORTED_TEMPLATE' : '' })
window.audit = { mode: params.get('mode') || 'populated', delay: Number(params.get('delay') || 0), snapshotError: params.has('snapshotError'), historyError: params.has('historyError'), saveError: params.has('saveError'), resolveError: params.has('resolveError'), checks: 0, writes: [], run: null, requests: [], config }
const response = (body, status = 200) => new Response(status === 204 ? null : JSON.stringify(body), { status, headers: { 'content-type': 'application/json' } })
window.fetch = async (input, options = {}) => {
  const uri = typeof input === 'string' ? input : input.url
  const url = new URL(uri, location.origin)
  if (!url.pathname.startsWith('/api/')) return nativeFetch(input, options)
  const method = options.method || input.method || 'GET'
  const body = options.body ? JSON.parse(options.body) : null
  const action = window.audit
  action.requests.push({ path: url.pathname, query: url.search, method, body })
  if (action.delay) await new Promise(resolve => setTimeout(resolve, action.delay))
  if (url.pathname === '/api/v1/admin/host-connectivity') {
    if (action.snapshotError) return response({ code: 'CONNECTIVITY_STORAGE_UNAVAILABLE', message: 'Fixture unavailable' }, 503)
    const unconfigured = action.mode === 'unconfigured' || !config.remnawaveUserId
    const accessible = action.mode === 'empty' || unconfigured ? [] : hostIds.map((id, index) => ({ hostUuid: id, remark: ['Tokyo transport', 'Frankfurt endpoint', 'Hidden subscription host'][index], address: `edge-${index + 1}.example`, port: 443, latest: attempt(index, ['connected', 'failed', 'unsupported'][index]) }))
    return response({ config: { ...config, remnawaveUserId: unconfigured ? 0 : config.remnawaveUserId }, user: unconfigured ? null : { id: config.remnawaveUserId, username: accountNames[config.remnawaveUserId], status: 'ACTIVE' }, run: action.run, hosts: accessible, stale: action.mode === 'stale', errorCode: action.mode === 'stale' ? 'CONNECTIVITY_SUBSCRIPTION_UNAVAILABLE' : unconfigured ? 'CONNECTIVITY_NOT_CONFIGURED' : action.mode === 'empty' ? 'CONNECTIVITY_NO_ACCESSIBLE_HOSTS' : '' })
  }
  if (url.pathname.endsWith('/host-connectivity/history')) {
    if (action.historyError) return response({ code: 'CONNECTIVITY_STORAGE_UNAVAILABLE', message: 'Fixture unavailable' }, 503)
    if (action.mode === 'empty' || action.mode === 'unconfigured') return response({ items: [], nextCursor: null })
    const selected = url.searchParams.get('hostUuid')
    const cursor = url.searchParams.get('cursor')
    const items = selected ? [attempt(hostIds.indexOf(selected), 'connected', cursor ? 20 : 10)] : [attempt(0, 'connected', cursor ? 20 : 10), attempt(1, 'failed', cursor ? 20 : 10)]
    return response({ items, nextCursor: cursor ? null : 'fixture-next-cursor' })
  }
  if (url.pathname.endsWith('/host-connectivity/test-user/resolve')) {
    if (action.resolveError) return response({ code: 'CONNECTIVITY_ACCOUNT_UNAVAILABLE', message: 'Fixture account unavailable' }, 400)
    accountNames[9] = body.username
    return response({ id: 9, username: body.username, status: 'ACTIVE' })
  }
  if (url.pathname.endsWith('/host-connectivity/checks')) {
    action.checks++
    action.run = { id: '99999999-9999-4999-8999-999999999999', status: 'running', trigger: 'manual', startedAt: now(), finishedAt: null, total: 3, completed: 0, errorCode: '' }
    setTimeout(() => { action.run = { ...action.run, status: 'completed', completed: 3, finishedAt: now() } }, 6000)
    return response(action.run, 202)
  }
  if (url.pathname === '/api/v1/admin/settings/connectivity.config' && method === 'PUT') {
    if (action.saveError) return response({ code: 'CONNECTIVITY_ACCOUNT_UNAVAILABLE', message: 'Fixture account unavailable' }, 400)
    Object.assign(config, JSON.parse(body.value)); action.writes.push(JSON.parse(body.value)); return response(null, 204)
  }
  if (url.pathname === '/api/v1/admin/settings') return response({ items: [{ key: 'connectivity.config', value: JSON.stringify(config), configured: true, encrypted: false, category: 'application', updatedAt: now() }] })
  const activity = { timezone: 'Asia/Shanghai', dailyRewardMinTxb: '0.00', dailyRewardMinTxbMinor: '0', dailyRewardMaxTxb: '0.00', dailyRewardMaxTxbMinor: '0', groupMessageThreshold: 0, groupMessageRewardTxb: '0.00', groupMessageRewardTxbMinor: '0' }
  if (url.pathname === '/api/v1/admin/activity-settings') return response(activity)
  const limits = { minimum: { currency: 'TXB', minor: '100' }, maximum: { currency: 'TXB', minor: '100000' }, updatedAt: now() }
  if (url.pathname === '/api/v1/balance') return response({ balance: { currency: 'TXB', minor: '10000' }, addAmountLimits: limits })
  if (url.pathname === '/api/v1/admin/billing/amount-limits') return response(limits)
  return response({ items: [] })
}
await import('/src/styles/main.css')
const [{ createApp, h }, { createPinia }, { default: ui }, { default: UApp }, { default: Component }, i18n, routerApi] = await Promise.all([
  import('vue'), import('pinia'), import('@nuxt/ui/vue-plugin'), import('@nuxt/ui/components/App.vue'),
  import(params.has('full') ? '/src/components/admin/AdminSettingsPanel.vue' : '/src/components/admin/connectivity/AdminConnectivitySettings.vue'), import('/src/i18n/index.ts'), import('vue-router'),
])
window.auditLocale = i18n.setLocale
const app = createApp({ render: () => h(UApp, null, { default: () => h('main', { style: { width: 'min(100%, 1000px)', margin: '0 auto', padding: '12px' } }, [h(Component, { ref: instance => { window.auditComponent = instance } })]) }) })
app.config.globalProperties.$t = i18n.t
const router = routerApi.createRouter({ history: routerApi.createMemoryHistory(), routes: [{ path: '/', component: { render: () => null } }] })
app.use(createPinia()); app.use(router); app.use(ui); await router.push('/'); await router.isReady(); app.mount('#app')
window.auditReady = true
