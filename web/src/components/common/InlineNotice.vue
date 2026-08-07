<script setup lang="ts">
import { computed } from 'vue'
import { PhCheckCircle, PhInfo, PhWarningCircle } from '@phosphor-icons/vue'

const props = withDefaults(defineProps<{
  tone?: 'info' | 'warning' | 'success'
  title?: string
}>(), {
  tone: 'info',
  title: undefined,
})

const icon = computed(() => ({
  info: PhInfo,
  warning: PhWarningCircle,
  success: PhCheckCircle,
})[props.tone])
</script>

<template>
  <div class="notice" :class="`notice--${tone}`" role="status">
    <component :is="icon" :size="20" weight="fill" aria-hidden="true" />
    <div>
      <strong v-if="title">{{ title }}</strong>
      <p><slot /></p>
    </div>
  </div>
</template>
