<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import { en, zh_cn } from '@nuxt/ui/locale'

import AppShell from '@/components/layout/AppShell.vue'
import AuthGate from '@/components/session/AuthGate.vue'
import LoadingScreen from '@/components/session/LoadingScreen.vue'
import { useSessionStore } from '@/stores/session'
import { useI18n } from '@/i18n'

const route = useRoute()
const sessionStore = useSessionStore()
const immersive = computed(() => route.meta.immersive === true)
const { locale } = useI18n()
const uiLocale = computed(() => locale.value === 'zh-CN' ? zh_cn : en)
</script>

<template>
  <UApp :locale="uiLocale" :toaster="{ position: 'top-center' }">
    <LoadingScreen v-if="sessionStore.status === 'idle' || sessionStore.status === 'loading'" />
    <AuthGate
      v-else-if="sessionStore.status === 'error'"
      :message="sessionStore.error ?? $t('auth.authenticationFailed')"
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
  </UApp>
</template>
