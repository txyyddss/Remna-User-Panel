import { computed, readonly, ref } from 'vue'
import type { PaymentResponse } from '@/types'

const payment = ref<PaymentResponse | null>(null)
const visible = ref(false)

export function usePaymentSheet() {
    function openPaymentSheet(nextPayment: PaymentResponse) {
        payment.value = nextPayment
        visible.value = true
    }

    function closePaymentSheet() {
        visible.value = false
        payment.value = null
    }

    return {
        payment: readonly(payment),
        visible: readonly(visible),
        hasPayment: computed(() => visible.value && payment.value !== null),
        openPaymentSheet,
        closePaymentSheet,
    }
}
