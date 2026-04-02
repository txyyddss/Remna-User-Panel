import { ref } from 'vue'

const visible = ref(false)
const title = ref('')
const message = ref('')
let resolveFn: ((value: boolean) => void) | null = null

export function useConfirm() {
    function confirm(opts: { title?: string; message: string }): Promise<boolean> {
        title.value = opts.title || 'Confirm'
        message.value = opts.message
        visible.value = true
        return new Promise<boolean>((resolve) => {
            resolveFn = resolve
        })
    }

    function handleResult(result: boolean) {
        visible.value = false
        if (resolveFn) {
            resolveFn(result)
            resolveFn = null
        }
    }

    return {
        visible,
        title,
        message,
        confirm,
        accept: () => handleResult(true),
        dismiss: () => handleResult(false),
    }
}
