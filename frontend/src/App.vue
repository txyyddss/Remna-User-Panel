<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import AppToast from '@/components/AppToast.vue'
import AppConfirm from '@/components/AppConfirm.vue'
import PaymentInstructionModal from '@/components/payments/PaymentInstructionModal.vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const isTelegram = ref(false)
let refreshDebounce: number | null = null

const navItems = computed(() => {
  const items = [
    { path: '/', label: 'Home', icon: 'home' },
    { path: '/sub', label: 'Sub', icon: 'sub' },
    { path: '/combos', label: 'Plans', icon: 'plans' },
    { path: '/credits', label: userStore.appConfig?.credit_name || 'Credits', icon: 'credits' },
    { path: '/jellyfin', label: 'Jellyfin', icon: 'jellyfin' },
    { path: '/info', label: 'Usage', icon: 'usage' },
  ]

  if (userStore.isAdmin) {
    items.push({ path: '/admin', label: 'Admin', icon: 'admin' })
  }

  return items
})

async function waitForTelegramInitData(timeoutMs = 2500): Promise<boolean> {
  const startedAt = Date.now()
  while (Date.now() - startedAt < timeoutMs) {
    const tg = window.Telegram?.WebApp
    if (tg?.initData) {
      return true
    }
    await new Promise((resolve) => window.setTimeout(resolve, 120))
  }
  return false
}

async function bootstrapTelegramApp() {
  const ready = await waitForTelegramInitData()
  if (!ready) {
    router.replace('/blocked')
    return
  }

  const tg = window.Telegram?.WebApp
  if (!tg) {
    router.replace('/blocked')
    return
  }

  isTelegram.value = true
  tg.ready()
  tg.expand()
  tg.setBackgroundColor('#07070d')
  tg.setHeaderColor('#07070d')

  // Disable vertical swipes if available (prevents accidental close on mobile)
  if (typeof (tg as any).disableVerticalSwipes === 'function') {
    ;(tg as any).disableVerticalSwipes()
  }

  // BackButton handler
  tg.BackButton.onClick(() => {
    if (route.path !== '/') {
      router.back()
    }
  })

  // Re-expand when returning from external links (mobile fix)
  document.addEventListener('visibilitychange', () => {
    if (!document.hidden && isTelegram.value) {
      tg.expand()
    }
  })

  try {
    const access = await userStore.bootstrapMiniAppAccess()
    if (!access.group_joined) {
      router.replace({ path: '/blocked', query: { reason: 'group' } })
      return
    }

    await userStore.refreshState()
    if (userStore.error?.includes('group membership required')) {
      router.replace({ path: '/blocked', query: { reason: 'group' } })
      return
    }

    userStore.startAutoRefresh()
  } catch {
    router.replace('/blocked')
  }
}

onMounted(() => {
  void bootstrapTelegramApp()
})

// Manage BackButton visibility based on route
watch(() => route.path, (path) => {
  if (!isTelegram.value) return
  const tg = window.Telegram?.WebApp
  if (!tg) return
  if (path === '/' || path === '/blocked') {
    tg.BackButton.hide()
  } else {
    tg.BackButton.show()
  }
}, { immediate: false })

// Debounced background refresh on route change
watch(() => route.fullPath, () => {
  if (isTelegram.value && route.path !== '/blocked') {
    if (refreshDebounce !== null) {
      window.clearTimeout(refreshDebounce)
    }
    refreshDebounce = window.setTimeout(() => {
      void userStore.refreshState({ background: true })
      refreshDebounce = null
    }, 350)
  }
})

onUnmounted(() => {
  userStore.stopAutoRefresh()
  if (refreshDebounce !== null) {
    window.clearTimeout(refreshDebounce)
  }
})
</script>

<template>
  <div class="app-shell" v-if="isTelegram || $route.path === '/blocked'">
    <router-view v-slot="{ Component }">
      <transition name="slide-up" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>

    <nav class="bottom-nav" v-if="isTelegram && $route.path !== '/blocked'">
      <router-link
        v-for="item in navItems"
        :key="item.path"
        :to="item.path"
        class="nav-item"
        :class="{ active: $route.path === item.path }"
      >
        <svg v-if="item.icon === 'home'" class="nav-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/></svg>
        <svg v-else-if="item.icon === 'sub'" class="nav-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
        <svg v-else-if="item.icon === 'plans'" class="nav-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>
        <svg v-else-if="item.icon === 'credits'" class="nav-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="16"/><line x1="8" y1="12" x2="16" y2="12"/></svg>
        <svg v-else-if="item.icon === 'jellyfin'" class="nav-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="20" rx="2.18" ry="2.18"/><line x1="7" y1="2" x2="7" y2="22"/><line x1="17" y1="2" x2="17" y2="22"/><line x1="2" y1="12" x2="22" y2="12"/><line x1="2" y1="7" x2="7" y2="7"/><line x1="2" y1="17" x2="7" y2="17"/><line x1="17" y1="7" x2="22" y2="7"/><line x1="17" y1="17" x2="22" y2="17"/></svg>
        <svg v-else-if="item.icon === 'usage'" class="nav-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="20" x2="18" y2="10"/><line x1="12" y1="20" x2="12" y2="4"/><line x1="6" y1="20" x2="6" y2="14"/></svg>
        <svg v-else-if="item.icon === 'admin'" class="nav-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>
        <span>{{ item.label }}</span>
      </router-link>
    </nav>

    <AppToast />
    <AppConfirm />
    <PaymentInstructionModal />
  </div>

  <div v-else class="loading-page">
    <div class="loading-spinner"></div>
  </div>
</template>

<style scoped>
.app-shell {
  min-height: 100dvh;
}

.nav-icon {
  width: 20px;
  height: 20px;
  opacity: 0.7;
  transition: opacity 0.2s;
}

.nav-item.active .nav-icon {
  opacity: 1;
}
</style>
