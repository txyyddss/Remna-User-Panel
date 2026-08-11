import { onMounted, readonly, shallowRef } from 'vue'

import type { ActiveQuestionnaire, QuestionnaireParticipation } from '@/api/features'
import { featuresApi } from '@/api/features'
import { localizedError } from '@/i18n'
import { createUuid } from '@/utils/browserCompatibility'
import { notifyHaptic, openExternalLink } from '@/utils/telegram'

export function useQuestionnaires() {
  const questionnaire = shallowRef<ActiveQuestionnaire | null>(null)
  const participation = shallowRef<QuestionnaireParticipation | null>(null)
  const loading = shallowRef(true)
  const joining = shallowRef(false)
  const error = shallowRef<string | null>(null)
  const joinKeys = new Map<string, string>()

  async function load(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      questionnaire.value = await featuresApi.getActiveQuestionnaire()
      participation.value = questionnaire.value?.participation ?? null
    } catch (caught) {
      error.value = localizedError(caught, 'errors.questionnaireUnavailable')
    } finally {
      loading.value = false
    }
  }

  async function openQuestionnaire(): Promise<void> {
    if (!questionnaire.value || joining.value) return
    joining.value = true
    error.value = null
    try {
      if (!participation.value) {
        const id = questionnaire.value.id
        const key = joinKeys.get(id) ?? createUuid()
        joinKeys.set(id, key)
        participation.value = await featuresApi.joinQuestionnaire(id, key)
        joinKeys.delete(id)
      }
      openExternalLink(questionnaire.value.formUrl)
    } catch (caught) {
      error.value = localizedError(caught, 'errors.questionnaireOpen')
      notifyHaptic('error')
    } finally {
      joining.value = false
    }
  }

  onMounted(() => void load())

  return {
    questionnaire: readonly(questionnaire),
    participation: readonly(participation),
    loading: readonly(loading),
    joining: readonly(joining),
    error: readonly(error),
    load,
    openQuestionnaire,
  }
}
