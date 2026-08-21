<script setup lang="ts">
import { shallowRef } from 'vue'

import { useI18n } from '@/i18n'
import type { Locale } from '@/i18n/generated'
import { selectionHaptic } from '@/utils/telegram'

interface Props {
  showLabel?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showLabel: false,
})

const open = shallowRef(false)
const { locale, locales, setLocale, t } = useI18n()

function localeLabel(value: Locale): string {
  return value === 'en' ? t('app.english') : t('app.simplifiedChinese')
}

function selectLocale(value: Locale): void {
  if (value === locale.value) {
    open.value = false
    return
  }
  setLocale(value)
  selectionHaptic()
  open.value = false
}
</script>

<template>
  <UPopover
    v-model:open="open"
    :content="{ side: 'top', align: 'end', sideOffset: 12, collisionPadding: 16 }"
  >
    <UButton
      class="language-control"
      color="neutral"
      variant="ghost"
      icon="i-ph-translate"
      :label="props.showLabel ? $t('app.language') : undefined"
      :aria-label="$t('app.language')"
    />
    <template #content>
      <div class="language-control__panel" :aria-label="$t('app.language')" role="group">
        <UButton
          v-for="value in locales"
          :key="value"
          class="language-control__option"
          color="neutral"
          :variant="locale === value ? 'soft' : 'ghost'"
          :aria-pressed="locale === value"
          @click="selectLocale(value)"
        >
          {{ localeLabel(value) }}
        </UButton>
      </div>
    </template>
  </UPopover>
</template>
