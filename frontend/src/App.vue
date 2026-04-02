<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const isTelegram = ref(false)

const navItems = computed(() => {
  const items = [
    { path: '/', label: 'Home' },
    { path: '/sub', label: 'Subscription' },
    { path: '/combos', label: 'Plans' },
    { path: '/credits', label: userStore.appConfig?.credit_name || 'Credits' },
    { path: '/jellyfin', label: 'Jellyfin' },
    { path: '/info', label: 'Usage' },
  ]

  if (userStore.isAdmin) {
    items.push({ path: '/admin', label: 'Admin' })
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
  tg.setBackgroundColor('#0b1220')
  tg.setHeaderColor('#0b1220')

  tg.BackButton.onClick(() => {
    if (route.path !== '/') {
      router.back()
    }
  })

  await userStore.refreshState()
  if (userStore.error?.includes('group membership required')) {
    router.replace({ path: '/blocked', query: { reason: 'group' } })
    return
  }

  userStore.startAutoRefresh()
}

onMounted(() => {
  void bootstrapTelegramApp()
})

watch(() => route.fullPath, () => {
  if (isTelegram.value && route.path !== '/blocked') {
    void userStore.refreshState({ background: true })
  }
})

onUnmounted(() => {
  userStore.stopAutoRefresh()
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
        <span>{{ item.label }}</span>
      </router-link>
    </nav>
  </div>

  <div v-else class="loading-page">
    <div class="loading-spinner"></div>
  </div>
</template>

<style scoped>
.app-shell {
  min-height: 100dvh;
}
</style>
