<script setup lang="ts">
import { useRouter } from 'vue-router'

const items = [
  { to: '/questionnaire', labelKey: 'nav.questionnaire', noteKey: 'dashboard.questionnaireNote', icon: 'i-ph-list-checks' },
  { to: '/emby', labelKey: 'nav.emby', noteKey: 'dashboard.embyNote', icon: 'i-ph-monitor-play' },
]

const router = useRouter()

function goTo(to: string): void {
  void router.push(to).catch(() => undefined)
}
</script>

<template>
  <section class="section-block home-around">
    <div class="section-heading">
      <h2>{{ $t('dashboard.aroundTx') }}</h2>
      <span class="section-heading__meta">{{ $t('dashboard.memberTools') }}</span>
    </div>
    <div class="home-around__links">
      <UButton
        v-for="item in items"
        :key="item.to"
        type="button"
        color="neutral"
        variant="ghost"
        class="home-around__link"
        @click="goTo(item.to)"
      >
        <span class="feature-icon feature-icon--small"><UIcon :name="item.icon" /></span>
        <span><strong>{{ $t(item.labelKey) }}</strong><small>{{ $t(item.noteKey) }}</small></span>
        <UIcon name="i-ph-arrow-right" />
      </UButton>
    </div>
  </section>
</template>
