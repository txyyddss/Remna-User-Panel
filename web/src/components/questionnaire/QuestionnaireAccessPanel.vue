<script setup lang="ts">
import type { ActiveQuestionnaire, QuestionnaireParticipation } from '@/api/features'
import MarkdownContent from '@/components/common/MarkdownContent.vue'
import { useClipboard } from '@/composables/useClipboard'
import { formatDateTime, txbInputFromMinor } from '@/utils/format'

defineProps<{
  questionnaire: ActiveQuestionnaire
  participation: QuestionnaireParticipation | null
  joining: boolean
}>()
defineEmits<{ open: [] }>()

const { copied, copy } = useClipboard()
</script>

<template>
  <section class="section-block questionnaire-card">
    <div class="questionnaire-card__heading">
      <span class="feature-icon"><UIcon name="i-ph-clipboard-text" /></span>
      <div>
        <h2>{{ questionnaire.title }}</h2>
        <p>{{ $t('questionnaire.reward', { amount: txbInputFromMinor(questionnaire.rewardTxbMinor) }) }}</p>
      </div>
    </div>
    <MarkdownContent :source="questionnaire.description" />
    <div v-if="participation" class="validation-code">
      <span>{{ $t('questionnaire.validationCode') }}</span>
      <div>
        <code>{{ participation.validationCode }}</code>
        <UButton
          class="icon-button"
          color="neutral"
          variant="ghost"
          :icon="copied ? 'i-ph-check-circle' : 'i-ph-copy'"
          :aria-label="copied ? $t('questionnaire.codeCopied') : $t('questionnaire.copyCode')"
          @click="copy(participation.validationCode)"
        />
      </div>
      <small>{{ $t('questionnaire.codeHint') }}</small>
    </div>
    <div class="questionnaire-card__footer">
      <span>{{ questionnaire.closesAt ? $t('questionnaire.closes', { date: formatDateTime(questionnaire.closesAt) }) : $t('questionnaire.noClose') }}</span>
      <UButton
        icon="i-ph-arrow-square-out"
        :disabled="joining"
        :loading="joining"
        :label="joining ? $t('questionnaire.preparing') : participation ? $t('questionnaire.openAgain') : $t('questionnaire.getCode')"
        @click="$emit('open')"
      />
    </div>
  </section>
</template>

<style scoped>
.questionnaire-card { display: grid; gap: 1rem; }
.questionnaire-card__heading { display: flex; align-items: center; gap: 0.75rem; }
.questionnaire-card__heading h2, .questionnaire-card__heading p { margin: 0; }
.questionnaire-card__heading h2 { font-size: 1.2rem; }
.questionnaire-card__heading p { margin-top: 0.3rem; color: var(--text-muted); font-size: 0.76rem; }
.validation-code { padding: 0.9rem; border: 1px solid #415c4b; border-radius: var(--radius-control); background: var(--accent-soft); }
.validation-code > span, .validation-code small { display: block; color: var(--text-muted); font-size: 0.68rem; }
.validation-code > div { display: flex; align-items: center; justify-content: space-between; gap: 0.6rem; margin: 0.45rem 0; }
.validation-code code { overflow-wrap: anywhere; color: var(--accent); font-family: var(--font-mono); font-size: 1.05rem; font-weight: 700; letter-spacing: 0.06em; }
.questionnaire-card__footer { display: flex; flex-direction: column; gap: 0.75rem; padding-top: 1rem; border-top: 1px solid var(--line); }
.questionnaire-card__footer > span { color: var(--text-faint); font-size: 0.68rem; }
@media (min-width: 640px) { .questionnaire-card__footer { flex-direction: row; align-items: center; justify-content: space-between; } }
</style>
