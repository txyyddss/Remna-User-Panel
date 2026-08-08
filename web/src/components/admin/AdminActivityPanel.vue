<script setup lang="ts">
import { onMounted, reactive, shallowRef } from 'vue'
import { PhGift, PhPencilSimple, PhPlus } from '@phosphor-icons/vue'

import type { ActivitySettings, BetGame, CouponDefinition, LuckyDrawAdmin, LuckyDrawWrite } from '@/api/features'
import { featuresApi } from '@/api/features'
import AdminActivitySettings from '@/components/admin/activity/AdminActivitySettings.vue'
import AdminLuckyDrawEditor from '@/components/admin/activity/AdminLuckyDrawEditor.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'

const games = shallowRef<BetGame[]>([])
const draws = shallowRef<LuckyDrawAdmin[]>([])
const coupons = shallowRef<CouponDefinition[]>([])
const settings = shallowRef<ActivitySettings | null>(null)
const editingGameId = shallowRef<string | null | undefined>(undefined)
const editingDraw = shallowRef<LuckyDrawAdmin | null | undefined>(undefined)
const loading = shallowRef(true)
const busy = shallowRef(false)
const error = shallowRef<string | null>(null)
const gameDraft = reactive({ name: '', icon: 'dice', description: '', winChancePercent: 50, minStake: '1.00', maxStake: '20.00', returnMultiplier: 2, enabled: true })

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  try {
    const [settingsResponse, gameResponse, drawResponse, couponResponse] = await Promise.all([
      featuresApi.getAdminActivitySettings(),
      featuresApi.getAdminActivityGames(),
      featuresApi.getAdminLuckyDraws(),
      featuresApi.getAdminCoupons(),
    ])
    settings.value = settingsResponse
    games.value = gameResponse.items
    draws.value = drawResponse.items
    coupons.value = couponResponse.items
  } catch (caught) {
    error.value = caught instanceof Error ? caught.message : 'Activity configuration could not be loaded.'
  } finally {
    loading.value = false
  }
}

function editGame(game?: BetGame): void {
  editingGameId.value = game?.id ?? null
  Object.assign(gameDraft, game ? {
    name: game.name,
    icon: game.icon,
    description: game.description,
    winChancePercent: game.winChanceBps / 100,
    minStake: txbInputFromMinor(game.minimumStakeMinor),
    maxStake: txbInputFromMinor(game.maximumStakeMinor),
    returnMultiplier: game.returnMultiplierBps / 10000,
    enabled: game.enabled,
  } : { name: '', icon: 'dice', description: '', winChancePercent: 50, minStake: '1.00', maxStake: '20.00', returnMultiplier: 2, enabled: true })
}

async function saveSettings(value: { timezone: string; dailyRewardTxb: string; groupMessageThreshold: number; groupMessageRewardTxb: string }): Promise<void> {
  busy.value = true
  error.value = null
  try {
    settings.value = await featuresApi.saveAdminActivitySettings(value)
  } catch (caught) {
    error.value = caught instanceof Error ? caught.message : 'Check-in settings could not be saved.'
  } finally {
    busy.value = false
  }
}

async function saveGame(): Promise<void> {
  const minimumStakeMinor = moneyFromTxbInput(gameDraft.minStake)
  const maximumStakeMinor = moneyFromTxbInput(gameDraft.maxStake)
  if (!minimumStakeMinor || !maximumStakeMinor) return
  busy.value = true
  error.value = null
  try {
    await featuresApi.saveAdminActivityGame(editingGameId.value ?? null, {
      name: gameDraft.name,
      icon: gameDraft.icon,
      description: gameDraft.description,
      winChanceBps: Math.round(gameDraft.winChancePercent * 100),
      minimumStakeMinor,
      maximumStakeMinor,
      returnMultiplierBps: Math.round(gameDraft.returnMultiplier * 10000),
      enabled: gameDraft.enabled,
    })
    editingGameId.value = undefined
    games.value = (await featuresApi.getAdminActivityGames()).items
  } catch (caught) {
    error.value = caught instanceof Error ? caught.message : 'The game could not be saved.'
  } finally {
    busy.value = false
  }
}

async function saveDraw(value: LuckyDrawWrite): Promise<void> {
  busy.value = true
  error.value = null
  try {
    await featuresApi.saveAdminLuckyDraw(editingDraw.value?.id ?? null, value)
    editingDraw.value = undefined
    draws.value = (await featuresApi.getAdminLuckyDraws()).items
  } catch (caught) {
    error.value = caught instanceof Error ? caught.message : 'The lucky draw could not be saved.'
  } finally {
    busy.value = false
  }
}

onMounted(() => void load())
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>Activity</h2><p>Set check-in rewards, transparent game odds, and weighted draws.</p></div>
    </div>

    <AdminActivitySettings :settings="settings" :busy="busy" @save="saveSettings" />

    <section class="activity-admin-section">
      <div class="activity-admin-section__heading"><div><h3>Betting games</h3><p>Members see odds, limits, and total return before staking.</p></div><button class="button button--primary" type="button" @click="editGame()"><PhPlus :size="18" />New game</button></div>
      <form v-if="editingGameId !== undefined" class="catalog-editor" @submit.prevent="saveGame">
        <div class="catalog-editor__heading"><h3>{{ editingGameId ? 'Edit game' : 'New game' }}</h3></div>
        <label><span>Name</span><input v-model.trim="gameDraft.name" required maxlength="80" /></label>
        <label><span>Icon key</span><input v-model.trim="gameDraft.icon" required maxlength="30" /></label>
        <label class="catalog-editor__wide"><span>Description</span><textarea v-model.trim="gameDraft.description" rows="2" maxlength="300" /></label>
        <label><span>Win chance, percent</span><input v-model.number="gameDraft.winChancePercent" type="number" min="0.01" max="99.99" step="0.01" required /></label>
        <label><span>Total return multiplier</span><input v-model.number="gameDraft.returnMultiplier" type="number" min="1.01" step="0.01" required /></label>
        <TxbAmountField id="game-min-stake" v-model="gameDraft.minStake" label="Minimum stake" min-minor="1" required />
        <TxbAmountField id="game-max-stake" v-model="gameDraft.maxStake" label="Maximum stake" min-minor="1" required />
        <SwitchField id="game-enabled" v-model="gameDraft.enabled" label="Available to members" help="Disabled games remain in immutable result history." />
        <div class="button-row"><button class="button button--secondary" type="button" @click="editingGameId = undefined">Cancel</button><button class="button button--primary" type="submit" :disabled="busy">{{ busy ? 'Saving' : 'Save game' }}</button></div>
      </form>
      <div v-if="loading" class="admin-loading">Loading games</div>
      <div v-else class="admin-list">
        <article v-for="game in games" :key="game.id" class="admin-list-row"><div><strong>{{ game.name }}</strong><small>{{ (game.winChanceBps / 100).toFixed(2) }}% win · {{ (game.returnMultiplierBps / 10000).toFixed(2) }}× return · {{ txbInputFromMinor(game.minimumStakeMinor) }} to {{ txbInputFromMinor(game.maximumStakeMinor) }} TXB</small></div><button class="icon-button" type="button" :aria-label="`Edit ${game.name}`" @click="editGame(game)"><PhPencilSimple :size="18" /></button></article>
        <div v-if="!games.length" class="empty-inline"><div><h3>No games</h3><p>Create a game with explicit odds.</p></div></div>
      </div>
    </section>

    <section class="activity-admin-section">
      <div class="activity-admin-section__heading"><div><h3>Lucky draws</h3><p>Configure a fee and a complete weighted prize table.</p></div><button class="button button--primary" type="button" @click="editingDraw = null"><PhPlus :size="18" />New draw</button></div>
      <AdminLuckyDrawEditor v-if="editingDraw !== undefined" :draw="editingDraw" :coupons="coupons" :busy="busy" @save="saveDraw" @cancel="editingDraw = undefined" />
      <div class="admin-list">
        <article v-for="draw in draws" :key="draw.id" class="admin-list-row"><div class="admin-list-row__icon"><PhGift :size="19" /></div><div><strong>{{ draw.name }}</strong><small>{{ txbInputFromMinor(draw.feeTxbMinor) }} TXB fee · {{ draw.prizes.length }} prizes · {{ draw.enabled ? 'Available' : 'Disabled' }}</small></div><button class="icon-button" type="button" :aria-label="`Edit ${draw.name}`" @click="editingDraw = draw"><PhPencilSimple :size="18" /></button></article>
        <div v-if="!loading && !draws.length" class="empty-inline"><div><h3>No lucky draws</h3><p>Add a weighted draw when prizes are ready.</p></div></div>
      </div>
    </section>

    <p v-if="error" class="field-error admin-error" role="alert">{{ error }}</p>
  </section>
</template>

<style scoped>
.activity-admin-section { display: grid; gap: 0.8rem; padding: 1rem; border-top: 1px solid var(--line); }
.activity-admin-section__heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 0.8rem; }
.activity-admin-section__heading h3, .activity-admin-section__heading p { margin: 0; }
.activity-admin-section__heading p { margin-top: 0.25rem; color: var(--text-muted); font-size: 0.78rem; }
.admin-list-row__icon { display: grid; width: 34px; height: 34px; place-items: center; border-radius: 10px; color: var(--accent); background: var(--accent-soft); }
.admin-error { margin: 1rem; }
</style>
