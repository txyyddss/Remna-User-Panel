<script setup lang="ts">
import { computed, nextTick, useTemplateRef, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { useI18n } from '@/i18n'
import { useSessionStore } from '@/stores/session'
import { focusWithoutScrolling } from '@/utils/dom'
import { telegramFullscreenState } from '@/utils/telegram'
import LanguageControl from './LanguageControl.vue'
import { desktopNavigationItems, mobileNavigationItems } from './navigation'

const route = useRoute()
const router = useRouter()
const sessionStore = useSessionStore()
const { t } = useI18n()

const activePath = computed(() => route.path)
const mobileItems = computed(() => mobileNavigationItems(sessionStore.isAdmin))
const desktopItems = computed(() => desktopNavigationItems(t, sessionStore.isAdmin))
const showBackButton = computed(() => !['/', '/home'].includes(route.path))
const mainContent = useTemplateRef<globalThis.HTMLElement>('mainContent')
const isFullscreen = telegramFullscreenState()
const greetingName = computed(() => t('nav.fullscreenGreeting', {
  name: sessionStore.user?.firstName?.trim() || t('nav.memberFallback'),
}))
const greetingUsername = computed(() => sessionStore.user?.username?.trim() || sessionStore.user?.telegramUsername?.trim() || '')

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
  if (to === '/admin/settings') return activePath.value.startsWith('/admin')
  return activePath.value === to || (to === '/home' && activePath.value === '/')
}
</script>

<template>
  <div class="app-frame" :class="{ 'app-frame--fullscreen': isFullscreen }">
    <a class="skip-link" href="#main-content">{{ $t('nav.skip') }}</a>
    <UDashboardGroup class="app-dashboard" storage="local" storage-key="tx-carpool-shell" unit="rem">
      <UDashboardSidebar class="side-rail app-dashboard__sidebar" :default-size="15" :min-size="13" :max-size="20" :collapsed-size="4.5" resizable collapsible :toggle="false">
        <template #header="{ collapsed, collapse }">
          <header class="side-rail__header">
            <RouterLink class="side-rail__brand" to="/home">
              <span class="side-rail__brand-mark" aria-hidden="true"><UIcon name="i-ph-users-three" /></span>
              <span class="side-rail__brand-name">{{ $t('app.name') }}</span>
            </RouterLink>
            <UButton
              color="neutral"
              variant="ghost"
              :icon="collapsed ? 'i-ph-sidebar-simple' : 'i-ph-sidebar-simple-duotone'"
              :aria-label="$t(collapsed ? 'nav.expandSidebar' : 'nav.collapseSidebar')"
              @click="collapse(!collapsed)"
            />
          </header>
        </template>
        <template #default="{ collapsed }">
          <nav class="side-rail__nav" :aria-label="$t('nav.primary')">
            <UNavigationMenu :items="desktopItems" orientation="vertical" :collapsed="collapsed" color="primary" variant="pill" :tooltip="true" :popover="true" />
          </nav>
        </template>
        <template #footer>
          <footer class="side-rail__footer">
            <div class="side-rail__member">
              <UIcon name="i-ph-user-circle" />
              <div>
                <strong>{{ sessionStore.user?.username || sessionStore.user?.firstName || $t('nav.memberFallback') }}</strong>
                <span>{{ $t('nav.member') }}</span>
              </div>
            </div>
            <LanguageControl show-label />
          </footer>
        </template>
      </UDashboardSidebar>
      <div class="app-frame__content">
        <div v-if="isFullscreen" class="app-greeting" role="status">
          <strong>{{ greetingName }}</strong>
          <span v-if="greetingUsername">@{{ greetingUsername }}</span>
        </div>
        <main id="main-content" ref="mainContent" class="app-main" tabindex="-1">
          <slot />
        </main>
      </div>
    </UDashboardGroup>

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
        data-haptic="navigate"
        @click="goTo(item.to)"
      >
        <UIcon :name="item.icon" />
        <span>{{ $t(item.labelKey) }}</span>
      </UButton>
    </nav>
  </div>
</template>
