<script setup lang="ts">
import type { Component } from 'vue'
import { computed, nextTick, useTemplateRef, watch } from 'vue'
import {
  PhCirclesFour,
  PhCompass,
  PhGameController,
  PhHouse,
  PhListChecks,
  PhMonitorPlay,
  PhShieldCheck,
  PhWallet,
} from '@phosphor-icons/vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { useSessionStore } from '@/stores/session'
import { initials } from '@/utils/format'

interface NavItem {
  label: string
  to: string
  icon: Component
}

const route = useRoute()
const router = useRouter()
const sessionStore = useSessionStore()

const primaryItems: NavItem[] = [
  { label: 'Home', to: '/home', icon: PhHouse },
  { label: 'Explore', to: '/catalog', icon: PhCompass },
  { label: 'Balance', to: '/balance', icon: PhWallet },
]

const extraItems: NavItem[] = [
  { label: 'Activity', to: '/activity', icon: PhGameController },
  { label: 'Questionnaire', to: '/questionnaire', icon: PhListChecks },
  { label: 'Emby', to: '/emby', icon: PhMonitorPlay },
]

const displayName = computed(() => [sessionStore.user?.firstName, sessionStore.user?.lastName].filter(Boolean).join(' ') || 'TX Member')
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
    <a class="skip-link" href="#main-content">Skip to content</a>
    <aside class="side-rail" aria-label="Primary navigation">
      <RouterLink class="side-rail__brand" to="/home" aria-label="TX Carpool home">
        <span class="brand-mark"><PhCirclesFour :size="21" weight="fill" /></span>
        <span>TX Carpool</span>
      </RouterLink>

      <nav class="side-rail__nav">
        <RouterLink
          v-for="item in primaryItems"
          :key="item.to"
          :to="item.to"
          class="nav-item"
          :class="{ 'nav-item--active': isActive(item.to) }"
        >
          <component :is="item.icon" :size="20" />
          <span>{{ item.label }}</span>
        </RouterLink>
        <div class="side-rail__divider" />
        <RouterLink
          v-for="item in extraItems"
          :key="item.to"
          :to="item.to"
          class="nav-item"
          :class="{ 'nav-item--active': isActive(item.to) }"
        >
          <component :is="item.icon" :size="20" />
          <span>{{ item.label }}</span>
        </RouterLink>
        <RouterLink
          v-if="sessionStore.isAdmin"
          to="/admin/settings"
          class="nav-item"
          :class="{ 'nav-item--active': activePath.startsWith('/admin') }"
        >
          <PhShieldCheck :size="20" />
          <span>Admin</span>
        </RouterLink>
      </nav>

      <div class="side-rail__profile">
        <span class="avatar">{{ userInitials }}</span>
        <span class="side-rail__profile-copy">
          <strong>{{ displayName }}</strong>
          <small>{{ sessionStore.user?.username ? `@${sessionStore.user.username}` : 'Telegram member' }}</small>
        </span>
      </div>
    </aside>

    <div class="app-frame__content">
      <header class="mobile-header">
        <RouterLink class="mobile-header__brand" to="/home" aria-label="TX Carpool home">
          <span class="brand-mark"><PhCirclesFour :size="19" weight="fill" /></span>
          <strong>TX Carpool</strong>
        </RouterLink>
        <span class="avatar avatar--small">{{ userInitials }}</span>
      </header>
      <main id="main-content" ref="mainContent" class="app-main" tabindex="-1">
        <slot />
      </main>
    </div>

    <nav class="bottom-nav" aria-label="Primary navigation">
      <RouterLink
        v-for="item in primaryItems"
        :key="item.to"
        :to="item.to"
        class="bottom-nav__item"
        :class="{ 'bottom-nav__item--active': isActive(item.to) }"
      >
        <component :is="item.icon" :size="21" />
        <span>{{ item.label }}</span>
      </RouterLink>
    </nav>
  </div>
</template>
