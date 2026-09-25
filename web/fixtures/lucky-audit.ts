import '@/styles/main.css'
import { createApp } from 'vue'
import ui from '@nuxt/ui/vue-plugin'
import { featuresApi } from '@/api/features'
import { t } from '@/i18n'
import LuckyAudit from './LuckyAudit.vue'

const prizes = Array.from({ length: 20 }, (_, index) => ({ id: `prize-${index + 1}`, name: index === 19 ? 'Annual bonus · 500 TXB' : index % 3 === 0 ? `${index + 1} days added` : `${(index + 1) * 5} TXB` }))
const state = { delay: 1200, failNext: false, selectedId: 'prize-20', rewardKind: 'txb_delta' }
const overview = {
  balance: { currency: 'TXB', minor: '10000', display: '100.00 TXB' },
  timeZone: 'Asia/Shanghai', checkedInToday: true, dailyRewardMinTxbMinor: '50', dailyRewardMaxTxbMinor: '100',
  games: [], draws: [{ id: 'draw-1', name: 'Weekend rewards', description: 'One draw, one recorded result.', feeTxbMinor: '100', enabled: true, prizes }], recentResults: [],
  groupMessageReward: { enabled: false, localDate: '2026-09-25', messageCount: 0, threshold: 0, rewardMinor: '0', rewarded: false },
}
featuresApi.getActivity = async () => overview as never
featuresApi.drawLuckyPrize = async () => {
  await new Promise((resolve) => setTimeout(resolve, state.delay))
  if (state.failNext) { state.failNext = false; throw new Error('Constructed network failure') }
  const prize = prizes.find((candidate) => candidate.id === state.selectedId) ?? prizes[19]
  const reward = state.rewardKind === 'none' ? { kind: 'none' } : state.rewardKind === 'negative' ? { kind: 'txb_delta', txbDeltaMinor: '-50' } : { kind: 'txb_delta', txbDeltaMinor: '50000' }
  return { id: `result-${Date.now()}`, kind: 'draw', outcome: 'complete', message: '', drawId: 'draw-1', prizeId: prize.id, prizeName: prize.name, reward, balanceAfter: { currency: 'TXB', minor: '59900', display: '599.00 TXB' }, createdAt: new Date().toISOString() } as never
}
const auditState = Object.assign(state, { overview })
;(window as Window & { __audit?: typeof auditState }).__audit = auditState
const app = createApp(LuckyAudit)
app.config.globalProperties.$t = t
app.use(ui)
app.mount('#app')
