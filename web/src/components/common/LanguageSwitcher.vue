<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/i18n'
import type { Locale } from '@/i18n/generated'

const { locale, locales, setLocale, t } = useI18n()
const items = computed(() => locales.value.map((value) => ({
  value,
  label: value === 'en' ? t('app.english') : t('app.simplifiedChinese'),
})))
</script>

<template>
  <USelect
    class="language-switcher"
    :model-value="locale"
    :items="items"
    value-key="value"
    :aria-label="t('app.language')"
    @update:model-value="setLocale($event as Locale)"
  />
</template>
