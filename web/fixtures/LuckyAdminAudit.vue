<script setup lang="ts">
import { en } from '@nuxt/ui/locale'
import AdminActivityPanel from '@/components/admin/AdminActivityPanel.vue'
</script>
<template>
  <UApp :locale="en"><main id="main-content" style="max-width: 74rem; margin: 0 auto; padding: 1rem;"><AdminActivityPanel /></main></UApp>
</template>
