<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/i18n'
import { mobileNavigationItems } from './navigation'

const props = defineProps<{ isAdmin: boolean }>()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const items = computed(() => mobileNavigationItems(props.isAdmin).map(item => ({
  label: t(item.labelKey), icon: item.icon, value: item.to,
})))

function isCurrent(value: string): boolean {
  if (route.path.startsWith('/admin')) return value === '/admin/settings'
  return value === route.path || (value === '/home' && route.path === '/')
}

function normalizeRouteValue(value: unknown): string | undefined {
  if (typeof value === 'string' || typeof value === 'number') return String(value)
  if (value && typeof value === 'object') {
    if ('value' in value && (typeof value.value === 'string' || typeof value.value === 'number')) {
      return String(value.value)
    }
    if ('to' in value && typeof value.to === 'string') {
      return value.to
    }
  }
  return undefined
}

function navigate(value: unknown): void {
  const nextValue = normalizeRouteValue(value)
  if (!nextValue) return
  void router.push(nextValue).catch(() => undefined)
}
</script>

<template>
  <nav class="bottom-nav" :class="{ 'bottom-nav--admin': props.isAdmin }" :aria-label="t('nav.primary')">
    <div class="flex w-full justify-around">
      <UButton
        v-for="item in items"
        :key="item.value"
        :label="item.label"
        :icon="item.icon"
        :color="isCurrent(item.value) ? 'primary' : 'neutral'"
        variant="ghost"
        class="bottom-nav__item grow basis-0 flex-col gap-1 py-1"
        :data-state="isCurrent(item.value) ? 'active' : undefined"
        :aria-current="isCurrent(item.value) ? 'page' : undefined"
        data-haptic="navigate"
        @click="navigate(item.value)"
      />
    </div>
  </nav>
</template>
