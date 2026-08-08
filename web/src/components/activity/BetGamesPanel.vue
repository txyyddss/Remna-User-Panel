<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import { PhDiceFive, PhTrendUp } from '@phosphor-icons/vue'

import type { BetGame } from '@/api/features'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'

const props = defineProps<{
  games: readonly BetGame[]
  busy: boolean
}>()
const emit = defineEmits<{ bet: [payload: { gameId: string; stakeTxbMinor: string }] }>()

const selectedId = shallowRef<string | null>(null)
const stake = shallowRef('')
const selected = computed(() => props.games.find((game) => game.id === selectedId.value))
const stakeMinor = computed(() => moneyFromTxbInput(stake.value))
const canBet = computed(() => {
  if (!selected.value || !stakeMinor.value) return false
  const amount = BigInt(stakeMinor.value)
  return amount >= BigInt(selected.value.minimumStakeMinor)
    && amount <= BigInt(selected.value.maximumStakeMinor)
})

watch(() => props.games, (games) => {
  if (!selectedId.value || !games.some((game) => game.id === selectedId.value)) {
    selectedId.value = games.find((game) => game.enabled)?.id ?? null
  }
}, { immediate: true })

watch(selected, (game) => {
  stake.value = game ? txbInputFromMinor(game.minimumStakeMinor) : ''
}, { immediate: true })

function submit(): void {
  if (!selected.value || !canBet.value) return
  emit('bet', { gameId: selected.value.id, stakeTxbMinor: stakeMinor.value })
}
</script>

<template>
  <section class="section-block bet-panel">
    <div class="section-heading section-heading--stacked">
      <h2>Bet games</h2>
      <p>Odds and total return are shown before you confirm. One result per tap.</p>
    </div>
    <div v-if="games.length" class="bet-grid">
      <button
        v-for="game in games"
        :key="game.id"
        class="bet-option"
        :class="{ 'bet-option--selected': game.id === selectedId }"
        type="button"
        :disabled="!game.enabled || busy"
        :aria-pressed="game.id === selectedId"
        @click="selectedId = game.id"
      >
        <span class="bet-option__icon"><PhDiceFive :size="21" /></span>
        <span><strong>{{ game.name }}</strong><small>{{ game.description || `${txbInputFromMinor(game.minimumStakeMinor)} to ${txbInputFromMinor(game.maximumStakeMinor)} TXB` }}</small></span>
        <span class="bet-option__odds">{{ (game.winChanceBps / 100).toFixed(2) }}%</span>
      </button>
    </div>
    <div v-if="selected" class="bet-form">
      <TxbAmountField
        id="bet-stake"
        v-model="stake"
        label="Stake"
        :min-minor="selected.minimumStakeMinor"
        :max-minor="selected.maximumStakeMinor"
        :hint="`${txbInputFromMinor(selected.minimumStakeMinor)} to ${txbInputFromMinor(selected.maximumStakeMinor)} TXB`"
        required
      />
      <div class="bet-disclosure">
        <PhTrendUp :size="19" />
        <span><strong>{{ (selected.returnMultiplierBps / 10000).toFixed(2) }}× total return</strong><small>A loss returns 0 TXB.</small></span>
      </div>
      <button class="button button--primary" type="button" :disabled="!canBet || busy" @click="submit">
        {{ busy ? 'Resolving securely' : 'Confirm bet' }}
      </button>
    </div>
    <div v-else class="empty-inline"><div><h3>No bet games</h3><p>An administrator can publish games when they are ready.</p></div></div>
  </section>
</template>

<style scoped>
.bet-grid { display: grid; gap: 0.55rem; }
.bet-option { min-height: 68px; display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: 0.7rem; padding: 0.7rem; border: 1px solid var(--line); border-radius: var(--radius-control); color: var(--text); background: var(--surface-raised); text-align: left; cursor: pointer; }
.bet-option--selected { border-color: #557763; background: var(--accent-soft); }
.bet-option__icon { width: 40px; height: 40px; display: grid; place-items: center; border-radius: 12px; color: var(--accent); background: #15241b; }
.bet-option strong, .bet-option small { display: block; }
.bet-option strong { font-size: 0.82rem; }
.bet-option small { margin-top: 0.2rem; color: var(--text-faint); font-size: 0.68rem; }
.bet-option__odds { color: var(--accent); font-family: var(--font-mono); font-size: 0.72rem; }
.bet-form { display: grid; gap: 0.8rem; margin-top: 1rem; padding-top: 1rem; border-top: 1px solid var(--line); }
.bet-disclosure { min-height: 52px; display: flex; align-items: center; gap: 0.65rem; padding: 0.65rem; border-radius: var(--radius-control); color: var(--accent); background: var(--accent-soft); }
.bet-disclosure strong, .bet-disclosure small { display: block; }
.bet-disclosure strong { font-size: 0.76rem; }
.bet-disclosure small { margin-top: 0.2rem; color: var(--text-muted); font-size: 0.65rem; }

@media (min-width: 720px) {
  .bet-form { grid-template-columns: minmax(220px, 1fr) minmax(180px, 0.8fr) auto; align-items: end; }
}
</style>
