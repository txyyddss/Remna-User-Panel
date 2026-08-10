<script setup lang="ts">
import { shallowRef } from 'vue'
import { PhArrowClockwise, PhChartBar, PhPencilSimple, PhPlus, PhTrash } from '@phosphor-icons/vue'

import { featuresApi, type AdminStatistics, type StatisticsQuery } from '@/api/features'
import type { Combo, SquadProduct, SquadProductWrite } from '@/api/types'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { useI18n } from '@/i18n'
import { formatBytes, formatMoney } from '@/utils/format'
import AdminCatalogEditor from './AdminCatalogEditor.vue'
import AdminSectionState from './AdminSectionState.vue'
import AdminSquadEditor from './AdminSquadEditor.vue'
import AdminStatisticsPanel from './AdminStatisticsPanel.vue'

const combos = useAdminSection<Combo>('combos')
const squads = useAdminSection<SquadProduct>('squad-products')
const editing = shallowRef<Combo | null | undefined>(undefined)
const deleting = shallowRef<Combo | null>(null)
const editingSquad = shallowRef<SquadProduct | null>(null)
const statisticsTarget = shallowRef<{ kind: 'combo' | 'squad'; id: string; title: string } | null>(null)
const { t } = useI18n()

function loadStatistics(query: StatisticsQuery): Promise<AdminStatistics> {
  const target = statisticsTarget.value
  if (!target) return Promise.reject(new Error(t('adminCatalog.chooseRecord')))
  return target.kind === 'combo'
    ? featuresApi.getAdminComboStatistics(target.id, query)
    : featuresApi.getAdminSquadStatistics(target.id, query)
}

async function save(payload: Record<string, unknown>): Promise<void> {
  const success = editing.value?.id
    ? await combos.update(editing.value.id, payload)
    : await combos.create(payload)
  if (success) editing.value = undefined
}

async function remove(): Promise<void> {
  if (!deleting.value) return
  const success = await combos.remove(deleting.value.id)
  if (success) deleting.value = null
}

function importSquads(): void {
  void squads.perform(() => import('@/api/client').then(({ api }) => api.importAdminSquadProducts()))
}

async function saveSquad(payload: SquadProductWrite): Promise<void> {
  if (!editingSquad.value) return
  const id = editingSquad.value.id
  const success = await squads.perform(() => import('@/api/client').then(({ api }) => api.updateAdminSquadProduct(id, payload)))
  if (success) editingSquad.value = null
}
</script>

<template>
  <section class="admin-panel">
    <div class="admin-panel__heading">
      <div><h2>{{ t('adminCatalog.title') }}</h2><p>{{ t('adminCatalog.copy') }}</p></div>
      <button class="button button--primary" type="button" @click="editing = null"><PhPlus :size="18" /> {{ t('adminCatalog.newCombo') }}</button>
    </div>
    <AdminCatalogEditor v-if="editing !== undefined" :combo="editing ?? undefined" :squads="squads.items.value" :busy="combos.busy.value" @save="save" @cancel="editing = undefined" />
    <AdminSectionState :loading="combos.loading.value" :error="combos.error.value" @retry="combos.load()">
      <div class="admin-list">
        <article v-for="combo in combos.items.value" :key="combo.id" class="admin-list-row admin-list-row--catalog">
          <div><strong>{{ combo.name }}</strong><small>{{ t('adminCatalog.comboSummary', { traffic: formatBytes(combo.trafficLimitBytes), days: combo.validityDays }) }}</small></div>
          <strong>{{ formatMoney(combo.price) }}</strong>
          <StatusBadge :tone="combo.active ? 'success' : 'neutral'" :label="combo.active ? t('adminCatalog.available') : t('adminCatalog.paused')" />
          <div class="row-actions">
            <button class="icon-button" type="button" :aria-label="t('adminCatalog.statisticsFor', { name: combo.name })" @click="statisticsTarget = { kind: 'combo', id: combo.id, title: t('adminCatalog.statisticsTitle', { name: combo.name }) }"><PhChartBar :size="18" /></button>
            <button class="icon-button" type="button" :aria-label="t('adminCatalog.editNamed', { name: combo.name })" @click="editing = combo"><PhPencilSimple :size="18" /></button>
            <button class="icon-button icon-button--danger" type="button" :aria-label="t('adminCatalog.hideNamed', { name: combo.name })" @click="deleting = combo"><PhTrash :size="18" /></button>
          </div>
        </article>
        <div v-if="!combos.items.value.length" class="empty-inline"><div><h3>{{ t('adminCatalog.noCombos') }}</h3><p>{{ t('adminCatalog.noCombosHint') }}</p></div></div>
      </div>
    </AdminSectionState>
    <AdminStatisticsPanel v-if="statisticsTarget?.kind === 'combo'" :title="statisticsTarget.title" :load="loadStatistics" @close="statisticsTarget = null" />

    <div class="admin-subsection-heading">
      <div><h3>{{ t('adminCatalog.squads') }}</h3><p>{{ t('adminCatalog.squadsHint') }}</p></div>
      <button class="button button--secondary" type="button" :disabled="squads.busy.value" @click="importSquads"><PhArrowClockwise :size="18" /> {{ t('adminCatalog.refreshSquads') }}</button>
    </div>
    <AdminSquadEditor
      v-if="editingSquad"
      :squad="editingSquad"
      :busy="squads.busy.value"
      @save="saveSquad"
      @cancel="editingSquad = null"
    />
    <AdminSectionState :loading="squads.loading.value" :error="squads.error.value" @retry="squads.load()">
      <div class="admin-list admin-list--compact">
        <article v-for="squad in squads.items.value" :key="squad.id" class="admin-list-row">
          <div><strong>{{ squad.name }}</strong><small>{{ squad.description || squad.remnaSquadUuid }}</small></div>
          <strong>{{ formatMoney(squad.price) }}</strong>
          <StatusBadge :tone="squad.upstreamPresent && squad.visible ? 'success' : 'neutral'" :label="squad.upstreamPresent && squad.visible ? t('adminCatalog.available') : t('adminCatalog.hidden')" />
          <div class="row-actions">
            <button class="icon-button" type="button" :aria-label="t('adminCatalog.statisticsFor', { name: squad.name })" @click="statisticsTarget = { kind: 'squad', id: squad.id, title: t('adminCatalog.statisticsTitle', { name: squad.name }) }"><PhChartBar :size="18" /></button>
            <button class="icon-button" type="button" :aria-label="t('adminCatalog.editNamed', { name: squad.name })" @click="editingSquad = squad">
              <PhPencilSimple :size="18" />
            </button>
          </div>
        </article>
        <div v-if="!squads.items.value.length" class="empty-inline"><div><h3>{{ t('adminCatalog.noSquads') }}</h3><p>{{ t('adminCatalog.noSquadsHint') }}</p></div></div>
      </div>
    </AdminSectionState>
    <AdminStatisticsPanel v-if="statisticsTarget?.kind === 'squad'" :title="statisticsTarget.title" :load="loadStatistics" @close="statisticsTarget = null" />

    <ConfirmDialog
      :open="deleting !== null"
      :title="t('adminCatalog.hideTitle')"
      :description="t('adminCatalog.hideDescription')"
      :confirm-label="t('adminCatalog.hideCombo')"
      :busy="combos.busy.value"
      danger
      @update:open="!$event && (deleting = null)"
      @confirm="remove"
    />
  </section>
</template>
