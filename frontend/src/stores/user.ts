import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api } from '@/api'

export const useUserStore = defineStore('user', () => {
    const user = ref<any>(null)
    const subscription = ref<any>(null)
    const jellyfin = ref<any>(null)
    const appConfig = ref<any>(null)
    const loading = ref(false)
    const error = ref<string | null>(null)

    const isAdmin = computed(() => user.value?.is_admin || false)
    const hasSubscription = computed(() => !!subscription.value)
    const hasJellyfin = computed(() => !!jellyfin.value)
    const credit = computed(() => user.value?.credit || 0)
    const telegramName = computed(() => user.value?.telegram_name || 'User')

    async function fetchMe() {
        loading.value = true
        error.value = null
        try {
            const data = await api.getMe()
            user.value = data.user
            subscription.value = data.subscription
            jellyfin.value = data.jellyfin
            appConfig.value = data.config
        } catch (e: any) {
            error.value = e.message
        } finally {
            loading.value = false
        }
    }

    async function refreshCredit() {
        try {
            const data = await api.getCreditBalance()
            if (user.value) {
                user.value.credit = data.balance
            }
        } catch (e) { }
    }

    return {
        user, subscription, jellyfin, appConfig, loading, error,
        isAdmin, hasSubscription, hasJellyfin, credit, telegramName,
        fetchMe, refreshCredit,
    }
})
