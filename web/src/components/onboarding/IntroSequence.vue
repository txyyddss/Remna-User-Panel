<script setup lang="ts">
import { onMounted } from 'vue'
import { PhArrowRight } from '@phosphor-icons/vue'

import { useIntroSequence } from '@/composables/useIntroSequence'
import type { OnboardingWelcomeMessage } from '@/api/features'

const emit = defineEmits<{ complete: [] }>()
const props = defineProps<{ messages: readonly OnboardingWelcomeMessage[] }>()
const { index, message, progress, start, skip } = useIntroSequence({
  messages: () => props.messages,
  onComplete: () => emit('complete'),
})

onMounted(start)
</script>

<template>
  <section class="intro-sequence">
    <button class="text-button intro-sequence__skip" type="button" @click="skip">
      {{ $t('common.skip') }}
      <PhArrowRight :size="17" />
    </button>
    <div class="intro-sequence__center">
      <Transition name="intro-copy" mode="out-in">
        <h1 :key="index">{{ message }}</h1>
      </Transition>
    </div>
    <div class="intro-sequence__progress" aria-hidden="true">
      <span :style="{ transform: `scaleX(${progress / 100})` }" />
    </div>
  </section>
</template>
