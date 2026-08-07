<script setup lang="ts">
import { CheckboxIndicator, CheckboxRoot } from 'reka-ui'
import { PhArrowRight, PhCheck, PhLinkBreak, PhShieldWarning } from '@phosphor-icons/vue'

const accepted = defineModel<boolean>({ required: true })

defineProps<{ loading: boolean }>()
defineEmits<{ submit: [] }>()
</script>

<template>
  <section class="onboarding-panel">
    <header class="onboarding-panel__header">
      <span class="feature-icon feature-icon--warning"><PhShieldWarning :size="25" weight="fill" /></span>
      <h1>Your link stays yours.</h1>
      <p>Sharing a subscription link exposes private access and can result in immediate account suspension.</p>
    </header>

    <div class="agreement-callout">
      <PhLinkBreak :size="23" aria-hidden="true" />
      <p>Never post, forward, sell, or share your subscription URL with another person.</p>
    </div>

    <label class="checkbox-row">
      <CheckboxRoot v-model="accepted" class="checkbox-control">
        <CheckboxIndicator class="checkbox-indicator">
          <PhCheck :size="16" weight="bold" />
        </CheckboxIndicator>
      </CheckboxRoot>
      <span>I understand and will keep my subscription link private.</span>
    </label>

    <button class="button button--primary button--wide" type="button" :disabled="!accepted || loading" @click="$emit('submit')">
      {{ loading ? 'Creating your account' : 'Finish setup' }}
      <PhArrowRight :size="19" />
    </button>
  </section>
</template>
