<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useI18n } from '@/i18n'
import { mobileNavigationItems } from './navigation'

const { isAdmin } = defineProps<{ isAdmin: boolean }>()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const items = computed(() => mobileNavigationItems(isAdmin).map(item => ({
  label: t(item.labelKey), icon: item.icon, value: item.to,
})))

function isCurrent(value: string): boolean {
  if (route.path.startsWith('/admin')) return value === '/admin/settings'
  return value === route.path || (value === '/home' && route.path === '/')
}

function navigate(value: string): void {
  void router.push(value).catch(() => undefined)
}
</script>

<template>
  <nav class="bottom-nav" :class="{ 'bottom-nav--admin': isAdmin }" :aria-label="t('nav.primary')">
    <div class="flex w-full justify-around">
      <button
        v-for="item in items"
        :key="item.value"
        type="button"
        class="bottom-nav__item grow basis-0 flex-col gap-1 py-1 data-[state=active]:text-primary"
        :data-state="isCurrent(item.value) ? 'active' : undefined"
        :aria-current="isCurrent(item.value) ? 'page' : undefined"
        data-haptic="navigate"
        @click="navigate(item.value)"
      >
        <span class="size-5" :class="item.icon" aria-hidden="true" />
        <span class="text-[10px]/3">{{ item.label }}</span>
      </button>
    </div>
  </nav>
</template>
