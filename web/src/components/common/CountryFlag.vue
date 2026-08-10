<script setup lang="ts">
import cn from '@/assets/flags/cn.svg'
import de from '@/assets/flags/de.svg'
import gb from '@/assets/flags/gb.svg'
import hk from '@/assets/flags/hk.svg'
import jp from '@/assets/flags/jp.svg'
import sg from '@/assets/flags/sg.svg'
import us from '@/assets/flags/us.svg'

const props = defineProps<{ code: string }>()
const flags: Record<string, string> = { CN: cn, DE: de, GB: gb, HK: hk, JP: jp, SG: sg, US: us }
</script>

<template>
  <span class="country-flag" :title="code.toUpperCase()">
    <img v-if="flags[code.toUpperCase()]" :src="flags[code.toUpperCase()]" alt="" />
    <span v-else>{{ code.toUpperCase() || '—' }}</span>
  </span>
</template>

<style scoped>
.country-flag { width: 28px; height: 20px; display: inline-grid; place-items: center; overflow: hidden; border: 1px solid var(--line); border-radius: 5px; color: var(--text-muted); background: var(--surface); font-family: var(--font-mono); font-size: 0.58rem; }
.country-flag img { width: 100%; height: 100%; object-fit: cover; }
</style>
