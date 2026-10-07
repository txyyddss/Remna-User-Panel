import { computed, onScopeDispose, readonly, shallowRef, watch, type Ref } from 'vue'
import { adminOperationsApi, type AdminEntitlement, type AdminEntitlementRefundQuote } from '@/api/adminOperations'
import { localizedError, t } from '@/i18n'
import { txbInputFromMinor } from '@/utils/format'

export function useAdminRefundQuote(open: Readonly<Ref<boolean>>, item: () => Pick<AdminEntitlement, 'id' | 'userId'> | null) {
  const input = shallowRef('')
  const quote = shallowRef<AdminEntitlementRefundQuote | null>(null)
  const loading = shallowRef(false)
  const error = shallowRef<string | null>(null)
  let manuallyEdited = false
  let version = 0
  const amount = computed({ get: () => input.value, set: (value: string) => { manuallyEdited = true; input.value = value } })

  async function load(): Promise<void> {
    const target = item()
    if (!open.value || !target) return
    const ticket = ++version
    loading.value = true
    error.value = null
    try {
      const next = await adminOperationsApi.getEntitlementRefundQuote(target.userId, target.id)
      if (ticket !== version || !open.value || item()?.id !== target.id || item()?.userId !== target.userId) return
      quote.value = next
      if (next.suggestedRefund) {
        if (!manuallyEdited) input.value = txbInputFromMinor(next.suggestedRefund.minor)
      } else error.value = t('adminRefundQuote.unavailable')
    } catch (caught) {
      if (ticket === version) error.value = localizedError(caught, 'adminRefundQuote.unavailable')
    } finally {
      if (ticket === version) loading.value = false
    }
  }

  watch(() => [open.value, item()?.id, item()?.userId], () => {
    version += 1
    manuallyEdited = false
    input.value = ''
    quote.value = null
    error.value = null
    loading.value = false
    if (open.value && item()) void load()
  }, { immediate: true })
  onScopeDispose(() => { version += 1 })
  return { amount, quote: readonly(quote), loading: readonly(loading), error: readonly(error), retry: load }
}
