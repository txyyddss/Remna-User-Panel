<script setup lang="ts">
import { computed } from 'vue'
import type { SubscriptionHost, SubscriptionSquad } from '@/api/subscription'
import { useI18n } from '@/i18n'
import UptimeBar from './UptimeBar.vue'

const props = defineProps<{ squad: Readonly<SubscriptionSquad>; hosts: readonly Readonly<SubscriptionHost>[] }>()
const { t } = useI18n()
const visibleHosts = computed(() => props.hosts.filter(host => props.squad.hostUuids.includes(host.uuid)))
</script>

<template>
  <section class="squad-uptime">
    <UptimeBar :timeline="squad.timeline" :label="squad.name" />
    <details v-if="visibleHosts.length" class="squad-uptime__details">
      <summary>{{ t('subscription.hostDetails', { count: visibleHosts.length }) }}</summary>
      <div class="squad-uptime__hosts"><UptimeBar v-for="host in visibleHosts" :key="host.uuid" :timeline="host.timeline" :label="host.name || host.uuid" host /></div>
    </details>
    <p v-else>{{ t('subscription.noSquadHosts') }}</p>
  </section>
</template>

<style scoped>
.squad-uptime { display: grid; gap: 0.7rem; padding: 1rem 0; border-bottom: 1px solid var(--line); }
.squad-uptime:last-child { border-bottom: 0; }
.squad-uptime__details summary { min-height: 44px; padding: 0.65rem 0; color: var(--text-muted); font-size: 0.8rem; cursor: pointer; }
.squad-uptime__details summary:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
.squad-uptime__hosts { display: grid; gap: 1.15rem; padding: 0.5rem 0 0.5rem 0.65rem; }
.squad-uptime p { margin: 0; color: var(--text-muted); font-size: 0.8rem; }
</style>
