<script setup lang="ts">
import { useI18n } from '@/i18n'
import type { Locale } from '@/i18n/generated'

const { locale, locales, setLocale, t } = useI18n()
</script>

<template>
  <label class="language-switcher">
    <span class="sr-only">{{ t('app.language') }}</span>
    <select :value="locale" :aria-label="t('app.language')" @change="setLocale(($event.target as HTMLSelectElement).value as Locale)">
      <option v-for="item in locales" :key="item" :value="item">{{ item === 'en' ? t('app.english') : t('app.simplifiedChinese') }}</option>
    </select>
  </label>
</template>
