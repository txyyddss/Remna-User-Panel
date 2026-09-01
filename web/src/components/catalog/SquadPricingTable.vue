<script setup lang="ts">
import { computed } from 'vue'
import type { PricingTableSection, PricingTableTier } from '@nuxt/ui'

import type { SquadProduct } from '@/api/types'
import MarkdownContent from '@/components/common/MarkdownContent.vue'
import { useI18n } from '@/i18n'
import { formatMoney } from '@/utils/format'
import { selectionHaptic } from '@/utils/telegram'
import SquadNodeBlocks from './SquadNodeBlocks.vue'

type SquadType = NonNullable<SquadProduct['profile']>['type']
type ProfileKey = 'upstreamCarriers' | 'isp' | 'location' | 'ct' | 'cu' | 'cm'

const props = defineProps<{
  profileType: SquadType
  squads: readonly SquadProduct[]
  selectedIds: readonly string[]
  includedIds: readonly string[]
  featuredIds: readonly string[]
}>()
const emit = defineEmits<{
  toggle: [id: string]
  openGeocheck: [node: SquadProduct['accessibleNodes'][number]]
}>()
const { t } = useI18n()

const typeKey = computed(() => props.profileType === 'china_optimized'
  ? 'chinaOptimized'
  : props.profileType === 'international_network' ? 'internationalNetwork' : 'broadband')
const tiers = computed<PricingTableTier[]>(() => props.squads.map(squad => ({
  id: squad.id,
  title: squad.name,
  price: formatMoney(squad.price),
  description: squad.description,
  badge: squad.activationRequired ? t('catalog.activationRequired') : undefined,
  highlight: isSelected(squad.id),
})))
const profileFeatures = computed(() => {
  if (props.profileType === 'international_network') {
    return [feature('upstream', t('squadProfile.upstreamCarriers'), squad => profileValue(squad, 'upstreamCarriers'))]
  }
  if (props.profileType === 'broadband') {
    return [
      feature('isp', t('squadProfile.isp'), squad => profileValue(squad, 'isp')),
      feature('static-ip', t('squadProfile.static'), squad => squad.profile?.type === 'broadband' && !squad.profile.dynamic),
      feature('location', t('squadProfile.location'), squad => profileValue(squad, 'location')),
    ]
  }
  return [
    feature('ct-route', t('squadProfile.ct'), squad => profileValue(squad, 'ct')),
    feature('cu-route', t('squadProfile.cu'), squad => profileValue(squad, 'cu')),
    feature('cm-route', t('squadProfile.cm'), squad => profileValue(squad, 'cm')),
  ]
})
const sections = computed<PricingTableSection[]>(() => [{
  title: t('catalog.squadDetails'),
  features: [
    feature('remaining', t('catalog.remaining'), remaining),
    feature('port', t('squadProfile.port'), port),
    feature('description', t('catalog.description'), squad => squad.description),
    ...profileFeatures.value,
    feature('node-count', t('catalog.nodeCount'), squad => String(squad.accessibleNodes.length)),
    feature('nodes', t('catalog.nodes'), squad => squad.accessibleNodes.length > 0),
  ],
}])

function feature(id: string, title: string, value: (squad: SquadProduct) => string | boolean) {
  return { id, title, tiers: Object.fromEntries(props.squads.map(squad => [squad.id, value(squad)])) }
}
function squadFor(id: string): SquadProduct | undefined { return props.squads.find(squad => squad.id === id) }
function portFor(id: string): string { const squad = squadFor(id); return squad ? port(squad) : '' }
function remainingFor(id: string): string { const squad = squadFor(id); return squad ? remaining(squad) : '' }
function isSelected(id: string): boolean { return props.selectedIds.includes(id) }
function isIncluded(id: string): boolean { return props.includedIds.includes(id) }
function isFeatured(id: string): boolean { return props.featuredIds.includes(id) }
function isFull(squad: SquadProduct): boolean { return !isIncluded(squad.id) && squad.stockRemaining === 0 && !squad.stockHeldByCurrentUser }
function profileValue(squad: SquadProduct, key: ProfileKey): string {
  const profile = squad.profile
  if (key === 'upstreamCarriers') return profile?.type === 'international_network' ? profile.upstreamCarriers.join(', ') : ''
  if (key === 'isp') return profile?.type === 'broadband' ? profile.isp : ''
  if (key === 'location') return profile?.type === 'broadband' ? profile.location : ''
  if (key === 'ct') return profile?.type === 'china_optimized' ? profile.ct : ''
  if (key === 'cu') return profile?.type === 'china_optimized' ? profile.cu : ''
  return profile?.type === 'china_optimized' ? profile.cm : ''
}
function port(squad: SquadProduct): string {
  const profile = squad.profile
  if (!profile || !('portMbps' in profile)) return ''
  return profile.portMbps === null ? '' : `${profile.portMbps} ${t('squadProfile.mbps')}`
}
function remaining(squad: SquadProduct): string {
  if (squad.stockLimit === null || squad.stockLimit === undefined) return ''
  if (squad.stockLimit <= 0) return '0%'
  const count = Math.max(0, Math.min(squad.stockLimit, squad.stockRemaining ?? squad.stockLimit))
  return `${Math.round(count * 100 / squad.stockLimit)}%`
}
function toggle(squad: SquadProduct): void {
  if (isIncluded(squad.id) || isFull(squad)) return
  selectionHaptic()
  emit('toggle', squad.id)
}
</script>

<template>
  <section class="squad-pricing-table" :class="`squad-pricing-table--${profileType}`">
    <h3>{{ $t(`squadProfile.types.${typeKey}`) }}</h3>
    <UPricingTable :tiers="tiers" :sections="sections" :caption="$t('catalog.squadTypeCaption', { type: $t(`squadProfile.types.${typeKey}`) })">
      <template #tier-title="{ tier }">
        <span class="squad-pricing-table__title">
          <UIcon v-if="isFeatured(tier.id)" name="i-lucide-crown" :aria-label="$t('catalog.featured')" />
          <span>{{ tier.title }}</span>
        </span>
      </template>
      <template #tier-badge="{ tier }">
        <UBadge v-if="squadFor(tier.id)?.activationRequired" color="warning" variant="subtle" :label="$t('catalog.activationRequired')" />
        <UBadge v-if="squadFor(tier.id) && isFull(squadFor(tier.id)!)" color="error" variant="subtle" :label="$t('catalog.full')" />
      </template>
      <template #feature-remaining-value="{ tier }">
        <span>{{ remainingFor(tier.id) }}</span>
      </template>
      <template #feature-port-value="{ tier }">
        <span>{{ portFor(tier.id) }}</span>
      </template>
      <template #feature-description-value="{ tier }">
        <MarkdownContent v-if="squadFor(tier.id)?.description" :source="squadFor(tier.id)?.description ?? ''" compact />
      </template>
      <template #feature-nodes-value="{ tier }">
        <SquadNodeBlocks v-if="squadFor(tier.id)" :nodes="squadFor(tier.id)?.accessibleNodes ?? []" @open-geocheck="emit('openGeocheck', $event)" />
      </template>
      <template #tier-button="{ tier }">
        <UButton
          v-if="!isIncluded(tier.id)"
          block
          size="lg"
          :disabled="squadFor(tier.id) ? isFull(squadFor(tier.id)!) : true"
          :color="isSelected(tier.id) ? 'success' : 'neutral'"
          :variant="isSelected(tier.id) ? 'solid' : 'outline'"
          :label="$t(squadFor(tier.id) && isFull(squadFor(tier.id)!) ? 'catalog.full' : isSelected(tier.id) ? 'catalog.selectedBadge' : 'catalog.selectSquad')"
          :aria-pressed="isSelected(tier.id)"
          data-haptic="select"
          @click="squadFor(tier.id) && toggle(squadFor(tier.id)!)"
        />
      </template>
    </UPricingTable>
  </section>
</template>

<style scoped>
.squad-pricing-table { min-width: 0; display: grid; gap: 0.55rem; }
.squad-pricing-table > h3 { margin: 0; font-size: 0.92rem; }
.squad-pricing-table__title { min-width: 0; display: flex; align-items: center; gap: 0.35rem; }
.squad-pricing-table__title > span { overflow-wrap: anywhere; }
.squad-pricing-table__title :deep(svg) { flex: 0 0 auto; color: var(--warning); }
.squad-pricing-table :deep([data-slot='tierWrapper']) { min-width: 11rem; }
.squad-pricing-table :deep([data-slot='featureValue']) { min-width: 0; }
</style>
