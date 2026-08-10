<script setup lang="ts">
import type { OnboardingAgreement } from '@/api/features'
import { agreementIcon } from './agreementIcons'

defineProps<{ agreements: readonly OnboardingAgreement[]; selectedIds: readonly string[]; allAccepted: boolean; loading: boolean; showAction?: boolean }>()
defineEmits<{ submit: []; toggle: [id: string] }>()
</script>

<template>
  <section class="onboarding-panel">
    <header class="onboarding-panel__header">
      <span class="feature-icon feature-icon--warning"><UIcon name="i-ph-shield-warning-fill" /></span>
      <h1>{{ $t('onboarding.agreementTitle') }}</h1>
      <p>{{ $t('onboarding.agreementCopy') }}</p>
    </header>

    <div class="agreement-list">
      <label v-for="agreement in agreements" :key="agreement.id" class="agreement-callout agreement-callout--selectable">
        <UIcon :name="agreementIcon(agreement.icon)" aria-hidden="true" />
        <span><strong>{{ agreement.title }}</strong><small>{{ agreement.body }}</small></span>
        <UCheckbox
          :model-value="selectedIds.includes(agreement.id)"
          :aria-label="agreement.title"
          @update:model-value="$emit('toggle', agreement.id)"
        />
      </label>
    </div>

    <UButton
      v-if="showAction !== false"
      block
      trailing-icon="i-ph-arrow-right"
      :disabled="!allAccepted || loading"
      :loading="loading"
      :label="loading ? $t('onboarding.finishing') : $t('onboarding.finish')"
      @click="$emit('submit')"
    />
  </section>
</template>

<style scoped>
.agreement-list { display: grid; gap: 0.65rem; }
.agreement-callout--selectable { display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: start; cursor: pointer; }
.agreement-callout--selectable strong, .agreement-callout--selectable small { display: block; }
.agreement-callout--selectable small { margin-top: 0.25rem; color: var(--text-muted); font-size: 0.72rem; line-height: 1.5; }
</style>
