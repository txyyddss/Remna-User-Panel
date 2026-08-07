<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import ComingSoonPanel from '@/components/placeholders/ComingSoonPanel.vue'

const route = useRoute()
const product = computed(() => String(route.meta.product ?? 'Coming soon'))
</script>

<template>
  <ComingSoonPanel :product="product" />
</template>
