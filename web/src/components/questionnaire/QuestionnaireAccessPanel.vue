<script setup lang="ts">
import { PhArrowSquareOut, PhCheckCircle, PhClipboardText, PhCopy } from '@phosphor-icons/vue'

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
      <span class="feature-icon"><PhClipboardText :size="23" /></span>
      <div><h2>{{ questionnaire.title }}</h2><p>Reward {{ txbInputFromMinor(questionnaire.rewardTxbMinor) }} TXB after validation.</p></div>
    </div>
    <MarkdownContent :source="questionnaire.description" />
    <div v-if="participation" class="validation-code">
      <span>Your validation code</span>
      <div><code>{{ participation.validationCode }}</code><button class="icon-button" type="button" :aria-label="copied ? 'Code copied' : 'Copy validation code'" @click="copy(participation.validationCode)"><PhCheckCircle v-if="copied" :size="19" /><PhCopy v-else :size="19" /></button></div>
      <small>Paste this exact code into the external form. You can return here to retrieve it.</small>
    </div>
    <div class="questionnaire-card__footer">
      <span>{{ questionnaire.closesAt ? `Closes ${formatDateTime(questionnaire.closesAt)}` : 'No closing time announced' }}</span>
      <button class="button button--primary" type="button" :disabled="joining" @click="$emit('open')">
        <PhArrowSquareOut :size="18" /> {{ joining ? 'Preparing code' : participation ? 'Open form again' : 'Get code and open form' }}
      </button>
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

@media (min-width: 640px) {
  .questionnaire-card__footer { flex-direction: row; align-items: center; justify-content: space-between; }
}
</style>
