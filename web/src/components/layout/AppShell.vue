<script setup lang="ts">
import { computed, nextTick, useTemplateRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useCommunityAccess } from '@/composables/useCommunityAccess'
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
const { activeCombo: hasValidCombo, refresh: refreshCommunityAccess } = useCommunityAccess()

const activePath = computed(() => route.path)
const mobileItems = computed(() => mobileNavigationItems(sessionStore.isAdmin))
const desktopItems = computed(() => desktopNavigationItems(t, sessionStore.isAdmin, hasValidCombo.value))
const showBackButton = computed(() => !['/', '/home'].includes(route.path))
const appContent = useTemplateRef<globalThis.HTMLDivElement>('appContent')
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

watch(() => route.path, (_next, previous) => {
  if (!previous) return
  void refreshCommunityAccess()
  void nextTick()
    .then(() => {
      if (appContent.value) {
        appContent.value.scrollTop = 0
        appContent.value.scrollLeft = 0
      }
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
      <UDashboardSidebar class="side-rail app-dashboard__sidebar" :default-size="15" :min-size="13" :max-size="20" resizable>
        <template #default>
          <nav class="side-rail__nav" :aria-label="$t('nav.primary')">
            <UNavigationMenu :items="desktopItems" orientation="vertical" color="primary" variant="pill" />
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
      <div ref="appContent" class="app-frame__content">
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
