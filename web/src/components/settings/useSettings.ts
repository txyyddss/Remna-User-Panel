import { onMounted, onScopeDispose, readonly, shallowRef } from 'vue'
import { preferencesApi, type GroupMemberTag } from '@/api/preferences'
import { memberOperationsApi } from '@/api/memberOperations'
import type { TrafficResetAutomation } from '@/api/types'
import { localizedError } from '@/i18n'

export function useSettings() {
  const automation = shallowRef<TrafficResetAutomation | null>(null)
  const automationLoading = shallowRef(true)
  const automationSaving = shallowRef(false)
  const automationError = shallowRef<string | null>(null)
  const tag = shallowRef<GroupMemberTag | null>(null)
  const tagLoading = shallowRef(true)
  const tagSaving = shallowRef(false)
  const tagError = shallowRef<string | null>(null)
  let disposed = false
  let tagRevision = 0

  async function loadAutomation(): Promise<void> {
    automationLoading.value = true
    try { automation.value = await memberOperationsApi.getTrafficResetAutomation(); automationError.value = null }
    catch (caught) { automationError.value = localizedError(caught, 'purchaseOperations.automation.loadFailed') }
    finally { automationLoading.value = false }
  }
  async function setAutomation(enabled: boolean): Promise<void> {
    if (automationSaving.value || automation.value?.enabled === enabled) return
    automationSaving.value = true
    automationError.value = null
    try { automation.value = await memberOperationsApi.updateTrafficResetAutomation(enabled) }
    catch (caught) { automationError.value = localizedError(caught, 'purchaseOperations.automation.saveFailed') }
    finally { automationSaving.value = false }
  }
  async function loadTag(): Promise<void> {
    const revision = ++tagRevision
    tagLoading.value = true
    try {
      const next = await preferencesApi.getTag()
      if (!disposed && revision === tagRevision) { tag.value = next; tagError.value = null }
    } catch (caught) {
      if (!disposed && revision === tagRevision) tagError.value = localizedError(caught, 'settings.tag.loadFailed')
    } finally {
      if (!disposed && revision === tagRevision) tagLoading.value = false
    }
  }
  async function saveTag(value: string): Promise<void> {
    if (tagSaving.value) return
    const revision = ++tagRevision
    tagSaving.value = true
    tagError.value = null
    try {
      const next = await preferencesApi.updateTag(value)
      if (!disposed && revision === tagRevision) tag.value = next
    } catch (caught) {
      if (!disposed && revision === tagRevision) tagError.value = localizedError(caught, 'settings.tag.saveFailed')
    } finally {
      if (!disposed && revision === tagRevision) tagSaving.value = false
    }
  }
  onMounted(() => { void loadAutomation(); void loadTag() })
  onScopeDispose(() => { disposed = true; tagRevision += 1 })
  return { automation: readonly(automation), automationLoading: readonly(automationLoading), automationSaving: readonly(automationSaving), automationError: readonly(automationError), tag: readonly(tag), tagLoading: readonly(tagLoading), tagSaving: readonly(tagSaving), tagError: readonly(tagError), loadAutomation, setAutomation, loadTag, saveTag }
}
