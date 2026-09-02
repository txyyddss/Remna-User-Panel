<script setup lang="ts">
const steps = ['choose', 'confirm', 'return'] as const
</script>

<template>
  <section class="community-guide" aria-labelledby="community-guide-title">
    <p class="community-guide__eyebrow">{{ $t('community.guideEyebrow') }}</p>
    <h2 id="community-guide-title">{{ $t('community.guideTitle') }}</h2>
    <p class="community-guide__description">{{ $t('community.guideDescription') }}</p>
    <ol class="community-guide__steps">
      <li v-for="(step, index) in steps" :key="step">
        <span class="community-guide__number" aria-hidden="true">{{ String(index + 1).padStart(2, '0') }}</span>
        <span>{{ $t(`community.steps.${step}`) }}</span>
      </li>
    </ol>
  </section>
</template>

<style scoped>
.community-guide {
  display: grid;
  gap: 0.5rem;
  padding: 1rem;
  border: 1px solid var(--line);
  border-left: 3px solid var(--accent);
  border-radius: var(--radius-panel);
  background: var(--surface-raised);
}

.community-guide__eyebrow,
.community-guide h2,
.community-guide__description,
.community-guide__steps { margin: 0; }
.community-guide__eyebrow { color: var(--accent); font-size: 0.68rem; font-weight: 800; letter-spacing: 0.08em; text-transform: uppercase; }
.community-guide h2 { font-size: 1.05rem; line-height: 1.2; }
.community-guide__description { color: var(--text-muted); font-size: 0.82rem; line-height: 1.45; }
.community-guide__steps { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.35rem; margin-top: 0.35rem; padding: 0; list-style: none; }
.community-guide__steps li { display: grid; align-content: start; gap: 0.35rem; color: var(--text); font-size: 0.68rem; font-weight: 700; line-height: 1.25; }
.community-guide__number { display: grid; width: 1.7rem; height: 1.7rem; flex: 0 0 auto; place-items: center; border: 1px solid var(--line-strong); border-radius: 0.55rem; color: var(--accent); background: var(--accent-soft); font-family: var(--font-mono); font-size: 0.64rem; }

@media (min-width: 900px) {
  .community-guide__steps { grid-template-columns: 1fr; gap: 0.5rem; }
  .community-guide__steps li { display: flex; align-items: center; gap: 0.55rem; font-size: 0.78rem; line-height: 1.35; }
}
</style>
