<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'

import type { Purchase } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useAutoRenewal } from '@/composables/useAutoRenewal'
import { t } from '@/i18n'
import { formatDate, formatMoney } from '@/utils/format'

const props = defineProps<{ purchase: Purchase }>()
const emit = defineEmits<{ changed: [] }>()

const open = shallowRef(false)
const switchValue = shallowRef(props.purchase.autoRenewEnabled)
const { renewal, loading, updating, error, enabled, canEnable, ineligibleReason, reset, load, setEnabled } = useAutoRenewal(
  () => props.purchase.id,
  () => props.purchase.autoRenewEnabled,
)

const actionColor = computed(() => enabled.value ? 'success' : 'error')
const actionLabel = computed(() => t(enabled.value ? 'home.autoRenewalEnabledAction' : 'home.autoRenewalDisabledAction'))
const eligibilityMessage = computed(() => localizedReason(ineligibleReason.value, 'home.autoRenewalUnavailable'))
const showSwitch = computed(() => canEnable.value || enabled.value)

watch(open, (next) => {
  if (next) void load()
})
watch(renewal, (next) => {
  if (next) switchValue.value = next.enabled
})
watch(() => props.purchase.id, () => {
  reset()
  switchValue.value = props.purchase.autoRenewEnabled
  if (open.value) void load()
})

function localizedReason(reason: string | null, fallback: string): string {
  if (!reason) return t(fallback)
  const key = `home.autoRenewalReason.${reason}`
  const localized = t(key)
  return localized === key ? t(fallback) : localized
}

async function updateRenewal(next: boolean): Promise<void> {
  const previous = enabled.value
  if (next && !canEnable.value) {
    switchValue.value = previous
    return
  }
  switchValue.value = next
  if (await setEnabled(next)) emit('changed')
  else switchValue.value = previous
}
</script>

<template>
  <div class="auto-renewal-control">
    <UButton block class="auto-renewal-control__action" :color="actionColor" variant="soft" :label="actionLabel" data-haptic @click="open = true" />
    <UModal v-model:open="open" :title="$t('home.autoRenewalTitle')" :description="$t('home.autoRenewalHint')">
      <template #body>
        <div class="auto-renewal-control__dialog">
          <USkeleton v-if="loading" class="h-36" />
          <template v-else-if="renewal">
            <dl class="auto-renewal-control__quote">
              <div><dt>{{ $t('home.autoRenewalGross') }}</dt><dd>{{ formatMoney(renewal.grossPrice) }}</dd></div>
              <div><dt>{{ $t('home.autoRenewalDiscount') }}</dt><dd>{{ formatMoney(renewal.discount) }}</dd></div>
              <div><dt>{{ $t('home.autoRenewalPrice') }}</dt><dd>{{ formatMoney(renewal.netPrice) }}</dd></div>
              <div><dt>{{ $t('home.autoRenewalChargeDate') }}</dt><dd>{{ formatDate(renewal.scheduledAt) }}</dd></div>
              <div><dt>{{ $t('home.autoRenewalNextCycleDate') }}</dt><dd>{{ formatDate(renewal.nextCycleEndsAt) }}</dd></div>
            </dl>
            <USwitch v-if="showSwitch" v-model="switchValue" :color="enabled ? 'success' : 'error'" :label="$t('home.autoRenewalSwitch')" :loading="updating" :disabled="updating" data-haptic @update:model-value="updateRenewal" />
            <InlineNotice v-else tone="warning">{{ eligibilityMessage }}</InlineNotice>
          </template>
          <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
        </div>
      </template>
    </UModal>
  </div>
</template>

<style scoped>
.auto-renewal-control, .auto-renewal-control__dialog { display: grid; gap: 0.75rem; }
.auto-renewal-control__action { justify-content: center; }
.auto-renewal-control__quote { display: grid; gap: 0.45rem; margin: 0; }
.auto-renewal-control__quote > div { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; padding-bottom: 0.45rem; border-bottom: 1px solid var(--line); }
.auto-renewal-control__quote dt { color: var(--text-faint); font-size: 0.7rem; }
.auto-renewal-control__quote dd { margin: 0; font-family: var(--font-mono); font-size: 0.78rem; font-weight: 700; text-align: right; }
</style>
