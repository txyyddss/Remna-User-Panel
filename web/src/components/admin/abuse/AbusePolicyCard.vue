<script setup lang="ts">
import { reactive, shallowRef, watch } from 'vue'
import type { AbusePolicy, AbusePunishment } from '@/api/abuse'
import { useI18n } from '@/i18n'

const props = defineProps<{ policy: AbusePolicy; punishments: AbusePunishment[]; busy: boolean }>()
const emit = defineEmits<{ savePolicy: [value: AbusePolicy]; savePunishment: [value: AbusePunishment] }>()
const { t } = useI18n()
const form = reactive<AbusePolicy>({ ...props.policy })
const ladder = shallowRef(props.punishments.map(item => ({ ...item })))

watch(() => props.policy, value => Object.assign(form, value))
watch(() => props.punishments, value => { ladder.value = value.map(item => ({ ...item })) })

const actionLabel = (action: string) => t(`abuse.action.${action}`)
</script>

<template>
  <div class="policy-stack">
    <section class="card">
      <h3>{{ t('adminAbuse.policyTitle') }}</h3>
      <UForm :state="form" @submit="emit('savePolicy', { ...form })">
        <UFormField :label="t('adminAbuse.globalEnabled')">
          <USwitch v-model="form.globalEnabled" />
        </UFormField>
        <UFormField :label="t('adminAbuse.globalLimit')">
          <UInput v-model.number="form.globalLimit" min="0" type="number" />
        </UFormField>
        <UFormField :label="t('adminAbuse.validity')">
          <UInput v-model.number="form.warningValidityDays" min="1" type="number" />
        </UFormField>
        <UFormField :label="t('adminAbuse.warningCooldown')">
          <UInput v-model.number="form.warningCooldownMinutes" min="0" max="525600" type="number" />
        </UFormField>
        <UButton type="submit" :loading="busy" :label="t('common.save')" />
      </UForm>
    </section>

    <section class="card">
      <h3>{{ t('adminAbuse.punishmentsTitle') }}</h3>
      <UForm
        v-for="item in ladder"
        :key="item.action"
        :state="item"
        class="ladder"
        @submit="emit('savePunishment', { ...item })"
      >
        <strong>{{ actionLabel(item.action) }}</strong>
        <USwitch v-model="item.enabled" />
        <UInput v-model.number="item.incidentThreshold" type="number" />
        <UInput v-model.number="item.durationMinutes" type="number" />
        <UButton type="submit" size="sm" :loading="busy" :label="t('common.save')" />
      </UForm>
    </section>
  </div>
</template>

<style scoped>
.policy-stack,
.card,
.card :deep(form) {
  display: grid;
  gap: 0.75rem;
}

.card {
  padding: 1rem;
  border: 1px solid var(--line);
  border-radius: var(--radius-panel);
  background: var(--surface-raised);
}

.ladder {
  grid-template-columns: minmax(0, 1fr) auto 4.5rem 4.5rem auto;
  align-items: end;
  border-top: 1px solid var(--line);
  padding-top: 0.7rem;
}

@media (max-width: 620px) {
  .ladder {
    grid-template-columns: minmax(0, 1fr) auto;
  }

  .ladder strong {
    grid-column: 1 / -1;
  }
}
</style>
