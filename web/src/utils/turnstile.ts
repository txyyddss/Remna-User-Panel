export interface Turnstile {
  render: (element: HTMLElement, options: {
    sitekey: string; action: string; theme: 'dark'; size: 'flexible'; language: string
    callback: (token: string) => void
    'error-callback': () => void
    'expired-callback': () => void
  }) => string
  remove: (id: string) => void
  reset: (id: string) => void
}

declare global { interface Window { turnstile?: Turnstile } }
let loading: Promise<Turnstile> | null = null

export function loadTurnstile(): Promise<Turnstile> {
  if (window.turnstile) return Promise.resolve(window.turnstile)
  if (loading) return loading
  loading = new Promise<Turnstile>((resolve, reject) => {
    const script = document.createElement('script')
    const timeout = globalThis.setTimeout(() => fail(), 15_000)
    function fail(): void {
      globalThis.clearTimeout(timeout)
      script.remove()
      reject(new Error())
    }
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit'
    script.async = true
    script.onload = () => {
      globalThis.clearTimeout(timeout)
      if (window.turnstile) resolve(window.turnstile)
      else fail()
    }
    script.onerror = fail
    document.head.append(script)
  }).catch(error => { loading = null; throw error })
  return loading
}
