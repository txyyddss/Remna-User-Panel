<script setup lang="ts">
import { computed } from 'vue'

import type { OnboardingLocalizedContent } from '@/api/features'
import { useI18n } from '@/i18n'
import { newWelcomePair, normalizeContentID, welcomeContent, welcomePairs, type BilingualWelcome } from './onboardingEditor'

const props = defineProps<{ content: OnboardingLocalizedContent }>()
const emit = defineEmits<{ 'update:content': [content: OnboardingLocalizedContent] }>()
const { t } = useI18n()
const pairs = computed(() => welcomePairs(props.content))
const canAdd = computed(() => pairs.value.length < 20)

function commit(next: BilingualWelcome[]): void {
  emit('update:content', welcomeContent(next))
}

function updateID(index: number, id: string): void {
  const normalizedID = normalizeContentID(id)
  commit(pairs.value.map((pair, position) => position === index
    ? { ...pair, id: normalizedID, english: { ...pair.english, id: normalizedID }, chinese: { ...pair.chinese, id: normalizedID } }
    : pair))
}

function updateText(index: number, locale: 'english' | 'chinese', text: string): void {
  commit(pairs.value.map((pair, position) => position === index
    ? { ...pair, [locale]: { ...pair[locale], text } }
    : pair))
}

function move(index: number, direction: -1 | 1): void {
  const target = index + direction
  if (target < 0 || target >= pairs.value.length) return
  const next = [...pairs.value]
  const current = next[index]
  next[index] = next[target]
  next[target] = current
  commit(next)
}

function remove(index: number): void {
  if (pairs.value.length <= 1) return
  commit(pairs.value.filter((_, position) => position !== index))
}

function add(): void {
  if (canAdd.value) commit([...pairs.value, newWelcomePair()])
}
</script>

<template>
  <section class="welcome-editor" :aria-label="t('adminOnboarding.welcome')">
    <article v-for="(pair, index) in pairs" :key="pair.id" class="welcome-card">
      <header class="welcome-card__heading">
        <div><UBadge color="neutral" variant="soft" :label="t('adminOnboarding.cardNumber', { number: index + 1 })" /></div>
        <div class="welcome-card__actions">
          <UButton color="neutral" variant="ghost" square icon="i-ph-arrow-up" :disabled="index === 0" :aria-label="t('adminOnboarding.moveUp')" @click="move(index, -1)" />
          <UButton color="neutral" variant="ghost" square icon="i-ph-arrow-down" :disabled="index === pairs.length - 1" :aria-label="t('adminOnboarding.moveDown')" @click="move(index, 1)" />
          <UButton color="error" variant="ghost" square icon="i-ph-trash" :disabled="pairs.length <= 1" :aria-label="t('adminOnboarding.removeCard')" @click="remove(index)" />
        </div>
      </header>
      <UFormField :label="t('adminOnboarding.identifier')">
        <UInput :model-value="pair.id" :maxlength="64" @update:model-value="updateID(index, String($event))" />
      </UFormField>
      <div class="welcome-card__locales">
        <UFormField :label="t('app.english')">
          <UTextarea :model-value="pair.english.text" :rows="4" :maxlength="1000" @update:model-value="updateText(index, 'english', String($event))" />
        </UFormField>
        <UFormField :label="t('app.simplifiedChinese')">
          <UTextarea :model-value="pair.chinese.text" :rows="4" :maxlength="1000" @update:model-value="updateText(index, 'chinese', String($event))" />
        </UFormField>
      </div>
    </article>
    <UButton color="neutral" variant="outline" icon="i-ph-plus" :disabled="!canAdd" :label="t('adminOnboarding.addWelcome')" @click="add" />
  </section>
</template>

<style scoped>
.welcome-editor { display: grid; gap: 0.75rem; padding: 0 1rem; }
.welcome-card { display: grid; gap: 0.7rem; padding: 0.85rem; border: 1px solid var(--line); border-radius: var(--radius-card); background: var(--surface-raised); }
.welcome-card__heading, .welcome-card__actions { display: flex; align-items: center; gap: 0.25rem; }
.welcome-card__heading { justify-content: space-between; }
.welcome-card__locales { display: grid; gap: 0.7rem; }
@media (min-width: 760px) { .welcome-card__locales { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
</style>
