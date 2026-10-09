<script setup lang="ts">
import { computed, onScopeDispose, shallowRef, watch } from 'vue'
import type { SubscriptionHost } from '@/api/subscription'
import CountryFlag from '@/components/common/CountryFlag.vue'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useClipboard } from '@/composables/useClipboard'
import { useI18n } from '@/i18n'
import { createLatestRequest } from '@/utils/latestRequest'

const props = defineProps<{ host: Readonly<SubscriptionHost> }>()
const { t } = useI18n()
const { copy, copied } = useClipboard()
const open = shallowRef(false)
const qr = shallowRef<string | null>(null)
const qrError = shallowRef(false)
const copyError = shallowRef(false)
const latest = createLatestRequest()
const flags = computed(() => props.host.countryCodes.length ? props.host.countryCodes : [''])
const linkError = computed(() => t(props.host.linkErrorCode === 'CONNECTIVITY_LINK_AMBIGUOUS' ? 'subscription.ambiguousLink' : 'subscription.noLink'))
watch(() => [open.value, props.host.link] as const, async ([show, link]) => {
  const token = latest.begin()
  qr.value = null
  qrError.value = false
  if (!show || !link) { open.value = false; return }
  try {
    const { default: QRCode } = await import('qrcode')
    const value = await QRCode.toDataURL(link, { width: 280, margin: 3, errorCorrectionLevel: 'M', color: { dark: '#111111', light: '#ffffff' } })
    if (latest.isCurrent(token)) qr.value = value
  } catch { if (latest.isCurrent(token)) qrError.value = true }
})
onScopeDispose(() => latest.dispose())
async function copyLink(): Promise<void> { if (props.host.link) copyError.value = !await copy(props.host.link) }
</script>

<template>
  <li class="host-link">
    <div class="host-link__row">
      <span class="host-link__flags"><CountryFlag v-for="country in flags" :key="country" :code="country" /></span>
      <strong class="host-link__name">{{ host.name || host.uuid }}</strong>
      <UButton color="neutral" variant="ghost" :icon="copied ? 'i-ph-check-bold' : 'i-ph-copy'" :disabled="!host.link" :aria-label="t(copied ? 'common.copied' : 'subscription.copyHost', { name: host.name })" data-haptic="copy" @click="copyLink" />
      <UButton color="neutral" variant="ghost" icon="i-ph-qr-code" :disabled="!host.link" :aria-label="t('subscription.qrHost', { name: host.name })" data-haptic="open" @click="open = true" />
    </div>
    <small v-if="!host.link">{{ linkError }}</small>
    <InlineNotice v-if="copyError" tone="warning">{{ t('subscription.copyFailed') }}</InlineNotice>
    <UModal v-model:open="open" :title="host.name || host.uuid" :description="t('subscription.qrHint')" close-icon="i-ph-x-bold">
      <template #body>
        <div class="host-link__qr">
          <img v-if="qr" :src="qr" :alt="t('subscription.qrHost', { name: host.name })" width="280" height="280" />
          <InlineNotice v-else-if="qrError" tone="warning">{{ t('subscription.qrFailed') }}</InlineNotice>
          <USkeleton v-else class="host-link__qr-loading" />
        </div>
      </template>
    </UModal>
  </li>
</template>

<style scoped>
.host-link { display: grid; gap: 0.4rem; padding: 0.65rem 0; border-bottom: 1px solid var(--line); }
.host-link:last-child { border-bottom: 0; }
.host-link__row { display: flex; align-items: center; gap: 0.35rem; min-width: 0; }
.host-link__flags { display: flex; flex-wrap: wrap; gap: 0.15rem; max-width: 80px; flex-shrink: 0; }
.host-link__name { flex: 1; min-width: 0; overflow-wrap: anywhere; font-size: 0.85rem; font-weight: 600; }
.host-link__row :deep(button) { min-width: 44px; min-height: 44px; flex-shrink: 0; }
.host-link small { color: var(--text-muted); font-size: 0.75rem; }
.host-link__qr { display: grid; place-items: center; min-height: 280px; }
.host-link__qr img { max-width: 100%; height: auto; background: white; border-radius: 6px; }
.host-link__qr-loading { width: 280px; max-width: 100%; height: 280px; }
</style>
