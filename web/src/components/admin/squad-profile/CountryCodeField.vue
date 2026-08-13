<script setup lang="ts">
import { shallowRef, watch } from 'vue'

import { isCountryCode, normalizeCountryCode } from '@/components/squad-profile/profile'
import { useI18n } from '@/i18n'

const props = defineProps<{
  id: string
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const { t } = useI18n()
const lastValidCode = shallowRef('')
const pendingInput = shallowRef<string | null>(null)

watch(() => props.modelValue, (value) => {
  const normalized = normalizeCountryCode(value)
  if (pendingInput.value === normalized) {
    pendingInput.value = null
    if (isCountryCode(normalized)) lastValidCode.value = normalized
    return
  }
  pendingInput.value = null
  lastValidCode.value = isCountryCode(normalized) ? normalized : ''
}, { immediate: true })

function update(value: string): void {
  const normalized = normalizeCountryCode(value)
  pendingInput.value = normalized
  emit('update:modelValue', normalized)
}

function restoreIfBlank(): void {
  if (!props.modelValue && lastValidCode.value) update(lastValidCode.value)
}
</script>

<template>
  <UFormField :name="props.id" :label="t('squadProfile.country')" :hint="t('squadProfile.countryHint')" required>
    <UInput
      :id="props.id"
      :model-value="props.modelValue"
      :placeholder="t('squadProfile.countryPlaceholder')"
      maxlength="2"
      minlength="2"
      pattern="[A-Za-z]{2}"
      inputmode="text"
      autocomplete="off"
      autocapitalize="characters"
      spellcheck="false"
      required
      @update:model-value="update"
      @blur="restoreIfBlank"
    />
  </UFormField>
</template>

