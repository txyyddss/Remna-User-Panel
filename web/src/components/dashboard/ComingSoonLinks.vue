<script setup lang="ts">
import { PhArrowRight, PhGameController, PhListChecks, PhMonitorPlay } from '@phosphor-icons/vue'
import { RouterLink } from 'vue-router'

const items = [
  { to: '/games', label: 'Games', note: 'Community play', icon: PhGameController },
  { to: '/questionnaire', label: 'Questionnaire', note: 'Member feedback', icon: PhListChecks },
  { to: '/emby', label: 'Emby', note: 'Shared media', icon: PhMonitorPlay },
]
</script>

<template>
  <section class="section-block more-section">
    <div class="section-heading">
      <h2>Around TX</h2>
      <span class="section-heading__meta">In progress</span>
    </div>
    <div class="more-links">
      <RouterLink v-for="item in items" :key="item.to" :to="item.to" class="more-link">
        <span class="feature-icon feature-icon--small"><component :is="item.icon" :size="20" /></span>
        <span><strong>{{ item.label }}</strong><small>{{ item.note }}</small></span>
        <PhArrowRight :size="18" />
      </RouterLink>
    </div>
  </section>
</template>
