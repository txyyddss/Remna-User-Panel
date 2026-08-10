<script setup lang="ts">
import { CheckboxIndicator, CheckboxRoot } from 'reka-ui'
import { PhArrowRight, PhCheck, PhShieldWarning } from '@phosphor-icons/vue'

import type { OnboardingAgreement } from '@/api/features'
import { agreementIcon } from './agreementIcons'

defineProps<{ agreements: readonly OnboardingAgreement[]; selectedIds: readonly string[]; allAccepted: boolean; loading: boolean }>()
defineEmits<{ submit: []; toggle: [id: string] }>()
</script>

<template>
  <section class="onboarding-panel">
    <header class="onboarding-panel__header">
      <span class="feature-icon feature-icon--warning"><PhShieldWarning :size="25" weight="fill" /></span>
      <h1>{{ $t('onboarding.agreementTitle') }}</h1>
      <p>{{ $t('onboarding.agreementCopy') }}</p>
    </header>

    <div class="agreement-list">
      <label v-for="agreement in agreements" :key="agreement.id" class="agreement-callout agreement-callout--selectable">
        <component :is="agreementIcon(agreement.icon)" :size="23" aria-hidden="true" />
        <span><strong>{{ agreement.title }}</strong><small>{{ agreement.body }}</small></span>
        <CheckboxRoot class="checkbox-control" :model-value="selectedIds.includes(agreement.id)" @update:model-value="$emit('toggle', agreement.id)">
          <CheckboxIndicator class="checkbox-indicator"><PhCheck :size="16" weight="bold" /></CheckboxIndicator>
        </CheckboxRoot>
      </label>
    </div>

    <button class="button button--primary button--wide" type="button" :disabled="!allAccepted || loading" @click="$emit('submit')">
      {{ loading ? $t('onboarding.finishing') : $t('onboarding.finish') }}
      <PhArrowRight :size="19" />
    </button>
  </section>
</template>

<style scoped>
.agreement-list { display: grid; gap: 0.65rem; }
.agreement-callout--selectable { display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: start; cursor: pointer; }
.agreement-callout--selectable strong, .agreement-callout--selectable small { display: block; }
.agreement-callout--selectable small { margin-top: 0.25rem; color: var(--text-muted); font-size: 0.72rem; line-height: 1.5; }
</style>
