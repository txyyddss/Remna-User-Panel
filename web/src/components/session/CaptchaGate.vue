<script setup lang="ts">
import { onMounted, onUnmounted, shallowRef, useTemplateRef, watch } from 'vue'

import { api } from '@/api/client'
import type { Session } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import LanguageControl from '@/components/layout/LanguageControl.vue'
import { localizedError, useI18n } from '@/i18n'
import { loadTurnstile, type Turnstile } from '@/utils/turnstile'

const props = defineProps<{ challenge: NonNullable<Session['captcha']> }>()
const emit = defineEmits<{ verified: [session: Session] }>()
const { locale, t } = useI18n()
const container = useTemplateRef<globalThis.HTMLElement>('widget')
const loading = shallowRef(true)
const busy = shallowRef(false)
const error = shallowRef<string | null>(null)
let provider: Turnstile | null = null
let widgetId: string | null = null
let disposed = false
let generation = 0

async function verify(token: string): Promise<void> {
  if (busy.value || disposed) return
  busy.value = true
  error.value = null
  try {
    const session = await api.verifyCaptcha(token)
    if (session.captcha?.required) throw new Error(t('captcha.failed'))
    if (!disposed) emit('verified', session)
  } catch (caught) {
    if (!disposed) {
      error.value = localizedError(caught, 'captcha.failed')
      if (widgetId !== null) provider?.reset(widgetId)
    }
  } finally { busy.value = false }
}

async function render(): Promise<void> {
  const version = ++generation
  loading.value = true
  error.value = null
  if (widgetId !== null) provider?.remove(widgetId)
  widgetId = null
  try {
    if (!props.challenge.siteKey) throw new Error(t('captcha.unavailable'))
    provider = await loadTurnstile()
    if (disposed || version !== generation || !container.value) return
    widgetId = provider.render(container.value, {
      sitekey: props.challenge.siteKey, action: props.challenge.action, theme: 'dark', size: 'flexible',
      language: locale.value === 'zh-CN' ? 'zh-cn' : 'en',
      callback: token => void verify(token),
      'error-callback': () => { error.value = t('captcha.unavailable') },
      'expired-callback': () => { error.value = t('captcha.expired'); if (widgetId !== null) provider?.reset(widgetId) },
    })
  } catch { if (!disposed) error.value = t('captcha.unavailable') }
  finally { if (!disposed && version === generation) loading.value = false }
}

onMounted(() => void render())
watch(locale, () => { if (!busy.value) void render() })
onUnmounted(() => { disposed = true; generation++; if (widgetId !== null) provider?.remove(widgetId) })
</script>

<template>
  <main class="auth-screen captcha-screen" :aria-busy="loading || busy">
    <div class="auth-screen__copy">
      <p class="eyebrow">{{ t('captcha.eyebrow') }}</p>
      <h1>{{ t('captcha.title') }}</h1>
      <p>{{ t('captcha.description') }}</p>
    </div>
    <p v-if="loading || busy" role="status">{{ t(busy ? 'captcha.verifying' : 'captcha.loading') }}</p>
    <div ref="widget" class="captcha-screen__widget" />
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <UButton v-if="error" :disabled="busy" icon="i-ph-arrow-clockwise" :label="t('captcha.retry')" @click="render" />
    <footer class="auth-screen__locale"><LanguageControl /></footer>
  </main>
</template>

<style scoped>
.captcha-screen { max-width: 28rem; margin-inline: auto; }
.captcha-screen__widget { width: 100%; min-width: 0; min-height: 65px; }
.captcha-screen > p { color: var(--text-muted); font-size: 0.85rem; }
</style>
