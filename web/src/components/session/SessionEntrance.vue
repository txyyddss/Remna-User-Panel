<script setup lang="ts">
import type { SessionStatus } from '@/stores/session'

import LoadingScreen from './LoadingScreen.vue'
import { useLoadingSequence } from './useLoadingSequence'

const props = defineProps<{ status: SessionStatus }>()
const { loading, showing } = useLoadingSequence(() => props.status)
</script>

<template>
  <div class="session-entrance">
    <Transition name="session-page" appear>
      <div v-if="!loading" v-show="!showing" class="session-entrance__page">
        <slot />
      </div>
    </Transition>
    <Transition name="session-cover">
      <LoadingScreen v-if="showing" />
    </Transition>
  </div>
</template>
