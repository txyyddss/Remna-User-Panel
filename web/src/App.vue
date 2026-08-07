<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'

import AppShell from '@/components/layout/AppShell.vue'
import AuthGate from '@/components/session/AuthGate.vue'
import LoadingScreen from '@/components/session/LoadingScreen.vue'
import { useSessionStore } from '@/stores/session'

const route = useRoute()
const sessionStore = useSessionStore()
const immersive = computed(() => route.meta.immersive === true)
</script>

<template>
  <LoadingScreen v-if="sessionStore.status === 'idle' || sessionStore.status === 'loading'" />
  <AuthGate
    v-else-if="sessionStore.status === 'error'"
    :message="sessionStore.error ?? 'Authentication failed.'"
    @retry="sessionStore.bootstrap(true)"
  />
  <template v-else>
    <RouterView v-if="immersive" v-slot="{ Component, route: currentRoute }">
      <Transition name="route" mode="out-in">
        <component :is="Component" :key="currentRoute.fullPath" />
      </Transition>
    </RouterView>
    <AppShell v-else>
      <RouterView v-slot="{ Component, route: currentRoute }">
        <Transition name="route" mode="out-in">
          <component :is="Component" :key="currentRoute.fullPath" />
        </Transition>
      </RouterView>
    </AppShell>
  </template>
</template>
