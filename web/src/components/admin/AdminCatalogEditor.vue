<script setup lang="ts">
import { reactive, watch } from 'vue'
import { PhFloppyDisk, PhX } from '@phosphor-icons/vue'

import type { Combo, ResetCadence } from '@/api/types'

const props = defineProps<{ combo?: Combo; busy: boolean }>()
const emit = defineEmits<{
  cancel: []
  save: [payload: Record<string, unknown>]
}>()

const draft = reactive({
  name: '',
  description: '',
  priceTxbMinor: '',
  validityDays: 30,
  trafficLimitBytes: '',
  resetStrategy: 'MONTH' as ResetCadence,
  squadProductIds: [] as string[],
  active: true,
})

watch(() => props.combo, (combo) => {
  Object.assign(draft, combo ? {
    name: combo.name,
    description: combo.description,
    priceTxbMinor: combo.price.minor,
    validityDays: combo.validityDays,
    trafficLimitBytes: combo.trafficLimitBytes,
    resetStrategy: combo.resetStrategy,
    squadProductIds: combo.includedSquads.map((squad) => squad.id),
    active: combo.active,
  } : {
    name: '', description: '', priceTxbMinor: '', validityDays: 30,
    trafficLimitBytes: '', resetStrategy: 'MONTH', squadProductIds: [], active: true,
  })
}, { immediate: true })

function submit(): void {
  emit('save', { ...draft })
}
</script>

<template>
  <form class="catalog-editor" @submit.prevent="submit">
    <div class="catalog-editor__heading">
      <h3>{{ combo ? `Edit ${combo.name}` : 'New combo' }}</h3>
      <button class="icon-button" type="button" aria-label="Close editor" @click="$emit('cancel')"><PhX :size="19" /></button>
    </div>
    <label><span>Name</span><input v-model.trim="draft.name" required maxlength="80" /></label>
    <label class="catalog-editor__wide"><span>Description</span><textarea v-model.trim="draft.description" required rows="3" /></label>
    <label><span>Price, TXB minor units</span><input v-model="draft.priceTxbMinor" required inputmode="numeric" pattern="[0-9]+" /></label>
    <label><span>Validity days</span><input v-model.number="draft.validityDays" required type="number" min="1" /></label>
    <label><span>Traffic limit, bytes</span><input v-model="draft.trafficLimitBytes" required inputmode="numeric" pattern="[0-9]+" /></label>
    <label><span>Reset cadence</span><select v-model="draft.resetStrategy"><option>DAY</option><option>WEEK</option><option>MONTH</option></select></label>
    <label class="switch-row"><input v-model="draft.active" type="checkbox" /><span>Available to users</span></label>
    <button class="button button--primary catalog-editor__wide" type="submit" :disabled="busy">
      <PhFloppyDisk :size="18" /> {{ busy ? 'Saving' : 'Save combo' }}
    </button>
  </form>
</template>
