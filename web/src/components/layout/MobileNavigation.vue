<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { LayoutGroup, motion } from 'motion-v'
import { motionSpring } from '@/composables/motionPresets'
import { useMotionPreferences } from '@/composables/useMotionPreferences'
import { useI18n } from '@/i18n'
import { mobileNavigationItems } from './navigation'

const props = defineProps<{ isAdmin: boolean }>()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const { reducedMotion } = useMotionPreferences()
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
    <LayoutGroup>
      <div class="bottom-nav__list" role="tablist">
        <UButton
          v-for="item in items"
          :key="item.value"
          type="button"
          class="bottom-nav__item"
          :class="{ 'bottom-nav__item--active': item.value === active }"
          color="neutral"
          variant="ghost"
          role="tab"
          :aria-selected="item.value === active"
          :tabindex="item.value === active ? 0 : -1"
          data-haptic="navigate"
          @click="navigate(item.value)"
          @keydown.enter.prevent="navigate(item.value)"
        >
          <motion.span v-if="item.value === active" layout-id="mobile-nav-indicator" class="bottom-nav__indicator" :transition="reducedMotion ? { duration: 0.08 } : motionSpring" aria-hidden="true" />
          <motion.span class="bottom-nav__icon" :animate="item.value === active ? { scale: reducedMotion ? 1 : [0.94, 1.04, 1] } : { scale: 1 }" :transition="reducedMotion ? { duration: 0.08 } : { duration: 0.18, ease: 'easeOut' }">
            <UIcon :name="item.icon" />
          </motion.span>
          <span class="bottom-nav__label">{{ item.label }}</span>
        </UButton>
      </div>
    </LayoutGroup>
  </nav>
</template>
