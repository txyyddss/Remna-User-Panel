<script setup lang="ts">
import { PhArrowRight, PhAt } from '@phosphor-icons/vue'

const username = defineModel<string>({ required: true })

defineProps<{
  valid: boolean
  hint: string
  loading: boolean
}>()

defineEmits<{ submit: [] }>()
</script>

<template>
  <section class="onboarding-panel">
    <header class="onboarding-panel__header">
      <p class="eyebrow">Your handle</p>
      <h1>Choose a quiet identity.</h1>
      <p>This permanent username becomes your private Remnawave account name.</p>
    </header>

    <form class="form-stack" @submit.prevent="$emit('submit')">
      <label class="field-label" for="carpool-username">Username</label>
      <div class="input-shell" :class="{ 'input-shell--valid': valid }">
        <PhAt :size="20" aria-hidden="true" />
        <input
          id="carpool-username"
          v-model.trim="username"
          name="username"
          type="text"
          inputmode="text"
          autocomplete="off"
          autocapitalize="none"
          spellcheck="false"
          maxlength="9"
          placeholder="river"
          aria-describedby="username-hint"
        />
      </div>
      <p id="username-hint" class="field-hint" :class="{ 'field-hint--valid': valid }">{{ hint }}</p>

      <button class="button button--primary button--wide" type="submit" :disabled="!valid || loading">
        {{ loading ? 'Checking availability' : 'Reserve username' }}
        <PhArrowRight :size="19" />
      </button>
    </form>
  </section>
</template>
