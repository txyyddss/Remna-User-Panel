import { ref } from 'vue'

export interface ToastItem {
    id: number
    message: string
    type: 'success' | 'error' | 'info'
}

const toasts = ref<ToastItem[]>([])
let nextId = 0

function addToast(message: string, type: ToastItem['type'], durationMs = 3000) {
    const id = nextId++
    toasts.value.push({ id, message, type })
    window.setTimeout(() => {
        toasts.value = toasts.value.filter((t) => t.id !== id)
    }, durationMs)
}

export function useToast() {
    return {
        toasts,
        success: (msg: string) => addToast(msg, 'success'),
        error: (msg: string) => addToast(msg, 'error'),
        info: (msg: string) => addToast(msg, 'info'),
        dismiss: (id: number) => {
            toasts.value = toasts.value.filter((t) => t.id !== id)
        },
    }
}
