<script setup lang="ts">
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useQuestionnaires } from '@/composables/useQuestionnaires'
import QuestionnaireAccessPanel from './QuestionnaireAccessPanel.vue'

const { questionnaire, participation, loading, joining, error, load, openQuestionnaire } = useQuestionnaires()
</script>

<template>
  <div v-auto-animate class="page page--questionnaire">
    <header class="page-header">
      <h1>{{ $t('questionnaire.title') }}</h1>
    </header>
    <SkeletonBlock v-if="loading" height="24rem" />
    <template v-else-if="questionnaire">
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      <QuestionnaireAccessPanel :questionnaire="questionnaire" :participation="participation" :joining="joining" @open="openQuestionnaire" />
    </template>
    <section v-else class="section-block empty-inline">
      <div><h3>{{ $t('questionnaire.none') }}</h3><p>{{ $t('questionnaire.noneHint') }}</p></div>
      <UButton color="neutral" variant="outline" :label="$t('common.refresh')" data-haptic="refresh" @click="load" />
    </section>
  </div>
</template>
