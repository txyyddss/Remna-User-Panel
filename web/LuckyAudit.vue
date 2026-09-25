<script setup lang="ts">
import { en } from '@nuxt/ui/locale'
import ActivityPage from '@/components/activity/ActivityPage.vue'
</script>
<template>
  <UApp :locale="en"><main id="main-content" style="max-width: 70rem; margin: 0 auto; padding: 1rem;"><ActivityPage /></main></UApp>
</template>
