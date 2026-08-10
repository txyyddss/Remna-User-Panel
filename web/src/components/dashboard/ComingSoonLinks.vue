<script setup lang="ts">
import { RouterLink } from 'vue-router'

const items = [
  { to: '/activity', labelKey: 'nav.activity', noteKey: 'dashboard.activityNote', icon: 'i-ph-game-controller' },
  { to: '/questionnaire', labelKey: 'nav.questionnaire', noteKey: 'dashboard.questionnaireNote', icon: 'i-ph-list-checks' },
  { to: '/emby', labelKey: 'nav.emby', noteKey: 'dashboard.embyNote', icon: 'i-ph-monitor-play' },
]
</script>

<template>
  <section class="section-block more-section">
    <div class="section-heading">
      <h2>{{ $t('dashboard.aroundTx') }}</h2>
      <span class="section-heading__meta">{{ $t('dashboard.memberTools') }}</span>
    </div>
    <div class="more-links">
      <RouterLink v-for="item in items" :key="item.to" :to="item.to" class="more-link">
        <span class="feature-icon feature-icon--small"><UIcon :name="item.icon" /></span>
        <span><strong>{{ $t(item.labelKey) }}</strong><small>{{ $t(item.noteKey) }}</small></span>
        <UIcon name="i-ph-arrow-right" />
      </RouterLink>
    </div>
  </section>
</template>
