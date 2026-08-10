<script setup lang="ts">
import { onMounted } from 'vue'

import type { OnboardingWelcomeMessage } from '@/api/features'
import { useIntroSequence } from '@/composables/useIntroSequence'

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
    <UButton
      class="intro-sequence__skip"
      color="neutral"
      variant="ghost"
      trailing-icon="i-ph-arrow-right"
      :label="$t('common.skip')"
      @click="skip"
    />
    <div class="intro-sequence__center">
      <Transition name="intro-copy" mode="out-in">
        <h1 :key="index">{{ message }}</h1>
      </Transition>
    </div>
    <UProgress class="intro-sequence__progress" :model-value="progress" :max="100" aria-hidden="true" />
  </section>
</template>
