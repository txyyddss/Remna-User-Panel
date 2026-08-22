<script setup lang="ts">
import { computed } from 'vue'

import type { SquadProfile } from '@/api/types'
import MarkdownContent from '@/components/common/MarkdownContent.vue'
import { useI18n } from '@/i18n'
import { countryName, profileTypeMeta } from './profile'
import CarrierLogo from './CarrierLogo.vue'

const props = withDefaults(defineProps<{
  name?: string
  profile: SquadProfile | null
  description?: string
  compact?: boolean
  presentation?: 'default' | 'member'
}>(), { name: '', description: '', compact: false, presentation: 'default' })

defineSlots<{ facts?: () => unknown, headingMeta?: () => unknown }>()

const { locale, t } = useI18n()
const typeMeta = computed(() => props.profile ? profileTypeMeta(props.profile.type) : null)
const typeClass = computed(() => props.profile ? `squad-profile-summary--${props.profile.type}` : '')
const portText = computed(() => {
  if (!props.profile || !('portMbps' in props.profile)) return ''
  return props.profile.portMbps === null ? t('squadProfile.unlimited') : `${props.profile.portMbps} ${t('squadProfile.mbps')}`
})
const countryText = computed(() => {
  if (!props.profile || !('countryCode' in props.profile)) return ''
  return countryName(props.profile.countryCode, locale.value)
})
const memberProfileText = computed(() => {
  const label = typeMeta.value ? t(typeMeta.value.labelKey) : t('squadProfile.unconfigured')
  return portText.value ? `${label} ${portText.value}` : label
})
</script>

<template>
  <div
    class="squad-profile-summary"
    :class="[typeClass, { 'squad-profile-summary--compact': compact, 'squad-profile-summary--member': presentation === 'member' }]"
  >
    <div class="squad-profile-summary__heading">
      <span v-if="presentation === 'member'" class="squad-profile-summary__identity-icon"><UIcon :name="typeMeta?.icon ?? 'i-ph-info'" aria-hidden="true" /></span>
      <UIcon v-else :name="typeMeta?.icon ?? 'i-ph-info'" aria-hidden="true" />
      <div class="squad-profile-summary__heading-copy">
        <span v-if="presentation === 'member' && name" class="squad-profile-summary__name-row"><strong>{{ name }}</strong><slot name="headingMeta" /></span>
        <strong v-else>{{ typeMeta ? t(typeMeta.labelKey) : t('squadProfile.unconfigured') }}</strong>
        <span v-if="presentation === 'member' && name">{{ memberProfileText }}</span>
      </div>
    </div>
    <div v-if="profile || $slots.facts" class="squad-profile-summary__facts">
      <span v-if="profile?.type === 'broadband'"><UIcon name="i-ph-broadcast" />{{ profile.isp }}</span>
      <span v-if="profile?.type === 'broadband' && presentation !== 'member'"><UIcon name="i-ph-gauge" />{{ profile.portMbps }} {{ t('squadProfile.mbps') }}</span>
      <span v-if="profile?.type === 'broadband'"><UIcon name="i-ph-arrows-clockwise" />{{ profile.dynamic ? t('squadProfile.dynamic') : t('squadProfile.static') }}</span>
      <span v-if="profile?.type === 'broadband' && presentation !== 'member'"><UIcon name="i-ph-map-pin" />{{ profile.location }}</span>
      <template v-else-if="profile?.type === 'china_optimized'">
        <span class="squad-profile-summary__carrier-route"><CarrierLogo carrier="telecom" :label="t('squadProfile.carriers.chinaTelecom')" />{{ profile.ct }}</span>
        <span class="squad-profile-summary__carrier-route"><CarrierLogo carrier="unicom" :label="t('squadProfile.carriers.chinaUnicom')" />{{ profile.cu }}</span>
        <span class="squad-profile-summary__carrier-route"><CarrierLogo carrier="mobile" :label="t('squadProfile.carriers.chinaMobile')" />{{ profile.cm }}</span>
        <span v-if="presentation !== 'member'"><UIcon name="i-ph-gauge" />{{ portText }}</span>
        <span v-if="presentation !== 'member'"><UIcon name="i-ph-map-pin" />{{ countryText }}</span>
      </template>
      <template v-else-if="profile?.type === 'international_network'">
        <span v-if="presentation !== 'member'"><UIcon name="i-ph-gauge" />{{ portText }}</span>
        <span v-if="presentation !== 'member'"><UIcon name="i-ph-map-pin" />{{ countryText }}</span>
        <span><UIcon name="i-ph-broadcast" />{{ profile.upstreamCarriers.join(', ') }}</span>
      </template>
      <slot name="facts" />
    </div>
    <MarkdownContent v-if="description" class="squad-profile-summary__description" :source="description" compact />
  </div>
</template>

<style scoped>
.squad-profile-summary { display: grid; gap: 0.55rem; min-width: 0; }
.squad-profile-summary__heading { display: flex; align-items: center; gap: 0.4rem; color: var(--text-muted); font-size: 0.72rem; }
.squad-profile-summary__heading-copy { display: grid; min-width: 0; gap: 0.08rem; }
.squad-profile-summary__name-row { min-width: 0; display: flex; align-items: center; justify-content: space-between; gap: 0.5rem; }
.squad-profile-summary__heading-copy strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.squad-profile-summary__facts { display: flex; flex-wrap: wrap; gap: 0.35rem 0.7rem; color: var(--text-faint); font-size: 0.68rem; line-height: 1.45; }
.squad-profile-summary__facts > span { display: inline-flex; align-items: center; gap: 0.25rem; min-width: 0; }
.squad-profile-summary__carrier-route { font-weight: 600; }
.squad-profile-summary__facts :deep(svg) { flex: 0 0 auto; }
.squad-profile-summary--compact { gap: 0.35rem; }
.squad-profile-summary--member { --squad-profile-tone: var(--accent); --squad-profile-tone-soft: color-mix(in srgb, var(--accent) 11%, var(--surface)); --squad-profile-tone-line: color-mix(in srgb, var(--accent) 32%, var(--line)); gap: 0.6rem; }
.squad-profile-summary--member.squad-profile-summary--china_optimized { --squad-profile-tone: var(--warning); --squad-profile-tone-soft: color-mix(in srgb, var(--warning) 11%, var(--surface)); --squad-profile-tone-line: color-mix(in srgb, var(--warning) 32%, var(--line)); }
.squad-profile-summary--member.squad-profile-summary--international_network { --squad-profile-tone: #9ebddd; --squad-profile-tone-soft: color-mix(in srgb, #9ebddd 11%, var(--surface)); --squad-profile-tone-line: color-mix(in srgb, #9ebddd 32%, var(--line)); }
.squad-profile-summary--member .squad-profile-summary__heading { align-items: flex-start; color: var(--text); }
.squad-profile-summary--member .squad-profile-summary__identity-icon { width: 2.5rem; height: 2.5rem; display: inline-grid; flex: 0 0 auto; place-items: center; color: var(--squad-profile-tone); font-size: 1.75rem; }
.squad-profile-summary--member .squad-profile-summary__heading-copy span { color: var(--squad-profile-tone); font-size: 0.66rem; font-weight: 700; }
.squad-profile-summary--member .squad-profile-summary__facts { gap: 0.35rem; }
.squad-profile-summary--member .squad-profile-summary__facts > span { max-width: 100%; padding: 0.27rem 0.38rem; border: 1px solid var(--squad-profile-tone-line); border-radius: 7px; color: var(--text-muted); background: color-mix(in srgb, var(--squad-profile-tone-soft) 70%, var(--surface)); }
.squad-profile-summary--member .squad-profile-summary__description { color: var(--text-muted); }
.squad-profile-summary--member.squad-profile-summary--compact { gap: 0.45rem; }
</style>
