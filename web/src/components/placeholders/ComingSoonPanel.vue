<script setup lang="ts">
import type { Component } from 'vue'
import { computed } from 'vue'
import { PhArrowLeft, PhGameController, PhListChecks, PhMonitorPlay } from '@phosphor-icons/vue'
import { RouterLink } from 'vue-router'

const props = defineProps<{ product: string }>()

const content = computed<{ icon: Component; title: string; body: string }>(() => ({
  Games: {
    icon: PhGameController,
    title: 'Play is pulling in.',
    body: 'Small community games and member challenges will arrive here in a future release.',
  },
  Questionnaire: {
    icon: PhListChecks,
    title: 'Your voice gets a room.',
    body: 'Member questionnaires will make it easy to shape plans, regions, and community events.',
  },
  Emby: {
    icon: PhMonitorPlay,
    title: 'Movie night comes later.',
    body: 'Emby access and viewing controls are planned for a dedicated, privacy-minded experience.',
  },
}[props.product] ?? {
  icon: PhGameController,
  title: 'Something useful is coming.',
  body: 'This part of TX Carpool is still being prepared.',
}))
</script>

<template>
  <div class="page coming-soon-page">
    <section class="coming-soon-panel">
      <div class="coming-soon-panel__visual" aria-hidden="true">
        <span class="coming-soon-panel__orbit" />
        <span class="coming-soon-panel__icon"><component :is="content.icon" :size="42" weight="duotone" /></span>
      </div>
      <div class="coming-soon-panel__copy">
        <p class="eyebrow">{{ product }}</p>
        <h1>{{ content.title }}</h1>
        <p>{{ content.body }}</p>
        <RouterLink class="button button--secondary" to="/home">
          <PhArrowLeft :size="19" />
          Back home
        </RouterLink>
      </div>
    </section>
  </div>
</template>
