import { computed, onMounted, readonly, shallowRef } from 'vue'

import type { AffiliateOverview, AffiliateReferralPage } from '@/api/features'
import { featuresApi } from '@/api/features'
import { localizedError } from '@/i18n'

export function useAffiliateCentre() {
  const overview = shallowRef<AffiliateOverview | null>(null)
  const referrals = shallowRef<AffiliateReferralPage | null>(null)
  const page = shallowRef(1)
  const loading = shallowRef(true)
  const referralsLoading = shallowRef(false)
  const error = shallowRef<string | null>(null)

  async function load(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const [summary, rows] = await Promise.all([featuresApi.getAffiliates(), featuresApi.getAffiliateReferrals(page.value)])
      overview.value = summary
      referrals.value = rows
    } catch (caught) {
      error.value = localizedError(caught, 'affiliates.loadFailed')
    } finally {
      loading.value = false
    }
  }

  async function setPage(next: number): Promise<void> {
    if (next === page.value || next < 1) return
    page.value = next
    referralsLoading.value = true
    try {
      referrals.value = await featuresApi.getAffiliateReferrals(next)
    } catch (caught) {
      error.value = localizedError(caught, 'affiliates.referralsFailed')
    } finally {
      referralsLoading.value = false
    }
  }

  onMounted(() => void load())
  return { overview: computed(() => overview.value), referrals: computed(() => referrals.value), page: computed(() => page.value), loading: readonly(loading), referralsLoading: readonly(referralsLoading), error: readonly(error), load, setPage }
}
