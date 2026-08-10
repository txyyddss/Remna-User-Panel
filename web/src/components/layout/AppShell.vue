<script setup lang="ts">
import { computed, nextTick, useTemplateRef, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import LanguageSwitcher from '@/components/common/LanguageSwitcher.vue'
import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { useI18n } from '@/i18n'
import { useSessionStore } from '@/stores/session'
import { initials } from '@/utils/format'

interface NavItem {
  labelKey: string
  to: string
  icon: string
}

const route = useRoute()
const router = useRouter()
const sessionStore = useSessionStore()
const { t } = useI18n()

const primaryItems: NavItem[] = [
  { labelKey: 'nav.home', to: '/home', icon: 'i-ph-house' },
  { labelKey: 'nav.explore', to: '/catalog', icon: 'i-ph-compass' },
  { labelKey: 'nav.balance', to: '/balance', icon: 'i-ph-wallet' },
]

const extraItems: NavItem[] = [
  { labelKey: 'nav.activity', to: '/activity', icon: 'i-ph-game-controller' },
  { labelKey: 'nav.questionnaire', to: '/questionnaire', icon: 'i-ph-list-checks' },
  { labelKey: 'nav.emby', to: '/emby', icon: 'i-ph-monitor-play' },
]

const displayName = computed(() => [sessionStore.user?.firstName, sessionStore.user?.lastName].filter(Boolean).join(' ') || t('nav.memberFallback'))
const userInitials = computed(() => initials(displayName.value))
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
    <a class="skip-link" href="#main-content">{{ $t('nav.skip') }}</a>
    <aside class="side-rail" :aria-label="$t('nav.primary')">
      <RouterLink class="side-rail__brand" to="/home" :aria-label="$t('nav.homeLabel')">
        <span class="brand-mark"><UIcon name="i-ph-circles-four-fill" /></span>
        <span>{{ $t('app.name') }}</span>
      </RouterLink>

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
        <USeparator class="side-rail__divider" />
        <RouterLink
          v-for="item in extraItems"
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

      <div class="side-rail__profile">
        <UAvatar :text="userInitials" />
        <span class="side-rail__profile-copy">
          <strong>{{ displayName }}</strong>
          <small>{{ sessionStore.user?.username ? `@${sessionStore.user.username}` : $t('nav.member') }}</small>
        </span>
      </div>
    </aside>

    <div class="app-frame__content">
      <header class="mobile-header">
        <RouterLink class="mobile-header__brand" to="/home" :aria-label="$t('nav.homeLabel')">
          <span class="brand-mark"><UIcon name="i-ph-circles-four-fill" /></span>
          <strong>{{ $t('app.name') }}</strong>
        </RouterLink>
        <LanguageSwitcher />
        <UAvatar :text="userInitials" size="sm" />
      </header>
      <main id="main-content" ref="mainContent" class="app-main" tabindex="-1">
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
    </nav>
  </div>
</template>
