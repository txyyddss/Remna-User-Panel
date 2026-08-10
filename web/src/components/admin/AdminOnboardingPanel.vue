<script setup lang="ts">
import { onMounted, reactive, shallowRef } from 'vue'
import { PhCloudArrowUp, PhFloppyDisk } from '@phosphor-icons/vue'

import { featuresApi, type OnboardingBundle, type OnboardingLocalizedContent } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { localizedError, useI18n } from '@/i18n'

type BundleKind = 'welcome' | 'agreements'

const { t } = useI18n()
const kind = shallowRef<BundleKind>('welcome')
const bundle = shallowRef<OnboardingBundle | null>(null)
const editors = reactive({ en: '[]', 'zh-CN': '[]' })
const loading = shallowRef(true)
const busy = shallowRef<'save' | 'publish' | null>(null)
const error = shallowRef<string | null>(null)
const message = shallowRef<string | null>(null)

function syncEditors(next: OnboardingBundle): void {
  editors.en = JSON.stringify(next.draft.en, null, 2)
  editors['zh-CN'] = JSON.stringify(next.draft['zh-CN'], null, 2)
}

async function load(nextKind: BundleKind = kind.value): Promise<void> {
  kind.value = nextKind
  loading.value = true
  error.value = null
  message.value = null
  try {
    bundle.value = await featuresApi.getAdminOnboardingBundle(nextKind)
    syncEditors(bundle.value)
  } catch (caught) {
    error.value = localizedError(caught, 'errors.adminLoad')
  } finally {
    loading.value = false
  }
}

function contentFromEditors(): OnboardingLocalizedContent {
  const english: unknown = JSON.parse(editors.en)
  const chinese: unknown = JSON.parse(editors['zh-CN'])
  if (!Array.isArray(english) || !Array.isArray(chinese)) throw new Error(t('adminOnboarding.arrayRequired'))
  return { en: english, 'zh-CN': chinese } as OnboardingLocalizedContent
}

async function save(): Promise<void> {
  if (!bundle.value || busy.value) return
  busy.value = 'save'
  error.value = null
  message.value = null
  try {
    bundle.value = await featuresApi.saveAdminOnboardingDraft(kind.value, bundle.value.draftRevision, contentFromEditors())
    syncEditors(bundle.value)
    message.value = t('adminOnboarding.saved')
  } catch (caught) {
    error.value = localizedError(caught, 'errors.adminAction')
  } finally {
    busy.value = null
  }
}

async function publish(): Promise<void> {
  if (!bundle.value || busy.value) return
  busy.value = 'publish'
  error.value = null
  message.value = null
  try {
    bundle.value = await featuresApi.publishAdminOnboarding(kind.value, bundle.value.draftRevision)
    syncEditors(bundle.value)
    message.value = t('adminOnboarding.published')
  } catch (caught) {
    error.value = localizedError(caught, 'errors.adminAction')
  } finally {
    busy.value = null
  }
}

onMounted(() => void load())
</script>

<template>
  <section class="admin-panel onboarding-editor">
    <div class="admin-panel__heading">
      <div><h2>{{ t('adminOnboarding.title') }}</h2><p>{{ t('adminOnboarding.copy') }}</p></div>
      <label class="compact-field"><span>{{ t('adminOnboarding.bundle') }}</span><select :value="kind" @change="load(($event.target as HTMLSelectElement).value as BundleKind)"><option value="welcome">{{ t('adminOnboarding.welcome') }}</option><option value="agreements">{{ t('adminOnboarding.agreements') }}</option></select></label>
    </div>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <InlineNotice v-if="message" tone="success">{{ message }}</InlineNotice>
    <div v-if="loading" class="admin-loading">{{ t('common.loading') }}</div>
    <template v-else-if="bundle">
      <div class="revision-strip" role="status"><span>{{ t('adminOnboarding.draftRevision', { revision: bundle.draftRevision }) }}</span><span>{{ t('adminOnboarding.publishedRevision', { revision: bundle.publishedRevision }) }}</span></div>
      <div class="locale-editor-grid">
        <label><span>{{ t('app.english') }}</span><textarea v-model="editors.en" rows="18" spellcheck="false" /></label>
        <label><span>{{ t('app.simplifiedChinese') }}</span><textarea v-model="editors['zh-CN']" rows="18" spellcheck="false" /></label>
      </div>
      <p class="field-hint">{{ kind === 'welcome' ? t('adminOnboarding.welcomeHint') : t('adminOnboarding.agreementHint') }}</p>
      <div class="button-row">
        <button class="button button--secondary" type="button" :disabled="Boolean(busy)" @click="save"><PhFloppyDisk :size="18" />{{ busy === 'save' ? t('common.saving') : t('adminOnboarding.saveDraft') }}</button>
        <button class="button button--primary" type="button" :disabled="Boolean(busy)" @click="publish"><PhCloudArrowUp :size="18" />{{ busy === 'publish' ? t('adminOnboarding.publishing') : t('adminOnboarding.publish') }}</button>
      </div>
    </template>
  </section>
</template>

<style scoped>
.onboarding-editor { display: grid; gap: 0.9rem; }
.compact-field { display: grid; gap: 0.3rem; min-width: 180px; }
.compact-field span, .locale-editor-grid label > span { color: var(--text-muted); font-size: 0.7rem; font-weight: 700; }
.revision-strip { display: flex; flex-wrap: wrap; gap: 0.45rem; padding: 0 1rem; }
.revision-strip span { padding: 0.35rem 0.55rem; border-radius: 999px; color: var(--text-muted); background: var(--surface-raised); font-size: 0.66rem; }
.locale-editor-grid { display: grid; gap: 0.8rem; padding: 0 1rem; }
.locale-editor-grid label { display: grid; gap: 0.4rem; min-width: 0; }
.locale-editor-grid textarea { width: 100%; resize: vertical; font-family: var(--font-mono); font-size: 0.7rem; line-height: 1.55; }
.onboarding-editor > .field-hint, .onboarding-editor > .button-row { margin-inline: 1rem; }
@media (min-width: 900px) { .locale-editor-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
</style>
