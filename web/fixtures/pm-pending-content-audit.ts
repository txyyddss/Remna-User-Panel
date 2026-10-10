import '@/styles/main.css'
import { createApp } from 'vue'
import { createRouter, createMemoryHistory } from 'vue-router'
import ui from '@nuxt/ui/vue-plugin'
import { pmApi } from '@/api/pm'
import { setLocale, t } from '@/i18n'
import PMPendingContentAudit from './PMPendingContentAudit.vue'
const date = new Date().toISOString()
pmApi.deliveries = async () => ({ items: ['PM_CONTENT_EXPIRED', 'PM_CONTENT_INVALID'].map((errorCode, index) => ({ operationId: `pending-content-${index}`, status: 'failed', direction: 'outbound', sourceChatId: '-100123', sourceMessageId: String(50 + index), resultMessageId: null, topicId: '50', errorCode, readAt: null, readSource: '', createdAt: date })) })
setLocale(new URLSearchParams(location.search).has('zh') ? 'zh-CN' : 'en')
const app = createApp(PMPendingContentAudit)
app.config.globalProperties.$t = t
app.use(ui)
app.use(createRouter({ history: createMemoryHistory(), routes: [{ path: '/:pathMatch(.*)*', component: PMPendingContentAudit }] }))
app.mount('#app')
