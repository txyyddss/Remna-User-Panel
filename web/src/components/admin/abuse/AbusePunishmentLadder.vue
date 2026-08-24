<script setup lang="ts">
import { shallowRef, watch } from 'vue'
import type { AbusePunishment } from '@/api/abuse'
import { useI18n } from '@/i18n'

const props = defineProps<{ punishments: AbusePunishment[]; busy: boolean }>()
const emit = defineEmits<{ save: [value: AbusePunishment] }>()
type FormError = { name?: string; message: string }
const { t } = useI18n()
const ladder = shallowRef(props.punishments.map(item => ({ ...item })))

watch(() => props.punishments, value => { ladder.value = value.map(item => ({ ...item })) })

function validate(value: Partial<AbusePunishment>): FormError[] {
  const errors: FormError[] = []
  const { incidentThreshold = 0, durationMinutes = 0 } = value
  if (!Number.isInteger(incidentThreshold) || incidentThreshold < 1 || incidentThreshold > 100000) errors.push({ name: 'incidentThreshold', message: t('adminAbuse.invalidThreshold') })
  if (!Number.isInteger(durationMinutes) || durationMinutes < 1 || durationMinutes > 525600) errors.push({ name: 'durationMinutes', message: t('adminAbuse.invalidDuration') })
  return errors
}
</script>

<template>
  <section class="card ladder-card">
    <div>
      <p class="eyebrow">{{ t('adminAbuse.escalationEyebrow') }}</p>
      <h3>{{ t('adminAbuse.punishmentsTitle') }}</h3>
    </div>
    <UForm
      v-for="item in ladder"
      :key="item.action"
      :state="item"
      :validate="validate"
      class="ladder"
      @submit="emit('save', { ...item })"
    >
      <strong>{{ t(`abuse.action.${item.action}`) }}</strong>
      <UFormField name="enabled" :label="t('adminAbuse.ruleEnabled')">
        <USwitch v-model="item.enabled" />
      </UFormField>
      <UFormField name="incidentThreshold" :label="t('adminAbuse.incidentThreshold')">
        <UInputNumber v-model="item.incidentThreshold" :min="1" :max="100000" :step="1" :disable-wheel-change="true" />
      </UFormField>
      <UFormField name="durationMinutes" :label="t('adminAbuse.duration')">
        <UInputNumber v-model="item.durationMinutes" :min="1" :max="525600" :step="1" :disable-wheel-change="true" />
      </UFormField>
      <UButton type="submit" size="sm" :loading="busy" :label="t('common.save')" />
    </UForm>
  </section>
</template>

<style scoped>
.card,
.ladder {
  display: grid;
  gap: 0.85rem;
}

.card {
  padding: 1rem;
  border: 1px solid var(--line);
  border-radius: var(--radius-panel);
  background: var(--surface-raised);
}

.ladder-card {
  grid-column: 1 / -1;
}

.ladder {
  grid-template-columns: minmax(9rem, 1.4fr) minmax(6.5rem, 0.6fr) repeat(2, minmax(8rem, 1fr)) auto;
  align-items: end;
  border-top: 1px solid var(--line);
  padding-top: 0.85rem;
}

@media (max-width: 880px) {
  .ladder { grid-template-columns: 1fr 1fr; }
  .ladder > :first-child { grid-column: 1 / -1; }
  .ladder > :last-child { justify-self: start; }
}
</style>
