<script setup lang="ts">
import { computed, watch } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import { en, zh_cn } from '@nuxt/ui/locale'

import AppShell from '@/components/layout/AppShell.vue'
import AuthGate from '@/components/session/AuthGate.vue'
import SessionEntrance from '@/components/session/SessionEntrance.vue'
import AppErrorBoundary from '@/components/session/AppErrorBoundary.vue'
import { useSessionStore } from '@/stores/session'
import { useDisplayCurrency } from '@/composables/useDisplayCurrency'
import { useI18n } from '@/i18n'
import { isTelegramWebAppDetected } from '@/utils/telegram'

const route = useRoute()
const sessionStore = useSessionStore()
const displayCurrency = useDisplayCurrency()
const immersive = computed(() => route.meta.immersive === true)
const browserPublic = computed(() => route.meta.browserPublic === true && !isTelegramWebAppDetected())
const { locale } = useI18n()
const uiLocale = computed(() => locale.value === 'zh-CN' ? zh_cn : en)

watch(() => sessionStore.user?.id, (userID) => {
  if (!userID) return
  displayCurrency.hydrate(sessionStore.user?.displayCurrency)
  void displayCurrency.refresh()
}, { immediate: true })
</script>

<template>
  <UApp :locale="uiLocale" :toaster="{ position: 'top-center' }">
    <AppErrorBoundary>
      <RouterView v-if="browserPublic" v-slot="{ Component, route: currentRoute }">
        <component :is="Component" :key="currentRoute.fullPath" />
      </RouterView>
      <SessionEntrance v-else :status="sessionStore.status">
        <AuthGate
          v-if="sessionStore.status === 'error'"
          :message="sessionStore.error ?? $t('auth.authenticationFailed')"
          @retry="sessionStore.bootstrap(true)"
        />
        <template v-else>
          <RouterView v-if="immersive" v-slot="{ Component, route: currentRoute }">
            <component :is="Component" :key="currentRoute.fullPath" />
          </RouterView>
          <AppShell v-else>
            <RouterView v-slot="{ Component, route: currentRoute }">
              <component :is="Component" :key="currentRoute.fullPath" />
            </RouterView>
          </AppShell>
        </template>
      </SessionEntrance>
    </AppErrorBoundary>
  </UApp>
</template>
