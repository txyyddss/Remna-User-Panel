import { onMounted, watch, type Ref } from 'vue'

import type { Combo, Purchase } from '@/api/types'

interface CatalogStepPersistenceOptions {
  activeStep: Ref<number>
  userId: Readonly<Ref<string | undefined>>
  loading: Readonly<Ref<boolean>>
  quoting: Readonly<Ref<boolean>>
  quoteUsable: Readonly<Ref<boolean>>
  purchase: Readonly<Ref<Purchase | null>>
  selectedCombo: Readonly<Ref<Combo | undefined>>
  selectedSquadIds: Readonly<Ref<readonly string[]>>
  selectedCouponGrantId: Ref<string | null>
  refreshQuote: () => Promise<boolean>
}

const catalogScrollOptions = { top: 0, left: 0, behavior: 'auto' as const }

export function useCatalogStepPersistence(options: CatalogStepPersistenceOptions) {
  const stepKey = () => options.userId.value ? `txc-catalog-step:v2:${options.userId.value}` : null

  function scrollCatalogToTop(): void {
    try {
      globalThis.scrollTo?.(catalogScrollOptions)
    } catch {
      globalThis.scrollTo?.(0, 0)
    }

    const content = globalThis.document?.querySelector<globalThis.HTMLElement>('.app-frame__content')
    if (!content) return

    try {
      content.scrollTo(catalogScrollOptions)
    } catch {
      content.scrollTop = 0
      content.scrollLeft = 0
    }
  }

  async function restoreStepQuote(): Promise<void> {
    if (
      ![3, 4].includes(options.activeStep.value)
      || options.loading.value
      || options.quoting.value
      || options.quoteUsable.value
      || options.purchase.value
      || !options.selectedCombo.value
    ) return

    await options.refreshQuote()
  }

  onMounted(() => {
    try {
      const key = stepKey()
      const value = key ? Number(globalThis.sessionStorage?.getItem(key)) : Number.NaN
      if (value >= 1 && value <= 4) options.activeStep.value = value
    } catch {
      // Storage is optional in restricted WebViews.
    }

    void restoreStepQuote()
  })

  watch(options.activeStep, (value) => {
    try {
      const key = stepKey()
      if (key) globalThis.sessionStorage?.setItem(key, String(value))
    } catch {
      // Storage is optional in restricted WebViews.
    }

    scrollCatalogToTop()
  })

  watch(
    [options.activeStep, options.loading, options.selectedCombo, options.selectedSquadIds, options.selectedCouponGrantId],
    () => { void restoreStepQuote() },
    { deep: true },
  )

  function clearPersistedStep(): void {
    try {
      const key = stepKey()
      if (key) globalThis.sessionStorage?.removeItem(key)
    } catch {
      // Storage is optional in restricted WebViews.
    }
  }

  return { clearPersistedStep }
}
