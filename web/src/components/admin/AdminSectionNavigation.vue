<script setup lang="ts">
interface AdminNavigationSection {
  value: string
  label: string
  icon: string
}

interface AdminNavigationGroup {
  label: string
  sections: readonly AdminNavigationSection[]
}

defineProps<{
  groups: readonly AdminNavigationGroup[]
  activeSection: string
  label: string
}>()

const emit = defineEmits<{ select: [section: string] }>()
</script>

<template>
  <nav class="admin-section-navigation" :aria-label="label">
    <div
      v-for="(group, groupIndex) in groups"
      :key="group.label"
      class="admin-section-navigation__group"
      role="group"
      :aria-labelledby="`admin-section-group-${groupIndex}`"
    >
      <span :id="`admin-section-group-${groupIndex}`" class="admin-section-navigation__label">
        {{ group.label }}
      </span>
      <div class="admin-section-navigation__items">
        <UButton
          v-for="section in group.sections"
          :key="section.value"
          type="button"
          class="admin-section-navigation__link"
          :class="{ 'admin-section-navigation__link--active': activeSection === section.value }"
          :color="activeSection === section.value ? 'primary' : 'neutral'"
          :variant="activeSection === section.value ? 'soft' : 'ghost'"
          :icon="section.icon"
          :label="section.label"
          :aria-current="activeSection === section.value ? 'page' : undefined"
          data-haptic="navigate"
          @click="emit('select', section.value)"
        />
      </div>
    </div>
  </nav>
</template>

<style scoped>
.admin-section-navigation { display: none; }

@media (min-width: 900px) {
  .admin-section-navigation {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 1px;
    overflow: hidden;
    margin-bottom: 0.9rem;
    border: 1px solid var(--line);
    border-radius: var(--radius-panel);
    background: var(--line);
  }
}

.admin-section-navigation__group {
  min-width: 0;
  padding: 0.75rem;
  background: var(--surface-raised);
}

.admin-section-navigation__label {
  display: block;
  margin-bottom: 0.45rem;
  color: var(--text-faint);
  font-family: var(--font-mono);
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.admin-section-navigation__items {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.35rem;
}

.admin-section-navigation__link {
  width: 100%;
  min-width: 0;
  min-height: 44px;
  justify-content: flex-start;
}

.admin-section-navigation__link:only-child { grid-column: 1 / -1; }
.admin-section-navigation__link--active { box-shadow: inset 3px 0 0 var(--accent); }
.admin-section-navigation__link :deep([data-slot='label']) { min-width: 0; overflow-wrap: anywhere; white-space: normal; text-align: left; }
</style>
