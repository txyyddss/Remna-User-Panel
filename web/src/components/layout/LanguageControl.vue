<script setup lang="ts">
import { en, zh_cn } from '@nuxt/ui/locale'

import { useI18n } from '@/i18n'
import { selectionHaptic } from '@/utils/telegram'

interface Props {
  showLabel?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showLabel: false,
})

const uiLocales = [en, zh_cn]
const { locale, setLocale } = useI18n()

function selectLocale(value: string): void {
  if (value === locale.value) return
  if (value !== 'en' && value !== 'zh-CN') return
  setLocale(value)
  selectionHaptic()
}
</script>

<template>
  <div class="language-control">
    <span v-if="props.showLabel" class="language-control__label">{{ $t('app.language') }}</span>
    <ULocaleSelect
      :model-value="locale"
      :locales="uiLocales"
      color="neutral"
      variant="ghost"
      size="sm"
      :search-input="false"
      :aria-label="$t('app.language')"
      @update:model-value="selectLocale"
    />
  </div>
</template>
