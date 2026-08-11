<script setup lang="ts">
import { computed } from 'vue'

import type { OnboardingLocalizedContent } from '@/api/features'
import { agreementIcon } from '@/components/onboarding/agreementIcons'
import { useI18n } from '@/i18n'
import {
  agreementColors,
  agreementContent,
  agreementIconKeys,
  agreementPairs,
  newAgreementPair,
  normalizeContentID,
  type AgreementColor,
  type BilingualAgreement,
  type EditableAgreement,
} from './onboardingEditor'

const props = defineProps<{ content: OnboardingLocalizedContent }>()
const emit = defineEmits<{ 'update:content': [content: OnboardingLocalizedContent] }>()
const { t } = useI18n()
const pairs = computed(() => agreementPairs(props.content))
const canAdd = computed(() => pairs.value.length < 30)

function commit(next: BilingualAgreement[]): void {
  emit('update:content', agreementContent(next))
}

function change(index: number, update: (pair: BilingualAgreement) => BilingualAgreement): void {
  commit(pairs.value.map((pair, position) => position === index ? update(pair) : pair))
}

function updateID(index: number, id: string): void {
  const normalizedID = normalizeContentID(id)
  change(index, (pair) => ({ ...pair, id: normalizedID, english: { ...pair.english, id: normalizedID }, chinese: { ...pair.chinese, id: normalizedID } }))
}

function updateText(index: number, locale: 'english' | 'chinese', field: 'title' | 'body' | 'pageTitle', value: string): void {
  change(index, (pair) => ({ ...pair, [locale]: { ...pair[locale], [field]: value } as EditableAgreement }))
}

function setIcon(index: number, icon: BilingualAgreement['english']['icon']): void {
  change(index, (pair) => ({ ...pair, english: { ...pair.english, icon }, chinese: { ...pair.chinese, icon } }))
}

function setColor(index: number, color: AgreementColor): void {
  change(index, (pair) => ({ ...pair, english: { ...pair.english, color }, chinese: { ...pair.chinese, color } }))
}

function move(index: number, direction: -1 | 1): void {
  const target = index + direction
  if (target < 0 || target >= pairs.value.length) return
  const pageTitles = { english: pairs.value[0].english.pageTitle, chinese: pairs.value[0].chinese.pageTitle }
  const next = [...pairs.value]
  const current = next[index]
  next[index] = next[target]
  next[target] = current
  commit(next.map((pair, position) => ({
    ...pair,
    english: { ...pair.english, ...(position === 0 && pageTitles.english ? { pageTitle: pageTitles.english } : { pageTitle: undefined }) },
    chinese: { ...pair.chinese, ...(position === 0 && pageTitles.chinese ? { pageTitle: pageTitles.chinese } : { pageTitle: undefined }) },
  })))
}

function remove(index: number): void {
  if (pairs.value.length <= 1) return
  const pageTitles = { english: pairs.value[0].english.pageTitle, chinese: pairs.value[0].chinese.pageTitle }
  commit(pairs.value.filter((_, position) => position !== index).map((pair, position) => ({
    ...pair,
    english: { ...pair.english, ...(position === 0 && pageTitles.english ? { pageTitle: pageTitles.english } : { pageTitle: undefined }) },
    chinese: { ...pair.chinese, ...(position === 0 && pageTitles.chinese ? { pageTitle: pageTitles.chinese } : { pageTitle: undefined }) },
  })))
}

function add(): void {
  if (canAdd.value) commit([...pairs.value, newAgreementPair()])
}
</script>

<template>
  <section class="agreement-editor" :aria-label="t('adminOnboarding.agreements')">
    <article v-for="(pair, index) in pairs" :key="pair.id" class="agreement-card" :class="`agreement-card--${pair.english.color ?? 'warning'}`">
      <header class="agreement-card__heading">
        <div class="agreement-card__identity"><span class="agreement-card__icon"><UIcon :name="agreementIcon(pair.english.icon)" /></span><UBadge color="neutral" variant="soft" :label="t('adminOnboarding.cardNumber', { number: index + 1 })" /></div>
        <div class="agreement-card__actions">
          <UButton color="neutral" variant="ghost" square icon="i-ph-arrow-up" :disabled="index === 0" :aria-label="t('adminOnboarding.moveUp')" @click="move(index, -1)" />
          <UButton color="neutral" variant="ghost" square icon="i-ph-arrow-down" :disabled="index === pairs.length - 1" :aria-label="t('adminOnboarding.moveDown')" @click="move(index, 1)" />
          <UButton color="error" variant="ghost" square icon="i-ph-trash" :disabled="pairs.length <= 1" :aria-label="t('adminOnboarding.removeCard')" @click="remove(index)" />
        </div>
      </header>
      <UFormField :label="t('adminOnboarding.identifier')"><UInput :model-value="pair.id" :maxlength="64" @update:model-value="updateID(index, String($event))" /></UFormField>
      <div class="agreement-card__controls">
        <UFormField :label="t('adminOnboarding.icon')">
          <div class="icon-picker" :aria-label="t('adminOnboarding.icon')">
            <UButton v-for="icon in agreementIconKeys" :key="icon" color="neutral" variant="outline" square :class="{ 'is-selected': pair.english.icon === icon }" :icon="agreementIcon(icon)" :aria-label="t(`adminOnboarding.icons.${icon}`)" @click="setIcon(index, icon)" />
          </div>
        </UFormField>
        <UFormField :label="t('adminOnboarding.color')">
          <div class="color-picker" :aria-label="t('adminOnboarding.color')">
            <UButton v-for="color in agreementColors" :key="color" color="neutral" variant="outline" square :class="['color-picker__item', `color-picker__item--${color}`, { 'is-selected': pair.english.color === color }]" :icon="pair.english.color === color ? 'i-ph-check' : undefined" :aria-label="t(`adminOnboarding.colors.${color}`)" @click="setColor(index, color)" />
          </div>
        </UFormField>
      </div>
      <div v-if="index === 0" class="agreement-card__locales">
        <UFormField :label="t('adminOnboarding.englishPageTitle')"><UInput :model-value="pair.english.pageTitle ?? ''" :maxlength="200" @update:model-value="updateText(index, 'english', 'pageTitle', String($event))" /></UFormField>
        <UFormField :label="t('adminOnboarding.chinesePageTitle')"><UInput :model-value="pair.chinese.pageTitle ?? ''" :maxlength="200" @update:model-value="updateText(index, 'chinese', 'pageTitle', String($event))" /></UFormField>
      </div>
      <div class="agreement-card__locales">
        <UFormField :label="t('app.english')"><UInput :model-value="pair.english.title" :maxlength="200" @update:model-value="updateText(index, 'english', 'title', String($event))" /></UFormField>
        <UFormField :label="t('app.simplifiedChinese')"><UInput :model-value="pair.chinese.title" :maxlength="200" @update:model-value="updateText(index, 'chinese', 'title', String($event))" /></UFormField>
      </div>
      <div class="agreement-card__locales">
        <UFormField :label="t('adminOnboarding.englishBody')"><UTextarea :model-value="pair.english.body" :rows="4" :maxlength="2000" @update:model-value="updateText(index, 'english', 'body', String($event))" /></UFormField>
        <UFormField :label="t('adminOnboarding.chineseBody')"><UTextarea :model-value="pair.chinese.body" :rows="4" :maxlength="2000" @update:model-value="updateText(index, 'chinese', 'body', String($event))" /></UFormField>
      </div>
    </article>
    <UButton color="neutral" variant="outline" icon="i-ph-plus" :disabled="!canAdd" :label="t('adminOnboarding.addAgreement')" @click="add" />
  </section>
</template>

<style scoped>
.agreement-editor { display: grid; gap: 0.75rem; padding: 0 1rem; }
.agreement-card { display: grid; gap: 0.7rem; padding: 0.85rem; border: 1px solid var(--line); border-radius: var(--radius-card); background: var(--surface-raised); }
.agreement-card--accent { border-color: var(--accent); }.agreement-card--success { border-color: var(--success); }.agreement-card--warning { border-color: var(--warning); }.agreement-card--danger { border-color: var(--danger); }
.agreement-card__heading, .agreement-card__identity, .agreement-card__actions, .icon-picker, .color-picker { display: flex; align-items: center; gap: 0.3rem; }
.agreement-card__heading { justify-content: space-between; }.agreement-card__icon { display: grid; width: 30px; height: 30px; place-items: center; border-radius: 9px; color: var(--accent); background: var(--accent-soft); }
.agreement-card__controls, .agreement-card__locales { display: grid; gap: 0.7rem; }.icon-picker :deep(button.is-selected), .color-picker :deep(button.is-selected) { box-shadow: 0 0 0 2px var(--text); }
.color-picker__item { color: var(--text); }.color-picker__item--accent { background: var(--accent-soft); }.color-picker__item--success { background: var(--success-soft); }.color-picker__item--warning { background: var(--warning-soft); }.color-picker__item--danger { background: var(--danger-soft); }.color-picker__item--neutral { background: var(--surface); }
@media (min-width: 760px) { .agreement-card__controls { grid-template-columns: repeat(2, minmax(0, 1fr)); }.agreement-card__locales { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
</style>
