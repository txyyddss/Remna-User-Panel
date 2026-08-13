<script setup lang="ts">
import { computed } from 'vue'

import type { SquadProfile } from '@/api/types'
import MarkdownContent from '@/components/common/MarkdownContent.vue'
import { useI18n } from '@/i18n'
import { countryName, profileTypeMeta } from './profile'

const props = withDefaults(defineProps<{
  profile: SquadProfile | null
  description?: string
  compact?: boolean
}>(), { description: '', compact: false })

const { locale, t } = useI18n()
const typeMeta = computed(() => props.profile ? profileTypeMeta(props.profile.type) : null)
const portText = computed(() => {
  if (!props.profile || !('portMbps' in props.profile)) return ''
  return props.profile.portMbps === null ? t('squadProfile.unlimited') : `${props.profile.portMbps} ${t('squadProfile.mbps')}`
})
const countryText = computed(() => {
  if (!props.profile || !('countryCode' in props.profile)) return ''
  return countryName(props.profile.countryCode, locale.value)
})
</script>

<template>
  <div class="squad-profile-summary" :class="{ 'squad-profile-summary--compact': compact }">
    <div class="squad-profile-summary__heading">
      <UIcon :name="typeMeta?.icon ?? 'i-ph-info'" aria-hidden="true" />
      <strong>{{ typeMeta ? t(typeMeta.labelKey) : t('squadProfile.unconfigured') }}</strong>
    </div>
    <div v-if="profile" class="squad-profile-summary__facts">
      <span v-if="profile.type === 'broadband'"><UIcon name="i-ph-broadcast" />{{ profile.isp }}</span>
      <span v-if="profile.type === 'broadband'"><UIcon name="i-ph-gauge" />{{ profile.portMbps }} {{ t('squadProfile.mbps') }}</span>
      <span v-if="profile.type === 'broadband'"><UIcon name="i-ph-arrows-clockwise" />{{ profile.dynamic ? t('squadProfile.dynamic') : t('squadProfile.static') }}</span>
      <span v-if="profile.type === 'broadband'"><UIcon name="i-ph-map-pin" />{{ profile.location }}</span>
      <template v-else-if="profile.type === 'china_optimized'">
        <span><UIcon name="i-ph-broadcast" />CT {{ profile.ct }}</span>
        <span><UIcon name="i-ph-broadcast" />CU {{ profile.cu }}</span>
        <span><UIcon name="i-ph-broadcast" />CM {{ profile.cm }}</span>
        <span><UIcon name="i-ph-gauge" />{{ portText }}</span>
        <span><UIcon name="i-ph-map-pin" />{{ countryText }}</span>
      </template>
      <template v-else>
        <span><UIcon name="i-ph-gauge" />{{ portText }}</span>
        <span><UIcon name="i-ph-map-pin" />{{ countryText }}</span>
        <span><UIcon name="i-ph-broadcast" />{{ profile.upstreamCarriers.join(', ') }}</span>
      </template>
    </div>
    <MarkdownContent v-if="description" :source="description" compact />
  </div>
</template>

<style scoped>
.squad-profile-summary { display: grid; gap: 0.55rem; min-width: 0; }
.squad-profile-summary__heading { display: flex; align-items: center; gap: 0.4rem; color: var(--text-muted); font-size: 0.72rem; }
.squad-profile-summary__facts { display: flex; flex-wrap: wrap; gap: 0.35rem 0.7rem; color: var(--text-faint); font-size: 0.68rem; line-height: 1.45; }
.squad-profile-summary__facts span { display: inline-flex; align-items: center; gap: 0.25rem; min-width: 0; }
.squad-profile-summary__facts :deep(svg) { flex: 0 0 auto; }
.squad-profile-summary--compact { gap: 0.35rem; }
</style>
