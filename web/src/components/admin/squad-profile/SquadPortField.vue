<script setup lang="ts">
import SwitchField from '@/components/common/SwitchField.vue'
import { useI18n } from '@/i18n'

const props = withDefaults(defineProps<{
  id: string
  label: string
  allowUnlimited?: boolean
}>(), { allowUnlimited: false })
const port = defineModel<number | undefined>({ required: true })
const unlimited = defineModel<boolean>('unlimited', { required: true })
const { t } = useI18n()
</script>

<template>
  <div class="port-field">
    <UFormField :name="id" :label="label" :required="!allowUnlimited">
      <UInputNumber v-model="port" :min="1" :step="1" :disabled="allowUnlimited && unlimited" :placeholder="t('squadProfile.portPlaceholder')" />
    </UFormField>
    <SwitchField v-if="props.allowUnlimited" :id="`${id}-unlimited`" v-model="unlimited" :label="t('squadProfile.unlimitedPort')" :help="t('squadProfile.unlimitedPortHint')" />
  </div>
</template>

<style scoped>
.port-field { display: grid; gap: 0.5rem; }
</style>
