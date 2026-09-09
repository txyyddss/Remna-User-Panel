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
function navigate(value: string | number): void {
  void router.push(String(value)).catch(() => undefined)
}
</script>

<template>
  <nav class="bottom-nav" :class="{ 'bottom-nav--admin': isAdmin }" :aria-label="t('nav.primary')">
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