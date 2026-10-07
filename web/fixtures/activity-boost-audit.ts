import '@/styles/main.css'
import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import ui from '@nuxt/ui/vue-plugin'

import { featuresApi } from '@/api/features'
import type { ActivityOverview, ActivityResult } from '@/api/features'
import { ApiError } from '@/api/http'
import { setLocale, t } from '@/i18n'
import LuckyAudit from './LuckyAudit.vue'

const parameters = new URLSearchParams(location.search)
const state = {
  count: Number(parameters.get('count') ?? 0), unavailable: parameters.has('unavailable'), networkFailure: parameters.has('error'), delay: Number(parameters.get('delay') ?? 0),
  checkedIn: false, rewarded: false, enabled: true, messageCount: 0,
  loads: 0, claims: 0, balanceMinor: 10000,
}
const link = 'https://t.me/boost?c=123456'
featuresApi.getActivity = async (): Promise<ActivityOverview> => {
  state.loads++
  if (state.delay) await new Promise(resolve => setTimeout(resolve, state.delay))
  if (state.networkFailure) throw new Error('Constructed activity network failure')
  return {
    balance: { currency: 'TXB', minor: String(state.balanceMinor), display: '' },
    timeZone: 'Asia/Shanghai', checkedInToday: state.checkedIn,
    dailyRewardMinTxbMinor: String(Math.floor((125 * state.count + 1) / 2)),
    dailyRewardMaxTxbMinor: String(Math.floor((225 * state.count + 1) / 2)),
    groupBoost: {
      state: state.unavailable ? 'unavailable' : state.count ? 'boosted' : 'required',
      count: state.unavailable ? null : state.count, boostUrl: link,
    },
    games: [], draws: [], recentResults: [],
    groupMessageReward: {
      enabled: state.enabled, localDate: '2026-10-07', threshold: state.enabled ? 5 : 0,
      messageCount: state.messageCount, rewarded: state.rewarded,
      rewardMinor: String(Math.floor((125 * state.count + 1) / 2)),
    },
  }
}
featuresApi.checkIn = async (): Promise<ActivityResult> => {
  state.claims++
  await new Promise(resolve => setTimeout(resolve, 500))
  if (!state.count || state.unavailable) throw new ApiError(state.unavailable ? 503 : 403, {
    code: state.unavailable ? 'GROUP_BOOST_UNAVAILABLE' : 'GROUP_BOOST_REQUIRED', message: '', requestId: 'audit',
  })
  const reward = Math.floor((125 * state.count + 1) / 2)
  state.checkedIn = true
  state.balanceMinor += reward
  return {
    id: 'boost-check-in', kind: 'check_in', outcome: 'complete', message: '',
    reward: { kind: 'txb_delta', txbDeltaMinor: String(reward) },
    balanceAfter: { currency: 'TXB', minor: String(state.balanceMinor), display: '' },
    createdAt: new Date().toISOString(),
  }
}
Object.assign(window, { __boostAudit: state, __boostAuditLocale: setLocale })
const app = createApp(LuckyAudit)
const router = createRouter({ history: createWebHistory(), routes: [{ path: '/fixtures/activity-boost-audit.html', component: LuckyAudit }] })
app.config.globalProperties.$t = t
app.use(router)
app.use(ui)
void router.isReady().then(() => app.mount('#app'))
