<script setup lang="ts">
import type { ConnectionNode } from '@/api/types'
import CountryFlag from '@/components/common/CountryFlag.vue'
import { formatDateTime } from '@/utils/format'
import type { ConnectionTarget } from './types'

const props = withDefaults(defineProps<{
  nodes: readonly ConnectionNode[]
  disabled?: boolean
}>(), { disabled: false })

const emit = defineEmits<{ drop: [target: ConnectionTarget] }>()

function select(node: ConnectionNode, index: number): void {
  const connection = node.ips[index]
  if (!connection) return
  emit('drop', { nodeName: node.name, countryCode: node.countryCode, connection })
}
</script>

<template>
  <div class="connection-nodes">
    <article v-for="node in props.nodes" :key="node.uuid" class="connection-node">
      <header class="connection-node__header">
        <CountryFlag :code="node.countryCode" />
        <div>
          <h2>{{ node.name }}</h2>
          <p>{{ $t('connections.connectionCount', { count: node.ips.length }) }}</p>
        </div>
      </header>
      <div class="connection-node__ips">
        <div v-for="(connection, index) in node.ips" :key="connection.handle" class="connection-ip">
          <div class="connection-ip__identity">
            <UIcon name="i-ph-device-mobile" aria-hidden="true" />
            <div>
              <code>{{ connection.ip }}</code>
              <span>{{ $t('connections.lastSeen', { date: formatDateTime(connection.lastSeen) }) }}</span>
            </div>
          </div>
          <UButton
            type="button"
            color="error"
            variant="ghost"
            square
            icon="i-ph-plugs-connected"
            :disabled="disabled"
            :aria-label="$t('connections.disconnectIp', { ip: connection.ip })"
            data-haptic="heavy"
            @click="select(node, index)"
          />
        </div>
      </div>
    </article>
  </div>
</template>
