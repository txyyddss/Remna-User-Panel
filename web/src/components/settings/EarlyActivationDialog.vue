<script setup lang="ts">
import { computed, shallowRef, watch } from 'vue'
import type { EarlyActivationQuote } from '@/api/types'
import { useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'

const open = defineModel<boolean>('open', { required: true })
defineProps<{ quote: EarlyActivationQuote | null; busy: boolean }>()
const emit = defineEmits<{ confirm: [confirmation: string] }>()
const { t } = useI18n()
const confirmation = shallowRef('')
const confirmed = computed(() => confirmation.value.trim() === t('settings.combo.confirmation'))
watch(open, () => { confirmation.value = '' })
</script>

<template>
  <UModal v-model:open="open" :title="$t('settings.combo.activateTitle')" :description="$t('settings.combo.forfeitWarning')" :dismissible="!busy" :close="!busy">
    <template #body>
      <div class="early-activation">
        <p v-if="quote">{{ $t('settings.combo.forfeitUntil', { date: formatDateTime(quote.currentValidUntil) }) }}</p>
        <p v-if="quote">{{ $t('settings.combo.newPeriod', { start: formatDateTime(quote.newValidFrom), end: formatDateTime(quote.newValidUntil) }) }}</p>
        <UFormField :label="$t('settings.combo.confirmPrompt', { text: $t('settings.combo.confirmation') })">
          <UInput v-model="confirmation" class="w-full" :disabled="busy" autocomplete="off" :aria-label="$t('settings.combo.confirmPrompt', { text: $t('settings.combo.confirmation') })" />
        </UFormField>
      </div>
    </template>
    <template #footer>
      <UButton color="neutral" variant="outline" :label="$t('common.cancel')" :disabled="busy" @click="open = false" />
      <UButton color="error" :label="$t('settings.combo.activate')" :disabled="!confirmed || busy || !quote?.eligible" :loading="busy" @click="emit('confirm', confirmation.trim())" />
    </template>
  </UModal>
</template>

<style scoped>
.early-activation { display: grid; gap: 1rem; }
.early-activation p { margin: 0; color: var(--text-muted); font-size: 0.875rem; line-height: 1.5; }
</style>
