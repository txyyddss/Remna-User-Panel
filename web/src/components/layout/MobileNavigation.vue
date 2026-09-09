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
const active = computed(() => route.path.startsWith('/admin') ? '/admin/settings'
  : items.value.some(item => item.value === route.path) ? route.path : '/home')

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

function handleNavClick(event: MouseEvent): void {
  const target = event.target as HTMLElement | null
  const trigger = target?.closest('.bottom-nav__item') as HTMLElement | null
  if (!trigger) return

  const nav = trigger.closest('.bottom-nav') as HTMLElement | null
  const index = nav ? Array.from(nav.querySelectorAll('.bottom-nav__item')).indexOf(trigger) : -1
  const nextValue = index >= 0 ? items.value[index]?.value : undefined
  if (!nextValue) return

  void router.push(String(nextValue)).catch(() => undefined)
}
</script>

<template>
  <nav class="bottom-nav" :class="{ 'bottom-nav--admin': props.isAdmin }" :aria-label="t('nav.primary')" @click="handleNavClick">
    <UTabs
      :model-value="active" :items="items" :content="false" color="primary" class="w-full"
      :ui="{
        list: 'justify-around w-full bg-transparent p-0',
        trigger: 'bottom-nav__item grow basis-0 flex-col gap-1 py-1 data-[state=active]:text-primary',
        indicator: 'bg-primary/10 shadow-none duration-320 ease-in-out',
        label: 'text-[10px]/3', leadingIcon: 'size-5',
      }"
      data-haptic="navigate"
      @update:model-value="navigate"
    />
  </nav>
</template>
