<script setup lang="ts">
import { shallowRef } from 'vue'
import { PhArrowClockwise, PhPencilSimple, PhPlus, PhTrash } from '@phosphor-icons/vue'

import type { Combo, SquadProduct, SquadProductWrite } from '@/api/types'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useAdminSection } from '@/composables/useAdminSection'
import { formatBytes, formatMoney } from '@/utils/format'
import AdminCatalogEditor from './AdminCatalogEditor.vue'
import AdminSectionState from './AdminSectionState.vue'
import AdminSquadEditor from './AdminSquadEditor.vue'

const combos = useAdminSection<Combo>('combos')
const squads = useAdminSection<SquadProduct>('squad-products')
const editing = shallowRef<Combo | null | undefined>(undefined)
const deleting = shallowRef<Combo | null>(null)
const editingSquad = shallowRef<SquadProduct | null>(null)

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
      <div><h2>Catalog</h2><p>Manage combo terms and imported Remnawave squads.</p></div>
      <button class="button button--primary" type="button" @click="editing = null"><PhPlus :size="18" /> New combo</button>
    </div>
    <AdminCatalogEditor v-if="editing !== undefined" :combo="editing ?? undefined" :squads="squads.items.value" :busy="combos.busy.value" @save="save" @cancel="editing = undefined" />
    <AdminSectionState :loading="combos.loading.value" :error="combos.error.value" @retry="combos.load()">
      <div class="admin-list">
        <article v-for="combo in combos.items.value" :key="combo.id" class="admin-list-row admin-list-row--catalog">
          <div><strong>{{ combo.name }}</strong><small>{{ formatBytes(combo.trafficLimitBytes) }} for {{ combo.validityDays }} days</small></div>
          <strong>{{ formatMoney(combo.price) }}</strong>
          <StatusBadge :tone="combo.active ? 'success' : 'neutral'" :label="combo.active ? 'Available' : 'Paused'" />
          <div class="row-actions">
            <button class="icon-button" type="button" :aria-label="`Edit ${combo.name}`" @click="editing = combo"><PhPencilSimple :size="18" /></button>
            <button class="icon-button icon-button--danger" type="button" :aria-label="`Delete ${combo.name}`" @click="deleting = combo"><PhTrash :size="18" /></button>
          </div>
        </article>
        <div v-if="!combos.items.value.length" class="empty-inline"><div><h3>No combos</h3><p>Create the first plan to open the catalog.</p></div></div>
      </div>
    </AdminSectionState>

    <div class="admin-subsection-heading">
      <div><h3>Internal squads</h3><p>Descriptions and pricing remain local after import.</p></div>
      <button class="button button--secondary" type="button" :disabled="squads.busy.value" @click="importSquads"><PhArrowClockwise :size="18" /> Import</button>
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
          <StatusBadge :tone="squad.upstreamPresent && squad.visible ? 'success' : 'neutral'" :label="squad.upstreamPresent && squad.visible ? 'Available' : 'Hidden'" />
          <div class="row-actions">
            <button class="icon-button" type="button" :aria-label="`Edit ${squad.name}`" @click="editingSquad = squad">
              <PhPencilSimple :size="18" />
            </button>
          </div>
        </article>
        <div v-if="!squads.items.value.length" class="empty-inline"><div><h3>No squads imported</h3><p>Import from Remnawave to configure optional access.</p></div></div>
      </div>
    </AdminSectionState>

    <ConfirmDialog
      :open="deleting !== null"
      title="Delete this combo?"
      description="Existing purchases keep their saved price and terms. The combo disappears from new purchases."
      confirm-label="Delete combo"
      :busy="combos.busy.value"
      danger
      @update:open="!$event && (deleting = null)"
      @confirm="remove"
    />
  </section>
</template>
