<script setup lang="ts">
import { reactive, watch } from 'vue'
import { PhFloppyDisk, PhX } from '@phosphor-icons/vue'

import type { SquadProduct, SquadProductWrite } from '@/api/types'

const props = defineProps<{
  squad: SquadProduct
  busy: boolean
}>()

const emit = defineEmits<{
  cancel: []
  save: [payload: SquadProductWrite]
}>()

const draft = reactive({
  description: '',
  priceTxbMinor: '',
  visible: false,
})

watch(() => props.squad, (squad) => {
  Object.assign(draft, {
    description: squad.description,
    priceTxbMinor: squad.price.minor,
    visible: squad.visible,
  })
}, { immediate: true })

function submit(): void {
  emit('save', {
    remnaSquadUuid: props.squad.remnaSquadUuid,
    name: props.squad.name,
    description: draft.description.trim(),
    priceTxbMinor: draft.priceTxbMinor,
    visible: draft.visible,
  })
}
</script>

<template>
  <form class="catalog-editor squad-editor" @submit.prevent="submit">
    <div class="catalog-editor__heading">
      <div>
        <h3>Edit {{ squad.name }}</h3>
        <p>Remnawave identity is read-only. Description, price, and visibility are local.</p>
      </div>
      <button class="icon-button" type="button" aria-label="Close squad editor" @click="$emit('cancel')">
        <PhX :size="19" />
      </button>
    </div>
    <label class="catalog-editor__wide">
      <span>Description</span>
      <textarea v-model="draft.description" rows="3" maxlength="1000" />
    </label>
    <label>
      <span>Price, TXB minor units</span>
      <input v-model="draft.priceTxbMinor" required inputmode="numeric" pattern="[0-9]+" />
    </label>
    <label class="switch-row">
      <input v-model="draft.visible" type="checkbox" />
      <span>Visible in the user catalog</span>
    </label>
    <button class="button button--primary catalog-editor__wide" type="submit" :disabled="busy">
      <PhFloppyDisk :size="18" /> {{ busy ? 'Saving' : 'Save squad' }}
    </button>
  </form>
</template>
