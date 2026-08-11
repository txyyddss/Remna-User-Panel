<script setup lang="ts">
import { computed, nextTick, useTemplateRef, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { useSessionStore } from '@/stores/session'
import { focusWithoutScrolling } from '@/utils/dom'

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
const adminItem: NavItem = { labelKey: 'nav.admin', to: '/admin/settings', icon: 'i-ph-shield-check' }

const activePath = computed(() => route.path)
const mobileItems = computed(() => sessionStore.isAdmin ? [...primaryItems, adminItem] : primaryItems)
const showBackButton = computed(() => !['/', '/home'].includes(route.path))
const mainContent = useTemplateRef<globalThis.HTMLElement>('mainContent')

function goBack(): void {
  try {
    void Promise.resolve(router.back()).catch(() => undefined)
  } catch {
    // The router may already be disposing the current view.
  }
}

function goTo(to: string): void {
  void router.push(to).catch(() => undefined)
}

useTelegramBackButton(showBackButton, goBack)

watch(() => route.fullPath, (_next, previous) => {
  if (!previous) return
  void nextTick()
    .then(() => {
      if (mainContent.value) focusWithoutScrolling(mainContent.value)
    })
    .catch(() => undefined)
})

function isActive(to: string): boolean {
  if (to === adminItem.to) return activePath.value.startsWith('/admin')
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
          data-haptic
        >
          <UIcon :name="item.icon" />
          <span>{{ $t(item.labelKey) }}</span>
        </RouterLink>
        <RouterLink
          v-if="sessionStore.isAdmin"
          :to="adminItem.to"
          class="nav-item"
          :class="{ 'nav-item--active': isActive(adminItem.to) }"
          data-haptic
        >
          <UIcon :name="adminItem.icon" />
          <span>{{ $t(adminItem.labelKey) }}</span>
        </RouterLink>
      </nav>
    </aside>

    <div class="app-frame__content">
      <main ref="mainContent" class="app-main" tabindex="-1">
        <slot />
      </main>
    </div>

    <nav class="bottom-nav" :class="{ 'bottom-nav--admin': sessionStore.isAdmin }" :aria-label="$t('nav.primary')">
      <UButton
        v-for="item in mobileItems"
        :key="item.to"
        type="button"
        color="neutral"
        variant="ghost"
        class="bottom-nav__item"
        :class="{ 'bottom-nav__item--active': isActive(item.to) }"
        :aria-current="isActive(item.to) ? 'page' : undefined"
        data-haptic
        @click="goTo(item.to)"
      >
        <UIcon :name="item.icon" />
        <span>{{ $t(item.labelKey) }}</span>
      </UButton>
    </nav>
  </div>
</template>
