import { computed, shallowRef, type Ref } from 'vue'

import type { FeaturePaymentMethod } from '@/api/features'
import { moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'
import { positivePaymentMinor } from './paymentOrderHelpers'

interface PaymentConfigurationOptions {
  minimumMinor?: () => string
  maximumMinor?: () => string
}

export function usePaymentConfiguration(options: PaymentConfigurationOptions, stage: Ref<string>) {
  const amount = shallowRef('20.00')
  const selectedMethodId = shallowRef<string | null>(null)
  const amountMinor = computed(() => moneyFromTxbInput(amount.value))
  const minimumMinor = computed(() => positivePaymentMinor(options.minimumMinor?.(), 100n))
  const maximumMinor = computed(() => positivePaymentMinor(options.maximumMinor?.(), 10_000_000_000n))
  const amountValid = computed(() => amountMinor.value !== ''
    && BigInt(amountMinor.value) >= minimumMinor.value
    && BigInt(amountMinor.value) <= maximumMinor.value)
  const canCreate = computed(() => amountValid.value && selectedMethodId.value !== null && stage.value === 'configure')

  function reset(methods: readonly FeaturePaymentMethod[]): void {
    if (!amountValid.value) amount.value = defaultAmount()
    const selected = methods.find((method) => method.id === selectedMethodId.value && method.available)
    selectedMethodId.value = selected?.id ?? methods.find((method) => method.available)?.id ?? null
  }

  function chooseMethod(methodId: string): void {
    if (stage.value === 'configure') selectedMethodId.value = methodId
  }

  function defaultAmount(): string {
    const preferred = 2000n
    const minor = preferred < minimumMinor.value
      ? minimumMinor.value
      : preferred > maximumMinor.value ? maximumMinor.value : preferred
    return txbInputFromMinor(minor.toString())
  }

  return {
    amount, selectedMethodId, amountMinor, minimumMinor, maximumMinor,
    amountValid, canCreate, reset, chooseMethod,
  }
}
