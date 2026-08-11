<script setup lang="ts">
import { computed, onMounted, shallowRef } from 'vue'

import { featuresApi, type OnboardingBundle, type OnboardingLocalizedContent } from '@/api/features'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { localizedError, useI18n } from '@/i18n'
import OnboardingAgreementEditor from './onboarding/OnboardingAgreementEditor.vue'
import OnboardingWelcomeEditor from './onboarding/OnboardingWelcomeEditor.vue'

type BundleKind = 'welcome' | 'agreements'

const { t } = useI18n()
const kind = shallowRef<BundleKind>('welcome')
const bundle = shallowRef<OnboardingBundle | null>(null)
const draft = shallowRef<OnboardingLocalizedContent | null>(null)
const loading = shallowRef(true)
const busy = shallowRef<'save' | 'publish' | null>(null)
const error = shallowRef<string | null>(null)
const message = shallowRef<string | null>(null)
const kindItems = computed(() => [
  { value: 'welcome', label: t('adminOnboarding.welcome') },
  { value: 'agreements', label: t('adminOnboarding.agreements') },
])

function syncDraft(next: OnboardingBundle): void {
  draft.value = next.draft
}

async function load(nextKind: BundleKind = kind.value): Promise<void> {
  kind.value = nextKind
  loading.value = true
  error.value = null
  message.value = null
  try {
    bundle.value = await featuresApi.getAdminOnboardingBundle(nextKind)
    syncDraft(bundle.value)
  } catch (caught) {
    error.value = localizedError(caught, 'errors.adminLoad')
  } finally {
    loading.value = false
  }
}

async function save(): Promise<void> {
  if (!bundle.value || !draft.value || busy.value) return
  busy.value = 'save'
  error.value = null
  message.value = null
  try {
    bundle.value = await featuresApi.saveAdminOnboardingDraft(kind.value, bundle.value.draftRevision, draft.value)
    syncDraft(bundle.value)
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
    syncDraft(bundle.value)
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
      <UFormField :label="t('adminOnboarding.bundle')"><USelect :model-value="kind" :items="kindItems" value-key="value" @update:model-value="load($event as BundleKind)" /></UFormField>
    </div>
    <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
    <InlineNotice v-if="message" tone="success">{{ message }}</InlineNotice>
    <USkeleton v-if="loading" class="m-4 h-40" />
    <template v-else-if="bundle && draft">
      <div class="revision-strip" role="status"><UBadge color="neutral" variant="soft" :label="t('adminOnboarding.draftRevision', { revision: bundle.draftRevision })" /><UBadge color="neutral" variant="soft" :label="t('adminOnboarding.publishedRevision', { revision: bundle.publishedRevision })" /></div>
      <p class="field-hint">{{ kind === 'welcome' ? t('adminOnboarding.welcomeHint') : t('adminOnboarding.agreementHint') }}</p>
      <OnboardingWelcomeEditor v-if="kind === 'welcome'" :content="draft" @update:content="draft = $event" />
      <OnboardingAgreementEditor v-else :content="draft" @update:content="draft = $event" />
      <div class="button-row">
        <UButton color="neutral" variant="outline" icon="i-ph-floppy-disk" :disabled="Boolean(busy)" :loading="busy === 'save'" :label="busy === 'save' ? t('common.saving') : t('adminOnboarding.saveDraft')" @click="save" />
        <UButton icon="i-ph-cloud-arrow-up" :disabled="Boolean(busy)" :loading="busy === 'publish'" :label="busy === 'publish' ? t('adminOnboarding.publishing') : t('adminOnboarding.publish')" @click="publish" />
      </div>
    </template>
  </section>
</template>

<style scoped>
.onboarding-editor { display: grid; gap: 0.9rem; }
.revision-strip { display: flex; flex-wrap: wrap; gap: 0.45rem; padding: 0 1rem; }
.onboarding-editor > .field-hint, .onboarding-editor > .button-row { margin-inline: 1rem; }
</style>
