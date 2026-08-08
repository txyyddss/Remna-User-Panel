<script setup lang="ts">
import { reactive, watch } from 'vue'
import { PhFloppyDisk, PhX } from '@phosphor-icons/vue'

import type { SquadProduct, SquadProductWrite } from '@/api/types'
import MarkdownContent from '@/components/common/MarkdownContent.vue'
import SwitchField from '@/components/common/SwitchField.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'

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
  priceTxb: '',
  visible: false,
})

watch(() => props.squad, (squad) => {
  Object.assign(draft, {
    description: squad.description,
    priceTxb: txbInputFromMinor(squad.price.minor),
    visible: squad.visible,
  })
}, { immediate: true })

function submit(): void {
  const priceTxbMinor = moneyFromTxbInput(draft.priceTxb)
  if (!priceTxbMinor) return
  emit('save', {
    remnaSquadUuid: props.squad.remnaSquadUuid,
    name: props.squad.name,
    description: draft.description.trim(),
    priceTxbMinor,
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
    <div class="catalog-editor__wide markdown-preview"><span>Preview</span><MarkdownContent :source="draft.description || 'Description preview'" compact /></div>
    <TxbAmountField id="squad-price" v-model="draft.priceTxb" label="Price" min-minor="0" required />
    <SwitchField id="squad-visible" v-model="draft.visible" label="Visible in the user catalog" help="Upstream-missing squads remain unavailable even when visible." />
    <button class="button button--primary catalog-editor__wide" type="submit" :disabled="busy">
      <PhFloppyDisk :size="18" /> {{ busy ? 'Saving' : 'Save squad' }}
    </button>
  </form>
</template>

<style scoped>
.markdown-preview {
  padding: 0.7rem;
  border: 1px solid var(--line);
  border-radius: var(--radius-control);
  background: var(--surface);
}

.markdown-preview > span {
  display: block;
  margin-bottom: 0.45rem;
  color: var(--text-faint);
  font-size: 0.68rem;
  font-weight: 700;
}
</style>
