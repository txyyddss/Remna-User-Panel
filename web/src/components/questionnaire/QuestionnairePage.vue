<script setup lang="ts">
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useQuestionnaires } from '@/composables/useQuestionnaires'
import QuestionnaireAccessPanel from './QuestionnaireAccessPanel.vue'

const { questionnaire, participation, loading, joining, error, load, openQuestionnaire } = useQuestionnaires()
</script>

<template>
  <div class="page page--questionnaire">
    <header class="page-header">
      <p class="eyebrow">Member feedback</p>
      <h1>Share your experience.</h1>
      <p>Your private validation code connects one eligible response to one reward.</p>
    </header>
    <SkeletonBlock v-if="loading" height="24rem" />
    <template v-else-if="questionnaire">
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      <QuestionnaireAccessPanel :questionnaire="questionnaire" :participation="participation" :joining="joining" @open="openQuestionnaire" />
    </template>
    <section v-else class="section-block empty-inline">
      <div><h3>No active questionnaire</h3><p>Completed questionnaires remain in your reward history.</p></div>
      <button class="button button--secondary" type="button" @click="load">Refresh</button>
    </section>
  </div>
</template>
