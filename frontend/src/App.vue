<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const isTelegram = ref(false)

const navItems = computed(() => {
  const items = [
    { path: '/', icon: '🏠', label: 'Home' },
    { path: '/sub', icon: '📡', label: 'Sub' },
    { path: '/credits', icon: '💎', label: userStore.appConfig?.credit_name || 'TXB' },
    { path: '/jellyfin', icon: '🎬', label: 'Video' },
    { path: '/info', icon: '📊', label: 'Info' },
  ]
  if (userStore.isAdmin) {
    items.push({ path: '/admin', icon: '⚙️', label: 'Admin' })
  }
  return items
})

onMounted(async () => {
  const tg = window.Telegram?.WebApp
  if (tg && tg.initData) {
    isTelegram.value = true
    tg.ready()
    tg.expand()
    tg.setBackgroundColor('#0a0a0f')
    tg.setHeaderColor('#0a0a0f')

    // Back button
    tg.BackButton.onClick(() => {
      if (route.path !== '/') {
        router.back()
      }
    })

    await userStore.refreshState()
    userStore.startAutoRefresh()
  } else {
    router.replace('/blocked')
  }
})

watch(() => route.fullPath, () => {
  if (isTelegram.value) {
    userStore.refreshState({ background: true })
  }
})

onUnmounted(() => {
  userStore.stopAutoRefresh()
})
</script>

<template>
  <div class="app" v-if="isTelegram || $route.path === '/blocked'">
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
        <span class="nav-icon">{{ item.icon }}</span>
        <span>{{ item.label }}</span>
      </router-link>
    </nav>
  </div>
  <div v-else class="loading-page">
    <div class="loading-spinner"></div>
  </div>
</template>

<style scoped>
.app {
  min-height: 100dvh;
}
</style>
