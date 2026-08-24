<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import { useRouter } from 'vue-router'

import type { Purchase, SquadProduct } from '@/api/types'
import CatalogSquadStep from '@/components/catalog/CatalogSquadStep.vue'
import SquadActivationDialog from '@/components/catalog/SquadActivationDialog.vue'
import { useCatalogSquadPresentation } from '@/components/catalog/useCatalogSquadPresentation'
import { useSquadAddition } from '@/composables/useSquadAddition'

const open = defineModel<boolean>('open', { required: true })
const props = defineProps<{ active: Purchase }>()
const emit = defineEmits<{ changed: [] }>()
const router = useRouter()
const step = shallowRef(1)
const activationOpen = shallowRef(false)
const activationTarget = shallowRef<SquadProduct | null>(null)
const activationPrompting = shallowRef(false)
let activationResolver: ((code: string | null) => void) | null = null
const emptyIncludedIds = computed<readonly string[]>(() => [])
const comboId = computed<string | null>(() => props.active.comboId)
const activeSquadUuids = computed(() => props.active.squadUuids)
const {
  selectedSquadIds, quote, purchase, loading, quoting, purchasing, needsBalance, error,
  visibleSquads, selectedSquads, activationSquads, reset, toggleSquad, load, refreshQuote, confirmPurchase,
} = useSquadAddition({ purchaseId: () => props.active.id, activeSquadUuids: () => activeSquadUuids.value })
const { featuredIds, orderedIds } = useCatalogSquadPresentation(visibleSquads, emptyIncludedIds, comboId)

watch(open, (visible) => {
  if (!visible) {
    if (activationResolver) resolveActivation(null)
    return
  }
  step.value = 1
  reset()
  void load()
})

async function continueToCheckout(): Promise<void> {
  if (!selectedSquadIds.value.length || !(await refreshQuote())) return
  step.value = 2
}

async function requestActivationCode(squad: SquadProduct): Promise<string | null> {
  activationTarget.value = squad
  activationOpen.value = true
  return new Promise((resolve) => { activationResolver = resolve })
}

function resolveActivation(code: string | null): void {
  const resolve = activationResolver
  activationResolver = null
  activationOpen.value = false
  activationTarget.value = null
  resolve?.(code)
}

async function confirm(): Promise<void> {
  if (activationPrompting.value) return
  activationPrompting.value = true
  try {
    const codes: Record<string, string> = {}
    for (const squad of activationSquads.value) {
      const code = await requestActivationCode(squad)
      if (!code) return
      codes[squad.remnaSquadUuid] = code
    }
    if (await confirmPurchase(codes)) emit('changed')
  } finally {
    activationPrompting.value = false
  }
}

function close(): void {
  open.value = false
}

function goHome(): void {
  close()
  void router.push('/home')
}
</script>

<template>
  <UModal v-model:open="open" :title="$t('home.squadAddition.title')" :description="$t('home.squadAddition.description')" scrollable>
    <template #body>
      <div class="squad-addition-dialog">
        <UStepper v-model="step" :items="[{ title: $t('home.squadAddition.steps.choose') }, { title: $t('home.squadAddition.steps.checkout') }]" :linear="true" :disabled="loading || purchasing" size="sm" />
        <USkeleton v-if="loading" class="h-48" />
        <CatalogSquadStep v-else-if="step === 1" :squads="visibleSquads" :selected-ids="selectedSquadIds" :included-ids="emptyIncludedIds" :featured-ids="featuredIds" :ordered-ids="orderedIds" @toggle="toggleSquad" />
        <SquadAdditionCheckout v-else :squads="selectedSquads" :quote="quote" :purchase="purchase" :quoting="quoting" :purchasing="purchasing || activationPrompting" :needs-balance="needsBalance" :error="error" @back="step = 1" @confirm="confirm" @home="goHome" />
        <UAlert v-if="step === 1 && error" color="warning" variant="soft" icon="i-ph-warning-circle" :description="error" />
      </div>
    </template>
    <template v-if="step === 1" #footer>
      <UButton color="neutral" variant="outline" :label="$t('common.cancel')" data-haptic="dismiss" @click="close" />
      <UButton :disabled="loading || !selectedSquadIds.length" :label="$t('catalog.continue')" data-haptic="confirm" @click="continueToCheckout" />
    </template>
  </UModal>
  <SquadActivationDialog v-model:open="activationOpen" :squad="activationTarget" @submit="resolveActivation" @cancel="resolveActivation(null)" />
</template>

<style scoped>
.squad-addition-dialog { display: grid; gap: 1rem; min-width: 0; padding-bottom: max(0.25rem, env(safe-area-inset-bottom)); }
</style>
