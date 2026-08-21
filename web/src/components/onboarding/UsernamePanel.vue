<script setup lang="ts">
const username = defineModel<string>({ required: true })

defineProps<{
  valid: boolean
  hint: string
  loading: boolean
  showAction?: boolean
}>()

defineEmits<{ submit: [] }>()
</script>

<template>
  <section class="onboarding-panel">
    <header class="onboarding-panel__header">
      <p class="eyebrow">{{ $t('onboarding.usernameEyebrow') }}</p>
      <h1>{{ $t('onboarding.usernameTitle') }}</h1>
      <p>{{ $t('onboarding.usernameCopy') }}</p>
    </header>

    <form class="form-stack" @submit.prevent="$emit('submit')">
      <UFormField name="username" :label="$t('onboarding.username')" :description="hint">
        <UInput
          id="carpool-username"
          v-model.trim="username"
          icon="i-ph-at"
          name="username"
          type="text"
          inputmode="text"
          autocomplete="off"
          autocapitalize="none"
          :spellcheck="false"
          :maxlength="9"
          :placeholder="$t('onboarding.usernamePlaceholder')"
        />
      </UFormField>

      <UButton
        v-if="showAction !== false"
        block
        type="submit"
        trailing-icon="i-ph-arrow-right"
        :disabled="!valid || loading"
        :loading="loading"
        :label="loading ? $t('onboarding.checkingAvailability') : $t('onboarding.reserveUsername')"
        data-haptic="confirm"
      />
    </form>
  </section>
</template>
