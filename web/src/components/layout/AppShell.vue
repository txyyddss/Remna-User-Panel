<script setup lang="ts">
import { computed, nextTick, useTemplateRef, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { useSessionStore } from '@/stores/session'
import LanguageControl from './LanguageControl.vue'

interface NavItem {
  labelKey: string
  to: string
  icon: string
}

const route = useRoute()
const router = useRouter()
const sessionStore = useSessionStore()

const primaryItems: NavItem[] = [
  { labelKey: 'nav.home', to: '/home', icon: 'i-ph-house' },
  { labelKey: 'nav.explore', to: '/catalog', icon: 'i-ph-compass' },
  { labelKey: 'nav.activity', to: '/activity', icon: 'i-ph-game-controller' },
]

const activePath = computed(() => route.path)
const showBackButton = computed(() => !['/', '/home'].includes(route.path))
const mainContent = useTemplateRef<globalThis.HTMLElement>('mainContent')

useTelegramBackButton(showBackButton, () => router.back())

watch(() => route.fullPath, async (_next, previous) => {
  if (!previous) return
  await nextTick()
  mainContent.value?.focus({ preventScroll: true })
})

function isActive(to: string): boolean {
  return activePath.value === to || (to === '/home' && activePath.value === '/')
}
</script>

<template>
  <div class="app-frame">
    <aside class="side-rail" :aria-label="$t('nav.primary')">
      <nav class="side-rail__nav">
        <RouterLink
          v-for="item in primaryItems"
          :key="item.to"
          :to="item.to"
          class="nav-item"
          :class="{ 'nav-item--active': isActive(item.to) }"
        >
          <UIcon :name="item.icon" />
          <span>{{ $t(item.labelKey) }}</span>
        </RouterLink>
        <RouterLink
          v-if="sessionStore.isAdmin"
          to="/admin/settings"
          class="nav-item"
          :class="{ 'nav-item--active': activePath.startsWith('/admin') }"
        >
          <UIcon name="i-ph-shield-check" />
          <span>{{ $t('nav.admin') }}</span>
        </RouterLink>
      </nav>
      <div class="side-rail__footer"><LanguageControl /></div>
    </aside>

    <div class="app-frame__content">
      <main ref="mainContent" class="app-main" tabindex="-1">
        <slot />
      </main>
    </div>

    <nav class="bottom-nav" :aria-label="$t('nav.primary')">
      <RouterLink
        v-for="item in primaryItems"
        :key="item.to"
        :to="item.to"
        class="bottom-nav__item"
        :class="{ 'bottom-nav__item--active': isActive(item.to) }"
      >
        <UIcon :name="item.icon" />
        <span>{{ $t(item.labelKey) }}</span>
      </RouterLink>
      <div class="bottom-nav__locale">
        <LanguageControl />
        <span class="sr-only">{{ $t('app.language') }}</span>
      </div>
    </nav>
  </div>
</template>
