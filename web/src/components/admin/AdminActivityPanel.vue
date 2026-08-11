<script setup lang="ts">
import { computed, onMounted, reactive, shallowRef } from 'vue'

import type { AdminStatistics, BetGame, CouponDefinition, LuckyDrawAdmin, LuckyDrawWrite, StatisticsQuery } from '@/api/features'
import { featuresApi } from '@/api/features'
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
const editingGameId = shallowRef<string | null | undefined>(undefined)
const editingDraw = shallowRef<LuckyDrawAdmin | null | undefined>(undefined)
const loading = shallowRef(true)
const busy = shallowRef(false)
const error = shallowRef<string | null>(null)
const gameDraft = reactive({ name: '', icon: 'dice', description: '', winChancePercent: 50, minStake: '1.00', maxStake: '20.00', returnMultiplier: 2, enabled: true })
const statisticsTarget = shallowRef<{ kind: 'game' | 'draw'; id: string; title: string } | null>(null)
const deleting = shallowRef<{ kind: 'game' | 'draw'; id: string; name: string } | null>(null)
const { t } = useI18n()
const iconItems = computed(() => ['dice', 'coin', 'cards', 'target', 'trophy', 'lightning', 'sparkle'].map((value) => ({ value, label: t(`adminActivityManagement.icons.${value}`) })))

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
    const [gameResponse, drawResponse, couponResponse] = await Promise.all([
      featuresApi.getAdminActivityGames(),
      featuresApi.getAdminLuckyDraws(),
      featuresApi.getAdminCoupons(),
    ])
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

    <section class="activity-admin-section">
      <div class="activity-admin-section__heading"><div><h3>{{ t('adminActivityManagement.games') }}</h3><p>{{ t('adminActivityManagement.gamesHint') }}</p></div><UButton icon="i-ph-plus" :label="t('adminActivityManagement.newGame')" @click="editGame()" /></div>
      <form v-if="editingGameId !== undefined" class="catalog-editor" @submit.prevent="saveGame">
        <div class="catalog-editor__heading"><h3>{{ editingGameId ? t('adminActivityManagement.editGame') : t('adminActivityManagement.newGame') }}</h3></div>
        <UFormField name="game-name" :label="t('adminActivityManagement.name')" required><UInput v-model.trim="gameDraft.name" class="w-full" :maxlength="80" /></UFormField>
        <UFormField name="game-icon" :label="t('adminActivityManagement.icon')"><USelect v-model="gameDraft.icon" class="w-full" :items="iconItems" /></UFormField>
        <UFormField class="catalog-editor__wide" name="game-description" :label="t('adminActivityManagement.description')"><UTextarea v-model.trim="gameDraft.description" class="w-full" :rows="2" :maxlength="300" /></UFormField>
        <UFormField name="win-chance" :label="t('adminActivityManagement.winChance')" required><UInput v-model.number="gameDraft.winChancePercent" class="w-full" type="number" :min="0.01" :max="99.99" :step="0.01" /></UFormField>
        <UFormField name="return-multiplier" :label="t('adminActivityManagement.returnMultiplier')" required><UInput v-model.number="gameDraft.returnMultiplier" class="w-full" type="number" :min="1.01" :step="0.01" /></UFormField>
        <TxbAmountField id="game-min-stake" v-model="gameDraft.minStake" :label="t('adminActivityManagement.minimumStake')" min-minor="1" required />
        <TxbAmountField id="game-max-stake" v-model="gameDraft.maxStake" :label="t('adminActivityManagement.maximumStake')" min-minor="1" required />
        <SwitchField id="game-enabled" v-model="gameDraft.enabled" :label="t('adminActivityManagement.available')" :help="t('adminActivityManagement.availableHint')" />
        <div class="button-row"><UButton color="neutral" variant="outline" :label="t('common.cancel')" @click="editingGameId = undefined" /><UButton type="submit" :loading="busy" :disabled="busy" :label="busy ? t('common.saving') : t('adminActivityManagement.saveGame')" /></div>
      </form>
      <USkeleton v-if="loading" class="h-24 w-full" />
      <div v-else v-auto-animate class="admin-list">
        <article v-for="game in games" :key="game.id" class="admin-list-row"><div><strong>{{ game.name }}</strong><small>{{ t('adminActivityManagement.gameSummary', { chance: (game.winChanceBps / 100).toFixed(2), multiplier: (game.returnMultiplierBps / 10000).toFixed(2), minimum: txbInputFromMinor(game.minimumStakeMinor), maximum: txbInputFromMinor(game.maximumStakeMinor) }) }}</small></div><div class="row-actions"><UButton color="neutral" variant="ghost" square icon="i-ph-chart-bar" :aria-label="t('adminActivityManagement.statisticsFor', { name: game.name })" @click="statisticsTarget = { kind: 'game', id: game.id, title: t('adminActivityManagement.statisticsTitle', { name: game.name }) }" /><UButton color="neutral" variant="ghost" square icon="i-ph-pencil-simple" :aria-label="t('adminActivityManagement.editNamed', { name: game.name })" @click="editGame(game)" /><UButton color="error" variant="ghost" square icon="i-ph-trash" :aria-label="t('adminActivityManagement.deleteNamed', { name: game.name })" @click="deleting = { kind: 'game', id: game.id, name: game.name }" /></div></article>
        <div v-if="!games.length" class="empty-inline"><div><h3>{{ t('adminActivityManagement.noGames') }}</h3><p>{{ t('adminActivityManagement.noGamesHint') }}</p></div></div>
      </div>
      <AdminStatisticsPanel v-if="statisticsTarget?.kind === 'game'" :title="statisticsTarget.title" :load="loadStatistics" @close="statisticsTarget = null" />
    </section>

    <section class="activity-admin-section">
      <div class="activity-admin-section__heading"><div><h3>{{ t('adminActivityManagement.draws') }}</h3><p>{{ t('adminActivityManagement.drawsHint') }}</p></div><UButton icon="i-ph-plus" :label="t('adminActivityManagement.newDraw')" @click="editingDraw = null" /></div>
      <AdminLuckyDrawEditor v-if="editingDraw !== undefined" :draw="editingDraw" :coupons="coupons" :busy="busy" @save="saveDraw" @cancel="editingDraw = undefined" />
      <div v-auto-animate class="admin-list">
        <article v-for="draw in draws" :key="draw.id" class="admin-list-row"><div class="admin-list-row__icon"><UIcon name="i-ph-gift" /></div><div><strong>{{ draw.name }}</strong><small>{{ t('adminActivityManagement.drawSummary', { fee: txbInputFromMinor(draw.feeTxbMinor), prizes: draw.prizes.length, status: draw.enabled ? t('adminActivityManagement.availableStatus') : t('adminActivityManagement.disabledStatus') }) }}</small></div><div class="row-actions"><UButton color="neutral" variant="ghost" square icon="i-ph-chart-bar" :aria-label="t('adminActivityManagement.statisticsFor', { name: draw.name })" @click="statisticsTarget = { kind: 'draw', id: draw.id, title: t('adminActivityManagement.statisticsTitle', { name: draw.name }) }" /><UButton color="neutral" variant="ghost" square icon="i-ph-pencil-simple" :aria-label="t('adminActivityManagement.editNamed', { name: draw.name })" @click="editingDraw = draw" /><UButton color="error" variant="ghost" square icon="i-ph-trash" :aria-label="t('adminActivityManagement.deleteNamed', { name: draw.name })" @click="deleting = { kind: 'draw', id: draw.id, name: draw.name }" /></div></article>
        <div v-if="!loading && !draws.length" class="empty-inline"><div><h3>{{ t('adminActivityManagement.noDraws') }}</h3><p>{{ t('adminActivityManagement.noDrawsHint') }}</p></div></div>
      </div>
      <AdminStatisticsPanel v-if="statisticsTarget?.kind === 'draw'" :title="statisticsTarget.title" :load="loadStatistics" @close="statisticsTarget = null" />
    </section>

    <UAlert v-if="error" class="admin-error" color="warning" variant="soft" icon="i-ph-warning" :description="error" />
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
