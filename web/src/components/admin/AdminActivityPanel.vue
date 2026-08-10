<script setup lang="ts">
import { onMounted, reactive, shallowRef } from 'vue'
import { PhChartBar, PhGift, PhPencilSimple, PhPlus, PhTrash } from '@phosphor-icons/vue'

import type { ActivitySettings, AdminStatistics, BetGame, CouponDefinition, LuckyDrawAdmin, LuckyDrawWrite, StatisticsQuery } from '@/api/features'
import { featuresApi } from '@/api/features'
import AdminActivitySettings from '@/components/admin/activity/AdminActivitySettings.vue'
import AdminLuckyDrawEditor from '@/components/admin/activity/AdminLuckyDrawEditor.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { localizedError, useI18n } from '@/i18n'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'
import AdminStatisticsPanel from './AdminStatisticsPanel.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'

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
const statisticsTarget = shallowRef<{ kind: 'game' | 'draw'; id: string; title: string } | null>(null)
const deleting = shallowRef<{ kind: 'game' | 'draw'; id: string; name: string } | null>(null)
const { t } = useI18n()

function loadStatistics(query: StatisticsQuery): Promise<AdminStatistics> {
  if (!statisticsTarget.value) return Promise.reject(new Error(t('adminActivityManagement.chooseActivity')))
  return statisticsTarget.value.kind === 'game'
    ? featuresApi.getAdminActivityGameStatistics(statisticsTarget.value.id, query)
    : featuresApi.getAdminLuckyDrawStatistics(statisticsTarget.value.id, query)
}

async function removeActivity(): Promise<void> {
  if (!deleting.value || busy.value) return
  busy.value = true
  error.value = null
  try {
    if (deleting.value.kind === 'game') await featuresApi.deleteAdminActivityGame(deleting.value.id)
    else await featuresApi.deleteAdminLuckyDraw(deleting.value.id)
    deleting.value = null
    await load()
  } catch (caught) {
    error.value = localizedError(caught, 'adminActivityManagement.deleteFailed')
  } finally {
    busy.value = false
  }
}

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
    error.value = localizedError(caught, 'adminActivityManagement.loadFailed')
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

async function saveSettings(value: { timezone: string; groupMessageThreshold: number }): Promise<void> {
  busy.value = true
  error.value = null
  try {
    const saved = await featuresApi.saveAdminActivitySettings(value)
    if (settings.value) settings.value = { ...settings.value, ...saved }
  } catch (caught) {
    error.value = localizedError(caught, 'adminActivityManagement.settingsFailed')
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
    error.value = localizedError(caught, 'adminActivityManagement.gameSaveFailed')
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
    error.value = localizedError(caught, 'adminActivityManagement.drawSaveFailed')
  } finally {
    busy.value = false
  }
}

onMounted(() => void load())
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>{{ t('adminActivityManagement.title') }}</h2><p>{{ t('adminActivityManagement.copy') }}</p></div>
    </div>

    <AdminActivitySettings :settings="settings" :busy="busy" @save="saveSettings" />

    <section class="activity-admin-section">
      <div class="activity-admin-section__heading"><div><h3>{{ t('adminActivityManagement.games') }}</h3><p>{{ t('adminActivityManagement.gamesHint') }}</p></div><button class="button button--primary" type="button" @click="editGame()"><PhPlus :size="18" />{{ t('adminActivityManagement.newGame') }}</button></div>
      <form v-if="editingGameId !== undefined" class="catalog-editor" @submit.prevent="saveGame">
        <div class="catalog-editor__heading"><h3>{{ editingGameId ? t('adminActivityManagement.editGame') : t('adminActivityManagement.newGame') }}</h3></div>
        <label><span>{{ t('adminActivityManagement.name') }}</span><input v-model.trim="gameDraft.name" required maxlength="80" /></label>
        <label><span>{{ t('adminActivityManagement.icon') }}</span><select v-model="gameDraft.icon"><option value="dice">{{ t('adminActivityManagement.icons.dice') }}</option><option value="coin">{{ t('adminActivityManagement.icons.coin') }}</option><option value="cards">{{ t('adminActivityManagement.icons.cards') }}</option><option value="target">{{ t('adminActivityManagement.icons.target') }}</option><option value="trophy">{{ t('adminActivityManagement.icons.trophy') }}</option><option value="lightning">{{ t('adminActivityManagement.icons.lightning') }}</option><option value="sparkle">{{ t('adminActivityManagement.icons.sparkle') }}</option></select></label>
        <label class="catalog-editor__wide"><span>{{ t('adminActivityManagement.description') }}</span><textarea v-model.trim="gameDraft.description" rows="2" maxlength="300" /></label>
        <label><span>{{ t('adminActivityManagement.winChance') }}</span><input v-model.number="gameDraft.winChancePercent" type="number" min="0.01" max="99.99" step="0.01" required /></label>
        <label><span>{{ t('adminActivityManagement.returnMultiplier') }}</span><input v-model.number="gameDraft.returnMultiplier" type="number" min="1.01" step="0.01" required /></label>
        <TxbAmountField id="game-min-stake" v-model="gameDraft.minStake" :label="t('adminActivityManagement.minimumStake')" min-minor="1" required />
        <TxbAmountField id="game-max-stake" v-model="gameDraft.maxStake" :label="t('adminActivityManagement.maximumStake')" min-minor="1" required />
        <SwitchField id="game-enabled" v-model="gameDraft.enabled" :label="t('adminActivityManagement.available')" :help="t('adminActivityManagement.availableHint')" />
        <div class="button-row"><button class="button button--secondary" type="button" @click="editingGameId = undefined">{{ t('common.cancel') }}</button><button class="button button--primary" type="submit" :disabled="busy">{{ busy ? t('common.saving') : t('adminActivityManagement.saveGame') }}</button></div>
      </form>
      <div v-if="loading" class="admin-loading">{{ t('adminActivityManagement.loadingGames') }}</div>
      <div v-else class="admin-list">
        <article v-for="game in games" :key="game.id" class="admin-list-row"><div><strong>{{ game.name }}</strong><small>{{ t('adminActivityManagement.gameSummary', { chance: (game.winChanceBps / 100).toFixed(2), multiplier: (game.returnMultiplierBps / 10000).toFixed(2), minimum: txbInputFromMinor(game.minimumStakeMinor), maximum: txbInputFromMinor(game.maximumStakeMinor) }) }}</small></div><div class="row-actions"><button class="icon-button" type="button" :aria-label="t('adminActivityManagement.statisticsFor', { name: game.name })" @click="statisticsTarget = { kind: 'game', id: game.id, title: t('adminActivityManagement.statisticsTitle', { name: game.name }) }"><PhChartBar :size="18" /></button><button class="icon-button" type="button" :aria-label="t('adminActivityManagement.editNamed', { name: game.name })" @click="editGame(game)"><PhPencilSimple :size="18" /></button><button class="icon-button icon-button--danger" type="button" :aria-label="t('adminActivityManagement.deleteNamed', { name: game.name })" @click="deleting = { kind: 'game', id: game.id, name: game.name }"><PhTrash :size="18" /></button></div></article>
        <div v-if="!games.length" class="empty-inline"><div><h3>{{ t('adminActivityManagement.noGames') }}</h3><p>{{ t('adminActivityManagement.noGamesHint') }}</p></div></div>
      </div>
      <AdminStatisticsPanel v-if="statisticsTarget?.kind === 'game'" :title="statisticsTarget.title" :load="loadStatistics" @close="statisticsTarget = null" />
    </section>

    <section class="activity-admin-section">
      <div class="activity-admin-section__heading"><div><h3>{{ t('adminActivityManagement.draws') }}</h3><p>{{ t('adminActivityManagement.drawsHint') }}</p></div><button class="button button--primary" type="button" @click="editingDraw = null"><PhPlus :size="18" />{{ t('adminActivityManagement.newDraw') }}</button></div>
      <AdminLuckyDrawEditor v-if="editingDraw !== undefined" :draw="editingDraw" :coupons="coupons" :busy="busy" @save="saveDraw" @cancel="editingDraw = undefined" />
      <div class="admin-list">
        <article v-for="draw in draws" :key="draw.id" class="admin-list-row"><div class="admin-list-row__icon"><PhGift :size="19" /></div><div><strong>{{ draw.name }}</strong><small>{{ t('adminActivityManagement.drawSummary', { fee: txbInputFromMinor(draw.feeTxbMinor), prizes: draw.prizes.length, status: draw.enabled ? t('adminActivityManagement.availableStatus') : t('adminActivityManagement.disabledStatus') }) }}</small></div><div class="row-actions"><button class="icon-button" type="button" :aria-label="t('adminActivityManagement.statisticsFor', { name: draw.name })" @click="statisticsTarget = { kind: 'draw', id: draw.id, title: t('adminActivityManagement.statisticsTitle', { name: draw.name }) }"><PhChartBar :size="18" /></button><button class="icon-button" type="button" :aria-label="t('adminActivityManagement.editNamed', { name: draw.name })" @click="editingDraw = draw"><PhPencilSimple :size="18" /></button><button class="icon-button icon-button--danger" type="button" :aria-label="t('adminActivityManagement.deleteNamed', { name: draw.name })" @click="deleting = { kind: 'draw', id: draw.id, name: draw.name }"><PhTrash :size="18" /></button></div></article>
        <div v-if="!loading && !draws.length" class="empty-inline"><div><h3>{{ t('adminActivityManagement.noDraws') }}</h3><p>{{ t('adminActivityManagement.noDrawsHint') }}</p></div></div>
      </div>
      <AdminStatisticsPanel v-if="statisticsTarget?.kind === 'draw'" :title="statisticsTarget.title" :load="loadStatistics" @close="statisticsTarget = null" />
    </section>

    <p v-if="error" class="field-error admin-error" role="alert">{{ error }}</p>
    <ConfirmDialog :open="Boolean(deleting)" :title="t('adminActivityManagement.deleteTitle', { name: deleting?.name ?? t('adminActivityManagement.activity') })" :description="t('adminActivityManagement.deleteDescription')" :confirm-label="t('adminActivityManagement.deletePermanently')" :busy="busy" danger @update:open="!$event && (deleting = null)" @confirm="removeActivity" />
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
