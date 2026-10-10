<script setup lang="ts">
import type { IPLookupAdminSettings } from '@/api/ipLookup'
import SwitchField from '@/components/common/SwitchField.vue'
import { providerName } from '@/components/ip-lookup/presentation'

const settings = defineModel<IPLookupAdminSettings>({ required: true })

function clearChanged(id: string, value: boolean | 'indeterminate'): void {
  if (value === true && settings.value.credentials?.[id]) settings.value.credentials[id].value = ''
}
</script>

<template>
  <section class="lookup-providers">
    <h2>{{ $t('ipLookup.providerHeading') }}</h2>
    <div v-for="provider in settings.providers" :key="provider.id" class="provider-settings">
      <SwitchField :id="`provider-${provider.id}`" v-model="provider.enabled" :label="providerName(provider.id)" :help="settings.configured?.[provider.id] ? $t('ipLookup.credentialConfigured') : $t('ipLookup.credentialMissing')" />
      <UFormField v-if="provider.id === 'maxmind' || provider.id === 'scamalytics'" :name="`${provider.id}-account`" :label="provider.id === 'scamalytics' ? $t('ipLookup.scamalyticsUsername') : $t('ipLookup.accountId')"><UInput v-model.trim="provider.accountId" class="w-full" :inputmode="provider.id === 'maxmind' ? 'numeric' : 'text'" /></UFormField>
      <template v-if="settings.credentials?.[provider.id]">
        <UFormField :name="`${provider.id}-credential`" :label="provider.id === 'maxmind' ? $t('ipLookup.licenseKey') : $t('ipLookup.apiKey')" :hint="$t('ipLookup.credentialHint')">
          <UInput v-model="settings.credentials[provider.id]!.value" class="w-full" type="password" autocomplete="new-password" :disabled="settings.credentials[provider.id]!.clear" />
        </UFormField>
        <UCheckbox v-if="settings.configured?.[provider.id]" v-model="settings.credentials[provider.id]!.clear" :label="$t('ipLookup.clearCredential')" @update:model-value="clearChanged(provider.id, $event)" />
      </template>
    </div>
  </section>
</template>

<style scoped>
.lookup-providers { display: grid; gap: 0.9rem; }
.lookup-providers h2 { font-size: 1rem; margin: 0; }
.lookup-providers p { color: var(--text-muted); font-size: 0.82rem; margin: 0; }
.provider-settings { display: grid; gap: 0.7rem; padding-block: 1rem; border-top: 1px solid var(--line); }
</style>
